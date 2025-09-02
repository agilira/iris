# Iris Idle Strategies Implementation

## Overview

Following the wise counsel of the ancient Gemini, we have implemented a comprehensive system of configurable **Idle Strategies** for the Iris logging library. This addresses the critical issue of CPU consumption when the consumer loop has no work to process.

## The Problem

The original implementation used a hardcoded spinning strategy that consumed ~100% of one CPU core even when no logging was occurring. This was problematic for:

- Microservices running on resource-limited clusters
- Production environments where CPU efficiency matters
- Applications with variable logging workloads

## The Solution: Configurable Idle Strategies

We've implemented 5 different idle strategies, each providing different trade-offs between latency and CPU usage:

### 1. SpinningIdleStrategy
- **Latency**: Minimum possible (~nanoseconds)
- **CPU Usage**: ~100% of one core
- **Best for**: Ultra-low latency requirements where CPU is not a concern
- **Use case**: High-frequency trading, real-time systems

```go
config := &iris.Config{
    IdleStrategy: iris.NewSpinningIdleStrategy(),
    // ... other config
}
```

### 2. SleepingIdleStrategy
- **Latency**: ~1-10ms (configurable)
- **CPU Usage**: ~1-10% (configurable)
- **Best for**: Balanced CPU usage and latency in production
- **Parameters**: `sleepDuration`, `maxSpins` (spin before sleeping)

```go
// Low CPU, higher latency
iris.NewSleepingIdleStrategy(5*time.Millisecond, 0)

// Hybrid: spin briefly then sleep
iris.NewSleepingIdleStrategy(time.Millisecond, 1000)
```

### 3. YieldingIdleStrategy
- **Latency**: ~microseconds to low milliseconds
- **CPU Usage**: ~10-50% (depending on configuration)
- **Best for**: Moderate CPU reduction with reasonable latency
- **Parameters**: `maxSpins` (spins before yielding to scheduler)

```go
// More aggressive yielding
iris.NewYieldingIdleStrategy(100)

// Conservative yielding
iris.NewYieldingIdleStrategy(10000)
```

### 4. ChannelIdleStrategy
- **Latency**: ~microseconds (channel wake-up time)
- **CPU Usage**: Near 0% when idle
- **Best for**: Minimum CPU usage, low-throughput scenarios
- **Parameters**: `timeout` (max wait before checking shutdown)

```go
// No timeout - maximum efficiency
iris.NewChannelIdleStrategy(0)

// With timeout for responsive shutdown
iris.NewChannelIdleStrategy(100*time.Millisecond)
```

### 5. ProgressiveIdleStrategy (Default)
- **Latency**: Adaptive (starts minimum, increases when idle)
- **CPU Usage**: Adaptive (starts high, reduces over time)
- **Best for**: Variable workloads, general use
- **Behavior**: 
  - Hot spin for 1000 iterations (minimum latency)
  - Occasional yielding up to 10000 iterations
  - Progressive sleep with exponential backoff
  - Resets to hot spin when work is found

```go
iris.NewProgressiveIdleStrategy() // Used by default
```

## Predefined Strategies

For convenience, we provide predefined strategies:

```go
// Ultra-low latency, maximum CPU
config.IdleStrategy = iris.SpinningStrategy

// Good all-around performance (default)
config.IdleStrategy = iris.BalancedStrategy  

// Minimize CPU usage
config.IdleStrategy = iris.EfficientStrategy

// Hybrid approach
config.IdleStrategy = iris.HybridStrategy
```

## Implementation Architecture

### Core Interface
```go
type IdleStrategy interface {
    Idle() bool    // Called when no work available
    Reset()        // Called when work is found
    String() string // Human-readable name
}
```

### Integration Points

1. **ZephyrosLight**: Updated to use configurable idle strategy instead of hardcoded spinning
2. **Config**: Added `IdleStrategy` field with intelligent default
3. **Builder Pattern**: Added `WithIdleStrategy()` method
4. **Ring Buffer**: Updated to pass idle strategy to underlying engine

### Backward Compatibility

- **Default Behavior**: Uses `ProgressiveIdleStrategy` (similar to original but more efficient)
- **API**: All existing code continues to work without changes
- **Performance**: Default performance is equal or better than original

## Performance Characteristics

| Strategy | Latency | CPU Usage | Best Use Case |
|----------|---------|-----------|---------------|
| Spinning | ~ns | ~100% | Ultra-low latency |
| Sleeping | ~1-10ms | ~1-10% | Production balance |
| Yielding | ~μs-ms | ~10-50% | Moderate efficiency |
| Channel | ~μs | ~0% | Low throughput |
| Progressive | Adaptive | Adaptive | General use (default) |

## Usage Examples

### Basic Configuration
```go
config := &iris.Config{
    Output:       iris.WrapWriter(os.Stdout),
    Encoder:      iris.NewJSONEncoder(),
    IdleStrategy: iris.NewSleepingIdleStrategy(time.Millisecond, 100),
}

logger, err := iris.New(*config)
```

### High-Performance Setup
```go
config := &iris.Config{
    IdleStrategy: iris.SpinningStrategy,  // Ultra-low latency
    Capacity:     8192,                   // Large buffer
    BatchSize:    256,                    // High throughput
}
```

### Resource-Efficient Setup
```go
config := &iris.Config{
    IdleStrategy: iris.EfficientStrategy, // Low CPU usage
    Capacity:     1024,                   // Moderate buffer
    BatchSize:    32,                     // Balanced batching
}
```

## Testing and Validation

The implementation includes comprehensive tests:

- **Basic functionality** for all strategies
- **Parameter validation** for configurable strategies
- **Integration tests** with the ring buffer
- **Performance characteristics** verification
- **Default behavior** validation

## Configuration Guide

### Strategy Selection
Choose the appropriate strategy based on your requirements:

- **Need minimum latency**: Use `SpinningStrategy`
- **Need CPU efficiency**: Use `EfficientStrategy` 
- **Need balance**: Use `BalancedStrategy` (default)
- **Variable workload**: Use `ProgressiveIdleStrategy`

## Technical Implementation Details

### ZephyrosLight Integration
The `LoopProcess()` method was updated to use the configurable strategy:

```go
func (z *ZephyrosLight[T]) LoopProcess() {
    for z.closed.Load() == 0 {
        processed := z.ProcessBatch()
        
        if processed > 0 {
            z.idleStrategy.Reset() // Work found
        } else {
            z.idleStrategy.Idle()  // Use configured strategy
        }
    }
}
```

### Memory Layout
Idle strategies are lightweight and don't impact the zero-allocation write path. The strategy interface adds minimal overhead to the consumer loop.

### Thread Safety
All idle strategies are designed for single-consumer use (as per Iris architecture) and don't require additional synchronization.

## Conclusion

This implementation successfully addresses Gemini's concern about CPU consumption while maintaining the ultra-high performance characteristics of Iris. Users can now choose the appropriate trade-off between latency and CPU usage based on their specific requirements.

The default `ProgressiveIdleStrategy` provides excellent performance for most use cases without requiring manual tuning, while advanced users can fine-tune their configuration for specific requirements.
