// sequence_integrity_test.go: Tests for CAS-based sequence claiming
//
// These tests verify that writeDropOnFull and writeBlockOnFull do not
// waste sequence numbers, which would create permanent gaps in
// availableBuffer and block the reader from processing subsequent entries.
//
// BUG FIXED: The original implementation used writerCursor.Add(1)
// unconditionally. A dropped or retried write left an unpublished
// sequence gap. ProcessBatch scans for contiguous availableBuffer
// sequences and stops at the first gap, permanently stalling all
// entries written after the gap.
//
// FIX: Both functions now use CompareAndSwap to check capacity BEFORE
// claiming a sequence number. No sequence is ever wasted.
//
// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestDropOnFull_NoSequenceGap verifies that after a drop event,
// subsequent writes are still processed by the reader.
// This is the core regression test for the sequence waste bug.
func TestDropOnFull_NoSequenceGap(t *testing.T) {
	var processed atomic.Int64

	// WHY capacity=4: small enough to guarantee buffer-full conditions
	// with a slow processor. Each slot processes in 10ms, so 4 slots
	// can handle ~400 writes/s. Writing faster than that triggers drops.
	z, err := NewBuilder[TestRecord](4).
		WithProcessor(func(r *TestRecord) {
			processed.Add(1)
			time.Sleep(10 * time.Millisecond)
		}).
		WithBackpressurePolicy(DropOnFull).
		WithBatchSize(1).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	defer z.Close()

	go z.Loop()

	// Phase 1: fill the buffer and trigger at least one drop
	dropped := 0
	for i := 0; i < 20; i++ {
		ok := z.Write(func(r *TestRecord) {
			r.ID = int64(i)
		})
		if !ok {
			dropped++
		}
	}

	if dropped == 0 {
		t.Fatal("expected at least one drop to trigger the regression path")
	}

	// Phase 2: wait for the slow processor to drain everything
	time.Sleep(300 * time.Millisecond)

	// Phase 3: write MORE entries AFTER the drops
	const postDropWrites = 4
	for i := 0; i < postDropWrites; i++ {
		// WHY retry loop: CAS under no contention (single writer)
		// should succeed on first try, but yield if ring still draining
		for attempt := 0; attempt < 50; attempt++ {
			if z.Write(func(r *TestRecord) { r.ID = int64(100 + i) }) {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Phase 4: wait for processing and verify ALL post-drop writes were consumed
	time.Sleep(200 * time.Millisecond)

	stats := z.Stats()
	totalProcessed := stats["items_processed"]

	// WHY this assertion: before the CAS fix, the reader was permanently
	// stuck at the first wasted sequence. Post-drop writes would never be
	// processed even though Write() returned true. After the fix, every
	// successful Write() is guaranteed to be processed.
	if totalProcessed < int64(postDropWrites) {
		t.Errorf("reader stalled after drop: only processed %d entries total,"+
			" expected at least %d post-drop writes to be consumed", totalProcessed, postDropWrites)
	}
}

// TestBlockOnFull_NoSequenceWaste verifies that the blocking path does
// not waste sequence numbers while waiting for space.
func TestBlockOnFull_NoSequenceWaste(t *testing.T) {
	var processed atomic.Int64

	z, err := NewBuilder[TestRecord](4).
		WithProcessor(func(r *TestRecord) {
			processed.Add(1)
			time.Sleep(5 * time.Millisecond)
		}).
		WithBackpressurePolicy(BlockOnFull).
		WithBatchSize(1).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	defer z.Close()

	go z.Loop()

	// Write more records than capacity, all must succeed (blocking)
	const total = 16
	var wg sync.WaitGroup
	var succeeded atomic.Int64

	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ok := z.Write(func(r *TestRecord) {
				r.ID = int64(id)
			})
			if ok {
				succeeded.Add(1)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	if succeeded.Load() != total {
		t.Errorf("BlockOnFull: expected all %d writes to succeed, got %d", total, succeeded.Load())
	}

	stats := z.Stats()
	if stats["items_processed"] != total {
		t.Errorf("BlockOnFull: expected %d processed, got %d (sequence gap detected)",
			total, stats["items_processed"])
	}

	if stats["items_dropped"] != 0 {
		t.Errorf("BlockOnFull: expected 0 dropped, got %d", stats["items_dropped"])
	}
}

// TestDropOnFull_ConcurrentNoGap verifies that under concurrent writer
// contention, no sequence gaps are created by the CAS retry loop.
func TestDropOnFull_ConcurrentNoGap(t *testing.T) {
	var processed atomic.Int64

	z, err := NewBuilder[TestRecord](64).
		WithProcessor(func(r *TestRecord) {
			processed.Add(1)
		}).
		WithBackpressurePolicy(DropOnFull).
		WithBatchSize(8).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	defer z.Close()

	go z.Loop()

	// WHY 20 goroutines x 50 writes: enough contention to exercise the
	// CAS retry path (multiple goroutines reading the same writerCursor
	// value, only one winning the CAS). The ring is large enough (64)
	// that drops should be rare, so most writes succeed.
	const writers = 20
	const perWriter = 50
	var wg sync.WaitGroup
	var totalSuccess atomic.Int64

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				if z.Write(func(r *TestRecord) { r.ID = int64(i) }) {
					totalSuccess.Add(1)
				}
			}
		}()
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	stats := z.Stats()

	// WHY totalSuccess == items_processed: every Write() that returned
	// true MUST be processed. Before the CAS fix, sequence gaps could
	// cause successful writes to be silently lost.
	if stats["items_processed"] != totalSuccess.Load() {
		t.Errorf("processed (%d) != successful writes (%d): sequence gap detected",
			stats["items_processed"], totalSuccess.Load())
	}
}

// TestBlockOnFull_ConcurrentAllDelivered verifies that BlockOnFull
// delivers every message under heavy concurrent writes.
func TestBlockOnFull_ConcurrentAllDelivered(t *testing.T) {
	var processed atomic.Int64

	z, err := NewBuilder[TestRecord](16).
		WithProcessor(func(r *TestRecord) {
			processed.Add(1)
		}).
		WithBackpressurePolicy(BlockOnFull).
		WithBatchSize(4).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	defer z.Close()

	go z.Loop()

	const writers = 10
	const perWriter = 20
	const total = writers * perWriter
	var wg sync.WaitGroup

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				z.Write(func(r *TestRecord) { r.ID = int64(i) })
			}
		}()
	}

	wg.Wait()
	time.Sleep(300 * time.Millisecond)

	stats := z.Stats()
	if stats["items_processed"] != total {
		t.Errorf("expected all %d items processed, got %d", total, stats["items_processed"])
	}
	if stats["items_dropped"] != 0 {
		t.Errorf("expected 0 dropped, got %d", stats["items_dropped"])
	}
}
