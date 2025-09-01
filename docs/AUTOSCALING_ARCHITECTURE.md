# IRIS Auto-Scaling Architecture

## Overview

IRIS implements an auto-scaling logging architecture inspired by Lethe's adaptive buffer scaling patterns. The system automatically transitions between SingleRing and MPSC modes based on real-time performance metrics to optimize throughput and latency under varying workload conditions.

## Architecture Overview

### Dual Architecture System

```go
type AutoScalingLogger struct {
    // Current mode (atomic)
    mode atomic.Uint32 // SingleRingMode or MPSCMode
    
    // Logger implementations
    singleRingLogger *Logger // Ultra-fast single-threaded (~25ns/op)
    mpscLogger       *Logger // Multi-producer high-contention (~35ns/op per thread)
    
    // Performance monitoring (inspired by Lethe)
    metrics AutoScalingMetrics
    config  AutoScalingConfig
    
    // Zero-loss transition control
    transitionMu sync.RWMutex
}
```

### Scaling Modes

| Mode | Performance | Best For | Characteristics |
|------|-------------|----------|-----------------|
| **SingleRing** | ~25ns/op | Low contention, single producers | Ultra-fast, minimal overhead |
| **MPSC** | ~35ns/op per thread | High contention, multiple goroutines | Scales with concurrent producers |

## Performance Monitoring (Inspired by Lethe)

### Performance Monitoring

```go
type AutoScalingMetrics struct {
    // Write frequency tracking
    writeCount       atomic.Uint64 // Total write count
    recentWriteCount atomic.Uint64 // Writes in measurement window
    
    // Contention metrics
    contentionCount  atomic.Uint64 // Failed writes due to contention
    contentionRatio  atomic.Uint32 // Contention percentage
    
    // Latency tracking
    avgLatency       atomic.Uint64 // Average latency in nanoseconds
    recentLatency    atomic.Uint64 // Recent latency window
    
    // Goroutine monitoring
    activeGoroutines atomic.Uint32 // Current active writers
}
```

### Scaling Triggers

**Scale to MPSC when:**
- Write frequency ≥ 1000 writes/sec
- Contention ratio ≥ 10%
- Average latency ≥ 1ms
- Active goroutines ≥ 3

**Scale to SingleRing when:**
- Write frequency ≤ 100 writes/sec
- Contention ratio ≤ 1%
- Average latency ≤ 100µs

## Configuration

### Default Production Configuration

```go
config := iris.DefaultAutoScalingConfig()
// ScaleToMPSCWriteThreshold:    1000 writes/sec
// ScaleToMPSCContentionRatio:   10% contention
// ScaleToMPSCLatencyThreshold:  1ms latency
// ScaleToMPSCGoroutineCount:    3 active goroutines
// MeasurementWindow:            100ms
// ScalingCooldown:              1s
// StabilityRequirement:         3 consecutive measurements
```

### Custom Configuration

```go
scalingConfig := iris.AutoScalingConfig{
    // Aggressive scaling for high-throughput
    ScaleToMPSCWriteThreshold:    500,  // Scale earlier
    ScaleToMPSCContentionRatio:   5,    // More sensitive
    ScaleToMPSCLatencyThreshold:  500 * time.Microsecond,
    
    // Conservative scale-down
    ScaleToSingleWriteThreshold:  100,
    ScaleToSingleContentionRatio: 1,
    
    // Fast adaptation
    MeasurementWindow:    50 * time.Millisecond,
    ScalingCooldown:      500 * time.Millisecond,
    StabilityRequirement: 2,
}
```

## Usage Examples

### Basic Auto-Scaling Logger

```go
package main

import (
    "os"
    "github.com/agilira/iris"
)

func main() {
    // Create auto-scaling logger
    config := iris.Config{
        Level:   iris.Info,
        Output:  os.Stdout,
        Encoder: iris.NewJSONEncoder(),
    }
    
    autoLogger, err := iris.NewAutoScalingLogger(
        config, 
        iris.DefaultAutoScalingConfig(),
    )
    if err != nil {
        panic(err)
    }
    
    // Start auto-scaling system
    autoLogger.Start()
    defer autoLogger.Close()
    
    // Use normally - auto-scaling is transparent
    autoLogger.Info("Message will auto-scale based on load")
}
```

### Multi-Goroutine High-Contention Example

```go
func highContentionLogging(autoLogger *iris.AutoScalingLogger) {
    var wg sync.WaitGroup
    
    // Launch multiple goroutines (will trigger MPSC scaling)
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 1000; j++ {
                autoLogger.Info("High contention message",
                    iris.Int("goroutine", id),
                    iris.Int("message", j),
                )
            }
        }(i)
    }
    
    wg.Wait()
    
    // Check scaling results
    stats := autoLogger.GetScalingStats()
    fmt.Printf("Mode: %s, Scale operations: %d\n", 
        stats.CurrentMode, stats.TotalScaleOperations)
}
```

## Performance Characteristics

### Automatic Optimization

1. **Low Load**: SingleRing mode (25ns/op) - optimal for single producers
2. **High Load**: MPSC mode (35ns/op per thread) - scales with goroutines
3. **Zero Loss**: Atomic transitions with no log message loss
4. **Adaptive**: Real-time performance monitoring and adjustment

### Benchmarking Results

```
BenchmarkAutoScaling/SingleProducer    50000000    25.41 ns/op    0 B/op    0 allocs/op
BenchmarkAutoScaling/MultiProducer     30000000    35.67 ns/op    0 B/op    0 allocs/op
BenchmarkAutoScaling/ScaleTransition         100    15430 ns/op   0 B/op    0 allocs/op
```

## Monitoring and Statistics

### Real-Time Statistics

```go
stats := autoLogger.GetScalingStats()

type AutoScalingStats struct {
    CurrentMode          AutoScalingMode // Current scaling mode
    TotalScaleOperations uint64         // Total scaling operations
    ScaleToMPSCCount     uint64         // Scale to MPSC count
    ScaleToSingleCount   uint64         // Scale to Single count
    TotalWrites          uint64         // Total log writes
    ContentionCount      uint64         // Contention events
    ActiveGoroutines     uint32         // Current active goroutines
}
```

### Monitoring Example

```go
// Monitor scaling in real-time
go func() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := autoLogger.GetScalingStats()
        fmt.Printf("Mode: %s, Writes: %d, Goroutines: %d\n",
            stats.CurrentMode, stats.TotalWrites, stats.ActiveGoroutines)
    }
}()
```

## Technical Implementation

### Adaptive Scaling Algorithm

The auto-scaling architecture adapts Lethe's `tryAdaptiveResize` and `shouldScaleToMPSC` patterns for logging architecture transitions:

1. **Lethe Pattern**: Adaptive buffer resizing based on contention metrics
2. **Iris Adaptation**: Adaptive architecture switching based on workload patterns
3. **Zero Loss**: Atomic transitions using `sync.RWMutex` for consistency
4. **Performance Metrics**: Real-time monitoring similar to Lethe's latency tracking

### Scaling Decision Logic

```go
// Scaling decision based on Lethe's shouldScaleToMPSC logic
func (asl *AutoScalingLogger) determinePreferredMode(metrics scalingMetrics) AutoScalingMode {
    // Scale up conditions
    if metrics.writesPerSecond >= asl.config.ScaleToMPSCWriteThreshold ||
       metrics.contentionRatio >= asl.config.ScaleToMPSCContentionRatio ||
       metrics.avgLatency >= asl.config.ScaleToMPSCLatencyThreshold ||
       metrics.activeGoroutines >= asl.config.ScaleToMPSCGoroutineCount {
        return MPSCMode // Scale up similar to Lethe's buffer scaling
    }
    
    // Scale down conditions
    if metrics.writesPerSecond <= asl.config.ScaleToSingleWriteThreshold &&
       metrics.contentionRatio <= asl.config.ScaleToSingleContentionRatio &&
       metrics.avgLatency <= asl.config.ScaleToSingleLatencyMax {
        return SingleRingMode
    }
    
    return AutoScalingMode(asl.mode.Load()) // Maintain current mode
}
```

### Atomic Transitions

```go
func (asl *AutoScalingLogger) performScaling(targetMode AutoScalingMode) {
    // Lock for exclusive access during transition
    asl.transitionMu.Lock()
    defer asl.transitionMu.Unlock()
    
    // Atomic mode switch
    asl.mode.Store(uint32(targetMode))
    
    // No log loss - both loggers remain active
    // Only routing changes atomically
}
```

## Architecture Characteristics

### Key Features

1. **Auto-Scaling Architecture**: Dynamic transitions between logging architectures based on workload
2. **Lethe-Inspired Metrics**: Adapts proven adaptive scaling patterns for logging systems
3. **Zero Loss Transitions**: Guaranteed message delivery during scaling operations
4. **Real-Time Adaptation**: Responds to workload changes with minimal latency
5. **Performance Optimized**: Automatically selects optimal architecture for current load patterns

### Benefits

- **Adaptive Performance**: Transitions from static logging architectures to dynamic systems
- **Optimal Throughput**: Performance optimization across varying workload patterns
- **Operational Simplicity**: Self-tuning systems requiring minimal manual configuration
- **Scalability**: Automatic adaptation to application growth and load changes

## Future Development

### Planned Features

1. **Machine Learning Integration**: Predictive scaling based on historical workload patterns
2. **Multi-Tier Scaling**: Additional intermediate modes for fine-grained performance optimization
3. **Cross-Application Learning**: Shared scaling intelligence across multiple deployments
4. **Hardware-Aware Scaling**: CPU core count and NUMA topology-aware optimization

The auto-scaling architecture provides a foundation for advanced adaptive logging systems that optimize performance automatically based on runtime conditions.
