// errors.go: Error definitions for internal Zephyros Light
//
// Simplified error handling for IRIS internal Zephyros Light implementation.
// This provides essential error types without the full error management
// system available in commercial Zephyros.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package zephyroslite

import "errors"

// Core errors for Zephyros Light operations
var (
	// ErrMissingProcessor is returned when no processor function is provided
	ErrMissingProcessor = errors.New("processor function is required")

	// ErrInvalidCapacity is returned when ring capacity is invalid
	ErrInvalidCapacity = errors.New("capacity must be power of two and greater than zero")

	// ErrInvalidBatchSize is returned when batch size is invalid
	ErrInvalidBatchSize = errors.New("batch size must be positive and not exceed capacity")

	// ErrRingClosed is returned when operations are attempted on closed ring
	ErrRingClosed = errors.New("ring buffer is closed")
)
