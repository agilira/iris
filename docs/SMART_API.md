# IRIS Smart API Guide

## Overview

The IRIS Smart API represents a revolutionary approach to logging configuration. Instead of requiring manual setup of complex parameters, the Smart API uses intelligent auto-detection to configure optimal settings based on your environment and usage patterns.

## Philosophy: Zero Configuration, Maximum Performance

```go
// ❌ OLD WAY: Complex manual configuration
logger, err := iris.New(iris.Config{
    Level:              iris.Info,
    Output:             os.Stdout,
    Encoder:            iris.NewJSONEncoder(),
    Capacity:           65536,
    BatchSize:          32,
    Architecture:       iris.ThreadedRings,
    NumRings:           4,
    BackpressurePolicy: iris.DropOnFull,
    IdleStrategy:       iris.NewProgressiveIdleStrategy(),
})

// ✅ NEW WAY: Smart API handles everything
logger, err := iris.New(iris.Config{})
```

## Smart Detection Features

### Architecture Auto-Detection

```go
// Smart API automatically selects:
func detectOptimalArchitecture() Architecture {
    if runtime.NumCPU() >= 4 {
        return ThreadedRings // Better for production multi-core systems
    }
    return SingleRing // Simpler for single-core or development
}
```

**Benefits:**
- **Multi-core systems (≥4 CPUs)**: ThreadedRings for optimal concurrency
- **Single/dual-core systems**: SingleRing for minimal overhead
- **Automatic scaling**: No manual tuning required

### Capacity Optimization

```go
// Smart capacity based on system resources
func detectOptimalCapacity() int64 {
    cpus := int64(runtime.NumCPU())
    capacity := cpus * 8192  // 8KB per CPU core
    
    if capacity < 8192 {
        capacity = 8192      // Minimum: 8KB
    }
    if capacity > 65536 {
        capacity = 65536     // Maximum: 64KB
    }
    return capacity
}
```

**Results:**
- **Memory efficient**: No waste on single-core systems
- **Scalable**: Grows with CPU count for multi-core systems
- **Bounded**: Never exceeds reasonable limits

### Encoder Intelligence

```go
// Smart encoder selection based on context
func detectEncoderFromOptions(opts ...Option) Encoder {
    if isDevelopmentMode(opts) {
        return NewTextEncoder() // Human-readable for development
    }
    return NewJSONEncoder() // Structured for production
}
```

**Automatic Selection:**
- **Production mode**: JSON for structured logging and parsing
- **Development mode**: Text for human readability
- **Context aware**: Based on `iris.Development()` option

### Level Auto-Detection

```go
// Priority order for level detection:
// 1. Development mode → Debug level
// 2. IRIS_LEVEL environment variable
// 3. Default to Info level
func detectLevelFromOptions(opts ...Option) Level {
    if isDevelopmentMode(opts) {
        return Debug
    }
    if envLevel := os.Getenv("IRIS_LEVEL"); envLevel != "" {
        if level, err := ParseLevel(envLevel); err == nil {
            return level
        }
    }
    return Info
}
```

**Environment Variable Support:**
```bash
export IRIS_LEVEL=debug   # Sets debug level
export IRIS_LEVEL=error   # Sets error level
# Supports: debug, info, warn, error
```

### ⚡ Time Optimization

```go
// Ultra-fast cached time for performance
TimeFn: timecache.CachedTime // 121x faster than time.Now()
```

**Performance Benefits:**
- **121x faster**: Cached time vs `time.Now()`
- **Consistent timestamps**: All logs in same batch share timestamp
- **Zero allocations**: Pre-formatted time strings

## Migration Guide

### From Complex Configuration

```go
// ❌ BEFORE: Manual configuration (still works!)
logger, err := iris.New(iris.Config{
    Level:              iris.Info,
    Output:             os.Stdout,
    Encoder:            iris.NewJSONEncoder(),
    Capacity:           32768,
    BatchSize:          16,
    Architecture:       iris.ThreadedRings,
    NumRings:           4,
    BackpressurePolicy: iris.DropOnFull,
    IdleStrategy:       iris.NewProgressiveIdleStrategy(),
})

// ✅ AFTER: Smart API (same result, zero config!)
logger, err := iris.New(iris.Config{})
```

### Overriding Smart Defaults

You can still override specific settings:

```go
// Smart API with custom output
logger, err := iris.New(iris.Config{
    Output: myCustomWriter, // Override: custom output
    // Everything else: auto-detected
})

// Smart API with custom level
logger, err := iris.New(iris.Config{
    Level: iris.Error, // Override: error level only
    // Everything else: auto-detected
})
```

## Performance Comparison

### Before Smart API
```
Configuration: Manual (15-20 lines of setup)
Hot Path Allocations: 5-6 allocs/op
Encoding Performance: 800+ ns/op
Memory per Record: 10KB
User Experience: Complex, error-prone
```

### After Smart API
```
Configuration: Auto (1 line of setup)
Hot Path Allocations: 1-3 allocs/op (-67%)
Encoding Performance: 324-537 ns/op (+40-60%)
Memory per Record: 2.5KB (-75%)
User Experience: Simple, foolproof
```

## Advanced Usage

### Development Workflow

```go
func main() {
    // Development: Human-readable logs with debug level
    logger, _ := iris.New(iris.Config{}, iris.Development())
    logger.Start()
    
    logger.Debug("Starting service")
    logger.Info("Ready to serve requests")
}
```

### Production Deployment

```go
func main() {
    // Production: JSON logs with info level
    logger, _ := iris.New(iris.Config{})
    logger.Start()
    
    logger.Info("Service started", iris.Str("version", "1.0.0"))
}
```

### Container/Kubernetes Deployment

```bash
# Set environment variable for log level
export IRIS_LEVEL=warn

# Or in Kubernetes manifest:
env:
- name: IRIS_LEVEL
  value: "info"
```

```go
// Application code (same for all environments)
logger, _ := iris.New(iris.Config{}) // Reads IRIS_LEVEL automatically
```

## Benefits Summary

### For Developers
- **Instant Setup**: One line to get started
- **Zero Tuning**: No performance knobs to adjust
- **Context Aware**: Automatically adapts to dev vs prod

### For Operations
- **Optimal Performance**: Automatically tuned for your hardware
- **Environment Driven**: Control via environment variables
- **Production Ready**: Secure defaults for all scenarios

### For Organizations
- **Reduced Complexity**: Less training and documentation needed
- **Fewer Bugs**: No configuration mistakes possible
- **Better Performance**: Optimized automatically
---

Iris • an AGILira fragment