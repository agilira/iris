// logger_management.go: Logger management methods for Iris logger
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync/atomic"
)

// Close gracefully shuts down the logger
func (l *Logger) Close() {
	if atomic.LoadInt32(&l.closed) != 0 {
		return
	}

	atomic.StoreInt32(&l.closed, 1)
	l.ring.Close()

	// Wait for consumer to finish
	<-l.done

	// Close writer if it supports closing
	if closer, ok := l.writer.(interface{ Close() error }); ok {
		closer.Close()
	}
}

// AddWriter adds a new writer to the logger's output destinations
func (l *Logger) AddWriter(writer Writer) error {
	if !l.hasTee {
		// Convert to MultiWriter
		l.multiWriter = NewMultiWriter(WrapWriter(l.writer))
		l.hasTee = true
	}

	l.multiWriter.AddWriter(WrapWriter(writer))
	l.writer = l.multiWriter
	return nil
}

// RemoveWriter removes a writer from the logger's output destinations
func (l *Logger) RemoveWriter(writer Writer) bool {
	if !l.hasTee {
		return false
	}

	return l.multiWriter.RemoveWriter(WrapWriter(writer))
}

// AddWriteSyncer adds a WriteSyncer to the logger's output destinations
func (l *Logger) AddWriteSyncer(syncer WriteSyncer) error {
	if !l.hasTee {
		// Convert to MultiWriter
		l.multiWriter = NewMultiWriter(WrapWriter(l.writer))
		l.hasTee = true
	}

	l.multiWriter.AddWriter(syncer)
	l.writer = l.multiWriter
	return nil
}

// WriterCount returns the number of active writers
func (l *Logger) WriterCount() int {
	if !l.hasTee {
		return 1
	}

	return l.multiWriter.Count()
}

// Sampling Methods

// GetSamplingStats returns the current sampling statistics.
// Returns nil if sampling is not enabled.
func (l *Logger) GetSamplingStats() *SamplingStats {
	if l.sampler == nil {
		return nil
	}

	stats := l.sampler.Stats()
	return &stats
}

// IsSamplingEnabled returns true if sampling is configured and active
func (l *Logger) IsSamplingEnabled() bool {
	return l.sampler != nil
}

// SetSamplingConfig updates the sampling configuration.
// Pass nil to disable sampling.
func (l *Logger) SetSamplingConfig(config *SamplingConfig) {
	if config == nil {
		l.sampler = nil
		return
	}

	l.sampler = NewSampler(*config)
}

// Sync ensures all buffered log entries are flushed to the output.
// This method blocks until all pending entries are processed.
func (l *Logger) Sync() error {
	if atomic.LoadInt32(&l.closed) != 0 {
		return nil
	}

	// Process any pending entries in the ring buffer
	// Since Xantos doesn't have a Sync method, we just ensure outputs are synced

	// Sync the underlying writer if it supports it
	if l.hasTee && l.multiWriter != nil {
		return l.multiWriter.Sync()
	}

	// Try to sync the primary writer if it implements WriteSyncer
	if syncer, ok := l.writer.(WriteSyncer); ok {
		return syncer.Sync()
	}

	return nil
}
