// text_encoder_performance_test.go: Performance tests for TextEncoder
//
// This file contains benchmarks to measure the impact of changes on TextEncoder
// performance, helping identify any regressions from field array optimizations.
//
// Copyright (c) 2025 AGILira
// Series: IRIS Logging Library - Performance Optimization
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"testing"
	"time"
)

// BenchmarkTextEncoder_Simple tests basic text encoder performance
func BenchmarkTextEncoder_Simple(b *testing.B) {
	encoder := NewTextEncoder()
	record := NewRecord(Info, "test message")
	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		encoder.Encode(record, now, &buf)
	}
}

// BenchmarkTextEncoder_WithFields tests text encoder with multiple fields
func BenchmarkTextEncoder_WithFields(b *testing.B) {
	encoder := NewTextEncoder()
	record := NewRecord(Info, "Performance test message")

	// Add multiple fields like in the failing test
	record.AddField(Str("user", "john_doe"))
	record.AddField(Secret("password", "secret123"))
	record.AddField(Int64("count", 42))
	record.AddField(Float64("ratio", 3.14159))
	record.AddField(Bool("active", true))

	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		encoder.Encode(record, now, &buf)
	}
}

// BenchmarkTextEncoder_MaxFields tests text encoder with maximum fields (32)
func BenchmarkTextEncoder_MaxFields(b *testing.B) {
	encoder := NewTextEncoder()
	record := NewRecord(Info, "Max fields test")

	// Fill with maximum fields
	for i := 0; i < 32; i++ {
		record.AddField(Str("field", "value"))
	}

	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		encoder.Encode(record, now, &buf)
	}
}
