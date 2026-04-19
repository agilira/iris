// encoder_console_security_test.go: Security test suite for ConsoleEncoder
//
// THREAT MODEL:
//   CWE-116: Improper Encoding or Escaping of Output
//     - ESC (0x1B) in user content starts ANSI sequences that can change
//       terminal colors, move cursors, or execute terminal commands, making
//       a malicious log line appear as a different level or be invisible.
//   CWE-93: Improper Neutralization of CRLF Sequences (Log Injection)
//     - Newlines in message or field values forge extra log lines,
//       enabling log tampering and SIEM/alerting evasion.
//   CWE-20: Improper Input Validation
//     - Control characters (tab, backspace, null) can misalign log output
//       or truncate strings at the OS boundary.
//   CWE-200: Exposure of Sensitive Information
//     - Secret fields must always be redacted regardless of encoder.
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

var consoleSecTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// --- CWE-116: ANSI injection ---

func TestSecurity_ConsoleEncoder_ANSIInjectionInMsg(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: An attacker injects ESC sequences into a log message via user input
	// (e.g. a Telegram message). The terminal renders the sequence, changing the
	// displayed log level color or erasing previous output, making attacks
	// look like innocent INFO entries.
	// MITIGATION EXPECTED: ESC (0x1B < 0x20) replaced with '_'.
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "\033[31mERROR\033[0m injected")
	enc.Encode(rec, consoleSecTime, &buf)
	out := buf.String()
	if strings.Contains(out, "\033") {
		t.Fatalf("ANSI injection in message not neutralized: %q", out)
	}
}

func TestSecurity_ConsoleEncoder_ANSIInjectionInFieldValue(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: Field value contains ANSI codes to reset color after a WARN/ERROR
	// entry, making subsequent INFO lines appear as errors in the terminal.
	// MITIGATION EXPECTED: ESC stripped from field values.
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Warn, "suspicious")
	rec.AddField(Str("user_input", "\033[0mINFO legit looking line"))
	enc.Encode(rec, consoleSecTime, &buf)
	out := buf.String()
	if strings.Contains(out, "\033") {
		t.Fatalf("ANSI injection in field value not neutralized: %q", out)
	}
}

func TestSecurity_ConsoleEncoder_ANSIInjectionInFieldKey(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: A malformed field key containing ESC resets color mid-line.
	// MITIGATION EXPECTED: consoleEscape applied to field keys.
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "test")
	rec.AddField(Str("\033[31mbadkey\033[0m", "val"))
	enc.Encode(rec, consoleSecTime, &buf)
	out := buf.String()
	if strings.Contains(out, "\033") {
		t.Fatalf("ANSI injection in field key not neutralized: %q", out)
	}
}

func TestSecurity_ConsoleEncoder_ANSIColorizedOutputStillSafe(t *testing.T) {
	// ATTACK VECTOR: CWE-116
	// IMPACT: Verify that enabling colorize=true (encoder's own ANSI codes) does
	// NOT allow user-injected ANSI to pass through alongside the encoder's codes.
	// MITIGATION EXPECTED: user content sanitized; only encoder-emitted codes present.
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(true)
	var buf bytes.Buffer
	rec := NewRecord(Info, "\033[31mfake error\033[0m")
	enc.Encode(rec, consoleSecTime, &buf)
	out := buf.String()

	// Strip known encoder ANSI codes, then verify no ESC remains.
	stripped := strings.ReplaceAll(out, ansiCyan, "")
	stripped = strings.ReplaceAll(stripped, ansiReset, "")
	if strings.Contains(stripped, "\033") {
		t.Fatalf("user ESC survived after stripping encoder codes: %q", out)
	}
}

// --- CWE-93: Log injection via newlines ---

func TestSecurity_ConsoleEncoder_NewlineInjectionInMsg(t *testing.T) {
	// ATTACK VECTOR: CWE-93
	// IMPACT: A newline in the message forges an extra log line, allowing
	// an attacker to inject a fake ERROR or FATAL entry into the log stream.
	// MITIGATION EXPECTED: exactly one line produced per Encode call.
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "legit\nFATAL forged entry")
	enc.Encode(rec, consoleSecTime, &buf)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("newline injection produced %d lines (expected 1):\n%s", len(lines), buf.String())
	}
}

func TestSecurity_ConsoleEncoder_NewlineInjectionInFieldValue(t *testing.T) {
	// ATTACK VECTOR: CWE-93
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "ok")
	rec.AddField(Str("data", "normal\nERROR injected line"))
	enc.Encode(rec, consoleSecTime, &buf)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("newline injection in field produced %d lines:\n%s", len(lines), buf.String())
	}
}

func TestSecurity_ConsoleEncoder_CRLFInjection(t *testing.T) {
	// ATTACK VECTOR: CWE-93
	// IMPACT: \r moves the cursor to the start of the line; a following write
	// overwrites the visible content, hiding the real log entry.
	// MITIGATION EXPECTED: \r neutralized.
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "real\r\nFAKE")
	enc.Encode(rec, consoleSecTime, &buf)
	out := buf.String()
	if strings.Contains(out, "\r") {
		t.Fatalf("CR not neutralized: %q", out)
	}
}

// --- CWE-20: Control character neutralization ---

func TestSecurity_ConsoleEncoder_ControlCharsNeutralized(t *testing.T) {
	// ATTACK VECTOR: CWE-20
	// IMPACT: Tab, backspace, null byte can misalign or truncate log output.
	// MITIGATION EXPECTED: all chars < 0x20 and 0x7F replaced with '_'.
	t.Parallel()
	dangerous := []struct {
		name  string
		input string
	}{
		{"tab", "val\t"},
		{"backspace", "val\x08"},
		{"null", "val\x00"},
		{"bell", "val\x07"},
		{"del", "val\x7F"},
	}
	enc := NewConsoleEncoderWithColorize(false)
	for _, tc := range dangerous {
		var buf bytes.Buffer
		rec := NewRecord(Info, tc.input)
		enc.Encode(rec, consoleSecTime, &buf)
		// Trim the encoder's own trailing newline before checking for control chars.
		out := strings.TrimSuffix(buf.String(), "\n")
		for _, c := range out {
			if c < 0x20 || c == 0x7F {
				t.Errorf("%s: control char 0x%02X not neutralized in %q", tc.name, c, out)
			}
		}
	}
}

// --- CWE-200: Secret redaction ---

func TestSecurity_ConsoleEncoder_SecretRedacted(t *testing.T) {
	// ATTACK VECTOR: CWE-200
	// IMPACT: Sensitive values (API keys, tokens) must never appear in logs.
	// MITIGATION EXPECTED: Secret fields always emit [REDACTED].
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "auth")
	rec.AddField(Secret("api_key", "sk-abc123secret"))
	enc.Encode(rec, consoleSecTime, &buf)
	out := buf.String()
	if strings.Contains(out, "sk-abc123secret") {
		t.Fatalf("secret value leaked in console output: %q", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Fatalf("expected [REDACTED] marker in output: %q", out)
	}
}
