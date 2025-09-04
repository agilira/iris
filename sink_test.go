// sink_test.go: Comprehensive test suite for iris logging sink functionality
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to safely close file ignoring expected errors
func safeCloseSinkFile(t *testing.T, file *os.File) {
	if err := file.Close(); err != nil {
		t.Errorf("Failed to close file: %v", err)
	}
}

// mockWriter implements io.Writer for testing
type mockWriter struct {
	data []byte
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.data = append(m.data, p...)
	return len(p), nil
}

// mockWriteSyncer implements WriteSyncer for testing
type mockWriteSyncer struct {
	mockWriter
	syncCalls int
	syncErr   error
}

func (m *mockWriteSyncer) Sync() error {
	m.syncCalls++
	return m.syncErr
}

// TestWriteSyncerInterface tests that our types implement WriteSyncer
func TestWriteSyncerInterface(t *testing.T) {
	// Test nopSyncer implements WriteSyncer
	var _ WriteSyncer = nopSyncer{}

	// Test fileSyncer implements WriteSyncer
	var _ WriteSyncer = fileSyncer{}

	// Test mockWriteSyncer implements WriteSyncer
	var _ WriteSyncer = &mockWriteSyncer{}
}

// TestNopSyncer tests nopSyncer functionality
func TestNopSyncer(t *testing.T) {
	buf := &bytes.Buffer{}
	syncer := nopSyncer{buf}

	// Test Write functionality
	data := []byte("test data")
	n, err := syncer.Write(data)
	if err != nil {
		t.Errorf("Expected no error from Write, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("Expected buffer to contain %q, got %q", data, buf.Bytes())
	}

	// Test Sync functionality (should be no-op)
	err = syncer.Sync()
	if err != nil {
		t.Errorf("Expected Sync to return nil, got %v", err)
	}
}

// TestFileSyncer tests fileSyncer functionality
func TestFileSyncer(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.log")

	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()

	syncer := fileSyncer{file}

	// Test Write functionality
	data := []byte("test file data")
	n, err := syncer.Write(data)
	if err != nil {
		t.Errorf("Expected no error from Write, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Test Sync functionality
	err = syncer.Sync()
	if err != nil {
		t.Errorf("Expected Sync to succeed, got %v", err)
	}

	// Verify data was written to file
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}
	if !bytes.Equal(content, data) {
		t.Errorf("Expected file to contain %q, got %q", data, content)
	}
}

// TestWrapWriter tests WrapWriter with different input types
func TestWrapWriter(t *testing.T) {
	tests := []struct {
		name     string
		writer   io.Writer
		wantType string
	}{
		{
			name:     "os.File should return fileSyncer",
			writer:   createTempFile(t),
			wantType: "fileSyncer",
		},
		{
			name:     "WriteSyncer should return as-is",
			writer:   &mockWriteSyncer{},
			wantType: "mockWriteSyncer",
		},
		{
			name:     "bytes.Buffer should return nopSyncer",
			writer:   &bytes.Buffer{},
			wantType: "nopSyncer",
		},
		{
			name:     "strings.Builder should return nopSyncer",
			writer:   &strings.Builder{},
			wantType: "nopSyncer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapWriter(tt.writer)

			// result is guaranteed to be WriteSyncer from function signature
			// Just verify we can call its methods
			if result == nil {
				t.Error("WrapWriter returned nil")
				return
			}

			// Check specific type
			switch tt.wantType {
			case "fileSyncer":
				if _, ok := result.(fileSyncer); !ok {
					t.Errorf("Expected fileSyncer, got %T", result)
				}
			case "mockWriteSyncer":
				if _, ok := result.(*mockWriteSyncer); !ok {
					t.Errorf("Expected mockWriteSyncer, got %T", result)
				}
			case "nopSyncer":
				if _, ok := result.(nopSyncer); !ok {
					t.Errorf("Expected nopSyncer, got %T", result)
				}
			}
		})
	}
}

// TestWrapWriterFunctionality tests that wrapped writers work correctly
func TestWrapWriterFunctionality(t *testing.T) {
	testData := []byte("test write and sync")

	// Test with buffer (nopSyncer)
	buf := &bytes.Buffer{}
	syncer := WrapWriter(buf)

	n, err := syncer.Write(testData)
	if err != nil {
		t.Errorf("Expected no error writing to buffer syncer, got %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}

	err = syncer.Sync()
	if err != nil {
		t.Errorf("Expected no error syncing buffer, got %v", err)
	}

	if !bytes.Equal(buf.Bytes(), testData) {
		t.Errorf("Expected buffer to contain %q, got %q", testData, buf.Bytes())
	}

	// Test with file (fileSyncer)
	file := createTempFile(t)
	defer safeCloseSinkFile(t, file)

	fileSyncer := WrapWriter(file)

	n, err = fileSyncer.Write(testData)
	if err != nil {
		t.Errorf("Expected no error writing to file syncer, got %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes to file, wrote %d", len(testData), n)
	}

	err = fileSyncer.Sync()
	if err != nil {
		t.Errorf("Expected no error syncing file, got %v", err)
	}
}

// TestNewFileSyncer tests NewFileSyncer constructor
func TestNewFileSyncer(t *testing.T) {
	file := createTempFile(t)
	defer safeCloseSinkFile(t, file)

	syncer := NewFileSyncer(file)

	// Verify type
	_, ok := syncer.(fileSyncer)
	if !ok {
		t.Errorf("Expected fileSyncer, got %T", syncer)
	}

	// Test functionality
	data := []byte("file syncer test")
	n, err := syncer.Write(data)
	if err != nil {
		t.Errorf("Expected no error writing, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	err = syncer.Sync()
	if err != nil {
		t.Errorf("Expected no error syncing, got %v", err)
	}
}

// TestNewNopSyncer tests NewNopSyncer constructor
func TestNewNopSyncer(t *testing.T) {
	buf := &bytes.Buffer{}
	syncer := NewNopSyncer(buf)

	// Verify type
	_, ok := syncer.(nopSyncer)
	if !ok {
		t.Errorf("Expected nopSyncer, got %T", syncer)
	}

	// Test functionality
	data := []byte("nop syncer test")
	n, err := syncer.Write(data)
	if err != nil {
		t.Errorf("Expected no error writing, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	err = syncer.Sync()
	if err != nil {
		t.Errorf("Expected no error syncing, got %v", err)
	}

	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("Expected buffer to contain %q, got %q", data, buf.Bytes())
	}
}

// TestIsFileSyncer tests IsFileSyncer utility function
func TestIsFileSyncer(t *testing.T) {
	tests := []struct {
		name     string
		syncer   WriteSyncer
		expected bool
	}{
		{
			name:     "fileSyncer should return true",
			syncer:   NewFileSyncer(createTempFile(t)),
			expected: true,
		},
		{
			name:     "nopSyncer should return false",
			syncer:   NewNopSyncer(&bytes.Buffer{}),
			expected: false,
		},
		{
			name:     "mockWriteSyncer should return false",
			syncer:   &mockWriteSyncer{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFileSyncer(tt.syncer)
			if result != tt.expected {
				t.Errorf("Expected IsFileSyncer to return %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsNopSyncer tests IsNopSyncer utility function
func TestIsNopSyncer(t *testing.T) {
	tests := []struct {
		name     string
		syncer   WriteSyncer
		expected bool
	}{
		{
			name:     "nopSyncer should return true",
			syncer:   NewNopSyncer(&bytes.Buffer{}),
			expected: true,
		},
		{
			name:     "fileSyncer should return false",
			syncer:   NewFileSyncer(createTempFile(t)),
			expected: false,
		},
		{
			name:     "mockWriteSyncer should return false",
			syncer:   &mockWriteSyncer{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNopSyncer(tt.syncer)
			if result != tt.expected {
				t.Errorf("Expected IsNopSyncer to return %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestSyncBehavior tests different sync behaviors
func TestSyncBehavior(t *testing.T) {
	// Test nopSyncer always succeeds
	nop := NewNopSyncer(&bytes.Buffer{})
	for i := 0; i < 100; i++ {
		if err := nop.Sync(); err != nil {
			t.Errorf("nopSyncer.Sync() should never fail, got %v", err)
		}
	}

	// Test fileSyncer with valid file
	file := createTempFile(t)
	defer safeCloseSinkFile(t, file)

	fileSync := NewFileSyncer(file)
	if err := fileSync.Sync(); err != nil {
		t.Errorf("fileSyncer.Sync() with valid file should succeed, got %v", err)
	}

	// Test mockWriteSyncer with controlled error
	mock := &mockWriteSyncer{}
	if err := mock.Sync(); err != nil {
		t.Errorf("mockWriteSyncer.Sync() should succeed by default, got %v", err)
	}
	if mock.syncCalls != 1 {
		t.Errorf("Expected 1 sync call, got %d", mock.syncCalls)
	}
}

// TestConcurrentAccess tests concurrent access to syncers (with individual buffers)
func TestConcurrentAccess(t *testing.T) {
	// Each goroutine gets its own buffer to avoid race conditions
	syncers := make([]WriteSyncer, 10)
	for i := range syncers {
		syncers[i] = NewNopSyncer(&bytes.Buffer{})
	}

	// Run concurrent writes and syncs
	done := make(chan bool, 20)

	// Start 10 writers (each with its own syncer)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			data := []byte("concurrent test")
			for j := 0; j < 100; j++ {
				_, _ = syncers[id].Write(data)
			}
		}(i)
	}

	// Start 10 syncers (each with its own syncer)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				_ = syncers[id].Sync()
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Test should complete without race conditions or panics
}

// TestPerformance tests performance characteristics (not a benchmark)
func TestPerformance(t *testing.T) {
	data := []byte("performance test data")

	// Test nopSyncer performance
	nop := NewNopSyncer(&bytes.Buffer{})
	for i := 0; i < 1000; i++ {
		_, _ = nop.Write(data)
		_ = nop.Sync()
	}

	// Test fileSyncer performance
	file := createTempFile(t)
	defer safeCloseSinkFile(t, file)

	fileSync := NewFileSyncer(file)
	for i := 0; i < 100; i++ { // Fewer iterations for file I/O
		_, _ = fileSync.Write(data)
		_ = fileSync.Sync()
	}

	// If we reach here, performance is acceptable for testing
}

// Helper function to create temporary file for testing
func createTempFile(t *testing.T) *os.File {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.log")

	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	t.Cleanup(func() {
		_ = file.Close()
	})

	return file
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	// Test WrapWriter with nil (should panic, but let's not test that)
	// Instead test with unusual but valid writers

	// Test with empty write
	syncer := NewNopSyncer(&bytes.Buffer{})
	n, err := syncer.Write([]byte{})
	if err != nil {
		t.Errorf("Expected no error writing empty slice, got %v", err)
	}
	if n != 0 {
		t.Errorf("Expected to write 0 bytes, wrote %d", n)
	}

	// Test with multiple syncs
	for i := 0; i < 10; i++ {
		if err := syncer.Sync(); err != nil {
			t.Errorf("Multiple syncs should not fail, got %v on sync %d", err, i)
		}
	}
}

// TestIntegrationScenarios tests realistic usage scenarios
func TestIntegrationScenarios(t *testing.T) {
	scenarios := []struct {
		name   string
		setup  func() WriteSyncer
		verify func(WriteSyncer) error
	}{
		{
			name: "File logging scenario",
			setup: func() WriteSyncer {
				return WrapWriter(createTempFile(t))
			},
			verify: func(ws WriteSyncer) error {
				// Simulate logging operations
				_, err := ws.Write([]byte("INFO: Application started\n"))
				if err != nil {
					return err
				}
				_, err = ws.Write([]byte("DEBUG: Processing request\n"))
				if err != nil {
					return err
				}
				return ws.Sync()
			},
		},
		{
			name: "Buffer logging scenario",
			setup: func() WriteSyncer {
				return WrapWriter(&bytes.Buffer{})
			},
			verify: func(ws WriteSyncer) error {
				// Simulate high-volume logging
				for i := 0; i < 1000; i++ {
					_, err := ws.Write([]byte("log entry\n"))
					if err != nil {
						return err
					}
				}
				return ws.Sync()
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			syncer := scenario.setup()
			err := scenario.verify(syncer)
			if err != nil {
				t.Errorf("Scenario verification failed: %v", err)
			}
		})
	}
}

// TestAddSync tests the AddSync alias function
func TestAddSync(t *testing.T) {
	t.Run("AddSync with buffer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ws := AddSync(buf)

		if ws == nil {
			t.Error("AddSync returned nil")
		}

		// Test basic functionality
		data := []byte("test")
		n, err := ws.Write(data)
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected %d bytes, got %d", len(data), n)
		}

		err = ws.Sync()
		if err != nil {
			t.Errorf("Sync failed: %v", err)
		}
	})

	t.Run("AddSync with WriteSyncer", func(t *testing.T) {
		mock := &mockWriteSyncer{}
		ws := AddSync(mock)

		// Should return the same WriteSyncer
		if ws != mock {
			t.Error("AddSync should return same WriteSyncer when input already implements interface")
		}
	})
} // TestMultiWriteSyncer tests the MultiWriteSyncer functionality
func TestMultiWriteSyncer(t *testing.T) {
	t.Run("Multiple writers", func(t *testing.T) {
		buf1 := &mockWriteSyncer{}
		buf2 := &mockWriteSyncer{}
		buf3 := &mockWriteSyncer{}

		multi := MultiWriteSyncer(buf1, buf2, buf3)

		// Test writing
		data := []byte("test message")
		n, err := multi.Write(data)
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected %d bytes written, got %d", len(data), n)
		}

		// Verify all writers received the data
		for i, buf := range []*mockWriteSyncer{buf1, buf2, buf3} {
			if !bytes.Equal(buf.data, data) {
				t.Errorf("Writer %d did not receive correct data", i)
			}
		}

		// Test sync
		err = multi.Sync()
		if err != nil {
			t.Errorf("Sync failed: %v", err)
		}

		// Verify all writers were synced
		for i, buf := range []*mockWriteSyncer{buf1, buf2, buf3} {
			if buf.syncCalls != 1 {
				t.Errorf("Writer %d: expected 1 sync call, got %d", i, buf.syncCalls)
			}
		}
	})

	t.Run("Nil writers filtered", func(t *testing.T) {
		buf1 := &mockWriteSyncer{}
		buf2 := &mockWriteSyncer{}

		multi := MultiWriteSyncer(buf1, nil, buf2, nil)

		data := []byte("test")
		_, err := multi.Write(data)
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}

		// Only non-nil writers should have data
		if !bytes.Equal(buf1.data, data) {
			t.Error("buf1 did not receive data")
		}
		if !bytes.Equal(buf2.data, data) {
			t.Error("buf2 did not receive data")
		}
	})

	t.Run("Write error handling", func(t *testing.T) {
		buf1 := &mockWriteSyncer{}
		buf2 := &errorWriter{error: io.ErrShortWrite}
		buf3 := &mockWriteSyncer{}

		multi := MultiWriteSyncer(buf1, buf2, buf3)

		_, err := multi.Write([]byte("test"))
		if err == nil {
			t.Error("Expected error from multi writer")
		}
		if err != io.ErrShortWrite {
			t.Errorf("Expected ErrShortWrite, got %v", err)
		}
	})
}

// TestMultiWriter tests the MultiWriter convenience function
func TestMultiWriter(t *testing.T) {
	t.Run("Multiple io.Writers", func(t *testing.T) {
		buf1 := &bytes.Buffer{}
		buf2 := &bytes.Buffer{}
		buf3 := &bytes.Buffer{}

		multi := MultiWriter(buf1, buf2, buf3)

		data := []byte("test message")
		n, err := multi.Write(data)
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected %d bytes written, got %d", len(data), n)
		}

		// Verify all buffers received the data
		for i, buf := range []*bytes.Buffer{buf1, buf2, buf3} {
			if !bytes.Equal(buf.Bytes(), data) {
				t.Errorf("Buffer %d did not receive correct data", i)
			}
		}

		// Test sync (should not error for buffers)
		err = multi.Sync()
		if err != nil {
			t.Errorf("Sync failed: %v", err)
		}
	})

	t.Run("Mixed writer types", func(t *testing.T) {
		buf := &bytes.Buffer{}
		mock := &mockWriteSyncer{}

		multi := MultiWriter(buf, mock)

		data := []byte("mixed types")
		_, err := multi.Write(data)
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}

		// Verify both received data
		if !bytes.Equal(buf.Bytes(), data) {
			t.Error("Buffer did not receive data")
		}
		if !bytes.Equal(mock.data, data) {
			t.Error("Mock writer did not receive data")
		}

		// Test sync
		err = multi.Sync()
		if err != nil {
			t.Errorf("Sync failed: %v", err)
		}

		if mock.syncCalls != 1 {
			t.Errorf("Expected 1 sync call, got %d", mock.syncCalls)
		}
	})
}

// errorWriter for testing error conditions
type errorWriter struct {
	error
}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, e.error
}

func (e *errorWriter) Sync() error {
	return e.error
}
