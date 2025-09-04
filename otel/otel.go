// otel.go: OpenTelemetry integration for Iris
//
// This package provides seamless integration between Iris and OpenTelemetry,
// enabling automatic trace correlation using Iris's existing context.Context system.
// The integration leverages Iris's ContextExtractor for optimal performance.
//
// Key Features:
//   - Automatic trace/span ID extraction from context.Context
//   - Uses Iris's existing context system for zero-allocation performance
//   - Smart resource detection (service name, version, environment)
//   - Baggage propagation for distributed context
//   - Compatible with Iris's ContextLogger pattern
//
// Usage:
//   import "github.com/agilira/iris/otel"
//
//   logger, _ := iris.New(iris.Config{})
//
//   // Create OpenTelemetry-aware context logger
//   ctx := trace.WithSpanContext(context.Background(), spanCtx)
//   otelLogger := otel.WithTracing(logger, ctx)
//   otelLogger.Info("Processing request", iris.Str("user", "john"))
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package otel

import (
	"context"
	"os"
	"runtime/debug"
	"strings"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"

	"github.com/agilira/iris"
)

// WithTracing creates an OpenTelemetry-aware context logger from an IRIS logger.
// This function extracts trace information from the context and creates a ContextLogger
// with the appropriate fields pre-populated using IRIS's existing context system.
func WithTracing(logger *iris.Logger, ctx context.Context) *iris.ContextLogger {
	// Extract OpenTelemetry information and store in context
	enrichedCtx := ExtractOTelContext(ctx)

	// Create custom extractor for OpenTelemetry fields
	extractor := createOTelExtractor()

	// Use IRIS's existing context system
	otelLogger := logger.WithContextExtractor(enrichedCtx, extractor)

	// Add baggage fields
	otelLogger = WithBaggage(otelLogger, ctx)

	// Add resource fields
	resourceFields := extractResourceFields()
	if len(resourceFields) > 0 {
		otelLogger = otelLogger.With(resourceFields...)
	}

	return otelLogger
}

// createOTelExtractor builds a ContextExtractor for OpenTelemetry integration
func createOTelExtractor() *iris.ContextExtractor {
	keys := make(map[iris.ContextKey]string)

	// Add standard OpenTelemetry keys
	keys["otel.trace_id"] = "trace_id"
	keys["otel.span_id"] = "span_id"
	keys["otel.sampled"] = "trace_sampled"

	return &iris.ContextExtractor{
		Keys:     keys,
		MaxDepth: 10,
	}
}

// ExtractOTelContext extracts OpenTelemetry trace information from context
// and returns a new context with the information stored as values that
// IRIS's ContextExtractor can access.
func ExtractOTelContext(ctx context.Context) context.Context {
	// Extract span context
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ctx
	}

	sc := span.SpanContext()
	if !sc.IsValid() {
		return ctx
	}

	// Store trace information in context for IRIS's extractor
	ctx = context.WithValue(ctx, iris.ContextKey("otel.trace_id"), sc.TraceID().String())
	ctx = context.WithValue(ctx, iris.ContextKey("otel.span_id"), sc.SpanID().String())

	if sc.IsSampled() {
		ctx = context.WithValue(ctx, iris.ContextKey("otel.sampled"), "true")
	}

	return ctx
}

// WithBaggage extracts OpenTelemetry baggage and adds it to the context logger
func WithBaggage(logger *iris.ContextLogger, ctx context.Context) *iris.ContextLogger {
	bag := baggage.FromContext(ctx)
	if bag.Len() == 0 {
		return logger
	}

	fields := make([]iris.Field, 0, minInt(bag.Len(), 10)) // Max 10 baggage items
	count := 0

	// Iterate through baggage members correctly using Members() slice
	members := bag.Members()
	for _, member := range members {
		if count >= 10 {
			break
		}

		key := "baggage." + member.Key()
		value := member.Value()

		fields = append(fields, iris.Str(key, value))
		count++
	}

	if len(fields) > 0 {
		return logger.With(fields...)
	}

	return logger
}

// extractResourceFields detects service metadata automatically
func extractResourceFields() []iris.Field {
	fields := make([]iris.Field, 0, 4)

	// Service name from environment or binary name
	if serviceName := getServiceName(); serviceName != "" {
		fields = append(fields, iris.Str("service.name", serviceName))
	}

	// Service version from build info or environment
	if version := getServiceVersion(); version != "" {
		fields = append(fields, iris.Str("service.version", version))
	}

	// Environment from standard variables
	if env := getEnvironment(); env != "" {
		fields = append(fields, iris.Str("deployment.environment", env))
	}

	return fields
}

// getServiceName extracts service name from various sources
func getServiceName() string {
	// Try OpenTelemetry environment variable
	if serviceName := os.Getenv("OTEL_SERVICE_NAME"); serviceName != "" {
		return serviceName
	}

	// Try generic service name
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		return serviceName
	}

	// Fallback to binary name from build info
	if info, ok := debug.ReadBuildInfo(); ok {
		if path := info.Main.Path; path != "" {
			parts := strings.Split(path, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}

	return ""
}

// getServiceVersion extracts service version from various sources
func getServiceVersion() string {
	// Try OpenTelemetry environment variable
	if version := os.Getenv("OTEL_SERVICE_VERSION"); version != "" {
		return version
	}

	// Try generic service version
	if version := os.Getenv("SERVICE_VERSION"); version != "" {
		return version
	}

	// Fallback to build info
	if info, ok := debug.ReadBuildInfo(); ok {
		if version := info.Main.Version; version != "" && version != "(devel)" {
			return version
		}
	}

	return ""
}

// getEnvironment extracts environment from various sources
func getEnvironment() string {
	// Try OpenTelemetry resource attributes
	if env := os.Getenv("OTEL_RESOURCE_ATTRIBUTES"); env != "" {
		if strings.Contains(env, "deployment.environment=") {
			parts := strings.Split(env, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "deployment.environment=") {
					return strings.TrimPrefix(part, "deployment.environment=")
				}
			}
		}
	}

	// Try generic environment variable
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}

	return ""
}

// minInt returns the minimum of two integers
// Note: Using custom name to avoid built-in min redefinition
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
