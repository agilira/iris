// sink.go: High-performance I/O synchronization utilities for iris logging
//
// This file provides efficient writer synchronization primitives optimized for
// high-throughput logging scenarios. The WriteSyncer interface and related
// implementations ensure data durability while maintaining minimal overhead.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"context"
	"io"
	"os"
)

// WriteSyncer combines io.Writer with the ability to synchronize written data
// to persistent storage. This interface is essential for ensuring data durability
// in logging scenarios where data loss is unacceptable.
//
// Performance considerations:
// - Sync() should be called judiciously as it may involve expensive syscalls
// - Implementations should be thread-safe for concurrent logging scenarios
// - Zero allocations in hot paths for maximum throughput
type WriteSyncer interface {
	io.Writer
	Sync() error
}

// SyncWriter provides enhanced writer capabilities for external output destinations
// such as Loki, Kafka, Prometheus, etc. This interface enables modular output
// architecture where specialized writers are maintained as separate modules.
//
// SyncWriter extends basic io.Writer with structured record processing, allowing
// external writer modules to access Iris's rich Record format with fields,
// levels, and metadata while maintaining zero dependencies in the core library.
//
// Performance considerations:
// - WriteRecord() should be non-blocking or implement internal buffering
// - Implementations should handle backpressure gracefully
// - Background processing recommended for network/disk operations
type SyncWriter interface {
	// WriteRecord writes a structured log record to the destination.
	// Should handle the record asynchronously to avoid blocking Iris's hot path.
	WriteRecord(record *Record) error

	// Close releases any resources and flushes pending data.
	// Should ensure all data is safely written before returning.
	io.Closer
}

// SyncReader provides the ability to read log records from external logging systems
// and integrate them into Iris's high-performance processing pipeline. This interface
// enables Iris to act as a universal logging accelerator for existing logger implementations.
//
// The SyncReader operates in background goroutines and feeds records into Iris's
// lock-free ring buffer, allowing existing loggers (slog, logrus, zap) to benefit
// from Iris's performance and advanced features without code changes.
//
// Performance considerations:
// - Read() operates in separate goroutines to avoid blocking Iris's hot path
// - Implementations should handle backpressure gracefully
// - Context cancellation should be respected for clean shutdowns
type SyncReader interface {
	// Read retrieves the next log record from the external logging system.
	// Returns nil when no more records are available or context is cancelled.
	// Implementations should block until a record is available or context expires.
	Read(ctx context.Context) (*Record, error)

	// Close releases any resources associated with the reader.
	// Should be called when the reader is no longer needed.
	io.Closer
}

// nopSyncer wraps any io.Writer and provides a no-op Sync() implementation.
// This is used for writers that don't require explicit synchronization
// (e.g., in-memory buffers, network connections) or where sync is handled
// at a different layer.
//
// Performance: Zero-cost abstraction with inline Sync() calls
type nopSyncer struct{ io.Writer }

// Sync implements WriteSyncer.Sync() with a no-op operation.
// Returns nil immediately, assuming the underlying writer handles
// synchronization internally or doesn't require it.
func (n nopSyncer) Sync() error { return nil }

// fileSyncer wraps *os.File to provide explicit file synchronization.
// This ensures that written data is flushed to persistent storage,
// which is critical for durability guarantees in logging systems.
//
// Performance: Uses os.File.Sync() syscall which may block briefly
type fileSyncer struct{ *os.File }

// Sync implements WriteSyncer.Sync() by calling the underlying file's Sync().
// This forces a flush of all written data to persistent storage, ensuring
// durability at the cost of potential I/O blocking.
func (f fileSyncer) Sync() error { return f.File.Sync() }

// WrapWriter intelligently converts any io.Writer into a WriteSyncer.
// This function provides automatic detection and wrapping of different writer
// types to ensure optimal performance and correct synchronization behavior.
//
// Type-specific optimizations:
// - *os.File: Uses fileSyncer for explicit sync() syscalls
// - WriteSyncer: Returns as-is (already implements interface)
// - Other writers: Uses nopSyncer (no-op sync for non-file writers)
//
// Performance: Zero allocations for WriteSyncer inputs, minimal overhead
// for type switching in other cases.
//
// Usage patterns:
//   - File logging: WrapWriter(file) -> fileSyncer (with sync)
//   - Buffer logging: WrapWriter(buffer) -> nopSyncer (no sync needed)
//   - Network logging: WrapWriter(conn) -> nopSyncer (sync at protocol level)
func WrapWriter(w io.Writer) WriteSyncer {
	switch t := w.(type) {
	case *os.File:
		return fileSyncer{t}
	case WriteSyncer:
		return t
	default:
		return nopSyncer{w}
	}
}

// AddSync is an alias for WrapWriter for familiarity with zap
func AddSync(w io.Writer) WriteSyncer { return WrapWriter(w) }

// multiWS implements fan-out to multiple destinations
type multiWS struct{ ws []WriteSyncer }

// MultiWriteSyncer creates a WriteSyncer that duplicates writes to multiple writers
func MultiWriteSyncer(writers ...WriteSyncer) WriteSyncer {
	cp := make([]WriteSyncer, 0, len(writers))
	for _, w := range writers {
		if w != nil {
			cp = append(cp, w)
		}
	}
	return &multiWS{ws: cp}
}

// MultiWriter accepts io.Writer interfaces, wraps them and creates a MultiWriteSyncer
func MultiWriter(writers ...io.Writer) WriteSyncer {
	ws := make([]WriteSyncer, 0, len(writers))
	for _, w := range writers {
		ws = append(ws, WrapWriter(w))
	}
	return MultiWriteSyncer(ws...)
}

func (m *multiWS) Write(p []byte) (int, error) {
	var firstErr error
	for _, w := range m.ws {
		if _, err := w.Write(p); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	// ritorna len(p) se almeno uno ha scritto, preservando primo errore
	if firstErr != nil {
		return 0, firstErr
	}
	return len(p), nil
}

func (m *multiWS) Sync() error {
	var firstErr error
	for _, w := range m.ws {
		if err := w.Sync(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// NewFileSyncer creates a WriteSyncer specifically for file operations.
// This function provides explicit file syncing capabilities and should be
// used when you need guaranteed durability for file-based logging.
//
// Performance: Direct file operations with explicit sync control
func NewFileSyncer(file *os.File) WriteSyncer {
	return fileSyncer{file}
}

// NewNopSyncer creates a WriteSyncer that performs no synchronization.
// This is useful for scenarios where sync is handled externally or
// where the underlying writer doesn't support/need synchronization.
//
// Performance: Zero-cost wrapper with inline no-op sync
func NewNopSyncer(w io.Writer) WriteSyncer {
	return nopSyncer{w}
}

// IsFileSyncer checks if a WriteSyncer is backed by a file.
// This can be useful for conditional logic based on the underlying
// writer type, such as applying different buffering strategies.
func IsFileSyncer(ws WriteSyncer) bool {
	_, ok := ws.(fileSyncer)
	return ok
}

// IsNopSyncer checks if a WriteSyncer uses no-op synchronization.
// This can help optimize write patterns when sync operations are
// known to be no-ops.
func IsNopSyncer(ws WriteSyncer) bool {
	_, ok := ws.(nopSyncer)
	return ok
}

// readerLogger extends a standard Logger to support background processing
// of external log sources through SyncReader interfaces. This enables
// existing logger integrations while maintaining Iris's performance characteristics.
//
// The readerLogger runs SyncReaders in separate goroutines that feed into
// the main Iris ring buffer, allowing external loggers to benefit from
// Iris's advanced features (OpenTelemetry, Loki integration, security, etc.)
// without performance impact on the primary logging path.
type readerLogger struct {
	*Logger
	readers []SyncReader
	done    chan struct{}
}

// NewReaderLogger creates a logger that processes both direct logging calls
// and background readers. The underlying Logger performance is preserved while
// external log sources are processed asynchronously.
//
// Parameters:
//   - config: Standard Iris logger configuration
//   - readers: External log sources to process in background
//   - opts: Standard Iris logger options
//
// Returns:
//   - *readerLogger: Extended logger with reader support
//   - error: Configuration or setup error
//
// Performance: Zero impact on direct logging, background readers operate
// in separate goroutines feeding into the same high-performance ring buffer.
func NewReaderLogger(config Config, readers []SyncReader, opts ...Option) (*readerLogger, error) {
	// Create standard Iris logger (performance unchanged)
	base, err := New(config, opts...)
	if err != nil {
		return nil, err
	}

	rl := &readerLogger{
		Logger:  base,
		readers: readers,
		done:    make(chan struct{}),
	}

	return rl, nil
}

// Start begins both the standard Iris logger and background reader processing.
// Each SyncReader runs in its own goroutine, feeding records into the main
// Iris ring buffer for processing with full feature support.
func (rl *readerLogger) Start() {
	// Start the base Iris logger (unchanged performance)
	rl.Logger.Start()

	// Start background readers
	for _, reader := range rl.readers {
		go rl.processReader(reader)
	}
}

// Close gracefully shuts down both the logger and all background readers.
// Ensures all buffered records from readers are processed before termination.
func (rl *readerLogger) Close() error {
	// Signal all readers to stop
	close(rl.done)

	// Close all readers
	for _, reader := range rl.readers {
		if err := reader.Close(); err != nil {
			// Log error but continue closing others
			rl.Logger.Error("Failed to close reader", String("error", err.Error()))
		}
	}

	// Close the base logger
	return rl.Logger.Close()
}

// processReader handles a single SyncReader in a background goroutine.
// Records from the reader are fed into Iris's ring buffer, gaining access
// to all Iris features (OpenTelemetry, Loki, security, sampling, etc.)
func (rl *readerLogger) processReader(reader SyncReader) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Monitor done channel for shutdown
	go func() {
		<-rl.done
		cancel()
	}()

	for {
		// Read record from external source
		record, err := reader.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled, clean shutdown
				return
			}
			// Log error and continue
			rl.Logger.Error("Reader error", String("error", err.Error()))
			continue
		}

		if record == nil {
			// No more records or reader closed
			return
		}

		// Feed into Iris ring buffer (gains all Iris features automatically)
		rl.Logger.Write(func(slot *Record) {
			*slot = *record // Direct copy - all Iris processing applies
		})
	}
}
