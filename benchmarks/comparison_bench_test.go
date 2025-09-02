// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package benchmarks

import (
	"io"
	"log/slog"
	"testing"
	"time"

	// Iris
	"github.com/agilira/iris"

	// Zap
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	// Zerolog
	"github.com/rs/zerolog"

	// Logrus
	"github.com/sirupsen/logrus"

	// Apex
	"github.com/apex/log"
	apexjson "github.com/apex/log/handlers/json"

	// Go-kit
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"

	// Log15
	log15 "github.com/inconshreveable/log15/v3"
)

// =============================================================================
// IRIS BENCHMARKS
// =============================================================================

// newIrisLogger creates a logger with optimal benchmark configuration - copied from iris_benchmarks_test.go
func newIrisLogger() *iris.Logger {
	logger, err := iris.New(iris.Config{
		Level:              iris.Debug,
		Output:             iris.WrapWriter(io.Discard),
		Encoder:            iris.NewJSONEncoder(),
		Capacity:           64, // Minimum power of 2 (2^6 = 64)
		BatchSize:          1,  // No batching overhead
		Architecture:       iris.SingleRing,
		NumRings:           1,
		BackpressurePolicy: 0, // 0 = DropOnFull by design
	})
	if err != nil {
		panic(err)
	}
	// Don't call Start() - keep it synchronous for benchmarking
	return logger
}

// withBenchedIrisLogger replicates the pattern from iris_benchmarks_test.go
func withBenchedIrisLogger(b *testing.B, f func(*iris.Logger)) {
	logger := newIrisLogger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func BenchmarkIris_NoContext(b *testing.B) {
	withBenchedIrisLogger(b, func(log *iris.Logger) {
		log.Info("No context.")
	})
}

func BenchmarkIris_10Fields(b *testing.B) {
	withBenchedIrisLogger(b, func(log *iris.Logger) {
		log.Info("Ten fields, passed at the log site.",
			iris.Int("one", 1),
			iris.Int("two", 2),
			iris.Int("three", 3),
			iris.Int("four", 4),
			iris.Int("five", 5),
			iris.Int("six", 6),
			iris.Int("seven", 7),
			iris.Int("eight", 8),
			iris.Int("nine", 9),
			iris.Int("ten", 10),
		)
	})
}

func BenchmarkIris_DisabledWithoutFields(b *testing.B) {
	logger, err := iris.New(iris.Config{
		Level:              iris.Error, // Disabled level
		Output:             iris.WrapWriter(io.Discard),
		Encoder:            iris.NewJSONEncoder(),
		Capacity:           64, // Minimum power of 2 (2^6 = 64)
		BatchSize:          1,
		Architecture:       iris.SingleRing,
		NumRings:           1,
		BackpressurePolicy: 0, // 0 = DropOnFull by design
	})
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Logging at a disabled level without any structured context.")
		}
	})
}

func BenchmarkIris_WithoutFields(b *testing.B) {
	withBenchedIrisLogger(b, func(log *iris.Logger) {
		log.Info("Logging without any structured context.")
	})
}

func BenchmarkIris_AddingFields(b *testing.B) {
	withBenchedIrisLogger(b, func(log *iris.Logger) {
		log.Info("Logging with additional context at each log site.",
			iris.Int("int", 1),
			iris.String("string", "value"),
			iris.Time("time", time.Unix(0, 0)),
			iris.String("user1_name", "Jane Doe"),
			iris.String("user2_name", "Jane Doe"),
			iris.String("error", "fail"),
		)
	})
}

func BenchmarkIris_AccumulatedContext(b *testing.B) {
	logger := newIrisLogger().With(
		iris.Int("int", 1),
		iris.String("string", "value"),
		iris.Time("time", time.Unix(0, 0)),
		iris.String("user1_name", "Jane Doe"),
		iris.String("user2_name", "Jane Doe"),
		iris.String("error", "fail"),
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Logging with some accumulated context.")
		}
	})
}

// =============================================================================
// ZAP BENCHMARKS
// =============================================================================

// newZapLogger creates a logger with optimal benchmark configuration
func newZapLogger() *zap.Logger {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zapcore.InfoLevel)
	logger := zap.New(core)
	return logger
}

// withBenchedZapLogger replicates the pattern for Zap
func withBenchedZapLogger(b *testing.B, f func(*zap.Logger)) {
	logger := newZapLogger()
	defer logger.Sync()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func BenchmarkZap_NoContext(b *testing.B) {
	withBenchedZapLogger(b, func(log *zap.Logger) {
		log.Info("No context.")
	})
}

func BenchmarkZap_10Fields(b *testing.B) {
	withBenchedZapLogger(b, func(log *zap.Logger) {
		log.Info("Ten fields, passed at the log site.",
			zap.Int("one", 1),
			zap.Int("two", 2),
			zap.Int("three", 3),
			zap.Int("four", 4),
			zap.Int("five", 5),
			zap.Int("six", 6),
			zap.Int("seven", 7),
			zap.Int("eight", 8),
			zap.Int("nine", 9),
			zap.Int("ten", 10),
		)
	})
}

func BenchmarkZap_DisabledWithoutFields(b *testing.B) {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zapcore.ErrorLevel)
	logger := zap.New(core)
	defer logger.Sync()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Logging at a disabled level without any structured context.")
		}
	})
}

func BenchmarkZap_WithoutFields(b *testing.B) {
	withBenchedZapLogger(b, func(log *zap.Logger) {
		log.Info("Logging without any structured context.")
	})
}

func BenchmarkZap_AddingFields(b *testing.B) {
	withBenchedZapLogger(b, func(log *zap.Logger) {
		log.Info("Logging with additional context at each log site.",
			zap.Int("int", 1),
			zap.String("string", "value"),
			zap.Time("time", time.Unix(0, 0)),
			zap.String("user1_name", "Jane Doe"),
			zap.String("user2_name", "Jane Doe"),
			zap.String("error", "fail"),
		)
	})
}

func BenchmarkZap_AccumulatedContext(b *testing.B) {
	logger := newZapLogger().With(
		zap.Int("int", 1),
		zap.String("string", "value"),
		zap.Time("time", time.Unix(0, 0)),
		zap.String("user1_name", "Jane Doe"),
		zap.String("user2_name", "Jane Doe"),
		zap.String("error", "fail"),
	)
	defer logger.Sync()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Logging with some accumulated context.")
		}
	})
}

// =============================================================================
// ZEROLOG BENCHMARKS
// =============================================================================

// newZerologLogger creates a logger with optimal benchmark configuration
func newZerologLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.InfoLevel)
}

// withBenchedZerologLogger replicates the pattern for Zerolog
func withBenchedZerologLogger(b *testing.B, f func(zerolog.Logger)) {
	logger := newZerologLogger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func BenchmarkZerolog_NoContext(b *testing.B) {
	withBenchedZerologLogger(b, func(log zerolog.Logger) {
		log.Info().Msg("No context.")
	})
}

func BenchmarkZerolog_10Fields(b *testing.B) {
	withBenchedZerologLogger(b, func(log zerolog.Logger) {
		log.Info().
			Int("one", 1).
			Int("two", 2).
			Int("three", 3).
			Int("four", 4).
			Int("five", 5).
			Int("six", 6).
			Int("seven", 7).
			Int("eight", 8).
			Int("nine", 9).
			Int("ten", 10).
			Msg("Ten fields, passed at the log site.")
	})
}

func BenchmarkZerolog_DisabledWithoutFields(b *testing.B) {
	logger := zerolog.New(io.Discard).Level(zerolog.ErrorLevel)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Msg("Logging at a disabled level without any structured context.")
		}
	})
}

func BenchmarkZerolog_WithoutFields(b *testing.B) {
	withBenchedZerologLogger(b, func(log zerolog.Logger) {
		log.Info().Msg("Logging without any structured context.")
	})
}

func BenchmarkZerolog_AddingFields(b *testing.B) {
	withBenchedZerologLogger(b, func(log zerolog.Logger) {
		log.Info().
			Int("int", 1).
			Str("string", "value").
			Time("time", time.Unix(0, 0)).
			Str("user1_name", "Jane Doe").
			Str("user2_name", "Jane Doe").
			Str("error", "fail").
			Msg("Logging with additional context at each log site.")
	})
}

func BenchmarkZerolog_AccumulatedContext(b *testing.B) {
	logger := newZerologLogger().With().
		Int("int", 1).
		Str("string", "value").
		Time("time", time.Unix(0, 0)).
		Str("user1_name", "Jane Doe").
		Str("user2_name", "Jane Doe").
		Str("error", "fail").
		Logger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Msg("Logging with some accumulated context.")
		}
	})
}

// =============================================================================
// LOGRUS BENCHMARKS
// =============================================================================

// newLogrusLogger creates a logger with optimal benchmark configuration
func newLogrusLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	return logger
}

// withBenchedLogrusLogger replicates the pattern for Logrus
func withBenchedLogrusLogger(b *testing.B, f func(*logrus.Logger)) {
	logger := newLogrusLogger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func BenchmarkLogrus_NoContext(b *testing.B) {
	withBenchedLogrusLogger(b, func(log *logrus.Logger) {
		log.Info("No context.")
	})
}

func BenchmarkLogrus_10Fields(b *testing.B) {
	withBenchedLogrusLogger(b, func(log *logrus.Logger) {
		log.WithFields(logrus.Fields{
			"one":   1,
			"two":   2,
			"three": 3,
			"four":  4,
			"five":  5,
			"six":   6,
			"seven": 7,
			"eight": 8,
			"nine":  9,
			"ten":   10,
		}).Info("Ten fields, passed at the log site.")
	})
}

func BenchmarkLogrus_DisabledWithoutFields(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.ErrorLevel)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Logging at a disabled level without any structured context.")
		}
	})
}

func BenchmarkLogrus_WithoutFields(b *testing.B) {
	withBenchedLogrusLogger(b, func(log *logrus.Logger) {
		log.Info("Logging without any structured context.")
	})
}

func BenchmarkLogrus_AddingFields(b *testing.B) {
	withBenchedLogrusLogger(b, func(log *logrus.Logger) {
		log.WithFields(logrus.Fields{
			"int":        1,
			"string":     "value",
			"time":       time.Unix(0, 0),
			"user1_name": "Jane Doe",
			"user2_name": "Jane Doe",
			"error":      "fail",
		}).Info("Logging with additional context at each log site.")
	})
}

func BenchmarkLogrus_AccumulatedContext(b *testing.B) {
	logger := newLogrusLogger().WithFields(logrus.Fields{
		"int":        1,
		"string":     "value",
		"time":       time.Unix(0, 0),
		"user1_name": "Jane Doe",
		"user2_name": "Jane Doe",
		"error":      "fail",
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Logging with some accumulated context.")
		}
	})
}

// =============================================================================
// SLOG BENCHMARKS
// =============================================================================

// newSlogLogger creates a logger with optimal benchmark configuration
func newSlogLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// withBenchedSlogLogger replicates the pattern for Slog
func withBenchedSlogLogger(b *testing.B, f func(*slog.Logger)) {
	logger := newSlogLogger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func BenchmarkSlog_NoContext(b *testing.B) {
	withBenchedSlogLogger(b, func(log *slog.Logger) {
		log.Info("No context.")
	})
}

func BenchmarkSlog_10Fields(b *testing.B) {
	withBenchedSlogLogger(b, func(log *slog.Logger) {
		log.Info("Ten fields, passed at the log site.",
			slog.Int("one", 1),
			slog.Int("two", 2),
			slog.Int("three", 3),
			slog.Int("four", 4),
			slog.Int("five", 5),
			slog.Int("six", 6),
			slog.Int("seven", 7),
			slog.Int("eight", 8),
			slog.Int("nine", 9),
			slog.Int("ten", 10),
		)
	})
}

func BenchmarkSlog_DisabledWithoutFields(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Logging at a disabled level without any structured context.")
		}
	})
}

func BenchmarkSlog_WithoutFields(b *testing.B) {
	withBenchedSlogLogger(b, func(log *slog.Logger) {
		log.Info("Logging without any structured context.")
	})
}

func BenchmarkSlog_AddingFields(b *testing.B) {
	withBenchedSlogLogger(b, func(log *slog.Logger) {
		log.Info("Logging with additional context at each log site.",
			slog.Int("int", 1),
			slog.String("string", "value"),
			slog.Time("time", time.Unix(0, 0)),
			slog.String("user1_name", "Jane Doe"),
			slog.String("user2_name", "Jane Doe"),
			slog.String("error", "fail"),
		)
	})
}

func BenchmarkSlog_AccumulatedContext(b *testing.B) {
	logger := newSlogLogger().With(
		slog.Int("int", 1),
		slog.String("string", "value"),
		slog.Time("time", time.Unix(0, 0)),
		slog.String("user1_name", "Jane Doe"),
		slog.String("user2_name", "Jane Doe"),
		slog.String("error", "fail"),
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Logging with some accumulated context.")
		}
	})
}

// =============================================================================
// APEX BENCHMARKS
// =============================================================================

// newApexLogger creates a logger with optimal benchmark configuration
func newApexLogger() *log.Logger {
	logger := &log.Logger{
		Handler: apexjson.New(io.Discard),
		Level:   log.InfoLevel,
	}
	return logger
}

// withBenchedApexLogger replicates the pattern for Apex
func withBenchedApexLogger(b *testing.B, f func(*log.Logger)) {
	logger := newApexLogger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func BenchmarkApex_NoContext(b *testing.B) {
	withBenchedApexLogger(b, func(log *log.Logger) {
		log.Info("No context.")
	})
}

func BenchmarkApex_10Fields(b *testing.B) {
	withBenchedApexLogger(b, func(logger *log.Logger) {
		logger.WithFields(log.Fields{
			"one":   1,
			"two":   2,
			"three": 3,
			"four":  4,
			"five":  5,
			"six":   6,
			"seven": 7,
			"eight": 8,
			"nine":  9,
			"ten":   10,
		}).Info("Ten fields, passed at the log site.")
	})
}

func BenchmarkApex_DisabledWithoutFields(b *testing.B) {
	logger := &log.Logger{
		Handler: apexjson.New(io.Discard),
		Level:   log.ErrorLevel,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Logging at a disabled level without any structured context.")
		}
	})
}

func BenchmarkApex_WithoutFields(b *testing.B) {
	withBenchedApexLogger(b, func(log *log.Logger) {
		log.Info("Logging without any structured context.")
	})
}

func BenchmarkApex_AddingFields(b *testing.B) {
	withBenchedApexLogger(b, func(logger *log.Logger) {
		logger.WithFields(log.Fields{
			"int":        1,
			"string":     "value",
			"time":       time.Unix(0, 0),
			"user1_name": "Jane Doe",
			"user2_name": "Jane Doe",
			"error":      "fail",
		}).Info("Logging with additional context at each log site.")
	})
}

func BenchmarkApex_AccumulatedContext(b *testing.B) {
	logger := newApexLogger()
	ctx := logger.WithFields(log.Fields{
		"int":    1,
		"string": "value",
		"time":   time.Unix(0, 0),
	})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx.Info("Logging with accumulated context.")
		}
	})
}

// =============================================================================
// GO-KIT BENCHMARKS
// =============================================================================

// newGoKitLogger creates a go-kit logger with JSON output to io.Discard
func newGoKitLogger() kitlog.Logger {
	return kitlog.NewJSONLogger(io.Discard)
}

// withBenchedGoKitLogger sets up go-kit logger for benchmarking following the Iris pattern
func withBenchedGoKitLogger(b *testing.B, fn func(kitlog.Logger)) {
	logger := newGoKitLogger()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fn(logger)
		}
	})
}

func BenchmarkGoKit_NoContext(b *testing.B) {
	withBenchedGoKitLogger(b, func(logger kitlog.Logger) {
		level.Info(logger).Log("msg", "The quick brown fox jumps over the lazy dog.")
	})
}

func BenchmarkGoKit_10Fields(b *testing.B) {
	withBenchedGoKitLogger(b, func(logger kitlog.Logger) {
		level.Info(logger).Log(
			"msg", "The quick brown fox jumps over the lazy dog.",
			"rate", 15,
			"low", 16,
			"high", 123.2,
			"cog", 999.99,
			"int", 1,
			"string", "value",
			"time", time.Unix(0, 0),
			"user1_name", "Jane Doe",
			"user2_name", "Jane Doe",
			"error", "fail",
		)
	})
}

func BenchmarkGoKit_DisabledWithoutFields(b *testing.B) {
	logger := kitlog.NewNopLogger() // Disabled logger
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			level.Info(logger).Log("msg", "The quick brown fox jumps over the lazy dog.")
		}
	})
}

func BenchmarkGoKit_WithoutFields(b *testing.B) {
	withBenchedGoKitLogger(b, func(logger kitlog.Logger) {
		level.Info(logger).Log("msg", "The quick brown fox jumps over the lazy dog.")
	})
}

func BenchmarkGoKit_AddingFields(b *testing.B) {
	withBenchedGoKitLogger(b, func(logger kitlog.Logger) {
		level.Info(kitlog.With(logger,
			"int", 1,
			"string", "value",
			"time", time.Unix(0, 0),
			"user1_name", "Jane Doe",
			"user2_name", "Jane Doe",
			"error", "fail",
		)).Log("msg", "Logging with additional context at each log site.")
	})
}

func BenchmarkGoKit_AccumulatedContext(b *testing.B) {
	logger := newGoKitLogger()
	ctx := kitlog.With(logger,
		"int", 1,
		"string", "value",
		"time", time.Unix(0, 0),
	)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			level.Info(ctx).Log("msg", "Logging with accumulated context.")
		}
	})
}

// =============================================================================
// LOG15 BENCHMARKS
// =============================================================================

// newLog15Logger creates a log15 logger with JSON output to io.Discard
func newLog15Logger() log15.Logger {
	logger := log15.New()
	logger.SetHandler(log15.StreamHandler(io.Discard, log15.JsonFormat()))
	return logger
}

// withBenchedLog15Logger sets up log15 logger for benchmarking following the Iris pattern
func withBenchedLog15Logger(b *testing.B, fn func(log15.Logger)) {
	logger := newLog15Logger()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fn(logger)
		}
	})
}

func BenchmarkLog15_NoContext(b *testing.B) {
	withBenchedLog15Logger(b, func(logger log15.Logger) {
		logger.Info("The quick brown fox jumps over the lazy dog.")
	})
}

func BenchmarkLog15_10Fields(b *testing.B) {
	withBenchedLog15Logger(b, func(logger log15.Logger) {
		logger.Info("The quick brown fox jumps over the lazy dog.",
			"rate", 15,
			"low", 16,
			"high", 123.2,
			"cog", 999.99,
			"int", 1,
			"string", "value",
			"time", time.Unix(0, 0),
			"user1_name", "Jane Doe",
			"user2_name", "Jane Doe",
			"error", "fail",
		)
	})
}

func BenchmarkLog15_DisabledWithoutFields(b *testing.B) {
	logger := log15.New()
	logger.SetHandler(log15.DiscardHandler()) // Disabled logger
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("The quick brown fox jumps over the lazy dog.")
		}
	})
}

func BenchmarkLog15_WithoutFields(b *testing.B) {
	withBenchedLog15Logger(b, func(logger log15.Logger) {
		logger.Info("The quick brown fox jumps over the lazy dog.")
	})
}

func BenchmarkLog15_AddingFields(b *testing.B) {
	withBenchedLog15Logger(b, func(logger log15.Logger) {
		logger.Info("Logging with additional context at each log site.",
			"int", 1,
			"string", "value",
			"time", time.Unix(0, 0),
			"user1_name", "Jane Doe",
			"user2_name", "Jane Doe",
			"error", "fail",
		)
	})
}

func BenchmarkLog15_AccumulatedContext(b *testing.B) {
	logger := newLog15Logger()
	ctx := logger.New(
		"int", 1,
		"string", "value",
		"time", time.Unix(0, 0),
	)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx.Info("Logging with accumulated context.")
		}
	})
}
