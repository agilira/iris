// migration_helpers.go: Safe helpers for BinaryField migration
//
// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0

package iris

import "unsafe"

// Migration context to store original Field data for safe access
var migrationFieldContext = make(map[*BinaryField]*Field)

// storeMigrationContext stores the original Field for safe access
func storeMigrationContext(bf *BinaryField, original *Field) {
	migrationFieldContext[bf] = original
}

// getMigrationContext retrieves the original Field for safe access
func getMigrationContext(bf *BinaryField) *Field {
	return migrationFieldContext[bf]
}

// Safe accessor methods for migration phase
func (bf *BinaryField) GetKeySafe() string {
	if original := getMigrationContext(bf); original != nil {
		return original.Key
	}
	return ""
}

func (bf *BinaryField) GetStringSafe() string {
	if original := getMigrationContext(bf); original != nil {
		if original.Type == StringType {
			return original.String
		}
		if original.Type == ErrorType && original.Err != nil {
			return original.Err.Error()
		}
		if original.Type == ByteStringType {
			return string(original.Bytes)
		}
	}
	return ""
}

func (bf *BinaryField) GetIntSafe() int64 {
	if original := getMigrationContext(bf); original != nil {
		return original.Int
	}
	return int64(bf.Data)
}

func (bf *BinaryField) GetBoolSafe() bool {
	if original := getMigrationContext(bf); original != nil {
		return original.Bool
	}
	return bf.Data == 1
}

func (bf *BinaryField) GetFloatSafe() float64 {
	if original := getMigrationContext(bf); original != nil {
		return original.Float
	}
	return *(*float64)(unsafe.Pointer(&bf.Data))
}
