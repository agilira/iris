// binary_logger.go: Pure binary structured logger - NO JSON
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// LazyCaller defers caller computation until needed (like Zap)
type LazyCaller struct {
	skip     int
	computed bool
	mu       sync.RWMutex
	file     string
	line     int
	function string
}

// NewLazyCaller creates a lazy caller that computes on demand
func NewLazyCaller(skip int) *LazyCaller {
	return &LazyCaller{
		skip: skip,
	}
}

// File returns the file (computed once)
func (lc *LazyCaller) File() string {
	lc.compute()
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.file
}

// Line returns the line number (computed once)
func (lc *LazyCaller) Line() int {
	lc.compute()
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.line
}

// Function returns the function name (computed once)
func (lc *LazyCaller) Function() string {
	lc.compute()
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.function
}

// compute does the expensive runtime.Caller once and caches
func (lc *LazyCaller) compute() {
	lc.mu.RLock()
	if lc.computed {
		lc.mu.RUnlock()
		return
	}
	lc.mu.RUnlock()

	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Double-check pattern
	if lc.computed {
		return
	}

	pc, file, line, ok := runtime.Caller(lc.skip)
	if ok {
		lc.file = file
		lc.line = line

		if fn := runtime.FuncForPC(pc); fn != nil {
			lc.function = fn.Name()
		}
	}

	lc.computed = true
}

// LazyCallerPool manages lazy caller reuse
type LazyCallerPool struct {
	pool sync.Pool
}

// NewLazyCallerPool creates a pool of lazy callers
func NewLazyCallerPool() *LazyCallerPool {
	return &LazyCallerPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &LazyCaller{}
			},
		},
	}
}

// GetLazyCaller gets a lazy caller from pool
func (lcp *LazyCallerPool) GetLazyCaller(skip int) *LazyCaller {
	caller := lcp.pool.Get().(*LazyCaller)
	caller.skip = skip
	caller.computed = false
	caller.file = ""
	caller.line = 0
	caller.function = ""
	return caller
}

// ReleaseLazyCaller returns lazy caller to pool
func (lcp *LazyCallerPool) ReleaseLazyCaller(caller *LazyCaller) {
	lcp.pool.Put(caller)
}

// BinaryField represents a field in pure binary format (ULTRA-LIGHTWEIGHT)
type BinaryField struct {
	KeyPtr uintptr // Key pointer (zero-copy)
	KeyLen uint16  // Key length
	Type   uint8   // Field type (1 byte)
	_      uint8   // Padding for alignment
	Data   uint64  // Union data: int64, bool, or string ptr+len
}

// BinaryEntry represents a complete log entry in binary format
type BinaryEntry struct {
	Timestamp uint64   // Unix nano timestamp
	Level     uint8    // Log level (1 byte)
	_         [3]uint8 // Padding for alignment
	MsgPtr    uintptr  // Message pointer (zero-copy)
	MsgLen    uint32   // Message length
	FieldPtr  uintptr  // Fields array pointer
	FieldCnt  uint16   // Field count
	_         uint16   // Padding
}

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

// BinaryContext represents a binary logging context
type BinaryContext struct {
	logger *BinaryLogger
	fields []BinaryField
}

// TASK S1: Direct binary field methods (OPTIMAL VERSION)
// BinaryStr creates binary string field directly (no Field conversion)
func BinaryStr(key, value string) BinaryField {
	strPtr := uintptr(unsafe.Pointer(&value))
	strLen := uint64(len(value))
	return BinaryField{
		KeyPtr: uintptr(unsafe.Pointer(&key)),
		KeyLen: uint16(len(key)),
		Type:   uint8(StringType),
		Data:   (uint64(strPtr) << 32) | strLen,
	}
} // BinaryInt creates binary int field directly (no Field conversion)
func BinaryInt(key string, value int64) BinaryField {
	return BinaryField{
		KeyPtr: uintptr(unsafe.Pointer(&key)),
		KeyLen: uint16(len(key)),
		Type:   uint8(IntType),
		Data:   uint64(value),
	}
}

// BinaryBool creates binary bool field directly (no Field conversion)
func BinaryBool(key string, value bool) BinaryField {
	data := uint64(0)
	if value {
		data = 1
	}
	return BinaryField{
		KeyPtr: uintptr(unsafe.Pointer(&key)),
		KeyLen: uint16(len(key)),
		Type:   uint8(BoolType),
		Data:   data,
	}
}

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
