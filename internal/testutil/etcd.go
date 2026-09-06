package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	etcd "github.com/testcontainers/testcontainers-go/modules/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func StartEtcd(t *testing.T) *clientv3.Client {
	t.Helper()
	ctx := context.Background()

	etcdContainer, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14")
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	endpoint, err := etcdContainer.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("etcd endpoint: %v", err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("etcd client: %v", err)
	}
	t.Cleanup(func() {
		_ = cli.Close()
		if err := testcontainers.TerminateContainer(etcdContainer); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})
	return cli
}
