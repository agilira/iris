// xantos.go: Ultra-high performance lock-free ring buffer
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package notus

import (
	"runtime"
	"sync/atomic"
	"time"
)

// Performance-critical constants based on empirical optimization
const (
	// OPT-002: Reader cache refresh strategy
	// Refresh cached reader every 32 writes using bit masking.
	// 32 is an optimal power-of-two that balances the cost of atomic Load()
	// with the risk of operating on stale reader position data.
	readerCacheRefreshMask = 31 // 32 - 1 for bitwise AND operation

	// OPT-002: Spin-then-yield strategy
	// Spin for 16384 iterations before yielding to scheduler.
	// This value provides optimal balance between CPU efficiency and latency
	// for SPSC workloads based on extensive empirical testing.
	spinYieldMask = (1 << 14) - 1 // Bitmask for branchless modulo (16384 iterations)
)

// ProcessorFunc is the ultra-fast processing function signature
type ProcessorFunc[T any] func(*T)

// Notus is the ultra-high performance SPSC processor
type Notus[T any] struct {
	// Ring buffer core
	buffer   []T
	capacity int64
	mask     int64 // capacity - 1 for bit masking

	// Writer state (SPSC optimized)
	writerCursor   AtomicPaddedInt64 // Published sequence
	writerPosition PaddedInt64       // Cached writer position

	// Reader state (SPSC optimized)
	readerCursor AtomicPaddedInt64 // Reader sequence
	cachedReader int64             // Cached reader position (SPSC optimization)

	// SPSC performance counters
	batchPublishSize int64 // Batch publish threshold

	// Processor function
	processor ProcessorFunc[T]

	// Batching configuration
	batchSize int64

	// Control
	closed AtomicPaddedInt64 // 0 = open, 1 = closed

	// Cache line padding to prevent false sharing
	_ [64]byte
}

// NewNotus creates a new ultra-high performance processor
// Deprecated: Use NewBuilder instead for better configuration
func NewNotus[T any](capacity int64, processor ProcessorFunc[T]) (*Notus[T], error) {
	return NewBuilder[T](capacity).WithProcessor(processor).Build()
}

// Write publishes a new item to the ring buffer (ULTRA-OPTIMIZED SPSC path)
func (n *Notus[T]) Write(writerFunc func(*T)) bool {
	// SAFETY: Keep closed check for production reliability
	if n.closed.Load() != 0 {
		return false
	}

	// OPTIMIZATION: Direct atomic read + increment (eliminates temp variable)
	nextPos := atomic.LoadInt64(&n.writerPosition.Value) + 1

	// OPTIMIZATION 2: Position-based refresh (no counter overhead)
	// Refresh cached reader every 32 writes using bit masking
	if nextPos&readerCacheRefreshMask == 0 {
		n.cachedReader = n.readerCursor.Load()
	}

	// OPTIMIZATION 3: Single capacity check with bit operations
	// Combine the two capacity checks into one and use mask operation
	if nextPos-n.cachedReader > n.capacity {
		// Only refresh if really needed
		fresh := n.readerCursor.Load()
		if nextPos-fresh > n.capacity {
			return false
		}
		n.cachedReader = fresh
	}

	// OPTIMIZATION 4: Direct slot calculation with pre-computed mask
	slot := &n.buffer[nextPos&n.mask]

	writerFunc(slot)

	// OPTIMIZATION 7: Update position atomically after function call
	atomic.StoreInt64(&n.writerPosition.Value, nextPos)

	// OPTIMIZATION 8: Optimized batch publish logic
	if n.batchPublishSize == 1 || nextPos&(n.batchPublishSize-1) == 0 {
		n.writerCursor.Store(nextPos)
	}

	return true
}

// Flush ensures all pending writes are visible to reader
func (n *Notus[T]) Flush() {
	// OPTIMIZATION: Direct atomic read and store
	currentPos := atomic.LoadInt64(&n.writerPosition.Value)
	n.writerCursor.Store(currentPos)
}

// ProcessBatch processes items in a single batch (ULTRA-OPTIMIZED SPSC)
func (n *Notus[T]) ProcessBatch() int {
	current := n.readerCursor.Load()
	available := n.writerCursor.Load()

	// OPTIMIZATION 1: Early exit with single comparison
	if available <= current {
		return 0
	}

	count := available - current

	// OPTIMIZATION: Branchless minimum calculation (inspired by five-vee-disruptor)
	// Replace: if count > n.batchSize { count = n.batchSize }
	diff := count - n.batchSize
	branchlessMask := diff >> 63 // 0 if count >= batchSize, -1 if count < batchSize
	count = n.batchSize + (diff & branchlessMask)

	// OPTIMIZATION 2: Local variable caching for hot path
	buffer := n.buffer
	mask := n.mask
	processor := n.processor

	// OPTIMIZATION 3: Aggressive unrolling for small batches (most common case)
	if count == 1 {
		// Single item - most common case, zero loop overhead
		nextIdx := (current + 1) & mask
		processor(&buffer[nextIdx])
	} else if count <= 4 {
		// Small batch unrolling for maximum speed
		seq := current + 1
		for i := int64(0); i < count; i++ {
			processor(&buffer[(seq+i)&mask])
		}
	} else {
		// OPTIMIZATION 4: Regular loop for larger batches
		endSeq := current + count
		for seq := current + 1; seq <= endSeq; seq++ {
			processor(&buffer[seq&mask])
		}
	}

	// OPTIMIZATION 5: Single atomic store at the end
	newReaderPos := current + count
	n.readerCursor.Store(newReaderPos)
	return int(count)
} // LoopProcess continuously processes items (blocking)
func (n *Notus[T]) LoopProcess() {
	spins := 0

	for n.closed.Load() == 0 {
		if n.ProcessBatch() == 0 {
			// OPTIMIZATION: Branchless yield strategy from five-vee-disruptor
			spins++
			if spins&spinYieldMask == 0 {
				// Yield CPU to other goroutines after optimal spin iterations
				runtime.Gosched()
			}
		} else {
			spins = 0 // Reset spin counter when work is found
		}
	}

	// Process remaining items after close - ensure ALL items are processed
	// Keep trying until we're absolutely sure there's nothing left
	consecutiveEmpty := 0
	for consecutiveEmpty < 3 { // Require 3 consecutive empty attempts
		n.Flush() // Always ensure writes are visible first
		processed := n.ProcessBatch()
		if processed > 0 {
			// Found work, reset empty counter and continue
			consecutiveEmpty = 0
			continue
		}

		// No work found, increment empty counter
		consecutiveEmpty++

		// Wait a bit for any in-flight writes to complete
		time.Sleep(time.Microsecond)

		// Try one more time after wait
		if n.ProcessBatch() > 0 {
			consecutiveEmpty = 0 // Reset if we found work after wait
		}
	}
}

// Close gracefully shuts down the processor
func (n *Notus[T]) Close() {
	n.closed.Store(1)
	n.Flush() // Ensure all writes are published
}

// Stats returns performance statistics
func (n *Notus[T]) Stats() map[string]int64 {
	writerPos := n.writerCursor.Load()
	readerPos := n.readerCursor.Load()

	return map[string]int64{
		"writer_position": writerPos,
		"reader_position": readerPos,
		"buffer_size":     n.capacity,
		"items_buffered":  writerPos - readerPos,
		"closed":          n.closed.Load(),
	}
}
