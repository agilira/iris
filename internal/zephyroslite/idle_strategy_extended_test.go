// idle_strategy_extended_test.go: Comprehensive tests for idle strategies
//
// This file provides complete test coverage for all idle strategies
// in the zephyroslite package to improve overall coverage.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"os"
	"testing"
	"time"
)

func TestSpinningIdleStrategy(t *testing.T) {
	strategy := NewSpinningIdleStrategy()

	// Test Idle method
	result := strategy.Idle()
	if !result {
		t.Errorf("Expected Idle() = true, got false")
	}

	// Test Reset method explicitly - ensure it's called and doesn't panic
	strategy.Reset()

	// Test that Reset can be called multiple times safely
	strategy.Reset()
	strategy.Reset()

	// Test String method
	if strategy.String() != "spinning" {
		t.Errorf("Expected String() = 'spinning', got %s", strategy.String())
	}

	// Test that Reset doesn't affect functionality
	result2 := strategy.Idle()
	if !result2 {
		t.Errorf("Expected Idle() = true after Reset, got false")
	}
}

func TestSleepingIdleStrategy_ProgressiveBehavior(t *testing.T) {
	// Skip timing-sensitive tests in CI environments where scheduler is unpredictable
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping timing-sensitive idle strategy test in CI environment")
	}

	strategy := NewSleepingIdleStrategy(time.Millisecond, 2)

	// Test initial state
	if strategy.String() != "sleeping" {
		t.Errorf("Expected String() = 'sleeping', got %s", strategy.String())
	}

	// Test Idle method (should spin first)
	start := time.Now()
	result := strategy.Idle()
	elapsed := time.Since(start)

	// First call should not sleep (spinning)
	if !result {
		t.Errorf("Expected first Idle() = true, got false")
	}
	if elapsed > time.Millisecond/2 {
		t.Errorf("First Idle() should not sleep, but took %v", elapsed)
	}

	// Second call should still spin
	start = time.Now()
	result = strategy.Idle()
	elapsed = time.Since(start)

	if !result {
		t.Errorf("Expected second Idle() = true, got false")
	}
	if elapsed > time.Millisecond/2 {
		t.Errorf("Second Idle() should still spin, but took %v", elapsed)
	}

	// Third call should sleep
	start = time.Now()
	result = strategy.Idle()
	elapsed = time.Since(start)

	if !result {
		t.Errorf("Expected third Idle() = true, got false")
	}
	if elapsed < time.Millisecond/2 {
		t.Errorf("Third Idle() should sleep for ~1ms, but took only %v", elapsed)
	}

	// Test Reset
	strategy.Reset()

	// After reset, should not sleep again
	start = time.Now()
	result = strategy.Idle()
	elapsed = time.Since(start)

	if !result {
		t.Errorf("Expected Idle() after reset = true, got false")
	}
	if elapsed > time.Millisecond/2 {
		t.Errorf("Idle() after reset should not sleep, but took %v", elapsed)
	}
}

func TestYieldingIdleStrategy(t *testing.T) {
	strategy := NewYieldingIdleStrategy(2)

	// Test String method
	if strategy.String() != "yielding" {
		t.Errorf("Expected String() = 'yielding', got %s", strategy.String())
	}

	// Test Idle method (first calls should just spin)
	result := strategy.Idle()
	if !result {
		t.Errorf("Expected Idle() = true, got false")
	}

	result = strategy.Idle()
	if !result {
		t.Errorf("Expected second Idle() = true, got false")
	}

	// Third call should yield
	result = strategy.Idle()
	if !result {
		t.Errorf("Expected third Idle() (with yield) = true, got false")
	}

	// Test Reset method
	strategy.Reset()
}

func TestChannelIdleStrategy(t *testing.T) {
	strategy := NewChannelIdleStrategy(10 * time.Millisecond)

	// Test String method
	if strategy.String() != "channel" {
		t.Errorf("Expected String() = 'channel', got %s", strategy.String())
	}

	// Test WakeUp method first
	strategy.WakeUp()

	// Test Idle method - should not block if signal is present
	start := time.Now()
	result := strategy.Idle()
	elapsed := time.Since(start)

	if !result {
		t.Errorf("Expected Idle() = true, got false")
	}
	if elapsed > 5*time.Millisecond {
		t.Errorf("Idle() should not block when signal present, but took %v", elapsed)
	}

	// Test Reset method
	strategy.Reset()

	// Test Idle again after reset signal
	start = time.Now()
	result = strategy.Idle()
	elapsed = time.Since(start)

	if !result {
		t.Errorf("Expected Idle() after Reset = true, got false")
	}
	if elapsed > 5*time.Millisecond {
		t.Errorf("Idle() after Reset should not block, but took %v", elapsed)
	}
}

func TestChannelIdleStrategy_Timeout(t *testing.T) {
	// Skip timing-sensitive tests in CI environments where scheduler is unpredictable
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping timing-sensitive timeout test in CI environment")
	}

	strategy := NewChannelIdleStrategy(5 * time.Millisecond)

	// Test timeout behavior
	start := time.Now()
	result := strategy.Idle()
	elapsed := time.Since(start)

	if !result {
		t.Errorf("Expected Idle() with timeout = true, got false")
	}
	// More lenient timing for test stability
	if elapsed < 2*time.Millisecond || elapsed > 20*time.Millisecond {
		t.Errorf("Expected timeout around 5ms, but took %v", elapsed)
	}
}

func TestProgressiveIdleStrategy(t *testing.T) {
	strategy := NewProgressiveIdleStrategy()

	// Test String method
	if strategy.String() != "progressive" {
		t.Errorf("Expected String() = 'progressive', got %s", strategy.String())
	}

	// Test hot spin phase
	for i := 0; i < 100; i++ {
		result := strategy.Idle()
		if !result {
			t.Errorf("Expected Idle() in hot spin = true, got false")
		}
	}

	// Test Reset
	strategy.Reset()

	// After reset, should start from hot spin again
	result := strategy.Idle()
	if !result {
		t.Errorf("Expected Idle() after reset = true, got false")
	}
}

func TestIdleStrategyTypes(t *testing.T) {
	strategies := []struct {
		name     string
		strategy IdleStrategy
	}{
		{"Spinning", NewSpinningIdleStrategy()},
		{"Sleeping", NewSleepingIdleStrategy(time.Millisecond, 1)},
		{"Yielding", NewYieldingIdleStrategy(100)},
		{"Channel", NewChannelIdleStrategy(time.Millisecond)},
		{"Progressive", NewProgressiveIdleStrategy()},
	}

	for _, test := range strategies {
		t.Run(test.name, func(t *testing.T) {
			// Test that all strategies implement the interface correctly
			result := test.strategy.Idle()
			if !result {
				t.Errorf("%s: Expected Idle() = true, got false", test.name)
			}

			test.strategy.Reset()

			stringResult := test.strategy.String()
			if stringResult == "" {
				t.Errorf("%s: String() should not be empty", test.name)
			}
		})
	}
}

func TestChannelIdleStrategy_ConcurrentWakeUp(t *testing.T) {
	strategy := NewChannelIdleStrategy(time.Second) // Long timeout for this test

	// Test multiple concurrent WakeUps don't block
	for i := 0; i < 10; i++ {
		go strategy.WakeUp()
	}

	// Give goroutines time to execute
	time.Sleep(10 * time.Millisecond)

	// Strategy should still work normally
	start := time.Now()
	result := strategy.Idle()
	elapsed := time.Since(start)

	if !result {
		t.Errorf("Expected Idle() = true after concurrent WakeUps, got false")
	}
	if elapsed > 10*time.Millisecond {
		t.Errorf("Idle() should not block after WakeUps, but took %v", elapsed)
	}
}

func TestSleepingIdleStrategy_EdgeCases(t *testing.T) {
	// Test with zero sleep duration (should use default)
	strategy := NewSleepingIdleStrategy(0, 0)
	if strategy.String() != "sleeping" {
		t.Errorf("Expected String() = 'sleeping', got %s", strategy.String())
	}

	result := strategy.Idle()
	if !result {
		t.Errorf("Expected Idle() = true, got false")
	}

	// Test with negative maxSpins (should use 0)
	strategy2 := NewSleepingIdleStrategy(time.Microsecond, -5)
	result = strategy2.Idle()
	if !result {
		t.Errorf("Expected Idle() with negative maxSpins = true, got false")
	}
}

func TestYieldingIdleStrategy_EdgeCases(t *testing.T) {
	// Test with zero or negative maxSpins (should use default)
	strategy := NewYieldingIdleStrategy(0)
	if strategy.String() != "yielding" {
		t.Errorf("Expected String() = 'yielding', got %s", strategy.String())
	}

	strategy2 := NewYieldingIdleStrategy(-10)
	result := strategy2.Idle()
	if !result {
		t.Errorf("Expected Idle() with negative maxSpins = true, got false")
	}
}

func BenchmarkSpinningIdleStrategy(b *testing.B) {
	strategy := NewSpinningIdleStrategy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.Idle()
	}
}

func BenchmarkSleepingIdleStrategy(b *testing.B) {
	strategy := NewSleepingIdleStrategy(time.Nanosecond, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.Idle()
		strategy.Reset() // Always reset to avoid sleeping
	}
}

func BenchmarkYieldingIdleStrategy(b *testing.B) {
	strategy := NewYieldingIdleStrategy(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.Idle()
	}
}

func BenchmarkChannelIdleStrategy(b *testing.B) {
	strategy := NewChannelIdleStrategy(0) // No timeout

	// Pre-fill with signals to avoid blocking
	for i := 0; i < b.N; i++ {
		strategy.WakeUp()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.Idle()
	}
}

func BenchmarkProgressiveIdleStrategy(b *testing.B) {
	strategy := NewProgressiveIdleStrategy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.Idle()
		if i%1000 == 0 {
			strategy.Reset() // Reset occasionally to stay in hot spin
		}
	}
}
