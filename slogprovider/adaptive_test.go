// adaptive_test.go: Unit tests for AdaptiveHandler (slog -> AdaptiveLogger bridge)
//
// WHY separate from provider_test.go: AdaptiveHandler is a fundamentally
// different integration path -- synchronous direct writes into AdaptiveLogger
// instead of async channel-based SyncReader. Testing them together would
// obscure which path a failure belongs to.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package slogprovider

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/agilira/iris"
)

// newTestAdaptiveLogger creates an AdaptiveLogger wired to a bufferedWriter.
// Returns the logger (started) and the buffer for output inspection.
// WHY helper: every test needs this setup; duplicating it is a maintenance trap.
func newTestAdaptiveLogger(t *testing.T) (*iris.AdaptiveLogger, *bufferedWriter) {
	t.Helper()
	buf := &bufferedWriter{}

	al, err := iris.NewAdaptiveLogger(iris.DefaultScalerConfig(iris.Config{
		Output:  buf,
		Encoder: iris.NewJSONEncoder(),
		Level:   iris.Debug,
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger failed: %v", err)
	}
	if err := al.Start(); err != nil {
		t.Fatalf("AdaptiveLogger.Start() failed: %v", err)
	}
	return al, buf
}

func TestNewAdaptiveHandler(t *testing.T) {
	al, buf := newTestAdaptiveLogger(t)
	_ = buf
	defer func() {
		if err := al.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	h := NewAdaptiveHandler(al)
	if h == nil {
		t.Fatal("NewAdaptiveHandler returned nil")
	}
}

func TestAdaptiveHandler_Handle_BasicMessage(t *testing.T) {
	al, buf := newTestAdaptiveLogger(t)
	defer func() {
		if err := al.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	h := NewAdaptiveHandler(al)
	logger := slog.New(h)
	logger.Info("hello from slog", "user", "antonio")

	// Flush
	time.Sleep(50 * time.Millisecond)
	if err := al.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "hello from slog") {
		t.Errorf("expected message in output, got: %s", output)
	}
	if !strings.Contains(output, `"user":"antonio"`) {
		t.Errorf("expected user field in output, got: %s", output)
	}
}

func TestAdaptiveHandler_Handle_AllLevels(t *testing.T) {
	al, buf := newTestAdaptiveLogger(t)
	defer func() {
		if err := al.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	h := NewAdaptiveHandler(al)
	logger := slog.New(h)

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	time.Sleep(50 * time.Millisecond)
	if err := al.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}

	output := buf.String()
	for _, expected := range []string{"debug msg", "info msg", "warn msg", "error msg"} {
		if !strings.Contains(output, expected) {
			t.Errorf("missing %q in output: %s", expected, output)
		}
	}
}

func TestAdaptiveHandler_Handle_TypePreservation(t *testing.T) {
	al, buf := newTestAdaptiveLogger(t)
	defer func() {
		if err := al.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	h := NewAdaptiveHandler(al)
	logger := slog.New(h)

	logger.Info("typed fields",
		"str", "value",
		"num", 42,
		"flag", true,
		"pi", 3.14,
	)

	time.Sleep(50 * time.Millisecond)
	if err := al.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}

	output := buf.String()
	// WHY individual checks: each type must survive the slog->iris conversion
	if !strings.Contains(output, `"str":"value"`) {
		t.Errorf("string field missing: %s", output)
	}
	if !strings.Contains(output, `"num":42`) {
		t.Errorf("int field missing: %s", output)
	}
	if !strings.Contains(output, `"flag":true`) {
		t.Errorf("bool field missing: %s", output)
	}
	if !strings.Contains(output, `"pi":3.14`) {
		t.Errorf("float field missing: %s", output)
	}
}

func TestAdaptiveHandler_Enabled_AlwaysTrue(t *testing.T) {
	al, _ := newTestAdaptiveLogger(t)
	defer func() {
		if err := al.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	h := NewAdaptiveHandler(al)
	ctx := context.Background()

	// WHY always true: level filtering is iris's job, not slog's.
	// Returning false here would prevent records from ever reaching iris.
	levels := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	for _, l := range levels {
		if !h.Enabled(ctx, l) {
			t.Errorf("Enabled(%v) = false, want true", l)
		}
	}
}

func TestAdaptiveHandler_WithAttrs_PassThrough(t *testing.T) {
	al, _ := newTestAdaptiveLogger(t)
	defer func() {
		if err := al.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	h := NewAdaptiveHandler(al)
	h2 := h.WithAttrs([]slog.Attr{slog.String("static", "val")})

	// WHY: WithAttrs must return a valid handler (not nil, not panic)
	if h2 == nil {
		t.Fatal("WithAttrs returned nil")
	}
}

func TestAdaptiveHandler_WithGroup_PassThrough(t *testing.T) {
	al, _ := newTestAdaptiveLogger(t)
	defer func() {
		if err := al.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	h := NewAdaptiveHandler(al)
	h2 := h.WithGroup("mygroup")

	if h2 == nil {
		t.Fatal("WithGroup returned nil")
	}
}

func TestAdaptiveHandler_ConcurrentWrites(t *testing.T) {
	al, buf := newTestAdaptiveLogger(t)
	defer func() {
		if err := al.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	h := NewAdaptiveHandler(al)
	logger := slog.New(h)

	// WHY 20 goroutines: exercises AdaptiveLogger's contention detection
	// and validates that the handler is safe for concurrent use
	var wg sync.WaitGroup
	for g := 0; g < 20; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				logger.Info("concurrent msg", "goroutine", id, "iter", i)
			}
		}(g)
	}
	wg.Wait()

	time.Sleep(100 * time.Millisecond)
	if err := al.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// We sent 20*50 = 1000 messages. Allow some loss from backpressure,
	// but we should see a substantial number of them.
	if len(lines) < 500 {
		t.Errorf("expected at least 500 lines from 1000 writes, got %d", len(lines))
	}
}

func TestAdaptiveHandler_HandleError_NilContext(t *testing.T) {
	al, _ := newTestAdaptiveLogger(t)
	defer func() {
		if err := al.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	h := NewAdaptiveHandler(al)

	// WHY: the handler ignores ctx (signature uses _ context.Context).
	// We pass context.TODO() to confirm Handle works with a non-nil ctx.
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "minimal ctx", 0)
	if err := h.Handle(context.TODO(), record); err != nil {
		t.Errorf("Handle(context.TODO()) returned error: %v", err)
	}
}
