# Provider Development Guide

## Overview

This guide provides technical specifications and implementation guidance for developing SyncReader providers for external logging libraries. Providers enable integration between external logging systems and Iris's high-performance logging pipeline.

## Development Requirements

### Prerequisites

- Go 1.21 or later
- Understanding of target logging library
- Familiarity with Go concurrency patterns
- Knowledge of context handling and resource management

### Dependencies

```go
// Minimum required dependencies
require (
    github.com/agilira/iris v{latest}
    {target-logging-library} v{version}
)
```

## Implementation Specification

### Module Structure

```
iris-provider-{library}/
├── go.mod                    # Module definition
├── provider.go               # Main implementation
├── provider_test.go          # Unit tests
├── integration_test.go       # Integration tests
├── benchmark_test.go         # Performance tests
├── README.md                 # Documentation
└── examples/
    └── basic_usage.go        # Usage example
```

### Core Implementation

#### Provider Structure

```go
package {library}provider

type Provider struct {
    // Buffer for incoming records
    records chan {ExternalRecord}
    
    // Shutdown signaling
    closed  chan struct{}
    once    sync.Once
    
    // Configuration
    config Config
    
    // External library integration fields
    // (implementation-specific)
}

type Config struct {
    BufferSize int
    // Additional configuration fields
}
```

#### Constructor Implementation

```go
func New(config Config) *Provider {
    // Validate configuration
    if config.BufferSize <= 0 {
        config.BufferSize = 1000 // Default buffer size
    }
    
    return &Provider{
        records: make(chan {ExternalRecord}, config.BufferSize),
        closed:  make(chan struct{}),
        config:  config,
    }
}
```

#### SyncReader Interface Implementation

```go
func (p *Provider) Read(ctx context.Context) (*iris.Record, error) {
    select {
    case record := <-p.records:
        return p.convertRecord(record), nil
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-p.closed:
        return nil, nil
    }
}

func (p *Provider) Close() error {
    p.once.Do(func() {
        close(p.closed)
        // Additional cleanup if required
    })
    return nil
}
```

#### External Library Integration

The provider must implement the interface required by the external logging library:

```go
// Example for slog.Handler interface
func (p *Provider) Handle(ctx context.Context, record slog.Record) error {
    select {
    case p.records <- record:
        return nil
    case <-p.closed:
        return fmt.Errorf("provider closed")
    default:
        // Buffer full - drop record to prevent blocking
        return nil
    }
}

func (p *Provider) Enabled(ctx context.Context, level slog.Level) bool {
    return true // Let Iris handle level filtering
}

// Additional interface methods as required
```

### Record Conversion

#### Conversion Function

```go
func (p *Provider) convertRecord(external {ExternalRecord}) *iris.Record {
    // Create Iris record
    record := iris.NewRecord(
        p.convertLevel(external.Level),
        external.Message,
    )
    
    // Convert fields
    for _, field := range external.Fields {
        irisField := p.convertField(field)
        if !record.AddField(irisField) {
            // Field limit reached (32 fields)
            break
        }
    }
    
    return record
}
```

#### Level Mapping

```go
func (p *Provider) convertLevel(external {ExternalLevel}) iris.Level {
    switch external {
    case {ExternalDebug}:
        return iris.Debug
    case {ExternalInfo}:
        return iris.Info
    case {ExternalWarn}:
        return iris.Warn
    case {ExternalError}, {ExternalFatal}:
        return iris.Error
    default:
        return iris.Info
    }
}
```

#### Field Conversion

```go
func (p *Provider) convertField(external {ExternalField}) iris.Field {
    key := external.Key
    value := external.Value
    
    switch v := value.(type) {
    case string:
        return iris.String(key, v)
    case int:
        return iris.Int64(key, int64(v))
    case int64:
        return iris.Int64(key, v)
    case float64:
        return iris.Float64(key, v)
    case bool:
        return iris.Bool(key, v)
    case time.Time:
        return iris.Time(key, v)
    case time.Duration:
        return iris.Dur(key, v)
    default:
        // Fallback to string representation
        return iris.String(key, fmt.Sprintf("%v", v))
    }
}
```

## Testing Requirements

### Unit Tests

```go
func TestProvider_Read(t *testing.T) {
    provider := New(Config{BufferSize: 10})
    defer provider.Close()
    
    // Test normal operation
    ctx := context.Background()
    
    // Add test record
    testRecord := {ExternalRecord}{
        Level:   {ExternalInfo},
        Message: "test message",
    }
    provider.records <- testRecord
    
    // Read and verify
    record, err := provider.Read(ctx)
    assert.NoError(t, err)
    assert.Equal(t, iris.Info, record.Level)
    assert.Equal(t, "test message", record.Msg)
}

func TestProvider_ContextCancellation(t *testing.T) {
    provider := New(Config{BufferSize: 10})
    defer provider.Close()
    
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately
    
    record, err := provider.Read(ctx)
    assert.Nil(t, record)
    assert.Equal(t, context.Canceled, err)
}

func TestProvider_Close(t *testing.T) {
    provider := New(Config{BufferSize: 10})
    
    err := provider.Close()
    assert.NoError(t, err)
    
    // Subsequent reads should return nil
    ctx := context.Background()
    record, err := provider.Read(ctx)
    assert.Nil(t, record)
    assert.Nil(t, err)
}
```

### Integration Tests

```go
func TestProvider_Integration(t *testing.T) {
    provider := New(Config{BufferSize: 100})
    defer provider.Close()
    
    // Create ReaderLogger
    var buf bytes.Buffer
    logger, err := iris.NewReaderLogger(iris.Config{
        Output:  iris.WrapWriter(&buf),
        Encoder: iris.NewJSONEncoder(),
        Level:   iris.Debug,
    }, []iris.SyncReader{provider})
    require.NoError(t, err)
    defer logger.Close()
    
    logger.Start()
    
    // Use external library
    externalLogger := {ExternalLibrary}.New(provider)
    externalLogger.Info("integration test", "key", "value")
    
    // Wait for processing
    time.Sleep(100 * time.Millisecond)
    logger.Sync()
    
    // Verify output
    output := buf.String()
    assert.Contains(t, output, "integration test")
    assert.Contains(t, output, "key")
    assert.Contains(t, output, "value")
}
```

### Benchmark Tests

```go
func BenchmarkProvider_Handle(b *testing.B) {
    provider := New(Config{BufferSize: 10000})
    defer provider.Close()
    
    record := {ExternalRecord}{
        Level:   {ExternalInfo},
        Message: "benchmark message",
    }
    
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        provider.Handle(ctx, record)
    }
}

func BenchmarkProvider_Conversion(b *testing.B) {
    provider := New(Config{BufferSize: 1})
    defer provider.Close()
    
    record := {ExternalRecord}{
        Level:   {ExternalInfo},
        Message: "benchmark message",
        Fields: []Field{
            {Key: "string", Value: "value"},
            {Key: "int", Value: 123},
            {Key: "bool", Value: true},
        },
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        provider.convertRecord(record)
    }
}
```

## Performance Guidelines

### Buffer Sizing

- **Small Applications**: 100-1000 records
- **Medium Applications**: 1000-10000 records
- **High Volume Applications**: 10000+ records
- **Memory Constraint**: Record size × buffer size < available memory

### Optimization Techniques

1. **Avoid Allocations**: Reuse objects where possible
2. **Efficient Conversion**: Use type switches instead of reflection
3. **Minimal Processing**: Defer complex operations to Iris pipeline
4. **Buffer Management**: Drop records rather than block when buffer is full

### Performance Targets

| Operation | Target Performance |
|-----------|-------------------|
| Record ingestion | < 100 ns/op |
| Record conversion | < 1000 ns/op |
| Overall throughput | 10x+ improvement over direct library |

## Documentation Requirements

### README.md Structure

```markdown
# Iris Provider: {Library}

## Installation
go get github.com/agilira/iris-provider-{library}

## Usage
[Basic usage example]

## Configuration
[Configuration options]

## Performance
[Performance characteristics]

## License
[License information]
```

### Code Documentation

- Document all exported types and functions
- Include usage examples in documentation comments
- Explain configuration options and their impact
- Document performance characteristics and limitations

## Release Process

### Versioning

Follow semantic versioning:
- **Major**: Breaking changes to public API
- **Minor**: New features, backward compatible
- **Patch**: Bug fixes, backward compatible

### Release Checklist

1. Update version in documentation
2. Run full test suite
3. Update CHANGELOG.md
4. Tag release in git
5. Publish to pkg.go.dev
6. Update provider ecosystem documentation

## Support and Maintenance

### Issue Handling

- Respond to issues within 48 hours
- Provide clear reproduction steps for bug reports
- Label issues appropriately (bug, enhancement, question)
- Maintain issue templates for consistent reporting

### Security

- Monitor for security vulnerabilities in dependencies
- Follow responsible disclosure for security issues
- Keep dependencies updated to latest secure versions
- Document security considerations in README

### Community

- Welcome community contributions via pull requests
- Maintain contributor guidelines
- Provide clear development setup instructions
- Acknowledge contributors in release notes

---

Iris • an AGILira fragment