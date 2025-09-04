// security_test.go: Security features test suite for Iris logging library
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSecretFieldRedaction(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create logger with buffer output
	logger, err := New(Config{
		Level:  Debug,
		Output: WrapWriter(&buf),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseSecurityLogger(t, logger)

	logger.Start()

	// Log with sensitive data
	logger.Info("User authentication",
		Secret("password", "super_secret_password123"),
		Secret("api_key", "sk-1234567890abcdef"),
		Str("username", "john_doe"), // Regular field for comparison
	)

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Verify that sensitive data is redacted
	if strings.Contains(output, "super_secret_password123") {
		t.Error("Password was not redacted in log output")
	}
	if strings.Contains(output, "sk-1234567890abcdef") {
		t.Error("API key was not redacted in log output")
	}

	// Verify that redaction marker is present
	if !strings.Contains(output, `"password":"[REDACTED]"`) {
		t.Error("Password field redaction marker not found")
	}
	if !strings.Contains(output, `"api_key":"[REDACTED]"`) {
		t.Error("API key field redaction marker not found")
	}

	// Verify that non-sensitive data is still present
	if !strings.Contains(output, `"username":"john_doe"`) {
		t.Error("Username field was incorrectly redacted")
	}

	t.Logf("Log output: %s", output)
}

func TestSecretFieldType(t *testing.T) {
	// Test that Secret() creates the correct field type
	field := Secret("test_key", "sensitive_value")

	if field.K != "test_key" {
		t.Errorf("Expected key 'test_key', got '%s'", field.K)
	}
	if field.T != kindSecret {
		t.Errorf("Expected type kindSecret (%d), got %d", kindSecret, field.T)
	}
	if field.Str != "sensitive_value" {
		t.Errorf("Expected value 'sensitive_value', got '%s'", field.Str)
	}
}

func TestMultipleSecretFields(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New(Config{
		Level:  Info,
		Output: WrapWriter(&buf),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseSecurityLogger(t, logger)

	logger.Start()

	// Log with multiple secret fields
	logger.Warn("Security event",
		Secret("password", "password123"),
		Secret("credit_card", "4111-1111-1111-1111"),
		Secret("ssn", "123-45-6789"),
		Str("event_type", "login_attempt"),
		Int("attempt_count", 3),
	)

	time.Sleep(50 * time.Millisecond)
	output := buf.String()

	// Verify all secrets are redacted
	sensitiveData := []string{
		"password123",
		"4111-1111-1111-1111",
		"123-45-6789",
	}

	for _, sensitive := range sensitiveData {
		if strings.Contains(output, sensitive) {
			t.Errorf("Sensitive data '%s' was not redacted", sensitive)
		}
	}

	// Verify redaction markers
	redactionMarkers := []string{
		`"password":"[REDACTED]"`,
		`"credit_card":"[REDACTED]"`,
		`"ssn":"[REDACTED]"`,
	}

	for _, marker := range redactionMarkers {
		if !strings.Contains(output, marker) {
			t.Errorf("Redaction marker '%s' not found", marker)
		}
	}

	// Verify non-sensitive data is preserved
	if !strings.Contains(output, `"event_type":"login_attempt"`) {
		t.Error("Non-sensitive string field was incorrectly redacted")
	}
	if !strings.Contains(output, `"attempt_count":3`) {
		t.Error("Non-sensitive integer field was incorrectly processed")
	}
}

// Helper function for safe logger cleanup
func safeCloseSecurityLogger(t *testing.T, logger *Logger) {
	if err := logger.Close(); err != nil {
		t.Logf("Warning: Error closing logger in test: %v", err)
	}
}
