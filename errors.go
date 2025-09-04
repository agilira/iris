// errors.go: Error handling integration for Iris logging library
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/agilira/go-errors"
)

// LoggerError codes - specific error codes for the iris logging library
const (
	// Core logging errors
	ErrCodeLoggerCreation errors.ErrorCode = "IRIS_LOGGER_CREATION"
	ErrCodeLoggerNotFound errors.ErrorCode = "IRIS_LOGGER_NOT_FOUND"
	ErrCodeLoggerDisabled errors.ErrorCode = "IRIS_LOGGER_DISABLED"
	ErrCodeLoggerClosed   errors.ErrorCode = "IRIS_LOGGER_CLOSED"

	// Configuration errors
	ErrCodeInvalidConfig errors.ErrorCode = "IRIS_INVALID_CONFIG"
	ErrCodeInvalidLevel  errors.ErrorCode = "IRIS_INVALID_LEVEL"
	ErrCodeInvalidFormat errors.ErrorCode = "IRIS_INVALID_FORMAT"
	ErrCodeInvalidOutput errors.ErrorCode = "IRIS_INVALID_OUTPUT"

	// Field and encoding errors
	ErrCodeInvalidField      errors.ErrorCode = "IRIS_INVALID_FIELD"
	ErrCodeEncodingFailed    errors.ErrorCode = "IRIS_ENCODING_FAILED"
	ErrCodeFieldTypeMismatch errors.ErrorCode = "IRIS_FIELD_TYPE_MISMATCH"
	ErrCodeBufferOverflow    errors.ErrorCode = "IRIS_BUFFER_OVERFLOW"

	// Writer and output errors
	ErrCodeWriterNotAvailable errors.ErrorCode = "IRIS_WRITER_NOT_AVAILABLE"
	ErrCodeWriteFailed        errors.ErrorCode = "IRIS_WRITE_FAILED"
	ErrCodeFlushFailed        errors.ErrorCode = "IRIS_FLUSH_FAILED"
	ErrCodeSyncFailed         errors.ErrorCode = "IRIS_SYNC_FAILED"

	// Performance and resource errors
	ErrCodeMemoryAllocation errors.ErrorCode = "IRIS_MEMORY_ALLOCATION"
	ErrCodePoolExhausted    errors.ErrorCode = "IRIS_POOL_EXHAUSTED"
	ErrCodeTimeout          errors.ErrorCode = "IRIS_TIMEOUT"
	ErrCodeResourceLimit    errors.ErrorCode = "IRIS_RESOURCE_LIMIT"

	// Ring buffer errors
	ErrCodeRingInvalidCapacity  errors.ErrorCode = "IRIS_RING_INVALID_CAPACITY"
	ErrCodeRingInvalidBatchSize errors.ErrorCode = "IRIS_RING_INVALID_BATCH_SIZE"
	ErrCodeRingMissingProcessor errors.ErrorCode = "IRIS_RING_MISSING_PROCESSOR"
	ErrCodeRingClosed           errors.ErrorCode = "IRIS_RING_CLOSED"
	ErrCodeRingBuildFailed      errors.ErrorCode = "IRIS_RING_BUILD_FAILED"

	// Hook and middleware errors
	ErrCodeHookExecution   errors.ErrorCode = "IRIS_HOOK_EXECUTION"
	ErrCodeMiddlewareChain errors.ErrorCode = "IRIS_MIDDLEWARE_CHAIN"
	ErrCodeFilterFailed    errors.ErrorCode = "IRIS_FILTER_FAILED"

	// File and rotation errors
	ErrCodeFileOpen         errors.ErrorCode = "IRIS_FILE_OPEN"
	ErrCodeFileWrite        errors.ErrorCode = "IRIS_FILE_WRITE"
	ErrCodeFileRotation     errors.ErrorCode = "IRIS_FILE_ROTATION"
	ErrCodePermissionDenied errors.ErrorCode = "IRIS_PERMISSION_DENIED"
)

// ErrorHandler represents a function that handles errors within the logging system
type ErrorHandler func(err *errors.Error)

// defaultErrorHandler is the default error handler that prints to stderr
var defaultErrorHandler ErrorHandler = func(err *errors.Error) {
	// Avoid infinite recursion by not using the logger to log errors
	fmt.Fprintf(os.Stderr, "[IRIS ERROR] %s: %s\n", err.Code, err.Message)
	if err.Cause != nil {
		fmt.Fprintf(os.Stderr, "[IRIS ERROR] Caused by: %v\n", err.Cause)
	}
}

// currentErrorHandler holds the current error handler
var currentErrorHandler = defaultErrorHandler

// SetErrorHandler sets a custom error handler for the iris logging system
// This allows applications to customize how logging errors are handled
func SetErrorHandler(handler ErrorHandler) {
	if handler == nil {
		currentErrorHandler = defaultErrorHandler
		return
	}
	currentErrorHandler = handler
}

// GetErrorHandler returns the current error handler
func GetErrorHandler() ErrorHandler {
	return currentErrorHandler
}

// handleError processes an error using the current error handler
func handleError(err *errors.Error) {
	if err == nil {
		return
	}

	// Add runtime context if not already present
	if err.Context == nil {
		err.Context = make(map[string]interface{})
	}

	// Add system information for debugging
	err.Context["go_version"] = runtime.Version()
	err.Context["goroutines"] = runtime.NumGoroutine()

	// Call the error handler
	currentErrorHandler(err)
}

// NewLoggerError creates a new logger-specific error with standard context
func NewLoggerError(code errors.ErrorCode, message string) *errors.Error {
	err := errors.New(code, message).
		WithSeverity("error").
		WithContext("component", "iris_logger").
		WithContext("timestamp", time.Now().UTC())

	// Add caller information for debugging
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			_ = err.WithContext("caller_func", fn.Name())
		}
		_ = err.WithContext("caller_file", file)
		_ = err.WithContext("caller_line", line)
	}

	return err
}

// NewLoggerErrorWithField creates a logger error with field and value information
func NewLoggerErrorWithField(code errors.ErrorCode, message, field, value string) *errors.Error {
	err := errors.NewWithField(code, message, field, value).
		WithSeverity("error").
		WithContext("component", "iris_logger").
		WithContext("timestamp", time.Now().UTC())

	return err
}

// WrapLoggerError wraps an existing error with logger-specific context
func WrapLoggerError(originalErr error, code errors.ErrorCode, message string) *errors.Error {
	err := errors.Wrap(originalErr, code, message).
		WithSeverity("error").
		WithContext("component", "iris_logger").
		WithContext("timestamp", time.Now().UTC())

	// Add caller information
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			_ = err.WithContext("caller_func", fn.Name())
		}
		_ = err.WithContext("caller_file", file)
		_ = err.WithContext("caller_line", line)
	}

	return err
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if irisErr, ok := err.(*errors.Error); ok {
		return irisErr.IsRetryable()
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) errors.ErrorCode {
	if irisErr, ok := err.(*errors.Error); ok {
		return irisErr.ErrorCode()
	}
	return ""
}

// GetUserMessage extracts a user-friendly message from an error
func GetUserMessage(err error) string {
	if irisErr, ok := err.(*errors.Error); ok {
		return irisErr.UserMessage()
	}
	return err.Error()
}

// IsLoggerError checks if an error is an iris logger error
func IsLoggerError(err error, code errors.ErrorCode) bool {
	return errors.HasCode(err, code)
}

// RecoverWithError recovers from a panic and converts it to a logger error
func RecoverWithError(code errors.ErrorCode) *errors.Error {
	if r := recover(); r != nil {
		message := fmt.Sprintf("Panic recovered: %v", r)

		// Create error with stack trace
		err := NewLoggerError(code, message)

		// Add panic information to context
		_ = err.WithContext("panic_value", r)
		_ = err.WithContext("recovery_time", time.Now().UTC())

		// Add stack trace from panic point
		buf := make([]byte, 4096)
		stackSize := runtime.Stack(buf, false)
		_ = err.WithContext("panic_stack", string(buf[:stackSize]))

		return err
	}
	return nil
}

// SafeExecute executes a function safely, handling any panics
func SafeExecute(fn func() error, operation string) error {
	defer func() {
		if err := RecoverWithError(ErrCodeLoggerExecution); err != nil {
			_ = err.WithContext("operation", operation)
			handleError(err)
		}
	}()

	return fn()
}

// ErrCodeLoggerExecution represents the error code for logger execution failures
const ErrCodeLoggerExecution errors.ErrorCode = "IRIS_LOGGER_EXECUTION"

// validateErrorCodes ensures all error codes follow naming conventions
func validateErrorCodes() {
	// This function can be used during development to ensure error codes
	// follow the IRIS_ prefix convention and are properly categorized

	codes := []errors.ErrorCode{
		ErrCodeLoggerCreation, ErrCodeLoggerNotFound, ErrCodeLoggerDisabled,
		ErrCodeLoggerClosed, ErrCodeInvalidConfig, ErrCodeInvalidLevel,
		ErrCodeInvalidFormat, ErrCodeInvalidOutput, ErrCodeInvalidField,
		ErrCodeEncodingFailed, ErrCodeFieldTypeMismatch, ErrCodeBufferOverflow,
		ErrCodeWriterNotAvailable, ErrCodeWriteFailed, ErrCodeFlushFailed,
		ErrCodeSyncFailed, ErrCodeMemoryAllocation, ErrCodePoolExhausted,
		ErrCodeTimeout, ErrCodeResourceLimit, ErrCodeRingInvalidCapacity,
		ErrCodeRingInvalidBatchSize, ErrCodeRingMissingProcessor, ErrCodeRingClosed,
		ErrCodeRingBuildFailed, ErrCodeHookExecution,
		ErrCodeMiddlewareChain, ErrCodeFilterFailed, ErrCodeFileOpen,
		ErrCodeFileWrite, ErrCodeFileRotation, ErrCodePermissionDenied,
		ErrCodeLoggerExecution,
	}

	for _, code := range codes {
		if len(string(code)) == 0 {
			panic("Empty error code detected")
		}
		if string(code)[:5] != "IRIS_" {
			panic(fmt.Sprintf("Error code %s does not follow IRIS_ prefix convention", code))
		}
	}
}

// init validates error codes during package initialization
func init() {
	validateErrorCodes()
}
