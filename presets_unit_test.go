// presets_unit_test.go: Unit tests for configuration presets
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"os"
	"testing"
	"time"
)

// TestNewDevelopment tests development preset creation
func TestNewDevelopment(t *testing.T) {
	logger, err := NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create development logger: %v", err)
	}
	defer logger.Close()

	// Test that logger was created successfully
	if logger == nil {
		t.Error("Development logger should not be nil")
	}

	// Test logging functionality
	logger.Info("test development message", String("preset", "development"))
}

// TestNewProduction tests production preset creation
func TestNewProduction(t *testing.T) {
	logger, err := NewProduction()
	if err != nil {
		t.Fatalf("Failed to create production logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Error("Production logger should not be nil")
	}

	logger.Info("test production message", String("preset", "production"))
}

// TestNewExample tests example preset creation
func TestNewExample(t *testing.T) {
	logger, err := NewExample()
	if err != nil {
		t.Fatalf("Failed to create example logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Error("Example logger should not be nil")
	}

	logger.Info("test example message", String("preset", "example"))
}

// TestNewUltraFast tests ultra-fast preset creation
func TestNewUltraFast(t *testing.T) {
	logger, err := NewUltraFast()
	if err != nil {
		t.Fatalf("Failed to create ultra-fast logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Error("Ultra-fast logger should not be nil")
	}

	logger.Info("test ultra-fast message", String("preset", "ultrafast"))
}

// TestNewFastText tests fast text preset creation
func TestNewFastText(t *testing.T) {
	logger, err := NewFastText()
	if err != nil {
		t.Fatalf("Failed to create fast text logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Error("Fast text logger should not be nil")
	}

	logger.Info("test fast text message", String("preset", "fasttext"))
}

// TestNewDevelopmentWithStackTrace tests development with stack trace preset
func TestNewDevelopmentWithStackTrace(t *testing.T) {
	logger, err := NewDevelopmentWithStackTrace()
	if err != nil {
		t.Fatalf("Failed to create development with stack trace logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Error("Development with stack trace logger should not be nil")
	}

	logger.Error("test error with stack trace", String("preset", "dev_stacktrace"))
}

// TestNewDebugWithStackTrace tests debug with stack trace preset
func TestNewDebugWithStackTrace(t *testing.T) {
	logger, err := NewDebugWithStackTrace()
	if err != nil {
		t.Fatalf("Failed to create debug with stack trace logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Error("Debug with stack trace logger should not be nil")
	}

	logger.Warn("test warning with stack trace", String("preset", "debug_stacktrace"))
}

// TestDevelopmentConfig tests development configuration function
func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()

	if config.Level != DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", config.Level)
	}

	if config.BufferSize != 1024 {
		t.Errorf("Expected BufferSize 1024, got %d", config.BufferSize)
	}

	// Test that config can create a logger
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger from development config: %v", err)
	}
	defer logger.Close()
}

// TestProductionConfig tests production configuration function
func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	if config.Level != InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}

	if config.BufferSize != 8192 {
		t.Errorf("Expected BufferSize 8192, got %d", config.BufferSize)
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger from production config: %v", err)
	}
	defer logger.Close()
}

// TestExampleConfig tests example configuration function
func TestExampleConfig(t *testing.T) {
	config := ExampleConfig()

	if config.Level != InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}

	if !config.DisableTimestamp {
		t.Error("Expected DisableTimestamp to be true for example config")
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger from example config: %v", err)
	}
	defer logger.Close()
}

// TestAllPresetsCreateValidLoggers tests that all presets create working loggers
func TestAllPresetsCreateValidLoggers(t *testing.T) {
	presets := []struct {
		name    string
		creator func() (*Logger, error)
	}{
		{"Development", NewDevelopment},
		{"Production", NewProduction},
		{"Example", NewExample},
		{"UltraFast", NewUltraFast},
		{"FastText", NewFastText},
		{"DevelopmentWithStackTrace", NewDevelopmentWithStackTrace},
		{"DebugWithStackTrace", NewDebugWithStackTrace},
	}

	for _, preset := range presets {
		t.Run(preset.name, func(t *testing.T) {
			logger, err := preset.creator()
			if err != nil {
				t.Fatalf("Failed to create %s logger: %v", preset.name, err)
			}
			defer logger.Close()

			// Test basic logging
			logger.Info("test message", String("preset", preset.name))
		})
	}
}

// TestConfigCreators tests all config creator functions
func TestConfigCreators(t *testing.T) {
	configs := []struct {
		name    string
		creator func() Config
	}{
		{"Development", DevelopmentConfig},
		{"Production", ProductionConfig},
		{"Example", ExampleConfig},
		{"DevelopmentWithStackTrace", DevelopmentWithStackTraceConfig},
		{"DebugWithStackTrace", DebugWithStackTraceConfig},
	}

	for _, config := range configs {
		t.Run(config.name, func(t *testing.T) {
			cfg := config.creator()

			// Test that config creates a valid logger
			logger, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create logger from %s config: %v", config.name, err)
			}
			defer logger.Close()
		})
	}
}

// TestPresetPerformanceCharacteristics tests performance aspects of presets
func TestPresetPerformanceCharacteristics(t *testing.T) {
	// Test that UltraFast preset has appropriate settings for performance
	config := Config{
		Level:            InfoLevel,
		Writer:           DiscardSyncer, // Use discard for testing
		Format:           BinaryFormat,
		BufferSize:       16384,
		BatchSize:        256,
		DisableTimestamp: true,
		EnableCaller:     false,
		UltraFast:        true,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create ultra-fast logger: %v", err)
	}
	defer logger.Close()

	// Test high-frequency logging
	for i := 0; i < 1000; i++ {
		logger.Info("performance test", Int("iteration", i))
	}
}

// TestNewUltraFastFile tests the NewUltraFastFile function
func TestNewUltraFastFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := "/tmp/iris_test_ultrafast.log"

	// Clean up before and after test
	defer func() {
		if err := os.Remove(tmpFile); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to clean up test file: %v", err)
		}
	}()

	logger, err := NewUltraFastFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create ultra fast file logger: %v", err)
	}
	defer logger.Close()

	// Test logging
	logger.Info("test ultra fast file logging")

	// Verify the file was created and has content
	time.Sleep(100 * time.Millisecond) // Allow time for async writes

	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

// TestNewUltraFastFileInvalidPath tests error handling for invalid file paths
func TestNewUltraFastFileInvalidPath(t *testing.T) {
	// Try to create logger with invalid path
	invalidPath := "/nonexistent/directory/file.log"

	logger, err := NewUltraFastFile(invalidPath)
	if err == nil {
		if logger != nil {
			logger.Close()
		}
		t.Error("Expected error when creating logger with invalid path")
	}
}

// TestNewUltraFastNetwork tests the NewUltraFastNetwork function
func TestNewUltraFastNetwork(t *testing.T) {
	// Create a mock network writer using a discard writer for simplicity
	writer := DiscardSyncer

	logger, err := NewUltraFastNetwork(writer)
	if err != nil {
		t.Fatalf("Failed to create ultra fast network logger: %v", err)
	}
	defer logger.Close()

	// Test logging - just verify it doesn't crash
	logger.Info("test ultra fast network logging")

	// Allow time for async processing
	time.Sleep(100 * time.Millisecond)

	// For discard writer, we just test that it completes without error
	// In a real test, you'd use a proper mock writer that captures data
}
