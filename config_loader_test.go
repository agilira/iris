// config_loader_test.go: Tests for configuration loading
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"os"
	"testing"
)

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
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configJSON); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

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
	defer os.Remove(tmpFileInvalid.Name())

	if _, err := tmpFileInvalid.WriteString(invalidJSON); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFileInvalid.Close()

	_, err = LoadConfigFromJSON(tmpFileInvalid.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	// Test empty file
	emptyFile, err := os.CreateTemp("", "iris_config_empty_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(emptyFile.Name())
	emptyFile.Close()

	_, err = LoadConfigFromJSON(emptyFile.Name())
	if err == nil {
		t.Error("Expected error for empty JSON file, got nil")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Test with all environment variables set
	os.Setenv("IRIS_LEVEL", "warn")
	os.Setenv("IRIS_CAPACITY", "4096")
	os.Setenv("IRIS_ENABLE_CALLER", "true")
	os.Setenv("IRIS_DEVELOPMENT", "1")
	os.Setenv("IRIS_FORMAT", "console")
	os.Setenv("IRIS_OUTPUT", "stderr")

	defer func() {
		os.Unsetenv("IRIS_LEVEL")
		os.Unsetenv("IRIS_CAPACITY")
		os.Unsetenv("IRIS_ENABLE_CALLER")
		os.Unsetenv("IRIS_DEVELOPMENT")
		os.Unsetenv("IRIS_FORMAT")
		os.Unsetenv("IRIS_OUTPUT")
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
	os.Setenv("IRIS_CAPACITY", "invalid_number")
	config3, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv should not error on invalid values: %v", err)
	}
	// Invalid capacity should be ignored, using default value
	if config3.Capacity != 0 {
		t.Errorf("Expected default capacity (0) for invalid value, got %d", config3.Capacity)
	}
	os.Unsetenv("IRIS_CAPACITY")

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
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(configJSON); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		tmpFile.Close()

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
		os.Setenv("IRIS_IDLE_STRATEGY", "channel")
		defer os.Unsetenv("IRIS_IDLE_STRATEGY")

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
		os.Setenv("IRIS_IDLE_STRATEGY", "invalid_strategy")
		defer os.Unsetenv("IRIS_IDLE_STRATEGY")

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
