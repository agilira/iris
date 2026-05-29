// autoscaling_idlefactory_test.go: per-ring idle strategy isolation
//
// Pins that Config.IdleStrategyFactory yields a DISTINCT idle-strategy instance
// per ring. The AdaptiveLogger runs two rings (singleLogger always, multiLogger
// after scale-up); a single shared stateful strategy — notably a channel-based
// one whose wakeup is a buffer-of-one — lets one ring's consumer steal the
// wakeup meant for the other, stranding records. A per-ring factory removes the
// sharing.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"sync"
	"testing"

	"github.com/agilira/iris/internal/zephyroslite"
)

func TestAdaptiveLogger_IdleStrategyFactory_DistinctPerRing(t *testing.T) {
	var mu sync.Mutex
	produced := make([]zephyroslite.IdleStrategy, 0, 2)

	buf := &bufferedSyncer{}
	cfg := Config{
		Level:   Debug,
		Output:  buf,
		Encoder: NewJSONEncoder(),
		IdleStrategyFactory: func() zephyroslite.IdleStrategy {
			s := zephyroslite.NewSleepingIdleStrategy(0, 0)
			mu.Lock()
			produced = append(produced, s)
			mu.Unlock()
			return s
		},
	}

	sc := DefaultScalerConfig(cfg)
	sc.GoroutineThreshold = 2

	al, err := NewAdaptiveLogger(sc)
	if err != nil {
		t.Fatalf("NewAdaptiveLogger failed: %v", err)
	}
	defer safeCloseAdaptiveLogger(t, al)
	if err := al.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Force a deterministic scale-up so the multiLogger ring is constructed.
	holding := make(chan struct{})
	release := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = al.getLogger()
		close(holding)
		<-release
		al.releaseWriter()
	}()
	<-holding
	_ = al.getLogger() // activeWriters=2 → scale up → multi ring built
	al.releaseWriter()
	close(release)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(produced) != 2 {
		t.Fatalf("factory invoked %d times, want 2 (one per ring: single + multi)", len(produced))
	}
	if produced[0] == produced[1] {
		t.Fatal("single and multi ring share the SAME idle-strategy instance: " +
			"factory result was not isolated per ring")
	}
}
