// field_extended_test.go: Tests for extended field types (Dur, TimeField, Bytes, Err)
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

// TestExtendedFieldTypes tests the new field types (Dur, TimeField, Bytes, Err)
func TestExtendedFieldTypes(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Start()

	// Test Duration field
	duration := 5 * time.Second
	logger.Info("duration test", Dur("duration", duration))

	// Test Time field
	testTime := time.Date(2025, 8, 22, 12, 0, 0, 0, time.UTC)
	logger.Info("time test", TimeField("timestamp", testTime))

	// Test Bytes field
	testBytes := []byte("hello world")
	logger.Info("bytes test", Bytes("data", testBytes))

	// Test Error field
	testError := errors.New("test error")
	logger.Info("error test", Err(testError))

	// Test Named Error field
	logger.Info("named error test", NamedErr("custom_error", testError))

	// Test nil error (should be empty string)
	logger.Info("nil error test", Err(nil))

	// Give time for async processing
	time.Sleep(50 * time.Millisecond)

	logger.Sync()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 6 {
		t.Errorf("Expected 6 log lines, got %d", len(lines))
	}

	// Test Duration encoding
	var durEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &durEntry); err != nil {
		t.Errorf("Failed to parse duration log entry: %v", err)
	}
	if durEntry["duration"] != float64(5000000000) { // 5 seconds in nanoseconds
		t.Errorf("Expected duration 5000000000, got %v", durEntry["duration"])
	}

	// Test Time encoding
	var timeEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[1]), &timeEntry); err != nil {
		t.Errorf("Failed to parse time log entry: %v", err)
	}
	expectedTimeStr := "2025-08-22T12:00:00Z"
	if timeEntry["timestamp"] != expectedTimeStr {
		t.Errorf("Expected time %s, got %v", expectedTimeStr, timeEntry["timestamp"])
	}

	// Test Bytes encoding (should be JSON array of integers)
	var bytesEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[2]), &bytesEntry); err != nil {
		t.Errorf("Failed to parse bytes log entry: %v", err)
	}
	// The bytes "hello world" should be encoded as array of byte values
	expectedBytes := []interface{}{104.0, 101.0, 108.0, 108.0, 111.0, 32.0, 119.0, 111.0, 114.0, 108.0, 100.0}
	bytesData, ok := bytesEntry["data"].([]interface{})
	if !ok {
		t.Errorf("Expected bytes data to be array, got %T", bytesEntry["data"])
	} else if len(bytesData) != len(expectedBytes) {
		t.Errorf("Expected %d bytes, got %d", len(expectedBytes), len(bytesData))
	}

	// Test Error encoding
	var errorEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[3]), &errorEntry); err != nil {
		t.Errorf("Failed to parse error log entry: %v", err)
	}
	if errorEntry["error"] != "test error" {
		t.Errorf("Expected error 'test error', got %v", errorEntry["error"])
	}

	// Test Named Error encoding
	var namedErrorEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[4]), &namedErrorEntry); err != nil {
		t.Errorf("Failed to parse named error log entry: %v", err)
	}
	if namedErrorEntry["custom_error"] != "test error" {
		t.Errorf("Expected custom_error 'test error', got %v", namedErrorEntry["custom_error"])
	}

	// Test nil error encoding (should be empty string)
	var nilErrorEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[5]), &nilErrorEntry); err != nil {
		t.Errorf("Failed to parse nil error log entry: %v", err)
	}
	if nilErrorEntry["error"] != "" {
		t.Errorf("Expected nil error to be empty string, got %v", nilErrorEntry["error"])
	}
}

// TestFieldValueAccessors tests the value accessor methods for extended field types
func TestFieldValueAccessors(t *testing.T) {
	// Test Duration field accessor
	duration := 2 * time.Minute
	durField := Dur("test_dur", duration)

	if durField.DurationValue() != duration {
		t.Errorf("Expected duration %v, got %v", duration, durField.DurationValue())
	}

	// Test Time field accessor
	testTime := time.Now()
	timeField := TimeField("test_time", testTime)

	retrievedTime := timeField.TimeValue()
	if !retrievedTime.Equal(testTime) {
		t.Errorf("Expected time %v, got %v", testTime, retrievedTime)
	}

	// Test Bytes field accessor
	testBytes := []byte("test data")
	bytesField := Bytes("test_bytes", testBytes)

	retrievedBytes := bytesField.BytesValue()
	if string(retrievedBytes) != string(testBytes) {
		t.Errorf("Expected bytes %s, got %s", string(testBytes), string(retrievedBytes))
	}

	// Test wrong type accessors return zero values
	strField := Str("test_str", "hello")

	if strField.DurationValue() != 0 {
		t.Errorf("Expected zero duration for string field, got %v", strField.DurationValue())
	}

	if !strField.TimeValue().IsZero() {
		t.Errorf("Expected zero time for string field, got %v", strField.TimeValue())
	}

	if strField.BytesValue() != nil {
		t.Errorf("Expected nil bytes for string field, got %v", strField.BytesValue())
	}
}
