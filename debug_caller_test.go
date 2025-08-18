// debug_caller_test.go: Debug caller stack to understand the issue
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"runtime"
	"testing"
)

// TestDebugCallerStack helps us understand the call stack
func TestDebugCallerStack(t *testing.T) {
	t.Log("=== DEBUGGING CALLER STACK ===")

	// Print the entire call stack to understand what's happening
	for i := 0; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			t.Logf("Skip %d: No more frames", i)
			break
		}

		funcName := "unknown"
		if fn := runtime.FuncForPC(pc); fn != nil {
			funcName = fn.Name()
		}

		t.Logf("Skip %d: %s:%d in %s", i, file, line, funcName)
	}

	// Now test what happens when we call through the logger
	t.Log("\n=== TESTING THROUGH LOGGER ===")
	testLoggerCaller(t)
}

func testLoggerCaller(t *testing.T) {
	falsePtr := false
	config := Config{
		Level:                InfoLevel,
		Writer:               NewDiscardWriter(),
		Format:               JSONFormat,
		EnableCaller:         true,
		EnableCallerFunction: &falsePtr,
		CallerSkip:           2, // Current default
		BufferSize:           1024,
		BatchSize:            8,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// This should capture the caller info for this function
	caller := logger.getCaller()
	t.Logf("Logger caller with skip 2: %s:%d", caller.File, caller.Line)

	// Test with different skip values
	for skip := 1; skip <= 5; skip++ {
		logger.callerSkip = skip
		caller := logger.getCaller()
		if caller.Valid {
			t.Logf("Skip %d: %s:%d", skip, caller.File, caller.Line)
		} else {
			t.Logf("Skip %d: Invalid caller", skip)
		}
	}
}
