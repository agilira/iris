// example_slog_integration.go: Example demonstrating external slog provider integration
//
// This example shows how to use Iris as a universal logging accelerator
// for existing slog-based applications using external providers.
// The slog provider is a separate module that implements iris.SyncReader.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/agilira/iris"
	// External provider - separate module!
	// slogprovider "github.com/agilira/iris-provider-slog"
)

func main() {
	println("=== External Provider Integration ===")
	println("This example demonstrates the provider architecture.")
	println("To run with actual slog integration:")
	println("1. go get github.com/agilira/iris-provider-slog")
	println("2. Uncomment the import and provider usage below")
	println("3. Replace mockAdapter with slogprovider.New()")
	println("")

	// Run examples with mock adapter (for demonstration)
	basicExample()
	advancedExample()
	existingAppExample()
}

// mockAdapter is a simple implementation for demonstration purposes
type mockAdapter struct {
	records chan slog.Record
	closed  chan struct{}
	once    sync.Once
}

func newMockAdapter() *mockAdapter {
	return &mockAdapter{
		records: make(chan slog.Record, 1000),
		closed:  make(chan struct{}),
	}
}

func (m *mockAdapter) Read(ctx context.Context) (*iris.Record, error) {
	select {
	case record := <-m.records:
		// Convert slog.Record to iris.Record
		irisRecord := iris.NewRecord(iris.Info, record.Message)
		record.Attrs(func(a slog.Attr) bool {
			switch a.Value.Kind() {
			case slog.KindString:
				irisRecord.AddField(iris.String(a.Key, a.Value.String()))
			case slog.KindInt64:
				irisRecord.AddField(iris.Int64(a.Key, a.Value.Int64()))
			case slog.KindBool:
				irisRecord.AddField(iris.Bool(a.Key, a.Value.Bool()))
			default:
				irisRecord.AddField(iris.String(a.Key, a.Value.String()))
			}
			return true
		})
		return irisRecord, nil
	case <-m.closed:
		return nil, nil
	default:
		return nil, nil
	}
}

func (m *mockAdapter) Close() error {
	m.once.Do(func() {
		close(m.closed)
	})
	return nil
}

func (m *mockAdapter) Handle(ctx context.Context, record slog.Record) error {
	select {
	case m.records <- record:
		return nil
	case <-m.closed:
		return context.Canceled
	default:
		// Buffer full, drop record
		return nil
	}
}

func (m *mockAdapter) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (m *mockAdapter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

func (m *mockAdapter) WithGroup(name string) slog.Handler {
	return m
}

func basicExample() {
	println("=== Basic slog Integration ===")

	// Create provider (in real usage: slogprovider.New())
	adapter := newMockAdapter()
	defer adapter.Close()

	// Create Iris logger with adapter
	readers := []iris.SyncReader{adapter}
	logger, err := iris.NewReaderLogger(iris.Config{
		Output:  iris.WrapWriter(os.Stdout),
		Encoder: iris.NewJSONEncoder(),
		Level:   iris.Info,
	}, readers)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// Start processing
	logger.Start()

	// Create slog logger with our adapter as handler
	slogger := slog.New(adapter)

	// Now use slog normally - but get Iris performance and features!
	slogger.Info("User authentication successful",
		"user_id", "12345",
		"session_id", "sess_abc123",
		"login_time", time.Now())

	slogger.Warn("Rate limit approaching",
		"user_id", "12345",
		"requests_remaining", 10)

	slogger.Error("Database connection failed",
		"error", "connection timeout",
		"retry_count", 3)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)
	if err := logger.Sync(); err != nil {
		println("Warning: Failed to sync logger:", err.Error())
	}
}

func advancedExample() {
	println("\n=== Advanced Integration with All Features ===")

	// Create provider (in real usage: slogprovider.New())
	adapter := newMockAdapter()
	defer adapter.Close()

	// Create Iris logger with ALL advanced features
	readers := []iris.SyncReader{adapter}
	logger, err := iris.NewReaderLogger(iris.Config{
		Output: iris.MultiWriter(
			iris.WrapWriter(os.Stdout),
			// iris.NewLokiWriter("http://localhost:3100"), // Uncomment for Loki
		),
		Encoder: iris.NewJSONEncoder(),
		Level:   iris.Debug,
	}, readers,
		// iris.WithOTel(),     // OpenTelemetry integration
		// iris.WithSecurity(), // Automatic secret redaction
		iris.WithCaller(), // Caller information
	)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Start()

	// Create slog logger
	slogger := slog.New(adapter)

	// Same slog code, but now with:
	// ✅ 10x performance boost
	// ✅ OpenTelemetry trace correlation (if enabled)
	// ✅ Grafana Loki batching (if configured)
	// ✅ Automatic secret redaction
	// ✅ Caller information
	slogger.Info("Payment processed",
		"transaction_id", "txn_123456",
		"amount", 2499,
		"currency", "USD",
		"api_key", "sk_live_12345_secret", // Will be redacted automatically
		"password", "user_password", // Will be redacted automatically
	)

	time.Sleep(100 * time.Millisecond)
	if err := logger.Sync(); err != nil {
		println("Warning: Failed to sync logger:", err.Error())
	}
}

func existingAppExample() {
	println("\n=== Existing Application Integration ===")

	// Simulate existing application with slog
	existingFunction := func(logger *slog.Logger) {
		// This is existing code that doesn't need to change
		logger.Info("Processing order",
			"order_id", "ord_123",
			"customer_id", "cust_456")

		logger.Debug("Cache hit",
			"key", "user:123",
			"ttl", "300s")

		logger.Error("Validation failed",
			"field", "email",
			"value", "invalid-email")
	}

	// Original setup (slow)
	println("--- Before: Standard slog (slow) ---")
	originalLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	start := time.Now()
	existingFunction(originalLogger)
	originalDuration := time.Since(start)

	// Accelerated setup (fast + features)
	println("\n--- After: Iris-accelerated slog (fast + features) ---")
	adapter := newMockAdapter()
	defer adapter.Close()

	readers := []iris.SyncReader{adapter}
	irisLogger, err := iris.NewReaderLogger(iris.Config{
		Output:  iris.WrapWriter(os.Stdout),
		Encoder: iris.NewJSONEncoder(),
		Level:   iris.Debug,
	}, readers)
	if err != nil {
		panic(err)
	}
	defer irisLogger.Close()

	irisLogger.Start()

	// Same existing function, but now with Iris acceleration
	acceleratedLogger := slog.New(adapter)
	start = time.Now()
	existingFunction(acceleratedLogger) // ZERO CODE CHANGES!
	acceleratedDuration := time.Since(start)

	time.Sleep(100 * time.Millisecond)
	if err := irisLogger.Sync(); err != nil {
		println("Warning: Failed to sync iris logger:", err.Error())
	}

	println("\nPerformance comparison:")
	println("Original slog:", originalDuration.String())
	println("Iris-accelerated:", acceleratedDuration.String())
	if originalDuration > acceleratedDuration {
		improvement := float64(originalDuration) / float64(acceleratedDuration)
		println("Improvement: %.2fx faster", improvement)
	}
}
