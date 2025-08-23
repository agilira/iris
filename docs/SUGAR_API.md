# IRIS Sugar API Documentation

## 🍯 Overview

IRIS provides "sugared" APIs for developers who prefer printf-style logging syntax. These convenience methods offer familiar formatting while maintaining IRIS's high performance characteristics.

**Trade-off**: Sugar APIs sacrifice zero-allocation guarantees for developer convenience and familiar syntax.

## 📋 Available Sugar Methods

### Debug Level
```go
func (l *Logger) Debugf(format string, args ...any) bool
```

### Info Level  
```go
func (l *Logger) Infof(format string, args ...any) bool
```

### Warn Level
```go
func (l *Logger) Warnf(format string, args ...any) bool
```

### Error Level
```go
func (l *Logger) Errorf(format string, args ...any) bool
```

## 🚀 Usage Examples

### Basic Printf-Style Logging
```go
logger, err := iris.New(iris.Config{
    Level:   iris.Debug,
    Output:  os.Stdout,
    Encoder: iris.NewJSONEncoder(),
})
if err != nil {
    log.Fatal(err)
}
logger.Start()
defer logger.Close()

// Sugar API examples
logger.Debugf("Debug: %s = %d", "count", 10)
logger.Infof("User %s logged in with ID %d", "john", 123)
logger.Warnf("Warning: %d errors found in %s", 3, "database")
logger.Errorf("Error: %s failed with code %d", "operation", 500)
```

**Output:**
```json
{"level":"debug","msg":"Debug: count = 10","time":"2025-08-23T10:30:00Z"}
{"level":"info","msg":"User john logged in with ID 123","time":"2025-08-23T10:30:01Z"}
{"level":"warn","msg":"Warning: 3 errors found in database","time":"2025-08-23T10:30:02Z"}
{"level":"error","msg":"Error: operation failed with code 500","time":"2025-08-23T10:30:03Z"}
```

### Complex Formatting
```go
// Multiple data types
logger.Infof("Process %d completed in %v with status %t", 
    processID, duration, success)

// String formatting with precision
logger.Debugf("Coordinates: lat=%.6f, lon=%.6f", latitude, longitude)

// Conditional logging with formatting
if retryCount > 0 {
    logger.Warnf("Retry attempt #%d for operation %q", retryCount, operation)
}
```

### Error Handling with Context
```go
func processRequest(userID int, data []byte) error {
    logger.Debugf("Processing request for user %d (size: %d bytes)", userID, len(data))
    
    if err := validateData(data); err != nil {
        logger.Errorf("Validation failed for user %d: %v", userID, err)
        return err
    }
    
    logger.Infof("Successfully processed request for user %d", userID)
    return nil
}
```

## ⚡ Performance Characteristics

### Sugar API Performance
- **Allocation**: ✅ Allocates memory for string formatting
- **Speed**: ~50-100ns per call (vs ~25ns for structured API)
- **Level Check**: ✅ Fast atomic level checking (early exit)
- **Sampling**: ✅ Supports intelligent sampling
- **Thread Safety**: ✅ Fully thread-safe

### Internal Implementation
```go
func (l *Logger) logf(level Level, format string, args ...any) bool {
    // Fast level check (atomic operation)
    if !l.shouldLog(level) {
        return true  // Early exit - zero work done
    }
    
    // String formatting (allocation occurs here)
    var sb strings.Builder
    sb.Grow(len(format) + 32)  // Pre-allocate buffer
    sb.WriteString(fmt.Sprintf(format, args...))
    
    // Delegate to structured logging path
    return l.log(level, sb.String())
}
```

## 📊 Sugar vs Structured API Comparison

| Feature | Sugar API | Structured API |
|---------|-----------|----------------|
| **Syntax** | `logger.Infof("User %s", name)` | `logger.Info("User login", iris.Str("user", name))` |
| **Familiarity** | ✅ Printf-style (familiar) | ❌ Requires learning field syntax |
| **Performance** | ⚠️ ~50-100ns + allocations | ✅ ~25ns + zero allocations |
| **Type Safety** | ❌ Runtime format errors | ✅ Compile-time type safety |
| **Parsing** | ❌ Harder to parse logs | ✅ Structured, easy to parse |
| **Security** | ⚠️ Format string injection risk | ✅ Injection-safe by design |

## 🛡️ Security Considerations

### Format String Safety
```go
// ❌ DANGEROUS: User input in format string
userInput := "malicious%s%s%s"
logger.Infof(userInput, "data")  // Potential format string attack

// ✅ SAFE: User input as argument
logger.Infof("User input: %s", userInput)  // Safe formatting

// ✅ BETTER: Use structured API for user data
logger.Info("User input received", iris.Str("input", userInput))
```

### Input Sanitization
Sugar APIs automatically sanitize output through IRIS's encoder, but format strings themselves should never contain user input.

## 🔄 Migration Examples

### From Standard Log Package
```go
// Before (standard log)
log.Printf("User %s logged in at %v", username, time.Now())

// After (IRIS Sugar)
logger.Infof("User %s logged in at %v", username, time.Now())
```

### From Logrus
```go
// Before (logrus)
logrus.Infof("Processing %d items in %v", count, duration)

// After (IRIS Sugar)
logger.Infof("Processing %d items in %v", count, duration)
```

### From Zap Sugar
```go
// Before (zap sugar)
sugar.Infof("Failed to fetch URL: %s", url)

// After (IRIS Sugar)
logger.Infof("Failed to fetch URL: %s", url)
```

## 🎯 Best Practices

### When to Use Sugar APIs
✅ **Good for:**
- Rapid prototyping and development
- Simple logging without complex structure
- Migrating from printf-style loggers
- Console/debug output
- Low-frequency logging

❌ **Avoid for:**
- High-performance production systems
- Structured log analysis requirements
- Logs that need machine parsing
- High-frequency logging (>10K msgs/sec)

### Hybrid Approach
```go
// Use structured API for high-frequency operational logs
logger.Info("Request processed",
    iris.Str("method", "GET"),
    iris.Str("path", "/api/users"),
    iris.Int("status", 200),
    iris.Duration("duration", elapsed),
)

// Use sugar API for occasional debug/error messages
logger.Debugf("Cache miss for key %q, fetching from database", cacheKey)
logger.Errorf("Unexpected error in %s: %v", functionName, err)
```

### Format String Guidelines
```go
// ✅ Good: Clear, readable format strings
logger.Infof("User %s (ID: %d) performed action %s", username, userID, action)

// ❌ Bad: Unclear format strings
logger.Infof("%s: %d, %s", username, userID, action)

// ✅ Good: Consistent formatting
logger.Infof("Database query completed in %v (rows: %d)", duration, rowCount)

// ❌ Bad: Inconsistent formatting  
logger.Infof("DB took %v returned %d", duration, rowCount)
```

## 🔧 Advanced Usage

### Conditional Logging with Sugar
```go
// Efficient: Check level before expensive operations
if logger.Level() <= iris.Debug {
    expensiveData := generateDebugInfo()
    logger.Debugf("Debug info: %+v", expensiveData)
}
```

### Error Wrapping with Sugar
```go
func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        logger.Errorf("Failed to open file %s: %v", filename, err)
        return fmt.Errorf("file processing failed: %w", err)
    }
    defer file.Close()
    
    logger.Infof("Successfully opened file %s", filename)
    // ... processing logic
}
```

### Integration with Context
```go
// Combine with context loggers for the best of both worlds
func handleRequest(ctx context.Context, logger *iris.Logger) {
    contextLogger := logger.WithContext(ctx)
    
    // Use structured API for key operational metrics
    contextLogger.Info("Request started",
        iris.Str("endpoint", "/api/users"),
        iris.Time("start_time", time.Now()),
    )
    
    // Use sugar API for debugging and errors
    if err := validateRequest(); err != nil {
        logger.Errorf("Request validation failed: %v", err)
        return
    }
    
    logger.Debugf("Processing %d items", len(items))
}
```

## 📝 Return Values

All sugar methods return `bool`:
- `true`: Message was successfully logged or filtered out
- `false`: Message was dropped due to buffer overflow

```go
if !logger.Errorf("Critical error: %v", err) {
    // Message was dropped - buffer might be full
    // Consider using direct I/O for critical errors
    fmt.Fprintf(os.Stderr, "CRITICAL: Logger buffer full, error: %v\n", err)
}
```

## 🔗 Related Documentation

- [IRIS Core API](../README.md) - Main library documentation
- [Performance Guide](BENCHMARK_NOTES.md) - Performance characteristics
- [Security Reference](SECURITY_REFERENCE.md) - Security best practices
- [Field Types](../README.md#-field-types) - Structured logging fields

---

**💡 Pro Tip**: Use sugar APIs for development and debugging, structured APIs for production and high-performance scenarios. IRIS makes it easy to mix and match both approaches in the same application.
