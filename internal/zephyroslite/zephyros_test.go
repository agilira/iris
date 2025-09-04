// zephyros_test.go: Tests for ZephyrosLight ring buffer
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"sync"
	"testing"
	"time"
)

// TestBackpressurePolicy_String tests the String method for BackpressurePolicy
func TestBackpressurePolicy_String(t *testing.T) {
	tests := []struct {
		policy   BackpressurePolicy
		expected string
	}{
		{DropOnFull, "DropOnFull"},
		{BlockOnFull, "BlockOnFull"},
	}

	for _, test := range tests {
		result := test.policy.String()
		if result != test.expected {
			t.Errorf("Expected %s for policy %v, got %s", test.expected, test.policy, result)
		}
	}
}

// TestRecord simple test record for ring buffer testing
type TestRecord struct {
	ID      int64
	Message string
	Value   int
}

// TestZephyrosLight_Builder tests the builder pattern
func TestZephyrosLight_Builder(t *testing.T) {
	t.Run("Builder_ValidConfiguration", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex

		processor := func(record *TestRecord) {
			mu.Lock()
			processed = append(processed, *record)
			mu.Unlock()
		}

		// Test valid configuration
		z, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			WithBatchSize(32).
			Build()

		if err != nil {
			t.Fatalf("Expected successful build, got error: %v", err)
		}
		if z == nil {
			t.Fatal("Expected non-nil ZephyrosLight")
		}

		// Verify configuration
		stats := z.Stats()
		if stats["buffer_size"] != 1024 {
			t.Errorf("Expected buffer size 1024, got %d", stats["buffer_size"])
		}
		if stats["batch_size"] != 32 {
			t.Errorf("Expected batch size 32, got %d", stats["batch_size"])
		}
	})

	t.Run("Builder_InvalidCapacity", func(t *testing.T) {
		processor := func(record *TestRecord) {}

		// Test invalid capacity (not power of two)
		_, err := NewBuilder[TestRecord](1000).
			WithProcessor(processor).
			Build()

		if err != ErrInvalidCapacity {
			t.Errorf("Expected ErrInvalidCapacity, got %v", err)
		}

		// Test zero capacity
		_, err = NewBuilder[TestRecord](0).
			WithProcessor(processor).
			Build()

		if err != ErrInvalidCapacity {
			t.Errorf("Expected ErrInvalidCapacity for zero capacity, got %v", err)
		}

		// Test negative capacity
		_, err = NewBuilder[TestRecord](-1).
			WithProcessor(processor).
			Build()

		if err != ErrInvalidCapacity {
			t.Errorf("Expected ErrInvalidCapacity for negative capacity, got %v", err)
		}
	})

	t.Run("Builder_MissingProcessor", func(t *testing.T) {
		// Test missing processor
		_, err := NewBuilder[TestRecord](1024).
			WithBatchSize(32).
			Build()

		if err != ErrMissingProcessor {
			t.Errorf("Expected ErrMissingProcessor, got %v", err)
		}
	})

	t.Run("Builder_InvalidBatchSize", func(t *testing.T) {
		processor := func(record *TestRecord) {}

		// Test zero batch size
		_, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			WithBatchSize(0).
			Build()

		if err != ErrInvalidBatchSize {
			t.Errorf("Expected ErrInvalidBatchSize for zero batch size, got %v", err)
		}

		// Test batch size exceeding capacity
		_, err = NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			WithBatchSize(2048).
			Build()

		if err != ErrInvalidBatchSize {
			t.Errorf("Expected ErrInvalidBatchSize for oversized batch, got %v", err)
		}
	})
}

// TestZephyrosLight_BuilderMethods tests the builder pattern methods
func TestZephyrosLight_BuilderMethods(t *testing.T) {
	t.Run("WithBackpressurePolicy", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex
		processor := func(record *TestRecord) {
			mu.Lock()
			processed = append(processed, *record)
			mu.Unlock()
		}

		// Test with DropOnFull policy
		z1, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			WithBackpressurePolicy(DropOnFull).
			Build()

		if err != nil {
			t.Fatalf("Expected successful build with DropOnFull, got error: %v", err)
		}
		z1.Close()

		// Test with BlockOnFull policy
		z2, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			WithBackpressurePolicy(BlockOnFull).
			Build()

		if err != nil {
			t.Fatalf("Expected successful build with BlockOnFull, got error: %v", err)
		}
		z2.Close()
	})

	t.Run("WithIdleStrategy", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex
		processor := func(record *TestRecord) {
			mu.Lock()
			processed = append(processed, *record)
			mu.Unlock()
		}

		// Test with different idle strategies
		strategies := []IdleStrategy{
			NewSpinningIdleStrategy(),
			NewSleepingIdleStrategy(1*time.Microsecond, 10),
			NewYieldingIdleStrategy(10),
			NewChannelIdleStrategy(1 * time.Microsecond),
			NewProgressiveIdleStrategy(),
		}

		for i, strategy := range strategies {
			z, err := NewBuilder[TestRecord](1024).
				WithProcessor(processor).
				WithIdleStrategy(strategy).
				Build()

			if err != nil {
				t.Fatalf("Expected successful build with strategy %d, got error: %v", i, err)
			}
			z.Close()
		}
	})
}

// TestZephyrosLight_BasicOperations tests basic write and process operations
func TestZephyrosLight_BasicOperations(t *testing.T) {
	t.Run("Write_And_Process", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex

		processor := func(record *TestRecord) {
			mu.Lock()
			processed = append(processed, *record)
			mu.Unlock()
		}

		z, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			WithBatchSize(10).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}

		// Write some records
		for i := 0; i < 5; i++ {
			success := z.Write(func(r *TestRecord) {
				r.ID = int64(i)
				r.Message = "test message"
				r.Value = i * 10
			})
			if !success {
				t.Errorf("Write %d failed", i)
			}
		}

		// Process records
		count := z.ProcessBatch()
		if count != 5 {
			t.Errorf("Expected 5 records processed, got %d", count)
		}

		// Verify processed records
		mu.Lock()
		if len(processed) != 5 {
			t.Errorf("Expected 5 processed records, got %d", len(processed))
		}
		for i, record := range processed {
			if record.ID != int64(i) {
				t.Errorf("Record %d: expected ID %d, got %d", i, i, record.ID)
			}
			if record.Message != "test message" {
				t.Errorf("Record %d: expected 'test message', got '%s'", i, record.Message)
			}
			if record.Value != i*10 {
				t.Errorf("Record %d: expected value %d, got %d", i, i*10, record.Value)
			}
		}
		mu.Unlock()
	})

	t.Run("Write_After_Close", func(t *testing.T) {
		processor := func(record *TestRecord) {}

		z, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}

		// Close the ring
		z.Close()

		// Try to write after close
		success := z.Write(func(r *TestRecord) {
			r.ID = 1
		})

		if success {
			t.Error("Expected write to fail after close")
		}

		// Check dropped count
		stats := z.Stats()
		if stats["items_dropped"] != 1 {
			t.Errorf("Expected 1 dropped item, got %d", stats["items_dropped"])
		}
	})

	t.Run("Buffer_Full_Behavior", func(t *testing.T) {
		processor := func(record *TestRecord) {}

		// Use very small buffer to test full condition
		z, err := NewBuilder[TestRecord](4).
			WithProcessor(processor).
			WithBatchSize(2). // Use smaller batch size for small buffer
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}

		// Fill buffer beyond capacity
		successCount := 0
		for i := 0; i < 10; i++ {
			success := z.Write(func(r *TestRecord) {
				r.ID = int64(i)
			})
			if success {
				successCount++
			}
		}

		// Should accept some writes but not all
		if successCount >= 10 {
			t.Error("Expected some writes to fail due to buffer full")
		}

		stats := z.Stats()
		if stats["items_dropped"] == 0 {
			t.Error("Expected some items to be dropped due to buffer full")
		}
	})
}

// TestZephyrosLight_BackpressurePolicies tests different backpressure behaviors
func TestZephyrosLight_BackpressurePolicies(t *testing.T) {
	t.Run("BlockOnFull_Policy", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex

		processor := func(record *TestRecord) {
			mu.Lock()
			defer mu.Unlock()
			processed = append(processed, *record)
			// Very minimal processing time to avoid deadlock
		}

		// Use small buffer with BlockOnFull policy to trigger writeBlockOnFull
		z, err := NewBuilder[TestRecord](4).
			WithProcessor(processor).
			WithBackpressurePolicy(BlockOnFull).
			WithBatchSize(1).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}
		defer z.Close()

		// Start processing loop
		go z.Loop()

		// Write records sequentially to test blocking behavior
		const numRecords = 8
		successCount := 0

		for i := 0; i < numRecords; i++ {
			success := z.Write(func(r *TestRecord) {
				r.ID = int64(i)
				r.Message = "blocking test"
				r.Value = i * 10
			})
			if success {
				successCount++
			}
			// Small delay to allow processing
			time.Sleep(1 * time.Millisecond)
		}

		// Give time for all records to be processed
		time.Sleep(50 * time.Millisecond)

		// With BlockOnFull, all writes should eventually succeed
		if successCount != numRecords {
			t.Errorf("Expected %d successful writes, got %d", numRecords, successCount)
		}

		stats := z.Stats()
		droppedCount := stats["items_dropped"]
		if droppedCount != 0 {
			t.Errorf("Expected 0 dropped items with BlockOnFull policy, got %d", droppedCount)
		}
	})

	t.Run("DropOnFull_Policy", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex

		processor := func(record *TestRecord) {
			mu.Lock()
			defer mu.Unlock()
			processed = append(processed, *record)
			// Simulate very slow processing to force buffer full
			time.Sleep(5 * time.Millisecond)
		}

		// Use small buffer with DropOnFull policy
		z, err := NewBuilder[TestRecord](4).
			WithProcessor(processor).
			WithBackpressurePolicy(DropOnFull).
			WithBatchSize(1).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}
		defer z.Close()

		// Start processing loop
		go z.Loop()

		// Write many records quickly to fill buffer and trigger drops
		const numRecords = 20
		successCount := 0
		for i := 0; i < numRecords; i++ {
			success := z.Write(func(r *TestRecord) {
				r.ID = int64(i)
				r.Message = "drop test"
				r.Value = i * 10
			})
			if success {
				successCount++
			}
		}

		// Give time for processing
		time.Sleep(100 * time.Millisecond)

		stats := z.Stats()
		droppedCount := stats["items_dropped"]

		// With DropOnFull policy, some items should be dropped
		if droppedCount == 0 {
			t.Error("Expected some items to be dropped with DropOnFull policy")
		}

		// Total processed + dropped should equal attempts
		mu.Lock()
		processedCount := len(processed)
		mu.Unlock()

		if int64(processedCount)+droppedCount > numRecords {
			t.Errorf("Processed (%d) + Dropped (%d) should not exceed total writes (%d)",
				processedCount, droppedCount, numRecords)
		}
	})
}

// TestZephyrosLight_Flush tests the Flush method
func TestZephyrosLight_Flush(t *testing.T) {
	t.Run("Flush_Operation", func(t *testing.T) {
		processor := func(record *TestRecord) {}

		z, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}

		// Write some data
		z.Write(func(r *TestRecord) {
			r.ID = 1
		})

		// Call flush (should not panic or error)
		_ = z.Flush()

		// Flush is a no-op in this implementation, so just verify it doesn't crash
		stats := z.Stats()
		if stats["items_buffered"] < 0 {
			t.Error("Flush should not produce negative buffered items")
		}
	})
}

// TestZephyrosLight_Stats tests the statistics functionality
func TestZephyrosLight_Stats(t *testing.T) {
	t.Run("Stats_Tracking", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex

		processor := func(record *TestRecord) {
			mu.Lock()
			processed = append(processed, *record)
			mu.Unlock()
		}

		z, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			WithBatchSize(5).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}

		// Initial stats
		stats := z.Stats()
		expectedStats := map[string]int64{
			"writer_position": 0,
			"reader_position": 0,
			"buffer_size":     1024,
			"items_buffered":  0,
			"items_processed": 0,
			"items_dropped":   0,
			"closed":          0,
			"batch_size":      5,
		}

		for key, expected := range expectedStats {
			if stats[key] != expected {
				t.Errorf("Initial stats[%s]: expected %d, got %d", key, expected, stats[key])
			}
		}

		// Write and process some items
		for i := 0; i < 3; i++ {
			z.Write(func(r *TestRecord) {
				r.ID = int64(i)
			})
		}

		stats = z.Stats()
		if stats["writer_position"] != 3 {
			t.Errorf("Expected writer_position 3, got %d", stats["writer_position"])
		}
		if stats["items_buffered"] != 3 {
			t.Errorf("Expected items_buffered 3, got %d", stats["items_buffered"])
		}

		// Process items
		z.ProcessBatch()

		stats = z.Stats()
		if stats["items_processed"] != 3 {
			t.Errorf("Expected items_processed 3, got %d", stats["items_processed"])
		}
		if stats["items_buffered"] != 0 {
			t.Errorf("Expected items_buffered 0 after processing, got %d", stats["items_buffered"])
		}

		// Close and check stats
		z.Close()
		stats = z.Stats()
		if stats["closed"] != 1 {
			t.Errorf("Expected closed 1, got %d", stats["closed"])
		}
	})
}

// TestZephyrosLight_Concurrent tests concurrent operations
func TestZephyrosLight_Concurrent(t *testing.T) {
	t.Run("Concurrent_Writers", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex

		processor := func(record *TestRecord) {
			mu.Lock()
			processed = append(processed, *record)
			mu.Unlock()
		}

		z, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			WithBatchSize(50).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}

		const writers = 10
		const itemsPerWriter = 100
		var wg sync.WaitGroup
		wg.Add(writers)

		// Launch concurrent writers
		for w := 0; w < writers; w++ {
			go func(writerID int) {
				defer wg.Done()
				for i := 0; i < itemsPerWriter; i++ {
					z.Write(func(r *TestRecord) {
						r.ID = int64(writerID*1000 + i)
						r.Message = "concurrent test"
						r.Value = writerID
					})
				}
			}(w)
		}

		wg.Wait()

		// Process all items
		totalProcessed := 0
		for {
			count := z.ProcessBatch()
			if count == 0 {
				break
			}
			totalProcessed += count
		}

		stats := z.Stats()
		expectedItems := int64(writers * itemsPerWriter)

		if stats["items_processed"] < expectedItems-stats["items_dropped"] {
			t.Errorf("Expected processed + dropped >= %d, got processed=%d, dropped=%d",
				expectedItems, stats["items_processed"], stats["items_dropped"])
		}

		// Verify some items were processed
		if totalProcessed == 0 {
			t.Error("Expected some items to be processed")
		}
	})
}

// TestZephyrosLight_LoopProcess tests the consumer loop
func TestZephyrosLight_LoopProcess(t *testing.T) {
	t.Run("LoopProcess_ShortRun", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex

		processor := func(record *TestRecord) {
			mu.Lock()
			processed = append(processed, *record)
			mu.Unlock()
		}

		z, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}

		// Start consumer loop in background
		go z.LoopProcess()

		// Write some items
		for i := 0; i < 5; i++ {
			z.Write(func(r *TestRecord) {
				r.ID = int64(i)
				r.Message = "loop test"
			})
		}

		// Give time for processing
		time.Sleep(10 * time.Millisecond)

		// Close and wait for final processing
		z.Close()
		time.Sleep(10 * time.Millisecond)

		// Verify items were processed
		mu.Lock()
		processedCount := len(processed)
		mu.Unlock()

		if processedCount != 5 {
			t.Errorf("Expected 5 items processed, got %d", processedCount)
		}
	})
}

// TestZephyrosLight_Loop tests the Loop method specifically
func TestZephyrosLight_Loop(t *testing.T) {
	t.Run("Loop_Processing", func(t *testing.T) {
		processed := make([]TestRecord, 0)
		var mu sync.Mutex

		processor := func(record *TestRecord) {
			mu.Lock()
			processed = append(processed, *record)
			mu.Unlock()
		}

		z, err := NewBuilder[TestRecord](1024).
			WithProcessor(processor).
			WithBatchSize(10).
			Build()

		if err != nil {
			t.Fatalf("Failed to create ZephyrosLight: %v", err)
		}

		// Start Loop() method (not LoopProcess)
		done := make(chan bool)
		go func() {
			z.Loop() // This calls LoopProcess internally
			done <- true
		}()

		// Write test data
		const numItems = 20
		for i := 0; i < numItems; i++ {
			success := z.Write(func(r *TestRecord) {
				r.ID = int64(i)
				r.Message = "loop method test"
				r.Value = i * 100
			})
			if !success {
				t.Errorf("Write %d failed", i)
			}
		}

		// Give time for processing
		time.Sleep(50 * time.Millisecond)

		// Close the ring buffer
		z.Close()

		// Wait for Loop to exit
		select {
		case <-done:
			// Loop exited as expected
		case <-time.After(1 * time.Second):
			t.Error("Loop() did not exit within timeout")
		}

		// Verify all items were processed
		mu.Lock()
		processedCount := len(processed)
		mu.Unlock()

		if processedCount != numItems {
			t.Errorf("Expected %d items processed, got %d", numItems, processedCount)
		}

		// Verify processed items are correct
		mu.Lock()
		for i, record := range processed {
			if record.Message != "loop method test" {
				t.Errorf("Record %d: expected message 'loop method test', got '%s'", i, record.Message)
			}
		}
		mu.Unlock()
	})
}

// TestZephyrosLight_ResetIdleStrategy tests Reset method on idle strategies
func TestZephyrosLight_ResetIdleStrategy(t *testing.T) {
	t.Run("Reset_IdleStrategy", func(t *testing.T) {
		// Test that Reset is called on idle strategies
		strategies := []IdleStrategy{
			NewSpinningIdleStrategy(),
			NewSleepingIdleStrategy(1*time.Microsecond, 10),
			NewYieldingIdleStrategy(10),
			NewChannelIdleStrategy(1 * time.Microsecond),
			NewProgressiveIdleStrategy(),
		}

		for i, strategy := range strategies {
			// Call Reset to ensure it doesn't panic
			strategy.Reset()

			// Verify String method works (this was also 0% coverage)
			strRepr := strategy.String()
			if strRepr == "" {
				t.Errorf("Strategy %d returned empty string representation", i)
			}
		}
	})
}
