package testutil

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func StartRedis(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx,
		"redis:7",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	endpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("failed to get redis url: %v", err)
	}

	cli := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	t.Cleanup(func() {
		_ = cli.Close()
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})
	return cli
}
