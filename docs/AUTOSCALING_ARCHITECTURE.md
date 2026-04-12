# Iris Auto-Scaling Architecture

## Overview

Iris implements a lazy dual-mode auto-scaling logger that automatically transitions between SingleRing and MPSC modes based on real-time goroutine contention. The system starts in the fastest mode (SingleRing) and lazily initializes the multi-producer path only when needed.

Design philosophy:
- Present: Optimal for agentic OS (single producer, low latency)
- Future: Ready for robotics (multi-producer burst scenarios)

## Architecture

### Types and Constants

```go
// ScalingMode represents the current scaling mode
type ScalingMode uint32

const (
    SingleMode ScalingMode = iota // ~25ns/op single-producer
    MultiMode                     // ~35ns/op multi-producer
)
```

### ScalerConfig

```go
type ScalerConfig struct {
    GoroutineThreshold uint32        // Concurrent goroutines to trigger scale-up
    ScaleDownCooldown  time.Duration // Cooldown before scaling back down
    BaseConfig         Config        // Base config for logger creation
    Options            []Option      // Options for logger creation
}
```

### AdaptiveLogger

```go
type AdaptiveLogger struct {
    mode          atomic.Uint32            // Current ScalingMode
    singleLogger  *Logger                  // Always exists
    multiLogger   atomic.Pointer[Logger]   // Lazy-initialized via sync.Once
    activeWriters atomic.Uint32            // Contention detection
    lastMultiUse  atomic.Int64             // Timestamp for scale-down decision
    config        ScalerConfig
    // ... lifecycle (ctx, cancel, wg) and stats fields
}
```

### Scaling Modes

| Mode | Performance | Best For | Initialization |
|------|-------------|----------|----------------|
| **SingleMode** | ~25ns/op | Low contention, single producers | Always active |
| **MultiMode** | ~35ns/op per thread | High contention, multiple goroutines | Lazy (sync.Once) |

## Scaling Behavior

**Scale up to MultiMode when:**
- Active concurrent writers >= `GoroutineThreshold` (default: 2)
- Multi-producer logger is lazily created on first scale-up via `sync.Once`

**Scale down to SingleMode when:**
- No multi-producer activity for `ScaleDownCooldown` (default: 5s)
- Checked periodically by a background monitor goroutine

Zero log loss during transitions: both loggers remain active, mode switch is atomic.

## Configuration

### Default Production Configuration

```go
cfg := iris.Config{
    Output:  iris.WrapWriter(os.Stdout),
    Encoder: iris.NewJSONEncoder(),
}

scalerCfg := iris.DefaultScalerConfig(cfg)
// GoroutineThreshold: 2       -- scale up when 2+ goroutines write concurrently
// ScaleDownCooldown:  5s      -- stay in multi-mode for 5s after last concurrent use
```

### Custom Configuration

```go
scalerCfg := iris.ScalerConfig{
    GoroutineThreshold: 4,                    // More conservative scale-up
    ScaleDownCooldown:  10 * time.Second,     // Longer cooldown
    BaseConfig:         cfg,
    Options:            opts,
}
```

## Usage

### Creating and Using an AdaptiveLogger

```go
package main

import (
    "os"
    "github.com/agilira/iris"
)

func main() {
    cfg := iris.Config{
        Output:  iris.WrapWriter(os.Stdout),
        Encoder: iris.NewJSONEncoder(),
        Level:   iris.Info,
    }
    scalerCfg := iris.DefaultScalerConfig(cfg)

    al, err := iris.NewAdaptiveLogger(scalerCfg)
    if err != nil {
        panic(err)
    }

    if err := al.Start(); err != nil {
        panic(err)
    }
    defer al.Close()

    // Use normally -- auto-scaling is transparent
    al.Info("Single producer message", iris.Str("key", "value"))
}
```

### Multi-Goroutine Usage

```go
func highContentionLogging(al *iris.AdaptiveLogger) {
    var wg sync.WaitGroup

    // Launch multiple goroutines (triggers scale-up to MultiMode)
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 1000; j++ {
                al.Info("High contention message",
                    iris.Int("goroutine", id),
                    iris.Int("message", j),
                )
            }
        }(i)
    }

    wg.Wait()

    // Check scaling results
    stats := al.Stats()
    fmt.Printf("Mode: %s, Scale-ups: %d, Total writes: %d\n",
        stats.Mode, stats.ScaleUpCount, stats.TotalWrites)
}
```

## Monitoring

### AdaptiveStats

```go
stats := al.Stats()

type AdaptiveStats struct {
    Mode           ScalingMode // Current scaling mode
    ScaleUpCount   uint64      // Total scale-up operations
    ScaleDownCount uint64      // Total scale-down operations
    TotalWrites    uint64      // Total log writes
    ActiveWriters  uint32      // Current active concurrent writers
}
```

### Available Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `Stats()` | `AdaptiveStats` | Scaling statistics snapshot |
| `Mode()` | `ScalingMode` | Current mode (SingleMode or MultiMode) |
| `Start()` | `error` | Begins logger and scale-down monitor |
| `Close()` | `error` | Graceful shutdown of all loggers |

### Log Methods

`AdaptiveLogger` exposes `Debug`, `Info`, `Warn`, and `Error` methods with the same signature as `Logger`:

```go
al.Info(msg string, fields ...Field)
al.Debug(msg string, fields ...Field)
al.Warn(msg string, fields ...Field)
al.Error(msg string, fields ...Field)
```

## Internal Mechanics

1. Each log call increments `activeWriters` atomically
2. If `activeWriters >= GoroutineThreshold`, triggers scale-up
3. Multi-producer logger is initialized once via `sync.Once` (capacity: 4096)
4. Single-producer logger uses smaller capacity (1024)
5. Scale-down monitor runs at `ScaleDownCooldown / 2` interval
6. Scale-down occurs when `time.Since(lastMultiUse) > ScaleDownCooldown`

---

Iris -- an AGILira fragment
