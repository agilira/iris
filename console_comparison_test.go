// console_comparison_test.go: Compare Iris console encoding vs Zap console encoding
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BenchmarkIrisConsoleEncoding tests Iris console encoding performance
func BenchmarkIrisConsoleEncoding(b *testing.B) {
	encoder := NewConsoleEncoder(false) // No colors for fair comparison
	entry := &LogEntry{
		Level:     InfoLevel,
		Message:   "benchmark message",
		Timestamp: time.Now(),
		Fields: []Field{
			String("component", "benchmark"),
			Int("iteration", 100),
			Float("duration", 1.234),
		},
	}

	buf := make([]byte, 0, 256)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		encoder.EncodeLogEntry(entry, buf)
	}
}

// BenchmarkZapConsoleEncoding tests Zap console encoding performance
func BenchmarkZapConsoleEncoding(b *testing.B) {
	// Create Zap console encoder
	config := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(config)

	// Create Zap entry equivalent to our test
	zapEntry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "benchmark message",
	}

	fields := []zapcore.Field{
		zap.String("component", "benchmark"),
		zap.Int("iteration", 100),
		zap.Float64("duration", 1.234),
	}

	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		encoder.EncodeEntry(zapEntry, fields)
	}
}

// BenchmarkZapDevelopmentConsole tests Zap's development console setup
func BenchmarkZapDevelopmentConsole(b *testing.B) {
	// This is Zap's typical development console setup
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"/dev/null"} // Discard output for pure encoding test

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

// BenchmarkIrisConsoleEncodingColorized tests Iris with colors
func BenchmarkIrisConsoleEncodingColorized(b *testing.B) {
	encoder := NewConsoleEncoder(true) // With colors
	entry := &LogEntry{
		Level:     InfoLevel,
		Message:   "benchmark message",
		Timestamp: time.Now(),
		Fields: []Field{
			String("component", "benchmark"),
			Int("iteration", 100),
			Float("duration", 1.234),
		},
	}

	buf := make([]byte, 0, 256)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		encoder.EncodeLogEntry(entry, buf)
	}
}

// BenchmarkZapConsoleEncodingColorized tests Zap with colors
func BenchmarkZapConsoleEncodingColorized(b *testing.B) {
	// Create Zap console encoder with colors
	config := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // Colors!
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(config)

	zapEntry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "benchmark message",
	}

	fields := []zapcore.Field{
		zap.String("component", "benchmark"),
		zap.Int("iteration", 100),
		zap.Float64("duration", 1.234),
	}

	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		encoder.EncodeEntry(zapEntry, fields)
	}
}
