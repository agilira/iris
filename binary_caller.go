// binary_caller.go: Lazy caller computation for binary logging
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"runtime"
	"sync"
)

// LazyCaller defers caller computation until needed (like Zap)
type LazyCaller struct {
	skip     int
	computed bool
	mu       sync.RWMutex
	file     string
	line     int
	function string
}

// NewLazyCaller creates a lazy caller that computes on demand
func NewLazyCaller(skip int) *LazyCaller {
	return &LazyCaller{
		skip: skip,
	}
}

// File returns the file (computed once)
func (lc *LazyCaller) File() string {
	lc.compute()
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.file
}

// Line returns the line number (computed once)
func (lc *LazyCaller) Line() int {
	lc.compute()
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.line
}

// Function returns the function name (computed once)
func (lc *LazyCaller) Function() string {
	lc.compute()
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.function
}

// compute does the expensive runtime.Caller once and caches
func (lc *LazyCaller) compute() {
	lc.mu.RLock()
	if lc.computed {
		lc.mu.RUnlock()
		return
	}
	lc.mu.RUnlock()

	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Double-check pattern
	if lc.computed {
		return
	}

	pc, file, line, ok := runtime.Caller(lc.skip)
	if ok {
		lc.file = file
		lc.line = line

		if fn := runtime.FuncForPC(pc); fn != nil {
			lc.function = fn.Name()
		}
	}

	lc.computed = true
}

// LazyCallerPool manages lazy caller reuse
type LazyCallerPool struct {
	pool sync.Pool
}

// NewLazyCallerPool creates a pool of lazy callers
func NewLazyCallerPool() *LazyCallerPool {
	return &LazyCallerPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &LazyCaller{}
			},
		},
	}
}

// GetLazyCaller gets a lazy caller from pool
func (lcp *LazyCallerPool) GetLazyCaller(skip int) *LazyCaller {
	caller := lcp.pool.Get().(*LazyCaller)
	caller.skip = skip
	caller.computed = false
	caller.file = ""
	caller.line = 0
	caller.function = ""
	return caller
}

// ReleaseLazyCaller returns lazy caller to pool
func (lcp *LazyCallerPool) ReleaseLazyCaller(caller *LazyCaller) {
	lcp.pool.Put(caller)
}
