// zap_comparison_test.go: Scientific benchmarks comparing Iris vs Zap patterns
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
	"time"
)

// Scientific comparison with Zap-equivalent functionality
// These benchmarks replicate exact Zap usage patterns for fair comparison

// 1. Simple message (equivalent to zap.Info("message"))
func BenchmarkIris_SimpleLog(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 4096, // Standard buffer size
		BatchSize:  64,   // Standard batch size
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Simple log message")
	}
}

// 2. Structured log with 3 fields (equivalent to zap structured logging)
func BenchmarkIris_StructuredLog_3Fields(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 4096,
		BatchSize:  64,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("User action performed",
			Str("user", "john.doe"),
			Str("action", "login"),
			Int("duration_ms", 245),
		)
	}
}

// 3. Structured log with 5 fields (common Zap pattern)
func BenchmarkIris_StructuredLog_5Fields(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 4096,
		BatchSize:  64,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Request processed",
			Str("method", "GET"),
			Str("path", "/api/users"),
			Int("status", 200),
			Duration("duration", 150*time.Millisecond),
			Str("request_id", "req-12345"),
		)
	}
}

// 4. Error log with error field (equivalent to zap.Error)
func BenchmarkIris_ErrorLog(b *testing.B) {
	logger, _ := New(Config{
		Level:      DebugLevel, // Allow all levels
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 4096,
		BatchSize:  64,
	})
	defer logger.Close()

	testErr := Error(nil) // Create a test error field

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Error("Database connection failed",
			Str("database", "users_db"),
			Int("retry_count", 3),
			testErr,
		)
	}
}

// 5. Debug log (often filtered out)
func BenchmarkIris_DebugLog_Filtered(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel, // Debug will be filtered out
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 4096,
		BatchSize:  64,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Debug("This debug message will be filtered",
			Str("component", "auth"),
			Int("user_id", 12345),
		)
	}
}

// 6. High-volume production pattern
func BenchmarkIris_ProductionPattern(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 8192, // Production buffer size
		BatchSize:  128,  // Production batch size
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("HTTP request completed",
			Str("method", "POST"),
			Str("endpoint", "/api/v1/orders"),
			Int("status_code", 201),
			Duration("response_time", 89*time.Millisecond),
			Str("user_agent", "Mozilla/5.0"),
			Str("ip", "192.168.1.100"),
			Int("content_length", 1024),
			Str("trace_id", "trace-abc123"),
		)
	}
}

// 7. Mixed log levels (realistic usage)
func BenchmarkIris_MixedLevels(b *testing.B) {
	logger, _ := New(Config{
		Level:      DebugLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 4096,
		BatchSize:  64,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		switch i % 4 {
		case 0:
			logger.Debug("Debug message", Int("counter", i))
		case 1:
			logger.Info("Info message", Str("status", "ok"))
		case 2:
			logger.Warn("Warning message", Bool("retry", true))
		case 3:
			logger.Error("Error message", Str("error", "timeout"))
		}
	}
}

// 8. Zero-field message (pure message performance)
func BenchmarkIris_ZeroFieldsComparison(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 4096,
		BatchSize:  64,
	})
	defer logger.Close()

	message := "Application started successfully"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info(message)
	}
}

// 9. String concatenation pattern (what NOT to do, but often seen)
func BenchmarkIris_StringConcat(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 4096,
		BatchSize:  64,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// This is what people often do incorrectly instead of structured logging
		logger.Info("User john.doe performed action login in 245ms")
	}
}

// 10. Complex structured data (JSON-like)
func BenchmarkIris_ComplexStructured(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 4096,
		BatchSize:  64,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Complex operation completed",
			Str("service", "payment-processor"),
			Str("operation", "process_payment"),
			Str("payment_id", "pay_1234567890"),
			Str("merchant_id", "merch_abcdefgh"),
			Int("amount_cents", 15995),
			Str("currency", "USD"),
			Str("payment_method", "credit_card"),
			Bool("three_d_secure", true),
			Duration("processing_time", 1247*time.Millisecond),
			Float("risk_score", 0.23),
			Str("processor_response", "approved"),
			Str("auth_code", "123456"),
			Time("timestamp", time.Now()),
			Int("retry_attempt", 0),
			Bool("sandbox", false),
		)
	}
}
