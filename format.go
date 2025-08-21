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
func (e *FastTextEncoder) EncodeLogEntry(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) {
	e.Reset()

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

	// OPTIMIZATION 2: Batch timestamp and level in single operation
	if !timestamp.IsZero() {
		e.buf = timestamp.AppendFormat(e.buf, "15:04:05.000")
		e.buf = append(e.buf, ' ')
	}

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

	// Message - direct append, no intermediate space needed
	e.buf = append(e.buf, message...)

	// OPTIMIZATION 4: Optimized caller encoding
	if caller.Valid {
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

	// OPTIMIZATION 5: Batch field encoding
	if len(fields) > 0 {
		e.buf = append(e.buf, ' ')
		for i, field := range fields {
			if i > 0 {
				e.buf = append(e.buf, ' ')
			}
			e.buf = append(e.buf, field.Key...)
			e.buf = append(e.buf, '=')
			e.appendFieldValueFast(field) // Optimized version
		}
	}

	// Stack trace (if present)
	if stackTrace != "" {
		e.buf = append(e.buf, "\nStack trace:\n"...)
		e.buf = append(e.buf, stackTrace...)
	}

	e.buf = append(e.buf, '\n')
}

// appendFieldValueFast appends field value without quotes (ULTRA-OPTIMIZED)
func (e *FastTextEncoder) appendFieldValueFast(field Field) {
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
		e.buf = strconv.AppendUint(e.buf, uint64(field.Int), 10)

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

// BinaryEncoder provides ultra-fast binary encoding
type BinaryEncoder struct {
	buf []byte
}

// NewBinaryEncoder creates a new binary encoder
func NewBinaryEncoder() *BinaryEncoder {
	return &BinaryEncoder{
		buf: make([]byte, 0, 512), // OPTIMIZED: Larger buffer to reduce reallocations
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

// EncodeLogEntry encodes to binary format (ULTRA-FAST OPTIMIZED)
func (e *BinaryEncoder) EncodeLogEntry(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) {
	e.Reset()

	// OPTIMIZATION 1: Pre-calculate exact buffer size to avoid reallocations
	estimatedSize := 8 + 1 + 1 + 2 + len(message) // timestamp + level + flags + msglen + message
	if caller.Valid {
		estimatedSize += 1 + len(caller.File) + 2 // filelen + file + line
	}
	estimatedSize += 1 + len(fields)*20 // fieldcount + estimated field size

	// Ensure exact capacity
	if cap(e.buf) < estimatedSize {
		e.buf = make([]byte, 0, estimatedSize+50)
	}

	// Binary format: [timestamp:8bytes][level:1byte][flags:1byte][msglen:2bytes][msg][caller?][fieldcount:1byte][fields...]

	// OPTIMIZATION 2: Batch timestamp encoding using unsafe for max speed
	var ts int64
	if !timestamp.IsZero() {
		ts = timestamp.UnixNano()
	}

	// Ultra-fast 8-byte timestamp encoding in single operation
	tsBytes := [8]byte{
		byte(ts >> 56), byte(ts >> 48), byte(ts >> 40), byte(ts >> 32),
		byte(ts >> 24), byte(ts >> 16), byte(ts >> 8), byte(ts),
	}
	e.buf = append(e.buf, tsBytes[:]...)

	// OPTIMIZATION 3: Batch level and flags
	var flags byte
	if caller.Valid {
		flags |= 1
	}
	levelAndFlags := [2]byte{byte(level), flags}
	e.buf = append(e.buf, levelAndFlags[:]...)

	// OPTIMIZATION 4: Optimized message encoding
	msgLen := len(message)
	if msgLen > 65535 {
		msgLen = 65535 // Truncate very long messages
	}

	// Batch message length and content
	msgHeader := [2]byte{byte(msgLen >> 8), byte(msgLen)}
	e.buf = append(e.buf, msgHeader[:]...)
	e.buf = append(e.buf, message[:msgLen]...)

	// OPTIMIZATION 5: Optimized caller encoding
	if caller.Valid {
		fileLen := len(caller.File)
		if fileLen > 255 {
			fileLen = 255
		}

		e.buf = append(e.buf, byte(fileLen))
		e.buf = append(e.buf, caller.File[:fileLen]...)

		// Line number as 2 bytes
		line := caller.Line
		lineBytes := [2]byte{byte(line >> 8), byte(line)}
		e.buf = append(e.buf, lineBytes[:]...)
	}

	// OPTIMIZATION 6: Batch field count and encoding
	fieldCount := len(fields)
	if fieldCount > 255 {
		fieldCount = 255 // Limit fields
	}
	e.buf = append(e.buf, byte(fieldCount))

	// Ultra-fast field encoding
	for i := 0; i < fieldCount; i++ {
		e.encodeBinaryFieldFast(fields[i])
	}
}

// encodeBinaryFieldFast encodes a single field in binary format (ULTRA-OPTIMIZED)
func (e *BinaryEncoder) encodeBinaryFieldFast(field Field) {
	// OPTIMIZATION 1: Batch header encoding (type + keylen)
	keyLen := len(field.Key)
	if keyLen > 255 {
		keyLen = 255
	}

	// Field format: [type:1byte][keylen:1byte][key][value...]
	header := [2]byte{byte(field.Type), byte(keyLen)}
	e.buf = append(e.buf, header[:]...)
	e.buf = append(e.buf, field.Key[:keyLen]...)

	// OPTIMIZATION 2: Fast paths for hot types with reduced branching
	switch field.Type {
	case StringType:
		// FAST PATH: String encoding (most common)
		strLen := len(field.String)
		if strLen > 65535 {
			strLen = 65535
		}

		// Batch string length and content
		strHeader := [2]byte{byte(strLen >> 8), byte(strLen)}
		e.buf = append(e.buf, strHeader[:]...)
		e.buf = append(e.buf, field.String[:strLen]...)

	case IntType, Int64Type:
		// FAST PATH: Integer encoding (second most common)
		val := uint64(field.Int)
		intBytes := [8]byte{
			byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		}
		e.buf = append(e.buf, intBytes[:]...)

	case BoolType:
		// FAST PATH: Boolean (single byte)
		if field.Bool {
			e.buf = append(e.buf, 1)
		} else {
			e.buf = append(e.buf, 0)
		}

	case Float64Type:
		// OPTIMIZED: Float64 using unsafe for speed
		val := *(*uint64)(unsafe.Pointer(&field.Float))
		floatBytes := [8]byte{
			byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		}
		e.buf = append(e.buf, floatBytes[:]...)

	case Float32Type:
		// 4 bytes for float32
		val := *(*uint32)(unsafe.Pointer(&field.Float))
		floatBytes := [4]byte{
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		}
		e.buf = append(e.buf, floatBytes[:]...)

	// OPTIMIZATION 3: Group similar integer types
	case Int32Type, Int16Type, Int8Type, UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		val := uint64(field.Int)
		intBytes := [8]byte{
			byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		}
		e.buf = append(e.buf, intBytes[:]...)

	case DurationType, TimeType:
		// 8 bytes for duration/time
		val := uint64(field.Int)
		timeBytes := [8]byte{
			byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		}
		e.buf = append(e.buf, timeBytes[:]...)

	case ByteStringType:
		// Byte string encoding
		dataLen := len(field.Bytes)
		if dataLen > 65535 {
			dataLen = 65535
		}

		dataHeader := [2]byte{byte(dataLen >> 8), byte(dataLen)}
		e.buf = append(e.buf, dataHeader[:]...)
		e.buf = append(e.buf, field.Bytes[:dataLen]...)

	default:
		// Fallback for unknown types - encode as empty
		e.buf = append(e.buf, 0, 0) // zero-length value
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
