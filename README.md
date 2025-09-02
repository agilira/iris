# IRIS - Ultra-High Performance Logging Library

[![Go Report Card](https://goreportcard.com/badge/github.com/agilira/iris)](https://goreportcard.com/report/github.com/agilira/iris)
[![GoDoc](https://godoc.org/github.com/agilira/iris?status.svg)](https://godoc.org/github.com/agilira/iris)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

**IRIS** is an ultra-high performance logging library for Go, designed to be the fastest structured logger while providing enterprise-grade security features. Built on the Zephyros MPSC ring buffer, IRIS achieves exceptional performance without sacrificing safety or usability.

## 🚀 Key Features

### ⚡ Ultra-High Performance
- **103x faster timestamps** with intelligent time caching
- **Zero-allocation** logging paths for all field types
- **Lock-free MPSC** ring buffer with adaptive batching
- **Sub-nanosecond** level checking with atomic operations

### 🤖 Automatic Scaling (NEW!)
- **🔄 Zero-Configuration Auto-Scaling**: Automatically switches between SingleRing (~25ns) and MPSC (~35ns) modes
- **📊 Intelligent Load Detection**: Real-time monitoring of contention, latency, and throughput
- **⚡ Transparent Optimization**: No manual async/sync configuration needed - everything is automatic!
- **🎯 Production-Ready**: Self-tuning system that adapts to your application's workload patterns

### 🔒 Enterprise Security (NEW!)
- **🛡️ Sensitive Data Masking**: Automatic redaction of passwords, API keys, tokens
- **🚫 Log Injection Protection**: Complete defense against log manipulation attacks
- **🔐 Unicode Attack Prevention**: Protection against direction override exploits
- **✅ Zero Configuration**: Security enabled by default

### 🛡️ Data Integrity (CRITICAL FIX!)
- **💾 Guaranteed Persistence**: `Sync()` ensures all logs are written before return
- **🚨 Critical Bug Fixed**: Previous versions had data loss risk during shutdown
- **⏱️ Timeout Protection**: 5-second timeout prevents indefinite blocking
- **🔄 Migration Guide**: Complete upgrade documentation available

### � Context Integration (NEW!)
- **🎯 Context.Context Support**: Automatic extraction of context values
- **⚡ Performance Optimized**: Pre-extraction avoids O(n) context.Value() calls
- **🔧 Configurable**: Custom context key mapping and field naming
- **📊 Zero Allocations**: 40ns context extraction with zero memory overhead

### ⚙️ Configuration Management (NEW!)
- **📁 Multi-Source Loading**: JSON files, environment variables, defaults
- **🔄 Hot Reload**: Runtime configuration updates without restart
- **🐳 Container Ready**: Kubernetes ConfigMap and Docker environment support
- **🛡️ Secure**: Built-in validation and secure defaults

### �🎯 Developer Experience
- **Structured logging** with type-safe field constructors
- **Context-aware** field inheritance with `With()`
- **Named logger hierarchies** for component organization
- **Intelligent sampling** to reduce log noise

## 📦 Installation

```bash
go get github.com/agilira/iris
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "github.com/agilira/iris"
)

func main() {
    // Create logger with secure defaults
    logger, err := iris.New(iris.Config{
        Level:  iris.Info,
        Output: os.Stdout,
    })
    if err != nil {
        panic(err)
    }
    defer logger.Close()
    
    logger.Start()
    
    // Zero-allocation structured logging
    // 🤖 AUTOMATIC: No need to configure async/sync - 
    // Iris automatically optimizes performance based on your workload!
    logger.Info("User authenticated",
        iris.Str("username", "john_doe"),
        iris.Int("user_id", 12345),
        iris.Duration("response_time", time.Millisecond*150),
    )
}
```

### 🤖 Auto-Scaling Logger (Recommended for Production)

For production environments, use the AutoScalingLogger that automatically optimizes performance:

```go
package main

import (
    "github.com/agilira/iris"
)

func main() {
    // Create auto-scaling logger with zero configuration
    autoLogger, err := iris.NewAutoScalingLogger(
        iris.Config{
            Level:  iris.Info,
            Output: os.Stdout,
        },
        iris.DefaultAutoScalingConfig(), // 🎯 Production-ready defaults
    )
    if err != nil {
        panic(err)
    }
    defer autoLogger.Close()
    
    // Start the intelligent auto-scaling system
    autoLogger.Start()
    
    // Use normally - auto-scaling is completely transparent!
    // 🔄 Low load: Automatically uses SingleRing mode (~25ns)
    // ⚡ High load: Automatically switches to MPSC mode (~35ns)
    autoLogger.Info("This message will auto-scale based on your application load")
}
```

> **💡 Pro Tip**: Never manually configure async/sync modes! Iris AutoScalingLogger automatically detects your workload patterns and optimizes performance in real-time.

### 🔒 Secure Logging (Major Feature!)

IRIS automatically protects sensitive data and prevents injection attacks:

```go
// Sensitive data is automatically redacted
logger.Info("User login",
    iris.Str("username", "john_doe"),           // ✅ Visible in logs
    iris.Secret("password", "supersecret123"),  // ❌ Shows as [REDACTED]
    iris.Secret("api_key", "sk-1234567890"),    // ❌ Shows as [REDACTED]
)

// All user input is automatically sanitized against injection
logger.Info("User input received",
    iris.Str("filename", userFilename),        // Dangerous chars sanitized
    iris.Str("comment", userComment),          // Injection attempts blocked
)
```

**Output:**
```json
{"time":"2025-08-22T10:30:00Z","level":"info","msg":"User login","username":"john_doe","password":"[REDACTED]","api_key":"[REDACTED]"}
```

## 🔒 Security Features

### Sensitive Data Masking

Automatically redact confidential information:

```go
// Authentication data
iris.Secret("password", userPassword)
iris.Secret("session_token", sessionToken)
iris.Secret("refresh_token", refreshToken)

// API keys and secrets
iris.Secret("api_key", apiKey)
iris.Secret("webhook_secret", webhookSecret)
iris.Secret("connection_string", dbConnectionString)

// Financial information
iris.Secret("credit_card", creditCardNumber)
iris.Secret("bank_account", accountNumber)
iris.Secret("ssn", socialSecurityNumber)
```

### Log Injection Protection

Complete protection against log manipulation attacks:

```go
// All these injection attempts are automatically neutralized
maliciousInput := "normal\nlevel=error msg=\"SYSTEM HACKED\""
logger.Info("User input", iris.Str("data", maliciousInput))
// Result: data="normal_level_error msg_\"SYSTEM HACKED\""

maliciousKey := "user\" admin=\"true"
logger.Info("Action", iris.Str(maliciousKey, "value"))
// Result: user__admin__true="value"
```

**📖 For complete security documentation, see [docs/SECURE_BY_DESIGN.md](docs/SECURE_BY_DESIGN.md)**

## ⚡ Performance

IRIS delivers exceptional performance with security enabled:

```
BenchmarkTimeCache      1000000000    0.48 ns/op     0 allocs/op  (103x faster than time.Now)
BenchmarkSecretField    500000000     2.38 ns/op     0 allocs/op  (4.9% overhead)
BenchmarkJSONEncoder    2500000       423 ns/op      0 allocs/op
BenchmarkTextEncoder    2000000       481 ns/op      0 allocs/op
```

## � Sugar APIs (Printf-Style)

For developers who prefer familiar printf-style syntax:

```go
// Printf-style logging (convenience APIs)
logger.Debugf("Debug: %s = %d", "count", 10)
logger.Infof("User %s logged in with ID %d", "john", 123)
logger.Warnf("Warning: %d errors found", 3)
logger.Errorf("Error: %s failed with code %d", "operation", 500)

// Equivalent structured logging (zero-allocation)
logger.Debug("Debug", iris.Str("key", "count"), iris.Int("value", 10))
logger.Info("User login", iris.Str("user", "john"), iris.Int("id", 123))
```

**Trade-off**: Sugar APIs sacrifice zero-allocation guarantees for developer convenience.

**📖 For complete sugar API documentation, see [docs/SUGAR_API.md](docs/SUGAR_API.md)**

## �🎯 Advanced Usage

### Hierarchical Loggers

```go
// Create component-specific loggers
authLogger := logger.With(iris.Str("component", "auth"))
dbLogger := logger.With(iris.Str("component", "database"))

// Each maintains its context
authLogger.Info("Login attempt", iris.Str("user", "john"))
// Output: {"component":"auth","msg":"Login attempt","user":"john"}

dbLogger.Error("Connection failed", iris.Str("host", "db.example.com"))
// Output: {"component":"database","msg":"Connection failed","host":"db.example.com"}
```

### Custom Encoders

```go
// JSON encoder for production
jsonConfig := iris.Config{
    Encoder: iris.NewJSONEncoder(),
    Level:   iris.Info,
}

// Text encoder for development
textConfig := iris.Config{
    Encoder: iris.NewTextEncoder(),
    Level:   iris.Debug,
}
```

### Error Handling with Security

```go
logger.Error("Database operation failed",
    iris.Str("operation", "UPDATE"),
    iris.Str("table", "users"),
    iris.Secret("connection_string", dbURL),  // Never leaked
    iris.Str("error", err.Error()),           // Sanitized automatically
    iris.Int("affected_rows", rowCount),
)
```

## 🌐 Context Integration

### Automatic Context Extraction

```go
import "context"

// Create context with request information
ctx := context.Background()
ctx = context.WithValue(ctx, iris.RequestIDKey, "req-12345")
ctx = context.WithValue(ctx, iris.UserIDKey, "user-67890")

// Extract context once, use many times (optimized!)
contextLogger := logger.WithContext(ctx)

// All logs include context fields automatically
contextLogger.Info("Processing request")
contextLogger.Debug("Validating input") 
contextLogger.Info("Request completed")

// Output includes: request_id="req-12345", user_id="user-67890"
```

### Fast Context Methods

```go
// Optimized single-field extraction
requestLogger := logger.WithRequestID(ctx)
userLogger := logger.WithUserID(ctx)
traceLogger := logger.WithTraceID(ctx)

requestLogger.Info("Request processed")
// Output: {"request_id":"req-12345","msg":"Request processed",...}
```

### Custom Context Extraction

```go
// Define custom context keys and field mapping
extractor := &iris.ContextExtractor{
    Keys: map[iris.ContextKey]string{
        iris.RequestIDKey:              "req_id",      // Rename field
        iris.ContextKey("tenant_id"):   "tenant",      // Custom key
        iris.ContextKey("session_id"):  "session",     // Another custom
    },
}

contextLogger := logger.WithContextExtractor(ctx, extractor)
contextLogger.Info("Multi-tenant operation")
// Output includes: req_id, tenant, session
```

## ⚙️ Configuration Management

### JSON Configuration

```go
// config.json
{
  "level": "info",
  "format": "json", 
  "output": "stdout",
  "capacity": 8192,
  "batch_size": 32,
  "enable_caller": true
}

// Load configuration
config, err := iris.LoadConfigFromJSON("config.json")
if err != nil {
    log.Fatal(err)
}

logger, err := iris.New(config)
```

### Environment Variables

```bash
# Set via environment
export IRIS_LEVEL=debug
export IRIS_CAPACITY=16384
export IRIS_OUTPUT=/var/log/app.log
export IRIS_ENABLE_CALLER=true
```

```go
// Load from environment
config, err := iris.LoadConfigFromEnv()
if err != nil {
    log.Fatal(err) 
}

logger, err := iris.New(config)
```

### Multi-Source Configuration

```go
// Load with precedence: Environment > JSON > Defaults
config, err := iris.LoadConfigMultiSource("config.json")
if err != nil {
    log.Fatal(err)
}

// Environment variables override JSON settings
// JSON settings override built-in defaults
logger, err := iris.New(config)
```

## 🛠️ Configuration

### Production Configuration

```go
config := iris.Config{
    Level:     iris.Info,
    Capacity:  65536,        // Ring buffer size
    BatchSize: 32,           // Batch processing size
    Output:    os.Stdout,
    Encoder:   iris.NewJSONEncoder(),
    TimeFn:    timecache.CachedTime,  // 103x faster timestamps
}

logger, err := iris.New(config)
```

### Development Configuration

```go
config := iris.Config{
    Level:   iris.Debug,
    Output:  os.Stderr,
    Encoder: iris.NewTextEncoder(),  // Human-readable
}

logger, err := iris.New(config)
```

## 📊 Field Types

IRIS supports all common field types with zero allocations:

```go
// Strings and binary data
iris.Str("key", "value")
iris.Bytes("data", []byte{1, 2, 3})
iris.Secret("password", "secret")  // 🔒 Automatically redacted

// Numbers
iris.Int("count", 42)
iris.Int64("id", 1234567890)
iris.Uint64("size", 1024)
iris.Float64("ratio", 3.14159)

// Time and duration
iris.Time("timestamp", time.Now())
iris.Duration("elapsed", time.Millisecond*150)

// Boolean
iris.Bool("active", true)
```

## 🔄 Migration from Other Libraries

### From logrus

```go
// Before (logrus)
log.WithFields(log.Fields{
    "user": "john",
    "password": "secret",  // ❌ Potential leak
}).Info("User login")

// After (IRIS)
logger.Info("User login",
    iris.Str("user", "john"),
    iris.Secret("password", "secret"),  // ✅ Automatically redacted
)
```

### From zap

```go
// Before (zap)
logger.Info("User login",
    zap.String("user", "john"),
    zap.String("password", "secret"),  // ❌ Potential leak
)

// After (IRIS)
logger.Info("User login",
    iris.Str("user", "john"),
    iris.Secret("password", "secret"),  // ✅ Automatically redacted
)
```

## 📚 Documentation

- **[Security Guide](docs/SECURE_BY_DESIGN.md)** - Complete security features documentation
- **[Security Reference](docs/SECURITY_REFERENCE.md)** - Quick security reference
- **[API Documentation](docs/API.md)** - Complete API reference
- **[Performance Guide](docs/PERFORMANCE.md)** - Performance optimization tips

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Security

For security vulnerabilities, please see [SECURITY.md](SECURITY.md) for responsible disclosure guidelines.

## � Documentation

### Getting Started
- **[Quick Start Guide](docs/QUICK_START.md)** - 🚀 **START HERE** - Complete beginner's guide with auto-scaling concepts
- **[Auto-Scaling Architecture](docs/AUTOSCALING_ARCHITECTURE.md)** - Technical deep-dive into automatic performance optimization

### Core Features
- **[Security Reference](docs/SECURE_BY_DESIGN.md)** - Complete security features guide
- **[Sugar API Guide](docs/SUGAR_API.md)** - Printf-style logging API documentation
- **[API Documentation](docs/API.md)** - Full API reference
- **[Performance Guide](docs/PERFORMANCE.md)** - Optimization tips and benchmarks

### New Features
- **[Context Integration Guide](docs/CONTEXT_INTEGRATION.md)** - Complete context.Context integration
- **[Configuration Loading Guide](docs/CONFIGURATION_LOADING.md)** - Multi-source configuration management

### Advanced Topics
- **[Sync() Integration Guide](docs/SYNC_INTEGRATION_GUIDE.md)** - Complete data integrity and sync patterns
- **[Sync() Migration Guide](docs/SYNC_MIGRATION_GUIDE.md)** - Critical security update and migration
- **[Best Practices](docs/BEST_PRACTICES.md)** - Production deployment guidelines
- **[Security Reference](docs/SECURITY_REFERENCE.md)** - Security implementation details
- **[Contributing](CONTRIBUTING.md)** - Development and contribution guidelines

## �📄 License

IRIS is licensed under the [Mozilla Public License 2.0](LICENSE.md).

## 🏆 Why Choose IRIS?

| Feature | IRIS | zap | logrus | slog |
|---------|------|-----|--------|------|
| **Performance** | 🟢 Fastest | 🟡 Fast | 🔴 Slow | 🟡 Medium |
| **Security** | 🟢 Built-in | 🔴 None | 🔴 None | 🔴 None |
| **Zero Alloc** | 🟢 Yes | 🟡 Partial | 🔴 No | 🟡 Partial |
| **Injection Protection** | 🟢 Complete | 🔴 None | 🔴 None | 🔴 None |
| **Sensitive Data Masking** | 🟢 Automatic | 🔴 Manual | 🔴 Manual | 🔴 Manual |
| **Ease of Use** | 🟢 Simple | 🟡 Complex | 🟢 Simple | 🟡 Medium |

## 🎯 Roadmap

- [ ] **Advanced Security**: Regex-based PII detection
- [ ] **Observability**: OpenTelemetry integration
- [ ] **Formats**: Additional encoder formats (XML, YAML)
- [ ] **Sampling**: Advanced sampling strategies
- [ ] **Encryption**: Log encryption at rest
- [ ] **Digital Signatures**: Log integrity verification

## ❓ Frequently Asked Questions

### Q: Do I need to manually configure async/sync modes for production?
**A: NO!** This is the most important concept in IRIS. The `AutoScalingLogger` automatically handles all performance optimizations. Never manually configure async/sync modes - the auto-scaling system is designed to be completely transparent and optimal.

### Q: What's the difference between `iris.New()` and `iris.NewAutoScalingLogger()`?
**A:** 
- `iris.New()`: Basic logger, good for development or simple use cases
- `iris.NewAutoScalingLogger()`: **Recommended for production** - automatically optimizes performance based on your workload

### Q: How do I know if auto-scaling is working?
**A:** You can monitor it with:
```go
stats := autoLogger.GetScalingStats()
fmt.Printf("Current mode: %s, Scale operations: %d\n", 
    stats.CurrentMode, stats.TotalScaleOperations)
```

### Q: What performance should I expect?
**A:** 
- **Low contention**: ~25ns/op (SingleRing mode)
- **High contention**: ~35ns/op per thread (MPSC mode)
- **Automatic transitions**: Zero log loss during mode switches

### Q: Do I need to configure buffer sizes or batch parameters?
**A: NO!** The auto-scaling system automatically tunes all parameters based on your application's real-time performance metrics.

---

**Built with ❤️ by the AGILira team**

*Making logging fast, secure, and simple.*
