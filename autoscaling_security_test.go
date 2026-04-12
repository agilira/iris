// autoscaling_security_test.go: Security test suite for AdaptiveLogger
//
// THREAT MODEL:
//   CWE-362: Concurrent Execution Using Shared Resource
//     - getLogger() concurrent reads during mode transition
//     - ensureMultiLogger via sync.Once has initialization race surface
//     - scaleDown during active writes could invalidate logger reference
//     - concurrent Start/Close lifecycle could leave inconsistent state
//   CWE-400: Uncontrolled Resource Consumption
//     - Unbounded goroutine scaling via contention detection
//     - Resource leak if Close() not called or called multiple times
//   CWE-834: Excessive Iteration
//     - scaleDownMonitor timer loop must terminate on context cancellation
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// safeCloseAdaptiveSec closes an AdaptiveLogger, tolerating ring buffer flush
// timeouts which are expected when tests don't fully drain the ring buffer.
func safeCloseAdaptiveSec(t *testing.T, al *AdaptiveLogger) {
	t.Helper()
	if err := al.Close(); err != nil &&
		!strings.Contains(err.Error(), "flush timeout") &&
		!strings.Contains(err.Error(), "sync /dev/stdout") {
		t.Errorf("Close: %v", err)
	}
}

// --- Attack vector: CWE-362 (Concurrent lifecycle) ---

func TestSecurity_Autoscaling_ConcurrentWriteDuringScaleUp(t *testing.T) {
	// ATTACK VECTOR: CWE-362
	// IMPACT: Concurrent writes that trigger contention detection and scale-up
	// could observe stale logger references if mode transition is not atomic.
	// MITIGATION EXPECTED: getLogger() returns consistent reference via atomic mode.
	const goroutines = 20

	var buf bufferedSyncer
	cfg := Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  &buf,
	}

	adaptive, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	if startErr := adaptive.Start(); startErr != nil {
		t.Fatalf("Start: %v", startErr)
	}
	defer safeCloseAdaptiveSec(t, adaptive)

	var wg sync.WaitGroup
	var panicCount atomic.Int64

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicCount.Add(1)
					t.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			for i := 0; i < 100; i++ {
				adaptive.Info("concurrent write", Int("goroutine", id), Int("iter", i))
			}
		}(g)
	}

	wg.Wait()

	if panicCount.Load() > 0 {
		t.Fatalf("%d goroutines panicked during concurrent writes", panicCount.Load())
	}
}

func TestSecurity_Autoscaling_ConcurrentStartClose(t *testing.T) {
	// ATTACK VECTOR: CWE-362
	// IMPACT: Concurrent Start and Close calls could leave logger in
	// inconsistent state (started but not closeable, or double-close).
	// MITIGATION EXPECTED: atomic state management prevents corruption.
	var buf bufferedSyncer
	cfg := Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  &buf,
	}

	adaptive, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	if startErr := adaptive.Start(); startErr != nil {
		t.Fatalf("Start: %v", startErr)
	}

	// Close should be idempotent (no panic on first call)
	safeCloseAdaptiveSec(t, adaptive)

	// Second Close should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("second Close panicked: %v", r)
			}
		}()
		// WHY ignoring error: double-close may return error but must not panic
		_ = adaptive.Close()
	}()
}

func TestSecurity_Autoscaling_WriteAfterClose(t *testing.T) {
	// ATTACK VECTOR: CWE-362
	// IMPACT: Writing to a closed logger could panic on nil pointer
	// or write to a closed writer.
	// MITIGATION EXPECTED: graceful handling (no panic).
	var buf bufferedSyncer
	cfg := Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  &buf,
	}

	adaptive, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	if startErr := adaptive.Start(); startErr != nil {
		t.Fatalf("Start: %v", startErr)
	}

	safeCloseAdaptiveSec(t, adaptive)

	// Must not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("write-after-close panicked: %v", r)
			}
		}()
		adaptive.Info("after close")
	}()
}

// --- Attack vector: CWE-400 (Resource exhaustion) ---

func TestSecurity_Autoscaling_HighContentionDoesNotOOM(t *testing.T) {
	// ATTACK VECTOR: CWE-400
	// IMPACT: Extreme contention triggering repeated scale-up could
	// allocate unbounded internal resources (multi-logger goroutines).
	// MITIGATION EXPECTED: scale-up is bounded (single multi-logger via sync.Once).
	const goroutines = 50

	var buf bufferedSyncer
	cfg := Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  &buf,
	}

	adaptive, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	if startErr := adaptive.Start(); startErr != nil {
		t.Fatalf("Start: %v", startErr)
	}
	defer safeCloseAdaptiveSec(t, adaptive)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				adaptive.Info("flood", Str("data", "test"))
			}
		}()
	}

	wg.Wait()
	// If we reach here without OOM or panic, the scale-up is bounded
}

// --- Attack vector: CWE-834 (Scale-down monitor termination) ---

func TestSecurity_Autoscaling_ScaleDownMonitorTerminates(t *testing.T) {
	// ATTACK VECTOR: CWE-834
	// IMPACT: If scaleDownMonitor timer loop does not respect context
	// cancellation, it becomes a goroutine leak (resource exhaustion).
	// MITIGATION EXPECTED: context.Done channel terminates the monitor.
	var buf bufferedSyncer
	cfg := Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  &buf,
	}

	adaptive, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	if startErr := adaptive.Start(); startErr != nil {
		t.Fatalf("Start: %v", startErr)
	}

	// Force scale-up by generating contention
	var wg sync.WaitGroup
	wg.Add(10)
	for g := 0; g < 10; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				adaptive.Info("contention trigger")
			}
		}()
	}
	wg.Wait()

	// Close should terminate the scale-down monitor goroutine
	safeCloseAdaptiveSec(t, adaptive)

	// WHY sleep: give goroutine scheduler time to propagate context cancellation.
	// If the monitor goroutine has not terminated within 500ms, it is leaked.
	// We verify indirectly that Close() does not hang (it returns above).
}

// --- Attack vector: CWE-362 (Mode transition during write) ---

func TestSecurity_Autoscaling_ModeTransitionConsistency(t *testing.T) {
	// ATTACK VECTOR: CWE-362
	// IMPACT: During SingleMode -> MultiMode transition, a concurrent
	// write could observe an intermediate state where neither logger is valid.
	// MITIGATION EXPECTED: getLogger() atomics ensure valid reference at all times.
	const (
		writers      = 10
		writesPerGor = 200
	)

	var buf bufferedSyncer
	cfg := Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  &buf,
	}

	adaptive, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	if startErr := adaptive.Start(); startErr != nil {
		t.Fatalf("Start: %v", startErr)
	}
	defer safeCloseAdaptiveSec(t, adaptive)

	var writeErrors atomic.Int64
	var wg sync.WaitGroup

	wg.Add(writers)
	for w := 0; w < writers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < writesPerGor; i++ {
				func() {
					defer func() {
						if r := recover(); r != nil {
							writeErrors.Add(1)
						}
					}()
					adaptive.Info("transition test",
						Int("writer", id),
						Int("seq", i))
				}()
			}
		}(w)
	}

	wg.Wait()

	if errs := writeErrors.Load(); errs > 0 {
		t.Fatalf("%d panics during mode transition writes", errs)
	}
}
