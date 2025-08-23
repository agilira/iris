// main_test.go: Tests for configuration loading examples
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// captureStdout captures stdout during function execution
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestJSONConfigExample tests the JSON configuration loading example
func TestJSONConfigExample(t *testing.T) {
	output := captureStdout(func() {
		jsonConfigExample()
	})

	if output == "" {
		t.Error("Expected output from jsonConfigExample, got empty string")
	}

	// Check for expected content
	expectedContents := []string{
		"Loaded JSON config:",
		"Level=",
		"Capacity=",
		"BatchSize=",
	}

	for _, content := range expectedContents {
		if !contains(output, content) {
			t.Errorf("Expected output to contain %q, but it didn't. Output: %s", content, output)
		}
	}
}

// TestEnvConfigExample tests the environment variable configuration example
func TestEnvConfigExample(t *testing.T) {
	output := captureStdout(func() {
		envConfigExample()
	})

	if output == "" {
		t.Error("Expected output from envConfigExample, got empty string")
	}

	// Check for expected content
	expectedContents := []string{
		"Loaded env config:",
		"Level=",
		"Capacity=",
		"BatchSize=",
	}

	for _, content := range expectedContents {
		if !contains(output, content) {
			t.Errorf("Expected output to contain %q, but it didn't. Output: %s", content, output)
		}
	}
}

// TestMultiSourceConfigExample tests the multi-source configuration example
func TestMultiSourceConfigExample(t *testing.T) {
	output := captureStdout(func() {
		multiSourceConfigExample()
	})

	if output == "" {
		t.Error("Expected output from multiSourceConfigExample, got empty string")
	}

	// Check for expected content
	expectedContents := []string{
		"Multi-source config:",
		"Level=",
		"Capacity=",
		"BatchSize=",
	}

	for _, content := range expectedContents {
		if !contains(output, content) {
			t.Errorf("Expected output to contain %q, but it didn't. Output: %s", content, output)
		}
	}
}

// TestProductionDeploymentExample tests the production deployment example
func TestProductionDeploymentExample(t *testing.T) {
	output := captureStdout(func() {
		productionDeploymentExample()
	})

	if output == "" {
		t.Error("Expected output from productionDeploymentExample, got empty string")
	}

	// Check for expected content
	expectedContents := []string{
		"Production config:",
		"Level=",
		"Capacity=",
	}

	for _, content := range expectedContents {
		if !contains(output, content) {
			t.Errorf("Expected output to contain %q, but it didn't. Output: %s", content, output)
		}
	}
}

// TestLoadProductionConfig tests the loadProductionConfig function directly
func TestLoadProductionConfig(t *testing.T) {
	// Create temporary config directory
	configDir := filepath.Join(os.TempDir(), "iris_test_configs")
	os.MkdirAll(configDir, 0755)
	defer os.RemoveAll(configDir)

	tests := []struct {
		name        string
		setupEnv    func()
		cleanupEnv  func()
		expectError bool
	}{
		{
			name: "Production_Environment",
			setupEnv: func() {
				// Create production config
				prodConfig := `{"level": "info", "capacity": 8192, "batch_size": 16}`
				prodPath := filepath.Join(configDir, "production.json")
				os.WriteFile(prodPath, []byte(prodConfig), 0644)
				os.Setenv("ENVIRONMENT", "production")
			},
			cleanupEnv: func() {
				os.Unsetenv("ENVIRONMENT")
			},
			expectError: false,
		},
		{
			name: "Development_Environment",
			setupEnv: func() {
				// Create development config
				devConfig := `{"level": "debug", "capacity": 4096, "batch_size": 8}`
				devPath := filepath.Join(configDir, "development.json")
				os.WriteFile(devPath, []byte(devConfig), 0644)
				os.Setenv("ENVIRONMENT", "development")
			},
			cleanupEnv: func() {
				os.Unsetenv("ENVIRONMENT")
			},
			expectError: false,
		},
		{
			name: "Fallback_To_Env_Variables",
			setupEnv: func() {
				os.Setenv("ENVIRONMENT", "nonexistent")
				os.Setenv("IRIS_LEVEL", "warn")
				os.Setenv("IRIS_CAPACITY", "16384")
			},
			cleanupEnv: func() {
				os.Unsetenv("ENVIRONMENT")
				os.Unsetenv("IRIS_LEVEL")
				os.Unsetenv("IRIS_CAPACITY")
			},
			expectError: false,
		},
		{
			name: "Ultimate_Fallback_To_Defaults",
			setupEnv: func() {
				os.Setenv("ENVIRONMENT", "nonexistent")
				// Don't set any IRIS_* env variables
			},
			cleanupEnv: func() {
				os.Unsetenv("ENVIRONMENT")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			config, err := loadProductionConfig(configDir)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if config == nil {
				t.Error("Expected config to be non-nil")
			}
		})
	}
}

// TestMainFunction tests the main function execution
func TestMainFunction(t *testing.T) {
	output := captureStdout(func() {
		main()
	})

	if output == "" {
		t.Error("Expected output from main function, got empty string")
	}

	// Check for expected section headers
	expectedSections := []string{
		"=== Example 1: JSON Configuration Loading ===",
		"=== Example 2: Environment Variable Configuration ===",
		"=== Example 3: Multi-Source Configuration ===",
		"=== Example 4: Production Deployment Pattern ===",
	}

	for _, section := range expectedSections {
		if !contains(output, section) {
			t.Errorf("Expected output to contain section %q, but it didn't", section)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
