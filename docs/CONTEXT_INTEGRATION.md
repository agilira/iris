# Context Integration Guide

## Overview

IRIS provides optimized `context.Context` integration for structured logging in distributed systems. This feature enables automatic extraction and inclusion of context values in log records while maintaining zero-allocation performance characteristics.

## Key Features

- **Performance Optimized**: Pre-extraction strategy avoids O(n) context.Value() calls in hot path
- **Zero Allocations**: Context field extraction produces no allocations
- **Configurable**: Custom context key extraction with field name mapping
- **Thread Safe**: Safe for concurrent use across goroutines
- **Opt-in**: Zero overhead when not used

## Quick Start

### Basic Usage

```go
// Create context with values
ctx := context.Background()
ctx = context.WithValue(ctx, iris.RequestIDKey, "req-12345")
ctx = context.WithValue(ctx, iris.UserIDKey, "user-67890")

// Extract context once, use many times
contextLogger := logger.WithContext(ctx)

// All subsequent logs include context fields automatically
contextLogger.Info("User login", iris.Str("method", "oauth"))
contextLogger.Error("Authentication failed", iris.Str("reason", "invalid_token"))

// Output includes: request_id="req-12345", user_id="user-67890"
```

### Fast Methods for Common Cases

```go
// Optimized single-field extraction
requestLogger := logger.WithRequestID(ctx)
userLogger := logger.WithUserID(ctx)
traceLogger := logger.WithTraceID(ctx)

requestLogger.Info("Processing request")
// Output: {"request_id":"req-12345","msg":"Processing request",...}
```

## Performance Characteristics

### Benchmarks

```
BenchmarkContextValueExtraction-8    28638308    40.81 ns/op    0 B/op    0 allocs/op
```

### Performance Model

- **Context Extraction**: 40.81 ns one-time cost
- **Subsequent Logging**: Zero additional context overhead
- **Memory**: Zero allocations for context field storage

### Optimization Strategy

```go
// ❌ Inefficient: Repeated context.Value() calls
func handleRequest(ctx context.Context, logger *iris.Logger) {
    logger.Info("Start", iris.Str("request_id", ctx.Value("request_id").(string)))
    logger.Info("Process", iris.Str("request_id", ctx.Value("request_id").(string)))
    logger.Info("End", iris.Str("request_id", ctx.Value("request_id").(string)))
}

// ✅ Efficient: Pre-extraction with caching
func handleRequest(ctx context.Context, logger *iris.Logger) {
    contextLogger := logger.WithContext(ctx) // Extract once
    contextLogger.Info("Start")               // Use cached fields
    contextLogger.Info("Process")             // Use cached fields  
    contextLogger.Info("End")                 // Use cached fields
}
```

## Advanced Configuration

### Custom Context Extractor

```go
// Define custom context keys
type CustomKey string
const (
    OrganizationIDKey CustomKey = "org_id"
    TenantIDKey      CustomKey = "tenant_id"
)

// Create custom extractor
extractor := &iris.ContextExtractor{
    Keys: map[iris.ContextKey]string{
        iris.RequestIDKey:                "req_id",      // Rename field
        iris.ContextKey(OrganizationIDKey): "organization", // Custom key
        iris.ContextKey(TenantIDKey):       "tenant",       // Another custom key
    },
    MaxDepth: 10, // Limit context chain traversal
}

// Use custom extractor
contextLogger := logger.WithContextExtractor(ctx, extractor)
contextLogger.Info("Multi-tenant operation")
// Output includes: req_id, organization, tenant
```

### Combining Context with Manual Fields

```go
// Start with context extraction
contextLogger := logger.WithContext(ctx)

// Add additional fields
enrichedLogger := contextLogger.With(
    iris.Str("component", "auth"),
    iris.Str("operation", "login"),
)

enrichedLogger.Info("Authentication attempt")
// Output includes both context fields and manual fields
```

## Standard Context Keys

IRIS provides predefined keys for common use cases:

```go
const (
    RequestIDKey iris.ContextKey = "request_id"  // HTTP/gRPC request ID
    TraceIDKey   iris.ContextKey = "trace_id"    // Distributed tracing
    SpanIDKey    iris.ContextKey = "span_id"     // Tracing span ID
    UserIDKey    iris.ContextKey = "user_id"     // User identification
    SessionIDKey iris.ContextKey = "session_id"  // Session tracking
)
```

## Integration Patterns

### HTTP Middleware

```go
func ContextMiddleware(logger *iris.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract or generate request ID
            requestID := r.Header.Get("X-Request-ID")
            if requestID == "" {
                requestID = generateRequestID()
            }
            
            // Add to context
            ctx := context.WithValue(r.Context(), iris.RequestIDKey, requestID)
            
            // Create context logger
            contextLogger := logger.WithRequestID(ctx)
            
            // Store in request context for handlers
            ctx = context.WithValue(ctx, "logger", contextLogger)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Usage in handler
func MyHandler(w http.ResponseWriter, r *http.Request) {
    logger := r.Context().Value("logger").(*iris.ContextLogger)
    logger.Info("Handler called") // Includes request_id automatically
}
```

### gRPC Interceptor

```go
func ContextUnaryInterceptor(logger *iris.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // Extract metadata
        md, ok := metadata.FromIncomingContext(ctx)
        if ok {
            if reqID := md.Get("request-id"); len(reqID) > 0 {
                ctx = context.WithValue(ctx, iris.RequestIDKey, reqID[0])
            }
        }
        
        // Create context logger
        contextLogger := logger.WithContext(ctx)
        
        // Store in context
        ctx = context.WithValue(ctx, "logger", contextLogger)
        
        return handler(ctx, req)
    }
}
```

### Background Jobs

```go
func ProcessJob(ctx context.Context, logger *iris.Logger, jobID string) {
    // Create job-specific context
    jobCtx := context.WithValue(ctx, iris.ContextKey("job_id"), jobID)
    
    // Extract job context
    jobLogger := logger.WithContextExtractor(jobCtx, &iris.ContextExtractor{
        Keys: map[iris.ContextKey]string{
            iris.ContextKey("job_id"): "job_id",
            iris.UserIDKey:            "user_id",
        },
    })
    
    jobLogger.Info("Job started")
    // Process job...
    jobLogger.Info("Job completed")
}
```

## Best Practices

### 1. Extract Once, Use Many Times

```go
// ✅ Good: One extraction per request/operation
contextLogger := logger.WithContext(ctx)
contextLogger.Info("Operation started")
contextLogger.Debug("Processing step 1")
contextLogger.Debug("Processing step 2")
contextLogger.Info("Operation completed")
```

### 2. Use Fast Methods for Single Fields

```go
// ✅ Good: For single field extraction
requestLogger := logger.WithRequestID(ctx)
requestLogger.Info("Request processed")
```

### 3. Configure Extraction Scope

```go
// ✅ Good: Limit extraction to needed keys only
extractor := &iris.ContextExtractor{
    Keys: map[iris.ContextKey]string{
        iris.RequestIDKey: "request_id",
        iris.UserIDKey:    "user_id",
        // Only extract what you need
    },
    MaxDepth: 5, // Limit context traversal
}
```

### 4. Store Context Loggers in Request Context

```go
// ✅ Good: Avoid repeated extraction
ctx = context.WithValue(ctx, "logger", contextLogger)

// Later in the call chain
logger := ctx.Value("logger").(*iris.ContextLogger)
logger.Info("Deep in call stack") // Uses cached context fields
```

## Error Handling

Context integration is designed to be fault-tolerant:

```go
// Empty context - no errors, no fields extracted
emptyCtx := context.Background()
contextLogger := logger.WithContext(emptyCtx)
contextLogger.Info("Works fine") // Just message, no context fields

// Missing values - silently ignored
ctx := context.WithValue(context.Background(), "other_key", "value")
contextLogger := logger.WithContext(ctx) // No configured keys found
contextLogger.Info("Still works") // No context fields, no errors

// Wrong type values - silently skipped
ctx := context.WithValue(context.Background(), iris.RequestIDKey, 12345) // Not string
contextLogger := logger.WithContext(ctx) // Skips non-string value
contextLogger.Info("Type safe") // No request_id field
```

## Migration Guide

### From Manual Context Handling

```go
// Before: Manual context value extraction
func logWithContext(ctx context.Context, logger *iris.Logger, msg string) {
    fields := []iris.Field{}
    if reqID, ok := ctx.Value("request_id").(string); ok {
        fields = append(fields, iris.Str("request_id", reqID))
    }
    if userID, ok := ctx.Value("user_id").(string); ok {
        fields = append(fields, iris.Str("user_id", userID))
    }
    logger.Info(msg, fields...)
}

// After: Automatic context integration
func logWithContext(ctx context.Context, logger *iris.Logger, msg string) {
    contextLogger := logger.WithContext(ctx)
    contextLogger.Info(msg) // Context fields included automatically
}
```

### From Other Logging Libraries

```go
// Other library pattern
log.WithContext(ctx).WithField("key", "value").Info("message")

// IRIS equivalent
logger.WithContext(ctx).Info("message", iris.Str("key", "value"))

// Or combined approach
contextLogger := logger.WithContext(ctx)
enrichedLogger := contextLogger.With(iris.Str("key", "value"))
enrichedLogger.Info("message")
```

## Troubleshooting

### Context Values Not Appearing

1. **Check Key Types**: Ensure context keys match extraction configuration
2. **Verify Value Types**: Only string values are extracted by default
3. **Confirm Context Chain**: Values must be in the context passed to WithContext()

### Performance Issues

1. **Avoid Repeated Extraction**: Extract context once per operation
2. **Limit Extraction Scope**: Configure only needed keys
3. **Use Fast Methods**: For single-field cases, use WithRequestID() etc.

### Debug Context Extraction

```go
// Enable debug logging to see extraction results
extractor := &iris.ContextExtractor{
    Keys: map[iris.ContextKey]string{
        iris.RequestIDKey: "request_id",
    },
}

contextLogger := logger.WithContextExtractor(ctx, extractor)
// Check if fields were extracted by logging immediately
contextLogger.Debug("Context extraction test")
```

## Compatibility

- **Go Version**: Requires Go 1.18+
- **Context Package**: Compatible with standard library context
- **Thread Safety**: Safe for concurrent use
- **Zero Dependencies**: No external dependencies beyond standard library

## Performance Comparison

| Approach | Cost per Log | Allocations | Suitable For |
|----------|--------------|-------------|--------------|
| Manual ctx.Value() | ~100-200ns | 0-1 | Single logs |
| Context Pre-extraction | ~40ns once | 0 | Multiple logs |
| No Context | ~0ns | 0 | High frequency |

Choose the approach based on your logging pattern and performance requirements.
