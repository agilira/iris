// level.go: Logging levels for Iris
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"fmt"
	"os"
	"strings"
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

// String returns a lower-case ASCII representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case DPanicLevel:
		return "dpanic"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// CapitalString returns an all-caps ASCII representation of the log level.
func (l Level) CapitalString() string {
	return strings.ToUpper(l.String())
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
