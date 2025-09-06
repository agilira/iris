# ReaderLogger Integration Guide

## Overview

The ReaderLogger extends the standard Iris Logger to process external log sources through SyncReader interfaces. This document provides technical guidance for integrating external logging systems with Iris's high-performance pipeline.

## ReaderLogger Architecture

### Component Interaction

```
External Logger → SyncReader → ReaderLogger → Ring Buffer → Encoder → Output
                     ↑              ↑             ↑
                 Provider        Background     Features
                 Module          Goroutine      (OTel, Loki, etc.)
```

### Processing Model

1. **Source Independence**: Each SyncReader operates in a dedicated goroutine
2. **Unified Pipeline**: All records flow through the same Iris processing pipeline
3. **Feature Inheritance**: External logs automatically receive all Iris features
4. **Performance Isolation**: Reader failures do not affect core logger performance

## Configuration

### Basic Configuration

```go
config := iris.Config{
    Output:  iris.WrapWriter(os.Stdout),
    Encoder: iris.NewJSONEncoder(),
    Level:   iris.Info,
}

readers := []iris.SyncReader{
    provider1.New(provider1.Config{}),
    provider2.New(provider2.Config{}),
}

logger, err := iris.NewReaderLogger(config, readers)
```

### Advanced Configuration

```go
config := iris.Config{
    Output: iris.MultiWriter(
        iris.WrapWriter(os.Stdout),
        iris.NewLokiWriter(iris.LokiConfig{
            URL: "http://loki:3100",
            Labels: map[string]string{
                "service": "my-app",
                "source":  "external-readers",
            },
        }),
    ),
    Encoder:   iris.NewJSONEncoder(),
    Level:     iris.Debug,
    Capacity:  32768,
    BatchSize: 64,
}

options := []iris.Option{
    iris.WithOTel(),
    iris.WithCaller(),
    iris.AddStacktrace(iris.Error),
}

logger, err := iris.NewReaderLogger(config, readers, options...)
```

## Lifecycle Management

### Initialization Sequence

1. Create SyncReader instances with appropriate configuration
2. Create ReaderLogger with config, readers, and options
3. Call `Start()` to begin background processing
4. Begin using external logging libraries

### Shutdown Sequence

1. Stop external logging operations
2. Call `Close()` on ReaderLogger
3. ReaderLogger closes all SyncReader instances
4. All buffered records are processed before termination

### Example Lifecycle

```go
func main() {
    // Setup
    reader := provider.New(provider.Config{})
    logger, err := iris.NewReaderLogger(config, []iris.SyncReader{reader})
    if err != nil {
        log.Fatal(err)
    }
    
    // Start processing
    logger.Start()
    
    // Ensure clean shutdown
    defer func() {
        if err := logger.Close(); err != nil {
            log.Printf("Logger close error: %v", err)
        }
    }()
    
    // Application logic
    runApplication(reader)
}
```

## Performance Characteristics

### Throughput Expectations

| Component | Performance Target |
|-----------|-------------------|
| Direct Iris logging | 31 ns/op |
| SyncReader.Read() | 50-100 ns/op |
| Record conversion | 500-1000 ns/op |
| Overall throughput | 10-20x faster than direct external library |

### Memory Usage

- **Buffer Overhead**: SyncReader buffer size × record size
- **Processing Overhead**: Minimal additional allocation
- **Feature Overhead**: Standard Iris feature memory usage

### Latency Characteristics

- **Processing Latency**: Sub-millisecond record processing
- **Buffer Latency**: Dependent on buffer size and throughput
- **Feature Latency**: Standard Iris feature processing time

## Error Handling

### Reader Errors

Reader errors are handled gracefully:

1. **Temporary Errors**: Logged and processing continues
2. **Context Cancellation**: Clean shutdown initiated
3. **Reader Closure**: Reader goroutine terminates
4. **Fatal Errors**: Reader marked as failed, processing continues with other readers

### Error Logging

```go
// Reader errors are logged through the main logger
logger.Error("Reader error", 
    iris.String("reader", "slog-provider"),
    iris.String("error", err.Error()))
```

### Recovery Mechanisms

- **Automatic Retry**: Not implemented (readers should handle internally)
- **Circuit Breaking**: Not implemented (providers should implement if needed)
- **Failover**: Multiple readers provide natural redundancy

## Feature Integration

### OpenTelemetry Integration

External log records automatically receive:

- Trace ID extraction from context
- Span ID correlation
- Baggage propagation
- Resource attribute injection

### Loki Integration

Records from external sources are:

- Batched efficiently for Loki ingestion
- Labeled with provider-specific labels
- Formatted according to Loki requirements
- Delivered with optimal performance characteristics

### Security Features

External records receive:

- Automatic secret redaction
- Log injection protection
- Field sanitization
- Security audit logging

## Monitoring and Observability

### Metrics Collection

ReaderLogger provides metrics for:

- Records processed per reader
- Reader error rates
- Buffer utilization
- Processing latency

### Health Monitoring

```go
stats := logger.Stats()
for key, value := range stats {
    fmt.Printf("%s: %d\n", key, value)
}
// Output:
// capacity: 32768
// batch_size: 64
// size: 12
// processed: 15432
// dropped: 0
```

### Alerting Considerations

Monitor for:

- High reader error rates
- Buffer overflow conditions
- Processing latency increases
- Reader goroutine failures

## Troubleshooting

### Common Issues

#### High Memory Usage

**Symptoms**: Increasing memory consumption
**Causes**: 
- Oversized reader buffers
- Slow record processing
- External library memory leaks

**Solutions**:
- Reduce buffer sizes
- Optimize record conversion
- Profile external library usage

#### Performance Degradation

**Symptoms**: Increased logging latency
**Causes**:
- Reader conversion overhead
- Buffer contention
- Feature processing load

**Solutions**:
- Optimize record conversion logic
- Adjust buffer sizes
- Profile feature usage

#### Missing Records

**Symptoms**: Records not appearing in output
**Causes**:
- Reader buffer overflow
- Context cancellation
- Reader implementation errors

**Solutions**:
- Increase buffer sizes
- Check reader error logs
- Validate reader implementation

### Debugging Tools

#### Reader Status

```go
// Check reader goroutine status
stats := logger.Stats()
if stats["reader_errors"] > 0 {
    // Investigate reader issues
}
```

#### Record Tracking

```go
// Enable debug logging to track record flow
logger := iris.NewReaderLogger(config, readers, 
    iris.Development(),
    iris.WithCaller())
```

## Best Practices

### Configuration Guidelines

1. **Buffer Sizing**: Start with 1000-10000 record buffers
2. **Error Handling**: Always check ReaderLogger creation errors
3. **Lifecycle Management**: Use defer for proper cleanup
4. **Resource Limits**: Monitor memory usage in production

### Performance Optimization

1. **Efficient Conversion**: Minimize allocations in record conversion
2. **Batch Processing**: Use appropriate batch sizes for output
3. **Feature Selection**: Enable only required features
4. **Buffer Tuning**: Adjust based on actual throughput requirements

### Production Deployment

1. **Health Checks**: Monitor reader status and error rates
2. **Resource Monitoring**: Track memory and CPU usage
3. **Error Alerting**: Alert on reader failures and high error rates
4. **Performance Baseline**: Establish baseline metrics for comparison
