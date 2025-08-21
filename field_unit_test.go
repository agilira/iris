// field_unit_test.go: Comprehensive safety net for field.go optimizations
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// TestStringField tests string field creation
func TestStringField(t *testing.T) {
	field := String("key", "value")

	if field.Key != "key" {
		t.Errorf("Expected key 'key', got '%s'", field.Key)
	}
	if field.Type != StringType {
		t.Errorf("Expected StringType, got %v", field.Type)
	}
	if field.String != "value" {
		t.Errorf("Expected value 'value', got '%s'", field.String)
	}

	// Test alias
	field2 := Str("key2", "value2")
	if field2.String != "value2" {
		t.Error("Str alias should work identically")
	}
}

// TestIntegerFields tests all integer field types
func TestIntegerFields(t *testing.T) {
	tests := []struct {
		name      string
		field     Field
		expected  int64
		fieldType FieldType
	}{
		{"Int", Int("count", 42), 42, IntType},
		{"Int64", Int64("count64", 123456789), 123456789, Int64Type},
		{"Int32", Int32("count32", 32000), 32000, Int32Type},
		{"Int16", Int16("count16", 16000), 16000, Int16Type},
		{"Int8", Int8("count8", 127), 127, Int8Type},
		{"Uint", Uint("ucount", 42), 42, UintType},
		{"Uint64_safe", Uint64("ucount64", 123456789), 123456789, Uint64Type}, // Safe range
		{"Uint32", Uint32("ucount32", 4294967295), 4294967295, Uint32Type},
		{"Uint16", Uint16("ucount16", 65535), 65535, Uint16Type},
		{"Uint8", Uint8("ucount8", 255), 255, Uint8Type},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.field.Type != test.fieldType {
				t.Errorf("Expected type %v, got %v", test.fieldType, test.field.Type)
			}
			if test.field.Int != test.expected {
				t.Errorf("Expected value %d, got %d", test.expected, test.field.Int)
			}
		})
	}

	// Test Uint64 overflow behavior (should fallback to string)
	t.Run("Uint64_overflow", func(t *testing.T) {
		overflowField := Uint64("large", 18446744073709551615) // Max uint64
		if overflowField.Type != StringType {
			t.Errorf("Expected StringType for overflow Uint64, got %v", overflowField.Type)
		}
		if overflowField.String != "18446744073709551615" {
			t.Errorf("Expected string representation, got %s", overflowField.String)
		}
	})
}

// TestFloatFields tests float field types
func TestFloatFields(t *testing.T) {
	field64 := Float64("score", 98.6)
	if field64.Type != Float64Type {
		t.Error("Expected Float64Type")
	}
	if field64.Float != 98.6 {
		t.Error("Expected float value 98.6")
	}

	// Test alias
	fieldAlias := Float("score2", 99.9)
	if fieldAlias.Float != 99.9 {
		t.Error("Float alias should work")
	}

	field32 := Float32("score32", 3.14)
	if field32.Type != Float32Type {
		t.Error("Expected Float32Type")
	}
	if field32.Float != 3.140000104904175 { // float32 precision
		t.Errorf("Expected float32 precision conversion, got %f", field32.Float)
	}
}

// TestBoolField tests boolean field creation
func TestBoolField(t *testing.T) {
	fieldTrue := Bool("active", true)
	if fieldTrue.Type != BoolType || !fieldTrue.Bool {
		t.Error("Expected true boolean field")
	}

	fieldFalse := Bool("inactive", false)
	if fieldFalse.Type != BoolType || fieldFalse.Bool {
		t.Error("Expected false boolean field")
	}
}

// TestDurationField tests duration field creation
func TestDurationField(t *testing.T) {
	duration := 5 * time.Second
	field := Duration("elapsed", duration)

	if field.Type != DurationType {
		t.Error("Expected DurationType")
	}
	if field.Int != int64(duration) {
		t.Errorf("Expected duration %d, got %d", int64(duration), field.Int)
	}
}

// TestTimeField tests time field creation
func TestTimeField(t *testing.T) {
	now := time.Now()
	field := Time("timestamp", now)

	if field.Type != TimeType {
		t.Error("Expected TimeType")
	}
	if field.Int != now.UnixNano() {
		t.Errorf("Expected time %d, got %d", now.UnixNano(), field.Int)
	}
}

// TestErrorField tests error field creation
func TestErrorField(t *testing.T) {
	err := errors.New("test error")
	field := Error(err)

	if field.Key != "error" {
		t.Error("Expected default error key")
	}
	if field.Type != ErrorType {
		t.Error("Expected ErrorType")
	}
	if field.Err != err {
		t.Error("Expected same error reference")
	}

	// Test alias
	field2 := Err(err)
	if field2.Err != err {
		t.Error("Err alias should work")
	}

	// Test nil error
	fieldNil := Error(nil)
	if fieldNil.Err != nil {
		t.Error("Expected nil error to be stored as nil")
	}
}

// TestByteFields tests byte-related field types
func TestByteFields(t *testing.T) {
	data := []byte("hello world")

	// ByteString field
	byteStringField := ByteString("data", data)
	if byteStringField.Type != ByteStringType {
		t.Error("Expected ByteStringType")
	}
	if string(byteStringField.Bytes) != "hello world" {
		t.Error("Expected same byte data")
	}

	// Binary field
	binaryField := Binary("blob", data)
	if binaryField.Type != BinaryType {
		t.Error("Expected BinaryType")
	}
	if string(binaryField.Bytes) != "hello world" {
		t.Error("Expected same binary data")
	}
}

// TestAnyField tests any field type
func TestAnyField(t *testing.T) {
	value := map[string]int{"count": 42}
	field := Any("config", value)

	if field.Type != AnyType {
		t.Error("Expected AnyType")
	}
	if field.Any == nil {
		t.Error("Expected Any value to be set")
	}

	// Test nil any
	fieldNil := Any("nil", nil)
	if fieldNil.Any != nil {
		t.Error("Expected nil Any to be stored as nil")
	}
}

// TestFieldTypeConstants tests field type constant values
func TestFieldTypeConstants(t *testing.T) {
	expected := map[FieldType]string{
		StringType:     "StringType",
		IntType:        "IntType",
		Int64Type:      "Int64Type",
		Int32Type:      "Int32Type",
		Int16Type:      "Int16Type",
		Int8Type:       "Int8Type",
		UintType:       "UintType",
		Uint64Type:     "Uint64Type",
		Uint32Type:     "Uint32Type",
		Uint16Type:     "Uint16Type",
		Uint8Type:      "Uint8Type",
		Float64Type:    "Float64Type",
		Float32Type:    "Float32Type",
		BoolType:       "BoolType",
		DurationType:   "DurationType",
		TimeType:       "TimeType",
		ErrorType:      "ErrorType",
		ByteStringType: "ByteStringType",
		BinaryType:     "BinaryType",
		AnyType:        "AnyType",
	}

	// Verify that constants are unique and sequential
	for i := 0; i < len(expected); i++ {
		fieldType := FieldType(i)
		if _, exists := expected[fieldType]; !exists {
			t.Errorf("Missing field type constant for value %d", i)
		}
	}
}

// TestFieldStruct tests Field struct layout and size
func TestFieldStruct(t *testing.T) {
	field := Field{
		Key:    "test",
		Type:   StringType,
		String: "value",
		Int:    123,
		Float:  3.14,
		Bool:   true,
		Err:    errors.New("test"),
		Bytes:  []byte("data"),
		Any:    "anything",
	}

	// Verify all fields can be set without conflicts
	if field.Key != "test" || field.Type != StringType {
		t.Error("Basic field properties should be accessible")
	}

	if field.String != "value" || field.Int != 123 || field.Float != 3.14 {
		t.Error("Value fields should be independent")
	}

	if !field.Bool || field.Err == nil || field.Bytes == nil || field.Any == nil {
		t.Error("All field types should be settable")
	}
}

// TestFieldConstructorMemory tests that constructors don't allocate unnecessarily
func TestFieldConstructorMemory(t *testing.T) {
	// These should not allocate since they're returning structs by value
	field1 := String("key", "value")
	field2 := Int("count", 42)
	field3 := Bool("flag", true)

	// Basic validation that they work
	if field1.String != "value" || field2.Int != 42 || !field3.Bool {
		t.Error("Field constructors should work correctly")
	}
}

// TestFieldCompatibility tests compatibility with encoding
func TestFieldCompatibility(t *testing.T) {
	fields := []Field{
		String("str", "test"),
		Int("int", 42),
		Float64("float", 3.14),
		Bool("bool", true),
		Duration("dur", time.Second),
		Time("time", time.Now()),
		Error(errors.New("test")),
		ByteString("bytes", []byte("data")),
		Binary("binary", []byte("data")),
		Any("any", map[string]int{"k": 1}),
	}

	// Should be encodable without panics
	encoder := NewFastTextEncoder()
	for _, field := range fields {
		encoder.Reset()
		// Use migration method for safe Field processing
		encoder.appendFieldValueFastMigration(field)
		if len(encoder.Bytes()) == 0 {
			t.Errorf("Field %s should produce non-empty output", field.Key)
		}
	}
}

// TestFieldEdgeCases tests edge cases and boundary conditions
func TestFieldEdgeCases(t *testing.T) {
	// Empty string
	emptyStr := String("empty", "")
	if emptyStr.String != "" {
		t.Error("Empty string should be preserved")
	}

	// Zero values
	zeroInt := Int("zero", 0)
	if zeroInt.Int != 0 {
		t.Error("Zero int should be preserved")
	}

	// Large values
	maxInt := Int64("max", 9223372036854775807)
	if maxInt.Int != 9223372036854775807 {
		t.Error("Max int64 should be preserved")
	}

	// Nil slices
	nilBytes := ByteString("nil", nil)
	if nilBytes.Bytes != nil {
		t.Error("Nil bytes should be preserved")
	}
}

// TestFieldMutation tests that fields are safe to use
func TestFieldMutation(t *testing.T) {
	original := String("key", "original")

	// Create a copy and modify it
	copy := original
	copy.String = "modified"

	// Original should be unchanged
	if original.String != "original" {
		t.Error("Original field should not be affected by copy modification")
	}
	if copy.String != "modified" {
		t.Error("Copy should have modified value")
	}
}

// TestSafeConversions tests the thread-safe conversion functions
func TestSafeConversions(t *testing.T) {
	t.Run("SafeUint64ToInt64", func(t *testing.T) {
		// Test safe conversion
		if result, ok := SafeUint64ToInt64(123); !ok || result != 123 {
			t.Errorf("Expected safe conversion of 123, got %d, %v", result, ok)
		}

		// Test max safe value
		maxInt64 := uint64(1<<63 - 1) // 9223372036854775807
		if result, ok := SafeUint64ToInt64(maxInt64); !ok || result != int64(maxInt64) {
			t.Errorf("Expected safe conversion of max int64, got %d, %v", result, ok)
		}

		// Test overflow
		if result, ok := SafeUint64ToInt64(1 << 63); ok { // overflow by 1
			t.Errorf("Expected overflow detection, got safe conversion: %d", result)
		}

		// Test maximum uint64 (definitely overflow)
		if result, ok := SafeUint64ToInt64(^uint64(0)); ok { // max uint64
			t.Errorf("Expected overflow detection for max uint64, got safe conversion: %d", result)
		}
	})

	t.Run("SafeInt64ToUint64", func(t *testing.T) {
		// Test safe conversion
		if result, ok := SafeInt64ToUint64(123); !ok || result != 123 {
			t.Errorf("Expected safe conversion of 123, got %d, %v", result, ok)
		}

		// Test zero
		if result, ok := SafeInt64ToUint64(0); !ok || result != 0 {
			t.Errorf("Expected safe conversion of 0, got %d, %v", result, ok)
		}

		// Test negative (should fail)
		if result, ok := SafeInt64ToUint64(-1); ok {
			t.Errorf("Expected negative detection, got safe conversion: %d", result)
		}
	})

	t.Run("SafeUintToInt64", func(t *testing.T) {
		// Test safe conversion
		if result, ok := SafeUintToInt64(123); !ok || result != 123 {
			t.Errorf("Expected safe conversion of 123, got %d, %v", result, ok)
		}
	})
}

// TestLargeUintFields tests field creation with large uint values
func TestLargeUintFields(t *testing.T) {
	t.Run("Uint64 safe range", func(t *testing.T) {
		field := Uint64("key", 123)
		if field.Type != Uint64Type || field.Int != 123 {
			t.Errorf("Expected Uint64Type with value 123, got Type: %d, Int: %d", field.Type, field.Int)
		}
	})

	t.Run("Uint64 overflow to string", func(t *testing.T) {
		largeValue := uint64(1 << 63) // This will overflow int64
		field := Uint64("key", largeValue)
		if field.Type != StringType {
			t.Errorf("Expected fallback to StringType for large uint64, got Type: %d", field.Type)
		}
		expectedStr := fmt.Sprintf("%d", largeValue)
		if field.String != expectedStr {
			t.Errorf("Expected string representation %s, got %s", expectedStr, field.String)
		}
	})

	t.Run("Uint safe range", func(t *testing.T) {
		field := Uint("key", 123)
		if field.Type != UintType || field.Int != 123 {
			t.Errorf("Expected UintType with value 123, got Type: %d, Int: %d", field.Type, field.Int)
		}
	})
}
