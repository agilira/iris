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
		return uint64(field.Int)
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
		return strconv.FormatUint(uint64(field.Int), 10)
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

	switch a.Type {
	case StringType:
		return a.String == b.String
	case IntType, Int64Type, Int32Type, Int16Type, Int8Type,
		UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type,
		TimeType, DurationType:
		return a.Int == b.Int
	case Float64Type, Float32Type:
		return a.Float == b.Float
	case BoolType:
		return a.Bool == b.Bool
	case ErrorType:
		return (a.Err == nil && b.Err == nil) ||
			(a.Err != nil && b.Err != nil && a.Err.Error() == b.Err.Error())
	case BinaryType, ByteStringType:
		if len(a.Bytes) != len(b.Bytes) {
			return false
		}
		for i := range a.Bytes {
			if a.Bytes[i] != b.Bytes[i] {
				return false
			}
		}
		return true
	case AnyType:
		return fmt.Sprintf("%v", a.Any) == fmt.Sprintf("%v", b.Any)
	default:
		return false
	}
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
