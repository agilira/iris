// example_idle_strategies.go: Example demonstrating configurable idle strategies
//
// This example shows how to use different idle strategies to control CPU usage
// when the logger consumer has no work to process. Each strategy provides
// different trade-offs between latency and CPU consumption.

//go:build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/agilira/iris"
)

func main() {
	fmt.Println("=== Iris Idle Strategies Demo ===")
	fmt.Println()

	// Example 1: Ultra-low latency with spinning strategy
	fmt.Println("1. SpinningIdleStrategy - Ultra-low latency, high CPU usage")
	demoStrategy("Spinning", iris.NewSpinningIdleStrategy())

	// Example 2: CPU-efficient with sleeping strategy
	fmt.Println("\n2. SleepingIdleStrategy - Low CPU usage, higher latency")
	demoStrategy("Sleeping", iris.NewSleepingIdleStrategy(time.Millisecond, 0))

	// Example 3: Balanced with yielding strategy
	fmt.Println("\n3. YieldingIdleStrategy - Moderate CPU usage")
	demoStrategy("Yielding", iris.NewYieldingIdleStrategy(1000))

	// Example 4: Minimal CPU with channel strategy
	fmt.Println("\n4. ChannelIdleStrategy - Minimal CPU usage")
	demoStrategy("Channel", iris.NewChannelIdleStrategy(100*time.Millisecond))

	// Example 5: Adaptive with progressive strategy (default)
	fmt.Println("\n5. ProgressiveIdleStrategy - Adaptive (default)")
	demoStrategy("Progressive", iris.NewProgressiveIdleStrategy())

	// Example 6: Using predefined strategies
	fmt.Println("\n6. Predefined Strategies:")

	strategies := []struct {
		name     string
		strategy iris.IdleStrategy
		desc     string
	}{
		{"SpinningStrategy", iris.SpinningStrategy, "Ultra-low latency"},
		{"BalancedStrategy", iris.BalancedStrategy, "Good for most use cases"},
		{"EfficientStrategy", iris.EfficientStrategy, "Low CPU usage"},
		{"HybridStrategy", iris.HybridStrategy, "Spin then sleep"},
	}

	for _, s := range strategies {
		fmt.Printf("   - %s: %s\n", s.name, s.desc)
		demoStrategy(s.name, s.strategy)
	}

	fmt.Println("\n=== Performance Comparison ===")
	performanceComparison()
}

func demoStrategy(name string, strategy iris.IdleStrategy) {
	fmt.Printf("   Testing %s strategy...", name)

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
		fmt.Printf(" ERROR: %v\n", err)
		return
	}
	defer logger.Close()

	logger.Start()

	// Log a few messages
	start := time.Now()
	for i := 0; i < 3; i++ {
		logger.Info("Demo message",
			iris.String("strategy", name),
			iris.Int("message_id", i),
		)
	}

	// Ensure all messages are processed
	logger.Sync()

	elapsed := time.Since(start)
	fmt.Printf(" completed in %v\n", elapsed)
}

func performanceComparison() {
	fmt.Println("\nMeasuring idle CPU behavior (5 seconds each):")

	strategies := []struct {
		name     string
		strategy iris.IdleStrategy
	}{
		{"Spinning", iris.NewSpinningIdleStrategy()},
		{"Efficient", iris.EfficientStrategy},
		{"Balanced", iris.BalancedStrategy},
	}

	for _, s := range strategies {
		fmt.Printf("\n%s Strategy:\n", s.name)
		measureIdleBehavior(s.strategy)
	}
}

func measureIdleBehavior(strategy iris.IdleStrategy) {
	config := &iris.Config{
		Output:       iris.WrapWriter(os.Stdout),
		Encoder:      iris.NewJSONEncoder(),
		IdleStrategy: strategy,
		Capacity:     64,
		BatchSize:    16,
		Level:        iris.Error, // High level to avoid processing during measurement
	}

	logger, err := iris.New(*config)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
		return
	}
	defer logger.Close()

	logger.Start()

	fmt.Println("   Consumer started, measuring idle behavior...")
	fmt.Println("   (Note: Spinning strategy will use ~100% CPU, others should be lower)")

	// Let it idle for a few seconds
	time.Sleep(2 * time.Second)

	// Log one message to show it's still responsive
	logger.Error("Responsiveness test")
	logger.Sync()

	fmt.Println("   Strategy is responsive after idle period")
}
