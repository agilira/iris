// logger_factory.go: Factory functions and configuration builders for Iris logger
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"time"

	"github.com/agilira/go-errors"
	"github.com/agilira/zephyros"
)

// New creates a new Iris logger
func New(config Config) (*Logger, error) {
	// Apply default configuration
	config = applyDefaults(config)
	
	// Configure writers and outputs
	finalWriter, multiWriter, hasTee := configureWriters(config)
	
	// Apply ultra-fast mode optimizations
	config = applyUltraFastMode(config)
	
	// Configure caller settings
	config = configureCallerSettings(config)
	
	// Initialize sampler if needed
	sampler := initializeSampler(config)
	
	// Create base logger
	logger := createBaseLogger(config, finalWriter, multiWriter, hasTee, sampler)
	
	// Initialize encoders and functions
	if err := initializeEncoders(logger, config); err != nil {
		return nil, err
	}
	
	// Create Zephyros MPSC ring buffer
	if err := initializeRingBuffer(logger, config); err != nil {
		return nil, err
	}
	
	// Start consumer goroutine
	go logger.run()
	
	return logger, nil
}

// applyDefaults sets default configuration values
func applyDefaults(config Config) Config {
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
	return config
}

// configureWriters sets up the writer configuration
func configureWriters(config Config) (Writer, *MultiWriter, bool) {
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

	return finalWriter, multiWriter, hasTee
}

// applyUltraFastMode applies ultra-fast mode optimizations
func applyUltraFastMode(config Config) Config {
	if config.UltraFast {
		config.Format = BinaryFormat
		config.DisableTimestamp = true
		config.EnableCaller = false // Disable caller in ultra-fast mode
		falsePtr := false
		config.EnableCallerFunction = &falsePtr // Disable function names in ultra-fast mode
		config.BatchSize = 256                  // Larger batches for max throughput
	}
	return config
}

// configureCallerSettings sets up caller-related configuration
func configureCallerSettings(config Config) Config {
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

	return config
}

// initializeSampler creates sampler if needed
func initializeSampler(config Config) *Sampler {
	if config.SamplingConfig != nil {
		return NewSampler(*config.SamplingConfig)
	}
	return nil
}

// createBaseLogger creates the base logger instance
func createBaseLogger(config Config, finalWriter Writer, multiWriter *MultiWriter, hasTee bool, sampler *Sampler) *Logger {
	// Get caller function setting
	enableCallerFunction := false
	if config.EnableCallerFunction != nil {
		enableCallerFunction = *config.EnableCallerFunction
	}

	return &Logger{
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
}

// initializeEncoders sets up encoders based on format
func initializeEncoders(logger *Logger, config Config) error {
	switch config.Format {
	case JSONFormat:
		logger.jsonEncoder = NewJSONEncoder()
		// ULTRA-OPTIMIZATION: Pre-compute function pointer to eliminate method dispatch
		logger.encodeFunc = func(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) []byte {
			logger.jsonEncoder.EncodeLogEntry(timestamp, level, message, fields, caller, stackTrace)
			return logger.jsonEncoder.Bytes()
		}
	case ConsoleFormat:
		logger.consoleEncoder = NewConsoleEncoder(true) // Colorized by default
		// Console encoder function pointer (different signature)
		logger.encodeFunc = func(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) []byte {
			entry := &LogEntry{Timestamp: timestamp, Level: level, Message: message, Fields: fields, Caller: caller, StackTrace: stackTrace}
			var consoleBuf []byte
			return logger.consoleEncoder.EncodeLogEntry(entry, consoleBuf)
		}
	case FastTextFormat:
		logger.textEncoder = NewFastTextEncoder()
		logger.encodeFunc = func(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) []byte {
			// MIGRATION: Use direct Field->BinaryField conversion in encoder
			logger.textEncoder.EncodeLogEntryMigration(timestamp, level, message, fields, caller, stackTrace)
			return logger.textEncoder.Bytes()
		}
	case BinaryFormat:
		logger.binaryEncoder = NewBinaryEncoder()
		logger.encodeFunc = func(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string) []byte {
			// MIGRATION: Use direct Field->BinaryField conversion in encoder
			logger.binaryEncoder.EncodeLogEntryMigration(timestamp, level, message, fields, caller, stackTrace)
			return logger.binaryEncoder.Bytes()
		}
	}
	return nil
}

// initializeRingBuffer creates and configures the Zephyros ring buffer
func initializeRingBuffer(logger *Logger, config Config) error {
	var err error
	logger.ring, err = zephyros.NewBuilder[LogEntry](config.BufferSize).
		WithProcessor(logger.processLogEntry).
		WithBatchSize(config.BatchSize).
		Build()

	if err != nil {
		return errors.Wrap(err, ErrCodeBufferCreation, "failed to create Zephyros MPSC ring buffer")
	}
	return nil
}
