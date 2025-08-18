// level_test.go: Test the new log levels
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"testing"
)

func TestExtendedLogLevels(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{DPanicLevel, "dpanic"},
		{PanicLevel, "panic"},
		{FatalLevel, "fatal"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLevelOrdering(t *testing.T) {
	// Test that levels are ordered correctly
	if DebugLevel >= InfoLevel {
		t.Error("DebugLevel should be less than InfoLevel")
	}
	if InfoLevel >= WarnLevel {
		t.Error("InfoLevel should be less than WarnLevel")
	}
	if WarnLevel >= ErrorLevel {
		t.Error("WarnLevel should be less than ErrorLevel")
	}
	if ErrorLevel >= DPanicLevel {
		t.Error("ErrorLevel should be less than DPanicLevel")
	}
	if DPanicLevel >= PanicLevel {
		t.Error("DPanicLevel should be less than PanicLevel")
	}
	if PanicLevel >= FatalLevel {
		t.Error("PanicLevel should be less than FatalLevel")
	}
}

func TestLevelSpecialBehavior(t *testing.T) {
	// Test ShouldPanic
	if !DPanicLevel.ShouldPanic() {
		t.Error("DPanicLevel should panic")
	}
	if !PanicLevel.ShouldPanic() {
		t.Error("PanicLevel should panic")
	}
	if InfoLevel.ShouldPanic() {
		t.Error("InfoLevel should not panic")
	}

	// Test ShouldExit
	if !FatalLevel.ShouldExit() {
		t.Error("FatalLevel should exit")
	}
	if InfoLevel.ShouldExit() {
		t.Error("InfoLevel should not exit")
	}
}

// Note: We don't test DPanic and Panic methods here because they actually panic
// and we don't test Fatal because it calls os.Exit(1)
// These would be tested in integration tests with proper isolation
