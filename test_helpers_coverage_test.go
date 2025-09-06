// test_helpers_coverage_test.go: Test helpers functionality in Iris
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestUnit_CIFriendlyTimeout_SmartAdaptation tests timeout adaptation logic
func TestUnit_CIFriendlyTimeout_SmartAdaptation(t *testing.T) {
	testCases := []struct {
		name           string
		ciEnvironment  bool
		normalTimeout  time.Duration
		expectedFactor int
	}{
		{
			name:           "LocalDevelopment_FastExecution",
			ciEnvironment:  false,
			normalTimeout:  100 * time.Millisecond,
			expectedFactor: 1,
		},
		{
			name:           "CIEnvironment_ResourceConstrained",
			ciEnvironment:  true,
			normalTimeout:  100 * time.Millisecond,
			expectedFactor: 3,
		},
		{
			name:           "CIEnvironment_LongOperation",
			ciEnvironment:  true,
			normalTimeout:  5 * time.Second,
			expectedFactor: 3,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment
			if tc.ciEnvironment {
				t.Setenv("CI", "true")
			} else {
				t.Setenv("CI", "")
			}

			// Execute function under test
			actualTimeout := CIFriendlyTimeout(tc.normalTimeout)

			// Validate timeout scaling
			expectedTimeout := tc.normalTimeout * time.Duration(tc.expectedFactor)
			if actualTimeout != expectedTimeout {
				t.Errorf("Timeout should be scaled by factor %d in CI=%v environment. Expected: %v, Got: %v",
					tc.expectedFactor, tc.ciEnvironment, expectedTimeout, actualTimeout)
			}
		})
	}
}

// TestUnit_CIFriendlyRetryCount_IntelligentBackoff tests retry count adaptation
func TestUnit_CIFriendlyRetryCount_IntelligentBackoff(t *testing.T) {
	testScenarios := []struct {
		name               string
		ciEnvironment      bool
		normalRetries      int
		expectedMultiplier int
		rationale          string
	}{
		{
			name:               "LocalDev_OptimalRetries",
			ciEnvironment:      false,
			normalRetries:      3,
			expectedMultiplier: 1,
			rationale:          "Local environments have predictable resources",
		},
		{
			name:               "CI_SchedulerVariability",
			ciEnvironment:      true,
			normalRetries:      3,
			expectedMultiplier: 2,
			rationale:          "CI environments need more retries due to scheduler unpredictability",
		},
		{
			name:               "CI_SingleRetry_Boosted",
			ciEnvironment:      true,
			normalRetries:      1,
			expectedMultiplier: 2,
			rationale:          "Even single retries should be doubled in CI",
		},
	}

	for _, scenario := range testScenarios {
		scenario := scenario // Capture range variable
		t.Run(scenario.name, func(t *testing.T) {
			// Configure environment
			if scenario.ciEnvironment {
				t.Setenv("CI", "true")
			} else {
				t.Setenv("CI", "")
			}

			// Test retry count calculation
			actualRetries := CIFriendlyRetryCount(scenario.normalRetries)

			expectedRetries := scenario.normalRetries * scenario.expectedMultiplier
			if actualRetries != expectedRetries {
				t.Errorf("Retry count adaptation failed: %s. Expected: %d, Got: %d",
					scenario.rationale, expectedRetries, actualRetries)
			}
		})
	}
}

// TestUnit_CIFriendlySleep_AdaptiveTiming tests sleep duration adaptation
func TestUnit_CIFriendlySleep_AdaptiveTiming(t *testing.T) {
	testCases := []struct {
		name          string
		ciEnvironment bool
		sleepDuration time.Duration
		tolerance     time.Duration
	}{
		{
			name:          "LocalDev_NormalTiming",
			ciEnvironment: false,
			sleepDuration: 50 * time.Millisecond,
			tolerance:     15 * time.Millisecond,
		},
		{
			name:          "CI_ExtendedTiming",
			ciEnvironment: true,
			sleepDuration: 50 * time.Millisecond,
			tolerance:     25 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Environment setup
			if tc.ciEnvironment {
				t.Setenv("CI", "true")
			} else {
				t.Setenv("CI", "")
			}

			// Measure actual sleep duration
			startTime := time.Now()
			CIFriendlySleep(tc.sleepDuration)
			actualDuration := time.Since(startTime)

			// Calculate expected duration
			var expectedDuration time.Duration
			if tc.ciEnvironment {
				expectedDuration = tc.sleepDuration * 2
			} else {
				expectedDuration = tc.sleepDuration
			}

			// Validate sleep duration within tolerance
			diff := actualDuration - expectedDuration
			if diff < 0 {
				diff = -diff
			}
			if diff > tc.tolerance {
				t.Errorf("Sleep duration should be %v (±%v) but was %v",
					expectedDuration, tc.tolerance, actualDuration)
			}
		})
	}
}

// TestUnit_TestConfig_PlatformAware tests platform-specific configuration
func TestUnit_TestConfig_PlatformAware(t *testing.T) {
	t.Parallel()

	// Execute function under test
	config := TestConfig()

	// Validate core configuration components
	if config.Level != Debug {
		t.Errorf("Test config should use Debug level, got: %v", config.Level)
	}
	if config.Output == nil {
		t.Error("Output should be configured")
	}
	if config.Encoder == nil {
		t.Error("JSON encoder should be configured")
	}

	// Validate platform-specific capacity settings
	switch runtime.GOOS {
	case "darwin":
		if config.Capacity != 256 {
			t.Errorf("macOS should use smaller capacity (256) for memory compatibility, got: %d", config.Capacity)
		}
	default:
		if config.Capacity != 1024 {
			t.Errorf("Non-macOS platforms should use standard capacity (1024), got: %d", config.Capacity)
		}
	}

	// Verify config produces functional logger
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Test config should create valid logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger should be created")
	}

	// Test logger functionality
	logger.Info("TestConfig validation successful")
}

// TestUnit_TestConfigWithOutput_OutputValidation tests custom output configuration
func TestUnit_TestConfigWithOutput_OutputValidation(t *testing.T) {
	t.Parallel()

	// Create test output buffer with WriteSyncer interface
	var outputBuffer bytes.Buffer
	output := WrapWriter(&outputBuffer)

	// Execute function under test
	config := TestConfigWithOutput(output)

	// Validate configuration structure
	if config.Level != Debug {
		t.Errorf("Should use Debug level, got: %v", config.Level)
	}
	if config.Output == nil {
		t.Error("Custom output should be set")
	}
	if config.Encoder == nil {
		t.Error("JSON encoder should be configured")
	}

	// Validate platform-aware capacity
	switch runtime.GOOS {
	case "darwin":
		if config.Capacity != 256 {
			t.Errorf("macOS capacity optimization failed, expected: 256, got: %d", config.Capacity)
		}
	default:
		if config.Capacity != 1024 {
			t.Errorf("Standard platform capacity failed, expected: 1024, got: %d", config.Capacity)
		}
	}

	// Verify output functionality
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Should create logger with custom output: %v", err)
	}

	logger.Start()
	defer func() { _ = logger.Close() }()

	testMessage := "CustomOutput_ValidationMessage"
	logger.Info(testMessage)

	// Allow time for async processing
	CIFriendlySleep(50 * time.Millisecond)

	outputContent := outputBuffer.String()
	if !strings.Contains(outputContent, testMessage) {
		t.Errorf("Custom output should capture log messages. Expected to contain: %s, Got: %s",
			testMessage, outputContent)
	}
}

// TestUnit_TestConfigSmall_MinimalResourceUsage tests small resource configuration
func TestUnit_TestConfigSmall_MinimalResourceUsage(t *testing.T) {
	t.Parallel()

	// Execute function under test
	config := TestConfigSmall()

	// Validate minimal configuration
	if config.Level != Debug {
		t.Errorf("Small config should use Debug level, got: %v", config.Level)
	}
	if config.Output == nil {
		t.Error("Output should be configured")
	}
	if config.Encoder == nil {
		t.Error("Encoder should be present")
	}

	// Validate minimal resource allocation
	if config.Capacity != 64 {
		t.Errorf("Small config should use minimal capacity (64) regardless of platform, got: %d", config.Capacity)
	}

	// Test functionality with minimal resources
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Small config should create functional logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger should be instantiated")
	}

	logger.Start()
	defer func() { _ = logger.Close() }()

	// Verify logger handles multiple messages within small capacity
	for i := 0; i < 10; i++ {
		logger.Info("SmallConfig test message", Int("iteration", i))
	}

	// Validate stats reflect minimal configuration
	stats := logger.Stats()
	if stats == nil {
		t.Error("Logger stats should be available")
	}
}

// TestIntegration_TestHelpers_CrossPlatformBehavior tests helper functions across platforms
func TestIntegration_TestHelpers_CrossPlatformBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Test CI environment detection robustness
	originalEnvVars := []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "TRAVIS"}
	var originalValues []string

	// Store original values
	for _, envVar := range originalEnvVars {
		originalValues = append(originalValues, os.Getenv(envVar))
	}

	// Restore environment after test
	t.Cleanup(func() {
		for i, envVar := range originalEnvVars {
			if originalValues[i] != "" {
				_ = os.Setenv(envVar, originalValues[i])
			} else {
				_ = os.Unsetenv(envVar)
			}
		}
	})

	// Test various CI environment indicators
	ciIndicators := map[string]string{
		"CI":                     "true",
		"CONTINUOUS_INTEGRATION": "true",
		"GITHUB_ACTIONS":         "true",
		"TRAVIS":                 "true",
	}

	for envVar, value := range ciIndicators {
		t.Run("CI_Detection_"+envVar, func(t *testing.T) {
			// Clear all CI environment variables
			for _, clearVar := range originalEnvVars {
				_ = os.Unsetenv(clearVar)
			}

			// Set specific CI indicator
			_ = os.Setenv(envVar, value)

			// Verify CI detection
			if !IsCIEnvironment() {
				t.Errorf("Should detect CI environment with %s=%s", envVar, value)
			}

			// Verify helper functions adapt appropriately
			normalTimeout := 100 * time.Millisecond
			adaptedTimeout := CIFriendlyTimeout(normalTimeout)
			expectedTimeout := normalTimeout * 3
			if adaptedTimeout != expectedTimeout {
				t.Errorf("Timeout should be tripled in CI environment. Expected: %v, Got: %v",
					expectedTimeout, adaptedTimeout)
			}

			normalRetries := 5
			adaptedRetries := CIFriendlyRetryCount(normalRetries)
			expectedRetries := normalRetries * 2
			if adaptedRetries != expectedRetries {
				t.Errorf("Retries should be doubled in CI environment. Expected: %d, Got: %d",
					expectedRetries, adaptedRetries)
			}
		})
	}
}

// BenchmarkSuite_TestHelpers_PerformanceProfile benchmarks helper function performance
func BenchmarkSuite_TestHelpers_PerformanceProfile(b *testing.B) {
	b.Run("CIFriendlyTimeout", func(b *testing.B) {
		timeout := 100 * time.Millisecond
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			CIFriendlyTimeout(timeout)
		}
	})

	b.Run("CIFriendlyRetryCount", func(b *testing.B) {
		retries := 5
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			CIFriendlyRetryCount(retries)
		}
	})

	b.Run("TestConfig", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			TestConfig()
		}
	})

	b.Run("TestConfigSmall", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			TestConfigSmall()
		}
	})
}
