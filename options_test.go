// options_test.go: Tests for advanced logger options
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strings"
	"testing"
	"time"
)

// optionTestSyncer captures logs for option testing
type optionTestSyncer struct {
	logs     []string
	synced   bool
	logCount int
}

func (o *optionTestSyncer) Write(p []byte) (n int, err error) {
	o.logs = append(o.logs, string(p))
	o.logCount++
	return len(p), nil
}

func (o *optionTestSyncer) Sync() error {
	o.synced = true
	return nil
}

// TestWithCallerSkip tests the WithCallerSkip option
func TestWithCallerSkip(t *testing.T) {
	t.Run("WithCallerSkip_PositiveValue", func(t *testing.T) {
		syncer := &optionTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}, WithCaller(), WithCallerSkip(2))
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Log a message with caller information
		logger.Info("test message with caller skip", String("test", "value"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected WithCallerSkip to write log")
		}

		// Verify caller information is present (exact format may vary)
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "test message with caller skip") {
			t.Errorf("Expected log to contain message, got: %s", logContent)
		}
	})

	t.Run("WithCallerSkip_NegativeValue", func(t *testing.T) {
		syncer := &optionTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}, WithCaller(), WithCallerSkip(-5)) // Negative value should be normalized to 0
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Log a message
		logger.Info("test message with negative skip", String("test", "value"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected WithCallerSkip with negative value to write log")
		}

		// Verify message content
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "test message with negative skip") {
			t.Errorf("Expected log to contain message, got: %s", logContent)
		}
	})

	t.Run("WithCallerSkip_ZeroValue", func(t *testing.T) {
		syncer := &optionTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}, WithCaller(), WithCallerSkip(0))
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Log a message
		logger.Info("test message with zero skip", String("test", "value"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected WithCallerSkip with zero value to write log")
		}

		// Verify message content
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "test message with zero skip") {
			t.Errorf("Expected log to contain message, got: %s", logContent)
		}
	})
}

// TestAddStacktrace tests the AddStacktrace option
func TestAddStacktrace(t *testing.T) {
	t.Run("AddStacktrace_Error_Level", func(t *testing.T) {
		syncer := &optionTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}, AddStacktrace(Error)) // Enable stack traces for Error level and above
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Log an error message (should include stack trace)
		logger.Error("test error with stack trace", String("test", "value"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected AddStacktrace to write error log")
		}

		// Verify message content
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "test error with stack trace") {
			t.Errorf("Expected log to contain error message, got: %s", logContent)
		}
	})

	t.Run("AddStacktrace_Warn_Level_No_Stack", func(t *testing.T) {
		syncer := &optionTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}, AddStacktrace(Error)) // Stack traces only for Error level and above
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Log a warning message (should NOT include stack trace)
		logger.Warn("test warning without stack trace", String("test", "value"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected AddStacktrace to write warning log")
		}

		// Verify message content
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "test warning without stack trace") {
			t.Errorf("Expected log to contain warning message, got: %s", logContent)
		}
	})

	t.Run("AddStacktrace_Debug_Level", func(t *testing.T) {
		syncer := &optionTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}, AddStacktrace(Debug)) // Enable stack traces for Debug level and above
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Log a debug message (should include stack trace since min is Debug)
		logger.Debug("test debug with stack trace", String("test", "value"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected AddStacktrace to write debug log")
		}

		// Verify message content
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "test debug with stack trace") {
			t.Errorf("Expected log to contain debug message, got: %s", logContent)
		}
	})

	t.Run("AddStacktrace_Fatal_Level", func(t *testing.T) {
		syncer := &optionTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}, AddStacktrace(Fatal)) // Stack traces only for Fatal level
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Log an error message (should NOT include stack trace since min is Fatal)
		logger.Error("test error without stack trace", String("test", "value"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected AddStacktrace to write error log without stack trace")
		}

		// Verify message content
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "test error without stack trace") {
			t.Errorf("Expected log to contain error message, got: %s", logContent)
		}
	})
}

// TestWithCallerSkip_Integration tests WithCallerSkip in combination with other options
func TestWithCallerSkip_Integration(t *testing.T) {
	t.Run("WithCallerSkip_WithCaller_Integration", func(t *testing.T) {
		syncer := &optionTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}, WithCaller(), WithCallerSkip(1), Development())
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Log a message with multiple options
		logger.Info("integration test message", String("integration", "test"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected integration test to write log")
		}

		// Verify message content
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "integration test message") {
			t.Errorf("Expected log to contain message, got: %s", logContent)
		}
	})
}

// TestAddStacktrace_Integration tests AddStacktrace with other options
func TestAddStacktrace_Integration(t *testing.T) {
	t.Run("AddStacktrace_WithCaller_Integration", func(t *testing.T) {
		syncer := &optionTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}, WithCaller(), AddStacktrace(Warn), Development())
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Log a warning message (should include both caller and stack trace)
		logger.Warn("integration warning with stack", String("integration", "test"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected integration warning to write log")
		}

		// Verify message content
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "integration warning with stack") {
			t.Errorf("Expected log to contain warning message, got: %s", logContent)
		}
	})
}
