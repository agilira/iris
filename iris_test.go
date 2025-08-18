// iris_test.go: Tests for Iris logger
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestBasicLogging(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	writer := NewConsoleWriter(&buf)

	// Create logger
	logger, err := New(Config{
		Level:      DebugLevel,
		Writer:     writer,
		BufferSize: 1024,
		BatchSize:  32,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test basic logging
	logger.Info("Test message")
	logger.Debug("Debug message", Str("key", "value"))
	logger.Warn("Warning message", Int("number", 42))
	logger.Error("Error message", Err(err))

	// Force flush and give time for processing
	logger.ring.Flush()
	time.Sleep(50 * time.Millisecond)

	// Check output
	output := buf.String()
	if !strings.Contains(output, "Test message") {
		t.Error("Output should contain 'Test message'")
	}
	if !strings.Contains(output, "Debug message") {
		t.Error("Output should contain 'Debug message'")
	}
	if !strings.Contains(output, "Warning message") {
		t.Error("Output should contain 'Warning message'")
	}

	t.Logf("Output:\n%s", output)
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	writer := NewConsoleWriter(&buf)

	// Create logger with Info level (should filter out Debug)
	logger, err := New(Config{
		Level:      InfoLevel,
		Writer:     writer,
		BufferSize: 512,
		BatchSize:  16,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Debug("This should not appear")
	logger.Info("This should appear")
	logger.Warn("This should also appear")

	// Force flush and give time for processing
	logger.ring.Flush()
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if strings.Contains(output, "This should not appear") {
		t.Error("Debug message should be filtered out")
	}
	if !strings.Contains(output, "This should appear") {
		t.Error("Info message should appear")
	}
	if !strings.Contains(output, "This should also appear") {
		t.Error("Warn message should appear")
	}
}

func TestStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	writer := NewConsoleWriter(&buf)

	logger, err := New(Config{
		Level:  InfoLevel,
		Writer: writer,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test various field types
	logger.Info("Structured log",
		Str("string_field", "test_value"),
		Int("int_field", 123),
		Float("float_field", 3.14),
		Bool("bool_field", true),
		Duration("duration_field", 100*time.Millisecond),
	)

	// Force flush and give time for processing
	logger.ring.Flush()
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "string_field") {
		t.Error("Output should contain string_field")
	}
	if !strings.Contains(output, "test_value") {
		t.Error("Output should contain test_value")
	}
	if !strings.Contains(output, "int_field") {
		t.Error("Output should contain int_field")
	}
	if !strings.Contains(output, "123") {
		t.Error("Output should contain 123")
	}

	t.Logf("Structured output:\n%s", output)
}

func BenchmarkIrisLogging(b *testing.B) {
	var buf bytes.Buffer
	writer := NewConsoleWriter(&buf)

	logger, err := New(Config{
		Level:      InfoLevel,
		Writer:     writer,
		BufferSize: 4096,
		BatchSize:  128,
	})
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message", Int("iteration", i))
	}
}
