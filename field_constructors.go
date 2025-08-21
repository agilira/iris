// field_constructors.go: Field constructor functions for Iris logging
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"time"
)

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
	if safeValue, ok := SafeUintToInt64(value); ok {
		return Field{Key: key, Type: UintType, Int: safeValue}
	}
	// Fallback to string representation for very large values
	return Field{Key: key, Type: StringType, String: fmt.Sprintf("%d", value)}
}

// Uint64 creates a uint64 field
func Uint64(key string, value uint64) Field {
	if safeValue, ok := SafeUint64ToInt64(value); ok {
		return Field{Key: key, Type: Uint64Type, Int: safeValue}
	}
	// Fallback to string representation for very large values (> max int64)
	return Field{Key: key, Type: StringType, String: fmt.Sprintf("%d", value)}
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
func Binary(key string, value []byte) Field {
	return Field{Key: key, Type: BinaryType, Bytes: value}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Type: AnyType, Any: value}
}

// Secret creates a field for sensitive data that will be automatically redacted
// This is a CRITICAL SECURITY feature for preventing data leakage in logs
func Secret(key, value string) Field {
	return Field{
		Key:    key,
		Type:   SecretType,
		String: value, // Store the actual value internally
	}
}

// SecretAny creates a secret field for any sensitive data type
func SecretAny(key string, value interface{}) Field {
	return Field{
		Key:  key,
		Type: SecretType,
		Any:  value, // Store the actual value internally
	}
}
