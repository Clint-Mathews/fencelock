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
	"github.com/Clint-Mathews/fencelock/redislock"
)

// Both backends reject a stale write after A’s lease is gone.
// Single-instance Redis still fences this pause; replica-failover Redlock
// hazards are not reproduced here and must not be claimed as a CI failure.

func assertStaleWriteRejected(t *testing.T, leaseA, leaseB *lock.Lease) {
	t.Helper()
	resource := fencedstore.NewMemory()
	ctx := context.Background()
	if err := resource.Write(ctx, "partition-resource", leaseB.Token, []byte("B")); err != nil {
		t.Fatalf("client B write: %v", err)
	}
	err := resource.Write(ctx, "partition-resource", leaseA.Token, []byte("A-stale"))
	if !errors.Is(err, lock.ErrStaleToken) {
		t.Fatalf("expected stale write to be rejected, got %v", err)
	}
}

func TestPartitionScenario_Etcd(t *testing.T) {
	cliB := testutil.StartEtcd(t)
	stalledA := testutil.NewStallingEtcdClient(t, cliB.Endpoints())
	ctx := context.Background()

	leaseA, err := etcdlock.New(stalledA.Client).Acquire(ctx, "partition-resource", time.Second)
	if err != nil {
		t.Fatalf("client A acquire: %v", err)
	}
	stalledA.Stall()
	time.Sleep(3 * time.Second)

	leaseB, err := etcdlock.New(cliB).Acquire(ctx, "partition-resource", 5*time.Second)
	if err != nil {
		t.Fatalf("client B acquire: %v", err)
	}
	assertStaleWriteRejected(t, leaseA, leaseB)
}

func TestPartitionScenario_Redis(t *testing.T) {
	rdb := testutil.StartRedis(t)
	locker := redislock.New(rdb)
	ctx := context.Background()

	leaseA, err := locker.Acquire(ctx, "partition-resource", 200*time.Millisecond)
	if err != nil {
		t.Fatalf("client A acquire: %v", err)
	}
	time.Sleep(400 * time.Millisecond) // no KeepAlive; PX expiry is enough

	leaseB, err := locker.Acquire(ctx, "partition-resource", 5*time.Second)
	if err != nil {
		t.Fatalf("client B acquire: %v", err)
	}
	assertStaleWriteRejected(t, leaseA, leaseB)
}
