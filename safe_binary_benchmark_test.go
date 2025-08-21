// safe_binary_benchmark_test.go: Step 1.2 - Performance validation for Safe BinaryField API
//
// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0
//
// Questo file contiene benchmark per validare i performance gains
// delle nuove funzioni BinaryField vs legacy API in Step 1.2.

package iris

import (
	"testing"
)

// =============================================================================
// Step 1.2: Performance Benchmarks - BinaryField vs Legacy
// =============================================================================

// BenchmarkFieldCreation_Legacy benchmarks legacy field creation
func BenchmarkFieldCreation_Legacy(b *testing.B) {
	b.Run("Str", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Str("benchmark_key", "benchmark_value")
		}
	})

	b.Run("Int", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Int("benchmark_key", 12345)
		}
	})

	b.Run("Bool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Bool("benchmark_key", true)
		}
	})
}

// BenchmarkFieldCreation_NextGen benchmarks next-generation field creation
func BenchmarkFieldCreation_NextGen(b *testing.B) {
	b.Run("NextStr", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NextStr("benchmark_key", "benchmark_value")
		}
	})

	b.Run("NextInt", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NextInt("benchmark_key", 12345)
		}
	})

	b.Run("NextBool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NextBool("benchmark_key", true)
		}
	})
}

// BenchmarkFieldComposition_Legacy benchmarks complex field composition legacy
func BenchmarkFieldComposition_Legacy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fields := []Field{
			Str("user", "john_doe"),
			Int("age", 30),
			Bool("active", true),
			Str("email", "john@example.com"),
			Int("score", 95),
		}
		_ = fields
	}
}

// BenchmarkFieldComposition_NextGen benchmarks complex field composition next-gen
func BenchmarkFieldComposition_NextGen(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fields := []BinaryField{
			NextStr("user", "john_doe"),
			NextInt("age", 30),
			NextBool("active", true),
			NextStr("email", "john@example.com"),
			NextInt("score", 95),
		}
		_ = fields
	}
}

// BenchmarkConversion_ToLegacy benchmarks BinaryField to Legacy conversion
func BenchmarkConversion_ToLegacy(b *testing.B) {
	binaryFields := []BinaryField{
		NextStr("user", "john_doe"),
		NextInt("age", 30),
		NextBool("active", true),
		NextStr("email", "john@example.com"),
		NextInt("score", 95),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ToLegacyFields(binaryFields)
	}
}

// BenchmarkConversion_SingleField benchmarks single field conversion
func BenchmarkConversion_SingleField(b *testing.B) {
	bf := NextStr("test_key", "test_value")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = toLegacyField(bf)
	}
}

// BenchmarkMemoryFootprint compares memory footprint
func BenchmarkMemoryFootprint(b *testing.B) {
	b.Run("Legacy_1000_Fields", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fields := make([]Field, 1000)
			for j := 0; j < 1000; j++ {
				switch j % 3 {
				case 0:
					fields[j] = Str("key", "value")
				case 1:
					fields[j] = Int("key", j)
				case 2:
					fields[j] = Bool("key", j%2 == 0)
				}
			}
			_ = fields
		}
	})

	b.Run("NextGen_1000_Fields", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fields := make([]BinaryField, 1000)
			for j := 0; j < 1000; j++ {
				switch j % 3 {
				case 0:
					fields[j] = NextStr("key", "value")
				case 1:
					fields[j] = NextInt("key", j)
				case 2:
					fields[j] = NextBool("key", j%2 == 0)
				}
			}
			_ = fields
		}
	})
}

// BenchmarkConcurrentFieldCreation benchmarks concurrent field creation
func BenchmarkConcurrentFieldCreation(b *testing.B) {
	b.Run("Legacy_Concurrent", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = Str("concurrent", "test")
			}
		})
	})

	b.Run("NextGen_Concurrent", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = NextStr("concurrent", "test")
			}
		})
	})
}

// BenchmarkStringFieldVariations benchmarks different string field sizes
func BenchmarkStringFieldVariations(b *testing.B) {
	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{"small", "k", "v"},
		{"medium", "medium_key_name", "medium_length_value_content"},
		{"large", "very_long_key_name_for_testing_purposes", "very_long_value_content_that_represents_real_world_usage_scenarios"},
	}

	for _, tc := range testCases {
		b.Run("Legacy_"+tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = Str(tc.key, tc.value)
			}
		})

		b.Run("NextGen_"+tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = NextStr(tc.key, tc.value)
			}
		})
	}
}

// BenchmarkIntFieldVariations benchmarks different integer values
func BenchmarkIntFieldVariations(b *testing.B) {
	testCases := []struct {
		name  string
		value int
	}{
		{"zero", 0},
		{"small_positive", 42},
		{"small_negative", -42},
		{"large_positive", 1234567890},
		{"large_negative", -1234567890},
	}

	for _, tc := range testCases {
		b.Run("Legacy_"+tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = Int("test_key", tc.value)
			}
		})

		b.Run("NextGen_"+tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = NextInt("test_key", tc.value)
			}
		})
	}
}

// BenchmarkFullPipeline benchmarks the full logging pipeline
func BenchmarkFullPipeline(b *testing.B) {
	b.Run("Legacy_Pipeline", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fields := []Field{
				Str("operation", "benchmark"),
				Int("iteration", i),
				Bool("success", true),
			}
			_ = fields
		}
	})

	b.Run("NextGen_Pipeline_With_Conversion", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			binaryFields := []BinaryField{
				NextStr("operation", "benchmark"),
				NextInt("iteration", i),
				NextBool("success", true),
			}
			fields := ToLegacyFields(binaryFields)
			_ = fields
		}
	})

	b.Run("NextGen_Pipeline_Binary_Only", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fields := []BinaryField{
				NextStr("operation", "benchmark"),
				NextInt("iteration", i),
				NextBool("success", true),
			}
			_ = fields
		}
	})
}
