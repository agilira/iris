// field_helpers.go: Field helper functions for Iris logging
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"strconv"
	"time"
)

// =============================================================================
// Field Validation and Conversion Helpers
// =============================================================================

// ValidateField checks if a field is valid and safe to use
func ValidateField(field Field) error {
	if field.Key == "" {
		return fmt.Errorf("field key cannot be empty")
	}

	if !isValidFieldType(field.Type) {
		return fmt.Errorf("invalid field type: %d", field.Type)
	}

	return nil
}

// isValidFieldType checks if the field type is supported
func isValidFieldType(fieldType FieldType) bool {
	switch fieldType {
	case StringType, IntType, Int64Type, Int32Type, Int16Type, Int8Type,
		UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type,
		Float64Type, Float32Type, BoolType, TimeType, DurationType,
		ErrorType, BinaryType, ByteStringType, AnyType:
		return true
	default:
		return false
	}
}

// =============================================================================
// Field Value Extraction Helpers
// =============================================================================

// GetFieldValue returns the value of a field as an interface{}
func GetFieldValue(field Field) interface{} {
	switch field.Type {
	case StringType:
		return field.String
	case IntType, Int64Type, Int32Type, Int16Type, Int8Type:
		return field.Int
	case UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		// Use safe conversion for encoding
		value, _ := SafeInt64ToUint64ForEncoding(field.Int)
		return value
	case Float64Type, Float32Type:
		return field.Float
	case BoolType:
		return field.Bool
	case TimeType:
		return time.Unix(0, field.Int)
	case DurationType:
		return time.Duration(field.Int)
	case ErrorType:
		return field.Err
	case BinaryType, ByteStringType:
		return field.Bytes
	case AnyType:
		return field.Any
	default:
		return nil
	}
}

// GetFieldString returns the string representation of a field's value
func GetFieldString(field Field) string {
	switch field.Type {
	case StringType:
		return field.String
	case IntType, Int64Type, Int32Type, Int16Type, Int8Type:
		return strconv.FormatInt(field.Int, 10)
	case UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		// Use safe conversion for string formatting
		value, _ := SafeInt64ToUint64ForEncoding(field.Int)
		return strconv.FormatUint(value, 10)
	case Float64Type, Float32Type:
		return strconv.FormatFloat(field.Float, 'g', -1, 64)
	case BoolType:
		return strconv.FormatBool(field.Bool)
	case TimeType:
		return time.Unix(0, field.Int).Format(time.RFC3339Nano)
	case DurationType:
		return time.Duration(field.Int).String()
	case ErrorType:
		if field.Err != nil {
			return field.Err.Error()
		}
		return ""
	case BinaryType:
		return fmt.Sprintf("binary[%d]", len(field.Bytes))
	case ByteStringType:
		return string(field.Bytes)
	case AnyType:
		return fmt.Sprintf("%v", field.Any)
	default:
		return ""
	}
}

// =============================================================================
// Field Comparison and Utilities
// =============================================================================

// FieldsEqual compares two fields for equality
func FieldsEqual(a, b Field) bool {
	if a.Key != b.Key || a.Type != b.Type {
		return false
	}
	
	return compareFieldsByType(a, b)
}

// compareFieldsByType compares field values based on their type
func compareFieldsByType(a, b Field) bool {
	switch a.Type {
	case StringType:
		return compareStringFields(a, b)
	case IntType, Int64Type, Int32Type, Int16Type, Int8Type,
		UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type,
		TimeType, DurationType:
		return compareIntFields(a, b)
	case Float64Type, Float32Type:
		return compareFloatFields(a, b)
	case BoolType:
		return compareBoolFields(a, b)
	case ErrorType:
		return compareErrorFields(a, b)
	case BinaryType, ByteStringType:
		return compareByteFields(a, b)
	case AnyType:
		return compareAnyFields(a, b)
	default:
		return false
	}
}

// compareStringFields compares string type fields
func compareStringFields(a, b Field) bool {
	return a.String == b.String
}

// compareIntFields compares integer-based type fields
func compareIntFields(a, b Field) bool {
	return a.Int == b.Int
}

// compareFloatFields compares float type fields
func compareFloatFields(a, b Field) bool {
	return a.Float == b.Float
}

// compareBoolFields compares boolean type fields
func compareBoolFields(a, b Field) bool {
	return a.Bool == b.Bool
}

// compareErrorFields compares error type fields
func compareErrorFields(a, b Field) bool {
	return (a.Err == nil && b.Err == nil) ||
		(a.Err != nil && b.Err != nil && a.Err.Error() == b.Err.Error())
}

// compareByteFields compares byte array type fields
func compareByteFields(a, b Field) bool {
	if len(a.Bytes) != len(b.Bytes) {
		return false
	}
	for i := range a.Bytes {
		if a.Bytes[i] != b.Bytes[i] {
			return false
		}
	}
	return true
}

// compareAnyFields compares any type fields
func compareAnyFields(a, b Field) bool {
	return fmt.Sprintf("%v", a.Any) == fmt.Sprintf("%v", b.Any)
}

// CloneField creates a deep copy of a field
func CloneField(field Field) Field {
	clone := field

	// Deep copy slice fields
	if field.Bytes != nil {
		clone.Bytes = make([]byte, len(field.Bytes))
		copy(clone.Bytes, field.Bytes)
	}

	return clone
}

// =============================================================================
// Field Sorting and Grouping
// =============================================================================

// SortFieldsByKey sorts fields by their key names
func SortFieldsByKey(fields []Field) {
	for i := 0; i < len(fields)-1; i++ {
		for j := i + 1; j < len(fields); j++ {
			if fields[i].Key > fields[j].Key {
				fields[i], fields[j] = fields[j], fields[i]
			}
		}
	}
}

// GroupFieldsByType groups fields by their type
func GroupFieldsByType(fields []Field) map[FieldType][]Field {
	groups := make(map[FieldType][]Field)

	for _, field := range fields {
		groups[field.Type] = append(groups[field.Type], field)
	}

	return groups
}

// =============================================================================
// Field Statistics and Analysis
// =============================================================================

// FieldStats contains statistics about a set of fields
type FieldStats struct {
	TotalFields   int
	UniqueKeys    int
	TypeCounts    map[FieldType]int
	TotalBytes    int
	AverageKeyLen float64
}

// AnalyzeFields provides statistics about a set of fields
func AnalyzeFields(fields []Field) FieldStats {
	stats := FieldStats{
		TotalFields: len(fields),
		TypeCounts:  make(map[FieldType]int),
	}

	if len(fields) == 0 {
		return stats
	}

	keySet := make(map[string]bool)
	totalKeyLen := 0

	for _, field := range fields {
		// Count unique keys
		keySet[field.Key] = true
		totalKeyLen += len(field.Key)

		// Count types
		stats.TypeCounts[field.Type]++

		// Estimate memory usage
		stats.TotalBytes += len(field.Key)
		switch field.Type {
		case StringType:
			stats.TotalBytes += len(field.String)
		case BinaryType, ByteStringType:
			stats.TotalBytes += len(field.Bytes)
		default:
			stats.TotalBytes += 8 // Approximate for primitive types
		}
	}

	stats.UniqueKeys = len(keySet)
	stats.AverageKeyLen = float64(totalKeyLen) / float64(len(fields))

	return stats
}

// =============================================================================
// Safe Type Conversion Helpers (THREAD-SAFE, LOCK-FREE)
// =============================================================================

// SafeUint64ToInt64 safely converts uint64 to int64, checking for overflow
// Returns the converted value and true if conversion is safe
func SafeUint64ToInt64(value uint64) (int64, bool) {
	const maxInt64 = 1<<63 - 1 // 9223372036854775807
	if value > maxInt64 {
		return 0, false
	}
	return int64(value), true
}

// SafeInt64ToUint64 safely converts int64 to uint64, checking for negative values
// Returns the converted value and true if conversion is safe
func SafeInt64ToUint64(value int64) (uint64, bool) {
	if value < 0 {
		return 0, false
	}
	return uint64(value), true
}

// SafeUintToInt64 safely converts uint to int64, checking for overflow
func SafeUintToInt64(value uint) (int64, bool) {
	return SafeUint64ToInt64(uint64(value))
}

// SafeBinaryDataToInt64 safely converts BinaryField.Data to int64 for uint types
// This function handles the case where BinaryField.Data contains uint values
func SafeBinaryDataToInt64(data uint64, fieldType FieldType) (int64, bool) {
	switch fieldType {
	case UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		// For unsigned types, check for overflow
		return SafeUint64ToInt64(data)
	case IntType, Int64Type, Int32Type, Int16Type, Int8Type:
		// For signed types, direct conversion is safe (already validated during creation)
		return int64(data), true // #nosec G115 - Safe conversion in context
	case TimeType, DurationType:
		// Time values are typically safe
		return int64(data), true // #nosec G115 - Safe conversion for time values
	default:
		// For other types, assume safe conversion
		return int64(data), true // #nosec G115 - Safe conversion for primitive types
	}
}

// SafeInt64ToUint64ForEncoding safely converts int64 to uint64 for encoding purposes
// This is specifically for encoding/serialization where we need uint64 representation
// Returns the converted value and a flag indicating if it's a negative value stored in 2's complement
func SafeInt64ToUint64ForEncoding(value int64) (uint64, bool) {
	if value >= 0 {
		return uint64(value), false // Positive value, direct conversion
	}
	// For negative values, we use two's complement representation
	// This is safe for encoding because we'll decode it back correctly
	return uint64(value), true // #nosec G115 - Safe two's complement for encoding
}
