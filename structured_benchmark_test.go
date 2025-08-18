package iris

import (
	"testing"
	"time"
)

// BenchmarkStructured_WithFields - Pure structured operations (no JSON until output)
func BenchmarkStructured_WithFields(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewDiscardWriter(),
		BufferSize: 16384,
		BatchSize:  256,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create structured entry with fields (pure structured, no JSON yet)
		entry := logger.WithFieldsStructured(
			Str("service", "benchmark"),
			Int("iteration", i),
			Bool("success", true),
		)

		// Log message (uses existing infrastructure)
		entry.Info("Structured benchmark message")
	}
}

// BenchmarkStructured_ToJSON - Structured to JSON conversion when needed
func BenchmarkStructured_ToJSON(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		encoder := NewStructuredEncoder()

		// Add fields in structured format
		encoder.AddString("key", "value")
		encoder.AddInt("number", 42)
		encoder.AddBool("flag", true)

		// Convert to JSON only when needed (lazy)
		timestamp := time.Now()
		caller := Caller{File: "test.go", Line: 10, Function: "Test"}
		_ = encoder.ToJSON(timestamp, InfoLevel, "test message", caller)

		encoder.Reset()
	}
}

// BenchmarkStructured_MemoryProfile - Memory footprint analysis
func BenchmarkStructured_MemoryProfile(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		encoder := NewStructuredEncoder()
		encoder.AddString("key", "value")

		// Measure actual memory footprint
		footprint := encoder.MemoryFootprint()
		_ = footprint

		encoder.Reset()
	}
}
