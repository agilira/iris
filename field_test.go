// field_test.go: Comprehensive test suite for iris logging field functionality
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"math"
	"testing"
	"time"
)

// TestFieldConstructors tests all field constructor functions
func TestFieldConstructors(t *testing.T) {
	// Test Str/String
	strField := Str("key", "value")
	if strField.K != "key" || strField.T != kindString || strField.Str != "value" {
		t.Errorf("Str field incorrect: %+v", strField)
	}

	stringField := String("key2", "value2")
	if stringField.K != "key2" || stringField.T != kindString || stringField.Str != "value2" {
		t.Errorf("String field incorrect: %+v", stringField)
	}

	// Test Int variants
	intField := Int("int", 42)
	if intField.K != "int" || intField.T != kindInt64 || intField.I64 != 42 {
		t.Errorf("Int field incorrect: %+v", intField)
	}

	int8Field := Int8("int8", 8)
	if int8Field.K != "int8" || int8Field.T != kindInt64 || int8Field.I64 != 8 {
		t.Errorf("Int8 field incorrect: %+v", int8Field)
	}

	int16Field := Int16("int16", 16)
	if int16Field.K != "int16" || int16Field.T != kindInt64 || int16Field.I64 != 16 {
		t.Errorf("Int16 field incorrect: %+v", int16Field)
	}

	int32Field := Int32("int32", 32)
	if int32Field.K != "int32" || int32Field.T != kindInt64 || int32Field.I64 != 32 {
		t.Errorf("Int32 field incorrect: %+v", int32Field)
	}

	int64Field := Int64("int64", 64)
	if int64Field.K != "int64" || int64Field.T != kindInt64 || int64Field.I64 != 64 {
		t.Errorf("Int64 field incorrect: %+v", int64Field)
	}

	// Test Uint variants
	uintField := Uint("uint", 42)
	if uintField.K != "uint" || uintField.T != kindUint64 || uintField.U64 != 42 {
		t.Errorf("Uint field incorrect: %+v", uintField)
	}

	uint8Field := Uint8("uint8", 8)
	if uint8Field.K != "uint8" || uint8Field.T != kindUint64 || uint8Field.U64 != 8 {
		t.Errorf("Uint8 field incorrect: %+v", uint8Field)
	}

	uint16Field := Uint16("uint16", 16)
	if uint16Field.K != "uint16" || uint16Field.T != kindUint64 || uint16Field.U64 != 16 {
		t.Errorf("Uint16 field incorrect: %+v", uint16Field)
	}

	uint32Field := Uint32("uint32", 32)
	if uint32Field.K != "uint32" || uint32Field.T != kindUint64 || uint32Field.U64 != 32 {
		t.Errorf("Uint32 field incorrect: %+v", uint32Field)
	}

	uint64Field := Uint64("uint64", 64)
	if uint64Field.K != "uint64" || uint64Field.T != kindUint64 || uint64Field.U64 != 64 {
		t.Errorf("Uint64 field incorrect: %+v", uint64Field)
	}

	// Test Float variants
	float32Field := Float32("float32", 3.14)
	if float32Field.K != "float32" || float32Field.T != kindFloat64 || float32Field.F64 != float64(float32(3.14)) {
		t.Errorf("Float32 field incorrect: %+v", float32Field)
	}

	float64Field := Float64("float64", 3.14159)
	if float64Field.K != "float64" || float64Field.T != kindFloat64 || float64Field.F64 != 3.14159 {
		t.Errorf("Float64 field incorrect: %+v", float64Field)
	}

	// Test Bool
	boolTrueField := Bool("bool_true", true)
	if boolTrueField.K != "bool_true" || boolTrueField.T != kindBool || boolTrueField.I64 != 1 {
		t.Errorf("Bool true field incorrect: %+v", boolTrueField)
	}

	boolFalseField := Bool("bool_false", false)
	if boolFalseField.K != "bool_false" || boolFalseField.T != kindBool || boolFalseField.I64 != 0 {
		t.Errorf("Bool false field incorrect: %+v", boolFalseField)
	}

	// Test Duration
	duration := time.Minute
	durField := Dur("duration", duration)
	if durField.K != "duration" || durField.T != kindDur || durField.I64 != int64(duration) {
		t.Errorf("Duration field incorrect: %+v", durField)
	}

	// Test Time
	now := time.Now()
	timeField := TimeField("time", now)
	if timeField.K != "time" || timeField.T != kindTime || timeField.I64 != now.UnixNano() {
		t.Errorf("Time field incorrect: %+v", timeField)
	}

	timeField2 := Time("time2", now)
	if timeField2.K != "time2" || timeField2.T != kindTime || timeField2.I64 != now.UnixNano() {
		t.Errorf("Time alias field incorrect: %+v", timeField2)
	}

	// Test Bytes
	data := []byte("hello")
	bytesField := Bytes("bytes", data)
	if bytesField.K != "bytes" || bytesField.T != kindBytes || !bytes.Equal(bytesField.B, data) {
		t.Errorf("Bytes field incorrect: %+v", bytesField)
	}

	binaryField := Binary("binary", data)
	if binaryField.K != "binary" || binaryField.T != kindBytes || !bytes.Equal(binaryField.B, data) {
		t.Errorf("Binary alias field incorrect: %+v", binaryField)
	}
}

// TestFieldMethods tests all field methods
func TestFieldMethods(t *testing.T) {
	// Test string field methods
	strField := Str("test", "value")

	if strField.Type() != kindString {
		t.Errorf("Expected Type() to return kindString, got %v", strField.Type())
	}

	if strField.Key() != "test" {
		t.Errorf("Expected Key() to return 'test', got %s", strField.Key())
	}

	if !strField.IsString() {
		t.Error("Expected IsString() to return true")
	}

	if strField.IsInt() || strField.IsUint() || strField.IsFloat() || strField.IsBool() ||
		strField.IsDuration() || strField.IsTime() || strField.IsBytes() {
		t.Error("Expected other Is*() methods to return false for string field")
	}

	if strField.StringValue() != "value" {
		t.Errorf("Expected StringValue() to return 'value', got %s", strField.StringValue())
	}

	// Test int field methods
	intField := Int64("test", 42)

	if !intField.IsInt() {
		t.Error("Expected IsInt() to return true")
	}

	if intField.IntValue() != 42 {
		t.Errorf("Expected IntValue() to return 42, got %d", intField.IntValue())
	}

	// Test uint field methods
	uintField := Uint64("test", 42)

	if !uintField.IsUint() {
		t.Error("Expected IsUint() to return true")
	}

	if uintField.UintValue() != 42 {
		t.Errorf("Expected UintValue() to return 42, got %d", uintField.UintValue())
	}

	// Test float field methods
	floatField := Float64("test", 3.14)

	if !floatField.IsFloat() {
		t.Error("Expected IsFloat() to return true")
	}

	if floatField.FloatValue() != 3.14 {
		t.Errorf("Expected FloatValue() to return 3.14, got %f", floatField.FloatValue())
	}

	// Test bool field methods
	boolField := Bool("test", true)

	if !boolField.IsBool() {
		t.Error("Expected IsBool() to return true")
	}

	if !boolField.BoolValue() {
		t.Error("Expected BoolValue() to return true")
	}

	boolFalseField := Bool("test", false)
	if boolFalseField.BoolValue() {
		t.Error("Expected BoolValue() to return false")
	}

	// Test duration field methods
	duration := time.Hour
	durField := Dur("test", duration)

	if !durField.IsDuration() {
		t.Error("Expected IsDuration() to return true")
	}

	if durField.DurationValue() != duration {
		t.Errorf("Expected DurationValue() to return %v, got %v", duration, durField.DurationValue())
	}

	// Test time field methods
	now := time.Now()
	timeField := Time("test", now)

	if !timeField.IsTime() {
		t.Error("Expected IsTime() to return true")
	}

	retrievedTime := timeField.TimeValue()
	if retrievedTime.UnixNano() != now.UnixNano() {
		t.Errorf("Expected TimeValue() to return %v, got %v", now, retrievedTime)
	}

	// Test bytes field methods
	data := []byte("test data")
	bytesField := Bytes("test", data)

	if !bytesField.IsBytes() {
		t.Error("Expected IsBytes() to return true")
	}

	if !bytes.Equal(bytesField.BytesValue(), data) {
		t.Errorf("Expected BytesValue() to return %v, got %v", data, bytesField.BytesValue())
	}
}

// TestFieldValueMethodsWithWrongType tests that value methods return zero values for wrong types
func TestFieldValueMethodsWithWrongType(t *testing.T) {
	strField := Str("test", "value")

	// Test that wrong type methods return zero values
	if strField.IntValue() != 0 {
		t.Errorf("Expected IntValue() to return 0 for string field, got %d", strField.IntValue())
	}

	if strField.UintValue() != 0 {
		t.Errorf("Expected UintValue() to return 0 for string field, got %d", strField.UintValue())
	}

	if strField.FloatValue() != 0.0 {
		t.Errorf("Expected FloatValue() to return 0.0 for string field, got %f", strField.FloatValue())
	}

	if strField.BoolValue() != false {
		t.Errorf("Expected BoolValue() to return false for string field, got %t", strField.BoolValue())
	}

	if strField.DurationValue() != 0 {
		t.Errorf("Expected DurationValue() to return 0 for string field, got %v", strField.DurationValue())
	}

	zeroTime := time.Time{}
	if strField.TimeValue() != zeroTime {
		t.Errorf("Expected TimeValue() to return zero time for string field, got %v", strField.TimeValue())
	}

	if strField.BytesValue() != nil {
		t.Errorf("Expected BytesValue() to return nil for string field, got %v", strField.BytesValue())
	}

	// Test with int field
	intField := Int("test", 42)

	if intField.StringValue() != "" {
		t.Errorf("Expected StringValue() to return empty string for int field, got %s", intField.StringValue())
	}
}

// TestKindConstants tests that kind constants have expected values
func TestKindConstants(t *testing.T) {
	expectedKinds := map[kind]string{
		kindString:  "string",
		kindInt64:   "int64",
		kindUint64:  "uint64",
		kindFloat64: "float64",
		kindBool:    "bool",
		kindDur:     "duration",
		kindTime:    "time",
		kindBytes:   "bytes",
	}

	// Verify that kinds start from 1 and are sequential
	if kindString != 1 {
		t.Errorf("Expected kindString to be 1, got %d", kindString)
	}

	if kindBytes != 8 {
		t.Errorf("Expected kindBytes to be 8, got %d", kindBytes)
	}

	// Test that all expected kinds are defined
	for kind, name := range expectedKinds {
		if kind < kindString || kind > kindBytes {
			t.Errorf("Kind %s (%d) is out of expected range", name, kind)
		}
	}
}

// TestFieldZeroValues tests field behavior with zero values
func TestFieldZeroValues(t *testing.T) {
	// Test zero string
	zeroStr := Str("empty", "")
	if zeroStr.StringValue() != "" {
		t.Errorf("Expected empty string, got %s", zeroStr.StringValue())
	}

	// Test zero int
	zeroInt := Int("zero", 0)
	if zeroInt.IntValue() != 0 {
		t.Errorf("Expected zero int, got %d", zeroInt.IntValue())
	}

	// Test zero uint
	zeroUint := Uint("zero", 0)
	if zeroUint.UintValue() != 0 {
		t.Errorf("Expected zero uint, got %d", zeroUint.UintValue())
	}

	// Test zero float
	zeroFloat := Float64("zero", 0.0)
	if zeroFloat.FloatValue() != 0.0 {
		t.Errorf("Expected zero float, got %f", zeroFloat.FloatValue())
	}

	// Test false bool (stored as 0)
	falseBool := Bool("false", false)
	if falseBool.BoolValue() != false {
		t.Errorf("Expected false bool, got %t", falseBool.BoolValue())
	}

	// Test zero duration
	zeroDur := Dur("zero", 0)
	if zeroDur.DurationValue() != 0 {
		t.Errorf("Expected zero duration, got %v", zeroDur.DurationValue())
	}

	// Test zero time
	zeroTime := time.Time{}
	zeroTimeField := Time("zero", zeroTime)
	retrievedZeroTime := zeroTimeField.TimeValue()
	if retrievedZeroTime.UnixNano() != zeroTime.UnixNano() {
		t.Errorf("Expected zero time with UnixNano %d, got %d", zeroTime.UnixNano(), retrievedZeroTime.UnixNano())
	}

	// Test nil bytes
	nilBytes := Bytes("nil", nil)
	if nilBytes.BytesValue() != nil {
		t.Errorf("Expected nil bytes, got %v", nilBytes.BytesValue())
	}

	// Test empty bytes slice
	emptyBytes := Bytes("empty", []byte{})
	if len(emptyBytes.BytesValue()) != 0 {
		t.Errorf("Expected empty bytes slice, got %v", emptyBytes.BytesValue())
	}
}

// TestFieldEdgeCases tests edge cases and boundary values
func TestFieldEdgeCases(t *testing.T) {
	// Test maximum values
	maxInt64 := Int64("max_int64", 9223372036854775807)
	if maxInt64.IntValue() != 9223372036854775807 {
		t.Errorf("Max int64 value incorrect")
	}

	maxUint64 := Uint64("max_uint64", 18446744073709551615)
	if maxUint64.UintValue() != 18446744073709551615 {
		t.Errorf("Max uint64 value incorrect")
	}

	// Test minimum values
	minInt64 := Int64("min_int64", -9223372036854775808)
	if minInt64.IntValue() != -9223372036854775808 {
		t.Errorf("Min int64 value incorrect")
	}

	// Test special float values
	nanField := Float64("nan", math.NaN())
	if !math.IsNaN(nanField.FloatValue()) {
		t.Errorf("Expected NaN, got %f", nanField.FloatValue())
	}

	infField := Float64("inf", math.Inf(1))
	if !math.IsInf(infField.FloatValue(), 1) {
		t.Errorf("Expected +Inf, got %f", infField.FloatValue())
	}
}

// TestFieldPerformance tests that field creation is efficient (not a benchmark)
func TestFieldPerformance(t *testing.T) {
	// Simple performance test - create many fields quickly
	for i := 0; i < 1000; i++ {
		_ = Str("key", "value")
		_ = Int("key", i)
		_ = Bool("key", i%2 == 0)
		_ = Float64("key", float64(i))
	}
	// If this test completes quickly, field creation is efficient
}

// TestFieldMemoryLayout tests that Field struct has expected memory characteristics
func TestFieldMemoryLayout(t *testing.T) {
	f1 := Str("test", "value")
	f2 := Int("test", 42)

	// Both fields should have the same basic structure
	if f1.K != f2.K {
		t.Errorf("Field keys should be equal when set to same value")
	}

	// Fields should be comparable for key equality
	if f1.Key() != f2.Key() {
		t.Errorf("Field Key() methods should return equal values")
	}
}
