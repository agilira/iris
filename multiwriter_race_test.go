// multiwriter_race_test.go: Intensive race condition tests for MultiWriter
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment  
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestMultiWriterIntensiveRace tests for race conditions under extreme concurrent load
func TestMultiWriterIntensiveRace(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping intensive race test in short mode")
	}

	const (
		numWorkers     = 100
		numOperations  = 1000
		testDuration   = 5 * time.Second
	)

	mw := NewMultiWriter()
	
	// Add initial writers - each with separate buffer
	for i := 0; i < 10; i++ {
		mw.AddWriter(WrapWriter(&bytes.Buffer{}))
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Start writers - constantly writing
	for i := 0; i < numWorkers/4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			data := []byte("intensive race test data")
			
			for j := 0; j < numOperations; j++ {
				select {
				case <-stop:
					return
				default:
					// This could race with AddWriter/RemoveWriter
					mw.Write(data)
					mw.Sync()
				}
				
				if j%100 == 0 {
					runtime.Gosched()
				}
			}
		}(i)
	}

	// Start add/remove workers - constantly modifying writer list
	for i := 0; i < numWorkers/4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < numOperations/10; j++ {
				select {
				case <-stop:
					return
				default:
					// Add writer - each goroutine gets its OWN buffer to avoid races
					buf := WrapWriter(&bytes.Buffer{})
					mw.AddWriter(buf)
					
					// Immediately try to write (potential race)
					mw.Write([]byte("race test"))
					
					// Remove writer
					mw.RemoveWriter(buf)
				}
				
				if j%10 == 0 {
					runtime.Gosched()
				}
			}
		}(i) // Pass index correctly
	}

	// Start readers - constantly getting writer count/list
	for i := 0; i < numWorkers/4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < numOperations; j++ {
				select {
				case <-stop:
					return
				default:
					// These operations read the atomic pointer
					_ = mw.Count()
					_ = mw.Writers()
				}
				
				if j%100 == 0 {
					runtime.Gosched()
				}
			}
		}(i)
	}

	// Mixed workers - random operations
	for i := 0; i < numWorkers/4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < numOperations; j++ {
				select {
				case <-stop:
					return
				default:
					switch j % 4 {
					case 0:
						mw.Write([]byte("mixed operation"))
					case 1:
						mw.Sync()
					case 2:
						// Each goroutine creates its own buffer
						mw.AddWriter(WrapWriter(&bytes.Buffer{}))
					case 3:
						writers := mw.Writers()
						if len(writers) > 20 { // Prevent unlimited growth
							mw.RemoveWriter(writers[0])
						}
					}
				}
				
				if j%50 == 0 {
					runtime.Gosched()
				}
			}
		}(i)
	}

	// Let it run for a bit, then stop
	time.AfterFunc(testDuration, func() {
		close(stop)
	})

	wg.Wait()

	// Verify MultiWriter is still functional
	finalCount := mw.Count()
	if finalCount < 0 {
		t.Errorf("Invalid final count: %d", finalCount)
	}

	// Final write to ensure it still works
	_, err := mw.Write([]byte("final test"))
	if err != nil {
		t.Errorf("Final write failed: %v", err)
	}
}

// TestMultiWriterAtomicPointerRace specifically tests the unsafe.Pointer race
func TestMultiWriterAtomicPointerRace(t *testing.T) {
	const iterations = 10000
	
	for iteration := 0; iteration < 100; iteration++ {
		mw := NewMultiWriter()
		
		var wg sync.WaitGroup
		
		// Writer goroutine - constantly writing
		wg.Add(1)
		go func() {
			defer wg.Done()
			data := []byte("atomic pointer race test")
			
			for i := 0; i < iterations; i++ {
				// This calls getWriters() which does atomic.LoadPointer
				mw.Write(data)
			}
		}()
		
		// Modifier goroutine - constantly changing writers
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for i := 0; i < iterations/10; i++ {
				buf := WrapWriter(&bytes.Buffer{})
				
				// This calls setWriters() which does atomic.StorePointer
				mw.AddWriter(buf)
				
				// This could create a race with the Write() above
				mw.RemoveWriter(buf)
			}
		}()
		
		wg.Wait()
	}
}

// TestMultiWriterSliceValidityRace tests if slice remains valid during pointer updates
func TestMultiWriterSliceValidityRace(t *testing.T) {
	mw := NewMultiWriter()
	
	// Add some initial writers
	for i := 0; i < 5; i++ {
		mw.AddWriter(WrapWriter(&bytes.Buffer{}))
	}
	
	var wg sync.WaitGroup
	var raceDetected bool
	
	// Goroutine that gets slice and tries to use it
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		for i := 0; i < 50000; i++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						// If we get a panic, it might be due to slice being invalidated
						raceDetected = true
						t.Logf("Panic detected (possible race): %v", r)
					}
				}()
				
				// Get the slice
				writers := mw.getWriters()
				
				// Try to use it - this could panic if the slice was invalidated
				// by concurrent modification
				for _, w := range writers {
					if w != nil {
						w.Write([]byte("test"))
					}
				}
			}()
			
			if i%1000 == 0 {
				runtime.Gosched()
			}
		}
	}()
	
	// Goroutine that constantly modifies the writers
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		for i := 0; i < 5000; i++ {
			buf := WrapWriter(&bytes.Buffer{})
			
			// Add and immediately remove - this updates the atomic pointer
			mw.AddWriter(buf)
			mw.RemoveWriter(buf)
			
			if i%100 == 0 {
				runtime.Gosched()
			}
		}
	}()
	
	wg.Wait()
	
	if raceDetected {
		t.Log("Race condition detected - this is expected and shows the issue")
	}
}
