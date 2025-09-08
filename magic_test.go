// magic_test.go: Tests for Magic API functionality
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agilira/iris/internal/lethe"
)

func TestNewMagicLogger_FallbackMode(t *testing.T) {
	// Test fallback mode when no Lethe capabilities are registered
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test_fallback.log")

	// Ensure no Lethe capabilities are registered
	if lethe.HasLetheCapabilities() {
		t.Skip("Skipping fallback test because Lethe capabilities are registered")
	}

	logger, err := NewMagicLogger(logFile, Info)
	if err != nil {
		t.Fatalf("Failed to create magic logger in fallback mode: %v", err)
	}
	defer windowsSafeClose(t, logger)

	// Test basic logging
	logger.Start()
	logger.Info("Test message", String("mode", "fallback"))

	// Allow time for async write
	time.Sleep(100 * time.Millisecond)

	// Verify file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file was not created: %s", logFile)
	}

	// Read and verify content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Log file is empty")
	}

	t.Logf("Fallback mode test successful. Log content: %s", string(content))
}

func TestNewMagicLogger_WithLetheCapabilities(t *testing.T) {
	// Mock Lethe capability registration
	mockProvider := lethe.CapabilityProvider{
		Name: "lethe",
		CreateOptimizedSink: func(filename string, args ...interface{}) (interface{}, error) {
			// Mock sink that implements LetheWriter interface
			// For invalid paths, return error like real Lethe would
			if filepath.Dir(filename) == "/root/invalid/path/that/should/not/exist" {
				return nil, os.ErrPermission
			}
			return &mockLetheWriter{filename: filename}, nil
		},
		DetectWriter: func(writer interface{}) bool {
			_, ok := writer.(*mockLetheWriter)
			return ok
		},
	}

	// Register mock capability
	lethe.RegisterCapability(mockProvider)
	defer func() {
		// Note: In real implementation, we might want a way to unregister for testing
		// For now, we'll just verify the test works with the capability present
	}()

	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test_lethe.log")

	logger, err := NewMagicLogger(logFile, Info)
	if err != nil {
		t.Fatalf("Failed to create magic logger with Lethe capabilities: %v", err)
	}
	defer windowsSafeClose(t, logger)

	// Test logging with Lethe optimizations
	logger.Start()
	logger.Info("Test message", String("mode", "lethe-optimized"))

	// Allow time for async write
	time.Sleep(100 * time.Millisecond)

	t.Logf("Lethe-optimized mode test successful")
}

func TestMagicLogger_InvalidFile(t *testing.T) {
	// Test error handling with dangerous path patterns
	// Force standard file logger path by temporarily clearing Lethe capabilities
	oldCapabilities := lethe.HasLetheCapabilities()
	if oldCapabilities {
		// We need to test the standard file logger validation
		// so we temporarily disable Lethe for this test
		t.Skip("Skipping when Lethe is active - validation happens in standard file logger only")
	}

	testCases := []string{
		"/root/invalid/path/test.log",     // Unix dangerous path
		"../../../etc/passwd",             // Directory traversal
		"C:\\Windows\\System32\\test.log", // Windows dangerous path
		"\\root\\invalid\\path\\test.log", // Windows root path
	}

	for _, invalidPath := range testCases {
		t.Run(fmt.Sprintf("Path_%s", invalidPath), func(t *testing.T) {
			_, err := NewMagicLogger(invalidPath, Info)
			if err == nil {
				t.Errorf("Expected error for dangerous file path %s, but got none", invalidPath)
			} else {
				t.Logf("Correctly rejected dangerous path %s: %v", invalidPath, err)
			}
		})
	}
}

// mockLetheWriter implements the LetheWriter interface for testing
type mockLetheWriter struct {
	filename string
	data     []byte
}

func (m *mockLetheWriter) Write(data []byte) (int, error) {
	m.data = append(m.data, data...)
	return len(data), nil
}

func (m *mockLetheWriter) WriteOwned(data []byte) (int, error) {
	// Mock zero-copy optimization
	return m.Write(data)
}

func (m *mockLetheWriter) Sync() error {
	return nil
}

func (m *mockLetheWriter) Close() error {
	return nil
}

func (m *mockLetheWriter) GetOptimalBufferSize() int {
	return 16384 // Mock 16KB optimal size
}

func (m *mockLetheWriter) SupportsHotReload() bool {
	return true
}

func TestLetheDetection(t *testing.T) {
	// Test Lethe capability detection
	mock := &mockLetheWriter{filename: "test.log"}

	// Test detection
	detected := lethe.DetectLetheCapabilities(mock)
	if detected == nil {
		t.Error("Failed to detect Lethe capabilities in mock writer")
	}

	if !lethe.IsLetheWriter(mock) {
		t.Error("IsLetheWriter should return true for mock writer")
	}

	// Test non-Lethe writer
	normalFile, _ := os.CreateTemp("", "test")
	defer func() {
		if err := os.Remove(normalFile.Name()); err != nil {
			t.Logf("Warning: Error removing temp file: %v", err)
		}
	}()
	defer func() {
		if err := normalFile.Close(); err != nil {
			t.Logf("Warning: Error closing temp file: %v", err)
		}
	}()

	if lethe.IsLetheWriter(normalFile) {
		t.Error("IsLetheWriter should return false for normal file")
	}

	t.Log("Lethe detection tests successful")
}
