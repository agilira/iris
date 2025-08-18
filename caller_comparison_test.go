package iris

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BenchmarkIrisCallerInfo tests Iris with caller info enabled
func BenchmarkIrisCallerInfo(b *testing.B) {
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
		logger.Info("Benchmark message with caller info",
			String("key1", "value1"),
			Int("key2", 42),
			Duration("key3", time.Millisecond),
		)
	}
}

// BenchmarkZapCallerInfo tests Zap with caller info enabled
func BenchmarkZapCallerInfo(b *testing.B) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}
	config.DisableCaller = false // Enable caller info
	config.DisableStacktrace = true

	logger, err := config.Build()
	if err != nil {
		b.Fatalf("Failed to create Zap logger: %v", err)
	}
	defer logger.Sync()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message with caller info",
			zap.String("key1", "value1"),
			zap.Int("key2", 42),
			zap.Duration("key3", time.Millisecond),
		)
	}
}

// BenchmarkIrisCallerDisabled tests Iris without caller info
func BenchmarkIrisCallerDisabled(b *testing.B) {
	config := Config{
		Level:        InfoLevel,
		Writer:       NewDiscardWriter(),
		Format:       JSONFormat,
		EnableCaller: false,
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
		logger.Info("Benchmark message without caller info",
			String("key1", "value1"),
			Int("key2", 42),
			Duration("key3", time.Millisecond),
		)
	}
}

// BenchmarkZapCallerDisabled tests Zap without caller info
func BenchmarkZapCallerDisabled(b *testing.B) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}
	config.DisableCaller = true // Disable caller info
	config.DisableStacktrace = true

	logger, err := config.Build()
	if err != nil {
		b.Fatalf("Failed to create Zap logger: %v", err)
	}
	defer logger.Sync()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message without caller info",
			zap.String("key1", "value1"),
			zap.Int("key2", 42),
			zap.Duration("key3", time.Millisecond),
		)
	}
}

// BenchmarkIrisCallerConsole tests Iris console format with caller
func BenchmarkIrisCallerConsole(b *testing.B) {
	config := Config{
		Level:        InfoLevel,
		Writer:       NewDiscardWriter(),
		Format:       ConsoleFormat,
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
		logger.Info("Console benchmark with caller",
			String("key1", "value1"),
			Int("key2", 42),
		)
	}
}

// BenchmarkZapCallerConsole tests Zap console format with caller
func BenchmarkZapCallerConsole(b *testing.B) {
	config := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(config)
	core := zapcore.NewCore(encoder, zapcore.AddSync(&zapDiscardWriter{}), zapcore.InfoLevel)
	logger := zap.New(core, zap.AddCaller())
	defer logger.Sync()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Console benchmark with caller",
			zap.String("key1", "value1"),
			zap.Int("key2", 42),
		)
	}
}

// BenchmarkIrisCallerFastText tests Iris FastText format with caller
func BenchmarkIrisCallerFastText(b *testing.B) {
	config := Config{
		Level:        InfoLevel,
		Writer:       NewDiscardWriter(),
		Format:       FastTextFormat,
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
		logger.Info("FastText benchmark with caller",
			String("key1", "value1"),
			Int("key2", 42),
		)
	}
}

// zapDiscardWriter for Zap benchmarks
type zapDiscardWriter struct{}

func (w *zapDiscardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w *zapDiscardWriter) Sync() error {
	return nil
}
