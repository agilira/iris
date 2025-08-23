// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0
// Real-world benchmark test to verify processing is happening

package iris

import (
	"sync/atomic"
	"testing"
	"time"
)

// Counter sink implements WriteSyncer and counts written bytes
type counterSink struct {
	bytesWritten atomic.Int64
	callCount    atomic.Int32
}

func (cs *counterSink) Write(p []byte) (n int, err error) {
	cs.bytesWritten.Add(int64(len(p)))
	cs.callCount.Add(1)
	return len(p), nil
}

func (cs *counterSink) Sync() error {
	return nil
}

// BenchmarkRealProcessing verifies that logs are actually being processed
func BenchmarkRealProcessing(b *testing.B) {
	counter := &counterSink{}

	logger, err := New(Config{
		Level:     Info,
		Encoder:   NewJSONEncoder(),
		Output:    counter,
		Capacity:  1024,
		BatchSize: 64,
	})
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	// Start the logger before benchmark timing
	logger.Start()

	// Log some messages synchronously first to ensure the consumer is working
	for i := 0; i < 10; i++ {
		logger.Info("Warmup message", Int("i", i))
	}

	// Give consumer time to process
	time.Sleep(50 * time.Millisecond)

	initialCalls := counter.callCount.Load()
	initialBytes := counter.bytesWritten.Load()

	b.Logf("Pre-benchmark state: %d calls, %d bytes", initialCalls, initialBytes)

	// Reset timer for fair measurement
	b.ResetTimer()

	// Log messages in parallel (like standard benchmarks)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("This is a test message", String("key", "value"))
		}
	})

	// Stop timer before cleanup
	b.StopTimer()

	// Ensure messages are processed by sleeping
	time.Sleep(100 * time.Millisecond)

	// Close logger and process any remaining messages
	logger.Close()

	// Verify processing happened
	finalCalls := counter.callCount.Load()
	finalBytes := counter.bytesWritten.Load()

	addedCalls := finalCalls - initialCalls
	addedBytes := finalBytes - initialBytes

	b.ReportMetric(float64(addedCalls), "calls")
	b.ReportMetric(float64(addedBytes), "bytes")

	// Log statistics
	b.Logf("Processed during benchmark: %d calls, %d bytes", addedCalls, addedBytes)

	// Fail if no processing happened
	if finalCalls == 0 {
		b.Fatal("No log processing occurred at all!")
	} else if addedCalls == 0 {
		b.Fatal("No log processing occurred during benchmark phase!")
	}
}

// BenchmarkCompareProcessingSpeed directly compares performance of normal vs. discard sink
func BenchmarkCompareProcessingSpeed(b *testing.B) {
	b.Run("WithRealProcessing", func(b *testing.B) {
		counter := &counterSink{}
		logger, _ := New(Config{
			Level:   Info,
			Encoder: NewJSONEncoder(),
			Output:  counter,
		})
		logger.Start()
		defer logger.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Test message", Int("i", i))
		}
		b.StopTimer()

		b.ReportMetric(float64(counter.callCount.Load()), "calls")
		b.ReportMetric(float64(counter.bytesWritten.Load()), "bytes")
	})

	b.Run("WithDiscardSink", func(b *testing.B) {
		logger, _ := New(Config{
			Level:   Info,
			Encoder: NewJSONEncoder(),
			Output:  discardSyncer{},
		})
		logger.Start()
		defer logger.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Test message", Int("i", i))
		}
	})
}
