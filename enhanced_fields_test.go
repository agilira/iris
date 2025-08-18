// enhanced_fields_test.go: Test the enhanced field types
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestEnhancedFieldTypes(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected string
	}{
		{"Int32", Int32("test", 32), `"test":32`},
		{"Int16", Int16("test", 16), `"test":16`},
		{"Int8", Int8("test", 8), `"test":8`},
		{"Uint", Uint("test", 100), `"test":100`},
		{"Uint64", Uint64("test", 64), `"test":64`},
		{"Uint32", Uint32("test", 32), `"test":32`},
		{"Uint16", Uint16("test", 16), `"test":16`},
		{"Uint8", Uint8("test", 8), `"test":8`},
		{"Float32", Float32("test", 3.14), `"test":3.14`}, // Fixed: corrected precision
		{"ByteString", ByteString("test", []byte("hello")), `"test":"hello"`},
		{"Binary", Binary("test", []byte{1, 2, 3}), `"test":"` + base64.StdEncoding.EncodeToString([]byte{1, 2, 3}) + `"`},
		{"Any", Any("test", map[string]int{"count": 42}), `"test":{"count":42}`},
	}

	encoder := NewJSONEncoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder.buf = encoder.buf[:0] // Reset buffer
			encoder.encodeField(tt.field)

			result := string(encoder.buf)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestConsoleEncoderEnhancedFields(t *testing.T) {
	encoder := NewConsoleEncoder(false)

	entry := &LogEntry{
		Level:   InfoLevel,
		Message: "test",
		Fields: []Field{
			Int32("int32", 32),
			Uint64("uint64", 64),
			Float32("float32", 3.14),
			ByteString("bytes", []byte("hello")),
			Binary("binary", []byte{1, 2, 3}),
		},
	}

	var buf []byte
	result := string(encoder.EncodeLogEntry(entry, buf))

	expectedParts := []string{
		"int32=32",
		"uint64=64",
		"float32=3.14",
		"bytes=hello",
		"binary=<binary:3 bytes>",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected %q in result: %q", part, result)
		}
	}
}

func TestFieldTypeCompatibility(t *testing.T) {
	// Test that all field types can be created and encoded without errors
	fields := []Field{
		String("str", "value"),
		Int("int", 1),
		Int32("int32", 32),
		Int16("int16", 16),
		Int8("int8", 8),
		Uint("uint", 1),
		Uint64("uint64", 64),
		Uint32("uint32", 32),
		Uint16("uint16", 16),
		Uint8("uint8", 8),
		Float("float", 1.0),
		Float32("float32", 3.14),
		Bool("bool", true),
		ByteString("bytes", []byte("test")),
		Binary("binary", []byte{1, 2, 3}),
		Any("any", "anything"),
	}

	encoder := NewJSONEncoder()

	for _, field := range fields {
		encoder.buf = encoder.buf[:0]
		encoder.encodeField(field)

		// Should not panic and should produce valid output
		result := string(encoder.buf)
		if len(result) == 0 {
			t.Errorf("Field %s produced empty output", field.Key)
		}

		// Basic validation that it looks like JSON
		if !strings.Contains(result, field.Key) {
			t.Errorf("Field %s output doesn't contain key: %q", field.Key, result)
		}
	}
}

func BenchmarkEnhancedFieldTypes(b *testing.B) {
	encoder := NewJSONEncoder()

	fields := []Field{
		Int32("int32", 32),
		Uint64("uint64", 64),
		Float32("float32", 3.14),
		ByteString("bytes", []byte("hello world")),
		Binary("binary", []byte{1, 2, 3, 4, 5}),
		Any("any", map[string]int{"count": 42}),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, field := range fields {
			encoder.buf = encoder.buf[:0]
			encoder.encodeField(field)
		}
	}
}

func TestAnyFieldComplexTypes(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"string", "hello"},
		{"number", 42},
		{"slice", []int{1, 2, 3}},
		{"map", map[string]string{"key": "value"}},
		{"struct", struct{ Name string }{"test"}},
		{"nil", nil},
	}

	encoder := NewJSONEncoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := Any("test", tt.value)
			encoder.buf = encoder.buf[:0]
			encoder.encodeField(field)

			result := string(encoder.buf)

			// Verify it's valid JSON
			var decoded interface{}
			fullJSON := "{" + result + "}"
			if err := json.Unmarshal([]byte(fullJSON), &decoded); err != nil {
				t.Errorf("Invalid JSON produced for %s: %s", tt.name, result)
			}
		})
	}
}
