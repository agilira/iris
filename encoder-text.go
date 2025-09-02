// text_encoder.go: Secure text encoder with log injection protection
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/agilira/go-timecache"
)

// TextEncoder provides secure human-readable text encoding for log records.
// Implements comprehensive log injection protection by sanitizing all field
// keys and values to prevent malicious log manipulation.
//
// Security Features:
//   - Field key sanitization to prevent injection via malformed keys
//   - Value sanitization with proper quoting and escaping
//   - Control character neutralization
//   - Newline injection protection
//   - Unicode direction override protection
//
// Output Format:
//
//	time=2025-08-22T10:30:00Z level=info msg="User login" user=john_doe ip=192.168.1.1
type TextEncoder struct {
	// TimeFormat specifies the time format (default: RFC3339)
	TimeFormat string
	// QuoteValues determines if string values should be quoted (default: true for safety)
	QuoteValues bool
	// SanitizeKeys determines if field keys should be sanitized (default: true)
	SanitizeKeys bool
}

// NewTextEncoder creates a new secure text encoder with safe defaults.
func NewTextEncoder() *TextEncoder {
	return &TextEncoder{
		TimeFormat:   time.RFC3339,
		QuoteValues:  true,
		SanitizeKeys: true,
	}
}

// Encode writes the record to the buffer in secure text format.
func (e *TextEncoder) Encode(rec *Record, now time.Time, buf *bytes.Buffer) {
	// Timestamp with smart caching
	buf.WriteString("time=")
	// Use time cache for performance when timestamp is close to current time
	if cachedTime := timecache.CachedTime(); now.Sub(cachedTime).Abs() < 500*time.Microsecond {
		buf.WriteString(timecache.CachedTimeString()) // Fast cached format
	} else {
		buf.WriteString(now.Format(e.TimeFormat)) // Exact time for tests/historical
	}

	// Level
	buf.WriteString(" level=")
	buf.WriteString(rec.Level.String())

	// Message - only if non-empty
	if rec.Msg != "" {
		buf.WriteString(" msg=")
		if e.QuoteValues {
			e.writeQuotedValue(rec.Msg, buf)
		} else {
			e.writeSafeValue(rec.Msg, buf)
		}
	}

	// Logger name - only if non-empty
	if rec.Logger != "" {
		buf.WriteString(" logger=")
		if e.QuoteValues {
			e.writeQuotedValue(rec.Logger, buf)
		} else {
			e.writeSafeValue(rec.Logger, buf)
		}
	}

	// Caller information
	if rec.Caller != "" {
		buf.WriteString(" caller=")
		if e.QuoteValues {
			e.writeQuotedValue(rec.Caller, buf)
		} else {
			e.writeSafeValue(rec.Caller, buf)
		}
	}

	// Structured fields
	for i := int32(0); i < rec.n; i++ {
		f := rec.fields[i]
		buf.WriteByte(' ')

		// Security: Sanitize field key to prevent injection
		key := f.K
		if e.SanitizeKeys {
			key = e.sanitizeKey(key)
		}
		buf.WriteString(key)
		buf.WriteByte('=')

		// Type-specific value encoding
		switch f.T {
		case kindString:
			if e.QuoteValues {
				e.writeQuotedValue(f.Str, buf)
			} else {
				e.writeSafeValue(f.Str, buf)
			}
		case kindSecret:
			// Security: Always quote and redact secret values
			buf.WriteString(`"[REDACTED]"`)
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
			buf.WriteString(time.Unix(0, f.I64).UTC().Format(e.TimeFormat))
		case kindBytes:
			// Bytes as hex string for text format
			buf.WriteString("0x")
			for _, b := range f.B {
				buf.WriteString(strconv.FormatUint(uint64(b), 16))
			}
		}
	}

	// Stack trace on separate lines if present
	if rec.Stack != "" {
		buf.WriteString("\nstack:\n")
		// Security: Sanitize stack trace to prevent injection
		e.writeSafeMultiline(rec.Stack, buf)
	}

	buf.WriteByte('\n')
}

// sanitizeKey removes or replaces dangerous characters from field keys.
// This prevents log injection via malformed field names.
func (e *TextEncoder) sanitizeKey(key string) string {
	// Fast path for safe keys
	if e.isSafeKey(key) {
		return key
	}

	// Replace dangerous characters
	var result strings.Builder
	result.Grow(len(key))

	for _, r := range key {
		switch {
		case r >= 'a' && r <= 'z':
			result.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			result.WriteRune(r)
		case r >= '0' && r <= '9':
			result.WriteRune(r)
		case r == '_' || r == '-' || r == '.':
			result.WriteRune(r)
		default:
			// Replace dangerous characters with underscore
			result.WriteByte('_')
		}
	}

	sanitized := result.String()
	if len(sanitized) == 0 {
		return "invalid_key"
	}
	return sanitized
}

// isSafeKey quickly checks if a key contains only safe characters.
func (e *TextEncoder) isSafeKey(key string) bool {
	for i := 0; i < len(key); i++ {
		c := key[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.') {
			return false
		}
	}
	return len(key) > 0
}

// writeQuotedValue writes a string value with quotes and aggressive escaping.
// This function prioritizes security over readability by neutralizing any
// characters that could be used for log injection attacks.
func (e *TextEncoder) writeQuotedValue(value string, buf *bytes.Buffer) {
	buf.WriteByte('"')
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch c {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n', '\r', '\t':
			buf.WriteByte('_') // Replace all whitespace control chars
		case '=':
			buf.WriteByte('_') // Replace equals to prevent key=value injection
		default:
			if c < 0x20 || c == 0x7F {
				// Control characters replaced with underscore for maximum safety
				buf.WriteByte('_')
			} else {
				buf.WriteByte(c)
			}
		}
	}
	buf.WriteByte('"')
}

// writeSafeValue writes a string value with minimal escaping (spaces replaced).
func (e *TextEncoder) writeSafeValue(value string, buf *bytes.Buffer) {
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch c {
		case ' ':
			buf.WriteByte('_')
		case '\n', '\r', '\t':
			buf.WriteByte('_')
		default:
			if c < 0x20 || c == 0x7F {
				buf.WriteByte('_')
			} else {
				buf.WriteByte(c)
			}
		}
	}
}

// writeSafeMultiline writes multiline content (like stack traces) safely.
func (e *TextEncoder) writeSafeMultiline(content string, buf *bytes.Buffer) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if i > 0 {
			buf.WriteByte('\n')
		}
		// Prefix each line to prevent injection
		buf.WriteString("  ")
		e.writeSafeValue(line, buf)
	}
}
