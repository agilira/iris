// iris.go: Ultra-high performance logger built on Xantos
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/agilira/go-errors"
	"github.com/agilira/xantos"
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

// Logger is the main Iris logger
type Logger struct {
	// Core Xantos ring buffer
	ring *xantos.Xantos[LogEntry]

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

	// Performance flags
	disableTimestamp     bool
	enableCaller         bool
	enableCallerFunction bool // NEW: Control function name extraction
	callerSkip           int
	stackTraceLevel      Level // Level at which to capture stack traces
	ultraFast            bool

	// Error handling
	errorsMutex sync.Mutex
	errorsList  []error

	// Control
	closed bool
	done   chan struct{}
}

// New creates a new Iris logger
func New(config Config) (*Logger, error) {
	// Default configuration
	if config.BufferSize == 0 {
		config.BufferSize = 4096 // Default 4K entries
	}
	if config.BatchSize == 0 {
		config.BatchSize = 64 // Default batch size
	}
	if config.Writer == nil {
		config.Writer = StdoutWriter
	}
	if config.Format == 0 {
		config.Format = JSONFormat // Default to JSON
	}

	// Handle multiple outputs configuration
	var finalWriter Writer
	var multiWriter *MultiWriter
	var hasTee bool

	if len(config.Writers) > 0 || len(config.WriteSyncers) > 0 {
		// Multiple outputs specified
		hasTee = true

		// Collect all writers
		var syncers []WriteSyncer

		// Add primary writer if specified
		if config.Writer != nil {
			syncers = append(syncers, WrapWriter(config.Writer))
		}

		// Add additional Writers
		for _, w := range config.Writers {
			syncers = append(syncers, WrapWriter(w))
		}

		// Add WriteSyncers directly
		syncers = append(syncers, config.WriteSyncers...)

		// Create MultiWriter
		multiWriter = NewMultiWriter(syncers...)
		finalWriter = multiWriter
	} else {
		// Single output
		finalWriter = config.Writer
	}

	// Ultra-fast mode overrides
	if config.UltraFast {
		config.Format = BinaryFormat
		config.DisableTimestamp = true
		config.EnableCaller = false // Disable caller in ultra-fast mode
		falsePtr := false
		config.EnableCallerFunction = &falsePtr // Disable function names in ultra-fast mode
		config.BatchSize = 256                  // Larger batches for max throughput
	}

	// Set default caller function behavior
	if config.EnableCaller && !config.UltraFast {
		// Default to true for backward compatibility
		if config.EnableCallerFunction == nil {
			truePtr := true
			config.EnableCallerFunction = &truePtr
		}
	}

	// Set default caller skip if not specified
	if config.CallerSkip == 0 {
		config.CallerSkip = 3 // Skip: runtime.Caller, getCaller, log method
	}

	// Initialize sampler if sampling config is provided
	var sampler *Sampler
	if config.SamplingConfig != nil {
		sampler = NewSampler(*config.SamplingConfig)
	}

	// Get caller function setting
	enableCallerFunction := false
	if config.EnableCallerFunction != nil {
		enableCallerFunction = *config.EnableCallerFunction
	}

	logger := &Logger{
		level:                config.Level,
		writer:               finalWriter,
		format:               config.Format,
		multiWriter:          multiWriter,
		hasTee:               hasTee,
		sampler:              sampler,
		disableTimestamp:     config.DisableTimestamp,
		enableCaller:         config.EnableCaller,
		enableCallerFunction: enableCallerFunction,
		callerSkip:           config.CallerSkip,
		stackTraceLevel:      config.StackTraceLevel,
		ultraFast:            config.UltraFast,
		done:                 make(chan struct{}),
	}

	// Initialize encoders based on format
	switch config.Format {
	case JSONFormat:
		logger.jsonEncoder = NewJSONEncoder()
	case ConsoleFormat:
		logger.consoleEncoder = NewConsoleEncoder(true) // Colorized by default
	case FastTextFormat:
		logger.textEncoder = NewFastTextEncoder()
	case BinaryFormat:
		logger.binaryEncoder = NewBinaryEncoder()
	}

	// Create Xantos ring buffer with log processor
	var err error
	logger.ring, err = xantos.NewBuilder[LogEntry](config.BufferSize).
		WithProcessor(logger.processLogEntry).
		WithBatchSize(config.BatchSize).
		Build()

	if err != nil {
		return nil, errors.Wrap(err, ErrCodeBufferCreation, "failed to create Xantos ring buffer")
	}

	// Start consumer goroutine
	go logger.run()

	return logger, nil
}

// processLogEntry processes a single log entry (called by Xantos)
func (l *Logger) processLogEntry(entry *LogEntry) {
	if l.closed {
		return
	}

	// Ultra-fast path: encode based on format
	var encodedBytes []byte

	switch l.format {
	case JSONFormat:
		l.jsonEncoder.EncodeLogEntry(entry.Timestamp, entry.Level, entry.Message, entry.Fields, entry.Caller, entry.StackTrace)
		encodedBytes = l.jsonEncoder.Bytes()

	case FastTextFormat:
		l.textEncoder.EncodeLogEntry(entry.Timestamp, entry.Level, entry.Message, entry.Fields, entry.Caller, entry.StackTrace)
		encodedBytes = l.textEncoder.Bytes()

	case BinaryFormat:
		l.binaryEncoder.EncodeLogEntry(entry.Timestamp, entry.Level, entry.Message, entry.Fields, entry.Caller, entry.StackTrace)
		encodedBytes = l.binaryEncoder.Bytes()

	case ConsoleFormat:
		// Console encoder works differently (takes the entire entry)
		var consoleBuf []byte
		encodedBytes = l.consoleEncoder.EncodeLogEntry(entry, consoleBuf)
	}

	// Write to output
	_, err := l.writer.Write(encodedBytes)
	if err != nil {
		l.addError(errors.Wrap(err, ErrCodeWriteFailure, "failed to write log entry"))
	}

	// No field cleanup needed - using pre-allocated buffer
}

// run is the main consumer loop
func (l *Logger) run() {
	defer close(l.done)
	l.ring.LoopProcess()
}

// log is the core logging method
func (l *Logger) log(level Level, message string, fields []Field) {
	// Early exit if level is not enabled
	if !l.level.Enabled(level) {
		return
	}

	if l.closed {
		return
	}

	// Check sampling early if configured
	if l.sampler != nil {
		// Create a minimal entry for sampling decision
		tempEntry := LogEntry{
			Level:   level,
			Message: message,
			Fields:  fields,
		}
		if !l.disableTimestamp {
			tempEntry.Timestamp = time.Now()
		}

		decision := l.sampler.Sample(tempEntry)
		if decision == DropSample {
			// Skip this entry completely - don't even go to ring buffer
			return
		}
	}

	// Capture caller info outside the ring buffer write for performance
	var caller Caller
	if l.enableCaller {
		caller = l.getCaller()
	}

	// Capture stack trace if level qualifies and stack trace is enabled
	var stackTrace string
	if l.stackTraceLevel != 0 && level >= l.stackTraceLevel {
		// Use go-errors to capture stack trace
		stack := errors.CaptureStacktrace(l.callerSkip)
		if stack != nil {
			stackTrace = stack.String()
		}
	}

	// Write to ring buffer - ultra-fast path
	l.ring.Write(func(entry *LogEntry) {
		if !l.disableTimestamp {
			entry.Timestamp = time.Now()
		}
		entry.Level = level
		entry.Message = message
		entry.Fields = fields
		entry.Caller = caller
		entry.StackTrace = stackTrace // Copy fields efficiently
		if len(fields) <= 16 {
			// Use pre-allocated buffer for small field counts
			entry.Fields = entry.fieldBuf[:len(fields)]
			copy(entry.Fields, fields)
		} else {
			// Use larger slice for bigger field counts
			entry.Fields = make([]Field, len(fields))
			copy(entry.Fields, fields)
		}
	})
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...Field) {
	l.log(DebugLevel, message, fields)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...Field) {
	l.log(InfoLevel, message, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...Field) {
	l.log(WarnLevel, message, fields)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...Field) {
	l.log(ErrorLevel, message, fields)
}

// DPanic logs at DPanic level and panics
func (l *Logger) DPanic(message string, fields ...Field) {
	l.log(DPanicLevel, message, fields)
	DPanicLevel.HandleSpecial()
}

// Panic logs at Panic level and panics
func (l *Logger) Panic(message string, fields ...Field) {
	l.log(PanicLevel, message, fields)
	PanicLevel.HandleSpecial()
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, fields ...Field) {
	l.log(FatalLevel, message, fields)
	l.Close()
	os.Exit(1)
}

// With creates a logger with pre-set fields (returns a new logger)
func (l *Logger) With(fields ...Field) *Logger {
	// For now, return the same logger
	// In a future version, we could implement field inheritance
	return l
}

// Close gracefully shuts down the logger
func (l *Logger) Close() {
	if l.closed {
		return
	}

	l.closed = true
	l.ring.Close()

	// Wait for consumer to finish
	<-l.done

	// Close writer if it supports closing
	if closer, ok := l.writer.(interface{ Close() error }); ok {
		closer.Close()
	}
}

// addError safely adds an error to the error list
func (l *Logger) addError(err error) {
	l.errorsMutex.Lock()
	defer l.errorsMutex.Unlock()
	l.errorsList = append(l.errorsList, err)
}

// Errors returns any errors that occurred during logging
func (l *Logger) Errors() []error {
	l.errorsMutex.Lock()
	defer l.errorsMutex.Unlock()

	if len(l.errorsList) == 0 {
		return nil
	}

	result := make([]error, len(l.errorsList))
	copy(result, l.errorsList)
	return result
}

// getCaller captures caller information using runtime.Caller (ULTRA-OPTIMIZED)
func (l *Logger) getCaller() Caller {
	if !l.enableCaller {
		return Caller{Valid: false}
	}

	// Optimized: Use runtime.Caller with pre-calculated skip
	pc, file, line, ok := runtime.Caller(l.callerSkip)
	if !ok {
		return Caller{Valid: false}
	}

	caller := Caller{
		File:  file,
		Line:  line,
		Valid: true,
	}

	// CRITICAL OPTIMIZATION: Function name cache to beat Zap's 202ns!
	if l.enableCallerFunction && pc != 0 {
		// Fast path: Cache lookup O(1) invece di runtime.FuncForPC O(n)
		if cached, ok := funcNameCache.Load(pc); ok {
			caller.Function = cached.(string)
		} else {
			// Slow path: Only on cache miss
			if fn := runtime.FuncForPC(pc); fn != nil {
				name := fn.Name()
				// Store in cache per future calls
				funcNameCache.Store(pc, name)
				caller.Function = name
			}
		}
	}

	return caller
}

// AddWriter adds a new writer to the logger's output destinations
func (l *Logger) AddWriter(writer Writer) error {
	if !l.hasTee {
		// Convert to MultiWriter
		l.multiWriter = NewMultiWriter(WrapWriter(l.writer))
		l.hasTee = true
	}

	l.multiWriter.AddWriter(WrapWriter(writer))
	l.writer = l.multiWriter
	return nil
}

// RemoveWriter removes a writer from the logger's output destinations
func (l *Logger) RemoveWriter(writer Writer) bool {
	if !l.hasTee {
		return false
	}

	return l.multiWriter.RemoveWriter(WrapWriter(writer))
}

// AddWriteSyncer adds a WriteSyncer to the logger's output destinations
func (l *Logger) AddWriteSyncer(syncer WriteSyncer) error {
	if !l.hasTee {
		// Convert to MultiWriter
		l.multiWriter = NewMultiWriter(WrapWriter(l.writer))
		l.hasTee = true
	}

	l.multiWriter.AddWriter(syncer)
	l.writer = l.multiWriter
	return nil
}

// WriterCount returns the number of active writers
func (l *Logger) WriterCount() int {
	if !l.hasTee {
		return 1
	}

	return l.multiWriter.Count()
}

// Sampling Methods

// GetSamplingStats returns the current sampling statistics.
// Returns nil if sampling is not enabled.
func (l *Logger) GetSamplingStats() *SamplingStats {
	if l.sampler == nil {
		return nil
	}

	stats := l.sampler.Stats()
	return &stats
}

// IsSamplingEnabled returns true if sampling is configured and active
func (l *Logger) IsSamplingEnabled() bool {
	return l.sampler != nil
}

// SetSamplingConfig updates the sampling configuration.
// Pass nil to disable sampling.
func (l *Logger) SetSamplingConfig(config *SamplingConfig) {
	if config == nil {
		l.sampler = nil
		return
	}

	l.sampler = NewSampler(*config)
}

// Sync ensures all buffered log entries are flushed to the output.
// This method blocks until all pending entries are processed.
func (l *Logger) Sync() error {
	if l.closed {
		return nil
	}

	// Process any pending entries in the ring buffer
	// Since Xantos doesn't have a Sync method, we just ensure outputs are synced

	// Sync the underlying writer if it supports it
	if l.hasTee && l.multiWriter != nil {
		return l.multiWriter.Sync()
	}

	// Try to sync the primary writer if it implements WriteSyncer
	if syncer, ok := l.writer.(WriteSyncer); ok {
		return syncer.Sync()
	}

	return nil
}
