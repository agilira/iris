package iris

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestStackTraceSupport(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:           DebugLevel,
		Format:          JSONFormat,
		Writer:          &buf,
		StackTraceLevel: ErrorLevel, // Capture stack traces for Error and above
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log messages at different levels
	logger.Info("Info message")    // No stack trace
	logger.Warn("Warning message") // No stack trace
	logger.Error("Error message")  // Should have stack trace

	time.Sleep(50 * time.Millisecond)
	logger.ring.Flush()
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	t.Logf("Output: %s", output)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 log lines, got %d", len(lines))
	}

	// Check that Info and Warn don't have stack traces
	if strings.Contains(lines[0], "stacktrace") {
		t.Error("Info message should not have stack trace")
	}
	if strings.Contains(lines[1], "stacktrace") {
		t.Error("Warn message should not have stack trace")
	}

	// Check that Error has stack trace
	if !strings.Contains(lines[2], "stacktrace") {
		t.Error("Error message should have stack trace")
	}

	// Verify stack trace contains function names from the Go runtime
	if !strings.Contains(output, "testing.tRunner") {
		t.Error("Stack trace should contain testing runtime function names")
	}
}

func TestStackTraceDisabled(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  DebugLevel,
		Format: JSONFormat,
		Writer: &buf,
		// StackTraceLevel not set (defaults to 0 = disabled)
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Error("Error message") // Should not have stack trace

	time.Sleep(50 * time.Millisecond)
	logger.ring.Flush()
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	t.Logf("Output: %s", output)

	// Should not contain stack trace
	if strings.Contains(output, "stacktrace") {
		t.Error("Stack trace should be disabled when StackTraceLevel is 0")
	}
}

func TestStackTraceConsoleFormat(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:           DebugLevel,
		Format:          ConsoleFormat,
		Writer:          &buf,
		StackTraceLevel: ErrorLevel,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Error("Error with stack trace")

	time.Sleep(50 * time.Millisecond)
	logger.ring.Flush()
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	t.Logf("Console output: %s", output)

	// Should contain stack trace in console format
	if !strings.Contains(output, "Stack trace:") {
		t.Error("Console format should show stack trace")
	}

	if !strings.Contains(output, "testing.tRunner") {
		t.Error("Stack trace should contain testing runtime function names")
	}
}
