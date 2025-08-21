// binary_helpers.go: Helper functions for BinaryField creation (MEMORY-SAFE)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync"
)

// StringBuffer manages string data for binary fields (GC-SAFE)
type StringBuffer struct {
	data []byte
	refs []StringRef
}

// StringRef represents a reference to string data in buffer
type StringRef struct {
	Offset uint32
	Length uint32
}

// StringBufferPool manages reusable string buffers (ZAP-STYLE)
var stringBufferPool = sync.Pool{
	New: func() interface{} {
		return &StringBuffer{
			data: make([]byte, 0, 2048), // Larger for 104M/s throughput
			refs: make([]StringRef, 0, 32),
		}
	},
}

// GetStringBuffer gets a buffer from pool
func GetStringBuffer() *StringBuffer {
	return stringBufferPool.Get().(*StringBuffer)
}

// ReleaseStringBuffer returns buffer to pool
func ReleaseStringBuffer(buf *StringBuffer) {
	buf.data = buf.data[:0]
	buf.refs = buf.refs[:0]
	stringBufferPool.Put(buf)
}

// AddString adds string to buffer and returns reference (GC-SAFE)
func (sb *StringBuffer) AddString(s string) StringRef {
	offset := uint32(len(sb.data))
	sb.data = append(sb.data, s...)
	ref := StringRef{
		Offset: offset,
		Length: uint32(len(s)),
	}
	sb.refs = append(sb.refs, ref)
	return ref
}

// GetString retrieves string from buffer using reference
func (sb *StringBuffer) GetString(ref StringRef) string {
	return string(sb.data[ref.Offset : ref.Offset+ref.Length])
}

// MEMORY-SAFE: BinaryField methods with buffer pool

// BinaryStr creates a string field in binary format (GC-SAFE)
func BinaryStr(key string, value string) BinaryField {
	buffer := GetStringBuffer()
	keyRef := buffer.AddString(key)
	valueRef := buffer.AddString(value)
	
	return BinaryField{
		Buffer:   buffer,
		KeyRef:   keyRef,
		ValueRef: valueRef,
		Type:     uint8(StringType),
		Data:     0,
	}
}

// BinaryInt creates an integer field in binary format (GC-SAFE)
func BinaryInt(key string, value int64) BinaryField {
	buffer := GetStringBuffer()
	keyRef := buffer.AddString(key)
	
	return BinaryField{
		Buffer:   buffer,
		KeyRef:   keyRef,
		ValueRef: StringRef{}, // Not used for integers
		Type:     uint8(IntType),
		Data:     uint64(value),
	}
}

// BinaryBool creates a boolean field in binary format (GC-SAFE)
func BinaryBool(key string, value bool) BinaryField {
	buffer := GetStringBuffer()
	keyRef := buffer.AddString(key)
	
	var data uint64 = 0
	if value {
		data = 1
	}
	
	return BinaryField{
		Buffer:   buffer,
		KeyRef:   keyRef,
		ValueRef: StringRef{}, // Not used for booleans
		Type:     uint8(BoolType),
		Data:     data,
	}
}
