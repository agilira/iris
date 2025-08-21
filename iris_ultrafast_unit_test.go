package iris

import (
	"bytes"
	"testing"
)

// TestUltraFast tests the UltraFast() method of Logger
func TestUltraFast(t *testing.T) {
	t.Run("UltraFastWithBinaryFormat", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:            DebugLevel,
			Writer:           &buf,
			Format:           BinaryFormat,
			BufferSize:       1024,
			BatchSize:        32,
			EnableCaller:     false, // Required for ultra-fast
			DisableTimestamp: true,  // Required for ultra-fast
		}

		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Should be ultra-fast with binary format, no caller, no timestamp
		if !logger.UltraFast() {
			t.Error("Expected logger to be in UltraFast mode with BinaryFormat, no caller, no timestamp")
		}
	})

	t.Run("UltraFastWithJSONFormat", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:            DebugLevel,
			Writer:           &buf,
			Format:           JSONFormat,
			BufferSize:       1024,
			BatchSize:        32,
			EnableCaller:     false, // Required for ultra-fast
			DisableTimestamp: true,  // Required for ultra-fast
		}

		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Should be ultra-fast with JSON format, no caller, no timestamp
		if !logger.UltraFast() {
			t.Error("Expected logger to be in UltraFast mode with JSONFormat, no caller, no timestamp")
		}
	})

	t.Run("NotUltraFastWithCaller", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:            DebugLevel,
			Writer:           &buf,
			Format:           BinaryFormat,
			BufferSize:       1024,
			BatchSize:        32,
			EnableCaller:     true, // This disables ultra-fast
			DisableTimestamp: true,
		}

		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Should NOT be ultra-fast with caller enabled
		if logger.UltraFast() {
			t.Error("Expected logger to NOT be in UltraFast mode with caller enabled")
		}
	})

	t.Run("NotUltraFastWithTimestamp", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:            DebugLevel,
			Writer:           &buf,
			Format:           BinaryFormat,
			BufferSize:       1024,
			BatchSize:        32,
			EnableCaller:     false,
			DisableTimestamp: false, // This disables ultra-fast
		}

		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Should NOT be ultra-fast with timestamp enabled
		if logger.UltraFast() {
			t.Error("Expected logger to NOT be in UltraFast mode with timestamp enabled")
		}
	})

	t.Run("NotUltraFastWithConsoleFormat", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:            DebugLevel,
			Writer:           &buf,
			Format:           ConsoleFormat, // This disables ultra-fast
			BufferSize:       1024,
			BatchSize:        32,
			EnableCaller:     false,
			DisableTimestamp: true,
		}

		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Should NOT be ultra-fast with console format
		if logger.UltraFast() {
			t.Error("Expected logger to NOT be in UltraFast mode with ConsoleFormat")
		}
	})

	t.Run("NotUltraFastWithTextFormat", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Level:            DebugLevel,
			Writer:           &buf,
			Format:           FastTextFormat, // This disables ultra-fast
			BufferSize:       1024,
			BatchSize:        32,
			EnableCaller:     false,
			DisableTimestamp: true,
		}

		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Should NOT be ultra-fast with text format
		if logger.UltraFast() {
			t.Error("Expected logger to NOT be in UltraFast mode with FastTextFormat")
		}
	})

	t.Run("UltraFastDetectionLogic", func(t *testing.T) {
		// Test the complete boolean logic for ultra-fast detection
		testCases := []struct {
			name              string
			enableCaller      bool
			disableTimestamp  bool
			format            Format
			expectedUltraFast bool
		}{
			{"Perfect UltraFast Binary", false, true, BinaryFormat, true},
			{"Perfect UltraFast JSON", false, true, JSONFormat, true},
			{"Caller enabled", true, true, BinaryFormat, false},
			{"Timestamp enabled", false, false, BinaryFormat, false},
			{"Console format", false, true, ConsoleFormat, false},
			{"Text format", false, true, FastTextFormat, false},
			{"All disabled", true, false, ConsoleFormat, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var buf bytes.Buffer
				config := Config{
					Level:            DebugLevel,
					Writer:           &buf,
					Format:           tc.format,
					BufferSize:       1024,
					BatchSize:        32,
					EnableCaller:     tc.enableCaller,
					DisableTimestamp: tc.disableTimestamp,
				}

				logger, err := New(config)
				if err != nil {
					t.Fatalf("Failed to create logger: %v", err)
				}
				defer logger.Close()

				result := logger.UltraFast()
				if result != tc.expectedUltraFast {
					t.Errorf("Expected UltraFast()=%t, got %t for case: %s",
						tc.expectedUltraFast, result, tc.name)
				}
			})
		}
	})

	t.Run("UltraFastWithDifferentLevels", func(t *testing.T) {
		// Test that ultra-fast detection is independent of log level
		levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}

		for _, level := range levels {
			var buf bytes.Buffer
			config := Config{
				Level:            level,
				Writer:           &buf,
				Format:           BinaryFormat,
				BufferSize:       1024,
				BatchSize:        32,
				EnableCaller:     false,
				DisableTimestamp: true,
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger for level %v: %v", level, err)
			}

			if !logger.UltraFast() {
				t.Errorf("Expected UltraFast mode regardless of level, failed for %v", level)
			}

			logger.Close()
		}
	})
}
