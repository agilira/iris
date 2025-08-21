// conversion_layer_benchmark_test.go: Step 1.3 - Conversion Layer Performance
//
// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0
//
// Benchmark per validare le performance del conversion layer
// ottimizzato in Step 1.3, incluso reverse conversion.

package iris

import (
	"fmt"
	"testing"
)

// =============================================================================
// Step 1.3: Conversion Layer Benchmarks
// =============================================================================

// BenchmarkConversionLayer_Forward benchmarks forward conversion performance
func BenchmarkConversionLayer_Forward(b *testing.B) {
	binaryFields := []BinaryField{
		NextStr("user", "john_doe"),
		NextInt("age", 30),
		NextBool("active", true),
		NextStr("email", "john@example.com"),
		NextInt("score", 95),
	}

	b.Run("ToLegacyFields", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ToLegacyFields(binaryFields)
		}
	})

	b.Run("ToLegacyFields_WithCapacity", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ToLegacyFieldsWithCapacity(binaryFields, 10)
		}
	})
}

// BenchmarkConversionLayer_Reverse benchmarks reverse conversion performance
func BenchmarkConversionLayer_Reverse(b *testing.B) {
	legacyFields := []Field{
		Str("user", "john_doe"),
		Int("age", 30),
		Bool("active", true),
		Str("email", "john@example.com"),
		Int("score", 95),
	}

	b.Run("ToBinaryFields", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ToBinaryFields(legacyFields)
		}
	})

	b.Run("ToBinaryFields_WithCapacity", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ToBinaryFieldsWithCapacity(legacyFields, 10)
		}
	})
}

// BenchmarkConversionLayer_SingleField benchmarks single field conversions
func BenchmarkConversionLayer_SingleField(b *testing.B) {
	bf := NextStr("test", "value")
	field := Str("test", "value")

	b.Run("ToLegacy_Single", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = toLegacyField(bf)
		}
	})

	b.Run("ToBinary_Single", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ToBinaryField(field)
		}
	})
}

// BenchmarkConversionLayer_Scaling benchmarks conversion at different scales
func BenchmarkConversionLayer_Scaling(b *testing.B) {
	scales := []int{10, 100, 1000, 10000}

	for _, scale := range scales {
		b.Run(fmt.Sprintf("Forward_%d_fields", scale), func(b *testing.B) {
			binaryFields := make([]BinaryField, scale)
			for i := 0; i < scale; i++ {
				switch i % 3 {
				case 0:
					binaryFields[i] = NextStr("key", "value")
				case 1:
					binaryFields[i] = NextInt("key", i)
				case 2:
					binaryFields[i] = NextBool("key", i%2 == 0)
				}
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = ToLegacyFields(binaryFields)
			}
		})

		b.Run(fmt.Sprintf("Reverse_%d_fields", scale), func(b *testing.B) {
			legacyFields := make([]Field, scale)
			for i := 0; i < scale; i++ {
				switch i % 3 {
				case 0:
					legacyFields[i] = Str("key", "value")
				case 1:
					legacyFields[i] = Int("key", i)
				case 2:
					legacyFields[i] = Bool("key", i%2 == 0)
				}
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = ToBinaryFields(legacyFields)
			}
		})
	}
}

// BenchmarkConversionLayer_RoundTrip benchmarks full round-trip conversion
func BenchmarkConversionLayer_RoundTrip(b *testing.B) {
	originalBinary := []BinaryField{
		NextStr("user", "john"),
		NextInt("age", 30),
		NextBool("active", true),
	}

	b.Run("Binary_To_Legacy_To_Binary", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			legacy := ToLegacyFields(originalBinary)
			_ = ToBinaryFields(legacy)
		}
	})

	originalLegacy := []Field{
		Str("user", "john"),
		Int("age", 30),
		Bool("active", true),
	}

	b.Run("Legacy_To_Binary_To_Legacy", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			binary := ToBinaryFields(originalLegacy)
			_ = ToLegacyFields(binary)
		}
	})
}

// BenchmarkConversionLayer_MemoryEfficiency benchmarks memory efficiency
func BenchmarkConversionLayer_MemoryEfficiency(b *testing.B) {
	b.Run("WithCapacity_vs_Regular", func(b *testing.B) {
		binaryFields := []BinaryField{
			NextStr("test", "value"),
			NextInt("num", 42),
		}

		b.Run("Regular", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = ToLegacyFields(binaryFields)
			}
		})

		b.Run("WithCapacity", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = ToLegacyFieldsWithCapacity(binaryFields, 10)
			}
		})
	})
}

// BenchmarkConversionLayer_Concurrent benchmarks concurrent conversion
func BenchmarkConversionLayer_Concurrent(b *testing.B) {
	binaryFields := []BinaryField{
		NextStr("user", "test"),
		NextInt("count", 42),
		NextBool("flag", true),
	}

	legacyFields := []Field{
		Str("user", "test"),
		Int("count", 42),
		Bool("flag", true),
	}

	b.Run("Forward_Concurrent", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = ToLegacyFields(binaryFields)
			}
		})
	})

	b.Run("Reverse_Concurrent", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = ToBinaryFields(legacyFields)
			}
		})
	})
}
