# Log Sampling

Log sampling is a technique used to reduce log volume in high-throughput applications by selectively allowing only a subset of log messages to be processed. Iris implements a token bucket sampling algorithm that provides both burst capacity and sustained rate limiting.

## Overview

When applications generate high volumes of log data, several issues can arise:

- Overwhelming downstream log processing systems
- Excessive storage costs
- Performance degradation due to I/O overhead
- Log storms that can impact application stability

Iris sampling addresses these issues by implementing rate limiting at the logger level, ensuring that only a controlled number of log messages are processed while maintaining system performance.

## Token Bucket Algorithm

Iris implements a token bucket sampler that operates on the following principles:

1. **Capacity**: Maximum number of tokens the bucket can hold (burst capacity)
2. **Refill Rate**: Number of tokens added to the bucket per time interval
3. **Refill Interval**: Time period between token additions

Each log message consumes one token. If no tokens are available, the message is dropped.

### Algorithm Benefits

- **Burst Handling**: Allows temporary spikes in log volume up to bucket capacity
- **Rate Limiting**: Maintains sustained rate control over time
- **Performance**: Lock-free implementation using atomic operations
- **Predictable**: Deterministic behavior under load

## Configuration

### Config-based Configuration

```go
sampler := iris.NewTokenBucketSampler(
    100,              // capacity: 100 messages burst
    10,               // refill: 10 messages per interval
    time.Second,      // every: 1 second intervals
)

logger, err := iris.New(iris.Config{
    Level:   iris.Info,
    Output:  os.Stdout,
    Encoder: iris.NewJSONEncoder(),
    Sampler: sampler,
})
```

### Options-based Configuration

```go
sampler := iris.NewTokenBucketSampler(100, 10, time.Second)

logger, err := iris.New(iris.Config{
    Level:   iris.Info,
    Output:  os.Stdout,
    Encoder: iris.NewJSONEncoder(),
}, iris.WithSampler(sampler))
```

### Cloned Logger Configuration

```go
baseLogger, _ := iris.New(iris.Config{...})

// Create sampled variant
sampler := iris.NewTokenBucketSampler(50, 5, time.Second)
sampledLogger := baseLogger.WithOptions(iris.WithSampler(sampler))
```

## Parameter Guidelines

### Capacity Selection

The capacity parameter determines burst tolerance:

- **Low capacity (1-10)**: Strict rate limiting, minimal burst tolerance
- **Medium capacity (10-100)**: Balanced approach for typical applications
- **High capacity (100+)**: Accommodates large bursts, suitable for batch processing

### Refill Rate Selection

The refill rate controls sustained throughput:

- **Conservative (1-10/sec)**: For error logging or critical events
- **Moderate (10-100/sec)**: For general application logging
- **Aggressive (100+/sec)**: For debug logging or high-frequency events

### Timing Considerations

Refill intervals affect responsiveness:

- **Short intervals (milliseconds)**: More responsive to load changes
- **Medium intervals (seconds)**: Balanced performance and accuracy
- **Long intervals (minutes)**: Coarse-grained control

## Implementation Details

### Performance Characteristics

The token bucket sampler is optimized for high-performance scenarios:

- **Zero allocations**: No memory allocation during operation
- **Atomic operations**: Lock-free implementation using atomic integers
- **Cached time**: Uses cached timestamps to avoid expensive syscalls
- **Branch prediction**: Optimized for common case (sampler disabled)

### Thread Safety

The sampler implementation is fully thread-safe:

- **Concurrent access**: Multiple goroutines can safely call `Allow()`
- **No contention**: Lock-free design prevents blocking
- **Consistency**: Atomic operations ensure consistent state

### Hot Path Impact

When sampling is disabled (default), the performance impact is minimal:

```go
// Hot path check (when sampler is nil)
if l.sampler != nil && !l.sampler.Allow(level) {
    return false
}
```

Benchmark results show negligible overhead for the unsampled case:
- Without sampler: 0.40 ns/op
- With nil sampler: 0.40 ns/op (no measurable difference)

## Usage Patterns

### Error Rate Limiting

Prevent error log storms while ensuring visibility:

```go
errorSampler := iris.NewTokenBucketSampler(
    20,               // Allow 20 immediate errors
    5,                // Then 5 errors per minute
    time.Minute,
)

logger := logger.WithOptions(iris.WithSampler(errorSampler))
```

### Debug Logging Control

Control debug output volume in production:

```go
debugSampler := iris.NewTokenBucketSampler(
    100,              // 100 debug messages burst
    50,               // 50 per second sustained
    time.Second,
)

debugLogger := baseLogger.With(iris.Str("component", "debug")).
    WithOptions(iris.WithSampler(debugSampler))
```

### High-Frequency Event Sampling

Sample high-frequency events while maintaining statistical relevance:

```go
eventSampler := iris.NewTokenBucketSampler(
    10,               // Small burst
    1,                // 1 event per 10 seconds
    10*time.Second,
)

eventLogger := baseLogger.WithOptions(iris.WithSampler(eventSampler))
```

## Configuration Priority

When both `Config.Sampler` and `WithSampler()` option are specified, the configuration follows this precedence:

1. `Config.Sampler` takes highest priority
2. `WithSampler()` option is used if `Config.Sampler` is nil
3. Cloned loggers can override parent sampler with `WithSampler()`

Example:

```go
configSampler := iris.NewTokenBucketSampler(10, 1, time.Second)
optionSampler := iris.NewTokenBucketSampler(100, 10, time.Second)

logger, _ := iris.New(iris.Config{
    Sampler: configSampler,  // This takes priority
}, iris.WithSampler(optionSampler))  // This is ignored
```

## Monitoring and Observability

To monitor sampling effectiveness, implement hooks that track dropped messages:

```go
var droppedCount int64

dropHook := func(rec *iris.Record) {
    // This hook only runs for messages that pass sampling
    // Compare with total attempts to calculate drop rate
}

logger := logger.WithOptions(iris.WithHook(dropHook))
```

## Best Practices

### Production Deployment

1. **Start Conservative**: Begin with restrictive sampling and gradually increase
2. **Monitor Impact**: Track both dropped messages and system performance
3. **Component-Specific**: Use different sampling rates for different components
4. **Error Handling**: Never sample critical error messages

### Development and Testing

1. **Disable in Tests**: Use `nil` sampler for deterministic test behavior
2. **CI/CD Considerations**: Account for sampling in log-based monitoring
3. **Local Development**: Consider higher rates for debugging

### Performance Optimization

1. **Shared Samplers**: Reuse sampler instances across related loggers
2. **Appropriate Sizing**: Size buckets based on actual traffic patterns
3. **Monitoring Overhead**: Balance sampling accuracy with performance impact

## Troubleshooting

### Common Issues

**Messages Not Appearing**
- Check if sampling is too restrictive
- Verify token bucket parameters
- Ensure sufficient refill rate

**Performance Impact**
- Confirm sampler is nil when not needed
- Check for excessive token bucket contention
- Review refill interval timing

**Inconsistent Behavior**
- Verify thread-safe usage patterns
- Check for sampler sharing between unrelated loggers
- Review configuration precedence rules

### Debugging Sampling

To debug sampling behavior, temporarily log sampler state:

```go
sampler := iris.NewTokenBucketSampler(10, 1, time.Second)

// Custom sampler wrapper for debugging
type debugSampler struct {
    *iris.TokenBucketSampler
    logger *iris.Logger
}

func (d *debugSampler) Allow(level iris.Level) bool {
    allowed := d.TokenBucketSampler.Allow(level)
    d.logger.Debug("Sampler decision",
        iris.Bool("allowed", allowed),
        iris.Int64("tokens", d.tokens.Load()))
    return allowed
}
```
---

Iris • an AGILira fragment
