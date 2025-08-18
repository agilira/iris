// stacktrace_benchmark_test.go: Stack trace performance benchmarks
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"testing"
)

// BenchmarkStackTraceCapture tests performance when stack trace is actually captured
func BenchmarkStackTraceCapture(b *testing.B) {
	tests := []struct {
		name  string
		level Level
	}{
		{"ErrorLevel", ErrorLevel},
		{"WarnLevel", WarnLevel},
		{"InfoLevel", InfoLevel},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			config := Config{
				Level:           DebugLevel,
				Writer:          io.Discard,
				Format:          JSONFormat,
				BufferSize:      1024,
				BatchSize:       64,
				StackTraceLevel: test.level, // Stack trace captured at this level
			}

			logger, err := New(config)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Trigger stack trace capture
				logger.Error("Error message that will capture stack trace")
			}
		})
	}
}

// BenchmarkStackTraceVsNoStackTrace compares performance with and without stack trace
func BenchmarkStackTraceVsNoStackTrace(b *testing.B) {
	tests := []struct {
		name            string
		stackTraceLevel Level
	}{
		{"WithStackTrace", ErrorLevel},
		{"WithoutStackTrace", FatalLevel + 1}, // Level too high to trigger
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			config := Config{
				Level:           DebugLevel,
				Writer:          io.Discard,
				Format:          JSONFormat,
				BufferSize:      1024,
				BatchSize:       64,
				StackTraceLevel: test.stackTraceLevel,
			}

			logger, err := New(config)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Error("Error message")
			}
		})
	}
}

// BenchmarkStackTraceFormats tests stack trace performance across different formats
func BenchmarkStackTraceFormats(b *testing.B) {
	formats := []struct {
		name   string
		format Format
	}{
		{"JSON", JSONFormat},
		{"Console", ConsoleFormat},
		{"FastText", FastTextFormat},
	}

	for _, format := range formats {
		b.Run(format.name, func(b *testing.B) {
			config := Config{
				Level:           DebugLevel,
				Writer:          io.Discard,
				Format:          format.format,
				BufferSize:      1024,
				BatchSize:       64,
				StackTraceLevel: ErrorLevel, // Stack trace for errors
			}

			logger, err := New(config)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Error("Error with stack trace")
			}
		})
	}
}
