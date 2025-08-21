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
// Optimized to reduce allocations and improve performance
func (m *MultiOutputConfig) ToConfig(baseConfig Config) Config {
	config := baseConfig

	// Pre-allocate slices to avoid multiple allocations
	writers, syncers := m.prepareOutputSlices()
	
	// Add different types of outputs
	syncers = m.addConsoleOutputs(syncers)
	syncers = m.addFileOutputs(syncers)
	writers, syncers = m.addCustomOutputs(writers, syncers)
	
	// Apply to config
	m.applyOutputsToConfig(&config, writers, syncers)

	return config
}

// prepareOutputSlices pre-allocates slices with exact capacity
func (m *MultiOutputConfig) prepareOutputSlices() ([]io.Writer, []WriteSyncer) {
	// Pre-calculate total capacity to reduce allocations
	totalWriters := len(m.Writers)
	totalSyncers := len(m.WriteSyncers) + len(m.Files)
	if m.Console {
		totalSyncers++
	}
	if m.ConsoleErr {
		totalSyncers++
	}

	var writers []io.Writer
	var syncers []WriteSyncer

	if totalWriters > 0 {
		writers = make([]io.Writer, 0, totalWriters)
	}
	if totalSyncers > 0 {
		syncers = make([]WriteSyncer, 0, totalSyncers)
	}
	
	return writers, syncers
}

// addConsoleOutputs adds console and stderr outputs
func (m *MultiOutputConfig) addConsoleOutputs(syncers []WriteSyncer) []WriteSyncer {
	// Add console outputs first (most common case)
	if m.Console {
		syncers = append(syncers, StdoutWriteSyncer)
	}
	if m.ConsoleErr {
		syncers = append(syncers, StderrWriteSyncer)
	}
	return syncers
}

// addFileOutputs adds file-based outputs
func (m *MultiOutputConfig) addFileOutputs(syncers []WriteSyncer) []WriteSyncer {
	// Add file outputs - optimized error handling
	if len(m.Files) > 0 {
		for _, filename := range m.Files {
			if syncer, err := NewFileWriteSyncer(filename); err == nil {
				syncers = append(syncers, syncer)
			}
			// Note: silently skip invalid files to maintain backward compatibility
		}
	}
	return syncers
}

// addCustomOutputs adds custom writers and syncers
func (m *MultiOutputConfig) addCustomOutputs(writers []io.Writer, syncers []WriteSyncer) ([]io.Writer, []WriteSyncer) {
	// Add custom writers and syncers (already allocated exactly)
	if len(m.Writers) > 0 {
		writers = append(writers, m.Writers...)
	}
	if len(m.WriteSyncers) > 0 {
		syncers = append(syncers, m.WriteSyncers...)
	}
	return writers, syncers
}

// applyOutputsToConfig sets writers and syncers in config
func (m *MultiOutputConfig) applyOutputsToConfig(config *Config, writers []io.Writer, syncers []WriteSyncer) {
	// Set in config only if we have writers
	if len(writers) > 0 {
		config.Writers = writers
	}
	if len(syncers) > 0 {
		config.WriteSyncers = syncers
	}
}

// NewMultiOutputLogger creates a logger with multiple outputs using the convenient config
func NewMultiOutputLogger(multiConfig MultiOutputConfig, baseConfig Config) (*Logger, error) {
	config := multiConfig.ToConfig(baseConfig)
	return New(config)
}

// Convenience functions for common multiple output scenarios

// NewTeeLogger creates a logger that writes to both console and a file
// Optimized to reduce configuration overhead
func NewTeeLogger(filename string, level Level, format Format) (*Logger, error) {
	// Direct config construction instead of going through MultiOutputConfig
	config := Config{
		Level:        level,
		Format:       format,
		BufferSize:   4096,
		BatchSize:    64,
		WriteSyncers: make([]WriteSyncer, 0, 2), // Pre-allocate for console + file
	}

	// Add console output
	config.WriteSyncers = append(config.WriteSyncers, StdoutWriteSyncer)

	// Add file output
	if syncer, err := NewFileWriteSyncer(filename); err == nil {
		config.WriteSyncers = append(config.WriteSyncers, syncer)
	} else {
		return nil, err
	}

	return New(config)
}

// NewDevelopmentTeeLogger creates a development logger with console (colorized) and file output
// Optimized for development use case
func NewDevelopmentTeeLogger(filename string) (*Logger, error) {
	config := Config{
		Level:        DebugLevel,
		Format:       ConsoleFormat,
		BufferSize:   1024,
		BatchSize:    32,
		EnableCaller: true,
		WriteSyncers: make([]WriteSyncer, 0, 2),
	}

	config.WriteSyncers = append(config.WriteSyncers, StdoutWriteSyncer)

	if syncer, err := NewFileWriteSyncer(filename); err == nil {
		config.WriteSyncers = append(config.WriteSyncers, syncer)
	} else {
		return nil, err
	}

	return New(config)
}

// NewProductionTeeLogger creates a production logger with structured output to both stdout and file
// Optimized for production use case
func NewProductionTeeLogger(filename string) (*Logger, error) {
	config := Config{
		Level:        InfoLevel,
		Format:       JSONFormat,
		BufferSize:   8192,
		BatchSize:    128,
		WriteSyncers: make([]WriteSyncer, 0, 2),
	}

	config.WriteSyncers = append(config.WriteSyncers, StdoutWriteSyncer)

	if syncer, err := NewFileWriteSyncer(filename); err == nil {
		config.WriteSyncers = append(config.WriteSyncers, syncer)
	} else {
		return nil, err
	}

	return New(config)
}

// NewRotatingFileLogger creates a logger that writes to console and multiple rotating files
// Optimized for multiple file outputs
func NewRotatingFileLogger(files []string, level Level) (*Logger, error) {
	config := Config{
		Level:        level,
		Format:       JSONFormat,
		BufferSize:   4096,
		BatchSize:    64,
		WriteSyncers: make([]WriteSyncer, 0, len(files)+1), // Pre-allocate for console + files
	}

	// Add console first
	config.WriteSyncers = append(config.WriteSyncers, StdoutWriteSyncer)

	// Add files
	for _, filename := range files {
		if syncer, err := NewFileWriteSyncer(filename); err == nil {
			config.WriteSyncers = append(config.WriteSyncers, syncer)
		}
		// Continue with other files even if one fails
	}

	return New(config)
}
