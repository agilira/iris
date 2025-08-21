// multiwriter_unit_test.go: Unit tests for MultiWriter
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

// multiWriterMock is a mock implementation for testing
type multiWriterMock struct {
	data       *bytes.Buffer
	syncErr    error
	writeErr   error
	syncCalled bool
}

func newMultiWriterMock() *multiWriterMock {
	return &multiWriterMock{
		data: &bytes.Buffer{},
	}
}

func (m *multiWriterMock) Write(p []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	if m.data == nil {
		return 0, fmt.Errorf("data buffer is nil")
	}
	return m.data.Write(p)
}

func (m *multiWriterMock) Sync() error {
	m.syncCalled = true
	return m.syncErr
}

func (m *multiWriterMock) String() string {
	return m.data.String()
}

// TestNewMultiWriter tests MultiWriter creation
func TestNewMultiWriter(t *testing.T) {
	// Test with no writers
	mw := NewMultiWriter()
	if mw == nil {
		t.Error("NewMultiWriter should not return nil")
	}
	if mw.Count() != 0 {
		t.Errorf("Expected 0 writers, got %d", mw.Count())
	}

	// Test with multiple writers
	w1 := newMultiWriterMock()
	w2 := newMultiWriterMock()
	mw = NewMultiWriter(w1, w2)

	if mw.Count() != 2 {
		t.Errorf("Expected 2 writers, got %d", mw.Count())
	}
}

// TestMultiWriterWrite tests writing to multiple writers
func TestMultiWriterWrite(t *testing.T) {
	w1 := newMultiWriterMock()
	w2 := newMultiWriterMock()
	mw := NewMultiWriter(w1, w2)

	testData := []byte("test message")
	n, err := mw.Write(testData)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), n)
	}

	// Check both writers received the data
	if w1.String() != "test message" {
		t.Errorf("Writer 1 expected 'test message', got '%s'", w1.String())
	}

	if w2.String() != "test message" {
		t.Errorf("Writer 2 expected 'test message', got '%s'", w2.String())
	}
}

// TestMultiWriterWriteEmpty tests writing to empty MultiWriter
func TestMultiWriterWriteEmpty(t *testing.T) {
	mw := NewMultiWriter()

	testData := []byte("test message")
	n, err := mw.Write(testData)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), n)
	}
}

// TestMultiWriterWriteError tests error handling during write
func TestMultiWriterWriteError(t *testing.T) {
	w1 := newMultiWriterMock()
	w2 := newMultiWriterMock()
	w2.writeErr = errors.New("write error")

	mw := NewMultiWriter(w1, w2)

	testData := []byte("test message")
	_, err := mw.Write(testData)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	// First writer should still have written successfully
	if w1.String() != "test message" {
		t.Errorf("Writer 1 expected 'test message', got '%s'", w1.String())
	}
}

// TestMultiWriterSync tests syncing all writers
func TestMultiWriterSync(t *testing.T) {
	w1 := newMultiWriterMock()
	w2 := newMultiWriterMock()
	mw := NewMultiWriter(w1, w2)

	err := mw.Sync()
	if err != nil {
		t.Errorf("Unexpected sync error: %v", err)
	}

	if !w1.syncCalled {
		t.Error("Writer 1 Sync() was not called")
	}

	if !w2.syncCalled {
		t.Error("Writer 2 Sync() was not called")
	}
}

// TestMultiWriterSyncError tests sync error handling
func TestMultiWriterSyncError(t *testing.T) {
	w1 := newMultiWriterMock()
	w2 := newMultiWriterMock()
	w1.syncErr = errors.New("sync error")

	mw := NewMultiWriter(w1, w2)

	err := mw.Sync()
	if err == nil {
		t.Error("Expected sync error, got nil")
	}

	// Both should still be called
	if !w1.syncCalled {
		t.Error("Writer 1 Sync() was not called")
	}

	if !w2.syncCalled {
		t.Error("Writer 2 Sync() was not called")
	}
}

// TestAddWriter tests adding writers dynamically
func TestAddWriter(t *testing.T) {
	mw := NewMultiWriter()

	w1 := newMultiWriterMock()
	mw.AddWriter(w1)

	if mw.Count() != 1 {
		t.Errorf("Expected 1 writer, got %d", mw.Count())
	}

	// Test writing after adding
	testData := []byte("test message")
	mw.Write(testData)

	if w1.String() != "test message" {
		t.Errorf("Writer expected 'test message', got '%s'", w1.String())
	}
}

// TestRemoveWriter tests removing writers dynamically
func TestRemoveWriter(t *testing.T) {
	w1 := newMultiWriterMock()
	w2 := newMultiWriterMock()
	mw := NewMultiWriter(w1, w2)

	// Remove existing writer
	removed := mw.RemoveWriter(w1)
	if !removed {
		t.Error("Expected writer to be removed")
	}

	if mw.Count() != 1 {
		t.Errorf("Expected 1 writer, got %d", mw.Count())
	}

	// Remove non-existing writer
	w3 := newMultiWriterMock()
	removed = mw.RemoveWriter(w3)
	if removed {
		t.Error("Expected writer not to be removed")
	}
}

// TestWriters tests getting writers copy
func TestWriters(t *testing.T) {
	w1 := newMultiWriterMock()
	w2 := newMultiWriterMock()
	mw := NewMultiWriter(w1, w2)

	writers := mw.Writers()
	if len(writers) != 2 {
		t.Errorf("Expected 2 writers, got %d", len(writers))
	}

	// Modify returned slice should not affect original
	writers[0] = nil
	if mw.Count() != 2 {
		t.Error("Original MultiWriter was affected by slice modification")
	}
}

// TestWrapWriter tests writer wrapping functionality
func TestWrapWriter(t *testing.T) {
	buf := &bytes.Buffer{}

	// Test wrapping regular writer
	wrapped := WrapWriter(buf)
	if wrapped == nil {
		t.Error("WrapWriter returned nil")
	}

	// Test writing through wrapper
	testData := []byte("test message")
	n, err := wrapped.Write(testData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), n)
	}

	// Test sync (should not error)
	err = wrapped.Sync()
	if err != nil {
		t.Errorf("Unexpected sync error: %v", err)
	}

	// Test wrapping WriteSyncer (should return as-is)
	ws := newMultiWriterMock()
	wrapped2 := WrapWriter(ws)
	if wrapped2 != ws {
		t.Error("WrapWriter should return WriteSyncer as-is")
	}
}

// TestMultiWriterConcurrency tests concurrent access
func TestMultiWriterConcurrency(t *testing.T) {
	mw := NewMultiWriter()

	// Add writers concurrently
	done := make(chan bool, 10) // Buffered channel to prevent goroutine leaks
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			w := newMultiWriterMock()
			mw.AddWriter(w)
			mw.Write([]byte("test"))
			mw.Sync()
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if mw.Count() != 10 {
		t.Errorf("Expected 10 writers, got %d", mw.Count())
	}
}
