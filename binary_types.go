// binary_types.go: Binary logging types and structures
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

// BinaryField represents a field in pure binary format (ULTRA-LIGHTWEIGHT)
type BinaryField struct {
	KeyPtr uintptr // Key pointer (zero-copy)
	KeyLen uint16  // Key length
	Type   uint8   // Field type (1 byte)
	_      uint8   // Padding for alignment
	Data   uint64  // Union data: int64, bool, or string ptr+len
}

// BinaryEntry represents a complete log entry in binary format
type BinaryEntry struct {
	Timestamp uint64   // Unix nano timestamp
	Level     uint8    // Log level (1 byte)
	_         [3]uint8 // Padding for alignment
	MsgPtr    uintptr  // Message pointer (zero-copy)
	MsgLen    uint32   // Message length
	FieldPtr  uintptr  // Fields array pointer
	FieldCnt  uint16   // Field count
	_         uint16   // Padding
}

// BinaryContext represents a binary logging context
type BinaryContext struct {
	logger *BinaryLogger
	fields []BinaryField
}
