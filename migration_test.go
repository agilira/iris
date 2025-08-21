// migration_test.go: Step 2.1 test
package iris

import (
	"testing"
)

// TestStep2_1_NewAPI tests the new BinaryField API
func TestStep2_1_NewAPI(t *testing.T) {
	config := Config{Level: InfoLevel}
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test new API - should work with BinaryField
	logger.Info("Test message",
		Str("user", "john"),
		Int("age", 30),
		Bool("active", true),
	)

	t.Log("✅ New API works!")
}
