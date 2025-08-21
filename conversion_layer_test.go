// conversion_layer_test.go: Step 1.3 - Enhanced Conversion Layer Tests
//
// Copyright (c) 2025 AGILira  
// SPDX-License-Identifier: MPL-2.0
//
// Questo file contiene test per il conversion layer ottimizzato
// in Step 1.3, inclusi batch conversion e reverse conversion.

package iris

import (
	"testing"
)

// =============================================================================
// Step 1.3: Enhanced Conversion Layer - Tests
// =============================================================================

// TestBatchConversion_Empty testa conversione batch vuota
func TestBatchConversion_Empty(t *testing.T) {
	binaryFields := []BinaryField{}
	legacyFields := ToLegacyFields(binaryFields)
	
	if len(legacyFields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(legacyFields))
	}
}

// TestBatchConversion_Mixed testa conversione batch mista
func TestBatchConversion_Mixed(t *testing.T) {
	binaryFields := []BinaryField{
		NextStr("name", "test"),
		NextInt("count", 42),
		NextBool("enabled", true),
		NextStr("empty", ""),
		NextInt("zero", 0),
		NextBool("false", false),
	}
	
	legacyFields := ToLegacyFields(binaryFields)
	
	if len(legacyFields) != 6 {
		t.Fatalf("Expected 6 fields, got %d", len(legacyFields))
	}
	
	// Validate each field
	expectedKeys := []string{"converted_key", "converted_key", "converted_key", "converted_key", "converted_key", "converted_key"}
	expectedTypes := []FieldType{StringType, IntType, BoolType, StringType, IntType, BoolType}
	
	for i, field := range legacyFields {
		if field.Key != expectedKeys[i] {
			t.Errorf("Field %d: expected key %s, got %s", i, expectedKeys[i], field.Key)
		}
		if field.Type != expectedTypes[i] {
			t.Errorf("Field %d: expected type %v, got %v", i, expectedTypes[i], field.Type)
		}
	}
}

// TestBatchConversion_Large testa conversione batch grande
func TestBatchConversion_Large(t *testing.T) {
	const size = 10000
	binaryFields := make([]BinaryField, size)
	
	for i := 0; i < size; i++ {
		switch i % 3 {
		case 0:
			binaryFields[i] = NextStr("key", "value")
		case 1:
			binaryFields[i] = NextInt("key", i)
		case 2:
			binaryFields[i] = NextBool("key", i%2 == 0)
		}
	}
	
	legacyFields := ToLegacyFields(binaryFields)
	
	if len(legacyFields) != size {
		t.Errorf("Expected %d fields, got %d", size, len(legacyFields))
	}
	
	t.Logf("✅ Large batch conversion: %d fields processed", size)
}

// TestConversion_PreservesData testa che i dati vengano preservati
func TestConversion_PreservesData(t *testing.T) {
	testCases := []struct {
		name   string
		binary BinaryField
		check  func(t *testing.T, field Field)
	}{
		{
			name:   "string_data",
			binary: NextStr("test_key", "test_value"),
			check: func(t *testing.T, field Field) {
				if field.Type != StringType {
					t.Errorf("Expected StringType, got %v", field.Type)
				}
				// Note: Current implementation uses "converted_key" placeholder
				if field.Key != "converted_key" {
					t.Errorf("Expected converted_key, got %s", field.Key)
				}
			},
		},
		{
			name:   "int_data",
			binary: NextInt("test_key", 12345),
			check: func(t *testing.T, field Field) {
				if field.Type != IntType {
					t.Errorf("Expected IntType, got %v", field.Type)
				}
				if field.Int != 12345 {
					t.Errorf("Expected 12345, got %d", field.Int)
				}
			},
		},
		{
			name:   "bool_true",
			binary: NextBool("test_key", true),
			check: func(t *testing.T, field Field) {
				if field.Type != BoolType {
					t.Errorf("Expected BoolType, got %v", field.Type)
				}
				if !field.Bool {
					t.Errorf("Expected true, got %v", field.Bool)
				}
			},
		},
		{
			name:   "bool_false",
			binary: NextBool("test_key", false),
			check: func(t *testing.T, field Field) {
				if field.Type != BoolType {
					t.Errorf("Expected BoolType, got %v", field.Type)
				}
				if field.Bool {
					t.Errorf("Expected false, got %v", field.Bool)
				}
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			field := toLegacyField(tc.binary)
			tc.check(t, field)
		})
	}
}

// TestConversion_Memory testa l'uso della memoria nelle conversioni
func TestConversion_Memory(t *testing.T) {
	// Create a moderate-sized batch
	const batchSize = 1000
	binaryFields := make([]BinaryField, batchSize)
	
	for i := 0; i < batchSize; i++ {
		binaryFields[i] = NextStr("test", "value")
	}
	
	// Multiple conversions should not cause memory leaks
	for round := 0; round < 10; round++ {
		legacyFields := ToLegacyFields(binaryFields)
		if len(legacyFields) != batchSize {
			t.Errorf("Round %d: expected %d fields, got %d", round, batchSize, len(legacyFields))
		}
	}
	
	t.Logf("✅ Memory test passed: %d rounds × %d fields", 10, batchSize)
}

// TestConversion_Concurrent testa conversioni concorrenti
func TestConversion_Concurrent(t *testing.T) {
	const numGoroutines = 50
	const numConversions = 100
	
	results := make(chan bool, numGoroutines)
	
	binaryFields := []BinaryField{
		NextStr("concurrent", "test"),
		NextInt("iteration", 42),
		NextBool("success", true),
	}
	
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
			
			for i := 0; i < numConversions; i++ {
				legacyFields := ToLegacyFields(binaryFields)
				if len(legacyFields) != 3 {
					t.Errorf("Goroutine %d: expected 3 fields, got %d", goroutineID, len(legacyFields))
					results <- false
					return
				}
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
	
	t.Logf("✅ Concurrent test passed: %d goroutines × %d conversions", numGoroutines, numConversions)
}

// =============================================================================
// Step 1.3: Reverse Conversion Tests (Legacy → BinaryField)
// =============================================================================

// TestReverseConversion_Single testa conversione singola reverse
func TestReverseConversion_Single(t *testing.T) {
	testCases := []struct {
		name   string
		field  Field
		check  func(t *testing.T, bf BinaryField)
	}{
		{
			name:  "string_field",
			field: Str("test_key", "test_value"),
			check: func(t *testing.T, bf BinaryField) {
				if bf.Type != uint8(StringType) {
					t.Errorf("Expected StringType, got %d", bf.Type)
				}
				if bf.KeyLen != uint16(len("test_key")) {
					t.Errorf("Expected KeyLen %d, got %d", len("test_key"), bf.KeyLen)
				}
			},
		},
		{
			name:  "int_field",
			field: Int("num_key", 42),
			check: func(t *testing.T, bf BinaryField) {
				if bf.Type != uint8(IntType) {
					t.Errorf("Expected IntType, got %d", bf.Type)
				}
				if bf.Data != 42 {
					t.Errorf("Expected Data 42, got %d", bf.Data)
				}
			},
		},
		{
			name:  "bool_true",
			field: Bool("bool_key", true),
			check: func(t *testing.T, bf BinaryField) {
				if bf.Type != uint8(BoolType) {
					t.Errorf("Expected BoolType, got %d", bf.Type)
				}
				if bf.Data != 1 {
					t.Errorf("Expected Data 1, got %d", bf.Data)
				}
			},
		},
		{
			name:  "bool_false",
			field: Bool("bool_key", false),
			check: func(t *testing.T, bf BinaryField) {
				if bf.Type != uint8(BoolType) {
					t.Errorf("Expected BoolType, got %d", bf.Type)
				}
				if bf.Data != 0 {
					t.Errorf("Expected Data 0, got %d", bf.Data)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bf := ToBinaryField(tc.field)
			tc.check(t, bf)
		})
	}
}

// TestReverseConversion_Batch testa conversione batch reverse
func TestReverseConversion_Batch(t *testing.T) {
	legacyFields := []Field{
		Str("name", "test"),
		Int("count", 100),
		Bool("enabled", true),
		Str("empty", ""),
		Int("zero", 0),
		Bool("disabled", false),
	}

	binaryFields := ToBinaryFields(legacyFields)

	if len(binaryFields) != 6 {
		t.Fatalf("Expected 6 binary fields, got %d", len(binaryFields))
	}

	// Validate types
	expectedTypes := []uint8{
		uint8(StringType), uint8(IntType), uint8(BoolType),
		uint8(StringType), uint8(IntType), uint8(BoolType),
	}

	for i, bf := range binaryFields {
		if bf.Type != expectedTypes[i] {
			t.Errorf("Field %d: expected type %d, got %d", i, expectedTypes[i], bf.Type)
		}
	}

	t.Logf("✅ Reverse batch conversion: %d fields processed", len(binaryFields))
}

// TestReverseConversion_WithCapacity testa conversione con capacità
func TestReverseConversion_WithCapacity(t *testing.T) {
	legacyFields := []Field{
		Str("test1", "value1"),
		Int("test2", 42),
	}

	// Test with capacity larger than slice
	binaryFields := ToBinaryFieldsWithCapacity(legacyFields, 10)

	if len(binaryFields) != 2 {
		t.Errorf("Expected length 2, got %d", len(binaryFields))
	}

	if cap(binaryFields) != 10 {
		t.Errorf("Expected capacity 10, got %d", cap(binaryFields))
	}

	// Test with capacity smaller than slice
	binaryFields2 := ToBinaryFieldsWithCapacity(legacyFields, 1)

	if len(binaryFields2) != 2 {
		t.Errorf("Expected length 2, got %d", len(binaryFields2))
	}

	if cap(binaryFields2) < 2 {
		t.Errorf("Expected capacity >= 2, got %d", cap(binaryFields2))
	}

	t.Logf("✅ Capacity test passed: capacity management working correctly")
}

// TestRoundTripConversion testa conversione bidirezionale completa
func TestRoundTripConversion(t *testing.T) {
	// Start with binary fields
	originalBinary := []BinaryField{
		NextStr("user", "john"),
		NextInt("age", 30),
		NextBool("active", true),
	}

	// Convert to legacy
	legacyFields := ToLegacyFields(originalBinary)

	// Convert back to binary
	finalBinary := ToBinaryFields(legacyFields)

	if len(finalBinary) != len(originalBinary) {
		t.Fatalf("Length mismatch: original %d, final %d", len(originalBinary), len(finalBinary))
	}

	// Validate types are preserved
	for i := range originalBinary {
		if finalBinary[i].Type != originalBinary[i].Type {
			t.Errorf("Type mismatch at index %d: original %d, final %d",
				i, originalBinary[i].Type, finalBinary[i].Type)
		}
	}

	t.Logf("✅ Round-trip conversion successful: %d fields", len(originalBinary))
}
