# SyncReader Troubleshooting Guide

## Overview

This guide provides systematic troubleshooting procedures for SyncReader interface implementations and ReaderLogger integrations. It covers common issues, diagnostic procedures, and resolution strategies.

## Diagnostic Framework

### Error Categories

1. **Integration Errors**: Provider interface implementation issues
2. **Performance Issues**: Throughput or latency problems
3. **Resource Leaks**: Memory or goroutine leaks
4. **Configuration Problems**: Invalid or suboptimal configuration
5. **Concurrency Issues**: Race conditions or deadlocks

### Diagnostic Tools

```go
// Enable debug logging for troubleshooting
logger, err := iris.NewReaderLogger(iris.Config{
    Level:   iris.Debug,
    Output:  iris.WrapWriter(os.Stderr),
    Encoder: iris.NewTextEncoder(),
}, providers)

// Monitor goroutine count
func monitorGoroutines() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
    }
}

// Memory usage monitoring
func monitorMemory() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("Alloc: %d KB, Sys: %d KB\n", 
        m.Alloc/1024, m.Sys/1024)
}
```

## Common Issues and Solutions

### 1. Provider Not Receiving Records

#### Symptoms
- No output from ReaderLogger despite external library logging
- `Read()` method never returns records
- Empty log files

#### Diagnostic Steps

```go
// Add debug output to provider
func (p *Provider) Handle(ctx context.Context, record ExternalRecord) error {
    fmt.Printf("DEBUG: Received record: %+v\n", record) // Debug line
    select {
    case p.records <- record:
        fmt.Printf("DEBUG: Record queued successfully\n") // Debug line
        return nil
    case <-p.closed:
        return fmt.Errorf("provider closed")
    default:
        fmt.Printf("DEBUG: Buffer full, dropping record\n") // Debug line
        return nil
    }
}
```

#### Common Causes and Solutions

| Cause | Solution |
|-------|----------|
| External library not configured to use provider | Verify handler registration: `logger.SetHandler(provider)` |
| Provider constructor not called | Ensure `New()` is called and provider is passed to ReaderLogger |
| Buffer size too small | Increase buffer size in provider configuration |
| Context cancellation | Check context lifetime and cancellation conditions |

### 2. Performance Degradation

#### Symptoms
- Logging throughput below expected performance
- High CPU usage in logging code
- Memory allocation spikes during logging

#### Performance Profiling

```go
// CPU profiling
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Memory profiling in tests
func BenchmarkProvider_Performance(b *testing.B) {
    // ... setup code ...
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        provider.Handle(ctx, record)
    }
}
```

#### Performance Analysis

```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profile  
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

#### Performance Optimization

| Issue | Optimization |
|-------|-------------|
| High allocation rate | Pool field objects, reuse records |
| Slow field conversion | Optimize type switches, avoid reflection |
| Buffer contention | Increase buffer size or use multiple providers |
| Context overhead | Reuse contexts where appropriate |

### 3. Memory Leaks

#### Symptoms
- Continuously increasing memory usage
- Growing number of goroutines
- Eventually OOM or resource exhaustion

#### Leak Detection

```go
// Goroutine leak detection
func detectGoroutineLeaks(t *testing.T) {
    initialCount := runtime.NumGoroutine()
    
    // ... test code ...
    
    // Allow cleanup time
    time.Sleep(100 * time.Millisecond)
    runtime.GC()
    
    finalCount := runtime.NumGoroutine()
    if finalCount > initialCount {
        t.Errorf("Goroutine leak detected: %d -> %d", 
            initialCount, finalCount)
    }
}

// Memory leak detection
func detectMemoryLeaks(t *testing.T) {
    runtime.GC()
    var before runtime.MemStats
    runtime.ReadMemStats(&before)
    
    // ... test code ...
    
    runtime.GC()
    var after runtime.MemStats
    runtime.ReadMemStats(&after)
    
    if after.Alloc > before.Alloc*2 {
        t.Errorf("Potential memory leak: %d -> %d", 
            before.Alloc, after.Alloc)
    }
}
```

#### Common Leak Sources

| Source | Prevention |
|--------|------------|
| Unclosed providers | Always call `provider.Close()` in defer statements |
| Unreleased contexts | Use context.WithTimeout for bounded operations |
| Retained records | Clear record references after processing |
| Goroutine not terminated | Ensure proper shutdown signaling |

### 4. Record Loss or Corruption

#### Symptoms
- Missing log records in output
- Corrupted field values
- Incorrect log levels or timestamps

#### Validation Framework

```go
type RecordValidator struct {
    received int64
    processed int64
}

func (v *RecordValidator) TrackReceived() {
    atomic.AddInt64(&v.received, 1)
}

func (v *RecordValidator) TrackProcessed() {
    atomic.AddInt64(&v.processed, 1)
}

func (v *RecordValidator) GetStats() (received, processed int64) {
    return atomic.LoadInt64(&v.received), 
           atomic.LoadInt64(&v.processed)
}

// Usage in provider
func (p *Provider) Handle(ctx context.Context, record ExternalRecord) error {
    p.validator.TrackReceived()
    
    select {
    case p.records <- record:
        return nil
    default:
        // Record dropped due to full buffer
        return nil
    }
}

func (p *Provider) Read(ctx context.Context) (*iris.Record, error) {
    // ... read logic ...
    
    if record != nil {
        p.validator.TrackProcessed()
    }
    
    return record, err
}
```

#### Field Validation

```go
func validateFieldConversion(original ExternalField, converted iris.Field) error {
    // Type consistency check
    switch original.Value.(type) {
    case string:
        if converted.Type != iris.StringType {
            return fmt.Errorf("type mismatch: expected string, got %s", 
                converted.Type)
        }
    case int, int64:
        if converted.Type != iris.Int64Type {
            return fmt.Errorf("type mismatch: expected int64, got %s", 
                converted.Type)
        }
    // ... additional type checks ...
    }
    
    // Value consistency check
    if !reflect.DeepEqual(original.Key, converted.Key) {
        return fmt.Errorf("key mismatch: %s != %s", 
            original.Key, converted.Key)
    }
    
    return nil
}
```

### 5. Concurrency Issues

#### Symptoms
- Intermittent panics or crashes
- Deadlocks or hanging operations
- Data races detected by race detector

#### Race Detection

```bash
# Run tests with race detector
go test -race ./...

# Run application with race detector
go run -race main.go
```

#### Safe Concurrency Patterns

```go
// Thread-safe provider implementation
type Provider struct {
    mu      sync.RWMutex
    records chan ExternalRecord
    closed  chan struct{}
    once    sync.Once
    state   int32 // Use atomic operations
}

const (
    stateOpen = iota
    stateClosed
)

func (p *Provider) Close() error {
    p.once.Do(func() {
        atomic.StoreInt32(&p.state, stateClosed)
        close(p.closed)
    })
    return nil
}

func (p *Provider) isClosed() bool {
    return atomic.LoadInt32(&p.state) == stateClosed
}

func (p *Provider) Handle(ctx context.Context, record ExternalRecord) error {
    if p.isClosed() {
        return fmt.Errorf("provider closed")
    }
    
    select {
    case p.records <- record:
        return nil
    case <-p.closed:
        return fmt.Errorf("provider closed")
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## Configuration Issues

### Buffer Size Optimization

```go
// Calculate optimal buffer size
func calculateBufferSize(expectedTPS int, processingLatencyMs int) int {
    // Buffer size = TPS * latency in seconds * safety factor
    latencySeconds := float64(processingLatencyMs) / 1000.0
    safetyFactor := 2.0
    
    bufferSize := int(float64(expectedTPS) * latencySeconds * safetyFactor)
    
    // Minimum and maximum bounds
    if bufferSize < 100 {
        bufferSize = 100
    }
    if bufferSize > 100000 {
        bufferSize = 100000
    }
    
    return bufferSize
}
```

### Context Configuration

```go
// Proper context usage
func createProviderContext() (context.Context, context.CancelFunc) {
    // Use background context for long-running providers
    ctx := context.Background()
    
    // Add timeout for bounded operations
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    
    return ctx, cancel
}
```

## Monitoring and Observability

### Metrics Collection

```go
type ProviderMetrics struct {
    RecordsReceived   int64
    RecordsProcessed  int64
    RecordsDropped    int64
    ProcessingLatency time.Duration
    ErrorCount        int64
}

func (m *ProviderMetrics) IncrementReceived() {
    atomic.AddInt64(&m.RecordsReceived, 1)
}

func (m *ProviderMetrics) IncrementProcessed() {
    atomic.AddInt64(&m.RecordsProcessed, 1)
}

func (m *ProviderMetrics) IncrementDropped() {
    atomic.AddInt64(&m.RecordsDropped, 1)
}

func (m *ProviderMetrics) IncrementErrors() {
    atomic.AddInt64(&m.ErrorCount, 1)
}

func (m *ProviderMetrics) GetSnapshot() ProviderMetrics {
    return ProviderMetrics{
        RecordsReceived:  atomic.LoadInt64(&m.RecordsReceived),
        RecordsProcessed: atomic.LoadInt64(&m.RecordsProcessed),
        RecordsDropped:   atomic.LoadInt64(&m.RecordsDropped),
        ErrorCount:       atomic.LoadInt64(&m.ErrorCount),
    }
}
```

### Health Checks

```go
func (p *Provider) HealthCheck() error {
    if p.isClosed() {
        return fmt.Errorf("provider is closed")
    }
    
    // Check buffer usage
    bufferUsage := float64(len(p.records)) / float64(cap(p.records))
    if bufferUsage > 0.9 {
        return fmt.Errorf("buffer usage critical: %.1f%%", bufferUsage*100)
    }
    
    // Check error rate
    metrics := p.metrics.GetSnapshot()
    total := metrics.RecordsReceived
    if total > 0 {
        errorRate := float64(metrics.ErrorCount) / float64(total)
        if errorRate > 0.1 {
            return fmt.Errorf("error rate too high: %.1f%%", errorRate*100)
        }
    }
    
    return nil
}
```

## Testing Strategies

### Unit Test Template

```go
func TestProvider_Comprehensive(t *testing.T) {
    tests := []struct {
        name        string
        setup       func() *Provider
        operation   func(*Provider) error
        expectError bool
        cleanup     func(*Provider)
    }{
        {
            name: "normal operation",
            setup: func() *Provider {
                return New(Config{BufferSize: 10})
            },
            operation: func(p *Provider) error {
                ctx := context.Background()
                record := createTestRecord()
                return p.Handle(ctx, record)
            },
            expectError: false,
            cleanup: func(p *Provider) {
                p.Close()
            },
        },
        // ... additional test cases ...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider := tt.setup()
            defer tt.cleanup(provider)
            
            err := tt.operation(provider)
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Load Testing

```go
func TestProvider_LoadTest(t *testing.T) {
    provider := New(Config{BufferSize: 10000})
    defer provider.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Start consumer
    go func() {
        for {
            record, err := provider.Read(ctx)
            if err != nil || record == nil {
                return
            }
        }
    }()
    
    // Generate load
    const numGoroutines = 10
    const recordsPerGoroutine = 1000
    
    var wg sync.WaitGroup
    wg.Add(numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        go func() {
            defer wg.Done()
            for j := 0; j < recordsPerGoroutine; j++ {
                record := createTestRecord()
                provider.Handle(ctx, record)
            }
        }()
    }
    
    wg.Wait()
    
    // Verify no resource leaks
    runtime.GC()
    assert.True(t, runtime.NumGoroutine() < 20, "goroutine leak detected")
}
```

## Emergency Procedures

### Provider Recovery

```go
func (p *Provider) Recover() error {
    // Stop all operations
    p.Close()
    
    // Clear buffer
    for len(p.records) > 0 {
        <-p.records
    }
    
    // Reset state
    p.records = make(chan ExternalRecord, p.config.BufferSize)
    p.closed = make(chan struct{})
    p.once = sync.Once{}
    atomic.StoreInt32(&p.state, stateOpen)
    
    return nil
}
```

### Graceful Shutdown

```go
func (p *Provider) GracefulShutdown(timeout time.Duration) error {
    done := make(chan struct{})
    
    go func() {
        // Process remaining records
        for len(p.records) > 0 {
            time.Sleep(10 * time.Millisecond)
        }
        close(done)
    }()
    
    select {
    case <-done:
        return p.Close()
    case <-time.After(timeout):
        return fmt.Errorf("shutdown timeout exceeded")
    }
}
```

---

Iris • an AGILira fragment
