// binary_helpers.go: Helper functions for BinaryField creation (MEMORY-SAFE)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

// StringBuffer manages string data for binary fields (LOCK-FREE)
// NOTE: Legacy struct kept for compatibility - no longer used in production
type StringBuffer struct {
	// Removed unused fields to fix staticcheck warnings
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
	// For int64 values, check if we need to handle very large positive values
	// that might look like unsigned when converted to uint64
	if safeData, ok := SafeInt64ToUint64(value); ok {
		return BinaryField{
			Key:   key,
			Value: "", // Not used for integers
			Type:  uint8(IntType),
			Data:  safeData,
		}
	}
	// For negative values that are very large in magnitude, we should still store them
	// using uint64 representation (two's complement)
	data, _ := SafeInt64ToUint64ForEncoding(value)
	return BinaryField{
		Key:   key,
		Value: "", // Not used for integers
		Type:  uint8(IntType),
		Data:  data, // Safe two's complement representation
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

// DESIGNER API: eXtreme performance field constructors (ULTRA-FAST)
func XStr(key string, value string) XField {
	return BinaryStr(key, value)
}

func XInt(key string, value int64) XField {
	return BinaryInt(key, value)
}

func XBool(key string, value bool) XField {
	return BinaryBool(key, value)
}
