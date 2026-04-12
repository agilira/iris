// Package lethebridge provides the public interface for Lethe integration.
//
// This package enables the "H1 chip" pattern - when both Iris and Lethe
// are imported, they automatically work together without any configuration.
//
// Lethe registers its capabilities via init(), and Iris's Magic API
// detects and uses them automatically.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package lethebridge

import "sync"

// LetheWriter defines the enhanced interface that Lethe writers implement.
// When Iris detects this interface, it automatically enables optimizations.
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

// CapabilityProvider defines functions that can be registered by Lethe.
type CapabilityProvider struct {
	Name                string
	CreateOptimizedSink func(string, ...interface{}) (interface{}, error)
	DetectWriter        func(interface{}) bool
}

var (
	registry = make(map[string]CapabilityProvider)
	mu       sync.RWMutex
)

// RegisterCapability allows Lethe to register its capabilities at runtime.
// This is called by Lethe's init() function when both packages are imported.
func RegisterCapability(provider CapabilityProvider) {
	mu.Lock()
	defer mu.Unlock()
	registry[provider.Name] = provider
}

// HasLetheCapabilities checks if any Lethe providers are registered.
func HasLetheCapabilities() bool {
	mu.RLock()
	defer mu.RUnlock()
	return len(registry) > 0
}

// GetRegisteredCapabilities returns all registered Lethe capabilities.
func GetRegisteredCapabilities() []CapabilityProvider {
	mu.RLock()
	defer mu.RUnlock()

	capabilities := make([]CapabilityProvider, 0, len(registry))
	for _, provider := range registry {
		capabilities = append(capabilities, provider)
	}
	return capabilities
}

// GetLetheProvider returns the Lethe provider if available.
func GetLetheProvider() (CapabilityProvider, bool) {
	mu.RLock()
	defer mu.RUnlock()

	if provider, exists := registry["lethe"]; exists {
		return provider, true
	}
	return CapabilityProvider{}, false
}

// DetectLetheCapabilities checks if a writer supports Lethe optimizations.
// Returns the enhanced interface if available, nil otherwise.
func DetectLetheCapabilities(writer interface{}) LetheWriter {
	if letheWriter, ok := writer.(LetheWriter); ok {
		return letheWriter
	}
	return nil
}

// IsLetheWriter checks if a writer implements Lethe optimizations.
func IsLetheWriter(writer interface{}) bool {
	_, ok := writer.(LetheWriter)
	return ok
}
