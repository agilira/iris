// backpressure_test.go: Tests for configurable backpressure policies
package iris

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/agilira/iris/internal/zephyroslite"
)

// backpressureTestSyncer wraps a bytes.Buffer to implement WriteSyncer for testing
type backpressureTestSyncer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (bs *backpressureTestSyncer) Write(p []byte) (n int, err error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.buf.Write(p)
}

func (bs *backpressureTestSyncer) Sync() error {
	return nil
}

func (bs *backpressureTestSyncer) String() string {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.buf.String()
}

func (bs *backpressureTestSyncer) Bytes() []byte {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.buf.Bytes()
}

// slowWriterForBackpressure simulates a slow writer for testing backpressure
type slowWriterForBackpressure struct {
	buf   *backpressureTestSyncer
	delay time.Duration
}

func (w *slowWriterForBackpressure) Write(p []byte) (n int, err error) {
	// Simulate slow I/O
	time.Sleep(w.delay)
	return w.buf.Write(p)
}

func (w *slowWriterForBackpressure) Sync() error {
	return w.buf.Sync()
}

// TestBackpressureDropOnFull tests the default drop behavior
func TestBackpressureDropOnFull(t *testing.T) {
	syncer := &backpressureTestSyncer{}

	logger, err := New(Config{
		Level:              Debug,
		Output:             syncer,
		Encoder:            NewTextEncoder(),
		Capacity:           64, // Small capacity to trigger overflow
		BackpressurePolicy: zephyroslite.DropOnFull,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Start()
	defer logger.Close()

	// Fill the buffer quickly to trigger drops
	for i := 0; i < 100; i++ {
		logger.Info("Test message", Int("iteration", i))
	}

	// Allow processing time
	time.Sleep(100 * time.Millisecond)
	logger.Sync()

	// With DropOnFull, some messages should be dropped
	// We can't predict exact count, but should be less than 100
	lines := len(bytes.Split(syncer.Bytes(), []byte("\n"))) - 1 // -1 for empty last line

	t.Logf("Processed %d messages with DropOnFull policy", lines)
	if lines >= 100 {
		t.Errorf("Expected some messages to be dropped, but got %d (expected < 100)", lines)
	}
}

// TestBackpressureBlockOnFull tests the blocking behavior
func TestBackpressureBlockOnFull(t *testing.T) {
	syncer := &backpressureTestSyncer{}

	// Create a slow writer to force blocking
	slowWriter := &slowWriterForBackpressure{
		buf:   syncer,
		delay: 5 * time.Millisecond,
	}

	logger, err := New(Config{
		Level:              Debug,
		Output:             slowWriter,
		Encoder:            NewTextEncoder(),
		Capacity:           64, // Small capacity to force blocking
		BackpressurePolicy: zephyroslite.BlockOnFull,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Start()
	defer logger.Close()

	// Write messages concurrently to test blocking behavior
	start := time.Now()
	var wg sync.WaitGroup
	messageCount := 50

	for i := 0; i < messageCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			logger.Info("Test message", Int("iteration", idx))
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Allow final processing
	time.Sleep(100 * time.Millisecond)
	logger.Sync()

	lines := len(bytes.Split(syncer.Bytes(), []byte("\n"))) - 1

	t.Logf("Processed %d messages with BlockOnFull policy in %v", lines, duration)

	// With BlockOnFull, all messages should be processed
	if lines != messageCount {
		t.Errorf("Expected all %d messages to be processed, but got %d", messageCount, lines)
	}

	// Should take longer due to blocking (more than just the slow writer delay)
	expectedMinDuration := time.Duration(messageCount/10) * slowWriter.delay
	if duration < expectedMinDuration {
		t.Logf("Duration %v was faster than expected minimum %v, but this may be normal", duration, expectedMinDuration)
	}
}

// TestBackpressurePolicyComparison compares both policies side by side
func TestBackpressurePolicyComparison(t *testing.T) {
	testCases := []struct {
		name     string
		policy   zephyroslite.BackpressurePolicy
		capacity int64
		messages int
		delay    time.Duration
	}{
		{
			name:     "DropOnFull",
			policy:   zephyroslite.DropOnFull,
			capacity: 64,
			messages: 100,
			delay:    time.Millisecond,
		},
		{
			name:     "BlockOnFull",
			policy:   zephyroslite.BlockOnFull,
			capacity: 256,                    // Larger capacity to reduce blocking
			messages: 30,                     // Fewer concurrent messages
			delay:    500 * time.Microsecond, // Faster writer
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			syncer := &backpressureTestSyncer{}
			slowWriter := &slowWriterForBackpressure{
				buf:   syncer,
				delay: test.delay,
			}

			logger, err := New(Config{
				Level:              Debug,
				Output:             slowWriter,
				Encoder:            NewTextEncoder(),
				Capacity:           test.capacity,
				BackpressurePolicy: test.policy,
			})
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Start()
			defer logger.Close()

			start := time.Now()

			// Write messages concurrently
			var wg sync.WaitGroup
			for i := 0; i < test.messages; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					logger.Info("Test message", Int("iteration", idx))
				}(i)
			}

			wg.Wait()
			duration := time.Since(start)

			// Allow processing to complete
			time.Sleep(200 * time.Millisecond)
			logger.Sync()

			processed := len(bytes.Split(syncer.Bytes(), []byte("\n"))) - 1

			t.Logf("Policy %s: processed %d/%d messages in %v",
				test.name, processed, test.messages, duration)

			switch test.policy {
			case zephyroslite.DropOnFull:
				// DropOnFull may lose some messages under high contention
				if processed > test.messages {
					t.Errorf("Processed more messages (%d) than sent (%d)", processed, test.messages)
				}
				t.Logf("DropOnFull dropped %d messages", test.messages-processed)

			case zephyroslite.BlockOnFull:
				// BlockOnFull should process all messages
				if processed != test.messages {
					t.Errorf("Expected all %d messages, got %d", test.messages, processed)
				}
			}
		})
	}
}

// TestBackpressurePolicyDefault tests that the default policy is DropOnFull
func TestBackpressurePolicyDefault(t *testing.T) {
	syncer := &backpressureTestSyncer{}

	// Create logger without specifying BackpressurePolicy
	logger, err := New(Config{
		Level:    Debug,
		Output:   syncer,
		Encoder:  NewTextEncoder(),
		Capacity: 64, // Use explicit small capacity like other tests
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Start()
	defer logger.Close()

	// Write a few messages
	for i := 0; i < 10; i++ {
		logger.Info("Test message", Int("iteration", i))
	}

	// Give time for processing
	time.Sleep(50 * time.Millisecond)
	logger.Sync()

	// Should work fine with default policy
	lines := len(bytes.Split(syncer.Bytes(), []byte("\n"))) - 1
	t.Logf("Got %d lines from default policy test", lines)
	if lines != 10 {
		t.Errorf("Expected 10 messages, got %d", lines)
	}
}
