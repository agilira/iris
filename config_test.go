// config_test.go: Comprehensive test suite for iris logging configuration
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"os"
	"sync"
	"testing"
	"time"
)

// TestConfigDefaults tests that withDefaults applies correct default values
func TestConfigDefaults(t *testing.T) {
	config := &Config{}
	defaulted := config.withDefaults()

	if defaulted.Capacity != 1<<16 {
		t.Errorf("Expected default capacity 65536, got %d", defaulted.Capacity)
	}

	if defaulted.BatchSize != 32 {
		t.Errorf("Expected default batch size 32, got %d", defaulted.BatchSize)
	}

	if defaulted.Output == nil {
		t.Error("Expected default output to be set")
	}

	if defaulted.TimeFn == nil {
		t.Error("Expected default TimeFn to be set")
	}

	if defaulted.Level != Info {
		t.Errorf("Expected default level Info, got %v", defaulted.Level)
	}

	// Test that TimeFn works
	now := defaulted.TimeFn()
	if now.IsZero() {
		t.Error("Expected TimeFn to return valid time")
	}
}

// TestConfigWithDefaultsPreservesSetValues tests that existing values are preserved
func TestConfigWithDefaultsPreservesSetValues(t *testing.T) {
	buf := &bytes.Buffer{}
	customTime := func() time.Time { return time.Unix(1234567890, 0) }

	config := &Config{
		Capacity:  1024,
		BatchSize: 16,
		Output:    WrapWriter(buf),
		Level:     Debug,
		TimeFn:    customTime,
		Name:      "test-logger",
	}

	defaulted := config.withDefaults()

	if defaulted.Capacity != 1024 {
		t.Errorf("Expected preserved capacity 1024, got %d", defaulted.Capacity)
	}

	if defaulted.BatchSize != 16 {
		t.Errorf("Expected preserved batch size 16, got %d", defaulted.BatchSize)
	}

	if defaulted.Level != Debug {
		t.Errorf("Expected preserved level Debug, got %v", defaulted.Level)
	}

	if defaulted.Name != "test-logger" {
		t.Errorf("Expected preserved name 'test-logger', got %s", defaulted.Name)
	}

	// Test custom TimeFn
	expectedTime := time.Unix(1234567890, 0)
	if !defaulted.TimeFn().Equal(expectedTime) {
		t.Errorf("Expected custom time function to be preserved")
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorCode   string
	}{
		{
			name: "valid config",
			config: Config{
				Capacity:  1024,
				BatchSize: 32,
				Level:     Info,
			},
			expectError: false,
		},
		{
			name: "negative capacity",
			config: Config{
				Capacity:  -1,
				BatchSize: 32,
				Level:     Info,
			},
			expectError: true,
			errorCode:   "IRIS_INVALID_CONFIG",
		},
		{
			name: "zero capacity",
			config: Config{
				Capacity:  0,
				BatchSize: 32,
				Level:     Info,
			},
			expectError: true,
			errorCode:   "IRIS_INVALID_CONFIG",
		},
		{
			name: "negative batch size",
			config: Config{
				Capacity:  1024,
				BatchSize: -1,
				Level:     Info,
			},
			expectError: true,
			errorCode:   "IRIS_INVALID_CONFIG",
		},
		{
			name: "batch size exceeds capacity",
			config: Config{
				Capacity:  64,
				BatchSize: 128,
				Level:     Info,
			},
			expectError: true,
			errorCode:   "IRIS_INVALID_CONFIG",
		},
		{
			name: "invalid level",
			config: Config{
				Capacity:  1024,
				BatchSize: 32,
				Level:     Level(999),
			},
			expectError: true,
			errorCode:   "IRIS_INVALID_LEVEL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
					return
				}

				code := GetErrorCode(err)
				if string(code) != tt.errorCode {
					t.Errorf("Expected error code %s, got %s", tt.errorCode, string(code))
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got %v", err)
				}
			}
		})
	}
}

// TestConfigClone tests configuration cloning
func TestConfigClone(t *testing.T) {
	original := &Config{
		Capacity:  1024,
		BatchSize: 32,
		Output:    WrapWriter(&bytes.Buffer{}),
		Level:     Debug,
		TimeFn:    time.Now,
		Name:      "original",
	}

	clone := original.Clone()

	// Verify clone is not nil
	if clone == nil {
		t.Fatal("Clone returned nil")
	}

	// Verify clone has same values
	if clone.Capacity != original.Capacity {
		t.Errorf("Clone capacity mismatch: expected %d, got %d", original.Capacity, clone.Capacity)
	}

	if clone.BatchSize != original.BatchSize {
		t.Errorf("Clone batch size mismatch: expected %d, got %d", original.BatchSize, clone.BatchSize)
	}

	if clone.Level != original.Level {
		t.Errorf("Clone level mismatch: expected %v, got %v", original.Level, clone.Level)
	}

	if clone.Name != original.Name {
		t.Errorf("Clone name mismatch: expected %s, got %s", original.Name, clone.Name)
	}

	// Verify they are different objects
	if clone == original {
		t.Error("Clone should return a different object")
	}

	// Verify modifying clone doesn't affect original
	clone.Name = "modified"
	if original.Name == "modified" {
		t.Error("Modifying clone affected the original")
	}
}

// TestConfigCloneNil tests cloning nil config
func TestConfigCloneNil(t *testing.T) {
	var config *Config
	clone := config.Clone()

	if clone != nil {
		t.Error("Clone of nil config should return nil")
	}
}

// TestConfigAtomicLevel tests atomic level operations
func TestConfigAtomicLevel(t *testing.T) {
	config := &Config{Level: Warn}
	atomic := NewAtomicLevelFromConfig(config)

	// Test initial value
	if atomic.Load() != Warn {
		t.Errorf("Expected initial level Warn, got %v", atomic.Load())
	}

	// Test Store
	atomic.Store(Error)
	if atomic.Load() != Error {
		t.Errorf("Expected level Error after Store, got %v", atomic.Load())
	}

	// Test SetMin (alias for Store)
	atomic.SetMin(Debug)
	if atomic.Load() != Debug {
		t.Errorf("Expected level Debug after SetMin, got %v", atomic.Load())
	}
}

// TestConfigAtomicLevelConcurrency tests concurrent access to atomic level
func TestConfigAtomicLevelConcurrency(t *testing.T) {
	config := &Config{Level: Info}
	atomic := NewAtomicLevelFromConfig(config)

	var wg sync.WaitGroup
	levels := []Level{Debug, Info, Warn, Error}

	// Start multiple readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				_ = atomic.Load()
			}
		}()
	}

	// Start multiple writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				atomic.Store(levels[j%len(levels)])
			}
		}(i)
	}

	wg.Wait()

	// Verify final state is valid
	finalLevel := atomic.Load()
	if !IsValidLevel(finalLevel) {
		t.Errorf("Final level %v is not valid", finalLevel)
	}
}

// TestStats tests statistics functionality
func TestStats(t *testing.T) {
	config := &Config{}
	stats := config.GetStats()

	if stats == nil {
		t.Fatal("GetStats returned nil")
	}

	// Test initial values
	if stats.GetDropped() != 0 {
		t.Errorf("Expected initial dropped count 0, got %d", stats.GetDropped())
	}

	// Test increment
	stats.IncrementDropped()
	if stats.GetDropped() != 1 {
		t.Errorf("Expected dropped count 1 after increment, got %d", stats.GetDropped())
	}

	// Test multiple increments
	for i := 0; i < 10; i++ {
		stats.IncrementDropped()
	}
	if stats.GetDropped() != 11 {
		t.Errorf("Expected dropped count 11 after increments, got %d", stats.GetDropped())
	}

	// Test reset
	stats.Reset()
	if stats.GetDropped() != 0 {
		t.Errorf("Expected dropped count 0 after reset, got %d", stats.GetDropped())
	}
}

// TestStatsConcurrency tests concurrent statistics operations
func TestStatsConcurrency(t *testing.T) {
	config := &Config{}
	stats := config.GetStats()

	var wg sync.WaitGroup
	numGoroutines := 10
	incrementsPerGoroutine := 1000

	// Start concurrent incrementers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				stats.IncrementDropped()
			}
		}()
	}

	// Start concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 2000; j++ {
				_ = stats.GetDropped()
			}
		}()
	}

	wg.Wait()

	expectedCount := int64(numGoroutines * incrementsPerGoroutine)
	actualCount := stats.GetDropped()

	if actualCount != expectedCount {
		t.Errorf("Expected dropped count %d, got %d", expectedCount, actualCount)
	}
}

// TestConfigWithCustomOutput tests configuration with custom output
func TestConfigWithCustomOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		Output: WrapWriter(buf),
	}

	defaulted := config.withDefaults()

	// Write something to verify output works
	data := []byte("test output")
	n, err := defaulted.Output.Write(data)
	if err != nil {
		t.Errorf("Expected no error writing to output, got %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("Expected buffer to contain %q, got %q", data, buf.Bytes())
	}
}

// TestConfigWithFileOutput tests configuration with file output
func TestConfigWithFileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.log"

	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()

	config := &Config{
		Output: WrapWriter(file),
	}

	defaulted := config.withDefaults()

	// Test that we can sync
	err = defaulted.Output.Sync()
	if err != nil {
		t.Errorf("Expected no error syncing file output, got %v", err)
	}
}

// TestConfigPerformance tests that configuration operations are fast
func TestConfigPerformance(t *testing.T) {
	config := &Config{
		Capacity:  1024,
		BatchSize: 32,
		Level:     Info,
		Name:      "perf-test",
	}

	// Test multiple withDefaults calls (should be fast)
	for i := 0; i < 1000; i++ {
		_ = config.withDefaults()
	}

	// Test multiple validations (should be fast)
	for i := 0; i < 1000; i++ {
		_ = config.Validate()
	}

	// Test multiple clones (should be fast)
	for i := 0; i < 1000; i++ {
		_ = config.Clone()
	}

	// If we reach here, performance is acceptable for testing
}

// TestConfigEdgeCases tests edge cases and boundary conditions
func TestConfigEdgeCases(t *testing.T) {
	// Test maximum values
	config := &Config{
		Capacity:  1 << 30, // Large but valid capacity
		BatchSize: 1000,
		Level:     Error,
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Large valid values should not cause validation error: %v", err)
	}

	// Test equal capacity and batch size (valid edge case)
	config = &Config{
		Capacity:  100,
		BatchSize: 100,
		Level:     Info,
	}

	err = config.Validate()
	if err != nil {
		t.Errorf("Equal capacity and batch size should be valid: %v", err)
	}

	// Test zero batch size (should get default)
	config = &Config{
		Capacity:  1024,
		BatchSize: 0,
		Level:     Info,
	}

	defaulted := config.withDefaults()
	if defaulted.BatchSize <= 0 {
		t.Errorf("Zero batch size should get positive default, got %d", defaulted.BatchSize)
	}
}
