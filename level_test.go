// level_test.go: Comprehensive test suite for Iris logging level functionality
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestLevelConstants verifies that level constants have correct values
func TestLevelConstants(t *testing.T) {
	testCases := []struct {
		level    Level
		expected int32
		name     string
	}{
		{Debug, -1, "Debug"},
		{Info, 0, "Info"},
		{Warn, 1, "Warn"},
		{Error, 2, "Error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if int32(tc.level) != tc.expected {
				t.Errorf("Expected %s level to be %d, got %d", tc.name, tc.expected, int32(tc.level))
			}
		})
	}
}

// TestLevelString tests the String method for all levels
func TestLevelString(t *testing.T) {
	testCases := []struct {
		level    Level
		expected string
	}{
		{Debug, "debug"},
		{Info, "info"},
		{Warn, "warn"},
		{Error, "error"},
		{Level(-10), "unknown"},
		{Level(10), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Level_%d", int32(tc.level)), func(t *testing.T) {
			result := tc.level.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// TestLevelEnabled tests the Enabled method
func TestLevelEnabled(t *testing.T) {
	testCases := []struct {
		level    Level
		minLevel Level
		expected bool
		name     string
	}{
		{Debug, Debug, true, "Debug_enabled_at_Debug"},
		{Info, Debug, true, "Info_enabled_at_Debug"},
		{Warn, Debug, true, "Warn_enabled_at_Debug"},
		{Error, Debug, true, "Error_enabled_at_Debug"},
		{Debug, Info, false, "Debug_disabled_at_Info"},
		{Info, Info, true, "Info_enabled_at_Info"},
		{Warn, Info, true, "Warn_enabled_at_Info"},
		{Error, Info, true, "Error_enabled_at_Info"},
		{Debug, Warn, false, "Debug_disabled_at_Warn"},
		{Info, Warn, false, "Info_disabled_at_Warn"},
		{Warn, Warn, true, "Warn_enabled_at_Warn"},
		{Error, Warn, true, "Error_enabled_at_Warn"},
		{Debug, Error, false, "Debug_disabled_at_Error"},
		{Info, Error, false, "Info_disabled_at_Error"},
		{Warn, Error, false, "Warn_disabled_at_Error"},
		{Error, Error, true, "Error_enabled_at_Error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.level.Enabled(tc.minLevel)
			if result != tc.expected {
				t.Errorf("Expected %v for level %s enabled at %s, got %v",
					tc.expected, tc.level.String(), tc.minLevel.String(), result)
			}
		})
	}
}

// TestLevelConvenienceMethods tests IsDebug, IsInfo, IsWarn, IsError, IsDPanic, IsPanic, IsFatal methods
func TestLevelConvenienceMethods(t *testing.T) {
	testCases := []struct {
		level       Level
		isDebug     bool
		isInfo      bool
		isWarn      bool
		isError     bool
		isDPanic    bool
		isPanic     bool
		isFatal     bool
		description string
	}{
		{Debug, true, false, false, false, false, false, false, "Debug level"},
		{Info, false, true, false, false, false, false, false, "Info level"},
		{Warn, false, false, true, false, false, false, false, "Warn level"},
		{Error, false, false, false, true, false, false, false, "Error level"},
		{DPanic, false, false, false, false, true, false, false, "DPanic level"},
		{Panic, false, false, false, false, false, true, false, "Panic level"},
		{Fatal, false, false, false, false, false, false, true, "Fatal level"},
		{Level(-10), false, false, false, false, false, false, false, "Unknown level"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.level.IsDebug() != tc.isDebug {
				t.Errorf("IsDebug() expected %v, got %v", tc.isDebug, tc.level.IsDebug())
			}
			if tc.level.IsInfo() != tc.isInfo {
				t.Errorf("IsInfo() expected %v, got %v", tc.isInfo, tc.level.IsInfo())
			}
			if tc.level.IsWarn() != tc.isWarn {
				t.Errorf("IsWarn() expected %v, got %v", tc.isWarn, tc.level.IsWarn())
			}
			if tc.level.IsError() != tc.isError {
				t.Errorf("IsError() expected %v, got %v", tc.isError, tc.level.IsError())
			}
			if tc.level.IsDPanic() != tc.isDPanic {
				t.Errorf("IsDPanic() expected %v, got %v", tc.isDPanic, tc.level.IsDPanic())
			}
			if tc.level.IsPanic() != tc.isPanic {
				t.Errorf("IsPanic() expected %v, got %v", tc.isPanic, tc.level.IsPanic())
			}
			if tc.level.IsFatal() != tc.isFatal {
				t.Errorf("IsFatal() expected %v, got %v", tc.isFatal, tc.level.IsFatal())
			}
		})
	}
}

// TestParseLevel tests the ParseLevel function
func TestParseLevel(t *testing.T) {
	testCases := []struct {
		input       string
		expected    Level
		shouldError bool
		description string
	}{
		{"debug", Debug, false, "lowercase debug"},
		{"DEBUG", Debug, false, "uppercase debug"},
		{"Debug", Debug, false, "mixed case debug"},
		{"  debug  ", Debug, false, "debug with whitespace"},
		{"info", Info, false, "lowercase info"},
		{"INFO", Info, false, "uppercase info"},
		{"warn", Warn, false, "lowercase warn"},
		{"warning", Warn, false, "warning alias"},
		{"WARN", Warn, false, "uppercase warn"},
		{"error", Error, false, "lowercase error"},
		{"err", Error, false, "error alias"},
		{"ERROR", Error, false, "uppercase error"},
		{"dpanic", DPanic, false, "lowercase dpanic"},
		{"DPANIC", DPanic, false, "uppercase dpanic"},
		{"panic", Panic, false, "lowercase panic"},
		{"PANIC", Panic, false, "uppercase panic"},
		{"fatal", Fatal, false, "lowercase fatal"},
		{"FATAL", Fatal, false, "uppercase fatal"},
		{"", Info, false, "empty string defaults to info"},
		{"  ", Info, false, "whitespace only defaults to info after trim"},
		{"invalid", Info, true, "invalid level"},
		{"trace", Info, true, "unsupported level"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := ParseLevel(tc.input)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tc.input)
				}
				if result != tc.expected {
					t.Errorf("Expected level %s for invalid input %q, got %s",
						tc.expected.String(), tc.input, result.String())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("Expected level %s for input %q, got %s",
						tc.expected.String(), tc.input, result.String())
				}
			}
		})
	}
}

// TestLevelMarshalText tests the MarshalText method
func TestLevelMarshalText(t *testing.T) {
	testCases := []struct {
		level       Level
		expected    string
		shouldError bool
		description string
	}{
		{Debug, "debug", false, "Debug level"},
		{Info, "info", false, "Info level"},
		{Warn, "warn", false, "Warn level"},
		{Error, "error", false, "Error level"},
		{Level(-10), "", true, "Invalid level"},
		{Level(10), "", true, "Invalid level"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := tc.level.MarshalText()

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for level %d, but got none", int32(tc.level))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for level %s: %v", tc.level.String(), err)
				}
				if string(result) != tc.expected {
					t.Errorf("Expected %q, got %q", tc.expected, string(result))
				}
			}
		})
	}
}

// TestLevelUnmarshalText tests the UnmarshalText method
func TestLevelUnmarshalText(t *testing.T) {
	testCases := []struct {
		input       string
		expected    Level
		shouldError bool
		description string
	}{
		{"debug", Debug, false, "Valid debug"},
		{"info", Info, false, "Valid info"},
		{"warn", Warn, false, "Valid warn"},
		{"error", Error, false, "Valid error"},
		{"invalid", Info, true, "Invalid level"},
		{"", Info, false, "Empty string"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			var level Level
			err := level.UnmarshalText([]byte(tc.input))

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tc.input, err)
				}
				if level != tc.expected {
					t.Errorf("Expected level %s, got %s", tc.expected.String(), level.String())
				}
			}
		})
	}
}

// TestLevelUnmarshalTextNilPointer tests UnmarshalText with nil pointer
func TestLevelUnmarshalTextNilPointer(t *testing.T) {
	var level *Level
	err := level.UnmarshalText([]byte("info"))

	if err == nil {
		t.Error("Expected error when unmarshaling to nil pointer")
	}

	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("Expected error message to mention nil pointer, got: %v", err)
	}
}

// TestLevelJSONSerialization tests JSON serialization/deserialization
func TestLevelJSONSerialization(t *testing.T) {
	testCases := []struct {
		level Level
		json  string
	}{
		{Debug, `"debug"`},
		{Info, `"info"`},
		{Warn, `"warn"`},
		{Error, `"error"`},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("JSON_%s", tc.level.String()), func(t *testing.T) {
			// Test marshaling
			jsonBytes, err := json.Marshal(tc.level)
			if err != nil {
				t.Fatalf("Failed to marshal level %s: %v", tc.level.String(), err)
			}

			if string(jsonBytes) != tc.json {
				t.Errorf("Expected JSON %s, got %s", tc.json, string(jsonBytes))
			}

			// Test unmarshaling
			var level Level
			err = json.Unmarshal(jsonBytes, &level)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON %s: %v", tc.json, err)
			}

			if level != tc.level {
				t.Errorf("Expected level %s after round-trip, got %s",
					tc.level.String(), level.String())
			}
		})
	}
}

// TestAtomicLevel tests the AtomicLevel functionality
func TestAtomicLevel(t *testing.T) {
	// Test creation
	al := NewAtomicLevel(Info)
	if al == nil {
		t.Fatal("NewAtomicLevel returned nil")
	}

	// Test initial level
	if al.Level() != Info {
		t.Errorf("Expected initial level Info, got %s", al.Level().String())
	}

	// Test setting level
	al.SetLevel(Debug)
	if al.Level() != Debug {
		t.Errorf("Expected level Debug after SetLevel, got %s", al.Level().String())
	}

	// Test Enabled method
	if !al.Enabled(Debug) {
		t.Error("Debug should be enabled when AtomicLevel is Debug")
	}

	if !al.Enabled(Info) {
		t.Error("Info should be enabled when AtomicLevel is Debug (Info >= Debug)")
	}

	// Test String method
	if al.String() != "debug" {
		t.Errorf("Expected String() to return 'debug', got %s", al.String())
	}
}

// TestAtomicLevelSerialization tests AtomicLevel JSON serialization
func TestAtomicLevelSerialization(t *testing.T) {
	al := NewAtomicLevel(Warn)

	// Test MarshalText
	text, err := al.MarshalText()
	if err != nil {
		t.Fatalf("Failed to marshal AtomicLevel: %v", err)
	}

	if string(text) != "warn" {
		t.Errorf("Expected 'warn', got %s", string(text))
	}

	// Test UnmarshalText
	err = al.UnmarshalText([]byte("error"))
	if err != nil {
		t.Fatalf("Failed to unmarshal AtomicLevel: %v", err)
	}

	if al.Level() != Error {
		t.Errorf("Expected Error level after unmarshaling, got %s", al.Level().String())
	}
}

// TestAtomicLevelConcurrency tests AtomicLevel under concurrent conditions
func TestAtomicLevelConcurrency(t *testing.T) {
	al := NewAtomicLevel(Info)
	const numGoroutines = 100
	const operationsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines that read and write levels
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Alternate between setting different levels
				if j%2 == 0 {
					al.SetLevel(Debug)
				} else {
					al.SetLevel(Error)
				}

				// Read the level
				level := al.Level()

				// Check if level is valid
				if !IsValidLevel(level) {
					t.Errorf("Invalid level %d from goroutine %d", int32(level), routineID)
				}

				// Test Enabled method
				al.Enabled(Info)
			}
		}(i)
	}

	wg.Wait()

	// Final level should be valid
	finalLevel := al.Level()
	if !IsValidLevel(finalLevel) {
		t.Errorf("Final level %d is invalid", int32(finalLevel))
	}
}

// TestLevelFlag tests the LevelFlag functionality
func TestLevelFlag(t *testing.T) {
	var level Level = Info
	lf := NewLevelFlag(&level)

	if lf == nil {
		t.Fatal("NewLevelFlag returned nil")
	}

	// Test initial string representation
	if lf.String() != "info" {
		t.Errorf("Expected 'info', got %s", lf.String())
	}

	// Test Type method
	if lf.Type() != "level" {
		t.Errorf("Expected Type() to return 'level', got %s", lf.Type())
	}

	// Test Set method with valid levels
	validLevels := []string{"debug", "info", "warn", "error"}
	expectedLevels := []Level{Debug, Info, Warn, Error}

	for i, levelStr := range validLevels {
		err := lf.Set(levelStr)
		if err != nil {
			t.Errorf("Unexpected error setting level %s: %v", levelStr, err)
		}

		if level != expectedLevels[i] {
			t.Errorf("Expected level %s, got %s", expectedLevels[i].String(), level.String())
		}
	}

	// Test Set method with invalid level
	err := lf.Set("invalid")
	if err == nil {
		t.Error("Expected error for invalid level")
	}
}

// TestLevelFlagNilPointer tests LevelFlag with nil pointer
func TestLevelFlagNilPointer(t *testing.T) {
	lf := NewLevelFlag(nil)

	// Test String with nil level
	str := lf.String()
	if str != "info" {
		t.Errorf("Expected default 'info' for nil level, got %s", str)
	}

	// Test Set with nil level
	err := lf.Set("debug")
	if err == nil {
		t.Error("Expected error when setting level on nil pointer")
	}
}

// TestLevelFlagWithCommandLine tests LevelFlag integration with flag package
func TestLevelFlagWithCommandLine(t *testing.T) {
	var level Level = Info

	// Create a new flag set to avoid interfering with other tests
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Var(NewLevelFlag(&level), "level", "Set the log level")

	// Test parsing valid level
	err := fs.Parse([]string{"-level", "debug"})
	if err != nil {
		t.Fatalf("Failed to parse flag: %v", err)
	}

	if level != Debug {
		t.Errorf("Expected Debug level, got %s", level.String())
	}
}

// TestAllLevels tests the AllLevels function
func TestAllLevels(t *testing.T) {
	levels := AllLevels()
	expected := []Level{Debug, Info, Warn, Error}

	if len(levels) != len(expected) {
		t.Errorf("Expected %d levels, got %d", len(expected), len(levels))
	}

	for i, level := range levels {
		if level != expected[i] {
			t.Errorf("Expected level %s at index %d, got %s",
				expected[i].String(), i, level.String())
		}
	}
}

// TestAllLevelNames tests the AllLevelNames function
func TestAllLevelNames(t *testing.T) {
	names := AllLevelNames()
	expected := []string{"debug", "info", "warn", "error"}

	if len(names) != len(expected) {
		t.Errorf("Expected %d level names, got %d", len(expected), len(names))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("Expected name %s at index %d, got %s", expected[i], i, name)
		}
	}
}

// TestIsValidLevel tests the IsValidLevel function
func TestIsValidLevel(t *testing.T) {
	testCases := []struct {
		level    Level
		expected bool
		name     string
	}{
		{Debug, true, "Debug is valid"},
		{Info, true, "Info is valid"},
		{Warn, true, "Warn is valid"},
		{Error, true, "Error is valid"},
		{Level(-10), false, "Level -10 is invalid"},
		{Level(10), false, "Level 10 is invalid"},
		{Level(-2), false, "Level -2 is invalid"},
		{Level(3), false, "Level 3 is invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidLevel(tc.level)
			if result != tc.expected {
				t.Errorf("Expected IsValidLevel(%d) to be %v, got %v",
					int32(tc.level), tc.expected, result)
			}
		})
	}
}

// TestLevelComparison tests level comparison operations
func TestLevelComparison(t *testing.T) {
	// Test that levels are properly ordered
	if Debug >= Info {
		t.Error("Debug should be less than Info")
	}

	if Info >= Warn {
		t.Error("Info should be less than Warn")
	}

	if Warn >= Error {
		t.Error("Warn should be less than Error")
	}

	// Test equality - this seems redundant but ensures constants are properly defined
	debugCopy := Debug
	infoCopy := Info

	if Debug != debugCopy {
		t.Error("Debug should equal its copy")
	}

	if Info != infoCopy {
		t.Error("Info should equal its copy")
	}
}

// TestLevelNamesMapCompleteness ensures all levels are in the map
func TestLevelNamesMapCompleteness(t *testing.T) {
	// Test that all standard level names are in the map
	standardNames := []string{"debug", "info", "warn", "error"}

	for _, name := range standardNames {
		if _, exists := levelNamesMap[name]; !exists {
			t.Errorf("Level name %s not found in levelNamesMap", name)
		}
	}

	// Test aliases
	aliases := map[string]Level{
		"warning": Warn,
		"err":     Error,
	}

	for alias, expectedLevel := range aliases {
		if level, exists := levelNamesMap[alias]; !exists {
			t.Errorf("Alias %s not found in levelNamesMap", alias)
		} else if level != expectedLevel {
			t.Errorf("Alias %s maps to %s, expected %s",
				alias, level.String(), expectedLevel.String())
		}
	}
}

// TestLevelPerformance is a simple performance test (not a benchmark)
func TestLevelPerformance(t *testing.T) {
	// Test that String() method is fast for valid levels
	for i := 0; i < 1000; i++ {
		_ = Debug.String()
		_ = Info.String()
		_ = Warn.String()
		_ = Error.String()
	}

	// Test that Enabled() method is fast
	for i := 0; i < 1000; i++ {
		_ = Error.Enabled(Debug)
		_ = Warn.Enabled(Info)
		_ = Info.Enabled(Warn)
		_ = Debug.Enabled(Error)
	}

	// Test that ParseLevel is reasonably fast for common cases
	for i := 0; i < 100; i++ {
		_, _ = ParseLevel("info")
		_, _ = ParseLevel("debug")
		_, _ = ParseLevel("warn")
		_, _ = ParseLevel("error")
	}
}
