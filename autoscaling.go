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

	// Context fields accumulated via With()/Named() for propagation
	// WHY: multiLogger is lazy-initialized. Fields and name must be
	// accumulated here so that ensureMultiLogger can apply them when
	// the multi-producer logger is created on first contention.
	contextFields []Field
	contextName   string

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

// ensureMultiLogger lazily initializes multi-producer logger.
// WHY propagate context: With()/Named() accumulate fields and name in
// the AdaptiveLogger. When multiLogger is created lazily on first
// contention, it must inherit the same context so that scaling
// up/down never loses structured fields or component identity.
func (al *AdaptiveLogger) ensureMultiLogger() *Logger {
	al.multiOnce.Do(func() {
		multiCfg := al.config.BaseConfig
		multiCfg.Capacity = 4096 // Larger for multi-producer
		ml, err := New(multiCfg, al.config.Options...)
		if err != nil {
			return // Fall back to single logger
		}

		// Propagate accumulated context from With()/Named() calls
		if len(al.contextFields) > 0 {
			ml = ml.With(al.contextFields...)
		}
		if al.contextName != "" {
			ml = ml.Named(al.contextName)
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

// SetLevel atomically changes the minimum logging level on both internal loggers.
// WHY both: if multiLogger was lazily initialized, its level must stay in sync
// with singleLogger. A SetLevel that only updates singleLogger would cause
// filtered messages to reappear when the system scales up under contention.
func (al *AdaptiveLogger) SetLevel(level Level) {
	al.singleLogger.SetLevel(level)
	if ml := al.multiLogger.Load(); ml != nil {
		ml.SetLevel(level)
	}
}

// Level returns the current minimum logging level.
// Reads from singleLogger (always exists, always in sync with multiLogger).
func (al *AdaptiveLogger) Level() Level {
	return al.singleLogger.Level()
}

// Sync flushes both internal loggers' ring buffers and syncs the output.
// WHY both: if multiLogger was initialized, it may have buffered records
// that must reach the output before Sync returns.
func (al *AdaptiveLogger) Sync() error {
	var firstErr error
	if err := al.singleLogger.Sync(); err != nil {
		firstErr = err
	}
	if ml := al.multiLogger.Load(); ml != nil {
		if err := ml.Sync(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// With creates a child AdaptiveLogger that includes the given fields in
// every log record. The child shares the parent's ring buffers, output,
// and scaling machinery -- only the per-record context differs.
//
// WHY accumulate-and-clone: multiLogger is lazy. If we only called
// singleLogger.With() we would lose the fields when the system scales
// up under contention and creates multiLogger. By storing contextFields
// in the AdaptiveLogger struct, ensureMultiLogger can replay them.
//
// Thread Safety: safe to call from any goroutine. The returned child
// is independent and can be used concurrently with the parent.
func (al *AdaptiveLogger) With(fields ...Field) *AdaptiveLogger {
	if len(fields) == 0 {
		return al
	}

	// Build merged field slice: parent context + new fields
	merged := make([]Field, len(al.contextFields)+len(fields))
	copy(merged, al.contextFields)
	copy(merged[len(al.contextFields):], fields)

	child := &AdaptiveLogger{
		singleLogger:  al.singleLogger.With(fields...),
		config:        al.config,
		ctx:           al.ctx,
		cancel:        al.cancel,
		contextFields: merged,
		contextName:   al.contextName,
	}

	// Propagate to existing multiLogger if already initialized
	if ml := al.multiLogger.Load(); ml != nil {
		childMulti := ml.With(fields...)
		child.multiLogger.Store(childMulti)
	}
	// WHY no multiOnce copy: child gets a fresh sync.Once so that if
	// multiLogger was not yet initialized, ensureMultiLogger will create
	// it with the full accumulated context on first contention.

	// Share parent's scaling state (atomic counters are safe to share)
	child.mode.Store(al.mode.Load())

	return child
}

// Named creates a child AdaptiveLogger with a hierarchical name.
// Names are dot-separated: parent.Named("db") on a logger named "app"
// produces "app.db". The child shares scaling machinery with the parent.
//
// WHY accumulate: same reason as With() -- multiLogger is lazy and must
// inherit the full name chain when it is eventually created.
//
// Thread Safety: safe to call from any goroutine.
func (al *AdaptiveLogger) Named(name string) *AdaptiveLogger {
	if name == "" {
		return al
	}

	// Build hierarchical name
	fullName := name
	if al.contextName != "" {
		fullName = al.contextName + "." + name
	}

	child := &AdaptiveLogger{
		singleLogger:  al.singleLogger.Named(name),
		config:        al.config,
		ctx:           al.ctx,
		cancel:        al.cancel,
		contextFields: al.contextFields,
		contextName:   fullName,
	}

	// Propagate to existing multiLogger if already initialized
	if ml := al.multiLogger.Load(); ml != nil {
		childMulti := ml.Named(name)
		child.multiLogger.Store(childMulti)
	}

	child.mode.Store(al.mode.Load())

	return child
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
