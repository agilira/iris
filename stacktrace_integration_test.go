// stacktrace_integration_test.go: Integration tests for stacktrace with the logger
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestStacktraceIntegrationWithLogger tests that stacktrace works correctly with the logger
func TestStacktraceIntegrationWithLogger(t *testing.T) {
	buf := &bufferedSyncer{}

	logger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Capacity: 1024,
	}, AddStacktrace(Error))

	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Log at different levels
	logger.Debug("debug message") // No stack
	logger.Info("info message")   // No stack
	logger.Warn("warn message")   // No stack
	logger.Error("error message") // Should have stack

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Filter empty lines
	var validLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			validLines = append(validLines, line)
		}
	}

	// Should have 4 messages
	if len(validLines) != 4 {
		t.Errorf("Expected 4 log messages, got %d. Output: %s", len(validLines), output)
	}

	// Check each message
	for i, line := range validLines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse log line %d: %v", i, err)
			continue
		}

		level := logEntry["level"].(string)

		// Only Error level should have stack trace
		if level == "error" {
			if _, hasStack := logEntry["stack"]; !hasStack {
				t.Error("Error message should have stack trace")
			} else {
				stackTrace := logEntry["stack"].(string)
				if stackTrace == "" {
					t.Error("Stack trace should not be empty")
				}
				// Should contain at least one function name and line number
				if !strings.Contains(stackTrace, ".go:") {
					t.Errorf("Stack trace should contain Go source files, got: %s", stackTrace)
				}
				// Should contain either test function or testing framework
				if !strings.Contains(stackTrace, "Test") && !strings.Contains(stackTrace, "testing") {
					t.Errorf("Stack trace should contain test-related functions, got: %s", stackTrace)
				}
			}
		} else {
			if _, hasStack := logEntry["stack"]; hasStack {
				t.Errorf("Level %s should not have stack trace", level)
			}
		}
	}
}

// TestStacktraceWithDifferentLevels tests stacktrace configuration for different levels
func TestStacktraceWithDifferentLevels(t *testing.T) {
	tests := []struct {
		name            string
		stackLevel      Level
		logLevel        Level
		shouldHaveStack bool
	}{
		{"Debug_Stack_Debug_Log", Debug, Debug, true},
		{"Debug_Stack_Info_Log", Debug, Info, true},
		{"Debug_Stack_Error_Log", Debug, Error, true},
		{"Info_Stack_Debug_Log", Info, Debug, false},
		{"Info_Stack_Info_Log", Info, Info, true},
		{"Info_Stack_Error_Log", Info, Error, true},
		{"Error_Stack_Debug_Log", Error, Debug, false},
		{"Error_Stack_Info_Log", Error, Info, false},
		{"Error_Stack_Error_Log", Error, Error, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bufferedSyncer{}

			logger, err := New(Config{
				Output:   buf,
				Level:    Debug, // Always allow all levels
				Encoder:  NewJSONEncoder(),
				Capacity: 1024,
			}, AddStacktrace(tt.stackLevel))

			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer safeCloseIrisLogger(t, logger)

			logger.Start()

			// Log at the specified level
			switch tt.logLevel {
			case Debug:
				logger.Debug("test message")
			case Info:
				logger.Info("test message")
			case Warn:
				logger.Warn("test message")
			case Error:
				logger.Error("test message")
			}

			time.Sleep(50 * time.Millisecond)

			output := buf.String()
			if output == "" {
				t.Error("Expected log output")
				return
			}

			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry); err != nil {
				t.Errorf("Failed to parse log output: %v", err)
				return
			}

			_, hasStack := logEntry["stack"]
			if tt.shouldHaveStack && !hasStack {
				t.Error("Expected stack trace but none found")
			}
			if !tt.shouldHaveStack && hasStack {
				t.Error("Expected no stack trace but found one")
			}
		})
	}
}

// TestStacktraceDisabled tests that stacktrace can be disabled
func TestStacktraceDisabled(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create logger without stacktrace option (should be disabled by default)
	logger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Capacity: 1024,
	})

	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Log error messages that would normally have stack traces
	logger.Error("error without stack")
	logger.Error("another error without stack") // Use Error instead of Fatal

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse log line %d: %v", i, err)
			continue
		}

		if _, hasStack := logEntry["stack"]; hasStack {
			t.Error("Should not have stack trace when disabled")
		}
	}
}

// TestStacktraceWithClonedLogger tests that cloned loggers inherit stacktrace settings
func TestStacktraceWithClonedLogger(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create original logger with stacktrace
	originalLogger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Capacity: 1024,
	}, AddStacktrace(Warn))

	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, originalLogger)

	originalLogger.Start()

	// Clone the logger
	clonedLogger := originalLogger.WithOptions(WithCaller())

	// Both loggers should have the same stacktrace behavior
	originalLogger.Error("original error")
	clonedLogger.Error("cloned error")

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	stackCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			continue
		}

		if _, hasStack := logEntry["stack"]; hasStack {
			stackCount++
		}
	}

	// Both error messages should have stack traces
	if stackCount != 2 {
		t.Errorf("Expected 2 messages with stack traces, got %d", stackCount)
	}
}

// TestStacktraceOverride tests overriding stacktrace settings in cloned logger
func TestStacktraceOverride(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create original logger without stacktrace
	originalLogger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Capacity: 1024,
	})

	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, originalLogger)

	originalLogger.Start()

	// Clone with stacktrace enabled
	stackLogger := originalLogger.WithOptions(AddStacktrace(Error))

	// Original should not have stack, clone should have stack
	originalLogger.Error("original error") // No stack
	stackLogger.Error("stack error")       // Should have stack

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 log lines, got %d", len(lines))
	}

	// Check first message (original logger)
	var firstEntry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(lines[0])), &firstEntry); err != nil {
		t.Fatalf("Failed to parse first log line: %v", err)
	}

	if _, hasStack := firstEntry["stack"]; hasStack {
		t.Error("Original logger should not have stack trace")
	}

	// Check second message (cloned logger)
	var secondEntry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(lines[1])), &secondEntry); err != nil {
		t.Fatalf("Failed to parse second log line: %v", err)
	}

	if _, hasStack := secondEntry["stack"]; !hasStack {
		t.Error("Cloned logger should have stack trace")
	}
}
