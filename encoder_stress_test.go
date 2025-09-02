// encoder_stress_test.go: Stress tests for encoder performance and stability
//
// Tests stability and performance under high load, concurrent access,
// and memory pressure to identify regressions and bottlenecks that may
// not emerge in standard tests.
//
// Copyright (c) 2025 AGILira
// Series: IRIS Logging Library - Performance Optimization
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestTextEncoder_StressPerformance tests stability
func TestTextEncoder_StressPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Skip performance tests on CI where resources are throttled and unpredictable
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Stress tests disabled on CI due to resource variability")
	}

	encoder := NewTextEncoder()

	// Record with multiple field types for realistic stress
	record := &Record{
		Level:  Warn,
		Msg:    "High frequency stress test message",
		Logger: "stress.test",
		Caller: "stress_test.go:42",
		fields: [32]Field{},
		n:      0,
	}

	// Mix of field types that stress different code paths
	record.fields[0] = Str("component", "authentication")
	record.fields[1] = Secret("api_key", "sk-1234567890abcdef")
	record.fields[2] = Int64("user_id", 987654321)
	record.fields[3] = Float64("latency_ms", 45.678)
	record.fields[4] = Bool("success", true)
	record.fields[5] = Str("trace_id", "abc123def456ghi789")
	record.fields[6] = Int64("memory_mb", 512)
	record.fields[7] = Float64("cpu_percent", 23.45)
	record.n = 8

	now := time.Now()

	// stress test: 100k ops to verify stability
	t.Log("Starting stress test: 100,000 encoding operations")

	var buf bytes.Buffer
	start := time.Now()
	iterations := 100000

	// Monitoring allocations before the test
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < iterations; i++ {
		buf.Reset()
		encoder.Encode(record, now, &buf)

		// Every 10k iterations, check for memory leaks
		if i%10000 == 0 && i > 0 {
			runtime.GC()
			var mid runtime.MemStats
			runtime.ReadMemStats(&mid)
			if mid.Alloc > m1.Alloc*2 {
				t.Errorf("Potential memory leak detected at iteration %d: %d bytes", i, mid.Alloc)
			}
		}
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&m2)

	nanosPerOp := duration.Nanoseconds() / int64(iterations)
	t.Logf("Stress test completed: %d ns/op (%d iterations in %v)", nanosPerOp, iterations, duration)
	t.Logf("Memory delta: %d bytes allocated", m2.TotalAlloc-m1.TotalAlloc)

	// Ring buffer performance varies significantly under stress due to:
	// - GC pressure from 100k allocations
	// - CPU context switching
	// - Memory subsystem contention
	// - Scheduler variability under load
	// Conservative threshold: operations should complete in reasonable time
	if nanosPerOp > 10000 {
		t.Errorf("TextEncoder critically slow under stress: %d ns/op (expected <10μs)", nanosPerOp)
	} else if nanosPerOp > 5000 {
		t.Logf("TextEncoder stress performance below optimal: %d ns/op (target <3μs)", nanosPerOp)
	}
}

// TestTextEncoder_ConcurrentStress tests performance under concurrent access
func TestTextEncoder_ConcurrentStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent stress test in short mode")
	}

	// Skip performance tests on CI where resources are throttled and unpredictable
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Concurrent stress tests disabled on CI due to scheduler variability")
	}

	const numGoroutines = 10
	const operationsPerGoroutine = 5000

	encoder := NewTextEncoder()

	// Channels for coordination
	start := make(chan struct{})
	done := make(chan time.Duration, numGoroutines)

	// Record template
	createRecord := func() *Record {
		record := &Record{
			Level:  Info,
			Msg:    "Concurrent access stress test",
			fields: [32]Field{},
			n:      0,
		}
		record.fields[0] = Str("goroutine", "worker")
		record.fields[1] = Int64("ops", int64(operationsPerGoroutine))
		record.fields[2] = Bool("concurrent", true)
		record.n = 3
		return record
	}

	// Launch workers
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			record := createRecord()
			record.fields[3] = Int64("worker_id", int64(workerID))
			record.n = 4

			now := time.Now()
			var buf bytes.Buffer

			// Wait for start signal
			<-start

			workerStart := time.Now()
			for j := 0; j < operationsPerGoroutine; j++ {
				buf.Reset()
				encoder.Encode(record, now, &buf)
			}
			workerDuration := time.Since(workerStart)
			done <- workerDuration
		}(i)
	}

	t.Logf("Starting concurrent stress test: %d goroutines × %d operations",
		numGoroutines, operationsPerGoroutine)

	// Start all workers simultaneously
	testStart := time.Now()
	close(start)

	// Wait for completion and collect results
	go func() {
		wg.Wait()
		close(done)
	}()

	var totalDuration time.Duration
	var maxDuration time.Duration
	var minDuration time.Duration = time.Hour

	for workerDuration := range done {
		totalDuration += workerDuration
		if workerDuration > maxDuration {
			maxDuration = workerDuration
		}
		if workerDuration < minDuration {
			minDuration = workerDuration
		}
	}

	testDuration := time.Since(testStart)
	avgDuration := totalDuration / numGoroutines
	totalOps := numGoroutines * operationsPerGoroutine

	t.Logf("Concurrent stress test completed in %v", testDuration)
	t.Logf("Total operations: %d", totalOps)
	t.Logf("Worker durations - Avg: %v, Min: %v, Max: %v", avgDuration, minDuration, maxDuration)
	t.Logf("Throughput: %.0f ops/sec", float64(totalOps)/testDuration.Seconds())

	// Ring buffer concurrent performance varies due to:
	// - Goroutine scheduler contention
	// - CPU cache line bouncing
	// - Memory allocator synchronization
	// - Runtime system overhead
	// Conservative threshold for concurrent scenarios
	avgNanosPerOp := avgDuration.Nanoseconds() / operationsPerGoroutine
	if avgNanosPerOp > 25000 { // 25μs per op under heavy concurrency
		t.Errorf("Concurrent performance critically degraded: %d ns/op average", avgNanosPerOp)
	} else if avgNanosPerOp > 10000 {
		t.Logf("Concurrent performance below optimal: %d ns/op average (target <5μs)", avgNanosPerOp)
	}

	// Fairness check: scheduler should be reasonably fair even under load
	if maxDuration > minDuration*5 { // Increased tolerance for scheduler variations
		t.Errorf("Unfair scheduling detected: max %v vs min %v (ratio >5x)", maxDuration, minDuration)
	}
}

// TestTextEncoder_MemoryStress tests for memory leaks under heavy load
func TestTextEncoder_MemoryStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory stress test in short mode")
	}

	// Skip performance tests on CI where GC behavior is unpredictable
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Memory stress tests disabled on CI due to GC variability")
	}

	encoder := NewTextEncoder()

	// Record with fields that might cause allocations
	record := &Record{
		Level:  Error,
		Msg:    "Memory stress test with large content",
		Logger: "memory.stress.test.with.long.name",
		Caller: "very/long/path/to/memory_stress_test.go:123",
		fields: [32]Field{},
		n:      0,
	}

	// Fields with varying content to stress allocations
	record.fields[0] = Str("large_field", "This is a long string that might cause allocations and memory pressure during encoding operations")
	record.fields[1] = Secret("sensitive", "Another potentially long sensitive value that gets redacted")
	record.fields[2] = Str("json_like", `{"key":"value","nested":{"array":[1,2,3],"bool":true}}`)
	record.fields[3] = Int64("timestamp_ns", time.Now().UnixNano())
	record.n = 4

	now := time.Now()

	// Baseline memory measurement
	runtime.GC()
	runtime.GC() // Double GC
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	iterations := 50000
	buffers := make([]*bytes.Buffer, 10) // Pool of buffers to avoid contention
	for i := range buffers {
		buffers[i] = &bytes.Buffer{}
	}

	t.Logf("Starting memory stress test: %d iterations", iterations)

	for i := 0; i < iterations; i++ {
		buf := buffers[i%len(buffers)]
		buf.Reset()
		encoder.Encode(record, now, buf)

		// Periodic memory checks
		if i%5000 == 0 && i > 0 {
			runtime.GC()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			growth := float64(m.Alloc) / float64(m1.Alloc)
			if growth > 1.5 { // More than 50% growth
				t.Logf("Memory growth at iteration %d: %.2fx baseline (%d bytes)",
					i, growth, m.Alloc)
			}
		}
	}

	// Final memory measurement
	runtime.GC()
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	t.Logf("Memory stress test completed")
	t.Logf("Initial memory: %d bytes", m1.Alloc)
	t.Logf("Final memory: %d bytes", m2.Alloc)
	t.Logf("Total allocated during test: %d bytes", m2.TotalAlloc-m1.TotalAlloc)

	// Memory leak check: final memory should not be significantly higher
	if m2.Alloc > m1.Alloc*2 {
		t.Errorf("Potential memory leak: final memory %d bytes vs initial %d bytes",
			m2.Alloc, m1.Alloc)
	}
}
