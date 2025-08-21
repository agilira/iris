// format_text.go: Fast text encoder for Iris logging
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strconv"
	"time"
)

// FastTextEncoder provides ultra-fast text encoding
type FastTextEncoder struct {
	buf []byte
}

// NewFastTextEncoder creates a new fast text encoder
func NewFastTextEncoder() *FastTextEncoder {
	return &FastTextEncoder{
		buf: make([]byte, 0, 1024), // OPTIMIZED: Larger initial buffer to reduce reallocations
	}
}

// Reset resets the encoder for reuse
func (e *FastTextEncoder) Reset() {
	e.buf = e.buf[:0]
}

// Bytes returns the encoded bytes
func (e *FastTextEncoder) Bytes() []byte {
	return e.buf
}

// EncodeLogEntry encodes a log entry to fast text format
// Format: TIMESTAMP LEVEL MESSAGE [CALLER] field1=value1 field2=value2
func (e *FastTextEncoder) EncodeLogEntry(timestamp time.Time, level Level, message string, fields []BinaryField, caller Caller, stackTrace string) {
	e.Reset()

	// Pre-calculate capacity and prepare buffer
	e.prepareBuffer(timestamp, level, message, fields, caller, stackTrace)

	// Encode components in order
	e.appendTimestamp(timestamp)
	e.appendLevel(level)
	e.appendMessage(message)
	e.appendCaller(caller)
	e.appendBinaryFields(fields)
	e.appendStackTrace(stackTrace)

	e.buf = append(e.buf, '\n')
}

// prepareBuffer calculates required capacity and ensures buffer size
func (e *FastTextEncoder) prepareBuffer(timestamp time.Time, level Level, message string, fields []BinaryField, caller Caller, stackTrace string) {
	// OPTIMIZATION 1: Pre-calculate required capacity to minimize reallocations
	estimatedSize := len(message)
	if !timestamp.IsZero() {
		estimatedSize += 13 // timestamp format "15:04:05.000 " = 13 chars
	}
	// All levels are formatted to 6 chars (including space)
	_ = level // acknowledge parameter usage - all levels are 6 chars
	estimatedSize += 6
	for _, field := range fields {
		estimatedSize += len(field.GetKey()) + 20 // field + estimated value
	}
	if caller.Valid {
		estimatedSize += len(caller.File) + 10
	}
	if stackTrace != "" {
		estimatedSize += len(stackTrace) + 20
	}

	// Ensure capacity in one go
	if cap(e.buf) < estimatedSize {
		e.buf = make([]byte, 0, estimatedSize+100) // some extra margin
	}
}

// appendTimestamp adds timestamp to buffer
func (e *FastTextEncoder) appendTimestamp(timestamp time.Time) {
	if !timestamp.IsZero() {
		e.buf = timestamp.AppendFormat(e.buf, "15:04:05.000")
		e.buf = append(e.buf, ' ')
	}
}

// appendLevel adds level to buffer with consistent formatting
func (e *FastTextEncoder) appendLevel(level Level) {
	// OPTIMIZATION 3: Pre-computed level strings with consistent length
	switch level {
	case DebugLevel:
		e.buf = append(e.buf, "DEBUG "...)
	case InfoLevel:
		e.buf = append(e.buf, "INFO  "...) // FIXED: Consistent spacing
	case WarnLevel:
		e.buf = append(e.buf, "WARN  "...)
	case ErrorLevel:
		e.buf = append(e.buf, "ERROR "...)
	case FatalLevel:
		e.buf = append(e.buf, "FATAL "...)
	default:
		e.buf = append(e.buf, "UNKNW "...)
	}
}

// appendMessage adds message to buffer
func (e *FastTextEncoder) appendMessage(message string) {
	e.buf = append(e.buf, message...)
}

// appendCaller adds caller information to buffer
func (e *FastTextEncoder) appendCaller(caller Caller) {
	if !caller.Valid {
		return
	}

	e.buf = append(e.buf, " ["...)

	// Fast filename extraction using last slash
	filename := caller.File
	if idx := lastIndexByte(filename, '/'); idx >= 0 {
		filename = filename[idx+1:]
	}

	e.buf = append(e.buf, filename...)
	e.buf = append(e.buf, ':')
	e.buf = strconv.AppendInt(e.buf, int64(caller.Line), 10)
	e.buf = append(e.buf, ']')
}

// appendBinaryFields adds fields to buffer
func (e *FastTextEncoder) appendBinaryFields(fields []BinaryField) {
	if len(fields) == 0 {
		return
	}

	e.buf = append(e.buf, ' ')
	for i, field := range fields {
		if i > 0 {
			e.buf = append(e.buf, ' ')
		}
		e.buf = append(e.buf, field.GetKey()...)
		e.buf = append(e.buf, '=')
		e.appendFieldValueFast(field) // Optimized version
	}
}

// appendStackTrace adds stack trace to buffer
func (e *FastTextEncoder) appendStackTrace(stackTrace string) {
	if stackTrace != "" {
		e.buf = append(e.buf, "\nStack trace:\n"...)
		e.buf = append(e.buf, stackTrace...)
	}
}

// EncodeLogEntryMigration encodes directly from Field slice (MIGRATION METHOD)
func (e *FastTextEncoder) EncodeLogEntryMigration(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) {
	e.Reset()

	// Pre-calculate capacity and prepare buffer
	e.prepareBufferForMigration(timestamp, level, message, fields, caller, stackTrace)

	// Encode components in order (reuse methods where possible)
	e.appendTimestamp(timestamp)
	e.appendLevel(level)
	e.appendMessage(message)
	e.appendCaller(caller)
	e.appendMigrationFields(fields)
	e.appendStackTrace(stackTrace)

	e.buf = append(e.buf, '\n')
}

// prepareBufferForMigration calculates required capacity for Field slice
func (e *FastTextEncoder) prepareBufferForMigration(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) {
	// OPTIMIZATION 1: Pre-calculate required capacity to minimize reallocations
	estimatedSize := len(message)
	if !timestamp.IsZero() {
		estimatedSize += 13 // timestamp format "15:04:05.000 " = 13 chars
	}
	// All levels are formatted to 6 chars (including space)
	_ = level // acknowledge parameter usage - all levels are 6 chars
	estimatedSize += 6
	for _, field := range fields {
		estimatedSize += len(field.Key) + 20 // field + estimated value
	}
	if caller.Valid {
		estimatedSize += len(caller.File) + 10
	}
	if stackTrace != "" {
		estimatedSize += len(stackTrace) + 20
	}

	// Ensure capacity in one go
	if cap(e.buf) < estimatedSize {
		e.buf = make([]byte, 0, estimatedSize+100) // some extra margin
	}
}

// appendMigrationFields adds Field slice to buffer
func (e *FastTextEncoder) appendMigrationFields(fields []Field) {
	if len(fields) == 0 {
		return
	}

	e.buf = append(e.buf, ' ')
	for i, field := range fields {
		if i > 0 {
			e.buf = append(e.buf, ' ')
		}
		e.buf = append(e.buf, field.Key...)
		e.buf = append(e.buf, '=')
		e.appendFieldValueFastMigration(field) // Direct Field version
	}
}

// appendFieldValueFast appends field value without quotes (ULTRA-OPTIMIZED)
func (e *FastTextEncoder) appendFieldValueFast(field BinaryField) {
	// OPTIMIZATION: Fast paths for most common types first
	switch field.Type {
	case uint8(StringType), uint8(ByteStringType), uint8(ErrorType):
		// FAST PATH: Direct string append (most common case)
		e.buf = append(e.buf, field.GetString()...)
	case uint8(IntType), uint8(Int64Type), uint8(Int32Type), uint8(Int16Type), uint8(Int8Type):
		// FAST PATH: Integer append (second most common)
		e.buf = strconv.AppendInt(e.buf, field.GetInt(), 10)
	case uint8(BoolType):
		// OPTIMIZED: Single conditional append
		e.appendBoolValue(field.GetBool())
	case uint8(Float64Type), uint8(Float32Type):
		e.appendFloatValueFromBinary(field)
	case uint8(UintType), uint8(Uint64Type), uint8(Uint32Type), uint8(Uint16Type), uint8(Uint8Type):
		e.appendUintValue(field.GetInt())
	case uint8(DurationType), uint8(TimeType):
		e.appendTimeValueFromBinary(field)
	default:
		// Fallback for unknown types
		e.buf = append(e.buf, "<?>"...)
	}
}

// appendFloatValueFromBinary appends float value from BinaryField
func (e *FastTextEncoder) appendFloatValueFromBinary(field BinaryField) {
	if field.Type == uint8(Float64Type) {
		e.appendFloat64Value(field.GetFloat())
	} else {
		e.appendFloat32Value(field.GetFloat())
	}
}

// appendTimeValueFromBinary appends duration or time value from BinaryField
func (e *FastTextEncoder) appendTimeValueFromBinary(field BinaryField) {
	if field.Type == uint8(DurationType) {
		e.appendDurationValue(field.GetInt())
	} else {
		e.appendTimeValue(field.GetInt())
	}
}

// appendFieldValueFastMigration appends field value directly from Field (MIGRATION METHOD)
func (e *FastTextEncoder) appendFieldValueFastMigration(field Field) {
	// OPTIMIZATION: Fast paths for most common types with reduced switch overhead
	switch field.Type {
	case StringType:
		e.appendStringValue(field.String)
	case IntType, Int64Type, Int32Type, Int16Type, Int8Type:
		e.appendIntValue(field.Int)
	case BoolType:
		e.appendBoolValue(field.Bool)
	case Float64Type, Float32Type:
		e.appendFloatValueFromField(field)
	case UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		e.appendUintValue(field.Int)
	case DurationType, TimeType:
		e.appendTimeValueFromField(field)
	case ByteStringType:
		e.appendByteStringValue(field.Bytes)
	case ErrorType:
		e.appendErrorValue(field.Err)
	default:
		e.appendUnknownValue()
	}
}

// appendFloatValueFromField appends float value from Field
func (e *FastTextEncoder) appendFloatValueFromField(field Field) {
	if field.Type == Float64Type {
		e.appendFloat64Value(field.Float)
	} else {
		e.appendFloat32Value(field.Float)
	}
}

// appendTimeValueFromField appends duration or time value from Field
func (e *FastTextEncoder) appendTimeValueFromField(field Field) {
	if field.Type == DurationType {
		e.appendDurationValue(field.Int)
	} else {
		e.appendTimeValue(field.Int)
	}
}

// appendStringValue appends string value
func (e *FastTextEncoder) appendStringValue(s string) {
	e.buf = append(e.buf, s...)
}

// appendIntValue appends integer value
func (e *FastTextEncoder) appendIntValue(i int64) {
	e.buf = strconv.AppendInt(e.buf, i, 10)
}

// appendBoolValue appends boolean value
func (e *FastTextEncoder) appendBoolValue(b bool) {
	if b {
		e.buf = append(e.buf, "true"...)
	} else {
		e.buf = append(e.buf, "false"...)
	}
}

// appendFloat64Value appends 64-bit float value
func (e *FastTextEncoder) appendFloat64Value(f float64) {
	e.buf = strconv.AppendFloat(e.buf, f, 'f', -1, 64)
}

// appendFloat32Value appends 32-bit float value
func (e *FastTextEncoder) appendFloat32Value(f float64) {
	e.buf = strconv.AppendFloat(e.buf, f, 'f', -1, 32)
}

// appendUintValue appends unsigned integer value
func (e *FastTextEncoder) appendUintValue(i int64) {
	// Use safe conversion for encoding unsigned integers
	value, _ := SafeInt64ToUint64ForEncoding(i)
	e.buf = strconv.AppendUint(e.buf, value, 10)
}

// appendDurationValue appends duration value
func (e *FastTextEncoder) appendDurationValue(ns int64) {
	duration := time.Duration(ns)
	e.buf = append(e.buf, duration.String()...)
}

// appendTimeValue appends time value
func (e *FastTextEncoder) appendTimeValue(ns int64) {
	t := time.Unix(0, ns)
	e.buf = t.AppendFormat(e.buf, time.RFC3339)
}

// appendByteStringValue appends byte string value
func (e *FastTextEncoder) appendByteStringValue(bytes []byte) {
	e.buf = append(e.buf, string(bytes)...)
}

// appendErrorValue appends error value
func (e *FastTextEncoder) appendErrorValue(err error) {
	if err != nil {
		e.buf = append(e.buf, err.Error()...)
	} else {
		e.buf = append(e.buf, "<nil>"...)
	}
}

// appendUnknownValue appends placeholder for unknown types
func (e *FastTextEncoder) appendUnknownValue() {
	e.buf = append(e.buf, "<?>"...)
}
