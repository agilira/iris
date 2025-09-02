# Hot Reload Configuration Management

**Dynamic configuration updates without service restarts**

---

## Overview

Iris provides production-grade hot reload functionality that enables real-time configuration changes without requiring application restarts. This feature is built into the core logging library and provides immediate configuration updates with comprehensive audit trails for production environments.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration File Format](#configuration-file-format)
- [API Reference](#api-reference)
- [Audit Trail System](#audit-trail-system)
- [Error Handling](#error-handling)
- [Performance Characteristics](#performance-characteristics)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)
- [Examples](#examples)

---

## Quick Start

### Basic Setup

```go
package main

import (
    "os"
    "github.com/agilira/iris"
)

func main() {
    // Create logger
    logger, err := iris.New(iris.Config{
        Level:   iris.Info,
        Output:  os.Stdout,
        Encoder: iris.NewJSONEncoder(),
    })
    if err != nil {
        panic(err)
    }
    defer logger.Sync()
    logger.Start()
    defer logger.Close()

    // Enable hot reload
    watcher, err := iris.NewDynamicConfigWatcher("config.json", logger.AtomicLevel())
    if err != nil {
        logger.Error("Failed to enable hot reload", iris.Error(err))
        return
    }
    defer watcher.Stop()

    if err := watcher.Start(); err != nil {
        logger.Error("Failed to start config watcher", iris.Error(err))
        return
    }

    logger.Info("🔥 Hot reload enabled - edit config.json to change log levels!")

    // Your application logic here
    for {
        logger.Debug("Debug message")
        logger.Info("Info message")
        logger.Warn("Warning message")
        logger.Error("Error message")
        time.Sleep(2 * time.Second)
    }
}
```

### Configuration File (`config.json`)

```json
{
    "level": "info",
    "encoder": "json",
    "development": false
}
```

**To change log level:** Simply edit the file and save. Changes are applied within 2 seconds.

---

## Configuration File Format

### Supported Formats

Iris hot reload supports multiple configuration formats with automatic detection:

- **JSON** (`.json`)
- **YAML** (`.yaml`, `.yml`) 
- **TOML** (`.toml`)
- **HCL** (`.hcl`)
- **INI** (`.ini`)
- **Properties** (`.properties`)

### Configuration Schema

#### JSON Format
```json
{
    "level": "debug|info|warn|error|panic|fatal",
    "encoder": "json|text",
    "development": true|false,
    "capacity": 8192,
    "batch_size": 32,
    "enable_caller": true|false,
    "name": "logger-name"
}
```

#### YAML Format
```yaml
level: info
encoder: json
development: false
capacity: 8192
batch_size: 32
enable_caller: true
name: "production-logger"
```

#### TOML Format
```toml
level = "info"
encoder = "json"
development = false
capacity = 8192
batch_size = 32
enable_caller = true
name = "production-logger"
```

### Field Descriptions

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `level` | string | Minimum logging level | `"info"` |
| `encoder` | string | Output format encoder | `"json"` |
| `development` | boolean | Development mode features | `false` |
| `capacity` | integer | Ring buffer capacity | `8192` |
| `batch_size` | integer | Batch processing size | `32` |
| `enable_caller` | boolean | Include caller information | `false` |
| `name` | string | Logger instance name | `""` |

### Log Levels

Available levels in order of increasing severity:

- `debug` - Detailed diagnostic information
- `info` - General informational messages  
- `warn` - Warning conditions
- `error` - Error conditions
- `panic` - Panic conditions (logs then panics)
- `fatal` - Fatal conditions (logs then exits)

---

## API Reference

### DynamicConfigWatcher

#### Constructor

```go
func NewDynamicConfigWatcher(configPath string, atomicLevel *AtomicLevel) (*DynamicConfigWatcher, error)
```

Creates a new configuration watcher for the specified file path.

**Parameters:**
- `configPath` - Path to the configuration file to monitor
- `atomicLevel` - Atomic level instance from the logger (`logger.AtomicLevel()`)

**Returns:**
- `*DynamicConfigWatcher` - Watcher instance
- `error` - Error if file doesn't exist or watcher creation fails

**Example:**
```go
watcher, err := iris.NewDynamicConfigWatcher("app-config.json", logger.AtomicLevel())
if err != nil {
    return fmt.Errorf("failed to create watcher: %w", err)
}
```

#### Methods

##### Start()
```go
func (w *DynamicConfigWatcher) Start() error
```

Begins monitoring the configuration file for changes.

**Returns:**
- `error` - Error if watcher is already started or monitoring fails

##### Stop()
```go
func (w *DynamicConfigWatcher) Stop() error  
```

Stops monitoring the configuration file.

**Returns:**
- `error` - Error if watcher is not running

##### IsRunning()
```go
func (w *DynamicConfigWatcher) IsRunning() bool
```

Returns the current status of the watcher.

**Returns:**
- `bool` - `true` if actively monitoring, `false` otherwise

### AtomicLevel Integration

The hot reload system integrates directly with Iris's atomic level management:

```go
// Get atomic level from logger
atomicLevel := logger.AtomicLevel()

// Create watcher with the same atomic level
watcher, err := iris.NewDynamicConfigWatcher("config.json", atomicLevel)

// Changes to config file will automatically update logger.Level()
```

---

## Audit Trail System

### Overview

Iris provides comprehensive audit logging for all configuration changes, essential for compliance and security monitoring in production environments.

### Audit File Format

Audit entries are written to `iris-config-audit.jsonl` in JSON Lines format:

```json
{"timestamp":"2025-09-01T16:48:21.830+02:00","level":0,"event":"watch_start","component":"iris","file_path":"/app/config.json","process_id":1234,"process_name":"myapp","checksum":"45046fe016ec04459b742a63cfe4b9f10ca0274a18f5d100f2aab2af0738352f"}
{"timestamp":"2025-09-01T16:48:23.832+02:00","level":0,"event":"file_changed","component":"iris","file_path":"/app/config.json","process_id":1234,"process_name":"myapp","checksum":"ca2a0f50800d3747b39b4ba3b773ab7b44f8f27192d92480fd6a2cb78373c42a"}
```

### Audit Fields

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | ISO 8601 timestamp with timezone |
| `level` | integer | Audit severity level |
| `event` | string | Event type (`watch_start`, `file_changed`, `error`) |
| `component` | string | Source component (`iris`) |
| `file_path` | string | Absolute path to configuration file |
| `process_id` | integer | Process ID of the application |
| `process_name` | string | Process name |
| `checksum` | string | SHA-256 hash of file contents |

### Security Features

- **Tamper Detection:** SHA-256 checksums detect unauthorized modifications
- **Immutable Logs:** Append-only audit trail prevents tampering
- **Process Tracking:** Links changes to specific process instances
- **Timestamp Precision:** Microsecond-level timing for forensic analysis

### Audit Configuration

```go
// Audit is automatically enabled with default settings:
// - Output file: "iris-config-audit.jsonl"
// - Buffer size: 1000 entries
// - Flush interval: 5 seconds
// - Minimum level: Info (captures all changes)
```

---

## Error Handling

### Graceful Degradation

Iris hot reload is designed for production resilience:

1. **Invalid Configuration:** Falls back to safe defaults (Info level)
2. **File Access Errors:** Maintains current configuration
3. **Parsing Errors:** Logs error and continues with previous settings
4. **Network Issues:** Local file monitoring continues unaffected

### Error Categories

#### Configuration Errors
```go
// Invalid JSON syntax
{"level": "info",} // trailing comma

// Invalid level value  
{"level": "invalid_level"} // falls back to "info"

// Missing file
// Watcher creation fails with descriptive error
```

#### Runtime Errors
```go
// File permissions changed
// Error logged to audit trail, monitoring continues

// Disk full
// Continues with current configuration until resolved
```

### Error Logging

All errors are logged through Iris's standard error handling:

```go
logger.Error("Config reload failed", 
    iris.String("file", "/app/config.json"),
    iris.Error(err),
    iris.String("fallback", "maintaining current level"))
```

---

## Performance Characteristics

### Monitoring Overhead

- **Polling Interval:** 2 seconds (configurable)
- **CPU Overhead:** 12.10 nanoseconds per poll cycle
- **Memory Footprint:** < 1KB per watched file
- **I/O Operations:** Optimized with file modification time checks

### Performance Benchmarks

```
BenchmarkConfigReload-8     	50000000	12.10 ns/op	0 B/op	0 allocs/op
BenchmarkLevelCheck-8       	2000000000	0.45 ns/op	0 B/op	0 allocs/op
BenchmarkAuditWrite-8       	10000000	145.2 ns/op	48 B/op	1 allocs/op
```

### Scalability

- **Watched Files:** Unlimited (limited by system file handles)
- **Concurrent Watchers:** Thread-safe operation
- **Large Files:** Efficient for configurations up to 10MB
- **High-Frequency Changes:** Rate-limited to prevent resource exhaustion

---

## Production Deployment

### Recommended Configuration

```go
// Production setup with comprehensive monitoring
func setupProductionLogger() (*iris.Logger, *iris.DynamicConfigWatcher, error) {
    // Create logger with production defaults
    logger, err := iris.New(iris.Config{
        Level:    iris.Info,
        Capacity: 32768,  // Large buffer for high throughput
        BatchSize: 64,    // Efficient batching
        Output:   setupLogOutput(), // File or network destination
        Encoder:  iris.NewJSONEncoder(),
        Name:     "production-service",
    })
    if err != nil {
        return nil, nil, err
    }

    logger.Start()

    // Enable hot reload with monitoring
    watcher, err := iris.NewDynamicConfigWatcher("/etc/myapp/config.json", logger.AtomicLevel())
    if err != nil {
        logger.Warn("Hot reload disabled", iris.Error(err))
        return logger, nil, nil // Continue without hot reload
    }

    if err := watcher.Start(); err != nil {
        logger.Warn("Failed to start config watcher", iris.Error(err))
        return logger, nil, nil
    }

    logger.Info("Production logger initialized with hot reload enabled")
    return logger, watcher, nil
}
```

### Security Considerations

1. **File Permissions:** Restrict config file access to application user
2. **Audit Protection:** Secure audit log location with appropriate permissions
3. **Network Security:** Use secure file sharing for distributed configurations
4. **Change Management:** Implement approval workflows for production changes

### Monitoring Integration

```go
// Health check endpoint
func healthCheck(watcher *iris.DynamicConfigWatcher) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        status := map[string]interface{}{
            "hot_reload_enabled": watcher != nil,
            "watcher_running":    watcher != nil && watcher.IsRunning(),
            "config_file":        "/etc/myapp/config.json",
            "last_reload":        getLastReloadTime(),
        }
        
        json.NewEncoder(w).Encode(status)
    }
}
```

### Docker Deployment

```dockerfile
# Dockerfile example
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o myapp

FROM alpine:latest
RUN adduser -D appuser
COPY --from=builder /app/myapp /usr/local/bin/
COPY config.json /etc/myapp/config.json
RUN chown appuser:appuser /etc/myapp/config.json
USER appuser
CMD ["/usr/local/bin/myapp"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    build: .
    volumes:
      - ./config:/etc/myapp:ro  # Mount config directory
      - ./logs:/var/log/myapp   # Audit logs
    environment:
      - CONFIG_PATH=/etc/myapp/config.json
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.json: |
    {
      "level": "info",
      "encoder": "json",
      "capacity": 32768
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
      - name: app
        image: myapp:latest
        volumeMounts:
        - name: config-volume
          mountPath: /etc/myapp
          readOnly: true
        - name: audit-volume
          mountPath: /var/log/audit
      volumes:
      - name: config-volume
        configMap:
          name: app-config
      - name: audit-volume
        persistentVolumeClaim:
          claimName: audit-logs
```

---

## Troubleshooting

### Common Issues

#### Hot Reload Not Working

**Symptoms:** Configuration changes not applied

**Diagnosis:**
```go
// Check watcher status
if !watcher.IsRunning() {
    logger.Error("Config watcher is not running")
}

// Check file permissions
if _, err := os.Stat("config.json"); err != nil {
    logger.Error("Config file access error", iris.Error(err))
}
```

**Solutions:**
1. Verify file exists and is readable
2. Check file permissions (644 or 755)
3. Ensure watcher was started successfully
4. Review audit logs for error messages

#### Audit Trail Missing

**Symptoms:** No audit entries in `iris-config-audit.jsonl`

**Causes:**
- Insufficient disk space
- Permission denied on audit file location
- Audit buffer not flushed

**Solutions:**
```go
// Force audit flush before shutdown
watcher.Stop() // Automatically flushes audit buffer

// Check audit file permissions
info, err := os.Stat("iris-config-audit.jsonl")
if err != nil {
    logger.Error("Audit file error", iris.Error(err))
}
```

#### Performance Issues

**Symptoms:** High CPU usage or memory consumption

**Diagnosis:**
```go
// Monitor file change frequency
// Excessive changes may indicate issue
```

**Solutions:**
1. Increase polling interval for less critical applications
2. Implement rate limiting for configuration changes
3. Use file locking to prevent partial reads during writes

### Debug Mode

Enable verbose logging for troubleshooting:

```go
// Enable debug level to see detailed hot reload activity
logger.SetLevel(iris.Debug)

// Watch for config reload messages
// [IRIS] Configuration reloaded from /path/to/config.json - Level: debug
```

### Log Analysis

```bash
# Monitor audit trail in real-time
tail -f iris-config-audit.jsonl

# Count configuration changes
grep "file_changed" iris-config-audit.jsonl | wc -l

# Verify checksums for integrity
grep "checksum" iris-config-audit.jsonl | cut -d'"' -f8 | sort | uniq
```

---

## Examples

### Example 1: Web Service with Hot Reload

```go
package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/agilira/iris"
)

func main() {
    // Setup logger
    logger, err := iris.New(iris.Config{
        Level:   iris.Info,
        Output:  os.Stdout,
        Encoder: iris.NewJSONEncoder(),
    })
    if err != nil {
        panic(err)
    }
    defer logger.Sync()
    logger.Start()
    defer logger.Close()

    // Enable hot reload
    watcher, err := iris.NewDynamicConfigWatcher("config.json", logger.AtomicLevel())
    if err != nil {
        logger.Error("Hot reload disabled", iris.Error(err))
    } else {
        defer watcher.Stop()
        watcher.Start()
        logger.Info("🔥 Hot reload enabled")
    }

    // HTTP server with context-aware logging
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        logger.Debug("Processing request", 
            iris.String("method", r.Method),
            iris.String("path", r.URL.Path),
            iris.String("user_agent", r.UserAgent()))
        
        logger.Info("User request processed",
            iris.String("endpoint", "/api/users"),
            iris.Duration("response_time", time.Millisecond*150))
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "ok"}`))
    })

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    // Graceful shutdown
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        <-sigChan
        
        logger.Info("Shutting down server...")
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        server.Shutdown(ctx)
    }()

    logger.Info("Server starting", iris.String("addr", ":8080"))
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.Error("Server error", iris.Error(err))
    }
}
```

### Example 2: Microservice with Multiple Loggers

```go
package main

import (
    "github.com/agilira/iris"
)

func main() {
    // Main application logger
    appLogger, err := setupLogger("app", "app-config.json")
    if err != nil {
        panic(err)
    }
    defer appLogger.Close()

    // Database logger
    dbLogger, err := setupLogger("database", "db-config.json")
    if err != nil {
        panic(err)
    }
    defer dbLogger.Close()

    // API logger
    apiLogger, err := setupLogger("api", "api-config.json")
    if err != nil {
        panic(err)
    }
    defer apiLogger.Close()

    // Application logic
    appLogger.Info("Application started")
    dbLogger.Debug("Database connection established")
    apiLogger.Info("API server ready")
}

func setupLogger(name, configFile string) (*iris.Logger, error) {
    logger, err := iris.New(iris.Config{
        Level:   iris.Info,
        Output:  os.Stdout,
        Encoder: iris.NewJSONEncoder(),
        Name:    name,
    })
    if err != nil {
        return nil, err
    }

    logger.Start()

    // Hot reload per component
    watcher, err := iris.NewDynamicConfigWatcher(configFile, logger.AtomicLevel())
    if err != nil {
        logger.Warn("Hot reload disabled for component", 
            iris.String("component", name),
            iris.Error(err))
    } else {
        watcher.Start()
        logger.Info("Hot reload enabled", iris.String("component", name))
    }

    return logger, nil
}
```

### Example 3: Configuration Testing

```go
package main

import (
    "encoding/json"
    "os"
    "testing"
    "time"
    
    "github.com/agilira/iris"
)

func TestHotReload(t *testing.T) {
    // Create test config file
    configFile := "test-config.json"
    initialConfig := map[string]interface{}{
        "level": "info",
    }
    
    writeConfig(t, configFile, initialConfig)
    defer os.Remove(configFile)

    // Setup logger with hot reload
    logger, err := iris.New(iris.Config{
        Level:   iris.Info,
        Output:  os.Stdout,
        Encoder: iris.NewJSONEncoder(),
    })
    if err != nil {
        t.Fatal(err)
    }
    defer logger.Close()
    logger.Start()

    watcher, err := iris.NewDynamicConfigWatcher(configFile, logger.AtomicLevel())
    if err != nil {
        t.Fatal(err)
    }
    defer watcher.Stop()

    if err := watcher.Start(); err != nil {
        t.Fatal(err)
    }

    // Test level change
    if logger.Level() != iris.Info {
        t.Errorf("Expected Info level, got %v", logger.Level())
    }

    // Change config to debug
    debugConfig := map[string]interface{}{
        "level": "debug",
    }
    writeConfig(t, configFile, debugConfig)

    // Wait for reload
    time.Sleep(3 * time.Second)

    if logger.Level() != iris.Debug {
        t.Errorf("Expected Debug level after reload, got %v", logger.Level())
    }
}

func writeConfig(t *testing.T, filename string, config map[string]interface{}) {
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        t.Fatal(err)
    }
    
    if err := os.WriteFile(filename, data, 0644); err != nil {
        t.Fatal(err)
    }
}
```

---

## Conclusion

Iris hot reload provides production-grade dynamic configuration management with:

- **Zero-downtime updates** for production environments
- **Comprehensive audit trails** for compliance and security
- **Multi-format support** for flexible configuration management  
- **Production-ready performance** with minimal overhead
- **Robust error handling** and graceful degradation
