//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/redis"
	"github.com/vinaycharlie01/shroute/backend/test/containers"
)

func TestAdapter_PingAgainstRealRedis(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	addr := containers.Redis(ctx, t)

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
