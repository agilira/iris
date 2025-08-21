// migration_helpers.go: Safe helpers for BinaryField migration
//
// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0

package iris

// Migration context to store original Field data for safe access
var migrationFieldContext = make(map[*BinaryField]*Field)

// getMigrationContext retrieves the original Field for safe access
func getMigrationContext(bf *BinaryField) *Field {
	return migrationFieldContext[bf]
}
