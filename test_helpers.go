// test_helpers.go: Common test utilities for CI-friendly testing
//
// This file provides utilities to make tests more robust in CI environments
// where resources are limited and timing is unpredictable.

package iris

import (
	"os"
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
