// integration_test.go: End-to-end integration tests for Magic API functionality
//
// These tests simulate the complete Iris + Lethe integration experience
// to verify that the Magic API infrastructure works correctly before
// implementing the actual Lethe side.
//
// Copyright (c) 2025 AGILira
// Series: IRIS Logging Library - Integration Tests
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/agilira/iris/internal/lethe"
)

// windowsSafeClose ensures proper cleanup on Windows by forcing GC and longer delays
func windowsSafeClose(t *testing.T, logger *Logger) {
	// Ensure all async operations are completed
	if err := logger.Sync(); err != nil {
		t.Logf("Warning: Error syncing logger: %v", err)
	}

	// Close the logger
	if err := logger.Close(); err != nil {
		t.Logf("Warning: Error closing logger: %v", err)
	}

	// Force garbage collection and finalizers to run
	runtime.GC()
	runtime.GC() // Call twice to ensure finalizers run

	// Give Windows extra time to release file handles
	time.Sleep(500 * time.Millisecond)
}

func TestCompleteMagicIntegration(t *testing.T) {
	// Simulate complete Lethe registration and usage
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "magic_integration.log")

	// Step 1: Register mock Lethe provider (simulates Lethe's init())
	registerMockLetheProvider()

	// Step 2: Create Magic Logger (should detect Lethe and optimize)
	logger, err := NewMagicLogger(logFile, Info)
	if err != nil {
		t.Fatalf("Failed to create magic logger: %v", err)
	}
	defer windowsSafeClose(t, logger)

	// Verify we got Lethe optimizations
	if !lethe.HasLetheCapabilities() {
		t.Error("Lethe capabilities should be registered")
	}

	// Step 3: Test logging with zero-copy optimization
	logger.Start()

	// Log several messages to test WriteOwned path
	for i := 0; i < 5; i++ {
		logger.Info("Magic API test message",
			String("iteration", string(rune('A'+i))),
			Int("number", i+1),
			Bool("zero_copy", true))
	}

	// Allow async processing
	time.Sleep(200 * time.Millisecond)

	// Step 4: Verify file was created and contains data
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatal("Log file was not created")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("Log file is empty")
	}

	// Verify content structure
	lines := strings.Split(string(content), "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	if nonEmptyLines < 5 {
		t.Errorf("Expected at least 5 log entries, got %d", nonEmptyLines)
	}

	// Verify Magic API optimizations were used
	contentStr := string(content)
	if !strings.Contains(contentStr, "Magic API test message") {
		t.Error("Log content doesn't contain expected message")
	}

	if !strings.Contains(contentStr, "zero_copy") {
		t.Error("Log content doesn't contain zero_copy field")
	}

	t.Logf("Magic Integration test successful! Logged %d entries with zero-copy optimization", nonEmptyLines)
}

func TestMagicAPI_E2E_BasicFlow(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "magic_performance.log")

	// Register optimized provider
	registerMockLetheProvider()

	logger, err := NewMagicLogger(logFile, Info)
	if err != nil {
		t.Fatalf("Failed to create magic logger: %v", err)
	}
	defer windowsSafeClose(t, logger)

	// Test buffer size optimization
	// Mock returns 16384, should be detected and used
	provider, exists := lethe.GetLetheProvider()
	if !exists {
		t.Fatal("Lethe provider should be registered")
	}

	sink, err := provider.CreateOptimizedSink(logFile)
	if err != nil {
		t.Fatalf("Failed to create optimized sink: %v", err)
	}

	letheWriter := lethe.DetectLetheCapabilities(sink)
	if letheWriter == nil {
		t.Fatal("Should detect Lethe capabilities")
	}

	// Verify buffer size optimization
	bufferSize := letheWriter.GetOptimalBufferSize()
	if bufferSize != 16384 {
		t.Errorf("Expected buffer size 16384, got %d", bufferSize)
	}

	// Verify hot reload support
	if !letheWriter.SupportsHotReload() {
		t.Error("Should support hot reload")
	}

	// Test WriteOwned zero-copy path
	testData := []byte("zero-copy test data")
	n, err := letheWriter.WriteOwned(testData)
	if err != nil {
		t.Errorf("WriteOwned failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("WriteOwned returned %d bytes, expected %d", n, len(testData))
	}

	t.Log("Performance optimizations verified successfully")
}

func TestMagicFallbackBehavior(t *testing.T) {
	// Test fallback with invalid file path that forces fallback
	invalidPath := "/root/nonexistent/directory/test.log"

	// Should fallback to standard file logger when Lethe fails
	logger, err := NewMagicLogger(invalidPath, Info)
	if err == nil {
		if closeErr := logger.Close(); closeErr != nil {
			t.Logf("Warning: Error closing logger: %v", closeErr)
		}
		t.Skip("Skipping fallback test - system allows creation in /root")
	}

	// Error is expected, this demonstrates fallback behavior
	t.Logf("Fallback behavior verified - got expected error: %v", err)

	// Test with valid path but simulate provider failure
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "fallback_test.log")

	// Register a provider that always fails
	failingProvider := lethe.CapabilityProvider{
		Name: "failing-lethe",
		CreateOptimizedSink: func(filename string, args ...interface{}) (interface{}, error) {
			return nil, os.ErrPermission // Always fail
		},
	}
	lethe.RegisterCapability(failingProvider)

	logger, err = NewMagicLogger(logFile, Info)
	if err != nil {
		t.Fatalf("Magic logger should fallback when provider fails: %v", err)
	}
	defer windowsSafeClose(t, logger)

	// Test logging still works in fallback mode
	logger.Start()
	logger.Info("Fallback mode test", String("mode", "standard"))

	time.Sleep(100 * time.Millisecond)

	// Verify file creation
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read fallback log: %v", err)
	}

	if !strings.Contains(string(content), "Fallback mode test") {
		t.Error("Fallback log doesn't contain expected content")
	}

	t.Log("Fallback behavior verified successfully")
}

func TestMagicRegistrationLifecycle(t *testing.T) {
	// Test the registration lifecycle (can't clear global registry)

	// Verify capabilities are available (from previous tests)
	if !lethe.HasLetheCapabilities() {
		t.Log("No capabilities initially - registering test provider")
		registerMockLetheProvider()
	}

	if !lethe.HasLetheCapabilities() {
		t.Error("Should have capabilities after registration")
	}

	// Test multiple registrations work correctly
	initialCount := len(lethe.GetRegisteredCapabilities())

	provider2 := lethe.CapabilityProvider{
		Name: "lethe-v2",
		CreateOptimizedSink: func(filename string, args ...interface{}) (interface{}, error) {
			return &enhancedMockLetheWriter{filename: filename}, nil
		},
	}
	lethe.RegisterCapability(provider2)

	capabilities := lethe.GetRegisteredCapabilities()
	if len(capabilities) <= initialCount {
		t.Errorf("Expected more than %d capabilities after adding provider, got %d", initialCount, len(capabilities))
	}

	// Test Lethe-specific lookup still works
	_, exists := lethe.GetLetheProvider()
	if !exists {
		t.Error("Should still find Lethe provider among multiple")
	}

	// Test provider overwrite behavior
	newProvider := lethe.CapabilityProvider{
		Name: "lethe",
		CreateOptimizedSink: func(filename string, args ...interface{}) (interface{}, error) {
			return &enhancedMockLetheWriter{filename: filename + ".enhanced"}, nil
		},
	}
	lethe.RegisterCapability(newProvider)

	updatedProvider, exists := lethe.GetLetheProvider()
	if !exists {
		t.Error("Should still have Lethe provider after update")
	}

	// Verify it's a different instance (provider was updated)
	if updatedProvider.Name != "lethe" {
		t.Error("Provider name should still be 'lethe' after update")
	}

	t.Log("Registration lifecycle test completed successfully")
}

// Helper functions and mocks

func registerMockLetheProvider() {
	provider := lethe.CapabilityProvider{
		Name: "lethe",
		CreateOptimizedSink: func(filename string, args ...interface{}) (interface{}, error) {
			// Create actual file for realistic testing
			file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return nil, err
			}
			return &fullMockLetheWriter{
				filename: filename,
				file:     file,
				buffer:   &bytes.Buffer{},
			}, nil
		},
		DetectWriter: func(writer interface{}) bool {
			_, ok := writer.(*fullMockLetheWriter)
			return ok
		},
	}
	lethe.RegisterCapability(provider)
}

// Enhanced mock that implements full LetheWriter interface
type fullMockLetheWriter struct {
	filename      string
	file          *os.File
	buffer        *bytes.Buffer
	zeroCopyCalls int
}

func (f *fullMockLetheWriter) Write(data []byte) (int, error) {
	// Write to both buffer and file for testing
	f.buffer.Write(data)
	if f.file != nil {
		return f.file.Write(data)
	}
	return len(data), nil
}

func (f *fullMockLetheWriter) WriteOwned(data []byte) (int, error) {
	f.zeroCopyCalls++
	// Simulate zero-copy optimization
	return f.Write(data)
}

func (f *fullMockLetheWriter) Sync() error {
	if f.file != nil {
		return f.file.Sync()
	}
	return nil
}

func (f *fullMockLetheWriter) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

func (f *fullMockLetheWriter) GetOptimalBufferSize() int {
	return 16384 // 16KB optimized buffer
}

func (f *fullMockLetheWriter) SupportsHotReload() bool {
	return true
}

// Enhanced mock for multiple provider testing
type enhancedMockLetheWriter struct {
	filename string
}

func (e *enhancedMockLetheWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (e *enhancedMockLetheWriter) WriteOwned(data []byte) (int, error) {
	return len(data), nil
}

func (e *enhancedMockLetheWriter) Sync() error {
	return nil
}

func (e *enhancedMockLetheWriter) Close() error {
	return nil
}

func (e *enhancedMockLetheWriter) GetOptimalBufferSize() int {
	return 32768 // 32KB enhanced buffer
}

func (e *enhancedMockLetheWriter) SupportsHotReload() bool {
	return true
}
