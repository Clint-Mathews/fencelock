package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Clint-Mathews/fencelock/etcdlock"
	"github.com/Clint-Mathews/fencelock/fencedstore"
	"github.com/Clint-Mathews/fencelock/internal/testutil"
	"github.com/Clint-Mathews/fencelock/lock"
)

// TestFencingSavesIt is REQUIREMENTS.md Test B: the same pause/expire
// scenario as TestNaiveLockFails, with resource-side token checks.
func TestFencingSavesIt(t *testing.T) {
	cliB := testutil.StartEtcd(t)
	stalledA := testutil.NewStallingEtcdClient(t, cliB.Endpoints())
	lockerA := etcdlock.New(stalledA.Client)
	lockerB := etcdlock.New(cliB)
	resource := fencedstore.NewMemory()
	ctx := context.Background()

	leaseA, err := lockerA.Acquire(ctx, "resource-y", time.Second)
	if err != nil {
		t.Fatalf("client A acquire: %v", err)
	}
	if err := resource.Write(ctx, "resource-y", leaseA.Token, []byte("written by A")); err != nil {
		t.Fatalf("client A initial write should succeed: %v", err)
	}

	stalledA.Stall()
	time.Sleep(3 * time.Second)

	leaseB, err := lockerB.Acquire(ctx, "resource-y", 5*time.Second)
	if err != nil {
		t.Fatalf("client B acquire: %v", err)
	}
	if leaseB.Token <= leaseA.Token {
		t.Fatalf("expected B's token > A's token, got A=%d B=%d", leaseA.Token, leaseB.Token)
	}
	if err := resource.Write(ctx, "resource-y", leaseB.Token, []byte("written by B")); err != nil {
		t.Fatalf("client B write should succeed: %v", err)
	}

	err = resource.Write(ctx, "resource-y", leaseA.Token, []byte("written by A (stale!)"))
	if !errors.Is(err, lock.ErrStaleToken) {
		t.Fatalf("expected ErrStaleToken, got %v", err)
	}
	if got := string(resource.Get("resource-y")); got != "written by B" {
		t.Fatalf("expected resource to hold B's write, got %q", got)
	}
	t.Log("Fencing correctly rejected stale write")
}
