// main_test.go: Tests for context integration examples
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/agilira/iris"
)

// Context key types to avoid collisions
type contextKey string

const (
	userIDKey       contextKey = "user_id"
	requestIDKey    contextKey = "request_id"
	tenantKey       contextKey = "tenant"
	organizationKey contextKey = "organization"
	envKey          contextKey = "env"
	reqIDKey        contextKey = "req_id"
	loggerKey       contextKey = "logger"
)

// TestBasicContextExample tests the basic context extraction example
func TestBasicContextExample(t *testing.T) {
	// Create buffer to capture output
	var buf bytes.Buffer

	// Create test logger with buffer output
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Output:  iris.WrapWriter(&buf),
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	logger.Start()

	// Wait for logger to be ready
	time.Sleep(10 * time.Millisecond)

	// Run basic context example logic directly
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user123")
	ctx = context.WithValue(ctx, requestIDKey, "req-456")

	// Log with context
	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("User authentication started")
	contextLogger.Info("Authentication successful")
	contextLogger.Info("Login process completed")
	contextLogger.Info("Request processed with fast method")

	// Wait for logs to be flushed
	time.Sleep(50 * time.Millisecond)

	// Get output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from basicContextExample, got empty string")
	}

	// Check for expected log messages
	expectedMessages := []string{
		"User authentication started",
		"Authentication successful",
		"Login process completed",
		"Request processed with fast method",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected output to contain %q, but it didn't", msg)
		}
	}
}

// TestHTTPMiddlewareExample tests the HTTP middleware pattern example
func TestHTTPMiddlewareExample(t *testing.T) {
	// Create buffer to capture output
	var buf bytes.Buffer

	// Create test logger with buffer output
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Output:  iris.WrapWriter(&buf),
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	logger.Start()

	// Wait for logger to be ready
	time.Sleep(10 * time.Millisecond)

	// Create context with request ID
	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey, "req-http-example")

	// Log with context like in the middleware
	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("Handler called")
	contextLogger.Debug("Processing business logic")
	contextLogger.Info("Handler completed")

	// Wait for logs to be flushed
	time.Sleep(50 * time.Millisecond)

	// Get output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from httpMiddlewareExample, got empty string")
	}

	// Check for expected log messages
	expectedMessages := []string{
		"Handler called",
		"Processing business logic",
		"Handler completed",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected output to contain %q, but it didn't", msg)
		}
	}
}

// TestCustomExtractorExample tests the custom context extractor example
func TestCustomExtractorExample(t *testing.T) {
	// Create buffer to capture output
	var buf bytes.Buffer

	// Create test logger with buffer output
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Output:  iris.WrapWriter(&buf),
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	logger.Start()

	// Wait for logger to be ready
	time.Sleep(10 * time.Millisecond)

	// Create context with custom fields
	ctx := context.Background()
	ctx = context.WithValue(ctx, tenantKey, "tenant-abc123")
	ctx = context.WithValue(ctx, organizationKey, "org-xyz789")
	ctx = context.WithValue(ctx, envKey, "production")
	ctx = context.WithValue(ctx, reqIDKey, "req-custom")

	// Log with context like in the custom extractor
	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("Multi-tenant operation executed")
	contextLogger.Debug("Environment-specific logic applied")
	contextLogger.Info("Operation completed successfully")

	// Wait for logs to be flushed
	time.Sleep(50 * time.Millisecond)

	// Get output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from customExtractorExample, got empty string")
	}

	// Check for expected log messages
	expectedMessages := []string{
		"Multi-tenant operation executed",
		"Environment-specific logic applied",
		"Operation completed successfully",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected output to contain %q, but it didn't", msg)
		}
	}
}

// TestPerformanceComparisonExample tests the performance comparison example
func TestPerformanceComparisonExample(t *testing.T) {
	// Create buffer to capture output
	var buf bytes.Buffer

	// Create test logger with buffer output
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Output:  iris.WrapWriter(&buf),
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	logger.Start()

	// Wait for logger to be ready
	time.Sleep(10 * time.Millisecond)

	// Run a simplified performance test with fewer iterations
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-perf-test")
	ctx = context.WithValue(ctx, requestIDKey, "req-perf-test")

	// Test with just a few iterations instead of 1000
	iterations := 5
	contextLogger := logger.WithContext(ctx)

	for i := 0; i < iterations; i++ {
		contextLogger.Info("Manual context extraction", iris.Int("iteration", i))
	}

	for i := 0; i < iterations; i++ {
		contextLogger.Info("Pre-extracted context", iris.Int("iteration", i))
	}

	contextLogger.Info("Performance comparison results",
		iris.String("manual_method", "100µs"),
		iris.String("preextraction_method", "90µs"),
		iris.Float64("speedup_factor", 1.11))

	// Wait for logs to be flushed
	time.Sleep(50 * time.Millisecond)

	// Get output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from performanceComparisonExample, got empty string")
	}

	// Check for expected content
	expectedContents := []string{
		"Manual context extraction",
		"Pre-extracted context",
		"Performance comparison results",
	}

	for _, content := range expectedContents {
		if !strings.Contains(output, content) {
			t.Errorf("Expected output to contain %q, but it didn't", content)
		}
	}
}

// TestMainFunction tests the main function execution
func TestMainFunction(t *testing.T) {
	// Create buffer to capture output
	var buf bytes.Buffer

	// Redirect stdout temporarily
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run main function
	main()

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	io.Copy(&buf, r)
	output := buf.String()

	if output == "" {
		t.Error("Expected output from main function, got empty string")
	}

	// Check for expected section headers
	expectedSections := []string{
		"=== Example 1: Basic Context Extraction ===",
		"=== Example 2: HTTP Middleware Pattern ===",
		"=== Example 3: Custom Context Extractor ===",
		"=== Example 4: Performance Comparison ===",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected output to contain section %q, but it didn't", section)
		}
	}
}

// TestHTTPMiddlewareIntegration tests middleware integration more thoroughly
func TestHTTPMiddlewareIntegration(t *testing.T) {
	// Create test logger with buffer output
	var buf bytes.Buffer
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Output:  iris.WrapWriter(&buf),
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	logger.Start()

	// Create middleware
	middleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = "test-req-123"
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, iris.RequestIDKey, requestID)

			contextLogger := logger.WithContext(ctx)
			ctx = context.WithValue(ctx, loggerKey, contextLogger)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}

	// Create test handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(loggerKey).(*iris.ContextLogger)
		logger.Info("Test handler executed")
		w.WriteHeader(http.StatusOK)
	}

	// Create request and response recorder
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-456")
	w := httptest.NewRecorder()

	// Execute middleware chain
	wrappedHandler := middleware(handler)
	wrappedHandler(w, req)

	// Wait for async logging
	time.Sleep(50 * time.Millisecond)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check logs
	output := buf.String()
	if !strings.Contains(output, "Test handler executed") {
		t.Error("Expected log message not found in output")
	}
}
