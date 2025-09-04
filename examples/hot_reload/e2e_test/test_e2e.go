// test_e2e.go: End-to-end test for Iris hot reload functionality
//
// This comprehensive test verifies:
// 1. Hot reload detects configuration changes
// 2. Log level updates are applied correctly
// 3. Audit trail is generated properly
// 4. Multiple configuration changes work
// 5. Error handling for invalid configurations
//
// Copyright (c) 2025 AGILira
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/agilira/iris"
)

// TestConfig represents the JSON configuration structure
type TestConfig struct {
	Level       string `json:"level"`
	Development bool   `json:"development,omitempty"`
	Encoder     string `json:"encoder,omitempty"`
}

// TestResult captures test outcomes
type TestResult struct {
	TestName string
	Success  bool
	Error    error
	Duration time.Duration
	Details  string
}

// SyncBuffer wraps bytes.Buffer with Sync method for WriteSyncer interface
type SyncBuffer struct {
	*bytes.Buffer
}

// Sync implements the iris.WriteSyncer interface
func (sb *SyncBuffer) Sync() error {
	return nil
}

// NewSyncBuffer creates a new SyncBuffer instance
func NewSyncBuffer() *SyncBuffer {
	return &SyncBuffer{Buffer: &bytes.Buffer{}}
}

// E2ETest orchestrates comprehensive hot reload testing
type E2ETest struct {
	configFile string
	auditFile  string
	logger     *iris.Logger
	watcher    *iris.DynamicConfigWatcher
	output     *SyncBuffer
	results    []TestResult
	startTime  time.Time
}

func runE2ETests() {
	fmt.Println("🧪 Iris Hot Reload - End-to-End Test Suite")
	fmt.Println("==========================================")

	test := NewE2ETest()
	defer test.Cleanup()

	if err := test.Setup(); err != nil {
		fmt.Printf("❌ Setup failed: %v\n", err)
		os.Exit(1)
	}

	test.RunAllTests()
	test.PrintResults()

	if test.HasFailures() {
		os.Exit(1)
	}

	fmt.Println("\n🎉 All tests passed! Hot reload is working perfectly!")
}

// NewE2ETest creates a new E2ETest instance with default configuration
func NewE2ETest() *E2ETest {
	return &E2ETest{
		configFile: "test_config.json",
		auditFile:  "iris-config-audit.jsonl",
		output:     NewSyncBuffer(),
		results:    make([]TestResult, 0),
		startTime:  time.Now(),
	}
}

// Setup initializes the test environment and creates necessary configuration files
func (t *E2ETest) Setup() error {
	fmt.Println("🔧 Setting up test environment...")

	// Clean up any existing files
	if err := os.Remove(t.configFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove existing config file: %v\n", err)
	}
	if err := os.Remove(t.auditFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove existing audit file: %v\n", err)
	}

	// Create initial configuration - use standard Level type
	config := iris.Config{
		Level:   iris.Info,
		Output:  t.output,
		Encoder: iris.NewJSONEncoder(),
	}

	logger, err := iris.New(config)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Start the logger for async processing
	logger.Start()

	t.logger = logger

	// Create initial config file
	if err := t.writeConfig("info"); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Set up watcher using logger's built-in atomic level
	watcher, err := iris.NewDynamicConfigWatcher(t.configFile, logger.AtomicLevel())
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	t.watcher = watcher

	// Start watching
	if err := watcher.Start(); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	fmt.Println("✅ Test environment ready")
	return nil
}

// RunAllTests executes the complete test suite for hot reload functionality
func (t *E2ETest) RunAllTests() {
	fmt.Println("\n🚀 Running comprehensive test suite...")

	t.runTest("Initial Level Check", t.testInitialLevel)
	t.runTest("Hot Reload to Debug", t.testHotReloadDebug)
	t.runTest("Hot Reload to Error", t.testHotReloadError)
	t.runTest("Hot Reload to Warn", t.testHotReloadWarn)
	t.runTest("Invalid Config Fallback", t.testInvalidConfig)
	t.runTest("Rapid Changes", t.testRapidChanges)
	t.runTest("Audit Trail Verification", t.testAuditTrail)
	t.runTest("Log Output Verification", t.testLogOutput)
}

func (t *E2ETest) runTest(name string, testFunc func() error) {
	fmt.Printf("  📋 %s... ", name)
	start := time.Now()

	err := testFunc()
	duration := time.Since(start)

	result := TestResult{
		TestName: name,
		Success:  err == nil,
		Error:    err,
		Duration: duration,
	}

	if err == nil {
		fmt.Printf("✅ (%v)\n", duration.Truncate(time.Millisecond))
	} else {
		fmt.Printf("❌ (%v): %v\n", duration.Truncate(time.Millisecond), err)
	}

	t.results = append(t.results, result)
}

func (t *E2ETest) testInitialLevel() error {
	if t.logger.Level() != iris.Info {
		return fmt.Errorf("expected initial level Info, got %s", t.logger.Level())
	}
	return nil
}

func (t *E2ETest) testHotReloadDebug() error {
	t.output.Reset()

	// Change to debug level
	if err := t.writeConfig("debug"); err != nil {
		return err
	}

	// Wait for reload
	if err := t.waitForReload(iris.Debug, 5*time.Second); err != nil {
		return err
	}

	// Force a sync before checking output
	if err := t.logger.Sync(); err != nil {
		fmt.Printf("Warning: failed to sync logger: %v\n", err)
	}

	// Test debug message appears
	t.logger.Debug("test debug message")
	t.logger.Info("test info message")

	// Force another sync to ensure messages are flushed
	if err := t.logger.Sync(); err != nil {
		fmt.Printf("Warning: failed to sync logger: %v\n", err)
	}

	output := t.output.String()
	fmt.Printf("DEBUG OUTPUT: %q\n", output) // Debug print

	if !strings.Contains(output, "test debug message") {
		return fmt.Errorf("debug message not found in output")
	}
	if !strings.Contains(output, "test info message") {
		return fmt.Errorf("info message not found in output")
	}

	return nil
}

func (t *E2ETest) testHotReloadError() error {
	t.output.Reset()

	// Change to error level
	if err := t.writeConfig("error"); err != nil {
		return err
	}

	// Wait for reload
	if err := t.waitForReload(iris.Error, 5*time.Second); err != nil {
		return err
	}

	// Force a sync before checking output
	if err := t.logger.Sync(); err != nil {
		fmt.Printf("Warning: failed to sync logger: %v\n", err)
	}

	// Test only error messages appear
	t.logger.Debug("should not appear")
	t.logger.Info("should not appear")
	t.logger.Warn("should not appear")
	t.logger.Error("should appear")

	// Force another sync to ensure messages are flushed
	if err := t.logger.Sync(); err != nil {
		fmt.Printf("Warning: failed to sync logger: %v\n", err)
	}

	output := t.output.String()
	fmt.Printf("ERROR OUTPUT: %q\n", output) // Debug print

	if strings.Contains(output, "should not appear") {
		return fmt.Errorf("lower level messages appeared when they shouldn't")
	}
	if !strings.Contains(output, "should appear") {
		return fmt.Errorf("error message not found in output")
	}

	return nil
}

func (t *E2ETest) testHotReloadWarn() error {
	// Change to warn level
	if err := t.writeConfig("warn"); err != nil {
		return err
	}

	// Wait for reload
	return t.waitForReload(iris.Warn, 5*time.Second)
}

func (t *E2ETest) testInvalidConfig() error {
	// First, ensure we're at a known level (warn)
	if err := t.writeConfig("warn"); err != nil {
		return err
	}

	if err := t.waitForReload(iris.Warn, 5*time.Second); err != nil {
		return err
	}

	originalLevel := t.logger.Level()
	fmt.Printf("Original level before invalid config: %s\n", originalLevel)

	// Write invalid config
	invalidConfig := `{"level": "invalid_level"}`
	if err := os.WriteFile(t.configFile, []byte(invalidConfig), 0600); err != nil {
		return err
	}

	// Wait for processing - with invalid config, system should fallback to Info
	time.Sleep(3 * time.Second)

	currentLevel := t.logger.Level()
	fmt.Printf("Level after invalid config: %s\n", currentLevel)

	// System should fallback to Info level (safe default) for invalid config
	if currentLevel != iris.Info {
		return fmt.Errorf("invalid config should fallback to Info level, got %s", currentLevel)
	}

	return nil
}

func (t *E2ETest) testRapidChanges() error {
	levels := []string{"debug", "info", "warn", "error", "info"}
	expectedLevels := []iris.Level{iris.Debug, iris.Info, iris.Warn, iris.Error, iris.Info}

	for i, level := range levels {
		if err := t.writeConfig(level); err != nil {
			return err
		}

		if err := t.waitForReload(expectedLevels[i], 3*time.Second); err != nil {
			return fmt.Errorf("rapid change %d failed: %w", i+1, err)
		}

		// Small delay between changes
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (t *E2ETest) testAuditTrail() error {
	// Ensure we have an audit file
	if _, err := os.Stat(t.auditFile); os.IsNotExist(err) {
		return fmt.Errorf("audit file does not exist: %s", t.auditFile)
	}

	// Read audit file
	content, err := os.ReadFile(t.auditFile)
	if err != nil {
		return fmt.Errorf("failed to read audit file: %w", err)
	}

	// Check for audit entries
	lines := strings.Split(string(content), "\n")
	auditEntries := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			auditEntries++
		}
	}

	if auditEntries == 0 {
		return fmt.Errorf("no audit entries found")
	}

	fmt.Printf(" [%d audit entries]", auditEntries)
	return nil
}

func (t *E2ETest) testLogOutput() error {
	t.output.Reset()

	// Ensure we're at info level
	if err := t.writeConfig("info"); err != nil {
		return err
	}

	if err := t.waitForReload(iris.Info, 3*time.Second); err != nil {
		return err
	}

	// Generate test messages
	t.logger.Debug("debug message")
	t.logger.Info("info message", iris.String("key", "value"))
	t.logger.Warn("warn message")
	t.logger.Error("error message")

	// Force sync to ensure all messages are flushed
	if err := t.logger.Sync(); err != nil {
		fmt.Printf("Warning: failed to sync logger: %v\n", err)
	}

	output := t.output.String()
	fmt.Printf("LOG OUTPUT: %q\n", output) // Debug print

	// Debug should not appear
	if strings.Contains(output, "debug message") {
		return fmt.Errorf("debug message appeared at info level")
	}

	// Others should appear
	expectedMessages := []string{"info message", "warn message", "error message"}
	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			return fmt.Errorf("expected message not found: %s", msg)
		}
	}

	// Check JSON structure
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line != "" {
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
				return fmt.Errorf("invalid JSON log entry: %w", err)
			}
		}
	}

	return nil
}

func (t *E2ETest) writeConfig(level string) error {
	config := TestConfig{
		Level:   level,
		Encoder: "json",
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(t.configFile, data, 0600)
}

func (t *E2ETest) waitForReload(expectedLevel iris.Level, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if t.logger.Level() == expectedLevel {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for level change to %s (current: %s)",
		expectedLevel, t.logger.Level())
}

// PrintResults displays a comprehensive summary of all test results
func (t *E2ETest) PrintResults() {
	fmt.Println("\n📊 Test Results Summary")
	fmt.Println("======================")

	passed := 0
	failed := 0
	totalDuration := time.Duration(0)

	for _, result := range t.results {
		if result.Success {
			passed++
		} else {
			failed++
		}
		totalDuration += result.Duration
	}

	fmt.Printf("📈 Total Tests: %d\n", len(t.results))
	fmt.Printf("✅ Passed: %d\n", passed)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("⏱️  Total Duration: %v\n", totalDuration.Truncate(time.Millisecond))
	fmt.Printf("🎯 Success Rate: %.1f%%\n", float64(passed)/float64(len(t.results))*100)

	if failed > 0 {
		fmt.Println("\n❌ Failed Tests:")
		for _, result := range t.results {
			if !result.Success {
				fmt.Printf("   • %s: %v\n", result.TestName, result.Error)
			}
		}
	}

	// Print audit file contents for verification
	if content, err := os.ReadFile(t.auditFile); err == nil {
		fmt.Println("\n📋 Audit Trail Contents:")
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		lineNum := 1
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				fmt.Printf("   %d: %s\n", lineNum, line)
				lineNum++
			}
		}
	}
}

// HasFailures returns true if any test in the suite failed
func (t *E2ETest) HasFailures() bool {
	for _, result := range t.results {
		if !result.Success {
			return true
		}
	}
	return false
}

// Cleanup removes temporary files and resources created during testing
func (t *E2ETest) Cleanup() {
	fmt.Println("\n🧹 Cleaning up test environment...")

	if t.watcher != nil {
		if err := t.watcher.Stop(); err != nil {
			fmt.Printf("Warning: failed to stop watcher: %v\n", err)
		}
	}

	if t.logger != nil {
		if err := t.logger.Sync(); err != nil {
			fmt.Printf("Warning: failed to sync logger: %v\n", err)
		}
		if err := t.logger.Close(); err != nil {
			fmt.Printf("Warning: failed to close logger: %v\n", err)
		}
	}

	// Clean up test files
	if err := os.Remove(t.configFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove config file: %v\n", err)
	}
	// Keep audit file for inspection

	fmt.Println("✅ Cleanup complete")
}
