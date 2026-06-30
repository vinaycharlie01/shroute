//go:build integration

package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/redis"
)

func TestAdapter_PingAgainstRealRedis(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("start redis container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("terminate redis container: %v", err)
		}
	})

	addr, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("endpoint: %v", err)
	}

	adapter, err := redis.New(ctx, addr, "", 0)
	if err != nil {
		t.Fatalf("redis.New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := adapter.Close(); err != nil {
			t.Logf("close adapter: %v", err)
		}
	})

	if got := adapter.Name(); got != "redis" {
		t.Errorf("Name() = %q, want %q", got, "redis")
	}
	if err := adapter.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}
