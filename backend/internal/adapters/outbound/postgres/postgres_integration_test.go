//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/postgres"
)

func TestAdapter_PingAgainstRealPostgres(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("shroute"),
		tcpostgres.WithUsername("shroute"),
		tcpostgres.WithPassword("shroute"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("terminate postgres container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	adapter, err := postgres.New(ctx, dsn)
	if err != nil {
		t.Fatalf("postgres.New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := adapter.Close(); err != nil {
			t.Logf("close adapter: %v", err)
		}
	})

	if got := adapter.Name(); got != "postgres" {
		t.Errorf("Name() = %q, want %q", got, "postgres")
	}
	if err := adapter.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}
