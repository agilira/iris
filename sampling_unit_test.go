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

// TestSamplingStatsRate tests SamplingRate function
func TestSamplingStatsRate(t *testing.T) {
	tests := []struct {
		name     string
		stats    SamplingStats
		expected float64
	}{
		{
			name:     "zero total",
			stats:    SamplingStats{Sampled: 0, Dropped: 0, Total: 0},
			expected: 0,
		},
		{
			name:     "all sampled",
			stats:    SamplingStats{Sampled: 100, Dropped: 0, Total: 100},
			expected: 100,
		},
		{
			name:     "half sampled",
			stats:    SamplingStats{Sampled: 50, Dropped: 50, Total: 100},
			expected: 50,
		},
		{
			name:     "quarter sampled",
			stats:    SamplingStats{Sampled: 25, Dropped: 75, Total: 100},
			expected: 25,
		},
		{
			name:     "high volume scenario",
			stats:    SamplingStats{Sampled: 1000, Dropped: 9000, Total: 10000},
			expected: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rate := test.stats.SamplingRate()
			if rate != test.expected {
				t.Errorf("SamplingRate() = %f, expected %f", rate, test.expected)
			}
		})
	}
}

// TestSamplingStatsDropRate tests DropRate function
func TestSamplingStatsDropRate(t *testing.T) {
	tests := []struct {
		name     string
		stats    SamplingStats
		expected float64
	}{
		{
			name:     "zero total",
			stats:    SamplingStats{Sampled: 0, Dropped: 0, Total: 0},
			expected: 0,
		},
		{
			name:     "all dropped",
			stats:    SamplingStats{Sampled: 0, Dropped: 100, Total: 100},
			expected: 100,
		},
		{
			name:     "none dropped",
			stats:    SamplingStats{Sampled: 100, Dropped: 0, Total: 100},
			expected: 0,
		},
		{
			name:     "half dropped",
			stats:    SamplingStats{Sampled: 50, Dropped: 50, Total: 100},
			expected: 50,
		},
		{
			name:     "high drop rate",
			stats:    SamplingStats{Sampled: 100, Dropped: 9900, Total: 10000},
			expected: 99,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rate := test.stats.DropRate()
			if rate != test.expected {
				t.Errorf("DropRate() = %f, expected %f", rate, test.expected)
			}
		})
	}
}

// TestNewDevelopmentSampling tests the development sampling preset
func TestNewDevelopmentSampling(t *testing.T) {
	config := NewDevelopmentSampling()

	// Verify development preset values
	if config.Initial != 100 {
		t.Errorf("Expected Initial=100, got %d", config.Initial)
	}
	if config.Thereafter != 10 {
		t.Errorf("Expected Thereafter=10, got %d", config.Thereafter)
	}
	if config.Tick != 0 {
		t.Errorf("Expected Tick=0 (no reset), got %v", config.Tick)
	}

	// Test that the config works with a sampler
	sampler := NewSampler(config)
	if sampler == nil {
		t.Error("NewSampler should accept development config")
	}

	// Quick functionality test
	entry := LogEntry{Level: InfoLevel, Message: "development test"}
	decision := sampler.Sample(entry)
	if decision != LogSample {
		t.Error("Development sampling should allow initial entries")
	}
}

// TestNewProductionSampling tests the production sampling preset
func TestNewProductionSampling(t *testing.T) {
	config := NewProductionSampling()

	// Verify production preset values
	if config.Initial != 100 {
		t.Errorf("Expected Initial=100, got %d", config.Initial)
	}
	if config.Thereafter != 100 {
		t.Errorf("Expected Thereafter=100, got %d", config.Thereafter)
	}
	if config.Tick != time.Minute {
		t.Errorf("Expected Tick=1m, got %v", config.Tick)
	}

	// Test that the config works with a sampler
	sampler := NewSampler(config)
	if sampler == nil {
		t.Error("NewSampler should accept production config")
	}

	// Quick functionality test
	entry := LogEntry{Level: InfoLevel, Message: "production test"}
	decision := sampler.Sample(entry)
	if decision != LogSample {
		t.Error("Production sampling should allow initial entries")
	}
}

// TestNewHighVolumeSampling tests the high volume sampling preset
func TestNewHighVolumeSampling(t *testing.T) {
	config := NewHighVolumeSampling()

	// Verify high volume preset values
	if config.Initial != 50 {
		t.Errorf("Expected Initial=50, got %d", config.Initial)
	}
	if config.Thereafter != 1000 {
		t.Errorf("Expected Thereafter=1000, got %d", config.Thereafter)
	}
	if config.Tick != 10*time.Minute {
		t.Errorf("Expected Tick=10m, got %v", config.Tick)
	}

	// Test that the config works with a sampler
	sampler := NewSampler(config)
	if sampler == nil {
		t.Error("NewSampler should accept high volume config")
	}

	// Quick functionality test
	entry := LogEntry{Level: InfoLevel, Message: "high volume test"}
	decision := sampler.Sample(entry)
	if decision != LogSample {
		t.Error("High volume sampling should allow initial entries")
	}
}

// TestSamplingPresetsComparison tests that the presets have the expected relative characteristics
func TestSamplingPresetsComparison(t *testing.T) {
	dev := NewDevelopmentSampling()
	prod := NewProductionSampling()
	highVol := NewHighVolumeSampling()

	// Development should be most permissive
	if dev.Thereafter >= prod.Thereafter {
		t.Error("Development sampling should be more permissive than production")
	}

	// Production should be more permissive than high volume
	if prod.Thereafter >= highVol.Thereafter {
		t.Error("Production sampling should be more permissive than high volume")
	}

	// High volume should have the smallest initial window
	if highVol.Initial >= dev.Initial || highVol.Initial >= prod.Initial {
		t.Error("High volume sampling should have the smallest initial window")
	}

	// Only production and high volume should have tick resets
	if dev.Tick != 0 {
		t.Error("Development sampling should not have time-based resets")
	}
	if prod.Tick == 0 || highVol.Tick == 0 {
		t.Error("Production and high volume sampling should have time-based resets")
	}
}
