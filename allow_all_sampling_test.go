package iris

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestAllowAllSampling(t *testing.T) {
	var buf bytes.Buffer

	// Sampling that allows everything: Initial=10, Thereafter=1
	samplingConfig := SamplingConfig{
		Initial:    10, // Allow first 10 messages
		Thereafter: 1,  // Then allow every message
	}

	config := Config{
		Level:          DebugLevel,
		Format:         JSONFormat,
		Writer:         &buf,
		SamplingConfig: &samplingConfig,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log 5 messages - all should be allowed
	for i := 0; i < 5; i++ {
		logger.Info(fmt.Sprintf("message %d", i))
	}

	time.Sleep(50 * time.Millisecond)
	logger.ring.Flush()
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	fmt.Printf("Output: '%s'\n", output)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var validLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			validLines = append(validLines, line)
		}
	}

	fmt.Printf("Valid lines count: %d\n", len(validLines))
	for i, line := range validLines {
		fmt.Printf("Line %d: %s\n", i, line)
	}

	stats := logger.GetSamplingStats()
	if stats != nil {
		fmt.Printf("Stats: Total=%d, Sampled=%d, Dropped=%d\n", stats.Total, stats.Sampled, stats.Dropped)
	}

	// All 5 messages should be in output
	if len(validLines) != 5 {
		t.Errorf("Expected 5 log lines, got %d", len(validLines))
	}

	// All messages should be sampled, none dropped
	if stats != nil && (stats.Total != 5 || stats.Sampled != 5 || stats.Dropped != 0) {
		t.Errorf("Expected all messages to be sampled (Total=5, Sampled=5, Dropped=0), got Total=%d, Sampled=%d, Dropped=%d",
			stats.Total, stats.Sampled, stats.Dropped)
	}
}
