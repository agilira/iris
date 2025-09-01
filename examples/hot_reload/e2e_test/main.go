// run_e2e_test.go: Runner for end-to-end hot reload tests
//
// This executes the comprehensive test suite to verify Iris hot reload
// functionality beyond any reasonable doubt.
//
// Copyright (c) 2025 AGILira
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
		os.Chdir(originalDir)
	}()
	
	runE2ETests()
}
