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
