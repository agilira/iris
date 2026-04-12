# Iris Encoders

Iris provides two built-in encoders. Each encoder implements the `Encoder` interface and can be configured independently.

## Encoder Interface

```go
type Encoder interface {
    Encode(rec *Record, now time.Time, buf *bytes.Buffer)
}
```

## Available Encoders

### 1. JSON Encoder

**File:** `encoder-json.go`

Implements NDJSON (newline-delimited JSON) format with zero-reflection encoding.

```go
encoder := iris.NewJSONEncoder()

// Customization
encoder.TimeKey = "timestamp"    // default: "ts"
encoder.LevelKey = "severity"    // default: "level"
encoder.MsgKey = "message"       // default: "msg"
encoder.RFC3339 = false          // default: true (uses UnixNano if false)
```

**Security features:**
- Field keys pass through `quoteString()` (CWE-116 protection)
- Control character neutralization in values
- Newline injection protection
- No reflection, no `encoding/json`

**Use Cases:**
- Structured logging systems
- Log aggregation (ELK, Splunk, etc.)
- API logging and monitoring
- Machine-readable logs

**Output Example:**
```json
{"ts":"2025-09-06T14:30:45.123Z","level":"info","msg":"User login","user":"john","ip":"192.168.1.1"}
```

### 2. Text Encoder

**File:** `encoder-text.go`

Provides secure human-readable text encoding with comprehensive log injection protection.

```go
encoder := iris.NewTextEncoder()

// Customization
encoder.TimeFormat = time.RFC3339Nano  // default: time.RFC3339
encoder.QuoteValues = false            // default: true
encoder.SanitizeKeys = false           // default: true
```

**Security Features:**
- Field key sanitization
- Value sanitization with proper quoting
- Control character neutralization
- Newline injection protection
- Unicode direction override protection

**Use Cases:**
- Production logs requiring security
- System logs in untrusted environments
- Compliance and audit logging
- General-purpose structured text logs

**Output Example:**
```
time=2025-09-06T14:30:45Z level=info msg="User login" user=john ip=192.168.1.1
```

## Configuration Examples

### Production (JSON)

```go
cfg := iris.Config{
    Output:  iris.WrapWriter(os.Stdout),
    Encoder: iris.NewJSONEncoder(),
    Level:   iris.Info,
}

logger, err := iris.New(cfg)
```

### Production (Text, secure)

```go
textEncoder := iris.NewTextEncoder()
textEncoder.QuoteValues = true
textEncoder.SanitizeKeys = true

cfg := iris.Config{
    Output:  iris.WrapWriter(os.Stdout),
    Encoder: textEncoder,
    Level:   iris.Warn,
}

logger, err := iris.New(cfg)
```

## Field Types

Both encoders support the full range of Iris field types:

```go
logger.Info("Complex log entry",
    iris.Str("service", "auth"),
    iris.Int64("user_id", 12345),
    iris.Float64("response_time", 0.045),
    iris.Bool("success", true),
    iris.Duration("timeout", 30*time.Second),
    iris.Time("login_time", time.Now()),
    iris.Bytes("payload", []byte("binary data")),
    iris.NamedError("error", err),
    iris.Any("metadata", complexObject),
)
```

## Resilience

Encoders are designed to never panic:

- Invalid field types are encoded as strings
- Nil values are handled gracefully
- Malformed data is sanitized (text encoder)
- Buffers are reused from a pool with zero allocations during normal operation

---

Iris -- an AGILira fragment
