// encoder-cnsl.go: Human-readable console encoder for Iris logging records
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"strconv"
	"strings"
	"time"
)

// ConsoleEncoder implements an encoder for human-readable console output.
// It provides configurable time formatting, level casing, and optional ANSI color support
// for development and debugging environments.
type ConsoleEncoder struct {
	// TimeFormat specifies the time layout for timestamps (default: time.RFC3339Nano)
	TimeFormat string

	// LevelCasing controls level text casing: "upper" (default) or "lower"
	LevelCasing string

	// EnableColor enables ANSI color codes for different log levels (default: false)
	EnableColor bool
}

// NewConsoleEncoder creates a new console encoder with default settings.
// By default, it uses RFC3339Nano time format, uppercase level casing, and no colors.
func NewConsoleEncoder() *ConsoleEncoder {
	return &ConsoleEncoder{
		TimeFormat:  time.RFC3339Nano,
		LevelCasing: "upper",
		EnableColor: false,
	}
}

// NewColorConsoleEncoder creates a new console encoder with ANSI colors enabled.
// This is useful for development environments where color output enhances readability.
func NewColorConsoleEncoder() *ConsoleEncoder {
	return &ConsoleEncoder{
		TimeFormat:  time.RFC3339Nano,
		LevelCasing: "upper",
		EnableColor: true,
	}
}

// Encode writes a log record to the buffer in console-friendly format.
// The output format is: timestamp level message key=value key=value...
func (e *ConsoleEncoder) Encode(rec *Record, now time.Time, buf *bytes.Buffer) {
	// Set defaults
	timeFormat := e.TimeFormat
	if timeFormat == "" {
		timeFormat = time.RFC3339Nano
	}

	// Format level with casing preference
	levelStr := rec.Level.String()
	if e.LevelCasing == "" || strings.EqualFold(e.LevelCasing, "upper") {
		levelStr = strings.ToUpper(levelStr)
	}

	// Apply color if enabled
	if e.EnableColor {
		levelStr = colorizeLevel(rec.Level, levelStr)
	}

	// Pre-allocate buffer space for better performance
	buf.Grow(128)

	// Write timestamp
	buf.WriteString(now.Format(timeFormat))
	buf.WriteByte(' ')

	// Write level
	buf.WriteString(levelStr)

	// Write message if present
	if rec.Msg != "" {
		buf.WriteByte(' ')
		buf.WriteString(rec.Msg)
	}

	// Write all fields as key=value pairs
	for i := int32(0); i < rec.n; i++ {
		field := rec.fields[i]
		buf.WriteByte(' ')
		buf.WriteString(field.K)
		buf.WriteByte('=')
		encodeConsoleValue(field, buf)
	}

	buf.WriteByte('\n')
}

// encodeConsoleValue writes a field value to the buffer using console-appropriate formatting.
// Values are formatted without JSON encoding for better readability.
func encodeConsoleValue(field Field, buf *bytes.Buffer) {
	switch field.T {
	case kindString:
		writeMaybeQuoted(field.Str, buf)
	case kindInt64:
		buf.WriteString(strconv.FormatInt(field.I64, 10))
	case kindUint64:
		buf.WriteString(strconv.FormatUint(field.U64, 10))
	case kindFloat64:
		buf.WriteString(strconv.FormatFloat(field.F64, 'f', -1, 64))
	case kindBool:
		if field.I64 != 0 {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case kindDur:
		buf.WriteString(time.Duration(field.I64).String())
	case kindTime:
		buf.WriteString(time.Unix(0, field.I64).Format(time.RFC3339Nano))
	case kindBytes:
		buf.WriteByte('<')
		buf.WriteString(strconv.Itoa(len(field.B)))
		buf.WriteString("B>")
	}
}

// writeMaybeQuoted writes a string to the buffer, adding quotes only if necessary.
// Strings containing spaces, quotes, or control characters are quoted and escaped.
func writeMaybeQuoted(s string, buf *bytes.Buffer) {
	if s == "" {
		buf.WriteString(`""`)
		return
	}

	// Check if quoting is needed
	needsQuoting := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '"', '\\', '\n', '\r':
			needsQuoting = true
			goto quote_check_done
		}
	}

quote_check_done:
	if !needsQuoting {
		buf.WriteString(s)
		return
	}

	quoteString(s, buf)
}

// colorizeLevel applies ANSI color codes to level strings based on severity.
// Colors are chosen to provide good visibility and semantic meaning:
// - Debug: gray (low importance)
// - Info: blue (informational)
// - Warn: yellow (caution)
// - Error: red (problems)
// - DPanic/Panic/Fatal: bright red (critical)
func colorizeLevel(level Level, levelStr string) string {
	const (
		gray      = "\x1b[90m"
		blue      = "\x1b[34m"
		yellow    = "\x1b[33m"
		red       = "\x1b[31m"
		magenta   = "\x1b[35m"
		brightRed = "\x1b[91m"
		reset     = "\x1b[0m"
	)

	switch level {
	case Debug:
		return gray + levelStr + reset
	case Info:
		return blue + levelStr + reset
	case Warn:
		return yellow + levelStr + reset
	case Error:
		return red + levelStr + reset
	case DPanic:
		return magenta + levelStr + reset
	case Panic:
		return brightRed + levelStr + reset
	case Fatal:
		return brightRed + levelStr + reset
	default:
		return levelStr
	}
}
