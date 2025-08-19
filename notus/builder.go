// builder.go: Builder pattern for NOTUS configuration
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package notus

import (
	"fmt"
)

var (
	// ErrCapacity is returned when capacity is not a power of two
	ErrCapacity = fmt.Errorf("capacity must be a power of two")

	// ErrMissingProcessor is returned when no processor function is provided
	ErrMissingProcessor = fmt.Errorf("missing processor function")
)

// Builder builds a NOTUS disruptor with fluent configuration
type Builder[T any] struct {
	capacity         int64
	processor        ProcessorFunc[T]
	batchSize        int64 // Batch size for processing
	batchPublishSize int64 // Batch publish size for SPSC optimization
} // NewBuilder creates a new NOTUS builder with the specified capacity
func NewBuilder[T any](capacity int64) *Builder[T] {
	// OPTIMIZATION: Intelligent default batch size based on capacity
	defaultBatchSize := int64(64) // Safe default
	if capacity >= 1024 {
		defaultBatchSize = 256 // Optimal for larger buffers
	} else if capacity >= 64 {
		defaultBatchSize = 16 // Appropriate for small buffers
	} else if capacity < 64 {
		defaultBatchSize = 1 // Minimal for very small buffers
	}

	// OPTIMIZATION: Intelligent default batch publish size
	defaultBatchPublishSize := int64(64)
	if capacity < 64 {
		// Find largest power-of-2 that is reasonable for small buffers
		defaultBatchPublishSize = 1
		for defaultBatchPublishSize*4 <= capacity {
			defaultBatchPublishSize *= 2
		}
		// Ensure at least size 1, max size 8 for small buffers
		if defaultBatchPublishSize > 8 {
			defaultBatchPublishSize = 8
		}
	}

	return &Builder[T]{
		capacity:         capacity,
		batchSize:        defaultBatchSize,
		batchPublishSize: defaultBatchPublishSize,
	}
}

// WithProcessor sets the processing function for items
func (b *Builder[T]) WithProcessor(processor ProcessorFunc[T]) *Builder[T] {
	b.processor = processor
	return b
}

// WithBatchSize sets the batch size for processing
func (b *Builder[T]) WithBatchSize(batchSize int64) *Builder[T] {
	b.batchSize = batchSize
	return b
}

// WithBatchPublishSize sets the batch publish size for SPSC optimization
// Lower values = lower latency, higher values = higher throughput
func (b *Builder[T]) WithBatchPublishSize(size int64) *Builder[T] {
	b.batchPublishSize = size
	return b
}

// Build creates and initializes a new NOTUS disruptor
func (b *Builder[T]) Build() (*Notus[T], error) {
	// OPTIMIZATION: Single capacity validation with bit operations
	if b.capacity <= 0 || (b.capacity&(b.capacity-1)) != 0 {
		return nil, ErrCapacity
	}

	// OPTIMIZATION: Early validation to avoid allocation if invalid
	if b.processor == nil {
		return nil, ErrMissingProcessor
	}

	// OPTIMIZATION: Combined validation checks
	if b.batchSize <= 0 || b.batchSize > b.capacity {
		if b.batchSize <= 0 {
			return nil, fmt.Errorf("batch size must be positive, got %d", b.batchSize)
		}
		return nil, fmt.Errorf("batch size (%d) cannot exceed capacity (%d)", b.batchSize, b.capacity)
	}

	// OPTIMIZATION: Power-of-2 check using bit manipulation for publish size
	if b.batchPublishSize > 1 && (b.batchPublishSize&(b.batchPublishSize-1)) != 0 {
		return nil, fmt.Errorf("batch publish size must be power of 2, got %d", b.batchPublishSize)
	}

	// OPTIMIZATION: Pre-compute mask during construction
	mask := b.capacity - 1

	n := &Notus[T]{
		buffer:           make([]T, b.capacity),
		capacity:         b.capacity,
		mask:             mask,
		processor:        b.processor,
		batchSize:        b.batchSize,
		batchPublishSize: b.batchPublishSize,
		cachedReader:     -1, // Initialize SPSC cache
	}

	// OPTIMIZATION: Batch initialize atomic fields for better cache locality
	n.writerCursor.Store(-1)
	n.readerCursor.Store(-1)
	n.closed.Store(0)
	n.writerPosition.Value = -1

	return n, nil
}
