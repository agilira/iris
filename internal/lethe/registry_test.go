// registry_test.go: Tests for Lethe capability registry
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package lethe

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRegisterCapability(t *testing.T) {
	// Save original state
	originalRegistry := make(map[string]CapabilityProvider)
	mu.Lock()
	for k, v := range registry {
		originalRegistry[k] = v
	}
	// Clear registry for test
	registry = make(map[string]CapabilityProvider)
	mu.Unlock()

	// Restore original state after test
	defer func() {
		mu.Lock()
		registry = originalRegistry
		mu.Unlock()
	}()

	provider := CapabilityProvider{
		Name: "test-provider",
		CreateOptimizedSink: func(filename string, args ...interface{}) (interface{}, error) {
			return &mockRegistryWriter{filename: filename}, nil
		},
		DetectWriter: func(writer interface{}) bool {
			_, ok := writer.(*mockRegistryWriter)
			return ok
		},
	}

	// Test registration
	RegisterCapability(provider)

	// Verify registration
	if !HasLetheCapabilities() {
		t.Error("HasLetheCapabilities should return true after registration")
	}

	capabilities := GetRegisteredCapabilities()
	if len(capabilities) != 1 {
		t.Errorf("Expected 1 registered capability, got %d", len(capabilities))
	}

	if capabilities[0].Name != "test-provider" {
		t.Errorf("Expected provider name 'test-provider', got '%s'", capabilities[0].Name)
	}
}

func TestGetLetheProvider(t *testing.T) {
	// Save and clear registry
	originalRegistry := make(map[string]CapabilityProvider)
	mu.Lock()
	for k, v := range registry {
		originalRegistry[k] = v
	}
	registry = make(map[string]CapabilityProvider)
	mu.Unlock()

	defer func() {
		mu.Lock()
		registry = originalRegistry
		mu.Unlock()
	}()

	// Test when no Lethe provider is registered
	_, exists := GetLetheProvider()
	if exists {
		t.Error("GetLetheProvider should return false when no Lethe provider is registered")
	}

	// Register a non-Lethe provider
	nonLetheProvider := CapabilityProvider{Name: "other-provider"}
	RegisterCapability(nonLetheProvider)

	_, exists = GetLetheProvider()
	if exists {
		t.Error("GetLetheProvider should return false when only non-Lethe providers are registered")
	}

	// Register Lethe provider
	letheProvider := CapabilityProvider{
		Name: "lethe",
		CreateOptimizedSink: func(filename string, args ...interface{}) (interface{}, error) {
			return &mockRegistryWriter{filename: filename}, nil
		},
	}
	RegisterCapability(letheProvider)

	provider, exists := GetLetheProvider()
	if !exists {
		t.Error("GetLetheProvider should return true when Lethe provider is registered")
	}

	if provider.Name != "lethe" {
		t.Errorf("Expected Lethe provider name 'lethe', got '%s'", provider.Name)
	}
}

func TestMultipleProviders(t *testing.T) {
	// Save and clear registry
	originalRegistry := make(map[string]CapabilityProvider)
	mu.Lock()
	for k, v := range registry {
		originalRegistry[k] = v
	}
	registry = make(map[string]CapabilityProvider)
	mu.Unlock()

	defer func() {
		mu.Lock()
		registry = originalRegistry
		mu.Unlock()
	}()

	// Register multiple providers
	providers := []CapabilityProvider{
		{Name: "provider1"},
		{Name: "provider2"},
		{Name: "lethe"},
		{Name: "provider3"},
	}

	for _, provider := range providers {
		RegisterCapability(provider)
	}

	// Test count
	capabilities := GetRegisteredCapabilities()
	if len(capabilities) != 4 {
		t.Errorf("Expected 4 registered capabilities, got %d", len(capabilities))
	}

	// Test that all providers are present
	names := make(map[string]bool)
	for _, cap := range capabilities {
		names[cap.Name] = true
	}

	expectedNames := []string{"provider1", "provider2", "lethe", "provider3"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("Expected provider '%s' to be registered", name)
		}
	}

	// Test Lethe-specific lookup
	letheProvider, exists := GetLetheProvider()
	if !exists {
		t.Error("Should find Lethe provider among multiple providers")
	}
	if letheProvider.Name != "lethe" {
		t.Errorf("Expected Lethe provider, got '%s'", letheProvider.Name)
	}
}

func TestConcurrentRegistration(t *testing.T) {
	// Save and clear registry
	originalRegistry := make(map[string]CapabilityProvider)
	mu.Lock()
	for k, v := range registry {
		originalRegistry[k] = v
	}
	registry = make(map[string]CapabilityProvider)
	mu.Unlock()

	defer func() {
		mu.Lock()
		registry = originalRegistry
		mu.Unlock()
	}()

	const numGoroutines = 10
	const providersPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent registration
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < providersPerGoroutine; j++ {
				provider := CapabilityProvider{
					Name: fmt.Sprintf("provider-%d-%d", goroutineID, j),
				}
				RegisterCapability(provider)
			}
		}(i)
	}

	wg.Wait()

	// Verify all providers were registered
	capabilities := GetRegisteredCapabilities()
	expectedCount := numGoroutines * providersPerGoroutine
	if len(capabilities) != expectedCount {
		t.Errorf("Expected %d registered capabilities, got %d", expectedCount, len(capabilities))
	}

	// Verify thread safety by checking HasLetheCapabilities
	if !HasLetheCapabilities() {
		t.Error("HasLetheCapabilities should return true after concurrent registrations")
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	// Save and clear registry
	originalRegistry := make(map[string]CapabilityProvider)
	mu.Lock()
	for k, v := range registry {
		originalRegistry[k] = v
	}
	registry = make(map[string]CapabilityProvider)
	mu.Unlock()

	defer func() {
		mu.Lock()
		registry = originalRegistry
		mu.Unlock()
	}()

	const numWriters = 5
	const numReaders = 10
	const duration = 100 * time.Millisecond

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start writers
	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			defer wg.Done()
			counter := 0
			for {
				select {
				case <-done:
					return
				default:
					provider := CapabilityProvider{
						Name: fmt.Sprintf("writer-%d-provider-%d", writerID, counter),
					}
					RegisterCapability(provider)
					counter++
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	// Start readers
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func(readerID int) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_ = HasLetheCapabilities()
					_ = GetRegisteredCapabilities()
					_, _ = GetLetheProvider()
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	// Let it run for a while
	time.Sleep(duration)
	close(done)
	wg.Wait()

	// Verify registry is still functional
	if !HasLetheCapabilities() {
		t.Error("Registry should still be functional after concurrent access")
	}

	capabilities := GetRegisteredCapabilities()
	if len(capabilities) == 0 {
		t.Error("Should have registered capabilities after concurrent access")
	}
}

func TestRegistryProviderOverwrite(t *testing.T) {
	// Save and clear registry
	originalRegistry := make(map[string]CapabilityProvider)
	mu.Lock()
	for k, v := range registry {
		originalRegistry[k] = v
	}
	registry = make(map[string]CapabilityProvider)
	mu.Unlock()

	defer func() {
		mu.Lock()
		registry = originalRegistry
		mu.Unlock()
	}()

	// Register first provider
	provider1 := CapabilityProvider{
		Name: "lethe",
		CreateOptimizedSink: func(filename string, args ...interface{}) (interface{}, error) {
			return "first", nil
		},
	}
	RegisterCapability(provider1)

	// Register second provider with same name (should overwrite)
	provider2 := CapabilityProvider{
		Name: "lethe",
		CreateOptimizedSink: func(filename string, args ...interface{}) (interface{}, error) {
			return "second", nil
		},
	}
	RegisterCapability(provider2)

	// Should still have only one provider
	capabilities := GetRegisteredCapabilities()
	if len(capabilities) != 1 {
		t.Errorf("Expected 1 capability after overwrite, got %d", len(capabilities))
	}

	// Should be the second provider
	letheProvider, exists := GetLetheProvider()
	if !exists {
		t.Error("Should have Lethe provider after overwrite")
	}

	// Test that it's the new provider
	result, err := letheProvider.CreateOptimizedSink("test.log")
	if err != nil {
		t.Errorf("CreateOptimizedSink failed: %v", err)
	}
	if result != "second" {
		t.Errorf("Expected 'second', got %v (provider was not overwritten)", result)
	}
}

// Mock types for testing

type mockRegistryWriter struct {
	filename string
	data     []byte
}

func (m *mockRegistryWriter) Write(data []byte) (int, error) {
	m.data = append(m.data, data...)
	return len(data), nil
}

func (m *mockRegistryWriter) WriteOwned(data []byte) (int, error) {
	return m.Write(data)
}

func (m *mockRegistryWriter) Sync() error {
	return nil
}

func (m *mockRegistryWriter) Close() error {
	return nil
}

func (m *mockRegistryWriter) GetOptimalBufferSize() int {
	return 4096
}

func (m *mockRegistryWriter) SupportsHotReload() bool {
	return true
}
