package lock

import (
	"context"
	"time"
)

// Locker grants leases over a named resource, each carrying a
// monotonically increasing fencing token. Implementations must never
// compute the token client-side
type Locker interface {
	// Acquire blocks until the lock is held or ctx is cancelled
	Acquire(ctx context.Context, resource string, ttl time.Duration) (*Lease, error)

	// TryAcquire is non-blocking: it returns ErrNotAcquired immediately if
	// the lock is currently held by someone else.
	TryAcquire(ctx context.Context, resource string, ttl time.Duration) (*Lease, error)
}

type Lease struct {
	Resource  string
	Token     int64
	ExpiresAt time.Time

	// releaseFn/refreshFn are supplied by the backend implementation that
	// created this Lease; callers only ever see the exported methods below.
	releaseFn func(ctx context.Context) error
	refreshFn func(ctx context.Context, ttl time.Duration) error
}

// NewLease is used by backend implementations to construct a Lease with its
// backend-specific release/refresh behavior attached
func NewLease(resource string, token int64, expiresAt time.Time,
	releaseFn func(ctx context.Context) error,
	refreshFn func(ctx context.Context, ttl time.Duration) error,
) *Lease {
	return &Lease{
		Resource:  resource,
		Token:     token,
		ExpiresAt: expiresAt,
		releaseFn: releaseFn,
		refreshFn: refreshFn,
	}
}

// Release gives up the lock early
func (l *Lease) Release(ctx context.Context) error {
	return l.releaseFn(ctx)
}

// Refresh renews the lease for another call (heartbeat)
func (l *Lease) Refresh(ctx context.Context, ttl time.Duration) error {
	if err := l.refreshFn(ctx, ttl); err != nil {
		return err
	}
	l.ExpiresAt = time.Now().Add(ttl)
	return nil
}

// Valid reports whether the lease has NOT yet expired locally. This is a
// client-side, best-effort convenience only. Never use
// this as the basis of deciding a write is safe; that's FencedResource's job.
func (l *Lease) Valid() bool {
	return time.Now().Before(l.ExpiresAt)
}
