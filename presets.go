// presets.go: Configuration presets for common use cases
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync"
)

// Cached preset configurations to avoid repeated allocations
var (
	developmentConfigCache Config
	productionConfigCache  Config
	exampleConfigCache     Config
	ultraFastConfigCache   Config
	fastTextConfigCache    Config
	devStackTraceCache     Config
	debugStackTraceCache   Config

	configCacheOnce sync.Once
)

// initializeConfigCache initializes all preset configurations once
func initializeConfigCache() {
	developmentConfigCache = Config{
		Level:      DebugLevel,
		Writer:     StdoutWriter,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	productionConfigCache = Config{
		Level:      InfoLevel,
		Writer:     StdoutWriter,
		Format:     JSONFormat,
		BufferSize: 8192,
		BatchSize:  128,
	}

	exampleConfigCache = Config{
		Level:            InfoLevel,
		Writer:           StdoutWriter,
		Format:           JSONFormat,
		BufferSize:       512,
		BatchSize:        16,
		DisableTimestamp: true,
	}

	ultraFastConfigCache = Config{
		Level:            InfoLevel,
		Writer:           StderrWriteSyncer,
		Format:           BinaryFormat,
		BufferSize:       16384,
		BatchSize:        256,
		DisableTimestamp: true,
		EnableCaller:     false,
		UltraFast:        true,
	}

	fastTextConfigCache = Config{
		Level:      DebugLevel,
		Writer:     StdoutWriter,
		Format:     FastTextFormat,
		BufferSize: 2048,
		BatchSize:  64,
	}

	devStackTraceCache = Config{
		Level:           DebugLevel,
		Writer:          StdoutWriter,
		Format:          ConsoleFormat,
		BufferSize:      1024,
		BatchSize:       32,
		StackTraceLevel: ErrorLevel,
	}

	debugStackTraceCache = Config{
		Level:           DebugLevel,
		Writer:          StdoutWriter,
		Format:          ConsoleFormat,
		BufferSize:      1024,
		BatchSize:       16,
		StackTraceLevel: WarnLevel,
		EnableCaller:    true,
	}
}

// NewDevelopment creates a logger suitable for development
// Features: console output with colors, debug level, human-readable format
func NewDevelopment() (*Logger, error) {
	configCacheOnce.Do(initializeConfigCache)
	return New(developmentConfigCache)
}

// NewProduction creates a logger suitable for production
// Features: JSON output, info level, optimized for performance
func NewProduction() (*Logger, error) {
	configCacheOnce.Do(initializeConfigCache)
	return New(productionConfigCache)
}

// NewExample creates a logger suitable for examples and testing
// Features: deterministic output, no timestamps, consistent formatting
func NewExample() (*Logger, error) {
	configCacheOnce.Do(initializeConfigCache)
	return New(exampleConfigCache)
}

// NewUltraFast creates a logger optimized for maximum performance
// Features: binary format, minimal allocations, maximum throughput
func NewUltraFast() (*Logger, error) {
	configCacheOnce.Do(initializeConfigCache)
	return New(ultraFastConfigCache)
}

// NewFastText creates a logger with fast text output
// Features: human-readable text, fast encoding, good for development
func NewFastText() (*Logger, error) {
	configCacheOnce.Do(initializeConfigCache)
	return New(fastTextConfigCache)
}

// NewDevelopmentWithStackTrace creates a development logger with stack trace support
// Features: console output with colors, debug level, stack traces for errors
func NewDevelopmentWithStackTrace() (*Logger, error) {
	configCacheOnce.Do(initializeConfigCache)
	return New(devStackTraceCache)
}

// NewDebugWithStackTrace creates a debug logger with comprehensive stack trace support
// Features: all levels logged, stack traces for warnings and above
func NewDebugWithStackTrace() (*Logger, error) {
	configCacheOnce.Do(initializeConfigCache)
	return New(debugStackTraceCache)
}

// DevelopmentConfig returns a development configuration
// Use this if you want to customize the development preset
func DevelopmentConfig() Config {
	configCacheOnce.Do(initializeConfigCache)
	return developmentConfigCache
}

// ProductionConfig returns a production configuration
// Use this if you want to customize the production preset
func ProductionConfig() Config {
	configCacheOnce.Do(initializeConfigCache)
	return productionConfigCache
}

// ExampleConfig returns an example configuration
// Use this if you want to customize the example preset
func ExampleConfig() Config {
	configCacheOnce.Do(initializeConfigCache)
	return exampleConfigCache
}

// DevelopmentWithStackTraceConfig returns a development configuration with stack traces
// Use this if you want to customize the development with stack trace preset
func DevelopmentWithStackTraceConfig() Config {
	configCacheOnce.Do(initializeConfigCache)
	return devStackTraceCache
}

// DebugWithStackTraceConfig returns a debug configuration with comprehensive stack traces
// Use this if you want to customize the debug with stack trace preset
func DebugWithStackTraceConfig() Config {
	configCacheOnce.Do(initializeConfigCache)
	return debugStackTraceCache
}

// NewUltraFastFile creates a high-performance binary logger for file output
// Features: binary format to file, maximum performance, production-ready
func NewUltraFastFile(filePath string) (*Logger, error) {
	fileWriter, err := NewFileWriter(filePath)
	if err != nil {
		return nil, err
	}

	config := ultraFastConfigCache
	config.Writer = fileWriter
	config.DisableTimestamp = false // Keep timestamp for production files

	return New(config)
}

// NewUltraFastNetwork creates a high-performance binary logger for network output
// Features: binary format to network, maximum throughput, production-ready
func NewUltraFastNetwork(writer WriteSyncer) (*Logger, error) {
	config := ultraFastConfigCache
	config.Writer = writer
	config.BufferSize = 32768       // Extra large buffer for network
	config.BatchSize = 512          // Large batches for network efficiency
	config.DisableTimestamp = false // Keep timestamp for production

	return New(config)
}
