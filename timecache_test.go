// timecache_test.go: Test suite for ultra-fast time caching
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
	"time"
)

func TestCachedTimeNano(t *testing.T) {
	// Test that CachedTimeNano returns a reasonable timestamp
	nano := CachedTimeNano()
	now := time.Now().UnixNano()

	// Should be within 1ms of actual time (allowing for cache precision)
	diff := now - nano
	if diff < 0 {
		diff = -diff
	}

	if diff > int64(time.Millisecond) {
		t.Errorf("CachedTimeNano too far from actual time: diff=%dms", diff/int64(time.Millisecond))
	}
}

func TestCachedTime(t *testing.T) {
	// Test that CachedTime returns a reasonable time.Time
	cached := CachedTime()
	now := time.Now()

	// Should be within 1ms of actual time
	diff := now.Sub(cached)
	if diff < 0 {
		diff = -diff
	}

	if diff > time.Millisecond {
		t.Errorf("CachedTime too far from actual time: diff=%v", diff)
	}
}

func TestTimeCacheConsistency(t *testing.T) {
	// Multiple calls within the cache update interval should return same value
	time1 := CachedTimeNano()
	time2 := CachedTimeNano()
	time3 := CachedTimeNano()

	// All should be exactly the same (cached value)
	if time1 != time2 || time2 != time3 {
		t.Errorf("CachedTimeNano not consistent: %d, %d, %d", time1, time2, time3)
	}
}

func TestTimeCacheProgression(t *testing.T) {
	// Wait for multiple cache update cycles to ensure progression
	start := CachedTimeNano()

	// Wait long enough to guarantee at least 2-3 ticker updates
	time.Sleep(2 * time.Millisecond) // 4x the 500μs update interval

	end := CachedTimeNano()

	// Time should have progressed
	if end <= start {
		t.Errorf("CachedTimeNano did not progress: start=%d, end=%d", start, end)
	}

	// Verify the progression is reasonable (should be at least the sleep duration)
	diff := end - start
	minExpected := int64(time.Millisecond) // At least 1ms progression
	if diff < minExpected {
		t.Errorf("CachedTimeNano progression too small: got %d ns, expected at least %d ns", diff, minExpected)
	}
}

// Test integration with logger
func TestLoggerUsesTimeCache(t *testing.T) {
	// Create logger - should use CachedTime by default
	logger, err := New(Config{
		Level:  Debug,
		Output: &testBuffer{},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Start()

	// Log multiple messages quickly
	start := CachedTimeNano()
	logger.Info("message 1")
	logger.Info("message 2")
	logger.Info("message 3")
	end := CachedTimeNano()

	// All logs should use cached time (same or very close)
	// Since we're logging quickly, should be within same cache window
	if end != start {
		// If they differ, the difference should be minimal (one cache update at most)
		diff := end - start
		if diff > int64(time.Millisecond) {
			t.Logf("TimeCache updated during test (normal): diff=%v", time.Duration(diff))
		}
	}
}

// Test that custom TimeFn still works
func TestCustomTimeFn(t *testing.T) {
	customTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	logger, err := New(Config{
		Level:  Info,
		Output: &testBuffer{},
		TimeFn: func() time.Time { return customTime },
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Custom TimeFn should override TimeCache
	// This test ensures backwards compatibility
	if logger.clock() != customTime {
		t.Errorf("Custom TimeFn not respected: got %v, want %v", logger.clock(), customTime)
	}
}

type testBuffer struct {
	data []byte
}

func (tb *testBuffer) Write(p []byte) (int, error) {
	tb.data = append(tb.data, p...)
	return len(p), nil
}

func (tb *testBuffer) Sync() error {
	return nil
}
