// detection_test.go: Tests for Lethe detection capabilities
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package lethe

import (
	"os"
	"testing"
)

func TestDetectLetheCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		writer   interface{}
		expected bool
	}{
		{
			name:     "nil writer",
			writer:   nil,
			expected: false,
		},
		{
			name:     "standard file",
			writer:   &os.File{},
			expected: false,
		},
		{
			name:     "mock Lethe writer",
			writer:   &mockDetectionLetheWriter{},
			expected: true,
		},
		{
			name:     "non-writer interface",
			writer:   "string",
			expected: false,
		},
		{
			name:     "struct without methods",
			writer:   struct{}{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLetheCapabilities(tt.writer)

			if tt.expected && result == nil {
				t.Errorf("Expected to detect Lethe capabilities, but got nil")
			}

			if !tt.expected && result != nil {
				t.Errorf("Expected no Lethe capabilities, but got: %+v", result)
			}
		})
	}
}

func TestIsLetheWriter(t *testing.T) {
	tests := []struct {
		name     string
		writer   interface{}
		expected bool
	}{
		{
			name:     "nil writer",
			writer:   nil,
			expected: false,
		},
		{
			name:     "standard file",
			writer:   &os.File{},
			expected: false,
		},
		{
			name:     "mock Lethe writer",
			writer:   &mockDetectionLetheWriter{},
			expected: true,
		},
		{
			name:     "partial interface implementation",
			writer:   &partialWriter{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLetheWriter(tt.writer)

			if result != tt.expected {
				t.Errorf("IsLetheWriter() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestLetheWriterInterface(t *testing.T) {
	// Test that our mock properly implements the interface
	var _ LetheWriter = &mockDetectionLetheWriter{}

	writer := &mockDetectionLetheWriter{}

	// Test all interface methods
	data := []byte("test data")

	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned %d bytes, expected %d", n, len(data))
	}

	n, err = writer.WriteOwned(data)
	if err != nil {
		t.Errorf("WriteOwned failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("WriteOwned returned %d bytes, expected %d", n, len(data))
	}

	if err := writer.Sync(); err != nil {
		t.Errorf("Sync failed: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	bufferSize := writer.GetOptimalBufferSize()
	if bufferSize <= 0 {
		t.Errorf("GetOptimalBufferSize returned %d, expected positive value", bufferSize)
	}

	hotReload := writer.SupportsHotReload()
	if !hotReload {
		t.Error("SupportsHotReload should return true for mock writer")
	}
}

func TestDetectionWithInterfaceUpgrading(t *testing.T) {
	// Test detection when a writer can be upgraded to LetheWriter
	basicWriter := &basicWriter{}

	// Should not be detected as LetheWriter
	if IsLetheWriter(basicWriter) {
		t.Error("Basic writer should not be detected as LetheWriter")
	}

	// Upgrade to LetheWriter
	upgradedWriter := &upgradedWriter{basicWriter}

	// Should now be detected
	if !IsLetheWriter(upgradedWriter) {
		t.Error("Upgraded writer should be detected as LetheWriter")
	}

	detected := DetectLetheCapabilities(upgradedWriter)
	if detected == nil {
		t.Error("Failed to detect capabilities in upgraded writer")
	}
}

// Test mocks and helper types

type mockDetectionLetheWriter struct {
	data []byte
}

func (m *mockDetectionLetheWriter) Write(data []byte) (int, error) {
	m.data = append(m.data, data...)
	return len(data), nil
}

func (m *mockDetectionLetheWriter) WriteOwned(data []byte) (int, error) {
	return m.Write(data)
}

func (m *mockDetectionLetheWriter) Sync() error {
	return nil
}

func (m *mockDetectionLetheWriter) Close() error {
	return nil
}

func (m *mockDetectionLetheWriter) GetOptimalBufferSize() int {
	return 8192
}

func (m *mockDetectionLetheWriter) SupportsHotReload() bool {
	return true
}

// partialWriter implements only some methods (should not be detected)
type partialWriter struct{}

func (p *partialWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (p *partialWriter) Sync() error {
	return nil
}

// basicWriter implements basic io.Writer
type basicWriter struct{}

func (b *basicWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

// upgradedWriter wraps basicWriter and adds Lethe capabilities
type upgradedWriter struct {
	*basicWriter
}

func (u *upgradedWriter) WriteOwned(data []byte) (int, error) {
	return u.Write(data)
}

func (u *upgradedWriter) Sync() error {
	return nil
}

func (u *upgradedWriter) Close() error {
	return nil
}

func (u *upgradedWriter) GetOptimalBufferSize() int {
	return 16384
}

func (u *upgradedWriter) SupportsHotReload() bool {
	return true
}
