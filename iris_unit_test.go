package iris

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// Mock writer for testing iris logger
type irisTestWriter struct {
	buffer     bytes.Buffer
	writeCount int
	syncCount  int
	writeErr   error
	syncErr    error
	mu         sync.Mutex
}

func newIrisTestWriter() *irisTestWriter {
	return &irisTestWriter{}
}

func (m *irisTestWriter) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeCount++
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.buffer.Write(p)
}

func (m *irisTestWriter) Sync() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.syncCount++
	return m.syncErr
}

func (m *irisTestWriter) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.buffer.String()
}

func (m *irisTestWriter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buffer.Reset()
	m.writeCount = 0
	m.syncCount = 0
}

func (m *irisTestWriter) WriteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.writeCount
}

func (m *irisTestWriter) SyncCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.syncCount
}

// TestLoggerCreation tests basic logger creation
func TestLoggerCreation(t *testing.T) {
	config := Config{
		Level:      InfoLevel,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger is nil")
	}

	// Test logger can be closed
	logger.Close()
}

// TestLoggerWithCustomWriter tests logger with custom writer
func TestLoggerWithCustomWriter(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:      DebugLevel,
		Writer:     writer,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log a message and wait for processing
	logger.Info("test message")
	time.Sleep(50 * time.Millisecond)

	if writer.WriteCount() == 0 {
		t.Error("Expected writer to be called")
	}
}

// TestLoggerLevels tests all logging levels
func TestLoggerLevels(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:      DebugLevel,
		Writer:     writer,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test all levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	time.Sleep(100 * time.Millisecond)

	if writer.WriteCount() != 4 {
		t.Errorf("Expected 4 writes, got %d", writer.WriteCount())
	}
}

// TestLoggerLevelFiltering tests level filtering
func TestLoggerLevelFiltering(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:      WarnLevel, // Only warn and above
		Writer:     writer,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test filtering
	logger.Debug("debug message") // Should be filtered
	logger.Info("info message")   // Should be filtered
	logger.Warn("warn message")   // Should pass
	logger.Error("error message") // Should pass

	time.Sleep(100 * time.Millisecond)

	if writer.WriteCount() != 2 {
		t.Errorf("Expected 2 writes, got %d", writer.WriteCount())
	}
}

// TestLoggerWithFields tests logging with fields
func TestLoggerWithFields(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:      InfoLevel,
		Writer:     writer,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log with fields
	logger.Info("test message", String("key", "value"), Int("count", 42))
	time.Sleep(100 * time.Millisecond)

	output := writer.String()
	if !strings.Contains(output, "test message") {
		t.Error("Expected message in output")
	}
	if !strings.Contains(output, "key") {
		t.Error("Expected field key in output")
	}
}

// TestLoggerCaller tests caller information
func TestLoggerCaller(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:        InfoLevel,
		Writer:       writer,
		Format:       JSONFormat,
		EnableCaller: true,
		BufferSize:   1024,
		BatchSize:    32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Info("test message with caller")
	time.Sleep(100 * time.Millisecond)

	output := writer.String()
	if !strings.Contains(output, "iris_unit_test.go") {
		t.Error("Expected caller file in output")
	}
}

// TestLoggerMultipleWriters tests multiple output destinations
func TestLoggerMultipleWriters(t *testing.T) {
	writer1 := newIrisTestWriter()
	writer2 := newIrisTestWriter()

	config := Config{
		Level:      InfoLevel,
		Writers:    []io.Writer{writer1, writer2}, // Use io.Writer slice
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Info("test message")
	time.Sleep(100 * time.Millisecond)

	if writer1.WriteCount() == 0 {
		t.Error("Expected writer1 to be called")
	}
	if writer2.WriteCount() == 0 {
		t.Error("Expected writer2 to be called")
	}
}

// TestLoggerFormats tests different output formats
func TestLoggerFormats(t *testing.T) {
	formats := []Format{JSONFormat, ConsoleFormat, FastTextFormat, BinaryFormat}

	for _, format := range formats {
		t.Run(fmt.Sprintf("Format_%d", format), func(t *testing.T) {
			writer := newIrisTestWriter()
			config := Config{
				Level:      InfoLevel,
				Writer:     writer,
				Format:     format,
				BufferSize: 1024,
				BatchSize:  32,
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("Failed to create logger with format %d: %v", format, err)
			}
			defer logger.Close()

			logger.Info("test message")
			time.Sleep(100 * time.Millisecond)

			if writer.WriteCount() == 0 {
				t.Errorf("Expected write for format %d", format)
			}
		})
	}
}

// TestLoggerUltraFast tests ultra-fast mode
func TestLoggerUltraFast(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:      InfoLevel,
		Writer:     writer,
		UltraFast:  true,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create ultra-fast logger: %v", err)
	}
	defer logger.Close()

	logger.Info("ultra fast message")
	time.Sleep(100 * time.Millisecond)

	if writer.WriteCount() == 0 {
		t.Error("Expected write in ultra-fast mode")
	}
}

// TestLoggerSampling tests log sampling
func TestLoggerSampling(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:  InfoLevel,
		Writer: writer,
		Format: JSONFormat,
		SamplingConfig: &SamplingConfig{
			Initial:    2,
			Thereafter: 5,
		},
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger with sampling: %v", err)
	}
	defer logger.Close()

	// Log multiple messages
	for i := 0; i < 10; i++ {
		logger.Info(fmt.Sprintf("message %d", i))
	}
	time.Sleep(200 * time.Millisecond)

	// Should have fewer writes due to sampling
	writeCount := writer.WriteCount()
	if writeCount >= 10 {
		t.Errorf("Expected sampling to reduce writes, got %d", writeCount)
	}
	if writeCount == 0 {
		t.Error("Expected at least some writes")
	}
}

// TestLoggerPanicRecovery tests panic recovery
func TestLoggerPanicRecovery(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:      InfoLevel,
		Writer:     writer,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test that panic methods actually panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Panic to panic")
		}
	}()

	logger.Panic("panic message")
}

// TestLoggerDefaults tests default configuration values
func TestLoggerDefaults(t *testing.T) {
	config := Config{} // Empty config

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger with defaults: %v", err)
	}
	defer logger.Close()

	// Should work with all defaults
	logger.Info("test with defaults")
	time.Sleep(50 * time.Millisecond)
}

// TestLoggerStackTrace tests stack trace capture
func TestLoggerStackTrace(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:           InfoLevel,
		Writer:          writer,
		Format:          JSONFormat,
		StackTraceLevel: ErrorLevel,
		BufferSize:      1024,
		BatchSize:       32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Error("error with stack trace")
	time.Sleep(100 * time.Millisecond)

	output := writer.String()
	if !strings.Contains(output, "error with stack trace") {
		t.Error("Expected error message in output")
	}
}

// TestLoggerSync tests sync functionality
func TestLoggerSync(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:      InfoLevel,
		Writer:     writer,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Info("test message")
	err = logger.Sync()
	if err != nil {
		t.Errorf("Sync failed: %v", err)
	}
}

// TestLoggerAddRemoveWriter tests dynamic writer management
func TestLoggerAddRemoveWriter(t *testing.T) {
	writer1 := newIrisTestWriter()
	config := Config{
		Level:      InfoLevel,
		Writer:     writer1,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Add second writer
	writer2 := newIrisTestWriter()
	err = logger.AddWriter(writer2)
	if err != nil {
		t.Errorf("Failed to add writer: %v", err)
	}

	if logger.WriterCount() != 2 {
		t.Errorf("Expected 2 writers, got %d", logger.WriterCount())
	}

	// Log message to both writers
	logger.Info("test message")
	time.Sleep(100 * time.Millisecond)

	if writer1.WriteCount() == 0 {
		t.Error("Expected writer1 to receive message")
	}
	if writer2.WriteCount() == 0 {
		t.Error("Expected writer2 to receive message")
	}

	// Remove writer
	removed := logger.RemoveWriter(writer2)
	if !removed {
		t.Error("Expected writer to be removed")
	}
}

// TestLoggerClosed tests logger behavior when closed
func TestLoggerClosed(t *testing.T) {
	writer := newIrisTestWriter()
	config := Config{
		Level:      InfoLevel,
		Writer:     writer,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Close logger
	logger.Close()

	// Try to log after closing
	initialCount := writer.WriteCount()
	logger.Info("should not be written")
	time.Sleep(50 * time.Millisecond)

	if writer.WriteCount() != initialCount {
		t.Error("Expected no writes after logger is closed")
	}
}

// TestLoggerAddWriteSyncer tests the AddWriteSyncer function
func TestLoggerAddWriteSyncer(t *testing.T) {
	writer1 := newIrisTestWriter()
	writer2 := newIrisTestWriter()

	config := Config{
		Writer: writer1,
		Level:  InfoLevel,
		Format: JSONFormat,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Initially should have 1 writer
	if logger.WriterCount() != 1 {
		t.Errorf("Expected 1 writer initially, got %d", logger.WriterCount())
	}

	// Add a second writer
	err = logger.AddWriteSyncer(WrapWriter(writer2))
	if err != nil {
		t.Fatalf("Failed to add writer syncer: %v", err)
	}

	// Should now have 2 writers
	if logger.WriterCount() != 2 {
		t.Errorf("Expected 2 writers after adding syncer, got %d", logger.WriterCount())
	}

	// Log a message - should go to both writers
	logger.Info("test message")
	time.Sleep(50 * time.Millisecond)

	if writer1.WriteCount() == 0 {
		t.Error("Expected writes to first writer")
	}
	if writer2.WriteCount() == 0 {
		t.Error("Expected writes to second writer")
	}

	// Both writers should have received the same message
	output1 := writer1.String()
	output2 := writer2.String()
	if output1 != output2 {
		t.Errorf("Writers received different output:\nWriter1: %s\nWriter2: %s", output1, output2)
	}
}

// TestLoggerGetSamplingStats tests the GetSamplingStats function
func TestLoggerGetSamplingStats(t *testing.T) {
	writer := newIrisTestWriter()

	// Test without sampling
	config := Config{
		Writer: writer,
		Level:  InfoLevel,
		Format: JSONFormat,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Should return nil when sampling is not enabled
	stats := logger.GetSamplingStats()
	if stats != nil {
		t.Error("Expected nil sampling stats when sampling is not enabled")
	}

	// Test with sampling enabled
	samplingConfig := &SamplingConfig{
		Initial:    1,
		Thereafter: 10,
	}

	configWithSampling := Config{
		Writer:         writer,
		Level:          InfoLevel,
		Format:         JSONFormat,
		SamplingConfig: samplingConfig,
	}

	loggerWithSampling, err := New(configWithSampling)
	if err != nil {
		t.Fatalf("Failed to create logger with sampling: %v", err)
	}
	defer loggerWithSampling.Close()

	// Should return valid stats when sampling is enabled
	stats = loggerWithSampling.GetSamplingStats()
	if stats == nil {
		t.Error("Expected non-nil sampling stats when sampling is enabled")
	} else {
		// Verify stats structure
		if stats.Sampled < 0 || stats.Dropped < 0 {
			t.Error("Sampling stats should have non-negative values")
		}
	}
}

// TestLoggerWith tests the With method for creating child loggers
func TestLoggerWith(t *testing.T) {
	writer := newIrisTestWriter()

	config := Config{
		Writer: writer,
		Level:  InfoLevel,
		Format: JSONFormat,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test With() with no fields - should return same logger
	noFieldLogger := logger.With()
	if noFieldLogger != logger {
		t.Error("With() with no fields should return the same logger instance")
	}

	// Test With() with fields
	childLogger := logger.With(Str("service", "test"), Int("version", 1))
	if childLogger == logger {
		t.Error("With() with fields should return a different logger instance")
	}

	// Test logging with child logger
	initialCount := writer.WriteCount()
	childLogger.Info("test message with pre-set fields")
	time.Sleep(50 * time.Millisecond)

	if writer.WriteCount() <= initialCount {
		t.Error("Expected write to occur after logging with child logger")
	}

	// Verify the pre-set fields are included in output
	output := writer.String()
	if !strings.Contains(output, "service") || !strings.Contains(output, "test") {
		t.Error("Expected pre-set fields to be included in log output")
	}
	if !strings.Contains(output, "version") || !strings.Contains(output, "1") {
		t.Error("Expected pre-set fields to be included in log output")
	}
}

// TestLoggerWithNested tests nested With() calls
func TestLoggerWithNested(t *testing.T) {
	writer := newIrisTestWriter()

	config := Config{
		Writer: writer,
		Level:  InfoLevel,
		Format: JSONFormat,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Create nested child loggers
	level1 := logger.With(Str("component", "auth"))
	level2 := level1.With(Str("user_id", "12345"))
	level3 := level2.With(Int("session_id", 98765))

	// Test logging from deepest level
	level3.Info("nested logger test")
	time.Sleep(50 * time.Millisecond)

	output := writer.String()

	// Verify all nested fields are present
	expectedFields := []string{"component", "auth", "user_id", "12345", "session_id", "98765"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field %q to be present in output: %s", field, output)
		}
	}
}
