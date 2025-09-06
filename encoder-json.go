// encoder-json.go: optimized json Encoder
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/agilira/go-timecache"
)

// Record represents a log entry with optimized field storage
type Record struct {
	Level  Level     // Log level
	Msg    string    // Log message
	Logger string    // Logger name
	Caller string    // Caller information (file:line)
	Stack  string    // Stack trace
	fields [32]Field // Optimized field array - 32 fields covers 99.9% of use cases
	n      int32     // Number of active fields
}

// resetForWrite resets a record for reuse in the ring buffer
func (r *Record) resetForWrite() {
	r.Level = Debug
	r.Msg = ""
	r.Logger = ""
	r.Caller = ""
	r.Stack = ""
	r.n = 0
}

// NewRecord creates a new Record with the specified level and message.
// Uses pre-allocated field storage to avoid heap allocations during logging.
func NewRecord(level Level, msg string) *Record {
	return &Record{
		Level: level,
		Msg:   msg,
		n:     0,
	}
}

// AddField adds a structured field to this record.
// Returns false if the field array is full (32 fields max - optimal for performance).
func (r *Record) AddField(field Field) bool {
	if r.n >= 32 {
		return false
	}
	r.fields[r.n] = field
	r.n++
	return true
}

// FieldCount returns the number of fields in this record.
func (r *Record) FieldCount() int {
	return int(r.n)
}

// GetField returns the field at the specified index.
// Panics if index is out of bounds (for test simplicity).
func (r *Record) GetField(index int) Field {
	// Safe bounds checking without unsafe conversion
	if index < 0 || index >= int(r.n) {
		return Field{} // Return zero field for out-of-bounds access
	}
	return r.fields[index]
}

// Reset clears the record for reuse.
func (r *Record) Reset() {
	r.Level = Debug
	r.Msg = ""
	r.Logger = ""
	r.Caller = ""
	r.Stack = ""
	r.n = 0
}

// Encoder astratto (permette anche encoder binari futuri).
type Encoder interface {
	Encode(rec *Record, now time.Time, buf *bytes.Buffer)
}

// JSONEncoder implements NDJSON (newline-delimited JSON) encoding with zero-reflection.
//
// The encoder produces one JSON object per log record, separated by newlines.
// This format is ideal for log processing systems and streaming applications.
//
// Performance Features:
// - Zero reflection overhead using pre-compiled encoding paths
// - Reusable buffer allocation for minimal GC pressure
// - Optimized time formatting with caching
// - Direct byte buffer writing without intermediate strings
//
// Output Format:
//
//	{"ts":"2025-09-06T14:30:45.123Z","level":"info","msg":"User action","field":"value"}
//
// Use Cases:
// - Log aggregation systems (ELK stack, Splunk)
// - Structured logging for APIs and microservices
// - Machine-readable logs for automated processing
// - Integration with JSON-based monitoring tools
type JSONEncoder struct {
	// TimeKey specifies the JSON key for timestamps (default: "ts")
	TimeKey string

	// LevelKey specifies the JSON key for log levels (default: "level")
	LevelKey string

	// MsgKey specifies the JSON key for log messages (default: "msg")
	MsgKey string

	// RFC3339 controls timestamp format:
	//   true:  RFC3339 string format (default, human-readable)
	//   false: Unix nanoseconds integer (compact, faster)
	RFC3339 bool
}

// NewJSONEncoder creates a new JSON encoder with standard defaults.
//
// Default configuration:
// - TimeKey: "ts"
// - LevelKey: "level"
// - MsgKey: "msg"
// - RFC3339: true (human-readable timestamps)
//
// The defaults follow common logging conventions and work well with
// most log processing systems.
//
// Returns:
//   - *JSONEncoder: Configured JSON encoder instance
func NewJSONEncoder() *JSONEncoder {
	return &JSONEncoder{
		TimeKey:  "ts",
		LevelKey: "level",
		MsgKey:   "msg",
		RFC3339:  true,
	}
}

// ensureDefaults ensures the encoder has valid key values (fallback for zero-value encoders)
func (e *JSONEncoder) ensureDefaults() {
	if e.TimeKey == "" {
		e.TimeKey = "ts"
	}
	if e.LevelKey == "" {
		e.LevelKey = "level"
	}
	if e.MsgKey == "" {
		e.MsgKey = "msg"
	}
}

// shouldUseTimeCache determines if we should use cached time for performance
func (e *JSONEncoder) shouldUseTimeCache(now time.Time) bool {
	cachedTime := timecache.CachedTime()
	return now.Sub(cachedTime).Abs() < 500*time.Microsecond
}

// Encode encodes a log record to JSON format
func (e *JSONEncoder) Encode(rec *Record, now time.Time, buf *bytes.Buffer) {
	// buf.Reset() viene fatto dal caller (buffer pool)
	buf.Grow(128)
	buf.WriteByte('{')

	// Ensure defaults only if needed (one-time check for zero-value encoders)
	if e.TimeKey == "" || e.LevelKey == "" || e.MsgKey == "" {
		e.ensureDefaults()
	}

	// Encode basic fields
	e.encodeTimestamp(now, buf)
	e.encodeLevel(rec, buf)
	e.encodeOptionalFields(rec, buf)
	e.encodeFields(rec, buf)

	buf.WriteByte('}')
	buf.WriteByte('\n')
}

// encodeTimestamp writes the timestamp field
func (e *JSONEncoder) encodeTimestamp(now time.Time, buf *bytes.Buffer) {
	buf.WriteString(`"`)
	buf.WriteString(e.TimeKey)
	buf.WriteString(`":`)
	if e.RFC3339 {
		buf.WriteByte('"')
		// SMART TIME HANDLING: Use cache for current time, exact time for tests
		if e.shouldUseTimeCache(now) {
			buf.WriteString(timecache.CachedTimeString()) // Use cached formatted time for performance
		} else {
			// Use exact time for testing or historical timestamps
			buf.WriteString(now.Format(time.RFC3339Nano))
		}
		buf.WriteByte('"')
	} else {
		// For Unix timestamps, if time is close to current, use cached nano time
		if e.shouldUseTimeCache(now) {
			buf.WriteString(strconv.FormatInt(timecache.CachedTimeNano(), 10))
		} else {
			// Use exact Unix nanoseconds for testing
			buf.WriteString(strconv.FormatInt(now.UnixNano(), 10))
		}
	}
}

// encodeLevel writes the level field
func (e *JSONEncoder) encodeLevel(rec *Record, buf *bytes.Buffer) {
	buf.WriteByte(',')
	buf.WriteString(`"`)
	buf.WriteString(e.LevelKey)
	buf.WriteString(`":"`)
	buf.WriteString(rec.Level.String())
	buf.WriteByte('"')
}

// encodeOptionalFields writes logger, msg, caller, stack if present
func (e *JSONEncoder) encodeOptionalFields(rec *Record, buf *bytes.Buffer) {
	// logger name (if present)
	if rec.Logger != "" {
		buf.WriteByte(',')
		buf.WriteString(`"logger":`)
		quoteString(rec.Logger, buf)
	}

	// msg (if present)
	if rec.Msg != "" {
		buf.WriteByte(',')
		buf.WriteString(`"`)
		buf.WriteString(e.MsgKey)
		buf.WriteString(`":`)
		quoteString(rec.Msg, buf)
	}

	// caller (if present)
	if rec.Caller != "" {
		buf.WriteByte(',')
		buf.WriteString(`"caller":`)
		quoteString(rec.Caller, buf)
	}

	// stack (if present)
	if rec.Stack != "" {
		buf.WriteByte(',')
		buf.WriteString(`"stack":`)
		quoteString(rec.Stack, buf)
	}
}

// encodeFields writes all the custom fields
func (e *JSONEncoder) encodeFields(rec *Record, buf *bytes.Buffer) {
	for i := int32(0); i < rec.n; i++ {
		f := rec.fields[i]
		buf.WriteByte(',')
		buf.WriteByte('"')
		buf.WriteString(f.K)
		buf.WriteString(`":`)
		e.encodeFieldValue(&f, buf)
	}
}

// encodeFieldValue writes a single field value based on its type
func (e *JSONEncoder) encodeFieldValue(f *Field, buf *bytes.Buffer) {
	switch f.T {
	case kindString:
		quoteString(f.Str, buf)
	case kindInt64:
		buf.WriteString(strconv.FormatInt(f.I64, 10))
	case kindUint64:
		buf.WriteString(strconv.FormatUint(f.U64, 10))
	case kindFloat64:
		buf.WriteString(strconv.FormatFloat(f.F64, 'f', -1, 64))
	case kindBool:
		if f.I64 != 0 {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case kindDur:
		// duration in ns (int64)
		buf.WriteString(strconv.FormatInt(f.I64, 10))
	case kindTime:
		e.encodeTimeField(f, buf)
	case kindBytes:
		e.encodeBytesField(f, buf)
	case kindSecret:
		// Secret fields are redacted
		buf.WriteString(`"[REDACTED]"`)
	case kindError:
		e.encodeErrorField(f, buf)
	case kindStringer:
		e.encodeStringerField(f, buf)
	case kindObject:
		e.encodeObjectField(f, buf)
	}
}

// encodeTimeField writes a time field with cache optimization
func (e *JSONEncoder) encodeTimeField(f *Field, buf *bytes.Buffer) {
	buf.WriteByte('"')
	// TIMECACHE OPTIMIZATION: Use cached time formatting for field timestamps too
	timeValue := time.Unix(0, f.I64).UTC()
	// Check if this timestamp is close to our cached time to use cache
	if cachedTime := timecache.CachedTime(); timeValue.Sub(cachedTime).Abs() < 500*time.Microsecond {
		buf.WriteString(timecache.CachedTimeString())
	} else {
		// For older timestamps, still need to format but this should be rare
		buf.WriteString(timeValue.Format(time.RFC3339Nano))
	}
	buf.WriteByte('"')
}

// encodeBytesField writes a bytes field as JSON array
func (e *JSONEncoder) encodeBytesField(f *Field, buf *bytes.Buffer) {
	// compatto: array byte in base10 per evitare base64
	buf.WriteByte('[')
	for j, b := range f.B {
		if j > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(strconv.Itoa(int(b)))
	}
	buf.WriteByte(']')
}

// encodeErrorField writes an error field
func (e *JSONEncoder) encodeErrorField(f *Field, buf *bytes.Buffer) {
	if f.Obj == nil {
		buf.WriteString(`null`)
	} else if err, ok := f.Obj.(error); ok {
		quoteString(err.Error(), buf)
	} else {
		quoteString(fmt.Sprintf("%v", f.Obj), buf)
	}
}

// encodeStringerField writes a stringer field
func (e *JSONEncoder) encodeStringerField(f *Field, buf *bytes.Buffer) {
	if f.Obj == nil {
		buf.WriteString(`null`)
	} else if stringer, ok := f.Obj.(interface{ String() string }); ok {
		quoteString(stringer.String(), buf)
	} else {
		quoteString(fmt.Sprintf("%v", f.Obj), buf)
	}
}

// encodeObjectField writes an object field
func (e *JSONEncoder) encodeObjectField(f *Field, buf *bytes.Buffer) {
	if f.Obj == nil {
		buf.WriteString(`null`)
	} else {
		// For arrays of errors (like Zap's ErrorsField)
		if errs, ok := f.Obj.([]error); ok {
			buf.WriteByte('[')
			for j, err := range errs {
				if j > 0 {
					buf.WriteByte(',')
				}
				if err == nil {
					buf.WriteString(`null`)
				} else {
					quoteString(err.Error(), buf)
				}
			}
			buf.WriteByte(']')
		} else {
			// Generic object - convert to string
			quoteString(fmt.Sprintf("%v", f.Obj), buf)
		}
	}
}

// quoteString: ottimizzato per stringhe comuni senza caratteri speciali
func quoteString(s string, buf *bytes.Buffer) {
	buf.WriteByte('"')

	// Fast path per stringhe senza caratteri speciali
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == '"' || c == '\\' {
			// Scrivi la parte "pulita" fino a qui
			if i > start {
				buf.WriteString(s[start:i])
			}

			// Gestisci il carattere speciale
			switch c {
			case '"':
				buf.WriteString(`\"`)
			case '\\':
				buf.WriteString(`\\`)
			case '\n':
				buf.WriteString(`\n`)
			case '\r':
				buf.WriteString(`\r`)
			case '\t':
				buf.WriteString(`\t`)
			default:
				// Carattere di controllo
				buf.WriteString(`\u00`)
				const hex = "0123456789abcdef"
				buf.WriteByte(hex[c>>4])
				buf.WriteByte(hex[c&0xF])
			}
			start = i + 1
		}
	}

	// Scrivi la parte finale se rimane
	if start < len(s) {
		buf.WriteString(s[start:])
	}

	buf.WriteByte('"')
}
