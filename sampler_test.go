// sampler_test.go: Test suite for iris log sampling functionality
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync"
	"testing"
	"time"
)

// TestNewTokenBucketSampler tests sampler creation with various parameters
func TestNewTokenBucketSampler(t *testing.T) {
	tests := []struct {
		name             string
		capacity         int64
		refill           int64
		every            time.Duration
		expectedCapacity int64
		expectedRefill   int64
		expectedEvery    time.Duration
	}{
		{
			name:             "valid parameters",
			capacity:         100,
			refill:           10,
			every:            time.Second,
			expectedCapacity: 100,
			expectedRefill:   10,
			expectedEvery:    time.Second,
		},
		{
			name:             "zero capacity gets default",
			capacity:         0,
			refill:           10,
			every:            time.Second,
			expectedCapacity: 1,
			expectedRefill:   10,
			expectedEvery:    time.Second,
		},
		{
			name:             "negative capacity gets default",
			capacity:         -5,
			refill:           10,
			every:            time.Second,
			expectedCapacity: 1,
			expectedRefill:   10,
			expectedEvery:    time.Second,
		},
		{
			name:             "zero refill gets default",
			capacity:         100,
			refill:           0,
			every:            time.Second,
			expectedCapacity: 100,
			expectedRefill:   1,
			expectedEvery:    time.Second,
		},
		{
			name:             "negative refill gets default",
			capacity:         100,
			refill:           -3,
			every:            time.Second,
			expectedCapacity: 100,
			expectedRefill:   1,
			expectedEvery:    time.Second,
		},
		{
			name:             "zero duration gets default",
			capacity:         100,
			refill:           10,
			every:            0,
			expectedCapacity: 100,
			expectedRefill:   10,
			expectedEvery:    time.Millisecond,
		},
		{
			name:             "negative duration gets default",
			capacity:         100,
			refill:           10,
			every:            -time.Second,
			expectedCapacity: 100,
			expectedRefill:   10,
			expectedEvery:    time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampler := NewTokenBucketSampler(tt.capacity, tt.refill, tt.every)

			if sampler.capacity != tt.expectedCapacity {
				t.Errorf("Expected capacity %d, got %d", tt.expectedCapacity, sampler.capacity)
			}

			if sampler.refill != tt.expectedRefill {
				t.Errorf("Expected refill %d, got %d", tt.expectedRefill, sampler.refill)
			}

			if sampler.every != tt.expectedEvery {
				t.Errorf("Expected every %v, got %v", tt.expectedEvery, sampler.every)
			}

			// Should start with full capacity
			if sampler.tokens.Load() != tt.expectedCapacity {
				t.Errorf("Expected initial tokens %d, got %d", tt.expectedCapacity, sampler.tokens.Load())
			}
		})
	}
}

// TestTokenBucketSamplerBasicAllow tests basic allow functionality
func TestTokenBucketSamplerBasicAllow(t *testing.T) {
	sampler := NewTokenBucketSampler(5, 1, time.Second)

	// Should allow first 5 requests (consume all tokens)
	for i := 0; i < 5; i++ {
		if !sampler.Allow(Info) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 6th request should be denied (no tokens left)
	if sampler.Allow(Info) {
		t.Error("Request 6 should be denied")
	}

	// Verify tokens are depleted
	if sampler.tokens.Load() != 0 {
		t.Errorf("Expected 0 tokens remaining, got %d", sampler.tokens.Load())
	}
}

// TestTokenBucketSamplerRefill tests token refill functionality
func TestTokenBucketSamplerRefill(t *testing.T) {
	sampler := NewTokenBucketSampler(10, 5, 100*time.Millisecond)

	// Consume all tokens
	for i := 0; i < 10; i++ {
		if !sampler.Allow(Info) {
			t.Errorf("Initial request %d should be allowed", i+1)
		}
	}

	// Should be denied now
	if sampler.Allow(Info) {
		t.Error("Should be denied after consuming all tokens")
	}

	// Wait for refill period
	time.Sleep(150 * time.Millisecond)

	// Should allow 5 more requests (refill amount)
	allowedAfterRefill := 0
	for i := 0; i < 10; i++ {
		if sampler.Allow(Info) {
			allowedAfterRefill++
		}
	}

	if allowedAfterRefill < 5 {
		t.Errorf("Expected at least 5 requests allowed after refill, got %d", allowedAfterRefill)
	}
}

// TestTokenBucketSamplerCapacityLimit tests that tokens don't exceed capacity
func TestTokenBucketSamplerCapacityLimit(t *testing.T) {
	sampler := NewTokenBucketSampler(3, 10, 50*time.Millisecond)

	// Start with 3 tokens, consume them all
	for i := 0; i < 3; i++ {
		sampler.Allow(Info)
	}

	// Wait for multiple refill periods (should add more than capacity)
	time.Sleep(200 * time.Millisecond)

	// Should only be able to consume up to capacity (3), not more
	allowedRequests := 0
	for i := 0; i < 10; i++ {
		if sampler.Allow(Info) {
			allowedRequests++
		}
	}

	if allowedRequests > 3 {
		t.Errorf("Allowed %d requests, should not exceed capacity of 3", allowedRequests)
	}
}

// TestTokenBucketSamplerLevelIgnored tests that log level is ignored
func TestTokenBucketSamplerLevelIgnored(t *testing.T) {
	sampler := NewTokenBucketSampler(4, 1, time.Second)

	levels := []Level{Debug, Info, Warn, Error}

	// All levels should be treated equally
	for i, level := range levels {
		if !sampler.Allow(level) {
			t.Errorf("Request %d with level %v should be allowed", i+1, level)
		}
	}

	// 5th request should be denied regardless of level
	if sampler.Allow(Debug) {
		t.Error("5th request should be denied regardless of level")
	}
}

// TestTokenBucketSamplerConcurrency tests concurrent access
func TestTokenBucketSamplerConcurrency(t *testing.T) {
	sampler := NewTokenBucketSampler(1000, 100, 10*time.Millisecond)

	var wg sync.WaitGroup
	var allowed, denied int64
	var mu sync.Mutex

	numGoroutines := 50
	requestsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localAllowed, localDenied := int64(0), int64(0)

			for j := 0; j < requestsPerGoroutine; j++ {
				if sampler.Allow(Info) {
					localAllowed++
				} else {
					localDenied++
				}

				// Small delay to allow some refill
				if j%20 == 0 {
					time.Sleep(time.Millisecond)
				}
			}

			mu.Lock()
			allowed += localAllowed
			denied += localDenied
			mu.Unlock()
		}()
	}

	wg.Wait()

	totalRequests := int64(numGoroutines * requestsPerGoroutine)
	if allowed+denied != totalRequests {
		t.Errorf("Total requests mismatch: allowed(%d) + denied(%d) != total(%d)", allowed, denied, totalRequests)
	}

	// Should have some allowed and some denied
	if allowed == 0 {
		t.Error("Expected some requests to be allowed")
	}

	if denied == 0 {
		t.Error("Expected some requests to be denied")
	}

	t.Logf("Concurrent test: %d allowed, %d denied out of %d total", allowed, denied, totalRequests)
}

// TestTokenBucketSamplerZeroCapacity tests edge case with minimum capacity
func TestTokenBucketSamplerZeroCapacity(t *testing.T) {
	sampler := NewTokenBucketSampler(0, 0, 0) // All should get defaults

	// Should allow exactly 1 request (default capacity)
	if !sampler.Allow(Info) {
		t.Error("Should allow first request with default capacity")
	}

	// Second request should be denied
	if sampler.Allow(Info) {
		t.Error("Should deny second request with capacity 1")
	}
}

// TestTokenBucketSamplerInterface tests that it implements Sampler interface
func TestTokenBucketSamplerInterface(t *testing.T) {
	var sampler Sampler = NewTokenBucketSampler(10, 1, time.Second)

	// Should be able to call Allow through interface
	if !sampler.Allow(Info) {
		t.Error("Should allow request through Sampler interface")
	}
}

// TestTokenBucketSamplerRapidRefill tests behavior with very fast refill
func TestTokenBucketSamplerRapidRefill(t *testing.T) {
	if IsCIEnvironment() {
		t.Skip("Skipping rapid refill test in CI due to timing sensitivity")
	}

	sampler := NewTokenBucketSampler(5, 3, time.Microsecond)

	// Consume initial tokens
	for i := 0; i < 5; i++ {
		sampler.Allow(Info)
	}

	// Wait a small amount for rapid refill
	time.Sleep(time.Millisecond)

	// Should have refilled significantly
	allowedAfterRefill := 0
	for i := 0; i < 20; i++ {
		if sampler.Allow(Info) {
			allowedAfterRefill++
		}
	}

	if allowedAfterRefill == 0 {
		t.Error("Expected some requests to be allowed after rapid refill")
	}

	// With rapid refill timing can be imprecise, just ensure it's reasonable
	if allowedAfterRefill > 15 {
		t.Errorf("Allowed %d requests, seems too many for rapid refill test", allowedAfterRefill)
	}
}

// TestTokenBucketSamplerLongWait tests behavior after long idle period
func TestTokenBucketSamplerLongWait(t *testing.T) {
	sampler := NewTokenBucketSampler(3, 2, 10*time.Millisecond)

	// Consume all tokens
	for i := 0; i < 3; i++ {
		sampler.Allow(Info)
	}

	// Wait much longer than refill period
	time.Sleep(100 * time.Millisecond)

	// Should be refilled to capacity (not more)
	allowed := 0
	for i := 0; i < 10; i++ {
		if sampler.Allow(Info) {
			allowed++
		}
	}

	if allowed != 3 {
		t.Errorf("Expected exactly 3 requests allowed after long wait, got %d", allowed)
	}
}

// TestTokenBucketSamplerBurstThenSustained tests burst followed by sustained rate
func TestTokenBucketSamplerBurstThenSustained(t *testing.T) {
	capacity := int64(10)
	refillRate := int64(2)
	refillPeriod := 50 * time.Millisecond

	sampler := NewTokenBucketSampler(capacity, refillRate, refillPeriod)

	// Phase 1: Burst (should allow up to capacity)
	burstAllowed := 0
	for i := 0; i < 20; i++ {
		if sampler.Allow(Info) {
			burstAllowed++
		}
	}

	if burstAllowed != int(capacity) {
		t.Errorf("Burst phase: expected %d allowed, got %d", capacity, burstAllowed)
	}

	// Phase 2: Sustained rate (should allow refill rate over time)
	time.Sleep(refillPeriod + 10*time.Millisecond)

	sustainedAllowed := 0
	for i := 0; i < 10; i++ {
		if sampler.Allow(Info) {
			sustainedAllowed++
		}
	}

	if sustainedAllowed < 1 || sustainedAllowed > int(refillRate) {
		t.Errorf("Sustained phase: expected 1-%d allowed, got %d", refillRate, sustainedAllowed)
	}
}
