//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/redis"
	"github.com/vinaycharlie01/shroute/backend/test/containers"
)

func TestCacheStore_StatsAgainstRealRedis(t *testing.T) {
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

	store := redis.NewCacheStore(adapter)
	stats, err := store.Stats(ctx)
	if err != nil {
		t.Fatalf("CacheStore.Stats() error = %v", err)
	}

	// A fresh Redis must report zero entries.
	if stats.EntryCount != 0 {
		t.Errorf("EntryCount = %d, want 0", stats.EntryCount)
	}
}

func TestCacheStore_ListAgainstRealRedis(t *testing.T) {
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

	// Seed test data: one key matching prefix "cache:", one not.
	client := adapter.Client()
	if err := client.Set(ctx, "cache:user:1", "alice", 0).Err(); err != nil {
		t.Fatalf("seed cache:user:1: %v", err)
	}
	if err := client.Set(ctx, "cache:user:2", "bob", 0).Err(); err != nil {
		t.Fatalf("seed cache:user:2: %v", err)
	}
	if err := client.Set(ctx, "session:abc", "data", 0).Err(); err != nil {
		t.Fatalf("seed session:abc: %v", err)
	}

	store := redis.NewCacheStore(adapter)
	entries, err := store.List(ctx, "cache:", 10)
	if err != nil {
		t.Fatalf("CacheStore.List() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("List() returned %d entries, want 2", len(entries))
	}

	// Must only contain keys starting with "cache:"
	for _, e := range entries {
		if len(e.Key) < 6 || e.Key[:6] != "cache:" {
			t.Errorf("List() returned unexpected key %q", e.Key)
		}
	}
}

func TestCacheStore_FlushAgainstRealRedis_Prefix(t *testing.T) {
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

	client := adapter.Client()
	if err := client.Set(ctx, "cache:user:1", "alice", 0).Err(); err != nil {
		t.Fatalf("seed cache:user:1: %v", err)
	}
	if err := client.Set(ctx, "other:key", "keep", 0).Err(); err != nil {
		t.Fatalf("seed other:key: %v", err)
	}

	store := redis.NewCacheStore(adapter)
	if err := store.Flush(ctx, "cache:"); err != nil {
		t.Fatalf("CacheStore.Flush() error = %v", err)
	}

	// Verify: cache: entries gone, other: entries remain.
	count, err := client.Exists(ctx, "cache:user:1").Result()
	if err != nil {
		t.Fatalf("exists cache:user:1: %v", err)
	}
	if count != 0 {
		t.Error("Flush(cache:) did not remove cache:user:1")
	}

	count, err = client.Exists(ctx, "other:key").Result()
	if err != nil {
		t.Fatalf("exists other:key: %v", err)
	}
	if count != 1 {
		t.Error("Flush(cache:) removed other:key which should have been kept")
	}
}
