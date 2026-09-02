package redislock

import (
	"context"
	"time"

	"github.com/Clint-Mathews/fencelock/lock"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// go:embed scripts/release.lua
var releaseScript string

// Locker implements lock.Locker against a single Redis instance.
//
// This is intentionally the "weaker" backend.
// It does not solve the Redlock critique;
// it demonstrates fencing tokens on top of a best-effort lock.

type Locker struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Locker {
	return &Locker{
		rdb: rdb,
	}
}

func (l *Locker) acquire(ctx context.Context, resource string, ttl time.Duration, blocking bool) (*lock.Lease, error) {
	lockKey := "fencelock:lock:" + resource
	tokenKey := "fencelock:token:" + resource
	owner := uuid.NewString()

	for {
		ok, err := l.rdb.SetNX(ctx, lockKey, owner, ttl).Result()
		if err != nil {
			return nil, err
		}
		if ok {
			break
		}
		if !blocking {
			return nil, lock.ErrNotAcquired
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
			continue
		}
	}

	// Token issuance: atomic INCR on a counter key that outlives any single
	// lock instance, so tokens keep increasing across acquire/release
	// cycles - this is the part naive Redis locks skip.
	token, err := l.rdb.Incr(ctx, tokenKey).Result()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(ttl)

	release := func(ctx context.Context) error {
		res, err := l.rdb.Eval(ctx, releaseScript, []string{lockKey}, owner).Result()
		if err != nil {
			return err
		}
		if n, _ := res.(int64); n == 0 {
			return lock.ErrLeaseExpired // we no longer owned the key
		}
		return nil
	}

	refresh := func(ctx context.Context, newTTL time.Duration) error {
		ok, err := l.rdb.Expire(ctx, lockKey, newTTL).Result()
		if err != nil {
			return err
		}
		if !ok {
			return lock.ErrLeaseExpired
		}
		return nil
	}

	return lock.NewLease(resource, token, expiresAt, release, refresh), nil
}

func (l *Locker) Acquire(ctx context.Context, resource string, ttl time.Duration) (*lock.Lease, error) {
	return l.acquire(ctx, resource, ttl, true)
}

func (l *Locker) TryAcquire(ctx context.Context, resource string, ttl time.Duration) (*lock.Lease, error) {
	return l.acquire(ctx, resource, ttl, false)
}
