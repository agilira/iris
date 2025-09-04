// examples/hot_reload/main.go: Demonstration of Iris hot reload with Argus
//
// This example shows how Iris can dynamically update its logging level
// when the configuration file changes, without requiring a restart.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agilira/iris"
)

func main() {
	fmt.Println("🔥 Iris Hot Reload Demo with Argus")
	fmt.Println("===================================")

	// Create initial configuration
	config := iris.Config{
		Level:   iris.Info,
		Output:  os.Stdout,
		Encoder: iris.NewJSONEncoder(),
	}

	// Create logger
	logger, err := iris.New(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer logger.Sync()

	// Create configuration file for hot reload
	configFile := "iris_config.json"
	createConfigFile(configFile, "info")

	// Set up hot reload watcher using logger's atomic level
	watcher, err := iris.NewDynamicConfigWatcher(configFile, logger.AtomicLevel())
	if err != nil {
		panic(fmt.Sprintf("Failed to create config watcher: %v", err))
	}
	defer watcher.Stop()

	// Start watching for config changes
	if err := watcher.Start(); err != nil {
		panic(fmt.Sprintf("Failed to start config watcher: %v", err))
	}

	fmt.Printf("✅ Hot reload enabled - watching %s\n", configFile)
	fmt.Println("\n🧪 Try changing the logging level in iris_config.json!")
	fmt.Println("   Available levels: debug, info, warn, error")
	fmt.Println("   Example: {\"level\": \"debug\"}")
	fmt.Println("\n📝 Log messages will appear below:")
	fmt.Println()

	// Simulate application activity with different log levels
	go func() {
		counter := 1
		for {
			// Log at different levels to show filtering
			logger.Debug("Debug message", iris.Int("counter", counter))
			time.Sleep(1 * time.Second)

			logger.Info("Info message", iris.Int("counter", counter))
			time.Sleep(1 * time.Second)

			logger.Warn("Warning message", iris.Int("counter", counter))
			time.Sleep(1 * time.Second)

			logger.Error("Error message", iris.Int("counter", counter))
			time.Sleep(2 * time.Second)

			counter++
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n\n🛑 Shutting down gracefully...")

	// Clean up config file
	if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: Failed to remove config file: %v\n", err)
	}
	fmt.Println("✅ Cleanup complete")
}

// createConfigFile creates the initial configuration file
func createConfigFile(filename, level string) {
	content := fmt.Sprintf(`{
    "level": "%s",
    "development": false,
    "encoder": "json"
}`, level)

	if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
		panic(fmt.Sprintf("Failed to create config file: %v", err))
	}
}
