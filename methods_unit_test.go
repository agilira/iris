package iris

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestFatalMethod tests the Fatal logging method
func TestFatalMethod(t *testing.T) {
	t.Run("FatalLogsMessage", func(t *testing.T) {
		// Test that Fatal logs the message before exiting
		// We need to run this in a subprocess to avoid exiting our test process
		if os.Getenv("TEST_FATAL") == "1" {
			// This code runs in the subprocess
			var buf bytes.Buffer
			config := Config{
				Level:      DebugLevel,
				Writer:     &buf,
				Format:     JSONFormat,
				BufferSize: 1024,
				BatchSize:  32,
			}
			logger, err := New(config)
			if err != nil {
				fmt.Printf("Failed to create logger: %v", err)
				os.Exit(2)
			}

			logger.Fatal("test fatal message", String("key", "value"))
			return
		}

		// Run the test in a subprocess
		cmd := exec.Command(os.Args[0], "-test.run=TestFatalMethod/FatalLogsMessage")
		cmd.Env = append(os.Environ(), "TEST_FATAL=1")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("Expected Fatal to exit with non-zero status")
		}

		// Verify exit status is 1
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 {
				t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
			}
		}
	})

	t.Run("FatalWithClosedLogger", func(t *testing.T) {
		// Test Fatal with a closed logger
		if os.Getenv("TEST_FATAL_CLOSED") == "1" {
			config := Config{
				Level:      DebugLevel,
				Writer:     &bytes.Buffer{},
				Format:     JSONFormat,
				BufferSize: 1024,
				BatchSize:  32,
			}
			logger, err := New(config)
			if err != nil {
				fmt.Printf("Failed to create logger: %v", err)
				os.Exit(2)
			}

			logger.Close()
			logger.Fatal("fatal with closed logger")
			return
		}

		cmd := exec.Command(os.Args[0], "-test.run=TestFatalMethod/FatalWithClosedLogger")
		cmd.Env = append(os.Environ(), "TEST_FATAL_CLOSED=1")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("Expected Fatal to exit with non-zero status even when logger is closed")
		}

		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 {
				t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
			}
		}
	})

	t.Run("FatalWithFields", func(t *testing.T) {
		// Test Fatal with various field types
		if os.Getenv("TEST_FATAL_FIELDS") == "1" {
			config := Config{
				Level:      DebugLevel,
				Writer:     &bytes.Buffer{},
				Format:     JSONFormat,
				BufferSize: 1024,
				BatchSize:  32,
			}
			logger, err := New(config)
			if err != nil {
				fmt.Printf("Failed to create logger: %v", err)
				os.Exit(2)
			}

			logger.Fatal("fatal with fields",
				String("string_field", "test"),
				Int("int_field", 42),
				Bool("bool_field", true),
				Duration("duration_field", time.Minute),
			)
			return
		}

		cmd := exec.Command(os.Args[0], "-test.run=TestFatalMethod/FatalWithFields")
		cmd.Env = append(os.Environ(), "TEST_FATAL_FIELDS=1")

		err := cmd.Run()
		if err == nil {
			t.Fatal("Expected Fatal to exit with non-zero status")
		}

		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 {
				t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
			}
		}
	})
}

// TestDPanicMethod tests the DPanic logging method
func TestDPanicMethod(t *testing.T) {
	t.Run("DPanicLogsAndPanics", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:      DebugLevel,
			Writer:     &buf,
			Format:     JSONFormat,
			BufferSize: 1024,
			BatchSize:  32,
		}
		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		// Test that DPanic panics
		defer func() {
			if r := recover(); r != nil {
				// Verify the panic message
				panicMsg := fmt.Sprintf("%v", r)
				if !strings.Contains(panicMsg, "dpanic level log") {
					t.Errorf("Expected panic message to contain 'dpanic level log', got %q", panicMsg)
				}

				// Verify that the log was written before panicking
				// Give the logger a moment to write
				time.Sleep(10 * time.Millisecond)
				logger.Close()

				output := buf.String()
				if !strings.Contains(output, "test dpanic message") {
					t.Errorf("Expected log output to contain 'test dpanic message', got %q", output)
				}
			} else {
				t.Error("Expected DPanic to panic, but it didn't")
			}
		}()

		logger.DPanic("test dpanic message", String("key", "value"))
	})

	t.Run("DPanicWithClosedLogger", func(t *testing.T) {
		config := Config{
			Level:      DebugLevel,
			Writer:     &bytes.Buffer{},
			Format:     JSONFormat,
			BufferSize: 1024,
			BatchSize:  32,
		}
		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Close()

		// Even with closed logger, DPanic should still panic
		defer func() {
			if r := recover(); r != nil {
				panicMsg := fmt.Sprintf("%v", r)
				if !strings.Contains(panicMsg, "dpanic level log") {
					t.Errorf("Expected panic message to contain 'dpanic level log', got %q", panicMsg)
				}
			} else {
				t.Error("Expected DPanic to panic even when logger is closed")
			}
		}()

		logger.DPanic("dpanic with closed logger")
	})

	t.Run("DPanicWithFields", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:      DebugLevel,
			Writer:     &buf,
			Format:     JSONFormat,
			BufferSize: 1024,
			BatchSize:  32,
		}
		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		defer func() {
			if r := recover(); r != nil {
				// Verify panic occurred
				panicMsg := fmt.Sprintf("%v", r)
				if !strings.Contains(panicMsg, "dpanic level log") {
					t.Errorf("Expected panic message to contain 'dpanic level log', got %q", panicMsg)
				}

				// Verify that fields were logged
				time.Sleep(10 * time.Millisecond)
				logger.Close()

				output := buf.String()
				if !strings.Contains(output, "dpanic with fields") {
					t.Errorf("Expected log output to contain 'dpanic with fields', got %q", output)
				}
			} else {
				t.Error("Expected DPanic to panic")
			}
		}()

		logger.DPanic("dpanic with fields",
			String("string_field", "test"),
			Int("int_field", 123),
			Bool("bool_field", false),
			Float64("float_field", 3.14),
		)
	})

	t.Run("DPanicWithLevelFiltering", func(t *testing.T) {
		// Test DPanic when level is set higher than DPanic
		var buf bytes.Buffer
		config := Config{
			Level:      FatalLevel, // Only log Fatal and above
			Writer:     &buf,
			Format:     JSONFormat,
			BufferSize: 1024,
			BatchSize:  32,
		}
		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		// DPanic should still panic even if the log is filtered out
		defer func() {
			if r := recover(); r != nil {
				panicMsg := fmt.Sprintf("%v", r)
				if !strings.Contains(panicMsg, "dpanic level log") {
					t.Errorf("Expected panic message to contain 'dpanic level log', got %q", panicMsg)
				}

				// Verify that the log was NOT written due to level filtering
				time.Sleep(10 * time.Millisecond)
				logger.Close()

				output := buf.String()
				if strings.Contains(output, "filtered dpanic") {
					t.Errorf("Expected log to be filtered out, but found it in output: %q", output)
				}
			} else {
				t.Error("Expected DPanic to panic even when level is filtered")
			}
		}()

		logger.DPanic("filtered dpanic message")
	})
}

// TestPanicMethod tests the Panic logging method for comparison
func TestPanicMethod(t *testing.T) {
	t.Run("PanicLogsAndPanics", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:      DebugLevel,
			Writer:     &buf,
			Format:     JSONFormat,
			BufferSize: 1024,
			BatchSize:  32,
		}
		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		defer func() {
			if r := recover(); r != nil {
				panicMsg := fmt.Sprintf("%v", r)
				if !strings.Contains(panicMsg, "panic level log") {
					t.Errorf("Expected panic message to contain 'panic level log', got %q", panicMsg)
				}

				time.Sleep(10 * time.Millisecond)
				logger.Close()

				output := buf.String()
				if !strings.Contains(output, "test panic message") {
					t.Errorf("Expected log output to contain 'test panic message', got %q", output)
				}
			} else {
				t.Error("Expected Panic to panic")
			}
		}()

		logger.Panic("test panic message", String("key", "value"))
	})
}

// TestMethodsLevelChecking tests level filtering for all methods
func TestMethodsLevelChecking(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      ErrorLevel, // Only Error and above
		Writer:     &buf,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test methods below threshold don't log
	logger.Debug("debug message") // Should be filtered
	logger.Info("info message")   // Should be filtered
	logger.Warn("warn message")   // Should be filtered

	// Test methods at or above threshold do log
	logger.Error("error message") // Should log

	// Give logger time to process
	time.Sleep(10 * time.Millisecond)
	logger.Close()

	output := buf.String()

	// Verify filtered messages don't appear
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should have been filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should have been filtered out")
	}
	if strings.Contains(output, "warn message") {
		t.Error("Warn message should have been filtered out")
	}

	// Verify error message appears
	if !strings.Contains(output, "error message") {
		t.Error("Error message should have been logged")
	}
}

// TestMethodsWithPreFields tests methods with pre-bound fields from With()
func TestMethodsWithPreFields(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      DebugLevel,
		Writer:     &buf,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create logger with pre-bound fields
	childLogger := logger.With(
		String("service", "test-service"),
		String("version", "1.0.0"),
	)

	// Test that pre-bound fields are included in all methods
	childLogger.Debug("debug with pre-fields", String("extra", "debug"))
	childLogger.Info("info with pre-fields", String("extra", "info"))
	childLogger.Warn("warn with pre-fields", String("extra", "warn"))
	childLogger.Error("error with pre-fields", String("extra", "error"))

	time.Sleep(10 * time.Millisecond)
	childLogger.Close()

	output := buf.String()

	// Verify all messages contain pre-bound fields
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if !strings.Contains(line, "service") || !strings.Contains(line, "test-service") {
			t.Errorf("Line %d missing service field: %s", i+1, line)
		}
		if !strings.Contains(line, "version") || !strings.Contains(line, "1.0.0") {
			t.Errorf("Line %d missing version field: %s", i+1, line)
		}
	}
}

// TestMethodsPerformance tests that methods have fast-path optimizations
func TestMethodsPerformance(t *testing.T) {
	// Test with disabled level to verify fast-path exits
	var buf bytes.Buffer
	config := Config{
		Level:      FatalLevel, // Disable all but Fatal
		Writer:     &buf,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// These should all take the fast path (early return)
	start := time.Now()
	for i := 0; i < 1000; i++ {
		logger.Debug("disabled debug")
		logger.Info("disabled info")
		logger.Warn("disabled warn")
		logger.Error("disabled error")
	}
	elapsed := time.Since(start)

	logger.Close()

	// Should be very fast since all calls take early return path
	if elapsed > 10*time.Millisecond {
		t.Errorf("Disabled logging took too long: %v (expected < 10ms)", elapsed)
	}

	// Verify no output was generated
	output := buf.String()
	if strings.TrimSpace(output) != "" {
		t.Errorf("Expected no output for disabled levels, got: %q", output)
	}
}

// TestLogMethod tests the core log method behavior
func TestLogMethod(t *testing.T) {
	t.Run("LogWithSampling", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:      DebugLevel,
			Writer:     &buf,
			Format:     JSONFormat,
			BufferSize: 1024,
			BatchSize:  32,
		}
		prodSampling := NewProductionSampling()
		config.SamplingConfig = &prodSampling

		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		// Log multiple messages - some should be sampled out
		for i := 0; i < 100; i++ {
			logger.Info(fmt.Sprintf("sampled message %d", i))
		}

		time.Sleep(10 * time.Millisecond)
		logger.Close()

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		// Production sampling should keep some messages but potentially fewer than all
		// The sampling logic allows the first 100 messages, then 1 per 100 thereafter
		if len(lines) == 0 {
			t.Error("Expected some messages to survive sampling")
		}

		// Since we send exactly 100 messages and Initial=100, all should pass
		// This test verifies sampling infrastructure is working
		if len(lines) < 50 || len(lines) > 100 {
			t.Errorf("Expected 50-100 lines with sampling, got %d lines", len(lines))
		}
	})

	t.Run("LogWithCallerInfo", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:        DebugLevel,
			Writer:       &buf,
			Format:       JSONFormat,
			BufferSize:   1024,
			BatchSize:    32,
			EnableCaller: true,
		}

		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		// Get current location info
		_, _, line, _ := runtime.Caller(0)
		logger.Info("message with caller") // This line should be captured
		expectedLine := line + 1

		time.Sleep(10 * time.Millisecond)
		logger.Close()

		output := buf.String()

		// Should contain caller information
		if !strings.Contains(output, "methods_unit_test.go") {
			t.Errorf("Expected output to contain caller file info, got: %s", output)
		}

		// Should contain line number (approximately)
		lineStr := fmt.Sprintf("%d", expectedLine)
		if !strings.Contains(output, lineStr) {
			t.Errorf("Expected output to contain line number %s, got: %s", lineStr, output)
		}
	})
}
