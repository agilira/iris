// sync_integration_test.go: Test critical Sync() functionality in IRIS logger
//
// This test ensures that logger.Sync() properly flushes all buffered log records
// to their destination, which is essential for data integrity.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

// Helper function for safe logger closing in sync integration tests
func safeCloseSyncLogger(t *testing.T, logger *Logger) {
	if logger != nil {
		if err := logger.Close(); err != nil {
			t.Logf("Failed to close logger: %v", err)
		}
	}
}

// TestLoggerSyncFlushesAllRecords tests that Sync() waits for all records to be processed
func TestLoggerSyncFlushesAllRecords(t *testing.T) {
	// Use a buffer to capture output
	var buf bytes.Buffer

	// Simple slow writer without additional mutex complications
	bufWrapper := &bufferWrapper{Buffer: &buf}
	slowWriter := &slowWriterWrapper{
		writer: bufWrapper,
		mu:     nil, // Don't use additional mutex
		delay:  10 * time.Millisecond,
	}

	logger, err := New(Config{
		Level:    Debug,
		Output:   slowWriter,
		Encoder:  NewTextEncoder(),
		Capacity: 64, // Must be power of 2
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Start()
	defer safeCloseSyncLogger(t, logger)

	// Write several log messages quickly
	const messageCount = 5
	for i := 0; i < messageCount; i++ {
		logger.Infof("Test message %d", i)
	}

	// At this point, messages might be in the ring buffer but not yet written
	// Sync() should block until all are processed

	startTime := time.Now()
	err = logger.Sync()
	syncDuration := time.Since(startTime)

	if err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}

	// Sync should have taken some time due to slow writer
	minExpectedDuration := time.Duration(messageCount) * slowWriter.delay / 2
	if syncDuration < minExpectedDuration {
		t.Errorf("Sync() completed too quickly (%v), expected at least %v",
			syncDuration, minExpectedDuration)
	}

	// After Sync(), all messages should be in the output
	output := buf.String()

	// Count messages in output
	messageLines := strings.Split(strings.TrimSpace(output), "\n")
	actualCount := 0
	for _, line := range messageLines {
		if strings.Contains(line, "Test message") {
			actualCount++
		}
	}

	if actualCount != messageCount {
		t.Errorf("Expected %d messages in output after Sync(), got %d. Output:\n%s",
			messageCount, actualCount, output)
	}
}

// TestLoggerSyncWithEmptyBuffer tests Sync() behavior when nothing is buffered
func TestLoggerSyncWithEmptyBuffer(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New(Config{
		Level:    Info,
		Output:   &bufferWrapper{Buffer: &buf},
		Encoder:  NewTextEncoder(),
		Capacity: 64, // Must be power of 2
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Start()
	defer safeCloseSyncLogger(t, logger)

	// Call Sync() without writing anything - should return quickly
	startTime := time.Now()
	err = logger.Sync()
	syncDuration := time.Since(startTime)

	if err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}

	// Should complete very quickly since nothing to flush
	maxExpectedDuration := 10 * time.Millisecond
	if syncDuration > maxExpectedDuration {
		t.Errorf("Sync() took too long (%v) for empty buffer, expected < %v",
			syncDuration, maxExpectedDuration)
	}
}

// TestLoggerSyncConcurrentWrites tests Sync() with concurrent logging
func TestLoggerSyncConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	var mu sync.Mutex

	bufWrapper := &bufferWrapper{Buffer: &buf, mu: &mu}
	slowWriter := &slowWriterWrapper{
		writer: bufWrapper,
		mu:     nil, // Don't double-lock - bufWrapper already has mutex
		delay:  5 * time.Millisecond,
	}

	logger, err := New(Config{
		Level:    Debug,
		Output:   slowWriter,
		Encoder:  NewTextEncoder(),
		Capacity: 128, // Must be power of 2
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Start()
	defer safeCloseSyncLogger(t, logger)

	const writerCount = 3
	const messagesPerWriter = 5

	// Start multiple goroutines writing logs
	var wg sync.WaitGroup
	for w := 0; w < writerCount; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < messagesPerWriter; i++ {
				logger.Infof("Writer %d message %d", writerID, i)
			}
		}(w)
	}

	// Wait for all writes to complete
	wg.Wait()

	// Now sync - should wait for all messages to be processed
	err = logger.Sync()
	if err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}

	// Check that all messages are in output
	mu.Lock()
	output := buf.String()
	mu.Unlock()

	expectedCount := writerCount * messagesPerWriter
	actualCount := strings.Count(output, "Writer")

	if actualCount != expectedCount {
		t.Errorf("Expected %d messages in output after Sync(), got %d. Output:\n%s",
			expectedCount, actualCount, output)
	}
}

// TestLoggerDeferSyncScenario tests the critical defer logger.Sync() use case
func TestLoggerDeferSyncScenario(t *testing.T) {
	var buf bytes.Buffer
	var mu sync.Mutex

	simulateApplicationShutdown := func() (string, error) {
		logger, err := New(Config{
			Level: Error,
			Output: &slowWriterWrapper{
				writer: &bufferWrapper{Buffer: &buf, mu: &mu},
				mu:     nil, // Don't double-lock - bufWrapper already has mutex
				delay:  15 * time.Millisecond,
			},
			Encoder:  NewTextEncoder(),
			Capacity: 64, // Must be power of 2
		})
		if err != nil {
			return "", err
		}

		logger.Start()
		defer func() {
			// This is the critical pattern - deferred Sync() before Close()
			if syncErr := logger.Sync(); syncErr != nil {
				t.Errorf("Deferred Sync() failed: %v", syncErr)
			}
			safeCloseSyncLogger(t, logger)
		}()

		// Simulate critical error logging at application shutdown
		logger.Error("Critical system error during shutdown")
		logger.Error("Database connection lost")
		logger.Error("Unable to save user data")

		// Function returns immediately, but deferred Sync() should ensure
		// all critical messages are written before Close()
		return "shutdown complete", nil
	}

	// Simulate the scenario
	result, err := simulateApplicationShutdown()
	if err != nil {
		t.Fatalf("Shutdown simulation failed: %v", err)
	}

	if result != "shutdown complete" {
		t.Errorf("Expected 'shutdown complete', got %s", result)
	}

	// Verify all critical messages were written
	mu.Lock()
	output := buf.String()
	mu.Unlock()

	criticalMessages := []string{
		"Critical system error during shutdown",
		"Database connection lost",
		"Unable to save user data",
	}

	for _, msg := range criticalMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Critical message '%s' not found in output after deferred Sync(). Output:\n%s", msg, output)
		}
	}

	// Count total error messages
	errorCount := strings.Count(output, "level=error")
	if errorCount != len(criticalMessages) {
		t.Errorf("Expected %d error messages, got %d in output:\n%s", len(criticalMessages), errorCount, output)
	}
}

// slowWriterWrapper simulates slow I/O operations to test timing
type slowWriterWrapper struct {
	writer WriteSyncer
	mu     *sync.Mutex
	delay  time.Duration
}

func (s *slowWriterWrapper) Write(p []byte) (n int, err error) {
	// Simulate slow I/O
	time.Sleep(s.delay)
	if s.mu != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	return s.writer.Write(p)
}

func (s *slowWriterWrapper) Sync() error {
	if s.mu != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	if syncer, ok := s.writer.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}

// bufferWrapper wraps bytes.Buffer to implement WriteSyncer
type bufferWrapper struct {
	*bytes.Buffer
	mu *sync.Mutex
}

func (b *bufferWrapper) Write(p []byte) (n int, err error) {
	if b.mu != nil {
		b.mu.Lock()
		defer b.mu.Unlock()
	}
	return b.Buffer.Write(p)
}

func (b *bufferWrapper) Sync() error {
	// bytes.Buffer doesn't need syncing, but we implement it for interface compliance
	return nil
}
