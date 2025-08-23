# IRIS Examples

This directory contains comprehensive examples demonstrating IRIS logging library features.

## Examples Overview

### Context Integration (`examples/context/`)

Demonstrates `context.Context` integration with automatic field extraction and performance optimization.

**Run the example:**
```bash
cd examples/context
go run main.go
```

**Features demonstrated:**
- Basic context field extraction
- HTTP middleware pattern
- Custom context extractors
- Performance comparison between manual and optimized approaches

### Configuration Loading (`examples/config/`)

Shows how to load configuration from JSON files, environment variables, and multi-source setups.

**Run the example:**
```bash
cd examples/config
go run main.go
```

**Features demonstrated:**
- JSON configuration loading
- Environment variable configuration
- Multi-source configuration with precedence
- Production deployment patterns

### Configuration Files (`examples/configs/`)

Sample configuration files for different environments:

- `development.json` - Debug-friendly configuration
- `staging.json` - Staging environment setup
- `production.json` - Production-optimized configuration  
- `microservice.json` - Microservice deployment config

## Running Examples

### Prerequisites

Make sure you have Go 1.18+ installed and IRIS module available:

```bash
# From the iris root directory
go mod tidy
```

### Context Integration Example

```bash
cd examples/context
go run main.go
```

This example will show:
1. Basic context extraction with request ID, user ID, and session ID
2. HTTP middleware pattern for web applications
3. Custom context extractors with field renaming
4. Performance comparison showing optimization benefits

### Configuration Loading Example

```bash
cd examples/config  
go run main.go
```

This example demonstrates:
1. Loading configuration from JSON files
2. Loading configuration from environment variables
3. Multi-source loading with precedence rules
4. Production deployment patterns with fallbacks

### Using Configuration Files

You can test configuration loading with the provided config files:

```bash
# Test with development config
IRIS_CONFIG=../configs/development.json go run examples/config/main.go

# Test with production config  
IRIS_CONFIG=../configs/production.json go run examples/config/main.go

# Test with environment variables
IRIS_LEVEL=debug IRIS_CAPACITY=4096 go run examples/config/main.go
```

## Integration Examples

### Kubernetes Deployment

Example Kubernetes deployment using ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iris-config
data:
  config.json: |
    {
      "level": "info",
      "format": "json",
      "output": "stdout",
      "capacity": 32768,
      "batch_size": 64,
      "enable_caller": false
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iris-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: myapp:latest
        env:
        - name: IRIS_LEVEL
          value: "info"
        volumeMounts:
        - name: config
          mountPath: /etc/iris
      volumes:
      - name: config
        configMap:
          name: iris-config
```

### Docker Compose

Example Docker Compose setup:

```yaml
version: '3.8'
services:
  app:
    image: myapp:latest
    environment:
      - IRIS_LEVEL=info
      - IRIS_FORMAT=json
      - IRIS_OUTPUT=stdout
      - IRIS_CAPACITY=32768
    volumes:
      - ./configs/production.json:/app/config.json:ro
```

### HTTP Service with Context

```go
package main

import (
    "context"
    "net/http"
    "github.com/agilira/iris"
)

func contextMiddleware(logger *iris.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract request ID
            requestID := r.Header.Get("X-Request-ID")
            if requestID == "" {
                requestID = generateRequestID()
            }
            
            // Create context logger
            ctx := context.WithValue(r.Context(), iris.RequestIDKey, requestID)
            contextLogger := logger.WithRequestID(ctx)
            
            // Store in context
            ctx = context.WithValue(ctx, "logger", contextLogger)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func handler(w http.ResponseWriter, r *http.Request) {
    logger := r.Context().Value("logger").(*iris.ContextLogger)
    logger.Info("Request handled")
}
```

## Performance Tips

### Context Integration

1. **Extract Once, Use Many Times**: Pre-extract context fields rather than calling `context.Value()` repeatedly
2. **Use Fast Methods**: For single fields, use `WithRequestID()`, `WithUserID()`, etc.
3. **Limit Extraction Scope**: Configure only needed keys in custom extractors

### Configuration Loading

1. **Load at Startup**: Configuration loading has ~14μs overhead, do it once at startup
2. **Cache Configuration**: Store loaded config and reuse for multiple logger instances
3. **Environment Override**: Use environment variables for container/cloud deployments

## Benchmarking

Run benchmarks to see performance characteristics:

```bash
# Context integration benchmarks
go test -bench=BenchmarkContext -benchmem

# Configuration loading benchmarks  
go test -bench=BenchmarkLoadConfig -benchmem

# Overall IRIS benchmarks
go test -bench=. -benchmem
```

## Troubleshooting

### Context Values Not Appearing

1. Check that context keys match the extractor configuration
2. Verify values are strings (other types are skipped)
3. Ensure context chain contains the values

### Configuration Loading Issues

1. Verify file permissions and paths
2. Check JSON syntax with a validator
3. Ensure environment variable names are correct (`IRIS_*`)

### Performance Issues

1. Avoid repeated context extraction
2. Use appropriate buffer sizes (Capacity, BatchSize)
3. Monitor memory usage with `-benchmem`

For more details, see the full documentation in the `docs/` directory.
