# Iris Binary Logger - User Guide

## Quick Start

### Installation
```bash
go get github.com/agilira/iris
```

### Basic Setup
```go
package main

import (
    "github.com/agilira/iris"
)

func main() {
    // Create logger
    logger := iris.NewIrisLogger(iris.InfoLevel)
    
    // Simple logging
    logger.Info("Application started")
    
    // Structured logging
    ctx := logger.XFields(
        iris.XStr("user", "john_doe"),
        iris.XInt("session_id", 12345),
        iris.XBool("authenticated", true),
    )
    ctx.Info("User logged in successfully")
}
```

## Console Output

### Default Output Format
By default, logs are written to `stdout` in binary format. To see human-readable output:

```bash
# Run your application and pipe to decoder
go run main.go | iris-decoder

# Or redirect to file and decode later
go run main.go > logs.bin
iris-decoder < logs.bin
```

### JSON Output Mode
For development and debugging, enable JSON output:

```go
logger := iris.NewIrisLogger(iris.DebugLevel)
logger.SetOutputFormat(iris.JSONFormat) // Human-readable JSON
```

### Output Example
```json
{
  "timestamp": "2025-08-21T10:30:45.123Z",
  "level": "info",
  "message": "User logged in successfully",
  "fields": {
    "user": "john_doe",
    "session_id": 12345,
    "authenticated": true
  }
}
```

## Development Workflow

### 1. Development Mode
```go
func setupDevLogger() *iris.XLogger {
    logger := iris.NewIrisLogger(iris.DebugLevel)
    logger.SetOutputFormat(iris.JSONFormat)
    logger.EnableCaller(true) // Show file:line info
    return logger
}
```

### 2. Production Mode
```go
func setupProdLogger() *iris.XLogger {
    logger := iris.NewIrisLogger(iris.InfoLevel)
    logger.SetOutputFormat(iris.BinaryFormat) // Maximum performance
    logger.EnableCaller(false) // Disable caller info
    return logger
}
```

### 3. Testing Mode
```go
func setupTestLogger() *iris.XLogger {
    logger := iris.NewIrisLogger(iris.ErrorLevel) // Only errors
    logger.SetOutput(io.Discard) // Silent during tests
    return logger
}
```

## Common Patterns

### Request Logging
```go
func handleRequest(w http.ResponseWriter, r *http.Request, logger *iris.XLogger) {
    // Create request context
    ctx := logger.XFields(
        iris.XStr("method", r.Method),
        iris.XStr("path", r.URL.Path),
        iris.XStr("remote_addr", r.RemoteAddr),
        iris.XStr("user_agent", r.UserAgent()),
    )
    
    start := time.Now()
    
    // Log request start
    ctx.Info("Request started")
    
    // Process request...
    
    // Log request completion
    ctx.XFields(
        iris.XInt("duration_ms", int(time.Since(start).Milliseconds())),
        iris.XInt("status_code", 200),
    ).Info("Request completed")
}
```

### Error Handling
```go
func processData(data []byte, logger *iris.XLogger) error {
    ctx := logger.XFields(
        iris.XInt("data_size", len(data)),
        iris.XStr("operation", "process_data"),
    )
    
    if len(data) == 0 {
        ctx.Error("Empty data received")
        return errors.New("empty data")
    }
    
    // Process data...
    if err := validate(data); err != nil {
        ctx.XFields(
            iris.XStr("error", err.Error()),
            iris.XStr("validation_stage", "format_check"),
        ).Error("Data validation failed")
        return err
    }
    
    ctx.Info("Data processed successfully")
    return nil
}
```

### Performance Monitoring
```go
func monitorPerformance(logger *iris.XLogger) {
    ctx := logger.XFields(
        iris.XStr("component", "api_server"),
        iris.XStr("version", "v1.2.3"),
    )
    
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        ctx.XFields(
            iris.XInt("goroutines", runtime.NumGoroutine()),
            iris.XInt("heap_mb", int(m.HeapAlloc/1024/1024)),
            iris.XInt("gc_cycles", int(m.NumGC)),
        ).Info("Performance metrics")
    }
}
```

## Debugging and Troubleshooting

### Enable Debug Logging
```go
logger := iris.NewIrisLogger(iris.DebugLevel)
logger.SetOutputFormat(iris.JSONFormat)

// Add caller information
logger.EnableCaller(true)

// Debug specific operations
ctx := logger.XFields(iris.XStr("debug_session", "session_123"))
ctx.Debug("Starting debug session")
```

### Memory Usage Monitoring
```go
func logMemoryUsage(ctx *iris.XContext) {
    footprint := ctx.MemoryFootprint()
    binarySize := ctx.GetBinarySize()
    
    ctx.XFields(
        iris.XInt("memory_footprint", footprint),
        iris.XInt("binary_size", binarySize),
    ).Debug("Context memory usage")
}
```

### Performance Profiling
```go
func profileLogging(logger *iris.XLogger) {
    start := time.Now()
    
    // Create context once, reuse multiple times
    ctx := logger.XFields(
        iris.XStr("operation", "batch_processing"),
        iris.XInt("batch_size", 1000),
    )
    
    // Log 1000 messages
    for i := 0; i < 1000; i++ {
        ctx.XFields(
            iris.XInt("item_id", i),
        ).Info("Processing item")
    }
    
    duration := time.Since(start)
    logger.XFields(
        iris.XInt("operations", 1000),
        iris.XInt("duration_ns", int(duration.Nanoseconds())),
        iris.XInt("ops_per_second", int(1000*time.Second/duration)),
    ).Info("Performance profile completed")
}
```

## Log Analysis

### Using Standard Tools
```bash
# Count log levels
iris-decoder < logs.bin | jq -r '.level' | sort | uniq -c

# Filter by field
iris-decoder < logs.bin | jq 'select(.fields.user == "john_doe")'

# Extract errors from last hour
iris-decoder < logs.bin | jq 'select(.level == "error" and (.timestamp | fromdateiso8601) > (now - 3600))'
```

### Custom Analysis Scripts
```go
// analyze_logs.go
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
)

type LogEntry struct {
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Timestamp string                 `json:"timestamp"`
    Fields    map[string]interface{} `json:"fields"`
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    errorCount := 0
    
    for scanner.Scan() {
        var entry LogEntry
        if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
            continue
        }
        
        if entry.Level == "error" {
            errorCount++
            fmt.Printf("ERROR: %s - %s\n", entry.Timestamp, entry.Message)
        }
    }
    
    fmt.Printf("Total errors: %d\n", errorCount)
}
```

## Configuration Examples

### Environment-Based Configuration
```go
func createLogger() *iris.XLogger {
    env := os.Getenv("ENV")
    
    switch env {
    case "development":
        logger := iris.NewIrisLogger(iris.DebugLevel)
        logger.SetOutputFormat(iris.JSONFormat)
        logger.EnableCaller(true)
        return logger
        
    case "staging":
        logger := iris.NewIrisLogger(iris.InfoLevel)
        logger.SetOutputFormat(iris.JSONFormat)
        return logger
        
    case "production":
        logger := iris.NewIrisLogger(iris.WarnLevel)
        logger.SetOutputFormat(iris.BinaryFormat)
        return logger
        
    default:
        return iris.NewIrisLogger(iris.InfoLevel)
    }
}
```

### Configuration File
```yaml
# config.yaml
logging:
  level: "info"
  format: "json"
  enable_caller: false
  output_file: "/var/log/app.log"
```

```go
type LogConfig struct {
    Level        string `yaml:"level"`
    Format       string `yaml:"format"`
    EnableCaller bool   `yaml:"enable_caller"`
    OutputFile   string `yaml:"output_file"`
}

func setupLoggerFromConfig(configPath string) (*iris.XLogger, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var config struct {
        Logging LogConfig `yaml:"logging"`
    }
    
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    level, err := iris.ParseLevel(config.Logging.Level)
    if err != nil {
        return nil, err
    }
    
    logger := iris.NewIrisLogger(level)
    
    if config.Logging.Format == "json" {
        logger.SetOutputFormat(iris.JSONFormat)
    }
    
    logger.EnableCaller(config.Logging.EnableCaller)
    
    if config.Logging.OutputFile != "" {
        file, err := os.OpenFile(config.Logging.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            return nil, err
        }
        logger.SetOutput(file)
    }
    
    return logger, nil
}
```

## Best Practices

### 1. Context Reuse
```go
// ✅ Good: Reuse context
baseCtx := logger.XFields(
    iris.XStr("service", "user-api"),
    iris.XStr("version", "1.0.0"),
)

// Use baseCtx for all operations in this service
baseCtx.Info("Service started")
baseCtx.XFields(iris.XStr("user_id", "123")).Info("User created")

// ❌ Bad: Create new context every time
logger.XFields(iris.XStr("service", "user-api")).Info("Service started")
logger.XFields(iris.XStr("service", "user-api")).Info("User created")
```

### 2. Field Naming
```go
// ✅ Good: Consistent, descriptive names
ctx := logger.XFields(
    iris.XStr("user_id", "12345"),
    iris.XStr("request_id", "req_67890"),
    iris.XInt("response_time_ms", 150),
    iris.XBool("cache_hit", true),
)

// ❌ Bad: Inconsistent, unclear names
ctx := logger.XFields(
    iris.XStr("user", "12345"),      // Missing _id suffix
    iris.XStr("reqId", "req_67890"), // CamelCase instead of snake_case
    iris.XInt("time", 150),          // Unclear units
    iris.XBool("cached", true),      // Unclear meaning
)
```

### 3. Error Context
```go
// ✅ Good: Rich error context
func processUser(userID string, logger *iris.XLogger) error {
    ctx := logger.XFields(
        iris.XStr("user_id", userID),
        iris.XStr("operation", "process_user"),
    )
    
    user, err := fetchUser(userID)
    if err != nil {
        ctx.XFields(
            iris.XStr("error", err.Error()),
            iris.XStr("stage", "fetch_user"),
        ).Error("Failed to fetch user")
        return err
    }
    
    // Continue processing...
}

// ❌ Bad: Minimal error context
func processUser(userID string, logger *iris.XLogger) error {
    user, err := fetchUser(userID)
    if err != nil {
        logger.Error("Error") // Useless
        return err
    }
}
```

### 4. Performance Optimization
```go
// ✅ Good: Check level before expensive operations
if logger.IsDebugEnabled() {
    expensiveData := generateDebugInfo() // Only when needed
    logger.XFields(
        iris.XStr("debug_data", expensiveData),
    ).Debug("Debug information")
}

// ❌ Bad: Always generate expensive data
expensiveData := generateDebugInfo() // Always executed
logger.XFields(
    iris.XStr("debug_data", expensiveData),
).Debug("Debug information") // Might be discarded
```

## Troubleshooting

### Common Issues

**Issue: No output visible**
```go
// Check if level is too high
logger := iris.NewIrisLogger(iris.ErrorLevel)
logger.Info("This won't appear") // Info < Error level
```

**Issue: Binary output unreadable**
```go
// Switch to JSON for debugging
logger.SetOutputFormat(iris.JSONFormat)
```

**Issue: Poor performance**
```go
// Avoid string concatenation in hot paths
// ❌ Bad
logger.Info("User " + userID + " logged in")

// ✅ Good
logger.XFields(iris.XStr("user_id", userID)).Info("User logged in")
```

**Issue: Memory leaks**
```go
// Always release fields when done
field := iris.XStr("temp", "value")
defer field.Release() // Return to pool
```

---

*For API reference and technical details, see [API.md](API.md)*
