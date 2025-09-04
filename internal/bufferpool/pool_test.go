// pool_test.go: Comprehensive test suite for buffer pool
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package bufferpool

import (
	"bytes"
	"os"
	"sync"
	"testing"
)

// TestGetReturnsCleanBuffer tests that Get() returns a clean buffer
func TestGetReturnsCleanBuffer(t *testing.T) {
	ResetStats()

	buf := Get()
	if buf == nil {
		t.Fatal("Get() returned nil buffer")
	}

	if buf.Len() != 0 {
		t.Errorf("Expected clean buffer with len=0, got len=%d", buf.Len())
	}

	if buf.Cap() < DefaultCapacity {
		t.Errorf("Expected buffer capacity >= %d, got %d", DefaultCapacity, buf.Cap())
	}

	Put(buf)
}

// TestPutWithNilBuffer tests that Put() handles nil gracefully
func TestPutWithNilBuffer(t *testing.T) {
	ResetStats()

	// Should not panic
	Put(nil)

	stats := GetStats()
	if stats.Puts != 0 {
		t.Errorf("Expected 0 puts with nil buffer, got %d", stats.Puts)
	}
}

// TestBufferReuse tests that buffers are properly reused
func TestBufferReuse(t *testing.T) {
	ResetStats()

	// Get and put a buffer
	buf1 := Get()
	buf1.WriteString("test data")
	Put(buf1)

	// Get another buffer - should be reused
	buf2 := Get()

	// Should be clean even though we wrote to it before
	if buf2.Len() != 0 {
		t.Errorf("Reused buffer should be clean, got len=%d", buf2.Len())
	}

	Put(buf2)

	stats := GetStats()
	if stats.Gets != 2 {
		t.Errorf("Expected 2 gets, got %d", stats.Gets)
	}
	if stats.Puts != 2 {
		t.Errorf("Expected 2 puts, got %d", stats.Puts)
	}
}

// TestOversizedBufferDrop tests that oversized buffers are dropped
func TestOversizedBufferDrop(t *testing.T) {
	ResetStats()

	buf := Get()

	// Make buffer oversized by writing large amount of data
	largeData := make([]byte, MaxBufferSize+1)
	buf.Write(largeData)

	if buf.Cap() <= MaxBufferSize {
		t.Skipf("Buffer didn't grow as expected, cap=%d", buf.Cap())
	}

	Put(buf)

	stats := GetStats()
	if stats.Drops != 1 {
		t.Errorf("Expected 1 drop for oversized buffer, got %d", stats.Drops)
	}

	// Get another buffer - should be fresh due to drop
	buf2 := Get()
	if buf2.Cap() > MaxBufferSize {
		t.Errorf("New buffer after drop should be normal size, got cap=%d", buf2.Cap())
	}

	Put(buf2)
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	ResetStats()

	const numGoroutines = 100
	const opsPerGoroutine = 50

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				buf := Get()
				buf.WriteString("concurrent test data")
				Put(buf)
			}
		}()
	}

	wg.Wait()

	stats := GetStats()
	expectedOps := int64(numGoroutines * opsPerGoroutine)

	if stats.Gets != expectedOps {
		t.Errorf("Expected %d gets, got %d", expectedOps, stats.Gets)
	}

	if stats.Puts != expectedOps {
		t.Errorf("Expected %d puts, got %d", expectedOps, stats.Puts)
	}

	// Should have fewer allocations than operations due to reuse
	if stats.Allocations >= expectedOps {
		t.Errorf("Expected allocations < %d, got %d", expectedOps, stats.Allocations)
	}
}

// TestStatsAccuracy tests that statistics are accurate
func TestStatsAccuracy(t *testing.T) {
	ResetStats()

	// Verify reset worked
	stats := GetStats()
	if stats.Gets != 0 || stats.Puts != 0 || stats.Allocations != 0 || stats.Drops != 0 {
		t.Errorf("Stats not reset properly: %+v", stats)
	}

	// Perform operations and check stats
	buf1 := Get()
	buf2 := Get()

	stats = GetStats()
	if stats.Gets != 2 {
		t.Errorf("Expected 2 gets, got %d", stats.Gets)
	}

	Put(buf1)
	Put(buf2)

	stats = GetStats()
	if stats.Puts != 2 {
		t.Errorf("Expected 2 puts, got %d", stats.Puts)
	}
}

// TestBufferCapacityGrowth tests buffer growth behavior
func TestBufferCapacityGrowth(t *testing.T) {
	ResetStats()

	buf := Get()
	initialCap := buf.Cap()

	// Write data to force growth
	data := make([]byte, initialCap*2)
	buf.Write(data)

	if buf.Cap() <= initialCap {
		t.Errorf("Buffer should have grown, initial=%d, current=%d", initialCap, buf.Cap())
	}

	Put(buf)

	// Get another buffer - capacity behavior depends on whether it was dropped
	buf2 := Get()
	Put(buf2)
}

// TestDefaultCapacity tests that new buffers have expected capacity
func TestDefaultCapacity(t *testing.T) {
	ResetStats()

	// Force allocation of new buffer
	bufs := make([]*bytes.Buffer, 10)
	for i := range bufs {
		bufs[i] = Get()
	}

	// Check that at least one has expected capacity
	foundExpectedCap := false
	for _, buf := range bufs {
		if buf.Cap() >= DefaultCapacity {
			foundExpectedCap = true
			break
		}
	}

	if !foundExpectedCap {
		t.Errorf("Expected at least one buffer with capacity >= %d", DefaultCapacity)
	}

	for _, buf := range bufs {
		Put(buf)
	}
}

// TestMaxBufferSizeConstant tests the MaxBufferSize constant
func TestMaxBufferSizeConstant(t *testing.T) {
	if MaxBufferSize != 1<<20 {
		t.Errorf("MaxBufferSize should be 1 MiB (1048576), got %d", MaxBufferSize)
	}
}

// TestDefaultCapacityConstant tests the DefaultCapacity constant
func TestDefaultCapacityConstant(t *testing.T) {
	if DefaultCapacity != 512 {
		t.Errorf("DefaultCapacity should be 512, got %d", DefaultCapacity)
	}
}

// TestPoolEfficiency tests that the pool actually reduces allocations
func TestPoolEfficiency(t *testing.T) {
	// Skip efficiency tests on CI where GC behavior varies significantly
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Pool efficiency tests disabled on CI due to GC unpredictability")
	}

	ResetStats()

	const numOps = 100

	// Perform many operations
	for i := 0; i < numOps; i++ {
		buf := Get()
		buf.WriteString("efficiency test")
		Put(buf)
	}

	stats := GetStats()

	// Buffer pool efficiency can vary due to:
	// - GC emptying the pool unexpectedly
	// - Initial pool warmup allocations
	// - Memory pressure causing pool drainage
	// - Runtime sync.Pool behavior variations
	// Conservative threshold: should prevent most allocations but allow GC variability
	efficiencyRatio := float64(stats.Allocations) / float64(stats.Gets)
	if efficiencyRatio > 0.5 { // 50% tolerance for GC and pool warmup
		t.Errorf("Pool critically inefficient: %.2f allocations per get (expected < 0.5)", efficiencyRatio)
	} else if efficiencyRatio > 0.2 {
		t.Logf("Pool efficiency below optimal: %.2f allocations per get (target < 0.1)", efficiencyRatio)
	}

	t.Logf("Pool efficiency: %.4f allocations per get (%d allocs for %d gets)",
		efficiencyRatio, stats.Allocations, stats.Gets)
}
