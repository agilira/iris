package iris

import (
	"testing"
)

// BenchmarkIris_WithFields_Micro - Isolated micro-benchmark per scientific measurement
func BenchmarkIris_WithFields_Micro(b *testing.B) {
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
		logger.Info("Micro benchmark",
			Str("service", "test"),
			Int("iteration", i),
			Bool("success", true),
		)
	}
}

// BenchmarkBuffer_SizeImpact - Test different buffer approaches scientifically
func BenchmarkBuffer_SizeImpact(b *testing.B) {
	// Test current approach vs different strategies
	b.Run("Current_64B_Pool", func(b *testing.B) {
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
			logger.Info("Buffer test",
				Str("service", "benchmark"),
				Int("iteration", i),
				Bool("success", true),
			)
		}
	})

	// Future: We'll test different pool strategies here
}

// BenchmarkMemory_Profile - Detailed memory allocation analysis
func BenchmarkMemory_Profile(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewDiscardWriter(),
		BufferSize: 16384,
		BatchSize:  256,
	})
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	// Exactly same pattern as WithFields benchmark
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark with fields",
			Str("service", "benchmark"),
			Int("iteration", i),
			Bool("success", true),
		)
	}
}

// BenchmarkZero_Allocation_Attempt - Target zero allocations
func BenchmarkZero_Allocation_Attempt(b *testing.B) {
	logger, _ := New(Config{
		Level:      InfoLevel,
		Writer:     NewDiscardWriter(),
		BufferSize: 16384,
		BatchSize:  256,
	})
	defer logger.Close()

	// Pre-allocate fields to avoid allocation in loop
	fields := []Field{
		Str("service", "benchmark"),
		Int("iteration", 0),
		Bool("success", true),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Update iteration field without reallocating
		fields[1] = Int("iteration", i)
		logger.log(InfoLevel, "Zero allocation test", fields)
	}
}
