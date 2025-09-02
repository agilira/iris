// text_encoder_test.go: Comprehensive security tests for TextEncoder
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTextEncoder_SecurityBasic(t *testing.T) {
	encoder := NewTextEncoder()

	// Create a simple record manually for testing
	record := &Record{
		Level:  Info,
		Msg:    "User login attempt",
		Logger: "test",
		Caller: "main.go:42",
		fields: [32]Field{},
		n:      0,
	}

	// Add fields with injection attempts
	record.fields[0] = Str("user", "admin")
	record.fields[1] = Str("malicious", "value\nlevel=error msg=\"HACKED\"")
	record.fields[2] = Secret("password", "secret123")
	record.n = 3

	now := time.Date(2025, 8, 22, 10, 30, 0, 0, time.UTC)

	var buf bytes.Buffer
	encoder.Encode(record, now, &buf)

	output := buf.String()
	t.Logf("Output: %s", output)

	// Security checks - aggressive sanitization should prevent injection
	// The encoder should replace dangerous characters with underscores
	if strings.Contains(output, "level=error") {
		t.Error("Log injection successful - level=error found in output!")
	}

	// The dangerous content should be neutralized (= becomes _)
	if !strings.Contains(output, "msg_") {
		t.Error("Expected msg= to be sanitized to msg_")
	}

	if !strings.Contains(output, "[REDACTED]") {
		t.Error("Secret value not properly redacted")
	}

	// Should contain escaped/replaced characters, not actual dangerous ones
	if strings.Count(output, "\n") > 1 { // Only final newline should exist
		t.Error("Newline injection not prevented")
	}
}

func TestTextEncoder_KeySanitization(t *testing.T) {
	encoder := NewTextEncoder()

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"safe_key", "user_id", "user_id"},
		{"safe_with_dots", "user.name", "user.name"},
		{"safe_with_dashes", "user-id", "user-id"},
		{"with_spaces", "user name", "user_name"},
		{"with_newlines", "user\nname", "user_name"},
		{"with_equals", "user=admin", "user_admin"},
		{"injection_attempt", "user\" level=\"error", "user__level__error"},
		{"unicode_attack", "user\u202e\u202d", "user__"},
		{"empty_key", "", "invalid_key"},
		{"only_special", "!@#$", "____"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := encoder.sanitizeKey(test.key)
			if result != test.expected {
				t.Errorf("sanitizeKey(%q) = %q, want %q", test.key, result, test.expected)
			}
		})
	}
}

func TestTextEncoder_ValueEscaping(t *testing.T) {
	encoder := NewTextEncoder()

	tests := []struct {
		name  string
		value string
		want  string // What should be in the output
	}{
		{"simple", "hello", `"hello"`},
		{"with_quotes", `hello "world"`, `"hello \"world\""`},
		{"with_newline", "hello\nworld", `"hello_world"`},         // Newlines replaced with underscore
		{"with_tab", "hello\tworld", `"hello_world"`},             // Tabs replaced with underscore
		{"control_chars", "hello\x01\x1fworld", `"hello__world"`}, // Control chars replaced
		{"with_equals", "key=value", `"key_value"`},               // Equals replaced with underscore
		{"injection_attempt", "value\" key2=\"hacked", `"value\" key2_\"hacked"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record := &Record{
				Level:  Info,
				Msg:    "test",
				fields: [32]Field{},
				n:      0,
			}
			record.fields[0] = Str("test_key", test.value)
			record.n = 1

			now := time.Now()
			var buf bytes.Buffer
			encoder.Encode(record, now, &buf)

			output := buf.String()
			if !strings.Contains(output, test.want) {
				t.Errorf("Expected %q in output, got: %s", test.want, output)
			}
		})
	}
}

func TestTextEncoder_InjectionPrevention(t *testing.T) {
	encoder := NewTextEncoder()

	record := &Record{
		Level:  Info,
		Msg:    "Legitimate message",
		fields: [32]Field{},
		n:      0,
	}

	// Field key injection
	record.fields[0] = Str("user\" extra_field=\"injected", "normal_value")

	// Field value injection with newline
	record.fields[1] = Str("data", "normal\nlevel=error msg=\"SYSTEM COMPROMISED\"")

	// Unicode direction override attack
	record.fields[2] = Str("weird\u202e\u202d", "value")
	record.n = 3

	now := time.Now()
	var buf bytes.Buffer
	encoder.Encode(record, now, &buf)

	output := buf.String()
	lines := strings.Split(output, "\n")

	t.Logf("Full output: %s", output)

	// Should only have 2 lines: log entry + final empty line
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d. Lines: %v", len(lines), lines)
	}

	// Should not contain dangerous patterns due to aggressive sanitization
	if strings.Contains(output, "level=error") {
		t.Error("Level injection succeeded - found level=error!")
	}

	// Content should be neutralized (= becomes _)
	if !strings.Contains(output, "msg_") {
		t.Error("Expected msg= to be sanitized to msg_")
	}

	// Should not contain injected field with equals
	if strings.Contains(output, "extra_field=") {
		t.Error("Field key injection succeeded!")
	}
}

func TestTextEncoder_SecretFieldRedaction(t *testing.T) {
	encoder := NewTextEncoder()

	record := &Record{
		Level:  Info,
		Msg:    "User authentication",
		fields: [32]Field{},
		n:      0,
	}

	// Add various secret fields
	record.fields[0] = Secret("password", "supersecret123!")
	record.fields[1] = Secret("api_key", "sk-1234567890abcdef")
	record.fields[2] = Secret("token", "")
	record.fields[3] = Str("username", "john_doe") // Non-secret for comparison
	record.n = 4

	now := time.Now()
	var buf bytes.Buffer
	encoder.Encode(record, now, &buf)

	output := buf.String()
	t.Logf("Output: %s", output)

	// All secret values should be redacted
	secretCount := strings.Count(output, "[REDACTED]")
	if secretCount != 3 {
		t.Errorf("Expected 3 redacted secrets, found %d", secretCount)
	}

	// Secret values should not appear in output
	if strings.Contains(output, "supersecret123!") {
		t.Error("Secret password value leaked!")
	}
	if strings.Contains(output, "sk-1234567890abcdef") {
		t.Error("Secret API key leaked!")
	}

	// Non-secret should still be visible
	if !strings.Contains(output, "john_doe") {
		t.Error("Non-secret value missing!")
	}
}

func TestTextEncoder_StackTraceSafety(t *testing.T) {
	encoder := NewTextEncoder()

	record := &Record{
		Level:  Error,
		Msg:    "Error occurred",
		fields: [32]Field{},
		n:      0,
	}

	// Stack trace with potential injection
	maliciousStack := "main.go:42\nlevel=info msg=\"INJECTED LOG\"\npanic.go:123"
	record.Stack = maliciousStack

	now := time.Now()
	var buf bytes.Buffer
	encoder.Encode(record, now, &buf)

	output := buf.String()
	t.Logf("Output: %s", output)

	// Should contain stack trace section
	if !strings.Contains(output, "stack:") {
		t.Error("Stack trace section missing")
	}

	// Should not allow injection through stack trace
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "level=info") && strings.Contains(line, "INJECTED LOG") {
			t.Error("Stack trace injection succeeded!")
		}
	}

	// Stack lines should be prefixed with spaces
	if !strings.Contains(output, "  main.go") {
		t.Error("Stack trace lines not properly prefixed")
	}
}

func TestTextEncoder_RealisticPerformance(t *testing.T) {
	encoder := NewTextEncoder()

	// Test realistic throughput over time instead of artificial micro-benchmarks
	iterations := 50000
	start := time.Now()
	buf := bytes.NewBuffer(make([]byte, 0, 256))

	for i := 0; i < iterations; i++ {
		// Simulate realistic logging with different timestamps and messages
		now := time.Now() // Real timestamp each time

		record := &Record{
			Level:  Info,
			Msg:    fmt.Sprintf("Request %d processed", i),
			fields: [32]Field{},
			n:      0,
		}

		// Vary field content realistically
		record.fields[0] = Str("user", fmt.Sprintf("user_%d", i%100))
		record.fields[1] = Int64("request_id", int64(i))
		record.fields[2] = Float64("duration", float64(i%1000)/100.0)
		record.n = 3

		buf.Reset()
		encoder.Encode(record, now, buf)
	}

	duration := time.Since(start)
	throughputPerSec := float64(iterations) / duration.Seconds()

	t.Logf("TextEncoder realistic throughput: %.0f ops/sec (%d iterations in %v)",
		throughputPerSec, iterations, duration)

	// Should achieve at least 100k ops/sec in realistic conditions
	if throughputPerSec < 100000 {
		t.Errorf("TextEncoder throughput too low: %.0f ops/sec (expected >100k)", throughputPerSec)
	}
}

func TestTextEncoder_ComplexInjectionScenario(t *testing.T) {
	encoder := NewTextEncoder()

	record := &Record{
		Level:  Warn,
		Msg:    "Security audit",
		fields: [32]Field{},
		n:      0,
	}

	// Complex injection attempt combining multiple techniques
	record.fields[0] = Str("user\n\"level\"=\"error", "admin\nlevel=fatal msg=\"BREACH\"")
	record.fields[1] = Secret("pass\x00word", "secret\n\r\t")
	record.fields[2] = Str("ip", "192.168.1.1\" attacker_ip=\"10.0.0.1")
	record.n = 3

	now := time.Now()
	var buf bytes.Buffer
	encoder.Encode(record, now, &buf)

	output := buf.String()
	t.Logf("Complex injection output: %s", output)

	// Verify no successful injection - aggressive sanitization should prevent this
	if strings.Contains(output, "level=error") || strings.Contains(output, "level=fatal") {
		t.Error("Level injection succeeded!")
	}

	// Content should be neutralized (= becomes _)
	if !strings.Contains(output, "msg_") {
		t.Error("Expected msg= to be sanitized to msg_")
	}

	if strings.Contains(output, "attacker_ip=") {
		t.Error("Field injection succeeded!")
	}

	// Should contain redacted secret
	if !strings.Contains(output, "[REDACTED]") {
		t.Error("Secret not redacted in complex scenario")
	}
}

// TestWriteSafeValue tests the writeSafeValue method with various character scenarios
func TestWriteSafeValue(t *testing.T) {
	encoder := NewTextEncoder()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal_Text",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "Text_With_Spaces",
			input:    "hello world",
			expected: "hello_world",
		},
		{
			name:     "Text_With_Newlines",
			input:    "hello\nworld",
			expected: "hello_world",
		},
		{
			name:     "Text_With_CarriageReturn",
			input:    "hello\rworld",
			expected: "hello_world",
		},
		{
			name:     "Text_With_Tabs",
			input:    "hello\tworld",
			expected: "hello_world",
		},
		{
			name:     "Text_With_Control_Characters",
			input:    "hello\x01\x02world",
			expected: "hello__world",
		},
		{
			name:     "Text_With_DEL_Character",
			input:    "hello\x7Fworld",
			expected: "hello_world",
		},
		{
			name:     "Mixed_Special_Characters",
			input:    "hello\n\r\t world\x01\x7F",
			expected: "hello____world__",
		},
		{
			name:     "Empty_String",
			input:    "",
			expected: "",
		},
		{
			name:     "Only_Special_Characters",
			input:    "\n\r\t \x01\x7F",
			expected: "______",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder.writeSafeValue(tt.input, &buf)

			result := buf.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
