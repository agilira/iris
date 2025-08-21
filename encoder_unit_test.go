// encoder_unit_test.go: Comprehensive safety net for encoder optimizations
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestJSONEncoderBasics tests core JSON encoder functionality
func TestJSONEncoderBasics(t *testing.T) {
	encoder := NewJSONEncoder()
	defer encoder.Reset()

	timestamp := time.Date(2025, 8, 21, 15, 30, 45, 123456789, time.UTC)
	caller := Caller{Valid: true, File: "test.go", Line: 42, Function: "TestFunc"}
	fields := []Field{
		Str("service", "test"),
		Int("count", 42),
		Bool("active", true),
	}

	encoder.EncodeLogEntry(timestamp, InfoLevel, "test message", fields, caller, "")

	// Verify JSON is valid
	var parsed map[string]interface{}
	if err := json.Unmarshal(encoder.Bytes(), &parsed); err != nil {
		t.Fatalf("Generated invalid JSON: %v\nJSON: %s", err, string(encoder.Bytes()))
	}

	// Verify essential fields
	if parsed["level"] != "INFO" {
		t.Errorf("Level not correctly encoded: got %v, want INFO", parsed["level"])
	}
	if parsed["message"] != "test message" {
		t.Error("Message not correctly encoded")
	}
	if parsed["service"] != "test" {
		t.Error("Service field not correctly encoded")
	}
	if parsed["count"] != float64(42) { // JSON numbers are float64
		t.Error("Count field not correctly encoded")
	}
	if parsed["active"] != true {
		t.Error("Active field not correctly encoded")
	}
}

// TestJSONEncoderFieldTypes tests all field types encoding
func TestJSONEncoderFieldTypes(t *testing.T) {
	encoder := NewJSONEncoder()
	defer encoder.Reset()

	fields := []Field{
		Str("string", "hello world"),
		Int("int", 123),
		Int64("int64", 9876543210),
		Int32("int32", 456),
		Float64("float64", 3.14159),
		Float32("float32", 2.718),
		Bool("bool_true", true),
		Bool("bool_false", false),
		Duration("duration", 5*time.Second),
		ByteString("bytes", []byte("test bytes")),
		Any("any_string", "any value"),
		Any("any_int", 789),
		Any("any_nil", nil),
	}

	encoder.EncodeLogEntry(time.Time{}, InfoLevel, "field types test", fields, Caller{}, "")

	// Verify JSON is valid
	var parsed map[string]interface{}
	if err := json.Unmarshal(encoder.Bytes(), &parsed); err != nil {
		t.Fatalf("Generated invalid JSON: %v\nJSON: %s", err, string(encoder.Bytes()))
	}

	// Verify specific field types
	if parsed["string"] != "hello world" {
		t.Error("String field not correctly encoded")
	}
	if parsed["int"] != float64(123) {
		t.Error("Int field not correctly encoded")
	}
	if parsed["bool_true"] != true || parsed["bool_false"] != false {
		t.Error("Bool fields not correctly encoded")
	}
	if parsed["any_string"] != "any value" {
		t.Error("Any string field not correctly encoded")
	}
	if parsed["any_int"] != float64(789) {
		t.Error("Any int field not correctly encoded")
	}
	if parsed["any_nil"] != nil {
		t.Error("Any nil field not correctly encoded")
	}
}

// TestJSONEncoderStringEscaping tests string escaping functionality
func TestJSONEncoderStringEscaping(t *testing.T) {
	encoder := NewJSONEncoder()
	defer encoder.Reset()

	specialStrings := []struct {
		input    string
		expected string
	}{
		{`hello "world"`, `hello \"world\"`},
		{"hello\nworld", `hello\nworld`},
		{"hello\tworld", `hello\tworld`},
		{"hello\rworld", `hello\rworld`},
		{`hello\world`, `hello\\world`},
		{"normal string", "normal string"},
	}

	for _, test := range specialStrings {
		encoder.Reset()

		fields := []Field{Str("test", test.input)}
		encoder.EncodeLogEntry(time.Time{}, InfoLevel, "escape test", fields, Caller{}, "")

		// Verify JSON is valid
		var parsed map[string]interface{}
		if err := json.Unmarshal(encoder.Bytes(), &parsed); err != nil {
			t.Fatalf("Generated invalid JSON for input %q: %v\nJSON: %s", test.input, err, string(encoder.Bytes()))
		}

		// Verify the escaped value matches expected
		if parsed["test"] != test.input {
			t.Errorf("String escaping failed for %q: got %q, want %q", test.input, parsed["test"], test.input)
		}
	}
}

// TestJSONEncoderCaller tests caller information encoding
func TestJSONEncoderCaller(t *testing.T) {
	encoder := NewJSONEncoder()
	defer encoder.Reset()

	// Test with caller
	caller := Caller{
		Valid:    true,
		File:     "/path/to/file.go",
		Line:     123,
		Function: "TestFunction",
	}

	encoder.EncodeLogEntry(time.Time{}, InfoLevel, "caller test", nil, caller, "")

	var parsed map[string]interface{}
	if err := json.Unmarshal(encoder.Bytes(), &parsed); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	if !strings.Contains(parsed["caller"].(string), "/path/to/file.go:123") {
		t.Error("Caller not correctly encoded")
	}
	if parsed["function"] != "TestFunction" {
		t.Error("Function not correctly encoded")
	}

	// Test without caller
	encoder.Reset()
	encoder.EncodeLogEntry(time.Time{}, InfoLevel, "no caller test", nil, Caller{}, "")

	parsed = make(map[string]interface{})
	json.Unmarshal(encoder.Bytes(), &parsed)

	if _, hasCaller := parsed["caller"]; hasCaller {
		t.Error("Caller should not be present when invalid")
	}
}

// TestJSONEncoderStackTrace tests stack trace encoding
func TestJSONEncoderStackTrace(t *testing.T) {
	encoder := NewJSONEncoder()
	defer encoder.Reset()

	stackTrace := "goroutine 1 [running]:\nmain.main()\n\t/path/to/main.go:42 +0x123"

	encoder.EncodeLogEntry(time.Time{}, ErrorLevel, "error with stack", nil, Caller{}, stackTrace)

	var parsed map[string]interface{}
	if err := json.Unmarshal(encoder.Bytes(), &parsed); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	if parsed["stacktrace"] != stackTrace {
		t.Error("Stack trace not correctly encoded")
	}
}

// TestJSONEncoderTimestamp tests timestamp encoding
func TestJSONEncoderTimestamp(t *testing.T) {
	encoder := NewJSONEncoder()
	defer encoder.Reset()

	timestamp := time.Date(2025, 8, 21, 15, 30, 45, 123456789, time.UTC)

	encoder.EncodeLogEntry(timestamp, InfoLevel, "timestamp test", nil, Caller{}, "")

	var parsed map[string]interface{}
	if err := json.Unmarshal(encoder.Bytes(), &parsed); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	timestampStr, ok := parsed["timestamp"].(string)
	if !ok {
		t.Error("Timestamp not encoded as string")
	}

	// Parse the timestamp back
	parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		t.Errorf("Invalid timestamp format: %v", err)
	}

	if !parsedTime.Equal(timestamp) {
		t.Errorf("Timestamp mismatch: got %v, want %v", parsedTime, timestamp)
	}
}

// TestJSONEncoderBufferReuse tests buffer pooling and reuse
func TestJSONEncoderBufferReuse(t *testing.T) {
	encoder := NewJSONEncoder()

	// First encode
	encoder.EncodeLogEntry(time.Time{}, InfoLevel, "first message", nil, Caller{}, "")
	firstResult := string(encoder.Bytes())

	// Reset and encode again
	encoder.Reset()
	encoder.EncodeLogEntry(time.Time{}, WarnLevel, "second message", nil, Caller{}, "")
	secondResult := string(encoder.Bytes())

	// Verify both are valid and different
	var first, second map[string]interface{}
	if err := json.Unmarshal([]byte(firstResult), &first); err != nil {
		t.Fatalf("First result invalid JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(secondResult), &second); err != nil {
		t.Fatalf("Second result invalid JSON: %v", err)
	}

	if first["message"] != "first message" {
		t.Error("First message not correct")
	}
	if second["message"] != "second message" {
		t.Error("Second message not correct")
	}
	if second["level"] != "WARN" {
		t.Error("Second level not correct")
	}

	encoder.Reset()
}

// TestJSONEncoderErrorHandling tests error field encoding
func TestJSONEncoderErrorHandling(t *testing.T) {
	encoder := NewJSONEncoder()
	defer encoder.Reset()

	// Create an error field
	err := &CustomError{Message: "test error"}
	fields := []Field{
		Error(err),
		{Key: "nil_error", Type: ErrorType, Err: nil},
	}

	encoder.EncodeLogEntry(time.Time{}, ErrorLevel, "error test", fields, Caller{}, "")

	var parsed map[string]interface{}
	if err := json.Unmarshal(encoder.Bytes(), &parsed); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	if parsed["error"] != "test error" {
		t.Error("Error field not correctly encoded")
	}
	if parsed["nil_error"] != nil {
		t.Error("Nil error field should be null")
	}
}

// CustomError for testing
type CustomError struct {
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}

// TestJSONEncoderEmptyMessage tests edge cases
func TestJSONEncoderEmptyMessage(t *testing.T) {
	encoder := NewJSONEncoder()
	defer encoder.Reset()

	// Empty message
	encoder.EncodeLogEntry(time.Time{}, InfoLevel, "", nil, Caller{}, "")

	var parsed map[string]interface{}
	if err := json.Unmarshal(encoder.Bytes(), &parsed); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	if parsed["message"] != "" {
		t.Error("Empty message not correctly encoded")
	}
}

// TestJSONEncoderLargeData tests performance with larger data
func TestJSONEncoderLargeData(t *testing.T) {
	encoder := NewJSONEncoder()
	defer encoder.Reset()

	// Create many fields
	fields := make([]Field, 50)
	for i := 0; i < 50; i++ {
		fields[i] = Str("field_"+strconv.Itoa(i), "value_"+strings.Repeat("x", 100))
	}

	encoder.EncodeLogEntry(time.Now(), InfoLevel, "large data test", fields, Caller{}, "")

	var parsed map[string]interface{}
	if err := json.Unmarshal(encoder.Bytes(), &parsed); err != nil {
		t.Fatalf("Generated invalid JSON for large data: %v", err)
	}

	// Verify we have all fields
	for i := 0; i < 50; i++ {
		fieldName := "field_" + strconv.Itoa(i)
		if _, exists := parsed[fieldName]; !exists {
			t.Errorf("Field %s missing from large data encoding", fieldName)
		}
	}
}
