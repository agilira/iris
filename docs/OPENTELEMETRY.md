# OpenTelemetry Integration

## Overview

Iris provides seamless OpenTelemetry integration for distributed tracing and observability in microservices architectures. The integration automatically correlates logs with traces, propagates baggage across service boundaries, and detects resource information without requiring modifications to the core Iris logger.

## Core Features

- **Automatic Trace Correlation**: Log entries include trace_id and span_id automatically
- **Baggage Propagation**: Distributed context propagation across service boundaries  
- **Resource Detection**: Automatic service name, version, and environment detection
- **Zero-Allocation Path**: Leverages Iris's ContextExtractor for optimal performance
- **Non-Intrusive Design**: Separate module that doesn't modify core Iris functionality

## Installation

```bash
go get github.com/agilira/iris/otel
```

## Basic Usage

### Quick Start

```go
import (
    "context"
    "github.com/agilira/iris"
    "github.com/agilira/iris/otel"
    "go.opentelemetry.io/otel"
)

// Create standard Iris logger
logger, _ := iris.New(iris.Config{})
logger.Start()
defer logger.Close()

// Get OpenTelemetry context from your tracer
tracer := otel.Tracer("my-service")
ctx, span := tracer.Start(context.Background(), "operation")
defer span.End()

// Create OpenTelemetry-aware logger
otelLogger := otel.WithTracing(logger, ctx)

// Log with automatic trace correlation
otelLogger.Info("Processing request",
    iris.Str("user_id", "12345"),
    iris.Int("items", 42),
)
```

Output includes trace correlation automatically:
```json
{
  "ts": "2025-09-02T10:30:45.123Z",
  "level": "info", 
  "msg": "Processing request",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "trace_sampled": "true",
  "user_id": "12345",
  "items": 42
}
```

### Advanced Usage with Baggage

```go
import (
    "go.opentelemetry.io/otel/baggage"
)

// Add correlation context
member, _ := baggage.NewMember("request.id", "req-123")
userTier, _ := baggage.NewMember("user.tier", "premium") 
bag, _ := baggage.New(member, userTier)
ctx = baggage.ContextWithBaggage(ctx, bag)

// Create logger with baggage propagation
otelLogger := otel.WithTracing(logger, ctx)

otelLogger.Info("User action completed")
```

Output includes baggage fields:
```json
{
  "ts": "2025-09-02T10:30:45.123Z",
  "level": "info",
  "msg": "User action completed", 
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "baggage.request.id": "req-123",
  "baggage.user.tier": "premium"
}
```

## API Reference

### WithTracing

Creates an OpenTelemetry-aware logger that automatically includes trace correlation and baggage propagation.

```go
func WithTracing(logger *iris.ContextLogger, ctx context.Context) *iris.ContextLogger
```

**Parameters:**
- `logger`: Base Iris logger instance
- `ctx`: Context containing OpenTelemetry span and baggage information

**Returns:** ContextLogger with automatic OpenTelemetry field extraction

### ExtractOTelContext

Extracts OpenTelemetry information into a context.Context for use with standard Iris context methods.

```go
func ExtractOTelContext(ctx context.Context) context.Context
```

**Parameters:**
- `ctx`: Source context containing OpenTelemetry information

**Returns:** Enhanced context.Context with OpenTelemetry fields

**Usage:**
```go
enrichedCtx := otel.ExtractOTelContext(ctx)
contextLogger := logger.WithContext(enrichedCtx)
contextLogger.Info("Message with trace correlation")
```

### WithBaggage

Creates a logger with baggage propagation from the provided context.

```go
func WithBaggage(logger *iris.ContextLogger, ctx context.Context) *iris.ContextLogger
```

**Parameters:**
- `logger`: Base Iris logger instance  
- `ctx`: Context containing OpenTelemetry baggage

**Returns:** ContextLogger with automatic baggage field extraction

## Field Extraction

The OpenTelemetry integration automatically extracts and includes the following fields:

### Trace Fields
- `trace_id`: Unique identifier for the distributed trace
- `span_id`: Unique identifier for the current span
- `trace_sampled`: Indicates if the trace is sampled ("true"/"false")

### Baggage Fields
- `baggage.<key>`: All baggage members are prefixed with "baggage."

### Resource Fields
- `service.name`: Service name from OpenTelemetry resource detection
- `service.version`: Service version when available
- `service.environment`: Deployment environment when available

## Performance Characteristics

The OpenTelemetry integration maintains Iris's zero-allocation logging performance:

- **Field Extraction**: Uses Iris's ContextExtractor pattern for zero-allocation field retrieval
- **Context Overhead**: Minimal overhead for context propagation 
- **Memory Usage**: No additional memory allocations during logging operations
- **Throughput**: Maintains Iris's high-throughput characteristics

## Integration Patterns

### Microservices Pattern

```go
// Service A: Initiating request
func handleRequest(w http.ResponseWriter, r *http.Request) {
    tracer := otel.Tracer("service-a")
    ctx, span := tracer.Start(r.Context(), "handle_request")
    defer span.End()
    
    otelLogger := otel.WithTracing(logger, ctx)
    otelLogger.Info("Request received", iris.Str("path", r.URL.Path))
    
    // Call Service B with propagated context
    callServiceB(ctx)
}

// Service B: Receiving propagated trace
func processData(ctx context.Context) {
    tracer := otel.Tracer("service-b") 
    ctx, span := tracer.Start(ctx, "process_data")
    defer span.End()
    
    otelLogger := otel.WithTracing(logger, ctx)
    otelLogger.Info("Processing data") // Same trace_id as Service A
}
```

### Middleware Pattern

```go
func otelMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tracer := otel.Tracer("http-server")
        ctx, span := tracer.Start(r.Context(), "http_request")
        defer span.End()
        
        // Add request context to logger
        otelLogger := otel.WithTracing(logger, ctx)
        
        // Store in request context for downstream handlers
        ctx = context.WithValue(ctx, "logger", otelLogger)
        r = r.WithContext(ctx)
        
        next.ServeHTTP(w, r)
    })
}
```

### Error Correlation Pattern

```go
func processWithErrorTracking(ctx context.Context) error {
    tracer := otel.Tracer("processor")
    ctx, span := tracer.Start(ctx, "complex_operation")
    defer span.End()
    
    otelLogger := otel.WithTracing(logger, ctx)
    
    if err := validateInput(); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        
        otelLogger.Error("Validation failed",
            iris.Err(err),
            iris.Str("input_type", "user_data"),
        )
        return err
    }
    
    otelLogger.Info("Operation completed successfully")
    return nil
}
```

## Best Practices

### Context Propagation

- Always propagate context.Context across service boundaries
- Use OpenTelemetry's HTTP and gRPC instrumentation for automatic context propagation
- Pass OpenTelemetry context to all logging operations for trace correlation

### Baggage Usage

- Use baggage for cross-cutting concerns like correlation IDs, user tiers, or feature flags
- Keep baggage payload small to minimize network overhead
- Avoid sensitive data in baggage as it's transmitted in HTTP headers

### Resource Configuration

- Configure OpenTelemetry resource detection with service.name, service.version
- Set deployment environment through OTEL_RESOURCE_ATTRIBUTES or resource builders
- Use consistent service naming across your infrastructure

### Performance Optimization

- Reuse ContextLogger instances when possible within the same trace context
- Avoid creating new OpenTelemetry-aware loggers for each log statement
- Use Iris's With() method to create child loggers with additional context

## Configuration

### Environment Variables

Standard OpenTelemetry environment variables work seamlessly:

```bash
export OTEL_SERVICE_NAME="my-service"
export OTEL_SERVICE_VERSION="1.2.3" 
export OTEL_RESOURCE_ATTRIBUTES="environment=production,region=us-west-2"
export OTEL_EXPORTER_OTLP_ENDPOINT="http://jaeger:14268/api/traces"
```

### Programmatic Configuration

```go
import (
    "go.opentelemetry.io/otel/sdk/resource"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Configure resource detection
res, err := resource.New(
    context.Background(),
    resource.WithAttributes(
        semconv.ServiceName("my-service"),
        semconv.ServiceVersion("1.2.3"),
        attribute.String("environment", "production"),
    ),
)

// Use with tracer provider
tp := trace.NewTracerProvider(
    trace.WithResource(res),
    trace.WithSampler(trace.AlwaysSample()),
)
otel.SetTracerProvider(tp)
```

## Troubleshooting

### Missing Trace Information

**Symptom**: Logs don't include trace_id or span_id

**Solutions**:
- Ensure OpenTelemetry tracer provider is configured
- Verify context.Context contains active span
- Check that span is started before creating OpenTelemetry logger

### Baggage Not Appearing

**Symptom**: Baggage fields missing from log output

**Solutions**:
- Verify baggage is set in context before logger creation
- Check baggage member names for invalid characters
- Ensure context propagation maintains baggage across service calls

### Performance Issues

**Symptom**: Decreased logging performance with OpenTelemetry

**Solutions**:
- Reuse ContextLogger instances within same trace context
- Avoid excessive baggage payload size
- Profile application to identify OpenTelemetry overhead sources

## Examples

See the `otel/` directory for complete integration examples and test cases demonstrating real-world usage patterns.

---

Iris • an AGILira fragment