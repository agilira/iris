// writer_unit_test.go: Unit tests for writer components
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockWriter for testing
type mockWriter struct {
	data    []byte
	writes  int
	failAt  int // Fail after N writes (0 = never fail)
	lastErr error
}

func (mw *mockWriter) Write(p []byte) (n int, err error) {
	mw.writes++
	if mw.failAt > 0 && mw.writes >= mw.failAt {
		mw.lastErr = io.ErrShortWrite
		return 0, mw.lastErr
	}
	mw.data = append(mw.data, p...)
	return len(p), nil
}

func (mw *mockWriter) String() string {
	return string(mw.data)
}

// TestConsoleWriter tests basic console writer functionality
func TestConsoleWriter(t *testing.T) {
	mock := &mockWriter{}
	writer := NewConsoleWriter(mock)

	// Test basic write
	data := []byte("test message\n")
	n, err := writer.Write(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	if mock.String() != "test message\n" {
		t.Errorf("Expected 'test message\\n', got '%s'", mock.String())
	}

	if mock.writes != 1 {
		t.Errorf("Expected 1 write, got %d", mock.writes)
	}
}

// TestConsoleWriterMultipleWrites tests multiple writes
func TestConsoleWriterMultipleWrites(t *testing.T) {
	mock := &mockWriter{}
	writer := NewConsoleWriter(mock)

	writes := []string{"first\n", "second\n", "third\n"}

	for _, write := range writes {
		_, err := writer.Write([]byte(write))
		if err != nil {
			t.Fatalf("Unexpected error on write: %v", err)
		}
	}

	expected := "first\nsecond\nthird\n"
	if mock.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, mock.String())
	}

	if mock.writes != 3 {
		t.Errorf("Expected 3 writes, got %d", mock.writes)
	}
}

// TestConsoleWriterErrorHandling tests error propagation
func TestConsoleWriterErrorHandling(t *testing.T) {
	mock := &mockWriter{failAt: 1}
	writer := NewConsoleWriter(mock)

	// First write should fail
	_, err := writer.Write([]byte("test"))
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err != io.ErrShortWrite {
		t.Errorf("Expected ErrShortWrite, got %v", err)
	}
}

// TestPredefinedWriters tests stdout/stderr writers
func TestPredefinedWriters(t *testing.T) {
	// Test that predefined writers are properly initialized
	if StdoutWriter == nil {
		t.Error("StdoutWriter should not be nil")
	}

	if StderrWriter == nil {
		t.Error("StderrWriter should not be nil")
	}

	// Test that they're different instances
	if StdoutWriter == StderrWriter {
		t.Error("StdoutWriter and StderrWriter should be different instances")
	}
}

// TestFileWriter tests file writer functionality
func TestFileWriter(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	// Create file writer
	writer, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer writer.Close()

	// Test write
	data := []byte("test log entry\n")
	n, err := writer.Write(data)

	if err != nil {
		t.Fatalf("Unexpected write error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	// Test sync
	err = writer.Sync()
	if err != nil {
		t.Fatalf("Unexpected sync error: %v", err)
	}

	// Close and verify file contents
	writer.Close()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != "test log entry\n" {
		t.Errorf("Expected 'test log entry\\n', got '%s'", string(content))
	}
}

// TestFileWriterAppend tests append functionality
func TestFileWriterAppend(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "append_test.log")

	// Write initial content
	err := os.WriteFile(filePath, []byte("initial\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Open with file writer (should append)
	writer, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer writer.Close()

	// Write additional content
	_, err = writer.Write([]byte("appended\n"))
	if err != nil {
		t.Fatalf("Unexpected write error: %v", err)
	}

	writer.Close()

	// Verify both entries exist
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "initial\nappended\n"
	if string(content) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(content))
	}
}

// TestFileWriterMultipleWrites tests multiple writes to file
func TestFileWriterMultipleWrites(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "multiple_writes.log")

	writer, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer writer.Close()

	writes := []string{"entry1\n", "entry2\n", "entry3\n"}

	for _, write := range writes {
		_, err := writer.Write([]byte(write))
		if err != nil {
			t.Fatalf("Unexpected write error: %v", err)
		}
	}

	writer.Close()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "entry1\nentry2\nentry3\n"
	if string(content) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(content))
	}
}

// TestFileWriterInvalidPath tests error handling for invalid paths
func TestFileWriterInvalidPath(t *testing.T) {
	// Try to create file in non-existent directory
	invalidPath := "/non/existent/directory/test.log"

	_, err := NewFileWriter(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}

	// Check that it's a path-related error
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Expected path error, got: %v", err)
	}
}

// TestFileWriterClose tests close functionality
func TestFileWriterClose(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "close_test.log")

	writer, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}

	// Write some data
	_, err = writer.Write([]byte("test\n"))
	if err != nil {
		t.Fatalf("Unexpected write error: %v", err)
	}

	// Close should not error
	err = writer.Close()
	if err != nil {
		t.Errorf("Unexpected close error: %v", err)
	}

	// Second close should not panic (idempotent)
	_ = writer.Close()
	// Note: os.File.Close() may return error on second close, but shouldn't panic
}
