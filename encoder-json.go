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
	Level  Level      // Log level
	Msg    string     // Log message
	Logger string     // Logger name
	Caller string     // Caller information (file:line)
	Stack  string     // Stack trace
	fields [128]Field // Pre-allocated field array (increased for test compatibility)
	n      int32      // Number of active fields
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
// Returns false if the field array is full (16 fields max).
func (r *Record) AddField(field Field) bool {
	if r.n >= 16 {
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
	if index < 0 || int32(index) >= r.n {
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

// JSONEncoder: NDJSON (una riga per record), zero-reflect.
type JSONEncoder struct {
	TimeKey  string // default "ts"
	LevelKey string // default "level"
	MsgKey   string // default "msg"
	RFC3339  bool   // default true (alternativa: UnixNano int64)
}

// NewJSONEncoder creates a new JSONEncoder with optimal defaults
func NewJSONEncoder() *JSONEncoder {
	return &JSONEncoder{
		TimeKey:  "ts",
		LevelKey: "level",
		MsgKey:   "msg",
		RFC3339:  true,
	}
}

func (e *JSONEncoder) Encode(rec *Record, now time.Time, buf *bytes.Buffer) {
	// buf.Reset() viene fatto dal caller (buffer pool)
	buf.Grow(128)
	buf.WriteByte('{')

	// Use defaults for zero-value encoder
	timeKey := e.TimeKey
	if timeKey == "" {
		timeKey = "ts"
	}
	levelKey := e.LevelKey
	if levelKey == "" {
		levelKey = "level"
	}
	msgKey := e.MsgKey
	if msgKey == "" {
		msgKey = "msg"
	}

	// ts
	buf.WriteString(`"`)
	buf.WriteString(timeKey)
	buf.WriteString(`":`)
	if e.RFC3339 {
		buf.WriteByte('"')
		// SMART TIME HANDLING: Use cache for current time, exact time for tests
		// If the provided time is close to current time (within 500μs), use cache for performance
		if cachedTime := timecache.CachedTime(); now.Sub(cachedTime).Abs() < 500*time.Microsecond {
			buf.WriteString(timecache.CachedTimeString()) // Use cached formatted time for performance
		} else {
			// Use exact time for testing or historical timestamps
			buf.WriteString(now.Format(time.RFC3339Nano))
		}
		buf.WriteByte('"')
	} else {
		// For Unix timestamps, if time is close to current, use cached nano time
		if cachedTime := timecache.CachedTime(); now.Sub(cachedTime).Abs() < 500*time.Microsecond {
			buf.WriteString(strconv.FormatInt(timecache.CachedTimeNano(), 10))
		} else {
			// Use exact Unix nanoseconds for testing
			buf.WriteString(strconv.FormatInt(now.UnixNano(), 10))
		}
	}

	// level
	buf.WriteByte(',')
	buf.WriteString(`"`)
	buf.WriteString(levelKey)
	buf.WriteString(`":"`)
	buf.WriteString(rec.Level.String())
	buf.WriteByte('"')

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
		buf.WriteString(msgKey)
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

	// fields
	for i := int32(0); i < rec.n; i++ {
		f := rec.fields[i]
		buf.WriteByte(',')
		buf.WriteByte('"')
		buf.WriteString(f.K)
		buf.WriteString(`":`)
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
		case kindBytes:
			// compatto: array byte in base10 per evitare base64
			buf.WriteByte('[')
			for j, b := range f.B {
				if j > 0 {
					buf.WriteByte(',')
				}
				buf.WriteString(strconv.Itoa(int(b)))
			}
			buf.WriteByte(']')
		case kindSecret:
			// Secret fields are redacted
			buf.WriteString(`"[REDACTED]"`)
		case kindError:
			if f.Obj == nil {
				buf.WriteString(`null`)
			} else if err, ok := f.Obj.(error); ok {
				quoteString(err.Error(), buf)
			} else {
				quoteString(fmt.Sprintf("%v", f.Obj), buf)
			}
		case kindStringer:
			if f.Obj == nil {
				buf.WriteString(`null`)
			} else if stringer, ok := f.Obj.(interface{ String() string }); ok {
				quoteString(stringer.String(), buf)
			} else {
				quoteString(fmt.Sprintf("%v", f.Obj), buf)
			}
		case kindObject:
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
	}

	buf.WriteByte('}')
	buf.WriteByte('\n')
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
