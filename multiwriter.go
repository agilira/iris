package iris

import (
	"io"
	"sync"
	"sync/atomic"
	"unsafe"
)

// WriteSyncer combines io.Writer and Sync functionality
type WriteSyncer interface {
	io.Writer
	Sync() error
}

// MultiWriter implements WriteSyncer for multiple outputs
// Optimized for high-performance concurrent access with lock-free reads
type MultiWriter struct {
	// Using atomic pointer for lock-free reads in common case
	writers unsafe.Pointer // *[]WriteSyncer
	mu      sync.RWMutex   // Only used for modifications
}

// NewMultiWriter creates a new MultiWriter that writes to all provided writers
func NewMultiWriter(writers ...WriteSyncer) *MultiWriter {
	var writersCopy []WriteSyncer

	if len(writers) == 0 {
		writersCopy = make([]WriteSyncer, 0)
	} else {
		// Create a copy to avoid external modifications
		writersCopy = make([]WriteSyncer, len(writers))
		copy(writersCopy, writers)
	}

	mw := &MultiWriter{}
	atomic.StorePointer(&mw.writers, unsafe.Pointer(&writersCopy))
	return mw
}

// getWriters returns the current writers slice using atomic load
func (mw *MultiWriter) getWriters() []WriteSyncer {
	return *(*[]WriteSyncer)(atomic.LoadPointer(&mw.writers))
}

// setWriters sets the writers slice using atomic store
func (mw *MultiWriter) setWriters(writers []WriteSyncer) {
	atomic.StorePointer(&mw.writers, unsafe.Pointer(&writers))
}

// Write implements io.Writer, writing to all underlying writers
// Optimized with lock-free read for the common case
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	writers := mw.getWriters()

	if len(writers) == 0 {
		return len(p), nil
	}

	// Fast path: single writer case (no locking needed)
	if len(writers) == 1 {
		return writers[0].Write(p)
	}

	// Multi-writer case: write to all writers, collecting any errors
	var firstErr error
	for i, w := range writers {
		nn, werr := w.Write(p)
		if werr != nil && firstErr == nil {
			firstErr = werr
		}
		// Return the bytes written by the first writer as the canonical count
		if i == 0 {
			n = nn
		}
	}

	return n, firstErr
}

// Sync implements WriteSyncer, syncing all underlying writers
// Optimized with lock-free read
func (mw *MultiWriter) Sync() error {
	writers := mw.getWriters()

	var firstErr error
	for _, w := range writers {
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

	writers := mw.getWriters()
	newWriters := make([]WriteSyncer, len(writers)+1)
	copy(newWriters, writers)
	newWriters[len(writers)] = writer
	mw.setWriters(newWriters)
}

// RemoveWriter removes a writer from the MultiWriter
func (mw *MultiWriter) RemoveWriter(writer WriteSyncer) bool {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	writers := mw.getWriters()
	for i, w := range writers {
		if w == writer {
			// Remove writer by swapping with last element and truncating
			newWriters := make([]WriteSyncer, len(writers)-1)
			copy(newWriters[:i], writers[:i])
			copy(newWriters[i:], writers[i+1:])
			mw.setWriters(newWriters)
			return true
		}
	}

	return false
}

// Writers returns a copy of the current writers slice
func (mw *MultiWriter) Writers() []WriteSyncer {
	writers := mw.getWriters()
	writersCopy := make([]WriteSyncer, len(writers))
	copy(writersCopy, writers)
	return writersCopy
}

// Count returns the number of writers
func (mw *MultiWriter) Count() int {
	writers := mw.getWriters()
	return len(writers)
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
