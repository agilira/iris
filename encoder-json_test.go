// encoder_test.go: Comprehensive test suite for iris JSON encoder
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestNewJSONEncoder tests encoder creation with defaults
func TestNewJSONEncoder(t *testing.T) {
	encoder := NewJSONEncoder()

	if encoder.TimeKey != "ts" {
		t.Errorf("Expected default TimeKey 'ts', got %s", encoder.TimeKey)
	}

	if encoder.LevelKey != "level" {
		t.Errorf("Expected default LevelKey 'level', got %s", encoder.LevelKey)
	}

	if encoder.MsgKey != "msg" {
		t.Errorf("Expected default MsgKey 'msg', got %s", encoder.MsgKey)
	}

	if !encoder.RFC3339 {
		t.Error("Expected default RFC3339 to be true")
	}
}

// TestNewRecord tests record creation and basic operations
func TestNewRecord(t *testing.T) {
	record := NewRecord(Info, "test message")

	if record.Level != Info {
		t.Errorf("Expected level Info, got %v", record.Level)
	}

	if record.Msg != "test message" {
		t.Errorf("Expected message 'test message', got %s", record.Msg)
	}

	if record.FieldCount() != 0 {
		t.Errorf("Expected 0 fields, got %d", record.FieldCount())
	}
}

// TestRecordAddField tests adding fields to records
func TestRecordAddField(t *testing.T) {
	record := NewRecord(Info, "test")

	// Add various field types
	fields := []Field{
		String("str", "value"),
		Int64("int", 42),
		Bool("bool", true),
		Float64("float", 3.14),
	}

	for i, field := range fields {
		if !record.AddField(field) {
			t.Errorf("Failed to add field %d", i)
		}
	}

	if record.FieldCount() != len(fields) {
		t.Errorf("Expected %d fields, got %d", len(fields), record.FieldCount())
	}

	// Test field retrieval
	for i, expected := range fields {
		actual := record.GetField(i)
		if actual.K != expected.K {
			t.Errorf("Field %d: expected key %s, got %s", i, expected.K, actual.K)
		}
		if actual.T != expected.T {
			t.Errorf("Field %d: expected type %v, got %v", i, expected.T, actual.T)
		}
	}
}

// TestRecordFieldLimit tests the 16-field limit
func TestRecordFieldLimit(t *testing.T) {
	record := NewRecord(Info, "test")

	// Add exactly 32 fields (the new optimized maximum)
	for i := 0; i < 32; i++ {
		field := String("field", "value")
		if !record.AddField(field) {
			t.Errorf("Failed to add field %d (should succeed)", i)
		}
	}

	// Try to add 33rd field (should fail)
	field := String("overflow", "value")
	if record.AddField(field) {
		t.Error("Should not be able to add 33rd field")
	}

	if record.FieldCount() != 32 {
		t.Errorf("Expected 32 fields, got %d", record.FieldCount())
	}
}

// TestRecordReset tests record reset functionality
func TestRecordReset(t *testing.T) {
	record := NewRecord(Error, "error message")
	record.AddField(String("key", "value"))

	// Verify initial state
	if record.Level != Error || record.Msg != "error message" || record.FieldCount() != 1 {
		t.Error("Record not in expected initial state")
	}

	// Reset and verify
	record.Reset()

	if record.Level != Debug {
		t.Errorf("Expected level Debug after reset, got %v", record.Level)
	}

	if record.Msg != "" {
		t.Errorf("Expected empty message after reset, got %s", record.Msg)
	}

	if record.FieldCount() != 0 {
		t.Errorf("Expected 0 fields after reset, got %d", record.FieldCount())
	}
}

// TestRecordGetFieldBounds tests bounds checking for GetField
func TestRecordGetFieldBounds(t *testing.T) {
	record := NewRecord(Info, "test")
	record.AddField(String("key", "value"))

	// Valid index
	field := record.GetField(0)
	if field.K != "key" {
		t.Errorf("Expected field key 'key', got %s", field.K)
	}

	// Invalid indices should return zero Field
	negativeField := record.GetField(-1)
	if negativeField.K != "" || negativeField.T != 0 {
		t.Error("Expected zero field for negative index")
	}

	outOfBoundsField := record.GetField(1)
	if outOfBoundsField.K != "" || outOfBoundsField.T != 0 {
		t.Error("Expected zero field for out-of-bounds index")
	}
}

// TestJSONEncoderBasic tests basic JSON encoding
func TestJSONEncoderBasic(t *testing.T) {
	encoder := NewJSONEncoder()
	record := NewRecord(Info, "test message")
	buf := &bytes.Buffer{}
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	encoder.Encode(record, now, buf)

	output := buf.String()

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Check required fields
	if parsed["ts"] != "2025-01-01T12:00:00Z" {
		t.Errorf("Expected timestamp '2025-01-01T12:00:00Z', got %v", parsed["ts"])
	}

	if parsed["level"] != "info" {
		t.Errorf("Expected level 'info', got %v", parsed["level"])
	}

	if parsed["msg"] != "test message" {
		t.Errorf("Expected message 'test message', got %v", parsed["msg"])
	}

	// Should end with newline
	if !strings.HasSuffix(output, "\n") {
		t.Error("Output should end with newline")
	}
}

// TestJSONEncoderWithFields tests encoding with structured fields
func TestJSONEncoderWithFields(t *testing.T) {
	encoder := NewJSONEncoder()
	record := NewRecord(Warn, "warning message")

	// Add various field types
	record.AddField(String("string_field", "text"))
	record.AddField(Int64("int_field", 42))
	record.AddField(Bool("bool_field", true))
	record.AddField(Float64("float_field", 3.14))

	buf := &bytes.Buffer{}
	now := time.Unix(1234567890, 0)

	encoder.Encode(record, now, buf)

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Check fields
	if parsed["string_field"] != "text" {
		t.Errorf("Expected string_field 'text', got %v", parsed["string_field"])
	}

	if parsed["int_field"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected int_field 42, got %v", parsed["int_field"])
	}

	if parsed["bool_field"] != true {
		t.Errorf("Expected bool_field true, got %v", parsed["bool_field"])
	}

	if parsed["float_field"] != 3.14 {
		t.Errorf("Expected float_field 3.14, got %v", parsed["float_field"])
	}
}

// TestJSONEncoderCustomKeys tests custom field keys
func TestJSONEncoderCustomKeys(t *testing.T) {
	encoder := &JSONEncoder{
		TimeKey:  "timestamp",
		LevelKey: "severity",
		MsgKey:   "message",
		RFC3339:  true,
	}

	record := NewRecord(Error, "custom keys test")
	buf := &bytes.Buffer{}
	now := time.Now()

	encoder.Encode(record, now, buf)

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Check custom keys exist
	if _, exists := parsed["timestamp"]; !exists {
		t.Error("Expected custom timestamp key")
	}

	if _, exists := parsed["severity"]; !exists {
		t.Error("Expected custom severity key")
	}

	if _, exists := parsed["message"]; !exists {
		t.Error("Expected custom message key")
	}

	// Check old keys don't exist
	if _, exists := parsed["ts"]; exists {
		t.Error("Old timestamp key should not exist")
	}
}

// TestJSONEncoderUnixTime tests Unix timestamp format
func TestJSONEncoderUnixTime(t *testing.T) {
	encoder := &JSONEncoder{
		TimeKey: "ts",
		RFC3339: false, // Use Unix nanoseconds
	}

	record := NewRecord(Info, "unix time test")
	buf := &bytes.Buffer{}
	now := time.Unix(1234567890, 123456789)

	encoder.Encode(record, now, buf)

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	expectedUnixNano := float64(now.UnixNano())
	if actualTime, ok := parsed["ts"].(float64); !ok || actualTime != expectedUnixNano {
		t.Errorf("Expected Unix nano %v, got %v", expectedUnixNano, parsed["ts"])
	}
} // TestJSONEncoderEmptyMessage tests encoding with empty message
func TestJSONEncoderEmptyMessage(t *testing.T) {
	encoder := NewJSONEncoder()
	record := NewRecord(Info, "") // Empty message
	buf := &bytes.Buffer{}
	now := time.Now()

	encoder.Encode(record, now, buf)

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Message field should not be present
	if _, exists := parsed["msg"]; exists {
		t.Error("Empty message should not create msg field")
	}
}

// TestJSONEncoderStringEscaping tests string escaping in JSON
func TestJSONEncoderStringEscaping(t *testing.T) {
	encoder := NewJSONEncoder()

	testCases := []struct {
		input    string
		expected string
	}{
		{`simple`, `"simple"`},
		{`"quoted"`, `"\"quoted\""`},
		{"line\nbreak", `"line\nbreak"`},
		{"tab\ttab", `"tab\ttab"`},
		{"carriage\rreturn", `"carriage\rreturn"`},
		{"back\\slash", `"back\\slash"`},
		{"\x01control", `"\u0001control"`},
	}

	for _, tc := range testCases {
		record := NewRecord(Info, tc.input)
		buf := &bytes.Buffer{}
		now := time.Now()

		encoder.Encode(record, now, buf)

		// Extract just the message field value from JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
			t.Errorf("Failed to parse JSON for input %q: %v", tc.input, err)
			continue
		}

		// Verify the message was properly escaped and can round-trip
		if parsed["msg"] != tc.input {
			t.Errorf("String escaping failed for %q: got %q", tc.input, parsed["msg"])
		}
	}
}

// TestJSONEncoderSpecialFieldTypes tests encoding of special field types
func TestJSONEncoderSpecialFieldTypes(t *testing.T) {
	encoder := NewJSONEncoder()
	record := NewRecord(Info, "special types")

	// Add special types
	duration := time.Hour + 2*time.Minute + 3*time.Second
	timestamp := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	byteData := []byte{1, 2, 3, 255}

	record.AddField(Dur("duration", duration))
	record.AddField(Time("timestamp", timestamp))
	record.AddField(Bytes("data", byteData))

	buf := &bytes.Buffer{}
	now := time.Now()

	encoder.Encode(record, now, buf)

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Duration should be nanoseconds
	expectedDurationNs := float64(duration.Nanoseconds())
	if parsed["duration"] != expectedDurationNs {
		t.Errorf("Expected duration %v ns, got %v", expectedDurationNs, parsed["duration"])
	}

	// Timestamp should be RFC3339 in UTC (exact match for UTC time)
	expectedTime := timestamp.Format(time.RFC3339Nano)
	if actualTime, ok := parsed["timestamp"].(string); !ok || actualTime != expectedTime {
		t.Errorf("Expected timestamp '%s', got %v", expectedTime, parsed["timestamp"])
	} // Bytes should be array of integers
	expectedBytes := []interface{}{float64(1), float64(2), float64(3), float64(255)}
	actualBytes := parsed["data"].([]interface{})
	if len(actualBytes) != len(expectedBytes) {
		t.Errorf("Expected %d bytes, got %d", len(expectedBytes), len(actualBytes))
	}
	for i, expected := range expectedBytes {
		if actualBytes[i] != expected {
			t.Errorf("Byte %d: expected %v, got %v", i, expected, actualBytes[i])
		}
	}
}

// TestJSONEncoderZeroValueEncoder tests encoding with zero-value encoder
func TestJSONEncoderZeroValueEncoder(t *testing.T) {
	var encoder JSONEncoder // Zero value (RFC3339 will be false)
	record := NewRecord(Info, "zero encoder test")
	buf := &bytes.Buffer{}
	now := time.Unix(1234567890, 0) // Use fixed time for predictable output

	encoder.Encode(record, now, buf)

	// Should still produce valid JSON with defaults
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Errorf("Zero-value encoder should produce valid JSON: %v", err)
	}

	// Should have default keys
	if _, exists := parsed["ts"]; !exists {
		t.Error("Expected default timestamp key 'ts'")
	}

	if _, exists := parsed["level"]; !exists {
		t.Error("Expected default level key 'level'")
	}

	// Timestamp should be Unix nanoseconds (since RFC3339 defaults to false)
	expectedUnixNano := float64(now.UnixNano())
	if actualTime, ok := parsed["ts"].(float64); !ok || actualTime != expectedUnixNano {
		t.Errorf("Expected Unix nanoseconds %v, got %v", expectedUnixNano, parsed["ts"])
	}
} // TestJSONEncoderPerformance tests that encoding is fast
func TestJSONEncoderPerformance(t *testing.T) {
	encoder := NewJSONEncoder()
	record := NewRecord(Info, "performance test message")

	// Add some fields
	record.AddField(String("service", "test-service"))
	record.AddField(Int64("user_id", 12345))
	record.AddField(Bool("success", true))
	record.AddField(Float64("duration", 0.123))

	buf := &bytes.Buffer{}
	now := time.Now()

	// Perform many encodings (should complete quickly)
	for i := 0; i < 1000; i++ {
		buf.Reset()
		encoder.Encode(record, now, buf)
	}

	// Final encode to verify still working
	buf.Reset()
	encoder.Encode(record, now, buf)

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Errorf("Performance test output is not valid JSON: %v", err)
	}
}
