package main

import (
	"testing"

	"github.com/agilira/iris"
	"github.com/agilira/iris/internal/zephyroslite"
)

func TestConfigLoaderBackpressurePolicy(t *testing.T) {
	tests := []struct {
		name           string
		configFile     string
		expectedPolicy zephyroslite.BackpressurePolicy
	}{
		{
			name:           "High Performance Config",
			configFile:     "../configs/high_performance.json",
			expectedPolicy: zephyroslite.DropOnFull,
		},
		{
			name:           "Reliable Config",
			configFile:     "../configs/reliable.json",
			expectedPolicy: zephyroslite.BlockOnFull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := iris.LoadConfigFromJSON(tt.configFile)
			if err != nil {
				t.Fatalf("LoadConfigFromJSON() error = %v", err)
			}

			if config.BackpressurePolicy != tt.expectedPolicy {
				t.Errorf("BackpressurePolicy = %v, want %v",
					config.BackpressurePolicy, tt.expectedPolicy)
			}

			// Test that we can create a logger with this config
			logger, err := iris.New(*config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			// Test logging functionality
			logger.Info("Test message", iris.String("test", "config_loader"))
			logger.Sync()
		})
	}
}

func TestConfigLoaderEnvironmentOverride(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedPolicy zephyroslite.BackpressurePolicy
	}{
		{"DropOnFull - drop", "drop", zephyroslite.DropOnFull},
		{"DropOnFull - drop_on_full", "drop_on_full", zephyroslite.DropOnFull},
		{"DropOnFull - droponful", "droponful", zephyroslite.DropOnFull},
		{"BlockOnFull - block", "block", zephyroslite.BlockOnFull},
		{"BlockOnFull - block_on_full", "block_on_full", zephyroslite.BlockOnFull},
		{"BlockOnFull - blockonful", "blockonful", zephyroslite.BlockOnFull},
		{"Default - invalid", "invalid", zephyroslite.DropOnFull},
		{"Default - empty", "", zephyroslite.DropOnFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			t.Setenv("IRIS_BACKPRESSURE_POLICY", tt.envValue)

			config, err := iris.LoadConfigFromEnv()
			if err != nil {
				t.Fatalf("LoadConfigFromEnv() error = %v", err)
			}

			if config.BackpressurePolicy != tt.expectedPolicy {
				t.Errorf("BackpressurePolicy = %v, want %v",
					config.BackpressurePolicy, tt.expectedPolicy)
			}
		})
	}
}

func TestBackpressurePolicyString(t *testing.T) {
	tests := []struct {
		policy   zephyroslite.BackpressurePolicy
		expected string
	}{
		{zephyroslite.DropOnFull, "DropOnFull"},
		{zephyroslite.BlockOnFull, "BlockOnFull"},
		{zephyroslite.BackpressurePolicy(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.policy.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
