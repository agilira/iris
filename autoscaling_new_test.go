// autoscaling_v2_test.go: Tests for AdaptiveLogger (lazy dual-mode)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

func safeCloseAdaptiveLogger(t *testing.T, logger *AdaptiveLogger) {
	if err := logger.Close(); err != nil &&
		!strings.Contains(err.Error(), "sync /dev/stdout: invalid argument") &&
		!strings.Contains(err.Error(), "sync /dev/stdout: bad file descriptor") {
		t.Errorf("Failed to close logger: %v", err)
	}
}

func TestScalingMode_String(t *testing.T) {
	tests := []struct {
		mode ScalingMode
		want string
	}{
		{SingleMode, "Single"},
		{MultiMode, "Multi"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("ScalingMode.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestDefaultScalerConfig(t *testing.T) {
	cfg := Config{Level: Info}
	sc := DefaultScalerConfig(cfg)

	if sc.GoroutineThreshold != 2 {
		t.Errorf("GoroutineThreshold = %d, want 2", sc.GoroutineThreshold)
	}
	if sc.ScaleDownCooldown != 5*time.Second {
		t.Errorf("ScaleDownCooldown = %v, want 5s", sc.ScaleDownCooldown)
	}
}

func TestNewAdaptiveLogger(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:   Info,
		Output:  WrapWriter(&buf),
		Encoder: NewJSONEncoder(),
	}

	al, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger failed: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	// Should start in SingleMode
	if al.Mode() != SingleMode {
		t.Errorf("Initial mode = %v, want SingleMode", al.Mode())
	}
}

func TestAdaptiveLogger_SingleProducerStaysSingle(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:   Debug,
		Output:  WrapWriter(&buf),
		Encoder: NewJSONEncoder(),
	}

	al, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger failed: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Single producer logging - should stay in SingleMode
	for i := 0; i < 100; i++ {
		al.Info("single producer message", Int("i", i))
	}

	time.Sleep(10 * time.Millisecond)

	if al.Mode() != SingleMode {
		t.Errorf("Mode = %v after single producer, want SingleMode", al.Mode())
	}

	stats := al.Stats()
	if stats.ScaleUpCount != 0 {
		t.Errorf("ScaleUpCount = %d, want 0 for single producer", stats.ScaleUpCount)
	}
}

func TestAdaptiveLogger_MultiProducerScalesUp(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:   Debug,
		Output:  WrapWriter(&buf),
		Encoder: NewJSONEncoder(),
	}

	sc := DefaultScalerConfig(cfg)
	sc.GoroutineThreshold = 2 // Scale when 2+ concurrent

	al, err := NewAdaptiveLogger(sc)
	if err != nil {
		t.Fatalf("NewAdaptiveLogger failed: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Multi-producer burst - should trigger scale up
	var wg sync.WaitGroup
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				al.Info("multi producer message", Int("goroutine", id), Int("i", i))
			}
		}(g)
	}
	wg.Wait()

	time.Sleep(50 * time.Millisecond)

	stats := al.Stats()
	if stats.ScaleUpCount == 0 {
		t.Error("Expected ScaleUpCount > 0 for multi-producer burst")
	}
	if stats.TotalWrites < 250 {
		t.Errorf("TotalWrites = %d, want >= 250", stats.TotalWrites)
	}
}

func TestAdaptiveLogger_ScaleDownAfterCooldown(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:   Debug,
		Output:  WrapWriter(&buf),
		Encoder: NewJSONEncoder(),
	}

	sc := DefaultScalerConfig(cfg)
	sc.GoroutineThreshold = 2
	sc.ScaleDownCooldown = 100 * time.Millisecond // Fast for testing

	al, err := NewAdaptiveLogger(sc)
	if err != nil {
		t.Fatalf("NewAdaptiveLogger failed: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Trigger scale up with burst
	var wg sync.WaitGroup
	for g := 0; g < 3; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				al.Info("burst", Int("g", id))
			}
		}(g)
	}
	wg.Wait()

	// Verify we scaled up
	time.Sleep(20 * time.Millisecond)
	if al.Stats().ScaleUpCount == 0 {
		t.Skip("Scale up didn't trigger - timing dependent test")
	}

	// Wait for scale down (cooldown + monitor interval)
	time.Sleep(300 * time.Millisecond)

	if al.Mode() != SingleMode {
		t.Errorf("Mode = %v after cooldown, want SingleMode", al.Mode())
	}

	stats := al.Stats()
	if stats.ScaleDownCount == 0 {
		t.Error("Expected ScaleDownCount > 0 after cooldown")
	}
}

func TestAdaptiveLogger_AllLogLevels(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:   Debug,
		Output:  WrapWriter(&buf),
		Encoder: NewTextEncoder(),
	}

	al, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger failed: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	al.Debug("debug message")
	al.Info("info message")
	al.Warn("warn message")
	al.Error("error message")

	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	for _, level := range []string{"debug", "info", "warn", "error"} {
		if !strings.Contains(strings.ToLower(output), level) {
			t.Errorf("Output missing %s level message", level)
		}
	}
}

func TestAdaptiveLogger_Stats(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:   Info,
		Output:  WrapWriter(&buf),
		Encoder: NewJSONEncoder(),
	}

	al, err := NewAdaptiveLogger(DefaultScalerConfig(cfg))
	if err != nil {
		t.Fatalf("NewAdaptiveLogger failed: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)

	if err := al.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	al.Info("test")
	time.Sleep(10 * time.Millisecond)

	stats := al.Stats()
	if stats.Mode != SingleMode {
		t.Errorf("Stats.Mode = %v, want SingleMode", stats.Mode)
	}
	if stats.TotalWrites != 1 {
		t.Errorf("Stats.TotalWrites = %d, want 1", stats.TotalWrites)
	}
}
