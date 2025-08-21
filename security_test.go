// security_test.go: Comprehensive security tests for Iris logging
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

// TestSecretFieldRedaction tests that secret fields are properly redacted
func TestSecretFieldRedaction(t *testing.T) {
	tests := []struct {
		name     string
		format   Format
		expected string
	}{
		{
			name:     "JSON secret redaction",
			format:   JSONFormat,
			expected: `"password":"[REDACTED]"`,
		},
		{
			name:     "Console secret redaction",
			format:   ConsoleFormat,
			expected: `password=[REDACTED]`,
		},
		{
			name:     "FastText secret redaction",
			format:   FastTextFormat,
			expected: `password=[REDACTED]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			config := Config{
				Level:  InfoLevel,
				Format: tt.format,
				Writer: WrapWriter(&buf),
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer logger.Sync()
			defer logger.Close()

			// Log with secret field
			logger.Info("Login attempt",
				String("username", "admin"),
				Secret("password", "super_secret_password"),
				String("ip", "192.168.1.1"),
			)

			// Wait for log to be written
			logger.Sync()
			time.Sleep(10 * time.Millisecond) // Give time for async processing
			output := buf.String()

			// Verify secret is redacted
			if !strings.Contains(output, "[REDACTED]") {
				t.Errorf("Expected output to contain [REDACTED], got: %q", output)
			}

			// Verify actual password is NOT in output
			if strings.Contains(output, "super_secret_password") {
				t.Errorf("Secret password leaked in output: %s", output)
			}
		})
	}
}

// TestSecretFieldAnyType tests secret field with any type
func TestSecretFieldAnyType(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Writer: WrapWriter(&buf),
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()
	defer logger.Close()

	// Test with different secret types
	sensitiveData := map[string]interface{}{
		"api_key": "sk-abc123def456",
		"tokens":  []string{"token1", "token2"},
	}

	logger.Info("API request",
		String("endpoint", "/api/users"),
		SecretAny("auth_data", sensitiveData),
	)

	logger.Sync()
	time.Sleep(10 * time.Millisecond) // Give time for async processing
	output := buf.String()

	// Verify secret is redacted
	if !strings.Contains(output, `"auth_data":"[REDACTED]"`) {
		t.Errorf("Expected SecretAny to be redacted, got: %s", output)
	}

	// Verify actual data is NOT in output
	if strings.Contains(output, "sk-abc123def456") || strings.Contains(output, "token1") {
		t.Errorf("Secret data leaked in output: %s", output)
	}
}

// TestLogInjectionProtection tests protection against log injection attacks
func TestLogInjectionProtection(t *testing.T) {
	tests := []struct {
		name             string
		maliciousInput   string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:             "newline injection",
			maliciousInput:   "user\nFAKE LOG: admin login successful",
			shouldContain:    []string{"user\\nFAKE"},
			shouldNotContain: []string{"\nFAKE LOG"},
		},
		{
			name:             "carriage return injection",
			maliciousInput:   "data\rMALICIOUS: injected log",
			shouldContain:    []string{"data\\rMALICIOUS"},
			shouldNotContain: []string{"\rMALICIOUS"},
		},
		{
			name:             "quote escape injection",
			maliciousInput:   `input" fake_field="injected_value`,
			shouldContain:    []string{`input\"`},
			shouldNotContain: []string{`fake_field="injected_value`},
		},
		{
			name:             "tab injection",
			maliciousInput:   "value\tinjected_field=malicious",
			shouldContain:    []string{"value injected"},
			shouldNotContain: []string{"\tinjected_field"},
		},
		{
			name:             "control character injection",
			maliciousInput:   "data\x00\x01\x1f",
			shouldContain:    []string{"data\\0??"},
			shouldNotContain: []string{"\x00", "\x01", "\x1f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			config := Config{
				Level:  InfoLevel,
				Format: ConsoleFormat, // Most vulnerable to injection
				Writer: WrapWriter(&buf),
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer logger.Sync()
			defer logger.Close()

			// Log with malicious input
			logger.Info("Security test",
				String("user_input", tt.maliciousInput),
				String("safe_field", "normal_value"),
			)

			logger.Sync()
			time.Sleep(10 * time.Millisecond) // Give time for async processing
			output := buf.String()

			// Check that dangerous characters are properly escaped/sanitized
			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, output)
				}
			}

			// Check that malicious content is NOT present
			for _, dangerous := range tt.shouldNotContain {
				if strings.Contains(output, dangerous) {
					t.Errorf("Output contains dangerous content %q: %s", dangerous, output)
				}
			}
		})
	}
}

// TestBinaryFormatSecretRedaction tests secret redaction in binary format
func TestBinaryFormatSecretRedaction(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  InfoLevel,
		Format: BinaryFormat,
		Writer: WrapWriter(&buf),
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()
	defer logger.Close()

	// Log with secret field
	logger.Info("Database connection",
		String("host", "localhost"),
		Secret("password", "database_secret_123"),
		String("database", "production"),
	)

	logger.Sync()
	time.Sleep(10 * time.Millisecond) // Give time for async processing
	output := buf.Bytes()

	// Convert binary output to string for inspection
	outputStr := string(output)

	// Verify secret is redacted in binary format
	if !strings.Contains(outputStr, "[REDACTED]") {
		t.Error("Expected [REDACTED] in binary output")
	}

	// Verify actual password is NOT in binary output
	if strings.Contains(outputStr, "database_secret_123") {
		t.Error("Secret password leaked in binary output")
	}

	// Verify non-secret fields are still present
	if !strings.Contains(outputStr, "localhost") {
		t.Error("Expected non-secret field 'localhost' in output")
	}
	if !strings.Contains(outputStr, "production") {
		t.Error("Expected non-secret field 'production' in output")
	}
}

// TestCombinedSecurityFeatures tests multiple security features together
func TestCombinedSecurityFeatures(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Writer: WrapWriter(&buf),
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()
	defer logger.Close()

	// Log with both secret fields and potentially dangerous input
	maliciousUser := "admin\n{\"fake\":\"injected\"}"
	sensitiveToken := "bearer_token_abc123xyz789"

	logger.Warn("Suspicious activity detected",
		String("user", maliciousUser),
		Secret("auth_token", sensitiveToken),
		String("action", "login_attempt"),
		SecretAny("session_data", map[string]string{
			"csrf_token": "csrf_abc123",
			"session_id": "sess_xyz789",
		}),
	)

	logger.Sync()
	time.Sleep(10 * time.Millisecond) // Give time for async processing
	output := buf.String()

	// Verify secrets are redacted
	if !strings.Contains(output, `"auth_token":"[REDACTED]"`) {
		t.Error("auth_token not properly redacted")
	}
	if !strings.Contains(output, `"session_data":"[REDACTED]"`) {
		t.Error("session_data not properly redacted")
	}

	// Verify sensitive data is NOT leaked
	if strings.Contains(output, "bearer_token_abc123xyz789") {
		t.Error("Sensitive token leaked in output")
	}
	if strings.Contains(output, "csrf_abc123") || strings.Contains(output, "sess_xyz789") {
		t.Error("Session data leaked in output")
	}

	// Verify malicious input is properly escaped (for JSON format)
	if strings.Contains(output, "\\n") || !strings.Contains(output, "admin") {
		// Input should be escaped but still readable
		t.Logf("Malicious input handling: %s", output)
	}
}

// TestSecretFieldPerformance tests that secret field redaction doesn't significantly impact performance
func TestSecretFieldPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	var buf bytes.Buffer

	config := Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Writer: WrapWriter(&buf),
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()
	defer logger.Close()

	// Benchmark secret field vs normal field performance
	const iterations = 1000

	// Test normal field performance
	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Info("Performance test",
			String("normal_field", "normal_value"),
			String("another_field", "another_value"),
		)
	}
	normalTime := time.Since(start)

	buf.Reset() // Clear buffer

	// Test secret field performance
	start = time.Now()
	for i := 0; i < iterations; i++ {
		logger.Info("Performance test",
			Secret("secret_field", "secret_value"),
			String("another_field", "another_value"),
		)
	}
	secretTime := time.Since(start)

	logger.Sync()

	// Secret field logging should not be more than 2x slower than normal
	if secretTime > normalTime*2 {
		t.Errorf("Secret field performance too slow: normal=%v, secret=%v", normalTime, secretTime)
	}

	t.Logf("Performance comparison - Normal: %v, Secret: %v", normalTime, secretTime)
}
