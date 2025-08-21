// console_unit_test.go: Comprehensive safety net for console encoder optimizations
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strings"
	"testing"
	"time"
)

// TestConsoleEncoderBasics tests basic console encoder functionality
func TestConsoleEncoderBasics(t *testing.T) {
	encoder := NewConsoleEncoder(false) // No color for easier testing

	entry := &LogEntry{
		Timestamp: time.Date(2025, 8, 21, 15, 30, 45, 123456789, time.UTC),
		Level:     InfoLevel,
		Message:   "test message",
		Fields: []Field{
			Str("service", "test"),
			Int("count", 42),
			Bool("active", true),
		},
		Caller: Caller{Valid: true, File: "/path/to/test.go", Line: 123},
	}

	buf := make([]byte, 0, 512)
	result := encoder.EncodeLogEntry(entry, buf)
	output := string(result)

	// Verify basic structure
	if !strings.Contains(output, "2025-08-21 15:30:45.123") {
		t.Error("Timestamp not correctly formatted")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Level not correctly formatted")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Message not present")
	}
	if !strings.Contains(output, "[test.go:123]") {
		t.Error("Caller info not correctly formatted")
	}
	if !strings.Contains(output, "service=test") {
		t.Error("String field not correctly formatted")
	}
	if !strings.Contains(output, "count=42") {
		t.Error("Int field not correctly formatted")
	}
	if !strings.Contains(output, "active=true") {
		t.Error("Bool field not correctly formatted")
	}
}

// TestConsoleEncoderColors tests colored output
func TestConsoleEncoderColors(t *testing.T) {
	encoder := NewConsoleEncoder(true) // With colors

	entry := &LogEntry{
		Level:   ErrorLevel,
		Message: "error message",
	}

	buf := make([]byte, 0, 512)
	result := encoder.EncodeLogEntry(entry, buf)
	output := string(result)

	// Should contain ANSI color codes
	if !strings.Contains(output, "\033[") {
		t.Error("No ANSI color codes found in colored output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Level text not present")
	}
}

// TestConsoleEncoderAllFieldTypes tests all field types
func TestConsoleEncoderAllFieldTypes(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	// Create various field types
	fields := []Field{
		Str("string", "hello"), // Simple string without spaces
		Int("int", 123),
		Int64("int64", 9876543210),
		Float64("float64", 3.14159),
		Bool("bool_true", true),
		Bool("bool_false", false),
		Duration("duration", 5*time.Second),
		ByteString("bytes", []byte("test bytes")),
		Any("any", "any value"),
	}

	entry := &LogEntry{
		Level:   InfoLevel,
		Message: "field types test",
		Fields:  fields,
	}

	buf := make([]byte, 0, 1024)
	result := encoder.EncodeLogEntry(entry, buf)
	output := string(result)

	// Verify specific field encodings
	if !strings.Contains(output, "string=hello") {
		t.Errorf("String field not correctly encoded, got: %s", output)
	}
	if !strings.Contains(output, "int=123") {
		t.Errorf("Int field not correctly encoded, got: %s", output)
	}
	if !strings.Contains(output, "int64=9876543210") {
		t.Errorf("Int64 field not correctly encoded, got: %s", output)
	}
	if !strings.Contains(output, "float64=3.14159") {
		t.Errorf("Float64 field not correctly encoded, got: %s", output)
	}
	if !strings.Contains(output, "bool_true=true") {
		t.Errorf("Bool true field not correctly encoded, got: %s", output)
	}
	if !strings.Contains(output, "bool_false=false") {
		t.Errorf("Bool false field not correctly encoded, got: %s", output)
	}
	if !strings.Contains(output, "duration=5s") {
		t.Error("Duration field not correctly encoded")
	}
}

// TestConsoleEncoderStringQuoting tests string quoting logic
func TestConsoleEncoderStringQuoting(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"simple", "simple", "value=simple"},
		{"with_space", "hello world", `value="hello world"`},
		{"with_quotes", `hello "world"`, `value="hello \"world\""`},
		{"with_equals", "key=value", `value="key=value"`},
		{"with_newline", "hello\nworld", `value="hello`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := &LogEntry{
				Level:   InfoLevel,
				Message: "quoting test",
				Fields:  []Field{Str("value", test.value)},
			}

			buf := make([]byte, 0, 512)
			result := encoder.EncodeLogEntry(entry, buf)
			output := string(result)

			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected %s in output, got: %s", test.expected, output)
			}
		})
	}
}

// TestConsoleEncoderStackTrace tests stack trace formatting
func TestConsoleEncoderStackTrace(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	stackTrace := "goroutine 1 [running]:\nmain.main()\n\t/path/to/main.go:42 +0x123"

	entry := &LogEntry{
		Level:      ErrorLevel,
		Message:    "error with stack",
		StackTrace: stackTrace,
	}

	buf := make([]byte, 0, 1024)
	result := encoder.EncodeLogEntry(entry, buf)
	output := string(result)

	if !strings.Contains(output, "Stack trace:") {
		t.Error("Stack trace header not present")
	}
	if !strings.Contains(output, "goroutine 1") {
		t.Error("Stack trace content not present")
	}
}

// TestConsoleEncoderLevels tests all log levels
func TestConsoleEncoderLevels(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	levels := []struct {
		level Level
		text  string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
	}

	for _, test := range levels {
		t.Run(test.text, func(t *testing.T) {
			entry := &LogEntry{
				Level:   test.level,
				Message: "level test",
			}

			buf := make([]byte, 0, 512)
			result := encoder.EncodeLogEntry(entry, buf)
			output := string(result)

			if !strings.Contains(output, test.text) {
				t.Errorf("Level text %s not found in output: %s", test.text, output)
			}
		})
	}
}

// TestConsoleEncoderCallerFormatting tests caller info formatting
func TestConsoleEncoderCallerFormatting(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	tests := []struct {
		name     string
		caller   Caller
		expected string
	}{
		{
			"full_path",
			Caller{Valid: true, File: "/long/path/to/file.go", Line: 42},
			"[file.go:42]",
		},
		{
			"relative_path",
			Caller{Valid: true, File: "handlers/user.go", Line: 123},
			"[user.go:123]",
		},
		{
			"no_caller",
			Caller{Valid: false},
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := &LogEntry{
				Level:   InfoLevel,
				Message: "caller test",
				Caller:  test.caller,
			}

			buf := make([]byte, 0, 512)
			result := encoder.EncodeLogEntry(entry, buf)
			output := string(result)

			if test.expected == "" {
				// Should not contain caller info
				if strings.Contains(output, "[") && strings.Contains(output, "]") {
					t.Error("Should not contain caller info when invalid")
				}
			} else {
				if !strings.Contains(output, test.expected) {
					t.Errorf("Expected %s in output, got: %s", test.expected, output)
				}
			}
		})
	}
}

// TestConsoleEncoderBufferReuse tests buffer reuse
func TestConsoleEncoderBufferReuse(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	entry := &LogEntry{
		Level:   InfoLevel,
		Message: "buffer reuse test",
		Fields:  []Field{Str("test", "value")},
	}

	// First encode
	buf := make([]byte, 0, 512)
	result1 := encoder.EncodeLogEntry(entry, buf)

	// Second encode with same buffer
	result2 := encoder.EncodeLogEntry(entry, result1[:0])

	// Results should be identical
	if string(result1) != string(result2) {
		t.Error("Buffer reuse produces different results")
	}
}

// TestConsoleEncoderNoTimestamp tests encoding without timestamp
func TestConsoleEncoderNoTimestamp(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	entry := &LogEntry{
		Level:   InfoLevel,
		Message: "no timestamp test",
	}

	buf := make([]byte, 0, 512)
	result := encoder.EncodeLogEntry(entry, buf)
	output := string(result)

	// Should not contain timestamp format
	if strings.Contains(output, "[2") || strings.Contains(output, "2025") {
		t.Error("Should not contain timestamp when not set")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Should still contain level")
	}
}

// TestConsoleEncoderEmptyFields tests encoding with no fields
func TestConsoleEncoderEmptyFields(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	entry := &LogEntry{
		Level:   InfoLevel,
		Message: "no fields test",
		Fields:  []Field{},
	}

	buf := make([]byte, 0, 512)
	result := encoder.EncodeLogEntry(entry, buf)
	output := string(result)

	if !strings.Contains(output, "INFO") {
		t.Error("Should contain level")
	}
	if !strings.Contains(output, "no fields test") {
		t.Error("Should contain message")
	}
}

// TestConsoleEncoderErrorFields tests error field encoding
func TestConsoleEncoderErrorFields(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	// Test with actual error
	err := &CustomError{Message: "test error"}
	fields := []Field{
		Error(err),
		{Key: "nil_error", Type: ErrorType, Err: nil},
	}

	entry := &LogEntry{
		Level:   ErrorLevel,
		Message: "error field test",
		Fields:  fields,
	}

	buf := make([]byte, 0, 512)
	result := encoder.EncodeLogEntry(entry, buf)
	output := string(result)

	if !strings.Contains(output, `error="test error"`) {
		t.Errorf("Error field not correctly encoded, got: %s", output)
	}
}

// TestConsoleEncoderBinaryFields tests binary field encoding
func TestConsoleEncoderBinaryFields(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	data := []byte("hello world binary data")
	fields := []Field{
		Binary("data", data),
	}

	entry := &LogEntry{
		Level:   InfoLevel,
		Message: "binary field test",
		Fields:  fields,
	}

	buf := make([]byte, 0, 512)
	result := encoder.EncodeLogEntry(entry, buf)
	output := string(result)

	if !strings.Contains(output, "data=<binary:23 bytes>") {
		t.Errorf("Binary field not correctly encoded, got: %s", output)
	}
}
