package testutil

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// StallingEtcdClient is Kleppman's paused process: it still has a Lease
// in memory, but KeepAlive cannot reach etcd after Stall()
type StallingEtcdClient struct {
	Client  *clientv3.Client
	mu      sync.Mutex
	conns   []net.Conn
	stalled bool
}

func NewStallingEtcdClient(t *testing.T, endpoints []string) *StallingEtcdClient {
	t.Helper()
	s := &StallingEtcdClient{}
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.stalled {
			return nil, errors.New("network stalled")
		}
		var d net.Dialer
		conn, err := d.DialContext(ctx, "tcp", addr)

		if err == nil {
			s.conns = append(s.conns, conn)
		}
		return conn, err
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithContextDialer(dialer),
		},
	})
	if err != nil {
		t.Fatalf("stalling etcd client: %v", err)
	}
	t.Cleanup(func() { _ = cli.Close() })
	s.Client = cli
	return s
}

func (s *StallingEtcdClient) Stall() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stalled = true
	for _, c := range s.conns {
		_ = c.Close()
	}
	s.conns = nil
}
