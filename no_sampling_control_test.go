package iris

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNoSamplingControl(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  DebugLevel,
		Format: JSONFormat,
		Writer: &buf,
		// No sampling config
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Log 5 messages without sampling
	for i := 0; i < 5; i++ {
		logger.Info(fmt.Sprintf("message %d", i))
		time.Sleep(1 * time.Millisecond) // Small delay between messages
	}

	time.Sleep(50 * time.Millisecond)
	logger.ring.Flush()
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	fmt.Printf("Output: '%s'\n", output)

	// Close logger explicitly
	logger.Close()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var validLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			validLines = append(validLines, line)
		}
	}

	fmt.Printf("Valid lines count: %d\n", len(validLines))

	// All 5 messages should be in output
	if len(validLines) != 5 {
		t.Errorf("Expected 5 log lines, got %d", len(validLines))
	}
}
