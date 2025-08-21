// binary_types.go: Binary logging types and structures (MEMORY-SAFE)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

// BinaryField represents a field in pure binary format (GC-SAFE FOR 104M/s)
type BinaryField struct {
	Buffer   *StringBuffer // Buffer reference (GC-safe)
	KeyRef   StringRef     // Key reference in buffer
	ValueRef StringRef     // Value reference (for strings)
	Type     uint8         // Field type
	Data     uint64        // Primitive data (int, bool, etc.)
}

// GetKey retrieves key string (GC-SAFE)
func (bf BinaryField) GetKey() string {
	if bf.Buffer == nil {
		return ""
	}
	return bf.Buffer.GetString(bf.KeyRef)
}

// GetString retrieves string value (GC-SAFE)
func (bf BinaryField) GetString() string {
	if bf.Buffer == nil || bf.Type != uint8(StringType) {
		return ""
	}
	return bf.Buffer.GetString(bf.ValueRef)
}

// GetInt retrieves integer value
func (bf BinaryField) GetInt() int64 {
	return int64(bf.Data)
}

// GetBool retrieves boolean value
func (bf BinaryField) GetBool() bool {
	return bf.Data != 0
}

// GetFloat retrieves float value
func (bf BinaryField) GetFloat() float64 {
	return float64(bf.Data)
}

// Release releases buffer back to pool (CRITICAL for 104M/s)
func (bf BinaryField) Release() {
	if bf.Buffer != nil {
		ReleaseStringBuffer(bf.Buffer)
	}
}

// BinaryEntry represents a complete log entry in binary format (SAFE)
type BinaryEntry struct {
	Timestamp uint64   // Unix nano timestamp
	Level     uint8    // Log level (1 byte)
	_         [3]uint8 // Padding for alignment
	MsgBuffer *StringBuffer // Message buffer (GC-safe)
	MsgRef    StringRef     // Message reference
	Fields    []BinaryField // Fields array (GC-safe)
}

// BinaryContext represents a binary logging context (SAFE)
type BinaryContext struct {
	logger *BinaryLogger
	fields []BinaryField
}
