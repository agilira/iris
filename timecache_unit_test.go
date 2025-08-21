// timecache_unit_test.go: Unit tests for time cache functionality
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

// TestTimeCacheBasicFunctionality tests basic time cache operations
func TestTimeCacheBasicFunctionality(t *testing.T) {
	// Test CachedTimeNano
	nano1 := CachedTimeNano()
	if nano1 <= 0 {
		t.Error("CachedTimeNano should return positive value")
	}

	// Test CachedTime
	time1 := CachedTime()
	if time1.IsZero() {
		t.Error("CachedTime should not return zero time")
	}

	// Verify consistency
	nano2 := time1.UnixNano()
	// Allow small difference due to cache precision
	diff := nano1 - nano2
	if diff < 0 {
		diff = -diff
	}
	if diff > int64(2*time.Millisecond) {
		t.Errorf("Time inconsistency too large: %d ns", diff)
	}
}

// TestTimeCacheUpdate verifies cache updates over time
func TestTimeCacheUpdate(t *testing.T) {
	initial := CachedTimeNano()

	// Wait for cache update (cache updates every 1ms)
	time.Sleep(5 * time.Millisecond)

	updated := CachedTimeNano()
	if updated <= initial {
		t.Error("Cache should update over time")
	}
}

// TestTimeCacheConcurrency tests concurrent access to time cache
func TestTimeCacheConcurrency(t *testing.T) {
	const numGoroutines = 100
	const numReads = 1000

	var wg sync.WaitGroup
	var errors int64

	// Concurrent readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < numReads; j++ {
				nano := CachedTimeNano()
				if nano <= 0 {
					atomic.AddInt64(&errors, 1)
					return
				}

				time := CachedTime()
				if time.IsZero() {
					atomic.AddInt64(&errors, 1)
					return
				}
			}
		}()
	}

	wg.Wait()

	if errors > 0 {
		t.Errorf("Found %d errors in concurrent access", errors)
	}
}

// TestTimeCachePrecision tests cache precision
func TestTimeCachePrecision(t *testing.T) {
	readings := make([]int64, 5)

	// Take readings with longer intervals
	for i := 0; i < 5; i++ {
		readings[i] = CachedTimeNano()
		time.Sleep(2 * time.Millisecond) // Wait longer than cache update interval
	}

	// Verify progression (at least some should progress)
	progressCount := 0
	for i := 1; i < len(readings); i++ {
		if readings[i] > readings[i-1] {
			progressCount++
		}
	}

	if progressCount < 2 { // At least half should progress
		t.Errorf("Time should progress more frequently: only %d/%d progressed",
			progressCount, len(readings)-1)
	}
}
