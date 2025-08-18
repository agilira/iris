// example_test.go: Example usage of Iris logger
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris_test

import (
	"time"

	"github.com/agilira/iris"
)

func ExampleNew_basic() {
	// Create a new logger
	logger, err := iris.New(iris.Config{
		Level:      iris.InfoLevel,
		Writer:     iris.StdoutWriter,
		BufferSize: 1024,
		BatchSize:  32,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// Basic logging
	logger.Info("Application started")
	logger.Info("User logged in", iris.Str("user", "john"), iris.Int("id", 123))
	logger.Warn("High memory usage", iris.Float("usage", 0.85))
	logger.Error("Database connection failed", iris.Err(err))

	// Structured logging with multiple fields
	logger.Info("Request processed",
		iris.Str("method", "GET"),
		iris.Str("path", "/api/users"),
		iris.Int("status", 200),
		iris.Duration("duration", 50*time.Millisecond),
	)
}

func ExampleNew_performance() {
	logger, _ := iris.New(iris.Config{
		Level:      iris.InfoLevel,
		BufferSize: 4096, // Larger buffer for high throughput
		BatchSize:  128,  // Larger batches for better performance
	})
	defer logger.Close()

	// High-frequency logging example
	for i := 0; i < 1000000; i++ {
		logger.Info("High frequency log",
			iris.Int("iteration", i),
			iris.Str("operation", "benchmark"),
		)
	}
}
