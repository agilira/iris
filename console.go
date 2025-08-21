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
	buf = e.appendTimestamp(buf, entry)
	buf = e.appendLevel(buf, entry)
	buf = e.appendMessage(buf, entry)
	buf = e.appendCaller(buf, entry)

	// Fields
	if len(entry.Fields) > 0 {
		buf = append(buf, ' ')
		buf = e.appendFields(buf, entry.Fields)
	}

	buf = e.appendStackTrace(buf, entry)

	buf = append(buf, '\n')
	return buf
}

// appendTimestamp adds formatted timestamp to buffer
func (e *ConsoleEncoder) appendTimestamp(buf []byte, entry *LogEntry) []byte {
	if entry.Timestamp.IsZero() {
		return buf
	}

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
	return buf
}

// appendLevel adds formatted level to buffer
func (e *ConsoleEncoder) appendLevel(buf []byte, entry *LogEntry) []byte {
	if e.colorize {
		return e.appendLevelWithColor(buf, entry.Level)
	}
	return e.appendLevelWithoutColor(buf, entry.Level)
}

// appendLevelWithColor adds colored level text
func (e *ConsoleEncoder) appendLevelWithColor(buf []byte, level Level) []byte {
	// Pre-calculate color codes to avoid repeated lookups
	var levelColor, levelText []byte
	switch level {
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
		color := e.levelColor(level)
		buf = append(buf, color...)
		buf = append(buf, e.levelText(level)...)
		buf = append(buf, Reset...)
		buf = append(buf, ' ')
		return buf
	}

	buf = append(buf, levelColor...)
	buf = append(buf, levelText...)
	buf = append(buf, Reset...)
	buf = append(buf, ' ')
	return buf
}

// appendLevelWithoutColor adds plain level text
func (e *ConsoleEncoder) appendLevelWithoutColor(buf []byte, level Level) []byte {
	// Fast path: direct level text without method calls
	switch level {
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
		buf = append(buf, e.levelText(level)...)
	}
	buf = append(buf, ' ')
	return buf
}

// appendMessage adds formatted message to buffer
func (e *ConsoleEncoder) appendMessage(buf []byte, entry *LogEntry) []byte {
	if e.colorize && entry.Level >= ErrorLevel {
		buf = append(buf, Bold...)
		buf = append(buf, entry.Message...)
		buf = append(buf, Reset...)
	} else {
		buf = append(buf, entry.Message...)
	}
	return buf
}

// appendCaller adds caller information to buffer
func (e *ConsoleEncoder) appendCaller(buf []byte, entry *LogEntry) []byte {
	if !entry.Caller.Valid {
		return buf
	}

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
	return buf
}

// appendStackTrace adds stack trace to buffer
func (e *ConsoleEncoder) appendStackTrace(buf []byte, entry *LogEntry) []byte {
	if entry.StackTrace == "" {
		return buf
	}

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
	case StringType, ByteStringType:
		return e.appendStringTypeValue(buf, field)
	case IntType, Int64Type, Int32Type, Int16Type, Int8Type:
		return strconv.AppendInt(buf, field.Int, 10)
	case UintType, Uint64Type, Uint32Type, Uint16Type, Uint8Type:
		return e.appendUintValue(buf, field.Int)
	case Float64Type, Float32Type:
		return strconv.AppendFloat(buf, field.Float, 'f', -1, 64)
	case BoolType:
		return e.appendBoolValue(buf, field.Bool)
	case DurationType, TimeType:
		return e.appendTimeTypeValue(buf, field)
	case ErrorType, BinaryType, AnyType:
		return e.appendSpecialTypeValue(buf, field)
	default:
		return append(buf, field.String...)
	}
}

// appendStringTypeValue handles string and byte string types
func (e *ConsoleEncoder) appendStringTypeValue(buf []byte, field Field) []byte {
	if field.Type == StringType {
		return e.appendStringValue(buf, field.String)
	}
	return e.appendByteStringValue(buf, field.Bytes)
}

// appendTimeTypeValue handles duration and time types
func (e *ConsoleEncoder) appendTimeTypeValue(buf []byte, field Field) []byte {
	if field.Type == DurationType {
		return e.appendDurationValue(buf, field.Int)
	}
	return e.appendTimeValue(buf, field.Int)
}

// appendSpecialTypeValue handles error, binary, and any types
func (e *ConsoleEncoder) appendSpecialTypeValue(buf []byte, field Field) []byte {
	switch field.Type {
	case ErrorType:
		return e.appendErrorValue(buf, field.Err)
	case BinaryType:
		return e.appendBinaryValue(buf, field.Bytes)
	case AnyType:
		return e.appendAnyValue(buf, field.Any)
	default:
		return buf
	}
}

// appendStringValue appends string value with quoting if needed
func (e *ConsoleEncoder) appendStringValue(buf []byte, s string) []byte {
	if needsQuotingFast(s) {
		buf = append(buf, '"')
		buf = append(buf, s...)
		buf = append(buf, '"')
	} else {
		buf = append(buf, s...)
	}
	return buf
}

// appendUintValue appends unsigned integer using safe conversion
func (e *ConsoleEncoder) appendUintValue(buf []byte, value int64) []byte {
	// Use safe conversion for encoding unsigned integers
	uintValue, _ := SafeInt64ToUint64ForEncoding(value)
	return strconv.AppendUint(buf, uintValue, 10)
}

// appendBoolValue appends boolean value
func (e *ConsoleEncoder) appendBoolValue(buf []byte, b bool) []byte {
	if b {
		return append(buf, "true"...)
	}
	return append(buf, "false"...)
}

// appendDurationValue appends duration value
func (e *ConsoleEncoder) appendDurationValue(buf []byte, nanos int64) []byte {
	dur := time.Duration(nanos)
	return append(buf, dur.String()...)
}

// appendTimeValue appends time value in RFC3339 format
func (e *ConsoleEncoder) appendTimeValue(buf []byte, nanos int64) []byte {
	t := time.Unix(0, nanos)
	return t.AppendFormat(buf, time.RFC3339)
}

// appendErrorValue appends error value with quoting if needed
func (e *ConsoleEncoder) appendErrorValue(buf []byte, err error) []byte {
	if err == nil {
		return buf
	}

	errStr := err.Error()
	if needsQuoting(errStr) {
		buf = append(buf, '"')
		buf = append(buf, errStr...)
		buf = append(buf, '"')
	} else {
		buf = append(buf, errStr...)
	}
	return buf
}

// appendByteStringValue appends byte string with quoting if needed
func (e *ConsoleEncoder) appendByteStringValue(buf []byte, bytes []byte) []byte {
	s := string(bytes)
	if needsQuoting(s) {
		buf = append(buf, '"')
		buf = append(buf, s...)
		buf = append(buf, '"')
	} else {
		buf = append(buf, s...)
	}
	return buf
}

// appendBinaryValue appends binary data summary
func (e *ConsoleEncoder) appendBinaryValue(buf []byte, bytes []byte) []byte {
	buf = append(buf, "<binary:"...)
	buf = strconv.AppendInt(buf, int64(len(bytes)), 10)
	buf = append(buf, " bytes>"...)
	return buf
}

// appendAnyValue appends any value representation
func (e *ConsoleEncoder) appendAnyValue(buf []byte, value interface{}) []byte {
	if value != nil {
		return append(buf, "<any>"...)
	}
	return append(buf, "<nil>"...)
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
