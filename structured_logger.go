package iris

import (
	"time"
)

// StructuredLogger extends Iris logger with structured internal encoding
func (l *Logger) WithFieldsStructured(fields ...Field) *StructuredEntry {
	entry := &StructuredEntry{
		logger:    l,
		encoder:   NewStructuredEncoder(),
		timestamp: time.Now(),
	}

	// Add fields to structured encoder (no JSON generation)
	for _, field := range fields {
		switch field.Type {
		case StringType:
			entry.encoder.AddString(field.Key, field.String)
		case IntType:
			entry.encoder.AddInt(field.Key, field.Int)
		case BoolType:
			entry.encoder.AddBool(field.Key, field.Bool)
		}
	}

	return entry
}

// StructuredEntry represents a log entry with structured fields
type StructuredEntry struct {
	logger    *Logger
	encoder   *StructuredEncoder
	timestamp time.Time
}

// Info logs with structured fields (lazy JSON conversion)
func (se *StructuredEntry) Info(message string) {
	if !se.logger.level.Enabled(InfoLevel) {
		return
	}

	// Convert structured fields back to Field slice for existing infrastructure
	fields := make([]Field, len(se.encoder.fields))
	for i, sfield := range se.encoder.fields {
		fields[i] = Field{
			Key:  sfield.Key,
			Type: sfield.Type,
		}

		switch sfield.Type {
		case StringType:
			fields[i].String = sfield.StrVal
		case IntType:
			fields[i].Int = sfield.IntVal
		case BoolType:
			fields[i].Bool = sfield.BoolVal
		}
	}

	// Use existing logger infrastructure
	se.logger.log(InfoLevel, message, fields)

	// Reset encoder back to pool
	se.encoder.Reset()
}

// MemoryFootprint returns internal structured memory usage (for benchmarking)
func (se *StructuredEntry) MemoryFootprint() int {
	return se.encoder.MemoryFootprint()
}
