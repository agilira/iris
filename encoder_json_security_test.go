// encoder_json_security_test.go: Security test suite for JSON encoder
//
// THREAT MODEL:
//   CWE-116: Improper Encoding or Escaping of Output
//     - Null bytes in field values can truncate JSON output
//     - Newline injection can split NDJSON records
//     - Backslash sequences can break JSON structure
//   CWE-20: Improper Input Validation
//     - Overlong UTF-8 sequences can bypass filtering
//     - Empty/whitespace-only keys can produce invalid JSON
//     - Extremely long strings can cause resource exhaustion
//   CWE-200: Exposure of Sensitive Information
//     - Secret() fields must NEVER leak plaintext in any code path
//     - Field key must not reveal secret value via injection
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// --- Attack vector: CWE-116 (Encoding/Escaping) ---

func TestSecurity_JSONEncoder_NullByteInFieldValue(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: Null bytes in C-based log parsers can truncate output,
	// hiding subsequent fields (including security-relevant ones).
	// MITIGATION EXPECTED: null bytes escaped as \u0000 in JSON output.
	enc := NewJSONEncoder()
	var buf bytes.Buffer
	rec := NewRecord(Info, "test")
	rec.AddField(Str("payload", "before\x00after"))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	enc.Encode(rec, now, &buf)
	output := buf.String()

	// Raw null byte must NOT appear in output
	if strings.ContainsRune(output, '\x00') {
		t.Fatal("raw null byte found in JSON output -- CWE-116 vulnerability")
	}

	// The escaped form must appear
	if !strings.Contains(output, `\u0000`) {
		t.Errorf("null byte not escaped to \\u0000, got: %s", output)
	}

	// Output must be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", err, output)
	}
}

func TestSecurity_JSONEncoder_NewlineInjection(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: Injected newlines in NDJSON split one record into multiple,
	// allowing an attacker to forge log entries.
	// MITIGATION EXPECTED: \n escaped as \\n inside JSON string values.
	enc := NewJSONEncoder()
	var buf bytes.Buffer
	rec := NewRecord(Info, "legit message")
	// Try to inject a fake second JSON record via the message field
	rec.Msg = "legit\n{\"level\":\"error\",\"msg\":\"FORGED\"}"
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	enc.Encode(rec, now, &buf)
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Must be exactly ONE NDJSON line
	if len(lines) != 1 {
		t.Fatalf("newline injection produced %d lines (expected 1):\n%s", len(lines), output)
	}

	// Must be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestSecurity_JSONEncoder_BackslashSequences(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: Unescaped backslashes can break JSON structure.
	// MITIGATION EXPECTED: backslashes doubled in output.
	enc := NewJSONEncoder()

	cases := []struct {
		name  string
		input string
	}{
		{"trailing_backslash", `value\`},
		{"double_backslash", `val\\ue`},
		{"backslash_quote", `val\"ue`},
		{"backslash_n_literal", `val\nue`},
		{"mixed", "tab\there\nnewline\\backslash\"quote"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			rec := NewRecord(Info, "test")
			rec.AddField(Str("data", tc.input))
			now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

			enc.Encode(rec, now, &buf)
			output := strings.TrimSpace(buf.String())

			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(output), &parsed); err != nil {
				t.Fatalf("invalid JSON for input %q: %v\nraw: %s", tc.input, err, output)
			}
		})
	}
}

func TestSecurity_JSONEncoder_ControlCharacters(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: Control characters (0x00-0x1F, 0x7F) can manipulate terminal
	// displays, corrupt log files, or bypass parsers.
	// MITIGATION EXPECTED: all control chars escaped.
	enc := NewJSONEncoder()

	// Build a string with all control characters
	var controlChars strings.Builder
	for i := 0; i < 0x20; i++ {
		controlChars.WriteByte(byte(i))
	}
	controlChars.WriteByte(0x7F) // DEL

	var buf bytes.Buffer
	rec := NewRecord(Info, "test")
	rec.AddField(Str("evil", controlChars.String()))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	enc.Encode(rec, now, &buf)
	output := strings.TrimSpace(buf.String())

	// Output must be valid JSON (json.Unmarshal rejects raw control chars)
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON with control chars: %v\nraw: %s", err, output)
	}
}

// --- Attack vector: CWE-200 (Secret field leak) ---

func TestSecurity_JSONEncoder_SecretNeverLeaks(t *testing.T) {
	// ATTACK VECTOR: CWE-200
	// IMPACT: If Secret() field leaks plaintext, passwords/API keys exposed in logs.
	// MITIGATION EXPECTED: Secret fields always produce "[REDACTED]", never the value.
	enc := NewJSONEncoder()

	secrets := []struct {
		key   string
		value string
	}{
		{"password", "P@ssw0rd!123"},
		{"api_key", "sk-live-4e2f8a9c1d3b5e7f"},
		{"token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
		{"credit_card", "4111-1111-1111-1111"},
		{"empty_secret", ""},
		// WHY null bytes in secret: attacker might try to truncate redaction
		{"null_secret", "before\x00after"},
	}

	for _, s := range secrets {
		t.Run(s.key, func(t *testing.T) {
			var buf bytes.Buffer
			rec := NewRecord(Info, "audit")
			rec.AddField(Secret(s.key, s.value))
			now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

			enc.Encode(rec, now, &buf)
			output := buf.String()

			// The actual secret value must NEVER appear
			if s.value != "" && strings.Contains(output, s.value) {
				t.Fatalf("SECRET LEAKED in JSON output for key %q: %s", s.key, output)
			}

			// The redaction marker must appear
			if !strings.Contains(output, `"[REDACTED]"`) {
				t.Errorf("redaction marker not found for key %q: %s", s.key, output)
			}
		})
	}
}

func TestSecurity_JSONEncoder_SecretKeyInjection(t *testing.T) {
	// ATTACK VECTOR: CWE-116 + CWE-200
	// IMPACT: A malicious field key like `key":"leaked","real_key` could
	// break JSON structure and leak the secret value as a separate field.
	// MITIGATION EXPECTED: key is properly quoted, no JSON breakout.
	enc := NewJSONEncoder()
	var buf bytes.Buffer

	maliciousKey := `password":"leaked","real`
	rec := NewRecord(Info, "test")
	rec.AddField(Secret(maliciousKey, "actual_secret"))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	enc.Encode(rec, now, &buf)
	output := strings.TrimSpace(buf.String())

	// Secret value must not appear
	if strings.Contains(output, "actual_secret") {
		t.Fatal("SECRET LEAKED via key injection")
	}

	// "leaked" should not appear as a standalone JSON value
	if strings.Contains(output, `"leaked"`) {
		t.Error("key injection created unauthorized JSON field")
	}
}

// --- Attack vector: CWE-20 (Input validation) ---

func TestSecurity_JSONEncoder_EmptyAndExtremeInputs(t *testing.T) {
	// ATTACK VECTOR: CWE-20
	// IMPACT: Empty keys, empty values, or extremely long strings
	// could cause panic or corrupt output.
	// MITIGATION EXPECTED: graceful handling, valid JSON output.
	enc := NewJSONEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		key  string
		val  string
	}{
		{"empty_key", "", "value"},
		{"empty_value", "key", ""},
		{"both_empty", "", ""},
		{"spaces_key", "   ", "value"},
		{"long_value", "key", strings.Repeat("A", 10000)},
		{"long_key", strings.Repeat("k", 10000), "value"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			rec := NewRecord(Info, "test")
			rec.AddField(Str(tc.key, tc.val))

			// Must not panic
			enc.Encode(rec, now, &buf)

			// Output must be non-empty
			if buf.Len() == 0 {
				t.Fatal("encoder produced empty output")
			}
		})
	}
}

// --- Fuzz target: FuzzJSONEncoder ---

func FuzzJSONEncoder(f *testing.F) {
	// Seeds: real attack patterns, not random noise
	seeds := []struct {
		msg   string
		key   string
		value string
	}{
		// Null byte injection
		{"msg", "key", "before\x00after"},
		// Newline injection (NDJSON splitting)
		{"msg", "key", "line1\nline2"},
		// JSON breakout via value
		{"msg", "key", `","evil":"injected`},
		// Backslash attacks
		{"msg", "key", `val\`},
		{"msg", "key", `val\\`},
		{"msg", "key", `val\"`},
		// Control characters
		{"msg", "key", "\x01\x02\x03\x1b[31m"},
		// Unicode direction overrides (CWE-1007)
		{"msg", "key", "normal\u202eesrever"},
		// Overlong UTF-8 (invalid sequences)
		{"msg", "key", "\xc0\xaf\xe0\x80\xaf"},
		// Format string patterns
		{"msg", "key", "%s%s%s%n%n%n"},
		// Empty edge cases
		{"", "", ""},
		// Very long strings
		{"msg", "k", strings.Repeat("A", 8192)},
	}

	for _, s := range seeds {
		f.Add(s.msg, s.key, s.value)
	}

	enc := NewJSONEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	f.Fuzz(func(t *testing.T, msg, key, value string) {
		var buf bytes.Buffer
		rec := NewRecord(Info, msg)
		rec.AddField(Str(key, value))

		// Must not panic
		enc.Encode(rec, now, &buf)
		output := strings.TrimSpace(buf.String())

		if len(output) == 0 {
			t.Fatal("encoder produced empty output")
		}

		// Output must be valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(output), &parsed); err != nil {
			t.Fatalf("invalid JSON: %v\nmsg=%q key=%q value=%q\nraw=%s",
				err, msg, key, value, output)
		}

		// Raw null bytes must never appear
		if strings.ContainsRune(output, '\x00') {
			t.Fatal("raw null byte in JSON output")
		}
	})
}

// --- Fuzz target: FuzzSecretField ---

func FuzzSecretField(f *testing.F) {
	// Seeds: attempt to defeat redaction via special characters
	seeds := []string{
		"simple_secret",
		"",
		"secret\x00with_null",
		"secret\nwith_newline",
		`secret","leaked":"true`,
		"secret\\\"escaped",
		strings.Repeat("x", 8192),
		"\u202ereverse",
		"%s%n%x",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	enc := NewJSONEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	f.Fuzz(func(t *testing.T, secretValue string) {
		var buf bytes.Buffer
		rec := NewRecord(Info, "audit")
		rec.AddField(Secret("secret_field", secretValue))

		enc.Encode(rec, now, &buf)
		output := buf.String()

		// The redaction marker must always appear
		if !strings.Contains(output, `"[REDACTED]"`) {
			t.Fatalf("redaction marker missing for value %q: %s", secretValue, output)
		}

		// The secret value must NEVER appear (if non-empty and not a
		// substring of the redaction marker itself)
		if secretValue != "" &&
			secretValue != "[REDACTED]" &&
			!strings.Contains("[REDACTED]", secretValue) &&
			strings.Contains(output, secretValue) {
			t.Fatalf("SECRET LEAKED: %q found in output: %s", secretValue, output)
		}
	})
}
