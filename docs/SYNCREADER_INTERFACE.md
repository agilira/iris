# SyncReader Interface Specification

## Overview

The SyncReader interface provides a standardized mechanism for integrating external logging systems with Iris's high-performance logging pipeline. This interface enables existing logging libraries to benefit from Iris's advanced features without requiring modifications to application code.

## Interface Definition

```go
type SyncReader interface {
    Read(ctx context.Context) (*Record, error)
    io.Closer
}
```

### Methods

#### Read(ctx context.Context) (*Record, error)

Retrieves the next log record from the external logging system.

**Parameters:**
- `ctx`: Context for cancellation and timeouts

**Returns:**
- `*Record`: Iris log record, or nil when no more records are available
- `error`: Error condition, including context cancellation

**Behavior:**
- Method should block until a record is available or context is cancelled
- Return `(nil, nil)` when the reader is closed and no more records exist
- Return `(nil, ctx.Err())` when context is cancelled
- Implementation must be goroutine-safe

#### Close() error

Releases resources associated with the reader.

**Returns:**
- `error`: Any error encountered during resource cleanup

**Behavior:**
- Must be idempotent (safe to call multiple times)
- Should signal Read() operations to return
- Must release all allocated resources

## Record Structure

The Record structure represents a log entry optimized for high-performance processing:

```go
type Record struct {
    Level  Level     // Log level (Debug, Info, Warn, Error)
    Msg    string    // Primary log message
    Logger string    // Logger name/identifier
    Caller string    // Caller information (file:line)
    Stack  string    // Stack trace information
    // Internal fields array (32 fields maximum)
}
```

### Record Methods

- `AddField(field Field) bool`: Adds a structured field to the record
- `FieldCount() int`: Returns the number of active fields
- `GetField(index int) Field`: Retrieves field at specified index

## Implementation Guidelines

### Performance Requirements

1. **Non-blocking Reads**: Use buffered channels or similar mechanisms to prevent blocking the source logging system
2. **Efficient Conversion**: Minimize allocations during record conversion
3. **Bounded Buffers**: Implement buffer limits to prevent memory exhaustion

### Error Handling

1. **Graceful Degradation**: Drop records rather than blocking when buffers are full
2. **Context Respect**: Always check context cancellation in Read() loops
3. **Resource Cleanup**: Ensure proper cleanup in Close() implementation

### Thread Safety

All SyncReader implementations must be safe for concurrent access:
- Multiple goroutines may call Read() simultaneously
- Close() may be called while Read() operations are in progress
- Internal state must be protected with appropriate synchronization

## Example Implementation Pattern

```go
type ExampleReader struct {
    records chan ExternalRecord
    closed  chan struct{}
    once    sync.Once
}

func (r *ExampleReader) Read(ctx context.Context) (*Record, error) {
    select {
    case rec := <-r.records:
        return r.convertRecord(rec), nil
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-r.closed:
        return nil, nil
    }
}

func (r *ExampleReader) Close() error {
    r.once.Do(func() {
        close(r.closed)
    })
    return nil
}
```

## Integration with ReaderLogger

The ReaderLogger processes SyncReader instances in background goroutines:

1. Each SyncReader runs in a dedicated goroutine
2. Records are fed into Iris's ring buffer for processing
3. All Iris features (encoding, output, security) apply automatically
4. Failed readers are logged but do not affect other readers

## Level Mapping

External logging levels should be mapped to Iris levels as follows:

| External Level | Iris Level | Numeric Value |
|----------------|------------|---------------|
| Trace/Verbose  | Debug      | -1            |
| Debug          | Debug      | -1            |
| Info           | Info       | 0             |
| Warn/Warning   | Warn       | 1             |
| Error/Fatal    | Error      | 2             |

## Field Conversion

External logging fields should be converted to Iris Field types:

- String values: `iris.String(key, value)`
- Numeric values: `iris.Int64()`, `iris.Float64()`, etc.
- Boolean values: `iris.Bool(key, value)`
- Time values: `iris.Time(key, value)`
- Duration values: `iris.Dur(key, value)`
- Complex types: Convert to string representation

## Testing Requirements

SyncReader implementations should include:

1. **Unit Tests**: Verify Read() and Close() behavior
2. **Concurrency Tests**: Test multiple goroutines accessing reader
3. **Resource Tests**: Verify proper cleanup and resource management
4. **Integration Tests**: Test with actual ReaderLogger instances
5. **Benchmark Tests**: Measure conversion and throughput performance
