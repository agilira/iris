// example_test.go: Example usage of Iris logger
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris_test

import (
	"github.com/agilira/iris"
)

func ExampleNew_basic() {
	// Create logger with eXtreme performance Designer API
	logger := iris.NewIrisLogger(iris.InfoLevel)

	// Basic logging with eXtreme fields
	baseCtx := logger.XFields(
		iris.XStr("app", "iris-demo"),
		iris.XStr("version", "2.0.0"),
	)

	baseCtx.Info("Application started")

	// User login context with eXtreme performance
	userCtx := logger.XFields(
		iris.XStr("user", "john"),
		iris.XInt("id", 123),
		iris.XBool("authenticated", true),
	)
	userCtx.Info("User logged in")

	// Performance monitoring with eXtreme fields
	perfCtx := logger.XFields(
		iris.XStr("metric", "memory_usage"),
		iris.XInt("usage_percent", 85),
	)
	perfCtx.Warn("High memory usage")

	// Request processing with eXtreme structured logging
	reqCtx := logger.XFields(
		iris.XStr("method", "GET"),
		iris.XStr("path", "/api/users"),
		iris.XInt("status", 200),
		iris.XInt("duration_ms", 50),
	)
	reqCtx.Info("Request processed")

	// Output:
}

func ExampleNew_performance() {
	// Create eXtreme performance logger
	logger := iris.NewIrisLogger(iris.InfoLevel)

	// Pre-create context for maximum performance (reuse pattern)
	ctx := logger.XFields(
		iris.XStr("operation", "benchmark"),
		iris.XStr("type", "high_frequency"),
	)

	// High-frequency logging with eXtreme performance - reuse context!
	for i := 0; i < 1000000; i++ {
		// Super fast: reuse the same context, only message changes
		ctx.Info("High frequency log")
	}

	// Alternative: if you need iteration number (slightly slower)
	for i := 0; i < 100; i++ {
		iterCtx := logger.XFields(
			iris.XStr("operation", "numbered_benchmark"),
			iris.XInt("iteration", int64(i)),
		)
		iterCtx.Info("Numbered iteration log")
	}

	// Output:
}

// ExampleNewIrisLogger demonstrates the new Designer API
func ExampleNewIrisLogger_basic() {
	// Create logger with beautiful Iris-branded API
	logger := iris.NewIrisLogger(iris.InfoLevel)

	// Simple context with eXtreme performance fields
	ctx := logger.XFields(
		iris.XStr("service", "auth-api"),
		iris.XInt("port", 8080),
		iris.XBool("ssl_enabled", true),
	)

	// Log with the beautiful new API - all levels supported!
	ctx.Info("Service started")
	ctx.InfoWithCaller("Service ready with caller info")
	ctx.Warn("High memory usage detected")
	ctx.Error("Database connection timeout")

	// Output:
}

// ExampleNewIrisLogger_advanced shows advanced Designer API usage
func ExampleNewIrisLogger_advanced() {
	logger := iris.NewIrisLogger(iris.InfoLevel)

	// Pre-create context for reuse (high-performance pattern)
	baseCtx := logger.XFields(
		iris.XStr("app", "iris-demo"),
		iris.XStr("version", "2.0.0"),
	)

	// Create specialized contexts
	authCtx := logger.XFields(
		iris.XStr("module", "auth"),
		iris.XBool("rate_limited", false),
	)

	userCtx := logger.XFields(
		iris.XStr("module", "user"),
		iris.XInt("active_sessions", 1024),
	)

	// Log with different contexts
	baseCtx.Info("Application initialized")
	authCtx.Info("Authentication module ready")
	userCtx.Info("User module ready")

	// Output:
}

// ExampleNewIrisLogger_fieldReuse shows optimal field reuse patterns
func ExampleNewIrisLogger_fieldReuse() {
	logger := iris.NewIrisLogger(iris.InfoLevel)

	// Create fields once, reuse multiple times (optimal performance)
	serviceField := iris.XStr("service", "payment-gateway")
	versionField := iris.XStr("version", "v1.2.3")

	// Use in different contexts
	startupCtx := logger.XFields(serviceField, versionField, iris.XStr("phase", "startup"))
	runtimeCtx := logger.XFields(serviceField, versionField, iris.XStr("phase", "runtime"))

	startupCtx.Info("Service initializing")
	runtimeCtx.Info("Processing payments")

	// Critical: Release buffers when done (for ultra-high performance)
	serviceField.Release()
	versionField.Release()

	// Output:
}
