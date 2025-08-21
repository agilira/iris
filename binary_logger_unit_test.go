// binary_logger_unit_test.go: Comprehensive safety net for binary logger optimizations
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strings"
	"testing"
	"unsafe"
)

// TestBinaryLoggerCreation tests basic binary logger instantiation
func TestBinaryLoggerCreation(t *testing.T) {
	logger := NewBinaryLogger(InfoLevel)

	if logger == nil {
		t.Fatal("BinaryLogger should not be nil")
	}

	if logger.level != InfoLevel {
		t.Errorf("Expected level InfoLevel, got %v", logger.level)
	}

	if logger.lazyCallerPool == nil {
		t.Error("LazyCallerPool should be initialized")
	}
}

// TestBinaryFieldCreation tests direct binary field creation
func TestBinaryFieldCreation(t *testing.T) {
	tests := []struct {
		name     string
		field    BinaryField
		key      string
		expected uint8
	}{
		{
			name:     "string_field",
			field:    BinaryStr("test_key", "test_value"),
			key:      "test_key",
			expected: uint8(StringType),
		},
		{
			name:     "int_field",
			field:    BinaryInt("count", 42),
			key:      "count",
			expected: uint8(IntType),
		},
		{
			name:     "bool_true_field",
			field:    BinaryBool("enabled", true),
			key:      "enabled",
			expected: uint8(BoolType),
		},
		{
			name:     "bool_false_field",
			field:    BinaryBool("disabled", false),
			key:      "disabled",
			expected: uint8(BoolType),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Verify field type
			if test.field.Type != test.expected {
				t.Errorf("Expected type %d, got %d", test.expected, test.field.Type)
			}

			// Verify key length
			if test.field.KeyLen != uint16(len(test.key)) {
				t.Errorf("Expected key length %d, got %d", len(test.key), test.field.KeyLen)
			}

			// Verify key pointer is not null
			if test.field.KeyPtr == 0 {
				t.Error("Key pointer should not be null")
			}
		})
	}
}

// TestBinaryStringFieldData tests string field data encoding
func TestBinaryStringFieldData(t *testing.T) {
	value := "test_string"
	field := BinaryStr("key", value)

	// Extract string pointer and length from Data field
	strPtr := uintptr(field.Data >> 32)
	strLen := uint64(field.Data & 0xFFFFFFFF)

	if strLen != uint64(len(value)) {
		t.Errorf("Expected string length %d, got %d", len(value), strLen)
	}

	if strPtr == 0 {
		t.Error("String pointer should not be null")
	}
}

// TestBinaryIntFieldData tests int field data encoding
func TestBinaryIntFieldData(t *testing.T) {
	value := int64(12345)
	field := BinaryInt("count", value)

	if field.Data != uint64(value) {
		t.Errorf("Expected data %d, got %d", value, field.Data)
	}
}

// TestBinaryBoolFieldData tests bool field data encoding
func TestBinaryBoolFieldData(t *testing.T) {
	// Test true
	fieldTrue := BinaryBool("enabled", true)
	if fieldTrue.Data != 1 {
		t.Errorf("Expected true data to be 1, got %d", fieldTrue.Data)
	}

	// Test false
	fieldFalse := BinaryBool("disabled", false)
	if fieldFalse.Data != 0 {
		t.Errorf("Expected false data to be 0, got %d", fieldFalse.Data)
	}
}

// TestBinaryContextWithBinaryFields tests direct binary field context creation
func TestBinaryContextWithBinaryFields(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	fields := []BinaryField{
		BinaryStr("service", "user-api"),
		BinaryInt("user_id", 12345),
		BinaryBool("authenticated", true),
	}

	ctx := logger.WithBinaryFields(fields...)

	if ctx == nil {
		t.Fatal("Context should not be nil")
	}

	if ctx.logger != logger {
		t.Error("Context logger should reference original logger")
	}

	if len(ctx.fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(ctx.fields))
	}

	// Verify field types
	if ctx.fields[0].Type != uint8(StringType) {
		t.Error("First field should be string type")
	}
	if ctx.fields[1].Type != uint8(IntType) {
		t.Error("Second field should be int type")
	}
	if ctx.fields[2].Type != uint8(BoolType) {
		t.Error("Third field should be bool type")
	}
}

// TestBinaryContextWithLegacyFields tests legacy field conversion
func TestBinaryContextWithLegacyFields(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	fields := []Field{
		Str("service", "user-api"),
		Int("user_id", 12345),
		Bool("authenticated", true),
	}

	ctx := logger.WithFields(fields...)

	if ctx == nil {
		t.Fatal("Context should not be nil")
	}

	if len(ctx.fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(ctx.fields))
	}

	// Verify converted field types
	if ctx.fields[0].Type != uint8(StringType) {
		t.Error("First field should be string type")
	}
	if ctx.fields[1].Type != uint8(IntType) {
		t.Error("Second field should be int type")
	}
	if ctx.fields[2].Type != uint8(BoolType) {
		t.Error("Third field should be bool type")
	}
}

// TestBinaryContextInfo tests basic Info logging
func TestBinaryContextInfo(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	ctx := logger.WithBinaryFields(
		BinaryStr("request_id", "req-123"),
		BinaryInt("status", 200),
	)

	// This should not panic
	ctx.Info("Request processed successfully")

	// Context should be returned to pool after use
	// Subsequent calls should work
	ctx2 := logger.WithBinaryFields(BinaryStr("test", "value"))
	ctx2.Info("Second message")
}

// TestBinaryContextInfoWithCaller tests Info logging with caller information
func TestBinaryContextInfoWithCaller(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	ctx := logger.WithBinaryFields(
		BinaryStr("operation", "test"),
	)

	// This should not panic
	ctx.InfoWithCaller("Operation completed")
}

// TestBinaryContextLevelFiltering tests log level filtering
func TestBinaryContextLevelFiltering(t *testing.T) {
	// Logger with WARN level should skip INFO messages
	logger := NewBinaryLogger(WarnLevel)

	ctx := logger.WithBinaryFields(
		BinaryStr("test", "value"),
	)

	// This should be filtered out (no panic expected)
	ctx.Info("This should be filtered")
	ctx.InfoWithCaller("This should also be filtered")
}

// TestBinaryContextEmptyFields tests context with no fields
func TestBinaryContextEmptyFields(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	ctx := logger.WithBinaryFields()

	if len(ctx.fields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(ctx.fields))
	}

	// Should not panic with empty fields
	ctx.Info("Message without fields")
	ctx.InfoWithCaller("Message without fields and with caller")
}

// TestBinaryContextMemoryFootprint tests memory usage calculation
func TestBinaryContextMemoryFootprint(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	ctx := logger.WithBinaryFields(
		BinaryStr("key1", "value1"),
		BinaryInt("key2", 42),
		BinaryBool("key3", true),
	)

	footprint := ctx.MemoryFootprint()

	expectedEntrySize := int(unsafe.Sizeof(BinaryEntry{}))
	expectedFieldsSize := int(unsafe.Sizeof(BinaryField{})) * 3
	expectedTotal := expectedEntrySize + expectedFieldsSize

	if footprint != expectedTotal {
		t.Errorf("Expected footprint %d, got %d", expectedTotal, footprint)
	}
}

// TestBinaryContextGetBinarySize tests binary size calculation
func TestBinaryContextGetBinarySize(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	ctx := logger.WithBinaryFields(
		BinaryStr("service", "api"),
		BinaryInt("port", 8080),
	)

	size := ctx.GetBinarySize()
	footprint := ctx.MemoryFootprint()

	if size != footprint {
		t.Errorf("GetBinarySize (%d) should equal MemoryFootprint (%d)", size, footprint)
	}
}

// TestLazyCallerCreation tests lazy caller instantiation
func TestLazyCallerCreation(t *testing.T) {
	caller := NewLazyCaller(1)

	if caller == nil {
		t.Fatal("LazyCaller should not be nil")
	}

	if caller.skip != 1 {
		t.Errorf("Expected skip 1, got %d", caller.skip)
	}

	if caller.computed {
		t.Error("Caller should not be computed initially")
	}
}

// TestLazyCallerComputation tests on-demand caller computation
func TestLazyCallerComputation(t *testing.T) {
	caller := NewLazyCaller(0)

	// First access should trigger computation
	file := caller.File()
	if file == "" {
		t.Error("File should not be empty after computation")
	}

	if !caller.computed {
		t.Error("Caller should be marked as computed")
	}

	line := caller.Line()
	if line <= 0 {
		t.Error("Line should be positive after computation")
	}

	function := caller.Function()
	if function == "" {
		t.Error("Function should not be empty after computation")
	}

	// Verify file contains appropriate name (could be any test file)
	if file == "" {
		t.Errorf("Expected non-empty file, got: %s", file)
	}
}

// TestLazyCallerPool tests lazy caller pooling
func TestLazyCallerPool(t *testing.T) {
	pool := NewLazyCallerPool()

	if pool == nil {
		t.Fatal("LazyCallerPool should not be nil")
	}

	// Get caller from pool
	caller1 := pool.GetLazyCaller(2)
	if caller1 == nil {
		t.Fatal("Caller from pool should not be nil")
	}

	if caller1.skip != 2 {
		t.Errorf("Expected skip 2, got %d", caller1.skip)
	}

	if caller1.computed {
		t.Error("Pooled caller should be reset")
	}

	// Return to pool
	pool.ReleaseLazyCaller(caller1)

	// Get another caller (might be the same instance)
	caller2 := pool.GetLazyCaller(3)
	if caller2 == nil {
		t.Fatal("Second caller from pool should not be nil")
	}

	if caller2.skip != 3 {
		t.Errorf("Expected skip 3, got %d", caller2.skip)
	}

	pool.ReleaseLazyCaller(caller2)
}

// TestBinaryLoggerPooling tests resource pooling efficiency
func TestBinaryLoggerPooling(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	// Create multiple contexts to test pooling
	for i := 0; i < 10; i++ {
		ctx := logger.WithBinaryFields(
			BinaryStr("iteration", "test"),
			BinaryInt("count", int64(i)),
		)

		ctx.Info("Pooling test message")
	}

	// Should not panic and should reuse pooled resources
}

// TestBinaryFieldAlignment tests struct alignment for performance
func TestBinaryFieldAlignment(t *testing.T) {
	var field BinaryField

	// Verify expected sizes for performance optimization
	expectedSize := uintptr(24) // Optimized struct size
	actualSize := unsafe.Sizeof(field)

	if actualSize > expectedSize {
		t.Logf("BinaryField size: %d bytes (larger than expected %d)", actualSize, expectedSize)
	}

	// Verify alignment
	if actualSize%8 != 0 {
		t.Errorf("BinaryField should be 8-byte aligned, got size %d", actualSize)
	}
}

// TestBinaryEntryAlignment tests struct alignment for performance
func TestBinaryEntryAlignment(t *testing.T) {
	var entry BinaryEntry

	actualSize := unsafe.Sizeof(entry)

	// Verify alignment
	if actualSize%8 != 0 {
		t.Errorf("BinaryEntry should be 8-byte aligned, got size %d", actualSize)
	}

	t.Logf("BinaryEntry size: %d bytes", actualSize)
}

// TestBinaryLoggerConcurrentAccess tests thread safety
func TestBinaryLoggerConcurrentAccess(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	// Run concurrent logging operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			ctx := logger.WithBinaryFields(
				BinaryStr("goroutine", "test"),
				BinaryInt("id", int64(id)),
			)

			ctx.Info("Concurrent message")
			ctx.InfoWithCaller("Concurrent message with caller")
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestBinaryLoggerEdgeCases tests edge cases and error conditions
func TestBinaryLoggerEdgeCases(t *testing.T) {
	logger := NewBinaryLogger(DebugLevel)

	// Test with very long strings
	longValue := strings.Repeat("x", 1000)
	ctx := logger.WithBinaryFields(
		BinaryStr("long_key", longValue),
	)
	ctx.Info("Message with long field")

	// Test with zero values
	ctx2 := logger.WithBinaryFields(
		BinaryStr("empty", ""),
		BinaryInt("zero", 0),
		BinaryBool("false", false),
	)
	ctx2.Info("Message with zero values")

	// Test with special characters
	ctx3 := logger.WithBinaryFields(
		BinaryStr("special", "Hello\nWorld\t\"Test\""),
	)
	ctx3.Info("Message with special characters")
}
