// console_benchmark_test.go: Performance analysis benchmarks for console encoder optimizations
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
	"time"
)

// BenchmarkConsoleEncoderPerformance benchmarks the overall console encoder performance
func BenchmarkConsoleEncoderPerformance(b *testing.B) {
	encoder := NewConsoleEncoder(false) // No colors for baseline

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     InfoLevel,
		Message:   "Request processed successfully",
		Fields: []Field{
			Str("service", "user-api"),
			Int("user_id", 12345),
			Bool("authenticated", true),
			Float64("duration_ms", 123.456),
			Str("endpoint", "/api/v1/users"),
		},
		Caller: Caller{Valid: true, File: "/app/handlers/user.go", Line: 42},
	}

	buf := make([]byte, 0, 512)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := encoder.EncodeLogEntry(entry, buf[:0])
		_ = result // Prevent optimization
	}
}

// BenchmarkConsoleEncoderWithColors benchmarks colored output performance
func BenchmarkConsoleEncoderWithColors(b *testing.B) {
	encoder := NewConsoleEncoder(true) // With colors

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     ErrorLevel, // Error level gets bold formatting
		Message:   "Error processing request",
		Fields: []Field{
			Str("error", "database connection failed"),
			Int("attempts", 3),
			Bool("fatal", true),
		},
		Caller: Caller{Valid: true, File: "/app/db/conn.go", Line: 156},
	}

	buf := make([]byte, 0, 512)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := encoder.EncodeLogEntry(entry, buf[:0])
		_ = result
	}
}

// BenchmarkConsoleEncoderFieldTypes benchmarks different field type performance
func BenchmarkConsoleEncoderFieldTypes(b *testing.B) {
	encoder := NewConsoleEncoder(false)
	buf := make([]byte, 0, 512)

	tests := []struct {
		name  string
		field Field
	}{
		{"String", Str("message", "test string value")},
		{"Int", Int("count", 42)},
		{"Int64", Int64("large_count", 9876543210)},
		{"Bool", Bool("enabled", true)},
		{"Float64", Float64("score", 98.765)},
		{"Duration", Duration("timeout", 30*time.Second)},
		{"ByteString", ByteString("data", []byte("hello world"))},
		{"Binary", Binary("blob", []byte("binary data content"))},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			entry := &LogEntry{
				Level:   InfoLevel,
				Message: "field test",
				Fields:  []Field{test.field},
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				result := encoder.EncodeLogEntry(entry, buf[:0])
				_ = result
			}
		})
	}
}

// BenchmarkConsoleEncoderStringQuoting benchmarks string quoting performance
func BenchmarkConsoleEncoderStringQuoting(b *testing.B) {
	encoder := NewConsoleEncoder(false)
	buf := make([]byte, 0, 512)

	tests := []struct {
		name  string
		value string
	}{
		{"NoQuoting", "simple_string_no_spaces"},
		{"WithSpaces", "string with spaces"},
		{"WithQuotes", `string with "quotes" inside`},
		{"WithSpecialChars", "string\nwith\tspecial\rcharacters"},
		{"Complex", `complex "string\nwith\tmixed\r\nspecial\\characters"`},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			entry := &LogEntry{
				Level:   InfoLevel,
				Message: "quoting test",
				Fields:  []Field{Str("test", test.value)},
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				result := encoder.EncodeLogEntry(entry, buf[:0])
				_ = result
			}
		})
	}
}

// BenchmarkConsoleEncoderBufferSizes benchmarks performance with different data sizes
func BenchmarkConsoleEncoderBufferSizes(b *testing.B) {
	encoder := NewConsoleEncoder(false)

	// Small data
	b.Run("SmallData", func(b *testing.B) {
		entry := &LogEntry{
			Level:   InfoLevel,
			Message: "small",
			Fields: []Field{
				Str("msg", "short"),
				Int("id", 1),
			},
		}
		buf := make([]byte, 0, 128)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result := encoder.EncodeLogEntry(entry, buf[:0])
			_ = result
		}
	})

	// Medium data
	b.Run("MediumData", func(b *testing.B) {
		fields := make([]Field, 10)
		for i := 0; i < 10; i++ {
			fields[i] = Str("field_"+string(rune('0'+i)), "medium_length_value_for_testing")
		}

		entry := &LogEntry{
			Timestamp: time.Now(),
			Level:     InfoLevel,
			Message:   "medium data test message",
			Fields:    fields,
			Caller:    Caller{Valid: true, File: "test.go", Line: 123},
		}
		buf := make([]byte, 0, 1024)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result := encoder.EncodeLogEntry(entry, buf[:0])
			_ = result
		}
	})

	// Large data
	b.Run("LargeData", func(b *testing.B) {
		fields := make([]Field, 25)
		for i := 0; i < 25; i++ {
			fields[i] = Str("field_"+string(rune('a'+i%26)), "this_is_a_longer_value_for_testing_large_data_scenarios_with_many_fields")
		}

		stackTrace := "goroutine 1 [running]:\nmain.main()\n\t/path/to/main.go:42 +0x123\nruntime.main()\n\t/usr/local/go/src/runtime/proc.go:250 +0x456"

		entry := &LogEntry{
			Timestamp:  time.Now(),
			Level:      ErrorLevel,
			Message:    "large data test with stack trace and many fields",
			Fields:     fields,
			Caller:     Caller{Valid: true, File: "/long/path/to/file.go", Line: 9999},
			StackTrace: stackTrace,
		}
		buf := make([]byte, 0, 4096)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result := encoder.EncodeLogEntry(entry, buf[:0])
			_ = result
		}
	})
}

// BenchmarkConsoleEncoderLevelFormatting benchmarks level formatting performance
func BenchmarkConsoleEncoderLevelFormatting(b *testing.B) {
	tests := []struct {
		name  string
		level Level
		color bool
	}{
		{"Debug_NoColor", DebugLevel, false},
		{"Info_NoColor", InfoLevel, false},
		{"Warn_NoColor", WarnLevel, false},
		{"Error_NoColor", ErrorLevel, false},
		{"Debug_WithColor", DebugLevel, true},
		{"Info_WithColor", InfoLevel, true},
		{"Warn_WithColor", WarnLevel, true},
		{"Error_WithColor", ErrorLevel, true},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			encoder := NewConsoleEncoder(test.color)
			entry := &LogEntry{
				Level:   test.level,
				Message: "level formatting test",
				Fields:  []Field{Str("test", "value")},
			}
			buf := make([]byte, 0, 256)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				result := encoder.EncodeLogEntry(entry, buf[:0])
				_ = result
			}
		})
	}
}

// BenchmarkConsoleEncoderTimestamp benchmarks timestamp formatting performance
func BenchmarkConsoleEncoderTimestamp(b *testing.B) {
	encoder := NewConsoleEncoder(false)
	timestamp := time.Now()
	buf := make([]byte, 0, 256)

	// With timestamp
	b.Run("WithTimestamp", func(b *testing.B) {
		entry := &LogEntry{
			Timestamp: timestamp,
			Level:     InfoLevel,
			Message:   "timestamp test",
			Fields:    []Field{Str("test", "value")},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result := encoder.EncodeLogEntry(entry, buf[:0])
			_ = result
		}
	})

	// Without timestamp
	b.Run("WithoutTimestamp", func(b *testing.B) {
		entry := &LogEntry{
			Level:   InfoLevel,
			Message: "no timestamp test",
			Fields:  []Field{Str("test", "value")},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result := encoder.EncodeLogEntry(entry, buf[:0])
			_ = result
		}
	})
}

// BenchmarkConsoleEncoderCaller benchmarks caller formatting performance
func BenchmarkConsoleEncoderCaller(b *testing.B) {
	encoder := NewConsoleEncoder(false)
	buf := make([]byte, 0, 256)

	// With caller
	b.Run("WithCaller", func(b *testing.B) {
		entry := &LogEntry{
			Level:   InfoLevel,
			Message: "caller test",
			Fields:  []Field{Str("test", "value")},
			Caller:  Caller{Valid: true, File: "/long/path/to/file.go", Line: 42},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result := encoder.EncodeLogEntry(entry, buf[:0])
			_ = result
		}
	})

	// Without caller
	b.Run("WithoutCaller", func(b *testing.B) {
		entry := &LogEntry{
			Level:   InfoLevel,
			Message: "no caller test",
			Fields:  []Field{Str("test", "value")},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result := encoder.EncodeLogEntry(entry, buf[:0])
			_ = result
		}
	})
}
