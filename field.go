// field.go: Structured fields for Iris logging
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"time"
)

// FieldType represents the type of a field value
type FieldType uint8

const (
	StringType FieldType = iota
	IntType
	Int64Type
	Int32Type
	Int16Type
	Int8Type
	UintType
	Uint64Type
	Uint32Type
	Uint16Type
	Uint8Type
	Float64Type
	Float32Type
	BoolType
	DurationType
	TimeType
	ErrorType
	ByteStringType
	BinaryType
	AnyType
)

// Field represents a structured log field
type Field struct {
	Key    string
	Type   FieldType
	String string
	Int    int64
	Float  float64
	Bool   bool
	Err    error
	Bytes  []byte      // For ByteString and Binary types
	Any    interface{} // For Any type
}

// String creates a string field
func String(key, value string) Field {
	return Field{
		Key:    key,
		Type:   StringType,
		String: value,
	}
}

// Str is an alias for String for brevity
func Str(key, value string) Field {
	return String(key, value)
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{
		Key:  key,
		Type: IntType,
		Int:  int64(value),
	}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{
		Key:  key,
		Type: Int64Type,
		Int:  value,
	}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{
		Key:   key,
		Type:  Float64Type,
		Float: value,
	}
}

// Float is an alias for Float64
func Float(key string, value float64) Field {
	return Float64(key, value)
}

// Bool creates a boolean field
func Bool(key string, value bool) Field {
	return Field{
		Key:  key,
		Type: BoolType,
		Bool: value,
	}
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
	return Field{
		Key:  key,
		Type: Int32Type,
		Int:  int64(value),
	}
}

// Int16 creates an int16 field
func Int16(key string, value int16) Field {
	return Field{
		Key:  key,
		Type: Int16Type,
		Int:  int64(value),
	}
}

// Int8 creates an int8 field
func Int8(key string, value int8) Field {
	return Field{
		Key:  key,
		Type: Int8Type,
		Int:  int64(value),
	}
}

// Uint creates a uint field
func Uint(key string, value uint) Field {
	return Field{
		Key:  key,
		Type: UintType,
		Int:  int64(value),
	}
}

// Uint64 creates a uint64 field
func Uint64(key string, value uint64) Field {
	return Field{
		Key:  key,
		Type: Uint64Type,
		Int:  int64(value),
	}
}

// Uint32 creates a uint32 field
func Uint32(key string, value uint32) Field {
	return Field{
		Key:  key,
		Type: Uint32Type,
		Int:  int64(value),
	}
}

// Uint16 creates a uint16 field
func Uint16(key string, value uint16) Field {
	return Field{
		Key:  key,
		Type: Uint16Type,
		Int:  int64(value),
	}
}

// Uint8 creates a uint8 field
func Uint8(key string, value uint8) Field {
	return Field{
		Key:  key,
		Type: Uint8Type,
		Int:  int64(value),
	}
}

// Float32 creates a float32 field
func Float32(key string, value float32) Field {
	return Field{
		Key:   key,
		Type:  Float32Type,
		Float: float64(value),
	}
}

// ByteString creates a field from a byte slice, displayed as a string
func ByteString(key string, value []byte) Field {
	return Field{
		Key:   key,
		Type:  ByteStringType,
		Bytes: value,
	}
}

// Binary creates a field from binary data (base64 encoded when displayed)
func Binary(key string, value []byte) Field {
	return Field{
		Key:   key,
		Type:  BinaryType,
		Bytes: value,
	}
}

// Any creates a field that can hold any value (uses reflection for encoding)
func Any(key string, value interface{}) Field {
	return Field{
		Key:  key,
		Type: AnyType,
		Any:  value,
	}
}
