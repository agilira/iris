// run_e2e_test.go: Runner for end-to-end hot reload tests
//
// This executes the comprehensive test suite to verify Iris hot reload
// functionality beyond any reasonable doubt.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("🧪 Starting Iris Hot Reload E2E Test Suite")
	fmt.Println("===========================================")

	// Change to test directory to avoid conflicts
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			fmt.Printf("Warning: failed to change back to original directory: %v\n", err)
		}
	}()

	runE2ETests()
}
