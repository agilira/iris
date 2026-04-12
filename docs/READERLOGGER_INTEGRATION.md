# ReaderLogger Integration Guide

## Overview

The ReaderLogger extends the standard Iris Logger to process external log sources through SyncReader interfaces. External log records flow through the same Iris processing pipeline, gaining access to all features (encoding, sampling, backpressure, security).

## Architecture

```
External Logger -> SyncReader -> ReaderLogger -> Ring Buffer -> Encoder -> Output
                      |               |
                  Provider         Background
                  Module           Goroutine
```

### Processing Model

1. **Source Independence**: Each SyncReader operates in a dedicated goroutine
2. **Unified Pipeline**: All records flow through the same Iris processing pipeline
3. **Feature Inheritance**: External logs receive all Iris features (encoding, sampling, security)
4. **Performance Isolation**: Reader failures do not affect core logger performance

## SyncReader Interface

```go
type SyncReader interface {
    Read(ctx context.Context) (*Record, error)
    Close() error
}
```

Providers implement `SyncReader` to bridge external logging systems into Iris.

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
if err != nil {
    log.Fatal(err)
}
```

### With Options

```go
config := iris.Config{
    Output:    iris.WrapWriter(os.Stdout),
    Encoder:   iris.NewJSONEncoder(),
    Level:     iris.Debug,
    Capacity:  32768,
    BatchSize: 64,
}

logger, err := iris.NewReaderLogger(config, readers)
```

### Multi-Output with MultiWriter

```go
config := iris.Config{
    Output: iris.MultiWriter(
        iris.WrapWriter(os.Stdout),
        iris.WrapWriter(logFile),
    ),
    Encoder: iris.NewJSONEncoder(),
    Level:   iris.Info,
}

logger, err := iris.NewReaderLogger(config, readers)
```

## Lifecycle Management

### Initialization Sequence

1. Create SyncReader instances with appropriate configuration
2. Create ReaderLogger with config and readers
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
    reader := provider.New(provider.Config{})
    logger, err := iris.NewReaderLogger(config, []iris.SyncReader{reader})
    if err != nil {
        log.Fatal(err)
    }

    logger.Start()
    defer func() {
        if err := logger.Close(); err != nil {
            log.Printf("Logger close error: %v", err)
        }
    }()

    runApplication(reader)
}
```

## Performance Characteristics

| Component | Typical Performance |
|-----------|-------------------|
| Direct Iris logging | ~31 ns/op |
| SyncReader.Read() | 50-100 ns/op (provider-dependent) |
| Record conversion | 500-1000 ns/op |

### Memory Usage

- **Buffer Overhead**: SyncReader buffer size x record size
- **Processing Overhead**: Minimal additional allocation
- **Feature Overhead**: Standard Iris feature memory usage

## Error Handling

### Reader Errors

Reader errors are handled gracefully by the background goroutine:

1. **Temporary Errors**: Logged via the main logger, processing continues
2. **Context Cancellation**: Clean shutdown initiated
3. **Fatal Errors**: Reader goroutine terminates, other readers continue

### Error Logging

```go
// Reader errors are logged through the main logger automatically
// "Failed to close reader" with error string field
```

## Monitoring

### Statistics

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

- High reader error rates (visible in logger output)
- Buffer overflow conditions (check `dropped` stat)
- Processing latency increases

## Troubleshooting

### High Memory Usage

**Causes:**
- Oversized reader buffers
- Slow record processing
- External library memory leaks

**Solutions:**
- Reduce buffer sizes via Config.Capacity
- Optimize record conversion
- Profile external library usage

### Performance Degradation

**Causes:**
- Reader conversion overhead
- Buffer contention
- Large batch sizes

**Solutions:**
- Optimize record conversion logic
- Adjust Capacity and BatchSize
- Profile with `go test -bench`

### Missing Records

**Causes:**
- Reader buffer overflow (check `dropped` stat)
- Context cancellation
- Reader implementation errors

**Solutions:**
- Increase Capacity
- Check reader error logs
- Validate SyncReader implementation

## Best Practices

1. **Buffer Sizing**: Start with default capacity, increase only if `dropped > 0`
2. **Error Handling**: Always check `NewReaderLogger` creation errors
3. **Lifecycle Management**: Use defer for proper `Close()` cleanup
4. **Resource Limits**: Monitor memory usage in production via `Stats()`

---

Iris -- an AGILira fragment
