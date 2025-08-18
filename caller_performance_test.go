// caller_performance_test.go: Focused caller performance tests
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"runtime"
	"testing"

	"go.uber.org/zap"
)

// BenchmarkIrisCallerOnly tests just caller overhead (no fields)
func BenchmarkIrisCallerOnly(b *testing.B) {
	config := Config{
		Level:        InfoLevel,
		Writer:       NewDiscardWriter(),
		Format:       JSONFormat,
		EnableCaller: true,
		CallerSkip:   3,
		BufferSize:   1024,
		BatchSize:    8,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Simple message with caller") // No fields
	}
}

// BenchmarkZapCallerOnly tests just Zap caller overhead (no fields)
func BenchmarkZapCallerOnly(b *testing.B) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}
	config.DisableCaller = false
	config.DisableStacktrace = true

	logger, err := config.Build()
	if err != nil {
		b.Fatalf("Failed to create Zap logger: %v", err)
	}
	defer logger.Sync()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Simple message with caller") // No fields
	}
}

// BenchmarkCallerFunction tests function name extraction performance
func BenchmarkCallerFunction(b *testing.B) {
	config := Config{
		Level:        InfoLevel,
		Writer:       NewDiscardWriter(),
		Format:       JSONFormat,
		EnableCaller: true,
		CallerSkip:   3,
		BufferSize:   1024,
		BatchSize:    8,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// This will trigger function name extraction
		caller := logger.getCaller()
		_ = caller.Function // Force function name extraction
	}
}

// BenchmarkCallerFileOnly tests caller without function names
func BenchmarkCallerFileOnly(b *testing.B) {
	// Mock a version that doesn't extract function names
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simplified caller info (file + line only)
		pc, file, line, ok := runtime.Caller(3)
		if ok {
			_ = pc
			_ = file
			_ = line
		}
	}
}
