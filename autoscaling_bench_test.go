// autoscaling_bench_test.go: Auto-scaling benchmarks
//
// This benchmark suite validates the performance characteristics of the
// auto-scaling logging architecture, demonstrating optimal
// performance across different workload patterns.
//
// Benchmark Categories:
//   1. SingleRing mode performance validation
//   2. MPSC mode performance validation
//   3. Auto-scaling transition overhead
//   4. Multi-workload adaptive performance
//
// Copyright (c) 2025 AGILira
// Series: IRIS Logging Library - Auto-Scaling
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"os"
	"sync"
	"testing"
	"time"
)

// BenchmarkAutoScaling_SingleProducer tests SingleRing mode performance
func BenchmarkAutoScaling_SingleProducer(b *testing.B) {
	config := Config{
		Level:    Info,
		Output:   os.Stdout,
		Encoder:  NewJSONEncoder(),
		Capacity: 2048,
	}

	// Configure for SingleRing preference (low thresholds)
	scalingConfig := AutoScalingConfig{
		ScaleToMPSCWriteThreshold:   100000,           // Very high threshold (won't trigger)
		ScaleToMPSCContentionRatio:  50,               // High contention threshold
		ScaleToMPSCLatencyThreshold: 10 * time.Second, // Very high latency
		ScaleToMPSCGoroutineCount:   100,              // High goroutine count

		ScaleToSingleWriteThreshold:  1, // Always prefer Single
		ScaleToSingleContentionRatio: 0,
		ScaleToSingleLatencyMax:      1 * time.Hour,

		MeasurementWindow:    100 * time.Millisecond,
		ScalingCooldown:      1 * time.Second,
		StabilityRequirement: 3,
	}

	autoLogger, err := NewAutoScalingLogger(config, scalingConfig)
	if err != nil {
		b.Fatal(err)
	}

	autoLogger.Start()
	defer autoLogger.Close()

	// Warm up and ensure SingleRing mode
	for i := 0; i < 100; i++ {
		autoLogger.Info("warmup")
	}
	time.Sleep(200 * time.Millisecond)

	// Verify we're in SingleRing mode
	if autoLogger.GetCurrentMode() != SingleRingMode {
		b.Fatalf("Expected SingleRing mode, got %s", autoLogger.GetCurrentMode())
	}

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark single-producer performance
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			autoLogger.Info("Benchmark message",
				String("producer", "single"),
				Int("iteration", 1),
			)
		}
	})
}

// BenchmarkAutoScaling_MultiProducer tests MPSC mode performance
func BenchmarkAutoScaling_MultiProducer(b *testing.B) {
	config := Config{
		Level:    Info,
		Output:   os.Stdout,
		Encoder:  NewJSONEncoder(),
		Capacity: 4096,
	}

	// Configure for MPSC preference (low thresholds)
	scalingConfig := AutoScalingConfig{
		ScaleToMPSCWriteThreshold:   1,                   // Always trigger MPSC
		ScaleToMPSCContentionRatio:  0,                   // Any contention triggers
		ScaleToMPSCLatencyThreshold: 1 * time.Nanosecond, // Any latency
		ScaleToMPSCGoroutineCount:   1,                   // Single goroutine triggers

		ScaleToSingleWriteThreshold:  0, // Never scale back
		ScaleToSingleContentionRatio: 0,
		ScaleToSingleLatencyMax:      0,

		MeasurementWindow:    50 * time.Millisecond,
		ScalingCooldown:      100 * time.Millisecond,
		StabilityRequirement: 1,
	}

	autoLogger, err := NewAutoScalingLogger(config, scalingConfig)
	if err != nil {
		b.Fatal(err)
	}

	autoLogger.Start()
	defer autoLogger.Close()

	// Trigger MPSC mode with multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				autoLogger.Info("trigger MPSC")
			}
		}()
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	// Verify we're in MPSC mode
	if autoLogger.GetCurrentMode() != MPSCMode {
		b.Fatalf("Expected MPSC mode, got %s", autoLogger.GetCurrentMode())
	}

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark multi-producer performance
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			autoLogger.Info("Benchmark message",
				String("producer", "multi"),
				Int("iteration", 1),
			)
		}
	})
}

// BenchmarkAutoScaling_Transition measures scaling transition overhead
func BenchmarkAutoScaling_Transition(b *testing.B) {
	config := Config{
		Level:    Info,
		Output:   os.Stdout,
		Encoder:  NewJSONEncoder(),
		Capacity: 2048,
	}

	// Configure for easy transitions
	scalingConfig := AutoScalingConfig{
		ScaleToMPSCWriteThreshold:   100, // Moderate threshold
		ScaleToMPSCContentionRatio:  5,
		ScaleToMPSCLatencyThreshold: 100 * time.Microsecond,
		ScaleToMPSCGoroutineCount:   2,

		ScaleToSingleWriteThreshold:  10,
		ScaleToSingleContentionRatio: 1,
		ScaleToSingleLatencyMax:      50 * time.Microsecond,

		MeasurementWindow:    20 * time.Millisecond,
		ScalingCooldown:      50 * time.Millisecond,
		StabilityRequirement: 1,
	}

	autoLogger, err := NewAutoScalingLogger(config, scalingConfig)
	if err != nil {
		b.Fatal(err)
	}

	autoLogger.Start()
	defer autoLogger.Close()

	b.ResetTimer()

	// Benchmark transition overhead
	for i := 0; i < b.N; i++ {
		// Trigger high load (should scale to MPSC)
		var wg sync.WaitGroup
		for g := 0; g < 3; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					autoLogger.Info("high load message")
				}
			}()
		}
		wg.Wait()

		// Wait for potential scaling
		time.Sleep(30 * time.Millisecond)

		// Trigger low load (should scale back to Single)
		for j := 0; j < 5; j++ {
			autoLogger.Info("low load message")
			time.Sleep(10 * time.Millisecond)
		}

		// Wait for potential scale-back
		time.Sleep(30 * time.Millisecond)
	}
}

// BenchmarkAutoScaling_AdaptiveWorkload tests real-world adaptive patterns
func BenchmarkAutoScaling_AdaptiveWorkload(b *testing.B) {
	config := Config{
		Level:    Info,
		Output:   os.Stdout,
		Encoder:  NewJSONEncoder(),
		Capacity: 2048,
	}

	autoLogger, err := NewAutoScalingLogger(config, DefaultAutoScalingConfig())
	if err != nil {
		b.Fatal(err)
	}

	autoLogger.Start()
	defer autoLogger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	// Simulate real-world adaptive workload patterns
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Burst pattern: high load followed by quiet period

			// High load burst (should trigger MPSC scaling)
			var wg sync.WaitGroup
			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < 10; j++ {
						autoLogger.Info("burst message",
							Int("goroutine", id),
							Int("message", j),
						)
					}
				}(i)
			}
			wg.Wait()

			// Quiet period (should scale back to SingleRing)
			for i := 0; i < 5; i++ {
				autoLogger.Info("quiet message", Int("iteration", i))
				time.Sleep(2 * time.Millisecond)
			}
		}
	})
}

// BenchmarkAutoScaling_CompareWithStatic compares auto-scaling vs static modes
func BenchmarkAutoScaling_CompareWithStatic(b *testing.B) {
	config := Config{
		Level:    Info,
		Output:   os.Stdout,
		Encoder:  NewJSONEncoder(),
		Capacity: 2048,
	}

	b.Run("AutoScaling", func(b *testing.B) {
		autoLogger, err := NewAutoScalingLogger(config, DefaultAutoScalingConfig())
		if err != nil {
			b.Fatal(err)
		}

		autoLogger.Start()
		defer autoLogger.Close()

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				autoLogger.Info("auto-scaling message", Int("iteration", 1))
			}
		})
	})

	b.Run("StaticSingleRing", func(b *testing.B) {
		staticLogger, err := New(config)
		if err != nil {
			b.Fatal(err)
		}

		staticLogger.Start()
		defer staticLogger.Close()

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				staticLogger.Info("static single message", Int("iteration", 1))
			}
		})
	})
}

// BenchmarkAutoScaling_ZeroAllocation verifies zero-allocation promise
func BenchmarkAutoScaling_ZeroAllocation(b *testing.B) {
	config := Config{
		Level:    Info,
		Output:   os.Stdout,
		Encoder:  NewJSONEncoder(),
		Capacity: 2048,
	}

	autoLogger, err := NewAutoScalingLogger(config, DefaultAutoScalingConfig())
	if err != nil {
		b.Fatal(err)
	}

	autoLogger.Start()
	defer autoLogger.Close()

	// Ensure consistent mode for allocation testing
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	b.ReportAllocs()

	// This should show 0 B/op and 0 allocs/op for the logging path
	for i := 0; i < b.N; i++ {
		autoLogger.Info("zero allocation test",
			String("key1", "value1"),
			Int("key2", 42),
			Bool("key3", true),
		)
	}
}

// BenchmarkAutoScaling_ScalingStats measures statistics collection overhead
func BenchmarkAutoScaling_ScalingStats(b *testing.B) {
	config := Config{
		Level:    Info,
		Output:   os.Stdout,
		Encoder:  NewJSONEncoder(),
		Capacity: 2048,
	}

	autoLogger, err := NewAutoScalingLogger(config, DefaultAutoScalingConfig())
	if err != nil {
		b.Fatal(err)
	}

	autoLogger.Start()
	defer autoLogger.Close()

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark statistics collection overhead
	for i := 0; i < b.N; i++ {
		stats := autoLogger.GetScalingStats()
		_ = stats.CurrentMode // Use the stats to prevent optimization
		_ = stats.TotalScaleOperations
		_ = stats.TotalWrites
	}
}
