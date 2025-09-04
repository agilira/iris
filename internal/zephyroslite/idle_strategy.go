// idle_strategy.go: Configurable idle strategies for Zephyros consumer loops
//
// This file implements various idle strategies that control CPU usage when
// the consumer loop has no work to process. These strategies provide different
// trade-offs between latency and CPU consumption.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"runtime"
	"sync/atomic"
	"time"
)

// IdleStrategy defines the interface for consumer idle behavior.
// When the consumer loop has no work to process, the idle strategy
// determines how to wait for new work while balancing latency and CPU usage.
type IdleStrategy interface {
	// Idle is called when no work is available.
	// Returns true if the caller should continue processing,
	// false if it should check for shutdown.
	Idle() bool

	// Reset is called when work is found to reset any internal state.
	Reset()

	// String returns a human-readable name for the strategy.
	String() string
}

// SpinningIdleStrategy provides ultra-low latency with maximum CPU usage.
// This strategy never yields the CPU and continuously checks for work.
// Best for: Ultra-low latency requirements where CPU consumption is not a concern.
// CPU Usage: ~100% of one core
// Latency: Minimum possible (~nanoseconds)
type SpinningIdleStrategy struct{}

// NewSpinningIdleStrategy creates a new spinning idle strategy.
func NewSpinningIdleStrategy() *SpinningIdleStrategy {
	return &SpinningIdleStrategy{}
}

func (s *SpinningIdleStrategy) Idle() bool {
	// Pure spin - no yielding, no sleeping
	return true
}

func (s *SpinningIdleStrategy) Reset() {
	// No state to reset
}

func (s *SpinningIdleStrategy) String() string {
	return "spinning"
}

// SleepingIdleStrategy reduces CPU usage with controlled latency increase.
// This strategy uses progressive backoff with configurable sleep duration.
// Best for: Balanced CPU usage and latency in production environments.
// CPU Usage: ~1-10% depending on configuration
// Latency: ~1-10ms depending on sleep duration
type SleepingIdleStrategy struct {
	sleepDuration time.Duration
	spins         int
	maxSpins      int
}

// NewSleepingIdleStrategy creates a new sleeping idle strategy.
// sleepDuration: How long to sleep when no work is found
// maxSpins: Number of spin iterations before sleeping (0 = sleep immediately)
func NewSleepingIdleStrategy(sleepDuration time.Duration, maxSpins int) *SleepingIdleStrategy {
	if sleepDuration <= 0 {
		sleepDuration = time.Millisecond // Default 1ms
	}
	if maxSpins < 0 {
		maxSpins = 0
	}
	return &SleepingIdleStrategy{
		sleepDuration: sleepDuration,
		maxSpins:      maxSpins,
	}
}

func (s *SleepingIdleStrategy) Idle() bool {
	if s.spins < s.maxSpins {
		s.spins++
		// Spin first, then sleep
		return true
	}

	// Sleep to reduce CPU usage
	time.Sleep(s.sleepDuration)
	return true
}

func (s *SleepingIdleStrategy) Reset() {
	s.spins = 0
}

func (s *SleepingIdleStrategy) String() string {
	return "sleeping"
}

// YieldingIdleStrategy provides a middle ground using runtime.Gosched().
// This strategy yields to the Go scheduler after a configurable number of spins.
// Best for: Moderate CPU reduction while maintaining reasonable latency.
// CPU Usage: ~10-50% depending on configuration
// Latency: ~microseconds to low milliseconds
type YieldingIdleStrategy struct {
	spins    int
	maxSpins int
}

// NewYieldingIdleStrategy creates a new yielding idle strategy.
// maxSpins: Number of spins before yielding to scheduler
func NewYieldingIdleStrategy(maxSpins int) *YieldingIdleStrategy {
	if maxSpins <= 0 {
		maxSpins = 1000 // Default: yield every 1000 spins
	}
	return &YieldingIdleStrategy{
		maxSpins: maxSpins,
	}
}

func (s *YieldingIdleStrategy) Idle() bool {
	s.spins++
	if s.spins >= s.maxSpins {
		runtime.Gosched()
		s.spins = 0
	}
	return true
}

func (s *YieldingIdleStrategy) Reset() {
	s.spins = 0
}

func (s *YieldingIdleStrategy) String() string {
	return "yielding"
}

// ChannelIdleStrategy provides efficient blocking wait using Go channels.
// This strategy puts the goroutine into an efficient wait state until
// new work arrives. Requires coordination with the producer side.
// Best for: Minimum CPU usage with acceptable latency for low-throughput scenarios.
// CPU Usage: Near 0% when idle
// Latency: ~microseconds (channel wake-up time)
type ChannelIdleStrategy struct {
	wakeupChan  chan struct{}
	timeoutChan <-chan time.Time
	timeout     time.Duration
	timer       *time.Timer
}

// NewChannelIdleStrategy creates a new channel-based idle strategy.
// timeout: Maximum time to wait before checking for shutdown (0 = no timeout)
func NewChannelIdleStrategy(timeout time.Duration) *ChannelIdleStrategy {
	strategy := &ChannelIdleStrategy{
		wakeupChan: make(chan struct{}, 1), // Buffered to prevent blocking
		timeout:    timeout,
	}

	if timeout > 0 {
		strategy.timer = time.NewTimer(timeout)
		strategy.timeoutChan = strategy.timer.C
	}

	return strategy
}

func (s *ChannelIdleStrategy) Idle() bool {
	if s.timeout > 0 {
		// Wait with timeout
		select {
		case <-s.wakeupChan:
			// Work available signal received
			if !s.timer.Stop() {
				// Drain timer channel if it fired
				select {
				case <-s.timer.C:
				default:
				}
			}
			s.timer.Reset(s.timeout)
			return true
		case <-s.timeoutChan:
			// Timeout - reset timer and continue
			s.timer.Reset(s.timeout)
			return true
		}
	} else {
		// Wait indefinitely
		<-s.wakeupChan
		return true
	}
}

func (s *ChannelIdleStrategy) Reset() {
	// Signal that work is available
	select {
	case s.wakeupChan <- struct{}{}:
	default:
		// Channel already has a signal, no need to add another
	}
}

func (s *ChannelIdleStrategy) String() string {
	return "channel"
}

// WakeUp signals the channel strategy that work may be available.
// This should be called by producers when they add work to the queue.
func (s *ChannelIdleStrategy) WakeUp() {
	select {
	case s.wakeupChan <- struct{}{}:
	default:
		// Channel already has a signal, no need to block
	}
}

// Progressive Idle Strategy provides adaptive behavior based on work patterns.
// This strategy starts with spinning for ultra-low latency, then progressively
// reduces CPU usage as idle time increases.
// Best for: Variable workload patterns where both low latency and low CPU usage are important.
type ProgressiveIdleStrategy struct {
	spins        int64 // atomic counter
	sleepCounter int64 // atomic counter

	// Configuration thresholds
	hotSpinThreshold  int           // Spins before first yield
	warmSpinThreshold int           // Spins before sleep starts
	sleepDuration     time.Duration // Initial sleep duration
	maxSleepDuration  time.Duration // Maximum sleep duration
}

// NewProgressiveIdleStrategy creates a new progressive idle strategy.
func NewProgressiveIdleStrategy() *ProgressiveIdleStrategy {
	return &ProgressiveIdleStrategy{
		hotSpinThreshold:  1000,             // Hot spin for 1000 iterations
		warmSpinThreshold: 10000,            // Then yield occasionally until 10000
		sleepDuration:     time.Microsecond, // Start with 1μs sleep
		maxSleepDuration:  time.Millisecond, // Up to 1ms max
	}
}

func (s *ProgressiveIdleStrategy) Idle() bool {
	spins := atomic.AddInt64(&s.spins, 1)

	if spins < int64(s.hotSpinThreshold) {
		// Hot spin for minimum latency
		return true
	} else if spins < int64(s.warmSpinThreshold) {
		// Occasional yield
		if spins&7 == 0 { // Every 8 iterations
			runtime.Gosched()
		}
		return true
	} else {
		// Progressive sleep with backoff
		sleepCounter := atomic.LoadInt64(&s.sleepCounter)
		shift := sleepCounter / 2
		if shift > 10 {
			shift = 10
		}
		sleepDuration := s.sleepDuration * time.Duration(1<<shift)
		if sleepDuration > s.maxSleepDuration {
			sleepDuration = s.maxSleepDuration
		}

		time.Sleep(sleepDuration)
		atomic.AddInt64(&s.sleepCounter, 1)
		atomic.StoreInt64(&s.spins, 0) // Reset spin count after sleep
		return true
	}
}

func (s *ProgressiveIdleStrategy) Reset() {
	atomic.StoreInt64(&s.spins, 0)
	atomic.StoreInt64(&s.sleepCounter, 0)
}

func (s *ProgressiveIdleStrategy) String() string {
	return "progressive"
}
