// converter.go: Binary to JSON conversion logic
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
	"unsafe"
)

// BinaryToJSONConverter handles conversion from Iris binary format to JSON
type BinaryToJSONConverter struct {
	pretty       bool
	levelFilter  string
	validateOnly bool
}

// NewBinaryToJSONConverter creates a new converter
func NewBinaryToJSONConverter(pretty bool) *BinaryToJSONConverter {
	return &BinaryToJSONConverter{
		pretty: pretty,
	}
}

// NewBinaryToJSONConverterWithOptions creates a new converter with advanced options
func NewBinaryToJSONConverterWithOptions(pretty bool, levelFilter string, validateOnly bool) *BinaryToJSONConverter {
	return &BinaryToJSONConverter{
		pretty:       pretty,
		levelFilter:  levelFilter,
		validateOnly: validateOnly,
	}
}

// LogEntry represents a single log entry for JSON output
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Caller    *CallerInfo            `json:"caller,omitempty"`
	Fields    map[string]interface{} `json:",inline,omitempty"`
}

// CallerInfo represents caller information
type CallerInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function,omitempty"`
}

// BinaryEntry represents the binary log entry structure (must match iris binary_logger.go)
type BinaryEntry struct {
	Timestamp uint64 // Unix nanoseconds
	Level     uint8  // Log level
	MsgPtr    uintptr
	MsgLen    uint32
	FieldPtr  uintptr
	FieldCnt  uint16
}

// BinaryField represents a binary field (must match iris binary_logger.go)
type BinaryField struct {
	KeyPtr   uintptr
	KeyLen   uint16
	ValuePtr uintptr
	ValueLen uint32
	Type     uint8 // Field type
}

// Field types (must match iris field types)
const (
	FieldTypeString = iota
	FieldTypeInt
	FieldTypeFloat64
	FieldTypeBool
	FieldTypeTime
	FieldTypeError
	FieldTypeAny
)

// Level constants (must match iris levels)
const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

//lint:ignore U1000 levelNames is needed for future binary parsing implementation
var levelNames = map[uint8]string{
	DebugLevel:  "debug",
	InfoLevel:   "info",
	WarnLevel:   "warn",
	ErrorLevel:  "error",
	DPanicLevel: "dpanic",
	PanicLevel:  "panic",
	FatalLevel:  "fatal",
} // Used by getLevelName for binary parsing (TODO: implement parseLineAsBinary)

// Convert reads binary log entries and writes JSON
func (c *BinaryToJSONConverter) Convert(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	encoder := json.NewEncoder(output)

	if c.pretty {
		encoder.SetIndent("", "  ")
	}

	return c.processLines(scanner, encoder)
}

func (c *BinaryToJSONConverter) processLines(scanner *bufio.Scanner, encoder *json.Encoder) error {

	lineNum := 0
	processedEntries := 0
	filteredEntries := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		entry, err := c.parseLogEntry(line)
		if err != nil {
			return fmt.Errorf("line %d: failed to parse as binary or JSON: %v", lineNum, err)
		}

		// Apply level filter if specified
		if c.shouldFilterEntry(entry) {
			filteredEntries++
			continue
		}

		// If validate-only mode, don't output
		if c.validateOnly {
			processedEntries++
			continue
		}

		if err := encoder.Encode(entry); err != nil {
			return fmt.Errorf("line %d: failed to encode JSON: %v", lineNum, err)
		}
		processedEntries++
	}

	return c.handleScannerResults(scanner, processedEntries, filteredEntries)
}

// parseLineAsBinary parses a true binary log entry
func (c *BinaryToJSONConverter) parseLineAsBinary() (*LogEntry, error) {
	// TODO: Implement true binary parsing when iris binary format is finalized
	// For now, this is a placeholder that will return an error to fall back to JSON
	return nil, fmt.Errorf("binary parsing not yet implemented")
}

// parseLineAsJSON parses JSON log entries (fallback/current format)
func (c *BinaryToJSONConverter) parseLineAsJSON(data []byte) (*LogEntry, error) {
	var rawEntry map[string]interface{}
	if err := json.Unmarshal(data, &rawEntry); err != nil {
		return nil, err
	}

	entry := &LogEntry{
		Fields: make(map[string]interface{}),
	}

	// Extract standard fields
	c.extractStandardFields(rawEntry, entry)

	// Handle caller info if present
	c.extractCallerInfo(rawEntry, entry)

	// Add remaining fields
	for k, v := range rawEntry {
		entry.Fields[k] = v
	}

	// If no extra fields, remove the empty map
	if len(entry.Fields) == 0 {
		entry.Fields = nil
	}

	return entry, nil
}

func (c *BinaryToJSONConverter) extractStandardFields(rawEntry map[string]interface{}, entry *LogEntry) {
	if ts, ok := rawEntry["timestamp"].(string); ok {
		entry.Timestamp = ts
		delete(rawEntry, "timestamp")
	}

	if level, ok := rawEntry["level"].(string); ok {
		entry.Level = level
		delete(rawEntry, "level")
	}

	if msg, ok := rawEntry["message"].(string); ok {
		entry.Message = msg
		delete(rawEntry, "message")
	}
}

func (c *BinaryToJSONConverter) extractCallerInfo(rawEntry map[string]interface{}, entry *LogEntry) {
	caller, ok := rawEntry["caller"]
	if !ok {
		return
	}

	callerMap, ok := caller.(map[string]interface{})
	if !ok {
		return
	}

	entry.Caller = &CallerInfo{}
	if file, ok := callerMap["file"].(string); ok {
		entry.Caller.File = file
	}
	if line, ok := callerMap["line"].(float64); ok {
		entry.Caller.Line = int(line)
	}
	if function, ok := callerMap["function"].(string); ok {
		entry.Caller.Function = function
	}

	delete(rawEntry, "caller")
}

// Helper functions for binary parsing (when implemented)

// getLevelName converts level code to string - for binary parsing implementation
func getLevelName(level uint8) string {
	if name, ok := levelNames[level]; ok {
		return name
	}
	return fmt.Sprintf("level_%d", level)
}

// GetAllLevelNames returns all available level names (exported for CLI validation)
func GetAllLevelNames() []string {
	names := make([]string, 0, len(levelNames))
	for _, name := range levelNames {
		names = append(names, name)
	}
	return names
}

// formatTimestamp converts nanoseconds to RFC3339 - for binary parsing implementation
func formatTimestamp(nanos uint64) string {
	// Timestamp values are typically safe for conversion
	// Only convert if the value is within safe range for int64
	if nanos <= 1<<63-1 {
		t := time.Unix(0, int64(nanos))
		return t.Format(time.RFC3339Nano)
	}
	// For very large timestamps, return a string representation
	return fmt.Sprintf("timestamp_overflow_%d", nanos)
}

// unsafeString converts a pointer and length to string (DANGEROUS - use carefully)
// For binary parsing implementation - converts binary field data to strings
func unsafeString(ptr uintptr, length uint32) string {
	if ptr == 0 || length == 0 {
		return ""
	}
	// #nosec G103 - unsafe.Pointer required for zero-allocation string conversion in export tool
	return *(*string)(unsafe.Pointer(&struct {
		data uintptr
		len  int
	}{ptr, int(length)}))
}

// parseLogEntry attempts to parse a log entry as binary first, then JSON
func (c *BinaryToJSONConverter) parseLogEntry(line []byte) (*LogEntry, error) {
	// For now, we'll handle the case where the binary logger is actually
	// writing JSON (as we saw in tests). Later we'll implement true binary parsing.
	entry, err := c.parseLineAsBinary()
	if err != nil {
		// If binary parsing fails, try JSON passthrough
		return c.parseLineAsJSON(line)
	}
	return entry, nil
}

// shouldFilterEntry checks if an entry should be filtered based on level
func (c *BinaryToJSONConverter) shouldFilterEntry(entry *LogEntry) bool {
	return c.levelFilter != "" && entry.Level != c.levelFilter
}

// handleScannerResults processes final scanner results and prints stats
func (c *BinaryToJSONConverter) handleScannerResults(scanner *bufio.Scanner, processedEntries, filteredEntries int) error {
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	// Print validation summary to stderr if in validate mode
	if c.validateOnly {
		fmt.Fprintf(os.Stderr, "Validation complete: %d entries processed, %d filtered\n", processedEntries, filteredEntries)
	}

	return nil
}

// init ensures future binary parsing functions are kept during compilation
func init() {
	// Prevent dead code elimination of functions needed for future binary parsing
	_ = getLevelName(InfoLevel) // Example usage to keep levelNames alive
	_ = formatTimestamp(0)
	_ = unsafeString(0, 0)
	_ = GetAllLevelNames() // Keeps levelNames definitively alive
}
