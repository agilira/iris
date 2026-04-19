// encoder-console.go: compact colorized encoder for interactive terminal sessions
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ansiReset  = "\033[0m"
	ansiGray   = "\033[90m"
	ansiCyan   = "\033[36m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
	ansiPurple = "\033[35m"
)

// ConsoleEncoder writes compact, optionally colorized log lines for interactive
// terminal sessions. It is not intended for production log aggregation —
// use JSONEncoder for machine-readable output.
//
// Output format:
//
//	15:04:05.000  INFO   mts  daemon starting  key=value ...
//
// Colors are emitted only when os.Stdout is a real TTY at construction time.
// When stdout is redirected to a file or pipe, output is plain text with no
// ANSI escape sequences.
//
// Security: all user-controlled content (message, field keys, string values)
// is sanitized to strip control characters including ESC (0x1B), preventing
// ANSI injection attacks (CWE-116) that could manipulate the terminal display.
type ConsoleEncoder struct {
	colorize bool
}

// NewConsoleEncoder creates a ConsoleEncoder, enabling colors only when
// os.Stdout is an interactive terminal.
func NewConsoleEncoder() *ConsoleEncoder {
	return &ConsoleEncoder{colorize: isTTY(os.Stdout)}
}

// NewConsoleEncoderWithColorize creates an encoder with explicit color control.
// WHY: isTTY(os.Stdout) returns false in test environments (no real TTY);
// this constructor lets tests exercise both the color and plain-text paths.
func NewConsoleEncoderWithColorize(colorize bool) *ConsoleEncoder {
	return &ConsoleEncoder{colorize: colorize}
}

// isTTY reports whether f is an interactive character device.
// WHY stdlib only: avoids adding golang.org/x/term to iris's dependency graph.
// os.ModeCharDevice covers /dev/tty, /dev/pts/*, and all pseudo-TTYs.
func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// Encode writes one log line to buf.
func (e *ConsoleEncoder) Encode(rec *Record, now time.Time, buf *bytes.Buffer) {
	buf.Grow(128)
	e.writeTimestamp(now, buf)
	e.writeLevel(rec, buf)
	e.writeLogger(rec, buf)
	e.writeMsg(rec, buf)
	e.writeFields(rec, buf)
	buf.WriteByte('\n')
}

func (e *ConsoleEncoder) writeTimestamp(now time.Time, buf *bytes.Buffer) {
	buf.WriteString(now.UTC().Format("15:04:05.000"))
	buf.WriteString("  ")
}

func (e *ConsoleEncoder) writeLevel(rec *Record, buf *bytes.Buffer) {
	label, color := consoleLevelAttrs(rec.Level)
	if e.colorize && color != "" {
		buf.WriteString(color)
		buf.WriteString(label)
		buf.WriteString(ansiReset)
	} else {
		buf.WriteString(label)
	}
	buf.WriteString("  ")
}

func (e *ConsoleEncoder) writeLogger(rec *Record, buf *bytes.Buffer) {
	if rec.Logger == "" {
		return
	}
	buf.WriteString(consoleEscape(rec.Logger))
	buf.WriteString("  ")
}

func (e *ConsoleEncoder) writeMsg(rec *Record, buf *bytes.Buffer) {
	if rec.Msg != "" {
		buf.WriteString(consoleEscape(rec.Msg))
	}
}

func (e *ConsoleEncoder) writeFields(rec *Record, buf *bytes.Buffer) {
	for i := int32(0); i < rec.n; i++ {
		buf.WriteString("  ")
		f := rec.fields[i]
		buf.WriteString(consoleEscape(f.K))
		buf.WriteByte('=')
		e.writeFieldValue(&f, buf)
	}
}

func (e *ConsoleEncoder) writeFieldValue(f *Field, buf *bytes.Buffer) {
	if consoleWriteScalar(f, buf) {
		return
	}
	consoleWriteComplex(f, buf)
}

// consoleWriteScalar handles primitive field types. Returns false for complex types.
func consoleWriteScalar(f *Field, buf *bytes.Buffer) bool {
	switch f.T {
	case kindString:
		consoleWriteString(f.Str, buf)
	case kindSecret:
		buf.WriteString("[REDACTED]")
	case kindInt64:
		buf.WriteString(strconv.FormatInt(f.I64, 10))
	case kindUint64:
		buf.WriteString(strconv.FormatUint(f.U64, 10))
	case kindFloat64:
		buf.WriteString(strconv.FormatFloat(f.F64, 'g', -1, 64))
	case kindBool:
		if f.I64 != 0 {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case kindDur:
		buf.WriteString(time.Duration(f.I64).String())
	case kindTime:
		buf.WriteString(time.Unix(0, f.I64).UTC().Format("15:04:05.000"))
	case kindBytes:
		buf.WriteString("0x")
		for _, b := range f.B {
			buf.WriteString(strconv.FormatUint(uint64(b), 16))
		}
	default:
		return false
	}
	return true
}

// consoleWriteComplex handles error, stringer, and object field types.
func consoleWriteComplex(f *Field, buf *bytes.Buffer) {
	if f.Obj == nil {
		buf.WriteString("nil")
		return
	}
	switch f.T {
	case kindError:
		if err, ok := f.Obj.(error); ok {
			consoleWriteString(err.Error(), buf)
			return
		}
	case kindStringer:
		if s, ok := f.Obj.(interface{ String() string }); ok {
			consoleWriteString(s.String(), buf)
			return
		}
	}
	consoleWriteString(fmt.Sprintf("%v", f.Obj), buf)
}

// consoleLevelAttrs returns the fixed-width uppercase label and ANSI color for l.
func consoleLevelAttrs(l Level) (label, color string) {
	switch l {
	case Debug:
		return "DEBUG", ansiGray
	case Info:
		return "INFO ", ansiCyan
	case Warn:
		return "WARN ", ansiYellow
	case Error:
		return "ERROR", ansiRed
	case Fatal:
		return "FATAL", ansiPurple
	default:
		return "?????", ""
	}
}

// consoleEscape returns s with all control characters (including ESC/0x1B)
// replaced by '_'. Fast path avoids allocation for clean strings.
//
// WHY ESC specifically: ANSI escape sequences start with ESC (0x1B = 27 < 0x20).
// Replacing it prevents user-controlled content from injecting color codes or
// terminal cursor commands (CWE-116 / terminal manipulation attack).
func consoleEscape(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] == 0x7F {
			return consoleEscapeSlow(s)
		}
	}
	return s
}

func consoleEscapeSlow(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == 0x7F {
			b.WriteByte('_')
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// consoleWriteString writes s quoted if it contains spaces, '=', or '"',
// otherwise bare. This keeps key=value pairs visually parseable while
// avoiding unnecessary noise for simple values like filenames or IDs.
func consoleWriteString(s string, buf *bytes.Buffer) {
	escaped := consoleEscape(s)
	if consoleNeedsQuoting(escaped) {
		buf.WriteByte('"')
		buf.WriteString(strings.ReplaceAll(escaped, `"`, `\"`))
		buf.WriteByte('"')
	} else {
		buf.WriteString(escaped)
	}
}

func consoleNeedsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '=' || c == '"' {
			return true
		}
	}
	return false
}
