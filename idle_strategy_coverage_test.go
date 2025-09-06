// idle_strategy_coverage_test.go: test idle_strategy complex functions
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"

	"github.com/agilira/iris/internal/zephyroslite"
)

// TestUnit_ZephyrosliteIdleStrategy_Reset tests the Reset method coverage
func TestUnit_ZephyrosliteIdleStrategy_Reset(t *testing.T) {
	t.Parallel()

	// Test SpinningIdleStrategy Reset() - this is the 0% coverage function
	strategy := zephyroslite.NewSpinningIdleStrategy()

	// Call Reset - this should be a no-op for spinning strategy
	strategy.Reset()

	// Verify strategy still works after reset
	result := strategy.Idle() // No parameters
	if !result {
		t.Error("Idle should return true for spinning strategy")
	}

	// Verify String method still works
	strategyName := strategy.String()
	if strategyName != "spinning" {
		t.Errorf("Expected 'spinning', got '%s'", strategyName)
	}
}

// TestUnit_ZephyrosliteIdleStrategy_ComprehensiveCoverage tests all idle strategies
func TestUnit_ZephyrosliteIdleStrategy_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		createStrategy    func() zephyroslite.IdleStrategy
		expectedStringOut string
		shouldCallReset   bool
	}{
		{
			name: "SpinningIdleStrategy_Complete",
			createStrategy: func() zephyroslite.IdleStrategy {
				return zephyroslite.NewSpinningIdleStrategy()
			},
			expectedStringOut: "spinning",
			shouldCallReset:   true,
		},
		{
			name: "SleepingIdleStrategy_Complete",
			createStrategy: func() zephyroslite.IdleStrategy {
				return zephyroslite.NewSleepingIdleStrategy(1, 1000)
			},
			expectedStringOut: "sleeping",
			shouldCallReset:   true,
		},
		{
			name: "YieldingIdleStrategy_Complete",
			createStrategy: func() zephyroslite.IdleStrategy {
				return zephyroslite.NewYieldingIdleStrategy(5)
			},
			expectedStringOut: "yielding",
			shouldCallReset:   true,
		},
		{
			name: "ChannelIdleStrategy_Complete",
			createStrategy: func() zephyroslite.IdleStrategy {
				return zephyroslite.NewChannelIdleStrategy(100)
			},
			expectedStringOut: "channel",
			shouldCallReset:   true,
		},
		{
			name: "ProgressiveIdleStrategy_Complete",
			createStrategy: func() zephyroslite.IdleStrategy {
				return zephyroslite.NewProgressiveIdleStrategy()
			},
			expectedStringOut: "progressive",
			shouldCallReset:   true,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create strategy
			strategy := tc.createStrategy()
			if strategy == nil {
				t.Fatal("Strategy should not be nil")
			}

			// Test String method
			stringOutput := strategy.String()
			if stringOutput != tc.expectedStringOut {
				t.Errorf("Expected String() to return '%s', got '%s'",
					tc.expectedStringOut, stringOutput)
			}

			// Test Idle method
			idleResult := strategy.Idle()
			// Note: Different strategies return different boolean values
			// We just verify the method can be called successfully
			_ = idleResult // Consume result to avoid unused variable warning

			// Test Reset method if applicable
			if tc.shouldCallReset {
				strategy.Reset() // This should cover the Reset() method

				// Verify strategy still works after reset
				postResetIdle := strategy.Idle()
				_ = postResetIdle // Method should execute without error
			}
		})
	}
}

// TestEdgeCase_ZephyrosliteIdleStrategy_SpecialScenarios tests edge cases
func TestEdgeCase_ZephyrosliteIdleStrategy_SpecialScenarios(t *testing.T) {
	t.Parallel()

	t.Run("SpinningStrategy_MultipleResets", func(t *testing.T) {
		t.Parallel()

		strategy := zephyroslite.NewSpinningIdleStrategy()

		// Call Reset multiple times
		for i := 0; i < 10; i++ {
			strategy.Reset()
		}

		// Should still work normally
		result := strategy.Idle()
		_ = result // Strategy should execute without error
	})

	t.Run("ChannelStrategy_WakeUpAfterReset", func(t *testing.T) {
		t.Parallel()

		strategy := zephyroslite.NewChannelIdleStrategy(50)

		// Reset and verify WakeUp still works
		strategy.Reset()

		// WakeUp should work after reset
		strategy.WakeUp()

		// Idle should still function
		result := strategy.Idle()
		_ = result // Method should execute without error
	})

	t.Run("AllStrategies_StringConsistency", func(t *testing.T) {
		t.Parallel()

		strategies := map[string]zephyroslite.IdleStrategy{
			"spinning":    zephyroslite.NewSpinningIdleStrategy(),
			"sleeping":    zephyroslite.NewSleepingIdleStrategy(1, 1000),
			"yielding":    zephyroslite.NewYieldingIdleStrategy(3),
			"channel":     zephyroslite.NewChannelIdleStrategy(100),
			"progressive": zephyroslite.NewProgressiveIdleStrategy(),
		}

		for expectedName, strategy := range strategies {
			actualName := strategy.String()
			if actualName != expectedName {
				t.Errorf("Strategy %s returned wrong string: expected '%s', got '%s'",
					expectedName, expectedName, actualName)
			}

			// Also test reset for coverage
			strategy.Reset()
		}
	})
}

// BenchmarkSuite_ZephyrosliteIdleStrategy_Performance benchmarks idle strategies
func BenchmarkSuite_ZephyrosliteIdleStrategy_Performance(b *testing.B) {
	strategies := map[string]func() zephyroslite.IdleStrategy{
		"Spinning": func() zephyroslite.IdleStrategy {
			return zephyroslite.NewSpinningIdleStrategy()
		},
		"Sleeping": func() zephyroslite.IdleStrategy {
			return zephyroslite.NewSleepingIdleStrategy(1, 1000)
		},
		"Yielding": func() zephyroslite.IdleStrategy {
			return zephyroslite.NewYieldingIdleStrategy(5)
		},
		"Channel": func() zephyroslite.IdleStrategy {
			return zephyroslite.NewChannelIdleStrategy(100)
		},
		"Progressive": func() zephyroslite.IdleStrategy {
			return zephyroslite.NewProgressiveIdleStrategy()
		},
	}

	for name, createFn := range strategies {
		strategy := createFn()

		b.Run(name+"_Reset", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				strategy.Reset()
			}
		})

		b.Run(name+"_Idle", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				strategy.Idle()
			}
		})

		b.Run(name+"_String", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = strategy.String()
			}
		})
	}
}
