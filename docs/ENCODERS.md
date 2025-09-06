# Iris Encoders

Iris provides four built-in encoders to cover different use cases and output formats. Each encoder implements the `Encoder` interface and can be configured independently.

## Available Encoders

### 1. JSON Encoder
**File:** `encoder-json.go`

Implements NDJSON (newline-delimited JSON) format with zero-reflection encoding.

```go
encoder := iris.NewJSONEncoder()

// Customization
encoder.TimeKey = "timestamp"    // default: "ts"
encoder.LevelKey = "severity"    // default: "level"  
encoder.MsgKey = "message"       // default: "msg"
encoder.RFC3339 = false          // default: true (uses UnixNano if false)
```

**Use Cases:**
- Structured logging systems
- Log aggregation (ELK, Splunk, etc.)
- API logging and monitoring
- Machine-readable logs

**Output Example:**
```json
{"ts":"2025-09-06T14:30:45.123Z","level":"info","msg":"User login","user":"john","ip":"192.168.1.1"}
```

### 2. Text Encoder
**File:** `encoder-text.go`

Provides secure human-readable text encoding with comprehensive log injection protection.

```go
encoder := iris.NewTextEncoder()

// Customization
encoder.TimeFormat = time.RFC3339Nano  // default: time.RFC3339
encoder.QuoteValues = false            // default: true
encoder.SanitizeKeys = false           // default: true
```

**Security Features:**
- Field key sanitization
- Value sanitization with proper quoting
- Control character neutralization
- Newline injection protection
- Unicode direction override protection

**Use Cases:**
- Production logs requiring security
- System logs in untrusted environments
- Compliance and audit logging
- General-purpose structured text logs

**Output Example:**
```
time=2025-09-06T14:30:45Z level=info msg="User login" user=john ip=192.168.1.1
```

### 3. Console Encoder
**File:** `encoder-cnsl.go`

Human-readable console output optimized for development and debugging.

```go
encoder := iris.NewConsoleEncoder()
// Or with colors
encoder := iris.NewColorConsoleEncoder()

// Customization
encoder.TimeFormat = time.Kitchen       // default: time.RFC3339Nano
encoder.LevelCasing = "lower"           // default: "upper"
encoder.EnableColor = true              // default: false
```

**Features:**
- Configurable time formatting
- Level casing control
- Optional ANSI color support
- Development-friendly output

**Use Cases:**
- Development environments
- Debugging and troubleshooting
- CLI applications
- Interactive terminals

**Output Example:**
```
2025-09-06T14:30:45.123456789Z INFO User login user=john ip=192.168.1.1
```

### 4. Binary Encoder
**File:** `encoder-binary.go`

Ultra-compact binary serialization optimized for maximum performance and minimal storage.

```go
encoder := iris.NewBinaryEncoder()
// Or compact version
encoder := iris.NewCompactBinaryEncoder()

// Customization
encoder.IncludeLoggerName = false       // default: true
encoder.IncludeCaller = true            // default: false
encoder.IncludeStack = true             // default: false
encoder.UseUnixNano = false             // default: true
```

**Features:**
- Zero reflection encoding
- Minimal allocations
- Optimized varint encoding
- Type-specific value serialization
- Configurable metadata inclusion

**Use Cases:**
- High-throughput systems
- Network protocols requiring binary formats
- Storage systems needing minimal space
- Log shipping with bandwidth constraints
- Performance-critical applications

**Binary Format:**
```
[MAGIC][VERSION][TIMESTAMP][LEVEL][LOGGER_LEN][LOGGER][MSG_LEN][MSG][CALLER_LEN][CALLER][STACK_LEN][STACK][FIELD_COUNT][FIELDS...]
```

## Configuration Examples

### Basic Setup
```go
import "github.com/agilira/iris"

// Choose your encoder
encoder := iris.NewJSONEncoder()        // JSON
// encoder := iris.NewTextEncoder()     // Text
// encoder := iris.NewConsoleEncoder()  // Console
// encoder := iris.NewBinaryEncoder()   // Binary

// Configure Iris
config := iris.NewConfig().
    WithEncoder(encoder).
    WithLevel(iris.Info)

logger := iris.NewLogger(config)
```

### Production Configuration
```go
// Secure text encoder for production
textEncoder := iris.NewTextEncoder()
textEncoder.QuoteValues = true
textEncoder.SanitizeKeys = true

config := iris.NewConfig().
    WithEncoder(textEncoder).
    WithLevel(iris.Warn).
    WithLoggerName("production-service")

logger := iris.NewLogger(config)
```

### Development Configuration
```go
// Colorful console output for development
consoleEncoder := iris.NewColorConsoleEncoder()
consoleEncoder.TimeFormat = time.Kitchen
consoleEncoder.LevelCasing = "lower"

config := iris.NewConfig().
    WithEncoder(consoleEncoder).
    WithLevel(iris.Debug).
    WithCaller(true)

logger := iris.NewLogger(config)
```

### High-Performance Configuration
```go
// Binary encoder for maximum performance
binaryEncoder := iris.NewCompactBinaryEncoder()
// Compact version excludes logger name by default

config := iris.NewConfig().
    WithEncoder(binaryEncoder).
    WithLevel(iris.Info).
    WithRingBuffer(8192)  // Larger buffer for high throughput

logger := iris.NewLogger(config)
```

## Encoder Interface

All encoders implement this interface:

```go
type Encoder interface {
    Encode(rec *Record, now time.Time, buf *bytes.Buffer)
}
```

This allows you to:
- Switch encoders without changing logging code
- Create custom encoders for specialized formats
- Test with different encoders easily
- Mix encoders in different environments

## Best Practices

1. **Choose the right encoder for your use case**
   - JSON: Machine processing and log aggregation
   - Text: Production security and human readability
   - Console: Development and debugging
   - Binary: High-performance and storage efficiency

2. **Configure encoders appropriately**
   - Enable security features (text encoder) in production
   - Use colors (console encoder) only in interactive terminals
   - Minimize metadata (binary encoder) for performance-critical paths

3. **Test with different encoders**
   - Validate output format in your log processing pipeline
   - Verify security features work with your data
   - Benchmark performance with realistic workloads

4. **Consider encoder consistency**
   - Use the same encoder across service instances
   - Document encoder configuration for your team
   - Version control encoder settings with your application

## Advanced Usage

### Custom Field Types
All encoders support the full range of Iris field types:

```go
logger.Info("Complex log entry",
    iris.String("service", "auth"),
    iris.Int64("user_id", 12345),
    iris.Float64("response_time", 0.045),
    iris.Bool("success", true),
    iris.Duration("timeout", 30*time.Second),
    iris.Time("login_time", time.Now()),
    iris.Bytes("payload", []byte("binary data")),
    iris.NamedError("error", err),
    iris.Any("metadata", complexObject),
)
```

### Error Handling
Encoders are designed to be resilient and never panic:

- Invalid field types are encoded as strings
- Nil values are handled gracefully
- Buffer overflow protection is built-in
- Malformed data is sanitized (text encoder)

### Buffer Management
Iris manages buffers efficiently across all encoders:

- Buffers are reused from a pool
- No allocations during normal operation
- Automatic buffer growth when needed
- Memory pressure adaptation

---

Iris • an AGILira fragment
