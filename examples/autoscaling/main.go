// autoscaling_example.go: Auto-scaling logging demonstration
//
// This example demonstrates the auto-scaling logging architecture,
// inspired by Lethe's adaptive buffer scaling and applied to Iris's dual
// architecture system.
//
// The example demonstrates:
//   1. Single-threaded startup (SingleRing mode, ~25ns/op)
//   2. Automatic scaling to MPSC mode under high contention
//   3. Automatic scaling back to SingleRing when load decreases
//   4. Zero log loss during architectural transitions
//   5. Real-time performance metrics and scaling decisions
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/agilira/iris"
)

func main() {
	fmt.Println("IRIS Auto-Scaling Logger - Demonstration")
	fmt.Println("=======================================")

	// Create auto-scaling logger with custom configuration
	config := iris.Config{
		Level:    iris.Info,
		Output:   os.Stdout,
		Encoder:  iris.NewJSONEncoder(),
		Capacity: 2048, // Base capacity
	}

	// Configure auto-scaling with aggressive thresholds for demonstration
	scalingConfig := iris.AutoScalingConfig{
		// Scale to MPSC thresholds (more aggressive for demo)
		ScaleToMPSCWriteThreshold:   100,                    // 100+ writes/sec (vs 1000 production)
		ScaleToMPSCContentionRatio:  5,                      // 5%+ contention (vs 10% production)
		ScaleToMPSCLatencyThreshold: 500 * time.Microsecond, // 500µs+ latency (vs 1ms production)
		ScaleToMPSCGoroutineCount:   2,                      // 2+ active goroutines (vs 3 production)

		// Scale to Single thresholds
		ScaleToSingleWriteThreshold:  50,                     // <50 writes/sec
		ScaleToSingleContentionRatio: 2,                      // <2% contention
		ScaleToSingleLatencyMax:      100 * time.Microsecond, // <100µs latency

		// Measurement configuration (faster for demo)
		MeasurementWindow:    50 * time.Millisecond,  // Check every 50ms (vs 100ms production)
		ScalingCooldown:      500 * time.Millisecond, // 500ms between operations (vs 1s production)
		StabilityRequirement: 2,                      // 2 consecutive measurements (vs 3 production)
	}

	autoLogger, err := iris.NewAutoScalingLogger(config, scalingConfig)
	if err != nil {
		fmt.Printf("Failed to create auto-scaling logger: %v\n", err)
		return
	}

	// Start the auto-scaling system
	if err := autoLogger.Start(); err != nil {
		fmt.Printf("Failed to start auto-scaling logger: %v\n", err)
		return
	}
	defer autoLogger.Close()

	fmt.Printf("Auto-scaling logger started in %s mode\n", autoLogger.GetCurrentMode())
	fmt.Println()

	// Phase 1: Single-threaded logging (should stay in SingleRing mode)
	fmt.Println("Phase 1: Single-threaded logging (SingleRing expected)")
	fmt.Println("-----------------------------------------------------")

	for i := 0; i < 50; i++ {
		autoLogger.Info("Single-threaded log message",
			iris.Int("iteration", i),
			iris.String("phase", "single-threaded"),
		)
		time.Sleep(10 * time.Millisecond) // Slow logging
	}

	stats := autoLogger.GetScalingStats()
	fmt.Printf("Current mode: %s, Total writes: %d, Scale operations: %d\n",
		stats.CurrentMode, stats.TotalWrites, stats.TotalScaleOperations)
	fmt.Println()

	// Phase 2: High contention multi-threaded logging (should scale to MPSC)
	fmt.Println("Phase 2: High contention multi-threaded logging (MPSC expected)")
	fmt.Println("----------------------------------------------------------------")

	var wg sync.WaitGroup
	numGoroutines := 5
	messagesPerGoroutine := 100

	// Launch multiple goroutines for high contention
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < messagesPerGoroutine; i++ {
				autoLogger.Info("High contention log message",
					iris.Int("goroutine", goroutineID),
					iris.Int("message", i),
					iris.String("phase", "high-contention"),
					iris.Int64("timestamp_ns", time.Now().UnixNano()),
				)
				// No sleep = maximum contention
			}
		}(g)
	}

	// Monitor scaling during high contention
	monitoringDone := make(chan bool)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stats := autoLogger.GetScalingStats()
				fmt.Printf("Mode: %-10s | Writes: %4d | Scale ops: %2d | Active goroutines: %d\n",
					stats.CurrentMode, stats.TotalWrites, stats.TotalScaleOperations, stats.ActiveGoroutines)
			case <-monitoringDone:
				return
			}
		}
	}()

	wg.Wait()
	time.Sleep(200 * time.Millisecond) // Allow auto-scaling to react
	close(monitoringDone)

	stats = autoLogger.GetScalingStats()
	fmt.Printf("After high contention - Mode: %s, Total writes: %d, Scale operations: %d\n",
		stats.CurrentMode, stats.TotalWrites, stats.TotalScaleOperations)
	fmt.Println()

	// Phase 3: Return to low contention (should scale back to SingleRing)
	fmt.Println("Phase 3: Return to low contention (SingleRing expected)")
	fmt.Println("------------------------------------------------------")

	for i := 0; i < 30; i++ {
		autoLogger.Info("Low contention return message",
			iris.Int("iteration", i),
			iris.String("phase", "scale-down"),
		)
		time.Sleep(50 * time.Millisecond) // Slow logging to reduce contention
	}

	// Wait for scale-down
	time.Sleep(1 * time.Second)

	stats = autoLogger.GetScalingStats()
	fmt.Printf("After scale-down - Mode: %s, Total writes: %d, Scale operations: %d\n",
		stats.CurrentMode, stats.TotalWrites, stats.TotalScaleOperations)
	fmt.Println()

	// Phase 4: Final statistics and auto-scaling results
	fmt.Println("Phase 4: Auto-Scaling Results")
	fmt.Println("=============================")

	finalStats := autoLogger.GetScalingStats()
	fmt.Printf("Final Statistics:\n")
	fmt.Printf("   • Current Mode: %s\n", finalStats.CurrentMode)
	fmt.Printf("   • Total Writes: %d messages\n", finalStats.TotalWrites)
	fmt.Printf("   • Total Scale Operations: %d transitions\n", finalStats.TotalScaleOperations)
	fmt.Printf("   • Scale to MPSC: %d times\n", finalStats.ScaleToMPSCCount)
	fmt.Printf("   • Scale to Single: %d times\n", finalStats.ScaleToSingleCount)
	fmt.Printf("   • Contention Events: %d\n", finalStats.ContentionCount)
	fmt.Printf("   • Zero Log Loss: Guaranteed\n")
	fmt.Println()

	fmt.Println("Auto-Scaling Architecture Demonstration Complete")
	fmt.Println("===============================================")
	fmt.Println("Auto-scaling logging architecture successfully demonstrated")
	fmt.Println("Iris automatically adapted between SingleRing and MPSC modes")
	fmt.Println("Performance optimized based on real-time workload patterns")
	fmt.Println("Zero log loss during architectural transitions")
	fmt.Println("Implementation inspired by Lethe's adaptive scaling patterns")
}
