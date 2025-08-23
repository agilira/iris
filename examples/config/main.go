// examples/configuration_loading.go: Demonstrates configuration loading
//
// This example shows how to use IRIS configuration loading from JSON files,
// environment variables, and multi-source configurations.
//
// To run: go run examples/configuration_loading.go
//
// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agilira/iris"
)

func main() {
	// Example 1: JSON configuration loading
	fmt.Println("=== Example 1: JSON Configuration Loading ===")
	jsonConfigExample()

	// Example 2: Environment variable configuration
	fmt.Println("\n=== Example 2: Environment Variable Configuration ===")
	envConfigExample()

	// Example 3: Multi-source configuration with precedence
	fmt.Println("\n=== Example 3: Multi-Source Configuration ===")
	multiSourceConfigExample()

	// Example 4: Production deployment pattern
	fmt.Println("\n=== Example 4: Production Deployment Pattern ===")
	productionDeploymentExample()
}

func jsonConfigExample() {
	// Create temporary config file
	configJSON := `{
  "level": "debug",
  "format": "json",
  "output": "stdout",
  "capacity": 8192,
  "batch_size": 16,
  "enable_caller": true,
  "name": "example-service"
}`

	tmpFile, err := os.CreateTemp("", "iris_example_*.json")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configJSON); err != nil {
		panic(err)
	}
	tmpFile.Close()

	// Load configuration from JSON
	config, err := iris.LoadConfigFromJSON(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded JSON config: Level=%v, Capacity=%d, BatchSize=%d\n",
		config.Level, config.Capacity, config.BatchSize)

	// Create logger with loaded configuration
	logger, err := iris.New(*config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Start()
	logger.Info("Logger created from JSON configuration")
	logger.Debug("Debug logging enabled from config")
}

func envConfigExample() {
	// Set environment variables
	os.Setenv("IRIS_LEVEL", "warn")
	os.Setenv("IRIS_CAPACITY", "16384")
	os.Setenv("IRIS_BATCH_SIZE", "32")
	os.Setenv("IRIS_ENABLE_CALLER", "false")
	os.Setenv("IRIS_NAME", "env-service")

	defer func() {
		// Clean up environment variables
		os.Unsetenv("IRIS_LEVEL")
		os.Unsetenv("IRIS_CAPACITY")
		os.Unsetenv("IRIS_BATCH_SIZE")
		os.Unsetenv("IRIS_ENABLE_CALLER")
		os.Unsetenv("IRIS_NAME")
	}()

	// Load configuration from environment
	config, err := iris.LoadConfigFromEnv()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded env config: Level=%v, Capacity=%d, BatchSize=%d\n",
		config.Level, config.Capacity, config.BatchSize)

	// Create logger with environment configuration
	logger, err := iris.New(*config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Start()
	logger.Info("Logger created from environment configuration")
	logger.Warn("Warning level logging enabled from environment")
}

func multiSourceConfigExample() {
	// Create JSON config file
	configJSON := `{
  "level": "info",
  "capacity": 4096,
  "batch_size": 8,
  "format": "json",
  "name": "multi-source-service"
}`

	tmpFile, err := os.CreateTemp("", "iris_multi_*.json")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configJSON)
	tmpFile.Close()

	// Set environment variables (these will override JSON settings)
	os.Setenv("IRIS_LEVEL", "debug")        // Overrides JSON "info"
	os.Setenv("IRIS_BATCH_SIZE", "64")      // Overrides JSON "8"
	// IRIS_CAPACITY not set, so JSON value (4096) will be used

	defer func() {
		os.Unsetenv("IRIS_LEVEL")
		os.Unsetenv("IRIS_BATCH_SIZE")
	}()

	// Load with multi-source precedence: Environment > JSON > Defaults
	config, err := iris.LoadConfigMultiSource(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Multi-source config: Level=%v (from env), Capacity=%d (from JSON), BatchSize=%d (from env)\n",
		config.Level, config.Capacity, config.BatchSize)

	// Create logger with multi-source configuration
	logger, err := iris.New(*config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Start()
	logger.Info("Logger created from multi-source configuration")
	logger.Debug("Debug level from environment override")
}

func productionDeploymentExample() {
	// Simulate production deployment scenario
	// In real production, these would be set by deployment environment

	// Create production config file
	prodConfigJSON := `{
  "level": "info",
  "format": "json",
  "output": "stdout",
  "capacity": 65536,
  "batch_size": 128,
  "enable_caller": false,
  "name": "production-service"
}`

	configDir := filepath.Join(os.TempDir(), "iris_configs")
	os.MkdirAll(configDir, 0755)
	defer os.RemoveAll(configDir)

	prodConfigPath := filepath.Join(configDir, "production.json")
	if err := os.WriteFile(prodConfigPath, []byte(prodConfigJSON), 0644); err != nil {
		panic(err)
	}

	// Simulate container environment variables
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("IRIS_LEVEL", "warn")      // Production override for safety
	os.Setenv("IRIS_CAPACITY", "131072") // Increased for high throughput

	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("IRIS_LEVEL")
		os.Unsetenv("IRIS_CAPACITY")
	}()

	// Production configuration loading pattern
	config, err := loadProductionConfig(configDir)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Production config: Level=%v, Capacity=%d, Format=JSON\n",
		config.Level, config.Capacity)

	// Create production logger
	logger, err := iris.New(*config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Start()

	// Simulate production logging
	logger.Info("Production service started")
	logger.Warn("High memory usage detected")
	logger.Info("Service health check passed")
}

// loadProductionConfig demonstrates a production-ready configuration loading pattern
func loadProductionConfig(configDir string) (*iris.Config, error) {
	// Determine environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	// Try environment-specific config file first
	configPath := filepath.Join(configDir, env+".json")
	
	// Try multi-source loading
	if config, err := iris.LoadConfigMultiSource(configPath); err == nil {
		fmt.Printf("Loaded configuration for environment: %s\n", env)
		return config, nil
	}

	// Fallback to environment variables only
	if config, err := iris.LoadConfigFromEnv(); err == nil {
		fmt.Println("Loaded configuration from environment variables")
		return config, nil
	}

	// Ultimate fallback to safe defaults
	fmt.Println("Using default configuration")
	config := &iris.Config{
		Level:     iris.Info,
		Capacity:  8192,
		BatchSize: 32,
		Output:    iris.WrapWriter(os.Stdout),
		Encoder:   iris.NewJSONEncoder(),
	}

	return config, nil
}
