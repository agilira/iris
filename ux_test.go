// ux_test.go: Test per API user-friendly super clean
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
)

// TestUXFriendlyAPI tests the super clean user experience API
func TestUXFriendlyAPI(t *testing.T) {
	// SUPER CLEAN API - UX optimized!
	logger := NewBLogger(InfoLevel) // Short name!

	ctx := logger.WithBFields( // Clean method name!
		BStr("user", "john"),  // Short constructor!
		BInt("age", 25),       // Short constructor!
		BBool("active", true), // Short constructor!
	)

	ctx.Info("Super clean API!") // Standard method!

	// Test that types are identical
	var logger2 *BLogger = NewBLogger(InfoLevel)
	var logger3 *BinaryLogger = NewBinaryLogger(InfoLevel)

	// Should be the same type under the hood
	if logger2 == nil || logger3 == nil {
		t.Error("Loggers should be valid")
	}

	// Test context types
	var ctx1 *BContext = logger.WithBFields(BStr("test", "value"))
	var ctx2 *BinaryContext = logger.WithBinaryFields(BStr("test", "value"))

	if ctx1 == nil || ctx2 == nil {
		t.Error("Contexts should be valid")
	}
}

// TestTypeAliasesEquivalence ensures aliases work identically to original types
func TestTypeAliasesEquivalence(t *testing.T) {
	// Test field aliases
	bf1 := BStr("key", "value")
	var bf2 BField = bf1
	var bf3 BinaryField = bf1

	if bf2.GetKey() != bf3.GetKey() {
		t.Error("BField and BinaryField should be identical")
	}

	// Test logger aliases
	logger1 := NewBLogger(InfoLevel)
	logger2 := NewBinaryLogger(InfoLevel)

	// Should be able to use interchangeably
	ctx1 := logger1.WithBFields(BStr("test", "1"))
	ctx2 := logger2.WithBFields(BStr("test", "2"))

	if ctx1 == nil || ctx2 == nil {
		t.Error("Both logger types should work identically")
	}
}
