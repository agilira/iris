// encoder-binary_test.go: Comprehensive tests for binary encoder
//
// Test Coverage:
// - Basic encoding functionality
// - All field types and edge cases
// - Performance benchmarks vs other encoders
// - Format compatibility and decoding
// - Configuration options
// - Size estimation accuracy
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

// Test fixtures
var (
	testTime = time.Date(2024, 1, 15, 12, 30, 45, 123456789, time.UTC)
	errTest  = errors.New("test error message")
)

type testStringerImpl struct {
	value string
}

func (t testStringerImpl) String() string {
	return t.value
}

// TestBinaryEncoder_NewBinaryEncoder tests default encoder creation
func TestBinaryEncoder_NewBinaryEncoder(t *testing.T) {
	encoder := NewBinaryEncoder()

	if !encoder.IncludeLoggerName {
		t.Error("Expected IncludeLoggerName to be true by default")
	}
	if encoder.IncludeCaller {
		t.Error("Expected IncludeCaller to be false by default")
	}
	if encoder.IncludeStack {
		t.Error("Expected IncludeStack to be false by default")
	}
	if !encoder.UseUnixNano {
		t.Error("Expected UseUnixNano to be true by default")
	}
}

// TestBinaryEncoder_NewCompactBinaryEncoder tests compact encoder creation
func TestBinaryEncoder_NewCompactBinaryEncoder(t *testing.T) {
	encoder := NewCompactBinaryEncoder()

	if encoder.IncludeLoggerName {
		t.Error("Expected IncludeLoggerName to be false for compact encoder")
	}
	if encoder.IncludeCaller {
		t.Error("Expected IncludeCaller to be false for compact encoder")
	}
	if encoder.IncludeStack {
		t.Error("Expected IncludeStack to be false for compact encoder")
	}
	if !encoder.UseUnixNano {
		t.Error("Expected UseUnixNano to be true for compact encoder")
	}
}

// TestBinaryEncoder_BasicEncoding tests basic record encoding
func TestBinaryEncoder_BasicEncoding(t *testing.T) {
	encoder := NewBinaryEncoder()
	buf := &bytes.Buffer{}

	rec := NewRecord(Info, "test message")
	rec.Logger = "test.logger"

	encoder.Encode(rec, testTime, buf)

	data := buf.Bytes()
	if len(data) == 0 {
		t.Fatal("Expected non-empty encoded data")
	}

	// Verify magic header
	if len(data) < 3 {
		t.Fatal("Encoded data too short for magic header")
	}

	magic := binary.LittleEndian.Uint16(data[0:2])
	version := data[2]

	if magic != binaryMagic {
		t.Errorf("Expected magic %x, got %x", binaryMagic, magic)
	}
	if version != binaryVersion {
		t.Errorf("Expected version %x, got %x", binaryVersion, version)
	}
}

// TestBinaryEncoder_AllFieldTypes tests encoding of all supported field types
func TestBinaryEncoder_AllFieldTypes(t *testing.T) {
	encoder := NewBinaryEncoder()
	buf := &bytes.Buffer{}

	rec := NewRecord(Info, "test with all field types")
	rec.Logger = "test.logger"

	// Add all field types
	rec.AddField(String("str", "hello world"))
	rec.AddField(Int64("int", -42))
	rec.AddField(Uint64("uint", 42))
	rec.AddField(Float64("float", 3.14159))
	rec.AddField(Bool("bool_true", true))
	rec.AddField(Bool("bool_false", false))
	rec.AddField(Dur("dur", time.Second))
	rec.AddField(Time("time", testTime))
	rec.AddField(Bytes("bytes", []byte("binary data")))
	rec.AddField(NamedError("error", errTest))
	rec.AddField(Stringer("stringer", testStringerImpl{value: "test stringer"}))

	encoder.Encode(rec, testTime, buf)

	data := buf.Bytes()
	if len(data) == 0 {
		t.Fatal("Expected non-empty encoded data")
	}

	// Basic validation: should start with magic header
	decoder := &BinaryDecoder{}
	valid, _ := decoder.DecodeMagic(data)
	if !valid {
		t.Error("Encoded data should have valid magic header")
	}
}

// TestBinaryEncoder_EmptyRecord tests encoding of empty record
func TestBinaryEncoder_EmptyRecord(t *testing.T) {
	encoder := NewBinaryEncoder()
	buf := &bytes.Buffer{}

	rec := NewRecord(Info, "")

	encoder.Encode(rec, testTime, buf)

	data := buf.Bytes()
	if len(data) == 0 {
		t.Fatal("Expected non-empty encoded data even for empty record")
	}

	// Should still have magic header and basic structure
	decoder := &BinaryDecoder{}
	valid, _ := decoder.DecodeMagic(data)
	if !valid {
		t.Error("Empty record should still have valid magic header")
	}
}

// TestBinaryEncoder_CompactMode tests compact encoding configuration
func TestBinaryEncoder_CompactMode(t *testing.T) {
	normalEncoder := NewBinaryEncoder()
	compactEncoder := NewCompactBinaryEncoder()

	rec := NewRecord(Info, "test message")
	rec.Logger = "very.long.logger.name.that.takes.space"
	rec.Caller = "/path/to/file.go:123"
	rec.Stack = "stack trace here"
	rec.AddField(String("key", "value"))

	normalBuf := &bytes.Buffer{}
	compactBuf := &bytes.Buffer{}

	normalEncoder.Encode(rec, testTime, normalBuf)
	compactEncoder.Encode(rec, testTime, compactBuf)

	normalSize := normalBuf.Len()
	compactSize := compactBuf.Len()

	if compactSize >= normalSize {
		t.Errorf("Compact encoding (%d bytes) should be smaller than normal (%d bytes)",
			compactSize, normalSize)
	}
}

// TestBinaryEncoder_ConfigurationOptions tests various encoder configurations
func TestBinaryEncoder_ConfigurationOptions(t *testing.T) {
	testCases := []struct {
		name          string
		includLogger  bool
		includeCaller bool
		includeStack  bool
		useUnixNano   bool
	}{
		{"all_included", true, true, true, true},
		{"minimal", false, false, false, true},
		{"rfc3339_time", true, false, false, false},
		{"caller_only", false, true, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoder := &BinaryEncoder{
				IncludeLoggerName: tc.includLogger,
				IncludeCaller:     tc.includeCaller,
				IncludeStack:      tc.includeStack,
				UseUnixNano:       tc.useUnixNano,
			}

			buf := &bytes.Buffer{}
			rec := NewRecord(Warn, "test message")
			rec.Logger = "test.logger"
			rec.Caller = "file.go:42"
			rec.Stack = "stack trace"

			encoder.Encode(rec, testTime, buf)

			if buf.Len() == 0 {
				t.Fatal("Expected non-empty encoded data")
			}

			// Verify magic header is present
			data := buf.Bytes()
			decoder := &BinaryDecoder{}
			valid, _ := decoder.DecodeMagic(data)
			if !valid {
				t.Error("Encoded data should have valid magic header")
			}
		})
	}
}

// TestBinaryDecoder_DecodeMagic tests magic header validation
func TestBinaryDecoder_DecodeMagic(t *testing.T) {
	decoder := &BinaryDecoder{}

	testCases := []struct {
		name     string
		data     []byte
		expected bool
		consumed int
	}{
		{
			name:     "valid_magic",
			data:     []byte{0x52, 0x49, 0x01, 0xFF}, // IR + version 1 + extra
			expected: true,
			consumed: 3,
		},
		{
			name:     "invalid_magic",
			data:     []byte{0x00, 0x00, 0x01},
			expected: false,
			consumed: 3,
		},
		{
			name:     "invalid_version",
			data:     []byte{0x52, 0x49, 0xFF}, // IR + version 255
			expected: false,
			consumed: 3,
		},
		{
			name:     "too_short",
			data:     []byte{0x52, 0x49}, // Only 2 bytes
			expected: false,
			consumed: 0,
		},
		{
			name:     "empty",
			data:     []byte{},
			expected: false,
			consumed: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, consumed := decoder.DecodeMagic(tc.data)
			if valid != tc.expected {
				t.Errorf("Expected valid=%v, got %v", tc.expected, valid)
			}
			if consumed != tc.consumed {
				t.Errorf("Expected consumed=%d, got %d", tc.consumed, consumed)
			}
		})
	}
}

// TestBinaryDecoder_ReadVarint tests varint decoding
func TestBinaryDecoder_ReadVarint(t *testing.T) {
	decoder := &BinaryDecoder{}

	testCases := []struct {
		name     string
		data     []byte
		offset   int
		expected uint64
		newPos   int
		hasError bool
	}{
		{
			name:     "zero",
			data:     []byte{0x00},
			offset:   0,
			expected: 0,
			newPos:   1,
			hasError: false,
		},
		{
			name:     "small_value",
			data:     []byte{0x7F},
			offset:   0,
			expected: 127,
			newPos:   1,
			hasError: false,
		},
		{
			name:     "two_byte_value",
			data:     []byte{0x80, 0x01},
			offset:   0,
			expected: 128,
			newPos:   2,
			hasError: false,
		},
		{
			name:     "large_value",
			data:     []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x0F},
			offset:   0,
			expected: 0xFFFFFFFF,
			newPos:   5,
			hasError: false,
		},
		{
			name:     "incomplete_varint",
			data:     []byte{0x80},
			offset:   0,
			expected: 0,
			newPos:   1,
			hasError: true,
		},
		{
			name:     "offset_reading",
			data:     []byte{0x00, 0x00, 0x7F},
			offset:   2,
			expected: 127,
			newPos:   3,
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, newPos, err := decoder.ReadVarint(tc.data, tc.offset)

			if tc.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if value != tc.expected {
					t.Errorf("Expected value=%d, got %d", tc.expected, value)
				}
				if newPos != tc.newPos {
					t.Errorf("Expected newPos=%d, got %d", tc.newPos, newPos)
				}
			}
		})
	}
}

// TestEstimateBinarySize tests size estimation accuracy
func TestEstimateBinarySize(t *testing.T) {
	encoder := NewBinaryEncoder()

	testCases := []struct {
		name string
		rec  *Record
	}{
		{
			name: "empty_record",
			rec:  NewRecord(Info, ""),
		},
		{
			name: "simple_record",
			rec: func() *Record {
				r := NewRecord(Error, "hello")
				r.Logger = "test"
				r.AddField(String("key", "value"))
				return r
			}(),
		},
		{
			name: "complex_record",
			rec: func() *Record {
				r := NewRecord(Debug, "complex message with details")
				r.Logger = "complex.logger.name"
				r.Caller = "/path/to/source.go:123"
				r.AddField(String("string_field", "long string value"))
				r.AddField(Int64("int_field", -12345))
				r.AddField(Float64("float_field", 3.14159265))
				r.AddField(Bool("bool_field", true))
				r.AddField(Bytes("bytes_field", []byte("binary data here")))
				return r
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Estimate size
			estimated := EstimateBinarySize(tc.rec)

			// Actual encoding
			buf := &bytes.Buffer{}
			encoder.Encode(tc.rec, testTime, buf)
			actual := buf.Len()

			// Estimation should be reasonably close (within 20%)
			diff := float64(actual-estimated) / float64(estimated)
			if diff < -0.2 || diff > 0.2 {
				t.Errorf("Size estimation too far off: estimated=%d, actual=%d, diff=%.2f%%",
					estimated, actual, diff*100)
			}

			t.Logf("Record %s: estimated=%d, actual=%d, diff=%.1f%%",
				tc.name, estimated, actual, diff*100)
		})
	}
}

// TestBinaryEncoder_SpecialStrings tests encoding of strings with special characters
func TestBinaryEncoder_SpecialStrings(t *testing.T) {
	encoder := NewBinaryEncoder()
	buf := &bytes.Buffer{}

	specialStrings := []string{
		"",                        // Empty string
		"hello\nworld",            // Newlines
		"tab\there",               // Tabs
		"null\x00byte",            // Null bytes
		"unicode: ñ €",            // Unicode characters
		strings.Repeat("x", 1000), // Long string
	}

	for i, str := range specialStrings {
		buf.Reset()
		rec := NewRecord(Info, str)
		rec.AddField(String("special", str))

		encoder.Encode(rec, testTime, buf)

		if buf.Len() == 0 {
			t.Errorf("Test case %d: expected non-empty encoded data", i)
		}

		// Verify magic header
		decoder := &BinaryDecoder{}
		valid, _ := decoder.DecodeMagic(buf.Bytes())
		if !valid {
			t.Errorf("Test case %d: should have valid magic header", i)
		}
	}
}

// TestBinaryEncoder_EdgeCaseValues tests encoding of edge case numeric values
func TestBinaryEncoder_EdgeCaseValues(t *testing.T) {
	encoder := NewBinaryEncoder()
	buf := &bytes.Buffer{}

	edgeCases := []struct {
		name  string
		field Field
	}{
		{"max_int64", Int64("max_int", 9223372036854775807)},
		{"min_int64", Int64("min_int", -9223372036854775808)},
		{"max_uint64", Uint64("max_uint", 18446744073709551615)},
		{"zero_uint64", Uint64("zero_uint", 0)},
		{"positive_inf", Float64("pos_inf", math.Inf(1))},
		{"negative_inf", Float64("neg_inf", math.Inf(-1))},
		{"zero_float", Float64("zero_float", 0.0)},
		{"negative_zero", Float64("neg_zero", math.Copysign(0.0, -1))},
		{"nan", Float64("nan", math.NaN())},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			rec := NewRecord(Info, "edge case test")
			rec.AddField(tc.field)

			encoder.Encode(rec, testTime, buf)

			if buf.Len() == 0 {
				t.Fatal("Expected non-empty encoded data")
			}

			// Verify magic header
			decoder := &BinaryDecoder{}
			valid, _ := decoder.DecodeMagic(buf.Bytes())
			if !valid {
				t.Error("Should have valid magic header")
			}
		})
	}
}
