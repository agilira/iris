// structured_logger_unit_test.go: Unit tests for structured logger
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
	"time"
)

// mockStructuredWriter for testing
type mockStructuredWriter struct {
	data   []byte
	writes int
}

func (msw *mockStructuredWriter) Write(p []byte) (n int, err error) {
	msw.writes++
	msw.data = append(msw.data, p...)
	return len(p), nil
}

func (msw *mockStructuredWriter) String() string {
	return string(msw.data)
}

// TestStructuredLoggerBasic tests basic structured logging functionality
func TestStructuredLoggerBasic(t *testing.T) {
	mock := &mockStructuredWriter{}

	config := Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(mock),
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  1,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test structured logging
	entry := logger.WithFieldsStructured(
		String("service", "test"),
		Int("port", 8080),
		Bool("debug", true),
	)

	entry.Info("test message")

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	if mock.writes == 0 {
		t.Error("Expected at least one write to mock writer")
	}

	output := mock.String()
	if output == "" {
		t.Error("Expected non-empty output")
	}

	// Check that structured fields are present in output
	if !contains(output, "service") {
		t.Error("Expected 'service' field in output")
	}
	if !contains(output, "test") {
		t.Error("Expected 'test' value in output")
	}
	if !contains(output, "port") {
		t.Error("Expected 'port' field in output")
	}
	if !contains(output, "8080") {
		t.Error("Expected '8080' value in output")
	}
}

// TestStructuredLoggerWithMultipleFields tests multiple field types
func TestStructuredLoggerWithMultipleFields(t *testing.T) {
	mock := &mockStructuredWriter{}

	config := Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(mock),
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  1,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test with various field types
	entry := logger.WithFieldsStructured(
		String("name", "john"),
		String("city", "new york"),
		Int("age", 30),
		Int("score", 95),
		Bool("active", true),
		Bool("verified", false),
	)

	entry.Info("user info")

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	output := mock.String()

	// Check all fields are present
	expectedFields := []string{"name", "john", "city", "new york", "age", "30", "score", "95", "active", "true", "verified", "false"}
	for _, expected := range expectedFields {
		if !contains(output, expected) {
			t.Errorf("Expected '%s' in output: %s", expected, output)
		}
	}
}

// TestStructuredLoggerMemoryFootprint tests memory footprint functionality
func TestStructuredLoggerMemoryFootprint(t *testing.T) {
	mock := &mockStructuredWriter{}

	config := Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(mock),
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  1,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Create structured entry
	entry := logger.WithFieldsStructured(
		String("test", "value"),
		Int("number", 42),
	)

	// Test memory footprint
	footprint := entry.MemoryFootprint()
	if footprint <= 0 {
		t.Errorf("Expected positive memory footprint, got %d", footprint)
	}
}

// TestStructuredLoggerLevelFiltering tests level filtering
func TestStructuredLoggerLevelFiltering(t *testing.T) {
	mock := &mockStructuredWriter{}

	config := Config{
		Level:      WarnLevel, // Only warn and above
		Writer:     NewConsoleWriter(mock),
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  1,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Info should be filtered out
	entry := logger.WithFieldsStructured(String("test", "filtered"))
	entry.Info("this should be filtered")

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	if mock.writes > 0 {
		t.Error("Expected no writes for filtered log level")
	}
}

// TestStructuredLoggerEmpty tests empty field handling
func TestStructuredLoggerEmpty(t *testing.T) {
	mock := &mockStructuredWriter{}

	config := Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(mock),
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  1,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test with no fields
	entry := logger.WithFieldsStructured()
	entry.Info("message without fields")

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	if mock.writes == 0 {
		t.Error("Expected at least one write even with no fields")
	}

	output := mock.String()
	if !contains(output, "message without fields") {
		t.Error("Expected message in output")
	}
}

// TestStructuredLoggerReset tests encoder reset functionality
func TestStructuredLoggerReset(t *testing.T) {
	mock := &mockStructuredWriter{}

	config := Config{
		Level:      InfoLevel,
		Writer:     NewConsoleWriter(mock),
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  1,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Create entry and log
	entry := logger.WithFieldsStructured(String("test", "first"))
	entry.Info("first message")

	// Memory footprint should be available before reset
	footprintBefore := entry.MemoryFootprint()

	// Give time for processing (reset happens in Info())
	time.Sleep(10 * time.Millisecond)

	// After logging, encoder should be reset
	footprintAfter := entry.MemoryFootprint()

	// Reset encoder typically should have lower or same footprint
	if footprintAfter > footprintBefore {
		t.Errorf("Expected footprint after reset (%d) to be <= before (%d)", footprintAfter, footprintBefore)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
