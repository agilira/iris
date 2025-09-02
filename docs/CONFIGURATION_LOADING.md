# Configuration Loading Guide

## Overview

Iris provides flexible configuration loading from multiple sources including JSON files and environment variables. This enables external configuration management for production deployments without requiring code changes or rebuilds.

## Key Features

- **Multi-Source Loading**: JSON files, environment variables, and programmatic defaults
- **Precedence System**: Environment variables override file configuration
- **Zero Dependencies**: Uses only Go standard library
- **Validation**: Built-in configuration validation with helpful error messages
- **Production Ready**: Designed for containerized and cloud-native deployments

## Quick Start

### JSON Configuration

```go
// Create config.json
{
  "level": "info",
  "format": "json",
  "output": "stdout",
  "capacity": 8192,
  "batch_size": 32,
  "enable_caller": true
}

// Load configuration
config, err := iris.LoadConfigFromJSON("config.json")
if err != nil {
    log.Fatal(err)
}

logger, err := iris.New(config)
if err != nil {
    log.Fatal(err)
}
```

### Environment Variables

```bash
# Set environment variables
export IRIS_LEVEL=debug
export IRIS_CAPACITY=16384
export IRIS_BATCH_SIZE=64
export IRIS_ENABLE_CALLER=true
```

```go
// Load from environment
config, err := iris.LoadConfigFromEnv()
if err != nil {
    log.Fatal(err)
}

logger, err := iris.New(config)
if err != nil {
    log.Fatal(err)
}
```

### Multi-Source Configuration

```go
// Load with precedence: Environment > JSON > Defaults
config, err := iris.LoadConfigMultiSource("config.json")
if err != nil {
    log.Fatal(err)
}

logger, err := iris.New(config)
if err != nil {
    log.Fatal(err)
}
```

## Performance Characteristics

### Benchmarks

```
BenchmarkLoadConfigFromJSON-8    85632    13979 ns/op    1440 B/op    15 allocs/op
```

### Performance Model

- **JSON Loading**: ~14μs per load operation
- **Environment Loading**: Sub-microsecond for standard variables
- **Startup Only**: Zero runtime overhead after configuration loading
- **Memory**: Minimal allocations, garbage collected after loading

## Configuration Format

### JSON Schema

```json
{
  "level": "debug|info|warn|error|panic|fatal",
  "format": "json|text",
  "output": "stdout|stderr|<file_path>",
  "capacity": 8192,
  "batch_size": 32,
  "enable_caller": true,
  "name": "logger_name"
}
```

### Environment Variables

| Environment Variable | JSON Field | Type | Description |
|---------------------|------------|------|-------------|
| `IRIS_LEVEL` | `level` | string | Log level (debug, info, warn, error, panic, fatal) |
| `IRIS_FORMAT` | `format` | string | Output format (json, text) |
| `IRIS_OUTPUT` | `output` | string | Output destination (stdout, stderr, file path) |
| `IRIS_CAPACITY` | `capacity` | int | Ring buffer capacity |
| `IRIS_BATCH_SIZE` | `batch_size` | int | Batch processing size |
| `IRIS_ENABLE_CALLER` | `enable_caller` | bool | Enable caller information |
| `IRIS_NAME` | `name` | string | Logger name |

### Configuration Precedence

1. **Environment Variables** (highest priority)
2. **JSON File Configuration**
3. **Default Values** (lowest priority)

```go
// Example: JSON sets level=info, ENV sets IRIS_LEVEL=debug
// Result: level=debug (environment wins)
config, err := iris.LoadConfigMultiSource("config.json")
// config.Level will be Debug due to environment override
```

## Advanced Usage

### Custom Output Destinations

```json
{
  "level": "info",
  "output": "/var/log/app.log",
  "format": "json"
}
```

```bash
# Environment variable for file output
export IRIS_OUTPUT=/var/log/app.log
```

### Production Configuration Example

```json
{
  "level": "info",
  "format": "json",
  "output": "stdout",
  "capacity": 65536,
  "batch_size": 64,
  "enable_caller": false,
  "name": "production-service"
}
```

### Development Configuration Example

```json
{
  "level": "debug",
  "format": "text",
  "output": "stdout",
  "capacity": 1024,
  "batch_size": 8,
  "enable_caller": true,
  "name": "dev-service"
}
```

### Container-Friendly Configuration

```dockerfile
# Dockerfile
ENV IRIS_LEVEL=info
ENV IRIS_FORMAT=json
ENV IRIS_OUTPUT=stdout
ENV IRIS_CAPACITY=32768
```

```go
// Application code - automatically picks up container env vars
config, err := iris.LoadConfigFromEnv()
if err != nil {
    log.Fatal(err)
}
```

## Configuration Validation

### Built-in Validation

```go
config, err := iris.LoadConfigFromJSON("config.json")
if err != nil {
    log.Fatal(err) // JSON parsing or file errors
}

// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatal(err) // Configuration validation errors
}
```

### Validation Rules

- **Capacity**: Must be positive, preferably power of 2
- **BatchSize**: Must be positive, should be ≤ Capacity
- **Level**: Must be valid log level
- **Output**: File paths must be writable
- **Format**: Must be supported format

### Error Handling

```go
config, err := iris.LoadConfigFromJSON("nonexistent.json")
if err != nil {
    // Handle file not found, permission errors, invalid JSON
    switch {
    case os.IsNotExist(err):
        log.Println("Config file not found, using defaults")
        config = iris.DefaultConfig()
    case strings.Contains(err.Error(), "invalid JSON"):
        log.Fatal("Configuration file contains invalid JSON")
    default:
        log.Fatal("Failed to load configuration:", err)
    }
}
```

## Configuration Templates

### Microservice Template

```json
{
  "level": "${ENVIRONMENT:info}",
  "format": "json",
  "output": "stdout",
  "capacity": 32768,
  "batch_size": 64,
  "enable_caller": false,
  "name": "${SERVICE_NAME:microservice}"
}
```

### High-Throughput Template

```json
{
  "level": "warn",
  "format": "json",
  "output": "stdout",
  "capacity": 131072,
  "batch_size": 256,
  "enable_caller": false,
  "name": "high-throughput-service"
}
```

### Debug Template

```json
{
  "level": "debug",
  "format": "text",
  "output": "stdout",
  "capacity": 1024,
  "batch_size": 1,
  "enable_caller": true,
  "name": "debug-service"
}
```

## Integration Patterns

### 12-Factor App Compliance

```go
// config/config.go
func Load() (*iris.Config, error) {
    // Try environment first (12-factor approach)
    if config, err := iris.LoadConfigFromEnv(); err == nil {
        return config, nil
    }
    
    // Fallback to config file
    if config, err := iris.LoadConfigFromJSON("config.json"); err == nil {
        return config, nil
    }
    
    // Use defaults as last resort
    config := iris.DefaultConfig()
    return &config, nil
}
```

### Kubernetes ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iris-config
data:
  config.json: |
    {
      "level": "info",
      "format": "json",
      "output": "stdout",
      "capacity": 32768,
      "batch_size": 64,
      "enable_caller": false
    }
```

```yaml
# deployment.yaml
spec:
  containers:
  - name: app
    env:
    - name: IRIS_LEVEL
      value: "info"
    volumeMounts:
    - name: config
      mountPath: /etc/iris
  volumes:
  - name: config
    configMap:
      name: iris-config
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    image: myapp:latest
    environment:
      - IRIS_LEVEL=info
      - IRIS_FORMAT=json
      - IRIS_OUTPUT=stdout
      - IRIS_CAPACITY=32768
    volumes:
      - ./config.json:/app/config.json:ro
```

## Hot Configuration Reload

### File Watcher Pattern

```go
import (
    "github.com/fsnotify/fsnotify"
)

func WatchConfig(logger *iris.Logger, configPath string) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        logger.Error("Failed to create config watcher", iris.Err(err))
        return
    }
    defer watcher.Close()

    err = watcher.Add(configPath)
    if err != nil {
        logger.Error("Failed to watch config file", iris.Err(err))
        return
    }

    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                logger.Info("Config file modified, reloading...")
                
                newConfig, err := iris.LoadConfigFromJSON(configPath)
                if err != nil {
                    logger.Error("Failed to reload config", iris.Err(err))
                    continue
                }
                
                // Apply new configuration
                logger.SetLevel(newConfig.Level)
                logger.Info("Configuration reloaded successfully")
            }
        case err := <-watcher.Errors:
            logger.Error("Config watcher error", iris.Err(err))
        }
    }
}
```

### Signal-Based Reload

```go
import (
    "os"
    "os/signal"
    "syscall"
)

func SetupConfigReload(logger *iris.Logger, configPath string) {
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGUSR1)

    go func() {
        for range sigs {
            logger.Info("Received SIGUSR1, reloading configuration...")
            
            newConfig, err := iris.LoadConfigFromJSON(configPath)
            if err != nil {
                logger.Error("Failed to reload config", iris.Err(err))
                continue
            }
            
            // Apply new configuration
            logger.SetLevel(newConfig.Level)
            logger.Info("Configuration reloaded successfully")
        }
    }()
}

// Usage: kill -USR1 <pid> to reload config
```

## Configuration Security

### Sensitive Data Handling

```json
{
  "level": "info",
  "output": "/var/log/app.log",
  "db_password": "DO_NOT_LOG_THIS"
}
```

```go
// Secure configuration loading
type SecureConfig struct {
    iris.Config
    DBPassword string `json:"db_password"`
}

func LoadSecureConfig(path string) (*SecureConfig, error) {
    var config SecureConfig
    
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    // Clear sensitive data from memory
    config.DBPassword = ""
    
    return &config, nil
}
```

### Environment Variable Security

```bash
# Use secure environment management
export IRIS_OUTPUT_FILE="/secure/logs/app.log"
# Avoid: export DB_PASSWORD="secret" # Visible in process list

# Better: Use mounted secrets
cat /run/secrets/db_password | read DB_PASSWORD
export IRIS_LEVEL=info
```

## Troubleshooting

### Configuration Not Loading

1. **Check File Permissions**: Ensure config file is readable
2. **Verify JSON Syntax**: Use JSON validator
3. **Environment Variables**: Check variable names and values
4. **Path Issues**: Use absolute paths or verify working directory

### Common Errors

```go
// File not found
config, err := iris.LoadConfigFromJSON("missing.json")
// Error: open missing.json: no such file or directory

// Invalid JSON
config, err := iris.LoadConfigFromJSON("invalid.json")
// Error: invalid character '}' looking for beginning of object key string

// Invalid level
config := &iris.Config{Level: "invalid"}
err := config.Validate()
// Error: invalid log level: invalid
```

### Debug Configuration Loading

```go
func DebugConfigLoad(path string) {
    fmt.Printf("Loading config from: %s\n", path)
    
    // Check file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        fmt.Printf("Config file does not exist\n")
        return
    }
    
    // Read raw content
    data, err := os.ReadFile(path)
    if err != nil {
        fmt.Printf("Failed to read file: %v\n", err)
        return
    }
    
    fmt.Printf("Raw config content:\n%s\n", string(data))
    
    // Try to load
    config, err := iris.LoadConfigFromJSON(path)
    if err != nil {
        fmt.Printf("Failed to load config: %v\n", err)
        return
    }
    
    fmt.Printf("Loaded config: %+v\n", config)
}
```

## Best Practices

### 1. Environment-Specific Configurations

```
configs/
  ├── development.json
  ├── staging.json
  └── production.json
```

```go
func LoadEnvironmentConfig() (*iris.Config, error) {
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }
    
    configPath := fmt.Sprintf("configs/%s.json", env)
    return iris.LoadConfigFromJSON(configPath)
}
```

### 2. Configuration Validation

```go
func LoadAndValidateConfig(path string) (*iris.Config, error) {
    config, err := iris.LoadConfigFromJSON(path)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return config, nil
}
```

### 3. Graceful Degradation

```go
func LoadConfigWithFallback() *iris.Config {
    // Try multi-source loading
    if config, err := iris.LoadConfigMultiSource("config.json"); err == nil {
        return config
    }
    
    // Try environment only
    if config, err := iris.LoadConfigFromEnv(); err == nil {
        return config
    }
    
    // Use safe defaults
    config := iris.DefaultConfig()
    config.Level = iris.Info
    config.Format = "json"
    
    return &config
}
```

### 4. Configuration Documentation

```go
// Document configuration in code
type ApplicationConfig struct {
    // Logging configuration
    Log iris.Config `json:"log"`
    
    // Server configuration  
    Server struct {
        Port int    `json:"port"`    // Server listening port
        Host string `json:"host"`    // Server bind address
    } `json:"server"`
}
```

## Configuration Integration

### From Hardcoded Configuration

```go
// Before: Hardcoded configuration
logger, _ := iris.New(iris.Config{
    Level:     iris.Info,
    Output:    os.Stdout,
    Capacity:  8192,
    BatchSize: 32,
})

// After: External configuration
config, err := iris.LoadConfigFromJSON("config.json")
if err != nil {
    log.Fatal(err)
}
logger, err := iris.New(config)
if err != nil {
    log.Fatal(err)
}
```

### From Other Configuration Libraries

```go
// Viper equivalent
// viper.SetConfigName("config")
// viper.ReadInConfig()

// IRIS equivalent
config, err := iris.LoadConfigFromJSON("config.json")
```

## Compatibility

- **Go Version**: Requires Go 1.18+
- **File Formats**: JSON (YAML and TOML support planned)
- **Environment**: POSIX-compliant environment variable handling
- **Containers**: Docker, Podman, Kubernetes compatible
- **Cloud**: AWS, GCP, Azure compatible

## Configuration Examples Repository

Complete configuration examples are available in the repository:

- `examples/configs/microservice.json` - Microservice configuration
- `examples/configs/high-throughput.json` - High-performance configuration  
- `examples/configs/development.json` - Development configuration
- `examples/configs/production.json` - Production configuration
- `examples/docker/` - Docker and container examples
- `examples/kubernetes/` - Kubernetes manifests
