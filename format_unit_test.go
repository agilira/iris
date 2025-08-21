package iris

import (
	"strings"
	"testing"
	"time"
)

// TestFastTextEncoder tests the FastTextEncoder
func TestFastTextEncoder(t *testing.T) {
	encoder := NewFastTextEncoder()
	timestamp := time.Date(2025, 8, 21, 10, 30, 45, 123456789, time.UTC)

	encoder.EncodeLogEntryMigration(
		timestamp,
		InfoLevel,
		"test message",
		[]Field{
			Str("service", "test"),
			Int("count", 42),
			Bool("active", true),
		},
		Caller{Valid: true, File: "/path/to/file.go", Line: 123},
		"",
	)

	output := string(encoder.Bytes())

	// Verify basic structure
	if !strings.Contains(output, "10:30:45.123") {
		t.Error("Expected timestamp in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Expected level in output")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Expected message in output")
	}
	if !strings.Contains(output, "service=test") {
		t.Error("Expected service field")
	}
	if !strings.Contains(output, "count=42") {
		t.Error("Expected count field")
	}
	if !strings.Contains(output, "active=true") {
		t.Error("Expected active field")
	}
	if !strings.Contains(output, "[file.go:123]") {
		t.Error("Expected caller info")
	}
}

// TestFastTextEncoderLevels tests all log levels
func TestFastTextEncoderLevels(t *testing.T) {
	encoder := NewFastTextEncoder()
	levels := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
	}

	for _, tc := range levels {
		t.Run(tc.expected, func(t *testing.T) {
			encoder.EncodeLogEntry(
				time.Time{},
				tc.level,
				"test",
				nil,
				Caller{},
				"",
			)

			output := string(encoder.Bytes())
			if !strings.Contains(output, tc.expected) {
				t.Errorf("Expected %s in output, got: %s", tc.expected, output)
			}
		})
	}
}

// TestFastTextEncoderFieldTypes tests all field types
func TestFastTextEncoderFieldTypes(t *testing.T) {
	encoder := NewFastTextEncoder()

	fields := []Field{
		Str("string_field", "hello"),
		Int("int_field", 42),
		Int64("int64_field", 9223372036854775807),
		Float64("float_field", 3.14159),
		Bool("bool_true", true),
		Bool("bool_false", false),
		Duration("duration_field", 5*time.Second),
		Time("time_field", time.Unix(1627890123, 0)),
	}

	encoder.EncodeLogEntryMigration(
		time.Time{},
		InfoLevel,
		"field test",
		fields,
		Caller{},
		"",
	)

	output := string(encoder.Bytes())

	// Verify all field types are present
	expectations := []string{
		"string_field=hello",
		"int_field=42",
		"int64_field=9223372036854775807",
		"float_field=3.14159", // Updated to match -1 precision
		"bool_true=true",
		"bool_false=false",
		"duration_field=5s",
	}

	for _, expected := range expectations {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected %s in output: %s", expected, output)
		}
	}
}

// TestFastTextEncoderStackTrace tests stack trace handling
func TestFastTextEncoderStackTrace(t *testing.T) {
	encoder := NewFastTextEncoder()
	stackTrace := "goroutine 1 [running]:\nmain.main()\n\t/path/to/main.go:42 +0x42"

	encoder.EncodeLogEntry(
		time.Time{},
		ErrorLevel,
		"error with stack",
		nil,
		Caller{},
		stackTrace,
	)

	output := string(encoder.Bytes())

	if !strings.Contains(output, "Stack trace:") {
		t.Error("Expected stack trace header")
	}
	if !strings.Contains(output, "goroutine 1") {
		t.Error("Expected stack trace content")
	}
}

// TestFastTextEncoderReset tests encoder reset functionality
func TestFastTextEncoderReset(t *testing.T) {
	encoder := NewFastTextEncoder()

	// First encoding
	encoder.EncodeLogEntry(
		time.Time{},
		InfoLevel,
		"first message",
		nil,
		Caller{},
		"",
	)

	if len(encoder.Bytes()) == 0 {
		t.Error("Expected content after first encoding")
	}

	// Reset and second encoding
	encoder.Reset()
	encoder.EncodeLogEntry(
		time.Time{},
		InfoLevel,
		"second message",
		nil,
		Caller{},
		"",
	)

	output := string(encoder.Bytes())
	if strings.Contains(output, "first message") {
		t.Error("Expected first message to be cleared after reset")
	}
	if !strings.Contains(output, "second message") {
		t.Error("Expected second message after reset")
	}
}

// TestBinaryEncoder tests the BinaryEncoder
func TestBinaryEncoder(t *testing.T) {
	encoder := NewBinaryEncoder()
	timestamp := time.Unix(1627890123, 456789000)

	encoder.EncodeLogEntry(
		timestamp,
		InfoLevel,
		"binary test",
		ToBinaryFields([]Field{
			Str("service", "test"),
			Int("id", 123),
			Bool("enabled", true),
		}),
		Caller{Valid: true, File: "test.go", Line: 42},
		"",
	)

	output := encoder.Bytes()

	// Verify binary structure (basic sanity checks)
	if len(output) < 20 {
		t.Error("Binary output too short")
	}

	// Check timestamp (first 8 bytes)
	expectedNano := timestamp.UnixNano()
	actualNano := int64(output[0])<<56 | int64(output[1])<<48 |
		int64(output[2])<<40 | int64(output[3])<<32 |
		int64(output[4])<<24 | int64(output[5])<<16 |
		int64(output[6])<<8 | int64(output[7])

	if actualNano != expectedNano {
		t.Errorf("Expected timestamp %d, got %d", expectedNano, actualNano)
	}

	// Check level (9th byte)
	if output[8] != byte(InfoLevel) {
		t.Errorf("Expected level %d, got %d", InfoLevel, output[8])
	}

	// Check flags (10th byte - should have caller bit set)
	if output[9]&1 == 0 {
		t.Error("Expected caller flag to be set")
	}
}

// TestBinaryEncoderFieldTypes tests binary encoding of different field types
func TestBinaryEncoderFieldTypes(t *testing.T) {
	encoder := NewBinaryEncoder()

	fields := []Field{
		Str("str", "test"),
		Int("int", 42),
		Float64("float", 3.14),
		Bool("bool", true),
		Duration("dur", time.Second),
	}

	encoder.EncodeLogEntry(
		time.Time{},
		InfoLevel,
		"test",
		ToBinaryFields(fields),
		Caller{},
		"",
	)

	output := encoder.Bytes()

	// Verify we have some reasonable output
	if len(output) < 50 {
		t.Error("Binary output seems too short for all fields")
	}
}

// TestBinaryEncoderReset tests binary encoder reset
func TestBinaryEncoderReset(t *testing.T) {
	encoder := NewBinaryEncoder()

	// First encoding
	encoder.EncodeLogEntry(
		time.Unix(1000, 0),
		InfoLevel,
		"first",
		nil,
		Caller{},
		"",
	)

	firstLen := len(encoder.Bytes())
	if firstLen == 0 {
		t.Error("Expected content after first encoding")
	}

	// Reset and encode again
	encoder.Reset()
	encoder.EncodeLogEntry(
		time.Unix(2000, 0),
		InfoLevel,
		"second",
		nil,
		Caller{},
		"",
	)

	// Should not contain data from first encoding
	output := encoder.Bytes()
	if len(output) >= firstLen*2 {
		t.Error("Reset may not have worked - output too long")
	}
}

// TestLastIndexByte tests the utility function
func TestLastIndexByte(t *testing.T) {
	tests := []struct {
		input    string
		char     byte
		expected int
	}{
		{"hello/world/test.go", '/', 11},
		{"no-slash", '/', -1},
		{"", '/', -1},
		{"a", 'a', 0},
		{"/leading/slash", '/', 8},
	}

	for _, tc := range tests {
		result := lastIndexByte(tc.input, tc.char)
		if result != tc.expected {
			t.Errorf("lastIndexByte(%q, %c) = %d, expected %d",
				tc.input, tc.char, result, tc.expected)
		}
	}
}

// TestFormatConstants tests format constants
func TestFormatConstants(t *testing.T) {
	// Verify format constants have expected values
	if JSONFormat != 0 {
		t.Error("JSONFormat should be 0")
	}
	if ConsoleFormat != 1 {
		t.Error("ConsoleFormat should be 1")
	}
	if FastTextFormat != 2 {
		t.Error("FastTextFormat should be 2")
	}
	if BinaryFormat != 3 {
		t.Error("BinaryFormat should be 3")
	}
}

// TestEncoderMemoryUsage tests memory efficiency
func TestEncoderMemoryUsage(t *testing.T) {
	// Test that encoders don't grow unbounded
	textEncoder := NewFastTextEncoder()
	binaryEncoder := NewBinaryEncoder()

	// Encode many entries
	for i := 0; i < 1000; i++ {
		textEncoder.EncodeLogEntry(
			time.Now(),
			InfoLevel,
			"memory test",
			ToBinaryFields([]Field{Int("iteration", i)}),
			Caller{},
			"",
		)
		textEncoder.Reset()

		binaryEncoder.EncodeLogEntry(
			time.Now(),
			InfoLevel,
			"memory test",
			ToBinaryFields([]Field{Int("iteration", i)}),
			Caller{},
			"",
		)
		binaryEncoder.Reset()
	}

	// Check that buffers haven't grown excessively
	if cap(textEncoder.buf) > 2048 {
		t.Errorf("Text encoder buffer grew too large: %d", cap(textEncoder.buf))
	}
	if cap(binaryEncoder.buf) > 1024 {
		t.Errorf("Binary encoder buffer grew too large: %d", cap(binaryEncoder.buf))
	}
}

// TestSanitizeForLogSafety tests the sanitizeForLogSafety function
func TestSanitizeForLogSafety(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal string", "hello world", "hello world"},
		{"empty string", "", ""},
		{"newline injection", "hello\nworld", "hello\\nworld"},
		{"carriage return", "hello\rworld", "hello\\rworld"},
		{"tab character", "hello\tworld", "hello world"},             // tab -> space
		{"null byte", "hello\x00world", "hello\\0world"},             // null -> \0
		{"control characters", "hello\x01\x1fworld", "hello??world"}, // control -> ?
		{"multiple issues", "hello\n\r\tworld\x00", "hello\\n\\r world\\0"},
		{"unicode safe", "hello 世界", "hello 世界"},
		{"quotes", "hello\"world", "hello\\\"world"},
		{"backslash", "hello\\world", "hello\\\\world"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := sanitizeForLogSafety(test.input)
			if result != test.expected {
				t.Errorf("sanitizeForLogSafety(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

// TestNeedsLogSafetySanitization tests the needsLogSafetySanitization function
func TestNeedsLogSafetySanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"safe string", "hello world", false},
		{"empty string", "", false},
		{"with newline", "hello\nworld", true},
		{"with carriage return", "hello\rworld", true},
		{"with tab", "hello\tworld", true},
		{"with null byte", "hello\x00world", true},
		{"with control char", "hello\x01world", true},
		{"unicode safe", "hello 世界", false},
		{"numbers and letters", "abc123XYZ", false},
		{"common punctuation", "hello, world!", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := needsLogSafetySanitization(test.input)
			if result != test.expected {
				t.Errorf("needsLogSafetySanitization(%q) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}

// TestNeedsQuotingFastSecure tests the needsQuotingFastSecure function edge cases
func TestNeedsQuotingFastSecure(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", true},
		{"simple word", "hello", false},
		{"with space", "hello world", true},
		{"with equals", "key=value", true},
		{"with quotes", "hello\"world", true},
		{"with backslash", "hello\\world", true},
		{"with newline", "hello\nworld", true},
		{"with tab", "hello\tworld", true},
		{"with brackets", "hello[world]", true},
		{"control character", "hello\x1fworld", true},
		{"del character", "hello\x7fworld", true},
		{"unicode safe", "hello世界", false},
		{"numbers", "12345", false},
		{"safe punctuation", "hello.world-test_123", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := needsQuotingFastSecure(test.input)
			if result != test.expected {
				t.Errorf("needsQuotingFastSecure(%q) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}
