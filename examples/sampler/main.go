// examples/sampler/main.go: Example demonstrating iris log sampling
//
// This example shows how to use the TokenBucketSampler to limit
// the rate of log messages in high-volume scenarios.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"time"

	"github.com/agilira/iris"
)

func main() {
	fmt.Println("🌸 Iris Log Sampling Example")
	fmt.Println("==============================")

	// Create a token bucket sampler that allows:
	// - Burst capacity: 5 messages
	// - Refill rate: 2 messages per second
	sampler := iris.NewTokenBucketSampler(5, 2, time.Second)

	// Create logger with sampler
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Encoder: iris.NewJSONEncoder(),
		Sampler: sampler,
	})
	if err != nil {
		panic(err)
	}

	logger.Start()
	defer logger.Close()

	fmt.Println("\n📊 Phase 1: Burst Test (sending 10 messages quickly)")
	fmt.Println("Expected: First 5 messages should pass, rest should be dropped")

	// Send 10 messages quickly - only first 5 should pass
	for i := 0; i < 10; i++ {
		logger.Info("Burst message", iris.Int("id", i))
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n⏰ Phase 2: Sustained Rate Test")
	fmt.Println("Expected: Messages should be throttled to ~2 per second")

	// Send messages at regular intervals to demonstrate sustained rate
	for i := 0; i < 8; i++ {
		logger.Info("Sustained message", iris.Int("id", i))
		time.Sleep(400 * time.Millisecond) // Slightly faster than refill rate
	}

	fmt.Println("\n🔄 Phase 3: Recovery Test")
	fmt.Println("Expected: After waiting, burst capacity should be restored")

	// Wait for bucket to refill
	time.Sleep(3 * time.Second)

	// Send another burst
	for i := 0; i < 6; i++ {
		logger.Info("Recovery message", iris.Int("id", i))
	}

	// Wait for final processing
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n✅ Sampling example completed!")
	fmt.Println("Note: Check the JSON log output above to see which messages passed through the sampler.")
}
