// encoder_console_fuzz_test.go
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

// FuzzConsoleEncode verifies that ConsoleEncoder never panics and always
// produces exactly one newline-terminated line, regardless of message or
// field value content. This covers the parse→escape→write pipeline end-to-end,
// exercising ANSI injection, log injection, and control character paths that
// security tests cover with known inputs.
func FuzzConsoleEncode(f *testing.F) {
	f.Add("normal message", "normal value")
	f.Add("", "")
	f.Add("\033[31mred\033[0m", "injected")
	f.Add("line1\nline2", "field\nvalue")
	f.Add("msg\r\nforged", "val\r\n")
	f.Add("tab\there", "val\ttab")
	f.Add("null\x00byte", "val\x00")
	f.Add(strings.Repeat("A", 1024), strings.Repeat("B", 1024))

	fuzzTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	f.Fuzz(func(t *testing.T, msg, fieldVal string) {
		t.Parallel()

		enc := NewConsoleEncoderWithColorize(false)
		var buf bytes.Buffer
		rec := NewRecord(Info, msg)
		rec.AddField(Str("k", fieldVal))

		// Must not panic.
		enc.Encode(rec, fuzzTime, &buf)

		out := buf.String()

		// Must end with exactly one newline.
		if !strings.HasSuffix(out, "\n") {
			t.Fatalf("output does not end with newline: %q", out)
		}
		if strings.Count(out, "\n") != 1 {
			t.Fatalf("output has %d newlines (expected 1): %q",
				strings.Count(out, "\n"), out)
		}

		// No raw ANSI codes in plain mode.
		if strings.Contains(out, "\033") {
			t.Fatalf("ANSI escape in plain output: %q", out)
		}
	})
}
