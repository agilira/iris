// console_test.go: Test the console encoder
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strings"
	"testing"
	"time"
)

func TestConsoleEncoder(t *testing.T) {
	encoder := NewConsoleEncoder(false) // No colors for testing

	entry := &LogEntry{
		Level:     InfoLevel,
		Message:   "test message",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Fields: []Field{
			String("user", "john"),
			Int("age", 30),
		},
	}

	var buf []byte
	result := encoder.EncodeLogEntry(entry, buf)

	expected := "[2025-01-01 12:00:00.000] INFO  test message user=john age=30\n"
	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestConsoleEncoderWithColors(t *testing.T) {
	encoder := NewConsoleEncoder(true) // With colors

	entry := &LogEntry{
		Level:     ErrorLevel,
		Message:   "error message",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Fields: []Field{
			String("component", "auth"),
		},
	}

	var buf []byte
	result := encoder.EncodeLogEntry(entry, buf)

	// Should contain color codes
	resultStr := string(result)
	if !strings.Contains(resultStr, Red) {
		t.Error("Expected red color code for error level")
	}
	if !strings.Contains(resultStr, Bold) {
		t.Error("Expected bold text for error message")
	}
	if !strings.Contains(resultStr, Cyan) {
		t.Error("Expected cyan color for field keys")
	}
}

func TestConsoleEncoderFieldTypes(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	entry := &LogEntry{
		Level:     InfoLevel,
		Message:   "test",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Fields: []Field{
			String("str", "value"),
			Int("int", 42),
			Float("float", 3.14),
			Bool("bool", true),
			Duration("dur", time.Second),
		},
	}

	var buf []byte
	result := string(encoder.EncodeLogEntry(entry, buf))

	tests := []string{
		"str=value",
		"int=42",
		"float=3.14",
		"bool=true",
		"dur=1s",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Expected %q in output: %q", test, result)
		}
	}
}

func TestConsoleEncoderQuoting(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	entry := &LogEntry{
		Level:     InfoLevel,
		Message:   "test",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Fields: []Field{
			String("normal", "value"),
			String("spaced", "value with spaces"),
			String("quoted", `value"with"quotes`),
		},
	}

	var buf []byte
	result := string(encoder.EncodeLogEntry(entry, buf))

	// Check quoting behavior
	if !strings.Contains(result, "normal=value") {
		t.Error("Normal values should not be quoted")
	}
	if !strings.Contains(result, `spaced="value with spaces"`) {
		t.Error("Values with spaces should be quoted")
	}
	if !strings.Contains(result, `quoted="value"with"quotes"`) {
		t.Error("Values with quotes should be quoted")
	}
}

func BenchmarkConsoleEncoder(b *testing.B) {
	encoder := NewConsoleEncoder(false)
	entry := &LogEntry{
		Level:     InfoLevel,
		Message:   "benchmark message",
		Timestamp: time.Now(),
		Fields: []Field{
			String("component", "benchmark"),
			Int("iteration", 100),
			Float("duration", 1.234),
		},
	}

	buf := make([]byte, 0, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.EncodeLogEntry(entry, buf)
	}
}
