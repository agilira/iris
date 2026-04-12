// autoscaling.go: Lazy dual-mode auto-scaling logger
//
// Implements zero-compromise auto-scaling:
// - Starts in SingleRing mode (25ns/op) for single-producer scenarios
// - Lazily initializes MPSC mode only when multi-producer detected
// - Automatic switching based on goroutine contention
// - Zero log loss during transitions
//
// Design Philosophy:
// - Present: Optimal for agentic OS (single producer, low latency)
// - Future: Ready for robotics (multi-producer burst scenarios)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// ScalingMode represents the current scaling mode
type ScalingMode uint32

const (
	// SingleMode - ultra-fast single-producer (~25ns/op)
	SingleMode ScalingMode = iota
	// MultiMode - multi-producer optimized (~35ns/op)
	MultiMode
)

func (m ScalingMode) String() string {
	if m == SingleMode {
		return "Single"
	}
	return "Multi"
}

// ScalerConfig defines auto-scaling behavior
type ScalerConfig struct {
	// Threshold: concurrent goroutines to trigger scale-up
	GoroutineThreshold uint32
	// Cooldown before scaling back down
	ScaleDownCooldown time.Duration
	// Base config for logger creation
	BaseConfig Config
	// Options for logger creation
	Options []Option
}

// DefaultScalerConfig returns production-ready defaults
func DefaultScalerConfig(cfg Config, opts ...Option) ScalerConfig {
	return ScalerConfig{
		GoroutineThreshold: 2,               // Scale up when 2+ goroutines
		ScaleDownCooldown:  5 * time.Second, // Stay in multi-mode for 5s
		BaseConfig:         cfg,
		Options:            opts,
	}
}

// AdaptiveLogger provides lazy dual-mode auto-scaling
type AdaptiveLogger struct {
	// Mode tracking
	mode atomic.Uint32

	// Loggers (singleLogger always exists, multiLogger lazy-initialized)
	singleLogger *Logger
	multiLogger  atomic.Pointer[Logger]
	multiOnce    sync.Once

	// Contention detection
	activeWriters atomic.Uint32
	lastMultiUse  atomic.Int64

	// Configuration
	config ScalerConfig

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Stats
	scaleUpCount   atomic.Uint64
	scaleDownCount atomic.Uint64
	totalWrites    atomic.Uint64
}

// NewAdaptiveLogger creates a lazy dual-mode auto-scaling logger
func NewAdaptiveLogger(cfg ScalerConfig) (*AdaptiveLogger, error) {
	// Create single-producer logger (always exists)
	singleCfg := cfg.BaseConfig
	singleCfg.Capacity = 1024 // Smaller for single producer
	singleLogger, err := New(singleCfg, cfg.Options...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	al := &AdaptiveLogger{
		singleLogger: singleLogger,
		config:       cfg,
		ctx:          ctx,
		cancel:       cancel,
	}
	al.mode.Store(uint32(SingleMode))

	return al, nil
}

// Start begins logger operations
func (al *AdaptiveLogger) Start() error {
	al.singleLogger.Start()

	// Start scale-down monitor
	al.wg.Add(1)
	go al.scaleDownMonitor()

	return nil
}

// Close gracefully shuts down
func (al *AdaptiveLogger) Close() error {
	al.cancel()
	al.wg.Wait()

	var err error
	if e := al.singleLogger.Close(); e != nil {
		err = e
	}
	if ml := al.multiLogger.Load(); ml != nil {
		if e := ml.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

// ensureMultiLogger lazily initializes multi-producer logger
func (al *AdaptiveLogger) ensureMultiLogger() *Logger {
	al.multiOnce.Do(func() {
		multiCfg := al.config.BaseConfig
		multiCfg.Capacity = 4096 // Larger for multi-producer
		ml, err := New(multiCfg, al.config.Options...)
		if err != nil {
			return // Fall back to single logger
		}
		ml.Start()
		al.multiLogger.Store(ml)
	})
	return al.multiLogger.Load()
}

// getLogger returns the appropriate logger based on current mode and contention
func (al *AdaptiveLogger) getLogger() *Logger {
	// Track active writers for contention detection
	writers := al.activeWriters.Add(1)

	// Check if we should scale up
	if writers >= al.config.GoroutineThreshold {
		if ScalingMode(al.mode.Load()) == SingleMode {
			// Scale up to multi-mode
			if ml := al.ensureMultiLogger(); ml != nil {
				al.mode.Store(uint32(MultiMode))
				al.scaleUpCount.Add(1)
			}
		}
		al.lastMultiUse.Store(time.Now().UnixNano())
	}

	// Return appropriate logger
	if ScalingMode(al.mode.Load()) == MultiMode {
		if ml := al.multiLogger.Load(); ml != nil {
			return ml
		}
	}
	return al.singleLogger
}

// releaseWriter decrements active writer count
func (al *AdaptiveLogger) releaseWriter() {
	al.activeWriters.Add(^uint32(0)) // Subtract 1
	al.totalWrites.Add(1)
}

// Info logs at Info level with automatic scaling
func (al *AdaptiveLogger) Info(msg string, fields ...Field) {
	logger := al.getLogger()
	defer al.releaseWriter()
	logger.Info(msg, fields...)
}

// Debug logs at Debug level
func (al *AdaptiveLogger) Debug(msg string, fields ...Field) {
	logger := al.getLogger()
	defer al.releaseWriter()
	logger.Debug(msg, fields...)
}

// Warn logs at Warn level
func (al *AdaptiveLogger) Warn(msg string, fields ...Field) {
	logger := al.getLogger()
	defer al.releaseWriter()
	logger.Warn(msg, fields...)
}

// Error logs at Error level
func (al *AdaptiveLogger) Error(msg string, fields ...Field) {
	logger := al.getLogger()
	defer al.releaseWriter()
	logger.Error(msg, fields...)
}

// scaleDownMonitor checks if we can scale back to single mode
func (al *AdaptiveLogger) scaleDownMonitor() {
	defer al.wg.Done()
	ticker := time.NewTicker(al.config.ScaleDownCooldown / 2)
	defer ticker.Stop()

	for {
		select {
		case <-al.ctx.Done():
			return
		case <-ticker.C:
			al.checkScaleDown()
		}
	}
}

// checkScaleDown scales back to single mode if idle
func (al *AdaptiveLogger) checkScaleDown() {
	if ScalingMode(al.mode.Load()) != MultiMode {
		return
	}

	lastUse := time.Unix(0, al.lastMultiUse.Load())
	if time.Since(lastUse) > al.config.ScaleDownCooldown {
		// No multi-producer activity, scale down
		al.mode.Store(uint32(SingleMode))
		al.scaleDownCount.Add(1)
	}
}

// Mode returns the current scaling mode
func (al *AdaptiveLogger) Mode() ScalingMode {
	return ScalingMode(al.mode.Load())
}

// Stats returns scaling statistics
func (al *AdaptiveLogger) Stats() AdaptiveStats {
	return AdaptiveStats{
		Mode:           al.Mode(),
		ScaleUpCount:   al.scaleUpCount.Load(),
		ScaleDownCount: al.scaleDownCount.Load(),
		TotalWrites:    al.totalWrites.Load(),
		ActiveWriters:  al.activeWriters.Load(),
	}
}

// AdaptiveStats provides scaling insights
type AdaptiveStats struct {
	Mode           ScalingMode
	ScaleUpCount   uint64
	ScaleDownCount uint64
	TotalWrites    uint64
	ActiveWriters  uint32
}

// =============================================================================
// Backward Compatibility Aliases
// =============================================================================

// AutoScalingMode is an alias for ScalingMode (backward compatibility)
// Deprecated: Use ScalingMode instead
type AutoScalingMode = ScalingMode

const (
	// SingleRingMode is an alias for SingleMode (backward compatibility)
	// Deprecated: Use SingleMode instead
	SingleRingMode = SingleMode
	// MPSCMode is an alias for MultiMode (backward compatibility)
	// Deprecated: Use MultiMode instead
	MPSCMode = MultiMode
)

// AutoScalingConfig provides backward compatibility with old API
// Deprecated: Use ScalerConfig instead
type AutoScalingConfig struct {
	ScaleToMPSCWriteThreshold    uint64
	ScaleToMPSCContentionRatio   uint32
	ScaleToMPSCLatencyThreshold  time.Duration
	ScaleToMPSCGoroutineCount    uint32
	ScaleToSingleWriteThreshold  uint64
	ScaleToSingleContentionRatio uint32
	ScaleToSingleLatencyMax      time.Duration
	MeasurementWindow            time.Duration
	ScalingCooldown              time.Duration
	StabilityRequirement         int
}

// AutoScalingLogger is an alias for AdaptiveLogger (backward compatibility)
// Deprecated: Use AdaptiveLogger instead
type AutoScalingLogger = AdaptiveLogger

// AutoScalingStats provides backward compatibility
// Deprecated: Use AdaptiveStats instead
type AutoScalingStats struct {
	CurrentMode          ScalingMode
	TotalScaleOperations uint64
	ScaleToMPSCCount     uint64
	ScaleToSingleCount   uint64
	TotalWrites          uint64
	ContentionCount      uint64
	ActiveGoroutines     uint32
}

// DefaultAutoScalingConfig returns a legacy-compatible config
// Deprecated: Use DefaultScalerConfig instead
func DefaultAutoScalingConfig() AutoScalingConfig {
	return AutoScalingConfig{
		ScaleToMPSCWriteThreshold:    1000,
		ScaleToMPSCContentionRatio:   10,
		ScaleToMPSCLatencyThreshold:  1 * time.Millisecond,
		ScaleToMPSCGoroutineCount:    3,
		ScaleToSingleWriteThreshold:  100,
		ScaleToSingleContentionRatio: 1,
		ScaleToSingleLatencyMax:      100 * time.Microsecond,
		MeasurementWindow:            100 * time.Millisecond,
		ScalingCooldown:              1 * time.Second,
		StabilityRequirement:         3,
	}
}

// NewAutoScalingLogger creates an AdaptiveLogger (backward compatibility)
// Deprecated: Use NewAdaptiveLogger instead
func NewAutoScalingLogger(cfg Config, scalingConfig AutoScalingConfig, opts ...Option) (*AdaptiveLogger, error) {
	sc := ScalerConfig{
		GoroutineThreshold: scalingConfig.ScaleToMPSCGoroutineCount,
		ScaleDownCooldown:  scalingConfig.ScalingCooldown,
		BaseConfig:         cfg,
		Options:            opts,
	}
	if sc.GoroutineThreshold == 0 {
		sc.GoroutineThreshold = 2
	}
	if sc.ScaleDownCooldown == 0 {
		sc.ScaleDownCooldown = 5 * time.Second
	}
	return NewAdaptiveLogger(sc)
}

// GetCurrentMode returns Mode() (backward compatibility)
// Deprecated: Use Mode() instead
func (al *AdaptiveLogger) GetCurrentMode() ScalingMode {
	return al.Mode()
}

// GetScalingStats returns legacy-compatible stats
// Deprecated: Use Stats() instead
func (al *AdaptiveLogger) GetScalingStats() AutoScalingStats {
	s := al.Stats()
	return AutoScalingStats{
		CurrentMode:          s.Mode,
		TotalScaleOperations: s.ScaleUpCount + s.ScaleDownCount,
		ScaleToMPSCCount:     s.ScaleUpCount,
		ScaleToSingleCount:   s.ScaleDownCount,
		TotalWrites:          s.TotalWrites,
		ContentionCount:      0, // Not tracked in new implementation
		ActiveGoroutines:     s.ActiveWriters,
	}
}
