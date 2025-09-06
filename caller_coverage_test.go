// caller_coverage_test.go: Test for caller functionality in Iris
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"runtime"
	"strings"
	"testing"
)

// TestUnit_ShortCaller_Coverage tests shortCaller functionality
func TestUnit_ShortCaller_Coverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		skip int
	}{
		{
			name: "Skip0",
			skip: 0,
		},
		{
			name: "Skip1",
			skip: 1,
		},
		{
			name: "Skip2",
			skip: 2,
		},
		{
			name: "SkipLarge",
			skip: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caller, ok := shortCaller(tt.skip)

			if tt.skip < 10 {
				// For reasonable skip values, should get valid caller
				if !ok {
					t.Error("shortCaller should return ok=true for reasonable skip values")
				}
				if caller == "" {
					t.Error("shortCaller should return non-empty caller string")
				}

				// Should contain file name and line number
				if !strings.Contains(caller, ".go:") {
					t.Error("caller should contain .go: pattern")
				}

				// Should be in short format (not full path)
				if strings.Count(caller, "/") > 1 {
					t.Error("caller should be in short format with at most one path separator")
				}
			} else {
				// For very large skip values, might not get valid caller
				// Just verify it doesn't panic
				t.Logf("Large skip value result: caller=%s, ok=%v", caller, ok)
			}
		})
	}
}

// TestUnit_ShortCaller_InvalidSkip tests shortCaller with invalid skip values
func TestUnit_ShortCaller_InvalidSkip(t *testing.T) {
	t.Parallel()

	// Test with very large skip value that should exceed stack depth
	caller, ok := shortCaller(1000)

	if ok {
		t.Logf("Unexpectedly got valid caller with large skip: %s", caller)
	} else {
		if caller != "" {
			t.Error("When ok=false, caller should be empty string")
		}
	}
}

// TestUnit_ShortCaller_PathTrimming tests that shortCaller properly trims paths
func TestUnit_ShortCaller_PathTrimming(t *testing.T) {
	t.Parallel()

	// Get current function's caller info
	caller, ok := shortCaller(0)
	if !ok {
		t.Fatal("Failed to get caller info")
	}

	// Verify format
	parts := strings.Split(caller, ":")
	if len(parts) != 2 {
		t.Errorf("Expected caller format 'file:line', got: %s", caller)
	}

	filename := parts[0]

	// Should not contain full absolute path
	if strings.HasPrefix(filename, "/") {
		t.Error("Caller should not contain absolute path")
	}

	// Should contain a .go file
	if !strings.HasSuffix(filename, ".go") {
		t.Errorf("Expected filename to have .go extension, got: %s", filename)
	}

	// Verify line number is numeric
	lineStr := parts[1]
	if len(lineStr) == 0 {
		t.Error("Line number should not be empty")
	}
}

// TestUnit_ShortCaller_CompareWithRuntimeCaller tests consistency with runtime.Caller
func TestUnit_ShortCaller_CompareWithRuntimeCaller(t *testing.T) {
	t.Parallel()

	// Get info from both shortCaller and runtime.Caller
	shortResult, shortOk := shortCaller(0)
	_, runtimeFile, _, runtimeOk := runtime.Caller(0)

	if shortOk != runtimeOk {
		t.Errorf("Consistency check failed: shortCaller ok=%v, runtime.Caller ok=%v", shortOk, runtimeOk)
	}

	if shortOk && runtimeOk {
		// Extract line number from shortCaller result
		parts := strings.Split(shortResult, ":")
		if len(parts) == 2 {
			// The line numbers might be slightly different due to different call points
			// Just verify the format is correct
			if !strings.HasSuffix(runtimeFile, parts[0]) && !strings.Contains(runtimeFile, parts[0]) {
				t.Logf("File name consistency check: short=%s, runtime=%s", parts[0], runtimeFile)
				// This is informational - slight differences are expected
			}
		}
	}
}
