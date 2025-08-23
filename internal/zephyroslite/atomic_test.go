// atomic_test.go: Tests for atomic operations in zephyroslite
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"sync"
	"testing"
)

// TestAtomicPaddedInt64_BasicOperations tests basic atomic operations
func TestAtomicPaddedInt64_BasicOperations(t *testing.T) {
	t.Run("Load_Store", func(t *testing.T) {
		var atomic AtomicPaddedInt64

		// Test initial value
		if atomic.Load() != 0 {
			t.Errorf("Expected initial value 0, got %d", atomic.Load())
		}

		// Test store and load
		atomic.Store(42)
		if atomic.Load() != 42 {
			t.Errorf("Expected 42 after store, got %d", atomic.Load())
		}

		// Test store negative value
		atomic.Store(-100)
		if atomic.Load() != -100 {
			t.Errorf("Expected -100 after store, got %d", atomic.Load())
		}
	})

	t.Run("Add_Operations", func(t *testing.T) {
		var atomic AtomicPaddedInt64

		// Test add positive
		result := atomic.Add(10)
		if result != 10 {
			t.Errorf("Expected Add(10) to return 10, got %d", result)
		}
		if atomic.Load() != 10 {
			t.Errorf("Expected value 10 after Add(10), got %d", atomic.Load())
		}

		// Test add negative (subtract)
		result = atomic.Add(-5)
		if result != 5 {
			t.Errorf("Expected Add(-5) to return 5, got %d", result)
		}
		if atomic.Load() != 5 {
			t.Errorf("Expected value 5 after Add(-5), got %d", atomic.Load())
		}

		// Test add zero
		result = atomic.Add(0)
		if result != 5 {
			t.Errorf("Expected Add(0) to return 5, got %d", result)
		}
	})

	t.Run("CompareAndSwap", func(t *testing.T) {
		var atomic AtomicPaddedInt64
		atomic.Store(100)

		// Test successful CAS
		success := atomic.CompareAndSwap(100, 200)
		if !success {
			t.Error("Expected CompareAndSwap(100, 200) to succeed")
		}
		if atomic.Load() != 200 {
			t.Errorf("Expected value 200 after successful CAS, got %d", atomic.Load())
		}

		// Test failed CAS
		success = atomic.CompareAndSwap(100, 300)
		if success {
			t.Error("Expected CompareAndSwap(100, 300) to fail")
		}
		if atomic.Load() != 200 {
			t.Errorf("Expected value unchanged at 200 after failed CAS, got %d", atomic.Load())
		}

		// Test CAS with same value
		success = atomic.CompareAndSwap(200, 200)
		if !success {
			t.Error("Expected CompareAndSwap(200, 200) to succeed")
		}
		if atomic.Load() != 200 {
			t.Errorf("Expected value 200 after CAS with same value, got %d", atomic.Load())
		}
	})
}

// TestAtomicPaddedInt64_Concurrent tests thread safety
func TestAtomicPaddedInt64_Concurrent(t *testing.T) {
	t.Run("Concurrent_Add", func(t *testing.T) {
		var atomic AtomicPaddedInt64
		const goroutines = 100
		const iterations = 1000

		var wg sync.WaitGroup
		wg.Add(goroutines)

		// Launch multiple goroutines doing Add operations
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					atomic.Add(1)
				}
			}()
		}

		wg.Wait()

		expected := int64(goroutines * iterations)
		if atomic.Load() != expected {
			t.Errorf("Expected %d after concurrent adds, got %d", expected, atomic.Load())
		}
	})

	t.Run("Concurrent_CAS", func(t *testing.T) {
		var atomic AtomicPaddedInt64
		const goroutines = 50
		var successCount int64
		var successAtomic AtomicPaddedInt64

		var wg sync.WaitGroup
		wg.Add(goroutines)

		// Launch multiple goroutines trying to CAS from 0 to their ID
		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				if atomic.CompareAndSwap(0, int64(id)) {
					successAtomic.Add(1)
				}
			}(i)
		}

		wg.Wait()

		// Exactly one should succeed
		successCount = successAtomic.Load()
		if successCount != 1 {
			t.Errorf("Expected exactly 1 successful CAS, got %d", successCount)
		}

		// Value should be one of the IDs
		value := atomic.Load()
		if value < 0 || value >= goroutines {
			t.Errorf("Expected value between 0 and %d, got %d", goroutines-1, value)
		}
	})
}

// TestAtomicPaddedInt64_MemoryLayout tests memory alignment
func TestAtomicPaddedInt64_MemoryLayout(t *testing.T) {
	t.Run("Memory_Alignment", func(t *testing.T) {
		// Create multiple atomic values to test false sharing prevention
		atomics := make([]AtomicPaddedInt64, 10)

		// Basic operations on all atomics
		for i := range atomics {
			atomics[i].Store(int64(i * 10))
		}

		// Verify values are independent
		for i := range atomics {
			expected := int64(i * 10)
			if atomics[i].Load() != expected {
				t.Errorf("Atomic %d: expected %d, got %d", i, expected, atomics[i].Load())
			}
		}

		// Concurrent operations to test false sharing prevention
		var wg sync.WaitGroup
		wg.Add(len(atomics))

		for i := range atomics {
			go func(idx int) {
				defer wg.Done()
				for j := 0; j < 1000; j++ {
					atomics[idx].Add(1)
				}
			}(i)
		}

		wg.Wait()

		// Verify final values
		for i := range atomics {
			expected := int64(i*10 + 1000)
			if atomics[i].Load() != expected {
				t.Errorf("Atomic %d: expected %d after concurrent adds, got %d", i, expected, atomics[i].Load())
			}
		}
	})
}

// TestMin tests the min helper function
func TestMin(t *testing.T) {
	t.Run("Min_Function", func(t *testing.T) {
		tests := []struct {
			a, b, expected int64
		}{
			{1, 2, 1},
			{2, 1, 1},
			{5, 5, 5},
			{-1, 1, -1},
			{-5, -10, -10},
			{0, 0, 0},
			{100, 50, 50},
		}

		for _, test := range tests {
			result := min(test.a, test.b)
			if result != test.expected {
				t.Errorf("min(%d, %d): expected %d, got %d", test.a, test.b, test.expected, result)
			}
		}
	})
}
