// iris.go: Ultra-high performance logger built on Xantos
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/agilira/zephyros"
)

// Logger is the main Iris logger
type Logger struct {
	// Core Zephyros MPSC ring buffer
	ring *zephyros.Zephyros[LogEntry]

	// Configuration
	level  Level
	writer Writer
	format Format

	// Multiple output support
	multiWriter *MultiWriter // For multiple output destinations
	hasTee      bool         // True if using multiple outputs

	// Sampling support
	sampler *Sampler // For log sampling, nil if sampling disabled

	// Encoders (only one will be used based on format)
	jsonEncoder    *JSONEncoder
	consoleEncoder *ConsoleEncoder
	textEncoder    *FastTextEncoder
	binaryEncoder  *BinaryEncoder

	// ULTRA-OPTIMIZATION: Pre-computed function pointers for zero method dispatch overhead
	// This eliminates virtual method calls and branch prediction penalties
	encodeFunc func(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) []byte

	// GENIUS OPTIMIZATION: Pre-bound fields for With() method
	preFields []Field // Fields that are always included

	// Performance flags
	disableTimestamp     bool
	enableCaller         bool
	enableCallerFunction bool // NEW: Control function name extraction
	callerSkip           int
	stackTraceLevel      Level // Level at which to capture stack traces
	ultraFast            bool

	// Error handling (MPSC-safe with atomic operations)
	closed int32 // Use atomic operations for thread safety
	done   chan struct{}
}

// processLogEntry processes a single log entry (called by Zephyros) - ULTRA-OPTIMIZED
func (l *Logger) processLogEntry(entry *LogEntry) {
	if atomic.LoadInt32(&l.closed) != 0 {
		return
	}

	// ULTRA-CRITICAL OPTIMIZATION: Direct function call eliminates switch overhead
	// Pre-computed function pointers eliminate method dispatch and branch misprediction
	// This is the HOTTEST path in the entire logger - every nanosecond counts
	encodedBytes := l.encodeFunc(entry.Timestamp, entry.Level, entry.Message, entry.Fields, entry.Caller, entry.StackTrace)

	// CRITICAL HOT PATH: Direct writer call eliminates method dispatch
	_, err := l.writer.Write(encodedBytes)
	if err != nil {
		// Ultra-fast error path: Pre-allocated error message
		const errMsg = "IRIS: failed to write log entry\n"
		_, _ = os.Stderr.WriteString(errMsg)
	}

	// GENIUS OPTIMIZATION: Zero entry to prevent memory leaks (from prototipo)
	// This releases references to large backing arrays and strings
	zeroLogEntry(entry)
}

// zeroLogEntry releases references to large backing arrays and strings (MEMORY OPTIMIZATION)
// This prevents memory leaks by ensuring large slices don't retain references
func zeroLogEntry(e *LogEntry) {
	// Clear string references to allow GC
	e.Message = ""
	e.StackTrace = ""

	// INTELLIGENT FIELD CLEANUP: Only clear large allocations
	if cap(e.Fields) > len(e.fieldBuf) {
		// Large slice was allocated - drop reference to allow GC
		e.Fields = nil
	} else {
		// Using pre-allocated buffer - just reset length
		e.Fields = e.fieldBuf[:0]
	}

	// Clear caller info
	e.Caller = Caller{}
}

// UltraFast returns true if logger is in ultra-fast mode (SMART AUTO-DETECTION)
// Based on configuration, automatically detects optimal performance mode
func (l *Logger) UltraFast() bool {
	// Ultra-fast when: not development mode AND using fast formats
	return !l.enableCaller &&
		l.disableTimestamp &&
		(l.format == BinaryFormat || l.format == JSONFormat)
}

// run is the main consumer loop
func (l *Logger) run() {
	defer close(l.done)
	l.ring.LoopProcess()
}
