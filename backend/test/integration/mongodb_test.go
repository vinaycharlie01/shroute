//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/mongodb"
	"github.com/vinaycharlie01/shroute/backend/test/containers"
)

func TestAdapter_PingAgainstRealMongoDB(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	uri := containers.MongoDB(ctx, t)

	adapter, err := mongodb.New(ctx, uri)
	if err != nil {
		t.Fatalf("mongodb.New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := adapter.Close(); err != nil {
			t.Logf("close adapter: %v", err)
		}
	})

	if got := adapter.Name(); got != "mongodb" {
		t.Errorf("Name() = %q, want %q", got, "mongodb")
	}
	if err := adapter.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}
