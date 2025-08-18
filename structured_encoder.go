// structured_encoder.go: Revolutionary structured encoding to beat Zap
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync"
	"time"
	"unsafe"
)

// StructuredField represents a field in binary format (ZAP-STYLE EFFICIENT)
type StructuredField struct {
	Key     string    // Key
	Type    FieldType // Type info
	IntVal  int64     // Unified integer storage
	StrVal  string    // String value (will optimize for zero-copy later)
	BoolVal bool      // Boolean value
}

// StructuredEncoder provides structured encoding WITHOUT JSON overhead
type StructuredEncoder struct {
	fields []StructuredField // Direct field storage
	cap    int               // Capacity
}

// Pool for StructuredField slices (ZAP-STYLE POOLING)
var structuredFieldPool = sync.Pool{
	New: func() interface{} {
		fields := make([]StructuredField, 0, 8) // Start with 8 fields capacity
		return &fields
	},
}

// NewStructuredEncoder creates zero-allocation structured encoder
func NewStructuredEncoder() *StructuredEncoder {
	fieldsPtr := structuredFieldPool.Get().(*[]StructuredField)
	fields := (*fieldsPtr)[:0] // Reset length, keep capacity

	return &StructuredEncoder{
		fields: fields,
		cap:    cap(fields),
	}
}

// AddString adds string field with efficient storage
func (e *StructuredEncoder) AddString(key, value string) {
	field := StructuredField{
		Key:    key,
		Type:   StringType,
		StrVal: value,
	}
	e.fields = append(e.fields, field)
}

// AddInt adds integer field
func (e *StructuredEncoder) AddInt(key string, value int64) {
	field := StructuredField{
		Key:    key,
		Type:   IntType,
		IntVal: value,
	}
	e.fields = append(e.fields, field)
}

// AddBool adds boolean field
func (e *StructuredEncoder) AddBool(key string, value bool) {
	field := StructuredField{
		Key:     key,
		Type:    BoolType,
		BoolVal: value,
	}
	e.fields = append(e.fields, field)
}

// MemoryFootprint returns the actual memory usage (like Zap)
func (e *StructuredEncoder) MemoryFootprint() int {
	size := int(unsafe.Sizeof(e.fields[0])) * len(e.fields)
	for _, field := range e.fields {
		size += len(field.Key) + len(field.StrVal) // Add string storage
	}
	return size
}

// ToJSON converts structured fields to JSON (LAZY CONVERSION)
func (e *StructuredEncoder) ToJSON(timestamp time.Time, level Level, message string, caller Caller) []byte {
	// Use existing JSONEncoder but feed it structured fields
	jsonEncoder := NewJSONEncoder()
	defer jsonEncoder.Reset()

	// Convert StructuredField to Field for JSON encoding
	fields := make([]Field, len(e.fields))
	for i, sfield := range e.fields {
		fields[i] = Field{
			Key:  sfield.Key,
			Type: sfield.Type,
		}

		switch sfield.Type {
		case StringType:
			fields[i].String = sfield.StrVal
		case IntType:
			fields[i].Int = sfield.IntVal
		case BoolType:
			fields[i].Bool = sfield.BoolVal
		}
	}

	// Generate JSON using existing optimized encoder
	jsonEncoder.EncodeLogEntry(timestamp, level, message, fields, caller, "")
	return jsonEncoder.Bytes()
}

// Reset returns structured encoder to pool
func (e *StructuredEncoder) Reset() {
	// Reset fields slice
	e.fields = e.fields[:0]
	// Return to pool
	structuredFieldPool.Put(&e.fields)
}
