// full_comparison_test.go: Full end-to-end comparison with Zap
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"

	"go.uber.org/zap"
)

// BenchmarkIrisFullConsoleLogging tests full Iris console logging pipeline
func BenchmarkIrisFullConsoleLogging(b *testing.B) {
	// Create Iris logger with console output to discard
	logger, err := New(Config{
		Level:  DebugLevel,
		Writer: NewDiscardWriter(), // Discard output for pure performance test
		Format: JSONFormat,         // We don't have console format in main config yet
	})
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			String("component", "benchmark"),
			Int("iteration", 100),
			Float("duration", 1.234),
		)
	}
}

// BenchmarkZapDevelopmentLogger tests Zap's development logger (full pipeline)
func BenchmarkZapDevelopmentLogger(b *testing.B) {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"/dev/null"} // Discard output

	logger, err := config.Build()
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Sync()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			zap.String("component", "benchmark"),
			zap.Int("iteration", 100),
			zap.Float64("duration", 1.234),
		)
	}
}

// BenchmarkZapProductionLogger tests Zap's production logger
func BenchmarkZapProductionLogger(b *testing.B) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"} // Discard output

	logger, err := config.Build()
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Sync()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			zap.String("component", "benchmark"),
			zap.Int("iteration", 100),
			zap.Float64("duration", 1.234),
		)
	}
}

// DiscardWriter is a writer that discards all writes
type DiscardWriter struct{}

func NewDiscardWriter() *DiscardWriter {
	return &DiscardWriter{}
}

func (w *DiscardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w *DiscardWriter) Close() error {
	return nil
}

// BenchmarkIrisSimpleConsoleLogging tests simple Iris logging (like our baseline)
func BenchmarkIrisSimpleConsoleLogging(b *testing.B) {
	logger, err := New(Config{
		Level:  DebugLevel,
		Writer: NewDiscardWriter(),
		Format: JSONFormat,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}
