# Iris Config Loader Examples

This directory demonstrates how to use Iris configuration loading system with backpressure policies.

## Files Overview

### Configuration Files (`../configs/`)

- **`high_performance.json`**: Configuration optimized for maximum throughput using `DropOnFull` policy
- **`reliable.json`**: Configuration optimized for guaranteed delivery using `BlockOnFull` policy

### Source Files

- **`main.go`**: Comprehensive demo showing all configuration loading methods
- **`main_test.go`**: Unit tests for configuration loading functionality

## Running the Demo

```bash
# Run the interactive demo
go run main.go

# Run the unit tests  
go test -v
```

## Demo Features

### 1. JSON Configuration Loading
Demonstrates loading logger configuration from JSON files with different backpressure policies.

### 2. Environment Variable Override
Shows how environment variables can override JSON configuration:
```bash
export IRIS_BACKPRESSURE_POLICY="block_on_full"
export IRIS_LEVEL="debug"
export IRIS_NAME="env-logger"
```

### 3. Multi-Source Configuration
Combines JSON base configuration with environment variable overrides following precedence rules.

### 4. Performance Comparison
Benchmarks the performance difference between `DropOnFull` and `BlockOnFull` policies.

## Expected Output

The demo produces output showing:
- ✓ Successful logger creation with each policy
- ✓ Environment variable override behavior
- ✓ Multi-source configuration precedence
- ✓ Performance comparison metrics

## Configuration Options

### Backpressure Policy Values

**In JSON:**
```json
{
  "backpressure_policy": "drop_on_full"  // or "block_on_full"
}
```

**In Environment Variables:**
- `IRIS_BACKPRESSURE_POLICY="drop_on_full"` (or variations: `drop`, `droponful`)
- `IRIS_BACKPRESSURE_POLICY="block_on_full"` (or variations: `block`, `blockonful`)

### Other Supported Configuration

| JSON Field | Environment Variable | Description |
|------------|---------------------|-------------|
| `level` | `IRIS_LEVEL` | Logging level (debug, info, warn, error, panic, fatal) |
| `format` | `IRIS_FORMAT` | Output format (json, text, console) |
| `output` | `IRIS_OUTPUT` | Output destination (stdout, stderr, file path) |
| `capacity` | `IRIS_CAPACITY` | Ring buffer capacity |
| `batch_size` | `IRIS_BATCH_SIZE` | Batch processing size |
| `name` | `IRIS_NAME` | Logger instance name |

## Configuration Precedence

1. **Environment Variables** (highest priority)
2. **JSON File Configuration**
3. **Default Values** (lowest priority)

## Use Cases

### High-Performance Applications
Use `high_performance.json` configuration for:
- Ad servers
- Real-time analytics
- High-frequency trading systems
- Game servers

### Mission-Critical Applications  
Use `reliable.json` configuration for:
- Financial transaction logging
- Audit systems
- Compliance logging
- Medical device logging

## Integration

To integrate config loading in your application:

```go
// Basic JSON loading
config, err := iris.LoadConfigFromJSON("config.json")
if err != nil {
    log.Fatal(err)
}

// Multi-source with environment overrides
config, err := iris.LoadConfigMultiSource("base-config.json")
if err != nil {
    log.Fatal(err)
}

// Create logger with loaded configuration
logger, err := iris.New(*config)
if err != nil {
    log.Fatal(err)
}
```

For more details, see the [Backpressure Policies Documentation](../../docs/BACKPRESSURE_POLICIES.md).
