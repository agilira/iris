// iris_test.go: Comprehensive tests for Iris Logger implementation
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

// bufferedSyncer wraps a bytes.Buffer to implement WriteSyncer
type bufferedSyncer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (bs *bufferedSyncer) Write(p []byte) (n int, err error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.buf.Write(p)
}

func (bs *bufferedSyncer) Sync() error {
	return nil
}

// Helper function for safe logger closing in test
func safeCloseIrisLogger(t *testing.T, logger *Logger) {
	if logger != nil {
		if err := logger.Close(); err != nil {
			t.Logf("Failed to close logger: %v", err)
		}
	}
}

func (bs *bufferedSyncer) String() string {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.buf.String()
}

// TestLoggerBasicOperations verifies core logger functionality
func TestLoggerBasicOperations(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	logger.Start()

	// Test basic logging
	logger.Debug("debug message", Str("key", "value"))
	logger.Info("info message", Int("count", 42))
	logger.Warn("warn message", Bool("flag", true))
	logger.Error("error message", Float64("value", 3.14))

	// Give time for async processing
	time.Sleep(50 * time.Millisecond)

	// Ensure all records are processed
	_ = logger.Sync()

	output := buf.String()
	t.Logf("Output: %q", output)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}

	// Verify JSON structure
	for i, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}

		if _, ok := logEntry["ts"]; !ok {
			t.Errorf("Line %d missing timestamp", i+1)
		}
		if _, ok := logEntry["level"]; !ok {
			t.Errorf("Line %d missing level", i+1)
		}
		if _, ok := logEntry["msg"]; !ok {
			t.Errorf("Line %d missing message", i+1)
		}
	}
}

// TestLoggerLevelFiltering verifies level filtering works correctly
func TestLoggerLevelFiltering(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Warn,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	// Start the logger for async processing
	logger.Start()

	// Only Warn and Error should be logged
	logger.Debug("debug - should not appear")
	logger.Info("info - should not appear")
	logger.Warn("warn - should appear")
	logger.Error("error - should appear")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines with Warn level, got %d", len(lines))
	}

	// Verify the right messages appear
	if !strings.Contains(output, "warn - should appear") {
		t.Error("Warn message not found in output")
	}
	if !strings.Contains(output, "error - should appear") {
		t.Error("Error message not found in output")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("Debug/Info messages found in output when they should be filtered")
	}
}

// TestLoggerWithFields verifies the With() method for adding base fields
func TestLoggerWithFields(t *testing.T) {
	buf := &bufferedSyncer{}
	baseLogger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, baseLogger)

	// Start the logger for async processing
	baseLogger.Start()

	// Create child logger with base fields
	childLogger := baseLogger.With(
		Str("service", "test-service"),
		Str("version", "1.0.0"),
	)

	childLogger.Info("test message", Str("extra", "field"))

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Verify base fields are present
	if !strings.Contains(output, `"service":"test-service"`) {
		t.Error("Base field 'service' not found in output")
	}
	if !strings.Contains(output, `"version":"1.0.0"`) {
		t.Error("Base field 'version' not found in output")
	}
	if !strings.Contains(output, `"extra":"field"`) {
		t.Error("Extra field not found in output")
	}
}

// TestLoggerNamed verifies the Named() method
func TestLoggerNamed(t *testing.T) {
	buf := &bufferedSyncer{}
	baseLogger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, baseLogger)

	// Start the logger for async processing
	baseLogger.Start()

	// Create named logger
	namedLogger := baseLogger.Named("test-component")
	namedLogger.Info("test message")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Verify logger name is present
	if !strings.Contains(output, `"logger":"test-component"`) {
		t.Error("Logger name not found in output")
	}
}

// TestLoggerSetLevel verifies dynamic level changes
func TestLoggerSetLevel(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	// Start the logger for async processing
	logger.Start()

	// Initial level allows debug
	logger.Debug("debug1")

	// Change to Info level
	logger.SetLevel(Info)
	logger.Debug("debug2 - should not appear")
	logger.Info("info1")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	// Verify level getter
	if logger.Level() != Info {
		t.Errorf("Expected level Info, got %v", logger.Level())
	}

	output := buf.String()

	if !strings.Contains(output, "debug1") {
		t.Error("First debug message should appear")
	}
	if strings.Contains(output, "debug2") {
		t.Error("Second debug message should not appear after level change")
	}
	if !strings.Contains(output, "info1") {
		t.Error("Info message should appear")
	}
}

// TestLoggerConcurrency verifies thread safety
func TestLoggerConcurrency(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	// Start the logger for async processing
	logger.Start()

	const numGoroutines = 10
	const messagesPerGoroutine = 5
	var wg sync.WaitGroup

	// Launch concurrent loggers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("concurrent message",
					Int("goroutine", id),
					Int("message", j))
			}
		}(i)
	}

	wg.Wait()

	// Wait longer for async processing of many messages
	time.Sleep(200 * time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	expectedLines := numGoroutines * messagesPerGoroutine
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify all lines are valid JSON
	for i, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}
	}
}

// TestLoggerHooks verifies hook functionality
func TestLoggerHooks(t *testing.T) {
	buf := &bufferedSyncer{}
	var hookCalled bool
	var hookRecord *Record

	hook := func(rec *Record) {
		hookCalled = true
		hookRecord = rec
	}

	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	}, WithHook(hook))
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	// Start the logger for async processing
	logger.Start()

	logger.Info("test hook message")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	if !hookCalled {
		t.Error("Hook was not called")
	}
	if hookRecord == nil {
		t.Error("Hook record is nil")
	}
}

// TestLoggerDevelopmentMode verifies development mode functionality
func TestLoggerDevelopmentMode(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Debug,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	}, Development())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	// Start the logger for async processing
	logger.Start()

	logger.Error("test error")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// In development mode, stack traces should be included
	if !strings.Contains(output, `"stack":`) {
		t.Error("Stack trace not found in development mode output")
	}
}

// TestLoggerWithCallerInfo verifies caller information capture
func TestLoggerWithCallerInfo(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	}, WithCaller())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	// Start the logger for async processing
	logger.Start()

	logger.Info("test message")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Debug: print the actual output
	t.Logf("Actual output: %s", output)

	// Caller info should be present
	if !strings.Contains(output, `"caller":`) {
		t.Error("Caller information not found in output")
	}
	if !strings.Contains(output, "iris_test.go") {
		t.Error("Source file not found in caller info")
	}
}

// TestShortCaller tests the shortCaller function with various path scenarios
func TestShortCaller(t *testing.T) {
	tests := []struct {
		name         string
		setupTest    func() (string, bool)
		expectOK     bool
		expectFormat string
	}{
		{
			name: "Valid_Caller",
			setupTest: func() (string, bool) {
				// This will call shortCaller with skip=1 to get this test function
				return shortCaller(1)
			},
			expectOK:     true,
			expectFormat: "iris_test.go:",
		},
		{
			name: "Invalid_Skip_Level",
			setupTest: func() (string, bool) {
				// Use a very high skip value to trigger the !ok case
				return shortCaller(100)
			},
			expectOK:     false,
			expectFormat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caller, ok := tt.setupTest()

			if ok != tt.expectOK {
				t.Errorf("Expected ok=%v, got ok=%v", tt.expectOK, ok)
			}

			if tt.expectOK {
				if !strings.Contains(caller, tt.expectFormat) {
					t.Errorf("Expected caller to contain %q, got %q", tt.expectFormat, caller)
				}
				// Should contain line number
				if !strings.Contains(caller, ":") {
					t.Error("Expected caller to contain line number separator ':'")
				}
			} else {
				if caller != tt.expectFormat {
					t.Errorf("Expected empty caller %q, got %q", tt.expectFormat, caller)
				}
			}
		})
	}
}

// TestLoggerStart tests the Start method with various scenarios
func TestLoggerStart(t *testing.T) {
	buf := &bufferedSyncer{}
	logger, err := New(Config{
		Level:   Info,
		Encoder: NewJSONEncoder(),
		Output:  buf,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseIrisLogger(t, logger)

	// Test multiple Start calls (should be idempotent)
	logger.Start()
	logger.Start() // Second call should be ignored
	logger.Start() // Third call should be ignored

	// Verify logger is working after multiple Start calls
	logger.Info("test message after multiple starts")

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "test message after multiple starts") {
		t.Error("Logger not working after multiple Start calls")
	}
}

// TestDPanic tests the DPanic method in both development and production modes
func TestDPanic(t *testing.T) {
	tests := []struct {
		name        string
		development bool
		expectPanic bool
	}{
		{
			name:        "Development_Mode_Should_Panic",
			development: true,
			expectPanic: true,
		},
		{
			name:        "Production_Mode_No_Panic",
			development: false,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bufferedSyncer{}

			var config Config
			if tt.development {
				config = Config{
					Level:   Debug,
					Encoder: NewJSONEncoder(),
					Output:  buf,
				}
			} else {
				config = Config{
					Level:   Debug,
					Encoder: NewJSONEncoder(),
					Output:  buf,
				}
			}

			logger, err := New(config, func() Option {
				if tt.development {
					return Development()
				}
				// Return a no-op option for production (default)
				return Option(func(o *loggerOptions) {})
			}())
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer safeCloseIrisLogger(t, logger)
			logger.Start()

			// Test the DPanic function
			defer func() {
				r := recover()
				if tt.expectPanic && r == nil {
					t.Error("Expected DPanic to panic in development mode, but it didn't")
				}
				if !tt.expectPanic && r != nil {
					t.Errorf("Expected DPanic not to panic in production mode, but it panicked with: %v", r)
				}
			}()

			logger.DPanic("test dpanic message", Str("mode", "test"))

			// If we reach here, no panic occurred
			if tt.expectPanic {
				t.Error("DPanic should have panicked but didn't")
			}
		})
	}
}

// TestSync tests the Sync method with different output types
func TestSync(t *testing.T) {
	tests := []struct {
		name     string
		setupOut func() WriteSyncer
		testFunc func(t *testing.T, logger *Logger, out WriteSyncer)
	}{
		{
			name: "Sync_With_Syncer_Output",
			setupOut: func() WriteSyncer {
				return &bufferedSyncer{}
			},
			testFunc: func(t *testing.T, logger *Logger, out WriteSyncer) {
				// Log something
				logger.Info("test sync message")

				// Wait for async processing before sync
				time.Sleep(50 * time.Millisecond)

				// Call Sync
				err := logger.Sync()
				if err != nil {
					t.Errorf("Sync returned error: %v", err)
				}

				// Wait a bit more after sync
				time.Sleep(10 * time.Millisecond)

				// Verify output was synced
				if bs, ok := out.(*bufferedSyncer); ok {
					output := bs.String()
					if !strings.Contains(output, "test sync message") {
						t.Error("Message not found after sync")
					}
				}
			},
		},
		{
			name: "Sync_With_Non_Syncer_Output",
			setupOut: func() WriteSyncer {
				// Create a simple writer that doesn't implement Sync
				return &simpleWriter{buf: &bytes.Buffer{}}
			},
			testFunc: func(t *testing.T, logger *Logger, out WriteSyncer) {
				// Log something
				logger.Info("test sync message")

				// Call Sync - should not return error even if output doesn't support sync
				err := logger.Sync()
				if err != nil {
					t.Errorf("Sync should not return error for non-syncer output, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tt.setupOut()

			logger, err := New(Config{
				Level:   Info,
				Encoder: NewJSONEncoder(),
				Output:  out,
			})
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer safeCloseIrisLogger(t, logger)
			logger.Start()

			// Wait for logger to start
			time.Sleep(10 * time.Millisecond)

			// Run the specific test
			tt.testFunc(t, logger, out)
		})
	}
}

// simpleWriter implements WriteSyncer without the Sync method for testing
type simpleWriter struct {
	buf *bytes.Buffer
}

func (sw *simpleWriter) Write(p []byte) (n int, err error) {
	return sw.buf.Write(p)
}

func (sw *simpleWriter) Sync() error {
	return nil
}
