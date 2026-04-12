// context.go: Optimized context integration for IRIS logging
//
// This implementation provides context.Context integration while maintaining
// IRIS's zero-allocation performance characteristics. Key optimizations:
// 1. Context field extraction is cached, not repeated per log call
// 2. Configurable context key extraction to avoid scanning all values
// 3. Optional context integration - zero overhead when not used
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"context"
)

// ContextKey represents a key type for context values that should be logged.
type ContextKey string

// Common context keys for standardized logging
const (
	RequestIDKey ContextKey = "request_id"
	TraceIDKey   ContextKey = "trace_id"
	SpanIDKey    ContextKey = "span_id"
	UserIDKey    ContextKey = "user_id"
	SessionIDKey ContextKey = "session_id"
)

// ContextExtractor defines which context keys should be extracted and logged.
// This prevents the performance overhead of scanning all context values.
type ContextExtractor struct {
	// Keys maps context keys to field names in log output
	Keys map[ContextKey]string

	// MaxDepth limits how deep to search in context chain (default: 10)
	MaxDepth int
}

// defaultContextExtractor returns a fresh extractor with the standard keys.
// WHY a function instead of a package-level var: a var pointing to a mutable
// struct violates INV-404 (zero package-level mutable variables). Any caller
// could modify Keys, introducing a data race. A function returns a new
// instance each time -- no shared mutable state.
func defaultContextExtractor() *ContextExtractor {
	return &ContextExtractor{
		Keys: map[ContextKey]string{
			RequestIDKey: "request_id",
			TraceIDKey:   "trace_id",
			SpanIDKey:    "span_id",
			UserIDKey:    "user_id",
			SessionIDKey: "session_id",
		},
		MaxDepth: 10,
	}
}

// ContextLogger wraps a Logger with pre-extracted context fields.
// This avoids context.Value() calls in the hot logging path.
type ContextLogger struct {
	logger *Logger
	fields []Field // Pre-extracted context fields
}

// contextOptions holds configuration for WithContext
type contextOptions struct {
	extractor *ContextExtractor
	keys      []ContextKey // Specific keys to extract (overrides extractor)
	fieldName string       // Custom field name (used with single key)
}

// ContextOption configures context extraction behavior
type ContextOption func(*contextOptions)

// WithExtractor uses a custom ContextExtractor for field extraction.
// Use this when you need full control over which keys are extracted
// and how they map to field names.
func WithExtractor(extractor *ContextExtractor) ContextOption {
	return func(opts *contextOptions) {
		opts.extractor = extractor
	}
}

// WithKeys extracts only the specified keys using their default field names.
// This is more efficient than using the full DefaultContextExtractor
// when you only need specific fields.
func WithKeys(keys ...ContextKey) ContextOption {
	return func(opts *contextOptions) {
		opts.keys = keys
	}
}

// WithKey extracts a single key with a custom field name.
// Optimized for the common case of extracting just one context value.
func WithKey(key ContextKey, fieldName string) ContextOption {
	return func(opts *contextOptions) {
		opts.keys = []ContextKey{key}
		opts.fieldName = fieldName
	}
}

// WithContext creates a new ContextLogger with fields extracted from context.
// This is the unified way to use context integration - extract once,
// log many times with the same context.
//
// Usage:
//
//	// Default extraction (all standard keys)
//	ctxLogger := logger.WithContext(ctx)
//
//	// Custom extractor
//	ctxLogger := logger.WithContext(ctx, WithExtractor(myExtractor))
//
//	// Specific keys only
//	ctxLogger := logger.WithContext(ctx, WithKeys(RequestIDKey, TraceIDKey))
//
//	// Single key with custom field name
//	ctxLogger := logger.WithContext(ctx, WithKey(RequestIDKey, "req_id"))
//
// Performance: O(k) where k is number of configured keys, not context depth.
func (l *Logger) WithContext(ctx context.Context, opts ...ContextOption) *ContextLogger {
	// Apply options
	options := &contextOptions{
		extractor: defaultContextExtractor(),
	}
	for _, opt := range opts {
		opt(options)
	}

	var fields []Field

	// Handle specific keys mode (more efficient)
	if len(options.keys) > 0 {
		fields = make([]Field, 0, len(options.keys))
		for _, key := range options.keys {
			if value := ctx.Value(key); value != nil {
				if strValue, ok := value.(string); ok && strValue != "" {
					// Use custom field name if single key with custom name
					fieldName := string(key)
					if options.fieldName != "" && len(options.keys) == 1 {
						fieldName = options.fieldName
					}
					fields = append(fields, Str(fieldName, strValue))
				}
			}
		}
		return &ContextLogger{logger: l, fields: fields}
	}

	// Handle extractor mode
	if options.extractor != nil && len(options.extractor.Keys) > 0 {
		fields = make([]Field, 0, len(options.extractor.Keys))
		for contextKey, fieldName := range options.extractor.Keys {
			if value := ctx.Value(contextKey); value != nil {
				if strValue, ok := value.(string); ok && strValue != "" {
					fields = append(fields, Str(fieldName, strValue))
				}
			}
		}
	}

	return &ContextLogger{logger: l, fields: fields}
}

// Logging methods for ContextLogger - all delegate to underlying logger
// with pre-extracted context fields automatically included.

// Debug logs a message at debug level with context fields
func (cl *ContextLogger) Debug(msg string, fields ...Field) {
	if cl.logger.level.Level() > Debug {
		return
	}
	allFields := append(cl.fields, fields...)
	cl.logger.Debug(msg, allFields...)
}

// Info logs a message at info level with context fields
func (cl *ContextLogger) Info(msg string, fields ...Field) {
	if cl.logger.level.Level() > Info {
		return
	}
	allFields := append(cl.fields, fields...)
	cl.logger.Info(msg, allFields...)
}

// Warn logs a message at warn level with context fields
func (cl *ContextLogger) Warn(msg string, fields ...Field) {
	if cl.logger.level.Level() > Warn {
		return
	}
	allFields := append(cl.fields, fields...)
	cl.logger.Warn(msg, allFields...)
}

// Error logs a message at error level with context fields
func (cl *ContextLogger) Error(msg string, fields ...Field) {
	if cl.logger.level.Level() > Error {
		return
	}
	allFields := append(cl.fields, fields...)
	cl.logger.Error(msg, allFields...)
}

// Fatal logs a message at fatal level with context fields and exits
func (cl *ContextLogger) Fatal(msg string, fields ...Field) {
	allFields := append(cl.fields, fields...)
	cl.logger.Fatal(msg, allFields...)
}

// With creates a new ContextLogger with additional fields.
// This preserves both context fields and manually added fields.
func (cl *ContextLogger) With(fields ...Field) *ContextLogger {
	newFields := make([]Field, len(cl.fields)+len(fields))
	copy(newFields, cl.fields)
	copy(newFields[len(cl.fields):], fields)

	return &ContextLogger{
		logger: cl.logger,
		fields: newFields,
	}
}

// WithAdditionalContext extracts additional context values without losing existing ones.
func (cl *ContextLogger) WithAdditionalContext(ctx context.Context, opts ...ContextOption) *ContextLogger {
	// Extract new context fields using the unified API
	newContextLogger := cl.logger.WithContext(ctx, opts...)

	// Combine existing fields with new context fields
	return cl.With(newContextLogger.fields...)
}
