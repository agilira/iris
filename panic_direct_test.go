// panic_direct_test.go: Direct test for Panic function to improve coverage
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
)

// TestLogger_Panic_DirectCoverage tests Panic method for coverage
func TestLogger_Panic_DirectCoverage(t *testing.T) {
	// Create a test that captures the panic without using subprocess
	defer func() {
		if r := recover(); r != nil {
			// Expected panic - this is good
			if r != "test panic for coverage" {
				t.Errorf("Expected panic message 'test panic for coverage', got: %v", r)
			}
		} else {
			t.Error("Expected Panic to panic, but it didn't")
		}
	}()

	// Create logger for testing
	syncer := &criticalTestSyncer{}
	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewTextEncoder(),
		Output:  syncer,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Start()
	defer safeClosePanicLogger(t, logger)

	// Call Panic - this should trigger the panic that we catch above
	logger.Panic("test panic for coverage", String("key", "value"))

	// This line should never be reached
	t.Error("This line should not be reached after Panic")
}

// Helper function for safe logger cleanup
func safeClosePanicLogger(t *testing.T, logger *Logger) {
	if err := logger.Close(); err != nil {
		t.Logf("Warning: Error closing logger in test: %v", err)
	}
}
