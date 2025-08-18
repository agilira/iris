# Iris: Ultra-High Performance Logger
### Built on Xantos Ring Buffer Engine

Iris is an ultra-high performance structured logger built as a wrapper around the Xantos lock-free ring buffer. It provides zap-like performance with the predictability and speed of Xantos' SPSC architecture.

## Features

- **Ultra-Fast**: Built on Xantos' 140M+ ops/sec ring buffer
- **Zero Allocations**: In steady state operations
- **Lock-Free**: Single Producer Single Consumer architecture
- **Structured Logging**: JSON, Console, and custom formats
- **Zero Dependencies**: Only uses AGILira libraries (Xantos + go-errors)
- **Predictable Latency**: ~7-10ns per log operation

## Quick Start

```go
package main

import (
    "github.com/agilira/xantos/iris"
)

func main() {
    logger := iris.New(iris.Config{
        Level:  iris.InfoLevel,
        Output: iris.ConsoleOutput,
        Format: iris.JSONFormat,
    })
    defer logger.Close()
    
    logger.Info("Hello from Iris!")
    logger.With("user", "john").Info("User logged in")
}
```

## Performance

Iris leverages Xantos' extreme performance characteristics:
- **Throughput**: 120M+ log entries/sec
- **Latency**: ~7-10ns per operation
- **Memory**: Zero allocations in steady state
- **Predictability**: Constant latency regardless of load

## Design Philosophy

Iris follows the same philosophy as Xantos: uncompromising performance through disciplined engineering. Built specifically for single-producer scenarios (which covers 99% of logging use cases), Iris eliminates the overhead of multi-producer synchronization while providing enterprise-grade structured logging capabilities.

---

*Iris • an AGILira fragment • Built on Xantos*
