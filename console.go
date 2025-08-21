// console.go: Human-readable console encoder for development
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strconv"
	"strings"
	"time"
)

// ConsoleEncoder formats log entries in human-readable format
type ConsoleEncoder struct {
	colorize bool
}

// Color codes for console output
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BrightRed = "\033[91m"
)

// NewConsoleEncoder creates a new console encoder
func NewConsoleEncoder(colorize bool) *ConsoleEncoder {
	return &ConsoleEncoder{
		colorize: colorize,
	}
}

// EncodeLogEntry formats a log entry as human-readable console output.
// Performance-optimized with hot path analysis and micro-optimizations.
//
//go:inline
func (e *ConsoleEncoder) EncodeLogEntry(entry *LogEntry, buf []byte) []byte {
	// Reset buffer
	buf = buf[:0]

	// Format: [TIMESTAMP] LEVEL MESSAGE [CALLER] field1=value1 field2=value2

	// Timestamp formatting - optimize common case of no timestamp
	if !entry.Timestamp.IsZero() {
		// Fast path: pre-calculate timestamp size and avoid repeated allocations
		const timestampSize = 23 // "[2006-01-02 15:04:05.000] "
		if cap(buf)-len(buf) < timestampSize {
			// Unlikely path: ensure capacity
			newBuf := make([]byte, len(buf), len(buf)+timestampSize+256)
			copy(newBuf, buf)
			buf = newBuf
		}

		buf = append(buf, '[')
		buf = entry.Timestamp.AppendFormat(buf, "2006-01-02 15:04:05.000")
		buf = append(buf, ']', ' ')
	}

	// Level formatting - hot path optimization
	if e.colorize {
		// Pre-calculate color codes to avoid repeated lookups
		var levelColor, levelText []byte
		switch entry.Level {
		case DebugLevel:
			levelColor = []byte(Magenta)
			levelText = []byte("DEBUG")
		case InfoLevel:
			levelColor = []byte(Blue)
			levelText = []byte("INFO ")
		case WarnLevel:
			levelColor = []byte(Yellow)
			levelText = []byte("WARN ")
		case ErrorLevel:
			levelColor = []byte(Red)
			levelText = []byte("ERROR")
		default:
			// Cold path: use method calls for uncommon levels
			color := e.levelColor(entry.Level)
			buf = append(buf, color...)
			buf = append(buf, e.levelText(entry.Level)...)
			buf = append(buf, Reset...)
			goto afterLevel
		}

		buf = append(buf, levelColor...)
		buf = append(buf, levelText...)
		buf = append(buf, Reset...)
	} else {
		// Fast path: direct level text without method calls
		switch entry.Level {
		case DebugLevel:
			buf = append(buf, "DEBUG"...)
		case InfoLevel:
			buf = append(buf, "INFO "...)
		case WarnLevel:
			buf = append(buf, "WARN "...)
		case ErrorLevel:
			buf = append(buf, "ERROR"...)
		default:
			// Cold path: use method call for uncommon levels
			buf = append(buf, e.levelText(entry.Level)...)
		}
	}

afterLevel:
	buf = append(buf, ' ')

	// Message
	if e.colorize && entry.Level >= ErrorLevel {
		buf = append(buf, Bold...)
		buf = append(buf, entry.Message...)
		buf = append(buf, Reset...)
	} else {
		buf = append(buf, entry.Message...)
	}

	// Caller information (show abbreviated file path)
	if entry.Caller.Valid {
		buf = append(buf, ' ')
		if e.colorize {
			buf = append(buf, Cyan...)
		}
		buf = append(buf, '[')

		// Show only filename, not full path
		filename := entry.Caller.File
		if idx := lastIndex(filename, "/"); idx >= 0 {
			filename = filename[idx+1:]
		}

		buf = append(buf, filename...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(entry.Caller.Line), 10)
		buf = append(buf, ']')

		if e.colorize {
			buf = append(buf, Reset...)
		}
	}

	// Fields
	if len(entry.Fields) > 0 {
		buf = append(buf, ' ')
		buf = e.appendFields(buf, entry.Fields)
	}

	// Stack trace
	if entry.StackTrace != "" {
		buf = append(buf, '\n')
		if e.colorize {
			buf = append(buf, Red...) // Red for stack trace
		}
		buf = append(buf, "Stack trace:"...)
		if e.colorize {
			buf = append(buf, Reset...)
		}
		buf = append(buf, '\n')
		buf = append(buf, entry.StackTrace...)
	}

	buf = append(buf, '\n')
	return buf
}

// levelColor returns the ANSI color code for a log level
func (e *ConsoleEncoder) levelColor(level Level) string {
	switch level {
	case DebugLevel:
		return Magenta
	case InfoLevel:
		return Blue
	case WarnLevel:
		return Yellow
	case ErrorLevel:
		return Red
	case DPanicLevel:
		return BrightRed + Bold
	case PanicLevel:
		return BrightRed + Bold
	case FatalLevel:
		return BrightRed + Bold
	default:
		return White
	}
}

// levelText returns the text representation of a log level
func (e *ConsoleEncoder) levelText(level Level) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO "
	case WarnLevel:
		return "WARN "
	case ErrorLevel:
		return "ERROR"
	case DPanicLevel:
		return "DPANIC"
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// appendFields formats and appends fields to the buffer.
// Performance-optimized with reduced method calls and hot path analysis.
//
//go:inline
func (e *ConsoleEncoder) appendFields(buf []byte, fields []Field) []byte {
	// Hot path: optimize common case of few fields
	for i, field := range fields {
		if i > 0 {
			buf = append(buf, ' ')
		}

		// Key formatting - avoid colorize method calls in hot path
		if e.colorize {
			buf = append(buf, Cyan...)
			buf = append(buf, field.Key...)
			buf = append(buf, Reset...)
		} else {
			buf = append(buf, field.Key...)
		}
		buf = append(buf, '=')

		// Value formatting - inline hot paths
		buf = e.appendFieldValue(buf, field)
	}
	return buf
}

// appendFieldValue appends a field value to the buffer.
// Ultra-optimized with hot path specialization and reduced allocations.
//
//go:inline
func (e *ConsoleEncoder) appendFieldValue(buf []byte, field Field) []byte {
	switch field.Type {
	case StringType:
		// Optimize string quoting check
		if needsQuotingFast(field.String) {
			buf = append(buf, '"')
			buf = append(buf, field.String...)
			buf = append(buf, '"')
		} else {
			buf = append(buf, field.String...)
		}
	case IntType, Int64Type, Int32Type, Int16Type, Int8Type:
		buf = strconv.AppendInt(buf, field.Int, 10)
	case UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		buf = strconv.AppendUint(buf, uint64(field.Int), 10)
	case Float64Type, Float32Type:
		buf = strconv.AppendFloat(buf, field.Float, 'f', -1, 64)
	case BoolType:
		if field.Bool {
			buf = append(buf, "true"...)
		} else {
			buf = append(buf, "false"...)
		}
	case DurationType:
		dur := time.Duration(field.Int)
		buf = append(buf, dur.String()...)
	case TimeType:
		t := time.Unix(0, field.Int)
		buf = t.AppendFormat(buf, time.RFC3339)
	case ErrorType:
		if field.Err != nil {
			if needsQuoting(field.Err.Error()) {
				buf = append(buf, '"')
				buf = append(buf, field.Err.Error()...)
				buf = append(buf, '"')
			} else {
				buf = append(buf, field.Err.Error()...)
			}
		}
	case ByteStringType:
		s := string(field.Bytes)
		if needsQuoting(s) {
			buf = append(buf, '"')
			buf = append(buf, s...)
			buf = append(buf, '"')
		} else {
			buf = append(buf, s...)
		}
	case BinaryType:
		buf = append(buf, "<binary:"...)
		buf = strconv.AppendInt(buf, int64(len(field.Bytes)), 10)
		buf = append(buf, " bytes>"...)
	case AnyType:
		if field.Any != nil {
			buf = append(buf, "<any>"...)
		} else {
			buf = append(buf, "<nil>"...)
		}
	default:
		buf = append(buf, field.String...)
	}
	return buf
}

// needsQuoting returns true if a string needs to be quoted
func needsQuoting(s string) bool {
	return strings.ContainsAny(s, " \t\n\r\"\\=")
}

// needsQuotingFast is an optimized version of needsQuoting for hot paths.
// Uses manual scanning instead of ContainsAny for better performance.
//
//go:inline
func needsQuotingFast(s string) bool {
	// Hot path: most strings don't need quoting
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Check for characters that require quoting
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '"' || c == '\\' || c == '=' {
			return true
		}
	}
	return false
}

// lastIndex returns the last index of substr in s, or -1 if not found
func lastIndex(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
