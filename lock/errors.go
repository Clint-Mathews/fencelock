package lock

import "errors"

var (
	// ErrNotAcquired is returned by TryAcquire when the lock is already held.
	ErrNotAcquired = errors.New("lock: not acquired")

	// ErrLeaseExpired is returned by Release/Refresh when the lease has
	// already expired server-side (session/lease lost).
	ErrLeaseExpired = errors.New("lock: lease expired")

	// ErrStaleToken is returned by FencedResource.Write when the supplied
	// token is not greater than or equal to the highest token already seen.
	ErrStaleToken = errors.New("lock: stale fencing token")
)
