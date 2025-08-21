// timecache.go: Ultra-fast time caching for IRIS MPSC logger
//
// Copyright (c) 2025 AGILira
// Series: an AGLIra fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync/atomic"
	"time"
)

// TimeCache provides cached time access to eliminate time.Now() allocations
// Optimized for MPSC logging where multiple goroutines need timestamps
type TimeCache struct {
	cachedTimeNano int64 // atomic int64 - current time in nanoseconds
	ticker         *time.Ticker
	stopCh         chan struct{}
}

// Global time cache instance for all IRIS loggers
var globalTimeCache = &TimeCache{}

func init() {
	// Initialize time cache
	globalTimeCache.cachedTimeNano = time.Now().UnixNano()
	globalTimeCache.ticker = time.NewTicker(500 * time.Microsecond) // 0.5ms precision for better accuracy
	globalTimeCache.stopCh = make(chan struct{})

	// Start background updater
	go globalTimeCache.updateLoop()
}

// updateLoop runs in background to update cached time
func (tc *TimeCache) updateLoop() {
	for {
		select {
		case <-tc.ticker.C:
			// Update cached time atomically - zero allocation
			atomic.StoreInt64(&tc.cachedTimeNano, time.Now().UnixNano())
		case <-tc.stopCh:
			tc.ticker.Stop()
			return
		}
	}
}

// CachedTimeNano returns cached time in nanoseconds (ZERO ALLOCATION)
func CachedTimeNano() int64 {
	return atomic.LoadInt64(&globalTimeCache.cachedTimeNano)
}

// CachedTime returns cached time as time.Time
func CachedTime() time.Time {
	nanos := atomic.LoadInt64(&globalTimeCache.cachedTimeNano)
	return time.Unix(0, nanos)
}

// StopTimeCache stops the global time cache (for testing)
func StopTimeCache() {
	close(globalTimeCache.stopCh)
}
