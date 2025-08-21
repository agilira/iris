// format_helpers.go: Helper functions for Iris formatting
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

// lastIndexByte returns the last index of byte c in string s, or -1 if not found
func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// sanitizeForLogSafety provides enhanced protection against log injection attacks
// This function goes beyond basic quoting to prevent malicious log format manipulation
func sanitizeForLogSafety(s string) string {
	if !needsLogSafetySanitization(s) {
		return s
	}

	// Replace or escape dangerous sequences that could cause log injection
	buf := make([]byte, 0, len(s)+20) // extra space for escaping
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\n', '\r':
			// Replace newlines with escaped versions to prevent log line injection
			buf = append(buf, '\\')
			if c == '\n' {
				buf = append(buf, 'n')
			} else {
				buf = append(buf, 'r')
			}
		case '\t':
			// Replace tabs with spaces to prevent field misalignment
			buf = append(buf, ' ')
		case '"':
			// Escape quotes to prevent field value escaping
			buf = append(buf, '\\', '"')
		case '\\':
			// Escape backslashes to prevent escape sequence injection
			buf = append(buf, '\\', '\\')
		case '\x00':
			// Replace null bytes (potential binary injection)
			buf = append(buf, '\\', '0')
		default:
			// For other control characters, replace with safe placeholder
			if c < 32 || c == 127 {
				buf = append(buf, '?')
			} else {
				buf = append(buf, c)
			}
		}
	}
	return string(buf)
}

// needsLogSafetySanitization checks if a string contains potentially dangerous characters
func needsLogSafetySanitization(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Check for control characters, newlines, quotes, and other dangerous chars
		if c < 32 || c == 127 || c == '"' || c == '\\' {
			return true
		}
	}
	return false
}

// needsQuotingFastSecure is an enhanced version that considers security implications
func needsQuotingFastSecure(s string) bool {
	if len(s) == 0 {
		return true // Empty strings should be quoted
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		// SECURITY: Quote if contains characters that could cause injection
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '"' || c == '\\' ||
			c == '=' || c == '[' || c == ']' || c < 32 || c == 127 {
			return true
		}
	}
	return false
}
