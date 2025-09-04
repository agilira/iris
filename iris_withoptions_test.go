// iris_withoptions_test.go: Tests for Logger.WithOptions function
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

// testSyncer wraps a bytes.Buffer to implement WriteSyncer for testing
type testSyncer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (ts *testSyncer) Write(p []byte) (n int, err error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.buf.Write(p)
}

func (ts *testSyncer) Sync() error {
	return nil
}

func (ts *testSyncer) String() string {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.buf.String()
}

// Helper function to safely close logger ignoring expected errors
func safeCloseWithOptionsLogger(t *testing.T, logger *Logger) {
	if err := logger.Close(); err != nil &&
		!strings.Contains(err.Error(), "sync /dev/stdout: invalid argument") &&
		!strings.Contains(err.Error(), "ring buffer flush failed") {
		t.Errorf("Failed to close logger: %v", err)
	}
}

// TestLogger_WithOptions tests Logger.WithOptions method
func TestLogger_WithOptions(t *testing.T) {
	buf := &testSyncer{}

	baseLogger, err := New(Config{
		Level:   Debug,
		Encoder: NewTextEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create base logger: %v", err)
	}
	baseLogger.Start()
	defer safeCloseWithOptionsLogger(t, baseLogger)

	// Create logger with options
	newLogger := baseLogger.WithOptions(WithCaller())

	// Verify new logger is different instance
	if newLogger == baseLogger {
		t.Error("WithOptions should return a new logger instance")
	}

	// Test that new logger works
	newLogger.Info("Test message")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	// Verify output contains the message
	output := buf.String()
	if !strings.Contains(output, "Test message") {
		t.Errorf("Expected 'Test message' in output: %s", output)
	}
}

// TestLogger_WithOptions_NoOptions tests WithOptions with no options
func TestLogger_WithOptions_NoOptions(t *testing.T) {
	buf := &testSyncer{}

	baseLogger, err := New(Config{
		Level:   Info,
		Encoder: NewTextEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create base logger: %v", err)
	}
	baseLogger.Start()
	defer safeCloseWithOptionsLogger(t, baseLogger)

	// Create logger with no options
	newLogger := baseLogger.WithOptions()

	// Should still return different instance
	if newLogger == baseLogger {
		t.Error("WithOptions should return a new logger instance even with no options")
	}

	// Test logging works
	newLogger.Info("No options test")
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "No options test") {
		t.Errorf("Expected 'No options test' in output: %s", output)
	}
}
