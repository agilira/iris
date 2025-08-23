package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/agilira/iris"
)

func main() {
	fmt.Println("=== Iris Config Loader Demo: Backpressure Policies ===")
	fmt.Println()

	// Demo 1: Load from JSON file
	fmt.Println("1. Loading high-performance configuration from JSON...")
	config1, err := iris.LoadConfigFromJSON("../configs/high_performance.json")
	if err != nil {
		log.Printf("Error loading high-performance config: %v", err)
	} else {
		logger1, err := iris.New(*config1)
		if err != nil {
			log.Printf("Error creating logger: %v", err)
		} else {
			fmt.Printf("   ✓ Created logger with policy: %v\n", config1.BackpressurePolicy)
			logger1.Info("High-performance logger initialized",
				iris.String("policy", config1.BackpressurePolicy.String()),
				iris.Int64("capacity", config1.Capacity))
			logger1.Sync()
		}
	}

	fmt.Println()

	// Demo 2: Load reliable configuration
	fmt.Println("2. Loading reliable configuration from JSON...")
	config2, err := iris.LoadConfigFromJSON("../configs/reliable.json")
	if err != nil {
		log.Printf("Error loading reliable config: %v", err)
	} else {
		logger2, err := iris.New(*config2)
		if err != nil {
			log.Printf("Error creating logger: %v", err)
		} else {
			fmt.Printf("   ✓ Created logger with policy: %v\n", config2.BackpressurePolicy)
			logger2.Info("Reliable logger initialized",
				iris.String("policy", config2.BackpressurePolicy.String()),
				iris.Int64("capacity", config2.Capacity))
			logger2.Sync()
		}
	}

	fmt.Println()

	// Demo 3: Environment variable override
	fmt.Println("3. Testing environment variable override...")
	os.Setenv("IRIS_BACKPRESSURE_POLICY", "block_on_full")
	os.Setenv("IRIS_LEVEL", "debug")
	os.Setenv("IRIS_NAME", "env-override-logger")

	config3, err := iris.LoadConfigFromEnv()
	if err != nil {
		log.Printf("Error loading from environment: %v", err)
	} else {
		logger3, err := iris.New(*config3)
		if err != nil {
			log.Printf("Error creating logger: %v", err)
		} else {
			fmt.Printf("   ✓ Environment override: policy=%v, name=%s\n",
				config3.BackpressurePolicy, config3.Name)
			logger3.Debug("Environment-configured logger active",
				iris.String("source", "environment"),
				iris.String("policy", config3.BackpressurePolicy.String()))
			logger3.Sync()
		}
	}

	fmt.Println()

	// Demo 4: Multi-source configuration (JSON + Environment)
	fmt.Println("4. Multi-source configuration (JSON base + Environment override)...")
	// Environment variables are already set from demo 3
	config4, err := iris.LoadConfigMultiSource("../configs/high_performance.json")
	if err != nil {
		log.Printf("Error loading multi-source config: %v", err)
	} else {
		logger4, err := iris.New(*config4)
		if err != nil {
			log.Printf("Error creating logger: %v", err)
		} else {
			fmt.Printf("   ✓ Multi-source: JSON base + ENV override\n")
			fmt.Printf("   ✓ Final policy: %v (from environment)\n", config4.BackpressurePolicy)
			fmt.Printf("   ✓ Final name: %s (from environment)\n", config4.Name)
			logger4.Info("Multi-source configuration active",
				iris.String("json_base", "high_performance.json"),
				iris.String("env_override", "true"),
				iris.String("final_policy", config4.BackpressurePolicy.String()))
			logger4.Sync()
		}
	}

	fmt.Println()

	// Demo 5: Performance comparison
	fmt.Println("5. Performance comparison between policies...")

	// High-performance logger (DropOnFull)
	hpConfig, _ := iris.LoadConfigFromJSON("../configs/high_performance.json")
	hpLogger, _ := iris.New(*hpConfig)

	// Reliable logger (BlockOnFull)
	reliableConfig, _ := iris.LoadConfigFromJSON("../configs/reliable.json")
	reliableLogger, _ := iris.New(*reliableConfig)

	messageCount := 1000

	// Test high-performance logger
	start := time.Now()
	for i := 0; i < messageCount; i++ {
		hpLogger.Info("Performance test message", iris.Int("iteration", i))
	}
	hpLogger.Sync()
	hpDuration := time.Since(start)

	// Test reliable logger
	start = time.Now()
	for i := 0; i < messageCount; i++ {
		reliableLogger.Info("Performance test message", iris.Int("iteration", i))
	}
	reliableLogger.Sync()
	reliableDuration := time.Since(start)

	fmt.Printf("   ✓ High-performance (DropOnFull): %d messages in %v\n", messageCount, hpDuration)
	fmt.Printf("   ✓ Reliable (BlockOnFull): %d messages in %v\n", messageCount, reliableDuration)
	fmt.Printf("   ✓ Performance ratio: %.2fx\n", float64(reliableDuration)/float64(hpDuration))

	// Clean up environment
	os.Unsetenv("IRIS_BACKPRESSURE_POLICY")
	os.Unsetenv("IRIS_LEVEL")
	os.Unsetenv("IRIS_NAME")

	fmt.Println()
	fmt.Println("=== Demo completed ===")
}
