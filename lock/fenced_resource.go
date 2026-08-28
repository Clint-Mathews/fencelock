package lock

import "context"

// FencedResource is anything that must enforce write ordering using
// fencing tokens instead of relying on the lock service for exclusivity.
type FencedResource interface {
	// Write only succeeds if token >= the highest token this resource has
	// already accepted for the given key. A stale token must be rejected
	// with ErrStaleToken and must not mutate state.
	Write(ctx context.Context, key string, token int64, data []byte) error
}
