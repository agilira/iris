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
	// Create temporary config file
	configJSON := `{
  "level": "debug",
  "format": "json",
  "output": "stdout",
  "capacity": 8192,
  "batch_size": 16,
  "enable_caller": true,
  "development": true
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
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("IRIS_LEVEL", "warn")
	os.Setenv("IRIS_CAPACITY", "4096")
	os.Setenv("IRIS_ENABLE_CALLER", "true")
	os.Setenv("IRIS_DEVELOPMENT", "1")

	defer func() {
		os.Unsetenv("IRIS_LEVEL")
		os.Unsetenv("IRIS_CAPACITY")
		os.Unsetenv("IRIS_ENABLE_CALLER")
		os.Unsetenv("IRIS_DEVELOPMENT")
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
