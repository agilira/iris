// iris_logging_methods_test.go: Tests for additional logging methods
//
// This file provides test coverage for logging methods that were missing
// coverage: InfoFields, Write, Stats, and formatted logging methods.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// logTestSyncer implements WriteSyncer for testing logging methods
type logTestSyncer struct {
	buf strings.Builder
	mu  sync.Mutex
}

func (lts *logTestSyncer) Write(p []byte) (n int, err error) {
	lts.mu.Lock()
	defer lts.mu.Unlock()
	return lts.buf.Write(p)
}

func (lts *logTestSyncer) Sync() error {
	return nil
}

func (lts *logTestSyncer) String() string {
	lts.mu.Lock()
	defer lts.mu.Unlock()
	return lts.buf.String()
}

// TestLogger_InfoFields tests InfoFields method
func TestLogger_InfoFields(t *testing.T) {
	buf := &logTestSyncer{}

	logger, err := New(Config{
		Level:   Info,
		Encoder: NewTextEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer safeCloseLoggingMethodsLogger(t, logger)

	// Test InfoFields with various field types
	result := logger.InfoFields("Info message with fields",
		Str("key1", "value1"),
		Int("key2", 42),
		Bool("key3", true))

	// Should return true when message is logged
	if !result {
		t.Error("InfoFields should return true when message is logged")
	}

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "Info message with fields") {
		t.Errorf("Expected message in output: %s", output)
	}
	if !strings.Contains(output, "key1") || !strings.Contains(output, "value1") {
		t.Errorf("Expected field key1=value1 in output: %s", output)
	}
	if !strings.Contains(output, "key2") || !strings.Contains(output, "42") {
		t.Errorf("Expected field key2=42 in output: %s", output)
	}
}

// TestLogger_InfoFields_LevelFiltering tests InfoFields with level filtering
func TestLogger_InfoFields_LevelFiltering(t *testing.T) {
	buf := &logTestSyncer{}

	// Create logger with Warn level (Info should be filtered out)
	logger, err := New(Config{
		Level:   Warn,
		Encoder: NewTextEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer safeCloseLoggingMethodsLogger(t, logger)

	// InfoFields returns true even when level is too low (early exit optimization)
	result := logger.InfoFields("This should be filtered out",
		Str("test", "value"))

	if !result {
		t.Error("InfoFields should return true even when message is filtered (early exit)")
	}

	time.Sleep(50 * time.Millisecond)

	// However, the message should not appear in output
	output := buf.String()
	if strings.Contains(output, "This should be filtered out") {
		t.Errorf("Message should be filtered out at Warn level: %s", output)
	}
}

// TestLogger_Write tests Write method (record filling interface)
func TestLogger_Write(t *testing.T) {
	buf := &logTestSyncer{}

	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewTextEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer safeCloseLoggingMethodsLogger(t, logger)

	// Test Write method with record filling function
	result := logger.Write(func(record *Record) {
		record.Msg = "Direct write test message"
		record.Level = Info
		record.AddField(Str("component", "test"))
		record.AddField(Int("count", 123))
	})

	if !result {
		t.Error("Write should return true when record is written")
	}

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "Direct write test message") {
		t.Errorf("Expected write message in output: %s", output)
	}
	if !strings.Contains(output, "component") || !strings.Contains(output, "test") {
		t.Errorf("Expected component field in output: %s", output)
	}
	if !strings.Contains(output, "count") || !strings.Contains(output, "123") {
		t.Errorf("Expected count field in output: %s", output)
	}
}

// TestLogger_Stats tests Stats method
func TestLogger_Stats(t *testing.T) {
	buf := &logTestSyncer{}

	logger, err := New(Config{
		Level:   Info,
		Encoder: NewTextEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer safeCloseLoggingMethodsLogger(t, logger)

	// Log some messages to generate stats
	logger.Info("Test message 1")
	logger.Info("Test message 2")
	logger.Warn("Warning message")

	time.Sleep(50 * time.Millisecond)

	// Get stats
	stats := logger.Stats()

	// Stats should not be nil
	if stats == nil {
		t.Error("Stats should not return nil")
	}

	// Should have some basic structure - exact validation depends on Stats implementation
	// Just verify we can call it without panic
	t.Logf("Stats returned: %+v", stats)
}

// TestLogger_FormattedMethods tests formatted logging methods
func TestLogger_FormattedMethods(t *testing.T) {
	buf := &logTestSyncer{}

	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewTextEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer safeCloseLoggingMethodsLogger(t, logger)

	tests := []struct {
		name     string
		logFunc  func()
		expected []string // Multiple possible patterns
	}{
		{
			name:     "Debugf",
			logFunc:  func() { logger.Debugf("Debug: %s = %d", "count", 10) },
			expected: []string{"Debug", "count", "10"},
		},
		{
			name:     "Infof",
			logFunc:  func() { logger.Infof("Info: %s at %v", "status", "ready") },
			expected: []string{"Info", "status", "ready"},
		},
		{
			name:     "Warnf",
			logFunc:  func() { logger.Warnf("Warning: %d errors found", 3) },
			expected: []string{"Warning", "3", "errors"},
		},
		{
			name:     "Errorf",
			logFunc:  func() { logger.Errorf("Error: %s failed with code %d", "operation", 500) },
			expected: []string{"Error", "operation", "500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear buffer
			buf.mu.Lock()
			buf.buf.Reset()
			buf.mu.Unlock()

			// Call formatted logging method
			tt.logFunc()

			time.Sleep(50 * time.Millisecond)

			output := buf.String()
			// Check that all expected parts are present in output
			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected %q in output: %s", exp, output)
				}
			}
		})
	}
}

// TestLogger_FormattedMethods_LevelFiltering tests formatted methods with level filtering
func TestLogger_FormattedMethods_LevelFiltering(t *testing.T) {
	buf := &logTestSyncer{}

	// Create logger with Error level (Debug, Info, Warn should be filtered)
	logger, err := New(Config{
		Level:   Error,
		Encoder: NewTextEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer safeCloseLoggingMethodsLogger(t, logger)

	// These should be filtered out
	logger.Debugf("Debug message %d", 1)
	logger.Infof("Info message %d", 2)
	logger.Warnf("Warn message %d", 3)

	// This should appear
	logger.Errorf("Error message %d", 4)

	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Should not contain filtered messages
	if strings.Contains(output, "Debug message 1") {
		t.Error("Debug message should be filtered at Error level")
	}
	if strings.Contains(output, "Info message 2") {
		t.Error("Info message should be filtered at Error level")
	}
	if strings.Contains(output, "Warn message 3") {
		t.Error("Warn message should be filtered at Error level")
	}

	// Should contain error message
	if !strings.Contains(output, "Error message 4") {
		t.Error("Error message should appear at Error level")
	}
}

// Helper function for safe logger cleanup
func safeCloseLoggingMethodsLogger(t *testing.T, logger *Logger) {
	if err := logger.Close(); err != nil {
		t.Logf("Warning: Error closing logger in test: %v", err)
	}
}
