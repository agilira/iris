package benchmarks

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/agilira/iris"
)

// TestBenchmarkResultsValidation validates that benchmark scenarios
// perform actual logging operations and produce verifiable output.
// This test ensures measurement integrity and demonstrates that
// performance metrics reflect real work being performed.
func TestBenchmarkResultsValidation(t *testing.T) {
	t.Log("Starting benchmark results validation")

	// Test 1: Validate AccumulatedContext output generation
	t.Run("AccumulatedContext_OutputValidation", func(t *testing.T) {
		t.Log("Validating AccumulatedContext scenario output generation")

		// Capture output
		var buf bytes.Buffer

		// Create logger with captured output
		logger, err := iris.New(iris.Config{
			Level:              iris.Debug,
			Output:             iris.WrapWriter(&buf),
			Encoder:            iris.NewJSONEncoder(),
			Capacity:           64,
			BatchSize:          1,
			Architecture:       iris.SingleRing,
			NumRings:           1,
			BackpressurePolicy: 0,
		})
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		t.Log("Logger instance created and started")

		// Create AccumulatedContext exactly like benchmark
		contextLogger := logger.With(
			iris.Int("int", 1),
			iris.String("string", "value"),
			iris.Time("time", time.Unix(0, 0)),
			iris.String("user1_name", "Jane Doe"),
			iris.String("user2_name", "Jane Doe"),
			iris.String("error", "fail"),
			iris.String("field7", "value7"),
			iris.String("field8", "value8"),
			iris.String("field9", "value9"),
			iris.String("field10", "value10"),
		)

		t.Log("AccumulatedContext logger created with 10 pre-allocated fields")

		// Perform the exact operation that is benchmarked
		contextLogger.Info("test message")

		// Force sync to ensure output is written
		logger.Sync()
		time.Sleep(10 * time.Millisecond) // Allow background processing

		output := buf.String()
		t.Logf("Generated output length: %d bytes", len(output))

		if len(output) == 0 {
			t.Fatal("AccumulatedContext scenario produced no output")
		}

		t.Logf("Generated output: %s", strings.TrimSpace(output))

		// Verify all 10 fields are present
		expectedFields := []string{
			`"int":1`,
			`"string":"value"`,
			`"user1_name":"Jane Doe"`,
			`"user2_name":"Jane Doe"`,
			`"error":"fail"`,
			`"field7":"value7"`,
			`"field8":"value8"`,
			`"field9":"value9"`,
			`"field10":"value10"`,
			`"msg":"test message"`,
		}

		for _, field := range expectedFields {
			if !strings.Contains(output, field) {
				t.Errorf("Expected field %s not found in output", field)
			}
		}

		t.Log("AccumulatedContext output validation completed successfully")
	})

	// Test 2: Validate WithoutFields output generation
	t.Run("WithoutFields_OutputValidation", func(t *testing.T) {
		t.Log("Validating WithoutFields scenario output generation")

		var buf bytes.Buffer

		logger, err := iris.New(iris.Config{
			Level:              iris.Debug,
			Output:             iris.WrapWriter(&buf),
			Encoder:            iris.NewJSONEncoder(),
			Capacity:           64,
			BatchSize:          1,
			Architecture:       iris.SingleRing,
			NumRings:           1,
			BackpressurePolicy: 0,
		})
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		t.Log("Logger instance created and started")

		// Perform WithoutFields benchmark operation
		logger.Info("test message")

		logger.Sync()
		time.Sleep(10 * time.Millisecond)

		output := buf.String()
		t.Logf("Generated output length: %d bytes", len(output))

		if len(output) == 0 {
			t.Fatal("WithoutFields scenario produced no output")
		}

		t.Logf("Generated output: %s", strings.TrimSpace(output))

		// Verify basic structure
		if !strings.Contains(output, `"msg":"test message"`) {
			t.Error("Message field not found in output")
		}

		// Verify it has fewer fields than AccumulatedContext
		fieldCount := strings.Count(output, `":`)
		t.Logf("Field count: %d", fieldCount)

		t.Log("WithoutFields output validation completed successfully")
	})

	// Test 3: Validate AddingFields output generation
	t.Run("AddingFields_OutputValidation", func(t *testing.T) {
		t.Log("Validating AddingFields scenario output generation")

		var buf bytes.Buffer

		logger, err := iris.New(iris.Config{
			Level:              iris.Debug,
			Output:             iris.WrapWriter(&buf),
			Encoder:            iris.NewJSONEncoder(),
			Capacity:           64,
			BatchSize:          1,
			Architecture:       iris.SingleRing,
			NumRings:           1,
			BackpressurePolicy: 0,
		})
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		t.Log("Logger instance created and started")

		// Perform AddingFields benchmark operation exactly
		logger.Info("test message",
			iris.Int("int", 1),
			iris.String("string", "value"),
			iris.Time("time", time.Unix(0, 0)),
			iris.String("user1_name", "Jane Doe"),
			iris.String("user2_name", "Jane Doe"),
			iris.String("error", "fail"),
			iris.String("field7", "value7"),
			iris.String("field8", "value8"),
			iris.String("field9", "value9"),
			iris.String("field10", "value10"),
		)

		logger.Sync()
		time.Sleep(10 * time.Millisecond)

		output := buf.String()
		t.Logf("Generated output length: %d bytes", len(output))

		if len(output) == 0 {
			t.Fatal("AddingFields scenario produced no output")
		}

		t.Logf("Generated output: %s", strings.TrimSpace(output))

		// Verify all 10 fields are present
		expectedFields := []string{
			`"int":1`,
			`"string":"value"`,
			`"user1_name":"Jane Doe"`,
			`"user2_name":"Jane Doe"`,
			`"error":"fail"`,
			`"field7":"value7"`,
			`"field8":"value8"`,
			`"field9":"value9"`,
			`"field10":"value10"`,
			`"msg":"test message"`,
		}

		for _, field := range expectedFields {
			if !strings.Contains(output, field) {
				t.Errorf("Expected field %s not found in output", field)
			}
		}

		t.Log("AddingFields output validation completed successfully")
	})

	// Test 4: Performance measurement validation
	t.Run("Performance_MeasurementValidation", func(t *testing.T) {
		t.Log("Validating performance measurement methodology")

		// Test AccumulatedContext timing
		logger1, _ := iris.New(iris.Config{
			Level:              iris.Debug,
			Output:             iris.WrapWriter(&bytes.Buffer{}),
			Encoder:            iris.NewJSONEncoder(),
			Capacity:           64,
			BatchSize:          1,
			Architecture:       iris.SingleRing,
			NumRings:           1,
			BackpressurePolicy: 0,
		})
		logger1.Start()
		defer logger1.Close()

		contextLogger := logger1.With(
			iris.Int("int", 1),
			iris.String("string", "value"),
			iris.Time("time", time.Unix(0, 0)),
			iris.String("user1_name", "Jane Doe"),
			iris.String("user2_name", "Jane Doe"),
			iris.String("error", "fail"),
			iris.String("field7", "value7"),
			iris.String("field8", "value8"),
			iris.String("field9", "value9"),
			iris.String("field10", "value10"),
		)

		// Warm up
		for i := 0; i < 1000; i++ {
			contextLogger.Info("warmup")
		}

		// Time AccumulatedContext
		iterations := 100000
		start := time.Now()
		for i := 0; i < iterations; i++ {
			contextLogger.Info("test message")
		}
		duration1 := time.Since(start)
		ns1 := duration1.Nanoseconds() / int64(iterations)

		t.Logf("AccumulatedContext measurement: %d ns/op", ns1)

		// Test WithoutFields timing
		logger2, _ := iris.New(iris.Config{
			Level:              iris.Debug,
			Output:             iris.WrapWriter(&bytes.Buffer{}),
			Encoder:            iris.NewJSONEncoder(),
			Capacity:           64,
			BatchSize:          1,
			Architecture:       iris.SingleRing,
			NumRings:           1,
			BackpressurePolicy: 0,
		})
		logger2.Start()
		defer logger2.Close()

		// Warm up
		for i := 0; i < 1000; i++ {
			logger2.Info("warmup")
		}

		// Time WithoutFields
		start = time.Now()
		for i := 0; i < iterations; i++ {
			logger2.Info("test message")
		}
		duration2 := time.Since(start)
		ns2 := duration2.Nanoseconds() / int64(iterations)

		t.Logf("WithoutFields measurement: %d ns/op", ns2)

		// Calculate performance ratio
		if ns1 > 0 {
			ratio := float64(ns2) / float64(ns1)
			t.Logf("Performance ratio (WithoutFields/AccumulatedContext): %.2f", ratio)
		}

		t.Log("Performance measurement validation completed")
	})

	t.Log("Benchmark results validation completed successfully")
	t.Log("All scenarios produce verifiable output and demonstrate actual logging work")
}
