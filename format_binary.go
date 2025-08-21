// format_binary.go: Ultra-fast binary encoder for Iris logging
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"time"
	"unsafe"
)

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
func (e *BinaryEncoder) EncodeLogEntry(timestamp time.Time, level Level, message string, fields []BinaryField, caller Caller, stackTrace string) {
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
func (e *BinaryEncoder) encodeBinaryFieldFast(field BinaryField) {
	// OPTIMIZATION 1: Batch header encoding (type + keylen)
	keyLen := len(field.GetKey())
	if keyLen > 255 {
		keyLen = 255
	}

	// Field format: [type:1byte][keylen:1byte][key][value...]
	header := [2]byte{byte(field.Type), byte(keyLen)}
	e.buf = append(e.buf, header[:]...)
	e.buf = append(e.buf, field.GetKey()[:keyLen]...)

	// OPTIMIZATION 2: Fast paths for hot types with reduced branching
	switch field.Type {
	case uint8(StringType), uint8(ByteStringType):
		e.encodeBinaryStringValue(field)
	case uint8(IntType), uint8(Int64Type):
		// FAST PATH: Integer encoding (second most common)
		e.encodeBinaryInt64Value(field.GetInt())
	case uint8(BoolType):
		e.encodeBinaryBoolValue(field.GetBool())
	case uint8(Float64Type), uint8(Float32Type):
		e.encodeBinaryFloatValue(field)
	case uint8(Int32Type), uint8(Int16Type), uint8(Int8Type), uint8(UintType), uint8(Uint64Type), uint8(Uint32Type), uint8(Uint16Type), uint8(Uint8Type):
		e.encodeBinaryInt64Value(field.GetInt())
	case uint8(DurationType), uint8(TimeType):
		e.encodeBinaryInt64Value(field.GetInt())
	case uint8(SecretType):
		// SECURITY: Store redacted marker in binary format
		redactedField := BinaryField{
			Key:   field.GetKey(),
			Value: "[REDACTED]",
			Type:  uint8(StringType),
		}
		e.encodeBinaryStringValue(redactedField)
	default:
		// Fallback for unknown types - encode as empty
		e.buf = append(e.buf, 0, 0) // zero-length value
	}
}

// EncodeLogEntryMigration encodes directly from Field slice (MIGRATION METHOD)
func (e *BinaryEncoder) EncodeLogEntryMigration(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) {
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

	// Ultra-fast field encoding (DIRECT FROM FIELD)
	for i := 0; i < fieldCount; i++ {
		e.encodeBinaryFieldFastMigration(fields[i])
	}
}

// encodeBinaryFieldFastMigration encodes directly from Field (MIGRATION METHOD)
func (e *BinaryEncoder) encodeBinaryFieldFastMigration(field Field) {
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
	case StringType, ByteStringType:
		e.encodeBinaryStringValueFromField(field)
	case IntType, Int64Type, Int32Type, Int16Type, Int8Type, UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		e.encodeBinaryInt64Value(field.Int)
	case BoolType:
		e.encodeBinaryBoolValue(field.Bool)
	case Float64Type, Float32Type:
		e.encodeBinaryFloatValueFromField(field)
	case DurationType, TimeType:
		e.encodeBinaryInt64Value(field.Int)
	case SecretType:
		// SECURITY: Store redacted marker in binary format (migration path)
		redacted := "[REDACTED]"
		redactedLen := len(redacted)
		strHeader := [2]byte{byte(redactedLen >> 8), byte(redactedLen)}
		e.buf = append(e.buf, strHeader[:]...)
		e.buf = append(e.buf, redacted...)
	default:
		// Fallback for unknown types - encode as empty
		e.buf = append(e.buf, 0, 0) // zero-length value
	}
}

// encodeBinaryStringValue encodes string or byte string value
func (e *BinaryEncoder) encodeBinaryStringValue(field BinaryField) {
	str := field.GetString()
	strLen := len(str)
	if strLen > 65535 {
		strLen = 65535
	}

	// Batch string length and content
	strHeader := [2]byte{byte(strLen >> 8), byte(strLen)}
	e.buf = append(e.buf, strHeader[:]...)
	e.buf = append(e.buf, str[:strLen]...)
}

// encodeBinaryInt64Value encodes 64-bit integer value
func (e *BinaryEncoder) encodeBinaryInt64Value(value int64) {
	val, _ := SafeInt64ToUint64ForEncoding(value)
	intBytes := [8]byte{
		byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
		byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
	}
	e.buf = append(e.buf, intBytes[:]...)
}

// encodeBinaryBoolValue encodes boolean value
func (e *BinaryEncoder) encodeBinaryBoolValue(value bool) {
	if value {
		e.buf = append(e.buf, 1)
	} else {
		e.buf = append(e.buf, 0)
	}
}

// encodeBinaryFloatValue encodes float32 or float64 value
func (e *BinaryEncoder) encodeBinaryFloatValue(field BinaryField) {
	if field.Type == uint8(Float64Type) {
		// OPTIMIZED: Float64 using unsafe for speed
		// #nosec G103 - unsafe.Pointer required for zero-allocation float64 binary encoding
		val := *(*uint64)(unsafe.Pointer(&field.Data))
		floatBytes := [8]byte{
			byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		}
		e.buf = append(e.buf, floatBytes[:]...)
	} else {
		// 4 bytes for float32
		floatVal := float32(field.GetFloat())
		// #nosec G103 - unsafe.Pointer required for zero-allocation float32 binary encoding
		val := *(*uint32)(unsafe.Pointer(&floatVal))
		floatBytes := [4]byte{
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		}
		e.buf = append(e.buf, floatBytes[:]...)
	}
}

// encodeBinaryStringValueFromField encodes string from Field struct
func (e *BinaryEncoder) encodeBinaryStringValueFromField(field Field) {
	var str string
	var strLen int

	if field.Type == StringType {
		str = field.String
		strLen = len(field.String)
	} else {
		// ByteStringType
		str = string(field.Bytes)
		strLen = len(field.Bytes)
	}

	if strLen > 65535 {
		strLen = 65535
	}

	// Batch string length and content
	strHeader := [2]byte{byte(strLen >> 8), byte(strLen)}
	e.buf = append(e.buf, strHeader[:]...)
	e.buf = append(e.buf, str[:strLen]...)
}

// encodeBinaryFloatValueFromField encodes float from Field struct
func (e *BinaryEncoder) encodeBinaryFloatValueFromField(field Field) {
	if field.Type == Float64Type {
		// OPTIMIZED: Float64 using unsafe for speed
		// #nosec G103 - unsafe.Pointer required for zero-allocation float64 binary encoding
		val := *(*uint64)(unsafe.Pointer(&field.Float))
		floatBytes := [8]byte{
			byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		}
		e.buf = append(e.buf, floatBytes[:]...)
	} else {
		// 4 bytes for float32
		// #nosec G103 - unsafe.Pointer required for zero-allocation float32 binary encoding
		val := *(*uint32)(unsafe.Pointer(&field.Float))
		floatBytes := [4]byte{
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		}
		e.buf = append(e.buf, floatBytes[:]...)
	}
}
