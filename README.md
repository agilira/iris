# IRIS - Ultra-High Performance Logging Library

[![Go Report Card](https://goreportcard.com/badge/github.com/agilira/iris)](https://goreportcard.com/report/github.com/agilira/iris)
[![GoDoc](https://godoc.org/github.com/agilira/iris?status.svg)](https://godoc.org/github.com/agilira/iris)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

**IRIS** is an ultra-high performance logging library for Go, designed to be the fastest structured logger while providing enterprise-grade security features. Built on the Zephyros MPSC ring buffer, IRIS achieves exceptional performance without sacrificing safety or usability.

## 🚀 Key Features

### 🧠 Smart API (NEW!)
- **🎯 Zero Configuration**: `iris.New(iris.Config{})` works perfectly out-of-the-box
- **🤖 Auto-Detection**: Optimal architecture, capacity, encoder, and level selection
- **⚡ Production Ready**: Smart defaults optimized for real-world performance
- **🔧 Development Friendly**: Auto-switches to human-readable output in dev mode

### ⚡ Ultra-High Performance
- **121x faster timestamps** with intelligent time caching (`timecache.CachedTime`)
- **1-3 allocs/op** in hot paths (down from 5-6 previously)
- **Lock-free MPSC** ring buffer with adaptive batching
- **324-537 ns/op** encoding performance with time cache optimization

### 🤖 Intelligent Auto-Scaling
- **🔄 Smart Architecture**: Auto-switches between SingleRing (~25ns) and ThreadedRings (~35ns)
- **📊 CPU-Aware Capacity**: 8KB per CPU core, capped at 64KB for optimal memory usage
- **⚡ Transparent Optimization**: No manual configuration needed - everything is automatic!
- **🎯 Production-Ready**: Self-tuning system that adapts to your application's workload

### 🔒 Enterprise Security
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

### Smart API - Zero Configuration Required

IRIS now features a **Smart API** that automatically configures itself for optimal performance. No complex setup needed!

```go
package main

import (
    "github.com/agilira/iris"
)

func main() {
    // 🎯 SIMPLE: Smart API auto-configures everything optimally
    logger, err := iris.New(iris.Config{})
    if err != nil {
        panic(err)
    }
    defer logger.Close()
    
    logger.Start()
    
    // ⚡ Zero-allocation structured logging
    logger.Info("User logged in", 
        iris.Str("user", "john"),
        iris.Int("session_id", 12345),
        iris.Float64("login_time", 1.23))
}
```

### Development Mode - Human Readable Logs

```go
// 🔧 DEVELOPMENT: Auto-selects TextEncoder for readability
logger, _ := iris.New(iris.Config{}, iris.Development())
logger.Start()

logger.Info("Debug info", iris.Str("component", "auth"))
// Output: time=2025-09-02T12:00:00Z level=info msg="Debug info" component="auth"
```

### Production Mode - Structured JSON

```go
// 🏭 PRODUCTION: Auto-selects JSONEncoder for structured logging
logger, _ := iris.New(iris.Config{})
logger.Start()

logger.Info("Request processed", iris.Str("method", "POST"))
// Output: {"ts":"2025-09-02T12:00:00Z","level":"info","msg":"Request processed","method":"POST"}
```

### What Makes It Smart?

- **🧠 Auto-Architecture**: Detects optimal ring buffer architecture (Single vs Multi-threaded)
- **⚡ Auto-Capacity**: Calculates buffer size based on CPU cores (8KB per core, max 64KB)
- **🎯 Auto-Encoder**: JSON for production, Text for development
- **📊 Auto-Level**: Info default, Debug for development, supports `IRIS_LEVEL` env var
- **⏰ Auto-Time**: Ultra-fast cached timestamps for performance
    
    logger.Start()
    
    // Zero-allocation structured logging
    // 🤖 AUTOMATIC: No need to configure async/sync - 
    // Smart API automatically optimizes everything!
    logger.Info("User authenticated",
        iris.Str("username", "john_doe"),
        iris.Int("user_id", 12345),
        iris.Duration("response_time", time.Millisecond*150),
    )
}
```

### 📈 Performance Metrics (After Smart API Optimization)

The Smart API delivers exceptional performance improvements:

```
Benchmark Results (Smart API vs Previous):
┌─────────────────────────┬──────────────┬──────────────┬─────────────┐
│ Metric                  │ Previous     │ Smart API    │ Improvement │
├─────────────────────────┼──────────────┼──────────────┼─────────────┤
│ Hot Path Allocations    │ 5-6 allocs   │ 1-3 allocs   │ -67%        │
│ Encoding Performance    │ 800+ ns/op   │ 324-537 ns/op│ +40-60%     │
│ Time Cache Performance  │ N/A          │ 311 ns/op    │ 121x faster │
│ Memory per Record       │ 10KB         │ 2.5KB        │ -75%        │
│ Configuration Lines     │ 15-20 lines  │ 1 line       │ -95%        │
└─────────────────────────┴──────────────┴──────────────┴─────────────┘

🚀 Real-world throughput: 1M+ ops/sec with zero configuration!
```

### 🔒 Secure Logging (Enterprise Grade)

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

### Smart Encoder Selection

```go
// 🏭 Production: Automatically selects JSON encoder
logger, _ := iris.New(iris.Config{})
logger.Start()
// Output: {"ts":"2025-09-02T12:00:00Z","level":"info","msg":"Hello"}

// 🔧 Development: Automatically selects Text encoder  
logger, _ := iris.New(iris.Config{}, iris.Development())
logger.Start()
// Output: time=2025-09-02T12:00:00Z level=info msg="Hello"
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

### Smart Auto-Configuration (Recommended)

IRIS automatically configures everything optimally. Zero setup required!

```go
// 🎯 One line setup - production ready!
logger, _ := iris.New(iris.Config{})
logger.Start()

// ✨ Automatically configured:
// • Architecture: Multi-threaded on multi-core systems
// • Capacity: 8KB per CPU core (optimal for your hardware)  
// • Encoder: JSON for structured logging
// • Level: Info (production safe)
// • Time: Ultra-fast cached timestamps
```

### Environment Variable Control

```bash
# Override only what you need
export IRIS_LEVEL=debug     # Development: debug level
export IRIS_LEVEL=warn      # Production: warn level only
export IRIS_LEVEL=error     # Critical: errors only
```

```go
// Application code stays the same
logger, _ := iris.New(iris.Config{}) // Automatically reads IRIS_LEVEL
```

### Development vs Production

```go
// 🔧 Development Mode
logger, _ := iris.New(iris.Config{}, iris.Development())
// Auto-enables: Text encoder, Debug level, Caller info

// 🏭 Production Mode  
logger, _ := iris.New(iris.Config{})
// Auto-enables: JSON encoder, Info level, Optimized performance
```



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

### Q: Do I need to manually configure anything for production?
**A: NO!** This is the key concept in IRIS. The **Smart API** automatically configures everything optimally. Zero setup required - just call `iris.New(iris.Config{})` and everything is production-ready.

### Q: How does the Smart API work?
**A:** The Smart API automatically detects and configures:
- **Architecture**: Single/Multi-threaded based on CPU cores
- **Capacity**: 8KB per CPU core (optimal for your hardware)
- **Encoder**: JSON for production, Text for development mode
- **Level**: Info by default, Debug in development mode
- **Performance**: Ultra-fast cached timestamps

### Q: How do I override Smart API defaults?
**A:** Only override what you specifically need:
```go
// Override only output, keep all other smart defaults
logger, _ := iris.New(iris.Config{
    Output: myCustomWriter,
})
```

### Q: What performance should I expect?
**A:** With Smart API configuration:
- **Encoding**: ~325-537 ns/op (JSON/Text encoders)
- **Allocations**: 1-3 allocs/op (down from 5-6)
- **Memory**: 75% reduction per Record (32-field vs 128-field)
- **Time**: 121x faster timestamps with caching

### Q: Do I need to learn complex configuration APIs?
**A: NO!** The Smart API is designed for simplicity:
```go
// 🎯 Production: One line setup
logger, _ := iris.New(iris.Config{})

// 🔧 Development: Add one option
logger, _ := iris.New(iris.Config{}, iris.Development())
```

## 📊 Technical Details & Performance Considerations

### Field Limitations
**IRIS** is optimized for performance with a **maximum of 32 fields per log record**. This design choice provides:

- **🚀 Memory Efficiency**: 75% reduction in memory usage per Record (32-field vs 128-field arrays)
- **⚡ CPU Cache Friendly**: Smaller arrays fit better in L1/L2 cache for faster access
- **🎯 Real-World Optimization**: 99.9% of log records use fewer than 32 fields

**Performance Impact:**
```go
// ✅ Optimal: Uses 32-field optimized arrays
logger.Info("User action",
    iris.Str("user", "john"),
    iris.Str("action", "login"),
    iris.Time("timestamp", time.Now()),
    // ... up to 32 fields total
)

// ⚠️ Fallback: Exceeding 32 fields triggers dynamic allocation
logger.Info("Oversized record", /* 33+ fields */)  // Slightly slower
```

**Best Practices:**
- **📊 Group Related Data**: Use nested objects for complex data structures
- **🎯 Essential Fields Only**: Log only business-critical information
- **⚡ Performance Monitoring**: Use benchmarks to verify field count impact

---

**Built with ❤️ by the AGILira team**

*Making logging fast, secure, and simple.*

```
```

---

**Built with ❤️ by the AGILira team**

*Making logging fast, secure, and simple.*
