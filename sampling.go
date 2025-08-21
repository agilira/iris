package iris

import (
	"sync/atomic"
	"time"
)

// SamplingDecision represents the decision made by the sampler
type SamplingDecision int

const (
	// LogSample indicates the log entry should be sampled (logged)
	LogSample SamplingDecision = iota
	// DropSample indicates the log entry should be dropped
	DropSample
)

// SamplingConfig controls the sampling behavior
type SamplingConfig struct {
	// Initial is the number of entries to log before sampling starts.
	// For example, if Initial is 100, the first 100 entries will always be logged.
	Initial int64

	// Thereafter is the sampling rate after the initial period.
	// For example, if Thereafter is 100, then after the initial period,
	// every 100th entry will be logged.
	Thereafter int64

	// Tick is the duration after which the sampling counters reset.
	// This creates time-based sampling windows.
	// If Tick is 0, counters never reset (count-based sampling only).
	Tick time.Duration

	// Hook is an optional function called when entries are sampled or dropped.
	// It receives the sampling decision and can be used for metrics.
	Hook func(entry LogEntry, decision SamplingDecision)
}

// Sampler implements the sampling logic
type Sampler struct {
	config         SamplingConfig
	counter        int64
	lastTick       int64
	dropped        int64
	sampled        int64
	hasHook        bool  // Pre-computed hook presence for performance
	isPowerOf2     bool  // True if Thereafter is power of 2
	thereafterMask int64 // Bitmask for power-of-2 optimization
}

// NewSampler creates a new sampler with the given configuration
func NewSampler(config SamplingConfig) *Sampler {
	if config.Initial <= 0 {
		config.Initial = 100
	}
	if config.Thereafter <= 0 {
		config.Thereafter = 100
	}

	sampler := &Sampler{
		config:   config,
		lastTick: CachedTimeNano(),   // Use cached time for initialization
		hasHook:  config.Hook != nil, // Pre-compute hook presence
	}

	// Optimize for power-of-2 Thereafter values
	if config.Thereafter > 0 && (config.Thereafter&(config.Thereafter-1)) == 0 {
		sampler.isPowerOf2 = true
		sampler.thereafterMask = config.Thereafter - 1
	}

	return sampler
}

// Sample determines whether a log entry should be sampled or dropped
func (s *Sampler) Sample(entry LogEntry) SamplingDecision {
	// Get current counter value and increment atomically
	n := atomic.AddInt64(&s.counter, 1)

	// Check if we need to reset based on tick duration
	if s.config.Tick > 0 {
		now := CachedTimeNano() // Use cached time instead of syscall
		lastTick := atomic.LoadInt64(&s.lastTick)

		if now-lastTick > int64(s.config.Tick) {
			// Try to reset - only one goroutine should succeed
			if atomic.CompareAndSwapInt64(&s.lastTick, lastTick, now) {
				atomic.StoreInt64(&s.counter, 1)
				n = 1
			}
		}
	}

	var decision SamplingDecision

	// Apply sampling logic
	if n <= s.config.Initial {
		// Within initial window - always sample
		decision = LogSample
		atomic.AddInt64(&s.sampled, 1)
	} else {
		// After initial window - sample every Nth entry
		// We want to sample the Nth entry after the initial period
		afterInitial := n - s.config.Initial

		var shouldSample bool
		if s.isPowerOf2 {
			// Fast bit-wise operation for power-of-2
			shouldSample = (afterInitial & s.thereafterMask) == 1
		} else {
			// Standard modulo for non-power-of-2
			shouldSample = afterInitial%s.config.Thereafter == 1
		}

		if shouldSample {
			decision = LogSample
			atomic.AddInt64(&s.sampled, 1)
		} else {
			decision = DropSample
			atomic.AddInt64(&s.dropped, 1)
		}
	}

	// Call hook if configured (optimized check)
	if s.hasHook {
		s.config.Hook(entry, decision)
	}

	return decision
}

// Stats returns the current sampling statistics
func (s *Sampler) Stats() SamplingStats {
	return SamplingStats{
		Sampled: atomic.LoadInt64(&s.sampled),
		Dropped: atomic.LoadInt64(&s.dropped),
		Total:   atomic.LoadInt64(&s.counter),
	}
}

// SamplingStats contains sampling statistics
type SamplingStats struct {
	Sampled int64 // Number of entries that were sampled (logged)
	Dropped int64 // Number of entries that were dropped
	Total   int64 // Total number of entries processed
}

// SamplingRate returns the current sampling rate as a percentage
func (s SamplingStats) SamplingRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Sampled) / float64(s.Total) * 100
}

// DropRate returns the current drop rate as a percentage
func (s SamplingStats) DropRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Dropped) / float64(s.Total) * 100
}

// Common sampling configurations

// NewDevelopmentSampling creates a sampling config suitable for development:
// - Initial: 100 entries always logged
// - Thereafter: every 10th entry logged
// - No time-based reset
func NewDevelopmentSampling() SamplingConfig {
	return SamplingConfig{
		Initial:    100,
		Thereafter: 10,
		Tick:       0,
	}
}

// NewProductionSampling creates a sampling config suitable for production:
// - Initial: 100 entries always logged
// - Thereafter: every 100th entry logged
// - Reset every minute
func NewProductionSampling() SamplingConfig {
	return SamplingConfig{
		Initial:    100,
		Thereafter: 100,
		Tick:       time.Minute,
	}
}

// NewHighVolumeSampling creates a sampling config for very high volume scenarios:
// - Initial: 50 entries always logged
// - Thereafter: every 1000th entry logged
// - Reset every 10 minutes
func NewHighVolumeSampling() SamplingConfig {
	return SamplingConfig{
		Initial:    50,
		Thereafter: 1000,
		Tick:       10 * time.Minute,
	}
}
