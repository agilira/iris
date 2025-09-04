# Iris Architecture

## Overview

Iris is an ultra-high performance structured logging library for Go, designed to be fast and secure while maintaining zero allocations and providing comprehensive features for production environments. The architecture is built around the embedded Zephyros MPSC ring buffer system, advanced field serialization, and intelligent performance optimizations. Iris challenges the performance of industry-standard loggers while offering advanced usability and feature completeness.

## System Architecture

### Design Principles

1. **Zero-Allocation Hot Paths**: No memory allocations during critical logging operations
2. **Lock-Free MPSC Ring Buffer**: Multi-producer, single-consumer architecture for maximum throughput
3. **Adaptive Performance**: Single-ring for maximum single-thread performance, threaded-rings for multi-producer scaling
4. **Type-Safe Structured Logging**: Compile-time type safety with runtime flexibility
5. **Encoder Abstraction**: Pluggable output formats (JSON, console, future binary formats)
6. **Security by Design**: Built-in protection against log injection and sensitive data exposure
7. **Production-Ready Features**: Comprehensive configuration, monitoring, and operational capabilities

## Module Structure

### Core Components (`iris.go`)

- `Logger`: Main logging interface with structured field support and contextual logging
- `Config`: Comprehensive configuration with intelligent defaults and validation
- `Level`: Atomic log level management with runtime adjustability
- `Field`: Type-safe structured field system with union-based storage

### Ring Buffer System (`ring.go`, `internal/zephyroslite/`)

Embedded Zephyros MPSC ring buffer providing ultra-low latency logging with two architectural modes:
- **SingleRing**: Optimized for single-producer scenarios (~25ns/op)
- **ThreadedRings**: Multi-producer scaling with excellent concurrency (~35ns/op)

### Field System (`field.go`)

Comprehensive type system supporting:
- Primitive types (string, int64, uint64, float64, bool)
- Temporal types (time.Time, time.Duration)
- Complex types ([]byte, error, fmt.Stringer, arbitrary objects)
- Security features (redacted secrets, safe error handling)

### Encoder Ecosystem (`encoder-*.go`)

- **JSONEncoder**: High-performance NDJSON with intelligent time caching
- **ConsoleEncoder**: Human-readable output with optional ANSI colors
- **Extensible Interface**: Support for future binary and custom formats

### Buffer Pool System (`internal/bufferpool/`)

Zero-allocation buffer management with intelligent sizing and reuse strategies.

## Detailed Architecture Diagram

> **Diagram Legend:**
> - **Solid arrows (→)**: Primary data flow and synchronous operations
> - **Dashed arrows (-.)**: Asynchronous processing and optional paths
> - **Color coding**: Functional layers with performance characteristics
> - **Performance metrics**: Actual benchmarks from production testing

```mermaid
graph TB
    subgraph "Application Layer"
        APP[Application Code<br/>Business Logic]
        CTX[Context Handling<br/>Request Tracing]
        MDW[Middleware<br/>Framework Integration]
    end

    subgraph "Iris Public API"
        LOGGER[Logger Interface<br/>Info/Debug/Warn/Error<br/>Structured Fields]
        WITH[With() Context<br/>Field Inheritance<br/>Named Loggers]
        LEVEL[Level Management<br/>Atomic Operations<br/>Runtime Adjustment]
        SMART[Smart API<br/>Automatic Configuration<br/>Convention over Config]
    end

    subgraph "Field System"
        FIELD[Field Construction<br/>Type-Safe Builders<br/>Union Storage]
        TYPES[Type System<br/>15 Supported Types<br/>Compile-Time Safety]
        VALIDATE[Validation<br/>Bounds Checking<br/>Security Filtering]
        SECRET[Security Features<br/>Secret Redaction<br/>Injection Protection]
    end

    subgraph "Configuration System"
        CONFIG[Config Struct<br/>Intelligent Defaults<br/>Validation Rules]
        LOADER[Config Loader<br/>JSON File Support<br/>Hot Reload Ready]
        ARCH[Architecture Choice<br/>SingleRing vs ThreadedRings<br/>Performance Optimization]
        BACKP[Backpressure Policy<br/>DropOnFull vs BlockOnFull<br/>Performance vs Reliability]
    end

    subgraph "Core Processing Pipeline"
        RECORD[Record Creation<br/>Field Assembly<br/>Level Checking<br/>~27-30ns logging path]
        SAMPLING[Sampling System<br/>Noise Reduction<br/>Rate Limiting]
        ENCODER[Encoder Selection<br/>Format Abstraction<br/>Pluggable System]
        BUFFER[Buffer Pool<br/>Zero Allocation<br/>Intelligent Sizing]
    end

    subgraph "Zephyros MPSC Ring Buffer"
        subgraph "SingleRing Architecture"
            SR[Single Ring<br/>~25ns/op<br/>Maximum Single-Thread Performance]
            SRCONS[Single Consumer<br/>Dedicated Goroutine<br/>Batch Processing]
        end
        
        subgraph "ThreadedRings Architecture"
            TR[Multiple Rings<br/>~35ns/op per thread<br/>Excellent Scaling]
            TRCONS[Multiple Consumers<br/>Load Balancing<br/>Adaptive Batching]
        end
        
        POLICY[Backpressure<br/>DropOnFull: ~0.7ns disabled<br/>BlockOnFull: guaranteed delivery]
        IDLE[Idle Strategies<br/>Spin, Yield, Sleep<br/>CPU vs Latency Trade-off]
    end

    subgraph "Encoder Implementations"
        JSON[JSON Encoder<br/>NDJSON Format<br/>Time Caching<br/>Character Escaping]
        CONSOLE[Console Encoder<br/>Human Readable<br/>ANSI Colors<br/>Key=Value Format]
        FUTURE[Future Encoders<br/>Binary Formats<br/>Custom Protocols<br/>Compression]
    end

    subgraph "Output Layer"
        STDOUT[Standard Output<br/>os.Stdout/Stderr<br/>Terminal Output]
        FILES[File Output<br/>Rotation Ready<br/>Append Mode]
        WRITERS[Custom Writers<br/>Network Sinks<br/>Monitoring Systems]
        DISCARD[Discard Output<br/>Benchmarking<br/>Testing]
    end

    subgraph "Time System"
        TIMECACHE[Time Cache<br/>go-timecache Integration<br/>Sub-microsecond Performance]
        TIMEFORMAT[Time Formatting<br/>RFC3339 vs Unix<br/>Precision Control]
    end

    subgraph "Performance Optimizations"
        ATOMIC[Atomic Operations<br/>Lock-Free Level Checks<br/>Memory Barriers]
        CACHE[Cache Optimization<br/>CPU Cache-Line Alignment<br/>False Sharing Prevention]
        PREFETCH[Memory Prefetch<br/>Predictable Access Patterns<br/>SIMD Ready]
        POOL[Resource Pooling<br/>Buffer Reuse<br/>GC Pressure Reduction]
    end

    subgraph "Monitoring & Observability"
        METRICS[Performance Metrics<br/>Throughput Monitoring<br/>Latency Histograms]
        HEALTH[Health Checks<br/>Ring Buffer Status<br/>Consumer Health]
        STATS[Statistics<br/>Message Counts<br/>Error Rates]
    end

    subgraph "Security Layer"
        INJECTION[Log Injection Protection<br/>Character Escaping<br/>Format Validation]
        REDACTION[Data Redaction<br/>Secret Fields<br/>PII Protection]
        AUDIT[Audit Capabilities<br/>Structured Logging<br/>Compliance Ready]
    end

    %% Application to API Flow
    APP --> LOGGER
    CTX --> WITH
    MDW --> LOGGER

    %% API to Field System
    LOGGER --> FIELD
    WITH --> FIELD
    FIELD --> TYPES
    TYPES --> VALIDATE
    VALIDATE --> SECRET

    %% Configuration Flow
    CONFIG --> ARCH
    CONFIG --> BACKP
    LOADER --> CONFIG
    ARCH -.-> SR
    ARCH -.-> TR
    BACKP --> POLICY

    %% Core Processing
    LOGGER --> RECORD
    FIELD --> RECORD
    RECORD --> SAMPLING
    SAMPLING --> ENCODER
    ENCODER --> BUFFER

    %% Ring Buffer Processing
    BUFFER --> SR
    BUFFER --> TR
    SR --> SRCONS
    TR --> TRCONS
    POLICY --> SR
    POLICY --> TR
    IDLE --> SRCONS
    IDLE --> TRCONS

    %% Encoder Selection
    ENCODER --> JSON
    ENCODER --> CONSOLE
    ENCODER -.-> FUTURE

    %% Output Flow
    JSON --> STDOUT
    JSON --> FILES
    JSON --> WRITERS
    JSON --> DISCARD
    CONSOLE --> STDOUT
    CONSOLE --> FILES

    %% Time Integration
    JSON --> TIMECACHE
    CONSOLE --> TIMECACHE
    TIMECACHE --> TIMEFORMAT

    %% Performance Integration
    RECORD --> ATOMIC
    SR --> CACHE
    TR --> CACHE
    BUFFER --> PREFETCH
    ENCODER --> POOL

    %% Monitoring Integration
    SRCONS --> METRICS
    TRCONS --> METRICS
    LOGGER --> HEALTH
    RECORD --> STATS

    %% Security Integration
    JSON --> INJECTION
    CONSOLE --> INJECTION
    SECRET --> REDACTION
    LOGGER --> AUDIT

    %% Async Processing Paths
    SRCONS -.-> JSON
    SRCONS -.-> CONSOLE
    TRCONS -.-> JSON
    TRCONS -.-> CONSOLE

    %% Configuration Dependencies
    CONFIG --> BUFFER
    CONFIG --> SAMPLING
    CONFIG --> TIMECACHE

    %% Smart API Integration
    SMART --> CONFIG
    SMART --> LOGGER

    %% Styling with semantic colors
    classDef application fill:#f8fafc,stroke:#1e293b,stroke-width:2px
    classDef api fill:#f0f9ff,stroke:#0369a1,stroke-width:2px
    classDef field fill:#f3e8ff,stroke:#7c3aed,stroke-width:2px
    classDef config fill:#fef3c7,stroke:#d97706,stroke-width:2px
    classDef processing fill:#ecfdf5,stroke:#059669,stroke-width:2px
    classDef ringbuffer fill:#eff6ff,stroke:#2563eb,stroke-width:2px
    classDef encoder fill:#f0fdf4,stroke:#16a34a,stroke-width:2px
    classDef output fill:#fefce8,stroke:#ca8a04,stroke-width:2px
    classDef time fill:#fcfcfc,stroke:#6b7280,stroke-width:2px
    classDef performance fill:#ecfeff,stroke:#0891b2,stroke-width:2px
    classDef monitoring fill:#fdf4ff,stroke:#a855f7,stroke-width:2px
    classDef security fill:#fef2f2,stroke:#dc2626,stroke-width:2px

    class APP,CTX,MDW application
    class LOGGER,WITH,LEVEL,SMART api
    class FIELD,TYPES,VALIDATE,SECRET field
    class CONFIG,LOADER,ARCH,BACKP config
    class RECORD,SAMPLING,ENCODER,BUFFER processing
    class SR,TR,SRCONS,TRCONS,POLICY,IDLE ringbuffer
    class JSON,CONSOLE,FUTURE encoder
    class STDOUT,FILES,WRITERS,DISCARD output
    class TIMECACHE,TIMEFORMAT time
    class ATOMIC,CACHE,PREFETCH,POOL performance
    class METRICS,HEALTH,STATS monitoring
    class INJECTION,REDACTION,AUDIT security
```

## Data Flow Summary

The architecture diagram illustrates the complete data flow through Iris components:

1. **Application Integration**: Seamless integration with application code, middleware, and context handling
2. **Type-Safe API**: Structured field construction with compile-time type safety and runtime validation
3. **Configuration Flexibility**: Intelligent defaults with comprehensive customization options
4. **High-Performance Processing**: Zero-allocation record creation with atomic level checking (~27-30ns logging path)
5. **Adaptive Ring Buffer**: Choice between single-ring (25ns) and threaded-rings (35ns) architectures
6. **Encoder Abstraction**: Pluggable output formats with format-specific optimizations
7. **Security Integration**: Built-in protection against injection attacks and sensitive data exposure

## Concurrency Model

### Single-Ring Architecture (Default)

```go
// Optimized for single-producer or low-contention scenarios
type SingleRingLogger struct {
    ring     *Ring              // Single MPSC ring buffer
    consumer *ConsumerGoroutine // Dedicated processing thread
    level    int64              // Atomic level checking
}
```

**Characteristics:**
- **~25ns/op** logging performance
- **Single consumer goroutine** for deterministic processing
- **Lock-free writes** with atomic sequence numbers
- **Optimal for**: Benchmarks, single-threaded applications, maximum single-thread throughput

### Threaded-Rings Architecture

```go
// Optimized for high-concurrency multi-producer scenarios
type ThreadedRingsLogger struct {
    rings     []*Ring           // Multiple rings for load distribution
    consumers []*Consumer       // Multiple consumer goroutines
    selector  RingSelector      // Load balancing strategy
}
```

**Characteristics:**
- **~35ns/op** per producer thread
- **Excellent scaling** with multiple producers (4x+ improvement)
- **Load balancing** across multiple rings
- **Optimal for**: Production servers, high-concurrency applications, multi-core scaling

## Adaptive Architecture

### Auto-Scaling Intelligence

Iris implements intelligent architecture adaptation through the AutoScaling Logger system, which dynamically transitions between SingleRing and ThreadedRings modes based on real-time performance metrics. This architectural adaptivity ensures optimal performance across varying workload patterns without manual intervention.

### Scaling Triggers

The auto-scaling system monitors key performance indicators and triggers architectural transitions based on:

#### Scale-to-MPSC Conditions
- **Write Frequency**: ≥1000 writes/second sustained load
- **Contention Ratio**: ≥10% failed write attempts due to buffer contention
- **Latency Threshold**: ≥1ms average logging latency indicating bottlenecks
- **Concurrency Level**: ≥3 active goroutines performing concurrent writes

#### Scale-to-Single Conditions  
- **Reduced Load**: <100 writes/second indicating low-contention scenarios
- **Low Contention**: <1% contention ratio showing minimal resource competition
- **Latency Optimization**: <100μs average latency enabling single-ring efficiency

### Architectural Decision Framework

The scaling decisions use stability requirements (3 consecutive measurements) and cooldown periods (1 second) to prevent oscillation while ensuring responsive adaptation to workload changes. The system maintains zero log loss during architectural transitions through atomic mode switching.

**Technical Implementation**: For complete technical details on metrics collection, scaling algorithms, and transition mechanisms, see [`docs/AUTOSCALING_ARCHITECTURE.md`](AUTOSCALING_ARCHITECTURE.md).

### Smart API Automatic Configuration

Iris Smart API provides intelligent auto-detection for optimal initial configuration:

```go
// Smart API automatically detects optimal settings
logger, err := iris.New(iris.Config{})

// Automatic detection includes:
// - Architecture: SingleRing (CPU < 4) vs ThreadedRings (CPU >= 4)
// - Capacity: Intelligent sizing based on CPU cores and available memory
// - Encoder: JSON for production, Console for development
// - Output: os.Stdout with intelligent defaults
// - Levels: Info level for production environments
```

**Smart Detection Logic:**
- **Architecture**: CPU-based selection (ThreadedRings for ≥4 cores, SingleRing otherwise)
- **Capacity**: Intelligent sizing based on available CPU cores
- **Environment**: Automatic detection of development vs production
- **Zero Configuration**: Optimal defaults without manual tuning

**Note**: The Smart API provides static auto-detection at startup. For dynamic runtime scaling between architectures based on real-time workload patterns, use the dedicated AutoScaling Logger system that continuously monitors and automatically transitions between SingleRing and MPSC modes.

## Field Type System

### Supported Types

Iris provides comprehensive type support with zero-allocation field construction:

| Category | Types | Storage | Performance |
|----------|-------|---------|-------------|
| **Primitives** | `string`, `int64`, `uint64`, `float64`, `bool` | Direct union fields | Zero allocation |
| **Temporal** | `time.Time`, `time.Duration` | Unix nanoseconds | Cache-optimized |
| **Binary** | `[]byte` | Direct slice reference | Zero-copy |
| **Complex** | `error`, `fmt.Stringer`, `interface{}` | Interface storage | Minimal allocation |
| **Security** | `Secret` (redacted strings) | Redacted output | Security-first |

### Field Construction Architecture

The field system uses union-based storage with type-safe builders, enabling zero-allocation field construction through pre-compiled type information and direct memory layout optimization.

## Encoder Architecture

### JSON Encoder

High-performance NDJSON (Newline Delimited JSON) encoder with advanced optimizations:

**Features:**
- **Time caching**: Integration with go-timecache for sub-microsecond timestamp formatting
- **Character escaping**: Comprehensive protection against log injection attacks
- **Streaming output**: Direct buffer writing without intermediate allocations
- **Format flexibility**: RFC3339 timestamps or Unix nanoseconds

### Console Encoder

Human-readable encoder optimized for development and debugging:

**Features:**
- **ANSI colors**: Optional color coding by log level
- **Key-value format**: Easy-to-read field representation
- **Smart quoting**: Automatic quoting only when necessary
- **Compact output**: Efficient space utilization

## Security Architecture

### Log Injection Protection

Comprehensive protection against log injection attacks through automatic character escaping, control character filtering, and Unicode normalization in all encoder implementations.

### Sensitive Data Protection

Built-in support for sensitive data redaction through `Secret` field types that automatically mask sensitive information in all output formats.

### Audit-Ready Logging

Structured logging format suitable for compliance and audit requirements:

- **Immutable records**: Log entries cannot be modified after creation
- **Timestamp precision**: Nanosecond-precision timestamps for accurate ordering
- **Structured fields**: Machine-readable format for automated analysis
- **Error context**: Comprehensive error information with stack traces when enabled

## Buffer Pool System

### Zero-Allocation Buffer Management

Intelligent buffer pooling system using sync.Pool with automatic sizing based on actual log entry patterns. The system includes performance monitoring, size limits for memory protection, and graceful degradation strategies.

**Architectural Features:**
- **Automatic sizing**: Buffers grow based on actual log entry sizes
- **Size limits**: Protection against memory bloat from oversized entries
- **Performance monitoring**: Statistics for optimization and debugging
- **Graceful degradation**: Fallback allocation when pool is exhausted

## Performance Optimization Strategies

### Atomic Operations

Lock-free implementation for critical paths:

```go
// Atomic log level checking (sub-nanosecond fast path)
func (l *Logger) isEnabled(level Level) bool {
    return atomic.LoadInt64(&l.level) <= int64(level)
}

// Fast disabled logging path (~0.7ns when disabled)
if !l.isEnabled(Debug) {
    return // Immediate return, no allocations
}
```

### Memory Layout Optimization

Cache-line aligned structures prevent false sharing in multi-core environments. Critical atomic fields are positioned to avoid cache line conflicts, with strategic padding to optimize memory access patterns.

### Time Optimization

Integration with go-timecache provides sub-microsecond timestamp performance through pre-computed time string caching, reducing time formatting overhead by 121x compared to standard time operations.

## Testing Architecture

### Test Strategy

- **Unit Tests**: Individual component validation with edge case coverage
- **Benchmark Tests**: Performance validation against regression baselines
- **Integration Tests**: End-to-end workflow validation with real encoders
- **Stress Tests**: High-concurrency and memory pressure validation
- **Security Tests**: Log injection and data redaction validation

### Performance Validation

Continuous benchmarking ensures performance characteristics remain within specified tolerances. Target performance maintains sub-30ns logging operations with zero allocations in hot paths.

## OpenTelemetry Integration

### Architectural Design

Iris integrates with OpenTelemetry through a dedicated `otel` package that leverages the existing `ContextExtractor` architecture without modifying core components. The integration maintains zero-allocation performance by using context value propagation and field extraction patterns.

### Integration Architecture

```
OpenTelemetry Context → ContextExtractor → Field Injection → Log Output
     ↓                       ↓                   ↓              ↓
Span/Baggage Data      Context Values      Automatic Fields   JSON/Console
```

### Technical Components

#### Context Extraction Layer
- **`ExtractOTelContext()`**: Extracts trace information into context values
- **`ContextExtractor`**: Reuses Iris's existing context field extraction system
- **Field Mapping**: Automatic mapping of OpenTelemetry data to log fields

#### Field Categories
- **Trace Fields**: `trace_id`, `span_id`, `trace_sampled`
- **Baggage Fields**: `baggage.*` (distributed context propagation)
- **Resource Fields**: `service.*` (service identification)

#### Performance Characteristics
- **Zero-Allocation**: Uses existing ContextExtractor pattern
- **Minimal Overhead**: Context value extraction without reflection
- **Reusable Components**: Context loggers can be cached within trace spans


---

Iris • an AGILira fragment
