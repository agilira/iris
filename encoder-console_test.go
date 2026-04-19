// encoder-console_test.go
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

var consoleTestTime = time.Date(2026, 4, 19, 15, 4, 5, 123000000, time.UTC)

func TestNewConsoleEncoder_Defaults(t *testing.T) {
	t.Parallel()
	enc := NewConsoleEncoder()
	// colorize depends on the test environment TTY; just verify no panic.
	_ = enc.colorize
}

func TestNewConsoleEncoderWithColorize(t *testing.T) {
	t.Parallel()
	plain := NewConsoleEncoderWithColorize(false)
	colored := NewConsoleEncoderWithColorize(true)
	if plain.colorize {
		t.Error("expected colorize=false")
	}
	if !colored.colorize {
		t.Error("expected colorize=true")
	}
}

func TestConsoleEncoder_TimestampFormat(t *testing.T) {
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "msg")
	enc.Encode(rec, consoleTestTime, &buf)
	if !strings.HasPrefix(buf.String(), "15:04:05.123") {
		t.Errorf("unexpected timestamp prefix: %s", buf.String())
	}
}

func TestConsoleEncoder_LevelLabels(t *testing.T) {
	t.Parallel()
	cases := []struct {
		level Level
		label string
	}{
		{Debug, "DEBUG"},
		{Info, "INFO "},
		{Warn, "WARN "},
		{Error, "ERROR"},
		{Fatal, "FATAL"},
	}
	for _, tc := range cases {
		enc := NewConsoleEncoderWithColorize(false)
		var buf bytes.Buffer
		rec := NewRecord(tc.level, "")
		enc.Encode(rec, consoleTestTime, &buf)
		if !strings.Contains(buf.String(), tc.label) {
			t.Errorf("level %v: expected label %q in %q", tc.level, tc.label, buf.String())
		}
	}
}

func TestConsoleEncoder_ColorizedLevel(t *testing.T) {
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(true)
	var buf bytes.Buffer
	rec := NewRecord(Info, "hello")
	enc.Encode(rec, consoleTestTime, &buf)
	out := buf.String()
	if !strings.Contains(out, ansiCyan) {
		t.Errorf("expected cyan ANSI code in colorized output: %q", out)
	}
	if !strings.Contains(out, ansiReset) {
		t.Errorf("expected reset ANSI code in colorized output: %q", out)
	}
}

func TestConsoleEncoder_PlainNoANSI(t *testing.T) {
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Error, "boom")
	enc.Encode(rec, consoleTestTime, &buf)
	out := buf.String()
	if strings.Contains(out, "\033") {
		t.Errorf("plain encoder must not emit ANSI codes: %q", out)
	}
}

func TestConsoleEncoder_LoggerAndMsg(t *testing.T) {
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "daemon ready")
	rec.Logger = "mts"
	enc.Encode(rec, consoleTestTime, &buf)
	out := buf.String()
	if !strings.Contains(out, "mts") {
		t.Errorf("expected logger name in output: %q", out)
	}
	if !strings.Contains(out, "daemon ready") {
		t.Errorf("expected message in output: %q", out)
	}
}

func TestConsoleEncoder_FieldTypes(t *testing.T) {
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "all fields")
	rec.AddField(String("str", "hello"))
	rec.AddField(Int("count", 42))
	rec.AddField(Bool("ok", true))
	rec.AddField(Float64("temp", 0.7))
	rec.AddField(Dur("elapsed", 150*time.Millisecond))
	rec.AddField(Secret("token", "s3cr3t"))
	enc.Encode(rec, consoleTestTime, &buf)
	out := buf.String()

	checks := []string{"str=hello", "count=42", "ok=true", "temp=0.7", "token=[REDACTED]"}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output: %q", want, out)
		}
	}
	if strings.Contains(out, "s3cr3t") {
		t.Error("secret value must not appear in output")
	}
}

func TestConsoleEncoder_QuotesStringWithSpaces(t *testing.T) {
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "")
	rec.AddField(String("msg", "hello world"))
	enc.Encode(rec, consoleTestTime, &buf)
	if !strings.Contains(buf.String(), `msg="hello world"`) {
		t.Errorf("expected quoted value: %q", buf.String())
	}
}

func TestConsoleEncoder_EmptyStringQuoted(t *testing.T) {
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "")
	rec.AddField(String("key", ""))
	enc.Encode(rec, consoleTestTime, &buf)
	if !strings.Contains(buf.String(), `key=""`) {
		t.Errorf("expected empty quoted value: %q", buf.String())
	}
}

func TestConsoleEncoder_SingleTrailingNewline(t *testing.T) {
	t.Parallel()
	enc := NewConsoleEncoderWithColorize(false)
	var buf bytes.Buffer
	rec := NewRecord(Info, "test")
	enc.Encode(rec, consoleTestTime, &buf)
	out := buf.String()
	if !strings.HasSuffix(out, "\n") {
		t.Error("output must end with newline")
	}
	if strings.Count(out, "\n") != 1 {
		t.Errorf("output must have exactly one newline, got: %q", out)
	}
}

func TestIsTTY_RegularFileIsNotTTY(t *testing.T) {
	t.Parallel()
	// WHY regular file: /dev/null is a character device (ModeCharDevice=true).
	// A regular temp file has no ModeCharDevice — the correct non-TTY test case.
	f, err := os.CreateTemp(t.TempDir(), "isTTY-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if isTTY(f) {
		t.Error("expected isTTY=false for a regular file")
	}
}
