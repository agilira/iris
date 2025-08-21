package iris

import (
	"strings"
	"testing"
	"time"
)

// TestFormatString tests the String() method for Format enum
func TestFormatString(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{JSONFormat, "json"},
		{ConsoleFormat, "console"},
		{FastTextFormat, "text"},
		{BinaryFormat, "binary"},
		{Format(999), "unknown"}, // Test unknown format
		{Format(-1), "unknown"},  // Test negative format
	}

	for _, test := range tests {
		result := test.format.String()
		if result != test.expected {
			t.Errorf("Format(%d).String() = %q, expected %q", int(test.format), result, test.expected)
		}
	}
}

// TestFormatStringConsistency tests that Format.String() is consistent with format constants
func TestFormatStringConsistency(t *testing.T) {
	// Test that all defined formats have proper string representations
	formats := []struct {
		format Format
		name   string
	}{
		{JSONFormat, "JSONFormat"},
		{ConsoleFormat, "ConsoleFormat"},
		{FastTextFormat, "FastTextFormat"},
		{BinaryFormat, "BinaryFormat"},
	}

	for _, f := range formats {
		str := f.format.String()
		if str == "unknown" {
			t.Errorf("%s should not return 'unknown', got: %q", f.name, str)
		}
		if str == "" {
			t.Errorf("%s should not return empty string", f.name)
		}
	}
}

// TestDefaultEncoderConfig tests the DefaultEncoderConfig function
func TestDefaultEncoderConfig(t *testing.T) {
	config := DefaultEncoderConfig()

	// Test that all fields have sensible defaults
	if config.BufferSize <= 0 {
		t.Errorf("Expected positive BufferSize, got %d", config.BufferSize)
	}

	if config.TimeFormat == "" {
		t.Error("Expected non-empty TimeFormat")
	}

	// Test specific expected values
	expectedBufferSize := 1024
	if config.BufferSize != expectedBufferSize {
		t.Errorf("Expected BufferSize %d, got %d", expectedBufferSize, config.BufferSize)
	}

	expectedTimeFormat := "2006-01-02T15:04:05.000Z07:00"
	if config.TimeFormat != expectedTimeFormat {
		t.Errorf("Expected TimeFormat %q, got %q", expectedTimeFormat, config.TimeFormat)
	}

	// Test boolean defaults
	if config.IncludeCaller != false {
		t.Errorf("Expected IncludeCaller to be false by default, got %t", config.IncludeCaller)
	}

	if config.IncludeStackTrace != false {
		t.Errorf("Expected IncludeStackTrace to be false by default, got %t", config.IncludeStackTrace)
	}
}

// TestDefaultEncoderConfigImmutability tests that DefaultEncoderConfig returns a fresh copy
func TestDefaultEncoderConfigImmutability(t *testing.T) {
	config1 := DefaultEncoderConfig()
	config2 := DefaultEncoderConfig()

	// Modify first config
	config1.BufferSize = 9999
	config1.TimeFormat = "modified"
	config1.IncludeCaller = true
	config1.IncludeStackTrace = true

	// Second config should be unchanged
	if config2.BufferSize == 9999 {
		t.Error("DefaultEncoderConfig should return independent copies")
	}
	if config2.TimeFormat == "modified" {
		t.Error("DefaultEncoderConfig should return independent copies")
	}
	if config2.IncludeCaller == true {
		t.Error("DefaultEncoderConfig should return independent copies")
	}
	if config2.IncludeStackTrace == true {
		t.Error("DefaultEncoderConfig should return independent copies")
	}

	// Verify second config has correct defaults
	expectedConfig := EncoderConfig{
		BufferSize:        1024,
		TimeFormat:        "2006-01-02T15:04:05.000Z07:00",
		IncludeCaller:     false,
		IncludeStackTrace: false,
	}

	if config2 != expectedConfig {
		t.Errorf("Second config should match default, got %+v", config2)
	}
}

// TestDefaultEncoderConfigTimeFormat tests that the time format is valid
func TestDefaultEncoderConfigTimeFormat(t *testing.T) {
	config := DefaultEncoderConfig()

	// Test that the time format can be used to format time
	now := time.Now()
	formatted := now.Format(config.TimeFormat)

	// Should not panic and should produce a non-empty string
	if formatted == "" {
		t.Error("Time format should produce non-empty formatted time")
	}

	// Should be in ISO 8601 format with timezone
	if len(formatted) < 20 { // Minimum reasonable length for ISO 8601
		t.Errorf("Formatted time seems too short: %q", formatted)
	}

	// Should contain expected components
	expectedComponents := []string{"T", ":", "-"}
	for _, component := range expectedComponents {
		if !strings.Contains(formatted, component) {
			t.Errorf("Expected time format to contain %q, formatted: %q", component, formatted)
		}
	}
}
