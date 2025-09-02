// record_performance_test.go: Performance tests for Record optimizations
//
// This file contains benchmarks to measure the impact of Record field storage
// optimizations, comparing different field array sizes and access patterns.
//
// Copyright (c) 2025 AGILira
// Series: IRIS Logging Library - Performance Optimization
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
)

// BenchmarkRecord_AddField_SingleField tests single field addition performance
func BenchmarkRecord_AddField_SingleField(b *testing.B) {
	record := NewRecord(Info, "test message")
	field := Str("key", "value")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		record.n = 0 // Reset
		record.AddField(field)
	}
}

// BenchmarkRecord_AddField_10Fields tests realistic field count
func BenchmarkRecord_AddField_10Fields(b *testing.B) {
	record := NewRecord(Info, "test message")
	field := Str("key", "value")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		record.n = 0 // Reset

		// Add 10 fields (realistic scenario)
		for j := 0; j < 10; j++ {
			record.AddField(field)
		}
	}
}

// BenchmarkRecord_AddField_32Fields tests maximum field count
func BenchmarkRecord_AddField_32Fields(b *testing.B) {
	record := NewRecord(Info, "test message")
	field := Str("key", "value")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		record.n = 0 // Reset

		// Add 32 fields (maximum)
		for j := 0; j < 32; j++ {
			record.AddField(field)
		}
	}
}

// BenchmarkRecord_FieldAccess tests field access performance
func BenchmarkRecord_FieldAccess(b *testing.B) {
	record := NewRecord(Info, "test message")

	// Pre-populate with 10 fields
	for i := 0; i < 10; i++ {
		record.AddField(Str("key", "value"))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Access all fields (simulating encoder iteration)
		for j := int32(0); j < record.n; j++ {
			_ = record.GetField(int(j))
		}
	}
}

// BenchmarkRecord_Creation tests record creation overhead
func BenchmarkRecord_Creation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		record := NewRecord(Info, "test message")
		_ = record
	}
}

// BenchmarkRecord_Reset tests record reset performance (important for pooling)
func BenchmarkRecord_Reset(b *testing.B) {
	record := NewRecord(Info, "test message")

	// Pre-populate with fields
	for i := 0; i < 10; i++ {
		record.AddField(Str("key", "value"))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		record.resetForWrite()
	}
}
