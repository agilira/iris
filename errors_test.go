// errors_test.go: Comprehensive test suite for Iris logging library error handling
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/agilira/go-errors"
)

// TestErrorCodes verifies all error codes follow naming conventions
func TestErrorCodes(t *testing.T) {
	testCases := []struct {
		name string
		code errors.ErrorCode
	}{
		{"Logger Creation", ErrCodeLoggerCreation},
		{"Logger Not Found", ErrCodeLoggerNotFound},
		{"Logger Disabled", ErrCodeLoggerDisabled},
		{"Logger Closed", ErrCodeLoggerClosed},
		{"Invalid Config", ErrCodeInvalidConfig},
		{"Invalid Level", ErrCodeInvalidLevel},
		{"Invalid Format", ErrCodeInvalidFormat},
		{"Invalid Output", ErrCodeInvalidOutput},
		{"Invalid Field", ErrCodeInvalidField},
		{"Encoding Failed", ErrCodeEncodingFailed},
		{"Field Type Mismatch", ErrCodeFieldTypeMismatch},
		{"Buffer Overflow", ErrCodeBufferOverflow},
		{"Writer Not Available", ErrCodeWriterNotAvailable},
		{"Write Failed", ErrCodeWriteFailed},
		{"Flush Failed", ErrCodeFlushFailed},
		{"Sync Failed", ErrCodeSyncFailed},
		{"Memory Allocation", ErrCodeMemoryAllocation},
		{"Pool Exhausted", ErrCodePoolExhausted},
		{"Timeout", ErrCodeTimeout},
		{"Resource Limit", ErrCodeResourceLimit},
		{"Hook Execution", ErrCodeHookExecution},
		{"Middleware Chain", ErrCodeMiddlewareChain},
		{"Filter Failed", ErrCodeFilterFailed},
		{"File Open", ErrCodeFileOpen},
		{"File Write", ErrCodeFileWrite},
		{"File Rotation", ErrCodeFileRotation},
		{"Permission Denied", ErrCodePermissionDenied},
		{"Logger Execution", ErrCodeLoggerExecution},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify error code is not empty
			if string(tc.code) == "" {
				t.Errorf("Error code for %s is empty", tc.name)
			}

			// Verify error code follows IRIS_ prefix convention
			if !strings.HasPrefix(string(tc.code), "IRIS_") {
				t.Errorf("Error code %s does not follow IRIS_ prefix convention", tc.code)
			}

			// Verify error code contains only uppercase letters, numbers, and underscores
			codeStr := string(tc.code)
			for _, char := range codeStr {
				if (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' {
					t.Errorf("Error code %s contains invalid character: %c", tc.code, char)
				}
			}
		})
	}
}

// TestNewLoggerError tests creating new logger errors
func TestNewLoggerError(t *testing.T) {
	testCases := []struct {
		name    string
		code    errors.ErrorCode
		message string
	}{
		{"Valid error", ErrCodeLoggerCreation, "Logger creation failed"},
		{"Another valid error", ErrCodeInvalidConfig, "Invalid configuration provided"},
		{"Empty message", ErrCodeTimeout, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewLoggerError(tc.code, tc.message)

			if err == nil {
				t.Fatal("NewLoggerError should not return nil")
			}

			if err.Code != tc.code {
				t.Errorf("Expected code %s, got %s", tc.code, err.Code)
			}

			if err.Message != tc.message {
				t.Errorf("Expected message %s, got %s", tc.message, err.Message)
			}

			if err.Severity != "error" {
				t.Errorf("Expected severity 'error', got %s", err.Severity)
			}

			// Check context
			if err.Context == nil {
				t.Error("Context should not be nil")
			}

			component, ok := err.Context["component"]
			if !ok || component != "iris_logger" {
				t.Errorf("Expected component 'iris_logger', got %v", component)
			}

			// Check timestamp
			if err.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}

			// Check caller information
			if _, ok := err.Context["caller_func"]; !ok {
				t.Error("Caller function should be in context")
			}

			if _, ok := err.Context["caller_file"]; !ok {
				t.Error("Caller file should be in context")
			}

			if _, ok := err.Context["caller_line"]; !ok {
				t.Error("Caller line should be in context")
			}
		})
	}
}

// TestNewLoggerErrorWithField tests creating errors with field information
func TestNewLoggerErrorWithField(t *testing.T) {
	testCases := []struct {
		name    string
		code    errors.ErrorCode
		message string
		field   string
		value   string
	}{
		{"Valid field error", ErrCodeInvalidField, "Invalid field type", "level", "invalid_level"},
		{"Empty field", ErrCodeFieldTypeMismatch, "Type mismatch", "", "some_value"},
		{"Empty value", ErrCodeInvalidField, "Invalid field", "format", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewLoggerErrorWithField(tc.code, tc.message, tc.field, tc.value)

			if err == nil {
				t.Fatal("NewLoggerErrorWithField should not return nil")
			}

			if err.Code != tc.code {
				t.Errorf("Expected code %s, got %s", tc.code, err.Code)
			}

			if err.Message != tc.message {
				t.Errorf("Expected message %s, got %s", tc.message, err.Message)
			}

			if err.Field != tc.field {
				t.Errorf("Expected field %s, got %s", tc.field, err.Field)
			}

			if err.Value != tc.value {
				t.Errorf("Expected value %s, got %s", tc.value, err.Value)
			}

			if err.Severity != "error" {
				t.Errorf("Expected severity 'error', got %s", err.Severity)
			}

			// Check context
			component, ok := err.Context["component"]
			if !ok || component != "iris_logger" {
				t.Errorf("Expected component 'iris_logger', got %v", component)
			}
		})
	}
}

// TestWrapLoggerError tests wrapping existing errors
func TestWrapLoggerError(t *testing.T) {
	originalErr := fmt.Errorf("original error message")
	code := ErrCodeEncodingFailed
	message := "Failed to encode log entry"

	wrappedErr := WrapLoggerError(originalErr, code, message)

	if wrappedErr == nil {
		t.Fatal("WrapLoggerError should not return nil")
	}

	if wrappedErr.Code != code {
		t.Errorf("Expected code %s, got %s", code, wrappedErr.Code)
	}

	if wrappedErr.Message != message {
		t.Errorf("Expected message %s, got %s", message, wrappedErr.Message)
	}

	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected cause to be original error, got %v", wrappedErr.Cause)
	}

	// Check that unwrapping works
	if wrappedErr.Unwrap() != originalErr {
		t.Error("Unwrap should return the original error")
	}

	// Check context
	component, ok := wrappedErr.Context["component"]
	if !ok || component != "iris_logger" {
		t.Errorf("Expected component 'iris_logger', got %v", component)
	}
}

// TestIsRetryableError tests checking if errors are retryable
func TestIsRetryableError(t *testing.T) {
	testCases := []struct {
		name      string
		err       error
		retryable bool
		setup     func() error
	}{
		{
			name:      "Non-retryable iris error",
			retryable: false,
			setup: func() error {
				return NewLoggerError(ErrCodeLoggerCreation, "Test error")
			},
		},
		{
			name:      "Retryable iris error",
			retryable: true,
			setup: func() error {
				return NewLoggerError(ErrCodeTimeout, "Timeout error").AsRetryable()
			},
		},
		{
			name:      "Standard Go error",
			retryable: false,
			setup: func() error {
				return fmt.Errorf("standard error")
			},
		},
		{
			name:      "Nil error",
			retryable: false,
			setup: func() error {
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.setup()
			result := IsRetryableError(err)

			if result != tc.retryable {
				t.Errorf("Expected retryable %v, got %v", tc.retryable, result)
			}
		})
	}
}

// TestGetErrorCode tests extracting error codes
func TestGetErrorCode(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		expectedCode errors.ErrorCode
	}{
		{
			name:         "Iris error",
			err:          NewLoggerError(ErrCodeLoggerCreation, "Test error"),
			expectedCode: ErrCodeLoggerCreation,
		},
		{
			name:         "Standard Go error",
			err:          fmt.Errorf("standard error"),
			expectedCode: "",
		},
		{
			name:         "Nil error",
			err:          nil,
			expectedCode: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := GetErrorCode(tc.err)

			if code != tc.expectedCode {
				t.Errorf("Expected code %s, got %s", tc.expectedCode, code)
			}
		})
	}
}

// TestGetUserMessage tests extracting user-friendly messages
func TestGetUserMessage(t *testing.T) {
	testCases := []struct {
		name            string
		err             error
		expectedMessage string
	}{
		{
			name:            "Iris error with user message",
			err:             NewLoggerError(ErrCodeLoggerCreation, "Technical message").WithUserMessage("User-friendly message"),
			expectedMessage: "User-friendly message",
		},
		{
			name:            "Iris error without user message",
			err:             NewLoggerError(ErrCodeLoggerCreation, "Technical message"),
			expectedMessage: "Technical message",
		},
		{
			name:            "Standard Go error",
			err:             fmt.Errorf("standard error"),
			expectedMessage: "standard error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			message := GetUserMessage(tc.err)

			if message != tc.expectedMessage {
				t.Errorf("Expected message %s, got %s", tc.expectedMessage, message)
			}
		})
	}
}

// TestIsLoggerError tests checking for specific logger error codes
func TestIsLoggerError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		code     errors.ErrorCode
		expected bool
	}{
		{
			name:     "Matching error code",
			err:      NewLoggerError(ErrCodeLoggerCreation, "Test error"),
			code:     ErrCodeLoggerCreation,
			expected: true,
		},
		{
			name:     "Non-matching error code",
			err:      NewLoggerError(ErrCodeLoggerCreation, "Test error"),
			code:     ErrCodeTimeout,
			expected: false,
		},
		{
			name:     "Standard Go error",
			err:      fmt.Errorf("standard error"),
			code:     ErrCodeLoggerCreation,
			expected: false,
		},
		{
			name:     "Wrapped error with matching code",
			err:      WrapLoggerError(fmt.Errorf("original"), ErrCodeTimeout, "Wrapped"),
			code:     ErrCodeTimeout,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsLoggerError(tc.err, tc.code)

			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestRecoverWithError tests panic recovery functionality in a controlled way
func TestRecoverWithError(t *testing.T) {
	// Test that RecoverWithError returns nil when there's no panic
	err := RecoverWithError(ErrCodeLoggerExecution)
	if err != nil {
		t.Errorf("RecoverWithError should return nil when no panic occurs, got: %v", err)
	}

	// Test the recovery pattern - simulate recovery with proper isolation
	t.Run("PanicRecoverySimulation", func(t *testing.T) {
		// Test string panic recovery
		var recoveredError *errors.Error
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Simulate what RecoverWithError does
					message := fmt.Sprintf("Panic recovered: %v", r)
					recoveredError = NewLoggerError(ErrCodeLoggerExecution, message)
					_ = recoveredError.WithContext("panic_value", r)
					_ = recoveredError.WithContext("recovery_time", time.Now().UTC())

					// Add stack trace
					buf := make([]byte, 4096)
					stackSize := runtime.Stack(buf, false)
					_ = recoveredError.WithContext("panic_stack", string(buf[:stackSize]))
				}
			}()
			panic("test panic string")
		}()

		if recoveredError == nil {
			t.Fatal("Expected panic recovery to create an error")
		}

		if !strings.Contains(recoveredError.Error(), "test panic string") {
			t.Errorf("Expected error message to contain panic value, got: %s", recoveredError.Error())
		}

		// Test error panic recovery
		recoveredError = nil
		originalError := fmt.Errorf("original error")
		func() {
			defer func() {
				if r := recover(); r != nil {
					message := fmt.Sprintf("Panic recovered: %v", r)
					recoveredError = NewLoggerError(ErrCodeLoggerExecution, message)
					_ = recoveredError.WithContext("panic_value", r)
				}
			}()
			panic(originalError)
		}()

		if recoveredError == nil {
			t.Fatal("Expected panic recovery to create an error for error panic")
		}

		if !strings.Contains(recoveredError.Error(), originalError.Error()) {
			t.Errorf("Expected error message to contain original error, got: %s", recoveredError.Error())
		}

		// Test integer panic recovery
		recoveredError = nil
		func() {
			defer func() {
				if r := recover(); r != nil {
					message := fmt.Sprintf("Panic recovered: %v", r)
					recoveredError = NewLoggerError(ErrCodeLoggerExecution, message)
					_ = recoveredError.WithContext("panic_value", r)
				}
			}()
			panic(42)
		}()

		if recoveredError == nil {
			t.Fatal("Expected panic recovery to create an error for integer panic")
		}

		if !strings.Contains(recoveredError.Error(), "42") {
			t.Errorf("Expected error message to contain '42', got: %s", recoveredError.Error())
		}
	})

}

// TestErrorCodeConstants tests that all error code constants are properly defined
func TestErrorCodeConstants(t *testing.T) {
	errorCodes := map[string]errors.ErrorCode{
		"ErrCodeLoggerCreation":     ErrCodeLoggerCreation,
		"ErrCodeLoggerNotFound":     ErrCodeLoggerNotFound,
		"ErrCodeLoggerDisabled":     ErrCodeLoggerDisabled,
		"ErrCodeLoggerClosed":       ErrCodeLoggerClosed,
		"ErrCodeInvalidConfig":      ErrCodeInvalidConfig,
		"ErrCodeInvalidLevel":       ErrCodeInvalidLevel,
		"ErrCodeInvalidFormat":      ErrCodeInvalidFormat,
		"ErrCodeInvalidOutput":      ErrCodeInvalidOutput,
		"ErrCodeInvalidField":       ErrCodeInvalidField,
		"ErrCodeEncodingFailed":     ErrCodeEncodingFailed,
		"ErrCodeFieldTypeMismatch":  ErrCodeFieldTypeMismatch,
		"ErrCodeBufferOverflow":     ErrCodeBufferOverflow,
		"ErrCodeWriterNotAvailable": ErrCodeWriterNotAvailable,
		"ErrCodeWriteFailed":        ErrCodeWriteFailed,
		"ErrCodeFlushFailed":        ErrCodeFlushFailed,
		"ErrCodeSyncFailed":         ErrCodeSyncFailed,
		"ErrCodeMemoryAllocation":   ErrCodeMemoryAllocation,
		"ErrCodePoolExhausted":      ErrCodePoolExhausted,
		"ErrCodeTimeout":            ErrCodeTimeout,
		"ErrCodeResourceLimit":      ErrCodeResourceLimit,
		"ErrCodeHookExecution":      ErrCodeHookExecution,
		"ErrCodeMiddlewareChain":    ErrCodeMiddlewareChain,
		"ErrCodeFilterFailed":       ErrCodeFilterFailed,
		"ErrCodeFileOpen":           ErrCodeFileOpen,
		"ErrCodeFileWrite":          ErrCodeFileWrite,
		"ErrCodeFileRotation":       ErrCodeFileRotation,
		"ErrCodePermissionDenied":   ErrCodePermissionDenied,
		"ErrCodeLoggerExecution":    ErrCodeLoggerExecution,
	}

	for name, code := range errorCodes {
		t.Run(name, func(t *testing.T) {
			if string(code) == "" {
				t.Errorf("Error code %s is empty", name)
			}

			if !strings.HasPrefix(string(code), "IRIS_") {
				t.Errorf("Error code %s does not have IRIS_ prefix: %s", name, code)
			}
		})
	}
}

// TestOSAwareness tests OS-specific behavior where applicable
func TestOSAwareness(t *testing.T) {
	// Test that the error handling works correctly on the current OS
	currentOS := runtime.GOOS

	// Create an error and verify it contains OS information in context if needed
	err := NewLoggerError(ErrCodeLoggerCreation, "OS-aware test")

	// Add OS-specific context
	_ = err.WithContext("os", currentOS)
	_ = err.WithContext("arch", runtime.GOARCH)

	if osValue, ok := err.Context["os"]; !ok || osValue != currentOS {
		t.Errorf("Expected OS %s in context, got %v", currentOS, osValue)
	}

	if archValue, ok := err.Context["arch"]; !ok || archValue != runtime.GOARCH {
		t.Errorf("Expected architecture %s in context, got %v", runtime.GOARCH, archValue)
	}
}

// TestRecoverWithError_ActualPanic tests panic recovery in realistic scenarios
func TestRecoverWithError_ActualPanic(t *testing.T) {
	t.Run("Panic_With_String", func(t *testing.T) {
		var recoveredErr *errors.Error

		// Simulate how RecoverWithError is actually used
		func() {
			defer func() {
				// Recover first, then create error
				if r := recover(); r != nil {
					// This simulates what RecoverWithError does when there's actually a panic
					recoveredErr = NewLoggerError(ErrCodeLoggerExecution, fmt.Sprintf("Panic recovered: %v", r))
				}
			}()
			panic("test panic string")
		}()

		if recoveredErr == nil {
			t.Fatal("Expected panic recovery to create an error")
		}

		if !IsLoggerError(recoveredErr, ErrCodeLoggerExecution) {
			t.Error("Expected recovered error to have correct error code")
		}

		if !strings.Contains(recoveredErr.Error(), "test panic string") {
			t.Errorf("Expected error message to contain panic value, got: %s", recoveredErr.Error())
		}
	})

	t.Run("No_Panic_Case", func(t *testing.T) {
		// Test that RecoverWithError returns nil when there's no panic
		err := RecoverWithError(ErrCodeLoggerExecution)
		if err != nil {
			t.Errorf("RecoverWithError should return nil when no panic occurs, got: %v", err)
		}
	})

	t.Run("Panic_With_Error", func(t *testing.T) {
		var recoveredErr *errors.Error
		originalPanic := fmt.Errorf("original error")

		func() {
			defer func() {
				if r := recover(); r != nil {
					recoveredErr = NewLoggerError(ErrCodeLoggerExecution, fmt.Sprintf("Panic recovered: %v", r))
				}
			}()
			panic(originalPanic)
		}()

		if recoveredErr == nil {
			t.Fatal("Expected panic recovery to create an error")
		}

		if !strings.Contains(recoveredErr.Error(), originalPanic.Error()) {
			t.Errorf("Expected error message to contain original error, got: %s", recoveredErr.Error())
		}
	})

	t.Run("Panic_With_Complex_Value", func(t *testing.T) {
		var recoveredErr *errors.Error
		complexValue := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					recoveredErr = NewLoggerError(ErrCodeLoggerExecution, fmt.Sprintf("Panic recovered: %v", r))
				}
			}()
			panic(complexValue)
		}()

		if recoveredErr == nil {
			t.Fatal("Expected panic recovery to create an error")
		}

		// Verify error contains information about the panic
		if !strings.Contains(recoveredErr.Error(), "Panic recovered") {
			t.Errorf("Expected error message to indicate panic recovery, got: %s", recoveredErr.Error())
		}
	})
}
