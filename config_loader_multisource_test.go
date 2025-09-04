// config_loader_multisource_test.go: Tests for LoadConfigMultiSource function
//
// This file provides comprehensive test coverage for LoadConfigMultiSource function
// including JSON file loading, environment variable overrides, error handling,
// and priority testing (env > json > defaults).
//
// Tests are OS-aware and validate all configuration precedence scenarios.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfigMultiSource validates LoadConfigMultiSource function
// Tests JSON loading, environment overrides, and priority system
func TestLoadConfigMultiSource(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"IRIS_LEVEL":      os.Getenv("IRIS_LEVEL"),
		"IRIS_FORMAT":     os.Getenv("IRIS_FORMAT"),
		"IRIS_OUTPUT":     os.Getenv("IRIS_OUTPUT"),
		"IRIS_CAPACITY":   os.Getenv("IRIS_CAPACITY"),
		"IRIS_BATCH_SIZE": os.Getenv("IRIS_BATCH_SIZE"),
		"IRIS_NAME":       os.Getenv("IRIS_NAME"),
	}

	// Clean environment for test
	envVars := []string{"IRIS_LEVEL", "IRIS_FORMAT", "IRIS_OUTPUT", "IRIS_CAPACITY", "IRIS_BATCH_SIZE", "IRIS_NAME"}
	for _, envVar := range envVars {
		if err := os.Unsetenv(envVar); err != nil {
			t.Errorf("Failed to unset env var %s: %v", envVar, err)
		}
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				if err := os.Setenv(key, value); err != nil {
					t.Errorf("Failed to restore env var %s: %v", key, err)
				}
			} else {
				if err := os.Unsetenv(key); err != nil {
					t.Errorf("Failed to unset env var %s: %v", key, err)
				}
			}
		}
	}()

	t.Run("defaults_only_no_json_no_env", func(t *testing.T) {
		config, err := LoadConfigMultiSource("")
		if err != nil {
			t.Errorf("LoadConfigMultiSource(\"\") unexpected error: %v", err)
		}

		if config.Level != Info {
			t.Errorf("Expected default level Info, got %v", config.Level)
		}
		if config.Encoder == nil {
			t.Error("Expected default JSON encoder, got nil")
		}
		if config.Output == nil {
			t.Error("Expected default stdout output, got nil")
		}
	})

	t.Run("json_file_only", func(t *testing.T) {
		// Create temporary JSON config file
		tmpDir := t.TempDir()
		jsonFile := filepath.Join(tmpDir, "test_config.json")

		testConfig := map[string]interface{}{
			"level":      "debug",
			"format":     "text",
			"output":     "stdout",
			"capacity":   2048,
			"batch_size": 64,
			"name":       "test-logger",
		}

		data, err := json.Marshal(testConfig)
		if err != nil {
			t.Fatalf("Failed to marshal test config: %v", err)
		}

		if err := os.WriteFile(jsonFile, data, 0644); err != nil { // #nosec G306 -- Test file permissions
			t.Fatalf("Failed to write test config file: %v", err)
		}

		config, err := LoadConfigMultiSource(jsonFile)
		if err != nil {
			t.Errorf("LoadConfigMultiSource(%q) unexpected error: %v", jsonFile, err)
		}

		if config.Level != Debug {
			t.Errorf("Expected level Debug from JSON, got %v", config.Level)
		}
		if config.Capacity != 2048 {
			t.Errorf("Expected capacity 2048 from JSON, got %d", config.Capacity)
		}
		if config.BatchSize != 64 {
			t.Errorf("Expected batch size 64 from JSON, got %d", config.BatchSize)
		}
		if config.Name != "test-logger" {
			t.Errorf("Expected name 'test-logger' from JSON, got %q", config.Name)
		}
	})

	t.Run("env_overrides_json", func(t *testing.T) {
		// Create temporary JSON config file
		tmpDir := t.TempDir()
		jsonFile := filepath.Join(tmpDir, "test_config.json")

		testConfig := map[string]interface{}{
			"level":      "debug",
			"format":     "text",
			"capacity":   1024,
			"batch_size": 32,
			"name":       "json-logger",
		}

		data, err := json.Marshal(testConfig)
		if err != nil {
			t.Fatalf("Failed to marshal test config: %v", err)
		}

		if err := os.WriteFile(jsonFile, data, 0644); err != nil { // #nosec G306 -- Test file permissions
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Set environment variables to override JSON
		if err := os.Setenv("IRIS_LEVEL", "error"); err != nil {
			t.Errorf("Failed to set IRIS_LEVEL: %v", err)
		}
		if err := os.Setenv("IRIS_FORMAT", "json"); err != nil {
			t.Errorf("Failed to set IRIS_FORMAT: %v", err)
		}
		if err := os.Setenv("IRIS_CAPACITY", "4096"); err != nil {
			t.Errorf("Failed to set IRIS_CAPACITY: %v", err)
		}
		if err := os.Setenv("IRIS_BATCH_SIZE", "128"); err != nil {
			t.Errorf("Failed to set IRIS_BATCH_SIZE: %v", err)
		}
		if err := os.Setenv("IRIS_NAME", "env-logger"); err != nil {
			t.Errorf("Failed to set IRIS_NAME: %v", err)
		}

		config, err := LoadConfigMultiSource(jsonFile)
		if err != nil {
			t.Errorf("LoadConfigMultiSource(%q) unexpected error: %v", jsonFile, err)
		}

		// Environment should override JSON
		if config.Level != Error {
			t.Errorf("Expected level Error from env override, got %v", config.Level)
		}
		if config.Capacity != 4096 {
			t.Errorf("Expected capacity 4096 from env override, got %d", config.Capacity)
		}
		if config.BatchSize != 128 {
			t.Errorf("Expected batch size 128 from env override, got %d", config.BatchSize)
		}
		if config.Name != "env-logger" {
			t.Errorf("Expected name 'env-logger' from env override, got %q", config.Name)
		}
	})

	t.Run("invalid_json_file", func(t *testing.T) {
		// Clear environment first
		envVars := []string{"IRIS_LEVEL", "IRIS_FORMAT", "IRIS_OUTPUT", "IRIS_CAPACITY", "IRIS_BATCH_SIZE", "IRIS_NAME"}
		for _, envVar := range envVars {
			if err := os.Unsetenv(envVar); err != nil {
				t.Errorf("Failed to unset env var %s: %v", envVar, err)
			}
		}

		// Test with non-existent file
		config, err := LoadConfigMultiSource("/nonexistent/file.json")
		// Should not error, just use defaults when JSON fails to load
		if err != nil {
			t.Errorf("LoadConfigMultiSource with invalid file should not error, got: %v", err)
		}

		// Should fall back to defaults
		if config.Level != Info {
			t.Errorf("Expected default level Info when JSON fails, got %v", config.Level)
		}
	})

	t.Run("partial_json_config", func(t *testing.T) {
		// Clear environment first
		envVars := []string{"IRIS_LEVEL", "IRIS_FORMAT", "IRIS_OUTPUT", "IRIS_CAPACITY", "IRIS_BATCH_SIZE", "IRIS_NAME"}
		for _, envVar := range envVars {
			if err := os.Unsetenv(envVar); err != nil {
				t.Errorf("Failed to unset env var %s: %v", envVar, err)
			}
		}

		// Create JSON with only some fields
		tmpDir := t.TempDir()
		jsonFile := filepath.Join(tmpDir, "partial_config.json")

		testConfig := map[string]interface{}{
			"level": "warn",
			"name":  "partial-logger",
			// Missing other fields - should use defaults
		}

		data, err := json.Marshal(testConfig)
		if err != nil {
			t.Fatalf("Failed to marshal partial config: %v", err)
		}

		if err := os.WriteFile(jsonFile, data, 0644); err != nil { // #nosec G306 -- Test file permissions
			t.Fatalf("Failed to write partial config file: %v", err)
		}

		config, err := LoadConfigMultiSource(jsonFile)
		if err != nil {
			t.Errorf("LoadConfigMultiSource(%q) unexpected error: %v", jsonFile, err)
		}

		// Should get JSON values where specified
		if config.Level != Warn {
			t.Errorf("Expected level Warn from partial JSON, got %v", config.Level)
		}
		if config.Name != "partial-logger" {
			t.Errorf("Expected name 'partial-logger' from partial JSON, got %q", config.Name)
		}

		// Should get defaults for missing fields
		if config.Encoder == nil {
			t.Error("Expected default encoder for missing JSON field")
		}
		if config.Output == nil {
			t.Error("Expected default output for missing JSON field")
		}
	})

	t.Run("env_only_no_json", func(t *testing.T) {
		// Set environment variables only
		if err := os.Setenv("IRIS_LEVEL", "fatal"); err != nil {
			t.Errorf("Failed to set IRIS_LEVEL: %v", err)
		}
		if err := os.Setenv("IRIS_FORMAT", "console"); err != nil {
			t.Errorf("Failed to set IRIS_FORMAT: %v", err)
		}
		if err := os.Setenv("IRIS_OUTPUT", "stderr"); err != nil {
			t.Errorf("Failed to set IRIS_OUTPUT: %v", err)
		}
		if err := os.Setenv("IRIS_CAPACITY", "8192"); err != nil {
			t.Errorf("Failed to set IRIS_CAPACITY: %v", err)
		}
		if err := os.Setenv("IRIS_BATCH_SIZE", "256"); err != nil {
			t.Errorf("Failed to set IRIS_BATCH_SIZE: %v", err)
		}
		if err := os.Setenv("IRIS_NAME", "env-only-logger"); err != nil {
			t.Errorf("Failed to set IRIS_NAME: %v", err)
		}

		config, err := LoadConfigMultiSource("")
		if err != nil {
			t.Errorf("LoadConfigMultiSource(\"\") with env vars unexpected error: %v", err)
		}

		if config.Level != Fatal {
			t.Errorf("Expected level Fatal from env, got %v", config.Level)
		}
		if config.Capacity != 8192 {
			t.Errorf("Expected capacity 8192 from env, got %d", config.Capacity)
		}
		if config.BatchSize != 256 {
			t.Errorf("Expected batch size 256 from env, got %d", config.BatchSize)
		}
		if config.Name != "env-only-logger" {
			t.Errorf("Expected name 'env-only-logger' from env, got %q", config.Name)
		}
	})

	t.Run("invalid_env_values", func(t *testing.T) {
		// Clear environment first
		envVars := []string{"IRIS_LEVEL", "IRIS_FORMAT", "IRIS_OUTPUT", "IRIS_CAPACITY", "IRIS_BATCH_SIZE", "IRIS_NAME"}
		for _, envVar := range envVars {
			if err := os.Unsetenv(envVar); err != nil {
				t.Errorf("Failed to unset env var %s: %v", envVar, err)
			}
		}

		// Set invalid environment values
		if err := os.Setenv("IRIS_LEVEL", "invalid-level"); err != nil {
			t.Errorf("Failed to set IRIS_LEVEL: %v", err)
		}
		if err := os.Setenv("IRIS_FORMAT", "invalid-format"); err != nil {
			t.Errorf("Failed to set IRIS_FORMAT: %v", err)
		}
		if err := os.Setenv("IRIS_CAPACITY", "not-a-number"); err != nil {
			t.Errorf("Failed to set IRIS_CAPACITY: %v", err)
		}
		if err := os.Setenv("IRIS_BATCH_SIZE", "also-not-a-number"); err != nil {
			t.Errorf("Failed to set IRIS_BATCH_SIZE: %v", err)
		}

		config, err := LoadConfigMultiSource("")
		// Should handle errors gracefully - invalid level defaults to Info
		if err == nil {
			// parseLevel handles invalid levels by defaulting to Info
			if config.Level != Info {
				t.Errorf("Expected default level Info for invalid env level, got %v", config.Level)
			}
		} else {
			// If there's an error, it's likely from invalid numeric values
			if config == nil {
				t.Error("Expected config even with invalid env values")
			}
		}
	})
}

// TestLoadConfigMultiSource_EnvironmentIntegration tests realistic environment scenarios
func TestLoadConfigMultiSource_EnvironmentIntegration(t *testing.T) {
	// Save and restore environment
	originalEnv := map[string]string{
		"IRIS_LEVEL":      os.Getenv("IRIS_LEVEL"),
		"IRIS_FORMAT":     os.Getenv("IRIS_FORMAT"),
		"IRIS_OUTPUT":     os.Getenv("IRIS_OUTPUT"),
		"IRIS_CAPACITY":   os.Getenv("IRIS_CAPACITY"),
		"IRIS_BATCH_SIZE": os.Getenv("IRIS_BATCH_SIZE"),
		"IRIS_NAME":       os.Getenv("IRIS_NAME"),
	}

	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				if err := os.Setenv(key, value); err != nil {
					t.Errorf("Failed to restore %s: %v", key, err)
				}
			} else {
				if err := os.Unsetenv(key); err != nil {
					t.Errorf("Failed to unset %s: %v", key, err)
				}
			}
		}
	}()

	t.Run("production_like_config", func(t *testing.T) {
		// Clear environment
		envVars := []string{"IRIS_LEVEL", "IRIS_FORMAT", "IRIS_OUTPUT", "IRIS_CAPACITY", "IRIS_BATCH_SIZE", "IRIS_NAME"}
		for _, envVar := range envVars {
			if err := os.Unsetenv(envVar); err != nil {
				t.Errorf("Failed to unset env var %s: %v", envVar, err)
			}
		}

		// Create production-like config

		// Create production JSON config
		tmpDir := t.TempDir()
		jsonFile := filepath.Join(tmpDir, "production.json")

		prodConfig := map[string]interface{}{
			"level":      "info",
			"format":     "json",
			"output":     "stdout",
			"capacity":   4096,
			"batch_size": 100,
			"name":       "production-api",
		}

		data, err := json.Marshal(prodConfig)
		if err != nil {
			t.Fatalf("Failed to marshal production config: %v", err)
		}

		if err := os.WriteFile(jsonFile, data, 0644); err != nil { // #nosec G306 -- Test file permissions
			t.Fatalf("Failed to write production config: %v", err)
		}

		// Override log level for debugging
		if err := os.Setenv("IRIS_LEVEL", "debug"); err != nil {
			t.Errorf("Failed to set IRIS_LEVEL: %v", err)
		}

		config, err := LoadConfigMultiSource(jsonFile)
		if err != nil {
			t.Errorf("Production config load error: %v", err)
		}

		// Should have env override for level
		if config.Level != Debug {
			t.Errorf("Expected Debug level from env override, got %v", config.Level)
		}

		// Should have JSON values for other settings
		if config.Name != "production-api" {
			t.Errorf("Expected production-api name from JSON, got %q", config.Name)
		}
		if config.Capacity != 4096 {
			t.Errorf("Expected capacity 4096 from JSON, got %d", config.Capacity)
		}
	})
}
