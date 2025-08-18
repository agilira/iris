// enhanced_fields_comparison_test.go: Compare Iris vs Zap on enhanced field types
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"

	"go.uber.org/zap"
)

// BenchmarkIrisEnhancedFields tests Iris with all enhanced field types
func BenchmarkIrisEnhancedFields(b *testing.B) {
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
		logger.Info("enhanced fields test",
			Int32("int32", 32),
			Uint64("uint64", 64),
			Float32("float32", 3.14),
			ByteString("bytes", []byte("hello world")),
			Binary("binary", []byte{1, 2, 3, 4, 5}),
		)
	}
}

// BenchmarkZapEnhancedFields tests Zap with equivalent enhanced field types
func BenchmarkZapEnhancedFields(b *testing.B) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}

	logger, err := config.Build()
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Sync()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("enhanced fields test",
			zap.Int32("int32", 32),
			zap.Uint64("uint64", 64),
			zap.Float32("float32", 3.14),
			zap.ByteString("bytes", []byte("hello world")),
			zap.Binary("binary", []byte{1, 2, 3, 4, 5}),
		)
	}
}

// BenchmarkIrisIntegerVariants tests Iris integer variants
func BenchmarkIrisIntegerVariants(b *testing.B) {
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
		logger.Info("integer variants test",
			Int8("int8", 8),
			Int16("int16", 16),
			Int32("int32", 32),
			Uint8("uint8", 8),
			Uint16("uint16", 16),
			Uint32("uint32", 32),
		)
	}
}

// BenchmarkZapIntegerVariants tests Zap integer variants
func BenchmarkZapIntegerVariants(b *testing.B) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}

	logger, err := config.Build()
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Sync()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("integer variants test",
			zap.Int8("int8", 8),
			zap.Int16("int16", 16),
			zap.Int32("int32", 32),
			zap.Uint8("uint8", 8),
			zap.Uint16("uint16", 16),
			zap.Uint32("uint32", 32),
		)
	}
}

// BenchmarkIrisBinaryFields tests Iris binary fields specifically
func BenchmarkIrisBinaryFields(b *testing.B) {
	logger, err := New(Config{
		Level:  DebugLevel,
		Writer: NewDiscardWriter(),
		Format: JSONFormat,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	data := []byte("this is some binary data that we want to encode")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("binary test",
			Binary("data", data),
			ByteString("text", data),
		)
	}
}

// BenchmarkZapBinaryFields tests Zap binary fields specifically
func BenchmarkZapBinaryFields(b *testing.B) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}

	logger, err := config.Build()
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Sync()

	data := []byte("this is some binary data that we want to encode")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("binary test",
			zap.Binary("data", data),
			zap.ByteString("text", data),
		)
	}
}

// BenchmarkIrisAnyFields tests Iris Any fields (most expensive)
func BenchmarkIrisAnyFields(b *testing.B) {
	logger, err := New(Config{
		Level:  DebugLevel,
		Writer: NewDiscardWriter(),
		Format: JSONFormat,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	complexData := map[string]interface{}{
		"name":  "test",
		"count": 42,
		"items": []string{"a", "b", "c"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("any fields test",
			Any("complex", complexData),
			Any("simple", "string"),
		)
	}
}

// BenchmarkZapAnyFields tests Zap Any fields equivalent
func BenchmarkZapAnyFields(b *testing.B) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}

	logger, err := config.Build()
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Sync()

	complexData := map[string]interface{}{
		"name":  "test",
		"count": 42,
		"items": []string{"a", "b", "c"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info("any fields test",
			zap.Any("complex", complexData),
			zap.Any("simple", "string"),
		)
	}
}
