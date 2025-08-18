// format.go: Ultra-fast formatting options for Iris
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strconv"
	"time"
	"unsafe"
)

// Format represents output format
type Format int

const (
	// JSONFormat produces JSON output (slower but structured)
	JSONFormat Format = iota
	// ConsoleFormat produces colorized console output (development)
	ConsoleFormat
	// FastTextFormat produces fast text output (ultra-fast)
	FastTextFormat
	// BinaryFormat produces binary output (fastest)
	BinaryFormat
)

// FastTextEncoder provides ultra-fast text encoding
type FastTextEncoder struct {
	buf []byte
}

// NewFastTextEncoder creates a new fast text encoder
func NewFastTextEncoder() *FastTextEncoder {
	return &FastTextEncoder{
		buf: make([]byte, 0, 512), // Smaller buffer for text
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
func (e *FastTextEncoder) EncodeLogEntry(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) {
	e.Reset()

	// Timestamp (simplified format for speed)
	if !timestamp.IsZero() {
		e.buf = timestamp.AppendFormat(e.buf, "15:04:05.000")
		e.buf = append(e.buf, ' ')
	}

	// Level (uppercase, fixed width)
	switch level {
	case DebugLevel:
		e.buf = append(e.buf, "DEBUG"...)
	case InfoLevel:
		e.buf = append(e.buf, "INFO "...)
	case WarnLevel:
		e.buf = append(e.buf, "WARN "...)
	case ErrorLevel:
		e.buf = append(e.buf, "ERROR"...)
	case FatalLevel:
		e.buf = append(e.buf, "FATAL"...)
	}
	e.buf = append(e.buf, ' ')

	// Message
	e.buf = append(e.buf, message...)

	// Caller (filename:line format for speed)
	if caller.Valid {
		e.buf = append(e.buf, ' ')
		e.buf = append(e.buf, '[')

		// Show only filename for speed
		filename := caller.File
		if idx := lastIndexByte(filename, '/'); idx >= 0 {
			filename = filename[idx+1:]
		}

		e.buf = append(e.buf, filename...)
		e.buf = append(e.buf, ':')
		e.buf = strconv.AppendInt(e.buf, int64(caller.Line), 10)
		e.buf = append(e.buf, ']')
	}

	// Fields (key=value format)
	for _, field := range fields {
		e.buf = append(e.buf, ' ')
		e.buf = append(e.buf, field.Key...)
		e.buf = append(e.buf, '=')
		e.appendFieldValue(field)
	}

	// Stack trace (if present)
	if stackTrace != "" {
		e.buf = append(e.buf, "\nStack trace:\n"...)
		e.buf = append(e.buf, stackTrace...)
	}

	e.buf = append(e.buf, '\n')
}

// appendFieldValue appends field value without quotes (ultra-fast)
func (e *FastTextEncoder) appendFieldValue(field Field) {
	switch field.Type {
	case StringType:
		e.buf = append(e.buf, field.String...)

	case IntType, Int64Type:
		e.buf = strconv.AppendInt(e.buf, field.Int, 10)

	case Float64Type:
		e.buf = strconv.AppendFloat(e.buf, field.Float, 'f', 3, 64)

	case BoolType:
		if field.Bool {
			e.buf = append(e.buf, "true"...)
		} else {
			e.buf = append(e.buf, "false"...)
		}

	case DurationType:
		duration := time.Duration(field.Int)
		e.buf = append(e.buf, duration.String()...)

	case TimeType:
		timestamp := time.Unix(0, field.Int)
		e.buf = timestamp.AppendFormat(e.buf, "15:04:05")

	case ErrorType:
		if field.Err != nil {
			e.buf = append(e.buf, field.Err.Error()...)
		} else {
			e.buf = append(e.buf, "nil"...)
		}
	}
}

// BinaryEncoder provides ultra-fast binary encoding
type BinaryEncoder struct {
	buf []byte
}

// NewBinaryEncoder creates a new binary encoder
func NewBinaryEncoder() *BinaryEncoder {
	return &BinaryEncoder{
		buf: make([]byte, 0, 256), // Even smaller for binary
	}
}

// Reset resets the encoder
func (e *BinaryEncoder) Reset() {
	e.buf = e.buf[:0]
}

// Bytes returns the encoded bytes
func (e *BinaryEncoder) Bytes() []byte {
	return e.buf
}

// EncodeLogEntry encodes to binary format (fastest possible)
func (e *BinaryEncoder) EncodeLogEntry(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) {
	e.Reset()

	// Binary format: [timestamp:8bytes][level:1byte][flags:1byte][msglen:2bytes][msg][caller?][fieldcount:1byte][fields...]

	// Timestamp as Unix nano (8 bytes, 0 if disabled)
	var ts int64
	if !timestamp.IsZero() {
		ts = timestamp.UnixNano()
	}
	e.buf = append(e.buf,
		byte(ts>>56), byte(ts>>48), byte(ts>>40), byte(ts>>32),
		byte(ts>>24), byte(ts>>16), byte(ts>>8), byte(ts))

	// Level (1 byte)
	e.buf = append(e.buf, byte(level))

	// Flags (1 byte: bit 0 = has caller)
	var flags byte
	if caller.Valid {
		flags |= 1
	}
	e.buf = append(e.buf, flags)

	// Message length + message
	msgLen := len(message)
	if msgLen > 65535 {
		msgLen = 65535 // Truncate very long messages
	}
	e.buf = append(e.buf, byte(msgLen>>8), byte(msgLen))
	e.buf = append(e.buf, message[:msgLen]...)

	// Caller info (if present)
	if caller.Valid {
		// File length + file
		fileLen := len(caller.File)
		if fileLen > 255 {
			fileLen = 255
		}
		e.buf = append(e.buf, byte(fileLen))
		e.buf = append(e.buf, caller.File[:fileLen]...)

		// Line number (2 bytes)
		line := caller.Line
		e.buf = append(e.buf, byte(line>>8), byte(line))
	}

	// Field count
	fieldCount := len(fields)
	if fieldCount > 255 {
		fieldCount = 255 // Limit fields
	}
	e.buf = append(e.buf, byte(fieldCount))

	// Fields (simplified binary encoding)
	for i, field := range fields[:fieldCount] {
		if i >= 255 {
			break
		}
		e.encodeBinaryField(field)
	}
}

// encodeBinaryField encodes a single field in binary format
func (e *BinaryEncoder) encodeBinaryField(field Field) {
	// Field format: [type:1byte][keylen:1byte][key][value...]

	// Type
	e.buf = append(e.buf, byte(field.Type))

	// Key length + key
	keyLen := len(field.Key)
	if keyLen > 255 {
		keyLen = 255
	}
	e.buf = append(e.buf, byte(keyLen))
	e.buf = append(e.buf, field.Key[:keyLen]...)

	// Value based on type
	switch field.Type {
	case StringType:
		strLen := len(field.String)
		if strLen > 65535 {
			strLen = 65535
		}
		e.buf = append(e.buf, byte(strLen>>8), byte(strLen))
		e.buf = append(e.buf, field.String[:strLen]...)

	case IntType, Int64Type:
		// 8 bytes for int64
		val := uint64(field.Int)
		e.buf = append(e.buf,
			byte(val>>56), byte(val>>48), byte(val>>40), byte(val>>32),
			byte(val>>24), byte(val>>16), byte(val>>8), byte(val))

	case Float64Type:
		// 8 bytes for float64
		val := *(*uint64)(unsafe.Pointer(&field.Float))
		e.buf = append(e.buf,
			byte(val>>56), byte(val>>48), byte(val>>40), byte(val>>32),
			byte(val>>24), byte(val>>16), byte(val>>8), byte(val))

	case BoolType:
		if field.Bool {
			e.buf = append(e.buf, 1)
		} else {
			e.buf = append(e.buf, 0)
		}

	case DurationType, TimeType:
		// 8 bytes for duration/time
		val := uint64(field.Int)
		e.buf = append(e.buf,
			byte(val>>56), byte(val>>48), byte(val>>40), byte(val>>32),
			byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
	}
}

// lastIndexByte returns the last index of byte c in string s, or -1 if not found
func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
