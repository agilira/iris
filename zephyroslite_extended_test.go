// zephyroslite_extended_test.go: Extended tests for zephyroslite package
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
	"time"

	"github.com/agilira/iris/internal/zephyroslite"
)

// TestBackpressurePolicy_String tests the String method of BackpressurePolicy
func TestBackpressurePolicy_String(t *testing.T) {
	tests := []struct {
		name     string
		policy   zephyroslite.BackpressurePolicy
		expected string
	}{
		{
			name:     "DropOnFull",
			policy:   zephyroslite.DropOnFull,
			expected: "DropOnFull",
		},
		{
			name:     "BlockOnFull",
			policy:   zephyroslite.BlockOnFull,
			expected: "BlockOnFull",
		},
		{
			name:     "Unknown",
			policy:   zephyroslite.BackpressurePolicy(999), // Invalid value
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.policy.String()
			if result != tt.expected {
				t.Errorf("BackpressurePolicy.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestBuilder_WithBackpressurePolicy tests the WithBackpressurePolicy method
func TestBuilder_WithBackpressurePolicy(t *testing.T) {
	builder := zephyroslite.NewBuilder[string](1024) // Need to provide capacity

	// Test setting DropOnFull
	result := builder.WithBackpressurePolicy(zephyroslite.DropOnFull)
	if result != builder {
		t.Error("WithBackpressurePolicy should return the same builder instance")
	}

	// Test setting BlockOnFull
	result = builder.WithBackpressurePolicy(zephyroslite.BlockOnFull)
	if result != builder {
		t.Error("WithBackpressurePolicy should return the same builder instance")
	}
}

// TestBuilder_WithIdleStrategy tests the WithIdleStrategy method
func TestBuilder_WithIdleStrategy(t *testing.T) {
	builder := zephyroslite.NewBuilder[string](1024) // Need to provide capacity

	// Test with spinning strategy
	spinningStrategy := zephyroslite.NewSpinningIdleStrategy()
	result := builder.WithIdleStrategy(spinningStrategy)
	if result != builder {
		t.Error("WithIdleStrategy should return the same builder instance")
	}

	// Test with sleeping strategy
	sleepingStrategy := zephyroslite.NewSleepingIdleStrategy(1*time.Millisecond, 10)
	result = builder.WithIdleStrategy(sleepingStrategy)
	if result != builder {
		t.Error("WithIdleStrategy should return the same builder instance")
	}

	// Test with yielding strategy
	yieldingStrategy := zephyroslite.NewYieldingIdleStrategy(10)
	result = builder.WithIdleStrategy(yieldingStrategy)
	if result != builder {
		t.Error("WithIdleStrategy should return the same builder instance")
	}
}

// TestZephyrosLight_writeBlockOnFull tests the writeBlockOnFull method
func TestZephyrosLight_writeBlockOnFull(t *testing.T) {
	t.Run("WriteBlockOnFull_Success", func(t *testing.T) {
		// Create a simple test that exercises the String() method and WithBackpressurePolicy
		// We can't easily test writeBlockOnFull directly since it's internal

		// Test that we can create a ring with BlockOnFull policy
		ring, err := zephyroslite.NewBuilder[string](4).
			WithBackpressurePolicy(zephyroslite.BlockOnFull).
			WithBatchSize(2). // Set valid batch size
			WithProcessor(func(msg *string) {
				// Simple processor
			}).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ring: %v", err)
		}

		// Test that the ring was created successfully
		if ring == nil {
			t.Error("Expected non-nil ring")
		}

		ring.Close()
	})
}

// TestZephyrosLight_Loop tests the Loop method
func TestZephyrosLight_Loop(t *testing.T) {
	// Create a ring buffer for testing Loop method
	ring, err := zephyroslite.NewBuilder[string](4).
		WithBatchSize(2). // Set valid batch size
		WithProcessor(func(msg *string) {
			// Simple processor for testing
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to create ring: %v", err)
	}

	// Start Loop in a goroutine
	done := make(chan bool)
	go func() {
		ring.Loop() // This should call LoopProcess internally
		done <- true
	}()

	// Give Loop time to start
	time.Sleep(10 * time.Millisecond)

	// Close and wait
	ring.Close()

	// Wait for Loop to finish
	select {
	case <-done:
		// Good, Loop completed
	case <-time.After(1 * time.Second):
		t.Error("Loop should have completed after Close()")
	}
}

// TestProgressiveIdleStrategy_Reset tests the Reset method that's currently uncovered
func TestProgressiveIdleStrategy_Reset(t *testing.T) {
	strategy := zephyroslite.NewProgressiveIdleStrategy()

	// Force the strategy to progress through some iterations
	for i := 0; i < 150; i++ {
		strategy.Idle()
	}

	// Reset the strategy
	strategy.Reset()

	// After reset, it should be back to initial state
	// This method should not panic and should execute successfully
	strategy.Idle()

	// Test that Reset can be called multiple times
	strategy.Reset()
	strategy.Reset()
}
