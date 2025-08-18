package iris

import (
	"testing"
)

// TASK S1: Direct binary fields benchmark (TESTA)
func BenchmarkBinary_WithBinaryFields_S1(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// S1: Direct binary field creation (no Field conversion)
		context := logger.WithBinaryFields(
			BinaryStr("service", "benchmark"),
			BinaryInt("iteration", int64(i)),
			BinaryBool("success", true),
		)

		// Log with pure binary format
		context.Info("S1 optimization test")
	}
}

// BenchmarkBinary_WithFields - Pure binary logging (NO JSON)
func BenchmarkBinary_WithFields(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Original: Field conversion approach
		context := logger.WithFields(
			Str("service", "benchmark"),
			Int("iteration", i),
			Bool("success", true),
		)

		// Log with pure binary format
		context.Info("Binary benchmark message")
	}
}

// BenchmarkBinary_MemoryProfile - Binary memory footprint analysis
func BenchmarkBinary_MemoryProfile(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		context := logger.WithFields(
			Str("key", "value"),
		)

		// Measure binary memory footprint
		footprint := context.MemoryFootprint()
		_ = footprint

		// Note: context cleanup happens in Info(), but we measure just structure
	}
}

// BenchmarkBinary_ZeroAllocation - Attempt zero allocation binary logging
func BenchmarkBinary_ZeroAllocation(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		context := logger.WithFields(
			Str("static", "value"),
		)

		// Minimal binary operation
		size := context.GetBinarySize()
		_ = size
	}
}

// BenchmarkBinary_WithBinaryFields_S2 tests zero-copy string optimization
func BenchmarkBinary_WithBinaryFields_S2(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// S2: Zero-copy string operations with StringHeader optimization
		ctx := logger.WithBinaryFields(
			BinaryStr("key", "value"),
			BinaryInt("id", 12345),
			BinaryBool("enabled", true),
		)
		ctx.Info("test message")
	}
}

// BenchmarkBinary_WithBinaryFields_S3 tests stack allocation optimization
func BenchmarkBinary_WithBinaryFields_S3(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// S3: Stack allocation for ≤4 fields (avoid pool overhead)
		ctx := logger.WithBinaryFields(
			BinaryStr("key", "value"),
			BinaryInt("id", 12345),
			BinaryBool("enabled", true),
		)
		ctx.Info("test message")
	}
}

// BenchmarkBinary_WithBinaryFields_S4 tests cached timestamp optimization
func BenchmarkBinary_WithBinaryFields_S4(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// S4: Cached timestamp to reduce time.Now() overhead
		ctx := logger.WithBinaryFields(
			BinaryStr("key", "value"),
			BinaryInt("id", 12345),
			BinaryBool("enabled", true),
		)
		ctx.Info("test message")
	}
}

// BenchmarkBinary_WithBinaryFields_S5 tests branchless bool optimization
func BenchmarkBinary_WithBinaryFields_S5(b *testing.B) {
	logger := NewBinaryLogger(InfoLevel)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// S5: Branchless bool operations to improve branch prediction
		ctx := logger.WithBinaryFields(
			BinaryStr("key", "value"),
			BinaryInt("id", 12345),
			BinaryBool("enabled", true),
		)
		ctx.Info("test message")
	}
}
