// sampler_integration_test.go: Integration tests for sampler in Iris
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

// TestSamplerIntegrationWithLogger tests that sampler works correctly with the logger
func TestSamplerIntegrationWithLogger(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create a sampler that allows only 2 messages per second
	sampler := NewTokenBucketSampler(2, 1, time.Second)

	logger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Sampler:  sampler,
		Capacity: 1024, // Safe capacity for CI
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Try to log 5 messages quickly - only first 2 should pass
	for i := 0; i < 5; i++ {
		logger.Info("message", Int("count", i))
	}

	// Wait for processing
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

	// Should have at least 1 message but no more than 2 (sampler capacity)
	if len(validLines) == 0 || len(validLines) > 2 {
		t.Errorf("Expected 1-2 log messages, got %d. Output: %s", len(validLines), output)
	}

	// Verify the messages that got through
	for i, line := range validLines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse log line %d: %v", i, err)
			continue
		}

		if logEntry["msg"] != "message" {
			t.Errorf("Line %d: expected message 'message', got %v", i, logEntry["msg"])
		}

		if count, ok := logEntry["count"].(float64); !ok || int(count) != i {
			t.Errorf("Line %d: expected count %d, got %v", i, i, logEntry["count"])
		}
	}
}

// TestSamplerRefillIntegration tests that refill works with the logger
func TestSamplerRefillIntegration(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create a sampler with capacity 2 and fast refill for testing
	sampler := NewTokenBucketSampler(2, 1, 50*time.Millisecond)

	logger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Sampler:  sampler,
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Log first message (should pass)
	logger.Info("first")

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Log second message immediately (should be blocked)
	logger.Info("second")

	// Wait for refill
	time.Sleep(100 * time.Millisecond)

	// Log third message (should pass after refill)
	logger.Info("third")

	// Wait for processing
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

	// Should have at least 1 message but could have 2 depending on timing
	if len(validLines) == 0 {
		t.Errorf("Expected at least 1 log messages after refill, got %d. Output: %s", len(validLines), output)
		return
	}

	// Check first message
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(validLines[0]), &logEntry); err != nil {
		t.Errorf("Failed to parse first log line: %v", err)
		return
	}

	if logEntry["msg"] != "first" {
		t.Errorf("First message: expected 'first', got %v", logEntry["msg"])
	}

	// If we have a second message, it should be "third"
	if len(validLines) > 1 {
		if err := json.Unmarshal([]byte(validLines[1]), &logEntry); err != nil {
			t.Errorf("Failed to parse second log line: %v", err)
			return
		}

		if logEntry["msg"] != "third" {
			t.Errorf("Second message: expected 'third', got %v", logEntry["msg"])
		}
	}
}

// TestSamplerDisabledIntegration tests logger behavior without sampler
func TestSamplerDisabledIntegration(t *testing.T) {
	buf := &bufferedSyncer{}

	logger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Sampler:  nil, // No sampler
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Log multiple messages quickly - all should pass
	messageCount := 5
	for i := 0; i < messageCount; i++ {
		logger.Info("message", Int("count", i))
	}

	// Wait for processing
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

	// Should have all messages (no sampling)
	if len(validLines) != messageCount {
		t.Errorf("Expected %d log messages without sampler, got %d. Output: %s", messageCount, len(validLines), output)
	}
}

// TestSamplerWithDifferentLevels tests that sampler treats all levels equally
func TestSamplerWithDifferentLevels(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create a sampler that allows only 3 messages
	sampler := NewTokenBucketSampler(3, 1, time.Hour) // Very slow refill

	logger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Sampler:  sampler,
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Log messages with different levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message") // This should be blocked

	// Wait for processing
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

	// Should have at most 3 messages (sampler capacity), likely fewer due to timing
	if len(validLines) == 0 || len(validLines) > 3 {
		t.Errorf("Expected 1-3 log messages with different levels, got %d. Output: %s", len(validLines), output)
		return
	}

	// Verify the messages that got through are among the expected ones
	expectedMessages := []string{"debug message", "info message", "warn message"}
	for i, line := range validLines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse log line %d: %v", i, err)
			continue
		}

		msg := logEntry["msg"].(string)
		found := false
		for _, expected := range expectedMessages {
			if msg == expected {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Line %d: unexpected message '%s'", i, msg)
		}
	}
}

// TestClonedLoggerInheritsSampler tests that cloned loggers inherit the sampler
func TestClonedLoggerInheritsSampler(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create a sampler that allows only 2 messages
	sampler := NewTokenBucketSampler(2, 1, time.Hour) // Very slow refill

	originalLogger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Sampler:  sampler,
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, originalLogger)

	originalLogger.Start()

	// Clone the logger
	clonedLogger := originalLogger.WithOptions(WithCaller())

	// Log one message with original logger
	originalLogger.Info("original message")

	// Log one message with cloned logger (should pass - 2nd token)
	clonedLogger.Info("cloned message")

	// Log another message with cloned logger (should be blocked - no tokens left)
	clonedLogger.Info("blocked message")

	// Wait for processing
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

	// Should have at most 2 messages (shared sampler capacity), likely fewer due to timing
	if len(validLines) == 0 || len(validLines) > 2 {
		t.Errorf("Expected 1-2 log messages from cloned logger sharing sampler, got %d. Output: %s", len(validLines), output)
		return
	}

	// Verify the messages that got through
	expectedMessages := []string{"original message", "cloned message"}
	for i, line := range validLines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse log line %d: %v", i, err)
			continue
		}

		msg := logEntry["msg"].(string)
		if i < len(expectedMessages) && msg != expectedMessages[i] {
			t.Errorf("Line %d: expected message '%s', got %s", i, expectedMessages[i], msg)
		}
	}
}
