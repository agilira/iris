package iris

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMultiWriter tests the basic MultiWriter functionality
func TestMultiWriter(t *testing.T) {
	var buf1, buf2, buf3 bytes.Buffer

	writer1 := WrapWriter(&buf1)
	writer2 := WrapWriter(&buf2)
	writer3 := WrapWriter(&buf3)

	mw := NewMultiWriter(writer1, writer2, writer3)

	testData := []byte("test message\n")

	n, err := mw.Write(testData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, got %d", len(testData), n)
	}

	// Check all buffers have the same content
	expected := string(testData)
	if buf1.String() != expected {
		t.Errorf("Buffer 1 content mismatch: got %q, want %q", buf1.String(), expected)
	}
	if buf2.String() != expected {
		t.Errorf("Buffer 2 content mismatch: got %q, want %q", buf2.String(), expected)
	}
	if buf3.String() != expected {
		t.Errorf("Buffer 3 content mismatch: got %q, want %q", buf3.String(), expected)
	}

	// Test Sync
	if err := mw.Sync(); err != nil {
		t.Errorf("Sync failed: %v", err)
	}

	// Test Count
	if count := mw.Count(); count != 3 {
		t.Errorf("Expected 3 writers, got %d", count)
	}
}

// TestMultiWriterAddRemove tests adding and removing writers dynamically
func TestMultiWriterAddRemove(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	writer1 := WrapWriter(&buf1)
	writer2 := WrapWriter(&buf2)

	mw := NewMultiWriter(writer1)

	// Initial state
	if count := mw.Count(); count != 1 {
		t.Errorf("Expected 1 writer initially, got %d", count)
	}

	// Add second writer
	mw.AddWriter(writer2)
	if count := mw.Count(); count != 2 {
		t.Errorf("Expected 2 writers after add, got %d", count)
	}

	// Write test data
	testData := []byte("after add\n")
	mw.Write(testData)

	expected := string(testData)
	if buf1.String() != expected {
		t.Errorf("Buffer 1 content mismatch: got %q, want %q", buf1.String(), expected)
	}
	if buf2.String() != expected {
		t.Errorf("Buffer 2 content mismatch: got %q, want %q", buf2.String(), expected)
	}

	// Remove first writer
	if !mw.RemoveWriter(writer1) {
		t.Error("Failed to remove writer1")
	}
	if count := mw.Count(); count != 1 {
		t.Errorf("Expected 1 writer after remove, got %d", count)
	}

	// Write more data
	buf1.Reset()
	buf2.Reset()

	mw.Write([]byte("after remove\n"))

	if buf1.Len() != 0 {
		t.Errorf("Buffer 1 should be empty after removal, got %q", buf1.String())
	}
	if buf2.String() != "after remove\n" {
		t.Errorf("Buffer 2 content mismatch: got %q, want %q", buf2.String(), "after remove\n")
	}
}

// TestLoggerMultipleOutputs tests logger with multiple outputs
func TestLoggerMultipleOutputs(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	config := Config{
		Level:      InfoLevel,
		Format:     JSONFormat,
		BufferSize: 256,
		BatchSize:  16,
		Writers:    []io.Writer{&buf1, &buf2},
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log a message
	logger.Info("test message", String("key", "value"))

	// Give some time for processing
	time.Sleep(50 * time.Millisecond)

	// Both buffers should have content
	content1 := buf1.String()
	content2 := buf2.String()

	if content1 == "" {
		t.Error("Buffer 1 is empty")
	}
	if content2 == "" {
		t.Error("Buffer 2 is empty")
	}
	if content1 != content2 {
		t.Errorf("Buffer contents don't match:\nBuffer 1: %q\nBuffer 2: %q", content1, content2)
	}

	// Check JSON structure
	if !strings.Contains(content1, "test message") {
		t.Error("Message not found in output")
	}
	if !strings.Contains(content1, `"key":"value"`) {
		t.Error("Field not found in output")
	}
}

// TestLoggerDynamicOutputs tests adding/removing outputs from logger
func TestLoggerDynamicOutputs(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	config := Config{
		Level:      InfoLevel,
		Format:     JSONFormat,
		BufferSize: 256,
		BatchSize:  16,
		Writer:     &buf1, // Use single Writer instead of Writers slice
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Initial count (single writer, no tee yet)
	if count := logger.WriterCount(); count != 1 {
		t.Errorf("Expected 1 writer initially, got %d", count)
	}

	// Add second writer (this will create MultiWriter)
	err = logger.AddWriter(&buf2)
	if err != nil {
		t.Fatalf("Failed to add writer: %v", err)
	}

	if count := logger.WriterCount(); count != 2 {
		t.Errorf("Expected 2 writers after add, got %d", count)
	}

	// Log message
	logger.Info("test message")
	time.Sleep(50 * time.Millisecond)

	// Both should have content
	if buf1.Len() == 0 {
		t.Error("Buffer 1 is empty")
	}
	if buf2.Len() == 0 {
		t.Error("Buffer 2 is empty")
	}
}

// TestTeeLogger tests the convenience tee logger function
func TestTeeLogger(t *testing.T) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "iris_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	logger, err := NewTeeLogger(tmpfile.Name(), InfoLevel, JSONFormat)
	if err != nil {
		t.Fatalf("Failed to create tee logger: %v", err)
	}
	defer logger.Close()

	// Log a message
	logger.Info("tee test message", String("test", "value"))

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Check file content
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "tee test message") {
		t.Error("Message not found in file output")
	}
	if !strings.Contains(contentStr, `"test":"value"`) {
		t.Error("Field not found in file output")
	}
}

// TestBufferedWriteSyncer tests the buffered writer functionality
func TestBufferedWriteSyncer(t *testing.T) {
	var buf bytes.Buffer
	base := WrapWriter(&buf)

	buffered := NewBufferedWriteSyncer(base, 10) // Small buffer for testing

	// Write data that fits in buffer
	data1 := []byte("test")
	n, err := buffered.Write(data1)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data1) {
		t.Errorf("Expected to write %d bytes, got %d", len(data1), n)
	}

	// Buffer should not be flushed yet
	if buf.Len() != 0 {
		t.Error("Buffer was flushed prematurely")
	}

	// Write more data to trigger flush
	data2 := []byte("_more_data_to_trigger_flush")
	buffered.Write(data2)

	// Now buffer should have content
	if buf.Len() == 0 {
		t.Error("Buffer was not flushed")
	}

	// Sync to flush remaining
	buffered.Sync()

	expected := string(data1) + string(data2)
	if buf.String() != expected {
		t.Errorf("Buffer content mismatch: got %q, want %q", buf.String(), expected)
	}
}

// TestDiscardWriteSyncer tests the discard writer
func TestDiscardWriteSyncer(t *testing.T) {
	discard := &DiscardWriteSyncer{}

	data := []byte("this will be discarded")
	n, err := discard.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, got %d", len(data), n)
	}

	if err := discard.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
}
