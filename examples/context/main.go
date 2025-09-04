// examples/context_integration.go: Demonstrates context.Context integration
//
// This example shows how to use IRIS context integration for distributed
// tracing and request correlation in web applications.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/agilira/iris"
)

// Context keys to avoid linter warnings (specific to main.go examples)
const (
	httpMethodKey   = iris.ContextKey("method")
	httpPathKey     = iris.ContextKey("path")
	httpLoggerKey   = iris.ContextKey("logger")
	httpTenantIDKey = iris.ContextKey("tenant_id")
	httpOrgIDKey    = iris.ContextKey("organization_id")
	httpEnvKey      = iris.ContextKey("environment")
)

func main() {
	// Create logger
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Output:  iris.WrapWriter(os.Stdout),
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Start()

	// Example 1: Basic context extraction
	fmt.Println("=== Example 1: Basic Context Extraction ===")
	basicContextExample(logger)

	// Example 2: HTTP middleware pattern
	fmt.Println("\n=== Example 2: HTTP Middleware Pattern ===")
	httpMiddlewareExample(logger)

	// Example 3: Custom context extractor
	fmt.Println("\n=== Example 3: Custom Context Extractor ===")
	customExtractorExample(logger)

	// Example 4: Performance comparison
	fmt.Println("\n=== Example 4: Performance Comparison ===")
	performanceComparisonExample(logger)
}

func basicContextExample(logger *iris.Logger) {
	// Create context with request information
	ctx := context.Background()
	ctx = context.WithValue(ctx, iris.RequestIDKey, "req-12345")
	ctx = context.WithValue(ctx, iris.UserIDKey, "user-67890")
	ctx = context.WithValue(ctx, iris.SessionIDKey, "sess-abcdef")

	// Extract context once, use many times (optimized!)
	contextLogger := logger.WithContext(ctx)

	// All logs include context fields automatically
	contextLogger.Info("User authentication started")
	contextLogger.Debug("Validating credentials")
	contextLogger.Info("Authentication successful")
	contextLogger.Debug("Creating session")
	contextLogger.Info("Login process completed")

	// Fast methods for single fields
	requestLogger := logger.WithRequestID(ctx)
	requestLogger.Info("Request processed with fast method")
}

func httpMiddlewareExample(logger *iris.Logger) {
	// Simulate HTTP middleware that adds context
	middleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract or generate request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
			}

			// Add to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, iris.RequestIDKey, requestID)
			ctx = context.WithValue(ctx, httpMethodKey, r.Method)
			ctx = context.WithValue(ctx, httpPathKey, r.URL.Path)

			// Create context logger once
			contextLogger := logger.WithContext(ctx)

			// Store in context for handlers
			ctx = context.WithValue(ctx, httpLoggerKey, contextLogger)

			// Call next handler with enriched context
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}

	// Simulate handler that uses context logger
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Get logger from context
		logger := r.Context().Value(httpLoggerKey).(*iris.ContextLogger)

		logger.Info("Handler called")
		logger.Debug("Processing business logic")
		logger.Info("Handler completed")
	}

	// Simulate request processing
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("X-Request-ID", "req-http-example")

	// Process through middleware
	wrappedHandler := middleware(handler)
	wrappedHandler(nil, req)
}

func customExtractorExample(logger *iris.Logger) {
	// Create context with custom values
	ctx := context.Background()
	ctx = context.WithValue(ctx, iris.RequestIDKey, "req-custom")
	ctx = context.WithValue(ctx, httpTenantIDKey, "tenant-abc123")
	ctx = context.WithValue(ctx, httpOrgIDKey, "org-xyz789")
	ctx = context.WithValue(ctx, httpEnvKey, "production")

	// Create custom extractor with field renaming
	extractor := &iris.ContextExtractor{
		Keys: map[iris.ContextKey]string{
			iris.RequestIDKey: "req_id",       // Rename field
			httpTenantIDKey:   "tenant",       // Custom key
			httpOrgIDKey:      "organization", // Another custom key
			httpEnvKey:        "env",          // Short name
		},
		MaxDepth: 10, // Limit context traversal
	}

	// Use custom extractor
	contextLogger := logger.WithContextExtractor(ctx, extractor)

	contextLogger.Info("Multi-tenant operation executed")
	contextLogger.Debug("Environment-specific logic applied")
	contextLogger.Info("Operation completed successfully")
}

func performanceComparisonExample(logger *iris.Logger) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, iris.RequestIDKey, "req-perf-test")
	ctx = context.WithValue(ctx, iris.UserIDKey, "user-perf-test")

	// Method 1: Inefficient - repeated context.Value() calls
	fmt.Println("Method 1: Manual context.Value() calls (inefficient)")
	start := time.Now()
	for i := 0; i < 1000; i++ {
		logger.Info("Manual context extraction",
			iris.Str("request_id", ctx.Value(iris.RequestIDKey).(string)),
			iris.Str("user_id", ctx.Value(iris.UserIDKey).(string)),
			iris.Int("iteration", i),
		)
	}
	elapsed1 := time.Since(start)

	// Method 2: Efficient - pre-extraction with caching
	fmt.Println("Method 2: Context pre-extraction (efficient)")
	start = time.Now()
	contextLogger := logger.WithContext(ctx) // Extract once
	for i := 0; i < 1000; i++ {
		contextLogger.Info("Pre-extracted context", // Use cached fields
			iris.Int("iteration", i),
		)
	}
	elapsed2 := time.Since(start)

	logger.Info("Performance comparison results",
		iris.Str("manual_method", elapsed1.String()),
		iris.Str("preextraction_method", elapsed2.String()),
		iris.Float64("speedup_factor", float64(elapsed1)/float64(elapsed2)),
	)
}
