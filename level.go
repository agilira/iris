package iris

import (
	"fmt"
	"os"
)

// Level represents the severity of a log entry
type Level int8

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = iota - 1
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
)

// Pre-computed string representations for ultra-fast lookups
var (
	levelNames = [...]string{
		0: "debug",  // DebugLevel = -1, index 0
		1: "info",   // InfoLevel = 0, index 1
		2: "warn",   // WarnLevel = 1, index 2
		3: "error",  // ErrorLevel = 2, index 3
		4: "dpanic", // DPanicLevel = 3, index 4
		5: "panic",  // PanicLevel = 4, index 5
		6: "fatal",  // FatalLevel = 5, index 6
	}

	levelNamesCapital = [...]string{
		0: "DEBUG",  // DebugLevel = -1, index 0
		1: "INFO",   // InfoLevel = 0, index 1
		2: "WARN",   // WarnLevel = 1, index 2
		3: "ERROR",  // ErrorLevel = 2, index 3
		4: "DPANIC", // DPanicLevel = 3, index 4
		5: "PANIC",  // PanicLevel = 4, index 5
		6: "FATAL",  // FatalLevel = 5, index 6
	}
)

// String returns a lower-case ASCII representation of the log level.
// Optimized with pre-computed lookup tables for zero allocations.
func (l Level) String() string {
	// Fast path: use lookup table for known levels
	if l >= DebugLevel && l <= FatalLevel {
		return levelNames[l-DebugLevel]
	}

	// Slow path: unknown levels - optimized fmt.Sprintf alternative
	return "Level(" + formatInt(int8(l)) + ")"
}

// CapitalString returns an all-caps ASCII representation of the log level.
// Optimized with pre-computed lookup tables for zero allocations.
func (l Level) CapitalString() string {
	// Fast path: use lookup table for known levels
	if l >= DebugLevel && l <= FatalLevel {
		return levelNamesCapital[l-DebugLevel]
	}

	// Slow path: unknown levels - optimized for minimal allocations
	return "LEVEL(" + formatInt(int8(l)) + ")"
}

// formatInt is an optimized integer to string conversion for small integers
// avoiding fmt.Sprintf overhead for the unknown level case
func formatInt(i int8) string {
	if i >= 0 && i <= 9 {
		return string('0' + byte(i))
	}
	if i >= -9 && i < 0 {
		return "-" + string('0'+byte(-i))
	}
	// For larger numbers, fall back to strconv (rare case)
	return fmt.Sprintf("%d", i)
}

// Enabled returns true if the given level is at or above this level.
func (l Level) Enabled(lvl Level) bool {
	return lvl >= l
}

// ShouldPanic returns true if the level should cause a panic.
func (l Level) ShouldPanic() bool {
	return l == DPanicLevel || l == PanicLevel
}

// ShouldExit returns true if the level should cause the program to exit.
func (l Level) ShouldExit() bool {
	return l == FatalLevel
}

// HandleSpecial handles special behavior for DPanic, Panic, and Fatal levels.
func (l Level) HandleSpecial() {
	switch l {
	case DPanicLevel:
		// In development, DPanic panics. In production, it's just an error.
		// For simplicity, we'll always panic for now.
		panic("dpanic level log")
	case PanicLevel:
		panic("panic level log")
	case FatalLevel:
		os.Exit(1)
	}
}
