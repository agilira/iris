package iris

import (
	"io"
	"testing"
)

// BenchmarkMultipleOutputs benchmarks logger performance with multiple outputs
func BenchmarkMultipleOutputs(b *testing.B) {
	b.Run("SingleOutput", func(b *testing.B) {
		config := Config{
			BufferSize: 1024,
			BatchSize:  32,
			Writer:     io.Discard,
			Format:     JSONFormat,
			Level:      InfoLevel,
		}

		logger, err := New(config)
		if err != nil {
			b.Fatal(err)
		}
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", String("iteration", "test"))
		}
	})

	b.Run("DualOutput", func(b *testing.B) {
		config := Config{
			BufferSize: 1024,
			BatchSize:  32,
			Writers:    []io.Writer{io.Discard, io.Discard},
			Format:     JSONFormat,
			Level:      InfoLevel,
		}

		logger, err := New(config)
		if err != nil {
			b.Fatal(err)
		}
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", String("iteration", "test"))
		}
	})

	b.Run("TripleOutput", func(b *testing.B) {
		config := Config{
			BufferSize: 1024,
			BatchSize:  32,
			Writers:    []io.Writer{io.Discard, io.Discard, io.Discard},
			Format:     JSONFormat,
			Level:      InfoLevel,
		}

		logger, err := New(config)
		if err != nil {
			b.Fatal(err)
		}
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", String("iteration", "test"))
		}
	})

	b.Run("QuintOutput", func(b *testing.B) {
		config := Config{
			BufferSize: 1024,
			BatchSize:  32,
			Writers:    []io.Writer{io.Discard, io.Discard, io.Discard, io.Discard, io.Discard},
			Format:     JSONFormat,
			Level:      InfoLevel,
		}

		logger, err := New(config)
		if err != nil {
			b.Fatal(err)
		}
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", String("iteration", "test"))
		}
	})
}

// BenchmarkMultiWriterPerformance tests MultiWriter performance directly
func BenchmarkMultiWriterPerformance(b *testing.B) {
	data := []byte("test log message\n")

	b.Run("SingleWriter", func(b *testing.B) {
		writer := WrapWriter(io.Discard)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			writer.Write(data)
		}
	})

	b.Run("MultiWriter_2", func(b *testing.B) {
		mw := NewMultiWriter(
			WrapWriter(io.Discard),
			WrapWriter(io.Discard),
		)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			mw.Write(data)
		}
	})

	b.Run("MultiWriter_5", func(b *testing.B) {
		mw := NewMultiWriter(
			WrapWriter(io.Discard),
			WrapWriter(io.Discard),
			WrapWriter(io.Discard),
			WrapWriter(io.Discard),
			WrapWriter(io.Discard),
		)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			mw.Write(data)
		}
	})

	b.Run("MultiWriter_10", func(b *testing.B) {
		writers := make([]WriteSyncer, 10)
		for i := range writers {
			writers[i] = WrapWriter(io.Discard)
		}
		mw := NewMultiWriter(writers...)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			mw.Write(data)
		}
	})
}
