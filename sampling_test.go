package iris

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSampler(t *testing.T) {
	t.Run("Basic sampling logic", func(t *testing.T) {
		config := SamplingConfig{
			Initial:    2,
			Thereafter: 3,
		}

		sampler := NewSampler(config)

		// First 2 entries should always be sampled
		entry := LogEntry{Message: "test"}

		decision1 := sampler.Sample(entry)
		if decision1 != LogSample {
			t.Errorf("Expected LogSample for entry 1, got %v", decision1)
		}

		decision2 := sampler.Sample(entry)
		if decision2 != LogSample {
			t.Errorf("Expected LogSample for entry 2, got %v", decision2)
		}

		// Next entry should be dropped (3rd entry, not divisible by 3)
		decision3 := sampler.Sample(entry)
		if decision3 != DropSample {
			t.Errorf("Expected DropSample for entry 3, got %v", decision3)
		}

		// 4th entry should be dropped
		decision4 := sampler.Sample(entry)
		if decision4 != DropSample {
			t.Errorf("Expected DropSample for entry 4, got %v", decision4)
		}

		// 5th entry should be sampled (3+2 = 5, first "thereafter" sample)
		decision5 := sampler.Sample(entry)
		if decision5 != LogSample {
			t.Errorf("Expected LogSample for entry 5, got %v", decision5)
		}
	})

	t.Run("Sampling stats", func(t *testing.T) {
		config := SamplingConfig{
			Initial:    1,
			Thereafter: 2,
		}

		sampler := NewSampler(config)
		entry := LogEntry{Message: "test"}

		// Sample 10 entries
		for i := 0; i < 10; i++ {
			sampler.Sample(entry)
		}

		stats := sampler.Stats()

		// Should have initial (1) + (10-1)/2 = 1 + 4 = 5 sampled entries
		expectedSampled := int64(5)
		if stats.Sampled != expectedSampled {
			t.Errorf("Expected %d sampled entries, got %d", expectedSampled, stats.Sampled)
		}

		// Should have 10 - 5 = 5 dropped entries
		expectedDropped := int64(5)
		if stats.Dropped != expectedDropped {
			t.Errorf("Expected %d dropped entries, got %d", expectedDropped, stats.Dropped)
		}

		// Total should be 10
		if stats.Total != 10 {
			t.Errorf("Expected 10 total entries, got %d", stats.Total)
		}

		// Sampling rate should be 50%
		expectedRate := 50.0
		actualRate := stats.SamplingRate()
		if actualRate != expectedRate {
			t.Errorf("Expected %.1f%% sampling rate, got %.1f%%", expectedRate, actualRate)
		}
	})

	t.Run("Time-based reset", func(t *testing.T) {
		config := SamplingConfig{
			Initial:    1,
			Thereafter: 2,
			Tick:       time.Millisecond * 10,
		}

		sampler := NewSampler(config)
		entry := LogEntry{Message: "test"}

		// First entry should be sampled (initial)
		decision1 := sampler.Sample(entry)
		if decision1 != LogSample {
			t.Errorf("Expected LogSample for first entry, got %v", decision1)
		}

		// Second entry should be dropped
		decision2 := sampler.Sample(entry)
		if decision2 != DropSample {
			t.Errorf("Expected DropSample for second entry, got %v", decision2)
		}

		// Wait for tick duration
		time.Sleep(time.Millisecond * 15)

		// After reset, first entry should be sampled again
		decision3 := sampler.Sample(entry)
		if decision3 != LogSample {
			t.Errorf("Expected LogSample after reset, got %v", decision3)
		}
	})
}

func TestSamplingConfig(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		config := SamplingConfig{} // Empty config
		sampler := NewSampler(config)

		// Should get default values
		if sampler.config.Initial != 100 {
			t.Errorf("Expected default Initial=100, got %d", sampler.config.Initial)
		}

		if sampler.config.Thereafter != 100 {
			t.Errorf("Expected default Thereafter=100, got %d", sampler.config.Thereafter)
		}
	})

	t.Run("Preset configurations", func(t *testing.T) {
		devConfig := NewDevelopmentSampling()
		if devConfig.Initial != 100 || devConfig.Thereafter != 10 || devConfig.Tick != 0 {
			t.Errorf("Unexpected development config: %+v", devConfig)
		}

		prodConfig := NewProductionSampling()
		if prodConfig.Initial != 100 || prodConfig.Thereafter != 100 || prodConfig.Tick != time.Minute {
			t.Errorf("Unexpected production config: %+v", prodConfig)
		}

		hvConfig := NewHighVolumeSampling()
		if hvConfig.Initial != 50 || hvConfig.Thereafter != 1000 || hvConfig.Tick != 10*time.Minute {
			t.Errorf("Unexpected high volume config: %+v", hvConfig)
		}
	})
}

func TestLoggerSampling(t *testing.T) {
	t.Run("Sampling integration", func(t *testing.T) {
		var buf bytes.Buffer

		samplingConfig := SamplingConfig{
			Initial:    2,
			Thereafter: 2,
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

		// Log 6 messages
		for i := 0; i < 6; i++ {
			logger.Info("test message", Int("i", i))
		}

		// Ensure processing
		logger.ring.Flush()
		time.Sleep(50 * time.Millisecond)

		// Should have 4 log entries:
		// - Entry 1: sampled (initial)
		// - Entry 2: sampled (initial)
		// - Entry 3: dropped (not divisible by 2)
		// - Entry 4: sampled (first "thereafter")
		// - Entry 5: dropped (not divisible by 2)
		// - Entry 6: sampled (second "thereafter")
		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		// Filter out empty lines
		var validLines []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				validLines = append(validLines, line)
			}
		}

		expectedLines := 4
		if len(validLines) != expectedLines {
			t.Errorf("Expected %d log lines, got %d\nOutput:\n%s", expectedLines, len(validLines), output)
		}

		// Check sampling stats
		stats := logger.GetSamplingStats()
		if stats == nil {
			t.Fatal("Expected sampling stats, got nil")
		}

		if stats.Total != 6 {
			t.Errorf("Expected 6 total entries, got %d", stats.Total)
		}

		if stats.Sampled != 4 {
			t.Errorf("Expected 4 sampled entries, got %d", stats.Sampled)
		}

		if stats.Dropped != 2 {
			t.Errorf("Expected 2 dropped entries, got %d", stats.Dropped)
		}
	})

	t.Run("No sampling when disabled", func(t *testing.T) {
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
		defer logger.Close()

		// Log 5 messages
		for i := 0; i < 5; i++ {
			logger.Info("test message", Int("i", i))
		}

		logger.ring.Flush()
		time.Sleep(50 * time.Millisecond)

		// Should have all 5 log entries
		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		// Filter out empty lines
		var validLines []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				validLines = append(validLines, line)
			}
		}

		if len(validLines) != 5 {
			t.Errorf("Expected 5 log lines, got %d", len(validLines))
		}

		// No sampling stats should be available
		stats := logger.GetSamplingStats()
		if stats != nil {
			t.Errorf("Expected no sampling stats, got %+v", stats)
		}

		if logger.IsSamplingEnabled() {
			t.Error("Expected sampling to be disabled")
		}
	})

	t.Run("Dynamic sampling configuration", func(t *testing.T) {
		var buf bytes.Buffer

		config := Config{
			Level:  DebugLevel,
			Format: JSONFormat,
			Writer: &buf,
		}

		logger, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Initially no sampling
		if logger.IsSamplingEnabled() {
			t.Error("Expected sampling to be disabled initially")
		}

		// Enable sampling
		samplingConfig := SamplingConfig{
			Initial:    1,
			Thereafter: 2,
		}
		logger.SetSamplingConfig(&samplingConfig)

		if !logger.IsSamplingEnabled() {
			t.Error("Expected sampling to be enabled after SetSamplingConfig")
		}

		// Disable sampling again
		logger.SetSamplingConfig(nil)

		if logger.IsSamplingEnabled() {
			t.Error("Expected sampling to be disabled after SetSamplingConfig(nil)")
		}
	})
}
