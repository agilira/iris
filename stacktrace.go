// stacktrace.go: Ultra-fast stacktrace capture for IRIS logger
//
// This implementation follows Zap's approach for high-performance stack trace
// capture using runtime.Callers and object pooling for zero-allocation paths.
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra library
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"runtime"
	"strconv"
	"sync"

	"github.com/agilira/iris/internal/bufferpool"
)

// stackPool pools Stack objects for reuse to reduce allocations
var stackPool = sync.Pool{
	New: func() any {
		return &Stack{
			storage: make([]uintptr, 64), // Pre-allocate for typical stack depth
		}
	},
}

// Stack represents a captured stack trace with program counters
type Stack struct {
	pcs     []uintptr       // Program counters; slice of storage
	frames  *runtime.Frames // Frame iterator for parsing
	storage []uintptr       // Backing storage for pool reuse
}

// Depth specifies how deep of a stack trace should be captured
type Depth int

const (
	// FirstFrame captures only the first frame (caller info)
	FirstFrame Depth = iota
	// FullStack captures the entire call stack
	FullStack
)

// CaptureStack captures a stack trace of the specified depth, skipping frames.
// skip=0 identifies the caller of CaptureStack.
// The caller must call FreeStack on the returned stack after using it.
func CaptureStack(skip int, depth Depth) *Stack {
	stack := stackPool.Get().(*Stack)

	switch depth {
	case FirstFrame:
		stack.pcs = stack.storage[:1]
	case FullStack:
		stack.pcs = stack.storage
	}

	// +2 to skip CaptureStack and runtime.Callers
	numFrames := runtime.Callers(skip+2, stack.pcs)

	// For full stack trace, expand storage if needed
	if depth == FullStack {
		pcs := stack.pcs
		for numFrames == len(pcs) {
			// Double the size until we capture all frames
			pcs = make([]uintptr, len(pcs)*2)
			numFrames = runtime.Callers(skip+2, pcs)
		}

		// Update storage if we expanded
		if len(pcs) > len(stack.storage) {
			stack.storage = pcs
		}
		stack.pcs = pcs[:numFrames]
	} else {
		stack.pcs = stack.pcs[:numFrames]
	}

	// Initialize frames iterator
	stack.frames = runtime.CallersFrames(stack.pcs)
	return stack
}

// FreeStack returns a Stack to the pool for reuse
func FreeStack(stack *Stack) {
	if stack == nil {
		return
	}

	// Reset for reuse
	stack.pcs = nil
	stack.frames = nil

	// Return to pool
	stackPool.Put(stack)
}

// Next returns the next frame in the stack trace
func (s *Stack) Next() (runtime.Frame, bool) {
	if s.frames == nil {
		return runtime.Frame{}, false
	}
	return s.frames.Next()
}

// FormatStack formats the entire stack trace into a string using buffer pooling
func (s *Stack) FormatStack() string {
	if s == nil {
		return ""
	}

	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	// Reset frames iterator
	s.frames = runtime.CallersFrames(s.pcs)

	nonEmpty := false
	// Format each frame, skipping the final runtime frame
	for frame, more := s.frames.Next(); more; frame, more = s.frames.Next() {
		if nonEmpty {
			buf.WriteByte('\n')
		}
		nonEmpty = true

		buf.WriteString(frame.Function)
		buf.WriteByte('\n')
		buf.WriteByte('\t')
		buf.WriteString(frame.File)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(frame.Line))
	}

	return buf.String()
}

// fastStacktrace is a high-performance replacement for debug.Stack()
// It uses the same approach as Zap for optimal performance
func fastStacktrace(skip int) string {
	stack := CaptureStack(skip+1, FullStack) // +1 to skip fastStacktrace itself
	defer FreeStack(stack)
	return stack.FormatStack()
}
