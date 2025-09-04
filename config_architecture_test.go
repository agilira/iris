// config_architecture_test.go: Comprehensive tests for Architecture enum functions
//
// This file provides complete test coverage for Architecture.String() and ParseArchitecture()
// functions to achieve 100% coverage for config.go architecture-related functions.
//
// Tests are comprehensive, OS-aware, and validate all possible code paths including
// edge cases, invalid inputs, and case-insensitive parsing.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"testing"
)

// TestArchitecture_String validates Architecture.String() method
// Tests all defined architecture values and unknown values
func TestArchitecture_String(t *testing.T) {
	tests := []struct {
		name     string
		arch     Architecture
		expected string
	}{
		{
			name:     "single_ring_architecture",
			arch:     SingleRing,
			expected: "single",
		},
		{
			name:     "threaded_rings_architecture",
			arch:     ThreadedRings,
			expected: "threaded",
		},
		{
			name:     "unknown_architecture_negative",
			arch:     Architecture(-1),
			expected: "unknown",
		},
		{
			name:     "unknown_architecture_high_value",
			arch:     Architecture(100),
			expected: "unknown",
		},
		{
			name:     "unknown_architecture_between_valid",
			arch:     Architecture(10),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.arch.String()
			if result != tt.expected {
				t.Errorf("Architecture.String() = %q, expected %q for architecture %d",
					result, tt.expected, int(tt.arch))
			}
		})
	}
}

// TestParseArchitecture validates ParseArchitecture() function
// Tests all valid and invalid architecture strings with case variations
func TestParseArchitecture(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Architecture
		expectError bool
		errorMsg    string
	}{
		// Valid SingleRing variations
		{
			name:        "parse_single_lowercase",
			input:       "single",
			expected:    SingleRing,
			expectError: false,
		},
		{
			name:        "parse_single_titlecase",
			input:       "Single",
			expected:    SingleRing,
			expectError: false,
		},
		{
			name:        "parse_single_uppercase",
			input:       "SINGLE",
			expected:    SingleRing,
			expectError: false,
		},

		// Valid ThreadedRings variations
		{
			name:        "parse_threaded_lowercase",
			input:       "threaded",
			expected:    ThreadedRings,
			expectError: false,
		},
		{
			name:        "parse_threaded_titlecase",
			input:       "Threaded",
			expected:    ThreadedRings,
			expectError: false,
		},
		{
			name:        "parse_threaded_uppercase",
			input:       "THREADED",
			expected:    ThreadedRings,
			expectError: false,
		},
		{
			name:        "parse_multi_lowercase",
			input:       "multi",
			expected:    ThreadedRings,
			expectError: false,
		},
		{
			name:        "parse_multi_titlecase",
			input:       "Multi",
			expected:    ThreadedRings,
			expectError: false,
		},
		{
			name:        "parse_multi_uppercase",
			input:       "MULTI",
			expected:    ThreadedRings,
			expectError: false,
		},

		// Invalid values - should return error and default to SingleRing
		{
			name:        "parse_invalid_empty",
			input:       "",
			expected:    SingleRing,
			expectError: true,
			errorMsg:    "unknown architecture: ",
		},
		{
			name:        "parse_invalid_random",
			input:       "random",
			expected:    SingleRing,
			expectError: true,
			errorMsg:    "unknown architecture: random",
		},
		{
			name:        "parse_invalid_mixed_case",
			input:       "SiNgLe",
			expected:    SingleRing,
			expectError: true,
			errorMsg:    "unknown architecture: SiNgLe",
		},
		{
			name:        "parse_invalid_partial_match",
			input:       "sing",
			expected:    SingleRing,
			expectError: true,
			errorMsg:    "unknown architecture: sing",
		},
		{
			name:        "parse_invalid_special_chars",
			input:       "single!",
			expected:    SingleRing,
			expectError: true,
			errorMsg:    "unknown architecture: single!",
		},
		{
			name:        "parse_invalid_numeric",
			input:       "123",
			expected:    SingleRing,
			expectError: true,
			errorMsg:    "unknown architecture: 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseArchitecture(tt.input)

			// Check architecture result
			if result != tt.expected {
				t.Errorf("ParseArchitecture(%q) = %d, expected %d",
					tt.input, int(result), int(tt.expected))
			}

			// Check error expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseArchitecture(%q) expected error but got nil", tt.input)
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("ParseArchitecture(%q) error = %q, expected %q",
						tt.input, err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParseArchitecture(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

// TestArchitecture_RoundTrip validates String() and ParseArchitecture() work together
// Tests that valid architectures can be converted to string and back
func TestArchitecture_RoundTrip(t *testing.T) {
	architectures := []Architecture{SingleRing, ThreadedRings}

	for _, arch := range architectures {
		t.Run(fmt.Sprintf("roundtrip_%s", arch.String()), func(t *testing.T) {
			// Convert to string
			str := arch.String()

			// Parse back to architecture
			parsed, err := ParseArchitecture(str)
			if err != nil {
				t.Errorf("Round-trip failed: ParseArchitecture(%q) error: %v", str, err)
			}

			// Verify they match
			if parsed != arch {
				t.Errorf("Round-trip failed: %s -> %q -> %s",
					arch.String(), str, parsed.String())
			}
		})
	}
}
