// autoscaling_unit_test.go: Comprehensive tests for autoscaling functionality
//
// This file provides complete test coverage for the AutoScalingLogger
// to ensure proper scaling behavior and performance monitoring.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/agilira/iris/internal/zephyroslite"
)

// Helper function to safely close logger ignoring stdout sync errors
func safeCloseLogger(t *testing.T, logger *AutoScalingLogger) {
	if err := logger.Close(); err != nil && !strings.Contains(err.Error(), "sync /dev/stdout: invalid argument") {
		t.Errorf("Failed to close logger: %v", err)
	}
}

func TestAutoScalingMode_String(t *testing.T) {
	tests := []struct {
		mode AutoScalingMode
		want string
	}{
		{SingleRingMode, "SingleRing"},
		{MPSCMode, "MPSC"},
		{AutoScalingMode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("AutoScalingMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultAutoScalingConfig(t *testing.T) {
	config := DefaultAutoScalingConfig()

	// Verify default values
	if config.ScaleToMPSCWriteThreshold != 1000 {
		t.Errorf("Expected ScaleToMPSCWriteThreshold 1000, got %d", config.ScaleToMPSCWriteThreshold)
	}
	if config.ScaleToSingleWriteThreshold != 100 {
		t.Errorf("Expected ScaleToSingleWriteThreshold 100, got %d", config.ScaleToSingleWriteThreshold)
	}
	if config.MeasurementWindow != 100*time.Millisecond {
		t.Errorf("Expected MeasurementWindow 100ms, got %v", config.MeasurementWindow)
	}
	if config.ScalingCooldown != 1*time.Second {
		t.Errorf("Expected ScalingCooldown 1s, got %v", config.ScalingCooldown)
	}
	if config.StabilityRequirement != 3 {
		t.Errorf("Expected StabilityRequirement 3, got %d", config.StabilityRequirement)
	}
}

func TestNewAutoScalingLogger(t *testing.T) {
	baseConfig := Config{
		Level:              Info,
		Output:             WrapWriter(os.Stdout),
		Encoder:            NewJSONEncoder(),
		BackpressurePolicy: zephyroslite.DropOnFull,
	}

	scalingConfig := DefaultAutoScalingConfig()

	logger, err := NewAutoScalingLogger(baseConfig, scalingConfig)
	if err != nil {
		t.Fatalf("Failed to create auto-scaling logger: %v", err)
	}
	defer safeCloseLogger(t, logger)

	// Verify initial state
	if logger.GetCurrentMode() != SingleRingMode {
		t.Errorf("Expected initial mode SingleRing, got %v", logger.GetCurrentMode())
	}

	if logger.GetCurrentMode() != SingleRingMode {
		t.Errorf("GetCurrentMode() = %v, want %v", logger.GetCurrentMode(), SingleRingMode)
	}
}

func TestAutoScalingLogger_StartStop(t *testing.T) {
	baseConfig := Config{
		Level:              Info,
		Output:             WrapWriter(os.Stdout),
		Encoder:            NewJSONEncoder(),
		BackpressurePolicy: zephyroslite.DropOnFull,
	}

	scalingConfig := DefaultAutoScalingConfig()
	scalingConfig.MeasurementWindow = 50 * time.Millisecond // Fast for testing

	logger, err := NewAutoScalingLogger(baseConfig, scalingConfig)
	if err != nil {
		t.Fatalf("Failed to create auto-scaling logger: %v", err)
	}

	// Start the logger
	err = logger.Start()
	if err != nil {
		t.Fatalf("Failed to start logger: %v", err)
	}

	// Give it time to start the scaling loop
	time.Sleep(100 * time.Millisecond)

	// Close the logger
	safeCloseLogger(t, logger)

	// Verify it can be called multiple times
	safeCloseLogger(t, logger)
}

func TestAutoScalingLogger_LoggingMethods(t *testing.T) {
	baseConfig := Config{
		Level:              Debug,
		Output:             WrapWriter(os.Stdout),
		Encoder:            NewJSONEncoder(),
		BackpressurePolicy: zephyroslite.DropOnFull,
	}

	scalingConfig := DefaultAutoScalingConfig()

	logger, err := NewAutoScalingLogger(baseConfig, scalingConfig)
	if err != nil {
		t.Fatalf("Failed to create auto-scaling logger: %v", err)
	}
	defer safeCloseLogger(t, logger)

	err = logger.Start()
	if err != nil {
		t.Fatalf("Failed to start logger: %v", err)
	}

	// Test all logging methods
	logger.Info("test info message")
	logger.Debug("test debug message")
	logger.Warn("test warn message")
	logger.Error("test error message")

	// Give time for metrics to update
	time.Sleep(50 * time.Millisecond)

	// Verify metrics were updated
	stats := logger.GetScalingStats()
	if stats.TotalWrites == 0 {
		t.Error("Expected TotalWrites > 0 after logging")
	}
}

func TestAutoScalingLogger_ScalingBehavior(t *testing.T) {
	baseConfig := Config{
		Level:              Info,
		Output:             WrapWriter(os.Stdout),
		Encoder:            NewJSONEncoder(),
		BackpressurePolicy: zephyroslite.DropOnFull,
	}

	scalingConfig := DefaultAutoScalingConfig()
	scalingConfig.ScaleToMPSCWriteThreshold = 10
	scalingConfig.ScaleToSingleWriteThreshold = 5
	scalingConfig.MeasurementWindow = 100 * time.Millisecond
	scalingConfig.ScalingCooldown = 200 * time.Millisecond

	logger, err := NewAutoScalingLogger(baseConfig, scalingConfig)
	if err != nil {
		t.Fatalf("Failed to create auto-scaling logger: %v", err)
	}
	defer safeCloseLogger(t, logger)

	err = logger.Start()
	if err != nil {
		t.Fatalf("Failed to start logger: %v", err)
	}

	// Generate high load to trigger scaling up
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				logger.Info("high load message")
				time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// Wait for scaling decision
	time.Sleep(300 * time.Millisecond)

	// Check if it scaled up (might or might not happen due to timing)
	stats := logger.GetScalingStats()
	if stats.TotalWrites == 0 {
		t.Error("Expected some logs to be processed")
	}

	// Current mode should be accessible
	mode := logger.GetCurrentMode()
	if mode != SingleRingMode && mode != MPSCMode {
		t.Errorf("Unexpected mode: %v", mode)
	}
}

func TestAutoScalingLogger_GetScalingStats(t *testing.T) {
	baseConfig := Config{
		Level:              Info,
		Output:             WrapWriter(os.Stdout),
		Encoder:            NewJSONEncoder(),
		BackpressurePolicy: zephyroslite.DropOnFull,
	}

	scalingConfig := DefaultAutoScalingConfig()

	logger, err := NewAutoScalingLogger(baseConfig, scalingConfig)
	if err != nil {
		t.Fatalf("Failed to create auto-scaling logger: %v", err)
	}
	defer safeCloseLogger(t, logger)

	if err := logger.Start(); err != nil {
		t.Fatalf("Failed to start logger: %v", err)
	}

	// Log some messages
	for i := 0; i < 5; i++ {
		logger.Info("test message")
	}

	time.Sleep(50 * time.Millisecond)

	stats := logger.GetScalingStats()

	// Verify stats structure
	if stats.CurrentMode != SingleRingMode && stats.CurrentMode != MPSCMode {
		t.Errorf("Unexpected CurrentMode: %v", stats.CurrentMode)
	}

	if stats.TotalWrites < 5 {
		t.Errorf("Expected TotalWrites >= 5, got %d", stats.TotalWrites)
	}

	// TotalScaleOperations and ContentionCount are uint64, so they're always >= 0
	// Just verify they can be accessed without error
	_ = stats.TotalScaleOperations
	_ = stats.ContentionCount
}

func TestAutoScalingLogger_WithFields(t *testing.T) {
	baseConfig := Config{
		Level:              Info,
		Output:             WrapWriter(os.Stdout),
		Encoder:            NewJSONEncoder(),
		BackpressurePolicy: zephyroslite.DropOnFull,
	}

	scalingConfig := DefaultAutoScalingConfig()

	logger, err := NewAutoScalingLogger(baseConfig, scalingConfig)
	if err != nil {
		t.Fatalf("Failed to create auto-scaling logger: %v", err)
	}
	defer safeCloseLogger(t, logger)

	if err := logger.Start(); err != nil {
		t.Fatalf("Failed to start logger: %v", err)
	}

	// Test logging with fields
	logger.Info("test message", Str("key", "value"), Int("number", 42))

	time.Sleep(50 * time.Millisecond)

	stats := logger.GetScalingStats()
	if stats.TotalWrites == 0 {
		t.Error("Expected TotalWrites > 0 after logging with fields")
	}
}

func TestAutoScalingLogger_PerformScaling_DirectTest(t *testing.T) {
	baseConfig := Config{
		Capacity:           1024,
		BatchSize:          8,
		Architecture:       SingleRing,
		BackpressurePolicy: zephyroslite.DropOnFull,
		IdleStrategy:       NewSpinningIdleStrategy(),
		Output:             AddSync(os.Stdout),
		Encoder:            NewTextEncoder(),
		Level:              Info,
	}

	scalingConfig := DefaultAutoScalingConfig()

	logger, err := NewAutoScalingLogger(baseConfig, scalingConfig)
	if err != nil {
		t.Fatalf("Failed to create auto-scaling logger: %v", err)
	}
	defer safeCloseLogger(t, logger)

	// Test same mode (no change)
	initialMode := logger.GetCurrentMode()
	logger.performScaling(initialMode)
	if logger.GetCurrentMode() != initialMode {
		t.Errorf("Expected mode to remain %v after scaling to same mode", initialMode)
	}

	// Test mode change
	targetMode := MPSCMode
	if initialMode == MPSCMode {
		targetMode = SingleRingMode
	}

	logger.performScaling(targetMode)
	if logger.GetCurrentMode() != targetMode {
		t.Errorf("Expected mode to change to %v, got %v", targetMode, logger.GetCurrentMode())
	}

	// Verify stats were updated
	stats := logger.GetScalingStats()
	if stats.TotalScaleOperations == 0 {
		t.Errorf("Expected TotalScaleOperations > 0 after scaling")
	}

	// Test scaling back
	originalMode := SingleRingMode
	if targetMode == SingleRingMode {
		originalMode = MPSCMode
	}

	logger.performScaling(originalMode)
	if logger.GetCurrentMode() != originalMode {
		t.Errorf("Expected mode to change back to %v, got %v", originalMode, logger.GetCurrentMode())
	}

	// Verify stats were updated again
	newStats := logger.GetScalingStats()
	if newStats.TotalScaleOperations <= stats.TotalScaleOperations {
		t.Errorf("Expected TotalScaleOperations to increase after second scaling")
	}
}

func TestAutoScalingLogger_ConcurrentScaling(t *testing.T) {
	baseConfig := Config{
		Capacity:           1024,
		BatchSize:          8,
		Architecture:       SingleRing,
		BackpressurePolicy: zephyroslite.DropOnFull,
		IdleStrategy:       NewSpinningIdleStrategy(),
		Output:             AddSync(os.Stdout),
		Encoder:            NewTextEncoder(),
		Level:              Info,
	}

	scalingConfig := DefaultAutoScalingConfig()

	logger, err := NewAutoScalingLogger(baseConfig, scalingConfig)
	if err != nil {
		t.Fatalf("Failed to create auto-scaling logger: %v", err)
	}
	defer safeCloseLogger(t, logger)

	// Test concurrent scaling operations
	const numGoroutines = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Alternate between modes
			if idx%2 == 0 {
				logger.performScaling(MPSCMode)
			} else {
				logger.performScaling(SingleRingMode)
			}
		}(i)
	}

	wg.Wait()

	// Should be in a valid mode
	currentMode := logger.GetCurrentMode()
	if currentMode != SingleRingMode && currentMode != MPSCMode {
		t.Errorf("Expected valid mode after concurrent scaling, got %v", currentMode)
	}

	// Should have performed some scaling operations
	stats := logger.GetScalingStats()
	if stats.TotalScaleOperations == 0 {
		t.Error("Expected some scaling operations after concurrent test")
	}
}
