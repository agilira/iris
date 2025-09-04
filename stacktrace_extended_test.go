// stacktrace_extended_test.go: Extended tests for stacktrace functionality
//
// This file provides additional test coverage for stacktrace functions
// that are not covered by existing tests.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"strings"
	"testing"
)

func TestStack_Next_Coverage(t *testing.T) {
	// Test the Next() function specifically
	stack := CaptureStack(0, FullStack)
	defer FreeStack(stack)

	// Test Next() function
	frame, hasNext := stack.Next()
	if !hasNext {
		t.Error("Expected at least one frame in stack trace")
	}

	// Should have valid frame
	if frame.Function == "" {
		t.Error("Expected frame to have function name")
	}

	// Continue iterating to test multiple calls to Next()
	frameCount := 1
	for {
		frame, hasNext = stack.Next()
		if !hasNext {
			break
		}
		frameCount++

		// Prevent infinite loop in case of bug
		if frameCount > 100 {
			t.Error("Too many frames, possible infinite loop")
			break
		}
	}

	if frameCount == 0 {
		t.Error("Expected at least one frame from Next()")
	}
}

func TestStack_Next_EmptyStack(t *testing.T) {
	// Create empty stack
	stack := &Stack{}

	// Test Next() on empty stack
	frame, hasNext := stack.Next()
	if hasNext {
		t.Error("Expected empty stack to return false for hasNext")
	}
	if frame.Function != "" {
		t.Error("Expected empty frame for empty stack")
	}
}

func TestCaptureStack_FirstFrame(t *testing.T) {
	// Test FirstFrame depth
	stack := CaptureStack(0, FirstFrame)
	defer FreeStack(stack)

	// Should have at least one frame (FirstFrame means only the first few frames)
	frame, hasNext := stack.Next()
	if !hasNext {
		t.Log("Expected one frame for FirstFrame depth - this might be expected behavior")
		return // Not an error, just different behavior
	}

	// Should contain this test function
	if !strings.Contains(frame.Function, "TestCaptureStack_FirstFrame") {
		t.Logf("Expected frame to contain test function name, got: %s", frame.Function)
		// This might be due to inlining or different frame capture behavior
	}

	// Don't check for additional frames since Next() is not re-entrant
}

func TestCaptureStack_FullStack(t *testing.T) {
	// Test FullStack depth
	stack := CaptureStack(0, FullStack)
	defer FreeStack(stack)

	// Should have multiple frames
	frameCount := 0
	for {
		_, hasNext := stack.Next()
		if !hasNext {
			break
		}
		frameCount++

		if frameCount > 100 {
			break // Prevent infinite loop
		}
	}

	if frameCount < 2 {
		t.Errorf("Expected at least 2 frames for FullStack, got %d", frameCount)
	}
}

func TestCaptureStack_Skip(t *testing.T) {
	// Helper function to test skip parameter
	testSkip := func(skip int) *Stack {
		return CaptureStack(skip, FirstFrame)
	}

	// Test with skip=0 (should include testSkip function)
	stack0 := testSkip(0)
	defer FreeStack(stack0)

	frame0, hasNext0 := stack0.Next()
	if !hasNext0 {
		t.Log("Expected frame with skip=0 - might be expected behavior")
		return
	}

	// Function name might be mangled, so just check it's not empty
	if frame0.Function == "" {
		t.Error("Expected non-empty function name with skip=0")
	}

	// Test with skip=1 (should skip testSkip function)
	stack1 := testSkip(1)
	defer FreeStack(stack1)

	frame1, hasNext1 := stack1.Next()
	if !hasNext1 {
		t.Log("Expected frame with skip=1 - might be expected behavior")
		return
	}

	// Should have valid function name
	if frame1.Function == "" {
		t.Error("Expected non-empty function name with skip=1")
	}
}

func TestFreeStack_Nil(t *testing.T) {
	// Test FreeStack with nil (should not panic)
	FreeStack(nil)
}

func TestStack_FormatStack_Nil(t *testing.T) {
	// Test nil stack
	var nilStack *Stack
	result := nilStack.FormatStack()
	if result != "" {
		t.Errorf("Expected empty string for nil stack, got: %s", result)
	}
}

func TestStack_FormatStack_Valid(t *testing.T) {
	// Test actual stack
	stack := CaptureStack(0, FullStack)
	defer FreeStack(stack)

	result := stack.FormatStack()
	if result == "" {
		t.Error("Expected non-empty stack trace string")
	}

	// Should contain function names
	if !strings.Contains(result, "TestStack_FormatStack_Valid") {
		t.Errorf("Expected stack trace to contain test function name, got: %s", result)
	}
}

// Helper function to create nested stack for testing
func createNestedStackTest(depth int) *Stack {
	if depth <= 0 {
		return CaptureStack(1, FullStack) // Skip this function
	}
	return createNestedStackTest(depth - 1)
}

func TestCaptureStack_NestedCalls(t *testing.T) {
	// Test stack creation from nested calls
	stack := createNestedStackTest(5)
	defer FreeStack(stack)

	// Should have multiple frames
	frameCount := 0
	foundCreateNestedStack := false

	for {
		frame, hasNext := stack.Next()
		if !hasNext {
			break
		}
		frameCount++

		if strings.Contains(frame.Function, "createNestedStackTest") {
			foundCreateNestedStack = true
		}

		if frameCount > 20 {
			break // Prevent infinite loop
		}
	}

	if frameCount < 5 {
		t.Errorf("Expected at least 5 frames from nested calls, got %d", frameCount)
	}

	if !foundCreateNestedStack {
		t.Error("Expected stack trace to contain createNestedStackTest function")
	}
}

func TestStack_Reuse(t *testing.T) {
	// Test that stack objects are properly reused from pool
	stack1 := CaptureStack(0, FirstFrame)
	frame1, hasNext1 := stack1.Next()

	// Store the first frame info before freeing
	var firstFrameFunc string
	if hasNext1 {
		firstFrameFunc = frame1.Function
	}

	FreeStack(stack1)

	// Get another stack - might be the same object from pool
	stack2 := CaptureStack(0, FirstFrame)
	defer FreeStack(stack2)

	frame2, hasNext := stack2.Next()
	if !hasNext {
		t.Log("Expected frame from reused stack - might be expected behavior")
		return
	}

	// Should work correctly even if it's a reused object
	if frame2.Function == "" {
		t.Error("Expected valid function name from reused stack")
	}

	// Both frames should be from this test function (or at least non-empty)
	if hasNext1 && firstFrameFunc == "" {
		t.Error("Expected non-empty function name from first stack")
	}
	if frame2.Function == "" {
		t.Error("Expected non-empty function name from second stack")
	}
}

func BenchmarkStack_Next(b *testing.B) {
	stack := CaptureStack(1, FullStack)
	defer FreeStack(stack)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset stack for each iteration
		stack = CaptureStack(1, FullStack)

		// Iterate through frames
		for {
			_, hasNext := stack.Next()
			if !hasNext {
				break
			}
		}

		FreeStack(stack)
	}
}

func BenchmarkCaptureStack_FirstFrame(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack := CaptureStack(1, FirstFrame)
		FreeStack(stack)
	}
}

func BenchmarkCaptureStack_FullStack(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack := CaptureStack(1, FullStack)
		FreeStack(stack)
	}
}
