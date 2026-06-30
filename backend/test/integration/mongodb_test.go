//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/mongodb"
)

func TestAdapter_PingAgainstRealMongoDB(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := tcmongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("start mongodb container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("terminate mongodb container: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

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
