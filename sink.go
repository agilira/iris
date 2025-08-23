// sink.go: High-performance I/O synchronization utilities for iris logging
//
// This file provides efficient writer synchronization primitives optimized for
// high-throughput logging scenarios. The WriteSyncer interface and related
// implementations ensure data durability while maintaining minimal overhead.
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
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

// AddSync: alias di WrapWriter per familiarità con zap.
func AddSync(w io.Writer) WriteSyncer { return WrapWriter(w) }

// MultiWriteSyncer: fan-out su più destinazioni.
type multiWS struct{ ws []WriteSyncer }

func MultiWriteSyncer(writers ...WriteSyncer) WriteSyncer {
	cp := make([]WriteSyncer, 0, len(writers))
	for _, w := range writers {
		if w != nil {
			cp = append(cp, w)
		}
	}
	return &multiWS{ws: cp}
}

// MultiWriter: accetta io.Writer, li wrappa e crea un MultiWriteSyncer.
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
