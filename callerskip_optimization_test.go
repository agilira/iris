// callerskip_optimization_test.go: Test for optimal caller skip values
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// BenchmarkCallerSkipValues tests different caller skip values
func BenchmarkCallerSkipValues(b *testing.B) {
	skipValues := []int{1, 2, 3, 4, 5}

	for _, skip := range skipValues {
		b.Run(fmt.Sprintf("Skip%d", skip), func(b *testing.B) {
			falsePtr := false // Disable function names for fair comparison
			config := Config{
				Level:                InfoLevel,
				Writer:               NewDiscardWriter(),
				Format:               JSONFormat,
				EnableCaller:         true,
				EnableCallerFunction: &falsePtr,
				CallerSkip:           skip,
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
				logger.Info("Skip optimization test")
			}
		})
	}
}

// BenchmarkRuntimeCallerDirect tests direct runtime.Caller performance
func BenchmarkRuntimeCallerDirect(b *testing.B) {
	skipValues := []int{1, 2, 3, 4, 5}

	for _, skip := range skipValues {
		b.Run(fmt.Sprintf("Skip%d", skip), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				pc, file, line, ok := runtime.Caller(skip)
				if ok {
					_ = pc
					_ = file
					_ = line
				}
			}
		})
	}
}

// TestCallerSkipAccuracy tests which skip value gives correct caller info
func TestCallerSkipAccuracy(t *testing.T) {
	skipValues := []int{1, 2, 3, 4} // Remove skip=5 since it goes beyond stack

	for _, skip := range skipValues {
		t.Run(fmt.Sprintf("Skip%d", skip), func(t *testing.T) {
			falsePtr := false
			config := Config{
				Level:                InfoLevel,
				Writer:               NewDiscardWriter(),
				Format:               JSONFormat,
				EnableCaller:         true,
				EnableCallerFunction: &falsePtr,
				CallerSkip:           skip,
				BufferSize:           1024,
				BatchSize:            8,
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer logger.Close()

			// This is the call we expect to be captured
			testCallerAccuracy(t, logger, skip)
		})
	}
}

func testCallerAccuracy(t *testing.T, logger *Logger, expectedSkip int) {
	caller := logger.getCaller()
	if !caller.Valid {
		t.Error("Expected valid caller info")
		return
	}

	// Check if the caller info makes sense
	t.Logf("Skip %d - File: %s, Line: %d", expectedSkip, caller.File, caller.Line)

	// For skip values 1-2, we expect our test file
	// For skip values 3+, we expect testing framework or runtime
	if expectedSkip <= 2 {
		if !strings.Contains(caller.File, "callerskip_optimization_test.go") {
			t.Errorf("Expected file to contain test filename, got: %s", caller.File)
		}
	} else {
		// For higher skip values, we expect testing framework or runtime
		if !strings.Contains(caller.File, "testing.go") && !strings.Contains(caller.File, "asm_") {
			t.Logf("Skip %d captured: %s (this may be valid)", expectedSkip, caller.File)
		}
	}
}
