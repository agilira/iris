package iris

import (
	"os"
	"sync"
)

// Common WriteSyncer implementations

// FileWriteSyncer wraps os.File to implement WriteSyncer
type FileWriteSyncer struct {
	file *os.File
	mu   sync.Mutex
}

// NewFileWriteSyncer creates a new FileWriteSyncer for the given filename
func NewFileWriteSyncer(filename string) (*FileWriteSyncer, error) {
	// #nosec G304 - File path is provided by application, not user input
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600) // Secure permissions
	if err != nil {
		return nil, err
	}

	return &FileWriteSyncer{file: file}, nil
}

// Write implements io.Writer
func (f *FileWriteSyncer) Write(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Write(p)
}

// Sync implements WriteSyncer
func (f *FileWriteSyncer) Sync() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Sync()
}

// Close closes the underlying file
func (f *FileWriteSyncer) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Close()
}

// BufferedWriteSyncer adds buffering to any WriteSyncer
type BufferedWriteSyncer struct {
	writer WriteSyncer
	buffer []byte
	pos    int
	mu     sync.Mutex
}

// NewBufferedWriteSyncer creates a buffered WriteSyncer with the specified buffer size
func NewBufferedWriteSyncer(writer WriteSyncer, bufferSize int) *BufferedWriteSyncer {
	return &BufferedWriteSyncer{
		writer: writer,
		buffer: make([]byte, bufferSize),
	}
}

// Write implements io.Writer with buffering
func (b *BufferedWriteSyncer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	totalWritten := 0

	for len(p) > 0 {
		// Calculate how much space is left in buffer
		spaceLeft := len(b.buffer) - b.pos

		if spaceLeft == 0 {
			// Buffer is full, flush it
			if err := b.flushUnsafe(); err != nil {
				return totalWritten, err
			}
			spaceLeft = len(b.buffer)
		}

		// Write as much as possible to buffer
		toCopy := len(p)
		if toCopy > spaceLeft {
			toCopy = spaceLeft
		}

		copy(b.buffer[b.pos:], p[:toCopy])
		b.pos += toCopy
		p = p[toCopy:]
		totalWritten += toCopy
	}

	return totalWritten, nil
}

// Sync flushes the buffer and syncs the underlying writer
func (b *BufferedWriteSyncer) Sync() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.flushUnsafe(); err != nil {
		return err
	}

	return b.writer.Sync()
}

// flushUnsafe flushes the buffer without locking (caller must hold lock)
func (b *BufferedWriteSyncer) flushUnsafe() error {
	if b.pos == 0 {
		return nil
	}

	_, err := b.writer.Write(b.buffer[:b.pos])
	b.pos = 0

	return err
}

// DiscardWriteSyncer is a WriteSyncer that discards all writes (useful for benchmarks)
type DiscardWriteSyncer struct{}

// Write implements io.Writer
func (d *DiscardWriteSyncer) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Sync implements WriteSyncer
func (d *DiscardWriteSyncer) Sync() error {
	return nil
}

// Common WriteSyncer instances
var (
	// StdoutWriteSyncer writes to os.Stdout
	StdoutWriteSyncer = &WriteSyncerWrapper{Writer: os.Stdout}

	// StderrWriteSyncer writes to os.Stderr
	StderrWriteSyncer = &WriteSyncerWrapper{Writer: os.Stderr}

	// DiscardWriteSyncer discards all writes
	DiscardSyncer = &DiscardWriteSyncer{}
)
