# Context Integration Guide

## Overview

Iris provides optimized `context.Context` integration for structured logging. Context field extraction is cached at creation time, avoiding repeated `context.Value()` calls in the hot logging path.

## Key Features

- **Performance Optimized**: Pre-extraction strategy avoids O(n) context.Value() calls per log
- **Zero Allocations**: Context field extraction produces no allocations after setup
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

### Extraction Variants

```go
// Default extraction (all 5 standard keys)
ctxLogger := logger.WithContext(ctx)

// Custom extractor
ctxLogger := logger.WithContext(ctx, iris.WithExtractor(myExtractor))

// Specific keys only (more efficient)
ctxLogger := logger.WithContext(ctx, iris.WithKeys(iris.RequestIDKey, iris.TraceIDKey))

// Single key with custom field name
ctxLogger := logger.WithContext(ctx, iris.WithKey(iris.RequestIDKey, "req_id"))
```

## Performance Characteristics

### Benchmarks

```
BenchmarkContextValueExtraction-8    28638308    40.81 ns/op    0 B/op    0 allocs/op
```

### Performance Model

- **Context Extraction**: ~41 ns one-time cost
- **Subsequent Logging**: Zero additional context overhead
- **Memory**: Zero allocations for context field storage

### Optimization Strategy

```go
// Inefficient: Repeated context.Value() calls
func handleRequest(ctx context.Context, logger *iris.Logger) {
    logger.Info("Start", iris.Str("request_id", ctx.Value("request_id").(string)))
    logger.Info("Process", iris.Str("request_id", ctx.Value("request_id").(string)))
    logger.Info("End", iris.Str("request_id", ctx.Value("request_id").(string)))
}

// Efficient: Pre-extraction with caching
func handleRequest(ctx context.Context, logger *iris.Logger) {
    contextLogger := logger.WithContext(ctx) // Extract once
    contextLogger.Info("Start")               // Use cached fields
    contextLogger.Info("Process")             // Use cached fields
    contextLogger.Info("End")                 // Use cached fields
}
```

## Standard Context Keys

Iris provides predefined keys for common use cases:

```go
const (
    RequestIDKey ContextKey = "request_id"  // HTTP/gRPC request ID
    TraceIDKey   ContextKey = "trace_id"    // Distributed tracing
    SpanIDKey    ContextKey = "span_id"     // Tracing span ID
    UserIDKey    ContextKey = "user_id"     // User identification
    SessionIDKey ContextKey = "session_id"  // Session tracking
)
```

The `defaultContextExtractor()` maps all 5 keys to their string equivalents by default.

## Advanced Configuration

### Custom Context Extractor

```go
type CustomKey string
const (
    OrganizationIDKey CustomKey = "org_id"
    TenantIDKey       CustomKey = "tenant_id"
)

extractor := &iris.ContextExtractor{
    Keys: map[iris.ContextKey]string{
        iris.RequestIDKey:                  "req_id",       // Rename field
        iris.ContextKey(OrganizationIDKey): "organization", // Custom key
        iris.ContextKey(TenantIDKey):       "tenant",       // Another custom key
    },
    MaxDepth: 10, // Limit context chain traversal
}

contextLogger := logger.WithContext(ctx, iris.WithExtractor(extractor))
contextLogger.Info("Multi-tenant operation")
// Output includes: req_id, organization, tenant
```

### Combining Context with Manual Fields

```go
contextLogger := logger.WithContext(ctx)

// ContextLogger supports With() to add additional fields
enrichedLogger := contextLogger.With(
    iris.Str("component", "auth"),
    iris.Str("operation", "login"),
)

enrichedLogger.Info("Authentication attempt")
// Output includes both context fields and manual fields
```

## ContextLogger API

`ContextLogger` exposes the same log methods as `Logger`:

```go
contextLogger.Debug(msg string, fields ...Field)
contextLogger.Info(msg string, fields ...Field)
contextLogger.Warn(msg string, fields ...Field)
contextLogger.Error(msg string, fields ...Field)
contextLogger.DPanic(msg string, fields ...Field)
contextLogger.With(fields ...Field) *ContextLogger
```

Each method automatically prepends the pre-extracted context fields to any additional fields passed at the call site.

## Integration Patterns

### HTTP Middleware

```go
func LoggingMiddleware(logger *iris.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestID := r.Header.Get("X-Request-ID")
            if requestID == "" {
                requestID = generateRequestID()
            }

            ctx := context.WithValue(r.Context(), iris.RequestIDKey, requestID)
            contextLogger := logger.WithContext(ctx,
                iris.WithKeys(iris.RequestIDKey))

            // Store in request context for handlers
            ctx = context.WithValue(ctx, "logger", contextLogger)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func MyHandler(w http.ResponseWriter, r *http.Request) {
    logger := r.Context().Value("logger").(*iris.ContextLogger)
    logger.Info("Handler called") // Includes request_id automatically
}
```

### Background Jobs

```go
func ProcessJob(ctx context.Context, logger *iris.Logger, jobID string) {
    jobCtx := context.WithValue(ctx, iris.ContextKey("job_id"), jobID)

    jobLogger := logger.WithContext(jobCtx, iris.WithExtractor(&iris.ContextExtractor{
        Keys: map[iris.ContextKey]string{
            iris.ContextKey("job_id"): "job_id",
            iris.UserIDKey:            "user_id",
        },
    }))

    jobLogger.Info("Job started")
    // ... process job ...
    jobLogger.Info("Job completed")
}
```

## Error Handling

Context integration is designed to be fault-tolerant:

```go
// Empty context -- no errors, no fields extracted
emptyCtx := context.Background()
contextLogger := logger.WithContext(emptyCtx)
contextLogger.Info("Works fine") // Just message, no context fields

// Missing values -- silently ignored
ctx := context.WithValue(context.Background(), "other_key", "value")
contextLogger := logger.WithContext(ctx) // No configured keys found
contextLogger.Info("Still works") // No context fields, no errors

// Wrong type values -- silently skipped (only string values extracted)
ctx := context.WithValue(context.Background(), iris.RequestIDKey, 12345)
contextLogger := logger.WithContext(ctx) // Skips non-string value
contextLogger.Info("Type safe") // No request_id field
```

## Best Practices

1. **Extract once, use many times**: Create ContextLogger per request/operation, not per log call
2. **Use WithKeys() for efficiency**: When you only need 1-2 keys, avoid scanning all 5 defaults
3. **Limit extraction scope**: Configure only keys you actually need
4. **Store in request context**: Avoid re-extraction deep in call chains

## Performance Comparison

| Approach | Cost per Log | Allocations | Suitable For |
|----------|--------------|-------------|--------------|
| Manual ctx.Value() | ~100-200ns | 0-1 | Single logs |
| Context Pre-extraction | ~41ns once | 0 | Multiple logs |
| No Context | ~0ns | 0 | High frequency |

---

Iris -- an AGILira fragment
