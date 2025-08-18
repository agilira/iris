// encoder.go: Manual JSON encoder for maximum performance
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"sync"
	"time"
)

// Optimized buffer pool per performance critiche (inspired by industry best practices)
var optimizedBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 64) // Start with 64B like industry leaders
		return &buf
	},
}

// JSONEncoder provides ultra-fast JSON encoding without reflection (OPTIMIZED)
type JSONEncoder struct {
	buf           []byte
	useBufferPool bool // Flag per abilitare buffer pooling
}

// NewJSONEncoder creates a new JSON encoder with optimized buffer strategy
func NewJSONEncoder() *JSONEncoder {
	return &JSONEncoder{
		buf:           nil,  // NO pre-allocation! Use pool instead
		useBufferPool: true, // ALWAYS use pooling
	}
}

// Reset resets the encoder for reuse (INDUSTRY-STANDARD OPTIMIZATION)
func (e *JSONEncoder) Reset() {
	if e.buf != nil {
		// Return buffer to pool (optimized pooling strategy)
		optimizedBufferPool.Put(&e.buf)
		e.buf = nil
	}
}

// getBuf gets a buffer from pool (OPTIMIZED pooling strategy)
func (e *JSONEncoder) getBuf() {
	if e.buf == nil {
		bufPtr := optimizedBufferPool.Get().(*[]byte)
		e.buf = (*bufPtr)[:0]
	}
}

// ensureCapacity ensures buffer has enough space (INTELLIGENT growth strategy)
func (e *JSONEncoder) ensureCapacity(needed int) {
	if cap(e.buf) < needed {
		// Need to grow - create larger buffer with smart sizing
		newBuf := make([]byte, len(e.buf), needed)
		copy(newBuf, e.buf)
		// Return old buffer to pool for reuse
		optimizedBufferPool.Put(&e.buf)
		e.buf = newBuf
	}
}

// Bytes returns the encoded JSON bytes
func (e *JSONEncoder) Bytes() []byte {
	return e.buf
}

// EncodeLogEntry encodes a complete log entry to JSON (MEMORY-OPTIMIZED)
func (e *JSONEncoder) EncodeLogEntry(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) {
	// CRITICAL: Get buffer from pool instead of Reset
	e.Reset()  // Return old buffer to pool
	e.getBuf() // Get fresh buffer from pool

	// TASK M1: Micro-tuned size estimation (IMPLEMENTA)
	// Zap uses 64B base, we need precise calculation not 100B
	baseSize := 75 // timestamp(35) + level(15) + message wrapper(25)
	messageSize := len(message)
	fieldsSize := 0
	for _, field := range fields {
		fieldsSize += len(field.Key) + 15 // key + quotes + colon + quotes + comma
		switch field.Type {
		case StringType:
			fieldsSize += len(field.String)
		case IntType:
			fieldsSize += 10 // max int digits
		case BoolType:
			fieldsSize += 5 // true/false
		default:
			fieldsSize += 20 // safe estimate for any type
		}
	}
	estimatedSize := baseSize + messageSize + fieldsSize
	if caller.Valid {
		estimatedSize += len(caller.File) + len(caller.Function) + 50
	}
	if stackTrace != "" {
		estimatedSize += len(stackTrace) + 20
	}

	// Ensure buffer has enough capacity
	if cap(e.buf) < estimatedSize {
		e.ensureCapacity(estimatedSize)
	}

	e.buf = append(e.buf, '{')

	// Timestamp
	if !timestamp.IsZero() {
		e.buf = append(e.buf, `"timestamp":"`...)
		e.buf = timestamp.AppendFormat(e.buf, time.RFC3339Nano)
		e.buf = append(e.buf, '"')
	}

	// Level
	if !timestamp.IsZero() {
		e.buf = append(e.buf, ',')
	}
	e.buf = append(e.buf, `"level":"`...)
	e.buf = append(e.buf, level.String()...)
	e.buf = append(e.buf, '"')

	// Caller information
	if caller.Valid {
		e.buf = append(e.buf, `,"caller":"`...)
		e.escapeStringFast(caller.File)
		e.buf = append(e.buf, ':')
		e.buf = strconv.AppendInt(e.buf, int64(caller.Line), 10)
		e.buf = append(e.buf, '"')

		if caller.Function != "" {
			e.buf = append(e.buf, `,"function":"`...)
			e.escapeStringFast(caller.Function)
			e.buf = append(e.buf, '"')
		}
	}

	// Stack trace
	if stackTrace != "" {
		e.buf = append(e.buf, `,"stacktrace":"`...)
		e.escapeStringFast(stackTrace)
		e.buf = append(e.buf, '"')
	}

	// Message
	e.buf = append(e.buf, `,"message":"`...)
	e.escapeStringFast(message)
	e.buf = append(e.buf, '"')

	// Fields
	for _, field := range fields {
		e.buf = append(e.buf, ',')
		e.encodeField(field)
	}

	e.buf = append(e.buf, "}\n"...)
}

// TASK M4: Inline key encoding optimization (IMPLEMENTA)
func (e *JSONEncoder) encodeField(field Field) {
	// M4: Inline key encoding to avoid slice overhead
	e.buf = append(e.buf, '"')
	e.buf = append(e.buf, field.Key...)
	e.buf = append(e.buf, '"', ':')

	// OPTIMIZATION: Fast paths per hot types (String, Int most common)
	switch field.Type {
	case StringType:
		// FAST PATH: String encoding (most common case)
		e.buf = append(e.buf, '"')
		e.escapeStringFast(field.String)
		e.buf = append(e.buf, '"')

	case IntType, Int64Type:
		// FAST PATH: Integer encoding (second most common)
		e.buf = strconv.AppendInt(e.buf, field.Int, 10)

	case BoolType:
		// M4: Optimized boolean with single append
		if field.Bool {
			e.buf = append(e.buf, 't', 'r', 'u', 'e')
		} else {
			e.buf = append(e.buf, 'f', 'a', 'l', 's', 'e')
		}

	case Float64Type:
		// OPTIMIZED: Float encoding
		e.buf = strconv.AppendFloat(e.buf, field.Float, 'f', -1, 64)

	// Consolidate similar integer types per reduce switch overhead
	case Int32Type, Int16Type, Int8Type:
		e.buf = strconv.AppendInt(e.buf, field.Int, 10)

	case UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		e.buf = strconv.AppendUint(e.buf, uint64(field.Int), 10)

	case Float32Type:
		e.buf = strconv.AppendFloat(e.buf, field.Float, 'f', -1, 32)

	case DurationType:
		e.buf = append(e.buf, '"')
		duration := time.Duration(field.Int)
		e.buf = append(e.buf, duration.String()...)
		e.buf = append(e.buf, '"')

	case ByteStringType:
		e.buf = append(e.buf, '"')
		e.escapeStringFast(string(field.Bytes))
		e.buf = append(e.buf, '"')

	case BinaryType:
		e.buf = append(e.buf, '"')
		encoded := base64.StdEncoding.EncodeToString(field.Bytes)
		e.buf = append(e.buf, encoded...)
		e.buf = append(e.buf, '"')

	case AnyType:
		// CRITICAL OPTIMIZATION: Avoid json.Marshal when possible
		if field.Any != nil {
			e.encodeAnyTypeFast(field.Any)
		} else {
			e.buf = append(e.buf, "null"...)
		}

	case ErrorType:
		if field.Err != nil {
			e.buf = append(e.buf, '"')
			e.escapeStringFast(field.Err.Error())
			e.buf = append(e.buf, '"')
		} else {
			e.buf = append(e.buf, "null"...)
		}

	default:
		// Fallback per unknown types
		e.buf = append(e.buf, '"')
		e.escapeStringFast(field.String)
		e.buf = append(e.buf, '"')
	}
}

// escapeStringFast escapes a string for JSON encoding (OPTIMIZED)
func (e *JSONEncoder) escapeStringFast(s string) {
	// OPTIMIZATION: Fast path per strings without special characters
	start := 0
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 32 || b == '"' || b == '\\' {
			// Found special character, flush previous good bytes
			if i > start {
				e.buf = append(e.buf, s[start:i]...)
			}

			// Handle special character
			switch b {
			case '"':
				e.buf = append(e.buf, `\"`...)
			case '\\':
				e.buf = append(e.buf, `\\`...)
			case '\n':
				e.buf = append(e.buf, `\n`...)
			case '\r':
				e.buf = append(e.buf, `\r`...)
			case '\t':
				e.buf = append(e.buf, `\t`...)
			default:
				// Other control characters - just append as is for now
				e.buf = append(e.buf, b)
			}
			start = i + 1
		}
	}

	// Append remaining good bytes
	if start < len(s) {
		e.buf = append(e.buf, s[start:]...)
	}
}

// encodeAnyTypeFast fast encoding per AnyType (avoid json.Marshal when possible)
func (e *JSONEncoder) encodeAnyTypeFast(value interface{}) {
	// CRITICAL OPTIMIZATION: Handle common types without reflection
	switch v := value.(type) {
	case string:
		e.buf = append(e.buf, '"')
		e.escapeStringFast(v)
		e.buf = append(e.buf, '"')
	case int:
		e.buf = strconv.AppendInt(e.buf, int64(v), 10)
	case int64:
		e.buf = strconv.AppendInt(e.buf, v, 10)
	case int32:
		e.buf = strconv.AppendInt(e.buf, int64(v), 10)
	case float64:
		e.buf = strconv.AppendFloat(e.buf, v, 'f', -1, 64)
	case float32:
		e.buf = strconv.AppendFloat(e.buf, float64(v), 'f', -1, 32)
	case bool:
		if v {
			e.buf = append(e.buf, "true"...)
		} else {
			e.buf = append(e.buf, "false"...)
		}
	case nil:
		e.buf = append(e.buf, "null"...)
	default:
		// Fallback to json.Marshal per complex types
		data, err := json.Marshal(value)
		if err != nil {
			e.buf = append(e.buf, "null"...)
		} else {
			e.buf = append(e.buf, data...)
		}
	}
}
