// text_encoder_profile_test.go: Performance profiling for TextEncoder optimization
//
// Identifica i bottleneck delle performance del TextEncoder per ridurre
// l'instabilità e migliorare le performance consistenti.

package iris

import (
	"bytes"
	"testing"
	"time"
)

// BenchmarkTextEncoder_Profiling replica il caso del test che fallisce
func BenchmarkTextEncoder_Profiling(b *testing.B) {
	encoder := NewTextEncoder()

	record := &Record{
		Level:  Info,
		Msg:    "Performance test message",
		fields: [32]Field{},
		n:      0,
	}

	// Stessi field del test che fallisce
	record.fields[0] = Str("user", "john_doe")
	record.fields[1] = Secret("password", "secret123")
	record.fields[2] = Int64("count", 42)
	record.fields[3] = Float64("ratio", 3.14159)
	record.fields[4] = Bool("active", true)
	record.n = 5

	now := time.Now()
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		encoder.Encode(record, now, &buf)
	}
}

// BenchmarkTextEncoder_SimpleProfile profila caso semplice
func BenchmarkTextEncoder_SimpleProfile(b *testing.B) {
	encoder := NewTextEncoder()
	record := NewRecord(Info, "test message")
	now := time.Now()
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		encoder.Encode(record, now, &buf)
	}
}
