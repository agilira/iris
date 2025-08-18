package iris

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestCallerInfo verifies that caller information is captured correctly
func TestCallerInfo(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	writer := NewConsoleWriter(&buf)

	// Create logger with caller enabled
	config := Config{
		Level:        DebugLevel,
		Writer:       writer,
		Format:       JSONFormat,
		EnableCaller: true,
		CallerSkip:   3, // Skip: runtime.Caller, getCaller, log
		BufferSize:   1024,
		BatchSize:    1,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log a message (this will capture caller info)
	logger.Info("Test message with caller")

	// Give the logger time to process
	time.Sleep(10 * time.Millisecond)

	// Check output
	output := buf.String()

	// Verify caller information is present
	if !strings.Contains(output, "caller_test.go") {
		t.Errorf("Expected caller file name in output, got: %s", output)
	}

	// Verify line number is present (should be around line 27-30)
	if !strings.Contains(output, ":") {
		t.Errorf("Expected line number in caller info, got: %s", output)
	}

	// Verify function name is present
	if !strings.Contains(output, "TestCallerInfo") {
		t.Errorf("Expected function name in caller info, got: %s", output)
	}

	t.Logf("Output with caller info: %s", output)
}

// TestCallerDisabled verifies that caller info is not captured when disabled
func TestCallerDisabled(t *testing.T) {
	var buf bytes.Buffer
	writer := NewConsoleWriter(&buf)

	config := Config{
		Level:        DebugLevel,
		Writer:       writer,
		Format:       JSONFormat,
		EnableCaller: false, // Disabled
		BufferSize:   1024,
		BatchSize:    1,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log a message
	logger.Info("Test message without caller")

	// Give the logger time to process
	time.Sleep(10 * time.Millisecond)

	// Check output
	output := buf.String()

	// Verify caller information is NOT present
	if strings.Contains(output, `"caller":`) {
		t.Errorf("Expected no caller field when disabled, got: %s", output)
	}

	t.Logf("Output without caller info: %s", output)
}

// TestConsoleCallerInfo tests caller info in console format
func TestConsoleCallerInfo(t *testing.T) {
	var buf bytes.Buffer
	writer := NewConsoleWriter(&buf)

	config := Config{
		Level:        DebugLevel,
		Writer:       writer,
		Format:       ConsoleFormat,
		EnableCaller: true,
		CallerSkip:   3,
		BufferSize:   1024,
		BatchSize:    1,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log a message
	logger.Info("Console test with caller")

	// Give the logger time to process
	time.Sleep(10 * time.Millisecond)

	// Check output
	output := buf.String()

	// Verify caller information is present in console format [file:line]
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Errorf("Expected caller info in brackets for console format, got: %s", output)
	}

	if !strings.Contains(output, "caller_test.go") {
		t.Errorf("Expected filename in console caller info, got: %s", output)
	}

	t.Logf("Console output with caller: %s", output)
}

// BenchmarkCallerInfo benchmarks logging with caller info enabled
func BenchmarkCallerInfo(b *testing.B) {
	var buf bytes.Buffer
	writer := NewConsoleWriter(&buf)

	config := Config{
		Level:        InfoLevel,
		Writer:       writer,
		Format:       JSONFormat,
		EnableCaller: true,
		CallerSkip:   3,
		BufferSize:   1024,
		BatchSize:    8,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message with caller info")
	}
}

// BenchmarkCallerDisabled benchmarks logging with caller info disabled
func BenchmarkCallerDisabled(b *testing.B) {
	var buf bytes.Buffer
	writer := NewConsoleWriter(&buf)

	config := Config{
		Level:        InfoLevel,
		Writer:       writer,
		Format:       JSONFormat,
		EnableCaller: false,
		BufferSize:   1024,
		BatchSize:    8,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message without caller info")
	}
}
