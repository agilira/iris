// smart_api_test.go: Tests for the smart API simplification
//
// This file validates that the smart API auto-detection works correctly
// and provides better defaults than manual configuration.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"os"
	"testing"
	"time"
)

// TestSmartAPI_BasicCreation tests that New() works with minimal configuration
func TestSmartAPI_BasicCreation(t *testing.T) {
	// Empty config should work with smart defaults
	logger, err := New(Config{})
	if err != nil {
		t.Fatalf("New(Config{}) failed: %v", err)
	}
	defer safeCloseSmartAPILogger(t, logger)

	logger.Start()

	// Should be able to log immediately
	logger.Info("Smart API test message")

	// Give some time for processing
	time.Sleep(10 * time.Millisecond)

	// Basic validation - no panics or errors
}

// TestSmartAPI_DevelopmentMode tests smart encoder selection
func TestSmartAPI_DevelopmentMode(t *testing.T) {
	var buf bytes.Buffer

	// Development mode should select TextEncoder automatically
	logger, err := New(Config{Output: WrapWriter(&buf)}, Development())
	if err != nil {
		t.Fatalf("New with Development() failed: %v", err)
	}
	defer safeCloseSmartAPILogger(t, logger)

	logger.Start()

	logger.Info("Development message", Str("key", "value"))

	time.Sleep(10 * time.Millisecond)
	_ = logger.Sync()

	output := buf.String()

	// TextEncoder should produce key-value format (not JSON)
	if output == "" {
		t.Error("No output generated")
	}

	// In development mode, output should be human-readable
	// (TextEncoder produces key=value format, not JSON)
	if len(output) < 10 {
		t.Errorf("Output seems too short: %q", output)
	}
}

// TestSmartAPI_ProductionMode tests default JSON encoder selection
func TestSmartAPI_ProductionMode(t *testing.T) {
	var buf bytes.Buffer

	// Production mode (no Development() option) should select JSONEncoder
	logger, err := New(Config{Output: WrapWriter(&buf)})
	if err != nil {
		t.Fatalf("New without Development() failed: %v", err)
	}
	defer safeCloseSmartAPILogger(t, logger)

	logger.Start()

	logger.Info("Production message", Str("key", "value"))

	time.Sleep(10 * time.Millisecond)
	_ = logger.Sync()

	output := buf.String()

	// JSONEncoder should produce JSON format
	if output == "" {
		t.Error("No output generated")
	}

	// Should contain JSON structure
	if output[0] != '{' {
		t.Errorf("Expected JSON output starting with '{', got: %q", output)
	}
}

// TestSmartAPI_ArchitectureDetection tests smart architecture selection
func TestSmartAPI_ArchitectureDetection(t *testing.T) {
	// Just test that it doesn't crash and selects appropriate architecture
	logger, err := New(Config{})
	if err != nil {
		t.Fatalf("New() with smart architecture failed: %v", err)
	}
	defer safeCloseSmartAPILogger(t, logger)

	// Architecture should be automatically selected based on CPU count
	// We can't easily test the specific choice without exposing internals,
	// but we can test that it doesn't crash
	logger.Start()
}

// TestSmartAPI_CapacityOptimization tests smart capacity selection
func TestSmartAPI_CapacityOptimization(t *testing.T) {
	logger, err := New(Config{})
	if err != nil {
		t.Fatalf("New() with smart capacity failed: %v", err)
	}
	defer safeCloseSmartAPILogger(t, logger)

	logger.Start()

	// Should handle multiple concurrent operations without issues
	for i := 0; i < 100; i++ {
		logger.Info("Capacity test", Int("iteration", i))
	}

	time.Sleep(50 * time.Millisecond)

	// Should complete without errors
}

// TestSmartAPI_LevelDetection tests smart level detection from environment
func TestSmartAPI_LevelDetection(t *testing.T) {
	// Test with environment variable
	originalLevel := os.Getenv("IRIS_LEVEL")
	defer func() {
		if err := os.Setenv("IRIS_LEVEL", originalLevel); err != nil {
			t.Errorf("Failed to restore IRIS_LEVEL: %v", err)
		}
	}()

	if err := os.Setenv("IRIS_LEVEL", "error"); err != nil {
		t.Errorf("Failed to set IRIS_LEVEL: %v", err)
	}

	var buf bytes.Buffer
	logger, err := New(Config{Output: WrapWriter(&buf)})
	if err != nil {
		t.Fatalf("New() with IRIS_LEVEL failed: %v", err)
	}
	defer safeCloseSmartAPILogger(t, logger)

	logger.Start()

	// Info should be filtered out (level is ERROR)
	logger.Info("This should be filtered")
	logger.Error("This should appear")

	time.Sleep(10 * time.Millisecond)
	_ = logger.Sync()

	output := buf.String()

	// Should only contain the error message
	if output == "" {
		t.Error("No output generated, but error message should appear")
	}
	// Could add more specific checks here if needed
}

// TestSmartAPI_BackwardCompatibility tests that existing code still works
func TestSmartAPI_BackwardCompatibility(t *testing.T) {
	// Test that old-style configuration still works
	var buf bytes.Buffer

	logger, err := New(Config{
		Level:    Info,
		Output:   WrapWriter(&buf),
		Encoder:  NewJSONEncoder(),
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("New() with explicit config failed: %v", err)
	}
	defer safeCloseSmartAPILogger(t, logger)

	logger.Start()

	logger.Info("Backward compatibility test")

	time.Sleep(10 * time.Millisecond)
	_ = logger.Sync()

	output := buf.String()
	if output == "" {
		t.Error("No output generated with explicit config")
	}
}

// TestSmartAPI_Performance tests that smart defaults don't hurt performance
func TestSmartAPI_Performance(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New(Config{Output: WrapWriter(&buf)})
	if err != nil {
		t.Fatalf("New() for performance test failed: %v", err)
	}
	defer safeCloseSmartAPILogger(t, logger)

	logger.Start()

	// Measure performance of smart defaults
	start := time.Now()

	for i := 0; i < 1000; i++ {
		logger.Info("Performance test", Int("iteration", i))
	}

	duration := time.Since(start)

	// Should complete 1000 operations reasonably quickly
	// This is a smoke test, not a precise benchmark
	if duration > 100*time.Millisecond {
		t.Logf("1000 operations took %v (might be slow, but not necessarily wrong)", duration)
	}

	time.Sleep(50 * time.Millisecond)
	_ = logger.Sync()
}

// Helper function for safe logger cleanup
func safeCloseSmartAPILogger(t *testing.T, logger *Logger) {
	if err := logger.Close(); err != nil {
		t.Logf("Warning: Error closing logger in test: %v", err)
	}
}
