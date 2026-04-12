// magic.go: Magic API for seamless Iris + Lethe integration
//
// This file provides the seamless automatic integration when both Iris and
// Lethe are imported together. Zero configuration, maximum performance.
//
// Usage (H1 style - zero config):
//
//	import (
//	    "github.com/agilira/iris"
//	    _ "github.com/agilira/lethe" // Import for side effects
//	)
//
//	// Just works!
//	logger := iris.Magic("app.log")
//
// Usage (with options):
//
//	logger := iris.Magic("app.log", iris.MagicLevel(iris.Debug))
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agilira/iris/lethebridge"
)

// MagicOption configures the Magic logger
type MagicOption func(*magicConfig)

type magicConfig struct {
	level Level
}

// MagicLevel sets the log level for the Magic logger
func MagicLevel(level Level) MagicOption {
	return func(c *magicConfig) {
		c.level = level
	}
}

// Magic creates a production-ready logger with zero configuration.
// This is the H1 chip pattern - import both Iris and Lethe, and it just works.
//
// When Lethe is imported:
//   - Automatic log rotation (100MB default)
//   - Compression enabled
//   - Zero-copy optimization
//   - Hot-reload ready
//
// When Lethe is not imported:
//   - Standard file logging
//   - Same API, graceful fallback
//
// Panics on error (use NewMagicLogger for error handling).
func Magic(filename string, opts ...MagicOption) *Logger {
	cfg := &magicConfig{
		level: Info, // Production default
	}
	for _, opt := range opts {
		opt(cfg)
	}

	logger, err := NewMagicLogger(filename, cfg.level)
	if err != nil {
		panic(fmt.Sprintf("iris.Magic: failed to create logger: %v", err))
	}

	logger.Start()
	return logger
}

// NewMagicLogger creates a logger with automatic Lethe optimization when available.
// This is the Magic API that provides seamless integration between Iris and Lethe.
//
// When Lethe is imported:
//   - Automatic zero-copy optimization via WriteOwned()
//   - Intelligent buffer sizing based on Lethe's recommendations
//   - Hot-reload configuration support
//   - Advanced rotation with compression
//
// When Lethe is not available:
//   - Graceful fallback to standard file logging
//   - Same API, no configuration changes needed
//
// Parameters:
//   - filename: Path to log file (will be created if needed)
//   - level: Minimum log level
//   - opts: Optional Iris configuration overrides
//
// Returns a fully configured Logger ready for high-performance logging.
func NewMagicLogger(filename string, level Level, opts ...Option) (*Logger, error) {
	// Check if Lethe capabilities are registered (runtime detection)
	if lethebridge.HasLetheCapabilities() {
		return createMagicLetheLogger(filename, level, opts...)
	}

	// Fallback to standard file logger
	return createStandardFileLogger(filename, level, opts...)
}

// createMagicLetheLogger creates an optimized logger using Lethe capabilities
func createMagicLetheLogger(filename string, level Level, opts ...Option) (*Logger, error) {
	provider, exists := lethebridge.GetLetheProvider()
	if !exists {
		return createStandardFileLogger(filename, level, opts...)
	}

	// Create optimized Lethe sink with smart defaults
	sink, err := provider.CreateOptimizedSink(filename,
		"maxSize", "100MB",
		"maxBackups", 5,
		"compress", true,
		"hotReload", true,
	)
	if err != nil {
		return createStandardFileLogger(filename, level, opts...)
	}

	// Detect if sink supports Lethe optimizations
	letheWriter := lethebridge.DetectLetheCapabilities(sink)
	if letheWriter == nil {
		return createStandardFileLogger(filename, level, opts...)
	}

	// Configure Iris with Lethe optimizations
	cfg := Config{
		Level:   level,
		Output:  letheWriter,
		Encoder: NewJSONEncoder(),
		// Use Lethe's recommended buffer size for optimal performance
		Capacity: int64(letheWriter.GetOptimalBufferSize()),
	}

	logger, err := New(cfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create magic logger: %w", err)
	}

	return logger, nil
}

// createStandardFileLogger creates a fallback logger when Lethe is not available
func createStandardFileLogger(filename string, level Level, opts ...Option) (*Logger, error) {
	// Validate filename to prevent directory traversal attacks
	cleanPath := filepath.Clean(filename)

	// Check for dangerous path patterns instead of exact equality
	if filepath.IsAbs(cleanPath) != filepath.IsAbs(filename) {
		return nil, fmt.Errorf("invalid file path: %s", filename)
	}

	// Prevent directory traversal attempts
	if containsTraversal(cleanPath) {
		return nil, fmt.Errorf("invalid file path: %s", filename)
	}

	// Use more restrictive file permissions (0600 instead of 0644)
	file, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", cleanPath, err)
	}

	cfg := Config{
		Level:    level,
		Output:   file,
		Encoder:  NewJSONEncoder(),
		Capacity: 8192, // Standard 8KB buffer
	}

	return New(cfg, opts...)
}

// containsTraversal checks for directory traversal patterns in a file path
func containsTraversal(path string) bool {
	// Check for common directory traversal patterns
	if strings.Contains(path, "..") || strings.Contains(path, "~") {
		return true
	}

	// Unix-specific dangerous paths
	if strings.HasPrefix(path, "/etc/") ||
		strings.HasPrefix(path, "/proc/") ||
		strings.HasPrefix(path, "/sys/") ||
		strings.HasPrefix(path, "/root/") {
		return true
	}

	// Windows-specific dangerous paths
	if strings.HasPrefix(path, "C:\\Windows\\") ||
		strings.HasPrefix(path, "C:\\Program Files\\") ||
		strings.HasPrefix(path, "\\root\\") {
		return true
	}

	return false
}
