// bugfix_test.go: Regression tests for 3 bugs found during code review
//
// 1. Hook panic kills consumer goroutine (no recover)
// 2. maxFields truncation is silent (data loss with no diagnostic)
// 3. Ring buffer sequence waste (tested at zephyroslite level)
//
// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync/atomic"
	"testing"
	"time"

	goerrors "github.com/agilira/go-errors"
)

// --- Hook Panic Recovery (Bug #1) ---

// TestHookPanic_ConsumerSurvives verifies that a panicking hook does not
// kill the consumer goroutine. After the panic, subsequent log entries
// must still be processed normally.
func TestHookPanic_ConsumerSurvives(t *testing.T) {
	buf := &bufferedSyncer{}
	var postPanicSeen atomic.Bool

	panicHook := func(rec *Record) {
		panic("intentional hook panic for testing")
	}
	witnessHook := func(rec *Record) {
		if rec.Msg == "after-panic" {
			postPanicSeen.Store(true)
		}
	}

	logger, err := New(Config{
		Level:    Info,
		Encoder:  NewJSONEncoder(),
		Output:   buf,
		Capacity: 1024,
	}, WithHook(panicHook), WithHook(witnessHook))
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// This triggers the panicking hook
	logger.Info("trigger-panic")
	time.Sleep(50 * time.Millisecond)

	// This MUST still be processed (consumer survived)
	logger.Info("after-panic")
	time.Sleep(50 * time.Millisecond)

	if !postPanicSeen.Load() {
		t.Error("consumer goroutine died after hook panic: post-panic log was never processed")
	}
}

// TestHookPanic_ErrorHandlerCalled verifies that a hook panic is
// reported via the ErrorHandler, not silently swallowed.
func TestHookPanic_ErrorHandlerCalled(t *testing.T) {
	buf := &bufferedSyncer{}
	var errorReported atomic.Bool

	panicHook := func(rec *Record) {
		panic("hook panic test")
	}

	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
		ErrorHandler: func(e *goerrors.Error) {
			if goerrors.HasCode(e, ErrCodeHookExecution) {
				errorReported.Store(true)
			}
		},
		Capacity: 1024,
	}, WithHook(panicHook))
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()
	logger.Info("trigger")
	time.Sleep(50 * time.Millisecond)

	if !errorReported.Load() {
		t.Error("ErrorHandler was not called with ErrCodeHookExecution after hook panic")
	}
}

// TestHookPanic_OtherHooksStillRun verifies that a panic in one hook
// does not prevent subsequent hooks from executing on the NEXT record.
func TestHookPanic_OtherHooksStillRun(t *testing.T) {
	buf := &bufferedSyncer{}
	var hookBCount atomic.Int64

	hookA := func(rec *Record) {
		if rec.Msg == "boom" {
			panic("hookA panic")
		}
	}
	hookB := func(rec *Record) {
		hookBCount.Add(1)
	}

	logger, err := New(Config{
		Level:    Info,
		Encoder:  NewJSONEncoder(),
		Output:   buf,
		Capacity: 1024,
	}, WithHook(hookA), WithHook(hookB))
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	logger.Info("boom") // hookA panics, hookB should still run for this record
	logger.Info("safe") // both hooks run normally
	time.Sleep(100 * time.Millisecond)

	if hookBCount.Load() < 2 {
		t.Errorf("hookB should have been called at least 2 times, got %d", hookBCount.Load())
	}
}

// --- maxFields Truncation Warning (Bug #3) ---

// TestMaxFieldsTruncation_CounterIncremented verifies that when a log
// record has more fields than maxFields, the truncated counter increments.
func TestMaxFieldsTruncation_CounterIncremented(t *testing.T) {
	buf := &bufferedSyncer{}

	logger, err := New(Config{
		Level:    Info,
		Encoder:  NewJSONEncoder(),
		Output:   buf,
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Build fields array larger than maxFields (32)
	fields := make([]Field, 40)
	for i := range fields {
		fields[i] = Int("field", i)
	}

	logger.Info("too-many-fields", fields...)
	time.Sleep(50 * time.Millisecond)

	stats := logger.Stats()
	if stats["truncated_fields"] != 1 {
		t.Errorf("expected truncated_fields=1, got %d", stats["truncated_fields"])
	}
}

// TestMaxFieldsTruncation_NormalNotCounted verifies that records with
// fewer than maxFields do NOT increment the truncation counter.
func TestMaxFieldsTruncation_NormalNotCounted(t *testing.T) {
	buf := &bufferedSyncer{}

	logger, err := New(Config{
		Level:    Info,
		Encoder:  NewJSONEncoder(),
		Output:   buf,
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	logger.Info("normal", Str("a", "1"), Int("b", 2))
	logger.Info("also-normal", Bool("c", true))
	time.Sleep(50 * time.Millisecond)

	stats := logger.Stats()
	if stats["truncated_fields"] != 0 {
		t.Errorf("expected truncated_fields=0 for normal records, got %d", stats["truncated_fields"])
	}
}

// TestMaxFieldsTruncation_ExactlyMaxFields verifies that exactly 32 fields
// does NOT trigger truncation (off-by-one check).
func TestMaxFieldsTruncation_ExactlyMaxFields(t *testing.T) {
	buf := &bufferedSyncer{}

	logger, err := New(Config{
		Level:    Info,
		Encoder:  NewJSONEncoder(),
		Output:   buf,
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Exactly maxFields (32) user fields, no base fields
	fields := make([]Field, 32)
	for i := range fields {
		fields[i] = Int("f", i)
	}

	logger.Info("exact-limit", fields...)
	time.Sleep(50 * time.Millisecond)

	stats := logger.Stats()
	if stats["truncated_fields"] != 0 {
		t.Errorf("exactly 32 fields should not truncate, got truncated_fields=%d", stats["truncated_fields"])
	}
}

// TestMaxFieldsTruncation_DataStillWritten verifies that even when fields
// are truncated, the first 32 are still written and the log entry is NOT
// dropped (partial data is better than no data).
func TestMaxFieldsTruncation_DataStillWritten(t *testing.T) {
	buf := &bufferedSyncer{}

	logger, err := New(Config{
		Level:    Info,
		Encoder:  NewJSONEncoder(),
		Output:   buf,
		Capacity: 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	fields := make([]Field, 40)
	for i := range fields {
		fields[i] = Int("field", i)
	}

	ok := logger.Info("truncated-but-written", fields...)
	time.Sleep(50 * time.Millisecond)

	if !ok {
		t.Error("log entry with too many fields should still be written (not dropped)")
	}

	output := buf.String()
	if output == "" {
		t.Error("expected output to contain the log entry")
	}
}
