// ring_test.go: Comprehensive tests for ultra-high performance ring buffer
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agilira/iris/internal/zephyroslite"
)

// Helper function for creating rings with default idle strategy in tests
func newTestRing(capacity, batchSize int64, processor ProcessorFunc) (*Ring, error) {
	return newRing(capacity, batchSize, SingleRing, 1, zephyroslite.DropOnFull, BalancedStrategy, processor)
}

func TestNewRing_ValidConfiguration(t *testing.T) {
	processed := int64(0)
	processor := func(r *Record) {
		atomic.AddInt64(&processed, 1)
	}

	ring, err := newTestRing(1024, 128, processor)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if ring == nil {
		t.Fatal("Expected ring to be created")
	}

	stats := ring.Stats()
	if stats["capacity"] != 1024 {
		t.Errorf("Expected capacity 1024, got %d", stats["capacity"])
	}
	if stats["batch_size"] != 128 {
		t.Errorf("Expected batch size 128, got %d", stats["batch_size"])
	}

	ring.Close()
}

func TestNewRing_AutoBatchSizing(t *testing.T) {
	processor := func(r *Record) {}

	testCases := []struct {
		capacity      int64
		expectedBatch int64
	}{
		{64, 16},    // capacity / 4 for small buffers
		{256, 64},   // capacity / 4 for medium buffers
		{1024, 256}, // optimal for larger buffers
		{4096, 256}, // optimal for larger buffers
	}

	for _, tc := range testCases {
		ring, err := newTestRing(tc.capacity, 0, processor) // 0 = auto-size
		if err != nil {
			t.Fatalf("Expected no error for capacity %d, got: %v", tc.capacity, err)
		}

		stats := ring.Stats()
		if stats["batch_size"] != tc.expectedBatch {
			t.Errorf("For capacity %d, expected batch size %d, got %d",
				tc.capacity, tc.expectedBatch, stats["batch_size"])
		}

		ring.Close()
	}
}

func TestNewRing_InvalidCapacity(t *testing.T) {
	processor := func(r *Record) {}

	invalidCapacities := []int64{0, -1, 3, 5, 6, 7, 9, 15, 17, 100, 1000}

	for _, capacity := range invalidCapacities {
		ring, err := newRing(capacity, 64, SingleRing, 1, zephyroslite.DropOnFull, BalancedStrategy, processor)
		if err == nil {
			t.Errorf("Expected error for invalid capacity %d, got nil", capacity)
			if ring != nil {
				ring.Close()
			}
		}
		if ring != nil {
			t.Errorf("Expected nil ring for invalid capacity %d", capacity)
		}
	}
}

func TestNewRing_InvalidBatchSize(t *testing.T) {
	processor := func(r *Record) {}

	testCases := []struct {
		capacity  int64
		batchSize int64
	}{
		{1024, 0},    // Invalid: zero
		{1024, -1},   // Invalid: negative
		{1024, 2048}, // Invalid: exceeds capacity
	}

	for _, tc := range testCases {
		ring, err := newRing(tc.capacity, tc.batchSize, SingleRing, 1, zephyroslite.DropOnFull, BalancedStrategy, processor)
		if tc.batchSize == 0 {
			// Zero batch size should auto-size, not error
			if err != nil {
				t.Errorf("Expected no error for auto-sizing batch, got: %v", err)
			}
		} else {
			if err == nil {
				t.Errorf("Expected error for invalid batch size %d, got nil", tc.batchSize)
				if ring != nil {
					ring.Close()
				}
			}
		}
	}
}

func TestNewRing_MissingProcessor(t *testing.T) {
	ring, err := newRing(1024, 128, SingleRing, 1, zephyroslite.DropOnFull, BalancedStrategy, nil)
	if err == nil {
		t.Error("Expected error for missing processor, got nil")
		if ring != nil {
			ring.Close()
		}
	}
	if ring != nil {
		t.Error("Expected nil ring for missing processor")
	}
}

func TestRing_Write(t *testing.T) {
	processed := make([]*Record, 0, 100)
	var mu sync.Mutex

	processor := func(r *Record) {
		mu.Lock()
		// Create a copy since the original might be reused
		recordCopy := *r
		processed = append(processed, &recordCopy)
		mu.Unlock()
	}

	ring, err := newTestRing(64, 16, processor)
	if err != nil {
		t.Fatalf("Failed to create ring: %v", err)
	}
	defer ring.Close()

	// Start processing
	go ring.Loop()

	// Write test records
	testRecords := []struct {
		level   Level
		message string
	}{
		{Info, "Test message 1"},
		{Warn, "Test message 2"},
		{Error, "Test message 3"},
	}

	for i, tr := range testRecords {
		success := ring.Write(func(r *Record) {
			r.Level = tr.level
			r.Msg = tr.message
		})
		if !success {
			t.Errorf("Failed to write record %d", i)
		}
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	processedCount := len(processed)
	mu.Unlock()

	if processedCount != len(testRecords) {
		t.Errorf("Expected %d processed records, got %d", len(testRecords), processedCount)
	}
}

func TestRing_HighThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high throughput test in short mode")
	}

	processed := int64(0)
	processor := func(r *Record) {
		atomic.AddInt64(&processed, 1)
	}

	ring, err := newTestRing(8192, 512, processor)
	if err != nil {
		t.Fatalf("Failed to create ring: %v", err)
	}
	defer ring.Close()

	// Start processing
	go ring.Loop()

	// High throughput write test
	const numRecords = 1000
	const numWorkers = 10

	var wg sync.WaitGroup
	recordsPerWorker := numRecords / numWorkers

	start := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < recordsPerWorker; j++ {
				for {
					success := ring.Write(func(r *Record) {
						r.Level = Info
						r.Msg = "High throughput test"
					})
					if success {
						break
					}
					// If ring is full, yield and retry
					time.Sleep(time.Microsecond)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Wait for all records to be processed
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt64(&processed) >= numRecords {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	finalProcessed := atomic.LoadInt64(&processed)
	t.Logf("Processed %d/%d records in %v", finalProcessed, numRecords, elapsed)
	t.Logf("Throughput: %.2f records/sec", float64(finalProcessed)/elapsed.Seconds())

	if finalProcessed < int64(numRecords)*9/10 { // Allow 10% loss due to ring being full
		t.Errorf("Processed too few records: %d/%d", finalProcessed, numRecords)
	}
}

func TestRing_Stats(t *testing.T) {
	processor := func(r *Record) {
		time.Sleep(time.Millisecond) // Slow processing to build up buffer
	}

	ring, err := newTestRing(256, 32, processor)
	if err != nil {
		t.Fatalf("Failed to create ring: %v", err)
	}
	defer ring.Close()

	// Start processing
	go ring.Loop()

	// Write some records
	for i := 0; i < 10; i++ {
		ring.Write(func(r *Record) {
			r.Level = Info
			r.Msg = "Test"
		})
	}

	stats := ring.Stats()

	// Verify expected stats fields
	expectedFields := []string{
		"writer_position", "reader_position", "buffer_size",
		"items_buffered", "closed", "capacity", "batch_size",
		"utilization_percent", "go_routines",
	}

	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Missing stats field: %s", field)
		}
	}

	// Verify basic stats values
	if stats["capacity"] != 256 {
		t.Errorf("Expected capacity 256, got %d", stats["capacity"])
	}
	if stats["batch_size"] != 32 {
		t.Errorf("Expected batch size 32, got %d", stats["batch_size"])
	}
	if stats["buffer_size"] != 256 {
		t.Errorf("Expected buffer size 256, got %d", stats["buffer_size"])
	}
}

func TestRing_Flush(t *testing.T) {
	processed := int64(0)
	processor := func(r *Record) {
		atomic.AddInt64(&processed, 1)
	}

	ring, err := newTestRing(64, 16, processor)
	if err != nil {
		t.Fatalf("Failed to create ring: %v", err)
	}
	defer ring.Close()

	// Start processing
	go ring.Loop()

	// Write a record
	ring.Write(func(r *Record) {
		r.Level = Info
		r.Msg = "Test flush"
	})

	// Flush to ensure visibility
	_ = ring.Flush()

	// Flush should be safe to call multiple times
	_ = ring.Flush()
	_ = ring.Flush()
}

func TestRing_ProcessBatch(t *testing.T) {
	processed := int64(0)
	processor := func(r *Record) {
		atomic.AddInt64(&processed, 1)
	}

	ring, err := newTestRing(64, 8, processor)
	if err != nil {
		t.Fatalf("Failed to create ring: %v", err)
	}
	defer ring.Close()

	// Write multiple records
	for i := 0; i < 5; i++ {
		ring.Write(func(r *Record) {
			r.Level = Info
			r.Msg = "Batch test"
		})
	}

	// Process batch manually
	count := ring.ProcessBatch()
	if count == 0 {
		t.Error("Expected to process at least some records")
	}

	if atomic.LoadInt64(&processed) == 0 {
		t.Error("Expected processor to be called")
	}
}

func TestRing_CloseGracefully(t *testing.T) {
	processed := int64(0)
	processor := func(r *Record) {
		atomic.AddInt64(&processed, 1)
		time.Sleep(time.Millisecond) // Simulate processing time
	}

	ring, err := newTestRing(64, 16, processor)
	if err != nil {
		t.Fatalf("Failed to create ring: %v", err)
	}

	// Start processing
	go ring.Loop()

	// Write records
	const numRecords = 20
	for i := 0; i < numRecords; i++ {
		ring.Write(func(r *Record) {
			r.Level = Info
			r.Msg = "Graceful close test"
		})
	}

	// Close and verify all records are processed
	ring.Close()

	// Wait a bit for final processing
	time.Sleep(100 * time.Millisecond)

	finalProcessed := atomic.LoadInt64(&processed)
	if finalProcessed != numRecords {
		t.Errorf("Expected %d processed records after close, got %d", numRecords, finalProcessed)
	}

	// Verify writes fail after close
	success := ring.Write(func(r *Record) {
		r.Level = Info
		r.Msg = "Should fail"
	})
	if success {
		t.Error("Expected write to fail after close")
	}
}
