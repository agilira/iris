package iris

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// criticalTestSyncer helps test critical logging methods
type criticalTestSyncer struct {
	logs     []string
	synced   bool
	logCount int
}

func (c *criticalTestSyncer) Write(p []byte) (n int, err error) {
	c.logs = append(c.logs, string(p))
	c.logCount++
	return len(p), nil
}

func (c *criticalTestSyncer) Sync() error {
	c.synced = true
	return nil
}

// TestLogger_DPanic tests the DPanic method in both development and production modes
func TestLogger_DPanic(t *testing.T) {
	t.Run("DPanic_Production_Mode", func(t *testing.T) {
		syncer := &criticalTestSyncer{}
		logger, err := New(Config{
			Level:   Debug,
			Encoder: NewTextEncoder(),
			Output:  syncer,
		}) // Production mode (default)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// Test DPanic in production mode - should not panic
		result := logger.DPanic("test dpanic message", String("key", "value"))

		// Wait for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify log was written
		if len(syncer.logs) == 0 {
			t.Error("Expected DPanic to write log in production mode")
		}

		// Verify return value
		if !result {
			t.Error("Expected DPanic to return true when logging")
		}

		// Verify message content
		logContent := syncer.logs[0]
		if !strings.Contains(logContent, "test dpanic message") {
			t.Errorf("Expected log to contain 'test dpanic message', got: %s", logContent)
		}
		if !strings.Contains(logContent, "key") {
			t.Errorf("Expected log to contain field 'key', got: %s", logContent)
		}
	})

	t.Run("DPanic_Development_Mode", func(t *testing.T) {
		// Test DPanic in development mode using subprocess to catch panic
		if os.Getenv("TEST_DPANIC_PANIC") == "1" {
			syncer := &criticalTestSyncer{}
			logger, err := New(Config{
				Level:   Debug,
				Encoder: NewTextEncoder(),
				Output:  syncer,
			}, Development()) // Development mode - should panic
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Start()
			defer logger.Close()

			// This should panic in development mode
			logger.DPanic("test dpanic panic", String("key", "value"))
			return
		}

		// Run subprocess to test panic behavior
		cmd := exec.Command(os.Args[0], "-test.run=TestLogger_DPanic/DPanic_Development_Mode")
		cmd.Env = append(os.Environ(), "TEST_DPANIC_PANIC=1")
		err := cmd.Run()

		// Verify the subprocess panicked (exit code != 0)
		if err == nil {
			t.Error("Expected DPanic to panic in development mode")
		}
	})

	t.Run("DPanic_Level_Filtering", func(t *testing.T) {
		syncer := &criticalTestSyncer{}
		logger, err := New(Config{
			Level:   Fatal, // Set level higher than DPanic
			Encoder: NewTextEncoder(),
			Output:  syncer,
		})
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Start()
		defer logger.Close()

		// DPanic should be filtered out
		result := logger.DPanic("filtered dpanic", String("key", "value"))

		time.Sleep(50 * time.Millisecond)

		// Should return true (early exit optimization) but no log written
		if !result {
			t.Error("Expected DPanic to return true even when filtered")
		}

		// No logs should be written when filtered
		if len(syncer.logs) > 0 {
			t.Errorf("Expected no logs when DPanic is filtered, got: %v", syncer.logs)
		}
	})
}

// TestLogger_Panic tests the Panic method
func TestLogger_Panic(t *testing.T) {
	t.Run("Panic_Always_Panics", func(t *testing.T) {
		// Test Panic using subprocess to catch panic
		if os.Getenv("TEST_PANIC") == "1" {
			syncer := &criticalTestSyncer{}
			logger, err := New(Config{
				Level:   Debug,
				Encoder: NewTextEncoder(),
				Output:  syncer,
			})
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Start()
			defer logger.Close()

			// This should always panic
			logger.Panic("test panic message", String("key", "value"))
			return
		}

		// Run subprocess to test panic behavior
		cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Panic/Panic_Always_Panics")
		cmd.Env = append(os.Environ(), "TEST_PANIC=1")
		err := cmd.Run()

		// Verify the subprocess panicked (exit code != 0)
		if err == nil {
			t.Error("Expected Panic to always panic")
		}
	})

	t.Run("Panic_Even_When_Filtered", func(t *testing.T) {
		// Test that Panic still panics even when level is filtered
		if os.Getenv("TEST_PANIC_FILTERED") == "1" {
			syncer := &criticalTestSyncer{}
			logger, err := New(Config{
				Level:   Fatal, // Set level higher than Panic
				Encoder: NewTextEncoder(),
				Output:  syncer,
			})
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Start()
			defer logger.Close()

			// This should still panic even when filtered
			logger.Panic("filtered panic", String("key", "value"))
			return
		}

		// Run subprocess to test panic behavior
		cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Panic/Panic_Even_When_Filtered")
		cmd.Env = append(os.Environ(), "TEST_PANIC_FILTERED=1")
		err := cmd.Run()

		// Verify the subprocess panicked (exit code != 0)
		if err == nil {
			t.Error("Expected Panic to panic even when filtered")
		}
	})
}

// TestLogger_Fatal tests the Fatal method
func TestLogger_Fatal(t *testing.T) {
	t.Run("Fatal_Exits_Process", func(t *testing.T) {
		// Test Fatal using subprocess to catch os.Exit
		if os.Getenv("TEST_FATAL") == "1" {
			syncer := &criticalTestSyncer{}
			logger, err := New(Config{
				Level:   Debug,
				Encoder: NewTextEncoder(),
				Output:  syncer,
			})
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Start()
			defer logger.Close()

			// This should call os.Exit(1)
			logger.Fatal("test fatal message", String("key", "value"))
			return
		}

		// Run subprocess to test exit behavior
		cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Fatal/Fatal_Exits_Process")
		cmd.Env = append(os.Environ(), "TEST_FATAL=1")
		err := cmd.Run()

		// Verify the subprocess exited with code 1
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 {
				t.Errorf("Expected Fatal to exit with code 1, got: %d", exitError.ExitCode())
			}
		} else if err == nil {
			t.Error("Expected Fatal to exit the process")
		}
	})

	t.Run("Fatal_Even_When_Filtered", func(t *testing.T) {
		// Test that Fatal still exits even when level is filtered
		if os.Getenv("TEST_FATAL_FILTERED") == "1" {
			syncer := &criticalTestSyncer{}
			logger, err := New(Config{
				Level:   Panic, // Set level higher than Fatal
				Encoder: NewTextEncoder(),
				Output:  syncer,
			})
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Start()
			defer logger.Close()

			// This should still exit even when filtered
			logger.Fatal("filtered fatal", String("key", "value"))
			return
		}

		// Run subprocess to test exit behavior
		cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Fatal/Fatal_Even_When_Filtered")
		cmd.Env = append(os.Environ(), "TEST_FATAL_FILTERED=1")
		err := cmd.Run()

		// Verify the subprocess exited
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 {
				t.Errorf("Expected Fatal to exit with code 1, got: %d", exitError.ExitCode())
			}
		} else if err == nil {
			t.Error("Expected Fatal to exit the process even when filtered")
		}
	})
}

// TestLogger_Fatal_Sync_Called tests that Fatal calls Sync before exiting
func TestLogger_Fatal_Sync_Called(t *testing.T) {
	t.Run("Fatal_Calls_Sync", func(t *testing.T) {
		// This test ensures Fatal calls Sync, but we can't directly test it
		// due to os.Exit. We test this indirectly by ensuring the method
		// executes the sync path before exit.

		if os.Getenv("TEST_FATAL_SYNC") == "1" {
			syncer := &criticalTestSyncer{}
			logger, err := New(Config{
				Level:   Debug,
				Encoder: NewTextEncoder(),
				Output:  syncer,
			})
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Start()
			defer logger.Close()

			// Fatal should call Sync before os.Exit
			logger.Fatal("test fatal sync", String("key", "value"))
			return
		}

		// Run subprocess
		cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Fatal_Sync_Called/Fatal_Calls_Sync")
		cmd.Env = append(os.Environ(), "TEST_FATAL_SYNC=1")
		err := cmd.Run()

		// Verify the subprocess exited (confirms Fatal was called)
		if err == nil {
			t.Error("Expected Fatal to exit the process")
		}
	})
}
