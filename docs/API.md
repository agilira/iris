# Iris eXtreme Logger API Reference

## Overview

Iris eXtreme Logger is an ultra-high performance structured logging system designed for applications requiring maximum throughput (104M+ operations/second) with zero-allocation, garbage collection-safe logging. The binary format provides superior performance compared to traditional JSON-based loggers.

## Quick Start

```go
// Create logger - DESIGNER API!
logger := NewIrisLogger(InfoLevel)

// Create context with fields - eXtreme PERFORMANCE!
ctx := logger.XFields(
    XStr("user_id", "12345"),
    XInt("request_size", 1024),
    XBool("authenticated", true),
)

// Log with context
ctx.Info("Request processed successfully")
```

## Table of Contents

- [Core Architecture](#core-architecture)
- [Logger](#1-logger-irislogger)
- [Fields](#2-fields-xfield)
- [Context](#3-context-xcontext)
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

### 1. Logger (`IrisLogger`)

#### Constructor

```go
func NewIrisLogger(level Level) *IrisLogger
```

**Parameters:**
- `level`: Minimum log level (DebugLevel, InfoLevel, WarnLevel, ErrorLevel)

**Returns:** Iris logger instance with initialized pools and GC-safe structures.

### 2. Fields (`XField`)

Binary fields represent structured data in a memory-optimized format using buffer references instead of direct string storage.

#### Field Types

```go
type XField struct {
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
func XStr(key string, value string) XField
```
- **Purpose:** Creates a string field with GC-safe buffer pooling
- **Performance:** Zero allocations after buffer pool warmup  
- **Memory:** Stored as buffer references, not direct strings
- **Design:** "X" = eXtreme performance

##### Integer Fields

```go
func XInt(key string, value int64) XField
```
- **Purpose:** Creates an integer field with inline storage
- **Performance:** Direct data storage in `Data` field, no additional allocations
- **Range:** Full int64 range supported
- **Design:** "X" = eXtreme performance

##### Boolean Fields

```go
func XBool(key string, value bool) XField
```
- **Purpose:** Creates a boolean field with compact storage
- **Performance:** Single bit stored as uint64, zero allocation
- **Storage:** `true` = 1, `false` = 0 in `Data` field
- **Design:** "X" = eXtreme performance

#### Field Methods

```go
func (xf XField) GetKey() string
```
- **Purpose:** Retrieves field key as string
- **Safety:** Returns empty string if buffer is nil

```go
func (xf XField) GetString() string
```
- **Purpose:** Retrieves string value for string fields
- **Validation:** Returns empty string for non-string field types

```go
func (xf XField) GetInt() int64
```
- **Purpose:** Retrieves integer value from `Data` field
- **Performance:** Direct memory access, zero allocation

```go
func (xf XField) GetBool() bool
```
- **Purpose:** Retrieves boolean value from `Data` field
- **Logic:** Returns `true` if `Data != 0`

```go
func (xf XField) Release()
```
- **Purpose:** Returns buffer to pool (critical for high-throughput)
- **Usage:** Call when field is no longer needed
- **Performance:** Essential for maintaining 104M+ ops/sec

### 3. Context (`XContext`)

Represents a logging context with pre-allocated fields for efficient logging operations.

#### Context Creation

```go
func (il *XLogger) XFields(fields ...XField) *XContext
```
- **Purpose:** Creates logging context with specified fields using eXtreme performance API
- **Performance:** Uses field pool for zero allocation
- **Pattern:** Variadic arguments for fluent API
- **Design:** Short, memorable method name for high-frequency use

#### Logging Methods

```go
func (xc *XContext) Info(message string)
```
- **Purpose:** Logs at INFO level with context fields
- **Performance:** Binary encoding, zero JSON marshaling
- **Behavior:** No-op if logger level > InfoLevel

```go
func (xc *XContext) InfoWithCaller(message string)
```
- **Purpose:** Logs at INFO level with lazy caller information
- **Performance:** Optimized caller computation, computed only when needed
- **Use Case:** Development and debugging scenarios requiring call site information

```go
func (xc *XContext) Debug(message string)
```
- **Purpose:** Logs at DEBUG level with context fields
- **Performance:** Optimized for development scenarios
- **Behavior:** No-op if logger level > DebugLevel

```go
func (xc *XContext) Warn(message string)
```
- **Purpose:** Logs at WARN level with context fields
- **Use Case:** Non-fatal issues requiring attention

```go
func (xc *XContext) Error(message string)
```
- **Purpose:** Logs at ERROR level with context fields
- **Use Case:** Error conditions requiring immediate attention

#### Diagnostic Methods

```go
func (xc *XContext) MemoryFootprint() int
```
- **Purpose:** Returns the total memory footprint of the context
- **Returns:** Memory usage in bytes
- **Use Case:** Performance profiling and optimization

```go
func (xc *XContext) GetBinarySize() int
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
logger := NewIrisLogger(InfoLevel)

// Single log with fields
ctx := logger.XFields(
    XStr("user_id", "12345"),
    XInt("request_size", 1024),
    XBool("authenticated", true),
)
ctx.Info("Request processed successfully")
```

### High-Performance Pattern

```go
logger := NewIrisLogger(InfoLevel)

// Pre-create context for reuse
ctx := logger.XFields(
    XStr("service", "api-gateway"),
    XStr("version", "v2.1.0"),
)

// Reuse context for multiple logs
for i := 0; i < 1000000; i++ {
    ctx.Info("Processing request")
}
```

### Field Reuse Pattern

```go
// Create fields once, reuse multiple times
userField := XStr("user", "john_doe")
serviceField := XStr("service", "auth")

ctx1 := logger.XFields(userField, serviceField, XStr("action", "login"))
ctx2 := logger.XFields(userField, serviceField, XStr("action", "logout"))

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
| `XFields` (3 fields) | 266 | 376 | 3 |
| `XStr()` creation | 21 | 0 | 0 |
| `XInt()` creation | 19 | 0 | 0 |
| `XBool()` creation | 18 | 0 | 0 |

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

1. **Start with `XStr/XInt/XBool`** constructors for type safety
2. **Use `XFields()`** for context creation
3. **Test with production load** to validate performance characteristics
4. **Profile regularly** to ensure zero-allocation goals are met

## Error Handling

### Nil Safety

All binary field operations are nil-safe and return zero values rather than panicking:

```go
var field XField  // Zero value
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
