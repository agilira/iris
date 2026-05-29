// zephyros_wakeup_test.go: Producer→consumer wakeup contract tests
//
// These tests pin the contract that a Write() performed AFTER the consumer
// has parked in IdleStrategy.Idle() still gets processed promptly. The
// regression they guard against: ChannelIdleStrategy(0) blocks the consumer
// indefinitely on its wakeup channel, and before this fix the producer never
// signalled it — so any record written once the ring drained was stranded
// forever (the consumer never woke to process it).
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"sync"
	"testing"
	"time"
)

// TestLoopProcess_ChannelIdle_WakesOnLateWrite is the core regression test.
//
// With a no-timeout ChannelIdleStrategy the consumer blocks in Idle() until
// explicitly woken. A producer that writes after the consumer has parked MUST
// wake it, otherwise the record is never processed. We deliberately let the
// consumer drain to empty (so it enters the blocking Idle) BEFORE writing.
func TestLoopProcess_ChannelIdle_WakesOnLateWrite(t *testing.T) {
	var mu sync.Mutex
	processed := 0
	processor := func(_ *TestRecord) {
		mu.Lock()
		processed++
		mu.Unlock()
	}

	z, err := NewBuilder[TestRecord](1024).
		WithProcessor(processor).
		WithIdleStrategy(NewChannelIdleStrategy(0)). // 0 = block until woken
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	go z.LoopProcess()
	defer z.Close()

	// Let the consumer reach the blocking Idle() on an empty ring.
	time.Sleep(20 * time.Millisecond)

	// Late write: must wake the parked consumer.
	if !z.Write(func(r *TestRecord) { r.Message = "late" }) {
		t.Fatal("Write returned false on a non-full ring")
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := processed
		mu.Unlock()
		if n == 1 {
			return // consumer woke and processed the late write
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("late write was never processed: consumer stayed parked in Idle() " +
		"(producer did not wake the channel idle strategy)")
}

// TestLoopProcess_ChannelIdle_WakesOnEveryBurst pins that the wakeup is not a
// one-shot: after the consumer re-parks between bursts, each subsequent burst
// must wake it again.
func TestLoopProcess_ChannelIdle_WakesOnEveryBurst(t *testing.T) {
	var mu sync.Mutex
	processed := 0
	processor := func(_ *TestRecord) {
		mu.Lock()
		processed++
		mu.Unlock()
	}

	z, err := NewBuilder[TestRecord](1024).
		WithProcessor(processor).
		WithIdleStrategy(NewChannelIdleStrategy(0)).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	go z.LoopProcess()
	defer z.Close()

	const bursts = 5
	for b := 0; b < bursts; b++ {
		time.Sleep(20 * time.Millisecond) // let the consumer re-park
		if !z.Write(func(r *TestRecord) { r.Message = "burst" }) {
			t.Fatalf("burst %d: Write returned false on a non-full ring", b)
		}
	}

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := processed
		mu.Unlock()
		if n == bursts {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	mu.Lock()
	got := processed
	mu.Unlock()
	t.Fatalf("expected %d records across bursts, got %d: consumer missed a wakeup", bursts, got)
}
