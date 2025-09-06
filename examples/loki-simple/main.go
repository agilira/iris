// simple_loki_example.go: Simple Loki integration example
//
// This is a basic example showing how to get started with iris + Loki quickly.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"

	"github.com/agilira/iris"
	// Note: To use Loki writer, you need to install the separate package:
	// go get github.com/agilira/iris-writer-loki
	// and import: "github.com/agilira/iris-writer-loki"
)

func main() {
	fmt.Println("🚀 Starting iris + Loki simple example...")
	fmt.Println("📋 Note: This example requires the iris-writer-loki package")
	fmt.Println("   Install with: go get github.com/agilira/iris-writer-loki")
	fmt.Println("   Then uncomment the code below and update imports")

	// Temporary simple logger example while Loki writer is in separate package
	logger, err := iris.New(iris.Config{
		Level:   iris.Info,
		Encoder: iris.NewConsoleEncoder(),
	})
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to create logger: %v", err))
	}
	defer logger.Close()
	logger.Start()

	logger.Info("🎉 Basic example running - see commented code for Loki integration")

	/* TODO: Uncomment when iris-writer-loki is installed

	// Step 1: Configure Loki writer
	lokiConfig := loki.WriterConfig{  // Note: this should be loki.WriterConfig, not iris.LokiWriterConfig
		Endpoint: "http://localhost:3100/loki/api/v1/push", // Your Loki endpoint
		Labels: map[string]string{
			"service": "simple-example",
			"env":     "development",
		},
		BatchSize:     100,             // Small batches for this example
		FlushInterval: 2 * time.Second, // Flush every 2 seconds
	}

	// Step 2: Create Loki writer
	lokiWriter, err := loki.NewWriter(lokiConfig)  // Note: this should be loki.NewWriter, not iris.WrapLokiWriter
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to create Loki writer: %v", err))
	}
	defer lokiWriter.Close()

	// Step 3: Create iris logger
	logger, err := iris.New(iris.Config{
		Level:   iris.Info,
		Output:  lokiWriter,
		Encoder: iris.NewJSONEncoder(),
	})
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to create logger: %v", err))
	}
	defer logger.Close()

	// Step 4: Start the logger (important!)
	logger.Start()

	fmt.Println("✅ Logger created and connected to Loki")

	// Step 5: Log some messages
	logger.Info("🎉 Application started",
		iris.Str("version", "1.0.0"),
		iris.Str("component", "main"))

	logger.Info("👤 User action",
		iris.Str("user_id", "user_123"),
		iris.Str("action", "login"),
		iris.Int64("session_id", 12345))

	logger.Warn("⚠️ Warning occurred",
		iris.Str("component", "database"),
		iris.Str("warning", "Connection pool getting full"),
		iris.Int64("active_connections", 95))

	logger.Error("❌ Error occurred",
		iris.Str("component", "payment"),
		iris.Str("error", "Credit card validation failed"),
		iris.Str("error_code", "CC_INVALID"))

	// Step 6: Log some structured data
	for i := 0; i < 20; i++ {
		logger.Info("📊 Processing request",
			iris.Int64("request_id", int64(1000+i)),
			iris.Str("method", "GET"),
			iris.Str("path", "/api/users"),
			iris.Dur("response_time", time.Duration(50+i*5)*time.Millisecond),
			iris.Int64("status_code", 200))
	}

	fmt.Println("📝 Logged 24 messages")

	// Step 7: Wait a bit to let batches flush
	fmt.Println("⏳ Waiting for batches to flush to Loki...")
	time.Sleep(5 * time.Second)

	// Step 8: Get statistics
	stats := lokiWriter.Stats()
	fmt.Printf("📊 Statistics:\n")
	fmt.Printf("   - Entries written: %d\n", stats["entries_written"])
	fmt.Printf("   - Batches sent: %d\n", stats["batches_sent"])
	fmt.Printf("   - Errors: %d\n", stats["errors"])
	fmt.Printf("   - Dropped entries: %d\n", stats["entries_dropped"])

	// Step 9: Final sync to ensure all logs are sent
	if err := logger.Sync(); err != nil {
		fmt.Printf("Warning: Failed to sync logger: %v\n", err)
	}
	fmt.Println("✅ All done! Check your Loki for the logs.")

	// Pro tip: In production, you might want to use MultiWriter
	// to send logs to both console AND Loki:
	//
	// output := iris.MultiWriter(os.Stdout, lokiWriter)
	// logger, err := iris.New(iris.Config{
	//     Output: output,
	//     ...
	// })

	*/ // End of commented Loki integration code
}
