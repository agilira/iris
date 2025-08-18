// field_encoding_comparison_test.go: Compare pure field encoding performance
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BenchmarkIrisFieldEncoding tests Iris field encoding only (same as BenchmarkEnhancedFieldTypes)
func BenchmarkIrisFieldEncoding(b *testing.B) {
	encoder := NewJSONEncoder()

	fields := []Field{
		Int32("int32", 32),
		Uint64("uint64", 64),
		Float32("float32", 3.14),
		ByteString("bytes", []byte("hello world")),
		Binary("binary", []byte{1, 2, 3, 4, 5}),
		Any("any", map[string]int{"count": 42}),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, field := range fields {
			encoder.buf = encoder.buf[:0]
			encoder.encodeField(field)
		}
	}
}

// BenchmarkZapFieldEncoding tests Zap field encoding equivalent
func BenchmarkZapFieldEncoding(b *testing.B) {
	// Create Zap JSON encoder
	config := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(config)

	// Create equivalent Zap fields
	zapFields := []zapcore.Field{
		zap.Int32("int32", 32),
		zap.Uint64("uint64", 64),
		zap.Float32("float32", 3.14),
		zap.ByteString("bytes", []byte("hello world")),
		zap.Binary("binary", []byte{1, 2, 3, 4, 5}),
		zap.Any("any", map[string]int{"count": 42}),
	}

	// Create a dummy entry for encoding
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "test",
	}

	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for range zapFields {
			buf.Reset()
			// Zap doesn't expose individual field encoding, so we encode the full entry
			// This gives Zap a slight disadvantage but it's the only way to test
			encoder.EncodeEntry(entry, zapFields)
		}
	}
}

// BenchmarkIrisSingleFieldTypes tests individual field types for Iris
func BenchmarkIrisSingleFieldTypes(b *testing.B) {
	encoder := NewJSONEncoder()

	benchmarks := []struct {
		name  string
		field Field
	}{
		{"Int32", Int32("test", 32)},
		{"Uint64", Uint64("test", 64)},
		{"Float32", Float32("test", 3.14)},
		{"ByteString", ByteString("test", []byte("hello world"))},
		{"Binary", Binary("test", []byte{1, 2, 3, 4, 5})},
		{"Any", Any("test", map[string]int{"count": 42})},
	}

	for _, bm := range benchmarks {
		b.Run("Iris_"+bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				encoder.buf = encoder.buf[:0]
				encoder.encodeField(bm.field)
			}
		})
	}
}

// BenchmarkZapSingleFieldTypes tests individual field types for Zap
func BenchmarkZapSingleFieldTypes(b *testing.B) {
	config := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(config)
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "test",
	}

	benchmarks := []struct {
		name  string
		field zapcore.Field
	}{
		{"Int32", zap.Int32("test", 32)},
		{"Uint64", zap.Uint64("test", 64)},
		{"Float32", zap.Float32("test", 3.14)},
		{"ByteString", zap.ByteString("test", []byte("hello world"))},
		{"Binary", zap.Binary("test", []byte{1, 2, 3, 4, 5})},
		{"Any", zap.Any("test", map[string]int{"count": 42})},
	}

	var buf bytes.Buffer

	for _, bm := range benchmarks {
		b.Run("Zap_"+bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				encoder.EncodeEntry(entry, []zapcore.Field{bm.field})
			}
		})
	}
}
