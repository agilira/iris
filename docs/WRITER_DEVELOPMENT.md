# Writer Development Guide

This guide covers developing external writer modules for Iris using the `SyncWriter` interface.

## Overview

Writers in Iris are external modules that implement the `SyncWriter` interface to send log records to various destinations. This modular approach keeps the core Iris library dependency-free while enabling powerful integrations.

## SyncWriter Interface

```go
type SyncWriter interface {
    WriteRecord(record *Record) error
    io.Closer
}
```

### Methods

- **WriteRecord(record *Record) error**: Processes a single log record for output
- **Close() error**: Gracefully shuts down the writer and flushes any pending data

## Creating a Writer Module

### 1. Module Structure

```
iris-writer-{destination}/
├── README.md
├── go.mod
├── go.sum
├── {destination}_writer.go
└── {destination}_writer_test.go
```

### 2. Implementation Template

```go
package {destination}writer

import (
    "sync/atomic"
    "github.com/agilira/iris"
)

type Config struct {
    // Configuration fields specific to your destination
    Endpoint    string
    BatchSize   int
    // ... other fields
}

type Writer struct {
    config Config
    // Internal state fields
    recordsWritten int64
    recordsDropped int64
    // ... other fields
}

func New(config Config) (*Writer, error) {
    // Validate configuration
    if config.Endpoint == "" {
        return nil, fmt.Errorf("endpoint is required")
    }
    
    // Set defaults
    if config.BatchSize == 0 {
        config.BatchSize = 1000
    }
    
    writer := &Writer{
        config: config,
        // Initialize other fields
    }
    
    // Start background processes if needed
    
    return writer, nil
}

func (w *Writer) WriteRecord(record *iris.Record) error {
    // Process the record
    // Handle batching, formatting, etc.
    
    atomic.AddInt64(&w.recordsWritten, 1)
    return nil
}

func (w *Writer) Close() error {
    // Flush any pending data
    // Stop background processes
    // Clean up resources
    return nil
}
```

### 3. Configuration Best Practices

- **Provide sensible defaults** for all optional configuration
- **Validate required fields** in the constructor
- **Use clear, descriptive field names**
- **Support common patterns** like batching, timeouts, retries

### 4. Performance Considerations

- **Batch records** when possible to reduce I/O overhead
- **Use background goroutines** for async processing
- **Implement proper backpressure** handling
- **Pool resources** like buffers and connections
- **Use atomic operations** for metrics

### 5. Error Handling

- **Non-blocking writes**: Never block the logging hot path
- **Graceful degradation**: Handle destination unavailability
- **Error callbacks**: Allow users to handle errors
- **Metrics**: Track success/failure rates

## Example: HTTP Writer

```go
func (w *Writer) WriteRecord(record *iris.Record) error {
    w.batchMu.Lock()
    w.batch = append(w.batch, record)
    shouldFlush := len(w.batch) >= w.config.BatchSize
    w.batchMu.Unlock()

    if shouldFlush {
        go w.flushBatch() // Non-blocking
    }

    atomic.AddInt64(&w.recordsWritten, 1)
    return nil
}

func (w *Writer) flushBatch() {
    w.batchMu.Lock()
    if len(w.batch) == 0 {
        w.batchMu.Unlock()
        return
    }
    currentBatch := w.batch
    w.batch = make([]*iris.Record, 0, w.config.BatchSize)
    w.batchMu.Unlock()

    if err := w.sendBatch(currentBatch); err != nil {
        atomic.AddInt64(&w.recordsDropped, int64(len(currentBatch)))
        if w.config.OnError != nil {
            w.config.OnError(err)
        }
    }
}
```

## Testing Writers

### Required Tests

1. **Constructor tests**: Valid/invalid configurations
2. **WriteRecord tests**: Basic functionality and error cases
3. **Close tests**: Proper cleanup and final flush
4. **Integration tests**: End-to-end with real destinations
5. **Performance tests**: Benchmarks for throughput

### Test Template

```go
func TestWriter_WriteRecord(t *testing.T) {
    writer, err := New(Config{
        Endpoint: "http://localhost:8080",
    })
    if err != nil {
        t.Fatalf("Failed to create writer: %v", err)
    }
    defer writer.Close()

    record := &iris.Record{
        Level: iris.Info,
        Msg:   "test message",
    }

    err = writer.WriteRecord(record)
    if err != nil {
        t.Errorf("WriteRecord() error = %v", err)
    }
}
```

## Documentation Requirements

Every writer module should include:

1. **README.md**: Overview, installation, usage examples
2. **Configuration reference**: All options with defaults
3. **Performance characteristics**: Throughput, latency, resource usage
4. **Error handling**: How errors are reported and handled

## External Dependencies

- **Use go-timecache**: For high-performance timestamps: `github.com/agilira/go-timecache`
- **Minimize dependencies**: Only add what's absolutely necessary
- **Version compatibility**: Support recent Go versions

## Publishing Guidelines

1. **Module naming**: `iris-writer-{destination}`
2. **Repository structure**: One writer per repository
3. **Versioning**: Follow semantic versioning
4. **Documentation**: Complete README and examples
5. **Testing**: Full test coverage with CI/CD

## Examples

- **iris-writer-loki**: Grafana Loki integration with batching and retries
- **iris-writer-s3**: AWS S3 output with compression and partitioning
- **iris-writer-kafka**: Apache Kafka producer with key routing

## Support

- **Core Interface**: Defined in main iris package
- **Community**: Discussion and examples in iris repository
- **Documentation**: Updated guides and best practices

---

Iris • an AGILira fragment
