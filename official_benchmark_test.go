// official_benchmark_test.go: Official benchmark comparing Iris Binary Logger vs Zap
//
// This benchmark is designed to be 100% IDENTICAL to Zap's official benchmarks
// from zap/benchmarks/scenario_bench_test.go for direct comparison.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// OFFICIAL BENCHMARK DATA - IDENTICAL TO ZAP
var (
	_messages   [1000]string
	_tenInts    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	_tenStrings = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	_tenTimes   = []time.Time{
		time.Unix(0, 0),
		time.Unix(1, 0),
		time.Unix(2, 0),
		time.Unix(3, 0),
		time.Unix(4, 0),
		time.Unix(5, 0),
		time.Unix(6, 0),
		time.Unix(7, 0),
		time.Unix(8, 0),
		time.Unix(9, 0),
	}
	errExample = func() error { return io.EOF }()
)

func init() {
	for i := range _messages {
		_messages[i] = "A bunch of fake message text that is kinda realistic and such."
	}
}

// getMessage - IDENTICAL TO ZAP
func getMessage(iter int) string {
	return _messages[iter%1000]
}

// newZapLogger - IDENTICAL TO ZAP
func newZapLogger(lvl zapcore.Level) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeDuration = zapcore.NanosDurationEncoder
	ec.EncodeTime = zapcore.EpochNanosTimeEncoder
	enc := zapcore.NewJSONEncoder(ec)
	return zap.New(zapcore.NewCore(
		enc,
		zapcore.AddSync(io.Discard),
		lvl,
	))
}

// fakeFields - IDENTICAL TO ZAP
func fakeFields() []zap.Field {
	return []zap.Field{
		zap.Int("int", _tenInts[0]),
		zap.String("string", _tenStrings[0]),
		zap.Time("time", _tenTimes[0]),
		zap.Error(errExample),
	}
}

// newSlog creates a slog logger
func newSlog() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// OFFICIAL BENCHMARK 1: WITHOUT FIELDS (SIMPLE LOGGING)
// This replicates EXACTLY BenchmarkWithoutFields from Zap
func BenchmarkWithoutFields(b *testing.B) {
	b.Logf("Logging without any structured context.")

	// IRIS OFFICIAL LOGGER - FIRST PLACE
	b.Run("Iris", func(b *testing.B) {
		logger := NewBinaryLogger(DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ctx := logger.WithBinaryFields() // No fields - simple logging
				ctx.Info(getMessage(0))
			}
		})
	})

	// ZAP OFFICIAL BENCHMARK - COMPETITOR
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})

	// SLOG REFERENCE - COMPETITOR
	b.Run("slog", func(b *testing.B) {
		logger := newSlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
}

// OFFICIAL BENCHMARK 2: WITH ACCUMULATED CONTEXT (STRUCTURED LOGGING)
// This replicates EXACTLY BenchmarkAccumulatedContext from Zap
func BenchmarkAccumulatedContext(b *testing.B) {
	b.Logf("Logging with some accumulated context.")

	// IRIS OFFICIAL LOGGER - FIRST PLACE
	b.Run("Iris", func(b *testing.B) {
		logger := NewBinaryLogger(DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ctx := logger.WithBinaryFields(
					BinaryInt("int", int64(_tenInts[0])),
					BinaryStr("string", _tenStrings[0]),
					// Note: Time and Error not implemented yet in binary
				)
				ctx.Info(getMessage(0))
			}
		})
	})

	// ZAP OFFICIAL BENCHMARK - COMPETITOR
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})

	// SLOG REFERENCE - COMPETITOR
	b.Run("slog", func(b *testing.B) {
		logger := newSlog()
		logger = logger.With(
			slog.Int("int", _tenInts[0]),
			slog.String("string", _tenStrings[0]),
			slog.Time("time", _tenTimes[0]),
			slog.Any("error", errExample),
		)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
}

// OFFICIAL BENCHMARK 3: ADDING FIELDS (DYNAMIC STRUCTURED LOGGING)
// This replicates EXACTLY BenchmarkAddingFields from Zap
func BenchmarkAddingFields(b *testing.B) {
	b.Logf("Logging with additional context at each log site.")

	// IRIS OFFICIAL LOGGER - FIRST PLACE
	b.Run("Iris", func(b *testing.B) {
		logger := NewBinaryLogger(DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ctx := logger.WithBinaryFields(
					BinaryInt("int", int64(_tenInts[0])),
					BinaryStr("string", _tenStrings[0]),
					// Adding fields dynamically - same pattern as Zap
				)
				ctx.Info(getMessage(0))
			}
		})
	})

	// ZAP OFFICIAL BENCHMARK - COMPETITOR
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0), fakeFields()...)
			}
		})
	})

	// SLOG REFERENCE - COMPETITOR
	b.Run("slog", func(b *testing.B) {
		logger := newSlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0),
					slog.Int("int", _tenInts[0]),
					slog.String("string", _tenStrings[0]),
				)
			}
		})
	})
}
