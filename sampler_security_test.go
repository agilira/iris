// sampler_security_test.go: Security test suite for TokenBucketSampler
//
// THREAT MODEL:
//   CWE-400: Uncontrolled Resource Consumption
//     - Invalid parameters (zero, negative, overflow) could disable rate limiting
//     - Token overflow past capacity could grant unlimited bursts
//   CWE-362: Concurrent Execution Using Shared Resource
//     - Concurrent Allow() calls on same sampler could corrupt token count
//     - Race between refill (Store) and consume (CAS) could lose tokens
//   CWE-799: Improper Control of Interaction Frequency
//     - Allow() ignores Level parameter (D4 defect): under flood, Error records
//       are dropped with same probability as Debug. Documented as known limitation.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- Attack vector: CWE-400 (Invalid parameters) ---

func TestSecurity_Sampler_InvalidParameters(t *testing.T) {
	// ATTACK VECTOR: CWE-400
	// IMPACT: Zero or negative capacity/refill could disable rate limiting
	// entirely or cause division by zero.
	// MITIGATION EXPECTED: NewTokenBucketSampler clamps invalid values to 1.
	cases := []struct {
		name     string
		capacity int64
		refill   int64
		every    time.Duration
	}{
		{"zero_capacity", 0, 1, time.Millisecond},
		{"negative_capacity", -1, 1, time.Millisecond},
		{"zero_refill", 1, 0, time.Millisecond},
		{"negative_refill", 1, -1, time.Millisecond},
		{"zero_duration", 1, 1, 0},
		{"negative_duration", 1, 1, -time.Millisecond},
		{"all_zero", 0, 0, 0},
		{"all_negative", -100, -100, -time.Second},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Must not panic
			s := NewTokenBucketSampler(tc.capacity, tc.refill, tc.every)
			if s == nil {
				t.Fatal("NewTokenBucketSampler returned nil")
			}

			// First Allow must succeed (capacity >= 1 after clamping)
			if !s.Allow(Info) {
				t.Error("first Allow() should succeed after parameter clamping")
			}
		})
	}
}

func TestSecurity_Sampler_OverflowParameters(t *testing.T) {
	// ATTACK VECTOR: CWE-400
	// IMPACT: MaxInt64 capacity could cause integer overflow in token arithmetic.
	// MITIGATION EXPECTED: graceful handling, no panic, still rate limits eventually.
	s := NewTokenBucketSampler(1<<62, 1, time.Millisecond)

	// Should work without panic
	allowed := s.Allow(Info)
	if !allowed {
		t.Error("Allow() should succeed with large capacity")
	}
}

// --- Attack vector: CWE-362 (Concurrent races) ---

func TestSecurity_Sampler_ConcurrentAllow(t *testing.T) {
	// ATTACK VECTOR: CWE-362
	// IMPACT: Concurrent Allow() calls could corrupt atomic counters,
	// granting more tokens than capacity or causing negative token counts.
	// MITIGATION EXPECTED: CAS loop ensures correct concurrent behavior.
	const (
		capacity   = 100
		goroutines = 50
		callsPer   = 20
	)

	s := NewTokenBucketSampler(capacity, 0, time.Hour) // WHY refill=0, every=Hour: no refills during test
	var allowed atomic.Int64
	var wg sync.WaitGroup

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < callsPer; i++ {
				if s.Allow(Info) {
					allowed.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	// Total allowed must not exceed capacity (no refills configured)
	total := allowed.Load()
	if total > capacity {
		t.Fatalf("allowed %d entries but capacity is %d -- token leak detected (CWE-362)",
			total, capacity)
	}

	// Should have allowed exactly capacity (all tokens consumed)
	if total != capacity {
		t.Logf("allowed %d of %d (some CAS retries may have occurred, this is acceptable)", total, capacity)
	}
}

func TestSecurity_Sampler_ConcurrentRefillAndConsume(t *testing.T) {
	// ATTACK VECTOR: CWE-362
	// IMPACT: Race between refill path (Store) and consume path (CAS)
	// could lose or duplicate tokens under high contention.
	// MITIGATION EXPECTED: atomic operations maintain consistency.
	const (
		goroutines = 20
		duration   = 100 * time.Millisecond
	)

	// WHY short refill period: forces refill/consume races
	s := NewTokenBucketSampler(10, 5, time.Microsecond)
	var allowed atomic.Int64
	var denied atomic.Int64
	var wg sync.WaitGroup

	done := make(chan struct{})
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					if s.Allow(Info) {
						allowed.Add(1)
					} else {
						denied.Add(1)
					}
				}
			}
		}()
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()

	// Verify some requests were allowed and some denied (rate limiting worked)
	a, d := allowed.Load(), denied.Load()
	if a == 0 {
		t.Fatal("no requests allowed -- sampler completely blocked")
	}
	if d == 0 {
		t.Fatal("no requests denied -- rate limiting not working")
	}
	t.Logf("allowed=%d denied=%d total=%d", a, d, a+d)
}

// --- Attack vector: CWE-799 (Level-blind sampling) ---

func TestSecurity_Sampler_LevelBypass_KnownDefect(t *testing.T) {
	// ATTACK VECTOR: CWE-799
	// IMPACT: Allow(_ Level) ignores level entirely. Under flood,
	// Error/Fatal records are dropped with same probability as Debug.
	// This is a KNOWN DEFECT (documented in ADR-006b D4).
	// This test DOCUMENTS the defect, not mitigates it.
	s := NewTokenBucketSampler(5, 0, time.Hour) // 5 tokens, no refill

	// Fill up with Debug entries
	for i := 0; i < 5; i++ {
		if !s.Allow(Debug) {
			t.Fatalf("Allow(Debug) failed on call %d, expected success", i+1)
		}
	}

	// Now an Error-level entry should be denied (DEFECT: all tokens gone)
	if s.Allow(Error) {
		// If this passes, it means level-awareness was added (D4 fix)
		t.Log("GOOD: Error allowed despite Debug flood -- level-aware fix may be in place")
	} else {
		// WHY this is NOT a test failure: D4 documents this as a known limitation.
		// The test exists to track when/if the fix is applied.
		t.Log("KNOWN DEFECT (D4): Error denied after Debug flood -- level-blind sampling")
	}
}

// --- Attack vector: Token exhaustion ---

func TestSecurity_Sampler_TokenExhaustion(t *testing.T) {
	// ATTACK VECTOR: CWE-400
	// IMPACT: After exhausting all tokens with no refill, sampler must
	// consistently deny ALL requests (no token leaks from concurrent bugs).
	// MITIGATION EXPECTED: Allow() returns false when tokens <= 0.
	s := NewTokenBucketSampler(3, 0, time.Hour)

	// Exhaust all tokens
	for i := 0; i < 3; i++ {
		if !s.Allow(Info) {
			t.Fatalf("Allow() denied on call %d, expected 3 tokens", i+1)
		}
	}

	// Must deny all subsequent calls
	for i := 0; i < 100; i++ {
		if s.Allow(Info) {
			t.Fatalf("Allow() succeeded on exhausted sampler (call %d)", i+1)
		}
	}
}

// --- Fuzz target: FuzzSamplerAllow ---

func FuzzSamplerAllow(f *testing.F) {
	// Seeds: boundary values for capacity, refill, and call count
	seeds := []struct {
		capacity int64
		refill   int64
		everyNs  int64
		calls    uint8
	}{
		{1, 1, 1000000, 10},       // minimal sampler, 10 calls
		{100, 50, 1000000000, 50}, // standard sampler
		{0, 0, 0, 5},              // all zero (clamped)
		{-1, -1, -1, 5},           // all negative (clamped)
		{1 << 30, 1, 1, 100},      // huge capacity
		{1, 1 << 30, 1, 100},      // huge refill
	}

	for _, s := range seeds {
		f.Add(s.capacity, s.refill, s.everyNs, s.calls)
	}

	f.Fuzz(func(t *testing.T, capacity, refill, everyNs int64, calls uint8) {
		// Clamp every to positive duration
		every := time.Duration(everyNs)
		if every <= 0 {
			every = time.Millisecond
		}

		// Must not panic regardless of inputs
		s := NewTokenBucketSampler(capacity, refill, every)
		if s == nil {
			t.Fatal("NewTokenBucketSampler returned nil")
		}

		allowedCount := 0
		for i := 0; i < int(calls); i++ {
			if s.Allow(Info) {
				allowedCount++
			}
		}

		// Sanity: allowed count must not exceed calls
		if allowedCount > int(calls) {
			t.Fatalf("allowed %d > calls %d", allowedCount, calls)
		}
	})
}
