package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/agilira/iris"
)

// Helper function to safely close logger ignoring expected errors
func safeCloseTestLogger(t *testing.T, logger *iris.Logger) {
	if err := logger.Close(); err != nil &&
		!strings.Contains(err.Error(), "sync /dev/stdout: invalid argument") &&
		!strings.Contains(err.Error(), "sync /dev/stdout: bad file descriptor") &&
		!strings.Contains(err.Error(), "ring buffer flush failed") {
		t.Errorf("Failed to close logger: %v", err)
	}
}

func TestLoggerCreation(t *testing.T) {
	// Test that logger can be created successfully
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Output:  os.Stdout,
		Encoder: &iris.TextEncoder{},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseTestLogger(t, logger)

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}
}

func TestLoggerWithCustomOutput(t *testing.T) {
	// Test logger with custom output buffer
	var buf bytes.Buffer

	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Encoder: iris.NewJSONEncoder(),
		Output:  iris.WrapWriter(&buf),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("Error closing logger: %v", err)
		}
	}()

	logger.Start()

	result := logger.Info("test message", iris.Str("key", "value"))
	if !result {
		t.Error("Expected Info log to return true")
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)
	if err := logger.Sync(); err != nil {
		t.Errorf("Error syncing logger: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("Expected output to contain key-value pair, got: %s", output)
	}
}

func TestLoggerStartAndSync(t *testing.T) {
	// Test logger start and sync operations
	var buf bytes.Buffer

	logger, err := iris.New(iris.Config{
		Level:   iris.Info,
		Encoder: iris.NewTextEncoder(),
		Output:  iris.WrapWriter(&buf),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("Error closing logger: %v", err)
		}
	}()

	// Start the logger
	logger.Start()

	// Log multiple messages
	logger.Info("message1")
	logger.Info("message2")
	logger.Info("message3")

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Sync to ensure all messages are written
	if err := logger.Sync(); err != nil {
		t.Errorf("Error syncing logger: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "message1") ||
		!strings.Contains(output, "message2") ||
		!strings.Contains(output, "message3") {
		t.Errorf("Expected all messages in output, got: %s", output)
	}
}

func TestLoggerWithDifferentLevels(t *testing.T) {
	// Test logger with different log levels
	var buf bytes.Buffer

	logger, err := iris.New(iris.Config{
		Level:   iris.Warn, // Only warn and above
		Encoder: iris.NewTextEncoder(),
		Output:  iris.WrapWriter(&buf),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("Error closing logger: %v", err)
		}
	}()

	logger.Start()

	// These should not appear (below warn level)
	logger.Debug("debug message")
	logger.Info("info message")

	// These should appear
	logger.Warn("warn message")
	logger.Error("error message")

	time.Sleep(100 * time.Millisecond)
	if err := logger.Sync(); err != nil {
		t.Errorf("Error syncing logger: %v", err)
	}

	output := buf.String()

	// Debug and Info should not be in output
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear when level is Warn")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should not appear when level is Warn")
	}

	// Warn and Error should be in output
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should appear when level is Warn")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should appear when level is Warn")
	}
}

func TestLoggerConfiguration(t *testing.T) {
	// Test different encoder configurations
	tests := []struct {
		name    string
		encoder iris.Encoder
	}{
		{"JSON Encoder", iris.NewJSONEncoder()},
		{"Text Encoder", iris.NewTextEncoder()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			logger, err := iris.New(iris.Config{
				Level:   iris.Debug,
				Encoder: tt.encoder,
				Output:  iris.WrapWriter(&buf),
			})
			if err != nil {
				t.Fatalf("Failed to create logger with %s: %v", tt.name, err)
			}
			defer func() {
				if err := logger.Close(); err != nil {
					t.Errorf("Failed to close logger: %v", err)
				}
			}()

			logger.Start()
			logger.Info("test message")
			time.Sleep(50 * time.Millisecond)
			if err := logger.Sync(); err != nil {
				t.Errorf("Failed to sync logger: %v", err)
			}

			if buf.Len() == 0 {
				t.Errorf("Expected output with %s, got empty buffer", tt.name)
			}
		})
	}
}

func TestMainFunctionality(t *testing.T) {
	// Test the main workflow similar to main function
	var buf bytes.Buffer

	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Encoder: iris.NewJSONEncoder(),
		Output:  iris.WrapWriter(&buf),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("Failed to close logger: %v", err)
		}
	}()

	// Simulate main function workflow
	logger.Start()

	result := logger.Info("test message", iris.Str("key", "value"))
	if !result {
		t.Error("Expected log operation to succeed")
	}

	// Give time for processing (as in main)
	time.Sleep(100 * time.Millisecond)

	if err := logger.Sync(); err != nil {
		t.Errorf("Failed to sync logger: %v", err)
	}

	// Verify output was generated
	if buf.Len() == 0 {
		t.Error("Expected some output after logging")
	}

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected 'test message' in output, got: %s", output)
	}
}

func TestMain(t *testing.T) {
	// Test the actual main function by capturing stdout
	// We'll redirect stdout temporarily to capture the output
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	os.Stdout = w

	// Run main in a goroutine so we can capture its output
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Main function panicked: %v", r)
			}
			done <- true
		}()
		main()
	}()

	// Wait for main to complete with timeout (reduced for faster CI)
	select {
	case <-done:
		// Success
	case <-time.After(15 * time.Second): // Ridotto da 25 a 15 secondi
		t.Fatal("Test timeout: main() took too long (likely Windows Sync issue)")
	}

	// Close writer and restore stdout
	if err := w.Close(); err != nil {
		t.Errorf("Error closing writer: %v", err)
	}
	os.Stdout = originalStdout

	// Read the captured output
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read output: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Errorf("Error closing reader: %v", err)
	}

	output := string(buf[:n])

	// Verify expected outputs from main function
	expectedMessages := []string{
		"Logger created",
		"Logger started",
		"Log result:",
	}

	for _, expected := range expectedMessages {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected '%s' in main output, got: %s", expected, output)
		}
	}

	// Check that either sync succeeded or had expected error
	syncSuccess := strings.Contains(output, "Logger synced")
	syncWarning := strings.Contains(output, "Warning:")

	if !syncSuccess && !syncWarning {
		t.Errorf("Expected either 'Logger synced' or sync warning in main output, got: %s", output)
	}
}
