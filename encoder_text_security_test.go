// encoder_text_security_test.go: Security test suite for Text encoder
//
// THREAT MODEL:
//   CWE-93: Improper Neutralization of CRLF Sequences in HTTP Headers (Log Injection)
//     - Newlines in field values or keys can forge extra log lines
//     - Carriage returns can overwrite visible output in terminals
//   CWE-116: Improper Encoding or Escaping of Output
//     - Control characters can manipulate terminal displays (ANSI escape)
//     - Tab injection can misalign structured log output
//   CWE-20: Improper Input Validation
//     - Field key bypass when SanitizeKeys=false
//     - Empty/whitespace keys produce ambiguous entries
//   CWE-200: Exposure of Sensitive Information
//     - Secret fields must always produce [REDACTED] in text output
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

// --- Attack vector: CWE-93 (Log Injection via CRLF) ---

func TestSecurity_TextEncoder_NewlineInjection(t *testing.T) {
	// ATTACK VECTOR: CWE-93
	// IMPACT: Injected newlines in text log entries forge additional lines,
	// allowing log tampering, SIEM evasion, or false alerting.
	// MITIGATION EXPECTED: newlines neutralized (replaced with _).
	enc := NewTextEncoder()
	var buf bytes.Buffer
	rec := NewRecord(Info, "legit")
	rec.AddField(Str("data", "line1\nINFO forged log entry"))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	enc.Encode(rec, now, &buf)
	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	// Must produce exactly ONE log line
	if len(lines) != 1 {
		t.Fatalf("newline injection produced %d lines (expected 1):\n%s", len(lines), output)
	}
}

func TestSecurity_TextEncoder_CRLFInjection(t *testing.T) {
	// ATTACK VECTOR: CWE-93
	// IMPACT: \r can overwrite visible portions of log line in terminals,
	// making a malicious entry look legitimate.
	// MITIGATION EXPECTED: \r neutralized.
	enc := NewTextEncoder()
	var buf bytes.Buffer
	rec := NewRecord(Info, "test")
	rec.AddField(Str("payload", "visible\r\nFORGED [ERROR] system breach"))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	enc.Encode(rec, now, &buf)
	output := buf.String()

	// Raw \r must not appear in output
	if strings.ContainsRune(output, '\r') {
		t.Fatal("raw carriage return found in text output -- CWE-93 vulnerability")
	}

	// Must still be a single line (trailing \n from encoder is OK)
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("CRLF injection produced %d lines:\n%s", len(lines), output)
	}
}

// --- Attack vector: CWE-116 (Control char / ANSI escape) ---

func TestSecurity_TextEncoder_ANSIEscapeInjection(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: ANSI escape sequences (e.g., \x1b[31m) can change terminal
	// colors, clear screen, or reposition cursor to hide malicious output.
	// MITIGATION EXPECTED: control chars including ESC neutralized.
	enc := NewTextEncoder()
	var buf bytes.Buffer
	rec := NewRecord(Info, "test")
	rec.AddField(Str("evil", "\x1b[31mRED_TEXT\x1b[0m"))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	enc.Encode(rec, now, &buf)
	output := buf.String()

	// ESC character (0x1b) must not appear
	if strings.ContainsRune(output, '\x1b') {
		t.Fatal("ANSI escape sequence not neutralized -- CWE-116 vulnerability")
	}
}

func TestSecurity_TextEncoder_ControlCharacters(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: Control characters 0x00-0x1F can corrupt log files,
	// confuse parsers, or manipulate terminal displays.
	// MITIGATION EXPECTED: all control chars replaced with _.
	enc := NewTextEncoder()
	var buf bytes.Buffer

	var controlChars strings.Builder
	for i := 0; i < 0x20; i++ {
		controlChars.WriteByte(byte(i))
	}
	controlChars.WriteByte(0x7F) // DEL

	rec := NewRecord(Info, "test")
	rec.AddField(Str("chars", controlChars.String()))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	enc.Encode(rec, now, &buf)
	output := buf.String()

	// No control characters should survive (except final \n from encoder)
	for i, c := range output {
		if c < 0x20 && c != '\n' {
			t.Errorf("control char 0x%02x at position %d not neutralized", c, i)
		}
		if c == 0x7F {
			t.Errorf("DEL char (0x7F) at position %d not neutralized", i)
		}
	}
}

func TestSecurity_TextEncoder_TabInjection(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: Tab characters in structured text logs can misalign columns,
	// causing field values to appear under wrong headers in SIEM parsers.
	// MITIGATION EXPECTED: tabs neutralized.
	enc := NewTextEncoder()
	var buf bytes.Buffer
	rec := NewRecord(Info, "test")
	rec.AddField(Str("data", "col1\tcol2\tcol3"))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	enc.Encode(rec, now, &buf)
	output := buf.String()

	if strings.ContainsRune(output, '\t') {
		t.Fatal("raw tab character found in text output -- CWE-116")
	}
}

// --- Attack vector: CWE-20 (Key sanitization) ---

func TestSecurity_TextEncoder_KeySanitization(t *testing.T) {
	// ATTACK VECTOR: CWE-20
	// IMPACT: Unsanitized keys with special characters (=, ", spaces)
	// can break key=value parsing in downstream systems.
	// MITIGATION EXPECTED: sanitizeKey replaces dangerous chars with _.
	enc := NewTextEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name      string
		key       string
		forbidden []string
	}{
		{"equals_in_key", "key=val", []string{"="}},
		{"quotes_in_key", `key"val`, []string{`"`}},
		{"spaces_in_key", "key val", []string{" "}},
		{"newline_in_key", "key\nval", []string{"\n"}},
		{"unicode_in_key", "key\u202eval", []string{"\u202e"}},
		{"null_in_key", "key\x00val", []string{"\x00"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			rec := NewRecord(Info, "test")
			rec.AddField(Str(tc.key, "value"))

			enc.Encode(rec, now, &buf)
			output := buf.String()

			for _, forbidden := range tc.forbidden {
				// The forbidden character should not appear in the key portion
				// of the output (before the = sign in key=value)
				eqIdx := strings.Index(output, "=")
				if eqIdx > 0 {
					keyPortion := output[:eqIdx]
					if strings.Contains(keyPortion, forbidden) {
						t.Errorf("forbidden char %q found in key portion: %s",
							forbidden, keyPortion)
					}
				}
			}
		})
	}
}

func TestSecurity_TextEncoder_EmptyAndExtremeInputs(t *testing.T) {
	// ATTACK VECTOR: CWE-20
	// IMPACT: Empty or extremely long inputs could cause panic or OOM.
	// MITIGATION EXPECTED: graceful handling without panic.
	enc := NewTextEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		key  string
		val  string
	}{
		{"empty_key", "", "value"},
		{"empty_value", "key", ""},
		{"both_empty", "", ""},
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

			if buf.Len() == 0 {
				t.Fatal("encoder produced empty output")
			}
		})
	}
}

// --- Attack vector: CWE-200 (Secret leak) ---

func TestSecurity_TextEncoder_SecretNeverLeaks(t *testing.T) {
	// ATTACK VECTOR: CWE-200
	// IMPACT: If Secret() field leaks plaintext in text encoder,
	// credentials exposed in logs.
	// MITIGATION EXPECTED: Secret fields produce [REDACTED].
	enc := NewTextEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	secrets := []struct {
		key   string
		value string
	}{
		{"password", "P@ssw0rd!123"},
		{"api_key", "sk-live-4e2f8a9c1d3b5e7f"},
		{"token", "eyJhbGciOiJIUzI1NiJ9.secret_payload"},
		{"empty_secret", ""},
		{"null_secret", "before\x00after"},
		{"newline_secret", "line1\nline2"},
	}

	for _, s := range secrets {
		t.Run(s.key, func(t *testing.T) {
			var buf bytes.Buffer
			rec := NewRecord(Info, "audit")
			rec.AddField(Secret(s.key, s.value))

			enc.Encode(rec, now, &buf)
			output := buf.String()

			// Secret value must NEVER appear
			if s.value != "" && strings.Contains(output, s.value) {
				t.Fatalf("SECRET LEAKED in text output for key %q: %s", s.key, output)
			}

			// Redaction marker must appear
			if !strings.Contains(output, "[REDACTED]") {
				t.Errorf("redaction marker not found for key %q: %s", s.key, output)
			}
		})
	}
}

// --- Fuzz target: FuzzTextEncoder ---

func FuzzTextEncoder(f *testing.F) {
	// Seeds: real attack patterns targeting text encoder specifically
	seeds := []struct {
		msg   string
		key   string
		value string
	}{
		// Log injection via newline
		{"msg", "key", "value\nINFO forged entry"},
		// CRLF injection
		{"msg", "key", "value\r\nFORGED"},
		// ANSI escape
		{"msg", "key", "\x1b[31mRED\x1b[0m"},
		// Tab misalignment
		{"msg", "key", "col1\tcol2"},
		// Null byte
		{"msg", "key", "before\x00after"},
		// Key=value breakout
		{"msg", "key=injected", "value"},
		// Unicode direction override
		{"msg", "key", "\u202eesrever"},
		// Format string patterns
		{"msg", "key", "%s%n%x"},
		// Empty edge cases
		{"", "", ""},
		// Long strings
		{"msg", "k", strings.Repeat("X", 8192)},
		// Backslash
		{"msg", "key", `val\ue`},
		// Equals in value
		{"msg", "key", "a=b=c"},
	}

	for _, s := range seeds {
		f.Add(s.msg, s.key, s.value)
	}

	enc := NewTextEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	f.Fuzz(func(t *testing.T, msg, key, value string) {
		var buf bytes.Buffer
		rec := NewRecord(Info, msg)
		rec.AddField(Str(key, value))

		// Must not panic
		enc.Encode(rec, now, &buf)
		output := buf.String()

		if len(output) == 0 {
			t.Fatal("encoder produced empty output")
		}

		// No raw control characters allowed (except final \n)
		trimmed := strings.TrimRight(output, "\n")
		for i, c := range trimmed {
			if c < 0x20 {
				t.Fatalf("control char 0x%02x at position %d in output: %q",
					c, i, trimmed)
			}
		}
	})
}
