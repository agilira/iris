// binary_helpers.go: Helper functions for BinaryField creation
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import "unsafe"

// TASK S1: Direct binary field methods (OPTIMAL VERSION)

// BinaryStr creates binary string field directly (no Field conversion)
func BinaryStr(key, value string) BinaryField {
	strPtr := uintptr(unsafe.Pointer(&value))
	strLen := uint64(len(value))
	return BinaryField{
		KeyPtr: uintptr(unsafe.Pointer(&key)),
		KeyLen: uint16(len(key)),
		Type:   uint8(StringType),
		Data:   (uint64(strPtr) << 32) | strLen,
	}
}

// BinaryInt creates binary int field directly (no Field conversion)
func BinaryInt(key string, value int64) BinaryField {
	return BinaryField{
		KeyPtr: uintptr(unsafe.Pointer(&key)),
		KeyLen: uint16(len(key)),
		Type:   uint8(IntType),
		Data:   uint64(value),
	}
}

// BinaryBool creates binary bool field directly (no Field conversion)
func BinaryBool(key string, value bool) BinaryField {
	data := uint64(0)
	if value {
		data = 1
	}
	return BinaryField{
		KeyPtr: uintptr(unsafe.Pointer(&key)),
		KeyLen: uint16(len(key)),
		Type:   uint8(BoolType),
		Data:   data,
	}
}
