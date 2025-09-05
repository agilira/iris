// test_helpers.go: Common test utilities for CI-friendly testing
//
// This file provides utilities to make tests more robust in CI environments
// where resources are limited and timing is unpredictable.

package iris

import (
	"bytes"
	"os"
	"runtime"
	"time"
)

// IsCIEnvironment returns true if running in a CI environment
func IsCIEnvironment() bool {
	return os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("CONTINUOUS_INTEGRATION") != ""
}

// CIFriendlyTimeout returns an appropriate timeout for the given operation
// In CI environments, timeouts are increased to account for resource constraints
func CIFriendlyTimeout(normalTimeout time.Duration) time.Duration {
	if IsCIEnvironment() {
		// Increase timeouts by 3x in CI
		return normalTimeout * 3
	}
	return normalTimeout
}

// CIFriendlyRetryCount returns an appropriate retry count for the given operation
// In CI environments, retry counts are increased to account for scheduler variability
func CIFriendlyRetryCount(normalRetries int) int {
	if IsCIEnvironment() {
		// Increase retries by 2x in CI
		return normalRetries * 2
	}
	return normalRetries
}

// CIFriendlySleep sleeps for an appropriate duration
// In CI environments, sleep durations are increased to allow for slower scheduling
func CIFriendlySleep(normalDuration time.Duration) {
	if IsCIEnvironment() {
		time.Sleep(normalDuration * 2)
	} else {
		time.Sleep(normalDuration)
	}
}

// TestConfig returns a basic configuration optimized for testing across platforms
//
// This function provides a consistent base configuration that works reliably
// on all platforms including macOS, which has different memory characteristics.
//
// Returns:
//   - Config: Platform-optimized configuration for testing
func TestConfig() Config {
	var buf bytes.Buffer

	config := Config{
		Level:   Debug,
		Output:  WrapWriter(&buf),
		Encoder: NewJSONEncoder(),
	}

	// Use smaller capacity on macOS for compatibility
	if runtime.GOOS == "darwin" {
		config.Capacity = 256 // Small but sufficient for tests
	} else {
		config.Capacity = 1024 // Standard test size
	}

	return config
}

// TestConfigWithOutput returns a test configuration with specified output
//
// Parameters:
//   - output: Output destination for log messages
//
// Returns:
//   - Config: Platform-optimized configuration with custom output
func TestConfigWithOutput(output WriteSyncer) Config {
	config := TestConfig()
	config.Output = output
	return config
}

// TestConfigSmall returns a minimal configuration for unit tests
//
// This configuration uses the smallest viable ring buffer size across
// all platforms for tests that need minimal resource usage.
//
// Returns:
//   - Config: Minimal configuration for resource-constrained tests
func TestConfigSmall() Config {
	var buf bytes.Buffer

	return Config{
		Level:     Debug,
		Output:    WrapWriter(&buf),
		Encoder:   NewTextEncoder(),
		Capacity:  64, // Minimal viable size across all platforms
		BatchSize: 4,  // Small batches for immediate processing
	}
}
