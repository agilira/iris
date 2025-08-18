package iris

import (
	"sync"
	"testing"
)

// BenchmarkCallerCache_BeforeAfter compara performance pre/post cache
func BenchmarkCallerCache_BeforeAfter(b *testing.B) {
	truePtr := true
	config := Config{
		Level:                InfoLevel,
		Writer:               NewDiscardWriter(),
		Format:               JSONFormat,
		EnableCaller:         true,
		EnableCallerFunction: &truePtr, // Pointer to true
		CallerSkip:           3,
		BufferSize:           1024,
		BatchSize:            8,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	// Pre-populate cache con alcune entries per test più realistico
	for i := 0; i < 10; i++ {
		logger.Info("warmup cache")
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("benchmark test message")
	}
}

// BenchmarkCallerCacheVsZap - Confronto diretto con Zap benchmark equivalent
func BenchmarkCallerCacheVsZap(b *testing.B) {
	b.Run("Iris_CallerCached", func(b *testing.B) {
		truePtr := true
		config := Config{
			Level:                InfoLevel,
			Writer:               NewDiscardWriter(),
			Format:               JSONFormat,
			EnableCaller:         true,
			EnableCallerFunction: &truePtr,
			CallerSkip:           3,
			BufferSize:           1024,
			BatchSize:            8,
		}

		logger, err := New(config)
		if err != nil {
			b.Fatal(err)
		}
		defer logger.Close()

		// Warmup cache
		logger.Info("warmup")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message")
		}
	})

	b.Run("Iris_CallerNoCaching", func(b *testing.B) {
		truePtr := true
		config := Config{
			Level:                InfoLevel,
			Writer:               NewDiscardWriter(),
			Format:               JSONFormat,
			EnableCaller:         true,
			EnableCallerFunction: &truePtr,
			CallerSkip:           3,
			BufferSize:           1024,
			BatchSize:            8,
		}

		logger, err := New(config)
		if err != nil {
			b.Fatal(err)
		}
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Clear cache ogni volta per forzare runtime.FuncForPC
			funcNameCache = sync.Map{}
			logger.Info("benchmark message")
		}
	})
}
