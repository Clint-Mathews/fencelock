package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Clint-Mathews/fencelock/etcdlock"
	"github.com/Clint-Mathews/fencelock/internal/testutil"
)

// TestConcurrencyStress races N goroutine against the same resource and
// assert: no two ever hold overlapping leases, and tokens strictly
// increase in order leases were granted. Run with -race.

func TestConcurrencyStress(t *testing.T) {
	cli := testutil.StartEtcd(t)
	locker := etcdlock.New(cli)
	ctx := context.Background()

	const goroutines = 20
	const roundsEach = 5

	var mu sync.Mutex
	var tokens []int64
	var holders int32 // must never exceed 1 while incremented under lock

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < roundsEach; r++ {
				lease, err := locker.Acquire(ctx, "stress-resource", 2*time.Second)
				if err != nil {
					t.Errorf("acquire: %v", err)
					return
				}
				mu.Lock()
				holders++
				if holders > 1 {
					mu.Unlock()
					t.Errorf("obverlapping lease holders detected: %d", holders)
					return
				}
				tokens = append(tokens, lease.Token)
				mu.Unlock()
				time.Sleep(5 * time.Millisecond) // hold briefly to widen the race window

				mu.Lock()
				holders--
				mu.Unlock()

				if err := lease.Release(ctx); err != nil {
					t.Errorf("release: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()

	for i := 1; i < len(tokens); i++ {
		if tokens[i] <= tokens[i-1] {
			t.Fatalf("tokens not strictly increasing at index %d: %v", i, tokens)
		}
	}
	t.Log("Concurrency stress test passed")
}
