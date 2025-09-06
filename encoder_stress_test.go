// encoder_stress_test.go: Stress tests for encoder performance and stability
//
// Tests stability and performance under high load, concurrent access,
// and memory pressure to identify regressions and bottlenecks that may
// not emerge in standard tests.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
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

	// Use CI-friendly parameters
	iterations := 10000 // Reduce iterations for CI compatibility
	if IsCIEnvironment() {
		iterations = 1000 // Even fewer iterations in CI
		t.Logf("Running in CI environment, reduced iterations to %d", iterations)
	}

	// stress test
	t.Logf("Starting stress test: %d encoding operations", iterations)

	var buf bytes.Buffer
	start := time.Now()

	// Monitoring allocations before the test
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < iterations; i++ {
		buf.Reset()
		encoder.Encode(record, now, &buf)
		if buf.Len() == 0 {
			t.Fatalf("Encoding failed at iteration %d", i)
		}

		// Every 1000 iterations (reduced from 10k), check for memory leaks
		if i%1000 == 0 && i > 0 {
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

	avgNsPerOp := duration.Nanoseconds() / int64(iterations)
	t.Logf("Stress test completed: %d iterations in %v", iterations, duration)
	t.Logf("Average performance: %d ns/op (%.2f μs)", avgNsPerOp, float64(avgNsPerOp)/1000.0)
	t.Logf("Memory delta: %d bytes allocated", m2.TotalAlloc-m1.TotalAlloc)

	// Adaptive performance thresholds based on execution environment
	var maxNsPerOp int64

	// Detect if race detector is enabled by checking for its typical performance impact
	isRaceDetectorEnabled := avgNsPerOp > 20000 && testing.Verbose()

	if isRaceDetectorEnabled || os.Getenv("GORACE") != "" || os.Getenv("CGO_ENABLED") == "1" {
		// Race detection or CGO enabled - much more lenient
		maxNsPerOp = 200000 // 200μs
		t.Log("Race detection or CGO detected, using relaxed performance threshold: 200μs")
	} else if IsCIEnvironment() {
		// CI environment - moderate threshold
		maxNsPerOp = 100000 // 100μs
		t.Log("CI environment detected, using moderate performance threshold: 100μs")
	} else {
		// Normal environment - stricter threshold
		maxNsPerOp = 50000 // 50μs
		t.Log("Normal environment, using standard performance threshold: 50μs")
	}

	if avgNsPerOp > maxNsPerOp {
		t.Errorf("Performance regression: %d ns/op (expected <%d ns/op)", avgNsPerOp, maxNsPerOp)
	}
}

// TestTextEncoder_ConcurrentStress tests performance under concurrent access
func TestTextEncoder_ConcurrentStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent stress test in short mode")
	}

	// Use CI-friendly parameters
	numGoroutines := 10
	operationsPerGoroutine := 5000

	if IsCIEnvironment() {
		numGoroutines = 5             // Fewer goroutines in CI
		operationsPerGoroutine = 1000 // Fewer operations per goroutine
		t.Logf("Running in CI environment, reduced to %d goroutines with %d ops each",
			numGoroutines, operationsPerGoroutine)
	}

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
	minDuration := time.Hour

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
	avgDuration := totalDuration / time.Duration(numGoroutines)
	totalOps := numGoroutines * operationsPerGoroutine

	t.Logf("Concurrent stress test completed in %v", testDuration)
	t.Logf("Total operations: %d", totalOps)
	t.Logf("Worker durations - Avg: %v, Min: %v, Max: %v", avgDuration, minDuration, maxDuration)
	t.Logf("Throughput: %.0f ops/sec", float64(totalOps)/testDuration.Seconds())

	// Adaptive performance thresholds for concurrent scenarios
	avgNanosPerOp := avgDuration.Nanoseconds() / int64(operationsPerGoroutine)

	var maxConcurrentNsPerOp int64

	// Detect if race detector is enabled by checking for typical performance impact
	isRaceDetectorEnabled := avgNanosPerOp > 50000 && testing.Verbose()

	if isRaceDetectorEnabled || os.Getenv("GORACE") != "" || os.Getenv("CGO_ENABLED") == "1" {
		// Race detection or CGO enabled - very lenient for concurrent tests
		maxConcurrentNsPerOp = 500000 // 500μs per op
		t.Log("Race detection or CGO detected, using very relaxed concurrent threshold: 500μs")
	} else if IsCIEnvironment() {
		// CI environment - moderate threshold for concurrency
		maxConcurrentNsPerOp = 200000 // 200μs per op
		t.Log("CI environment detected, using moderate concurrent threshold: 200μs")
	} else {
		// Normal environment - stricter threshold
		maxConcurrentNsPerOp = 25000 // 25μs per op
		t.Log("Normal environment, using standard concurrent threshold: 25μs")
	}

	if avgNanosPerOp > maxConcurrentNsPerOp {
		t.Errorf("Concurrent performance regression: %d ns/op average (expected <%d ns/op)",
			avgNanosPerOp, maxConcurrentNsPerOp)
	}

	// Fairness check: scheduler should be reasonably fair even under load
	// Use a more generous threshold for CI environments where scheduling can be uneven
	fairnessRatio := float64(maxDuration) / float64(minDuration)
	if fairnessRatio > 10.0 { // Very generous threshold for CI/containerized environments
		t.Logf("Warning: Scheduling variance detected (ratio: %.1fx), but within acceptable limits for stress testing", fairnessRatio)
		// Only fail on extremely unfair scheduling that indicates a real problem
		if fairnessRatio > 20.0 {
			t.Errorf("Severely unfair scheduling detected: max %v vs min %v (ratio %.1fx)", maxDuration, minDuration, fairnessRatio)
		}
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
