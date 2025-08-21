// multiwriter_proper_race_test.go: Fixed race condition tests for MultiWriter
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

// TestMultiWriterProperRaceTest tests for race conditions with NO shared buffers
func TestMultiWriterProperRaceTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race test in short mode")
	}

	const (
		numWorkers    = 50
		numOperations = 500
		testDuration  = 3 * time.Second
	)

	mw := NewMultiWriter()

	// NO initial shared writers - each goroutine will create its own

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Start writer-only workers - each with their OWN buffer
	for i := 0; i < numWorkers/4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each worker gets its OWN buffer that NO other goroutine touches
			myBuffer := WrapWriter(&bytes.Buffer{})
			mw.AddWriter(myBuffer)
			defer mw.RemoveWriter(myBuffer) // Clean up our own buffer

			data := []byte("worker exclusive buffer test data")

			for j := 0; j < numOperations; j++ {
				select {
				case <-stop:
					return
				default:
					// Write to ALL writers (including other workers' buffers)
					// This is safe because each buffer is only "owned" by one goroutine
					mw.Write(data)
					mw.Sync()
				}
			}
		}(i)
	}

	// Start add/remove workers - each creates temporary buffers
	for i := 0; i < numWorkers/4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations/10; j++ {
				select {
				case <-stop:
					return
				default:
					// Create temporary buffer just for this operation
					tempBuf := WrapWriter(&bytes.Buffer{})
					mw.AddWriter(tempBuf)

					// Write to all writers (including the temporary one)
					mw.Write([]byte("temp buffer test"))

					// Remove the temporary buffer
					mw.RemoveWriter(tempBuf)
				}
			}
		}(i)
	}

	// Start reader workers - just get counts and lists
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
			}
		}(i)
	}

	// Mixed workers - random operations with own buffers
	for i := 0; i < numWorkers/4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each mixed worker also gets its own permanent buffer
			myBuffer := WrapWriter(&bytes.Buffer{})
			mw.AddWriter(myBuffer)
			defer mw.RemoveWriter(myBuffer)

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
						// Create a temporary buffer
						tempBuf := WrapWriter(&bytes.Buffer{})
						mw.AddWriter(tempBuf)
						mw.RemoveWriter(tempBuf) // Immediately remove
					case 3:
						_ = mw.Count() // Just read the count
					}
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

// TestMultiWriterAtomicOperationsRace tests atomic pointer operations specifically
func TestMultiWriterAtomicOperationsRace(t *testing.T) {
	const iterations = 5000

	for iteration := 0; iteration < 50; iteration++ {
		mw := NewMultiWriter()

		var wg sync.WaitGroup

		// Writer goroutine - constantly writing with NO shared buffers
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Create our own buffer that nobody else touches
			myBuffer := WrapWriter(&bytes.Buffer{})
			mw.AddWriter(myBuffer)
			defer mw.RemoveWriter(myBuffer)

			data := []byte("atomic pointer race test")

			for i := 0; i < iterations; i++ {
				// This calls getWriters() which does atomic.LoadPointer
				mw.Write(data)
			}
		}()

		// Modifier goroutine - constantly changing writers list
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < iterations/10; i++ {
				// Each add/remove creates a NEW buffer that's not shared
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

// TestMultiWriterHighContentionRace tests high contention scenarios
func TestMultiWriterHighContentionRace(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high contention race test in short mode")
	}

	mw := NewMultiWriter()

	const numGoroutines = 100
	const numIterations = 1000

	var wg sync.WaitGroup

	// Each goroutine performs the same mix of operations but with its own resources
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine gets its own buffer
			myBuffer := WrapWriter(&bytes.Buffer{})
			mw.AddWriter(myBuffer)
			defer mw.RemoveWriter(myBuffer)

			for j := 0; j < numIterations; j++ {
				switch j % 5 {
				case 0:
					// Write to all buffers (safe because each buffer is owned by one goroutine)
					mw.Write([]byte("high contention test"))
				case 1:
					// Sync all writers
					mw.Sync()
				case 2:
					// Add temporary writer
					tempBuf := WrapWriter(&bytes.Buffer{})
					mw.AddWriter(tempBuf)
					mw.RemoveWriter(tempBuf)
				case 3:
					// Read operations
					_ = mw.Count()
				case 4:
					// Get writers list
					_ = mw.Writers()
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify still functional
	if mw.Count() < 0 {
		t.Error("MultiWriter in invalid state after high contention test")
	}
}
