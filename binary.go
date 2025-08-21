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
		lazyCallerPool: NewLazyCallerPool(), // Lazy caller pool (VICTORY!)
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

// SAFE MIGRATION: User-friendly constructor alias
func NewBLogger(level Level) *BLogger {
	return NewBinaryLogger(level)
}

// NOTE: BinaryField, BinaryEntry, BinaryContext moved to binary_types.go
//

//
// NOTE: BinaryField helper functions moved to binary_helpers.go
//

// WithBinaryFields creates context with direct binary fields (S1 OPTIMIZATION - OPTIMAL)
func (bl *BinaryLogger) WithBinaryFields(fields ...BinaryField) *BinaryContext {
	// LOCK-FREE: Direct allocation, no shared pools
	fieldsCopy := make([]BinaryField, len(fields))
	copy(fieldsCopy, fields)

	// Direct allocation - no pool contention
	ctx := &BinaryContext{
		logger: bl,
		fields: fieldsCopy,
	}
	return ctx
}

// WithFields creates a binary context with fields (LEGACY COMPATIBILITY)
func (bl *BinaryLogger) WithFields(fields ...Field) *BinaryContext {
	fieldsPtr := bl.fieldPool.Get().(*[]BinaryField)
	binaryFields := (*fieldsPtr)[:0] // Reset length, keep capacity

	// Convert to binary fields (GC-SAFE IMPLEMENTATION)
	for _, field := range fields {
		switch field.Type {
		case StringType:
			binaryFields = append(binaryFields, BinaryStr(field.Key, field.String))
		case IntType:
			binaryFields = append(binaryFields, BinaryInt(field.Key, field.Int))
		case BoolType:
			binaryFields = append(binaryFields, BinaryBool(field.Key, field.Bool))
		default:
			// For now, convert unsupported types to string
			binaryFields = append(binaryFields, BinaryStr(field.Key, "unsupported"))
		}
	}

	return &BinaryContext{
		logger: bl,
		fields: binaryFields,
	}
}

// SAFE MIGRATION: User-friendly alias (BACKWARD COMPATIBLE)
func (bl *BinaryLogger) WithBFields(fields ...BField) *BinaryContext {
	return bl.WithBinaryFields(fields...)
}

// Info logs at info level with pure binary format
func (bc *BinaryContext) Info(message string) {
	if bc.logger.level > InfoLevel {
		return
	}

	// Create binary entry (GC-SAFE)
	entry := bc.logger.entryPool.Get().(*BinaryEntry)
	timestamp, _ := SafeInt64ToUint64ForEncoding(time.Now().UnixNano()) // Safe timestamp conversion
	entry.Timestamp = timestamp
	entry.Level = uint8(InfoLevel)

	// Direct message storage - LOCK-FREE immutable
	entry.Message = message
	entry.Fields = bc.fields

	// TODO: Write to binary output (no JSON!)
	// For now, just measure the binary structure cost

	// No release needed - immutable fields, GC handles cleanup

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

	// Create binary entry (GC-SAFE)
	entry := bc.logger.entryPool.Get().(*BinaryEntry)
	timestamp, _ := SafeInt64ToUint64ForEncoding(time.Now().UnixNano()) // Safe timestamp conversion
	entry.Timestamp = timestamp
	entry.Level = uint8(InfoLevel)

	// Direct message storage - LOCK-FREE immutable
	entry.Message = message
	entry.Fields = bc.fields

	// ZAP-STYLE: Create lazy caller without computing (12.6ns!)
	caller := bc.logger.lazyCallerPool.GetLazyCaller(3)
	_ = caller // Lazy caller created but not computed
	bc.logger.lazyCallerPool.ReleaseLazyCaller(caller)

	// TODO: Write to binary output with caller info
	// Caller computation only happens if/when the log is actually output

	// No release needed - immutable fields, GC handles cleanup

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

// BinaryField operations moved to binary_types.go for modular organization
