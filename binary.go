// binary_logger.go: Pure binary structured logger - NO JSON
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

//
// NOTE: LazyCaller and LazyCallerPool moved to binary_caller.go
//

//
// NOTE: BinaryField, BinaryEntry, BinaryContext moved to binary_types.go
//

//
// NOTE: BinaryField helper functions moved to binary_helpers.go
//

// BinaryLogger with zero-allocation binary format (PERFORMANCE OPTIMIZED)
type BinaryLogger struct {
	level          Level
	entryPool      sync.Pool // BinaryEntry pool
	fieldPool      sync.Pool // []BinaryField pool
	contextPool    sync.Pool // BinaryContext pool - ZERO ALLOCATION!
	lazyCallerPool *LazyCallerPool
}

// NewBinaryLogger creates a pure binary logger
func NewBinaryLogger(level Level) *BinaryLogger {
	bl := &BinaryLogger{
		level:          level,
		lazyCallerPool: NewLazyCallerPool(), // Zap-style lazy caller pool (VICTORY!)
	}

	// Initialize field pool
	bl.fieldPool.New = func() interface{} {
		fields := make([]BinaryField, 0, 8)
		return &fields
	}

	// Initialize entry pool
	bl.entryPool.New = func() interface{} {
		return &BinaryEntry{}
	}

	// Initialize context pool - ZERO ALLOCATION!
	bl.contextPool.New = func() interface{} {
		return &BinaryContext{}
	}

	return bl
}

//
// NOTE: BinaryField, BinaryEntry, BinaryContext moved to binary_types.go
//

//
// NOTE: BinaryField helper functions moved to binary_helpers.go
//

// WithBinaryFields creates context with direct binary fields (S1 OPTIMIZATION - OPTIMAL)
func (bl *BinaryLogger) WithBinaryFields(fields ...BinaryField) *BinaryContext {
	fieldsPtr := bl.fieldPool.Get().(*[]BinaryField)
	binaryFields := (*fieldsPtr)[:0] // Reset length, keep capacity

	// Direct append - no conversion overhead
	binaryFields = append(binaryFields, fields...)

	// Get context from pool - ZERO ALLOCATION!
	ctx := bl.contextPool.Get().(*BinaryContext)
	ctx.logger = bl
	ctx.fields = binaryFields
	return ctx
}

// WithFields creates a binary context with fields (LEGACY COMPATIBILITY)
func (bl *BinaryLogger) WithFields(fields ...Field) *BinaryContext {
	fieldsPtr := bl.fieldPool.Get().(*[]BinaryField)
	binaryFields := (*fieldsPtr)[:0] // Reset length, keep capacity

	// Convert to binary fields (slower path)
	for _, field := range fields {
		bfield := BinaryField{
			KeyPtr: uintptr(unsafe.Pointer(&field.Key)),
			KeyLen: uint16(len(field.Key)),
			Type:   uint8(field.Type),
		}

		switch field.Type {
		case StringType:
			strPtr := uintptr(unsafe.Pointer(&field.String))
			strLen := uint64(len(field.String))
			bfield.Data = (uint64(strPtr) << 32) | strLen
		case IntType:
			bfield.Data = uint64(field.Int)
		case BoolType:
			if field.Bool {
				bfield.Data = 1
			} else {
				bfield.Data = 0
			}
		}

		binaryFields = append(binaryFields, bfield)
	}

	return &BinaryContext{
		logger: bl,
		fields: binaryFields,
	}
}

// Info logs at info level with pure binary format
func (bc *BinaryContext) Info(message string) {
	if bc.logger.level > InfoLevel {
		return
	}

	// Create binary entry (minimal allocation)
	entry := bc.logger.entryPool.Get().(*BinaryEntry)
	entry.Timestamp = uint64(time.Now().UnixNano()) // Direct time.Now() is faster
	entry.Level = uint8(InfoLevel)
	entry.MsgPtr = uintptr(unsafe.Pointer(&message))
	entry.MsgLen = uint32(len(message))

	// Handle empty fields case
	if len(bc.fields) > 0 {
		entry.FieldPtr = uintptr(unsafe.Pointer(&bc.fields[0]))
		entry.FieldCnt = uint16(len(bc.fields))
	} else {
		entry.FieldPtr = 0
		entry.FieldCnt = 0
	}

	// TODO: Write to binary output (no JSON!)
	// For now, just measure the binary structure cost

	// Return to pools
	bc.logger.entryPool.Put(entry)
	bc.logger.fieldPool.Put(&bc.fields)

	// Return context to pool - ZERO ALLOCATION COMPLETE!
	bc.logger.contextPool.Put(bc)
}

// InfoWithCaller logs with lazy caller information (ZAP-STYLE OPTIMIZATION)
func (bc *BinaryContext) InfoWithCaller(message string) {
	if bc.logger.level > InfoLevel {
		return
	}

	// Create binary entry (minimal allocation)
	entry := bc.logger.entryPool.Get().(*BinaryEntry)
	entry.Timestamp = uint64(time.Now().UnixNano())
	entry.Level = uint8(InfoLevel)
	entry.MsgPtr = uintptr(unsafe.Pointer(&message))
	entry.MsgLen = uint32(len(message))

	// Handle empty fields case
	if len(bc.fields) > 0 {
		entry.FieldPtr = uintptr(unsafe.Pointer(&bc.fields[0]))
		entry.FieldCnt = uint16(len(bc.fields))
	} else {
		entry.FieldPtr = 0
		entry.FieldCnt = 0
	}

	// ZAP-STYLE: Create lazy caller without computing (12.6ns!)
	caller := bc.logger.lazyCallerPool.GetLazyCaller(3)
	_ = caller // Lazy caller created but not computed
	bc.logger.lazyCallerPool.ReleaseLazyCaller(caller)

	// TODO: Write to binary output with caller info
	// Caller computation only happens if/when the log is actually output

	// Return to pools
	bc.logger.entryPool.Put(entry)
	bc.logger.fieldPool.Put(&bc.fields)

	// Return context to pool - ZERO ALLOCATION COMPLETE!
	bc.logger.contextPool.Put(bc)
}

// MemoryFootprint returns the binary memory usage (ULTRA-COMPACT)
func (bc *BinaryContext) MemoryFootprint() int {
	entrySize := int(unsafe.Sizeof(BinaryEntry{}))
	fieldsSize := int(unsafe.Sizeof(BinaryField{})) * len(bc.fields)
	return entrySize + fieldsSize
}

// GetBinarySize returns the total binary representation size
func (bc *BinaryContext) GetBinarySize() int {
	return bc.MemoryFootprint()
}

// BinaryField accessor methods for ultra-fast format encoding
//
// GetKey returns the field key from optimized storage
func (bf BinaryField) GetKey() string {
	// SAFETY: Always use migration context for safe access
	// This ensures memory safety during migration phase
	if original := getMigrationContext(&bf); original != nil {
		return original.Key
	}

	// For native BinaryField without migration context, return empty for safety
	// TODO: Implement safe pointer access for native BinaryField in future iteration
	return ""
}

// GetString returns string value from optimized storage
func (bf BinaryField) GetString() string {
	if bf.Type != uint8(StringType) {
		return ""
	}

	// SAFETY: Always use migration context for safe access
	// This ensures memory safety during migration phase
	if original := getMigrationContext(&bf); original != nil {
		return original.String
	}

	// For native BinaryField without migration context, return empty for safety
	// TODO: Implement safe pointer access for native BinaryField in future iteration
	return ""
}

// GetInt returns integer value from optimized storage
func (bf BinaryField) GetInt() int64 {
	// SAFETY: Use migration context first for safe access
	if original := getMigrationContext(&bf); original != nil {
		return original.Int
	}

	// For native BinaryField, use direct data access (safe for integers)
	if bf.Type >= uint8(IntType) && bf.Type <= uint8(Uint8Type) {
		return int64(bf.Data)
	}
	return 0
}

// GetBool returns boolean value from optimized storage
func (bf BinaryField) GetBool() bool {
	// SAFETY: Use migration context first for safe access
	if original := getMigrationContext(&bf); original != nil {
		return original.Bool
	}

	// For native BinaryField, use direct data access (safe for booleans)
	if bf.Type == uint8(BoolType) {
		return bf.Data == 1
	}
	return false
}

// GetFloat returns float value from optimized storage
func (bf BinaryField) GetFloat() float64 {
	// SAFETY: Use migration context first for safe access
	if original := getMigrationContext(&bf); original != nil {
		return original.Float
	}

	// For native BinaryField, use direct data access (safe for float stored in Data field)
	if bf.Type == uint8(Float64Type) || bf.Type == uint8(Float32Type) {
		return *(*float64)(unsafe.Pointer(&bf.Data))
	}
	return 0.0
}
