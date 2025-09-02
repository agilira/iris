// json_encoder_optimization_test.go: Performance benchmarks for JSONEncoder optimizations
//
// Confronta le performance prima e dopo le ottimizzazioni del hot path
// per verificare i miglioramenti di performance.

package iris

import (
	"bytes"
	"testing"
	"time"
)

// BenchmarkJSONEncoder_OptimizedHotPath testa le performance con l'ottimizzazione
func BenchmarkJSONEncoder_OptimizedHotPath(b *testing.B) {
	encoder := NewJSONEncoder() // Usa costruttore ottimizzato
	record := NewRecord(Info, "Performance test")
	record.AddField(Str("user", "john_doe"))
	record.AddField(Int64("request_id", 12345))
	record.AddField(Float64("duration", 1.23))

	now := time.Now()
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		encoder.Encode(record, now, &buf)
	}
}

// BenchmarkJSONEncoder_ZeroValueEncoder testa performance con encoder zero-value
func BenchmarkJSONEncoder_ZeroValueEncoder(b *testing.B) {
	encoder := &JSONEncoder{RFC3339: true} // Zero-value encoder (dovrebbe usare fallback)
	record := NewRecord(Info, "Performance test")
	record.AddField(Str("user", "john_doe"))
	record.AddField(Int64("request_id", 12345))
	record.AddField(Float64("duration", 1.23))

	now := time.Now()
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		encoder.Encode(record, now, &buf)
	}
}

// BenchmarkJSONEncoder_TimeCache testa impatto del time caching
func BenchmarkJSONEncoder_TimeCache(b *testing.B) {
	encoder := NewJSONEncoder()
	record := NewRecord(Info, "Time cache test")
	record.AddField(Str("operation", "benchmark"))

	// Usa tempo corrente per sfruttare il time cache
	now := time.Now()
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		encoder.Encode(record, now, &buf) // Stesso timestamp = cache hit
	}
}

// BenchmarkJSONEncoder_NoTimeCache testa performance senza time cache
func BenchmarkJSONEncoder_NoTimeCache(b *testing.B) {
	encoder := NewJSONEncoder()
	record := NewRecord(Info, "No time cache test")
	record.AddField(Str("operation", "benchmark"))

	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		// Usa timestamp diverso ogni volta = no cache
		encoder.Encode(record, time.Unix(1000000+int64(i), 0), &buf)
	}
}
