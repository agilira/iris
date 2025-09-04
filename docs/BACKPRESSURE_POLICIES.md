# Backpressure Policies Tutorial

Most users don't need to configure backpressure policies manually. The Smart API automatically selects optimal policies based on your workload. This guide is for advanced users who need specific behavior overrides.

## Quick Start (Recommended)

```go
// Most users: Smart API handles everything automatically
logger, _ := iris.New(iris.Config{})
logger.Start()

// Smart API automatically:
// - Selects optimal backpressure policy
// - Configures buffer sizes
// - Optimizes for your hardware
```

## 1. Overview

Iris provides configurable backpressure policies to handle different use cases when the ring buffer becomes full. This tutorial explains when and how to use each policy.

## 2. Available Policies

### DropOnFull (Default)
**Best for:** High-performance applications, ad servers, real-time systems, telemetry

**Behavior:** When the ring buffer is full, new log messages are immediately dropped without blocking the caller.

**Trade-offs:**
- ✅ Maximum performance (~6M+ messages/sec)
- ✅ Never blocks application threads
- ✅ Constant memory usage
- ❌ Some log messages may be lost under extreme load

**Use Cases:**
- Real-time trading systems
- Game servers
- Ad serving platforms
- High-frequency telemetry
- Performance-critical applications

### BlockOnFull
**Best for:** Audit systems, financial transactions, compliance logging, debugging

**Behavior:** When the ring buffer is full, callers block until space becomes available.

**Trade-offs:**
- ✅ Guaranteed delivery of all log messages
- ✅ No data loss
- ✅ Maintains message ordering
- ❌ May impact application performance under high load
- ❌ Potential for application slowdown

**Use Cases:**
- Financial transaction logging
- Audit trails
- Compliance systems
- Security event logging
- Critical error reporting

## 3. Configuration Examples

### 1. Programmatic Configuration

```go
package main

import (
    "os"
    "github.com/agilira/iris"
    "github.com/agilira/iris/internal/zephyroslite"
)

func main() {
    // Smart API: Optimal backpressure policy auto-selected
    logger, err := iris.New(iris.Config{})
    if err != nil {
        panic(err)
    }
    
    // Smart API automatically chooses optimal policy based on:
    // - System characteristics (CPU cores, memory)
    // - Workload patterns (detected at runtime)
    // - Output destination characteristics
    
    logger.Start()
    
    // For specific requirements, override only what you need:
    auditLogger, err := iris.New(iris.Config{
        BackpressurePolicy: zephyroslite.BlockOnFull, // Override only policy
        // Everything else: auto-detected optimally
    })
    if err != nil {
        panic(err)
    }
    
    auditLogger.Start()
    defer fastLogger.Close()
    defer auditLogger.Close()
    
    // High-frequency operations use fast logger
    for i := 0; i < 1000000; i++ {
        fastLogger.Info("High frequency event", iris.Int("iteration", i))
    }
    
    // Critical operations use audit logger
    auditLogger.Info("Financial transaction", 
        iris.Str("transaction_id", "tx-12345"),
        iris.Float64("amount", 1234.56))
}
```

### 2. Builder Pattern Configuration

```go
package main

import "github.com/agilira/iris"

func createLoggers() (*iris.Logger, *iris.Logger) {
    // Performance-optimized logger
    perfLogger := iris.NewBuilder().
        WithLevel(iris.Info).
        WithCapacity(16384).
        WithBackpressurePolicy(zephyroslite.DropOnFull).
        WithJSONEncoder().
        Build()
    
    // Compliance logger
    complianceLogger := iris.NewBuilder().
        WithLevel(iris.Debug).
        WithCapacity(2048).
        WithBackpressurePolicy(zephyroslite.BlockOnFull).
        WithTextEncoder().
        WithFileOutput("audit.log").
        Build()
    
    return perfLogger, complianceLogger
}
```

### 3. Configuration File (JSON)

```json
{
  "applications": {
    "web_server": {
      "level": "info",
      "output": "stdout",
      "encoder": "json",
      "capacity": 8192,
      "batch_size": 256,
      "backpressure_policy": "drop_on_full",
      "architecture": "single_ring"
    },
    "audit_service": {
      "level": "debug",
      "output": "/var/log/audit.log",
      "encoder": "text",
      "capacity": 2048,
      "batch_size": 64,
      "backpressure_policy": "block_on_full",
      "architecture": "single_ring"
    },
    "payment_processor": {
      "level": "info",
      "output": "/var/log/payments.log",
      "encoder": "json",
      "capacity": 4096,
      "batch_size": 128,
      "backpressure_policy": "block_on_full",
      "architecture": "single_ring"
    }
  }
}
```

## 4. Configuration Loading

### 3.1 JSON Configuration Files

Iris provides a comprehensive config loader that supports backpressure policies:

**high_performance.json:**
```json
{
  "level": "info",
  "format": "json",
  "output": "stdout",
  "capacity": 16384,
  "batch_size": 512,
  "backpressure_policy": "drop_on_full",
  "name": "high-performance-logger",
  "development": false,
  "enable_caller": false
}
```

**reliable.json:**
```json
{
  "level": "debug",
  "format": "json",
  "output": "app.log",
  "capacity": 8192,
  "batch_size": 256,
  "backpressure_policy": "block_on_full",
  "name": "reliable-logger",
  "development": true,
  "enable_caller": true
}
```

### 3.2 Loading from JSON

```go
// Load configuration from JSON file
config, err := iris.LoadConfigFromJSON("config.json")
if err != nil {
    log.Fatal(err)
}

logger, err := iris.New(*config)
if err != nil {
    log.Fatal(err)
}

logger.Info("Logger initialized", 
    iris.String("policy", config.BackpressurePolicy.String()))
```

### 3.3 Environment Variables

Configure backpressure policy via environment variables:

```bash
# Set backpressure policy
export IRIS_BACKPRESSURE_POLICY="block_on_full"  # or "drop_on_full"
export IRIS_LEVEL="debug"
export IRIS_FORMAT="json"
export IRIS_OUTPUT="stdout"
export IRIS_CAPACITY="8192"
export IRIS_BATCH_SIZE="256"
export IRIS_NAME="env-logger"
```

```go
// Load from environment
config, err := iris.LoadConfigFromEnv()
if err != nil {
    log.Fatal(err)
}

logger, err := iris.New(*config)
if err != nil {
    log.Fatal(err)
}
```

**Supported values for IRIS_BACKPRESSURE_POLICY:**
- `"drop"`, `"drop_on_full"`, `"droponful"` → DropOnFull
- `"block"`, `"block_on_full"`, `"blockonful"` → BlockOnFull
- Invalid/empty values default to DropOnFull

### 3.4 Multi-Source Configuration

Combine JSON configuration with environment variable overrides:

```go
// Load with precedence: Environment > JSON > Defaults
config, err := iris.LoadConfigMultiSource("base-config.json")
if err != nil {
    log.Fatal(err)
}

logger, err := iris.New(*config)
if err != nil {
    log.Fatal(err)
}
```

**Configuration Precedence (highest to lowest):**
1. Environment variables (IRIS_*)
2. JSON file configuration
3. Default values

### 3.5 Complete Example

```go
package main

import (
    "fmt"
    "os"
    "github.com/agilira/iris"
)

func main() {
    // Override specific settings via environment
    os.Setenv("IRIS_BACKPRESSURE_POLICY", "block_on_full")
    os.Setenv("IRIS_LEVEL", "debug")
    
    // Load configuration (JSON base + env overrides)
    config, err := iris.LoadConfigMultiSource("production.json")
    if err != nil {
        panic(err)
    }
    
    logger, err := iris.New(*config)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Logger created with policy: %s
", 
        config.BackpressurePolicy.String())
    
    logger.Info("Application started",
        iris.String("config_source", "multi-source"),
        iris.String("policy", config.BackpressurePolicy.String()))
    
    logger.Sync()
}
```

### Throughput Comparison

| Policy | Throughput | Latency | Memory | Data Loss |
|--------|------------|---------|---------|-----------|
| DropOnFull | ~6M+ msg/sec | ~15-20ns | Constant | Possible |
| BlockOnFull | Variable | Variable | Constant | None |

### Memory Usage

Both policies use constant memory (ring buffer size), but BlockOnFull may cause memory pressure in calling threads due to blocking.

### Latency Patterns

- **DropOnFull**: Consistent ultra-low latency
- **BlockOnFull**: Variable latency based on consumer speed

## 5. Performance Characteristics

### Capacity Tuning

```go
// For DropOnFull - larger buffers reduce drop probability
config := iris.Config{
    BackpressurePolicy: zephyroslite.DropOnFull,
    Capacity:           16384, // Power of 2, tune to workload
    BatchSize:          256,   // Balance between throughput and latency
}

// For BlockOnFull - smaller buffers reduce blocking time
config := iris.Config{
    BackpressurePolicy: zephyroslite.BlockOnFull,
    Capacity:           2048,  // Smaller for faster flushing
    BatchSize:          64,    // Smaller batches for lower latency
}
```

### Output Configuration

```go
// Fast outputs for DropOnFull
fastOutput := iris.WrapWriter(os.Stdout)
// or buffered file
fastOutput := iris.WrapWriter(bufio.NewWriter(file))

// Reliable outputs for BlockOnFull
reliableOutput := iris.WrapWriter(syncFile) // O_SYNC file
// or network with retries
reliableOutput := iris.NewNetworkSyncer("tcp://log-server:514")
```

### Monitoring and Alerting

```go
// Monitor drops in DropOnFull systems
stats := logger.Stats()
if drops := stats["drops"]; drops > threshold {
    // Alert: Increase capacity or optimize consumers
}

// Monitor blocking in BlockOnFull systems
if latency := measureLogLatency(); latency > threshold {
    // Alert: Optimize output or increase capacity
}
```

## 6. Best Practices

### From DropOnFull to BlockOnFull

```go
// Step 1: Increase output performance
output := iris.WrapWriter(bufio.NewWriterSize(file, 64*1024))

// Step 2: Reduce capacity for faster flushing
config.Capacity = config.Capacity / 2

// Step 3: Switch policy
config.BackpressurePolicy = zephyroslite.BlockOnFull

// Step 4: Monitor application performance
```

### From BlockOnFull to DropOnFull

```go
// Step 1: Increase capacity to reduce drops
config.Capacity = config.Capacity * 2

// Step 2: Switch policy
config.BackpressurePolicy = zephyroslite.DropOnFull

// Step 3: Implement drop monitoring
// Step 4: Add external log aggregation for redundancy
```

## 7. Troubleshooting Guide

### High Drop Rate (DropOnFull)

**Symptoms:** Missing log entries, high drop counters

**Solutions:**
1. Increase `Capacity` (power of 2)
2. Optimize output writer performance
3. Reduce log volume at source
4. Consider switching to BlockOnFull for critical logs

### Application Slowdown (BlockOnFull)

**Symptoms:** High logging latency, thread blocking

**Solutions:**
1. Optimize output performance (buffering, async writes)
2. Reduce `Capacity` for faster flushing
3. Increase `BatchSize` for better throughput
4. Consider switching to DropOnFull for non-critical logs

### Memory Issues

**Symptoms:** High memory usage, GC pressure

**Solutions:**
1. Reduce `Capacity` if too large
2. Check for log level misconfiguration
3. Verify output is not blocking indefinitely
4. Monitor goroutine counts

## 8. Troubleshooting

### Dynamic Policy Switching

```go
// Runtime policy changes require logger recreation
func switchToAuditMode(currentLogger *iris.Logger) *iris.Logger {
    currentLogger.Close()
    
    return iris.NewBuilder().
        WithLevel(iris.Debug).
        WithBackpressurePolicy(zephyroslite.BlockOnFull).
        WithFileOutput("audit-mode.log").
        Build()
}
```

### Policy-Specific Outputs

```go
// Use different outputs for different policies
type PolicyAwareLogger struct {
    fast  *iris.Logger // DropOnFull for performance
    audit *iris.Logger // BlockOnFull for guarantees
}

func (p *PolicyAwareLogger) Info(msg string, fields ...iris.Field) {
    p.fast.Info(msg, fields...)
}

func (p *PolicyAwareLogger) Audit(msg string, fields ...iris.Field) {
    p.audit.Info(msg, fields...)
}
```

## 10. Conclusion

Iris backpressure policies provide a powerful tool for balancing performance and reliability in high-throughput logging systems. The configuration system supports multiple deployment scenarios from simple hardcoded values to complex multi-source configurations.

### Key Takeaways

- **DropOnFull**: Choose for maximum performance in high-frequency scenarios
- **BlockOnFull**: Choose for guaranteed message delivery in critical systems  
- **Configuration**: Use JSON files and environment variables for flexible deployment
- **Monitoring**: Always monitor dropped message counts and buffer utilization
- **Testing**: Validate your choice under realistic load conditions

### Integration Summary

With the comprehensive configuration system, you can:
- Load policies from JSON configuration files
- Override settings via environment variables
- Combine multiple configuration sources with clear precedence
- Deploy the same application with different policies per environment
- Monitor and tune performance characteristics

This flexibility makes Iris suitable for diverse production environments, from high-performance ad servers to mission-critical financial systems.

## 9. Advanced Topics

### Custom Configuration Providers

You can extend the configuration system by implementing custom loaders:

```go
func LoadConfigFromDatabase(connectionString string) (*iris.Config, error) {
    // Custom logic to load from database
    config := &iris.Config{}
    
    // Query database for settings
    policy := queryBackpressurePolicy(connectionString)
    config.BackpressurePolicy = parseBackpressurePolicy(policy)
    
    return config, nil
}
```

### Dynamic Policy Switching

For advanced use cases, consider implementing runtime policy switching:

```go
type DynamicLogger struct {
    dropLogger  *iris.Logger
    blockLogger *iris.Logger
    useReliable bool
}

func (dl *DynamicLogger) Log(level iris.Level, msg string, fields ...iris.Field) {
    if dl.useReliable {
        dl.blockLogger.Log(level, msg, fields...)
    } else {
        dl.dropLogger.Log(level, msg, fields...)
    }
}
```

### Monitoring Integration

Integrate backpressure monitoring with your observability stack:

```go
func setupMetrics(logger *iris.Logger) {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            stats := logger.Stats()
            prometheus.GaugeVec.WithLabelValues("iris", "drops").
                Set(float64(stats["total_drops"]))
            prometheus.GaugeVec.WithLabelValues("iris", "writes").
                Set(float64(stats["total_writes"]))
        }
    }()
}

Choose your backpressure policy based on your specific requirements:

- **DropOnFull** when performance is critical and some log loss is acceptable
- **BlockOnFull** when data integrity is paramount and performance can be traded off

Both policies provide excellent performance characteristics when properly configured for your use case.

---

Iris • an AGILira fragment