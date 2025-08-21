package iris

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Mock WriteSyncer for testing
type multiOutputMock struct {
	buffer     bytes.Buffer
	writeErr   error
	syncErr    error
	writeCalls int
	syncCalls  int
}

func newMultiOutputMock() *multiOutputMock {
	return &multiOutputMock{}
}

func (m *multiOutputMock) Write(p []byte) (n int, err error) {
	m.writeCalls++
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.buffer.Write(p)
}

func (m *multiOutputMock) Sync() error {
	m.syncCalls++
	return m.syncErr
}

func (m *multiOutputMock) String() string {
	return m.buffer.String()
}

// TestMultiOutputConfig tests the basic configuration structure
func TestMultiOutputConfig(t *testing.T) {
	config := MultiOutputConfig{
		Console:    true,
		ConsoleErr: true,
		Files:      []string{"test.log", "debug.log"},
	}

	if !config.Console {
		t.Error("Console should be enabled")
	}
	if !config.ConsoleErr {
		t.Error("ConsoleErr should be enabled")
	}
	if len(config.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(config.Files))
	}
}

// TestToConfigConsoleOnly tests conversion with console output only
func TestToConfigConsoleOnly(t *testing.T) {
	multiConfig := MultiOutputConfig{
		Console: true,
	}

	baseConfig := Config{
		Level:  InfoLevel,
		Format: JSONFormat,
	}

	result := multiConfig.ToConfig(baseConfig)

	if result.Level != InfoLevel {
		t.Errorf("Expected level %v, got %v", InfoLevel, result.Level)
	}
	if result.Format != JSONFormat {
		t.Errorf("Expected format %v, got %v", JSONFormat, result.Format)
	}
	if len(result.WriteSyncers) != 1 {
		t.Errorf("Expected 1 WriteSyncer, got %d", len(result.WriteSyncers))
	}
}

// TestToConfigStderrOnly tests conversion with stderr output only
func TestToConfigStderrOnly(t *testing.T) {
	multiConfig := MultiOutputConfig{
		ConsoleErr: true,
	}

	baseConfig := Config{
		Level:  DebugLevel,
		Format: ConsoleFormat,
	}

	result := multiConfig.ToConfig(baseConfig)

	if len(result.WriteSyncers) != 1 {
		t.Errorf("Expected 1 WriteSyncer, got %d", len(result.WriteSyncers))
	}
}

// TestToConfigWithFiles tests conversion with file outputs
func TestToConfigWithFiles(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "test1.log")
	file2 := filepath.Join(tempDir, "test2.log")

	multiConfig := MultiOutputConfig{
		Files: []string{file1, file2},
	}

	baseConfig := Config{
		Level:  WarnLevel,
		Format: BinaryFormat,
	}

	result := multiConfig.ToConfig(baseConfig)

	if len(result.WriteSyncers) != 2 {
		t.Errorf("Expected 2 WriteSyncers, got %d", len(result.WriteSyncers))
	}
}

// TestToConfigWithCustomWriters tests conversion with custom writers
func TestToConfigWithCustomWriters(t *testing.T) {
	writer1 := &bytes.Buffer{}
	writer2 := &bytes.Buffer{}
	syncer1 := newMultiOutputMock()

	multiConfig := MultiOutputConfig{
		Writers:      []io.Writer{writer1, writer2},
		WriteSyncers: []WriteSyncer{syncer1},
	}

	baseConfig := Config{
		Level:  ErrorLevel,
		Format: JSONFormat,
	}

	result := multiConfig.ToConfig(baseConfig)

	if len(result.Writers) != 2 {
		t.Errorf("Expected 2 Writers, got %d", len(result.Writers))
	}
	if len(result.WriteSyncers) != 1 {
		t.Errorf("Expected 1 WriteSyncer, got %d", len(result.WriteSyncers))
	}
}

// TestToConfigMixed tests conversion with mixed outputs
func TestToConfigMixed(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "mixed.log")

	writer := &bytes.Buffer{}
	syncer := newMultiOutputMock()

	multiConfig := MultiOutputConfig{
		Console:      true,
		ConsoleErr:   true,
		Files:        []string{logFile},
		Writers:      []io.Writer{writer},
		WriteSyncers: []WriteSyncer{syncer},
	}

	baseConfig := Config{
		Level:  DebugLevel,
		Format: ConsoleFormat,
	}

	result := multiConfig.ToConfig(baseConfig)

	// Should have: stdout, stderr, file = 3 WriteSyncers + 1 custom = 4 total
	expectedSyncers := 4
	if len(result.WriteSyncers) != expectedSyncers {
		t.Errorf("Expected %d WriteSyncers, got %d", expectedSyncers, len(result.WriteSyncers))
	}

	// Should have 1 custom writer
	if len(result.Writers) != 1 {
		t.Errorf("Expected 1 Writer, got %d", len(result.Writers))
	}
}

// TestNewMultiOutputLogger tests logger creation
func TestNewMultiOutputLogger(t *testing.T) {
	multiConfig := MultiOutputConfig{
		Console: true,
	}

	baseConfig := Config{
		Level:      InfoLevel,
		Format:     JSONFormat,
		BufferSize: 1024,
		BatchSize:  32,
	}

	logger, err := NewMultiOutputLogger(multiConfig, baseConfig)
	if err != nil {
		t.Errorf("Unexpected error creating logger: %v", err)
	}
	if logger == nil {
		t.Error("Expected logger, got nil")
	}
}

// TestNewTeeLogger tests the convenience tee logger function
func TestNewTeeLogger(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "tee.log")

	logger, err := NewTeeLogger(logFile, InfoLevel, JSONFormat)
	if err != nil {
		t.Errorf("Unexpected error creating tee logger: %v", err)
	}
	if logger == nil {
		t.Error("Expected logger, got nil")
	}

	// Test that file was created (by writing something)
	logger.Info("Test message")
	time.Sleep(100 * time.Millisecond) // Allow async write

	// Check if file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file was not created: %v", err)
	}
}

// TestNewDevelopmentTeeLogger tests development logger
func TestNewDevelopmentTeeLogger(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "dev.log")

	logger, err := NewDevelopmentTeeLogger(logFile)
	if err != nil {
		t.Errorf("Unexpected error creating development logger: %v", err)
	}
	if logger == nil {
		t.Error("Expected logger, got nil")
	}

	// Test debug level works
	logger.Debug("Debug message")
	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Development log file was not created: %v", err)
	}
}

// TestNewProductionTeeLogger tests production logger
func TestNewProductionTeeLogger(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "prod.log")

	logger, err := NewProductionTeeLogger(logFile)
	if err != nil {
		t.Errorf("Unexpected error creating production logger: %v", err)
	}
	if logger == nil {
		t.Error("Expected logger, got nil")
	}

	// Test info level works
	logger.Info("Production message")
	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Production log file was not created: %v", err)
	}
}

// TestNewRotatingFileLogger tests rotating file logger
func TestNewRotatingFileLogger(t *testing.T) {
	tempDir := t.TempDir()
	logFile1 := filepath.Join(tempDir, "rotate1.log")
	logFile2 := filepath.Join(tempDir, "rotate2.log")

	logger, err := NewRotatingFileLogger([]string{logFile1, logFile2}, WarnLevel)
	if err != nil {
		t.Errorf("Unexpected error creating rotating logger: %v", err)
	}
	if logger == nil {
		t.Error("Expected logger, got nil")
	}

	// Test warn level works
	logger.Warn("Warning message")
	time.Sleep(100 * time.Millisecond)

	// At least one file should exist
	file1Exists := true
	file2Exists := true
	if _, err := os.Stat(logFile1); os.IsNotExist(err) {
		file1Exists = false
	}
	if _, err := os.Stat(logFile2); os.IsNotExist(err) {
		file2Exists = false
	}

	if !file1Exists && !file2Exists {
		t.Error("Neither rotating log file was created")
	}
}

// TestToConfigWithInvalidFiles tests handling of invalid file paths
func TestToConfigWithInvalidFiles(t *testing.T) {
	multiConfig := MultiOutputConfig{
		Files: []string{"/invalid/path/that/does/not/exist.log"},
	}

	baseConfig := Config{
		Level:  InfoLevel,
		Format: JSONFormat,
	}

	result := multiConfig.ToConfig(baseConfig)

	// Should handle invalid files gracefully (skip them)
	// The invalid file should be skipped, so we should have 0 WriteSyncers
	if len(result.WriteSyncers) != 0 {
		t.Errorf("Expected 0 WriteSyncers (invalid file should be skipped), got %d", len(result.WriteSyncers))
	}
}

// TestEmptyMultiOutputConfig tests empty configuration
func TestEmptyMultiOutputConfig(t *testing.T) {
	multiConfig := MultiOutputConfig{}

	baseConfig := Config{
		Level:  InfoLevel,
		Format: JSONFormat,
	}

	result := multiConfig.ToConfig(baseConfig)

	if len(result.Writers) != 0 {
		t.Errorf("Expected 0 Writers, got %d", len(result.Writers))
	}
	if len(result.WriteSyncers) != 0 {
		t.Errorf("Expected 0 WriteSyncers, got %d", len(result.WriteSyncers))
	}
}

// TestConfigurationPreservation tests that base config values are preserved
func TestConfigurationPreservation(t *testing.T) {
	multiConfig := MultiOutputConfig{
		Console: true,
	}

	baseConfig := Config{
		Level:        DebugLevel,
		Format:       BinaryFormat,
		BufferSize:   2048,
		BatchSize:    128,
		EnableCaller: true,
	}

	result := multiConfig.ToConfig(baseConfig)

	if result.Level != baseConfig.Level {
		t.Errorf("Level not preserved: expected %v, got %v", baseConfig.Level, result.Level)
	}
	if result.Format != baseConfig.Format {
		t.Errorf("Format not preserved: expected %v, got %v", baseConfig.Format, result.Format)
	}
	if result.BufferSize != baseConfig.BufferSize {
		t.Errorf("BufferSize not preserved: expected %d, got %d", baseConfig.BufferSize, result.BufferSize)
	}
	if result.BatchSize != baseConfig.BatchSize {
		t.Errorf("BatchSize not preserved: expected %d, got %d", baseConfig.BatchSize, result.BatchSize)
	}
	if result.EnableCaller != baseConfig.EnableCaller {
		t.Errorf("EnableCaller not preserved: expected %v, got %v", baseConfig.EnableCaller, result.EnableCaller)
	}
}
