// field.go: Structured fields for Iris logging (ULTRA-OPTIMIZED)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"time"
)

// FieldType represents the type of a field value (optimized enum)
type FieldType uint8

// OPTIMIZATION: Grouped by frequency of use (String/Int most common)
const (
	StringType FieldType = iota
	IntType
	BoolType    // Moved up - 3rd most common
	Float64Type // Moved up - 4th most common
	Int64Type
	Int32Type
	Int16Type
	Int8Type
	UintType
	Uint64Type
	Uint32Type
	Uint16Type
	Uint8Type
	Float32Type
	DurationType
	TimeType
	ErrorType
	ByteStringType
	BinaryType
	AnyType
)

// Field represents a structured log field (memory-optimized layout)
type Field struct {
	Key    string      // Most common access - first for cache efficiency
	Type   FieldType   // Enum type - compact
	String string      // String value - most common type
	Int    int64       // Integer value - second most common
	Float  float64     // Float value - third most common
	Bool   bool        // Boolean value - fourth most common
	Err    error       // Error interface
	Bytes  []byte      // For ByteString and Binary types
	Any    interface{} // For Any type - least common, last
}

// String creates a string field (ULTRA-OPTIMIZED hot path)
//
//go:inline
func String(key, value string) Field {
	return Field{
		Key:    key,
		Type:   StringType,
		String: value,
	}
}

// Str is an alias for String for brevity (ULTRA-OPTIMIZED hot path)
//
//go:inline
func Str(key, value string) Field {
	return Field{
		Key:    key,
		Type:   StringType,
		String: value,
	}
}

// Int creates an int field (ULTRA-OPTIMIZED hot path)
//
//go:inline
func Int(key string, value int) Field {
	return Field{
		Key:  key,
		Type: IntType,
		Int:  int64(value),
	}
}

// Int64 creates an int64 field (OPTIMIZED)
//
//go:inline
func Int64(key string, value int64) Field {
	return Field{
		Key:  key,
		Type: Int64Type,
		Int:  value,
	}
}

// Bool creates a boolean field (ULTRA-OPTIMIZED hot path)
//
//go:inline
func Bool(key string, value bool) Field {
	return Field{
		Key:  key,
		Type: BoolType,
		Bool: value,
	}
}

// Float64 creates a float64 field (ULTRA-OPTIMIZED hot path)
//
//go:inline
func Float64(key string, value float64) Field {
	return Field{
		Key:   key,
		Type:  Float64Type,
		Float: value,
	}
}

// Float is an alias for Float64 (OPTIMIZED)
//
//go:inline
func Float(key string, value float64) Field {
	return Float64(key, value)
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{
		Key:  key,
		Type: DurationType,
		Int:  int64(value),
	}
}

// Time creates a time field
func Time(key string, value time.Time) Field {
	return Field{
		Key:  key,
		Type: TimeType,
		Int:  value.UnixNano(),
	}
}

// Error creates an error field
func Error(err error) Field {
	return Field{
		Key:  "error",
		Type: ErrorType,
		Err:  err,
	}
}

// Err is an alias for Error for brevity
func Err(err error) Field {
	return Error(err)
}

// Integer variants for full Zap compatibility

// Int32 creates an int32 field
func Int32(key string, value int32) Field {
	return Field{Key: key, Type: Int32Type, Int: int64(value)}
}

// Int16 creates an int16 field
func Int16(key string, value int16) Field {
	return Field{Key: key, Type: Int16Type, Int: int64(value)}
}

// Int8 creates an int8 field
func Int8(key string, value int8) Field {
	return Field{Key: key, Type: Int8Type, Int: int64(value)}
}

// Uint creates a uint field
func Uint(key string, value uint) Field {
	return Field{Key: key, Type: UintType, Int: int64(value)}
}

// Uint64 creates a uint64 field
func Uint64(key string, value uint64) Field {
	return Field{Key: key, Type: Uint64Type, Int: int64(value)}
}

// Uint32 creates a uint32 field
func Uint32(key string, value uint32) Field {
	return Field{Key: key, Type: Uint32Type, Int: int64(value)}
}

// Uint16 creates a uint16 field
func Uint16(key string, value uint16) Field {
	return Field{Key: key, Type: Uint16Type, Int: int64(value)}
}

// Uint8 creates a uint8 field
func Uint8(key string, value uint8) Field {
	return Field{Key: key, Type: Uint8Type, Int: int64(value)}
}

// Float32 creates a float32 field
func Float32(key string, value float32) Field {
	return Field{Key: key, Type: Float32Type, Float: float64(value)}
}

// ByteString creates a field that uses the string value of a byte slice
func ByteString(key string, value []byte) Field {
	return Field{Key: key, Type: ByteStringType, Bytes: value}
}

// Binary creates a field from binary data (base64 encoded when displayed)
// Binary creates a binary field
func Binary(key string, value []byte) Field {
	return Field{Key: key, Type: BinaryType, Bytes: value}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Type: AnyType, Any: value}
}

// =============================================================================
// Next-Generation API - Step 1.1 Minimal Implementation
// =============================================================================

// NextStr creates a safe string field using BinaryField
// Step 1.1: Minimal implementation for compilation stability
func NextStr(key, value string) BinaryField {
	return BinaryField{
		KeyPtr: 0, // Safe: no pointer for now
		KeyLen: uint16(len(key)),
		Type:   uint8(StringType),
		Data:   uint64(len(value)), // Store just the length for now
	}
}

// NextInt creates a safe int field using BinaryField
func NextInt(key string, value int) BinaryField {
	return BinaryField{
		KeyPtr: 0, // Safe: no pointer for now
		KeyLen: uint16(len(key)),
		Type:   uint8(IntType),
		Data:   uint64(value),
	}
}

// NextBool creates a safe bool field using BinaryField
func NextBool(key string, value bool) BinaryField {
	var data uint64
	if value {
		data = 1
	}
	return BinaryField{
		KeyPtr: 0, // Safe: no pointer for now
		KeyLen: uint16(len(key)),
		Type:   uint8(BoolType),
		Data:   data,
	}
}

// ToLegacyFields converts BinaryField slice back to legacy Field slice
// Step 1.1: Basic conversion for compilation stability
func ToLegacyFields(binaryFields []BinaryField) []Field {
	if len(binaryFields) == 0 {
		return nil
	}

	fields := make([]Field, len(binaryFields))
	for i, bf := range binaryFields {
		fields[i] = toLegacyField(bf)
	}
	return fields
}

// toLegacyField converts a single BinaryField back to legacy Field
// Step 1.1: Basic conversion without unsafe pointers
func toLegacyField(bf BinaryField) Field {
	// Create a dummy key since we don't store the actual key yet
	key := "converted_key"

	switch FieldType(bf.Type) {
	case StringType:
		return Field{Key: key, Type: StringType, String: "converted_string"}
	case IntType:
		return Field{Key: key, Type: IntType, Int: int64(bf.Data)}
	case BoolType:
		return Field{Key: key, Type: BoolType, Bool: bf.Data == 1}
	default:
		return Field{Key: key, Type: StringType, String: "unknown"}
	}
}
