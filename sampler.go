// sampler.go: High-performance log sampling for rate limiting
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
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
