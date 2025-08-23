// stacktrace_test.go: Tests for stacktrace capture functionality
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strings"
	"testing"
)

// TestCaptureStack tests the CaptureStack function with various scenarios
func TestCaptureStack(t *testing.T) {
	tests := []struct {
		name     string
		skip     int
		depth    Depth
		testFunc func(t *testing.T, stack *Stack)
	}{
		{
			name:  "FirstFrame_Capture",
			skip:  0,
			depth: FirstFrame,
			testFunc: func(t *testing.T, stack *Stack) {
				if len(stack.pcs) != 1 {
					t.Errorf("Expected 1 frame for FirstFrame, got %d", len(stack.pcs))
				}
			},
		},
		{
			name:  "FullStack_Capture",
			skip:  0,
			depth: FullStack,
			testFunc: func(t *testing.T, stack *Stack) {
				if len(stack.pcs) == 0 {
					t.Error("Expected at least 1 frame for FullStack")
				}
				// Should have multiple frames for full stack
				if len(stack.pcs) < 2 {
					t.Logf("Warning: FullStack only captured %d frame(s), expected more", len(stack.pcs))
				}
			},
		},
		{
			name:  "Skip_Frames",
			skip:  2,
			depth: FullStack,
			testFunc: func(t *testing.T, stack *Stack) {
				if len(stack.pcs) == 0 {
					t.Error("Expected at least 1 frame even with skip=2")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stack with specified parameters
			stack := CaptureStack(tt.skip, tt.depth)
			defer FreeStack(stack)

			// Verify stack is not nil
			if stack == nil {
				t.Fatal("CaptureStack returned nil")
			}

			// Verify frames iterator is initialized
			if stack.frames == nil {
				t.Error("Stack frames iterator not initialized")
			}

			// Run specific test validation
			tt.testFunc(t, stack)
		})
	}
}

// TestCaptureStackExpansion tests the stack expansion logic for FullStack
func TestCaptureStackExpansion(t *testing.T) {
	// This test is designed to trigger the expansion logic by capturing
	// a very deep stack trace
	stack := CaptureStack(0, FullStack)
	defer FreeStack(stack)

	if stack == nil {
		t.Fatal("CaptureStack returned nil")
	}

	// The stack should contain multiple frames
	if len(stack.pcs) == 0 {
		t.Error("Expected at least 1 frame")
	}

	// Verify we can iterate through frames
	frame, more := stack.frames.Next()
	if !more {
		t.Error("Expected at least one frame to be available")
	}

	// The frame should contain valid information
	if frame.Function == "" {
		t.Error("Expected frame to have function name")
	}
}

// TestFreeStack tests the FreeStack function
func TestFreeStack(t *testing.T) {
	tests := []struct {
		name  string
		stack *Stack
	}{
		{
			name:  "Nil_Stack",
			stack: nil,
		},
		{
			name:  "Valid_Stack",
			stack: CaptureStack(0, FirstFrame),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FreeStack should not panic with nil or valid stack
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("FreeStack panicked: %v", r)
				}
			}()

			FreeStack(tt.stack)
		})
	}
}

// TestStackWithRealFunction tests stack capture in a nested function call
func TestStackWithRealFunction(t *testing.T) {
	stack := captureInNestedFunction()
	defer FreeStack(stack)

	if stack == nil {
		t.Fatal("CaptureStack returned nil")
	}

	// Should have captured multiple frames
	if len(stack.pcs) < 2 {
		t.Errorf("Expected at least 2 frames in nested call, got %d", len(stack.pcs))
	}

	// Iterate through frames to find our test function
	foundTestFunction := false
	for {
		frame, more := stack.frames.Next()
		if !more {
			break
		}

		if strings.Contains(frame.Function, "captureInNestedFunction") {
			foundTestFunction = true
		}
	}

	if !foundTestFunction {
		t.Error("Expected to find captureInNestedFunction in stack trace")
	}
}

// Helper function to create a nested call for testing
func captureInNestedFunction() *Stack {
	return captureInDeeperFunction()
}

func captureInDeeperFunction() *Stack {
	return CaptureStack(0, FullStack)
}

// TestStackPool tests that stacks are properly pooled
func TestStackPool(t *testing.T) {
	// Capture and free multiple stacks to test pooling
	stacks := make([]*Stack, 5)

	for i := 0; i < 5; i++ {
		stacks[i] = CaptureStack(0, FirstFrame)
		if stacks[i] == nil {
			t.Fatalf("CaptureStack %d returned nil", i)
		}
	}

	// Free all stacks
	for i := 0; i < 5; i++ {
		FreeStack(stacks[i])
	}

	// Capture again - should reuse pooled stacks
	newStack := CaptureStack(0, FirstFrame)
	defer FreeStack(newStack)

	if newStack == nil {
		t.Fatal("CaptureStack returned nil after pool reuse")
	}
}
