package iris

// StructuredLogger extends Iris logger with optimized field handling
func (l *Logger) WithFieldsStructured(fields ...Field) *StructuredEntry {
	// NO TIMESTAMP! Let the internal logger.log() handle it with CachedTime()
	entry := &StructuredEntry{
		logger: l,
		fields: fields, // Store fields directly - NO CONVERSION!
	}

	return entry
}

// StructuredEntry represents a log entry with pre-computed fields
type StructuredEntry struct {
	logger *Logger
	fields []Field // Direct field storage - NO INTERMEDIATE LAYER!
	// NO TIMESTAMP! Saves 24 bytes per entry + time.Now() call
}

// Info logs with structured fields (ZERO conversion overhead)
func (se *StructuredEntry) Info(message string) {
	if !se.logger.level.Enabled(InfoLevel) {
		return
	}

	// Direct pass-through to existing optimized infrastructure
	// NO CONVERSION! NO ALLOCATIONS! NO DOUBLE ENCODING!
	se.logger.log(InfoLevel, message, se.fields)
}

// MemoryFootprint returns memory usage (for benchmarking)
func (se *StructuredEntry) MemoryFootprint() int {
	memSize := 0
	for _, field := range se.fields {
		memSize += len(field.Key)
		switch field.Type {
		case StringType:
			memSize += len(field.String)
		case IntType:
			memSize += 8 // int64 size
		case BoolType:
			memSize += 1 // bool size
		}
	}
	return memSize
}
