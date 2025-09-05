// config_loader_test.go: Tests for configuration loading
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"os"
	"strings"
	"testing"
)

// Helper function to safely close config logger ignoring expected errors
func safeCloseConfigLogger(t *testing.T, logger *Logger) {
	if err := logger.Close(); err != nil &&
		!strings.Contains(err.Error(), "sync /dev/stdout: invalid argument") &&
		!strings.Contains(err.Error(), "sync /dev/stdout: bad file descriptor") &&
		!strings.Contains(err.Error(), "ring buffer flush failed") {
		t.Errorf("Failed to close logger: %v", err)
	}
}

func TestLoadConfigFromJSON(t *testing.T) {
	// Test valid JSON file
	configJSON := `{
  "level": "debug",
  "format": "json",
  "output": "stdout",
  "capacity": 8192,
  "batch_size": 16,
  "enable_caller": true,
  "development": true,
  "idle_strategy": "efficient"
}`

	tmpFile, err := os.CreateTemp("", "iris_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpFile.WriteString(configJSON); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load configuration
	config, err := LoadConfigFromJSON(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if config.Level != Debug {
		t.Errorf("Expected Debug level, got %v", config.Level)
	}
	if config.Capacity != 8192 {
		t.Errorf("Expected capacity 8192, got %d", config.Capacity)
	}
	if config.BatchSize != 16 {
		t.Errorf("Expected batch size 16, got %d", config.BatchSize)
	}
	if config.IdleStrategy == nil {
		t.Error("Expected idle strategy to be set")
	} else if config.IdleStrategy.String() != "sleeping" {
		t.Errorf("Expected 'sleeping' idle strategy, got %q", config.IdleStrategy.String())
	}

	// Test invalid file path
	_, err = LoadConfigFromJSON("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}

	// Test invalid JSON
	invalidJSON := `{
  "level": "debug",
  "format": "json"` // missing closing brace

	tmpFileInvalid, err := os.CreateTemp("", "iris_config_invalid_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFileInvalid.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpFileInvalid.WriteString(invalidJSON); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	if err := tmpFileInvalid.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	_, err = LoadConfigFromJSON(tmpFileInvalid.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	// Test empty file
	emptyFile, err := os.CreateTemp("", "iris_config_empty_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(emptyFile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()
	if err := emptyFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	_, err = LoadConfigFromJSON(emptyFile.Name())
	if err == nil {
		t.Error("Expected error for empty JSON file, got nil")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Test with all environment variables set
	if err := os.Setenv("IRIS_LEVEL", "warn"); err != nil {
		t.Errorf("Failed to set IRIS_LEVEL: %v", err)
	}
	if err := os.Setenv("IRIS_CAPACITY", "4096"); err != nil {
		t.Errorf("Failed to set IRIS_CAPACITY: %v", err)
	}
	if err := os.Setenv("IRIS_ENABLE_CALLER", "true"); err != nil {
		t.Errorf("Failed to set IRIS_ENABLE_CALLER: %v", err)
	}
	if err := os.Setenv("IRIS_DEVELOPMENT", "1"); err != nil {
		t.Errorf("Failed to set IRIS_DEVELOPMENT: %v", err)
	}
	if err := os.Setenv("IRIS_FORMAT", "console"); err != nil {
		t.Errorf("Failed to set IRIS_FORMAT: %v", err)
	}
	if err := os.Setenv("IRIS_OUTPUT", "stderr"); err != nil {
		t.Errorf("Failed to set IRIS_OUTPUT: %v", err)
	}

	defer func() {
		if err := os.Unsetenv("IRIS_LEVEL"); err != nil {
			t.Errorf("Failed to unset IRIS_LEVEL: %v", err)
		}
		if err := os.Unsetenv("IRIS_CAPACITY"); err != nil {
			t.Errorf("Failed to unset IRIS_CAPACITY: %v", err)
		}
		if err := os.Unsetenv("IRIS_ENABLE_CALLER"); err != nil {
			t.Errorf("Failed to unset IRIS_ENABLE_CALLER: %v", err)
		}
		if err := os.Unsetenv("IRIS_DEVELOPMENT"); err != nil {
			t.Errorf("Failed to unset IRIS_DEVELOPMENT: %v", err)
		}
		if err := os.Unsetenv("IRIS_FORMAT"); err != nil {
			t.Errorf("Failed to unset IRIS_FORMAT: %v", err)
		}
		if err := os.Unsetenv("IRIS_OUTPUT"); err != nil {
			t.Errorf("Failed to unset IRIS_OUTPUT: %v", err)
		}
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load env config: %v", err)
	}

	// Verify environment variables were parsed
	if config.Level != Warn {
		t.Errorf("Expected Warn level, got %v", config.Level)
	}
	if config.Capacity != 4096 {
		t.Errorf("Expected capacity 4096, got %d", config.Capacity)
	}

	// Test with invalid values (should be ignored, not error)
	if err := os.Setenv("IRIS_CAPACITY", "invalid_number"); err != nil {
		t.Errorf("Failed to set IRIS_CAPACITY: %v", err)
	}
	config3, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv should not error on invalid values: %v", err)
	}
	// Invalid capacity should be ignored, using default value
	if config3.Capacity != 0 {
		t.Errorf("Expected default capacity (0) for invalid value, got %d", config3.Capacity)
	}
	if err := os.Unsetenv("IRIS_CAPACITY"); err != nil {
		t.Errorf("Failed to unset IRIS_CAPACITY: %v", err)
	}

	// Test with no environment variables set
	config2, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load env config with defaults: %v", err)
	}
	if config2 == nil {
		t.Error("Expected non-nil config with defaults")
	}
}

func TestLevelParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", Debug},
		{"DEBUG", Debug},
		{"info", Info},
		{"INFO", Info},
		{"warn", Warn},
		{"warning", Warn},
		{"error", Error},
		{"ERROR", Error},
		{"panic", Panic},
		{"fatal", Fatal},
		{"invalid", Info}, // Default fallback
		{"", Info},        // Default fallback
	}

	for _, test := range tests {
		result := parseLevel(test.input)
		if result != test.expected {
			t.Errorf("parseLevel(%q) = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestIdleStrategyParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"spinning", "spinning"},
		{"SPINNING", "spinning"},
		{"sleeping", "sleeping"},
		{"SLEEPING", "sleeping"},
		{"yielding", "yielding"},
		{"YIELDING", "yielding"},
		{"channel", "channel"},
		{"CHANNEL", "channel"},
		{"progressive", "progressive"},
		{"PROGRESSIVE", "progressive"},
		{"balanced", "progressive"}, // BalancedStrategy is NewProgressiveIdleStrategy()
		{"BALANCED", "progressive"},
		{"invalid", "progressive"}, // Default fallback to BalancedStrategy (progressive)
		{"", "progressive"},        // Default fallback to BalancedStrategy (progressive)
	}

	for _, test := range tests {
		result := parseIdleStrategy(test.input)
		if result.String() != test.expected {
			t.Errorf("parseIdleStrategy(%q) = %s, want %s", test.input, result.String(), test.expected)
		}
	}
}

func TestIdleStrategyConfiguration(t *testing.T) {
	// Test JSON configuration
	t.Run("JSON", func(t *testing.T) {
		configJSON := `{
  "level": "info",
  "idle_strategy": "yielding"
}`

		tmpFile, err := os.CreateTemp("", "iris_config_idle_*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Errorf("Failed to remove temp file: %v", err)
			}
		}()

		if _, err := tmpFile.WriteString(configJSON); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		config, err := LoadConfigFromJSON(tmpFile.Name())
		if err != nil {
			t.Fatalf("LoadConfigFromJSON failed: %v", err)
		}

		if config.IdleStrategy == nil {
			t.Error("Expected idle strategy to be set")
		} else if config.IdleStrategy.String() != "yielding" {
			t.Errorf("Expected 'yielding' idle strategy, got %q", config.IdleStrategy.String())
		}
	})

	// Test environment variable configuration
	t.Run("Environment", func(t *testing.T) {
		if err := os.Setenv("IRIS_IDLE_STRATEGY", "channel"); err != nil {
			t.Errorf("Failed to set IRIS_IDLE_STRATEGY: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("IRIS_IDLE_STRATEGY"); err != nil {
				t.Errorf("Failed to unset IRIS_IDLE_STRATEGY: %v", err)
			}
		}()

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("LoadConfigFromEnv failed: %v", err)
		}

		if config.IdleStrategy == nil {
			t.Error("Expected idle strategy to be set")
		} else if config.IdleStrategy.String() != "channel" {
			t.Errorf("Expected 'channel' idle strategy, got %q", config.IdleStrategy.String())
		}
	})

	// Test invalid values fallback to default
	t.Run("InvalidValues", func(t *testing.T) {
		if err := os.Setenv("IRIS_IDLE_STRATEGY", "invalid_strategy"); err != nil {
			t.Errorf("Failed to set IRIS_IDLE_STRATEGY: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("IRIS_IDLE_STRATEGY"); err != nil {
				t.Errorf("Failed to unset IRIS_IDLE_STRATEGY: %v", err)
			}
		}()

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("LoadConfigFromEnv failed: %v", err)
		}

		if config.IdleStrategy == nil {
			t.Error("Expected idle strategy to be set")
		} else if config.IdleStrategy.String() != "progressive" {
			t.Errorf("Expected 'progressive' idle strategy for invalid input (BalancedStrategy), got %q", config.IdleStrategy.String())
		}
	})
}

// TestDynamicConfigWatcherLifecycle tests the lifecycle of DynamicConfigWatcher
func TestDynamicConfigWatcherLifecycle(t *testing.T) {
	t.Run("NewDynamicConfigWatcher", func(t *testing.T) {
		// Create a temp config file
		configJSON := `{"level": "info", "format": "json"}`
		tmpFile, err := os.CreateTemp("", "iris_dynamic_*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Errorf("Failed to remove temp file: %v", err)
			}
		}()

		_, err = tmpFile.WriteString(configJSON)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		// Test creating a new watcher
		watcher, err := NewDynamicConfigWatcher(tmpFile.Name(), nil)
		if err != nil {
			t.Fatalf("Expected successful watcher creation, got error: %v", err)
		}

		if watcher == nil {
			t.Fatal("Expected non-nil watcher")
		}

		// Clean up - only stop if watcher was actually started
		if watcher.IsRunning() {
			if err := watcher.Stop(); err != nil {
				t.Errorf("Failed to stop watcher: %v", err)
			}
		}
	})

	t.Run("Start_And_Stop", func(t *testing.T) {
		// Create a temp config file
		configJSON := `{"level": "debug", "format": "text"}`
		tmpFile, err := os.CreateTemp("", "iris_dynamic_*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Errorf("Failed to remove temp file: %v", err)
			}
		}()

		_, err = tmpFile.WriteString(configJSON)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		// Create watcher
		watcher, err := NewDynamicConfigWatcher(tmpFile.Name(), nil)
		if err != nil {
			t.Fatalf("Failed to create watcher: %v", err)
		}

		// Test Start
		err = watcher.Start()
		if err != nil {
			t.Fatalf("Expected successful start, got error: %v", err)
		}

		// Test IsRunning
		if !watcher.IsRunning() {
			t.Error("Expected watcher to be running after Start()")
		}

		// Test Stop
		err = watcher.Stop()
		if err != nil {
			t.Fatalf("Expected successful stop, got error: %v", err)
		}

		// Test IsRunning after stop
		if watcher.IsRunning() {
			t.Error("Expected watcher to not be running after Stop()")
		}
	})

	t.Run("Start_Nonexistent_File", func(t *testing.T) {
		// Test starting watcher with nonexistent file
		watcher, err := NewDynamicConfigWatcher("/nonexistent/file.json", nil)
		if err == nil {
			t.Error("Expected error when creating watcher with nonexistent file")
			if err := watcher.Stop(); err != nil {
				t.Errorf("Failed to stop watcher: %v", err)
			}
		}
	})

	t.Run("EnableDynamicLevel", func(t *testing.T) {
		// Create a temp config file
		configJSON := `{"level": "warn", "format": "json"}`
		tmpFile, err := os.CreateTemp("", "iris_dynamic_*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Errorf("Failed to remove temp file: %v", err)
			}
		}()

		_, err = tmpFile.WriteString(configJSON)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		// Create a logger to test with
		logger, err := New(Config{
			Level:    Info,
			Encoder:  NewTextEncoder(),
			Output:   os.Stdout,
			Capacity: 1024, // Safe capacity for CI
		})
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer safeCloseConfigLogger(t, logger)

		// Enable dynamic level watching
		watcher, err := EnableDynamicLevel(logger, tmpFile.Name())
		if err != nil {
			t.Fatalf("Expected successful EnableDynamicLevel, got error: %v", err)
		}
		defer func() {
			if err := watcher.Stop(); err != nil {
				t.Errorf("Failed to stop watcher: %v", err)
			}
		}()

		if !watcher.IsRunning() {
			t.Error("Expected watcher to be running after EnableDynamicLevel")
		}
	})
}

// TestParseBackpressurePolicy tests parseBackpressurePolicy function
func TestParseBackpressurePolicy(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"drop", "DropOnFull"},
		{"block", "BlockOnFull"},
		{"DROP", "DropOnFull"},
		{"BLOCK", "BlockOnFull"},
		{"invalid", "DropOnFull"}, // fallback to default
		{"", "DropOnFull"},        // fallback to default
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseBackpressurePolicy(tc.input)

			if result.String() != tc.expected {
				t.Errorf("Expected %q for input %q, got %q", tc.expected, tc.input, result.String())
			}
		})
	}
}
