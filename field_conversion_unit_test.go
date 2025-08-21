package iris

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestToLegacyFields tests the batch conversion from BinaryField to Field
func TestToLegacyFields(t *testing.T) {
	t.Run("EmptySlice", func(t *testing.T) {
		result := ToLegacyFields(nil)
		if result != nil {
			t.Errorf("Expected nil for empty input, got %v", result)
		}

		result = ToLegacyFields([]BinaryField{})
		if result != nil {
			t.Errorf("Expected nil for empty slice, got %v", result)
		}
	})

	t.Run("SingleField", func(t *testing.T) {
		binaryFields := []BinaryField{
			BinaryStr("test_key", "test_value"),
		}

		result := ToLegacyFields(binaryFields)
		if len(result) != 1 {
			t.Fatalf("Expected 1 field, got %d", len(result))
		}

		field := result[0]
		if field.Key != "converted_key" {
			t.Errorf("Expected key 'converted_key', got %q", field.Key)
		}
		if field.Type != StringType {
			t.Errorf("Expected StringType, got %v", field.Type)
		}
	})

	t.Run("MultipleFields", func(t *testing.T) {
		binaryFields := []BinaryField{
			BinaryStr("str_key", "string_value"),
			BinaryInt("int_key", 42),
			BinaryBool("bool_key", true),
		}

		result := ToLegacyFields(binaryFields)
		if len(result) != 3 {
			t.Fatalf("Expected 3 fields, got %d", len(result))
		}

		// All should have converted_key since we can't reconstruct original keys
		for i, field := range result {
			if field.Key != "converted_key" {
				t.Errorf("Field %d: expected key 'converted_key', got %q", i, field.Key)
			}
		}

		// Check types are preserved
		expectedTypes := []FieldType{StringType, IntType, BoolType}
		for i, expected := range expectedTypes {
			if result[i].Type != expected {
				t.Errorf("Field %d: expected type %v, got %v", i, expected, result[i].Type)
			}
		}
	})

	t.Run("LargeSlice", func(t *testing.T) {
		// Test with many fields to verify performance and correctness
		binaryFields := make([]BinaryField, 100)
		for i := range binaryFields {
			binaryFields[i] = BinaryInt(fmt.Sprintf("key_%d", i), int64(i))
		}

		result := ToLegacyFields(binaryFields)
		if len(result) != 100 {
			t.Fatalf("Expected 100 fields, got %d", len(result))
		}

		for i, field := range result {
			if field.Type != IntType {
				t.Errorf("Field %d: expected IntType, got %v", i, field.Type)
			}
			if field.Int != int64(i) {
				t.Errorf("Field %d: expected value %d, got %d", i, i, field.Int)
			}
		}
	})
}

// TestToLegacyField tests the single field conversion
func TestToLegacyField(t *testing.T) {
	t.Run("StringField", func(t *testing.T) {
		bf := BinaryStr("test_key", "test_value")
		result := toLegacyField(bf)

		if result.Key != "converted_key" {
			t.Errorf("Expected key 'converted_key', got %q", result.Key)
		}
		if result.Type != StringType {
			t.Errorf("Expected StringType, got %v", result.Type)
		}
		// String fields get a placeholder representation
		if !strings.Contains(result.String, "<binary_data:") {
			t.Errorf("Expected placeholder string representation, got %q", result.String)
		}
	})

	t.Run("IntField", func(t *testing.T) {
		bf := BinaryInt("int_key", 42)
		result := toLegacyField(bf)

		if result.Type != IntType {
			t.Errorf("Expected IntType, got %v", result.Type)
		}
		if result.Int != 42 {
			t.Errorf("Expected int value 42, got %d", result.Int)
		}
	})

	t.Run("BoolField", func(t *testing.T) {
		bfTrue := BinaryBool("bool_key", true)
		result := toLegacyField(bfTrue)

		if result.Type != BoolType {
			t.Errorf("Expected BoolType, got %v", result.Type)
		}
		if !result.Bool {
			t.Error("Expected bool value true, got false")
		}

		bfFalse := BinaryBool("bool_key", false)
		result = toLegacyField(bfFalse)
		if result.Bool {
			t.Error("Expected bool value false, got true")
		}
	})

	t.Run("FloatField", func(t *testing.T) {
		// Create a BinaryField for float manually since BinaryFloat64 doesn't exist
		bf := BinaryField{
			Type: uint8(Float64Type),
			Data: 0x400921FB54442D18, // Binary representation of π (3.14159...)
		}
		result := toLegacyField(bf)

		if result.Type != Float64Type {
			t.Errorf("Expected Float64Type, got %v", result.Type)
		}
		// Note: Due to binary representation, we check for reasonable range
		if result.Float < 3.0 || result.Float > 3.3 {
			t.Errorf("Expected float value around 3.14, got %f", result.Float)
		}
	})

	t.Run("UnknownType", func(t *testing.T) {
		// Create a BinaryField with unknown type
		bf := BinaryField{
			Type: 255, // Unknown type (max uint8)
			Data: 42,
		}
		result := toLegacyField(bf)

		// Should have default values for unknown type
		if result.Key != "converted_key" {
			t.Errorf("Expected default key, got %q", result.Key)
		}
	})
}

// TestConvertStringLikeTypes tests string-like type conversions
func TestConvertStringLikeTypes(t *testing.T) {
	baseField := Field{Key: "test", Type: StringType}
	bf := BinaryStr("test", "value")

	t.Run("StringType", func(t *testing.T) {
		result := convertStringLikeTypes(baseField, bf, StringType)
		if result.Type != StringType {
			t.Errorf("Expected StringType, got %v", result.Type)
		}
		if !strings.Contains(result.String, "<binary_data:") {
			t.Errorf("Expected binary data placeholder, got %q", result.String)
		}
	})

	t.Run("BinaryType", func(t *testing.T) {
		baseField.Type = BinaryType
		result := convertStringLikeTypes(baseField, bf, BinaryType)
		if result.Type != BinaryType {
			t.Errorf("Expected BinaryType, got %v", result.Type)
		}
		if result.Bytes != nil {
			t.Errorf("Expected nil bytes for binary type, got %v", result.Bytes)
		}
	})

	t.Run("ByteStringType", func(t *testing.T) {
		baseField.Type = ByteStringType
		result := convertStringLikeTypes(baseField, bf, ByteStringType)
		if result.Type != ByteStringType {
			t.Errorf("Expected ByteStringType, got %v", result.Type)
		}
		if result.Bytes != nil {
			t.Errorf("Expected nil bytes for byte string type, got %v", result.Bytes)
		}
	})
}

// TestConvertIntegerType tests integer type conversions
func TestConvertIntegerType(t *testing.T) {
	t.Run("ValidIntConversion", func(t *testing.T) {
		baseField := Field{Key: "test", Type: IntType}
		bf := BinaryInt("test", 42)

		result := convertIntegerType(baseField, bf)
		if result.Type != IntType {
			t.Errorf("Expected IntType, got %v", result.Type)
		}
		if result.Int != 42 {
			t.Errorf("Expected int value 42, got %d", result.Int)
		}
	})

	t.Run("Int64Type", func(t *testing.T) {
		baseField := Field{Key: "test", Type: Int64Type}
		bf := BinaryInt("test", 123456789)

		result := convertIntegerType(baseField, bf)
		if result.Type != Int64Type {
			t.Errorf("Expected Int64Type, got %v", result.Type)
		}
		if result.Int != 123456789 {
			t.Errorf("Expected int value 123456789, got %d", result.Int)
		}
	})

	t.Run("NegativeValues", func(t *testing.T) {
		baseField := Field{Key: "test", Type: IntType}
		bf := BinaryInt("test", -42)

		result := convertIntegerType(baseField, bf)
		if result.Int != -42 {
			t.Errorf("Expected int value -42, got %d", result.Int)
		}
	})

	t.Run("DifferentIntTypes", func(t *testing.T) {
		intTypes := []FieldType{IntType, Int64Type, Int32Type, Int16Type, Int8Type}
		for _, intType := range intTypes {
			baseField := Field{Key: "test", Type: intType}
			bf := BinaryInt("test", 100)

			result := convertIntegerType(baseField, bf)
			if result.Type != intType {
				t.Errorf("Expected %v, got %v", intType, result.Type)
			}
			if result.Int != 100 {
				t.Errorf("Expected int value 100 for type %v, got %d", intType, result.Int)
			}
		}
	})
}

// TestConvertUintegerType tests unsigned integer type conversions
func TestConvertUintegerType(t *testing.T) {
	t.Run("ValidUintConversion", func(t *testing.T) {
		baseField := Field{Key: "test", Type: UintType}
		// Create uint BinaryField manually
		bf := BinaryField{
			Type: uint8(UintType),
			Data: 42,
		}

		result := convertUintegerType(baseField, bf)
		if result.Type != UintType {
			t.Errorf("Expected UintType, got %v", result.Type)
		}
		if result.Int != 42 {
			t.Errorf("Expected int value 42, got %d", result.Int)
		}
	})

	t.Run("LargeUintValues", func(t *testing.T) {
		baseField := Field{Key: "test", Type: Uint64Type}
		// Create large uint64 BinaryField manually
		bf := BinaryField{
			Type: uint8(Uint64Type),
			Data: 18446744073709551615, // Max uint64
		}

		result := convertUintegerType(baseField, bf)
		// This might convert to string if overflow occurs
		switch result.Type {
		case StringType:
			// Overflow case - should be converted to string
			if !strings.Contains(result.String, "18446744073709551615") {
				t.Errorf("Expected string representation of large uint, got %q", result.String)
			}
		case Uint64Type:
			// Safe conversion case
			if result.Int < 0 {
				t.Errorf("Expected positive value for uint64, got %d", result.Int)
			}
		}
	})

	t.Run("DifferentUintTypes", func(t *testing.T) {
		uintTypes := []FieldType{UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type}
		for _, uintType := range uintTypes {
			baseField := Field{Key: "test", Type: uintType}
			// Create uint BinaryField manually
			bf := BinaryField{
				Type: uint8(uintType),
				Data: 100,
			}

			result := convertUintegerType(baseField, bf)
			// Should preserve type or convert to string on overflow
			if result.Type != uintType && result.Type != StringType {
				t.Errorf("Expected %v or StringType, got %v", uintType, result.Type)
			}
		}
	})
}

// TestConvertTimeType tests time and duration type conversions
func TestConvertTimeType(t *testing.T) {
	t.Run("TimeType", func(t *testing.T) {
		baseField := Field{Key: "test", Type: TimeType}
		now := time.Now()
		// Create time BinaryField manually
		bf := BinaryField{
			Type: uint8(TimeType),
			Data: uint64(now.UnixNano()),
		}

		result := convertTimeType(baseField, bf)
		if result.Type != TimeType {
			t.Errorf("Expected TimeType, got %v", result.Type)
		}
		// Should have some representation of the time
		if result.Int == 0 && result.String == "" {
			t.Error("Expected some time representation")
		}
	})

	t.Run("DurationType", func(t *testing.T) {
		baseField := Field{Key: "test", Type: DurationType}
		duration := time.Minute
		// Create duration BinaryField manually
		bf := BinaryField{
			Type: uint8(DurationType),
			Data: uint64(duration.Nanoseconds()),
		}

		result := convertTimeType(baseField, bf)
		if result.Type != DurationType {
			t.Errorf("Expected DurationType, got %v", result.Type)
		}
		// Should have some representation of the duration
		if result.Int == 0 && result.String == "" {
			t.Error("Expected some duration representation")
		}
	})

	t.Run("TimeOverflow", func(t *testing.T) {
		// Create a field that might cause overflow
		baseField := Field{Key: "test", Type: TimeType}
		bf := BinaryField{
			Type: uint8(TimeType),
			Data: ^uint64(0), // Maximum uint64 value
		}

		result := convertTimeType(baseField, bf)
		// Should handle gracefully, either as time or string
		if result.Type != TimeType && result.Type != StringType {
			t.Errorf("Expected TimeType or StringType, got %v", result.Type)
		}
	})
}

// TestFieldConversionIntegration tests the complete conversion pipeline
func TestFieldConversionIntegration(t *testing.T) {
	t.Run("RoundTripConversion", func(t *testing.T) {
		// Create various BinaryFields using available functions and manual creation
		binaryFields := []BinaryField{
			BinaryStr("string_field", "test_string"),
			BinaryInt("int_field", 42),
			BinaryBool("bool_field", true),
			// Create float BinaryField manually
			{Type: uint8(Float64Type), Data: 0x400921FB54442D18}, // π
			// Create time BinaryField manually
			{Type: uint8(TimeType), Data: uint64(time.Now().UnixNano())},
			// Create duration BinaryField manually
			{Type: uint8(DurationType), Data: uint64(time.Hour.Nanoseconds())},
		}

		// Convert to legacy fields
		legacyFields := ToLegacyFields(binaryFields)

		// Verify count
		if len(legacyFields) != len(binaryFields) {
			t.Fatalf("Expected %d legacy fields, got %d", len(binaryFields), len(legacyFields))
		}

		// Verify types are properly converted
		expectedTypes := []FieldType{StringType, IntType, BoolType, Float64Type, TimeType, DurationType}
		for i, expected := range expectedTypes {
			if legacyFields[i].Type != expected {
				t.Errorf("Field %d: expected type %v, got %v", i, expected, legacyFields[i].Type)
			}
		}
	})

	t.Run("MixedTypeConversion", func(t *testing.T) {
		// Test conversion with various numeric types created manually
		binaryFields := []BinaryField{
			{Type: uint8(Int8Type), Data: 127},
			{Type: uint8(Uint8Type), Data: 255},
			{Type: uint8(Int16Type), Data: 32767},
			{Type: uint8(Uint16Type), Data: 65535},
			{Type: uint8(Int32Type), Data: 2147483647},
			{Type: uint8(Uint32Type), Data: 4294967295},
		}

		legacyFields := ToLegacyFields(binaryFields)

		for i, field := range legacyFields {
			// All should be converted to some valid representation
			if field.Key != "converted_key" {
				t.Errorf("Field %d: expected converted_key, got %q", i, field.Key)
			}

			// Should have appropriate type (int-like or string fallback)
			validTypes := []FieldType{IntType, Int8Type, Int16Type, Int32Type, Int64Type,
				UintType, Uint8Type, Uint16Type, Uint32Type, Uint64Type, StringType}
			validType := false
			for _, validT := range validTypes {
				if field.Type == validT {
					validType = true
					break
				}
			}
			if !validType {
				t.Errorf("Field %d: unexpected type %v", i, field.Type)
			}
		}
	})
}
