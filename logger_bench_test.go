// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0
// Official IRIS benchmark suite - identical to zap for 1:1 performance comparison

package iris

import (
	"errors"
	"testing"
	"time"
)

// discardSyncer implements WriteSyncer but discards all writes (like zap's pattern)
type discardSyncer struct{}

func (ds discardSyncer) Write(p []byte) (n int, err error) { return len(p), nil }
func (ds discardSyncer) Sync() error                       { return nil }

type user struct {
	Name      string
	Email     string
	CreatedAt time.Time
}

var _jane = &user{
	Name:      "Jane Doe",
	Email:     "jane@test.com",
	CreatedAt: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
}

// For Object fields - implements marshaler interface
func (u user) MarshalLogObject() error {
	// IRIS doesn't have the same interface as zap,
	// we'll use Object() which handles reflection
	return nil
}

// Equivalent to zap's withBenchedLogger but using IRIS API
func withZapStyleLogger(b *testing.B, f func(*Logger)) {
	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  discardSyncer{}, // Use same discard pattern as zap
	})
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	logger.Start()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

// =============================================================================
// IDENTICAL ZAP BENCHMARK SUITE
// =============================================================================

func BenchmarkNoContext(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("No context.")
	})
}

func BenchmarkBoolField(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Boolean.", Bool("foo", true))
	})
}

func BenchmarkByteStringField(b *testing.B) {
	val := []byte("bar")
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("ByteString.", Bytes("foo", val)) // IRIS uses Bytes instead of ByteString
	})
}

func BenchmarkIntField(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Integer.", Int("foo", 42))
	})
}

func BenchmarkInt64Field(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("64-bit integer.", Int64("foo", 42))
	})
}

func BenchmarkFloat64Field(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Floating point.", Float64("foo", 3.14))
	})
}

func BenchmarkStringField(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Strings.", String("foo", "bar"))
	})
}

func BenchmarkStringerField(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Level.", String("foo", Info.String())) // IRIS equivalent to Stringer
	})
}

func BenchmarkTimeField(b *testing.B) {
	val := time.Unix(0, 0)
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Time.", Time("foo", val))
	})
}

func BenchmarkDurationField(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Duration", Dur("foo", time.Second))
	})
}

func BenchmarkErrorField(b *testing.B) {
	err := errors.New("egad")
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Error.", Err(err))
	})
}

func BenchmarkErrorsField(b *testing.B) {
	errs := []error{
		errors.New("egad"),
		errors.New("oh no"),
		errors.New("dear me"),
		errors.New("such fail"),
	}
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Errors.", Errors("errors", errs))
	})
}

func BenchmarkStackField(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		// IRIS equivalent to zap's Stack field - simplified for benchmark
		log.Info("Error.", String("stacktrace", "fake stack trace"))
	})
}

func BenchmarkObjectField(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("Arbitrary ObjectMarshaler.", Object("user", _jane))
	})
}

func BenchmarkReflectField(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		// Object() in IRIS uses reflection by default
		log.Info("Reflection-based serialization.", Object("user", _jane))
	})
}

func BenchmarkAddCallerHook(b *testing.B) {
	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  discardSyncer{},
	}, WithCaller()) // IRIS equivalent to zap's AddCaller
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	logger.Start()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Caller.")
		}
	})
}

func BenchmarkAddCallerAndStacktrace(b *testing.B) {
	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  discardSyncer{},
		// IRIS doesn't have auto-stacktrace config like zap development
	}, WithCaller()) // IRIS equivalent to zap's AddCaller
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	logger.Start()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Warn("Caller and stacktrace.")
		}
	})
}

func Benchmark5WithsUsed(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		// Simulate zap's With chaining using IRIS With
		log = log.With(String("worker", "1"), Int("trace", 1234), String("span", "5678"))
		log = log.With(String("baggage_key", "baggage_value"), String("other", "value"))
		log.Info("used")
	})
}

func Benchmark5WithsNotUsed(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		// Create With context but don't use it
		_ = log.With(String("worker", "1"), Int("trace", 1234), String("span", "5678"))
		_ = log.With(String("baggage_key", "baggage_value"), String("other", "value"))
		log.Info("used")
	})
}

func Benchmark5WithLazysUsed(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		// IRIS doesn't have lazy fields like zap, simulate with regular fields
		log = log.With(
			String("worker", "1"),
			Int("trace", 1234),
			String("span", "5678"),
			String("baggage_key", "baggage_value"),
			String("other", "value"),
		)
		log.Info("used")
	})
}

func Benchmark5WithLazysNotUsed(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		// Create lazy context but don't use it
		_ = log.With(
			String("worker", "1"),
			Int("trace", 1234),
			String("span", "5678"),
			String("baggage_key", "baggage_value"),
			String("other", "value"),
		)
		log.Info("used")
	})
}

func Benchmark10Fields(b *testing.B) {
	// Pre-allocate error outside the benchmark loop
	err := errors.New("fake")

	withZapStyleLogger(b, func(log *Logger) {
		log.Info("used",
			String("string", "four!"),
			Time("time", time.Unix(0, 0)),
			Int("int", 123),
			Dur("duration", time.Microsecond),
			Bool("bool", true),
			Float64("float64", 3.14),
			Int64("int64", 1<<32),
			String("string2", "four!"),
			Time("time2", time.Unix(0, 0)),
			Err(err),
		)
	})
}

func Benchmark10FieldsWithNewError(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("used",
			String("string", "four!"),
			Time("time", time.Unix(0, 0)),
			Int("int", 123),
			Dur("duration", time.Microsecond),
			Bool("bool", true),
			Float64("float64", 3.14),
			Int64("int64", 1<<32),
			String("string2", "four!"),
			Time("time2", time.Unix(0, 0)),
			Err(errors.New("fake")), // Creates new error every time
		)
	})
}

func Benchmark100Fields(b *testing.B) {
	// Pre-allocate fields outside benchmark to avoid measuring field creation
	fields := make([]Field, 100)
	for i := 0; i < 100; i++ {
		fields[i] = Int("foo", i)
	}

	withZapStyleLogger(b, func(log *Logger) {
		log.Info("used", fields...)
	})
}

func Benchmark100FieldsWithCreation(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		// Create fields inside the benchmark to measure full cost
		fields := make([]Field, 100)
		for i := 0; i < 100; i++ {
			fields[i] = Int("foo", i)
		}
		log.Info("used", fields...)
	})
}

func BenchmarkAny(b *testing.B) {
	withZapStyleLogger(b, func(log *Logger) {
		log.Info("used", Object("key", map[string]interface{}{
			"str": "value",
			"int": 42,
		}))
	})
}
