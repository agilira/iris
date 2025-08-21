// binary_helpers.go: Helper functions for BinaryField creation (MEMORY-SAFE)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

// StringBuffer manages string data for binary fields (LOCK-FREE)
type StringBuffer struct {
	data []byte
	refs []StringRef
}

// StringRef represents a reference to string data in buffer
type StringRef struct {
	Offset uint32
	Length uint32
}

// LOCK-FREE APPROACH: No shared buffers, immutable fields only

// Legacy buffer functions - DEPRECATED (kept for compatibility)
// StringBuffer is no longer used in production code

// LOCK-FREE: BinaryField methods - immutable data

// BinaryStr creates a string field in binary format (LOCK-FREE)
func BinaryStr(key string, value string) BinaryField {
	return BinaryField{
		Key:   key,
		Value: value,
		Type:  uint8(StringType),
		Data:  0,
	}
}

// BinaryInt creates an integer field in binary format (LOCK-FREE)
func BinaryInt(key string, value int64) BinaryField {
	return BinaryField{
		Key:   key,
		Value: "", // Not used for integers
		Type:  uint8(IntType),
		Data:  uint64(value),
	}
}

// BinaryBool creates a boolean field in binary format (LOCK-FREE)
func BinaryBool(key string, value bool) BinaryField {
	var data uint64 = 0
	if value {
		data = 1
	}

	return BinaryField{
		Key:   key,
		Value: "", // Not used for booleans
		Type:  uint8(BoolType),
		Data:  data,
	}
}

// SAFE MIGRATION: Short aliases for user-friendly API (BACKWARD COMPATIBLE)
func BStr(key string, value string) BField {
	return BinaryStr(key, value)
}

func BInt(key string, value int64) BField {
	return BinaryInt(key, value)
}

func BBool(key string, value bool) BField {
	return BinaryBool(key, value)
}
