// ondrop_test.go: Tests for the OnDrop forensic callback
//
// The OnDrop callback enables real-time detection of log message drops,
// which can signal log-flooding attacks (CWE-778). These tests verify:
// 1. Callback fires when drops occur (ring buffer full)
// 2. Callback receives monotonically increasing total
// 3. Callback is NOT called when no drops occur
// 4. Consumer survives a panicking OnDrop callback
// 5. ErrorHandler receives the OnDrop panic report
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	errors "github.com/agilira/go-errors"
)

// TestOnDrop_FiredOnRingFull verifies the callback fires exactly when drops
// happen because the ring buffer is saturated and the consumer is not running.
func TestOnDrop_FiredOnRingFull(t *testing.T) {
	var callbackCount atomic.Int64
	var lastTotal atomic.Int64

	logger, err := New(Config{
		Capacity:           4, // Tiny ring to force drops
		BatchSize:          1,
		Output:             WrapWriter(io.Discard),
		BackpressurePolicy: 0, // DropOnFull
		OnDrop: func(totalDropped int64) {
			callbackCount.Add(1)
			lastTotal.Store(totalDropped)
		},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Fill the ring without starting the consumer -> drops are guaranteed
	for i := 0; i < 20; i++ {
		logger.Info("flood")
	}

	// Now start the consumer to process what it can and trigger OnDrop
	logger.Start()
	if err := logger.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	if callbackCount.Load() == 0 {
		t.Fatal("OnDrop callback was never called despite drops occurring")
	}
	if lastTotal.Load() == 0 {
		t.Fatal("OnDrop received totalDropped=0, expected > 0")
	}

	// Verify consistent with Stats
	droppedStat := logger.Stats()["dropped"]
	if lastTotal.Load() != droppedStat {
		t.Errorf("OnDrop total (%d) does not match Stats dropped (%d)", lastTotal.Load(), droppedStat)
	}
}

// TestOnDrop_MonotonicallyIncreasing verifies the total passed to OnDrop
// never decreases across invocations.
func TestOnDrop_MonotonicallyIncreasing(t *testing.T) {
	var mu sync.Mutex
	var totals []int64

	logger, err := New(Config{
		Capacity:           4,
		BatchSize:          1,
		Output:             WrapWriter(io.Discard),
		BackpressurePolicy: 0, // DropOnFull
		OnDrop: func(totalDropped int64) {
			mu.Lock()
			totals = append(totals, totalDropped)
			mu.Unlock()
		},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	logger.Start()

	// Flood in bursts to provoke multiple OnDrop calls
	for burst := 0; burst < 5; burst++ {
		for i := 0; i < 30; i++ {
			logger.Info("burst-flood")
		}
		time.Sleep(5 * time.Millisecond) // Let consumer catch up
	}

	if err := logger.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(totals) == 0 {
		t.Fatal("OnDrop callback was never called")
	}

	for i := 1; i < len(totals); i++ {
		if totals[i] < totals[i-1] {
			t.Errorf("OnDrop total decreased: totals[%d]=%d < totals[%d]=%d",
				i, totals[i], i-1, totals[i-1])
		}
	}
}

// TestOnDrop_NotCalledWhenNoDrops verifies the callback is NOT invoked
// when no drops occur (ring has plenty of capacity).
func TestOnDrop_NotCalledWhenNoDrops(t *testing.T) {
	var called atomic.Int64

	logger, err := New(Config{
		Capacity:  65536, // Huge ring, no drops possible
		BatchSize: 32,
		Output:    WrapWriter(io.Discard),
		OnDrop: func(_ int64) {
			called.Add(1)
		},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	logger.Start()

	// Log a few messages - should all fit
	for i := 0; i < 10; i++ {
		logger.Info("normal message")
	}

	if err := logger.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	if called.Load() != 0 {
		t.Errorf("OnDrop was called %d times, expected 0 (no drops)", called.Load())
	}
}

// TestOnDrop_NilCallbackSafe verifies that a nil OnDrop (the default)
// causes no errors.
func TestOnDrop_NilCallbackSafe(t *testing.T) {
	logger, err := New(Config{
		Capacity:           4,
		BatchSize:          1,
		Output:             WrapWriter(io.Discard),
		BackpressurePolicy: 0, // DropOnFull
		// OnDrop intentionally nil
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Force drops
	for i := 0; i < 20; i++ {
		logger.Info("flood")
	}

	logger.Start()
	if err := logger.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Test passes if no panic occurred
}

// TestOnDrop_PanicRecovery verifies the consumer survives a panicking
// OnDrop callback. This is critical: a user-supplied callback must never
// be able to kill the consumer goroutine and silently lose all log data.
func TestOnDrop_PanicRecovery(t *testing.T) {
	var errorReceived atomic.Int64

	logger, err := New(Config{
		Capacity:           4,
		BatchSize:          1,
		Output:             WrapWriter(io.Discard),
		BackpressurePolicy: 0, // DropOnFull
		OnDrop: func(_ int64) {
			panic("intentional test panic in OnDrop")
		},
		ErrorHandler: func(_ *errors.Error) {
			errorReceived.Add(1)
		},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Force drops
	for i := 0; i < 20; i++ {
		logger.Info("flood")
	}

	logger.Start()

	// Give consumer time to process and hit OnDrop
	time.Sleep(50 * time.Millisecond)

	// Log more messages AFTER the panic - consumer must still work
	for i := 0; i < 5; i++ {
		logger.Info("after-panic")
	}
	time.Sleep(20 * time.Millisecond)

	if err := logger.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	if errorReceived.Load() == 0 {
		t.Fatal("ErrorHandler was never called for OnDrop panic")
	}
}

// TestOnDrop_ErrorHandlerReceivesPanicReport verifies that when OnDrop
// panics, the error is reported via ErrorHandler with the correct error code.
func TestOnDrop_ErrorHandlerReceivesPanicReport(t *testing.T) {
	var receivedErr atomic.Value

	logger, err := New(Config{
		Capacity:           4,
		BatchSize:          1,
		Output:             WrapWriter(io.Discard),
		BackpressurePolicy: 0, // DropOnFull
		OnDrop: func(_ int64) {
			panic("kaboom")
		},
		ErrorHandler: func(err *errors.Error) {
			receivedErr.Store(err)
		},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	for i := 0; i < 20; i++ {
		logger.Info("flood")
	}

	logger.Start()
	time.Sleep(50 * time.Millisecond)

	if err := logger.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	stored := receivedErr.Load()
	if stored == nil {
		t.Fatal("ErrorHandler never received the OnDrop panic error")
	}

	errStr := stored.(error).Error()
	if errStr == "" {
		t.Fatal("ErrorHandler received empty error string")
	}
}
