// zephyros_direct_test.go: Direct tests for specific uncovered zephyroslite functions
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync"
	"testing"
	"time"

	"github.com/agilira/iris/internal/zephyroslite"
)

// TestZephyroslite_DirectFunctionCalls tests specific functions that need coverage
func TestZephyroslite_DirectFunctionCalls(t *testing.T) {
	t.Run("Test_WithBackpressurePolicy_Coverage", func(t *testing.T) {
		// Test that WithBackpressurePolicy actually gets called and returns correct builder
		builder := zephyroslite.NewBuilder[string](4)

		// These calls should execute the WithBackpressurePolicy function
		builder1 := builder.WithBackpressurePolicy(zephyroslite.DropOnFull)
		builder2 := builder1.WithBackpressurePolicy(zephyroslite.BlockOnFull)

		if builder1 != builder || builder2 != builder {
			t.Error("WithBackpressurePolicy should return the same builder instance")
		}
	})

	t.Run("Test_WithIdleStrategy_Coverage", func(t *testing.T) {
		// Test that WithIdleStrategy actually gets called
		builder := zephyroslite.NewBuilder[string](4)

		// These calls should execute the WithIdleStrategy function
		spinStrategy := zephyroslite.NewSpinningIdleStrategy()
		builder1 := builder.WithIdleStrategy(spinStrategy)

		sleepStrategy := zephyroslite.NewSleepingIdleStrategy(1*time.Millisecond, 5)
		builder2 := builder1.WithIdleStrategy(sleepStrategy)

		if builder1 != builder || builder2 != builder {
			t.Error("WithIdleStrategy should return the same builder instance")
		}
	})

	t.Run("Test_BackpressurePolicy_String_Coverage", func(t *testing.T) {
		// Test all possible String() values for BackpressurePolicy
		policies := []struct {
			policy   zephyroslite.BackpressurePolicy
			expected string
		}{
			{zephyroslite.DropOnFull, "DropOnFull"},
			{zephyroslite.BlockOnFull, "BlockOnFull"},
			{zephyroslite.BackpressurePolicy(999), "Unknown"}, // Invalid value
		}

		for _, test := range policies {
			result := test.policy.String()
			if result != test.expected {
				t.Errorf("BackpressurePolicy(%d).String() = %s, want %s", test.policy, result, test.expected)
			}
		}
	})

	t.Run("Test_Loop_Method_Coverage", func(t *testing.T) {
		// Create a ring to test the Loop method specifically
		ring, err := zephyroslite.NewBuilder[string](4).
			WithBatchSize(2).
			WithProcessor(func(msg *string) {
				// Simple processor
			}).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ring: %v", err)
		}

		// Test Loop method by calling it in a goroutine
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			// This should call the Loop() method specifically
			ring.Loop()
		}()

		// Give it time to start
		time.Sleep(10 * time.Millisecond)

		// Close to stop the loop
		ring.Close()

		// Wait for completion
		wg.Wait()
	})

	t.Run("Test_writeBlockOnFull_Indirectly", func(t *testing.T) {
		// Test that forces the use of writeBlockOnFull by using BlockOnFull policy
		ring, err := zephyroslite.NewBuilder[string](2). // Small buffer
									WithBatchSize(1).
									WithBackpressurePolicy(zephyroslite.BlockOnFull). // This should use writeBlockOnFull
									WithProcessor(func(msg *string) {
				time.Sleep(50 * time.Millisecond) // Slow processing to create backpressure
			}).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ring: %v", err)
		}

		// Start consumer
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			ring.Loop()
		}()

		// Give consumer time to start
		time.Sleep(10 * time.Millisecond)

		// Write to trigger writeBlockOnFull
		success := ring.Write(func(slot *string) {
			*slot = "test message"
		})

		if !success {
			t.Error("Write should succeed with BlockOnFull policy")
		}

		// Close and cleanup
		ring.Close()
		wg.Wait()
	})
}
