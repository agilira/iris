package main

import (
	"fmt"
	"os"
	"time"

	"github.com/agilira/iris"
)

func main() {
	fmt.Println("=== Iris Idle Strategy Configuration Demo ===")

	// Demo 1: Configuration via JSON file
	fmt.Println("\n1. Loading configuration from JSON file...")
	config, err := iris.LoadConfigFromJSON("config_idle_strategy.json")
	if err != nil {
		fmt.Printf("Error loading JSON config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded idle strategy: %s\n", config.IdleStrategy.String())

	logger1, err := iris.New(*config)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer logger1.Close()

	logger1.Info("JSON configured logger with sleeping idle strategy")

	// Demo 2: Configuration via environment variables
	fmt.Println("\n2. Configuration via environment variables...")
	os.Setenv("IRIS_IDLE_STRATEGY", "yielding")
	os.Setenv("IRIS_LEVEL", "debug")
	defer func() {
		os.Unsetenv("IRIS_IDLE_STRATEGY")
		os.Unsetenv("IRIS_LEVEL")
	}()

	envConfig, err := iris.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("Error loading env config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Environment idle strategy: %s\n", envConfig.IdleStrategy.String())

	logger2, err := iris.New(*envConfig)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer logger2.Close()

	logger2.Debug("Environment configured logger with yielding idle strategy")

	// Demo 3: Programmatic configuration with different strategies
	fmt.Println("\n3. Programmatic configuration with different strategies...")

	strategies := []struct {
		name     string
		strategy iris.IdleStrategy
	}{
		{"Spinning", iris.SpinningStrategy},
		{"Efficient", iris.EfficientStrategy},
		{"Balanced", iris.BalancedStrategy},
		{"Hybrid", iris.HybridStrategy},
	}

	for _, s := range strategies {
		fmt.Printf("Creating logger with %s strategy (%s)...\n", s.name, s.strategy.String())

		config := iris.Config{
			Level:        iris.Info,
			IdleStrategy: s.strategy,
		}

		logger, err := iris.New(config)
		if err != nil {
			fmt.Printf("Error creating %s logger: %v\n", s.name, err)
			continue
		}

		logger.Info(fmt.Sprintf("Message from %s logger", s.name))
		logger.Close()
	}

	// Demo 4: Multi-source configuration priority
	fmt.Println("\n4. Multi-source configuration (JSON + Environment)...")

	// JSON has "sleeping", environment has "channel" - environment should override
	os.Setenv("IRIS_IDLE_STRATEGY", "channel")
	defer os.Unsetenv("IRIS_IDLE_STRATEGY")

	multiConfig, err := iris.LoadConfigMultiSource("config_idle_strategy.json")
	if err != nil {
		fmt.Printf("Error loading multi-source config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Multi-source result (env should override JSON): %s\n", multiConfig.IdleStrategy.String())

	logger4, err := iris.New(*multiConfig)
	if err != nil {
		fmt.Printf("Error creating multi-source logger: %v\n", err)
		os.Exit(1)
	}
	defer logger4.Close()

	logger4.Info("Multi-source configured logger")

	// Demo 5: Performance comparison (brief)
	fmt.Println("\n5. Brief performance demonstration...")

	performanceTest := func(name string, strategy iris.IdleStrategy) {
		config := iris.Config{
			Level:        iris.Info,
			IdleStrategy: strategy,
			Capacity:     1024,
		}

		logger, err := iris.New(config)
		if err != nil {
			fmt.Printf("Error creating %s logger: %v\n", name, err)
			return
		}
		defer logger.Close()

		start := time.Now()
		for i := 0; i < 1000; i++ {
			logger.Info("Performance test message", iris.Int("iteration", i))
		}
		duration := time.Since(start)

		fmt.Printf("%s strategy: 1000 messages in %v\n", name, duration)
	}

	performanceTest("Spinning", iris.SpinningStrategy)
	performanceTest("Efficient", iris.EfficientStrategy)
	performanceTest("Balanced", iris.BalancedStrategy)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("The idle strategy controls how the consumer loop behaves when no messages are available.")
	fmt.Println("Different strategies offer different CPU usage vs latency trade-offs:")
	fmt.Println("- Spinning: Lowest latency, highest CPU usage")
	fmt.Println("- Sleeping: Low CPU usage, higher latency")
	fmt.Println("- Yielding: Good balance, yields CPU to other goroutines")
	fmt.Println("- Channel: Event-driven, very low CPU when idle")
	fmt.Println("- Progressive/Balanced: Adaptive, starts fast then backs off")
}
