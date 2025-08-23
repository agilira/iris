// options.go: Advanced logger configuration and hooks system
//
// This module provides a sophisticated options system for the IRIS logger,
// enabling runtime configuration, caller information capture, stack traces,
// hooks, and development-specific behaviors. The options system is designed
// for zero-allocation operation in hot paths with hook execution deferred
// to the consumer thread to avoid contention.
//
// Key Features:
//   - Hook system executed in consumer thread (no contention)
//   - Caller information capture with configurable skip levels
//   - Stack trace collection for specified log levels
//   - Development mode behaviors (DPanic -> panic)
//   - Immutable option sets for thread safety
//
// Performance Design:
//   - Options are applied once during logger creation/cloning
//   - Hook execution happens in consumer thread only
//   - Zero allocations in producer threads for hook processing
//   - Caller information captured only when requested
//
// Copyright (c) 2025 AGILira
// Series: IRIS Logging Library
// SPDX-License-Identifier: MPL-2.0

package iris

// Hook represents a function executed in the consumer thread after log record processing.
//
// Hooks are executed in the consumer thread to avoid contention with producer
// threads. This design ensures maximum performance for logging operations while
// still allowing powerful post-processing capabilities.
//
// Hook functions receive the fully populated Record after encoding but before
// the buffer is returned to the pool. This allows for:
//   - Metrics collection
//   - Log forwarding to external systems
//   - Custom processing based on log content
//   - Development-time debugging
//
// Performance Notes:
//   - Executed in single consumer thread (no locks needed)
//   - Called after encoding is complete
//   - Should avoid blocking operations to maintain throughput
//
// Thread Safety: Hooks are called from single consumer thread only
type Hook func(rec *Record)

// loggerOptions contains immutable configuration for a logger instance.
//
// This structure holds all optional configuration that affects logger behavior.
// Options are immutable after logger creation to ensure thread safety and
// consistent behavior across the logger's lifetime.
//
// Design Principles:
//   - Immutable after construction for thread safety
//   - Zero-allocation access in hot paths
//   - Caller information only captured when requested
//   - Hook execution deferred to consumer thread
//
// Performance Features:
//   - Boolean flags for O(1) feature checks
//   - Pre-allocated hook slice to avoid runtime allocations
//   - Minimal memory footprint per logger instance
type loggerOptions struct {
	// Caller information configuration
	addCaller  bool // Enable caller information capture
	callerSkip int  // Number of stack frames to skip for caller detection

	// Stack trace configuration
	stackMin Level // Minimum level for stack trace capture (0 = disabled)

	// Development mode features
	development bool // Enable development-specific behaviors (DPanic -> panic)

	// Hook system
	hooks []Hook // Post-processing hooks executed in consumer thread
}

// Option represents a function that modifies logger options during construction.
//
// Options use the functional options pattern to provide a clean, extensible
// API for logger configuration. Each Option function modifies the options
// structure in place during logger creation or cloning.
//
// Pattern Benefits:
//   - Backward compatible API evolution
//   - Clear, self-documenting configuration
//   - Composable option sets
//   - Type-safe configuration
//
// Usage:
//
//	logger := logger.WithOptions(
//	    iris.WithCaller(),
//	    iris.AddStacktrace(iris.Error),
//	    iris.Development(),
//	)
type Option func(*loggerOptions)

// WithCaller enables caller information capture for log records.
//
// When enabled, the logger will capture the file name, line number, and
// function name of the calling code for each log record. This information
// is added to the log output for debugging and troubleshooting.
//
// Performance Impact:
//   - Adds runtime.Caller() call per log operation
//   - Minimal allocation for caller information
//   - Skip level optimization reduces overhead
//
// Returns:
//   - Option: Configuration function to enable caller capture
//
// Example:
//
//	logger := logger.WithOptions(iris.WithCaller())
//	logger.Info("message") // Will include caller info
func WithCaller() Option {
	return func(o *loggerOptions) { o.addCaller = true }
}

// WithCallerSkip sets the number of stack frames to skip for caller detection.
//
// This option is useful when the logger is wrapped by helper functions and
// you want the caller information to point to the actual calling code rather
// than the wrapper function.
//
// Parameters:
//   - n: Number of stack frames to skip (negative values are treated as 0)
//
// Common Skip Values:
//   - 0: Direct caller of log method
//   - 1: Skip one wrapper function
//   - 2+: Skip multiple wrapper layers
//
// Returns:
//   - Option: Configuration function to set caller skip level
//
// Example:
//
//	// Skip helper function to show actual caller
//	logger := logger.WithOptions(
//	    iris.WithCaller(),
//	    iris.WithCallerSkip(1),
//	)
func WithCallerSkip(n int) Option {
	return func(o *loggerOptions) {
		if n < 0 {
			n = 0
		}
		o.callerSkip = n
	}
}

// AddStacktrace enables stack trace capture for log levels at or above the specified minimum.
//
// Stack traces provide detailed call stack information for debugging complex
// issues. They are automatically captured for severe log levels (typically
// Error and above) to aid in troubleshooting.
//
// Parameters:
//   - min: Minimum log level for stack trace capture (Debug, Info, Warn, Error)
//
// Performance Impact:
//   - Stack trace capture is expensive (runtime.Stack() call)
//   - Only enabled for specified log levels to minimize overhead
//   - Stack traces are captured in producer thread but processed in consumer
//
// Returns:
//   - Option: Configuration function to enable stack trace capture
//
// Example:
//
//	// Capture stack traces for Error level and above
//	logger := logger.WithOptions(iris.AddStacktrace(iris.Error))
//	logger.Error("critical error") // Will include stack trace
//	logger.Warn("warning")         // No stack trace
func AddStacktrace(min Level) Option {
	return func(o *loggerOptions) { o.stackMin = min }
}

// Development enables development-specific behaviors for enhanced debugging.
//
// Development mode changes logger behavior to be more suitable for development
// and testing environments:
//   - DPanic level causes panic() in addition to logging
//   - Enhanced error reporting and validation
//   - More verbose debugging information
//
// This option should typically be disabled in production environments for
// optimal performance and stability.
//
// Returns:
//   - Option: Configuration function to enable development mode
//
// Example:
//
//	logger := logger.WithOptions(iris.Development())
//	logger.DPanic("development panic") // Will panic in dev mode, log in production
func Development() Option {
	return func(o *loggerOptions) {
		o.development = true
		// Enable stack traces for Error level and above in development mode
		if o.stackMin == StacktraceDisabled {
			o.stackMin = Error
		}
	}
}

// WithHook adds a post-processing hook to the logger.
//
// Hooks are functions executed in the consumer thread after log records are
// processed but before buffers are returned to the pool. This design ensures
// zero contention with producer threads while enabling powerful post-processing.
//
// Hook Use Cases:
//   - Metrics collection based on log content
//   - Log forwarding to external systems
//   - Custom alerting on specific log patterns
//   - Development-time debugging and validation
//
// Parameters:
//   - h: Hook function to execute (nil hooks are ignored)
//
// Performance Notes:
//   - Hooks are executed sequentially in consumer thread
//   - Should avoid blocking operations to maintain throughput
//   - No allocation overhead in producer threads
//
// Returns:
//   - Option: Configuration function to add the hook
//
// Example:
//
//	metricHook := func(rec *Record) {
//	    if rec.Level >= iris.Error {
//	        errorCounter.Inc()
//	    }
//	}
//	logger := logger.WithOptions(iris.WithHook(metricHook))
func WithHook(h Hook) Option {
	return func(o *loggerOptions) {
		if h != nil {
			o.hooks = append(o.hooks, h)
		}
	}
}

// newLoggerOptions creates a new loggerOptions with proper default values.
func newLoggerOptions() loggerOptions {
	return loggerOptions{
		addCaller:   false,
		callerSkip:  0,
		stackMin:    StacktraceDisabled, // Disabled by default
		development: false,
		hooks:       nil,
	}
}

// merge creates a new options set by cloning current options and applying new Option functions.
//
// This method implements the immutable options pattern, ensuring that applying
// options to a logger creates a new configuration without modifying the original.
// This design ensures thread safety and prevents unexpected configuration changes.
//
// Parameters:
//   - opts: Option functions to apply to the cloned configuration
//
// Returns:
//   - loggerOptions: New options set with applied modifications
//
// Implementation Notes:
//   - Performs shallow copy of original options
//   - Hook slices are shared between instances (hooks are immutable)
//   - All Option functions are applied in order
//   - Thread-safe through immutability
//
// Example (internal usage):
//
//	newOpts := logger.opts.merge(iris.WithCaller(), iris.Development())
func (o loggerOptions) merge(opts ...Option) loggerOptions {
	no := o // copy
	for _, fn := range opts {
		fn(&no)
	}
	return no
}
