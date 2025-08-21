package iris

import (
	"strings"
	"testing"
)

// TestLevelConstants tests that level constants have correct values
func TestLevelConstants(t *testing.T) {
	// Test level values are as expected
	if DebugLevel != -1 {
		t.Errorf("DebugLevel should be -1, got %d", DebugLevel)
	}
	if InfoLevel != 0 {
		t.Errorf("InfoLevel should be 0, got %d", InfoLevel)
	}
	if WarnLevel != 1 {
		t.Errorf("WarnLevel should be 1, got %d", WarnLevel)
	}
	if ErrorLevel != 2 {
		t.Errorf("ErrorLevel should be 2, got %d", ErrorLevel)
	}
	if DPanicLevel != 3 {
		t.Errorf("DPanicLevel should be 3, got %d", DPanicLevel)
	}
	if PanicLevel != 4 {
		t.Errorf("PanicLevel should be 4, got %d", PanicLevel)
	}
	if FatalLevel != 5 {
		t.Errorf("FatalLevel should be 5, got %d", FatalLevel)
	}
}

// TestLevelString tests the String() method for all levels
func TestLevelString(t *testing.T) {
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

	for _, test := range tests {
		result := test.level.String()
		if result != test.expected {
			t.Errorf("Level(%d).String() = %q, expected %q", test.level, result, test.expected)
		}
	}
}

// TestLevelStringUnknown tests String() method for unknown levels
func TestLevelStringUnknown(t *testing.T) {
	unknownLevel := Level(100)
	result := unknownLevel.String()
	expected := "Level(100)"
	if result != expected {
		t.Errorf("Level(100).String() = %q, expected %q", result, expected)
	}

	negativeLevel := Level(-10)
	result = negativeLevel.String()
	expected = "Level(-10)"
	if result != expected {
		t.Errorf("Level(-10).String() = %q, expected %q", result, expected)
	}
}

// TestLevelCapitalString tests the CapitalString() method
func TestLevelCapitalString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{DPanicLevel, "DPANIC"},
		{PanicLevel, "PANIC"},
		{FatalLevel, "FATAL"},
	}

	for _, test := range tests {
		result := test.level.CapitalString()
		if result != test.expected {
			t.Errorf("Level(%d).CapitalString() = %q, expected %q", test.level, result, test.expected)
		}
	}
}

// TestLevelCapitalStringUnknown tests CapitalString() for unknown levels
func TestLevelCapitalStringUnknown(t *testing.T) {
	unknownLevel := Level(42)
	result := unknownLevel.CapitalString()
	expected := "LEVEL(42)"
	if result != expected {
		t.Errorf("Level(42).CapitalString() = %q, expected %q", result, expected)
	}
}

// TestLevelEnabled tests the Enabled() method
func TestLevelEnabled(t *testing.T) {
	tests := []struct {
		baseLevel Level
		testLevel Level
		expected  bool
	}{
		// Debug level should enable everything
		{DebugLevel, DebugLevel, true},
		{DebugLevel, InfoLevel, true},
		{DebugLevel, WarnLevel, true},
		{DebugLevel, ErrorLevel, true},
		{DebugLevel, FatalLevel, true},

		// Info level should not enable debug
		{InfoLevel, DebugLevel, false},
		{InfoLevel, InfoLevel, true},
		{InfoLevel, WarnLevel, true},
		{InfoLevel, ErrorLevel, true},
		{InfoLevel, FatalLevel, true},

		// Warn level
		{WarnLevel, DebugLevel, false},
		{WarnLevel, InfoLevel, false},
		{WarnLevel, WarnLevel, true},
		{WarnLevel, ErrorLevel, true},
		{WarnLevel, FatalLevel, true},

		// Error level
		{ErrorLevel, DebugLevel, false},
		{ErrorLevel, InfoLevel, false},
		{ErrorLevel, WarnLevel, false},
		{ErrorLevel, ErrorLevel, true},
		{ErrorLevel, FatalLevel, true},

		// Fatal level
		{FatalLevel, DebugLevel, false},
		{FatalLevel, InfoLevel, false},
		{FatalLevel, WarnLevel, false},
		{FatalLevel, ErrorLevel, false},
		{FatalLevel, FatalLevel, true},
	}

	for _, test := range tests {
		result := test.baseLevel.Enabled(test.testLevel)
		if result != test.expected {
			t.Errorf("Level(%d).Enabled(%d) = %v, expected %v",
				test.baseLevel, test.testLevel, result, test.expected)
		}
	}
}

// TestLevelShouldPanic tests ShouldPanic() method
func TestLevelShouldPanic(t *testing.T) {
	tests := []struct {
		level    Level
		expected bool
	}{
		{DebugLevel, false},
		{InfoLevel, false},
		{WarnLevel, false},
		{ErrorLevel, false},
		{DPanicLevel, true},
		{PanicLevel, true},
		{FatalLevel, false},
	}

	for _, test := range tests {
		result := test.level.ShouldPanic()
		if result != test.expected {
			t.Errorf("Level(%d).ShouldPanic() = %v, expected %v", test.level, result, test.expected)
		}
	}
}

// TestLevelShouldExit tests ShouldExit() method
func TestLevelShouldExit(t *testing.T) {
	tests := []struct {
		level    Level
		expected bool
	}{
		{DebugLevel, false},
		{InfoLevel, false},
		{WarnLevel, false},
		{ErrorLevel, false},
		{DPanicLevel, false},
		{PanicLevel, false},
		{FatalLevel, true},
	}

	for _, test := range tests {
		result := test.level.ShouldExit()
		if result != test.expected {
			t.Errorf("Level(%d).ShouldExit() = %v, expected %v", test.level, result, test.expected)
		}
	}
}

// TestLevelHandleSpecialDPanic tests HandleSpecial for DPanic (should panic)
func TestLevelHandleSpecialDPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("DPanicLevel.HandleSpecial() should have panicked")
		} else {
			// Check panic message
			if msg, ok := r.(string); !ok || msg != "dpanic level log" {
				t.Errorf("DPanicLevel.HandleSpecial() panic message = %v, expected 'dpanic level log'", r)
			}
		}
	}()

	DPanicLevel.HandleSpecial()
}

// TestLevelHandleSpecialPanic tests HandleSpecial for Panic (should panic)
func TestLevelHandleSpecialPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("PanicLevel.HandleSpecial() should have panicked")
		} else {
			// Check panic message
			if msg, ok := r.(string); !ok || msg != "panic level log" {
				t.Errorf("PanicLevel.HandleSpecial() panic message = %v, expected 'panic level log'", r)
			}
		}
	}()

	PanicLevel.HandleSpecial()
}

// TestLevelHandleSpecialNormal tests HandleSpecial for normal levels (should not panic)
func TestLevelHandleSpecialNormal(t *testing.T) {
	normalLevels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel}

	for _, level := range normalLevels {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Level(%d).HandleSpecial() should not have panicked, but got: %v", level, r)
				}
			}()
			level.HandleSpecial()
		}()
	}
}

// TestLevelOrder tests that levels are ordered correctly
func TestLevelOrder(t *testing.T) {
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

// TestLevelStringConsistency tests that String() results are consistent
func TestLevelStringConsistency(t *testing.T) {
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, FatalLevel}

	for _, level := range levels {
		str1 := level.String()
		str2 := level.String()
		if str1 != str2 {
			t.Errorf("Level(%d).String() returned different results: %q vs %q", level, str1, str2)
		}

		cap1 := level.CapitalString()
		cap2 := level.CapitalString()
		if cap1 != cap2 {
			t.Errorf("Level(%d).CapitalString() returned different results: %q vs %q", level, cap1, cap2)
		}

		// CapitalString should be uppercase of String
		expectedCap := strings.ToUpper(str1)
		if cap1 != expectedCap {
			t.Errorf("Level(%d).CapitalString() = %q, expected %q", level, cap1, expectedCap)
		}
	}
}
