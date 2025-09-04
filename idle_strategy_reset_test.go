// idle_strategy_reset_test.go: Specific tests for Reset methods in idle strategies
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

// TestSpinningIdleStrategy_Reset tests the Reset method of SpinningIdleStrategy
func TestSpinningIdleStrategy_Reset(t *testing.T) {
	strategy := zephyroslite.NewSpinningIdleStrategy()

	// Call Reset - should not panic
	strategy.Reset()

	// Should still work after reset
	result := strategy.Idle()
	if !result {
		t.Error("SpinningIdleStrategy.Idle() should return true")
	}

	// Test multiple resets
	strategy.Reset()
	strategy.Reset()
}

// TestSleepingIdleStrategy_Reset tests the Reset method of SleepingIdleStrategy
func TestSleepingIdleStrategy_Reset(t *testing.T) {
	strategy := zephyroslite.NewSleepingIdleStrategy(1*time.Millisecond, 5)

	// Force some iterations first
	for i := 0; i < 10; i++ {
		strategy.Idle()
	}

	// Call Reset - should not panic
	strategy.Reset()

	// Should still work after reset
	result := strategy.Idle()
	if !result {
		t.Error("SleepingIdleStrategy.Idle() should return true")
	}
}

// TestYieldingIdleStrategy_Reset tests the Reset method of YieldingIdleStrategy
func TestYieldingIdleStrategy_Reset(t *testing.T) {
	strategy := zephyroslite.NewYieldingIdleStrategy(5)

	// Force some iterations first
	for i := 0; i < 10; i++ {
		strategy.Idle()
	}

	// Call Reset - should not panic
	strategy.Reset()

	// Should still work after reset
	result := strategy.Idle()
	if !result {
		t.Error("YieldingIdleStrategy.Idle() should return true")
	}
}

// TestChannelIdleStrategy_Reset tests the Reset method of ChannelIdleStrategy
func TestChannelIdleStrategy_Reset(t *testing.T) {
	strategy := zephyroslite.NewChannelIdleStrategy(1 * time.Millisecond)

	// Call Reset - should not panic
	strategy.Reset()

	// Should still work after reset (though it will return false quickly in channel strategy)
	strategy.Idle()
}

// TestIdleStrategy_String tests String methods for all strategies
func TestIdleStrategy_String(t *testing.T) {
	strategies := []struct {
		name     string
		strategy zephyroslite.IdleStrategy
		expected string
	}{
		{
			name:     "SpinningIdleStrategy",
			strategy: zephyroslite.NewSpinningIdleStrategy(),
			expected: "spinning",
		},
		{
			name:     "SleepingIdleStrategy",
			strategy: zephyroslite.NewSleepingIdleStrategy(1*time.Millisecond, 5),
			expected: "sleeping",
		},
		{
			name:     "YieldingIdleStrategy",
			strategy: zephyroslite.NewYieldingIdleStrategy(5),
			expected: "yielding",
		},
		{
			name:     "ChannelIdleStrategy",
			strategy: zephyroslite.NewChannelIdleStrategy(1 * time.Millisecond),
			expected: "channel",
		},
		{
			name:     "ProgressiveIdleStrategy",
			strategy: zephyroslite.NewProgressiveIdleStrategy(),
			expected: "progressive",
		},
	}

	for _, tt := range strategies {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.String()
			if result != tt.expected {
				t.Errorf("%s.String() = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}
