// field.go: High-performance structured logging fields for Iris logging library
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import "time"

// kind represents the type of data stored in a Field.
// Using uint8 for compact memory layout and fast comparisons.
type kind uint8

// Field type constants for efficient type checking and serialization.
// Each constant represents a different data type that can be stored in a Field.
const (
	// kindString represents string data
	kindString kind = iota + 1
	// kindInt64 represents signed 64-bit integer data
	kindInt64
	// kindUint64 represents unsigned 64-bit integer data
	kindUint64
	// kindFloat64 represents 64-bit floating-point data
	kindFloat64
	// kindBool represents boolean data
	kindBool
	// kindDur represents time.Duration data
	kindDur
	// kindTime represents time.Time data
	kindTime
	// kindBytes represents byte slice data
	kindBytes
	// kindSecret represents sensitive data that should be redacted
	kindSecret
	// kindError represents error data
	kindError
	// kindStringer represents fmt.Stringer data
	kindStringer
	// kindObject represents arbitrary object data (interface{})
	kindObject
)

// Field represents a key-value pair with type information for structured logging.
// It uses a union-like approach to minimize memory allocation and maximize performance.
// The T field indicates which of the value fields (I64, U64, F64, Str, B, Obj) contains the actual data.
type Field struct {
	// K is the field key/name
	K string
	// T indicates the type of data stored in this field
	T kind
	// I64 stores signed integers, bools (as 0/1), durations, and timestamps
	I64 int64
	// U64 stores unsigned integers
	U64 uint64
	// F64 stores floating-point numbers
	F64 float64
	// Str stores string values
	Str string
	// B stores byte slices
	B []byte
	// Obj stores arbitrary objects (errors, stringers, etc.)
	Obj interface{}
}

// Str creates a string field for logging.
// This is one of the most commonly used field types.
func Str(k, v string) Field { return Field{K: k, T: kindString, Str: v} }

// Secret creates a field for sensitive data that will be automatically redacted.
// The actual value is stored but will appear as "[REDACTED]" in log output.
// Use this for passwords, API keys, tokens, personal data, or any sensitive information.
//
// Example:
//
//	logger.Info("User login", iris.Secret("password", userPassword))
//	// Output: {"level":"info","msg":"User login","password":"[REDACTED]"}
//
// Security: This prevents accidental exposure of sensitive data in logs while
// maintaining the field structure for debugging purposes.
func Secret(k, v string) Field { return Field{K: k, T: kindSecret, Str: v} }

// Int creates a signed integer field from an int value.
// The int is converted to int64 for consistent storage.
func Int(k string, v int) Field { return Field{K: k, T: kindInt64, I64: int64(v)} }

// Int64 creates a signed 64-bit integer field.
// Use this for large integers or when you specifically need int64.
func Int64(k string, v int64) Field { return Field{K: k, T: kindInt64, I64: v} }

// Uint64 creates an unsigned 64-bit integer field.
// Use this for non-negative values that may exceed int64 range.
func Uint64(k string, v uint64) Field { return Field{K: k, T: kindUint64, U64: v} }

// Float64 creates a 64-bit floating-point field.
// Suitable for decimal numbers and scientific notation.
func Float64(k string, v float64) Field {
	return Field{K: k, T: kindFloat64, F64: v}
}

// Bool creates a boolean field.
// Internally stored as int64 (1 for true, 0 for false) for efficiency.
func Bool(k string, v bool) Field {
	if v {
		return Field{K: k, T: kindBool, I64: 1}
	}
	return Field{K: k, T: kindBool, I64: 0}
}

// Dur creates a duration field from time.Duration.
// Stored as int64 nanoseconds for precision and efficiency.
func Dur(k string, v time.Duration) Field { return Field{K: k, T: kindDur, I64: int64(v)} }

// TimeField creates a timestamp field from time.Time.
// Stored as Unix nanoseconds for high precision and compact representation.
func TimeField(k string, v time.Time) Field {
	return Field{K: k, T: kindTime, I64: v.UnixNano()}
}

// Bytes creates a byte slice field.
// Useful for binary data, encoded strings, or raw bytes.
func Bytes(k string, v []byte) Field { return Field{K: k, T: kindBytes, B: v} }

// Additional convenient constructors for common Go types

// Int8 creates a field from an int8 value.
func Int8(k string, v int8) Field { return Field{K: k, T: kindInt64, I64: int64(v)} }

// Int16 creates a field from an int16 value.
func Int16(k string, v int16) Field { return Field{K: k, T: kindInt64, I64: int64(v)} }

// Int32 creates a field from an int32 value.
func Int32(k string, v int32) Field { return Field{K: k, T: kindInt64, I64: int64(v)} }

// Uint creates a field from a uint value.
func Uint(k string, v uint) Field { return Field{K: k, T: kindUint64, U64: uint64(v)} }

// Uint8 creates a field from a uint8 value.
func Uint8(k string, v uint8) Field { return Field{K: k, T: kindUint64, U64: uint64(v)} }

// Uint16 creates a field from a uint16 value.
func Uint16(k string, v uint16) Field { return Field{K: k, T: kindUint64, U64: uint64(v)} }

// Uint32 creates a field from a uint32 value.
func Uint32(k string, v uint32) Field { return Field{K: k, T: kindUint64, U64: uint64(v)} }

// Float32 creates a field from a float32 value.
func Float32(k string, v float32) Field { return Field{K: k, T: kindFloat64, F64: float64(v)} }

// Time creates a timestamp field from time.Time (alias for TimeField for consistency).
func Time(k string, v time.Time) Field { return TimeField(k, v) }

// String creates a string field (alias for Str for consistency with Go naming).
func String(k, v string) Field { return Str(k, v) }

// Binary creates a byte slice field (alias for Bytes).
func Binary(k string, v []byte) Field { return Bytes(k, v) }

// Methods for Field type

// Type returns the kind of data stored in this field.
func (f Field) Type() kind {
	return f.T
}

// Key returns the field's key name.
func (f Field) Key() string {
	return f.K
}

// IsString returns true if the field contains string data.
func (f Field) IsString() bool {
	return f.T == kindString
}

// IsInt returns true if the field contains integer data.
func (f Field) IsInt() bool {
	return f.T == kindInt64
}

// IsUint returns true if the field contains unsigned integer data.
func (f Field) IsUint() bool {
	return f.T == kindUint64
}

// IsFloat returns true if the field contains floating-point data.
func (f Field) IsFloat() bool {
	return f.T == kindFloat64
}

// IsBool returns true if the field contains boolean data.
func (f Field) IsBool() bool {
	return f.T == kindBool
}

// IsDuration returns true if the field contains duration data.
func (f Field) IsDuration() bool {
	return f.T == kindDur
}

// IsTime returns true if the field contains timestamp data.
func (f Field) IsTime() bool {
	return f.T == kindTime
}

// IsBytes returns true if the field contains byte slice data.
func (f Field) IsBytes() bool {
	return f.T == kindBytes
}

// StringValue returns the string value if the field is a string, empty string otherwise.
func (f Field) StringValue() string {
	if f.T == kindString {
		return f.Str
	}
	return ""
}

// IntValue returns the int64 value if the field is an integer, 0 otherwise.
func (f Field) IntValue() int64 {
	if f.T == kindInt64 {
		return f.I64
	}
	return 0
}

// UintValue returns the uint64 value if the field is an unsigned integer, 0 otherwise.
func (f Field) UintValue() uint64 {
	if f.T == kindUint64 {
		return f.U64
	}
	return 0
}

// FloatValue returns the float64 value if the field is a float, 0.0 otherwise.
func (f Field) FloatValue() float64 {
	if f.T == kindFloat64 {
		return f.F64
	}
	return 0.0
}

// BoolValue returns the boolean value if the field is a bool, false otherwise.
func (f Field) BoolValue() bool {
	if f.T == kindBool {
		return f.I64 != 0
	}
	return false
}

// DurationValue returns the time.Duration value if the field is a duration, 0 otherwise.
func (f Field) DurationValue() time.Duration {
	if f.T == kindDur {
		return time.Duration(f.I64)
	}
	return 0
}

// TimeValue returns the time.Time value if the field is a timestamp, zero time otherwise.
func (f Field) TimeValue() time.Time {
	if f.T == kindTime {
		return time.Unix(0, f.I64)
	}
	return time.Time{}
}

// BytesValue returns the byte slice value if the field is bytes, nil otherwise.
func (f Field) BytesValue() []byte {
	if f.T == kindBytes {
		return f.B
	}
	return nil
}

// Error helpers (zap-like)

// Err creates an error field with key "error".
// If err is nil, returns a field with empty string (compatible but not elided).
func Err(err error) Field {
	if err == nil {
		return Str("error", "")
	}
	return Str("error", err.Error())
}

// NamedErr creates an error field with a custom key.
// If err is nil, returns a field with empty string (compatible but not elided).
func NamedErr(k string, err error) Field {
	if err == nil {
		return Str(k, "")
	}
	return Str(k, err.Error())
}

// ErrorField creates an error field for logging errors.
// Equivalent to NamedErr("error", err) but uses the proper error type for potential optimization.
func ErrorField(err error) Field {
	if err == nil {
		return Field{K: "error", T: kindError, Obj: nil}
	}
	return Field{K: "error", T: kindError, Obj: err}
}

// NamedError creates an error field with a custom key using proper error type.
func NamedError(k string, err error) Field {
	return Field{K: k, T: kindError, Obj: err}
}

// Stringer creates a stringer field for objects implementing fmt.Stringer.
func Stringer(k string, val interface{ String() string }) Field {
	return Field{K: k, T: kindStringer, Obj: val}
}

// Object creates an object field for arbitrary data.
func Object(k string, val interface{}) Field {
	return Field{K: k, T: kindObject, Obj: val}
}

// Errors creates a field for multiple errors (like Zap's ErrorsField).
func Errors(k string, errs []error) Field {
	return Field{K: k, T: kindObject, Obj: errs}
}
