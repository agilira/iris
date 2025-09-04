// config_loader.go: Configuration loading from multiple sources
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/agilira/argus"
	"github.com/agilira/iris/internal/zephyroslite"
)

// validateFilePath checks if a file path is safe to use
func validateFilePath(filename string) error {
	if filename == "" {
		return fmt.Errorf("empty file path")
	}

	// Clean the path to resolve any . or .. elements
	cleanPath := filepath.Clean(filename)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal: %s", filename)
	}

	return nil
}

// LoadConfigFromJSON loads logger configuration from a JSON file
func LoadConfigFromJSON(filename string) (*Config, error) {
	var config Config

	// Validate file path for security
	if err := validateFilePath(filename); err != nil {
		return &config, fmt.Errorf("invalid file path: %w", err)
	}

	data, err := os.ReadFile(filename) // #nosec G304 -- Path validation implemented above
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
		IdleStrategy       string `json:"idle_strategy"`
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
			file, err := os.OpenFile(jsonConfig.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
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

	// Set idle strategy
	if jsonConfig.IdleStrategy != "" {
		config.IdleStrategy = parseIdleStrategy(jsonConfig.IdleStrategy)
	}

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
		// Validate file path for security
		if err := validateFilePath(output); err != nil {
			return &config, fmt.Errorf("invalid output file path: %w", err)
		}
		// Assume it's a file path
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600) // #nosec G304 -- Path validation implemented above
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

	// IdleStrategy from IRIS_IDLE_STRATEGY
	if strategyStr := os.Getenv("IRIS_IDLE_STRATEGY"); strategyStr != "" {
		config.IdleStrategy = parseIdleStrategy(strategyStr)
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
			if jsonConfig.IdleStrategy != nil {
				config.IdleStrategy = jsonConfig.IdleStrategy
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
	if strategyStr := os.Getenv("IRIS_IDLE_STRATEGY"); strategyStr != "" {
		config.IdleStrategy = envConfig.IdleStrategy
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

// parseIdleStrategy converts a string to an IdleStrategy
func parseIdleStrategy(strategyStr string) IdleStrategy {
	switch strings.ToLower(strategyStr) {
	case "spinning":
		return NewSpinningIdleStrategy()
	case "sleeping":
		return NewSleepingIdleStrategy(1*time.Millisecond, 0) // Default 1ms sleep, no spin
	case "yielding":
		return NewYieldingIdleStrategy(1000) // Default 1000 spins before yield
	case "channel":
		return NewChannelIdleStrategy(100 * time.Millisecond) // Default 100ms timeout
	case "progressive", "balanced":
		return NewProgressiveIdleStrategy()
	case "efficient":
		return EfficientStrategy
	case "hybrid":
		return HybridStrategy
	default:
		return BalancedStrategy // Default strategy
	}
}

// DynamicConfigWatcher manages dynamic configuration changes using Argus
// Provides real-time hot reload of Iris logger configuration with audit trail
type DynamicConfigWatcher struct {
	configPath  string
	atomicLevel *AtomicLevel
	watcher     *argus.Watcher
	enabled     int32      // Use atomic int32 instead of bool for thread safety
	mu          sync.Mutex // Protect start/stop operations
}

// NewDynamicConfigWatcher creates a new dynamic config watcher for iris logger
// This enables runtime log level changes by watching the configuration file
//
// Parameters:
//   - configPath: Path to the JSON configuration file to watch
//   - atomicLevel: The atomic level instance from iris logger
//
// Example usage:
//
//	logger, err := iris.New(config)
//	if err != nil {
//	    return err
//	}
//
//	watcher, err := iris.NewDynamicConfigWatcher("config.json", logger.Level())
//	if err != nil {
//	    return err
//	}
//	defer watcher.Stop()
//
//	if err := watcher.Start(); err != nil {
//	    return err
//	}
//
// Now when you modify config.json and change the "level" field,
// the logger will automatically update its level without restart!
func NewDynamicConfigWatcher(configPath string, atomicLevel *AtomicLevel) (*DynamicConfigWatcher, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("config file does not exist: %w", err)
	}

	// Create Argus watcher with production-ready configuration
	config := argus.Config{
		PollInterval:         2 * time.Second, // Fast response for dev, efficient for prod
		OptimizationStrategy: argus.OptimizationAuto,

		// Enable audit trail for configuration changes
		Audit: argus.AuditConfig{
			Enabled:       true,
			OutputFile:    "iris-config-audit.jsonl",
			MinLevel:      argus.AuditInfo, // Capture all config changes
			BufferSize:    1000,
			FlushInterval: 5 * time.Second, // Faster flush for testing
		},

		// Error handling for config watcher
		ErrorHandler: func(err error, path string) {
			// Log errors through our own error system
			loggerErr := NewLoggerError(ErrCodeFileOpen,
				fmt.Sprintf("Config watcher error for %s: %v", path, err))
			GetErrorHandler()(loggerErr)
		},
	}

	watcher := argus.New(*config.WithDefaults())

	return &DynamicConfigWatcher{
		configPath:  configPath,
		atomicLevel: atomicLevel,
		watcher:     watcher,
		enabled:     0, // 0 = false, 1 = true for atomic int32
	}, nil
}

// Start begins watching the configuration file for changes
func (w *DynamicConfigWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if atomic.LoadInt32(&w.enabled) != 0 {
		return fmt.Errorf("watcher is already started")
	}

	// Load initial configuration
	if w.atomicLevel != nil {
		initialConfig, err := LoadConfigFromJSON(w.configPath)
		if err == nil {
			w.atomicLevel.SetLevel(initialConfig.Level)
		}
		// Don't fail on initial load error - just continue with current level
	}

	// Set up config file watcher with hot reload callback
	if err := w.watcher.Watch(w.configPath, func(event argus.ChangeEvent) {
		// Load and parse the updated configuration
		newConfig, err := LoadConfigFromJSON(event.Path)
		if err != nil {
			loggerErr := NewLoggerError(ErrCodeInvalidConfig,
				fmt.Sprintf("Failed to reload config from %s: %v", event.Path, err))
			GetErrorHandler()(loggerErr)
			return
		}

		// Update the atomic level with the new configuration
		if w.atomicLevel != nil {
			w.atomicLevel.SetLevel(newConfig.Level)
		}

		// Log successful config reload (using our own logger would create a loop!)
		// Instead we write to stderr for safety
		fmt.Fprintf(os.Stderr, "[IRIS] Configuration reloaded from %s - Level: %s\n",
			event.Path, newConfig.Level.String())
	}); err != nil {
		return fmt.Errorf("failed to setup file watcher: %w", err)
	}

	// Start the Argus watcher
	if err := w.watcher.Start(); err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}
	atomic.StoreInt32(&w.enabled, 1) // Set to 1 (true)
	return nil
}

// Stop stops watching the configuration file
func (w *DynamicConfigWatcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if atomic.LoadInt32(&w.enabled) == 0 {
		return fmt.Errorf("watcher is not started")
	}

	// Stop the Argus watcher
	if err := w.watcher.Stop(); err != nil {
		return fmt.Errorf("failed to stop file watcher: %w", err)
	}
	atomic.StoreInt32(&w.enabled, 0) // Set to 0 (false)
	return nil
}

// IsRunning returns true if the watcher is currently active
func (w *DynamicConfigWatcher) IsRunning() bool {
	return atomic.LoadInt32(&w.enabled) != 0
}

// EnableDynamicLevel creates and starts a config watcher for the given logger and config file
// This is a convenience function that combines NewDynamicConfigWatcher + Start
//
// Example:
//
//	logger, err := iris.New(config)
//	if err != nil {
//	    return err
//	}
//
//	watcher, err := iris.EnableDynamicLevel(logger, "config.json")
//	if err != nil {
//	    log.Printf("Dynamic level disabled: %v", err)
//	} else {
//	    defer watcher.Stop()
//	    log.Println("✅ Dynamic level changes enabled!")
//	}
func EnableDynamicLevel(logger *Logger, configPath string) (*DynamicConfigWatcher, error) {
	watcher, err := NewDynamicConfigWatcher(configPath, &logger.level)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic config watcher: %w", err)
	}

	if err := watcher.Start(); err != nil {
		return nil, fmt.Errorf("failed to start dynamic config watcher: %w", err)
	}

	return watcher, nil
}
