// detection.go: Lethe detection and interface definitions
//
// This package provides runtime detection of Lethe integration capabilities
// without creating hard dependencies. When both Iris and Lethe are imported
// together, automatic optimizations are unlocked through interface detection.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package lethe

// LetheWriter defines the enhanced interface that Lethe writers implement
// When Iris detects this interface, it automatically enables optimizations
type LetheWriter interface {
	// Standard WriteSyncer methods
	Write([]byte) (int, error)
	Sync() error
	Close() error

	// Lethe-specific optimization methods
	WriteOwned([]byte) (int, error) // Zero-copy write for owned buffers
	GetOptimalBufferSize() int      // Auto-tuning hint
	SupportsHotReload() bool        // Configuration hot-reload capability
}

// DetectLetheCapabilities checks if a writer supports Lethe optimizations
// Returns the enhanced interface if available, nil otherwise
func DetectLetheCapabilities(writer interface{}) LetheWriter {
	if letheWriter, ok := writer.(LetheWriter); ok {
		return letheWriter
	}
	return nil
}

// IsLetheWriter checks if a writer implements Lethe optimizations
func IsLetheWriter(writer interface{}) bool {
	_, ok := writer.(LetheWriter)
	return ok
}
