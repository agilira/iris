// sampler_options_test.go: Tests for sampler configuration through options
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
	"time"
)

// TestWithSamplerOption tests that WithSampler option works correctly
func TestWithSamplerOption(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create a sampler that allows only 2 messages
	sampler := NewTokenBucketSampler(2, 1, time.Hour)

	// Create logger using WithSampler option
	logger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Capacity: 1024,
	}, WithSampler(sampler))

	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Verify sampler is set
	if logger.sampler == nil {
		t.Fatal("Sampler should be set when using WithSampler option")
	}

	// Test that sampling works
	for i := 0; i < 5; i++ {
		logger.Info("test message", Int("id", i))
	}

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if output == "" {
		t.Error("Expected some log output with sampler")
	}

	// Count the number of messages
	lines := 0
	for _, char := range output {
		if char == '\n' {
			lines++
		}
	}

	// Should have limited number of messages due to sampling
	if lines > 3 { // Allow some flexibility for timing
		t.Errorf("Expected sampling to limit messages, got %d lines", lines)
	}
}

// TestWithSamplerNil tests that WithSampler(nil) disables sampling
func TestWithSamplerNil(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create logger with nil sampler
	logger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Capacity: 1024,
	}, WithSampler(nil))

	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Verify sampler is nil
	if logger.sampler != nil {
		t.Error("Sampler should be nil when using WithSampler(nil)")
	}

	// Test that all messages pass through
	messageCount := 5
	for i := 0; i < messageCount; i++ {
		logger.Info("test message", Int("id", i))
	}

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if output == "" {
		t.Error("Expected log output without sampler")
	}

	// Count the number of messages
	lines := 0
	for _, char := range output {
		if char == '\n' {
			lines++
		}
	}

	// Should have all messages without sampling
	if lines != messageCount {
		t.Errorf("Expected %d messages without sampling, got %d lines", messageCount, lines)
	}
}

// TestSamplerPriorityConfigVsOption tests that Config.Sampler takes priority over WithSampler
func TestSamplerPriorityConfigVsOption(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create two different samplers
	configSampler := NewTokenBucketSampler(1, 1, time.Hour)
	optionSampler := NewTokenBucketSampler(10, 1, time.Hour)

	// Create logger with both Config.Sampler and WithSampler option
	logger, err := New(Config{
		Output:   buf,
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Capacity: 1024,
		Sampler:  configSampler, // Config sampler should take priority
	}, WithSampler(optionSampler))

	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Config.Sampler should take priority over WithSampler option
	if logger.sampler != configSampler {
		t.Error("Config.Sampler should take priority over WithSampler option")
	}
}

// TestClonedLoggerInheritsOptions tests that cloned loggers can override sampler
func TestClonedLoggerInheritsOptions(t *testing.T) {
	buf := &bufferedSyncer{}

	// Create original logger without sampler
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

	// Verify original has no sampler
	if originalLogger.sampler != nil {
		t.Error("Original logger should not have sampler")
	}

	// Clone with sampler option
	sampler := NewTokenBucketSampler(3, 1, time.Hour)
	clonedLogger := originalLogger.WithOptions(WithSampler(sampler))

	// Verify cloned logger has sampler
	if clonedLogger.sampler != sampler {
		t.Error("Cloned logger should have the new sampler")
	}

	// Verify original logger still has no sampler
	if originalLogger.sampler != nil {
		t.Error("Original logger should still not have sampler")
	}
}
