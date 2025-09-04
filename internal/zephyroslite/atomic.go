// atomic.go: Simplified atomic operations for IRIS internal use
//
// This is a lightweight version of Zephyros atomic operations,
// embedded directly in IRIS to eliminate external dependencies
// while maintaining core performance characteristics.
//
// Features included:
//   - Basic atomic operations with cache-line padding
//   - Simplified padding (no advanced multi-level optimization)
//   - Essential operations only (Load, Store, Add, CompareAndSwap)
//
// Features removed (kept in commercial Zephyros):
//   - Advanced padding strategies
//   - Performance monitoring hooks
//   - Extended atomic operation variants
//   - Debug and profiling support
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"sync/atomic"
)

// AtomicPaddedInt64 provides cache-line padded atomic int64 operations
//
// This is a simplified version of the full Zephyros AtomicPaddedInt64,
// designed for IRIS internal use. It provides essential atomic operations
// with basic cache-line padding to prevent false sharing.
//
// Simplified Features:
//   - Standard 64-byte cache line padding
//   - Essential atomic operations (Load, Store, Add, CAS)
//   - Zero-allocation design
//
// Removed Features (commercial Zephyros only):
//   - Adaptive padding based on CPU architecture
//   - Performance monitoring and statistics
//   - Extended atomic operations (AddAndGet, etc.)
//   - Debug support and operation tracing
type AtomicPaddedInt64 struct {
	_   [64]byte // Pre-padding to prevent false sharing
	val int64    // Actual atomic value
	_   [64]byte // Post-padding to prevent false sharing
}

// Load atomically loads and returns the value
//
// This operation uses memory ordering semantics to ensure
// visibility across all CPU cores. It's optimized for high-frequency
// read operations in hot paths.
//
// Returns:
//   - int64: Current atomic value
//
// Performance: Sub-nanosecond operation with memory ordering guarantees
func (a *AtomicPaddedInt64) Load() int64 {
	return atomic.LoadInt64(&a.val)
}

// Store atomically stores the value
//
// This operation uses memory ordering semantics to ensure
// the value is visible to all CPU cores immediately.
//
// Parameters:
//   - val: Value to store atomically
//
// Performance: Sub-nanosecond operation with memory ordering guarantees
func (a *AtomicPaddedInt64) Store(val int64) {
	atomic.StoreInt64(&a.val, val)
}

// Add atomically adds delta to the value and returns the new value
//
// This operation is lock-free and provides atomic read-modify-write
// semantics. It's optimized for high-frequency increment/decrement
// operations.
//
// Parameters:
//   - delta: Value to add (can be negative for subtraction)
//
// Returns:
//   - int64: New value after addition
//
// Performance: Single atomic operation, typically 1-2 CPU cycles
func (a *AtomicPaddedInt64) Add(delta int64) int64 {
	return atomic.AddInt64(&a.val, delta)
}

// CompareAndSwap atomically compares and swaps the value
//
// This operation atomically compares the current value with 'old'
// and if they match, stores 'new'. It returns whether the swap
// was successful.
//
// Parameters:
//   - old: Expected current value
//   - new: New value to store if comparison succeeds
//
// Returns:
//   - bool: true if swap was successful, false if current value != old
//
// Performance: Single atomic operation, essential for lock-free algorithms
func (a *AtomicPaddedInt64) CompareAndSwap(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&a.val, old, new)
}

// helper functions for min/max operations (simplified versions)

// min returns the smaller of two int64 values
// Performance: Inlined for hot path usage, essential for batch processing
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
