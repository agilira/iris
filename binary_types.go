// binary_types.go: Binary logging types and structures (MEMORY-SAFE)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

// BinaryField represents a field in pure binary format (LOCK-FREE IMMUTABLE)
type BinaryField struct {
	Key   string // Direct string storage - no shared buffer
	Value string // Direct value storage - immutable
	Type  uint8  // Field type
	Data  uint64 // Primitive data (int, bool, etc.)
}

// GetKey retrieves key string (ZERO-COPY)
func (bf BinaryField) GetKey() string {
	return bf.Key
}

// GetString retrieves string value (ZERO-COPY)
func (bf BinaryField) GetString() string {
	if bf.Type != uint8(StringType) {
		return ""
	}
	return bf.Value
}

// GetInt retrieves integer value safely
func (bf BinaryField) GetInt() int64 {
	// Use the field type information to determine if safe conversion is needed
	fieldType := FieldType(bf.Type)
	if safeValue, ok := SafeBinaryDataToInt64(bf.Data, fieldType); ok {
		return safeValue
	}
	// If conversion is not safe, return 0 (this should not happen in practice
	// if the BinaryField was created correctly)
	return 0
}

// GetBool retrieves boolean value
func (bf BinaryField) GetBool() bool {
	return bf.Data != 0
}

// GetFloat retrieves float value
func (bf BinaryField) GetFloat() float64 {
	return float64(bf.Data)
}

// Release returns resources to pool (NO-OP for immutable fields)
func (bf BinaryField) Release() {
	// Immutable fields don't need release - GC handles cleanup
}

// BinaryEntry represents a complete log entry in binary format (LOCK-FREE)
type BinaryEntry struct {
	Timestamp uint64        // Unix nano timestamp
	Level     uint8         // Log level (1 byte)
	_         [3]uint8      // Padding for alignment
	Message   string        // Direct message storage (immutable)
	Fields    []BinaryField // Fields array (immutable)
}

// BinaryContext represents a binary logging context (SAFE)
type BinaryContext struct {
	logger *BinaryLogger
	fields []BinaryField
}

// SAFE MIGRATION: Alias per gradual transition (BACKWARD COMPATIBLE)
type BField = BinaryField     // Alias to BinaryField - zero overhead
type BContext = BinaryContext // Alias to BinaryContext - user friendly
type BLogger = BinaryLogger   // Alias to BinaryLogger - clean API
type BEntry = BinaryEntry     // Alias to BinaryEntry - short name
