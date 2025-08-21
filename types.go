// logger_types.go: Core types and data structures for Iris logger
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"sync"
	"time"

	"github.com/agilira/go-errors"
)

// Function name cache per ottimizzare caller info performance
// Usa sync.Map per accesso concorrente O(1) vs runtime.FuncForPC O(n)
var funcNameCache sync.Map // map[uintptr]string

// Error codes for Iris
const (
	ErrCodeBufferCreation errors.ErrorCode = "IRIS_BUFFER_CREATION"
	ErrCodeWriteFailure   errors.ErrorCode = "IRIS_WRITE_FAILURE"
)

// Caller represents caller information (file, line, function)
type Caller struct {
	File     string
	Line     int
	Function string
	Valid    bool // True if caller info was successfully captured
}

// LogEntry represents a single log entry in the ring buffer
type LogEntry struct {
	Timestamp  time.Time
	Level      Level
	Message    string
	Fields     []Field
	Caller     Caller // Caller information
	StackTrace string // Stack trace (only when enabled and level qualifies)
	// Pre-allocated field slice to avoid allocations
	fieldBuf [16]Field // Most log entries have < 16 fields
}

// Config holds the configuration for the Iris logger
type Config struct {
	// Level sets the minimum log level
	Level Level

	// Writer is the primary output destination
	Writer io.Writer

	// Writers is a slice of additional output destinations for multiple outputs
	Writers []io.Writer

	// WriteSyncers allows mixing Writer/Writers with WriteSyncer interfaces
	WriteSyncers []WriteSyncer

	// Development enables development mode with enhanced readability
	Development bool

	// EnableCaller adds caller information (file, line, function) to logs
	EnableCaller bool

	// EnableCallerFunction controls whether to include function names in caller info.
	// When false, only file and line are captured (faster).
	// Only used when EnableCaller is true.
	// If nil, defaults to true for backward compatibility.
	EnableCallerFunction *bool

	// CallerSkip is the number of stack frames to skip when determining caller info.
	// This is useful when the logger is wrapped by other functions.
	// Default is 0, meaning the immediate caller is used.
	CallerSkip int

	// SamplingConfig enables log sampling to reduce high-volume output.
	// If nil, no sampling is applied and all logs are written.
	SamplingConfig *SamplingConfig

	// StackTraceLevel enables stack trace capture for logs at or above this level.
	// Set to a level to enable stack traces, or use a level higher than Fatal to disable.
	// Example: StackTraceLevel: ErrorLevel (captures stack traces for Error, DPanic, Panic, Fatal)
	StackTraceLevel Level

	// Legacy fields for compatibility
	Format     Format // Output format
	BufferSize int64  // Ring buffer size
	BatchSize  int64  // Batch processing size

	// Performance optimizations
	DisableTimestamp bool // Skip timestamp for max speed
	UltraFast        bool // Enable all speed optimizations
}
