# IRIS Sync() Integration Guide

## Table of Contents
- [Overview](#overview)
- [Production Pattern (Auto-Scaling Logger)](#production-pattern-auto-scaling-logger)
- [Development Pattern (Basic Logger)](#development-pattern-basic-logger)
- [Critical Use Cases](#critical-use-cases)
- [Integration Patterns](#integration-patterns)
- [Performance Considerations](#performance-considerations)
- [Common Pitfalls](#common-pitfalls)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)
- [Examples](#examples)

## Overview

The `Sync()` method in IRIS ensures data integrity by guaranteeing that all buffered log records are written to the output destination before returning. With Iris's auto-scaling architecture, sync operations are optimized automatically based on your workload patterns.

### Why Sync() Matters

**Data Integrity**: Without proper synchronization, log records may remain in internal buffers during application shutdown, leading to data loss.

**Compliance**: Many regulatory environments require guaranteed log persistence for audit trails.

**Debugging**: Critical error logs must be persisted to aid in post-mortem analysis.

## Production Pattern (Auto-Scaling Logger)

For production environments, use the AutoScalingLogger which automatically handles performance optimization:

```go
func main() {
    // Production pattern: Auto-scaling logger
    autoLogger, err := iris.NewAutoScalingLogger(
        iris.Config{
            Level:   iris.Info,
            Output:  os.Stdout,
            Encoder: iris.NewJSONEncoder(),
        },
        iris.DefaultAutoScalingConfig(), // Handles all optimization automatically
    )
    if err != nil {
        panic(err)
    }
    
    autoLogger.Start() // Auto-scaling system activates
    defer func() {
        // CRITICAL: Always call Sync() before Close()
        if err := autoLogger.Sync(); err != nil {
            fmt.Fprintf(os.Stderr, "Failed to sync logs: %v\n", err)
        }
        autoLogger.Close()
    }()
    
    // Use normally - auto-scaling optimizes performance automatically
    autoLogger.Info("Application started")
    // ... work ...
}
```

## Development Pattern (Basic Logger)

For simple use cases or development environments:

```go
func main() {
    logger, err := iris.New(iris.Config{
        Level:   iris.Info,
        Output:  os.Stdout,
        Encoder: iris.NewJSONEncoder(),
    })
    if err != nil {
        panic(err)
    }
    
    logger.Start()
    defer func() {
        // CRITICAL: Always call Sync() before Close()
        if err := logger.Sync(); err != nil {
            fmt.Fprintf(os.Stderr, "Failed to sync logs: %v\n", err)
        }
        logger.Close()
    }()
    
    // Your application logic here
    logger.Info("Application started")
    // ... work ...
    logger.Info("Application shutting down")
}
```

## Critical Use Cases

### 1. Production Web Application (Auto-Scaling Pattern)

```go
func main() {
    // Production-ready auto-scaling logger
    autoLogger, err := iris.NewAutoScalingLogger(
        iris.Config{
            Level:   iris.Info,
            Output:  os.Stdout,
            Encoder: iris.NewJSONEncoder(),
        },
        iris.DefaultAutoScalingConfig(),
    )
    if err != nil {
        panic(err)
    }
    
    autoLogger.Start()
    defer func() {
        // CRITICAL: Always sync before shutdown
        if err := autoLogger.Sync(); err != nil {
            fmt.Fprintf(os.Stderr, "Failed to sync logs: %v\n", err)
        }
        autoLogger.Close()
    }()
    
    // Use normally - auto-scaling handles optimization
    autoLogger.Info("Server starting on :8080")
    
    // Graceful shutdown with signal handling
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            autoLogger.Info("Request received",
                iris.Str("method", r.Method),
                iris.Str("path", r.URL.Path),
            )
            w.WriteHeader(http.StatusOK)
        })
        http.ListenAndServe(":8080", nil)
    }()
    
    <-c
    autoLogger.Info("Shutting down gracefully")
    // Sync() called automatically in defer
}
```
```

### 2. Fatal Error Handling

```go
func handleCriticalError(logger *iris.Logger, err error) {
    logger.Error("Critical system failure", iris.Field{Key: "error", Value: err})
    
    // Ensure the error is persisted before terminating
    if syncErr := logger.Sync(); syncErr != nil {
        fmt.Fprintf(os.Stderr, "Failed to sync critical error: %v\n", syncErr)
    }
    
    os.Exit(1)
}
```

### 3. Batch Processing Checkpoints

```go
func processRecords(logger *iris.Logger, records []Record) error {
    for i, record := range records {
        if err := processRecord(record); err != nil {
            logger.Error("Failed to process record", 
                iris.Field{Key: "record_id", Value: record.ID},
                iris.Field{Key: "error", Value: err})
            
            // Sync every 100 records or on error
            if i%100 == 0 || err != nil {
                if syncErr := logger.Sync(); syncErr != nil {
                    return fmt.Errorf("sync failed: %w", syncErr)
                }
            }
        }
    }
    return nil
}
```

## Integration Patterns

### Pattern 1: Simple Synchronous

**Use Case**: Small applications, development environments
**Performance**: Good for low-volume logging
**Complexity**: Low

```go
logger.Info("Operation completed")
if err := logger.Sync(); err != nil {
    return fmt.Errorf("failed to sync logs: %w", err)
}
```

### Pattern 2: Deferred Sync

**Use Case**: Function-level guarantees, cleanup operations
**Performance**: Minimal overhead
**Complexity**: Low

```go
func criticalOperation() error {
    defer func() {
        if err := logger.Sync(); err != nil {
            log.Printf("Sync failed: %v", err)
        }
    }()
    
    logger.Info("Starting critical operation")
    // ... operation logic ...
    return nil
}
```

### Pattern 3: Periodic Sync

**Use Case**: High-throughput applications, balanced reliability
**Performance**: Excellent for high volume
**Complexity**: Medium

```go
type PeriodicLogger struct {
    logger   *iris.Logger
    interval time.Duration
    stop     chan struct{}
}

func (p *PeriodicLogger) Start() {
    ticker := time.NewTicker(p.interval)
    go func() {
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                if err := p.logger.Sync(); err != nil {
                    // Handle sync error
                    fmt.Printf("Periodic sync failed: %v\n", err)
                }
            case <-p.stop:
                return
            }
        }
    }()
}
```

### Pattern 4: Context-Aware Sync

**Use Case**: Request processing, transaction boundaries
**Performance**: Good for web applications
**Complexity**: Medium

```go
func handleRequest(ctx context.Context, logger *iris.Logger) error {
    // Create request-scoped logger
    reqLogger := logger.With(iris.Field{Key: "request_id", Value: requestID(ctx)})
    
    defer func() {
        if deadline, ok := ctx.Deadline(); ok {
            // Ensure sync completes before context deadline
            syncCtx, cancel := context.WithDeadline(context.Background(), deadline)
            defer cancel()
            
            done := make(chan error, 1)
            go func() {
                done <- reqLogger.Sync()
            }()
            
            select {
            case err := <-done:
                if err != nil {
                    log.Printf("Request sync failed: %v", err)
                }
            case <-syncCtx.Done():
                log.Printf("Request sync timeout")
            }
        }
    }()
    
    reqLogger.Info("Processing request")
    // ... request processing ...
    return nil
}
```

## Performance Considerations

### Sync() Performance Characteristics

| Operation | Typical Latency | Max Latency | Notes |
|-----------|----------------|-------------|-------|
| Empty buffer sync | < 1μs | < 10μs | Immediate return |
| Small batch (1-10 records) | 100μs - 1ms | 10ms | Depends on output destination |
| Large batch (100+ records) | 1ms - 10ms | 100ms | Linear with record count |
| Slow I/O destination | 10ms - 1s | 5s | Network/disk dependent |

### Optimization Guidelines

#### With Auto-Scaling Logger (Production):
- Use `iris.NewAutoScalingLogger()` for production environments
- Let the auto-scaling system handle all performance optimization
- No manual buffer size or batch configuration needed
- Focus only on proper Sync() calls during shutdown

#### With Basic Logger (Development):
- Use periodic sync for high-throughput applications
- Combine multiple log operations before syncing
- Suitable for development or simple use cases

#### Universal Guidelines:
- Always call Sync() before application shutdown
- Handle Sync() errors appropriately
- Use signal handlers for graceful shutdown

### Configuration Comparison

```go
// PRODUCTION: Auto-scaling handles everything
autoLogger, _ := iris.NewAutoScalingLogger(
    iris.Config{
        Level:   iris.Info,
        Output:  output,
        Encoder: iris.NewJSONEncoder(),
    },
    iris.DefaultAutoScalingConfig(), // Automatic optimization
)

// DEVELOPMENT: Manual configuration for simple cases
logger, _ := iris.New(iris.Config{
    Level:     iris.Info,
    Output:    output,
    Encoder:   iris.NewJSONEncoder(),
    Capacity:  4096,  // Manual buffer size
    BatchSize: 256,   // Manual batch processing
})
```

## Common Pitfalls

### 1. Mutex Deadlocks in Custom WriteSyncers

**Problem**: Using the same mutex in nested WriteSyncer implementations

```go
// ❌ WRONG: Deadlock risk
type SlowWriter struct {
    writer io.Writer
    mu     *sync.Mutex
}

type BufferWriter struct {
    buf *bytes.Buffer
    mu  *sync.Mutex  // Same mutex as SlowWriter!
}

func main() {
    var mu sync.Mutex
    buf := &BufferWriter{buf: &bytes.Buffer{}, mu: &mu}
    slow := &SlowWriter{writer: buf, mu: &mu}  // DEADLOCK!
}
```

**Solution**: Use separate mutexes or rely on inner writer's synchronization

```go
// ✅ CORRECT: Separate synchronization
type SlowWriter struct {
    writer iris.WriteSyncer
    delay  time.Duration
    // No mutex - delegate to inner writer
}

func (s *SlowWriter) Write(p []byte) (n int, err error) {
    time.Sleep(s.delay)  // Simulate slow I/O
    return s.writer.Write(p)
}

func (s *SlowWriter) Sync() error {
    return s.writer.Sync()
}
```

### 2. Ignoring Sync() Errors

**Problem**: Silent failures during critical operations

```go
// ❌ WRONG: Ignoring critical errors
logger.Error("Database connection failed")
logger.Sync()  // Error ignored!
os.Exit(1)
```

**Solution**: Always handle Sync() errors appropriately

```go
// ✅ CORRECT: Handle sync errors
logger.Error("Database connection failed")
if err := logger.Sync(); err != nil {
    fmt.Fprintf(os.Stderr, "Critical: Failed to persist error log: %v\n", err)
    // Consider alternative logging (file, syslog, etc.)
}
os.Exit(1)
```

### 3. Sync() in Hot Paths

**Problem**: Performance degradation from excessive syncing

```go
// ❌ WRONG: Sync in loop
for _, item := range items {
    logger.Info("Processing item", iris.Field{Key: "id", Value: item.ID})
    logger.Sync()  // Performance killer!
}
```

**Solution**: Batch operations and sync periodically

```go
// ✅ CORRECT: Periodic sync
for i, item := range items {
    logger.Info("Processing item", iris.Field{Key: "id", Value: item.ID})
    
    // Sync every 100 items or at the end
    if i%100 == 0 || i == len(items)-1 {
        if err := logger.Sync(); err != nil {
            return fmt.Errorf("sync failed at item %d: %w", i, err)
        }
    }
}
```

### 4. Race Conditions with Close()

**Problem**: Calling Sync() after Close() or concurrently

```go
// ❌ WRONG: Race condition
go func() {
    logger.Close()  // Goroutine 1
}()

go func() {
    logger.Sync()   // Goroutine 2 - may fail!
}()
```

**Solution**: Coordinate lifecycle properly

```go
// ✅ CORRECT: Proper coordination
var wg sync.WaitGroup

// Sync first, then close
if err := logger.Sync(); err != nil {
    log.Printf("Final sync failed: %v", err)
}
logger.Close()
```

## Troubleshooting

### Sync() Timeout Errors

**Symptom**: `flush timeout: target_pos=X, reader_pos=Y`

**Causes**:
1. Slow WriteSyncer implementation
2. Consumer goroutine not running
3. Deadlock in custom WriteSyncer
4. Excessive backpressure

**Diagnosis**:
```go
// Check logger stats before sync
stats := logger.Stats()
fmt.Printf("Pre-sync stats: %+v\n", stats)

err := logger.Sync()
if err != nil {
    stats := logger.Stats()
    fmt.Printf("Post-sync stats: %+v\n", stats)
    fmt.Printf("Sync error: %v\n", err)
}
```

**Solutions**:
1. Optimize WriteSyncer performance
2. Ensure logger.Start() was called
3. Review mutex usage in custom writers
4. Increase buffer capacity

### Consumer Not Processing

**Symptom**: `processed=0, size=N` (messages not being consumed)

**Causes**:
1. logger.Start() not called
2. Panic in consumer goroutine
3. Deadlock in processor function
4. WriteSyncer blocking indefinitely

**Diagnosis**:
```go
// Monitor stats over time
for i := 0; i < 10; i++ {
    stats := logger.Stats()
    fmt.Printf("Stats[%d]: %+v\n", i, stats)
    time.Sleep(100 * time.Millisecond)
}
```

### Memory Leaks

**Symptom**: Growing memory usage, high `size` in stats

**Causes**:
1. Sync() never called
2. WriteSyncer not draining properly
3. Buffer capacity too large for workload

**Solutions**:
1. Implement periodic sync
2. Review WriteSyncer implementation
3. Tune buffer size appropriately

## FAQ

### Q: How often should I call Sync()?

**A**: It depends on your use case:
- **Critical logs**: Immediately after logging
- **High-throughput**: Every 1-10 seconds or N operations
- **Application shutdown**: Always before Close()
- **Error conditions**: Immediately for critical errors

### Q: What's the performance impact of Sync()?

**A**: Sync() latency depends on:
- Number of buffered records
- WriteSyncer implementation speed
- I/O destination (file, network, etc.)

Typical range: 100μs - 100ms. Use periodic sync for high-throughput scenarios.

### Q: Can I call Sync() from multiple goroutines?

**A**: Yes, Sync() is thread-safe. However, multiple concurrent Sync() calls will serialize and each will wait for all previous records to be processed.

### Q: What happens if Sync() times out?

**A**: Sync() returns an error after 5 seconds. The logger remains functional, but some records may not have been written to the output destination.

### Q: Should I check Sync() errors?

**A**: Always check Sync() errors in production:
- Critical paths: Log to stderr and consider alternative output
- Non-critical paths: Log the error for monitoring

### Q: Can I customize the Sync() timeout?

**A**: Currently, the timeout is fixed at 5 seconds. This ensures bounded latency while allowing for slow I/O operations.

### Q: How does Sync() interact with sampling?

**A**: Sync() only affects records that pass through sampling. Dropped records are not included in sync operations.

## Examples

### Example 1: Web Server Integration

```go
package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "time"
    
    "github.com/agilira/iris"
)

func main() {
    logger, err := iris.New(iris.Config{
        Level:    iris.Info,
        Output:   os.Stdout,
        Encoder:  iris.NewJSONEncoder(),
        Capacity: 2048,
    })
    if err != nil {
        panic(err)
    }
    
    logger.Start()
    defer func() {
        logger.Info("Server shutting down")
        if err := logger.Sync(); err != nil {
            logger.Error("Failed to sync on shutdown", iris.Field{Key: "error", Value: err})
        }
        logger.Close()
    }()
    
    // Periodic sync for reliability
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            if err := logger.Sync(); err != nil {
                logger.Error("Periodic sync failed", iris.Field{Key: "error", Value: err})
            }
        }
    }()
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: createHandler(logger),
    }
    
    // Graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    
    go func() {
        <-c
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        server.Shutdown(ctx)
    }()
    
    logger.Info("Server starting on :8080")
    if err := server.ListenAndServe(); err != http.ErrServerClosed {
        logger.Error("Server failed", iris.Field{Key: "error", Value: err})
    }
}

func createHandler(logger *iris.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := generateRequestID()
        
        reqLogger := logger.With(
            iris.Field{Key: "request_id", Value: requestID},
            iris.Field{Key: "method", Value: r.Method},
            iris.Field{Key: "path", Value: r.URL.Path},
        )
        
        defer func() {
            duration := time.Since(start)
            reqLogger.Info("Request completed",
                iris.Field{Key: "duration_ms", Value: duration.Milliseconds()},
            )
            
            // Sync for errors or slow requests
            if duration > 5*time.Second {
                if err := reqLogger.Sync(); err != nil {
                    // Log sync failure to system logger
                    logger.Error("Failed to sync slow request", iris.Field{Key: "error", Value: err})
                }
            }
        }()
        
        reqLogger.Info("Request started")
        
        // Handle request...
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }
}
```

### Example 2: Batch Processing System

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/agilira/iris"
)

type BatchProcessor struct {
    logger       *iris.Logger
    syncInterval int
    syncTimeout  time.Duration
}

func NewBatchProcessor(logger *iris.Logger) *BatchProcessor {
    return &BatchProcessor{
        logger:       logger,
        syncInterval: 100,           // Sync every 100 operations
        syncTimeout:  30 * time.Second,
    }
}

func (bp *BatchProcessor) ProcessBatch(items []WorkItem) error {
    bp.logger.Info("Starting batch processing",
        iris.Field{Key: "batch_size", Value: len(items)},
    )
    
    var processed int
    var errors []error
    
    for i, item := range items {
        if err := bp.processItem(item); err != nil {
            errors = append(errors, err)
            bp.logger.Error("Item processing failed",
                iris.Field{Key: "item_id", Value: item.ID},
                iris.Field{Key: "error", Value: err},
            )
        } else {
            processed++
            bp.logger.Debug("Item processed successfully",
                iris.Field{Key: "item_id", Value: item.ID},
            )
        }
        
        // Periodic sync for reliability
        if (i+1)%bp.syncInterval == 0 {
            if err := bp.syncWithTimeout(); err != nil {
                bp.logger.Error("Periodic sync failed",
                    iris.Field{Key: "position", Value: i + 1},
                    iris.Field{Key: "error", Value: err},
                )
            }
        }
    }
    
    // Final sync
    if err := bp.syncWithTimeout(); err != nil {
        bp.logger.Error("Final batch sync failed", iris.Field{Key: "error", Value: err})
        return fmt.Errorf("failed to sync batch logs: %w", err)
    }
    
    bp.logger.Info("Batch processing completed",
        iris.Field{Key: "processed", Value: processed},
        iris.Field{Key: "errors", Value: len(errors)},
    )
    
    if len(errors) > 0 {
        return fmt.Errorf("batch completed with %d errors", len(errors))
    }
    
    return nil
}

func (bp *BatchProcessor) syncWithTimeout() error {
    done := make(chan error, 1)
    
    go func() {
        done <- bp.logger.Sync()
    }()
    
    select {
    case err := <-done:
        return err
    case <-time.After(bp.syncTimeout):
        return fmt.Errorf("sync timeout after %v", bp.syncTimeout)
    }
}

func (bp *BatchProcessor) processItem(item WorkItem) error {
    // Simulate work
    time.Sleep(10 * time.Millisecond)
    return nil
}

type WorkItem struct {
    ID   string
    Data interface{}
}
```

---

## Best Practices Summary

1. **Always call Sync() before application shutdown**
2. **Handle Sync() errors appropriately for your use case**
3. **Use periodic sync for high-throughput applications**
4. **Avoid calling Sync() in hot code paths**
5. **Be careful with mutex usage in custom WriteSyncers**
6. **Monitor Sync() performance and tune buffer sizes**
7. **Use context-aware patterns for bounded operations**
8. **Test your integration under load and failure conditions**
