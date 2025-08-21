// format_types.go: Format types and interfaces for Iris logging
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import "time"

// Format represents output format
type Format int

const (
	// JSONFormat produces JSON output (slower but structured)
	JSONFormat Format = iota
	// ConsoleFormat produces colorized console output (development)
	ConsoleFormat
	// FastTextFormat produces fast text output (ultra-fast)
	FastTextFormat
	// BinaryFormat produces binary output (fastest)
	BinaryFormat
)

// String returns the string representation of the format
func (f Format) String() string {
	switch f {
	case JSONFormat:
		return "json"
	case ConsoleFormat:
		return "console"
	case FastTextFormat:
		return "text"
	case BinaryFormat:
		return "binary"
	default:
		return "unknown"
	}
}

// Encoder interface for all log encoders
type Encoder interface {
	// Reset resets the encoder for reuse
	Reset()

	// Bytes returns the encoded bytes
	Bytes() []byte

	// EncodeLogEntry encodes a log entry with BinaryField slice
	EncodeLogEntry(timestamp time.Time, level Level, message string, fields []BinaryField, caller Caller, stackTrace string)

	// EncodeLogEntryMigration encodes a log entry with Field slice (migration support)
	EncodeLogEntryMigration(timestamp time.Time, level Level, message string, fields []Field, caller Caller, stackTrace string)
}

// EncoderConfig holds configuration for encoders
type EncoderConfig struct {
	// BufferSize initial buffer size for encoders
	BufferSize int

	// TimeFormat format for timestamps
	TimeFormat string

	// IncludeCaller whether to include caller information
	IncludeCaller bool

	// IncludeStackTrace whether to include stack traces
	IncludeStackTrace bool
}

// DefaultEncoderConfig returns the default encoder configuration
func DefaultEncoderConfig() EncoderConfig {
	return EncoderConfig{
		BufferSize:        1024,
		TimeFormat:        "2006-01-02T15:04:05.000Z07:00",
		IncludeCaller:     false,
		IncludeStackTrace: false,
	}
}
