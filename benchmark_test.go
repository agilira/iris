// benchmark_test.go: Advanced benchmarks for Iris vs Zap comparison
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"testing"
	"time"
)

// discard writer for pure performance testing
type discardWriter struct{}

func (d discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

var Discard = discardWriter{}

func BenchmarkIris_SimpleMessage(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 16384,
		BatchSize:  256,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Simple benchmark message")
	}
}

func BenchmarkIris_WithFields(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 16384,
		BatchSize:  256,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark with fields",
			Str("service", "benchmark"),
			Int("iteration", i),
			Bool("success", true),
		)
	}
}

func BenchmarkIris_ManyFields(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 16384,
		BatchSize:  256,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark with many fields",
			Str("service", "benchmark"),
			Str("component", "logger"),
			Str("method", "BenchmarkIris_ManyFields"),
			Int("iteration", i),
			Int("worker_id", i%10),
			Float("duration", 1.23),
			Bool("success", true),
			Duration("elapsed", 100*time.Millisecond),
		)
	}
}

func BenchmarkIris_HighThroughput(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 32768, // Very large buffer
		BatchSize:  512,   // Large batches
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(1) // Force single-threaded for SPSC

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.Info("High throughput message",
				Str("worker", "single"),
				Int("msg_id", i),
			)
			i++
		}
	})
}

func BenchmarkIris_ZeroFields(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(Discard),
		BufferSize: 16384,
		BatchSize:  256,
	})
	defer logger.Close()

	message := "This is a zero-field benchmark message"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info(message)
	}
}

// Benchmark encoder directly for raw performance
func BenchmarkEncoder_Direct(b *testing.B) {
	encoder := NewJSONEncoder()
	timestamp := time.Now()
	fields := []Field{
		Str("service", "test"),
		Int("id", 123),
		Bool("active", true),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		encoder.EncodeLogEntry(timestamp, InfoLevel, "Benchmark message", fields, Caller{Valid: false}, "")
		encoder.Reset() // Simulate reuse
	}
}

// Test raw Xantos performance for comparison
func BenchmarkXantos_Raw(b *testing.B) {
	// Direct Xantos usage for comparison
	ring, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(io.Discard),
		BufferSize: 16384,
		BatchSize:  256,
	})
	defer ring.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ring.ring.Write(func(entry *LogEntry) {
			entry.Level = InfoLevel
			entry.Message = "Raw benchmark"
			entry.Fields = entry.fieldBuf[:0]
		})
	}
}
