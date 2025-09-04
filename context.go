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

// DefaultContextExtractor provides sensible defaults for common use cases.
var DefaultContextExtractor = &ContextExtractor{
	Keys: map[ContextKey]string{
		RequestIDKey: "request_id",
		TraceIDKey:   "trace_id",
		SpanIDKey:    "span_id",
		UserIDKey:    "user_id",
		SessionIDKey: "session_id",
	},
	MaxDepth: 10,
}

// ContextLogger wraps a Logger with pre-extracted context fields.
// This avoids context.Value() calls in the hot logging path.
type ContextLogger struct {
	logger *Logger
	fields []Field // Pre-extracted context fields
}

// WithContext creates a new ContextLogger with fields extracted from context.
// This is the recommended way to use context integration - extract once,
// log many times with the same context.
//
// Performance: O(k) where k is number of configured keys, not context depth.
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	return l.WithContextExtractor(ctx, DefaultContextExtractor)
}

// WithContextExtractor creates a ContextLogger with custom extraction rules.
func (l *Logger) WithContextExtractor(ctx context.Context, extractor *ContextExtractor) *ContextLogger {
	var fields []Field

	// Pre-allocate for common case
	if len(extractor.Keys) > 0 {
		fields = make([]Field, 0, len(extractor.Keys))
	}

	// Extract configured keys only (not all context values)
	for contextKey, fieldName := range extractor.Keys {
		if value := ctx.Value(contextKey); value != nil {
			// Type assertion to string (most common case)
			if strValue, ok := value.(string); ok && strValue != "" {
				fields = append(fields, Str(fieldName, strValue))
			}
		}
	}

	return &ContextLogger{
		logger: l,
		fields: fields,
	}
}

// WithContextValue creates a ContextLogger with a single context value.
// Optimized for cases where you only need one context field.
func (l *Logger) WithContextValue(ctx context.Context, key ContextKey, fieldName string) *ContextLogger {
	var fields []Field

	if value := ctx.Value(key); value != nil {
		if strValue, ok := value.(string); ok && strValue != "" {
			fields = []Field{Str(fieldName, strValue)}
		}
	}

	return &ContextLogger{
		logger: l,
		fields: fields,
	}
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
func (cl *ContextLogger) WithAdditionalContext(ctx context.Context, extractor *ContextExtractor) *ContextLogger {
	// Extract new context fields
	newContextLogger := cl.logger.WithContextExtractor(ctx, extractor)

	// Combine existing fields with new context fields
	return cl.With(newContextLogger.fields...)
}

// Performance optimization: Context field extraction helpers

// WithRequestID extracts request ID with minimal allocations.
// Optimized for the most common use case.
func (l *Logger) WithRequestID(ctx context.Context) *ContextLogger {
	return l.WithContextValue(ctx, RequestIDKey, "request_id")
}

// WithTraceID extracts trace ID for distributed tracing.
func (l *Logger) WithTraceID(ctx context.Context) *ContextLogger {
	return l.WithContextValue(ctx, TraceIDKey, "trace_id")
}

// WithUserID extracts user ID from context for user-specific logging.
func (l *Logger) WithUserID(ctx context.Context) *ContextLogger {
	return l.WithContextValue(ctx, UserIDKey, "user_id")
}

// ContextMiddleware creates a middleware pattern for HTTP handlers.
// This demonstrates how to use context logging efficiently in web applications.
//
// Example usage:
//   handler := ContextMiddleware(logger)(http.HandlerFunc(myHandler))
//
// func ContextMiddleware(logger *Logger) func(http.Handler) http.Handler {
//     return func(next http.Handler) http.Handler {
//         return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//             // Extract request ID from header or generate one
//             requestID := r.Header.Get("X-Request-ID")
//             if requestID == "" {
//                 requestID = generateRequestID()
//             }
//
//             // Create context logger with request fields
//             contextLogger := logger.WithContext(r.Context()).With(
//                 Str("request_id", requestID),
//                 Str("method", r.Method),
//                 Str("path", r.URL.Path),
//             )
//
//             // Store in context for handlers
//             ctx := context.WithValue(r.Context(), "logger", contextLogger)
//             next.ServeHTTP(w, r.WithContext(ctx))
//         })
//     }
// }
//                 requestID = generateRequestID()
//             }
//
//             // Add to context
//             ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
//
//             // Create context logger once per request
//             contextLogger := logger.WithRequestID(ctx)
//
//             // Store in context for use by handlers
//             ctx = context.WithValue(ctx, "logger", contextLogger)
//
//             next.ServeHTTP(w, r.WithContext(ctx))
//         })
//     }
// }
