// otel_example_test.go: Example usage of OpenTelemetry integration
//
// This example demonstrates how to use IRIS with OpenTelemetry for
// automatic trace correlation and distributed tracing.

package otel_test

import (
	"context"
	"os"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/agilira/iris"
	irisotel "github.com/agilira/iris/otel"
)

func ExampleWithTracing() {
	// Set up a test tracer for demonstration
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)

	// Create IRIS logger
	logger, _ := iris.New(iris.Config{
		Output: os.Stdout,
		Level:  iris.Info,
	})
	logger.Start()
	defer logger.Close()

	// Create a new trace
	tracer := otel.Tracer("example")
	ctx, span := tracer.Start(context.Background(), "process_request")
	defer span.End()

	// Add some attributes to the span
	span.SetAttributes(
		attribute.String("user.id", "john_doe"),
		attribute.String("request.method", "POST"),
	)

	// Add baggage for distributed context
	member, _ := baggage.NewMember("correlation.id", "abc123")
	bag, _ := baggage.New(member)
	ctx = baggage.ContextWithBaggage(ctx, bag)

	// Create OpenTelemetry-aware logger
	otelLogger := irisotel.WithTracing(logger, ctx)

	// Log messages - trace information is automatically included
	otelLogger.Info("Processing user request",
		iris.Str("endpoint", "/api/users"),
		iris.Int("status", 200),
	)

	otelLogger.Info("Request completed successfully",
		iris.Float64("duration_ms", 45.67),
	)

	// Output will include trace_id, span_id, baggage.correlation.id, and resource fields
}

func TestExtractOTelContext(t *testing.T) {
	// This test shows how to manually extract OpenTelemetry context
	// for advanced use cases

	// Set up tracer
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)

	// Create a span context
	tracer := otel.Tracer("example")
	ctx, span := tracer.Start(context.Background(), "manual_extraction")
	defer span.End()

	// Extract OpenTelemetry context manually
	enrichedCtx := irisotel.ExtractOTelContext(ctx)

	// You can now use this context with IRIS's standard context methods
	logger, _ := iris.New(iris.Config{Output: os.Stdout})
	logger.Start()
	defer logger.Close()

	contextLogger := logger.WithContext(enrichedCtx)
	contextLogger.Info("This will include trace information")

	t.Log("OpenTelemetry context extracted and used with IRIS")
}
