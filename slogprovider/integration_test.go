// integration_test.go: Full integration tests with iris.NewReaderLogger
//
// WHY separate file: These tests exercise the complete pipeline
// (slog -> Provider -> Iris -> Encoder -> output). They import the parent
// iris package and verify end-to-end behavior including JSON encoding,
// multi-reader fan-in, and baseline performance.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package slogprovider

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/agilira/iris"
)

// bufferedWriter captures output for testing.
// WHY custom: We need both Write (io.Writer) and Sync (iris.Syncer)
// to satisfy iris.Config.Output without external dependencies.
// WHY mu: Write() is called from the ring processor goroutine while the test
// goroutine reads via String(). Mutex provides the required happens-before.
type bufferedWriter struct {
	mu   sync.Mutex
	data []byte
}

func (b *bufferedWriter) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bufferedWriter) Sync() error {
	return nil
}

func (b *bufferedWriter) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}

func TestFullIntegrationWithNewReaderLogger(t *testing.T) {
	// Create provider
	provider := New(100)
	defer func() {
		if err := provider.Close(); err != nil {
			t.Errorf("provider.Close() error = %v", err)
		}
	}()

	// Create buffered output
	buf := &bufferedWriter{}

	// Create ReaderLogger with provider
	readers := []iris.SyncReader{provider}
	logger, err := iris.NewReaderLogger(iris.Config{
		Output:  buf,
		Encoder: iris.NewJSONEncoder(),
		Level:   iris.Debug,
	}, readers)
	if err != nil {
		t.Fatalf("Failed to create ReaderLogger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("logger.Close() error = %v", err)
		}
	}()

	// Start the logger
	logger.Start()

	// Create slog logger using our provider
	slogger := slog.New(provider)

	// Log various levels and types
	slogger.Debug("Debug message", "key", "debug_value")
	slogger.Info("Info message", "user_id", "12345", "action", "login")
	slogger.Warn("Warning message", "component", "auth", "retry_count", 3)
	slogger.Error("Error message", "error_code", "AUTH_FAILED", "duration", 150)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Sync to ensure all records are written
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync failed: %v", err)
	}

	output := buf.String()
	t.Logf("Output: %s", output)

	// Verify all messages were processed
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}

	// Verify content
	testCases := []struct {
		level   string
		message string
		key     string
		value   string
	}{
		{"debug", "Debug message", "key", "debug_value"},
		{"info", "Info message", "user_id", "12345"},
		{"warn", "Warning message", "component", "auth"},
		{"error", "Error message", "error_code", "AUTH_FAILED"},
	}

	for i, tc := range testCases {
		if i >= len(lines) {
			t.Errorf("Missing log line %d", i)
			continue
		}

		line := lines[i]
		if !strings.Contains(line, `"level":"`+tc.level+`"`) {
			t.Errorf("Line %d: expected level %s, got: %s", i, tc.level, line)
		}
		if !strings.Contains(line, `"msg":"`+tc.message+`"`) {
			t.Errorf("Line %d: expected message %s, got: %s", i, tc.message, line)
		}
		if !strings.Contains(line, `"`+tc.key+`":"`+tc.value+`"`) {
			t.Errorf("Line %d: expected field %s=%s, got: %s", i, tc.key, tc.value, line)
		}
	}
}

func TestProviderWithMultipleReaders(t *testing.T) {
	// Create multiple providers
	provider1 := New(50)
	defer func() {
		if err := provider1.Close(); err != nil {
			t.Errorf("provider1.Close() error = %v", err)
		}
	}()

	provider2 := New(50)
	defer func() {
		if err := provider2.Close(); err != nil {
			t.Errorf("provider2.Close() error = %v", err)
		}
	}()

	// Create buffered output
	buf := &bufferedWriter{}

	// Create ReaderLogger with multiple providers
	readers := []iris.SyncReader{provider1, provider2}
	logger, err := iris.NewReaderLogger(iris.Config{
		Output:  buf,
		Encoder: iris.NewJSONEncoder(),
		Level:   iris.Info,
	}, readers)
	if err != nil {
		t.Fatalf("Failed to create ReaderLogger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("logger.Close() error = %v", err)
		}
	}()

	logger.Start()

	// Create slog loggers for each provider
	slogger1 := slog.New(provider1)
	slogger2 := slog.New(provider2)

	// Log from both
	slogger1.Info("Message from logger 1", "source", "logger1")
	slogger2.Info("Message from logger 2", "source", "logger2")

	// Give time for processing
	time.Sleep(100 * time.Millisecond)
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d", len(lines))
	}

	// Verify both messages are present (order may vary due to concurrent readers)
	if !strings.Contains(output, "Message from logger 1") {
		t.Error("Missing message from logger 1")
	}
	if !strings.Contains(output, "Message from logger 2") {
		t.Error("Missing message from logger 2")
	}
	if !strings.Contains(output, `"source":"logger1"`) {
		t.Error("Missing source field from logger 1")
	}
	if !strings.Contains(output, `"source":"logger2"`) {
		t.Error("Missing source field from logger 2")
	}
}

func TestProviderPerformanceBasic(t *testing.T) {
	provider := New(1000)
	defer func() {
		if err := provider.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	// Measure provider Handle performance
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	record.Add("key", "value")

	ctx := context.Background()

	// Warmup
	for i := 0; i < 100; i++ {
		if err := provider.Handle(ctx, record); err != nil {
			t.Errorf("Handle() warmup error = %v", err)
		}
	}

	start := time.Now()
	n := 1000
	for i := 0; i < n; i++ {
		if err := provider.Handle(ctx, record); err != nil {
			t.Errorf("Handle failed: %v", err)
		}
	}
	duration := time.Since(start)

	nsPerOp := duration.Nanoseconds() / int64(n)
	t.Logf("Handle performance: %d ns/op (%d ops in %v)", nsPerOp, n, duration)

	// WHY 1000ns threshold: slog default handlers run at ~1000+ ns/op.
	// Our channel-based Handle is a select + channel send, so sub-200ns typical.
	// With race detector overhead we allow 2x headroom.
	maxNsPerOp := 1000
	if testing.Short() {
		maxNsPerOp = 2000
	}
	if nsPerOp > int64(maxNsPerOp) {
		t.Errorf("Handle too slow: %d ns/op (expected < %d)", nsPerOp, maxNsPerOp)
	}
}
