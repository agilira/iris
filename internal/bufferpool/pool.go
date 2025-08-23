// pool.go: High-performance buffer pool for zero-allocation logging
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package bufferpool

import (
	"bytes"
	"sync"
	"sync/atomic"
)

// Pool statistics for monitoring and debugging
var (
	getCount   int64 // Number of Get() calls
	putCount   int64 // Number of Put() calls
	allocCount int64 // Number of new allocations
	dropCount  int64 // Number of oversized buffers dropped
)

// Configuration constants for buffer pool behavior
const (
	// MaxBufferSize is the maximum buffer capacity before dropping.
	// Buffers larger than this are discarded to prevent memory bloat.
	MaxBufferSize = 1 << 20 // 1 MiB

	// DefaultCapacity is the initial capacity hint for new buffers.
	// This reduces reallocations for typical log entry sizes.
	DefaultCapacity = 512 // 512 bytes
)

// pool is the global sync.Pool for reusing byte buffers.
// Using sync.Pool provides automatic garbage collection coordination
// and scales well across multiple goroutines.
var pool = sync.Pool{
	New: func() any {
		atomic.AddInt64(&allocCount, 1)
		// Pre-allocate with default capacity to reduce early reallocations
		buf := bytes.NewBuffer(make([]byte, 0, DefaultCapacity))
		return buf
	},
}

// Get restituisce un *bytes.Buffer pulito (Reset) dal pool.
// Incrementa automaticamente le statistiche e garantisce che il buffer
// sia pronto per l'uso immediato senza contenuti precedenti.
func Get() *bytes.Buffer {
	atomic.AddInt64(&getCount, 1)
	b := pool.Get().(*bytes.Buffer)
	b.Reset() // Ensure buffer is clean
	return b
}

// Put restituisce il buffer al pool. Se il buffer è cresciuto troppo,
// lo azzera per evitare growth non controllato della memoria.
// Questa strategia bilancia performance e utilizzo memoria.
func Put(b *bytes.Buffer) {
	if b == nil {
		return
	}

	atomic.AddInt64(&putCount, 1)

	// Taglio semplice: se il cap supera MaxBufferSize, rilascia il backing.
	// Questo previene che buffers molto grandi rimangano nel pool
	// consumando memoria inutilmente.
	if b.Cap() > MaxBufferSize {
		atomic.AddInt64(&dropCount, 1)
		// Sostituisce il buffer con uno nuovo per rilasciare la memoria
		*b = *bytes.NewBuffer(make([]byte, 0, DefaultCapacity))
	}

	b.Reset() // Clean buffer before returning to pool
	pool.Put(b)
}

// Stats returns current buffer pool statistics for monitoring.
// Useful for debugging memory usage and pool efficiency.
type Stats struct {
	Gets        int64 // Total number of Get() calls
	Puts        int64 // Total number of Put() calls
	Allocations int64 // Total number of new buffer allocations
	Drops       int64 // Total number of oversized buffers dropped
}

// GetStats returns a snapshot of current pool statistics.
// Thread-safe and can be called from multiple goroutines.
func GetStats() Stats {
	return Stats{
		Gets:        atomic.LoadInt64(&getCount),
		Puts:        atomic.LoadInt64(&putCount),
		Allocations: atomic.LoadInt64(&allocCount),
		Drops:       atomic.LoadInt64(&dropCount),
	}
}

// ResetStats resets all pool statistics to zero.
// Useful for benchmarking and testing scenarios.
func ResetStats() {
	atomic.StoreInt64(&getCount, 0)
	atomic.StoreInt64(&putCount, 0)
	atomic.StoreInt64(&allocCount, 0)
	atomic.StoreInt64(&dropCount, 0)
}
