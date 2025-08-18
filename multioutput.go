package iris

import "io"

// MultiOutputConfig provides convenient configuration for multiple outputs
type MultiOutputConfig struct {
	Console      bool          // Write to console (stdout)
	ConsoleErr   bool          // Write to stderr
	Files        []string      // File paths to write to
	Writers      []io.Writer   // Custom writers
	WriteSyncers []WriteSyncer // Custom WriteSyncers
}

// ToConfig converts MultiOutputConfig to a standard Config with multiple outputs
func (m *MultiOutputConfig) ToConfig(baseConfig Config) Config {
	config := baseConfig

	var writers []io.Writer
	var syncers []WriteSyncer

	// Add console outputs
	if m.Console {
		syncers = append(syncers, StdoutWriteSyncer)
	}
	if m.ConsoleErr {
		syncers = append(syncers, StderrWriteSyncer)
	}

	// Add file outputs
	for _, filename := range m.Files {
		if syncer, err := NewFileWriteSyncer(filename); err == nil {
			syncers = append(syncers, syncer)
		}
	}

	// Add custom writers and syncers
	writers = append(writers, m.Writers...)
	syncers = append(syncers, m.WriteSyncers...)

	// Set in config
	config.Writers = writers
	config.WriteSyncers = syncers

	return config
}

// NewMultiOutputLogger creates a logger with multiple outputs using the convenient config
func NewMultiOutputLogger(multiConfig MultiOutputConfig, baseConfig Config) (*Logger, error) {
	config := multiConfig.ToConfig(baseConfig)
	return New(config)
}

// Convenience functions for common multiple output scenarios

// NewTeeLogger creates a logger that writes to both console and a file
func NewTeeLogger(filename string, level Level, format Format) (*Logger, error) {
	multiConfig := MultiOutputConfig{
		Console: true,
		Files:   []string{filename},
	}

	baseConfig := Config{
		Level:      level,
		Format:     format,
		BufferSize: 4096,
		BatchSize:  64,
	}

	return NewMultiOutputLogger(multiConfig, baseConfig)
}

// NewDevelopmentTeeLogger creates a development logger with console (colorized) and file output
func NewDevelopmentTeeLogger(filename string) (*Logger, error) {
	multiConfig := MultiOutputConfig{
		Console: true,
		Files:   []string{filename},
	}

	baseConfig := Config{
		Level:        DebugLevel,
		Format:       ConsoleFormat,
		BufferSize:   1024,
		BatchSize:    32,
		EnableCaller: true,
	}

	return NewMultiOutputLogger(multiConfig, baseConfig)
}

// NewProductionTeeLogger creates a production logger with structured output to both stdout and file
func NewProductionTeeLogger(filename string) (*Logger, error) {
	multiConfig := MultiOutputConfig{
		Console: true,
		Files:   []string{filename},
	}

	baseConfig := Config{
		Level:      InfoLevel,
		Format:     JSONFormat,
		BufferSize: 8192,
		BatchSize:  128,
	}

	return NewMultiOutputLogger(multiConfig, baseConfig)
}

// NewRotatingFileLogger creates a logger that writes to console and multiple rotating files
func NewRotatingFileLogger(files []string, level Level) (*Logger, error) {
	multiConfig := MultiOutputConfig{
		Console: true,
		Files:   files,
	}

	baseConfig := Config{
		Level:      level,
		Format:     JSONFormat,
		BufferSize: 4096,
		BatchSize:  64,
	}

	return NewMultiOutputLogger(multiConfig, baseConfig)
}
