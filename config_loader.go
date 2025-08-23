// config_loader.go: Configuration loading from multiple sources
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/agilira/iris/internal/zephyroslite"
)

// LoadConfigFromJSON loads logger configuration from a JSON file
func LoadConfigFromJSON(filename string) (*Config, error) {
	var config Config

	data, err := os.ReadFile(filename)
	if err != nil {
		return &config, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON into a temporary structure
	var jsonConfig struct {
		Level              string `json:"level"`
		Format             string `json:"format"`
		Output             string `json:"output"`
		Capacity           int64  `json:"capacity"`
		BatchSize          int64  `json:"batch_size"`
		EnableCaller       bool   `json:"enable_caller"`
		Development        bool   `json:"development"`
		Name               string `json:"name"`
		BackpressurePolicy string `json:"backpressure_policy"`
	}

	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return &config, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	// Convert level string to Level enum
	config.Level = parseLevel(jsonConfig.Level)

	// Set encoder based on format
	switch strings.ToLower(jsonConfig.Format) {
	case "json":
		config.Encoder = NewJSONEncoder()
	case "text", "console":
		config.Encoder = NewTextEncoder()
	default:
		config.Encoder = NewJSONEncoder() // Default to JSON
	}

	// Set output
	switch strings.ToLower(jsonConfig.Output) {
	case "stdout":
		config.Output = WrapWriter(os.Stdout)
	case "stderr":
		config.Output = WrapWriter(os.Stderr)
	default:
		// Assume it's a file path
		if jsonConfig.Output != "" {
			file, err := os.OpenFile(jsonConfig.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return &config, fmt.Errorf("failed to open output file: %w", err)
			}
			config.Output = WrapWriter(file)
		} else {
			config.Output = WrapWriter(os.Stdout)
		}
	}

	// Set capacity and batch size
	if jsonConfig.Capacity > 0 {
		config.Capacity = jsonConfig.Capacity
	}
	if jsonConfig.BatchSize > 0 {
		config.BatchSize = jsonConfig.BatchSize
	}

	// Set name
	if jsonConfig.Name != "" {
		config.Name = jsonConfig.Name
	}

	// Set backpressure policy
	config.BackpressurePolicy = parseBackpressurePolicy(jsonConfig.BackpressurePolicy)

	return &config, nil
}

// LoadConfigFromEnv loads logger configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	var config Config

	// Level from IRIS_LEVEL
	if levelStr := os.Getenv("IRIS_LEVEL"); levelStr != "" {
		config.Level = parseLevel(levelStr)
	} else {
		config.Level = Info // Default level
	}

	// Format from IRIS_FORMAT
	format := os.Getenv("IRIS_FORMAT")
	switch strings.ToLower(format) {
	case "json":
		config.Encoder = NewJSONEncoder()
	case "text", "console":
		config.Encoder = NewTextEncoder()
	default:
		config.Encoder = NewJSONEncoder() // Default to JSON
	}

	// Output from IRIS_OUTPUT
	output := os.Getenv("IRIS_OUTPUT")
	switch strings.ToLower(output) {
	case "stdout", "":
		config.Output = WrapWriter(os.Stdout)
	case "stderr":
		config.Output = WrapWriter(os.Stderr)
	default:
		// Assume it's a file path
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return &config, fmt.Errorf("failed to open output file: %w", err)
		}
		config.Output = WrapWriter(file)
	}

	// Capacity from IRIS_CAPACITY
	if capacityStr := os.Getenv("IRIS_CAPACITY"); capacityStr != "" {
		if capacity, err := strconv.ParseInt(capacityStr, 10, 64); err == nil && capacity > 0 {
			config.Capacity = capacity
		}
	}

	// BatchSize from IRIS_BATCH_SIZE
	if batchStr := os.Getenv("IRIS_BATCH_SIZE"); batchStr != "" {
		if batchSize, err := strconv.ParseInt(batchStr, 10, 64); err == nil && batchSize > 0 {
			config.BatchSize = batchSize
		}
	}

	// Name from IRIS_NAME
	if name := os.Getenv("IRIS_NAME"); name != "" {
		config.Name = name
	}

	// BackpressurePolicy from IRIS_BACKPRESSURE_POLICY
	if policyStr := os.Getenv("IRIS_BACKPRESSURE_POLICY"); policyStr != "" {
		config.BackpressurePolicy = parseBackpressurePolicy(policyStr)
	}

	return &config, nil
}

// LoadConfigMultiSource loads configuration from multiple sources with precedence:
// 1. Environment variables (highest priority)
// 2. JSON file
// 3. Default values (lowest priority)
func LoadConfigMultiSource(jsonFile string) (*Config, error) {
	// Start with defaults
	config := Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  WrapWriter(os.Stdout),
	}

	// Load from JSON file if provided
	if jsonFile != "" {
		if jsonConfig, err := LoadConfigFromJSON(jsonFile); err == nil {
			// Merge JSON config (only non-zero values)
			if jsonConfig.Level != 0 {
				config.Level = jsonConfig.Level
			}
			if jsonConfig.Encoder != nil {
				config.Encoder = jsonConfig.Encoder
			}
			if jsonConfig.Output != nil {
				config.Output = jsonConfig.Output
			}
			if jsonConfig.Capacity > 0 {
				config.Capacity = jsonConfig.Capacity
			}
			if jsonConfig.BatchSize > 0 {
				config.BatchSize = jsonConfig.BatchSize
			}
			if jsonConfig.Name != "" {
				config.Name = jsonConfig.Name
			}
			if jsonConfig.BackpressurePolicy != zephyroslite.DropOnFull {
				config.BackpressurePolicy = jsonConfig.BackpressurePolicy
			}
		}
	}

	// Override with environment variables
	envConfig, err := LoadConfigFromEnv()
	if err != nil {
		return &config, err
	}

	// Apply environment overrides
	if levelStr := os.Getenv("IRIS_LEVEL"); levelStr != "" {
		config.Level = envConfig.Level
	}
	if format := os.Getenv("IRIS_FORMAT"); format != "" {
		config.Encoder = envConfig.Encoder
	}
	if output := os.Getenv("IRIS_OUTPUT"); output != "" {
		config.Output = envConfig.Output
	}
	if capacityStr := os.Getenv("IRIS_CAPACITY"); capacityStr != "" {
		config.Capacity = envConfig.Capacity
	}
	if batchStr := os.Getenv("IRIS_BATCH_SIZE"); batchStr != "" {
		config.BatchSize = envConfig.BatchSize
	}
	if name := os.Getenv("IRIS_NAME"); name != "" {
		config.Name = envConfig.Name
	}
	if policyStr := os.Getenv("IRIS_BACKPRESSURE_POLICY"); policyStr != "" {
		config.BackpressurePolicy = envConfig.BackpressurePolicy
	}

	return &config, nil
}

// parseLevel converts a string to a Level enum
func parseLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warn", "warning":
		return Warn
	case "error":
		return Error
	case "panic":
		return Panic
	case "fatal":
		return Fatal
	default:
		return Info // Default level
	}
}

// parseBackpressurePolicy converts a string to a BackpressurePolicy enum
func parseBackpressurePolicy(policyStr string) zephyroslite.BackpressurePolicy {
	switch strings.ToLower(policyStr) {
	case "drop", "drop_on_full", "droponful":
		return zephyroslite.DropOnFull
	case "block", "block_on_full", "blockonful":
		return zephyroslite.BlockOnFull
	default:
		return zephyroslite.DropOnFull // Default policy
	}
}
