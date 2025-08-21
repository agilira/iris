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
	estimatedSize := len(message) + 50 // base size for timestamp + level
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
	estimatedSize := len(message) + 50 // base size for timestamp + level
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
	// OPTIMIZATION: Fast paths for most common types with reduced switch overhead
	switch field.Type {
	case uint8(StringType):
		// FAST PATH: Direct string append (most common case)
		e.buf = append(e.buf, field.GetString()...)

	case uint8(IntType), uint8(Int64Type):
		// FAST PATH: Integer append (second most common)
		e.buf = strconv.AppendInt(e.buf, field.GetInt(), 10)

	case uint8(BoolType):
		// OPTIMIZED: Single conditional append
		if field.GetBool() {
			e.buf = append(e.buf, "true"...)
		} else {
			e.buf = append(e.buf, "false"...)
		}

	case uint8(Float64Type):
		// OPTIMIZED: Direct float append
		e.buf = strconv.AppendFloat(e.buf, field.GetFloat(), 'f', -1, 64)

	case uint8(Float32Type):
		e.buf = strconv.AppendFloat(e.buf, field.GetFloat(), 'f', -1, 32)

	// OPTIMIZATION: Group similar integer types to reduce switch branches
	case uint8(Int32Type), uint8(Int16Type), uint8(Int8Type):
		e.buf = strconv.AppendInt(e.buf, field.GetInt(), 10)

	case uint8(UintType), uint8(Uint64Type), uint8(Uint32Type), uint8(Uint16Type), uint8(Uint8Type):
		// Use safe conversion for encoding unsigned integers
		value, _ := SafeInt64ToUint64ForEncoding(field.GetInt())
		e.buf = strconv.AppendUint(e.buf, value, 10)

	case uint8(DurationType):
		// OPTIMIZED: Duration formatting
		duration := time.Duration(field.GetInt())
		e.buf = append(e.buf, duration.String()...)

	case uint8(TimeType):
		// Time formatting
		t := time.Unix(0, field.GetInt())
		e.buf = t.AppendFormat(e.buf, time.RFC3339)

	case uint8(ByteStringType):
		// Byte string as string
		e.buf = append(e.buf, field.GetString()...)

	case uint8(ErrorType):
		// Error formatting - for now use string representation
		e.buf = append(e.buf, field.GetString()...)

	default:
		// Fallback for unknown types
		e.buf = append(e.buf, "<?>"...)
	}
}

// appendFieldValueFastMigration appends field value directly from Field (MIGRATION METHOD)
func (e *FastTextEncoder) appendFieldValueFastMigration(field Field) {
	// OPTIMIZATION: Fast paths for most common types with reduced switch overhead
	switch field.Type {
	case StringType:
		// FAST PATH: Direct string append (most common case)
		e.buf = append(e.buf, field.String...)

	case IntType, Int64Type:
		// FAST PATH: Integer append (second most common)
		e.buf = strconv.AppendInt(e.buf, field.Int, 10)

	case BoolType:
		// OPTIMIZED: Single conditional append
		if field.Bool {
			e.buf = append(e.buf, "true"...)
		} else {
			e.buf = append(e.buf, "false"...)
		}

	case Float64Type:
		// OPTIMIZED: Direct float append
		e.buf = strconv.AppendFloat(e.buf, field.Float, 'f', -1, 64)

	case Float32Type:
		e.buf = strconv.AppendFloat(e.buf, field.Float, 'f', -1, 32)

	// OPTIMIZATION: Group similar integer types to reduce switch branches
	case Int32Type, Int16Type, Int8Type:
		e.buf = strconv.AppendInt(e.buf, field.Int, 10)

	case UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		// Use safe conversion for encoding unsigned integers
		value, _ := SafeInt64ToUint64ForEncoding(field.Int)
		e.buf = strconv.AppendUint(e.buf, value, 10)

	case DurationType:
		// OPTIMIZED: Duration formatting
		duration := time.Duration(field.Int)
		e.buf = append(e.buf, duration.String()...)

	case TimeType:
		// Time formatting
		t := time.Unix(0, field.Int)
		e.buf = t.AppendFormat(e.buf, time.RFC3339)

	case ByteStringType:
		// Byte string as string
		e.buf = append(e.buf, string(field.Bytes)...)

	case ErrorType:
		// Error formatting
		if field.Err != nil {
			e.buf = append(e.buf, field.Err.Error()...)
		} else {
			e.buf = append(e.buf, "<nil>"...)
		}

	default:
		// Fallback for unknown types
		e.buf = append(e.buf, "<?>"...)
	}
}
