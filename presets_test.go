// presets_test.go: Test configuration presets
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"testing"
)

func TestPresets(t *testing.T) {
	tests := []struct {
		name     string
		create   func() (*Logger, error)
		level    Level
		format   Format
		buffSize int64
	}{
		{"Development", NewDevelopment, DebugLevel, JSONFormat, 1024},
		{"Production", NewProduction, InfoLevel, JSONFormat, 8192},
		{"Example", NewExample, InfoLevel, JSONFormat, 512},
		{"UltraFast", NewUltraFast, InfoLevel, BinaryFormat, 16384},
		{"FastText", NewFastText, DebugLevel, FastTextFormat, 2048},
		{"DevelopmentWithStackTrace", NewDevelopmentWithStackTrace, DebugLevel, ConsoleFormat, 1024},
		{"DebugWithStackTrace", NewDebugWithStackTrace, DebugLevel, ConsoleFormat, 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := tt.create()
			if err != nil {
				t.Fatalf("Failed to create %s logger: %v", tt.name, err)
			}
			defer logger.Close()

			// Check level
			if logger.level != tt.level {
				t.Errorf("Expected level %v, got %v", tt.level, logger.level)
			}

			// Check format
			if logger.format != tt.format {
				t.Errorf("Expected format %v, got %v", tt.format, logger.format)
			}

			// Test that logger works
			logger.Info("test message", String("preset", tt.name))
		})
	}
}

func TestConfigFunctions(t *testing.T) {
	tests := []struct {
		name   string
		config func() Config
		level  Level
		format Format
	}{
		{"Development", DevelopmentConfig, DebugLevel, JSONFormat},
		{"Production", ProductionConfig, InfoLevel, JSONFormat},
		{"Example", ExampleConfig, InfoLevel, JSONFormat},
		{"DevelopmentWithStackTrace", DevelopmentWithStackTraceConfig, DebugLevel, ConsoleFormat},
		{"DebugWithStackTrace", DebugWithStackTraceConfig, DebugLevel, ConsoleFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config()

			if config.Level != tt.level {
				t.Errorf("Expected level %v, got %v", tt.level, config.Level)
			}

			if config.Format != tt.format {
				t.Errorf("Expected format %v, got %v", tt.format, config.Format)
			}

			// Test that config can be used to create logger
			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger from %s config: %v", tt.name, err)
			}
			defer logger.Close()

			logger.Info("test message", String("config", tt.name))
		})
	}
}

func TestExamplePreset(t *testing.T) {
	// Example preset should have deterministic output (no timestamps)
	logger, err := NewExample()
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// Check that timestamps are disabled
	if !logger.disableTimestamp {
		t.Error("Example preset should disable timestamps for deterministic output")
	}
}

func TestStackTracePresets(t *testing.T) {
	// Test stack trace configuration values
	devConfig := DevelopmentWithStackTraceConfig()
	debugConfig := DebugWithStackTraceConfig()

	if devConfig.StackTraceLevel != ErrorLevel {
		t.Errorf("DevelopmentWithStackTrace should have ErrorLevel stack traces, got %v", devConfig.StackTraceLevel)
	}

	if debugConfig.StackTraceLevel != WarnLevel {
		t.Errorf("DebugWithStackTrace should have WarnLevel stack traces, got %v", debugConfig.StackTraceLevel)
	}

	if devConfig.Format != ConsoleFormat {
		t.Errorf("DevelopmentWithStackTrace should use ConsoleFormat, got %v", devConfig.Format)
	}

	if debugConfig.Format != ConsoleFormat {
		t.Errorf("DebugWithStackTrace should use ConsoleFormat, got %v", debugConfig.Format)
	}

	if !debugConfig.EnableCaller {
		t.Error("DebugWithStackTrace should enable caller info")
	}
}

func TestUltraFastPreset(t *testing.T) {
	// UltraFast preset should have all optimizations enabled
	logger, err := NewUltraFast()
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// Check optimizations
	if !logger.ultraFast {
		t.Error("UltraFast preset should enable ultraFast mode")
	}
	if !logger.disableTimestamp {
		t.Error("UltraFast preset should disable timestamps")
	}
	if logger.enableCaller {
		t.Error("UltraFast preset should disable caller info")
	}
	if logger.format != BinaryFormat {
		t.Error("UltraFast preset should use binary format")
	}
}

func BenchmarkPresets(b *testing.B) {
	presets := []struct {
		name   string
		logger *Logger
	}{
		{"Development", createBenchmarkVersion(NewDevelopment)},
		{"Production", createBenchmarkVersion(NewProduction)},
		{"Example", createBenchmarkVersion(NewExample)},
		{"UltraFast", createBenchmarkVersion(NewUltraFast)},
		{"FastText", createBenchmarkVersion(NewFastText)},
		{"DevelopmentWithStackTrace", createBenchmarkVersion(NewDevelopmentWithStackTrace)},
		{"DebugWithStackTrace", createBenchmarkVersion(NewDebugWithStackTrace)},
	}

	for _, preset := range presets {
		b.Run(preset.name, func(b *testing.B) {
			logger := preset.logger

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("Benchmark test message")
			}
		})
	}
}

// createBenchmarkVersion creates a version of the logger with io.Discard for benchmarking
func createBenchmarkVersion(presetFunc func() (*Logger, error)) *Logger {
	logger, err := presetFunc()
	if err != nil {
		panic(err) // Should not happen in benchmarks
	}

	// Create a new config based on the original but with io.Discard writer
	config := Config{
		Level:      logger.level,
		Writer:     io.Discard, // Redirect to discard for benchmarks
		Format:     logger.format,
		BufferSize: 1024, // Standard buffer for benchmarks
		BatchSize:  64,   // Standard batch for benchmarks
	}

	benchLogger, err := New(config)
	if err != nil {
		panic(err) // Should not happen in benchmarks
	}

	return benchLogger
}

func BenchmarkPresetCreation(b *testing.B) {
	presets := []struct {
		name   string
		create func() (*Logger, error)
	}{
		{"Development", NewDevelopment},
		{"Production", NewProduction},
		{"Example", NewExample},
		{"UltraFast", NewUltraFast},
		{"FastText", NewFastText},
		{"DevelopmentWithStackTrace", NewDevelopmentWithStackTrace},
		{"DebugWithStackTrace", NewDebugWithStackTrace},
	}

	for _, preset := range presets {
		b.Run(preset.name+"_Creation", func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				logger, err := preset.create()
				if err != nil {
					b.Fatal(err)
				}
				logger.Close()
			}
		})
	}
}
