// iris_test.go: Comprehensive tests for Iris Logger implementation
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

// bufferedSyncer wraps a bytes.Buffer to implement WriteSyncer
type bufferedSyncer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (bs *bufferedSyncer) Write(p []byte) (n int, err error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.buf.Write(p)
}

func (bs *bufferedSyncer) Sync() error {
	return nil
}

func (bs *bufferedSyncer) String() string {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.buf.String()
}

// TestLoggerBasicOperations verifies core logger functionality
func TestLoggerBasicOperations(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Start()

	// Test basic logging
	logger.Debug("debug message", Str("key", "value"))
	logger.Info("info message", Int("count", 42))
	logger.Warn("warn message", Bool("flag", true))
	logger.Error("error message", Float64("value", 3.14))

	// Give time for async processing
	time.Sleep(50 * time.Millisecond)

	// Ensure all records are processed
	logger.Sync()

	output := buf.String()
	t.Logf("Output: %q", output)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}

	// Verify JSON structure
	for i, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}

		if _, ok := logEntry["ts"]; !ok {
			t.Errorf("Line %d missing timestamp", i+1)
		}
		if _, ok := logEntry["level"]; !ok {
			t.Errorf("Line %d missing level", i+1)
		}
		if _, ok := logEntry["msg"]; !ok {
			t.Errorf("Line %d missing message", i+1)
		}
	}
}

// TestLoggerLevelFiltering verifies level filtering works correctly
func TestLoggerLevelFiltering(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Warn,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Start the logger for async processing
	logger.Start()

	// Only Warn and Error should be logged
	logger.Debug("debug - should not appear")
	logger.Info("info - should not appear")
	logger.Warn("warn - should appear")
	logger.Error("error - should appear")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines with Warn level, got %d", len(lines))
	}

	// Verify the right messages appear
	if !strings.Contains(output, "warn - should appear") {
		t.Error("Warn message not found in output")
	}
	if !strings.Contains(output, "error - should appear") {
		t.Error("Error message not found in output")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("Debug/Info messages found in output when they should be filtered")
	}
}

// TestLoggerWithFields verifies the With() method for adding base fields
func TestLoggerWithFields(t *testing.T) {
	buf := &bufferedSyncer{}
	baseLogger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer baseLogger.Close()

	// Start the logger for async processing
	baseLogger.Start()

	// Create child logger with base fields
	childLogger := baseLogger.With(
		Str("service", "test-service"),
		Str("version", "1.0.0"),
	)

	childLogger.Info("test message", Str("extra", "field"))

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Verify base fields are present
	if !strings.Contains(output, `"service":"test-service"`) {
		t.Error("Base field 'service' not found in output")
	}
	if !strings.Contains(output, `"version":"1.0.0"`) {
		t.Error("Base field 'version' not found in output")
	}
	if !strings.Contains(output, `"extra":"field"`) {
		t.Error("Extra field not found in output")
	}
}

// TestLoggerNamed verifies the Named() method
func TestLoggerNamed(t *testing.T) {
	buf := &bufferedSyncer{}
	baseLogger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer baseLogger.Close()

	// Start the logger for async processing
	baseLogger.Start()

	// Create named logger
	namedLogger := baseLogger.Named("test-component")
	namedLogger.Info("test message")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Verify logger name is present
	if !strings.Contains(output, `"logger":"test-component"`) {
		t.Error("Logger name not found in output")
	}
}

// TestLoggerSetLevel verifies dynamic level changes
func TestLoggerSetLevel(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Start the logger for async processing
	logger.Start()

	// Initial level allows debug
	logger.Debug("debug1")

	// Change to Info level
	logger.SetLevel(Info)
	logger.Debug("debug2 - should not appear")
	logger.Info("info1")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	// Verify level getter
	if logger.Level() != Info {
		t.Errorf("Expected level Info, got %v", logger.Level())
	}

	output := buf.String()

	if !strings.Contains(output, "debug1") {
		t.Error("First debug message should appear")
	}
	if strings.Contains(output, "debug2") {
		t.Error("Second debug message should not appear after level change")
	}
	if !strings.Contains(output, "info1") {
		t.Error("Info message should appear")
	}
}

// TestLoggerConcurrency verifies thread safety
func TestLoggerConcurrency(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Start the logger for async processing
	logger.Start()

	const numGoroutines = 10
	const messagesPerGoroutine = 5
	var wg sync.WaitGroup

	// Launch concurrent loggers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("concurrent message",
					Int("goroutine", id),
					Int("message", j))
			}
		}(i)
	}

	wg.Wait()

	// Wait longer for async processing of many messages
	time.Sleep(200 * time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	expectedLines := numGoroutines * messagesPerGoroutine
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify all lines are valid JSON
	for i, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}
	}
}

// TestLoggerHooks verifies hook functionality
func TestLoggerHooks(t *testing.T) {
	buf := &bufferedSyncer{}
	var hookCalled bool
	var hookRecord *Record

	hook := func(rec *Record) {
		hookCalled = true
		hookRecord = rec
	}

	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	}, WithHook(hook))
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Start the logger for async processing
	logger.Start()

	logger.Info("test hook message")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	if !hookCalled {
		t.Error("Hook was not called")
	}
	if hookRecord == nil {
		t.Error("Hook record is nil")
	}
}

// TestLoggerDevelopmentMode verifies development mode functionality
func TestLoggerDevelopmentMode(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	}, Development())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Start the logger for async processing
	logger.Start()

	logger.Error("test error")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// In development mode, stack traces should be included
	if !strings.Contains(output, `"stack":`) {
		t.Error("Stack trace not found in development mode output")
	}
}

// TestLoggerWithCallerInfo verifies caller information capture
func TestLoggerWithCallerInfo(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	}, WithCaller())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Start the logger for async processing
	logger.Start()

	logger.Info("test message")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Debug: print the actual output
	t.Logf("Actual output: %s", output)

	// Caller info should be present
	if !strings.Contains(output, `"caller":`) {
		t.Error("Caller information not found in output")
	}
	if !strings.Contains(output, "iris_test.go") {
		t.Error("Source file not found in caller info")
	}
}
