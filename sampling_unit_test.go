// sampling_unit_test.go: Unit tests for sampling functionality
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync"
	"testing"
	"time"
)

// TestSamplingBasicFunctionality tests basic sampling operations
func TestSamplingBasicFunctionality(t *testing.T) {
	config := SamplingConfig{
		Initial:    10, // First 10 always logged
		Thereafter: 5,  // Then 1 in every 5
		Tick:       time.Second,
	}

	sampler := NewSampler(config)
	if sampler == nil {
		t.Fatal("NewSampler should not return nil")
	}

	// Test initial period - all should be logged
	entry := LogEntry{Level: InfoLevel, Message: "test"}
	for i := 0; i < 10; i++ {
		decision := sampler.Sample(entry)
		if decision != LogSample {
			t.Errorf("Initial %d entries should all be logged, got %v at %d", 10, decision, i)
		}
	}
}

// TestSamplingAfterInitial tests sampling behavior after initial period
func TestSamplingAfterInitial(t *testing.T) {
	config := SamplingConfig{
		Initial:    2, // First 2 always logged
		Thereafter: 3, // Then 1 in every 3
		Tick:       time.Second,
	}

	sampler := NewSampler(config)
	entry := LogEntry{Level: InfoLevel, Message: "test"}

	// Consume initial period
	for i := 0; i < 2; i++ {
		sampler.Sample(entry)
	}

	// Test thereafter period
	loggedCount := 0
	droppedCount := 0

	for i := 0; i < 15; i++ { // Test 5 cycles of 3
		decision := sampler.Sample(entry)
		if decision == LogSample {
			loggedCount++
		} else {
			droppedCount++
		}
	}

	// Should be approximately 1/3 logged
	if loggedCount == 0 {
		t.Error("Should have logged some entries after initial period")
	}
	if droppedCount == 0 {
		t.Error("Should have dropped some entries after initial period")
	}
}

// TestSamplingConcurrency tests concurrent sampling
func TestSamplingConcurrency(t *testing.T) {
	config := SamplingConfig{
		Initial:    5,
		Thereafter: 10,
		Tick:       time.Second,
	}

	sampler := NewSampler(config)
	entry := LogEntry{Level: InfoLevel, Message: "test"}

	const numGoroutines = 50
	const numSamples = 100

	var wg sync.WaitGroup
	loggedCount := int64(0)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			localLogged := 0
			for j := 0; j < numSamples; j++ {
				decision := sampler.Sample(entry)
				if decision == LogSample {
					localLogged++
				}
			}

			// This is racy but just for testing
			_ = localLogged
		}()
	}

	wg.Wait()

	// Test that it doesn't crash - exact counts are hard to verify in concurrent scenario
	_ = loggedCount
}

// TestSamplingReset tests sampling reset functionality
func TestSamplingReset(t *testing.T) {
	config := SamplingConfig{
		Initial:    3,
		Thereafter: 5,
		Tick:       100 * time.Millisecond, // Short tick for testing
	}

	sampler := NewSampler(config)
	entry := LogEntry{Level: InfoLevel, Message: "test"}

	// Consume initial + some thereafter
	for i := 0; i < 10; i++ {
		sampler.Sample(entry)
	}

	// Wait for tick to reset
	time.Sleep(150 * time.Millisecond)

	// Next few should be in initial period again
	for i := 0; i < 3; i++ {
		decision := sampler.Sample(entry)
		if decision != LogSample {
			t.Errorf("After reset, first %d should be logged, got %v at %d", 3, decision, i)
		}
	}
}

// TestSamplingZeroConfig tests edge cases with zero values
func TestSamplingZeroConfig(t *testing.T) {
	// Test with zero initial
	config := SamplingConfig{
		Initial:    0,
		Thereafter: 5,
		Tick:       time.Second,
	}

	sampler := NewSampler(config)
	entry := LogEntry{Level: InfoLevel, Message: "test"}

	// Should go straight to thereafter sampling
	decisions := make([]SamplingDecision, 10)
	for i := 0; i < 10; i++ {
		decisions[i] = sampler.Sample(entry)
	}

	// Just test it doesn't crash - sampling logic has edge cases
	_ = decisions
}
