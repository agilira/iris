# Provider Ecosystem Architecture

## Overview

The Iris provider ecosystem enables integration with external logging libraries through a modular architecture. This design separates the core Iris library from logging system integrations, allowing for independent development and maintenance of provider modules.

## Architecture Principles

### Core Library Isolation

The core Iris library maintains zero dependencies on external logging frameworks:

- Provides only the SyncReader interface definition
- Implements ReaderLogger for processing external log sources
- Maintains performance and feature guarantees
- Reduces maintenance burden and dependency conflicts

### Provider Module Independence

Each provider exists as a separate Go module:

- Independent versioning and release cycles
- Separate maintenance responsibilities
- Modular installation (`go get` only required providers)
- Community-driven development model

## Module Structure

### Core Library (`github.com/agilira/iris`)

```
iris/
├── sink.go              # SyncReader interface + ReaderLogger
├── iris.go              # Core logger implementation
├── field.go             # Field type system
├── level.go             # Level definitions
└── docs/
    ├── SYNCREADER_INTERFACE.md
    └── PROVIDER_ECOSYSTEM.md
```

### Provider Modules

Each provider follows the naming convention `iris-provider-{library}`:

```
iris-provider-slog/
├── go.mod               # module github.com/agilira/iris-provider-slog
├── provider.go          # SyncReader implementation
├── provider_test.go     # Unit and integration tests
├── README.md            # Provider-specific documentation
└── examples/            # Usage examples
```

## Provider Implementation Requirements

### Module Configuration

```go
// go.mod
module github.com/agilira/iris-provider-{library}

go 1.21

require github.com/agilira/iris v{version}
require {external-library} v{version}
```

### Package Structure

```go
// Provider package name should be descriptive
package slogprovider

import (
    "github.com/agilira/iris"
    "{external-library}"
)

// Primary type implements iris.SyncReader
type Provider struct {
    // Implementation-specific fields
}

// Constructor function
func New(config Config) *Provider {
    // Implementation
}
```

### Required Methods

All providers must implement:

1. **Constructor**: `New()` function with appropriate configuration
2. **SyncReader Interface**: `Read()` and `Close()` methods
3. **External Integration**: Methods required by external library interfaces

### Testing Requirements

Each provider module must include:

- Unit tests for all public methods
- Integration tests with Iris ReaderLogger
- Benchmark tests for performance measurement
- Example code demonstrating usage
- Documentation covering configuration options

## Supported Libraries

### Current Implementations

| Library | Module | Status |
|---------|--------|--------|
| log/slog | iris-provider-slog | Reference Implementation |

### Planned Implementations

| Library | Module | Priority |
|---------|--------|----------|
| logrus | iris-provider-logrus | High |
| zap | iris-provider-zap | High |
| log15 | iris-provider-log15 | Medium |
| apex/log | iris-provider-apex | Medium |

## Usage Patterns

### Basic Integration

```go
import (
    "github.com/agilira/iris"
    provider "github.com/agilira/iris-provider-{library}"
)

func main() {
    // Create provider instance
    p := provider.New(provider.Config{
        BufferSize: 1000,
    })
    defer p.Close()

    // Create Iris logger with provider
    readers := []iris.SyncReader{p}
    logger, err := iris.NewReaderLogger(iris.Config{
        Output:  iris.WrapWriter(os.Stdout),
        Encoder: iris.NewJSONEncoder(),
    }, readers)
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Close()

    logger.Start()

    // Use external library normally
    externalLogger := external.New(p) // p implements external interface
    externalLogger.Info("Message")    // Processed by Iris pipeline
}
```

### Advanced Configuration

```go
// Multi-provider setup
providers := []iris.SyncReader{
    slogprovider.New(slogprovider.Config{BufferSize: 1000}),
    logrusProvider.New(logrusProvider.Config{BufferSize: 500}),
}

logger, err := iris.NewReaderLogger(iris.Config{
    Output: iris.MultiWriter(
        iris.WrapWriter(os.Stdout),
        iris.NewLokiWriter(lokiConfig),
    ),
    Encoder: iris.NewJSONEncoder(),
}, providers,
    iris.WithOTel(),
    iris.WithCaller(),
)
```

## Performance Considerations

### Buffer Sizing

Provider buffer sizes should be configured based on:

- Expected log volume
- Processing latency requirements
- Memory constraints
- Backpressure handling strategy

### Conversion Efficiency

Record conversion should minimize:

- Memory allocations
- String manipulations
- Reflection usage
- Complex data structure traversals

### Backpressure Handling

Providers should implement appropriate backpressure strategies:

- Drop oldest records when buffer is full
- Implement configurable drop policies
- Provide metrics for dropped records
- Avoid blocking source logging operations

## Development Guidelines

### Provider Creation Process

1. **Repository Setup**: Create new repository with standard structure
2. **Interface Implementation**: Implement SyncReader interface
3. **External Integration**: Implement required external library interfaces
4. **Testing Suite**: Develop comprehensive test coverage
5. **Documentation**: Create README and usage examples
6. **Performance Validation**: Benchmark against direct library usage

### Code Quality Requirements

- Follow standard Go conventions and formatting
- Implement comprehensive error handling
- Provide detailed documentation comments
- Include usage examples in documentation
- Maintain test coverage above 90%

### Community Guidelines

- Accept community contributions via pull requests
- Maintain backward compatibility within major versions
- Provide migration guides for breaking changes
- Respond to issues and feature requests promptly
- Follow semantic versioning for releases

## Maintenance Model

### Core Library Responsibilities

- Maintain SyncReader interface stability
- Provide ReaderLogger implementation
- Document interface specifications
- Review provider ecosystem health

### Provider Responsibilities

- Implement and maintain SyncReader interface
- Ensure compatibility with external library updates
- Provide user support and documentation
- Maintain performance and reliability standards

### Community Responsibilities

- Report issues and provide feedback
- Contribute improvements and bug fixes
- Develop additional provider implementations
- Share usage patterns and best practices

---

Iris • an AGILira fragment