// encoder-cnsl_unit_test.go: Comprehensive tests for console encoder
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestNewConsoleEncoder tests the default console encoder creation
func TestNewConsoleEncoder(t *testing.T) {
	encoder := NewConsoleEncoder()

	if encoder == nil {
		t.Fatal("NewConsoleEncoder should not return nil")
	}

	// Check default values
	if encoder.TimeFormat != time.RFC3339Nano {
		t.Errorf("Expected TimeFormat to be RFC3339Nano, got %s", encoder.TimeFormat)
	}

	if encoder.LevelCasing != "upper" {
		t.Errorf("Expected LevelCasing to be 'upper', got %s", encoder.LevelCasing)
	}

	if encoder.EnableColor {
		t.Error("Expected EnableColor to be false by default")
	}
}

// TestNewColorConsoleEncoder tests the color console encoder creation
func TestNewColorConsoleEncoder(t *testing.T) {
	encoder := NewColorConsoleEncoder()

	if encoder == nil {
		t.Fatal("NewColorConsoleEncoder should not return nil")
	}

	// Check values for color encoder
	if encoder.TimeFormat != time.RFC3339Nano {
		t.Errorf("Expected TimeFormat to be RFC3339Nano, got %s", encoder.TimeFormat)
	}

	if encoder.LevelCasing != "upper" {
		t.Errorf("Expected LevelCasing to be 'upper', got %s", encoder.LevelCasing)
	}

	if !encoder.EnableColor {
		t.Error("Expected EnableColor to be true for color encoder")
	}
}

// TestConsoleEncoder_Encode tests the main Encode method with various scenarios
func TestConsoleEncoder_Encode(t *testing.T) {
	testTime := time.Date(2025, 8, 23, 17, 30, 0, 123456789, time.UTC)

	tests := []struct {
		name     string
		encoder  *ConsoleEncoder
		level    Level
		msg      string
		fields   []Field
		expected string
		contains []string // Patterns that should be present
	}{
		{
			name:    "basic_info_message",
			encoder: NewConsoleEncoder(),
			level:   Info,
			msg:     "test message",
			fields:  nil,
			contains: []string{
				"2025-08-23T17:30:00.123456789Z",
				"INFO",
				"test message",
			},
		},
		{
			name:    "message_with_fields",
			encoder: NewConsoleEncoder(),
			level:   Warn,
			msg:     "warning occurred",
			fields: []Field{
				Str("user", "john"),
				Int("count", 42),
			},
			contains: []string{
				"WARN",
				"warning occurred",
				"user=john",
				"count=42",
			},
		},
		{
			name:    "empty_message_with_fields",
			encoder: NewConsoleEncoder(),
			level:   Error,
			msg:     "",
			fields: []Field{
				Str("error", "database connection failed"),
				Bool("retryable", true),
			},
			contains: []string{
				"ERROR",
				"error=",
				"retryable=true",
			},
		},
		{
			name: "custom_time_format",
			encoder: &ConsoleEncoder{
				TimeFormat:  "2006-01-02 15:04:05",
				LevelCasing: "upper",
				EnableColor: false,
			},
			level:  Debug,
			msg:    "debug info",
			fields: nil,
			contains: []string{
				"2025-08-23 17:30:00",
				"DEBUG",
				"debug info",
			},
		},
		{
			name: "lowercase_levels",
			encoder: &ConsoleEncoder{
				TimeFormat:  time.RFC3339,
				LevelCasing: "lower",
				EnableColor: false,
			},
			level:  Info,
			msg:    "info message",
			fields: nil,
			contains: []string{
				"info", // Should be lowercase
				"info message",
			},
		},
		{
			name:    "with_colors_enabled",
			encoder: NewColorConsoleEncoder(),
			level:   Error,
			msg:     "error message",
			fields:  nil,
			contains: []string{
				"\x1b[31m", // Red color for error
				"ERROR",
				"\x1b[0m", // Reset color
				"error message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create record
			rec := &Record{
				Level: tt.level,
				Msg:   tt.msg,
				n:     int32(len(tt.fields)),
			}

			// Add fields to record
			for i, field := range tt.fields {
				if i < len(rec.fields) {
					rec.fields[i] = field
				}
			}

			// Encode
			var buf bytes.Buffer
			tt.encoder.Encode(rec, testTime, &buf)
			output := buf.String()

			// Check expected patterns
			for _, pattern := range tt.contains {
				if !strings.Contains(output, pattern) {
					t.Errorf("Expected output to contain '%s', got: %s", pattern, output)
				}
			}

			// Should end with newline
			if !strings.HasSuffix(output, "\n") {
				t.Error("Output should end with newline")
			}
		})
	}
}

// TestConsoleEncoder_EncodeConsoleValue tests field value encoding
func TestConsoleEncoder_EncodeConsoleValue(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected string
	}{
		{
			name:     "string_value",
			field:    Str("key", "simple"),
			expected: "simple",
		},
		{
			name:     "string_with_spaces",
			field:    Str("key", "hello world"),
			expected: `"hello world"`,
		},
		{
			name:     "string_with_quotes",
			field:    Str("key", `say "hello"`),
			expected: `"say \"hello\""`,
		},
		{
			name:     "empty_string",
			field:    Str("key", ""),
			expected: `""`,
		},
		{
			name:     "integer_value",
			field:    Int("key", 42),
			expected: "42",
		},
		{
			name:     "negative_integer",
			field:    Int("key", -123),
			expected: "-123",
		},
		{
			name:     "uint64_value",
			field:    Uint64("key", 12345),
			expected: "12345",
		},
		{
			name:     "float64_value",
			field:    Float64("key", 3.14159),
			expected: "3.14159",
		},
		{
			name:     "bool_true",
			field:    Bool("key", true),
			expected: "true",
		},
		{
			name:     "bool_false",
			field:    Bool("key", false),
			expected: "false",
		},
		{
			name:     "duration_value",
			field:    Dur("key", 5*time.Second),
			expected: "5s",
		},
		{
			name:     "duration_complex",
			field:    Dur("key", 1*time.Hour+30*time.Minute+45*time.Second),
			expected: "1h30m45s",
		},
		{
			name:     "time_value",
			field:    Time("key", time.Date(2025, 8, 23, 17, 30, 0, 0, time.UTC)),
			expected: "", // Will be checked separately for valid RFC3339 format
		},
		{
			name:     "bytes_value",
			field:    Bytes("key", []byte("hello")),
			expected: "<5B>",
		},
		{
			name:     "empty_bytes",
			field:    Bytes("key", []byte{}),
			expected: "<0B>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encodeConsoleValue(tt.field, &buf)
			result := buf.String()

			// Special handling for time_value test
			if tt.name == "time_value" {
				// Should be a valid RFC3339 timestamp
				_, err := time.Parse(time.RFC3339, result)
				if err != nil {
					// Try RFC3339Nano
					_, err = time.Parse(time.RFC3339Nano, result)
					if err != nil {
						t.Errorf("Expected valid RFC3339/RFC3339Nano timestamp, got '%s'", result)
					}
				}
				// Should contain 2025-08-23 and 17:30:00
				if !strings.Contains(result, "2025-08-23") {
					t.Errorf("Expected timestamp to contain date '2025-08-23', got '%s'", result)
				}
				return
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestConsoleEncoder_WriteMaybeQuoted tests string quoting logic
func TestConsoleEncoder_WriteMaybeQuoted(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple_string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string_with_space",
			input:    "hello world",
			expected: `"hello world"`,
		},
		{
			name:     "string_with_tab",
			input:    "hello\tworld",
			expected: `"hello\tworld"`,
		},
		{
			name:     "string_with_quotes",
			input:    `say "hello"`,
			expected: `"say \"hello\""`,
		},
		{
			name:     "string_with_backslash",
			input:    `path\to\file`,
			expected: `"path\\to\\file"`,
		},
		{
			name:     "string_with_newline",
			input:    "line1\nline2",
			expected: `"line1\nline2"`,
		},
		{
			name:     "string_with_carriage_return",
			input:    "line1\rline2",
			expected: `"line1\rline2"`,
		},
		{
			name:     "empty_string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "single_character",
			input:    "a",
			expected: "a",
		},
		{
			name:     "special_characters_no_quotes_needed",
			input:    "hello@#$%^&*()_+-={}[]|;':,.<>?/~`",
			expected: "hello@#$%^&*()_+-={}[]|;':,.<>?/~`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writeMaybeQuoted(tt.input, &buf)
			result := buf.String()

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestConsoleEncoder_ColorizeLevel tests ANSI color application
func TestConsoleEncoder_ColorizeLevel(t *testing.T) {
	tests := []struct {
		name               string
		level              Level
		levelStr           string
		shouldContainColor bool
		expectedColor      string
	}{
		{
			name:               "debug_level",
			level:              Debug,
			levelStr:           "DEBUG",
			shouldContainColor: true,
			expectedColor:      "\x1b[90m", // Gray
		},
		{
			name:               "info_level",
			level:              Info,
			levelStr:           "INFO",
			shouldContainColor: true,
			expectedColor:      "\x1b[34m", // Blue
		},
		{
			name:               "warn_level",
			level:              Warn,
			levelStr:           "WARN",
			shouldContainColor: true,
			expectedColor:      "\x1b[33m", // Yellow
		},
		{
			name:               "error_level",
			level:              Error,
			levelStr:           "ERROR",
			shouldContainColor: true,
			expectedColor:      "\x1b[31m", // Red
		},
		{
			name:               "dpanic_level",
			level:              DPanic,
			levelStr:           "DPANIC",
			shouldContainColor: true,
			expectedColor:      "\x1b[35m", // Magenta
		},
		{
			name:               "panic_level",
			level:              Panic,
			levelStr:           "PANIC",
			shouldContainColor: true,
			expectedColor:      "\x1b[91m", // Bright Red
		},
		{
			name:               "fatal_level",
			level:              Fatal,
			levelStr:           "FATAL",
			shouldContainColor: true,
			expectedColor:      "\x1b[91m", // Bright Red
		},
		{
			name:               "unknown_level",
			level:              Level(99), // Use a clearly invalid level
			levelStr:           "UNKNOWN",
			shouldContainColor: false,
			expectedColor:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeLevel(tt.level, tt.levelStr)

			if tt.shouldContainColor {
				// Should contain the expected color code
				if !strings.Contains(result, tt.expectedColor) {
					t.Errorf("Expected result to contain color code '%s', got: %s", tt.expectedColor, result)
				}

				// Should contain the reset code
				if !strings.Contains(result, "\x1b[0m") {
					t.Error("Expected result to contain reset code")
				}

				// Should contain the original level string
				if !strings.Contains(result, tt.levelStr) {
					t.Errorf("Expected result to contain level string '%s', got: %s", tt.levelStr, result)
				}
			} else {
				// Should be unchanged for unknown levels
				if result != tt.levelStr {
					t.Errorf("Expected result to be unchanged '%s' (len=%d), got '%s' (len=%d)",
						tt.levelStr, len(tt.levelStr), result, len(result))
				}
			}
		})
	}
}

// TestConsoleEncoder_IntegrationWithAllFieldTypes tests encoding with all field types
func TestConsoleEncoder_IntegrationWithAllFieldTypes(t *testing.T) {
	encoder := NewConsoleEncoder()
	testTime := time.Date(2025, 8, 23, 17, 30, 0, 0, time.UTC)

	// Create record with various field types
	rec := &Record{
		Level: Info,
		Msg:   "test with all field types",
		n:     15,
	}

	// Add various field types
	rec.fields[0] = Str("string", "hello world")
	rec.fields[1] = Int("int", 42)
	rec.fields[2] = Int64("int64", 9223372036854775807)
	rec.fields[3] = Uint64("uint64", 18446744073709551615)
	rec.fields[4] = Float64("float", 3.14159)
	rec.fields[5] = Bool("bool_true", true)
	rec.fields[6] = Bool("bool_false", false)
	rec.fields[7] = Dur("duration", 5*time.Minute)
	rec.fields[8] = Time("time", testTime)
	rec.fields[9] = Bytes("bytes", []byte("test"))
	rec.fields[10] = Str("empty", "")
	rec.fields[11] = Str("quoted", `contains "quotes"`)
	rec.fields[12] = Int("negative", -123)
	rec.fields[13] = Float32("float32", 2.71)
	rec.fields[14] = Str("spaces", "has spaces")

	var buf bytes.Buffer
	encoder.Encode(rec, testTime, &buf)
	output := buf.String()

	// Verify various components are present
	expectedPatterns := []string{
		"2025-08-23T17:30:00Z",
		"INFO",
		"test with all field types",
		"string=\"hello world\"",
		"int=42",
		"int64=9223372036854775807",
		"uint64=18446744073709551615",
		"float=3.14159",
		"bool_true=true",
		"bool_false=false",
		"duration=5m0s",
		"bytes=<4B>",
		"empty=\"\"",
		"quoted=\"contains \\\"quotes\\\"\"",
		"negative=-123",
		"float32=2.71",
		"spaces=\"has spaces\"",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Expected output to contain '%s', got: %s", pattern, output)
		}
	}

	// Special check for time field (timezone may vary)
	if !strings.Contains(output, "time=2025-08-23T") {
		t.Errorf("Expected output to contain time field with correct date, got: %s", output)
	}
}

// TestConsoleEncoder_EmptyTimeFormat tests behavior with empty time format
func TestConsoleEncoder_EmptyTimeFormat(t *testing.T) {
	encoder := &ConsoleEncoder{
		TimeFormat:  "", // Empty time format should use default
		LevelCasing: "upper",
		EnableColor: false,
	}

	testTime := time.Date(2025, 8, 23, 17, 30, 0, 123456789, time.UTC)
	rec := &Record{
		Level: Info,
		Msg:   "test message",
		n:     0,
	}

	var buf bytes.Buffer
	encoder.Encode(rec, testTime, &buf)
	output := buf.String()

	// Should use RFC3339Nano as default
	if !strings.Contains(output, "2025-08-23T17:30:00.123456789Z") {
		t.Errorf("Expected default RFC3339Nano format, got: %s", output)
	}
}

// TestConsoleEncoder_EmptyLevelCasing tests behavior with empty level casing
func TestConsoleEncoder_EmptyLevelCasing(t *testing.T) {
	encoder := &ConsoleEncoder{
		TimeFormat:  time.RFC3339,
		LevelCasing: "", // Empty should default to upper
		EnableColor: false,
	}

	testTime := time.Now()
	rec := &Record{
		Level: Info,
		Msg:   "test message",
		n:     0,
	}

	var buf bytes.Buffer
	encoder.Encode(rec, testTime, &buf)
	output := buf.String()

	// Should use uppercase as default
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected uppercase level, got: %s", output)
	}
}

// Benchmark tests for performance verification

// BenchmarkConsoleEncoder_Encode benchmarks basic encoding performance
func BenchmarkConsoleEncoder_Encode(b *testing.B) {
	encoder := NewConsoleEncoder()
	testTime := time.Now()
	rec := &Record{
		Level: Info,
		Msg:   "benchmark test message",
		n:     3,
	}
	rec.fields[0] = Str("user", "benchmarkuser")
	rec.fields[1] = Int("id", 12345)
	rec.fields[2] = Bool("active", true)

	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		encoder.Encode(rec, testTime, &buf)
	}
}

// BenchmarkConsoleEncoder_EncodeWithColors benchmarks encoding with colors
func BenchmarkConsoleEncoder_EncodeWithColors(b *testing.B) {
	encoder := NewColorConsoleEncoder()
	testTime := time.Now()
	rec := &Record{
		Level: Warn,
		Msg:   "benchmark warning message",
		n:     2,
	}
	rec.fields[0] = Str("component", "auth")
	rec.fields[1] = Dur("elapsed", 250*time.Millisecond)

	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		encoder.Encode(rec, testTime, &buf)
	}
}
