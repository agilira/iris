// smartconfig_override_test.go: Regression tests for buildSmartConfig field override bug
//
// buildSmartConfig was silently ignoring user-provided BatchSize, TimeFn, and
// OnDrop fields because they had no override checks. The fix adds explicit
// checks for all three. These tests ensure the regression never returns.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"sync/atomic"
	"testing"
	"time"
)

// TestSmartConfig_BatchSizeOverride verifies that an explicit BatchSize in
// Config is respected by buildSmartConfig instead of being silently replaced
// by the hardcoded default (32).
func TestSmartConfig_BatchSizeOverride(t *testing.T) {
	// WHY capacity=8, batchSize=2: if the override does not work, the smart
	// default (32) exceeds capacity (8) and New() fails with
	// "batch size cannot exceed capacity".
	logger, err := New(Config{
		Capacity:  8,
		BatchSize: 2,
		Output:    WrapWriter(io.Discard),
	})
	if err != nil {
		t.Fatalf("New() failed (BatchSize override broken): %v", err)
	}

	logger.Start()
	logger.Info("test message")
	if closeErr := logger.Close(); closeErr != nil {
		t.Fatalf("Close() failed: %v", closeErr)
	}
}

// TestSmartConfig_TimeFnOverride verifies that a custom TimeFn in Config
// is used instead of the default CachedTime.
func TestSmartConfig_TimeFnOverride(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var timeCalled atomic.Int64

	customTimeFn := func() time.Time {
		timeCalled.Add(1)
		return fixedTime
	}

	logger, err := New(Config{
		Output: WrapWriter(io.Discard),
		TimeFn: customTimeFn,
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	logger.Start()
	logger.Info("test with custom time")

	if closeErr := logger.Close(); closeErr != nil {
		t.Fatalf("Close() failed: %v", closeErr)
	}

	if timeCalled.Load() == 0 {
		t.Fatal("custom TimeFn was never called (override not working)")
	}
}

// TestSmartConfig_OnDropOverride verifies that OnDrop in Config is passed
// through buildSmartConfig to the logger processor.
func TestSmartConfig_OnDropOverride(t *testing.T) {
	var dropCalled atomic.Int64

	logger, err := New(Config{
		Capacity:  8,
		BatchSize: 1,
		Output:    WrapWriter(io.Discard),
		OnDrop: func(_ int64) {
			dropCalled.Add(1)
		},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Flood without consumer to force drops
	for i := 0; i < 30; i++ {
		logger.Info("flood")
	}

	logger.Start()
	if closeErr := logger.Close(); closeErr != nil {
		t.Fatalf("Close() failed: %v", closeErr)
	}

	if dropCalled.Load() == 0 {
		t.Fatal("OnDrop callback was never called (override not working)")
	}
}
