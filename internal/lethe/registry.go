// registry.go: Runtime registration system for Lethe integration
//
// This provides the automatic magic through runtime registration when both
// Iris and Lethe packages are imported together. Zero dependencies, pure magic.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package lethe

import "sync"

// CapabilityProvider defines functions that can be registered by Lethe
type CapabilityProvider struct {
	Name                string
	CreateOptimizedSink func(string, ...interface{}) (interface{}, error)
	DetectWriter        func(interface{}) bool
}

var (
	registry = make(map[string]CapabilityProvider)
	mu       sync.RWMutex
)

// RegisterCapability allows Lethe to register its capabilities at runtime
// This is called by Lethe's init() function when both packages are imported
func RegisterCapability(provider CapabilityProvider) {
	mu.Lock()
	defer mu.Unlock()
	registry[provider.Name] = provider
}

// GetRegisteredCapabilities returns all registered Lethe capabilities
func GetRegisteredCapabilities() []CapabilityProvider {
	mu.RLock()
	defer mu.RUnlock()

	capabilities := make([]CapabilityProvider, 0, len(registry))
	for _, provider := range registry {
		capabilities = append(capabilities, provider)
	}
	return capabilities
}

// HasLetheCapabilities checks if any Lethe providers are registered
func HasLetheCapabilities() bool {
	mu.RLock()
	defer mu.RUnlock()
	return len(registry) > 0
}

// GetLetheProvider returns the Lethe provider if available
func GetLetheProvider() (CapabilityProvider, bool) {
	mu.RLock()
	defer mu.RUnlock()

	if provider, exists := registry["lethe"]; exists {
		return provider, true
	}
	return CapabilityProvider{}, false
}
