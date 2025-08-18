// presets.go: Configuration presets for common use cases
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

// NewDevelopment creates a logger suitable for development
// Features: console output with colors, debug level, human-readable format
func NewDevelopment() (*Logger, error) {
	return New(Config{
		Level:      DebugLevel,
		Writer:     StdoutWriter,
		Format:     JSONFormat, // Structured but readable
		BufferSize: 1024,       // Smaller buffer for development
		BatchSize:  32,         // Smaller batches for immediate feedback
	})
}

// NewProduction creates a logger suitable for production
// Features: JSON output, info level, optimized for performance
func NewProduction() (*Logger, error) {
	return New(Config{
		Level:      InfoLevel,
		Writer:     StdoutWriter,
		Format:     JSONFormat, // Structured for parsing
		BufferSize: 8192,       // Larger buffer for throughput
		BatchSize:  128,        // Larger batches for efficiency
	})
}

// NewExample creates a logger suitable for examples and testing
// Features: deterministic output, no timestamps, consistent formatting
func NewExample() (*Logger, error) {
	return New(Config{
		Level:            InfoLevel,
		Writer:           StdoutWriter,
		Format:           JSONFormat,
		BufferSize:       512,  // Small buffer for examples
		BatchSize:        16,   // Small batches
		DisableTimestamp: true, // Deterministic output for examples
	})
}

// NewUltraFast creates a logger optimized for maximum performance
// Features: binary format, minimal allocations, maximum throughput
// PRODUCTION READY: Writes to stderr by default for binary logging
// For testing use: logger, _ := NewUltraFast(); logger.RemoveAll(); logger.AddWriteSyncer(DiscardSyncer)
func NewUltraFast() (*Logger, error) {
	return New(Config{
		Level:            InfoLevel,
		Writer:           StderrWriteSyncer, // Use stderr for production binary logs
		Format:           BinaryFormat,      // Fastest encoding
		BufferSize:       16384,             // Large buffer
		BatchSize:        256,               // Large batches
		DisableTimestamp: true,              // Skip timestamp for maximum speed
		EnableCaller:     false,             // Skip caller info for speed
		UltraFast:        true,              // Enable all optimizations
	})
}

// NewUltraFastFile creates a high-performance binary logger for file output
// Features: binary format to file, maximum performance, production-ready
func NewUltraFastFile(filePath string) (*Logger, error) {
	fileWriter, err := NewFileWriter(filePath)
	if err != nil {
		return nil, err
	}

	return New(Config{
		Level:            InfoLevel,
		Writer:           fileWriter,
		Format:           BinaryFormat, // Fastest encoding
		BufferSize:       16384,        // Large buffer
		BatchSize:        256,          // Large batches
		DisableTimestamp: false,        // Keep timestamp for production
		EnableCaller:     false,        // Skip caller for speed
		UltraFast:        true,         // Enable all optimizations
	})
}

// NewUltraFastNetwork creates a high-performance binary logger for network output
// Features: binary format to network, maximum throughput, production-ready
func NewUltraFastNetwork(writer WriteSyncer) (*Logger, error) {
	return New(Config{
		Level:            InfoLevel,
		Writer:           writer,
		Format:           BinaryFormat, // Fastest encoding
		BufferSize:       32768,        // Extra large buffer for network
		BatchSize:        512,          // Large batches for network efficiency
		DisableTimestamp: false,        // Keep timestamp for production
		EnableCaller:     false,        // Skip caller for speed
		UltraFast:        true,         // Enable all optimizations
	})
}

// NewFastText creates a logger with fast text output
// Features: human-readable text, fast encoding, good for development
func NewFastText() (*Logger, error) {
	return New(Config{
		Level:      DebugLevel,
		Writer:     StdoutWriter,
		Format:     FastTextFormat, // Ultra-fast text format
		BufferSize: 2048,
		BatchSize:  64,
	})
}

// NewDevelopmentWithStackTrace creates a development logger with stack trace support
// Features: console output with colors, debug level, stack traces for errors
func NewDevelopmentWithStackTrace() (*Logger, error) {
	return New(Config{
		Level:           DebugLevel,
		Writer:          StdoutWriter,
		Format:          ConsoleFormat, // Better for viewing stack traces
		BufferSize:      1024,
		BatchSize:       32,
		StackTraceLevel: ErrorLevel, // Stack traces for errors and panics
	})
}

// NewDebugWithStackTrace creates a debug logger with comprehensive stack trace support
// Features: all levels logged, stack traces for warnings and above
func NewDebugWithStackTrace() (*Logger, error) {
	return New(Config{
		Level:           DebugLevel,
		Writer:          StdoutWriter,
		Format:          ConsoleFormat, // Best for debugging
		BufferSize:      1024,
		BatchSize:       16,        // Immediate feedback
		StackTraceLevel: WarnLevel, // Stack traces for warn, error, and panic
		EnableCaller:    true,      // Include caller info for debugging
	})
}

// DevelopmentConfig returns a development configuration
// Use this if you want to customize the development preset
func DevelopmentConfig() Config {
	return Config{
		Level:      DebugLevel,
		Writer:     StdoutWriter,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}
}

// ProductionConfig returns a production configuration
// Use this if you want to customize the production preset
func ProductionConfig() Config {
	return Config{
		Level:      InfoLevel,
		Writer:     StdoutWriter,
		Format:     JSONFormat,
		BufferSize: 8192,
		BatchSize:  128,
	}
}

// ExampleConfig returns an example configuration
// Use this if you want to customize the example preset
func ExampleConfig() Config {
	return Config{
		Level:            InfoLevel,
		Writer:           StdoutWriter,
		Format:           JSONFormat,
		BufferSize:       512,
		BatchSize:        16,
		DisableTimestamp: true,
	}
}

// DevelopmentWithStackTraceConfig returns a development configuration with stack traces
// Use this if you want to customize the development with stack trace preset
func DevelopmentWithStackTraceConfig() Config {
	return Config{
		Level:           DebugLevel,
		Writer:          StdoutWriter,
		Format:          ConsoleFormat,
		BufferSize:      1024,
		BatchSize:       32,
		StackTraceLevel: ErrorLevel,
	}
}

// DebugWithStackTraceConfig returns a debug configuration with comprehensive stack traces
// Use this if you want to customize the debug with stack trace preset
func DebugWithStackTraceConfig() Config {
	return Config{
		Level:           DebugLevel,
		Writer:          StdoutWriter,
		Format:          ConsoleFormat,
		BufferSize:      1024,
		BatchSize:       16,
		StackTraceLevel: WarnLevel,
		EnableCaller:    true,
	}
}
