// idle_strategies.go: Public API for configurable consumer idle strategies
//
// This file provides convenient factory functions for creating different
// idle strategies that control CPU usage when the logger consumer has no
// work to process. Each strategy provides different trade-offs between
// latency and CPU consumption.
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"time"

	"github.com/agilira/iris/internal/zephyroslite"
)

// IdleStrategy defines the interface for consumer idle behavior.
// This type alias exposes the internal interface for configuration purposes.
type IdleStrategy = zephyroslite.IdleStrategy

// NewSpinningIdleStrategy creates an ultra-low latency idle strategy.
// This strategy provides the minimum possible latency by continuously
// checking for work without ever yielding the CPU.
//
// Best for: Ultra-low latency requirements where CPU consumption is not a concern
// CPU Usage: ~100% of one core when idle
// Latency: Minimum possible (~nanoseconds)
//
// Example:
//
//	config := &Config{
//	    IdleStrategy: NewSpinningIdleStrategy(),
//	    // ... other config
//	}
func NewSpinningIdleStrategy() IdleStrategy {
	return zephyroslite.NewSpinningIdleStrategy()
}

// NewSleepingIdleStrategy creates a CPU-efficient idle strategy with controlled latency.
// This strategy reduces CPU usage by sleeping when no work is available,
// with optional initial spinning for hybrid behavior.
//
// Parameters:
//   - sleepDuration: How long to sleep when no work is found (e.g., time.Millisecond)
//   - maxSpins: Number of spin iterations before sleeping (0 = sleep immediately)
//
// Best for: Balanced CPU usage and latency in production environments
// CPU Usage: ~1-10% depending on sleep duration and spin count
// Latency: ~1-10ms depending on sleep duration
//
// Examples:
//
//	// Low CPU usage, higher latency
//	NewSleepingIdleStrategy(5*time.Millisecond, 0)
//
//	// Hybrid: spin briefly then sleep
//	NewSleepingIdleStrategy(time.Millisecond, 1000)
func NewSleepingIdleStrategy(sleepDuration time.Duration, maxSpins int) IdleStrategy {
	return zephyroslite.NewSleepingIdleStrategy(sleepDuration, maxSpins)
}

// NewYieldingIdleStrategy creates a moderate CPU reduction strategy.
// This strategy spins for a configurable number of iterations before
// yielding to the Go scheduler, providing a middle ground between
// spinning and sleeping approaches.
//
// Parameters:
//   - maxSpins: Number of spins before yielding to scheduler
//
// Best for: Moderate CPU reduction while maintaining reasonable latency
// CPU Usage: ~10-50% depending on max spins configuration
// Latency: ~microseconds to low milliseconds
//
// Examples:
//
//	// More aggressive yielding (lower CPU, higher latency)
//	NewYieldingIdleStrategy(100)
//
//	// Conservative yielding (higher CPU, lower latency)
//	NewYieldingIdleStrategy(10000)
func NewYieldingIdleStrategy(maxSpins int) IdleStrategy {
	return zephyroslite.NewYieldingIdleStrategy(maxSpins)
}

// NewChannelIdleStrategy creates an efficient blocking wait strategy.
// This strategy puts the consumer goroutine into an efficient wait state
// using Go channels, providing near-zero CPU usage when idle.
//
// Parameters:
//   - timeout: Maximum time to wait before checking for shutdown (0 = no timeout)
//
// Best for: Minimum CPU usage with acceptable latency for low-throughput scenarios
// CPU Usage: Near 0% when idle
// Latency: ~microseconds (channel wake-up time)
//
// Note: This strategy works best with lower throughput workloads where
// the overhead of channel operations is acceptable.
//
// Examples:
//
//	// No timeout - maximum efficiency
//	NewChannelIdleStrategy(0)
//
//	// With timeout for responsive shutdown
//	NewChannelIdleStrategy(100*time.Millisecond)
func NewChannelIdleStrategy(timeout time.Duration) IdleStrategy {
	return zephyroslite.NewChannelIdleStrategy(timeout)
}

// NewProgressiveIdleStrategy creates an adaptive idle strategy.
// This strategy automatically adjusts its behavior based on work patterns,
// starting with spinning for ultra-low latency and progressively reducing
// CPU usage as idle time increases.
//
// This is the default strategy, providing good performance for most workloads
// without requiring manual tuning.
//
// Best for: Variable workload patterns where both low latency and low CPU usage are important
// CPU Usage: Adaptive - starts high, reduces over time when idle
// Latency: Starts at minimum, increases gradually when idle
//
// Behavior:
//   - Hot spin for first 1000 iterations (minimum latency)
//   - Occasional yielding up to 10000 iterations
//   - Progressive sleep with exponential backoff
//   - Resets to hot spin when work is found
//
// Example:
//
//	config := &Config{
//	    IdleStrategy: NewProgressiveIdleStrategy(),
//	    // ... other config
//	}
func NewProgressiveIdleStrategy() IdleStrategy {
	return zephyroslite.NewProgressiveIdleStrategy()
}

// Predefined idle strategies for common use cases

// SpinningStrategy provides ultra-low latency with maximum CPU usage.
// Equivalent to NewSpinningIdleStrategy().
var SpinningStrategy = NewSpinningIdleStrategy()

// BalancedStrategy provides good performance for most production workloads.
// Uses progressive strategy that adapts to workload patterns.
// Equivalent to NewProgressiveIdleStrategy().
var BalancedStrategy = NewProgressiveIdleStrategy()

// EfficientStrategy minimizes CPU usage for low-throughput scenarios.
// Uses 1ms sleep with no initial spinning.
// Equivalent to NewSleepingIdleStrategy(time.Millisecond, 0).
var EfficientStrategy = NewSleepingIdleStrategy(time.Millisecond, 0)

// HybridStrategy provides a good compromise between latency and CPU usage.
// Spins briefly then sleeps for 1ms.
// Equivalent to NewSleepingIdleStrategy(time.Millisecond, 1000).
var HybridStrategy = NewSleepingIdleStrategy(time.Millisecond, 1000)
