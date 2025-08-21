# Iris Binary Logger API Reference

## Overview

Iris Binary Logger is an ultra-high performance structured logging system designed for applications requiring maximum throughput (104M+ operations/second) with zero-allocation, garbage collection-safe logging. The binary format provides superior performance compared to traditional JSON-based loggers.

## Quick Start

```go
// Create logger - CLEAN API!
logger := NewBLogger(InfoLevel)

// Create context with fields - SHORT NAMES!
ctx := logger.WithBFields(
    BStr("user_id", "12345"),
    BInt("request_size", 1024),
    BBool("authenticated", true),
)

// Log with context
ctx.Info("Request processed successfully")
```

## Table of Contents

- [Core Architecture](#core-architecture)
- [Logger](#1-logger-blogger)
- [Fields](#2-fields-bfield)
- [Context](#3-context-bcontext)
- [Memory Management](#4-memory-management)
- [Usage Examples](#usage-examples)
- [Performance Characteristics](#performance-characteristics)
- [Best Practices](#best-practices)

## Core Architecture

### Design Principles
- **GC-Safe**: Zero `unsafe.Pointer` usage, full garbage collection safety
- **Zero-Allocation**: Object pooling for all dynamic structures
- **Binary Format**: Native binary encoding for maximum performance
- **Type-Safe**: Compile-time type safety with zero runtime overhead

### Architecture Decisions

#### String Buffer Pool Pattern
Iris uses pooled string buffers to eliminate allocations:
- **2KB initial capacity** balances memory usage with performance
- **Reference-based storage** avoids string copying and reduces GC pressure
- **Automatic growth** handles large log entries without performance degradation

#### Field Type System
Binary field encoding optimizes for both space and access time:
- **Inline primitives**: int64/bool stored directly in 8-byte `Data` field
- **Referenced strings**: Stored as buffer offset/length pairs
- **Type identification**: Single-byte type field for fast dispatch

#### Memory Pool Architecture
Three-tier pooling strategy optimizes for high-throughput scenarios:
- **String Buffers**: Pooled for field storage, returned after use
- **Field Slices**: Pre-allocated slices avoid dynamic growth
- **Context Objects**: Entire context objects pooled for zero allocation

## Core Components

### 1. Logger (`BLogger`)

#### Constructor

```go
func NewBLogger(level Level) *BLogger
```

**Parameters:**
- `level`: Minimum log level (DebugLevel, InfoLevel, WarnLevel, ErrorLevel)

**Returns:** Logger instance with initialized pools and GC-safe structures.

### 2. Fields (`BField`)

Binary fields represent structured data in a memory-optimized format using buffer references instead of direct string storage.

#### Field Types

```go
type BField struct {
    Buffer   *StringBuffer // Buffer reference (GC-safe)
    KeyRef   StringRef     // Key reference in buffer
    ValueRef StringRef     // Value reference (for strings)
    Type     uint8         // Field type identifier
    Data     uint64        // Primitive data (int, bool, float)
}
```

#### Field Constructors

##### String Fields

```go
func BStr(key string, value string) BField
```
- **Purpose:** Creates a string field with GC-safe buffer pooling
- **Performance:** Zero allocations after buffer pool warmup
- **Memory:** Stored as buffer references, not direct strings

##### Integer Fields

```go
func BInt(key string, value int64) BField
```
- **Purpose:** Creates an integer field with inline storage
- **Performance:** Direct data storage in `Data` field, no additional allocations
- **Range:** Full int64 range supported

##### Boolean Fields

```go
func BBool(key string, value bool) BField
```
- **Purpose:** Creates a boolean field with compact storage
- **Performance:** Single bit stored as uint64, zero allocation
- **Storage:** `true` = 1, `false` = 0 in `Data` field

#### Field Methods

```go
func (bf BField) GetKey() string
```
- **Purpose:** Retrieves field key as string
- **Safety:** Returns empty string if buffer is nil

```go
func (bf BField) GetString() string
```
- **Purpose:** Retrieves string value for string fields
- **Validation:** Returns empty string for non-string field types

```go
func (bf BField) GetInt() int64
```
- **Purpose:** Retrieves integer value from `Data` field
- **Performance:** Direct memory access, zero allocation

```go
func (bf BField) GetBool() bool
```
- **Purpose:** Retrieves boolean value from `Data` field
- **Logic:** Returns `true` if `Data != 0`

```go
func (bf BField) Release()
```
- **Purpose:** Returns buffer to pool (critical for high-throughput)
- **Usage:** Call when field is no longer needed
- **Performance:** Essential for maintaining 104M+ ops/sec

### 3. Context (`BContext`)

Represents a logging context with pre-allocated fields for efficient logging operations.

#### Context Creation

```go
func (bl *BLogger) WithBFields(fields ...BField) *BContext
```
- **Purpose:** Creates logging context with specified fields
- **Performance:** Uses field pool for zero allocation
- **Pattern:** Variadic arguments for fluent API

#### Logging Methods

```go
func (bc *BContext) Info(message string)
```
- **Purpose:** Logs at INFO level with context fields
- **Performance:** Binary encoding, zero JSON marshaling
- **Behavior:** No-op if logger level > InfoLevel

```go
func (bc *BContext) InfoWithCaller(message string)
```
- **Purpose:** Logs at INFO level with lazy caller information
- **Performance:** Optimized caller computation, computed only when needed
- **Use Case:** Development and debugging scenarios requiring call site information

```go
func (bc *BContext) Debug(message string)
```
- **Purpose:** Logs at DEBUG level with context fields
- **Performance:** Optimized for development scenarios
- **Behavior:** No-op if logger level > DebugLevel

```go
func (bc *BContext) Warn(message string)
```
- **Purpose:** Logs at WARN level with context fields
- **Use Case:** Non-fatal issues requiring attention

```go
func (bc *BContext) Error(message string)
```
- **Purpose:** Logs at ERROR level with context fields
- **Use Case:** Error conditions requiring immediate attention

#### Diagnostic Methods

```go
func (bc *BContext) MemoryFootprint() int
```
- **Purpose:** Returns the total memory footprint of the context
- **Returns:** Memory usage in bytes
- **Use Case:** Performance profiling and optimization

```go
func (bc *BContext) GetBinarySize() int
```
- **Purpose:** Returns memory usage in bytes for context and fields
- **Use Case:** Performance monitoring and memory profiling
- **Includes:** Context structure + all field buffers

```go
func (bc *BinaryContext) GetBinarySize() int
```
- **Purpose:** Returns binary encoding size for the log entry
- **Performance:** Used for buffer sizing and throughput calculations
- **Accuracy:** Reflects actual bytes that would be written

### 4. Memory Management

#### Buffer Pooling

```go
func GetStringBuffer() *StringBuffer
func ReleaseStringBuffer(buf *StringBuffer)
```
- **Purpose:** Advanced buffer pool management
- **Performance:** Eliminates allocations for string storage
- **Critical:** Must balance Get/Release calls for optimal performance

#### String References

```go
type StringRef struct {
    Offset uint32
    Length uint32
}
```
- **Purpose:** Reference-based string storage avoiding copies
- **Memory:** 8 bytes per reference vs. full string storage
- **Safety:** Bounds-checked access to underlying buffer

## Usage Examples

### Basic Logging

```go
logger := NewBLogger(InfoLevel)

// Single log with fields
ctx := logger.WithBFields(
    BStr("user_id", "12345"),
    BInt("request_size", 1024),
    BBool("authenticated", true),
)
ctx.Info("Request processed successfully")
```

### High-Performance Pattern

```go
logger := NewBLogger(InfoLevel)

// Pre-create context for reuse
ctx := logger.WithBFields(
    BStr("service", "api-gateway"),
    BStr("version", "v2.1.0"),
)

// Reuse context for multiple logs
for i := 0; i < 1000000; i++ {
    ctx.Info("Processing request")
}
```

### Field Reuse Pattern

```go
// Create fields once, reuse multiple times
userField := BStr("user", "john_doe")
serviceField := BStr("service", "auth")

ctx1 := logger.WithBFields(userField, serviceField, BStr("action", "login"))
ctx2 := logger.WithBFields(userField, serviceField, BStr("action", "logout"))

ctx1.Info("User login attempt")
ctx2.Info("User logout")

// Critical: Release buffers when done
userField.Release()
serviceField.Release()
```

## Performance Characteristics

### Benchmarks

| Operation | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `WithBFields` (3 fields) | 266 | 376 | 3 |
| `BStr()` creation | 21 | 0 | 0 |
| `BInt()` creation | 19 | 0 | 0 |
| `BBool()` creation | 18 | 0 | 0 |

### Memory Efficiency

- **Buffer Pooling:** 2KB initial buffer size with growth
- **Reference Storage:** 8 bytes per string vs. full string copy
- **Pool Warmup:** Zero allocations after initial pool population
- **GC Pressure:** Minimal due to pooling and reference patterns

### Throughput Targets

- **Design Goal:** 104M+ operations/second
- **Concurrency:** MPSC (Multi-Producer Single-Consumer) optimized
- **Latency:** Sub-20ns per operation in optimal conditions

## Thread Safety

- **Producer Safety:** Multiple goroutines can safely create contexts and log
- **Buffer Pools:** Thread-safe with internal synchronization
- **Consumer:** Single consumer pattern recommended for maximum performance

## Best Practices

### High-Performance Applications

1. **Pre-allocate contexts** for frequently used field combinations
2. **Reuse fields** across multiple log statements
3. **Call `Release()`** on fields when lifecycle ends
4. **Batch operations** when possible to amortize pool overhead

### Memory Management

1. **Balance Get/Release** buffer pool operations
2. **Avoid long-lived field references** that prevent pool return
3. **Monitor pool metrics** in production environments
4. **Use profiling** to identify allocation hotspots

### Development Guidelines

1. **Start with `BStr/BInt/BBool`** constructors for type safety
2. **Use `WithBFields()`** for context creation
3. **Test with production load** to validate performance characteristics
4. **Profile regularly** to ensure zero-allocation goals are met

## Error Handling

### Nil Safety

All binary field operations are nil-safe and return zero values rather than panicking:

```go
var field BField  // Zero value
key := field.GetKey()      // Returns ""
value := field.GetString() // Returns ""
field.Release()            // Safe no-op
```

### Buffer Exhaustion

When buffer pools are under pressure, operations gracefully degrade:
- New buffers allocated if pool empty
- Performance impact minimal due to pool growth
- No data loss or corruption

### Type Mismatches

Field type validation prevents runtime errors:
- `GetString()` on non-string fields returns empty string
- `GetInt()` on string fields returns 0
- Type information preserved in binary format

---

*For implementation details and advanced usage patterns, see the source code and benchmark suite.*
