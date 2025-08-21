// writesyncers_unit_test.go: Unit tests for writesyncers components
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// mockWriteSyncer for testing
type mockWriteSyncer struct {
	data    []byte
	writes  int
	syncs   int
	failAt  int // Fail after N writes (0 = never fail)
	lastErr error
}

func (mws *mockWriteSyncer) Write(p []byte) (n int, err error) {
	mws.writes++
	if mws.failAt > 0 && mws.writes >= mws.failAt {
		mws.lastErr = ErrMockFailure
		return 0, mws.lastErr
	}
	mws.data = append(mws.data, p...)
	return len(p), nil
}

func (mws *mockWriteSyncer) Sync() error {
	mws.syncs++
	if mws.failAt > 0 && mws.syncs >= mws.failAt {
		mws.lastErr = ErrMockFailure
		return mws.lastErr
	}
	return nil
}

func (mws *mockWriteSyncer) String() string {
	return string(mws.data)
}

// Custom error for testing
var ErrMockFailure = errors.New("mock failure")

// TestFileWriteSyncer tests basic file write syncer functionality
func TestFileWriteSyncer(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	// Create file write syncer
	syncer, err := NewFileWriteSyncer(filePath)
	if err != nil {
		t.Fatalf("Failed to create file write syncer: %v", err)
	}
	defer syncer.Close()

	// Test write
	data := []byte("test log entry\n")
	n, err := syncer.Write(data)

	if err != nil {
		t.Fatalf("Unexpected write error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	// Test sync
	err = syncer.Sync()
	if err != nil {
		t.Fatalf("Unexpected sync error: %v", err)
	}

	// Close and verify file contents
	syncer.Close()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != "test log entry\n" {
		t.Errorf("Expected 'test log entry\\n', got '%s'", string(content))
	}
}

// TestFileWriteSyncerConcurrency tests concurrent access
func TestFileWriteSyncerConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "concurrent_test.log")

	syncer, err := NewFileWriteSyncer(filePath)
	if err != nil {
		t.Fatalf("Failed to create file write syncer: %v", err)
	}
	defer syncer.Close()

	// Test concurrent writes
	done := make(chan bool, 2)

	// Goroutine 1
	go func() {
		for i := 0; i < 100; i++ {
			_, err := syncer.Write([]byte("A"))
			if err != nil {
				t.Errorf("Write error in goroutine 1: %v", err)
			}
		}
		done <- true
	}()

	// Goroutine 2
	go func() {
		for i := 0; i < 100; i++ {
			_, err := syncer.Write([]byte("B"))
			if err != nil {
				t.Errorf("Write error in goroutine 2: %v", err)
			}
		}
		done <- true
	}()

	// Wait for completion
	<-done
	<-done

	// Sync and close
	syncer.Sync()
	syncer.Close()

	// Verify file has correct length
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if len(content) != 200 { // 100 A's + 100 B's
		t.Errorf("Expected 200 bytes, got %d", len(content))
	}
}

// TestBufferedWriteSyncerBasic tests buffered write syncer functionality
func TestBufferedWriteSyncerBasic(t *testing.T) {
	mock := &mockWriteSyncer{}
	buffered := NewBufferedWriteSyncer(mock, 1024)

	// Test write
	data := []byte("test data\n")
	n, err := buffered.Write(data)

	if err != nil {
		t.Fatalf("Unexpected write error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	// Data should be buffered, not written to underlying writer yet
	if mock.writes != 0 {
		t.Errorf("Expected 0 writes to underlying writer, got %d", mock.writes)
	}

	// Test sync (should flush buffer)
	err = buffered.Sync()
	if err != nil {
		t.Fatalf("Unexpected sync error: %v", err)
	}

	// Now underlying writer should have been called
	if mock.writes != 1 {
		t.Errorf("Expected 1 write to underlying writer after sync, got %d", mock.writes)
	}

	if mock.syncs != 1 {
		t.Errorf("Expected 1 sync call, got %d", mock.syncs)
	}

	if mock.String() != "test data\n" {
		t.Errorf("Expected 'test data\\n', got '%s'", mock.String())
	}
}

// TestBufferedWriteSyncerBufferOverflow tests buffer overflow handling
func TestBufferedWriteSyncerBufferOverflow(t *testing.T) {
	mock := &mockWriteSyncer{}
	buffered := NewBufferedWriteSyncer(mock, 10) // Small buffer

	// Write data larger than buffer
	largeData := make([]byte, 25)
	for i := range largeData {
		largeData[i] = 'A'
	}

	n, err := buffered.Write(largeData)
	if err != nil {
		t.Fatalf("Unexpected write error: %v", err)
	}

	if n != len(largeData) {
		t.Errorf("Expected %d bytes written, got %d", len(largeData), n)
	}

	// Buffer should have been flushed multiple times
	if mock.writes < 2 {
		t.Errorf("Expected at least 2 writes to underlying writer, got %d", mock.writes)
	}

	// Sync to flush remaining data
	buffered.Sync()

	if len(mock.data) != 25 {
		t.Errorf("Expected 25 bytes in underlying writer, got %d", len(mock.data))
	}
}

// TestBufferedWriteSyncerMultipleWrites tests multiple small writes
func TestBufferedWriteSyncerMultipleWrites(t *testing.T) {
	mock := &mockWriteSyncer{}
	buffered := NewBufferedWriteSyncer(mock, 100)

	// Write multiple small chunks
	for i := 0; i < 5; i++ {
		data := []byte("chunk")
		_, err := buffered.Write(data)
		if err != nil {
			t.Fatalf("Unexpected write error: %v", err)
		}
	}

	// Should still be buffered
	if mock.writes != 0 {
		t.Errorf("Expected 0 writes to underlying writer, got %d", mock.writes)
	}

	// Sync should flush all
	buffered.Sync()

	if mock.writes != 1 {
		t.Errorf("Expected 1 write after sync, got %d", mock.writes)
	}

	expected := "chunkchunkchunkchunkchunk"
	if mock.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, mock.String())
	}
}

// TestDiscardWriteSyncerBasic tests discard syncer
func TestDiscardWriteSyncerBasic(t *testing.T) {
	discard := &DiscardWriteSyncer{}

	// Test write
	data := []byte("test data that will be discarded\n")
	n, err := discard.Write(data)

	if err != nil {
		t.Fatalf("Unexpected write error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	// Test sync
	err = discard.Sync()
	if err != nil {
		t.Fatalf("Unexpected sync error: %v", err)
	}
}

// TestFileWriteSyncerInvalidPath tests error handling for invalid paths
func TestFileWriteSyncerInvalidPath(t *testing.T) {
	// Try to create file in non-existent directory
	invalidPath := "/non/existent/directory/test.log"

	_, err := NewFileWriteSyncer(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

// TestBufferedWriteSyncerErrorHandling tests error propagation
func TestBufferedWriteSyncerErrorHandling(t *testing.T) {
	mock := &mockWriteSyncer{failAt: 1}
	buffered := NewBufferedWriteSyncer(mock, 10)

	// Write small data (stays in buffer)
	_, err := buffered.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Unexpected error writing to buffer: %v", err)
	}

	// Sync should trigger error
	err = buffered.Sync()
	if err == nil {
		t.Error("Expected sync error, got nil")
	}

	if err != ErrMockFailure {
		t.Errorf("Expected ErrMockFailure, got %v", err)
	}
}
