// idle_strategy_test.go: Tests for configurable idle strategies
//
// This file tests the various idle strategies to ensure they provide the expected
// behavior in terms of CPU usage and responsiveness.
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/agilira/iris/internal/zephyroslite"
)

func TestIdleStrategies_Basic(t *testing.T) {
	strategies := []struct {
		name     string
		strategy IdleStrategy
	}{
		{"Spinning", NewSpinningIdleStrategy()},
		{"Sleeping", NewSleepingIdleStrategy(time.Millisecond, 0)},
		{"Yielding", NewYieldingIdleStrategy(1000)},
		{"Channel", NewChannelIdleStrategy(10 * time.Millisecond)},
		{"Progressive", NewProgressiveIdleStrategy()},
	}

	for _, test := range strategies {
		t.Run(test.name, func(t *testing.T) {
			processed := int64(0)
			processor := func(r *Record) {
				atomic.AddInt64(&processed, 1)
			}

			ring, err := newRing(64, 16, SingleRing, 1, zephyroslite.DropOnFull, test.strategy, processor)
			if err != nil {
				t.Fatalf("Failed to create ring with %s strategy: %v", test.name, err)
			}
			defer ring.Close()

			// Start processing
			go ring.Loop()

			// Write a test record
			success := ring.Write(func(r *Record) {
				r.Level = Info
				r.Msg = "Test message"
			})

			if !success {
				t.Errorf("Failed to write record with %s strategy", test.name)
			}

			// Wait for processing
			deadline := time.Now().Add(100 * time.Millisecond)
			for time.Now().Before(deadline) {
				if atomic.LoadInt64(&processed) > 0 {
					break
				}
				time.Sleep(time.Millisecond)
			}

			if atomic.LoadInt64(&processed) == 0 {
				t.Errorf("Record not processed with %s strategy", test.name)
			}
		})
	}
}

func TestIdleStrategies_Predefined(t *testing.T) {
	strategies := []struct {
		name     string
		strategy IdleStrategy
	}{
		{"SpinningStrategy", SpinningStrategy},
		{"BalancedStrategy", BalancedStrategy},
		{"EfficientStrategy", EfficientStrategy},
		{"HybridStrategy", HybridStrategy},
	}

	for _, test := range strategies {
		t.Run(test.name, func(t *testing.T) {
			processed := int64(0)
			processor := func(r *Record) {
				atomic.AddInt64(&processed, 1)
			}

			ring, err := newRing(64, 16, SingleRing, 1, zephyroslite.DropOnFull, test.strategy, processor)
			if err != nil {
				t.Fatalf("Failed to create ring with %s: %v", test.name, err)
			}
			defer ring.Close()

			// Start processing
			go ring.Loop()

			// Write a test record
			ring.Write(func(r *Record) {
				r.Level = Info
				r.Msg = "Test message"
			})

			// Wait for processing
			time.Sleep(10 * time.Millisecond)

			if atomic.LoadInt64(&processed) == 0 {
				t.Errorf("Record not processed with %s", test.name)
			}
		})
	}
}

func TestProgressiveIdleStrategy_AdaptiveBehavior(t *testing.T) {
	strategy := NewProgressiveIdleStrategy()

	// Test that strategy provides different behavior when idle vs active
	// This is a basic test to ensure the strategy doesn't panic and behaves reasonably

	// Simulate idle behavior
	for i := 0; i < 5; i++ {
		if !strategy.Idle() {
			t.Error("Progressive strategy should continue when idle")
		}
	}

	// Reset (simulates work found)
	strategy.Reset()

	// Idle again should start fresh
	if !strategy.Idle() {
		t.Error("Progressive strategy should continue after reset")
	}
}

func TestSleepingIdleStrategy_Parameters(t *testing.T) {
	testCases := []struct {
		name          string
		sleepDuration time.Duration
		maxSpins      int
	}{
		{"NoSpinImmedateSleep", time.Millisecond, 0},
		{"ShortSpin", 500 * time.Microsecond, 100},
		{"LongSpin", 2 * time.Millisecond, 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := NewSleepingIdleStrategy(tc.sleepDuration, tc.maxSpins)

			// Test that strategy functions correctly
			start := time.Now()

			// Call idle a few times
			for i := 0; i < tc.maxSpins+2; i++ {
				if !strategy.Idle() {
					t.Error("Sleeping strategy should continue when idle")
				}
			}

			elapsed := time.Since(start)

			// With sleeping, we should see some delay
			if tc.maxSpins == 0 && elapsed < tc.sleepDuration/2 {
				t.Errorf("Expected some sleep delay with maxSpins=0, got %v", elapsed)
			}

			// Reset should work without error
			strategy.Reset()
		})
	}
}

func TestChannelIdleStrategy_WakeUp(t *testing.T) {
	strategy := NewChannelIdleStrategy(10 * time.Millisecond)

	// Test that wakeup works
	done := make(chan bool)

	go func() {
		// This should block until wakeup
		strategy.Idle()
		done <- true
	}()

	// Give goroutine time to start
	time.Sleep(time.Millisecond)

	// Wake up the strategy
	if channelStrategy, ok := strategy.(*zephyroslite.ChannelIdleStrategy); ok {
		channelStrategy.WakeUp()
	}

	// Should complete quickly after wakeup
	select {
	case <-done:
		// Good, wakeup worked
	case <-time.After(50 * time.Millisecond):
		t.Error("Channel strategy did not wake up in time")
	}
}

func TestYieldingIdleStrategy_GoschedCalls(t *testing.T) {
	strategy := NewYieldingIdleStrategy(3) // Yield after 3 spins

	// This test verifies that the yielding strategy calls runtime.Gosched
	// We can't easily test if Gosched is actually called, but we can test
	// that the strategy behaves correctly

	for i := 0; i < 10; i++ {
		if !strategy.Idle() {
			t.Error("Yielding strategy should continue when idle")
		}
	}

	strategy.Reset()

	// After reset, should start fresh
	if !strategy.Idle() {
		t.Error("Yielding strategy should continue after reset")
	}
}

func TestSpinningIdleStrategy_NeverYields(t *testing.T) {
	strategy := NewSpinningIdleStrategy()

	// Spinning strategy should never sleep or block
	start := time.Now()

	for i := 0; i < 1000; i++ {
		if !strategy.Idle() {
			t.Error("Spinning strategy should always continue")
		}
	}

	elapsed := time.Since(start)

	// Should complete very quickly with pure spinning
	if elapsed > 10*time.Millisecond {
		t.Errorf("Spinning strategy took too long: %v", elapsed)
	}

	// Reset should be instant
	strategy.Reset()
}

func TestIdleStrategy_StringRepresentation(t *testing.T) {
	strategies := []struct {
		strategy     IdleStrategy
		expectedName string
	}{
		{NewSpinningIdleStrategy(), "spinning"},
		{NewSleepingIdleStrategy(time.Millisecond, 0), "sleeping"},
		{NewYieldingIdleStrategy(1000), "yielding"},
		{NewChannelIdleStrategy(0), "channel"},
		{NewProgressiveIdleStrategy(), "progressive"},
	}

	for _, test := range strategies {
		if test.strategy.String() != test.expectedName {
			t.Errorf("Expected strategy name %q, got %q", test.expectedName, test.strategy.String())
		}
	}
}

func TestConfig_DefaultIdleStrategy(t *testing.T) {
	// Test that config uses BalancedStrategy by default
	config := &Config{
		Output:  &TestWriteSyncer{},
		Encoder: NewJSONEncoder(),
	}

	configWithDefaults := config.withDefaults()

	if configWithDefaults.IdleStrategy == nil {
		t.Error("Expected default idle strategy to be set")
	}

	// The default should be BalancedStrategy (progressive)
	if configWithDefaults.IdleStrategy.String() != "progressive" {
		t.Errorf("Expected default strategy to be 'progressive', got %q", configWithDefaults.IdleStrategy.String())
	}
}

func TestLogger_WithCustomIdleStrategy(t *testing.T) {
	// Test creating a logger with custom idle strategy
	config := &Config{
		Output:       &TestWriteSyncer{},
		Encoder:      NewJSONEncoder(),
		IdleStrategy: NewSleepingIdleStrategy(time.Millisecond, 100),
	}

	logger, err := New(*config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Start()

	// Log a message to verify it works
	logger.Info("Test message with custom idle strategy")

	// Give time for processing
	time.Sleep(10 * time.Millisecond)
	logger.Sync()
}

// TestWriteSyncer is a simple WriteSyncer for testing
type TestWriteSyncer struct {
	data []byte
}

func (t *TestWriteSyncer) Write(p []byte) (n int, err error) {
	t.data = append(t.data, p...)
	return len(p), nil
}

func (t *TestWriteSyncer) Sync() error {
	return nil
}
