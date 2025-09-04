// zephyroslite.go: Simplified MPSC ring buffer for IRIS internal use
//
// This is a lightweight version of Zephyros MPSC ring buffer,
// embedded directly in IRIS to eliminate external dependencies
// while maintaining core performance characteristics.
//
// Features included:
//   - Lock-free MPSC ring buffer
//   - Zero-allocation write operations
//   - Basic atomic operations with cache-line padding
//   - Fixed batch processing (no adaptive batching)
//   - Essential performance monitoring
//
// Features removed (kept in commercial Zephyros):
//   - Dynamic adaptive batching
//   - Advanced performance statistics
//   - Multi-ring ThreadedZephyros architecture
//   - Gemini strategy optimizations
//   - Extended monitoring and profiling
//   - Advanced idle strategies
//
// Performance target: ~15-20ns/op (vs 9ns commercial, 25ns current IRIS)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"fmt"
	"runtime"
	"time"
)

// ProcessorFunc is the processing function signature for log records
type ProcessorFunc[T any] func(*T)

// BackpressurePolicy defines how to handle ring buffer overflow
type BackpressurePolicy int

const (
	// DropOnFull drops new records when buffer is full (default)
	// Best for: High-performance applications, ad servers, real-time systems
	// Trade-off: Maximum performance, some log loss acceptable
	DropOnFull BackpressurePolicy = iota

	// BlockOnFull blocks the caller until buffer space is available
	// Best for: Audit systems, financial transactions, compliance logging
	// Trade-off: Guaranteed delivery, potential performance impact
	BlockOnFull
)

// String returns a string representation of the BackpressurePolicy
func (bp BackpressurePolicy) String() string {
	switch bp {
	case DropOnFull:
		return "DropOnFull"
	case BlockOnFull:
		return "BlockOnFull"
	default:
		return "Unknown"
	}
}

// ZephyrosLight is the simplified MPSC lock-free ring buffer for IRIS
//
// This implementation focuses on core MPSC performance while removing
// advanced features that are reserved for commercial Zephyros.
//
// Core Features:
//   - Lock-free MPSC operations
//   - Zero-allocation write path
//   - Cache-line padded atomic operations
//   - Fixed batch processing
//   - Essential statistics
//
// Simplified Design:
//   - Single ring only (no ThreadedZephyros)
//   - Fixed batch size (no adaptive batching)
//   - Basic padding (no advanced CPU-specific optimizations)
//   - Simplified idle strategy (no complex spinning algorithms)
type ZephyrosLight[T any] struct {
	// Ring buffer core
	buffer   []T
	capacity int64
	mask     int64 // capacity - 1 for bit masking

	// MPSC atomic cursors (cache-line padded)
	writerCursor AtomicPaddedInt64 // Producer claim sequence
	readerCursor AtomicPaddedInt64 // Consumer sequence

	// Availability tracking for MPSC coordination
	availableBuffer []AtomicPaddedInt64 // Per-slot availability markers

	// Configuration
	processor          ProcessorFunc[T]
	batchSize          int64
	backpressurePolicy BackpressurePolicy
	idleStrategy       IdleStrategy

	// Control
	closed AtomicPaddedInt64 // 0 = open, 1 = closed

	// Basic statistics (simplified)
	processed AtomicPaddedInt64 // Total processed count
	dropped   AtomicPaddedInt64 // Total dropped count

	// Cache line padding to prevent false sharing
	_ [64]byte
}

// Builder provides a fluent interface for creating ZephyrosLight instances
type Builder[T any] struct {
	capacity           int64
	processor          ProcessorFunc[T]
	batchSize          int64
	backpressurePolicy BackpressurePolicy
	idleStrategy       IdleStrategy
}

// NewBuilder creates a new builder for ZephyrosLight with specified capacity
//
// Parameters:
//   - capacity: Ring buffer size (must be power of two, e.g., 1024, 2048, 4096)
//
// Returns:
//   - *Builder[T]: Builder instance for fluent configuration
func NewBuilder[T any](capacity int64) *Builder[T] {
	return &Builder[T]{
		capacity:           capacity,
		batchSize:          64,         // Reasonable default, simpler than adaptive
		backpressurePolicy: DropOnFull, // Default to high-performance behavior
	}
}

// WithProcessor sets the processor function for log records
//
// Parameters:
//   - processor: Function to process each log record
//
// Returns:
//   - *Builder[T]: Builder instance for method chaining
func (b *Builder[T]) WithProcessor(processor ProcessorFunc[T]) *Builder[T] {
	b.processor = processor
	return b
}

// WithBatchSize sets the fixed batch size for processing
//
// Unlike commercial Zephyros, this uses a fixed batch size rather
// than adaptive batching for simplicity.
//
// Parameters:
//   - batchSize: Fixed number of items to process per batch
//
// Returns:
//   - *Builder[T]: Builder instance for method chaining
func (b *Builder[T]) WithBatchSize(batchSize int64) *Builder[T] {
	b.batchSize = batchSize
	return b
}

// WithBackpressurePolicy sets the behavior when the ring buffer is full
//
// Parameters:
//   - policy: DropOnFull (default) for maximum performance, BlockOnFull for guaranteed delivery
//
// Examples:
//   - DropOnFull: High-performance services, ad servers, real-time systems
//   - BlockOnFull: Audit systems, financial transactions, compliance logging
//
// Returns:
//   - *Builder[T]: Builder instance for method chaining
func (b *Builder[T]) WithBackpressurePolicy(policy BackpressurePolicy) *Builder[T] {
	b.backpressurePolicy = policy
	return b
}

// WithIdleStrategy sets the CPU usage strategy when no work is available
//
// Parameters:
//   - strategy: IdleStrategy implementation controlling CPU usage vs latency trade-offs
//
// Available strategies:
//   - NewSpinningIdleStrategy(): Ultra-low latency, ~100% CPU usage
//   - NewSleepingIdleStrategy(): Balanced CPU/latency, ~1-10% CPU usage
//   - NewYieldingIdleStrategy(): Moderate reduction, ~10-50% CPU usage
//   - NewChannelIdleStrategy(): Minimal CPU usage, ~microsecond latency
//   - NewProgressiveIdleStrategy(): Adaptive strategy for variable workloads
//
// Returns:
//   - *Builder[T]: Builder instance for method chaining
func (b *Builder[T]) WithIdleStrategy(strategy IdleStrategy) *Builder[T] {
	b.idleStrategy = strategy
	return b
}

// Build creates and initializes the ZephyrosLight ring buffer
//
// Returns:
//   - *ZephyrosLight[T]: Configured ring buffer ready for use
//   - error: Configuration validation error
func (b *Builder[T]) Build() (*ZephyrosLight[T], error) {
	// Validate capacity (must be power of two)
	if b.capacity <= 0 || (b.capacity&(b.capacity-1)) != 0 {
		return nil, ErrInvalidCapacity
	}

	// Validate processor
	if b.processor == nil {
		return nil, ErrMissingProcessor
	}

	// Validate batch size
	if b.batchSize <= 0 || b.batchSize > b.capacity {
		return nil, ErrInvalidBatchSize
	}

	// Default idle strategy to progressive for backward compatibility
	idleStrategy := b.idleStrategy
	if idleStrategy == nil {
		idleStrategy = NewProgressiveIdleStrategy() // Balanced default
	}

	// Create ring buffer
	z := &ZephyrosLight[T]{
		buffer:             make([]T, b.capacity),
		capacity:           b.capacity,
		mask:               b.capacity - 1,
		availableBuffer:    make([]AtomicPaddedInt64, b.capacity),
		processor:          b.processor,
		batchSize:          b.batchSize,
		backpressurePolicy: b.backpressurePolicy,
		idleStrategy:       idleStrategy,
	}

	// Initialize availability markers to invalid sequence
	for i := range z.availableBuffer {
		z.availableBuffer[i].Store(-1)
	}

	return z, nil
}

// Write adds an item to the ring buffer using zero-allocation pattern
//
// The behavior when the buffer is full depends on the configured BackpressurePolicy:
//   - DropOnFull: Returns false immediately (default, high-performance)
//   - BlockOnFull: Blocks until space becomes available (guaranteed delivery)
//
// Multiple producers can call this concurrently in both modes.
//
// Parameters:
//   - writerFunc: Function to populate the allocated slot (zero allocations)
//
// Returns:
//   - bool: true if successfully written, false if dropped or closed
//
// Performance: Target ~15-20ns/op (simplified vs 9ns commercial Zephyros)
func (z *ZephyrosLight[T]) Write(writerFunc func(*T)) bool {
	// Quick closed check
	if z.closed.Load() != 0 {
		z.dropped.Add(1)
		return false
	}

	switch z.backpressurePolicy {
	case DropOnFull:
		return z.writeDropOnFull(writerFunc)
	case BlockOnFull:
		return z.writeBlockOnFull(writerFunc)
	default:
		// Fallback to drop behavior for unknown policies
		return z.writeDropOnFull(writerFunc)
	}
}

// writeDropOnFull implements the original non-blocking behavior
func (z *ZephyrosLight[T]) writeDropOnFull(writerFunc func(*T)) bool {
	// MPSC: Claim sequence number atomically
	sequence := z.writerCursor.Add(1) - 1

	// Check if we're about to lap the reader (buffer full check)
	if sequence >= z.readerCursor.Load()+z.capacity {
		// Buffer full - drop the message
		z.dropped.Add(1)
		return false
	}

	// Write to allocated slot
	slot := &z.buffer[sequence&z.mask]
	writerFunc(slot)

	// Mark slot as available for reading
	z.availableBuffer[sequence&z.mask].Store(sequence)

	return true
}

// writeBlockOnFull implements blocking behavior for guaranteed delivery
func (z *ZephyrosLight[T]) writeBlockOnFull(writerFunc func(*T)) bool {
	// Block until we can successfully write or the ring is closed
	for {
		// Check if closed before each attempt
		if z.closed.Load() != 0 {
			z.dropped.Add(1)
			return false
		}

		// MPSC: Claim sequence number atomically
		sequence := z.writerCursor.Add(1) - 1

		// Check if we're about to lap the reader (buffer full check)
		currentReader := z.readerCursor.Load()
		if sequence < currentReader+z.capacity {
			// Space available - write the message
			slot := &z.buffer[sequence&z.mask]
			writerFunc(slot)

			// Mark slot as available for reading
			z.availableBuffer[sequence&z.mask].Store(sequence)

			return true
		}

		// Buffer full - yield and retry
		// We need to "rollback" the sequence claim since we can't use it
		// Note: This is a simplification - a full implementation would use
		// more sophisticated coordination to avoid sequence number waste
		runtime.Gosched()

		// Small delay to prevent tight spinning
		time.Sleep(time.Microsecond)
	}
}

// ProcessBatch processes available items in a single batch
//
// This is a simplified version that uses fixed batch size rather than
// the dynamic adaptive batching available in commercial Zephyros.
//
// Returns:
//   - int: Number of items processed in this batch
//
// Performance: Optimized for zero-allocation batch processing
func (z *ZephyrosLight[T]) ProcessBatch() int {
	current := z.readerCursor.Load()
	writerPos := z.writerCursor.Load()

	if current >= writerPos {
		return 0 // Nothing to process
	}

	// Use fixed batch size (simplified vs adaptive)
	maxProcess := min(z.batchSize, writerPos-current)

	// Scan for contiguous available sequences
	available := current - 1
	maxScan := current + maxProcess

	for seq := current; seq < maxScan; seq++ {
		if z.availableBuffer[seq&z.mask].Load() == seq {
			available = seq
		} else {
			break // Stop at first gap
		}
	}

	if available < current {
		return 0 // No contiguous sequence found
	}

	// Process the batch
	processed := int(available - current + 1)

	for seq := current; seq <= available; seq++ {
		idx := seq & z.mask
		z.processor(&z.buffer[idx])
		z.availableBuffer[idx].Store(-1) // Reset availability
	}

	// Update reader position
	z.readerCursor.Store(available + 1)
	z.processed.Add(int64(processed))

	return processed
}

// LoopProcess runs the consumer loop with configurable idle strategy
//
// This uses the configured IdleStrategy to control CPU usage when no work
// is available, providing different trade-offs between latency and CPU consumption.
func (z *ZephyrosLight[T]) LoopProcess() {
	for z.closed.Load() == 0 {
		processed := z.ProcessBatch()

		if processed > 0 {
			// Work found - reset idle strategy state
			z.idleStrategy.Reset()
		} else {
			// No work available - use idle strategy
			if !z.idleStrategy.Idle() {
				// Strategy indicates we should check for shutdown
				continue
			}
		}
	}

	// Final drain on close - process remaining items
	for z.ProcessBatch() > 0 {
		// Keep processing until empty
	}
}

// Close stops the processing loop and marks the ring as closed.
//
// The method is idempotent and thread-safe. After Close() is called,
// all Write() operations will return false and no new items will be processed.
func (z *ZephyrosLight[T]) Close() {
	z.closed.Store(1)
}

// Loop is an alias for LoopProcess for backward compatibility
// This method runs the consumer loop in the background
func (z *ZephyrosLight[T]) Loop() {
	z.LoopProcess()
}

// Flush waits for all messages currently in the ring buffer to be processed.
// Returns when all pending messages have been written to the output.
//
// WARNING: This method can block if the consumer is not running.
// Always ensure the consumer loop is active before calling Flush().
func (z *ZephyrosLight[T]) Flush() error {
	// Get current writer position - this is our target
	targetPosition := z.writerCursor.Load()

	// If nothing to flush, return immediately
	if targetPosition == 0 {
		return nil
	}

	// For DropOnFull policy, we still wait for all accepted messages to be processed
	// The policy affects write behavior, not flush behavior
	if z.backpressurePolicy == DropOnFull {
		// Use precise counting like BlockOnFull, but with more lenient timing
		initialProcessed := z.processed.Load()
		currentReader := z.readerCursor.Load()

		// Calculate how many items are pending processing
		pendingCount := targetPosition - currentReader
		if pendingCount <= 0 {
			return nil // Nothing pending
		}

		targetProcessed := initialProcessed + pendingCount
		timeout := time.Now().Add(3 * time.Second) // Slightly longer timeout

		for time.Now().Before(timeout) {
			currentProcessed := z.processed.Load()

			// Check if we've processed all items
			if currentProcessed >= targetProcessed {
				return nil
			}

			runtime.Gosched()
			time.Sleep(1 * time.Millisecond) // Faster polling for DropOnFull
		}

		// For DropOnFull, timeout is still an error since all accepted messages should be processed
		currentReader = z.readerCursor.Load()
		currentProcessed := z.processed.Load()
		return fmt.Errorf("flush timeout (DropOnFull): target_pos=%d, reader_pos=%d, target_processed=%d, current_processed=%d",
			targetPosition, currentReader, targetProcessed, currentProcessed)
	}

	// For BlockOnFull policy, use precise counting since no messages should be dropped
	// Get current processed count - we need to wait for this many more to be processed
	initialProcessed := z.processed.Load()
	currentReader := z.readerCursor.Load()

	// Calculate how many items are pending processing
	pendingCount := targetPosition - currentReader
	if pendingCount <= 0 {
		return nil // Nothing pending
	}

	targetProcessed := initialProcessed + int64(pendingCount)

	timeout := time.Now().Add(5 * time.Second)

	for time.Now().Before(timeout) {
		currentProcessed := z.processed.Load()

		// Check if we've processed enough items
		if currentProcessed >= targetProcessed {
			return nil
		}

		// Progressive backoff
		runtime.Gosched()
		time.Sleep(100 * time.Microsecond)
	}

	// Timeout occurred
	currentReader = z.readerCursor.Load()
	currentProcessed := z.processed.Load()
	return fmt.Errorf("flush timeout: target_pos=%d, reader_pos=%d, target_processed=%d, current_processed=%d",
		targetPosition, currentReader, targetProcessed, currentProcessed)
}

// Stats returns basic performance statistics
//
// This provides essential metrics without the comprehensive
// monitoring available in commercial Zephyros.
//
// Returns:
//   - map[string]int64: Basic performance metrics
func (z *ZephyrosLight[T]) Stats() map[string]int64 {
	writerPos := z.writerCursor.Load()
	readerPos := z.readerCursor.Load()

	return map[string]int64{
		"writer_position": writerPos,
		"reader_position": readerPos,
		"buffer_size":     z.capacity,
		"items_buffered":  writerPos - readerPos,
		"items_processed": z.processed.Load(),
		"items_dropped":   z.dropped.Load(),
		"closed":          z.closed.Load(),
		"batch_size":      z.batchSize,
	}
}
