// zephyros_flush_test.go: Test critical Flush() functionality
//
// This test ensures that Flush() correctly waits for all buffered items
// to be processed, which is essential for Sync() guarantees.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestFlushEnsuresProcessing tests that Flush() waits for all items to be processed
func TestFlushEnsuresProcessing(t *testing.T) {
	const capacity = 8
	const itemCount = 5

	// Counter to track processed items
	var processedCount atomic.Int32

	// Processor that increments counter (simulates log processing)
	processor := func(item *int) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		processedCount.Add(1)
	}

	z, err := NewBuilder[int](capacity).
		WithProcessor(processor).
		WithBatchSize(2).
		Build()
	if err != nil {
		t.Fatalf("Failed to create ZephyrosLight: %v", err)
	}

	// Start consumer in background
	go z.LoopProcess()

	// Write some items to the buffer
	for i := 0; i < itemCount; i++ {
		value := i
		if !z.Write(func(slot *int) {
			*slot = value
		}) {
			t.Fatalf("Failed to write item %d", i)
		}
	}

	// At this point, items are written but may not be processed yet
	// processedCount might be 0

	// Call Flush() - this should block until all items are processed
	startTime := time.Now()
	_ = z.Flush()
	flushDuration := time.Since(startTime)

	// After Flush(), all items should be processed
	finalCount := processedCount.Load()
	if finalCount != itemCount {
		t.Errorf("Expected %d items processed after Flush(), got %d", itemCount, finalCount)
	}

	// Flush should have taken some time (at least the processing time)
	// Since each item takes ~10ms and we process 2 at a time, expect at least 30ms
	minExpectedDuration := 25 * time.Millisecond
	if flushDuration < minExpectedDuration {
		t.Errorf("Flush() completed too quickly (%v), expected at least %v",
			flushDuration, minExpectedDuration)
	}

	z.Close()
}

// TestFlushWithNoItems tests Flush() behavior when buffer is empty
func TestFlushWithNoItems(t *testing.T) {
	const capacity = 8

	processor := func(item *int) {
		// Should not be called
		t.Error("Processor called but no items were written")
	}

	z, err := NewBuilder[int](capacity).
		WithProcessor(processor).
		WithBatchSize(2).
		Build()
	if err != nil {
		t.Fatalf("Failed to create ZephyrosLight: %v", err)
	}

	// Start consumer
	go z.LoopProcess()

	// Call Flush() without writing anything - should return immediately
	startTime := time.Now()
	_ = z.Flush()
	flushDuration := time.Since(startTime)

	// Should complete very quickly since nothing to flush
	maxExpectedDuration := 5 * time.Millisecond
	if flushDuration > maxExpectedDuration {
		t.Errorf("Flush() took too long (%v) for empty buffer, expected < %v",
			flushDuration, maxExpectedDuration)
	}

	z.Close()
}

// TestFlushConcurrentWrites tests Flush() with concurrent writes
func TestFlushConcurrentWrites(t *testing.T) {
	const capacity = 64 // Larger capacity to handle concurrent writes
	const writerCount = 3
	const itemsPerWriter = 5 // Reduced items per writer

	var processedCount atomic.Int32
	var mu sync.Mutex
	var processedItems []int

	processor := func(item *int) {
		mu.Lock()
		processedItems = append(processedItems, *item)
		mu.Unlock()
		processedCount.Add(1)
		time.Sleep(5 * time.Millisecond) // Simulate processing
	}

	z, err := NewBuilder[int](capacity).
		WithProcessor(processor).
		WithBatchSize(4).
		Build()
	if err != nil {
		t.Fatalf("Failed to create ZephyrosLight: %v", err)
	}

	// Start consumer
	go z.LoopProcess()

	// Start multiple writers
	var wg sync.WaitGroup
	for writer := 0; writer < writerCount; writer++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < itemsPerWriter; i++ {
				value := writerID*100 + i
				if !z.Write(func(slot *int) {
					*slot = value
				}) {
					t.Errorf("Writer %d failed to write item %d", writerID, i)
				}
			}
		}(writer)
	}

	// Wait for all writers to complete
	wg.Wait()

	// Now flush - should wait for all items to be processed
	_ = z.Flush()

	// Check that all items were processed
	expectedCount := writerCount * itemsPerWriter
	finalCount := processedCount.Load()
	if finalCount != int32(expectedCount) {
		t.Errorf("Expected %d items processed after Flush(), got %d", expectedCount, finalCount)
	}

	// Check that we have the right number of processed items
	mu.Lock()
	if len(processedItems) != expectedCount {
		t.Errorf("Expected %d processed items, got %d", expectedCount, len(processedItems))
	}
	mu.Unlock()

	z.Close()
}

// TestSyncIntegration tests that logger Sync() properly flushes all records
func TestSyncIntegration(t *testing.T) {
	// This test should be in the main iris package, but we include a note here
	// for the integration testing that should be done.
	t.Skip("Integration test - should be implemented in iris package to test logger.Sync()")
}
