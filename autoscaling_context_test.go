// autoscaling_context_test.go: Tests for AdaptiveLogger.With() and Named()
//
// Validates that structured context (fields and names) propagates correctly
// across the dual-mode scaling boundary. The core invariant is: scaling
// up or down must NEVER lose context fields or component identity.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// --- With() tests ---

func TestAdaptiveLogger_With_EmptyReturnsReceiver(t *testing.T) {
	buf := &bufferedSyncer{}
	al, err := NewAdaptiveLogger(DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	// With() with no fields must return the same instance (no allocation)
	child := al.With()
	if child != al {
		t.Error("With() with no fields must return the same AdaptiveLogger")
	}
}

func TestAdaptiveLogger_With_FieldsAppearInOutput(t *testing.T) {
	buf := &bufferedSyncer{}
	al, err := NewAdaptiveLogger(DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	child := al.With(String("component", "config"), Int("pid", 42))
	child.Info("boot ok")

	// Allow consumer to drain
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, `"component":"config"`) {
		t.Errorf("expected component field in output, got: %s", output)
	}
	if !strings.Contains(output, `"pid":42`) {
		t.Errorf("expected pid field in output, got: %s", output)
	}
}

func TestAdaptiveLogger_With_Chained(t *testing.T) {
	buf := &bufferedSyncer{}
	al, err := NewAdaptiveLogger(DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Chain With() calls -- both fields must appear
	child := al.With(String("app", "metis")).With(String("subsystem", "bus"))
	child.Info("event dispatched")

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, `"app":"metis"`) {
		t.Errorf("expected app field, got: %s", output)
	}
	if !strings.Contains(output, `"subsystem":"bus"`) {
		t.Errorf("expected subsystem field, got: %s", output)
	}
}

func TestAdaptiveLogger_With_ParentUnchanged(t *testing.T) {
	buf := &bufferedSyncer{}
	al, err := NewAdaptiveLogger(DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	_ = al.With(String("child_only", "yes"))

	// Parent must NOT have the child's field
	if len(al.contextFields) != 0 {
		t.Errorf("parent contextFields mutated: got %d, want 0", len(al.contextFields))
	}
}

func TestAdaptiveLogger_With_PropagateToExistingMultiLogger(t *testing.T) {
	buf := &bufferedSyncer{}
	cfg := DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	})
	cfg.GoroutineThreshold = 1 // Force immediate scale-up

	al, err := NewAdaptiveLogger(cfg)
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Force multiLogger creation by writing with contention
	al.Info("trigger scale-up")
	time.Sleep(20 * time.Millisecond)

	// Now With() on a logger that already has multiLogger
	child := al.With(String("after_scale", "true"))

	// Child must have propagated to its multiLogger
	if ml := child.multiLogger.Load(); ml == nil {
		t.Error("With() did not propagate to existing multiLogger")
	}
}

// --- Named() tests ---

func TestAdaptiveLogger_Named_EmptyReturnsReceiver(t *testing.T) {
	buf := &bufferedSyncer{}
	al, err := NewAdaptiveLogger(DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	child := al.Named("")
	if child != al {
		t.Error("Named(\"\") must return the same AdaptiveLogger")
	}
}

func TestAdaptiveLogger_Named_AppearsInOutput(t *testing.T) {
	buf := &bufferedSyncer{}
	al, err := NewAdaptiveLogger(DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	child := al.Named("config")
	child.Info("loaded")

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "config") {
		t.Errorf("expected logger name 'config' in output, got: %s", output)
	}
}

func TestAdaptiveLogger_Named_Hierarchical(t *testing.T) {
	buf := &bufferedSyncer{}
	al, err := NewAdaptiveLogger(DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	child := al.Named("app").Named("config")
	child.Info("dot-separated name")

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "app.config") {
		t.Errorf("expected hierarchical name 'app.config', got: %s", output)
	}

	// Verify contextName was accumulated correctly
	if child.contextName != "app.config" {
		t.Errorf("contextName = %q, want %q", child.contextName, "app.config")
	}
}

func TestAdaptiveLogger_Named_PropagateToExistingMultiLogger(t *testing.T) {
	buf := &bufferedSyncer{}
	cfg := DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	})
	cfg.GoroutineThreshold = 1

	al, err := NewAdaptiveLogger(cfg)
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Force multiLogger init
	al.Info("trigger")
	time.Sleep(20 * time.Millisecond)

	child := al.Named("bus")
	if ml := child.multiLogger.Load(); ml == nil {
		t.Error("Named() did not propagate to existing multiLogger")
	}
}

// --- Combined With()+Named() tests ---

func TestAdaptiveLogger_With_Named_Combined(t *testing.T) {
	buf := &bufferedSyncer{}
	al, err := NewAdaptiveLogger(DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	child := al.Named("store").With(String("db", "sqlite"))
	child.Warn("slow query")

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "store") {
		t.Errorf("expected name 'store' in output, got: %s", output)
	}
	if !strings.Contains(output, `"db":"sqlite"`) {
		t.Errorf("expected db field in output, got: %s", output)
	}
}

// --- Context propagation during scaling tests ---

func TestAdaptiveLogger_With_ContextSurvivesScaleUp(t *testing.T) {
	// WHY: This is the critical invariant. When a child created via With()
	// scales up from SingleMode to MultiMode, the multiLogger must inherit
	// the accumulated context fields. If it does not, log entries in
	// MultiMode would silently lose structured context.
	buf := &bufferedSyncer{}
	cfg := DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	})
	cfg.GoroutineThreshold = 2

	al, err := NewAdaptiveLogger(cfg)
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Create child BEFORE scale-up (multiLogger does not exist yet)
	child := al.Named("bus").With(String("trace_id", "abc-123"))

	// Force scale-up with concurrent writers
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			child.Info("concurrent write", Int("goroutine", id))
		}(i)
	}
	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	output := buf.String()
	// Every line must contain the context field
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("no output produced")
	}

	for i, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, `"trace_id":"abc-123"`) {
			t.Errorf("line %d missing trace_id field after scale-up: %s", i, line)
		}
	}
}

func TestAdaptiveLogger_Named_ContextSurvivesScaleUp(t *testing.T) {
	buf := &bufferedSyncer{}
	cfg := DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	})
	cfg.GoroutineThreshold = 2

	al, err := NewAdaptiveLogger(cfg)
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	child := al.Named("secrets")

	// Force scale-up
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			child.Info("concurrent", Int("g", id))
		}(i)
	}
	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("no output")
	}
	for i, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "secrets") {
			t.Errorf("line %d missing name 'secrets' after scale-up: %s", i, line)
		}
	}
}

// --- All log levels propagate context ---

func TestAdaptiveLogger_With_AllLevels(t *testing.T) {
	buf := &bufferedSyncer{}
	al, err := NewAdaptiveLogger(DefaultScalerConfig(Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
	}))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	child := al.With(String("env", "test"))
	child.Debug("d")
	child.Info("i")
	child.Warn("w")
	child.Error("e")

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	// Count occurrences of the context field
	count := strings.Count(output, `"env":"test"`)
	if count != 4 {
		t.Errorf("expected context field in all 4 levels, appeared %d times. Output:\n%s", count, output)
	}
}
