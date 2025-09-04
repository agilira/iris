// level.go: Logging level definitions and utilities for Iris logging library
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// Level represents the severity level of a log message.
// Levels are ordered from least to most severe: Debug < Info < Warn < Error < DPanic < Panic < Fatal
//
// Performance Notes:
// - Level is implemented as int32 for fast comparisons
// - Atomic operations used for thread-safe level changes
// - Zero allocation for level checks via inlined comparisons
type Level int32

// Log levels in order of increasing severity
const (
	Debug  Level = iota - 1 // Debug information, typically disabled in production
	Info                    // General information messages
	Warn                    // Warning messages for potentially harmful situations
	Error                   // Error messages for failure conditions
	DPanic                  // Development panic - panics in development, errors in production
	Panic                   // Panic level - logs message then panics
	Fatal                   // Fatal level - logs message then calls os.Exit(1)

	// StacktraceDisabled is a sentinel value used to disable stack trace collection
	StacktraceDisabled Level = -999
)

// levelNamesMap provides reverse lookup from string to level.
// Pre-computed map for faster parsing operations.
var levelNamesMap = map[string]Level{
	"debug":   Debug,
	"info":    Info,
	"warn":    Warn,
	"warning": Warn, // Alias for warn
	"error":   Error,
	"err":     Error, // Alias for error
	"dpanic":  DPanic,
	"panic":   Panic,
	"fatal":   Fatal,
	"":        Info, // Empty string defaults to Info
}

// String returns the string representation of the level.
// This is used for human-readable output and serialization.
func (l Level) String() string {
	switch l {
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Error:
		return "error"
	case DPanic:
		return "dpanic"
	case Panic:
		return "panic"
	case Fatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// Enabled determines if this level is enabled given a minimum level.
// This is a critical hot path function optimized for maximum performance.
func (l Level) Enabled(min Level) bool {
	return l >= min
}

// IsDebug returns true if the level is Debug.
// Convenience method for frequently checked debug level.
func (l Level) IsDebug() bool {
	return l == Debug
}

// IsInfo returns true if the level is Info.
// Convenience method for frequently checked info level.
func (l Level) IsInfo() bool {
	return l == Info
}

// IsWarn returns true if the level is Warn.
// Convenience method for frequently checked warn level.
func (l Level) IsWarn() bool {
	return l == Warn
}

// IsError returns true if the level is Error.
// Convenience method for frequently checked error level.
func (l Level) IsError() bool {
	return l == Error
}

// IsDPanic returns true if the level is DPanic.
// Convenience method for checking development panic level.
func (l Level) IsDPanic() bool {
	return l == DPanic
}

// IsPanic returns true if the level is Panic.
// Convenience method for checking panic level.
func (l Level) IsPanic() bool {
	return l == Panic
}

// IsFatal returns true if the level is Fatal.
// Convenience method for checking fatal level.
func (l Level) IsFatal() bool {
	return l == Fatal
}

// ParseLevel parses a string representation of a level and returns the corresponding Level.
// It handles common aliases and is case-insensitive.
// Returns Info level for empty strings as a sensible default.
func ParseLevel(s string) (Level, error) {
	// Normalize input: trim whitespace and convert to lowercase
	normalized := strings.ToLower(strings.TrimSpace(s))

	// Fast lookup using pre-computed map (includes empty string -> Info)
	if level, exists := levelNamesMap[normalized]; exists {
		return level, nil
	}

	// Return error for unknown levels
	return Info, fmt.Errorf("unknown level %q", s)
}

// MarshalText implements encoding.TextMarshaler for JSON/XML serialization.
// This method is optimized to avoid allocations in the common case.
func (l Level) MarshalText() ([]byte, error) {
	str := l.String()
	if str == "unknown" {
		return nil, fmt.Errorf("cannot marshal unknown level %d", l)
	}
	return []byte(str), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON/XML deserialization.
// This method provides detailed error information for debugging.
func (l *Level) UnmarshalText(b []byte) error {
	if l == nil {
		return fmt.Errorf("cannot unmarshal into nil Level pointer")
	}

	parsed, err := ParseLevel(string(b))
	if err != nil {
		return fmt.Errorf("failed to unmarshal level: %w", err)
	}

	*l = parsed
	return nil
}

// AtomicLevel provides atomic operations on Level values.
// This is useful for dynamically changing log levels in concurrent environments.
type AtomicLevel struct {
	level int32
}

// NewAtomicLevel creates a new AtomicLevel with the given initial level.
func NewAtomicLevel(level Level) *AtomicLevel {
	return &AtomicLevel{level: int32(level)}
}

// Level returns the current level atomically.
func (al *AtomicLevel) Level() Level {
	return Level(atomic.LoadInt32(&al.level))
}

// SetLevel sets the level atomically.
func (al *AtomicLevel) SetLevel(level Level) {
	atomic.StoreInt32(&al.level, int32(level))
}

// Enabled checks if the given level is enabled atomically.
// This is a high-performance method for checking levels in hot paths.
func (al *AtomicLevel) Enabled(level Level) bool {
	return level >= Level(atomic.LoadInt32(&al.level))
}

// String returns the string representation of the current level.
func (al *AtomicLevel) String() string {
	return al.Level().String()
}

// MarshalText implements encoding.TextMarshaler for AtomicLevel.
func (al *AtomicLevel) MarshalText() ([]byte, error) {
	return al.Level().MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler for AtomicLevel.
func (al *AtomicLevel) UnmarshalText(b []byte) error {
	var level Level
	if err := level.UnmarshalText(b); err != nil {
		return err
	}
	al.SetLevel(level)
	return nil
}

// LevelFlag is a command-line flag implementation for Level.
// It implements the flag.Value interface for easy CLI integration.
type LevelFlag struct {
	level *Level
}

// NewLevelFlag creates a new LevelFlag pointing to the given Level.
func NewLevelFlag(level *Level) *LevelFlag {
	return &LevelFlag{level: level}
}

// String returns the string representation of the level.
func (lf *LevelFlag) String() string {
	if lf.level == nil {
		return Info.String()
	}
	return lf.level.String()
}

// Set parses and sets the level from a string.
// This method is called by the flag package when parsing command-line arguments.
func (lf *LevelFlag) Set(s string) error {
	if lf.level == nil {
		return fmt.Errorf("cannot set level on nil LevelFlag")
	}

	parsed, err := ParseLevel(s)
	if err != nil {
		return fmt.Errorf("failed to set level flag: %w", err)
	}

	*lf.level = parsed
	return nil
}

// Type returns the type description for help text.
func (lf *LevelFlag) Type() string {
	return "level"
}

// AllLevels returns a slice of all valid levels in ascending order.
// This is useful for documentation, validation, and testing.
func AllLevels() []Level {
	return []Level{Debug, Info, Warn, Error}
}

// AllLevelNames returns a slice of all valid level names.
// This is useful for generating help text and validation messages.
func AllLevelNames() []string {
	levels := AllLevels()
	names := make([]string, len(levels))
	for i, level := range levels {
		names[i] = level.String()
	}
	return names
}

// IsValidLevel checks if the given level is a valid predefined level.
func IsValidLevel(level Level) bool {
	return level >= Debug && level <= Error
}
