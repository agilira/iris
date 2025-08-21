// field_types.go: Field types and constants for Iris logging
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

// FieldType represents the type of a field value (optimized enum)
type FieldType uint8

// OPTIMIZATION: Grouped by frequency of use (String/Int most common)
const (
	StringType FieldType = iota
	IntType
	BoolType    // Moved up - 3rd most common
	Float64Type // Moved up - 4th most common
	Int64Type
	Int32Type
	Int16Type
	Int8Type
	UintType
	Uint64Type
	Uint32Type
	Uint16Type
	Uint8Type
	Float32Type
	DurationType
	TimeType
	ErrorType
	ByteStringType
	BinaryType
	AnyType
)

// Field represents a structured log field (memory-optimized layout)
type Field struct {
	Key    string      // Most common access - first for cache efficiency
	Type   FieldType   // Enum type - compact
	String string      // String value - most common type
	Int    int64       // Integer value - second most common
	Float  float64     // Float value - third most common
	Bool   bool        // Boolean value - fourth most common
	Err    error       // Error interface
	Bytes  []byte      // For ByteString and Binary types
	Any    interface{} // For Any type - least common, last
}
