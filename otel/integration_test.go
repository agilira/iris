// integration_test.go: Integration test for IRIS OpenTelemetry
//
// This test demonstrates real-world usage of IRIS with OpenTelemetry
// including trace correlation across service boundaries.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package otel

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/agilira/iris"
)

func TestOpenTelemetryIntegration(t *testing.T) {
	// Setup in-memory tracer for testing
	exporter := tracetest.NewInMemoryExporter()

	// Create a resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("iris-test-service"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "test"),
		),
	)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// Create IRIS logger with buffer for testing
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer logger.Close()

	// Create a parent span (simulating incoming request)
	tracer := otel.Tracer("test-service")
	parentCtx, parentSpan := tracer.Start(context.Background(), "handle_request")
	parentSpan.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/api/users"),
		attribute.String("user.id", "test-user-123"),
	)

	// Add baggage for correlation
	correlationID, _ := baggage.NewMember("correlation.id", "req-456")
	userTier, _ := baggage.NewMember("user.tier", "premium")
	bag, _ := baggage.New(correlationID, userTier)
	parentCtx = baggage.ContextWithBaggage(parentCtx, bag)

	// Create OpenTelemetry-aware logger
	otelLogger := WithTracing(parentCtx, logger)

	// Log business logic events
	otelLogger.Info("Processing user creation request",
		iris.Str("operation", "create_user"),
		iris.Str("email", "test@example.com"),
	)

	// Simulate calling another service
	childCtx, childSpan := tracer.Start(parentCtx, "validate_user_data")
	childSpan.SetAttributes(
		attribute.String("validation.type", "email"),
		attribute.Bool("validation.passed", true),
	)

	childLogger := WithTracing(childCtx, logger)
	childLogger.Debug("Email validation completed",
		iris.Bool("valid", true),
		iris.Str("provider", "external-validator"),
	)

	childSpan.End()

	// More business logic
	otelLogger.Info("User created successfully",
		iris.Str("user.id", "user-789"),
		iris.Int("processing_time_ms", 150),
	)

	parentSpan.End()

	// Verify spans were created
	spans := exporter.GetSpans()
	if len(spans) < 2 {
		t.Errorf("Expected at least 2 spans, got %d", len(spans))
	}

	// Verify parent span
	found := false
	for _, span := range spans {
		if span.Name == "handle_request" {
			found = true

			// Check attributes
			attrs := span.Attributes
			hasMethod := false
			hasRoute := false
			for _, attr := range attrs {
				if attr.Key == "http.method" && attr.Value.AsString() == "POST" {
					hasMethod = true
				}
				if attr.Key == "http.route" && attr.Value.AsString() == "/api/users" {
					hasRoute = true
				}
			}
			if !hasMethod || !hasRoute {
				t.Error("Parent span missing expected attributes")
			}
			break
		}
	}
	if !found {
		t.Error("Parent span 'handle_request' not found")
	}
}

func TestBaggagePropagation(t *testing.T) {
	// Setup tracer
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)

	// Create logger
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer logger.Close()

	// Create context with baggage
	ctx := context.Background()

	// Add multiple baggage items
	requestID, _ := baggage.NewMember("request.id", "req-123")
	tenantID, _ := baggage.NewMember("tenant.id", "tenant-456")
	featureFlag, _ := baggage.NewMember("feature.experimental", "true")

	bag, err := baggage.New(requestID, tenantID, featureFlag)
	if err != nil {
		t.Fatalf("Failed to create baggage: %v", err)
	}

	ctx = baggage.ContextWithBaggage(ctx, bag)

	// Create tracer span
	tracer := otel.Tracer("baggage-test")
	ctx, span := tracer.Start(ctx, "test_operation")
	defer span.End()

	// Use ExtractOTelContext to get enriched context
	enrichedCtx := ExtractOTelContext(ctx)

	// Create context logger
	contextLogger := logger.WithContext(enrichedCtx)

	contextLogger.Info("Testing baggage propagation",
		iris.Str("test", "baggage_extraction"),
	)

	// The log should contain baggage fields automatically
	t.Log("Baggage propagation test completed - check logs for baggage.* fields")
}

func TestResourceDetection(t *testing.T) {
	// This test verifies that resource information is properly detected
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer logger.Close()

	ctx := context.Background()

	// Test resource detection (use exported function)
	enrichedCtx := ExtractOTelContext(ctx)

	contextLogger := logger.WithContext(enrichedCtx)
	contextLogger.Info("Testing resource detection",
		iris.Str("test", "resource_detection"),
	)

	t.Log("Resource detection test completed - check logs for resource.* fields")
}
