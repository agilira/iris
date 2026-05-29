// zephyros_closeleak_test.go: Close() must release a channel-parked consumer
//
// Pins that LoopProcess returns after Close() even when the consumer is parked
// in a no-timeout ChannelIdleStrategy. Without a wake on Close the consumer
// blocks on <-wakeupChan forever and the goroutine leaks.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import (
	"testing"
	"time"
)

func TestLoopProcess_ChannelIdle_CloseReleasesParkedConsumer(t *testing.T) {
	z, err := NewBuilder[TestRecord](1024).
		WithProcessor(func(_ *TestRecord) {}).
		WithIdleStrategy(NewChannelIdleStrategy(0)). // 0 = block until woken
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		z.LoopProcess()
		close(done)
	}()

	// Let the consumer drain the empty ring and park in Idle().
	time.Sleep(20 * time.Millisecond)

	z.Close()

	select {
	case <-done:
		// LoopProcess returned: Close released the parked consumer.
	case <-time.After(500 * time.Millisecond):
		t.Fatal("LoopProcess did not return after Close(): consumer leaked, " +
			"parked forever on the channel idle strategy")
	}
}
