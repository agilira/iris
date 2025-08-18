package iris

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSeparateBufferSampling(t *testing.T) {
	// Test with completely separate loggers and buffers

	// First: basic logging without sampling
	var buf1 bytes.Buffer
	config1 := Config{
		Level:  DebugLevel,
		Format: JSONFormat,
		Writer: &buf1,
	}

	logger1, err := New(config1)
	if err != nil {
		t.Fatalf("Failed to create logger1: %v", err)
	}
	defer logger1.Close()

	logger1.Info("basic message")
	time.Sleep(50 * time.Millisecond)
	logger1.Sync()

	output1 := buf1.String()
	fmt.Printf("Basic output: '%s'\n", output1)

	// Second: sampling with separate logger
	var buf2 bytes.Buffer
	samplingConfig := SamplingConfig{
		Initial:    1,
		Thereafter: 2,
	}

	config2 := Config{
		Level:          DebugLevel,
		Format:         JSONFormat,
		Writer:         &buf2,
		SamplingConfig: &samplingConfig,
	}

	logger2, err := New(config2)
	if err != nil {
		t.Fatalf("Failed to create logger2: %v", err)
	}
	defer logger2.Close()

	// Log 4 messages with individual sync
	for i := 0; i < 4; i++ {
		fmt.Printf("Logging message %d\n", i)
		logger2.Info(fmt.Sprintf("sampled message %d", i))

		// Sync after each message
		time.Sleep(20 * time.Millisecond)
		logger2.Sync()
	}

	output2 := buf2.String()
	fmt.Printf("Sampled output: '%s'\n", output2)

	time.Sleep(100 * time.Millisecond)
	logger2.ring.Flush()

	// Force another flush
	time.Sleep(50 * time.Millisecond)
	logger2.ring.Flush() // Wait a bit more and check for delayed messages
	time.Sleep(100 * time.Millisecond)

	// Check if there's any additional output
	finalOutput := buf2.String()
	if finalOutput != output2 {
		fmt.Printf("Additional output found: '%s'\n", finalOutput)
		output2 = finalOutput
	}

	lines := strings.Split(strings.TrimSpace(output2), "\n")
	var validLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			validLines = append(validLines, line)
		}
	}

	fmt.Printf("Valid lines count: %d\n", len(validLines))

	stats := logger2.GetSamplingStats()
	if stats != nil {
		fmt.Printf("Stats: Total=%d, Sampled=%d, Dropped=%d\n", stats.Total, stats.Sampled, stats.Dropped)
	}

	if len(validLines) == 0 {
		t.Fatal("Expected at least some log lines, got 0")
	}
}
