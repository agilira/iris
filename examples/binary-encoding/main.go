// examples/binary-encoding/main.go - Example demonstrating binary encoder usage
//
// This example shows how to use the Iris Binary Encoder for:
// - Ultra-fast log encoding with minimal allocations
// - Compact binary format for efficient storage/transmission
// - Configuration options for different use cases
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/agilira/iris"
)

func main() {
	fmt.Println("Iris Binary Encoder Example")
	fmt.Println("==============================")

	// Create a logger with binary encoder
	binaryLogger, err := createBinaryLogger()
	if err != nil {
		panic(err)
	}

	// Create a compact binary logger for bandwidth-constrained scenarios
	compactLogger, err := createCompactBinaryLogger()
	if err != nil {
		panic(err)
	}

	// Log various message types
	demonstrateBasicLogging(binaryLogger)
	demonstrateFieldTypes(binaryLogger)
	demonstrateCompactMode(compactLogger)
	demonstratePerformance()
}

// createBinaryLogger creates a logger using the binary encoder
func createBinaryLogger() (*iris.Logger, error) {
	encoder := iris.NewBinaryEncoder()

	config := iris.Config{
		Level:              iris.Debug,
		Output:             iris.WrapWriter(os.Stdout),
		Encoder:            encoder,
		Capacity:           64,
		BatchSize:          1,
		Architecture:       iris.SingleRing,
		NumRings:           1,
		BackpressurePolicy: 0,
	}

	return iris.New(config)
}

// createCompactBinaryLogger creates a logger using compact binary encoding
func createCompactBinaryLogger() (*iris.Logger, error) {
	encoder := iris.NewCompactBinaryEncoder()

	config := iris.Config{
		Level:              iris.Debug,
		Output:             iris.WrapWriter(os.Stdout),
		Encoder:            encoder,
		Capacity:           64,
		BatchSize:          1,
		Architecture:       iris.SingleRing,
		NumRings:           1,
		BackpressurePolicy: 0,
	}

	return iris.New(config)
}

// demonstrateBasicLogging shows basic binary logging
func demonstrateBasicLogging(logger *iris.Logger) {
	fmt.Println("\n📝 Basic Binary Logging:")
	fmt.Println("--")

	logger.Named("binary.demo").Info("Binary encoding demonstration")
	logger.Named("binary.demo").Info("Structured logging with binary format",
		iris.String("format", "binary"),
		iris.Bool("compact", true),
		iris.Int("version", 1),
	)
}

// demonstrateFieldTypes shows all supported field types in binary format
func demonstrateFieldTypes(logger *iris.Logger) {
	fmt.Println("\n🔢 All Field Types in Binary:")
	fmt.Println("--")

	logger.Named("fields.demo").Info("All field types encoded in binary format",
		iris.String("string_field", "text value"),
		iris.Int64("int64_field", -123456789),
		iris.Uint64("uint64_field", 987654321),
		iris.Float64("float64_field", 3.14159265359),
		iris.Bool("bool_field", true),
		iris.Dur("duration_field", 5*time.Minute),
		iris.Time("time_field", time.Now()),
		iris.Bytes("bytes_field", []byte("binary data")),
	)
}

// demonstrateCompactMode shows compact binary encoding
func demonstrateCompactMode(logger *iris.Logger) {
	fmt.Println("\n📦 Compact Binary Mode:")
	fmt.Println("--")

	logger.Info("Compact binary encoding (minimal size)")
	logger.Info("Space-optimized binary format",
		iris.String("mode", "compact"),
		iris.Int("saving", 40), // percentage space saving
	)
}

// demonstratePerformance shows performance characteristics
func demonstratePerformance() {
	fmt.Println("\n⚡ Performance Demonstration:")
	fmt.Println("--")

	// Create encoders for comparison
	binaryEncoder := iris.NewBinaryEncoder()
	jsonEncoder := iris.NewJSONEncoder()
	textEncoder := iris.NewTextEncoder()

	// Create test record
	record := iris.NewRecord(iris.Info, "performance test message")
	record.Logger = "perf.test"
	record.AddField(iris.String("test", "value"))
	record.AddField(iris.Int("number", 42))
	record.AddField(iris.Bool("flag", true))

	now := time.Now()

	// Test encoding performance
	fmt.Printf("Testing encoding performance for 100,000 operations...\n")

	// Binary encoding
	binaryBuf := &bytes.Buffer{}
	start := time.Now()
	for i := 0; i < 100000; i++ {
		binaryBuf.Reset()
		binaryEncoder.Encode(record, now, binaryBuf)
	}
	binaryTime := time.Since(start)
	binarySize := binaryBuf.Len()

	// JSON encoding
	jsonBuf := &bytes.Buffer{}
	start = time.Now()
	for i := 0; i < 100000; i++ {
		jsonBuf.Reset()
		jsonEncoder.Encode(record, now, jsonBuf)
	}
	jsonTime := time.Since(start)
	jsonSize := jsonBuf.Len()

	// Text encoding
	textBuf := &bytes.Buffer{}
	start = time.Now()
	for i := 0; i < 100000; i++ {
		textBuf.Reset()
		textEncoder.Encode(record, now, textBuf)
	}
	textTime := time.Since(start)
	textSize := textBuf.Len()

	// Display results
	fmt.Printf("\nPerformance Results:\n")
	fmt.Printf("%-10s | %-12s | %-10s | %-15s\n", "Format", "Time", "Size", "Speed Factor")
	fmt.Printf("%-10s | %-12s | %-10s | %-15s\n", "----------", "------------", "----------", "---------------")
	fmt.Printf("%-10s | %-12v | %-10d | %-15s\n", "Binary", binaryTime, binarySize, "1.0x (baseline)")
	fmt.Printf("%-10s | %-12v | %-10d | %.1fx slower\n", "JSON", jsonTime, jsonSize, float64(jsonTime)/float64(binaryTime))
	fmt.Printf("%-10s | %-12v | %-10d | %.1fx slower\n", "Text", textTime, textSize, float64(textTime)/float64(binaryTime))

	fmt.Printf("\nSize Comparison:\n")
	fmt.Printf("Binary: %d bytes (baseline)\n", binarySize)
	fmt.Printf("JSON:   %d bytes (%.1f%% larger)\n", jsonSize, float64(jsonSize-binarySize)/float64(binarySize)*100)
	fmt.Printf("Text:   %d bytes (%.1f%% larger)\n", textSize, float64(textSize-binarySize)/float64(binarySize)*100)

	fmt.Printf("\n✅ Binary encoding advantages:\n")
	fmt.Printf("   • %.1fx faster than JSON\n", float64(jsonTime)/float64(binaryTime))
	fmt.Printf("   • %.1fx faster than Text\n", float64(textTime)/float64(binaryTime))
	fmt.Printf("   • %d%% smaller than JSON\n", int((float64(jsonSize-binarySize)/float64(jsonSize))*100))
	fmt.Printf("   • %d%% smaller than Text\n", int((float64(textSize-binarySize)/float64(textSize))*100))
	fmt.Printf("   • Zero allocations for simple records\n")
	fmt.Printf("   • Self-describing format with magic header\n")
}
