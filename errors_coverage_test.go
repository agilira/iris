// errors_coverage_test.go: Test for error handling functionality in Iris
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"testing"

	"github.com/agilira/go-errors"
)

// TestUnit_RecoverWithError_NoPanic tests RecoverWithError when no panic occurs
func TestUnit_RecoverWithError_NoPanic(t *testing.T) {
	t.Parallel()

	var result *errors.Error

	// Execute function without panic
	func() {
		defer func() {
			result = RecoverWithError(ErrCodeLoggerCreation)
		}()
		// Normal execution, no panic
	}()

	// Should return nil when no panic
	if result != nil {
		t.Error("RecoverWithError should return nil when no panic occurs")
	}
}

// TestUnit_SafeExecute_Success tests SafeExecute with successful function
func TestUnit_SafeExecute_Success(t *testing.T) {
	t.Parallel()

	executed := false
	testFunc := func() error {
		executed = true
		return nil
	}

	err := SafeExecute(testFunc, "test_operation")

	if err != nil {
		t.Errorf("SafeExecute should return nil for successful function: %v", err)
	}

	if !executed {
		t.Error("Function should have been executed")
	}
}

// TestUnit_SafeExecute_WithError tests SafeExecute with function that returns error
func TestUnit_SafeExecute_WithError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("test error")
	testFunc := func() error {
		return expectedErr
	}

	err := SafeExecute(testFunc, "test_operation")

	if err != expectedErr {
		t.Errorf("SafeExecute should return the function's error, got: %v", err)
	}
}

// TestUnit_NewLoggerError_Coverage tests NewLoggerError function coverage
func TestUnit_NewLoggerError_Coverage(t *testing.T) {
	t.Parallel()

	err := NewLoggerError(ErrCodeLoggerCreation, "test message")

	if err == nil {
		t.Fatal("NewLoggerError should return non-nil error")
	}

	if !errors.HasCode(err, ErrCodeLoggerCreation) {
		t.Error("Error should have the specified error code")
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

// TestUnit_IsLoggerError_Coverage tests IsLoggerError function coverage
func TestUnit_IsLoggerError_Coverage(t *testing.T) {
	t.Parallel()

	// Test with logger error
	loggerErr := NewLoggerError(ErrCodeLoggerCreation, "test")
	if !IsLoggerError(loggerErr, ErrCodeLoggerCreation) {
		t.Error("IsLoggerError should return true for matching logger error")
	}

	// Test with different error code
	if IsLoggerError(loggerErr, ErrCodeInvalidConfig) {
		t.Error("IsLoggerError should return false for non-matching error code")
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard error")
	if IsLoggerError(stdErr, ErrCodeLoggerCreation) {
		t.Error("IsLoggerError should return false for standard error")
	}

	// Test with nil error
	if IsLoggerError(nil, ErrCodeLoggerCreation) {
		t.Error("IsLoggerError should return false for nil error")
	}
}
