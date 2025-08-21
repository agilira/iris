// binary_logger_benchmark_test.go: Performance analysis benchmarks for binary logger optimizations
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"testing"
)

// BenchmarkBinaryLoggerCreation benchmarks binary logger instantiation
func BenchmarkBinaryLoggerCreation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger := NewBinaryLogger(InfoLevel)
		_ = logger // Prevent optimization
	}
}

// BenchmarkBinaryFieldCreation benchmarks direct binary field creation
func BenchmarkBinaryFieldCreation(b *testing.B) {
	key := "test_key"
	value := "test_value"

	b.Run("BinaryStr", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			field := BinaryStr(key, value)
			_ = field
		}
	})

	b.Run("BinaryInt", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			field := BinaryInt(key, 42)
			_ = field
		}
	})

	b.Run("BinaryBool", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			field := BinaryBool(key, true)
			_ = field
		}
	})
}

// BenchmarkBinaryContextCreation benchmarks context creation performance
func BenchmarkBinaryContextCreation(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.Run("WithBinaryFields", func(b *testing.B) {
		fields := []BinaryField{
			BinaryStr("service", "user-api"),
			BinaryInt("user_id", 12345),
			BinaryBool("authenticated", true),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := logger.WithBinaryFields(fields...)
			_ = ctx
		}
	})

	b.Run("WithLegacyFields", func(b *testing.B) {
		fields := []Field{
			Str("service", "user-api"),
			Int("user_id", 12345),
			Bool("authenticated", true),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := logger.WithFields(fields...)
			_ = ctx
		}
	})
}

// BenchmarkBinaryContextLogging benchmarks actual logging performance
func BenchmarkBinaryContextLogging(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)
	message := "Request processed successfully"

	b.Run("InfoWithoutFields", func(b *testing.B) {
		ctx := logger.WithBinaryFields()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx.Info(message)
		}
	})

	b.Run("InfoWithBinaryFields", func(b *testing.B) {
		fields := []BinaryField{
			BinaryStr("service", "user-api"),
			BinaryInt("user_id", 12345),
			BinaryBool("authenticated", true),
			BinaryStr("endpoint", "/api/v1/users"),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := logger.WithBinaryFields(fields...)
			ctx.Info(message)
		}
	})

	b.Run("InfoWithLegacyFields", func(b *testing.B) {
		fields := []Field{
			Str("service", "user-api"),
			Int("user_id", 12345),
			Bool("authenticated", true),
			Str("endpoint", "/api/v1/users"),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := logger.WithFields(fields...)
			ctx.Info(message)
		}
	})
}

// BenchmarkBinaryContextLoggingWithCaller benchmarks logging with caller info
func BenchmarkBinaryContextLoggingWithCaller(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)
	message := "Operation completed"

	b.Run("InfoWithCaller", func(b *testing.B) {
		fields := []BinaryField{
			BinaryStr("operation", "test"),
			BinaryInt("duration_ms", 150),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := logger.WithBinaryFields(fields...)
			ctx.InfoWithCaller(message)
		}
	})
}

// BenchmarkBinaryFieldTypes benchmarks different field type performance
func BenchmarkBinaryFieldTypes(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)
	message := "field type test"

	tests := []struct {
		name  string
		field BinaryField
	}{
		{"String", BinaryStr("message", "test string value")},
		{"Int", BinaryInt("count", 42)},
		{"Bool", BinaryBool("enabled", true)},
		{"LongString", BinaryStr("description", "this is a longer string value for testing performance with larger data")},
		{"LargeInt", BinaryInt("large_count", 9876543210)},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ctx := logger.WithBinaryFields(test.field)
				ctx.Info(message)
			}
		})
	}
}

// BenchmarkBinaryLoggerScaling benchmarks performance with different numbers of fields
func BenchmarkBinaryLoggerScaling(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)
	message := "scaling test message"

	// Test with different numbers of fields
	fieldCounts := []int{1, 5, 10, 20}

	for _, count := range fieldCounts {
		b.Run(fmt.Sprintf("Fields_%d", count), func(b *testing.B) {
			// Pre-create fields
			fields := make([]BinaryField, count)
			for i := 0; i < count; i++ {
				fields[i] = BinaryStr(fmt.Sprintf("field_%d", i), fmt.Sprintf("value_%d", i))
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ctx := logger.WithBinaryFields(fields...)
				ctx.Info(message)
			}
		})
	}
}

// BenchmarkLazyCallerOperations benchmarks lazy caller performance
func BenchmarkLazyCallerOperations(b *testing.B) {
	b.Run("Creation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			caller := NewLazyCaller(2)
			_ = caller
		}
	})

	b.Run("FirstAccess", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			caller := NewLazyCaller(2)
			_ = caller.File() // First access triggers computation
		}
	})

	b.Run("CachedAccess", func(b *testing.B) {
		caller := NewLazyCaller(2)
		_ = caller.File() // Pre-compute

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = caller.File() // Should use cached value
		}
	})

	b.Run("PoolOperations", func(b *testing.B) {
		pool := NewLazyCallerPool()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			caller := pool.GetLazyCaller(2)
			pool.ReleaseLazyCaller(caller)
		}
	})
}

// BenchmarkBinaryMemoryOperations benchmarks memory-related operations
func BenchmarkBinaryMemoryOperations(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.Run("MemoryFootprint", func(b *testing.B) {
		ctx := logger.WithBinaryFields(
			BinaryStr("service", "api"),
			BinaryInt("port", 8080),
			BinaryBool("secure", true),
		)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = ctx.MemoryFootprint()
		}
	})

	b.Run("GetBinarySize", func(b *testing.B) {
		ctx := logger.WithBinaryFields(
			BinaryStr("service", "api"),
			BinaryInt("port", 8080),
		)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = ctx.GetBinarySize()
		}
	})
}

// BenchmarkBinaryLoggerLevelFiltering benchmarks level filtering performance
func BenchmarkBinaryLoggerLevelFiltering(b *testing.B) {
	// Create logger with WARN level (should filter INFO messages)
	logger := NewBinaryLogger(WarnLevel)
	message := "filtered message"

	b.Run("FilteredInfo", func(b *testing.B) {
		fields := []BinaryField{
			BinaryStr("test", "value"),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := logger.WithBinaryFields(fields...)
			ctx.Info(message) // Should be filtered out
		}
	})
}

// BenchmarkBinaryLoggerConcurrent benchmarks concurrent access performance
func BenchmarkBinaryLoggerConcurrent(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)
	message := "concurrent message"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := logger.WithBinaryFields(
				BinaryStr("worker", "test"),
				BinaryInt("iteration", 1),
			)
			ctx.Info(message)
		}
	})
}

// BenchmarkBinaryVsLegacyComparison compares binary vs legacy field performance
func BenchmarkBinaryVsLegacyComparison(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)
	message := "comparison test"

	b.Run("BinaryFields", func(b *testing.B) {
		fields := []BinaryField{
			BinaryStr("service", "user-api"),
			BinaryInt("user_id", 12345),
			BinaryBool("authenticated", true),
			BinaryStr("endpoint", "/api/v1/users"),
			BinaryInt("response_time_ms", 234),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := logger.WithBinaryFields(fields...)
			ctx.Info(message)
		}
	})

	b.Run("LegacyFields", func(b *testing.B) {
		fields := []Field{
			Str("service", "user-api"),
			Int("user_id", 12345),
			Bool("authenticated", true),
			Str("endpoint", "/api/v1/users"),
			Int("response_time_ms", 234),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			ctx := logger.WithFields(fields...)
			ctx.Info(message)
		}
	})
}
