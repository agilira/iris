package iris

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestDynamicConfigWatcher tests the dynamic config watcher API
func TestDynamicConfigWatcher(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	initialConfig := map[string]interface{}{
		"level":      "info",
		"format":     "json",
		"capacity":   65536,
		"batch_size": 32,
	}

	configData, err := json.Marshal(initialConfig)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create logger with initial config
	logger, err := New(Config{
		Level:    Info,
		Encoder:  NewTextEncoder(),
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test accessing atomic level
	atomicLevel := logger.AtomicLevel()
	if atomicLevel == nil {
		t.Fatal("AtomicLevel() returned nil")
	}

	// Test initial level
	if logger.Level() != Info {
		t.Errorf("Expected initial level Info, got %v", logger.Level())
	}

	// Test dynamic level change via atomic level
	atomicLevel.SetLevel(Debug)
	if logger.Level() != Debug {
		t.Errorf("Expected level Debug after SetLevel, got %v", logger.Level())
	}

	// Test watcher creation (placeholder implementation for now)
	watcher, err := NewDynamicConfigWatcher(configPath, atomicLevel)
	if err != nil {
		t.Fatalf("Failed to create dynamic config watcher: %v", err)
	}

	// Test watcher state
	if watcher.IsRunning() {
		t.Error("Watcher should not be running initially")
	}

	// Test starting watcher
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	if !watcher.IsRunning() {
		t.Error("Watcher should be running after Start()")
	}

	// Test stopping watcher
	if err := watcher.Stop(); err != nil {
		t.Fatalf("Failed to stop watcher: %v", err)
	}

	if watcher.IsRunning() {
		t.Error("Watcher should not be running after Stop()")
	}
}

// TestEnableDynamicLevel tests the convenience function
func TestEnableDynamicLevel(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dynamic_test.json")

	config := map[string]interface{}{
		"level": "warn",
	}

	configData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create logger
	logger, err := New(Config{
		Level:   Info,
		Encoder: NewTextEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test convenience function
	watcher, err := EnableDynamicLevel(logger, configPath)
	if err != nil {
		t.Fatalf("Failed to enable dynamic level: %v", err)
	}
	defer watcher.Stop()

	if !watcher.IsRunning() {
		t.Error("Watcher should be running after EnableDynamicLevel()")
	}
}

// TestAtomicLevelIntegration tests that AtomicLevel integrates properly with logger
func TestAtomicLevelIntegration(t *testing.T) {
	logger, err := New(Config{
		Level:   Error,
		Encoder: NewTextEncoder(),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test that AtomicLevel() returns a working reference
	atomicLevel := logger.AtomicLevel()

	// Change level via atomic level
	atomicLevel.SetLevel(Debug)

	// Verify change is reflected in logger
	if logger.Level() != Debug {
		t.Errorf("Expected logger level Debug, got %v", logger.Level())
	}

	// Change level via logger
	logger.SetLevel(Warn)

	// Verify change is reflected in atomic level
	if atomicLevel.Level() != Warn {
		t.Errorf("Expected atomic level Warn, got %v", atomicLevel.Level())
	}

	// Test that they reference the same underlying atomic value
	logger.SetLevel(Fatal)
	if atomicLevel.Level() != Fatal {
		t.Error("AtomicLevel and Logger level are not synchronized")
	}
}
