// sampler.go: High-performance log sampling for rate limiting
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync/atomic"
	"time"

	"github.com/agilira/go-timecache"
)

// Sampler defines the interface for log sampling strategies.
// Implementations control which log entries are allowed through
// to prevent overwhelming downstream systems.
type Sampler interface {
	// Allow determines if a log entry at the given level should be processed.
	// Returns true if the entry should be logged, false if it should be dropped.
	Allow(level Level) bool
}

// TokenBucketSampler implements rate limiting using a token bucket algorithm.
// Provides burst capacity with sustained rate limiting for high-volume logging.
type TokenBucketSampler struct {
	capacity int64         // Maximum tokens in bucket
	refill   int64         // Tokens added per refill period
	every    time.Duration // Refill period duration

	tokens atomic.Int64 // Current token count (atomic for concurrency)
	last   atomic.Int64 // Last refill timestamp in nanoseconds
}

// NewTokenBucketSampler creates a new token bucket sampler with the specified parameters.
// Validates inputs and sets reasonable defaults for invalid values.
//
// Parameters:
//   - capacity: Maximum number of tokens (burst capacity)
//   - refill: Number of tokens added per refill period
//   - every: Time duration between refills
//
// Returns a configured sampler ready for concurrent use.
func NewTokenBucketSampler(capacity, refill int64, every time.Duration) *TokenBucketSampler {
	// Validate and set defaults for parameters
	if capacity <= 0 {
		capacity = 1
	}
	if refill <= 0 {
		refill = 1
	}
	if every <= 0 {
		every = time.Millisecond
	}

	s := &TokenBucketSampler{
		capacity: capacity,
		refill:   refill,
		every:    every,
	}

	// Initialize with full capacity and current time
	s.tokens.Store(capacity)
	s.last.Store(timecache.CachedTimeNano())
	return s
}

// SamplerLimit defines the rate limit parameters for a single log level.
// Each level can have independent burst capacity and sustained rate.
type SamplerLimit struct {
	Capacity int64         // Maximum tokens (burst capacity)
	Refill   int64         // Tokens added per refill period
	Every    time.Duration // Refill period duration
}

// LevelAwareSampler applies independent rate limits per log level.
// Levels without a configured limit pass unconditionally, which
// ensures Error/Fatal are never silenced even under flood attacks.
type LevelAwareSampler struct {
	limits map[Level]*TokenBucketSampler
}

// NewLevelAwareSampler creates a sampler with per-level rate limits.
// Levels not present in the map are never sampled (always allowed).
// WHY: This is the firewall against log flood attacks (CWE-799).
// An attacker flooding Debug cannot drown Error/Fatal signals because
// each level draws from an independent token bucket.
func NewLevelAwareSampler(limits map[Level]SamplerLimit) *LevelAwareSampler {
	buckets := make(map[Level]*TokenBucketSampler, len(limits))
	for lvl, lim := range limits {
		buckets[lvl] = NewTokenBucketSampler(lim.Capacity, lim.Refill, lim.Every)
	}
	return &LevelAwareSampler{limits: buckets}
}

// DefaultLevelAwareSampler returns a production-ready configuration:
//
//	Error/Fatal: unlimited (no entry in map — always passes)
//	Warn:        100/s burst, 100/s sustained
//	Info:        50/s burst, 50/s sustained
//	Debug:       10/s burst, 10/s sustained
//
// WHY: these defaults are tuned for a daemon that logs to disk.
// High-severity levels must never be silenced; low-severity levels
// are throttled to prevent disk saturation under load.
func DefaultLevelAwareSampler() *LevelAwareSampler {
	return NewLevelAwareSampler(map[Level]SamplerLimit{
		Warn:  {Capacity: 100, Refill: 100, Every: time.Second},
		Info:  {Capacity: 50, Refill: 50, Every: time.Second},
		Debug: {Capacity: 10, Refill: 10, Every: time.Second},
	})
}

// Allow implements Sampler. If no limit is configured for the given level,
// the entry passes unconditionally. Otherwise it delegates to the
// per-level TokenBucketSampler.
func (s *LevelAwareSampler) Allow(level Level) bool {
	bucket, ok := s.limits[level]
	if !ok {
		return true // unconfigured levels always pass
	}
	return bucket.Allow(level)
}

// Allow implements the Sampler interface using token bucket rate limiting.
// Thread-safe implementation that refills tokens based on elapsed time
// and consumes tokens for allowed log entries.
//
// Parameters:
//   - level: Log level (unused in this implementation, all levels treated equally)
//
// Returns true if logging should proceed, false if rate limited.
func (s *TokenBucketSampler) Allow(_ Level) bool {
	now := timecache.CachedTimeNano()
	last := s.last.Load()

	// Calculate tokens to add based on elapsed time
	elapsed := now - last
	tokensToAdd := elapsed / s.every.Nanoseconds() * s.refill

	if tokensToAdd > 0 {
		// Update last time atomically
		if s.last.CompareAndSwap(last, now) {
			// Add tokens up to capacity
			current := s.tokens.Load()
			newTokens := current + tokensToAdd
			if newTokens > s.capacity {
				newTokens = s.capacity
			}
			s.tokens.Store(newTokens)
		}
	}

	// Try to consume a token
	for {
		current := s.tokens.Load()
		if current <= 0 {
			return false
		}
		if s.tokens.CompareAndSwap(current, current-1) {
			return true
		}
	}
}
