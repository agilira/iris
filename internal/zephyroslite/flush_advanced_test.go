package zephyroslite

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

// TestZephyrosLight_Flush_EmptyBuffer tests Flush() on an empty buffer
func TestZephyrosLight_Flush_EmptyBuffer(t *testing.T) {
	var processed int32

	buffer, err := NewBuilder[string](16).
		WithProcessor(func(value *string) {
			atomic.AddInt32(&processed, 1)
		}).
		WithBatchSize(1).
		WithBackpressurePolicy(BlockOnFull).
		WithIdleStrategy(NewSpinningIdleStrategy()).
		Build()
	if err != nil {
		t.Fatalf("Failed to build buffer: %v", err)
	}
	defer buffer.Close()

	// Start consumer
	go buffer.Loop()

	// Wait for consumer to start
	time.Sleep(10 * time.Millisecond)

	// Flush empty buffer - should return immediately
	start := time.Now()
	err = buffer.Flush()
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Flush() on empty buffer should not error: %v", err)
	}

	// Should be very fast (less than 100ms)
	threshold := 100 * time.Millisecond

	// In CI environments, be more lenient with timing
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		threshold = 500 * time.Millisecond // Much more lenient for CI
	}

	if duration > threshold {
		t.Errorf("Flush() on empty buffer took too long: %v (threshold: %v)", duration, threshold)
	}
}

// TestZephyrosLight_Flush_WithConsumer tests normal Flush() behavior
func TestZephyrosLight_Flush_WithConsumer(t *testing.T) {
	var processed int32

	buffer, err := NewBuilder[string](16).
		WithProcessor(func(value *string) {
			atomic.AddInt32(&processed, 1)
			// Small delay to simulate work
			time.Sleep(5 * time.Millisecond)
		}).
		WithBatchSize(1).
		WithBackpressurePolicy(BlockOnFull).
		WithIdleStrategy(NewSpinningIdleStrategy()).
		Build()
	if err != nil {
		t.Fatalf("Failed to build buffer: %v", err)
	}
	defer buffer.Close()

	// Start consumer
	go buffer.Loop()

	// Wait for consumer to start
	time.Sleep(10 * time.Millisecond)

	// Write several messages
	const numMessages = 5
	for i := 0; i < numMessages; i++ {
		success := buffer.Write(func(value *string) {
			*value = "test_message"
		})
		if !success {
			t.Fatalf("Failed to write message %d", i)
		}
	}

	// Flush should wait for all messages to be processed
	err = buffer.Flush()
	if err != nil {
		t.Errorf("Flush() should succeed: %v", err)
	}

	// All messages should be processed
	processedCount := atomic.LoadInt32(&processed)
	if processedCount != numMessages {
		t.Errorf("Expected %d messages processed, got %d", numMessages, processedCount)
	}
}

// TestZephyrosLight_Flush_WithoutConsumer tests Flush() timeout when consumer is not running
func TestZephyrosLight_Flush_WithoutConsumer(t *testing.T) {
	buffer, err := NewBuilder[string](16).
		WithProcessor(func(value *string) {
			// No-op processor
		}).
		WithBatchSize(1).
		WithBackpressurePolicy(BlockOnFull).
		WithIdleStrategy(NewSpinningIdleStrategy()).
		Build()
	if err != nil {
		t.Fatalf("Failed to build buffer: %v", err)
	}
	defer buffer.Close()

	// Write a message without starting consumer
	success := buffer.Write(func(value *string) {
		*value = "test_message"
	})
	if !success {
		t.Fatalf("Failed to write message")
	}

	// Flush should timeout since no consumer is running
	start := time.Now()
	err = buffer.Flush()
	duration := time.Since(start)

	if err == nil {
		t.Error("Flush() should timeout when no consumer is running")
	}

	// Should timeout around 5 seconds (BlockOnFull timeout)
	expectedTimeout := 5 * time.Second
	tolerance := 2 * time.Second // More tolerance for CI

	// In CI environments, be even more lenient with timing
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		tolerance = 3 * time.Second // Very lenient for CI
	}

	if duration < expectedTimeout-tolerance || duration > expectedTimeout+tolerance {
		t.Errorf("Flush() timeout duration unexpected: got %v, expected ~%v", duration, expectedTimeout)
	}

	// Error message should contain timeout information
	if err.Error() == "" {
		t.Error("Timeout error message should not be empty")
	}
}

// TestZephyrosLight_Flush_DropOnFull tests Flush() with DropOnFull policy
// This tests that Flush() behaves correctly when some messages may be dropped
func TestZephyrosLight_Flush_DropOnFull(t *testing.T) {
	var processed int32

	buffer, err := NewBuilder[string](4).
		WithProcessor(func(value *string) {
			atomic.AddInt32(&processed, 1)
			// Small delay
			time.Sleep(1 * time.Millisecond)
		}).
		WithBatchSize(1).
		WithBackpressurePolicy(DropOnFull).
		WithIdleStrategy(NewSpinningIdleStrategy()).
		Build()
	if err != nil {
		t.Fatalf("Failed to build buffer: %v", err)
	}
	defer buffer.Close()

	// Start consumer
	go buffer.Loop()

	// Wait for consumer to start
	time.Sleep(10 * time.Millisecond)

	// Write just a few messages that should all be accepted
	const numMessages = 3 // Well within buffer capacity
	var acceptedCount int

	for i := 0; i < numMessages; i++ {
		success := buffer.Write(func(value *string) {
			*value = fmt.Sprintf("message_%d", i)
		})
		if success {
			acceptedCount++
		}
		time.Sleep(1 * time.Millisecond) // Small delay to avoid overwhelming
	}

	// All messages should be accepted with small count and delay
	if acceptedCount != numMessages {
		t.Errorf("Expected %d messages to be accepted, got %d", numMessages, acceptedCount)
	}

	// Flush should succeed and be relatively quick
	start := time.Now()
	err = buffer.Flush()
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Flush() should succeed with DropOnFull policy: %v", err)
	}

	// Flush should be reasonably quick
	if elapsed > 2*time.Second {
		t.Errorf("Flush took too long (%v)", elapsed)
	}

	// All accepted messages should be processed
	processedCount := atomic.LoadInt32(&processed)

	if int(processedCount) != acceptedCount {
		t.Errorf("Expected %d messages processed (all accepted), got %d", acceptedCount, processedCount)
	}
}
