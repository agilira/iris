// security_test.go: Security test suite for ZephyrosLight MPSC ring buffer
//
// THREAT MODEL:
//
// ZephyrosLight is a lock-free MPSC ring buffer handling untrusted workloads
// from arbitrary goroutine counts. The attack surface includes:
//
//   CWE-362: Race condition on Close() + Write() (TOCTOU between closed check
//            and sequence claim). Writes can succeed after Close() due to the
//            gap between the closed flag check and the atomic sequence increment.
//
//   CWE-476: Nil writerFunc dereference. Write() accepts a func(*T) with no
//            nil guard. If caller passes nil, processor panics on invocation.
//
//   CWE-400: Sequence number exhaustion under BlockOnFull contention. Each
//            retry burns a sequence number without writing, creating phantom
//            claimed slots that cause consumer stalls.
//
//   CWE-834: Excessive iteration / infinite loop. ChannelIdleStrategy with
//            zero timeout blocks forever if no producer ever calls Reset().
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// securityRecord is a minimal payload for security tests.
type securityRecord struct {
	Value int64
}

// ---------------------------------------------------------------------------
// CWE-362: Race condition on Close() + concurrent Write()
// ATTACK VECTOR: Multiple goroutines writing while another goroutine closes.
// IMPACT: Write after close could corrupt ring state or panic.
// MITIGATION EXPECTED: Write returns false after Close, no panic, no data race.
// ---------------------------------------------------------------------------

func TestSecurity_WriteAfterClose_Race(t *testing.T) {
	t.Parallel()

	var processed atomic.Int64
	processor := func(r *securityRecord) {
		processed.Add(1)
	}

	ring, err := NewBuilder[securityRecord](64).
		WithProcessor(processor).
		WithBatchSize(8).
		WithBackpressurePolicy(DropOnFull).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Start consumer
	done := make(chan struct{})
	go func() {
		ring.LoopProcess()
		close(done)
	}()

	// Hammer writes from many goroutines while closing concurrently.
	// WHY 50 goroutines: enough to saturate scheduler and expose TOCTOU window.
	const writers = 50
	const writesPerGoroutine = 1000
	var wg sync.WaitGroup
	wg.Add(writers + 1)

	// Writers
	for i := 0; i < writers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				ring.Write(func(r *securityRecord) {
					r.Value = int64(id*writesPerGoroutine + j)
				})
				// Ignore return: some will be false after close
			}
		}(i)
	}

	// Closer: wait a bit then close mid-flight
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		ring.Close()
	}()

	wg.Wait()
	<-done

	// WHY no assertion on exact count: the race is intentional.
	// We only assert no panic and no data race (detected by -race flag).
	t.Logf("Processed %d records under close-race pressure", processed.Load())
}

// ---------------------------------------------------------------------------
// CWE-362: Double-close must be idempotent
// ATTACK VECTOR: Multiple goroutines competing to close the ring.
// IMPACT: Double close on channel panics if not guarded.
// MITIGATION EXPECTED: Close() is idempotent via atomic CAS.
// ---------------------------------------------------------------------------

func TestSecurity_DoubleClose_Idempotent(t *testing.T) {
	t.Parallel()

	processor := func(r *securityRecord) {}

	ring, err := NewBuilder[securityRecord](16).
		WithProcessor(processor).
		WithBatchSize(4).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Close from multiple goroutines concurrently.
	const closers = 20
	var wg sync.WaitGroup
	wg.Add(closers)
	for i := 0; i < closers; i++ {
		go func() {
			defer wg.Done()
			ring.Close() // Must not panic
		}()
	}
	wg.Wait()

	// Subsequent writes must return false, never panic.
	ok := ring.Write(func(r *securityRecord) {
		r.Value = 42
	})
	if ok {
		t.Error("Write after close should return false")
	}
}

// ---------------------------------------------------------------------------
// CWE-400: Full buffer with DropOnFull -- no resource leak
// ATTACK VECTOR: Flood producer fills buffer, dropped counter must be accurate.
// IMPACT: If drops are not counted, monitoring is blind to data loss.
// MITIGATION EXPECTED: dropped counter increments for every rejected write.
// ---------------------------------------------------------------------------

func TestSecurity_BufferFull_DropCounting(t *testing.T) {
	t.Parallel()

	const capacity = 16
	processor := func(r *securityRecord) {
		// Slow consumer: never drain
	}

	ring, err := NewBuilder[securityRecord](capacity).
		WithProcessor(processor).
		WithBatchSize(4).
		WithBackpressurePolicy(DropOnFull).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	defer ring.Close()

	// Fill buffer beyond capacity without consumer running.
	var accepted, rejected int
	for i := 0; i < capacity*4; i++ {
		ok := ring.Write(func(r *securityRecord) {
			r.Value = int64(i)
		})
		if ok {
			accepted++
		} else {
			rejected++
		}
	}

	if rejected == 0 {
		t.Error("Expected some writes to be rejected when buffer is full")
	}
	if accepted == 0 {
		t.Error("Expected some writes to succeed before buffer is full")
	}

	stats := ring.Stats()
	droppedCount := stats["items_dropped"]
	if droppedCount != int64(rejected) {
		t.Errorf("Dropped counter mismatch: stats=%d, counted=%d", droppedCount, rejected)
	}

	t.Logf("accepted=%d, rejected=%d, dropped_stat=%d", accepted, rejected, droppedCount)
}

// ---------------------------------------------------------------------------
// CWE-400: Concurrent producer flood with DropOnFull
// ATTACK VECTOR: Many goroutines saturating the ring simultaneously.
// IMPACT: Atomic operations under contention must not corrupt state.
// MITIGATION EXPECTED: No data race, no panic, total=accepted+dropped.
// ---------------------------------------------------------------------------

func TestSecurity_ConcurrentFlood_DropOnFull(t *testing.T) {
	t.Parallel()

	const capacity = 64
	var processed atomic.Int64
	processor := func(r *securityRecord) {
		processed.Add(1)
	}

	ring, err := NewBuilder[securityRecord](capacity).
		WithProcessor(processor).
		WithBatchSize(8).
		WithBackpressurePolicy(DropOnFull).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Start consumer
	done := make(chan struct{})
	go func() {
		ring.LoopProcess()
		close(done)
	}()

	// Flood from 100 goroutines
	const writers = 100
	const perWriter = 500
	var totalAccepted atomic.Int64
	var wg sync.WaitGroup
	wg.Add(writers)

	for i := 0; i < writers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perWriter; j++ {
				if ring.Write(func(r *securityRecord) {
					r.Value = 1
				}) {
					totalAccepted.Add(1)
				}
			}
		}()
	}

	wg.Wait()
	ring.Close()
	<-done

	stats := ring.Stats()
	total := totalAccepted.Load() + stats["items_dropped"]
	expected := int64(writers * perWriter)

	if total != expected {
		t.Errorf("Accounting mismatch: accepted(%d) + dropped(%d) = %d, expected %d",
			totalAccepted.Load(), stats["items_dropped"], total, expected)
	}

	t.Logf("processed=%d, accepted=%d, dropped=%d",
		processed.Load(), totalAccepted.Load(), stats["items_dropped"])
}

// ---------------------------------------------------------------------------
// CWE-20: Builder rejects invalid configurations
// ATTACK VECTOR: Malicious or buggy caller passes edge-case parameters.
// IMPACT: Invalid ring could panic or corrupt memory.
// MITIGATION EXPECTED: Build() returns descriptive error, ring is nil.
// ---------------------------------------------------------------------------

func TestSecurity_Builder_InvalidInputs(t *testing.T) {
	t.Parallel()

	processor := func(r *securityRecord) {}

	tests := []struct {
		name     string
		capacity int64
		batch    int64
		proc     ProcessorFunc[securityRecord]
		wantErr  error
	}{
		{"zero capacity", 0, 1, processor, ErrInvalidCapacity},
		{"negative capacity", -1, 1, processor, ErrInvalidCapacity},
		{"non-power-of-two capacity", 13, 1, processor, ErrInvalidCapacity},
		{"zero batch size", 16, 0, processor, ErrInvalidBatchSize},
		{"negative batch size", 16, -1, processor, ErrInvalidBatchSize},
		{"batch exceeds capacity", 16, 32, processor, ErrInvalidBatchSize},
		{"nil processor", 16, 4, nil, ErrMissingProcessor},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewBuilder[securityRecord](tc.capacity).
				WithBatchSize(tc.batch).
				WithBackpressurePolicy(DropOnFull)
			if tc.proc != nil {
				builder = builder.WithProcessor(tc.proc)
			}

			ring, err := builder.Build()
			if err == nil {
				if ring != nil {
					ring.Close()
				}
				t.Fatalf("Expected error %v, got nil", tc.wantErr)
			}
			if err != tc.wantErr {
				t.Errorf("Expected error %v, got %v", tc.wantErr, err)
			}
			if ring != nil {
				ring.Close()
				t.Error("Ring should be nil on error")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CWE-362: Flush under concurrent write pressure
// ATTACK VECTOR: Flush called while producers are still writing.
// IMPACT: Flush could hang forever or miss records.
// MITIGATION EXPECTED: Flush completes within timeout, consumer drains buffer.
// ---------------------------------------------------------------------------

func TestSecurity_FlushUnderPressure(t *testing.T) {
	t.Parallel()

	var processed atomic.Int64
	processor := func(r *securityRecord) {
		processed.Add(1)
	}

	ring, err := NewBuilder[securityRecord](128).
		WithProcessor(processor).
		WithBatchSize(16).
		WithBackpressurePolicy(DropOnFull).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Start consumer
	done := make(chan struct{})
	go func() {
		ring.LoopProcess()
		close(done)
	}()

	// Write a burst then flush
	const burst = 100
	for i := 0; i < burst; i++ {
		ring.Write(func(r *securityRecord) {
			r.Value = int64(i)
		})
	}

	// Flush must complete (not hang)
	flushDone := make(chan error, 1)
	go func() {
		flushDone <- ring.Flush()
	}()

	select {
	case err := <-flushDone:
		if err != nil {
			t.Errorf("Flush returned error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Flush hung under write pressure")
	}

	ring.Close()
	<-done

	if processed.Load() < int64(burst) {
		t.Errorf("Expected at least %d processed, got %d", burst, processed.Load())
	}
}

// ---------------------------------------------------------------------------
// CWE-834: Progressive idle strategy must not spin forever
// ATTACK VECTOR: Consumer with no producers spins indefinitely.
// IMPACT: CPU exhaustion, goroutine starvation.
// MITIGATION EXPECTED: Idle strategy yields CPU after progressive backoff.
// ---------------------------------------------------------------------------

func TestSecurity_IdleStrategy_ProgressiveBackoff(t *testing.T) {
	t.Parallel()

	strategy := NewProgressiveIdleStrategy()

	// Simulate 1000 idle cycles -- must not block or panic.
	for i := 0; i < 1000; i++ {
		strategy.Idle()
	}

	// Reset must not panic after heavy idle.
	strategy.Reset()

	// After reset, counters restart.
	for i := 0; i < 10; i++ {
		strategy.Idle()
	}
}

// ---------------------------------------------------------------------------
// CWE-362: Stats consistency under concurrent mutations
// ATTACK VECTOR: Reading stats while producers and consumer are active.
// IMPACT: Stats could return torn reads or inconsistent values.
// MITIGATION EXPECTED: Stats uses atomic loads, no torn reads.
// ---------------------------------------------------------------------------

func TestSecurity_Stats_ConcurrentRead(t *testing.T) {
	t.Parallel()

	var processed atomic.Int64
	processor := func(r *securityRecord) {
		processed.Add(1)
	}

	ring, err := NewBuilder[securityRecord](64).
		WithProcessor(processor).
		WithBatchSize(8).
		WithBackpressurePolicy(DropOnFull).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Start consumer
	done := make(chan struct{})
	go func() {
		ring.LoopProcess()
		close(done)
	}()

	// Write and read stats concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	// Writer goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			ring.Write(func(r *securityRecord) {
				r.Value = int64(i)
			})
		}
	}()

	// Stats reader goroutine -- must never panic or return negative values
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			stats := ring.Stats()
			if stats["items_processed"] < 0 {
				t.Errorf("Negative processed count: %d", stats["items_processed"])
			}
			if stats["items_dropped"] < 0 {
				t.Errorf("Negative dropped count: %d", stats["items_dropped"])
			}
		}
	}()

	wg.Wait()
	ring.Close()
	<-done
}
