package integration

import (
	"context"
	"testing"
	"time"

	"github.com/Clint-Mathews/fencelock/etcdlock"
	"github.com/Clint-Mathews/fencelock/internal/testutil"
)

// TestNaiveLockFails demonstrates the hazard: a client's lease expires
// while it's "paused", a second client acquires and writes, then the first
// client wakes up and writes anyway - with no fencing check, this corrupts
// state. This test intentionally does NOT use FencedResourec, to show the baseline problem.
func TestNaiveLockFails(t *testing.T) {
	cliB := testutil.StartEtcd(t)
	stalledA := testutil.NewStallingEtcdClient(t, cliB.Endpoints())
	lockerA := etcdlock.New(stalledA.Client)
	lockerB := etcdlock.New(cliB)
	ctx := context.Background()

	var sharedState string

	leaseA, err := lockerA.Acquire(ctx, "resource-x", 1*time.Second)
	if err != nil {
		t.Fatalf("client A acquire: %v", err)
	}

	stalledA.Stall()
	time.Sleep(3 * time.Second)

	leaseB, err := lockerB.Acquire(ctx, "resource-x", 5*time.Second)
	if err != nil {
		t.Fatalf("client B acquire: %v", err)
	}

	sharedState = "written by B"
	if leaseB.Token <= leaseA.Token {
		t.Fatalf("expected B's token > A's token, got A=%d B=%d", leaseA.Token, leaseB.Token)
	}

	sharedState = "written by A (stale!)"
	if sharedState != "written by A (stale!)" {
		t.Fatal("sanity check failed")
	}
	t.Log("BUG REPRODUCED: stale client A overwrote client B's write, unguarded")
}
