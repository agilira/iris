// Package iris provides a high-performance, structured logging library for Go applications.
//
// Iris is designed for production environments where performance, security, and reliability
// are critical. It offers zero-allocation logging paths, automatic memory management,
// and comprehensive security features including secure field handling and log injection prevention.
//
// # Key Features
//
//   - Smart API with zero-configuration setup and automatic optimization
//   - High-performance structured logging with zero-allocation fast paths
//   - Automatic memory management with buffer pooling and ring buffer architecture
//   - Comprehensive security features including field sanitization and injection prevention
//   - Multiple output formats: JSON, text, and console with smart formatting
//   - Dynamic configuration with hot-reload capabilities
//   - Built-in caller information and stack trace support
//   - Backpressure handling and automatic scaling
//   - OpenTelemetry integration support
//   - Extensive field types with type-safe APIs
//
// # Smart API - Zero Configuration
//
// The revolutionary Smart API automatically detects optimal settings for your environment:
//
//	// Smart API: Everything auto-configured
//	logger, err := iris.New(iris.Config{})
//	logger.Start()
//	logger.Info("Hello world", iris.String("user", "alice"))
//
// Smart features include:
//   - Architecture detection (SingleRing vs ThreadedRings based on CPU count)
//   - Capacity optimization (8KB per CPU core, bounded 8KB-64KB)
//   - Encoder selection (Text for development, JSON for production)
//   - Level detection (from environment or development mode)
//   - Time optimization (121x faster cached time)
//
// # Quick Start
//
// Basic usage with Smart API (recommended):
//
//	logger, err := iris.New(iris.Config{})
//	if err != nil {
//		panic(err)
//	}
//	logger.Start()
//	defer logger.Sync()
//
//	logger.Info("Application started", iris.String("version", "1.0.0"))
//
// Development mode with debug logging:
//
//	logger, err := iris.New(iris.Config{}, iris.Development())
//	logger.Start()
//	logger.Debug("Debug information visible")
//
// # Configuration
//
// While Smart API handles most scenarios, you can override specific settings:
//
//	// Override only what you need, rest is auto-detected
//	config := iris.Config{
//		Output: myCustomWriter,  // Custom output
//		Level:  iris.ErrorLevel, // Error level only
//		// Everything else: auto-optimized
//	}
//	logger, err := iris.New(config)
//
// Environment variable support:
//
//	export IRIS_LEVEL=debug  # Automatically detected by Smart API
//
// # Performance Optimizations
//
// Iris includes several performance optimizations automatically enabled by Smart API:
//
//   - Time caching for high-frequency logging scenarios (121x faster than time.Now())
//   - Buffer pooling to minimize garbage collection
//   - Ring buffer architecture for lock-free writes
//   - Smart idle strategies for CPU optimization
//   - Zero-allocation fast paths for common operations
//   - Architecture auto-detection based on system resources
//
// # Security Features
//
// Security is built into every aspect of Iris:
//
//   - Field sanitization prevents log injection attacks
//   - Secret field redaction protects sensitive data
//   - Caller verification prevents stack manipulation
//   - Safe string handling prevents buffer overflows
//
// # Field Types
//
// Iris supports a comprehensive set of field types with type-safe constructors:
//
//	logger.Info("User operation",
//		iris.String("user_id", "12345"),
//		iris.Int64("timestamp", time.Now().Unix()),
//		iris.Duration("elapsed", time.Since(start)),
//		iris.Error("error", err),
//		iris.Secret("password", "[REDACTED]"),
//	)
//
// # Advanced Usage
//
// For advanced scenarios, Iris provides:
//
//   - Custom encoders for specialized output formats
//   - Hierarchical loggers with inherited fields
//   - Sampling for high-volume scenarios
//   - Integration with monitoring systems
//   - Custom sink implementations
//   - Manual configuration overrides when needed
//
// # Error Handling
//
// Iris uses non-blocking error handling to maintain performance:
//
//	logger, err := iris.New(iris.Config{})
//	if err != nil {
//		// Handle configuration errors
//	}
//	logger.Start()
//
//	if dropped := logger.Dropped(); dropped > 0 {
//		// Handle dropped log entries
//	}
//
// # Performance Comparison
//
// Smart API delivers significant performance improvements:
//   - Hot Path Allocations: 1-3 allocs/op (67% reduction)
//   - Encoding Performance: 324-537 ns/op (40-60% improvement)
//   - Memory per Record: 2.5KB (75% reduction)
//   - Configuration: Zero lines vs 15-20 lines manually
//
// # Best Practices
//
//   - Use Smart API for all new projects (iris.New(iris.Config{}))
//   - Prefer structured fields instead of formatted messages
//   - Use typed field constructors (String, Int64, etc.)
//   - Leverage environment variables for deployment configuration
//   - Monitor dropped log entries in high-load scenarios
//   - Use iris.Development() for local development
//   - Use iris.Secret() for sensitive data fields
//
// For comprehensive documentation and examples, see:
// https://github.com/agilira/iris
package iris
