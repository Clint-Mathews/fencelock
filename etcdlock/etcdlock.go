package etcdlock

import (
	"context"
	"time"

	"github.com/Clint-Mathews/fencelock/lock"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Locker struct {
	client *clientv3.Client
}

func New(client *clientv3.Client) *Locker {
	return &Locker{
		client: client,
	}
}

func (l *Locker) acquire(ctx context.Context, resource string, ttl time.Duration, blocking bool) (*lock.Lease, error) {
	session, err := concurrency.NewSession(l.client, concurrency.WithTTL(int(ttl.Seconds())))
	if err != nil {
		return nil, err
	}
	mu := concurrency.NewMutex(session, "/fencelock/"+resource+"/")
	if blocking {
		if err := mu.Lock(ctx); err != nil {
			session.Close()
			return nil, err
		}
	} else {
		if err := mu.TryLock(ctx); err != nil {
			session.Close()
			if err == concurrency.ErrLocked {
				return nil, lock.ErrNotAcquired
			}
			return nil, err
		}
	}

	// mu.Key() is the etcd key the mutex created;
	// look up its CreateRevision to use as the fencing token.
	resp, err := l.client.Get(ctx, mu.Key())
	if err != nil || len(resp.Kvs) == 0 {
		mu.Unlock(ctx)
		session.Close()
		if err == nil {
			err = lock.ErrLeaseExpired
		}
		return nil, err
	}
	token := resp.Kvs[0].CreateRevision
	expiresAt := time.Now().Add(ttl)
	release := func(ctx context.Context) error {
		defer session.Close()
		return mu.Unlock(ctx)
	}
	refresh := func(ctx context.Context, newTTL time.Duration) error {
		// etcd session TTL is fixed at creation; "refresh" here means
		// keeping the session's keep alive going, which concurrency.Session
		// already does in the background via KeepAlive. Exposed for API
		// symmetry with the Locker interface and so a slow caller can
		// at least confirm the session is still alive
		select {
		case <-session.Done():
			return lock.ErrLeaseExpired
		default:
			return nil
		}
	}
	return lock.NewLease(resource, token, expiresAt, release, refresh), nil
}

func (l *Locker) Acquire(ctx context.Context, resource string, ttl time.Duration) (*lock.Lease, error) {
	return l.acquire(ctx, resource, ttl, true)
}

func (l *Locker) TryAcquire(ctx context.Context, resource string, ttl time.Duration) (*lock.Lease, error) {
	return l.acquire(ctx, resource, ttl, false)
}
