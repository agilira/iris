// cpu_usage_demo.go: Demo showing CPU usage differences between idle strategies
//
// This demo creates loggers with different idle strategies and shows their
// CPU usage characteristics when the consumer has no work to process.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

//go:build ignore

package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/agilira/iris"
)

func main() {
	fmt.Println("=== Iris Idle Strategies CPU Usage Demo ===")
	fmt.Println("This demo shows the CPU usage characteristics of different idle strategies")
	fmt.Println("when the consumer loop has no work to process.")
	fmt.Println()

	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()

	fmt.Println("Strategy comparison:")
	fmt.Println("1. SpinningStrategy: Continuously checks for work (high CPU)")
	fmt.Println("2. EfficientStrategy: Sleeps when idle (low CPU)")
	fmt.Println("3. BalancedStrategy: Adaptive approach (starts high, reduces over time)")
	fmt.Println()

	// Create loggers with different strategies
	strategies := []struct {
		name     string
		strategy iris.IdleStrategy
		desc     string
	}{
		{
			"Spinning",
			iris.SpinningStrategy,
			"Minimal latency, high CPU usage",
		},
		{
			"Efficient",
			iris.EfficientStrategy,
			"Low CPU usage, higher latency",
		},
		{
			"Balanced",
			iris.BalancedStrategy,
			"Adaptive: starts high, reduces when idle",
		},
	}

	for _, s := range strategies {
		fmt.Printf("=== %s Strategy ===\n", s.name)
		fmt.Printf("Description: %s\n", s.desc)

		logger := createLogger(s.strategy)

		fmt.Println("Starting consumer loop (will run for 3 seconds)...")

		// Start the logger and measure
		logger.Start()

		// Let it run idle for 3 seconds
		start := time.Now()
		time.Sleep(3 * time.Second)
		elapsed := time.Since(start)

		// Log one message to show responsiveness
		logger.Info("Responsiveness test after idle period")
		logger.Sync()

		// Stop the logger
		logger.Close()

		fmt.Printf("Strategy ran for %v while idle\n", elapsed)
		fmt.Printf("Note: %s\n", getNote(s.name))
		fmt.Println()
	}

	fmt.Println("=== Summary ===")
	fmt.Println("• SpinningStrategy provides ultra-low latency but consumes ~100% CPU")
	fmt.Println("• EfficientStrategy minimizes CPU usage but has slightly higher latency")
	fmt.Println("• BalancedStrategy adapts to workload patterns automatically")
	fmt.Println()
	fmt.Println("The choice depends on your specific requirements:")
	fmt.Println("- High-frequency trading: Use SpinningStrategy")
	fmt.Println("- Resource-constrained environments: Use EfficientStrategy")
	fmt.Println("- General applications: Use BalancedStrategy (default)")
}

func createLogger(strategy iris.IdleStrategy) *iris.Logger {
	config := &iris.Config{
		Output:       iris.WrapWriter(os.Stdout),
		Encoder:      iris.NewJSONEncoder(),
		IdleStrategy: strategy,
		Capacity:     1024,
		BatchSize:    32,
		Level:        iris.Info,
	}

	logger, err := iris.New(*config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}

	return logger
}

func getNote(strategyName string) string {
	switch strategyName {
	case "Spinning":
		return "You should observe high CPU usage in system monitor"
	case "Efficient":
		return "You should observe low CPU usage in system monitor"
	case "Balanced":
		return "CPU usage should start high, then decrease over time"
	default:
		return "Monitor CPU usage in system tools"
	}
}
