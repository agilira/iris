# Grafana Loki Integration Guide

This guide demonstrates how to integrate Iris with Grafana Loki using the external `iris-writer-loki` module.

## Overview

The Loki integration is provided through an external module that implements the `SyncWriter` interface. This modular approach keeps the core Iris library dependency-free while providing high-performance Loki integration with zero-allocation batched writes.

### Key Features

- **External Module**: Zero impact on core Iris library
- **Zero Allocations**: Write operations don't allocate memory in hot path
- **Async Batching**: Logs are batched and sent asynchronously to Loki
- **High Performance**: Uses timecache for optimized timestamp generation
- **Connection Pooling**: Reuses HTTP connections for optimal performance
- **Graceful Degradation**: Continues working even if Loki is unavailable
- **Configurable Backpressure**: Choose between dropping logs or blocking on errors
- **Automatic Retries**: Built-in retry logic with exponential backoff
- **Multi-tenant Support**: Full support for Loki's multi-tenancy features

## Installation

The Loki writer is available as a separate module:

```bash
go get github.com/agilira/iris-writer-loki
```

## Basic Usage

```go
package main

import (
    "log"
    "time"
    
    "github.com/agilira/iris"
    lokiwriter "github.com/agilira/iris-writer-loki"
)

func main() {
    // Configure Loki writer
    config := lokiwriter.Config{
        Endpoint:      "http://localhost:3100/loki/api/v1/push",
        TenantID:      "my-tenant",
        Labels:        map[string]string{"service": "my-app"},
        BatchSize:     1000,
        FlushInterval: time.Second,
        Timeout:       10 * time.Second,
        OnError: func(err error) {
            log.Printf("Loki writer error: %v", err)
        },
    }
    
    writer, err := lokiwriter.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer writer.Close()
    
    // Use with Iris logger
    logger := iris.New(iris.WithSyncWriter(writer))
    
    logger.Info("Application started",
        iris.String("version", "1.0.0"),
        iris.String("build", "abc123"))
}
```

## Configuration Options

The `lokiwriter.Config` struct provides comprehensive configuration options:

```go
type Config struct {
    Endpoint      string                 // Loki push endpoint URL
    TenantID      string                 // Optional tenant ID for multi-tenant setups
    Labels        map[string]string      // Static labels to attach to all log streams
    BatchSize     int                    // Number of records to batch (default: 1000)
    FlushInterval time.Duration          // Maximum time to wait before flushing (default: 1s)
    Timeout       time.Duration          // HTTP request timeout (default: 10s)
    OnError       func(error)           // Optional error callback function
    MaxRetries    int                   // Number of retry attempts (default: 3)
    RetryDelay    time.Duration         // Delay between retries (default: 100ms)
}
```

## Performance Configuration

### High-Throughput Scenarios

For applications with very high log volume (>10k logs/second):

```go
config := lokiwriter.Config{
    Endpoint:      "http://loki:3100/loki/api/v1/push",
    BatchSize:     5000,             // Large batches for efficiency
    FlushInterval: 10 * time.Second, // Less frequent flushes
    Timeout:       30 * time.Second, // Longer timeout for large batches
    Labels: map[string]string{
        "service": "high-volume-app",
        "env":     "production",
    },
}
```

### Low-Latency Scenarios

For applications where log delivery latency matters:

```go
config := lokiwriter.Config{
    Endpoint:      "http://loki:3100/loki/api/v1/push",
    BatchSize:     100,               // Smaller batches
    FlushInterval: 500 * time.Millisecond, // Frequent flushes
    Timeout:       5 * time.Second,   // Faster timeouts
    Labels: map[string]string{
        "service": "low-latency-app",
        "env":     "production",
    },
}
```

## Multi-Output Setup

Send logs to both console and Loki using multiple writers:

```go
import (
    "os"
    "github.com/agilira/iris"
    lokiwriter "github.com/agilira/iris-writer-loki"
)

func main() {
    // Create Loki writer
    lokiWriter, err := lokiwriter.New(lokiwriter.Config{
        Endpoint: "http://localhost:3100/loki/api/v1/push",
        Labels:   map[string]string{"service": "my-app"},
    })
    if err != nil {
        panic(err)
    }
    defer lokiWriter.Close()

    // Create console writer
    consoleWriter := os.Stdout

    // Use with Iris - note: this would require implementing multiple writers
    // or using a custom solution based on your needs
    logger := iris.New()
    
    // Log to both destinations
    logger.Info("Message goes to both console and Loki")
}
```

## Multi-Tenant Configuration

For multi-tenant Loki deployments:

```go
config := lokiwriter.Config{
    Endpoint: "http://loki:3100/loki/api/v1/push",
    TenantID: "customer-123", // Sets X-Scope-OrgID header
    Labels: map[string]string{
        "service":  "multi-tenant-app",
        "customer": "customer-123",
        "env":      "production",
    },
}
```

## Error Handling and Monitoring

### Error Callbacks

```go
config := lokiwriter.Config{
    Endpoint: "http://loki:3100/loki/api/v1/push",
    OnError: func(err error) {
        // Send to metrics system, alert, etc.
        // This callback should be fast and non-blocking
        log.Printf("Loki writer error: %v", err)
        
        // You could implement metrics tracking here
        // metrics.Counter("loki_errors").Inc()
        
        if strings.Contains(err.Error(), "timeout") {
            // Handle timeout errors specifically
            log.Println("Loki timeout detected - check network connectivity")
        }
    },
    Labels: map[string]string{
        "service": "monitored-app",
    },
}
```

### Performance Monitoring

The writer maintains internal metrics accessible through atomic operations:

```go
// Note: These would be internal fields in the actual implementation
// This is conceptual - actual implementation may vary
writer, err := lokiwriter.New(config)
if err != nil {
    panic(err)
}

// In the actual implementation, you would access these through 
// public methods or fields designed for monitoring
```

## Production Deployment

### Kubernetes Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  LOKI_ENDPOINT: "http://loki.logging.svc.cluster.local:3100/loki/api/v1/push"
  LOKI_BATCH_SIZE: "2000"
  LOKI_FLUSH_INTERVAL: "5s"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: my-app:latest
        env:
        - name: LOKI_ENDPOINT
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: LOKI_ENDPOINT
        - name: LOKI_BATCH_SIZE
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: LOKI_BATCH_SIZE
        - name: LOKI_FLUSH_INTERVAL
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: LOKI_FLUSH_INTERVAL
```

### Application Configuration

```go
import (
    "os"
    "strconv"
    "time"
)

func createLokiWriter() (*lokiwriter.Writer, error) {
    batchSize, _ := strconv.Atoi(os.Getenv("LOKI_BATCH_SIZE"))
    if batchSize == 0 {
        batchSize = 1000
    }
    
    flushInterval, _ := time.ParseDuration(os.Getenv("LOKI_FLUSH_INTERVAL"))
    if flushInterval == 0 {
        flushInterval = time.Second
    }
    
    config := lokiwriter.Config{
        Endpoint:      os.Getenv("LOKI_ENDPOINT"),
        BatchSize:     batchSize,
        FlushInterval: flushInterval,
        Labels: map[string]string{
            "service": os.Getenv("SERVICE_NAME"),
            "env":     os.Getenv("ENVIRONMENT"),
            "version": os.Getenv("APP_VERSION"),
        },
    }
    
    return lokiwriter.New(config)
}
```

### Docker Compose

```yaml
version: '3.8'
services:
  app:
    image: my-app:latest
    environment:
      - LOKI_ENDPOINT=http://loki:3100/loki/api/v1/push
      - LOKI_BATCH_SIZE=1000
      - LOKI_FLUSH_INTERVAL=5s
    depends_on:
      - loki

  loki:
    image: grafana/loki:latest
    ports:
      - "3100:3100"
    volumes:
      - ./loki-config.yaml:/etc/loki/local-config.yaml
    command: -config.file=/etc/loki/local-config.yaml
```

## Performance Characteristics

### Benchmarks

The external Loki writer maintains excellent performance characteristics while providing full Loki integration:

```
BenchmarkLokiWriter_WriteRecord-8    50000000    35.2 ns/op    0 allocs/op
BenchmarkLokiWriter_Handle-8         30000000    42.1 ns/op    0 allocs/op
```

The overhead is minimal (~35ns per write operation) while providing full batching and async processing.

### Memory Usage

- **Zero allocations** in the write hot path
- **Fixed memory usage** regardless of log volume  
- **Buffer pooling** for network operations using sync.Pool
- **Configurable batch sizes** for memory/latency trade-offs
- **Independent from core Iris**: No impact on core library memory usage

### Network Efficiency

- **Batched requests** reduce network overhead significantly
- **HTTP connection reuse** for optimal performance
- **Intelligent batching** based on size and time thresholds
- **Configurable timeouts** and retries for reliability
- **Async processing** to avoid blocking the hot logging path

## Troubleshooting

### Common Issues

**Logs not appearing in Loki:**
- Check Loki endpoint configuration and network connectivity
- Verify Loki is accepting logs at the specified endpoint
- Check that labels conform to Prometheus label naming rules
- Ensure the X-Scope-OrgID header is set correctly for multi-tenant setups

**High memory usage:**
- Reduce BatchSize if memory is constrained
- Check for Loki connectivity issues causing batches to accumulate
- Monitor error callback frequency - frequent errors may indicate issues

**Performance issues:**
- Increase BatchSize for higher throughput scenarios
- Reduce FlushInterval for lower latency requirements
- Check network latency between application and Loki
- Ensure Loki can handle the incoming log volume

### Debug Configuration

Enable detailed error reporting for troubleshooting:

```go
config := lokiwriter.Config{
    Endpoint: "http://loki:3100/loki/api/v1/push",
    OnError: func(err error) {
        // Detailed logging for debugging
        log.Printf("DEBUG: Loki writer error: %v", err)
        log.Printf("DEBUG: Error type: %T", err)
        
        // Check for specific error types
        if strings.Contains(err.Error(), "connection refused") {
            log.Println("DEBUG: Loki appears to be unreachable")
        }
        if strings.Contains(err.Error(), "400") {
            log.Println("DEBUG: Loki rejected the request - check log format")
        }
    },
    // ... other config
}
```

## Security Considerations

### Authentication

For production deployments with authentication, you can extend the writer or use a reverse proxy:

```go
// Option 1: Use reverse proxy with authentication
config := lokiwriter.Config{
    Endpoint: "http://authenticated-proxy:8080/loki/api/v1/push",
    // ... other config
}

// Option 2: In future versions, custom HTTP client support could be added
// This would allow direct authentication configuration
```

### Network Security

- Use HTTPS endpoints in production (`https://loki.example.com/loki/api/v1/push`)
- Configure proper TLS certificates for Loki
- Use network policies to restrict Loki access in Kubernetes
- Consider using service mesh for automatic mTLS between services
- Implement proper firewall rules to limit Loki endpoint access

## Best Practices

1. **Choose appropriate batch sizes**: 1000-5000 for high throughput, 100-500 for low latency
2. **Monitor error callbacks**: Always implement OnError to track writer health
3. **Use meaningful labels**: Keep labels cardinality low but meaningful for Loki performance
4. **Test failover scenarios**: Ensure your application continues working if Loki is unavailable
5. **Tune for your workload**: Benchmark with your actual log patterns and volume
6. **Graceful shutdown**: Always call `writer.Close()` to flush pending logs
7. **Resource management**: Monitor memory usage and adjust batch sizes accordingly

## Migration from Core Integration

If you're migrating from a previous version that had Loki integration in the core:

### Before (Core Integration)
```go
// Old approach - hypothetical core integration
logger, err := iris.New(iris.Config{
    Output: iris.NewLokiWriter(lokiConfig),
})
```

### After (External Module)
```go
// New approach - external module
import lokiwriter "github.com/agilira/iris-writer-loki"

writer, err := lokiwriter.New(lokiwriter.Config{
    Endpoint: "http://loki:3100/loki/api/v1/push",
    // ... config
})
if err != nil {
    panic(err)
}
defer writer.Close()

logger := iris.New(iris.WithSyncWriter(writer))
```

## Integration Examples

For complete production-ready examples, see the iris-writer-loki repository:

- Basic integration example
- High-throughput configuration
- Multi-tenant setup
- Kubernetes deployment patterns
- Error handling and monitoring
- Graceful shutdown patterns

---

**Related Documentation:**
- [Writer Development Guide](./WRITER_DEVELOPMENT.md) - Creating custom writers
- [SyncWriter Interface](./SYNCWRITER_INTERFACE.md) - Writer interface details  
- [Provider Ecosystem](./PROVIDER_ECOSYSTEM.md) - Modular architecture overview

---

Iris • an AGILira fragment