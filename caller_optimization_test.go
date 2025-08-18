// caller_optimization_test.go: Test for caller performance optimizations
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
)

// BenchmarkCallerWithoutFunction tests caller info without function names
func BenchmarkCallerWithoutFunction(b *testing.B) {
	falsePtr := false
	config := Config{
		Level:                InfoLevel,
		Writer:               NewDiscardWriter(),
		Format:               JSONFormat,
		EnableCaller:         true,
		EnableCallerFunction: &falsePtr, // NEW: Disable function names
		CallerSkip:           3,
		BufferSize:           1024,
		BatchSize:            8,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Caller test without function names")
	}
}

// BenchmarkCallerWithFunction tests caller info with function names
func BenchmarkCallerWithFunction(b *testing.B) {
	truePtr := true
	config := Config{
		Level:                InfoLevel,
		Writer:               NewDiscardWriter(),
		Format:               JSONFormat,
		EnableCaller:         true,
		EnableCallerFunction: &truePtr, // Enable function names
		CallerSkip:           3,
		BufferSize:           1024,
		BatchSize:            8,
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info("Caller test with function names")
	}
}

// BenchmarkGetCallerOptimized tests optimized getCaller method
func BenchmarkGetCallerOptimized(b *testing.B) {
	tests := []struct {
		name           string
		enableFunction bool
	}{
		{"WithoutFunction", false},
		{"WithFunction", true},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			enableFunctionPtr := test.enableFunction
			config := Config{
				Level:                InfoLevel,
				Writer:               NewDiscardWriter(),
				Format:               JSONFormat,
				EnableCaller:         true,
				EnableCallerFunction: &enableFunctionPtr,
				CallerSkip:           3,
				BufferSize:           1024,
				BatchSize:            8,
			}

			logger, err := New(config)
			if err != nil {
				b.Fatalf("Failed to create logger: %v", err)
			}
			defer logger.Close()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				caller := logger.getCaller()
				_ = caller.File     // Use file
				_ = caller.Line     // Use line
				_ = caller.Function // Use function (may be empty)
			}
		})
	}
}

// TestCallerFunctionControl tests that function control works
func TestCallerFunctionControl(t *testing.T) {
	tests := []struct {
		name           string
		enableFunction bool
		expectFunction bool
	}{
		{"FunctionEnabled", true, true},
		{"FunctionDisabled", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			enableFunctionPtr := test.enableFunction
			config := Config{
				Level:                InfoLevel,
				Writer:               NewDiscardWriter(),
				Format:               JSONFormat,
				EnableCaller:         true,
				EnableCallerFunction: &enableFunctionPtr,
				CallerSkip:           3,
				BufferSize:           1024,
				BatchSize:            8,
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer logger.Close()

			caller := logger.getCaller()

			if !caller.Valid {
				t.Fatal("Expected valid caller info")
			}

			if test.expectFunction {
				if caller.Function == "" {
					t.Error("Expected function name to be populated")
				}
			} else {
				if caller.Function != "" {
					t.Error("Expected function name to be empty")
				}
			}
		})
	}
}
