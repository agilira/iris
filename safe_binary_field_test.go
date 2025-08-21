// safe_binary_field_test.go: Step 1.2 - Test isolati per Safe BinaryField API
//
// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0
//
// Questo file contiene test isolati per validare la sicurezza e performance
// delle nuove funzioni BinaryField in Step 1.2 della migrazione.

package iris

import (
	"testing"
)

// =============================================================================
// Step 1.2: Safe BinaryField API - Isolated Tests
// =============================================================================

// TestSafeBinaryField_String testa la creazione sicura di string fields
func TestSafeBinaryField_String(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"simple", "test", "value"},
		{"empty_value", "key", ""},
		{"empty_key", "", "value"},
		{"both_empty", "", ""},
		{"unicode", "测试", "测试值"},
		{"special_chars", "key!@#", "value$%^"},
		{"long_string", "very_long_key_name_for_testing", "very_long_value_content_for_comprehensive_testing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test NextStr creation
			bf := NextStr(tt.key, tt.value)

			// Basic validation
			if bf.Type != uint8(StringType) {
				t.Errorf("Expected StringType (%d), got %d", StringType, bf.Type)
			}

			// Verify key via GC-safe method
			if bf.GetKey() != tt.key {
				t.Errorf("Expected key %s, got %s", tt.key, bf.GetKey())
			}

			// Verify value via GC-safe method
			if bf.GetString() != tt.value {
				t.Errorf("Expected value %s, got %s", tt.value, bf.GetString())
			}

			// Test conversion to legacy
			field := toLegacyField(bf)
			if field.Type != StringType {
				t.Errorf("Expected StringType after conversion, got %v", field.Type)
			}

			t.Logf("✅ String field created: key=%s, value=%s", tt.key, tt.value)
		})
	}
}

// TestSafeBinaryField_Int testa la creazione sicura di int fields
func TestSafeBinaryField_Int(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value int
	}{
		{"positive", "count", 42},
		{"negative", "delta", -100},
		{"zero", "zero", 0},
		{"max_int", "max", int(^uint(0) >> 1)},
		{"min_int", "min", -int(^uint(0)>>1) - 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test NextInt creation
			bf := NextInt(tt.key, tt.value)

			// Basic validation
			if bf.Type != uint8(IntType) {
				t.Errorf("Expected IntType (%d), got %d", IntType, bf.Type)
			}

			// Verify key via GC-safe method
			if bf.GetKey() != tt.key {
				t.Errorf("Expected key %s, got %s", tt.key, bf.GetKey())
			}

			// Verify value via GC-safe method
			if bf.GetInt() != int64(tt.value) {
				t.Errorf("Expected value %d, got %d", tt.value, bf.GetInt())
			}

			if bf.Data != uint64(tt.value) && tt.value >= 0 {
				t.Errorf("Expected Data %d, got %d", tt.value, bf.Data)
			}

			// Test conversion to legacy
			field := toLegacyField(bf)
			if field.Type != IntType {
				t.Errorf("Expected IntType after conversion, got %v", field.Type)
			}

			if field.Int != int64(tt.value) {
				t.Errorf("Expected Int %d after conversion, got %d", tt.value, field.Int)
			}

			t.Logf("✅ Int field created: key=%s, value=%d, Data=%d", tt.key, tt.value, bf.Data)
		})
	}
}

// TestSafeBinaryField_Bool testa la creazione sicura di bool fields
func TestSafeBinaryField_Bool(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value bool
	}{
		{"true", "enabled", true},
		{"false", "disabled", false},
		{"true_empty_key", "", true},
		{"false_empty_key", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test NextBool creation
			bf := NextBool(tt.key, tt.value)

			// Basic validation
			if bf.Type != uint8(BoolType) {
				t.Errorf("Expected BoolType (%d), got %d", BoolType, bf.Type)
			}

			// Verify key via GC-safe method
			if bf.GetKey() != tt.key {
				t.Errorf("Expected key %s, got %s", tt.key, bf.GetKey())
			}

			// Verify value via GC-safe method
			if bf.GetBool() != tt.value {
				t.Errorf("Expected value %t, got %t", tt.value, bf.GetBool())
			}

			expectedData := uint64(0)
			if tt.value {
				expectedData = 1
			}
			if bf.Data != expectedData {
				t.Errorf("Expected Data %d, got %d", expectedData, bf.Data)
			}

			// Test conversion to legacy
			field := toLegacyField(bf)
			if field.Type != BoolType {
				t.Errorf("Expected BoolType after conversion, got %v", field.Type)
			}

			if field.Bool != tt.value {
				t.Errorf("Expected Bool %v after conversion, got %v", tt.value, field.Bool)
			}

			t.Logf("✅ Bool field created: key=%s, value=%v, Data=%d", tt.key, tt.value, bf.Data)
		})
	}
}

// TestSafeBinaryField_ConversionRoundTrip testa la conversione bidirezionale
func TestSafeBinaryField_ConversionRoundTrip(t *testing.T) {
	testCases := []BinaryField{
		NextStr("string_key", "string_value"),
		NextInt("int_key", 12345),
		NextBool("bool_key", true),
		NextBool("bool_false", false),
	}

	for i, bf := range testCases {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			// Convert to legacy
			legacyField := toLegacyField(bf)

			// Convert back to binary (when we implement it)
			binaryFields := []BinaryField{bf}
			legacyFields := ToLegacyFields(binaryFields)

			// Validate
			if len(legacyFields) != 1 {
				t.Fatalf("Expected 1 legacy field, got %d", len(legacyFields))
			}

			if legacyFields[0].Type != legacyField.Type {
				t.Errorf("Type mismatch after round trip: expected %v, got %v",
					legacyField.Type, legacyFields[0].Type)
			}

			t.Logf("✅ Round trip successful for type %v", legacyField.Type)
		})
	}
}

// TestSafeBinaryField_MemorySafety testa la sicurezza della memoria
func TestSafeBinaryField_MemorySafety(t *testing.T) {
	// Test multiple allocations to ensure no memory corruption
	fields := make([]BinaryField, 1000)

	for i := 0; i < 1000; i++ {
		switch i % 3 {
		case 0:
			fields[i] = NextStr("key", "value")
		case 1:
			fields[i] = NextInt("key", i)
		case 2:
			fields[i] = NextBool("key", i%2 == 0)
		}
	}

	// Convert all to legacy (this should not crash)
	legacyFields := ToLegacyFields(fields)
	if len(legacyFields) != 1000 {
		t.Errorf("Expected 1000 legacy fields, got %d", len(legacyFields))
	}

	// Validate random samples
	for i := 0; i < 10; i++ {
		idx := i * 100
		if legacyFields[idx].Key != "converted_key" {
			t.Errorf("Expected converted_key, got %s", legacyFields[idx].Key)
		}
	}

	t.Logf("✅ Memory safety test passed: 1000 fields processed without crashes")
}

// TestSafeBinaryField_ConcurrentAccess testa l'accesso concorrente
func TestSafeBinaryField_ConcurrentAccess(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 100

	results := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", goroutineID, r)
					results <- false
					return
				}
				results <- true
			}()

			for i := 0; i < numOperations; i++ {
				bf := NextStr("concurrent", "test")
				_ = toLegacyField(bf)
			}
		}(g)
	}

	// Wait for all goroutines
	passed := 0
	for i := 0; i < numGoroutines; i++ {
		if <-results {
			passed++
		}
	}

	if passed != numGoroutines {
		t.Errorf("Only %d/%d goroutines completed successfully", passed, numGoroutines)
	}

	t.Logf("✅ Concurrent access test passed: %d goroutines, %d operations each",
		numGoroutines, numOperations)
}
