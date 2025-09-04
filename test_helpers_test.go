package iris

import (
	"os"
	"testing"
	"time"
)

func TestCIFriendlyTimeout(t *testing.T) {
	// Save original env at test start
	originalCI := os.Getenv("CI")
	originalGH := os.Getenv("GITHUB_ACTIONS")
	originalContInt := os.Getenv("CONTINUOUS_INTEGRATION")

	// Ensure clean state before starting
	defer func() {
		if originalCI != "" {
			os.Setenv("CI", originalCI)
		} else {
			os.Unsetenv("CI")
		}
		if originalGH != "" {
			os.Setenv("GITHUB_ACTIONS", originalGH)
		} else {
			os.Unsetenv("GITHUB_ACTIONS")
		}
		if originalContInt != "" {
			os.Setenv("CONTINUOUS_INTEGRATION", originalContInt)
		} else {
			os.Unsetenv("CONTINUOUS_INTEGRATION")
		}
	}()

	tests := []struct {
		name       string
		setCIEnv   bool
		timeout    time.Duration
		expectMult int
	}{
		{
			name:       "Normal_Environment",
			setCIEnv:   false,
			timeout:    100 * time.Millisecond,
			expectMult: 1,
		},
		{
			name:       "CI_Environment",
			setCIEnv:   true,
			timeout:    100 * time.Millisecond,
			expectMult: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment first
			os.Unsetenv("CI")
			os.Unsetenv("GITHUB_ACTIONS")
			os.Unsetenv("CONTINUOUS_INTEGRATION")

			if tt.setCIEnv {
				os.Setenv("CI", "true")
			}

			result := CIFriendlyTimeout(tt.timeout)
			expected := tt.timeout * time.Duration(tt.expectMult)

			if result != expected {
				t.Errorf("CIFriendlyTimeout() = %v, want %v", result, expected)
			}
		})
	}
}

func TestCIFriendlyRetryCount(t *testing.T) {
	// Save original env at test start
	originalCI := os.Getenv("CI")
	originalGH := os.Getenv("GITHUB_ACTIONS")
	originalContInt := os.Getenv("CONTINUOUS_INTEGRATION")

	// Ensure clean state before starting
	defer func() {
		if originalCI != "" {
			os.Setenv("CI", originalCI)
		} else {
			os.Unsetenv("CI")
		}
		if originalGH != "" {
			os.Setenv("GITHUB_ACTIONS", originalGH)
		} else {
			os.Unsetenv("GITHUB_ACTIONS")
		}
		if originalContInt != "" {
			os.Setenv("CONTINUOUS_INTEGRATION", originalContInt)
		} else {
			os.Unsetenv("CONTINUOUS_INTEGRATION")
		}
	}()

	tests := []struct {
		name       string
		setCIEnv   bool
		retries    int
		expectMult int
	}{
		{
			name:       "Normal_Environment",
			setCIEnv:   false,
			retries:    3,
			expectMult: 1,
		},
		{
			name:       "CI_Environment",
			setCIEnv:   true,
			retries:    3,
			expectMult: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment first
			os.Unsetenv("CI")
			os.Unsetenv("GITHUB_ACTIONS")
			os.Unsetenv("CONTINUOUS_INTEGRATION")

			if tt.setCIEnv {
				os.Setenv("CI", "true")
			}

			result := CIFriendlyRetryCount(tt.retries)
			expected := tt.retries * tt.expectMult

			if result != expected {
				t.Errorf("CIFriendlyRetryCount() = %v, want %v", result, expected)
			}
		})
	}
}

func TestCIFriendlySleep(t *testing.T) {
	// Save original env at test start
	originalCI := os.Getenv("CI")
	originalGH := os.Getenv("GITHUB_ACTIONS")
	originalContInt := os.Getenv("CONTINUOUS_INTEGRATION")

	// Ensure clean state before starting
	defer func() {
		if originalCI != "" {
			os.Setenv("CI", originalCI)
		} else {
			os.Unsetenv("CI")
		}
		if originalGH != "" {
			os.Setenv("GITHUB_ACTIONS", originalGH)
		} else {
			os.Unsetenv("GITHUB_ACTIONS")
		}
		if originalContInt != "" {
			os.Setenv("CONTINUOUS_INTEGRATION", originalContInt)
		} else {
			os.Unsetenv("CONTINUOUS_INTEGRATION")
		}
	}()

	tests := []struct {
		name       string
		setCIEnv   bool
		sleep      time.Duration
		expectMult int
	}{
		{
			name:       "Normal_Environment",
			setCIEnv:   false,
			sleep:      10 * time.Millisecond,
			expectMult: 1,
		},
		{
			name:       "CI_Environment",
			setCIEnv:   true,
			sleep:      10 * time.Millisecond,
			expectMult: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment first
			os.Unsetenv("CI")
			os.Unsetenv("GITHUB_ACTIONS")
			os.Unsetenv("CONTINUOUS_INTEGRATION")

			if tt.setCIEnv {
				os.Setenv("CI", "true")
			}

			start := time.Now()
			CIFriendlySleep(tt.sleep)
			elapsed := time.Since(start)

			expected := tt.sleep * time.Duration(tt.expectMult)
			// Allow 20% tolerance for timing variations
			tolerance := expected / 5

			if elapsed < expected-tolerance || elapsed > expected+tolerance {
				t.Errorf("CIFriendlySleep() took %v, expected around %v", elapsed, expected)
			}
		})
	}
}

func TestIsCIEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "No_CI_Env",
			envVars:  map[string]string{},
			expected: false,
		},
		{
			name:     "CI_Set",
			envVars:  map[string]string{"CI": "true"},
			expected: true,
		},
		{
			name:     "GITHUB_ACTIONS_Set",
			envVars:  map[string]string{"GITHUB_ACTIONS": "true"},
			expected: true,
		},
		{
			name:     "CONTINUOUS_INTEGRATION_Set",
			envVars:  map[string]string{"CONTINUOUS_INTEGRATION": "true"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			originalCI := os.Getenv("CI")
			originalGH := os.Getenv("GITHUB_ACTIONS")
			originalContInt := os.Getenv("CONTINUOUS_INTEGRATION")

			defer func() {
				os.Setenv("CI", originalCI)
				os.Setenv("GITHUB_ACTIONS", originalGH)
				os.Setenv("CONTINUOUS_INTEGRATION", originalContInt)
			}()

			// Clear all CI env vars
			os.Unsetenv("CI")
			os.Unsetenv("GITHUB_ACTIONS")
			os.Unsetenv("CONTINUOUS_INTEGRATION")

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			result := IsCIEnvironment()
			if result != tt.expected {
				t.Errorf("IsCIEnvironment() = %v, want %v", result, tt.expected)
			}
		})
	}
}
