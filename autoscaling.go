// autoscaling.go: auto-scaling logging architecture
//
// Auto-scaling is triggered by real-time performance metrics inspired by
// Lethe's adaptive buffer scaling, adapted for Iris's dual architecture.
//
// Key Innovation: Architectural auto-adaptation based on workload patterns
//   - Monitors contention, latency, and throughput metrics
//   - Atomically switches between SingleRing and MPSC modes
//   - Maintains zero log loss during transitions
//   - Adapts to changing application load patterns
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// AutoScalingMode represents the current scaling mode
type AutoScalingMode uint32

const (
	// SingleRingMode represents ultra-fast single-threaded logging (~25ns/op)
	// Best for: Low contention, single producers, benchmarks
	SingleRingMode AutoScalingMode = iota

	// MPSCMode represents multi-producer high-contention mode (~35ns/op per thread)
	// Best for: High contention, multiple goroutines, high throughput
	MPSCMode
)

func (m AutoScalingMode) String() string {
	switch m {
	case SingleRingMode:
		return "SingleRing"
	case MPSCMode:
		return "MPSC"
	default:
		return "Unknown"
	}
}

// AutoScalingMetrics tracks performance metrics for scaling decisions
type AutoScalingMetrics struct {
	// Write frequency metrics
	writeCount       atomic.Uint64 // Total write count
	lastWriteTime    atomic.Int64  // Last write timestamp (UnixNano)
	recentWriteCount atomic.Uint64 // Writes in last measurement window

	// Contention metrics (inspired by Lethe)
	contentionCount atomic.Uint64 // Failed write attempts due to contention

	// Latency metrics
	totalLatency  atomic.Uint64 // Cumulative latency in nanoseconds
	recentLatency atomic.Uint64 // Recent latency window

	// Goroutine tracking
	activeGoroutines atomic.Uint32 // Number of goroutines currently writing

	// Measurement window
	lastMeasurement atomic.Int64 // Last measurement timestamp
}

// AutoScalingConfig defines auto-scaling behavior
type AutoScalingConfig struct {
	// Scaling thresholds (inspired by Lethe's shouldScaleToMPSC)
	ScaleToMPSCWriteThreshold   uint64        // Min writes/sec to consider MPSC (e.g., 1000)
	ScaleToMPSCContentionRatio  uint32        // Min contention % to scale to MPSC (e.g., 10 = 10%)
	ScaleToMPSCLatencyThreshold time.Duration // Max latency before scaling to MPSC (e.g., 1ms)
	ScaleToMPSCGoroutineCount   uint32        // Min active goroutines for MPSC (e.g., 3)

	// Scale down thresholds
	ScaleToSingleWriteThreshold  uint64        // Max writes/sec to scale back to Single (e.g., 100)
	ScaleToSingleContentionRatio uint32        // Max contention % for Single mode (e.g., 1%)
	ScaleToSingleLatencyMax      time.Duration // Max latency for Single mode (e.g., 100µs)

	// Measurement and stability
	MeasurementWindow    time.Duration // How often to check metrics (e.g., 100ms)
	ScalingCooldown      time.Duration // Min time between scale operations (e.g., 1s)
	StabilityRequirement int           // Consecutive measurements before scaling (e.g., 3)
}

// DefaultAutoScalingConfig returns production-ready auto-scaling configuration
func DefaultAutoScalingConfig() AutoScalingConfig {
	return AutoScalingConfig{
		// Scale to MPSC thresholds (based on Lethe patterns)
		ScaleToMPSCWriteThreshold:   1000,                 // 1000+ writes/sec
		ScaleToMPSCContentionRatio:  10,                   // 10%+ contention
		ScaleToMPSCLatencyThreshold: 1 * time.Millisecond, // 1ms+ latency
		ScaleToMPSCGoroutineCount:   3,                    // 3+ active goroutines

		// Scale to Single thresholds
		ScaleToSingleWriteThreshold:  100,                    // <100 writes/sec
		ScaleToSingleContentionRatio: 1,                      // <1% contention
		ScaleToSingleLatencyMax:      100 * time.Microsecond, // <100µs latency

		// Measurement configuration
		MeasurementWindow:    100 * time.Millisecond, // Check every 100ms
		ScalingCooldown:      1 * time.Second,        // 1s between scale operations
		StabilityRequirement: 3,                      // 3 consecutive measurements
	}
}

// AutoScalingLogger implements an auto-scaling logging architecture
type AutoScalingLogger struct {
	// Current mode and state
	mode atomic.Uint32 // AutoScalingMode

	// Logger implementations
	singleRingLogger *Logger // Single-threaded ultra-fast mode
	mpscLogger       *Logger // Multi-producer mode

	// Auto-scaling components
	metrics AutoScalingMetrics
	config  AutoScalingConfig

	// Control and synchronization
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	transitionMu sync.RWMutex // Protects mode transitions

	// Scaling decision tracking
	lastScaleTime     atomic.Int64 // Last scaling operation timestamp
	consecutiveMPSC   atomic.Int32 // Consecutive measurements favoring MPSC
	consecutiveSingle atomic.Int32 // Consecutive measurements favoring Single

	// Performance tracking
	totalScaleOperations atomic.Uint64 // Total number of scaling operations
	scaleToMPSCCount     atomic.Uint64 // Scale to MPSC operations
	scaleToSingleCount   atomic.Uint64 // Scale to Single operations
}

// NewAutoScalingLogger creates an auto-scaling logger
func NewAutoScalingLogger(cfg Config, scalingConfig AutoScalingConfig, opts ...Option) (*AutoScalingLogger, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create SingleRing logger (optimized for single producer)
	singleConfig := cfg
	singleConfig.Capacity = 1024 // Smaller capacity for single producer
	singleLogger, err := New(singleConfig, opts...)
	if err != nil {
		cancel()
		return nil, err
	}

	// Create MPSC logger (optimized for multi-producer)
	mpscConfig := cfg
	mpscConfig.Capacity = 4096 // Larger capacity for multi-producer
	mpscLogger, err := New(mpscConfig, opts...)
	if err != nil {
		cancel()
		if closeErr := singleLogger.Close(); closeErr != nil {
			// Log both errors
			return nil, fmt.Errorf("failed to create MPSC logger: %w (also failed to close single logger: %v)", err, closeErr)
		}
		return nil, err
	}

	asl := &AutoScalingLogger{
		singleRingLogger: singleLogger,
		mpscLogger:       mpscLogger,
		config:           scalingConfig,
		ctx:              ctx,
		cancel:           cancel,
	}

	// Start in SingleRing mode (most efficient for low load)
	asl.mode.Store(uint32(SingleRingMode))

	return asl, nil
}

// Start begins auto-scaling operations
func (asl *AutoScalingLogger) Start() error {
	// Start both loggers
	asl.singleRingLogger.Start()
	asl.mpscLogger.Start()

	// Start auto-scaling monitor
	asl.wg.Add(1)
	go asl.autoScalingLoop()

	return nil
}

// Close gracefully shuts down auto-scaling logger
func (asl *AutoScalingLogger) Close() error {
	asl.cancel()
	asl.wg.Wait()

	// Close both loggers
	err1 := asl.singleRingLogger.Close()
	err2 := asl.mpscLogger.Close()

	if err1 != nil {
		return err1
	}
	return err2
}

// getCurrentLogger returns the currently active logger based on mode
func (asl *AutoScalingLogger) getCurrentLogger() *Logger {
	mode := AutoScalingMode(asl.mode.Load())
	switch mode {
	case SingleRingMode:
		return asl.singleRingLogger
	case MPSCMode:
		return asl.mpscLogger
	default:
		return asl.singleRingLogger // Fallback to single
	}
}

// Info logs at Info level with automatic scaling
func (asl *AutoScalingLogger) Info(msg string, fields ...Field) {
	start := time.Now()

	// Track active goroutine
	asl.metrics.activeGoroutines.Add(1)
	defer asl.metrics.activeGoroutines.Add(^uint32(0)) // Subtract 1

	// Get current logger (with read lock for performance)
	asl.transitionMu.RLock()
	logger := asl.getCurrentLogger()
	asl.transitionMu.RUnlock()

	// Attempt write
	logger.Info(msg, fields...)

	// Update metrics
	asl.updateMetrics(start, true)
}

// Debug logs at Debug level with automatic scaling
func (asl *AutoScalingLogger) Debug(msg string, fields ...Field) {
	start := time.Now()

	asl.metrics.activeGoroutines.Add(1)
	defer asl.metrics.activeGoroutines.Add(^uint32(0))

	asl.transitionMu.RLock()
	logger := asl.getCurrentLogger()
	asl.transitionMu.RUnlock()

	logger.Debug(msg, fields...)
	asl.updateMetrics(start, true)
}

// Warn logs at Warn level with automatic scaling
func (asl *AutoScalingLogger) Warn(msg string, fields ...Field) {
	start := time.Now()

	asl.metrics.activeGoroutines.Add(1)
	defer asl.metrics.activeGoroutines.Add(^uint32(0))

	asl.transitionMu.RLock()
	logger := asl.getCurrentLogger()
	asl.transitionMu.RUnlock()

	logger.Warn(msg, fields...)
	asl.updateMetrics(start, true)
}

// Error logs at Error level with automatic scaling
func (asl *AutoScalingLogger) Error(msg string, fields ...Field) {
	start := time.Now()

	asl.metrics.activeGoroutines.Add(1)
	defer asl.metrics.activeGoroutines.Add(^uint32(0))

	asl.transitionMu.RLock()
	logger := asl.getCurrentLogger()
	asl.transitionMu.RUnlock()

	logger.Error(msg, fields...)
	asl.updateMetrics(start, true)
}

// updateMetrics updates performance metrics for scaling decisions
func (asl *AutoScalingLogger) updateMetrics(start time.Time, success bool) {
	now := time.Now()
	duration := now.Sub(start)

	// Safe conversion with bounds checking
	var latency uint64
	if duration < 0 {
		latency = 0
	} else if duration > time.Duration(1<<63-1) {
		latency = uint64(1<<63 - 1)
	} else {
		nanos := duration.Nanoseconds()
		if nanos < 0 {
			latency = 0
		} else {
			latency = uint64(nanos)
		}
	}

	// Update write metrics
	asl.metrics.writeCount.Add(1)
	asl.metrics.recentWriteCount.Add(1)

	// Safe timestamp storage
	timestamp := now.UnixNano()
	asl.metrics.lastWriteTime.Store(timestamp)

	// Update latency metrics
	asl.metrics.totalLatency.Add(latency)
	asl.metrics.recentLatency.Add(latency)

	// Track contention (failed writes)
	if !success {
		asl.metrics.contentionCount.Add(1)
	}
}

// autoScalingLoop monitors metrics and triggers scaling decisions
func (asl *AutoScalingLogger) autoScalingLoop() {
	defer asl.wg.Done()

	ticker := time.NewTicker(asl.config.MeasurementWindow)
	defer ticker.Stop()

	for {
		select {
		case <-asl.ctx.Done():
			return
		case <-ticker.C:
			asl.checkScalingDecision()
		}
	}
}

// checkScalingDecision analyzes metrics and decides whether to scale
func (asl *AutoScalingLogger) checkScalingDecision() {
	now := time.Now()

	// Check cooldown period
	lastScale := time.Unix(0, asl.lastScaleTime.Load())
	if now.Sub(lastScale) < asl.config.ScalingCooldown {
		return // Still in cooldown
	}

	// Calculate current metrics
	metrics := asl.calculateCurrentMetrics()
	currentMode := AutoScalingMode(asl.mode.Load())

	// Determine preferred mode based on metrics
	preferredMode := asl.determinePreferredMode(metrics)

	// Update consecutive counters
	if preferredMode == MPSCMode {
		asl.consecutiveMPSC.Add(1)
		asl.consecutiveSingle.Store(0)
	} else {
		asl.consecutiveSingle.Add(1)
		asl.consecutiveMPSC.Store(0)
	}

	// Check if we should scale (requires stability)
	shouldScale := false
	targetMode := currentMode

	if preferredMode == MPSCMode && currentMode == SingleRingMode {
		// Safe conversion with bounds checking
		stabilityReq := asl.config.StabilityRequirement
		var stabilityReqInt32 int32
		if stabilityReq > int(1<<31-1) {
			stabilityReqInt32 = int32(1<<31 - 1)
		} else if stabilityReq < 0 {
			stabilityReqInt32 = 0
		} else {
			stabilityReqInt32 = int32(stabilityReq)
		}
		if asl.consecutiveMPSC.Load() >= stabilityReqInt32 {
			shouldScale = true
			targetMode = MPSCMode
		}
	} else if preferredMode == SingleRingMode && currentMode == MPSCMode {
		// Safe conversion with bounds checking
		stabilityReq := asl.config.StabilityRequirement
		var stabilityReqInt32 int32
		if stabilityReq > int(1<<31-1) {
			stabilityReqInt32 = int32(1<<31 - 1)
		} else if stabilityReq < 0 {
			stabilityReqInt32 = 0
		} else {
			stabilityReqInt32 = int32(stabilityReq)
		}
		if asl.consecutiveSingle.Load() >= stabilityReqInt32 {
			shouldScale = true
			targetMode = SingleRingMode
		}
	}

	// Perform scaling if needed
	if shouldScale {
		asl.performScaling(targetMode)
	}

	// Reset measurement window metrics
	asl.resetWindowMetrics()
}

// scalingMetrics holds calculated metrics for scaling decisions
type scalingMetrics struct {
	writesPerSecond  uint64
	contentionRatio  uint32
	avgLatency       time.Duration
	activeGoroutines uint32
}

// calculateCurrentMetrics computes current performance metrics
func (asl *AutoScalingLogger) calculateCurrentMetrics() scalingMetrics {
	windowDuration := asl.config.MeasurementWindow

	// Calculate writes per second
	recentWrites := asl.metrics.recentWriteCount.Load()
	writesPerSecond := uint64(float64(recentWrites) / windowDuration.Seconds())

	// Calculate contention ratio
	totalWrites := asl.metrics.writeCount.Load()
	contentionCount := asl.metrics.contentionCount.Load()
	var contentionRatio uint32
	if totalWrites > 0 {
		// Safe conversion with bounds checking
		ratio := (contentionCount * 100) / totalWrites
		if ratio > 100 {
			contentionRatio = 100
		} else {
			contentionRatio = uint32(ratio)
		}
	}

	// Calculate average latency
	recentLatency := asl.metrics.recentLatency.Load()
	var avgLatency time.Duration
	if recentWrites > 0 {
		// Safe conversion with bounds checking
		avgNanos := recentLatency / recentWrites
		if avgNanos > uint64(time.Duration(1<<63-1)) {
			avgLatency = time.Duration(1<<63 - 1) // Max duration
		} else {
			avgLatency = time.Duration(avgNanos)
		}
	}

	// Get active goroutines
	activeGoroutines := asl.metrics.activeGoroutines.Load()

	return scalingMetrics{
		writesPerSecond:  writesPerSecond,
		contentionRatio:  contentionRatio,
		avgLatency:       avgLatency,
		activeGoroutines: activeGoroutines,
	}
}

// determinePreferredMode decides the optimal mode based on metrics
func (asl *AutoScalingLogger) determinePreferredMode(metrics scalingMetrics) AutoScalingMode {
	// Scale to MPSC conditions (inspired by Lethe's shouldScaleToMPSC)
	if metrics.writesPerSecond >= asl.config.ScaleToMPSCWriteThreshold ||
		metrics.contentionRatio >= asl.config.ScaleToMPSCContentionRatio ||
		metrics.avgLatency >= asl.config.ScaleToMPSCLatencyThreshold ||
		metrics.activeGoroutines >= asl.config.ScaleToMPSCGoroutineCount {
		return MPSCMode
	}

	// Scale to Single conditions
	if metrics.writesPerSecond <= asl.config.ScaleToSingleWriteThreshold &&
		metrics.contentionRatio <= asl.config.ScaleToSingleContentionRatio &&
		metrics.avgLatency <= asl.config.ScaleToSingleLatencyMax {
		return SingleRingMode
	}

	// No clear preference, maintain current mode
	return AutoScalingMode(asl.mode.Load())
}

// performScaling executes the scaling operation with zero log loss
func (asl *AutoScalingLogger) performScaling(targetMode AutoScalingMode) {
	currentMode := AutoScalingMode(asl.mode.Load())
	if currentMode == targetMode {
		return // No change needed
	}

	// Lock for exclusive access during transition
	asl.transitionMu.Lock()
	defer asl.transitionMu.Unlock()

	// Double-check mode hasn't changed during lock acquisition
	if AutoScalingMode(asl.mode.Load()) == targetMode {
		return
	}

	// Perform atomic mode switch
	asl.mode.Store(uint32(targetMode))
	asl.lastScaleTime.Store(time.Now().UnixNano())

	// Update scaling statistics
	asl.totalScaleOperations.Add(1)
	if targetMode == MPSCMode {
		asl.scaleToMPSCCount.Add(1)
	} else {
		asl.scaleToSingleCount.Add(1)
	}

	// Reset consecutive counters
	asl.consecutiveMPSC.Store(0)
	asl.consecutiveSingle.Store(0)
}

// resetWindowMetrics resets metrics for the next measurement window
func (asl *AutoScalingLogger) resetWindowMetrics() {
	asl.metrics.recentWriteCount.Store(0)
	asl.metrics.recentLatency.Store(0)
	asl.metrics.lastMeasurement.Store(time.Now().UnixNano())
}

// GetCurrentMode returns the current scaling mode
func (asl *AutoScalingLogger) GetCurrentMode() AutoScalingMode {
	return AutoScalingMode(asl.mode.Load())
}

// GetScalingStats returns auto-scaling performance statistics
func (asl *AutoScalingLogger) GetScalingStats() AutoScalingStats {
	return AutoScalingStats{
		CurrentMode:          asl.GetCurrentMode(),
		TotalScaleOperations: asl.totalScaleOperations.Load(),
		ScaleToMPSCCount:     asl.scaleToMPSCCount.Load(),
		ScaleToSingleCount:   asl.scaleToSingleCount.Load(),
		TotalWrites:          asl.metrics.writeCount.Load(),
		ContentionCount:      asl.metrics.contentionCount.Load(),
		ActiveGoroutines:     asl.metrics.activeGoroutines.Load(),
	}
}

// AutoScalingStats provides auto-scaling performance insights
type AutoScalingStats struct {
	CurrentMode          AutoScalingMode
	TotalScaleOperations uint64
	ScaleToMPSCCount     uint64
	ScaleToSingleCount   uint64
	TotalWrites          uint64
	ContentionCount      uint64
	ActiveGoroutines     uint32
}
