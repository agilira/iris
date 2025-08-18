package iris

import (
	"io"
	"sync"
)

// WriteSyncer combines io.Writer and Sync functionality
type WriteSyncer interface {
	io.Writer
	Sync() error
}

// MultiWriter implements WriteSyncer for multiple outputs
type MultiWriter struct {
	writers []WriteSyncer
	mu      sync.RWMutex
}

// NewMultiWriter creates a new MultiWriter that writes to all provided writers
func NewMultiWriter(writers ...WriteSyncer) *MultiWriter {
	if len(writers) == 0 {
		return &MultiWriter{writers: []WriteSyncer{}}
	}

	// Create a copy to avoid external modifications
	writersCopy := make([]WriteSyncer, len(writers))
	copy(writersCopy, writers)

	return &MultiWriter{
		writers: writersCopy,
	}
}

// Write implements io.Writer, writing to all underlying writers
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	if len(mw.writers) == 0 {
		return len(p), nil
	}

	// Write to all writers, collecting any errors
	var firstErr error
	for _, w := range mw.writers {
		nn, werr := w.Write(p)
		if werr != nil && firstErr == nil {
			firstErr = werr
		}
		// Return the bytes written by the first writer as the canonical count
		if w == mw.writers[0] {
			n = nn
		}
	}

	return n, firstErr
}

// Sync implements WriteSyncer, syncing all underlying writers
func (mw *MultiWriter) Sync() error {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	var firstErr error
	for _, w := range mw.writers {
		if err := w.Sync(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// AddWriter adds a new writer to the MultiWriter
func (mw *MultiWriter) AddWriter(writer WriteSyncer) {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	mw.writers = append(mw.writers, writer)
}

// RemoveWriter removes a writer from the MultiWriter
func (mw *MultiWriter) RemoveWriter(writer WriteSyncer) bool {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	for i, w := range mw.writers {
		if w == writer {
			// Remove writer by swapping with last element and truncating
			mw.writers[i] = mw.writers[len(mw.writers)-1]
			mw.writers = mw.writers[:len(mw.writers)-1]
			return true
		}
	}

	return false
}

// Writers returns a copy of the current writers slice
func (mw *MultiWriter) Writers() []WriteSyncer {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	writers := make([]WriteSyncer, len(mw.writers))
	copy(writers, mw.writers)

	return writers
}

// Count returns the number of writers
func (mw *MultiWriter) Count() int {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	return len(mw.writers)
}

// WriteSyncerWrapper wraps an io.Writer to implement WriteSyncer
type WriteSyncerWrapper struct {
	io.Writer
}

// Sync is a no-op for WriteSyncerWrapper
func (w *WriteSyncerWrapper) Sync() error {
	return nil
}

// WrapWriter wraps an io.Writer to implement WriteSyncer
func WrapWriter(w io.Writer) WriteSyncer {
	if ws, ok := w.(WriteSyncer); ok {
		return ws
	}
	return &WriteSyncerWrapper{Writer: w}
}
