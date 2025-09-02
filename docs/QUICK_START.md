# Iris Quick Start Guide

## Getting Started in 30 Seconds

Iris features a Smart API that automatically configures itself for optimal performance. No complex setup, no manual tuning - just ultra-fast logging that works perfectly out-of-the-box.

## Installation

```bash
go get github.com/agilira/iris
```

## Core Concept: Smart Auto-Configuration

**Key Principle**: Iris intelligently detects and configures everything automatically. You **never** need to manually configure architectures, buffer sizes, encoders, or performance settings.

```go
// ✅ RIGHT: Smart API handles everything automatically
logger, _ := iris.New(iris.Config{})
logger.Start()

// Iris automatically:
// - Detects optimal architecture (Single vs Multi-threaded)
// - Calculates optimal capacity (8KB per CPU core)
// - Selects best encoder (JSON for production)
// - Sets appropriate level (Info default)
// - Enables time caching for performance
```

## Quick Start Examples

### 1. Simplest Possible Setup (Production Ready)

```go
package main

import "github.com/agilira/iris"

func main() {
    // One line setup - production ready!
    logger, _ := iris.New(iris.Config{})
    defer logger.Close()
    logger.Start()
    
    // ⚡ Ultra-fast structured logging
    logger.Info("Service started", iris.Str("version", "1.0.0"))
}
```

### 2. Development Mode (Human Readable)

```go
// Development mode: auto-selects TextEncoder
logger, _ := iris.New(iris.Config{}, iris.Development())
logger.Start()

logger.Info("Debug info", iris.Str("component", "auth"))
// Output: time=2025-09-02T12:00:00Z level=info msg="Debug info" component="auth"
```

### 3. Environment-Aware Level Setting

```go
// Environment variable support: IRIS_LEVEL=debug
logger, _ := iris.New(iris.Config{})
logger.Start()

// Automatically reads IRIS_LEVEL environment variable
// Supports: debug, info, warn, error
```

### 4. Web Application Example

```go
package main

import (
    "net/http"
    "github.com/agilira/iris"
)

func main() {
    // Production-ready web app logging
    logger, _ := iris.New(iris.Config{})
    defer logger.Close()
    logger.Start()
    
    http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        logger.Info("API request",
            iris.Str("method", r.Method),
            iris.Str("path", r.URL.Path),
            iris.Str("user_agent", r.UserAgent()))
    })
    
    logger.Info("Server starting", iris.Int("port", 8080))
    http.ListenAndServe(":8080", nil)
}
```

## Smart Auto-Detection Features

### Architecture Detection
```go
// Automatically selects optimal architecture:
// - SingleRing: For systems with < 4 CPU cores
// - ThreadedRings: For systems with >= 4 CPU cores
logger, _ := iris.New(iris.Config{}) // Smart choice made automatically
```

### Capacity Optimization
```go
// Auto-calculates optimal buffer capacity:
// - 8KB per CPU core
// - Minimum: 8KB
// - Maximum: 64KB
// No manual tuning required!
```

### Encoder Selection
```go
// Smart encoder selection:

// Production (default): JSON for structured logging
logger, _ := iris.New(iris.Config{})
// Output: {"ts":"2025-09-02T12:00:00Z","level":"info","msg":"Hello"}

// Development: Human-readable text
logger, _ := iris.New(iris.Config{}, iris.Development())
// Output: time=2025-09-02T12:00:00Z level=info msg="Hello"
```
    
    // Zero-allocation structured logging
    logger.Info("Application started",
        iris.Str("version", "1.0.0"),
        iris.Int("port", 8080),
    )
    
    logger.Debug("Debug information",
        iris.Str("component", "database"),
        iris.Bool("connected", true),
    )
}
```

### 2. Auto-Scaling Logger (Production Recommended)

```go
package main

import (
    "os"
    "sync"
    "github.com/agilira/iris"
)

func main() {
    //PRODUCTION PATTERN: Auto-scaling logger
    autoLogger, err := iris.NewAutoScalingLogger(
        iris.Config{
            Level:  iris.Info,
            Output: os.Stdout,
            Encoder: iris.NewJSONEncoder(), // Production-ready JSON
        },
        iris.DefaultAutoScalingConfig(), // 🤖 Intelligent defaults
    )
    if err != nil {
        panic(err)
    }
    defer autoLogger.Close()
    
    // Start the auto-scaling system
    autoLogger.Start()
    
    // Simulate different workload patterns
    // The logger automatically adapts to each pattern
    
    // Low contention scenario (uses SingleRing ~25ns/op)
    autoLogger.Info("Low load operation", iris.Str("type", "single"))
    
    // High contention scenario (automatically switches to MPSC ~35ns/op)
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                autoLogger.Info("High load operation",
                    iris.Int("goroutine", id),
                    iris.Int("operation", j),
                    iris.Str("type", "concurrent"),
                )
            }
        }(i)
    }
    wg.Wait()
    
    // Returns to low contention (automatically switches back to SingleRing)
    autoLogger.Info("Workload completed", iris.Str("status", "success"))
}
```

## How Auto-Scaling Works

### Automatic Mode Detection

Iris continuously monitors your application and automatically chooses the optimal logging mode:

| Workload Pattern | Auto-Selected Mode | Performance | When Used |
|------------------|-------------------|-------------|-----------|
| **Low Contention** | SingleRing | ~25ns/op | Single goroutine, low frequency |
| **High Contention** | MPSC | ~35ns/op per thread | Multiple goroutines, high frequency |

### Scaling Triggers (All Automatic)

The system automatically scales based on:

- **Write Frequency**: >1000 writes/sec → MPSC
- **ontention**: >10% contention ratio → MPSC  
- **Latency**: >1ms average latency → MPSC
- **Goroutines**: >3 concurrent writers → MPSC

### Zero Configuration Required

```go
// This is ALL you need for production:
autoLogger, _ := iris.NewAutoScalingLogger(config, iris.DefaultAutoScalingConfig())
autoLogger.Start()

// IRIS handles:
// ✅ Performance monitoring
// ✅ Mode switching decisions  
// ✅ Zero-loss transitions
// ✅ Optimal buffer sizing
// ✅ Batch processing tuning
```

## Built-in Security (Automatic)

Iris automatically protects your logs without configuration:

```go
// Sensitive data is automatically redacted
logger.Info("User login",
    iris.Str("username", "john_doe"),           // ✅ Visible
    iris.Secret("password", "supersecret123"),  // ❌ Automatically shows [REDACTED]
    iris.Secret("api_key", "sk-1234567890"),    // ❌ Automatically shows [REDACTED]
)

// User input is automatically sanitized against injection attacks
userInput := "malicious\nlevel=error msg=\"HACKED\""
logger.Info("User input received", 
    iris.Str("data", userInput), // ✅ Automatically sanitized
)
```

## Common Patterns

### Web Application Pattern

```go
func main() {
    // Create auto-scaling logger once
    autoLogger, _ := iris.NewAutoScalingLogger(
        iris.Config{Level: iris.Info, Output: os.Stdout},
        iris.DefaultAutoScalingConfig(),
    )
    autoLogger.Start()
    defer autoLogger.Close()
    
    // Use throughout your application
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Automatically optimizes for web request patterns
        autoLogger.Info("Request received",
            iris.Str("method", r.Method),
            iris.Str("path", r.URL.Path),
            iris.Str("user_agent", r.UserAgent()),
        )
    })
    
    autoLogger.Info("Server starting", iris.Int("port", 8080))
    http.ListenAndServe(":8080", nil)
}
```

### Microservice Pattern

```go
func main() {
    // Production microservice setup (Smart API)
    logger, _ := iris.New(iris.Config{})
    logger.Start()
    defer logger.Close()
    
    // Smart API automatically optimizes for microservices:
    // - JSON encoder for structured logs (aggregation-ready)
    // - Info level (production safe)
    // - Multi-threaded architecture for high throughput
    // - Optimal capacity based on system resources
    
    // Component-specific loggers (inherit Smart API optimization)
    authLogger := logger.With(iris.Str("component", "auth"))
    dbLogger := logger.With(iris.Str("component", "database"))
    
    // Each component uses optimized logging automatically
    authLogger.Info("Authentication service started")
    dbLogger.Info("Database connection established")
}
```

## ❌ Common Mistakes to Avoid

### ❌ Don't Manually Configure Performance

```go
// ❌ WRONG: Don't try to manually optimize
logger, _ := iris.New(iris.Config{
    Capacity:  10000,  // ❌ Don't manually set buffer sizes
    BatchSize: 100,    // ❌ Don't manually tune batching
    // ... manual async/sync settings // ❌ Don't configure modes
})

// ✅ RIGHT: Let auto-scaling handle optimization
autoLogger, _ := iris.NewAutoScalingLogger(config, iris.DefaultAutoScalingConfig())
```

### ❌ Don't Use Basic Logger for Production

```go
// ❌ WRONG: Basic logger for production
logger, _ := iris.New(config) // Limited to single mode

// ✅ RIGHT: Auto-scaling logger for production
autoLogger, _ := iris.NewAutoScalingLogger(config, iris.DefaultAutoScalingConfig())
```

### ❌ Don't Forget to Start the Logger

```go
autoLogger, _ := iris.NewAutoScalingLogger(config, iris.DefaultAutoScalingConfig())
// ❌ WRONG: Forgot to start - no auto-scaling!
autoLogger.Info("This won't be optimized")

// ✅ RIGHT: Always start the auto-scaling system
autoLogger.Start()
autoLogger.Info("This will be optimized automatically")
```

## Next Steps

1. **Read More**: Check out [AUTOSCALING_ARCHITECTURE.md](AUTOSCALING_ARCHITECTURE.md) for technical details
2. **Security**: Learn about automatic security features in [SECURE_BY_DESIGN.md](SECURE_BY_DESIGN.md)
3. **Configuration**: Explore advanced configuration in [CONFIGURATION_LOADING.md](CONFIGURATION_LOADING.md)
4. **Context**: Add context integration with [CONTEXT_INTEGRATION.md](CONTEXT_INTEGRATION.md)

## Key Takeaways

- **Zero Configuration**: IRIS handles all performance optimization automatically
- **Intelligent Scaling**: Real-time adaptation to your workload patterns
- **Automatic Security**: Built-in protection without configuration
- **Production Ready**: Just use `NewAutoScalingLogger` and `Start()` - that's it!

---
