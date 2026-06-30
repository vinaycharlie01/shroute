//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/mongodb"
	domainaudit "github.com/vinaycharlie01/shroute/backend/internal/domain/audit"
	"github.com/vinaycharlie01/shroute/backend/test/containers"
)

func TestAuditRepository_AppendAndListAgainstRealMongoDB(t *testing.T) {
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

	repo := mongodb.NewAuditRepository(adapter.Database())

	stored, err := repo.Append(ctx, domainaudit.Entry{
		Actor:  "user-1",
		Action: "login",
		Target: "session",
	})
	if err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if stored.ID == "" {
		t.Error("Append() did not assign an ID")
	}
	if stored.CreatedAt.IsZero() {
		t.Error("Append() did not assign CreatedAt")
	}

	entries, err := repo.List(ctx, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("List() returned %d entries, want 1", len(entries))
	}
	if entries[0].ID != stored.ID {
		t.Errorf("List()[0].ID = %q, want %q", entries[0].ID, stored.ID)
	}
}
