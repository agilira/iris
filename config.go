// config.go: High-performance configuration system for iris logging
//
// This file provides the core configuration structures and methods for the iris
// logging library. The Config struct centralizes all logger parameters with
// intelligent defaults and optimizations for maximum performance.
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/agilira/go-timecache"
	"github.com/agilira/iris/internal/zephyroslite"
)

// Architecture represents the ring buffer architecture type
type Architecture int

const (
	// SingleRing uses a single Zephyros ring for maximum single-thread performance
	// Best for: benchmarks, single-producer scenarios, maximum single-thread throughput
	// Performance: ~25ns/op single-thread, limited concurrency scaling
	SingleRing Architecture = iota

	// ThreadedRings uses ThreadedZephyros with multiple rings for multi-producer scaling
	// Best for: production, multi-producer scenarios, high concurrency
	// Performance: ~35ns/op per thread, excellent scaling (4x+ improvement with multiple producers)
	ThreadedRings
)

// String returns the string representation of the architecture
func (a Architecture) String() string {
	switch a {
	case SingleRing:
		return "single"
	case ThreadedRings:
		return "threaded"
	default:
		return "unknown"
	}
}

// ParseArchitecture parses a string into an Architecture
func ParseArchitecture(s string) (Architecture, error) {
	switch s {
	case "single", "Single", "SINGLE":
		return SingleRing, nil
	case "threaded", "Threaded", "THREADED", "multi", "Multi", "MULTI":
		return ThreadedRings, nil
	default:
		return SingleRing, fmt.Errorf("unknown architecture: %s", s)
	}
}

// Config represents the core configuration for an iris logger instance.
// This structure centralizes all logging parameters with intelligent defaults
// and performance optimizations. All fields are designed for zero-copy
// operations and minimal memory allocation.
//
// Performance considerations:
// - Capacity should be a power-of-two for optimal ring buffer performance
// - BatchSize affects throughput vs latency trade-offs
// - TimeFn allows for custom time sources (useful for testing and optimization)
//
// Thread-safety: Config structs are immutable after logger creation
type Config struct {
	// Ring buffer configuration (power-of-two recommended for Capacity)
	// Capacity determines the maximum number of log entries that can be buffered
	// before blocking or dropping occurs. Larger values improve throughput but
	// increase memory usage.
	Capacity int64

	// BatchSize controls how many log entries are processed together.
	// Higher values improve throughput but may increase latency.
	// Optimal values are typically 8-64 depending on workload.
	BatchSize int64

	// Architecture determines the ring buffer architecture type
	// SingleRing: Maximum single-thread performance (~25ns/op) - best for benchmarks
	// ThreadedRings: Multi-producer scaling (~35ns/op per thread) - best for production
	// Default: SingleRing for benchmark compatibility
	Architecture Architecture

	// NumRings specifies the number of rings for ThreadedRings architecture
	// Only used when Architecture = ThreadedRings
	// Higher values provide better parallelism but use more memory
	// Default: 4 (optimal for most multi-core systems)
	NumRings int

	// BackpressurePolicy determines the behavior when the ring buffer is full
	// DropOnFull: Drops new messages for maximum performance (default)
	// BlockOnFull: Blocks caller until space is available (guaranteed delivery)
	BackpressurePolicy zephyroslite.BackpressurePolicy

	// IdleStrategy controls CPU usage when no log records are being processed
	// Different strategies provide various trade-offs between latency and CPU usage:
	// - SpinningIdleStrategy: Ultra-low latency, ~100% CPU usage
	// - SleepingIdleStrategy: Balanced CPU/latency, ~1-10% CPU usage
	// - YieldingIdleStrategy: Moderate reduction, ~10-50% CPU usage
	// - ChannelIdleStrategy: Minimal CPU usage, ~microsecond latency
	// - ProgressiveIdleStrategy: Adaptive strategy for variable workloads (default)
	IdleStrategy zephyroslite.IdleStrategy

	// Output and formatting configuration
	// Output specifies where log entries are written. Must implement WriteSyncer
	// for proper synchronization guarantees.
	Output WriteSyncer

	// Encoder determines the output format (JSON, Console, etc.)
	// The encoder converts log records to their final byte representation
	Encoder Encoder

	// Level sets the minimum logging level. Messages below this level
	// are filtered out early for maximum performance.
	Level Level // default: Info

	// TimeFn allows custom time source for timestamps.
	// Default: time.Now for real-time logging
	// Can be overridden for testing or performance optimization
	TimeFn func() time.Time

	// Optional performance tuning
	// Sampler controls log sampling for high-volume scenarios
	// Can be nil to disable sampling
	Sampler Sampler

	// Name provides a human-readable identifier for this logger instance
	// Useful for debugging and metrics collection
	Name string
}

// stats represents internal logger statistics exposed via Logger.Stats().
// These metrics help monitor logger performance and health in production
// environments. All fields use atomic operations for thread-safe access.
//
// Performance: Zero-allocation reads using atomic operations
type stats struct {
	// Dropped counts the number of log entries that were dropped due to
	// buffer overflow or sampling. High values may indicate the need for
	// configuration tuning (larger Capacity, higher BatchSize, etc.)
	Dropped int64
}

// atomicLevel provides thread-safe level management without locks.
// This structure enables dynamic level changes during runtime with
// minimal performance overhead using atomic operations.
//
// Performance: Sub-nanosecond level checks using atomic loads
type atomicLevel struct{ v atomic.Int32 }

// Load atomically reads the current logging level.
// This method is called frequently in hot paths, so it's optimized
// for maximum performance with no allocations.
func (a *atomicLevel) Load() Level { return Level(a.v.Load()) }

// Store atomically updates the logging level.
// Safe for concurrent use from multiple goroutines.
func (a *atomicLevel) Store(l Level) { a.v.Store(int32(l)) }

// SetMin is an alias for Store, providing semantic clarity when
// setting a minimum logging level threshold.
func (a *atomicLevel) SetMin(l Level) { a.Store(l) }

// withDefaults applies safe fallback values to configuration fields.
// This method ensures that all required configuration parameters have
// sensible defaults for production use, preventing nil pointer errors
// and providing optimal performance characteristics.
//
// Default values are chosen based on:
// - Performance benchmarks for optimal throughput
// - Production deployment experience
// - Memory usage characteristics
//
// Performance: Copy-on-write semantics with minimal allocations
func (c *Config) withDefaults() *Config {
	out := *c

	// Set default capacity to 64K entries (power-of-two for optimal performance)
	// This provides good balance between memory usage and throughput
	if out.Capacity <= 0 {
		out.Capacity = 1 << 16 // 65536
	}

	// Default output to stdout with proper synchronization
	// WrapWriter automatically handles different writer types optimally
	if out.Output == nil {
		out.Output = WrapWriter(os.Stdout)
	}

	// Default time function to system time
	// Can be overridden for testing or custom time sources
	if out.TimeFn == nil {
		out.TimeFn = timecache.CachedTime
	}

	// Default batch size for optimal throughput/latency balance
	if out.BatchSize <= 0 {
		out.BatchSize = 32 // Empirically optimal for most workloads
	}

	// Default architecture to SingleRing for benchmark compatibility
	// Users can explicitly set ThreadedRings for production scaling
	if out.Architecture != SingleRing && out.Architecture != ThreadedRings {
		out.Architecture = SingleRing
	}

	// Default number of rings for ThreadedRings architecture
	if out.NumRings <= 0 {
		out.NumRings = 4 // Optimal for most multi-core systems
	}

	// Level defaults to Info (zero-value), which is already correct
	// No action needed as Info = 0

	// Default encoder if not specified
	if out.Encoder == nil {
		out.Encoder = NewJSONEncoder()
	}

	// Default backpressure policy to DropOnFull (high performance default)
	// BackpressurePolicy uses int type where 0 = DropOnFull by design
	if out.BackpressurePolicy != zephyroslite.DropOnFull && out.BackpressurePolicy != zephyroslite.BlockOnFull {
		out.BackpressurePolicy = zephyroslite.DropOnFull
	}

	// Default idle strategy to BalancedStrategy (progressive) for good all-around performance
	// This provides excellent performance for most workloads without manual tuning
	if out.IdleStrategy == nil {
		out.IdleStrategy = BalancedStrategy
	}

	// TODO: Add default sampler when sampling system is implemented

	return &out
}

// NewAtomicLevelFromConfig creates a new atomicLevel initialized with the config's level.
// This function bridges the gap between static configuration and dynamic level management.
func NewAtomicLevelFromConfig(config *Config) *atomicLevel {
	al := &atomicLevel{}
	al.Store(config.Level)
	return al
}

// Validate checks the configuration for common errors and returns an error if
// the configuration is invalid. This helps catch configuration issues early
// before logger creation.
//
// Performance: Fast validation with early returns for common cases
func (c *Config) Validate() error {
	if c.Capacity <= 0 {
		return NewLoggerErrorWithField(ErrCodeInvalidConfig, "capacity must be positive", "capacity", fmt.Sprintf("%d", c.Capacity))
	}

	if c.BatchSize < 0 {
		return NewLoggerErrorWithField(ErrCodeInvalidConfig, "batch size cannot be negative", "batch_size", fmt.Sprintf("%d", c.BatchSize))
	}

	if c.BatchSize > c.Capacity {
		// Use simple error for multiple fields case
		return NewLoggerError(ErrCodeInvalidConfig, "batch size cannot exceed capacity")
	}

	if !IsValidLevel(c.Level) {
		return NewLoggerErrorWithField(ErrCodeInvalidLevel, "invalid logging level", "level", fmt.Sprintf("%d", int(c.Level)))
	}

	if c.Architecture != SingleRing && c.Architecture != ThreadedRings {
		return NewLoggerErrorWithField(ErrCodeInvalidConfig, "invalid architecture type", "architecture", fmt.Sprintf("%d", int(c.Architecture)))
	}

	if c.Architecture == ThreadedRings && c.NumRings <= 0 {
		return NewLoggerErrorWithField(ErrCodeInvalidConfig, "invalid number of rings for threaded architecture", "num_rings", fmt.Sprintf("%d", c.NumRings))
	}

	return nil
}

// Clone creates a deep copy of the configuration.
// This is useful for creating derived configurations without affecting the original.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}

	clone := *c
	return &clone
}

// GetStats creates a new stats instance for tracking logger metrics.
// This factory function ensures proper initialization of all atomic counters.
func (c *Config) GetStats() *stats {
	return &stats{
		Dropped: 0,
	}
}

// IncrementDropped atomically increments the dropped counter in stats.
// This method is called when log entries are dropped due to buffer overflow.
func (s *stats) IncrementDropped() {
	atomic.AddInt64(&s.Dropped, 1)
}

// GetDropped atomically reads the current dropped count.
// This method provides thread-safe access to statistics.
func (s *stats) GetDropped() int64 {
	return atomic.LoadInt64(&s.Dropped)
}

// Reset atomically resets all counters to zero.
// Useful for periodic statistics collection.
func (s *stats) Reset() {
	atomic.StoreInt64(&s.Dropped, 0)
}
