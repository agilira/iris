// Package iris provides a high-performance, structured logging library for Go applications.
//
// Iris is designed for production environments where performance, security, and reliability
// are critical. It achieves zero-allocation logging on hot paths through lock-free ring
// buffers, buffer pooling, and type-safe field encoding.
//
// # Key Features
//
//   - Smart API with zero-configuration setup and automatic optimization
//   - Zero-allocation structured logging (~29 ns/op without fields, 0 allocs)
//   - Lock-free MPSC ring buffer architecture (ZephyrosLite)
//   - Built-in security: field sanitization, log injection prevention, secret redaction
//   - JSON and text encoders with safe key/value escaping
//   - Context-aware logging with context.Context integration
//   - Dynamic level changes via atomic operations
//   - Backpressure handling (drop-on-full or block-on-full policies)
//   - Configurable idle strategies for CPU/latency trade-offs
//   - Token-bucket sampling for high-volume scenarios
//   - Intelligent auto-scaling between SingleRing and ThreadedRings architectures
//   - SyncReader interface for integrating external log sources
//   - Modular writer ecosystem via SyncWriter interface
//
// # Smart API - Zero Configuration
//
// The Smart API automatically detects optimal settings for your environment:
//
//	logger, err := iris.New(iris.Config{})
//	if err != nil {
//		// handle error
//	}
//	logger.Start()
//	defer logger.Sync()
//
//	logger.Info("Hello world", iris.String("user", "alice"))
//
// Smart features include:
//   - Architecture detection (SingleRing vs ThreadedRings based on CPU count)
//   - Capacity optimization (power-of-two sizing, bounded by maxCapacity)
//   - Encoder selection (text for TTY, JSON otherwise)
//   - Level detection from IRIS_LEVEL environment variable
//   - Cached time source for high-frequency logging (go-timecache)
//
// # Configuration
//
// While Smart API handles most scenarios, any setting can be overridden:
//
//	logger, err := iris.New(iris.Config{
//		Output: myCustomWriter,
//		Level:  iris.Error,
//	})
//
// Config is validated at construction time: New() calls Validate() internally,
// rejecting invalid levels, out-of-range capacities, and mismatched batch sizes.
//
// # Performance
//
// Benchmark results on AMD Ryzen 5 7520U (go test -bench=. -benchmem, SingleRing):
//
//   - Message + 10 fields: ~34 ns/op, 0 allocs/op
//   - Accumulated context: ~25 ns/op, 0 allocs/op
//   - Adding 6 fields at call site: ~31 ns/op, 0 allocs/op
//   - No fields: ~29 ns/op, 0 allocs/op
//   - Disabled level (early exit): <1 ns/op, 0 allocs/op
//
// # Security
//
// Security is built into every layer:
//
//   - Field key and value sanitization prevents log injection (CWE-93, CWE-116)
//   - Secret field redaction protects sensitive data in output
//   - Encoder escapes all keys through quoteString (no raw writes)
//   - Input validation at construction (Capacity ceiling prevents CWE-400 OOM)
//
// # Field Types
//
// Type-safe field constructors minimize allocation and prevent type confusion:
//
//	logger.Info("Payment processed",
//		iris.Str("tx_id", "tx-123456"),
//		iris.Int64("amount_cents", 2499),
//		iris.Dur("elapsed", time.Since(start)),
//		iris.Secret("card", cardNumber),   // appears as [REDACTED] in output
//	)
//
// Available constructors: Str, String, Int, Int8, Int16, Int32, Int64,
// Uint, Uint8, Uint16, Uint32, Uint64, Float32, Float64, Bool,
// Dur, TimeField, Time, Bytes, Binary, Secret, Err, Stringer, Object.
//
// # Error Handling
//
// Logger creation returns errors for invalid configurations. Internal write
// errors are routed to Config.ErrorHandler (or stderr if nil). Dropped messages
// are tracked via logger.Stats()["dropped"].
//
//	logger, err := iris.New(iris.Config{})
//	if err != nil {
//		// invalid config: handle or exit
//	}
//	logger.Start()
//
// # Development Mode
//
// Development mode enables debug level, caller info, and text encoding:
//
//	logger, err := iris.New(iris.Config{}, iris.Development())
//
// # Context Integration
//
// ContextLogger carries fields extracted from context.Context:
//
//	cl := logger.WithContext(ctx, iris.WithKeys(iris.TraceIDKey))
//	cl.Info("request handled", iris.Int("status", 200))
//
// # Best Practices
//
//   - Use Smart API for all new projects: iris.New(iris.Config{})
//   - Prefer typed field constructors over formatted messages
//   - Use iris.Secret() for passwords, tokens, and PII
//   - Use iris.Development() for local development
//   - Monitor logger.Stats() in production for drop rate insight
//   - Set IRIS_LEVEL environment variable for deployment tuning
//
// For comprehensive documentation, see: https://github.com/agilira/iris
package iris
