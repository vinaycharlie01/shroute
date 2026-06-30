package cache_test

import (
	"context"
	"errors"
	"testing"

	appcache "github.com/vinaycharlie01/shroute/backend/internal/application/cache"
	"github.com/vinaycharlie01/shroute/backend/internal/application/ports/portsfakes"
	domaincache "github.com/vinaycharlie01/shroute/backend/internal/domain/cache"
)

func TestService_Stats(t *testing.T) {
	t.Parallel()

	fake := new(portsfakes.FakeCacheStore)
	want := domaincache.Stats{Hits: 10, Misses: 2, SizeBytes: 1024, EntryCount: 5}
	fake.StatsReturns(want, nil)

	svc := appcache.NewService(fake)
	got, err := svc.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats() unexpected error: %v", err)
	}

	if got != want {
		t.Errorf("Stats() = %+v, want %+v", got, want)
	}

	if fake.StatsCallCount() != 1 {
		t.Errorf("StatsCallCount() = %d, want 1", fake.StatsCallCount())
	}
}

func TestService_Stats_Error(t *testing.T) {
	t.Parallel()

	fake := new(portsfakes.FakeCacheStore)
	fake.StatsReturns(domaincache.Stats{}, errors.New("redis down"))

	svc := appcache.NewService(fake)
	_, err := svc.Stats(context.Background())
	if err == nil {
		t.Fatal("Stats() expected error, got nil")
	}
}

func TestService_List_RequiresPrefix(t *testing.T) {
	t.Parallel()

	svc := appcache.NewService(new(portsfakes.FakeCacheStore))
	_, err := svc.List(context.Background(), "", 10)
	if !errors.Is(err, appcache.ErrNoPrefix) {
		t.Fatalf("List() err = %v, want %v", err, appcache.ErrNoPrefix)
	}
}

func TestService_List_Success(t *testing.T) {
	t.Parallel()

	fake := new(portsfakes.FakeCacheStore)
	want := []domaincache.Entry{
		{Key: "cache:foo", SizeBytes: 100},
	}
	fake.ListReturns(want, nil)

	svc := appcache.NewService(fake)
	entries, err := svc.List(context.Background(), "cache:", 10)
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if len(entries) != 1 || entries[0].Key != "cache:foo" {
		t.Errorf("List() = %+v, want [{Key:cache:foo ...}]", entries)
	}
}

func TestService_List_DelegatesPrefix(t *testing.T) {
	t.Parallel()

	fake := new(portsfakes.FakeCacheStore)
	fake.ListReturns(nil, nil)

	svc := appcache.NewService(fake)
	_, _ = svc.List(context.Background(), "cache:reasoning:", 20)

	_, prefix, limit := fake.ListArgsForCall(0)
	if prefix != "cache:reasoning:" {
		t.Errorf("List prefix = %q, want %q", prefix, "cache:reasoning:")
	}
	if limit != 20 {
		t.Errorf("List limit = %d, want 20", limit)
	}
}

func TestService_Flush_RequiresPrefix(t *testing.T) {
	t.Parallel()

	svc := appcache.NewService(new(portsfakes.FakeCacheStore))
	err := svc.Flush(context.Background(), "")
	if !errors.Is(err, appcache.ErrNoPrefix) {
		t.Fatalf("Flush() err = %v, want %v", err, appcache.ErrNoPrefix)
	}
}

func TestService_Flush_Success(t *testing.T) {
	t.Parallel()

	fake := new(portsfakes.FakeCacheStore)
	fake.FlushReturns(nil)

	svc := appcache.NewService(fake)
	if err := svc.Flush(context.Background(), "cache:"); err != nil {
		t.Fatalf("Flush() unexpected error: %v", err)
	}

	if fake.FlushCallCount() != 1 {
		t.Errorf("FlushCallCount() = %d, want 1", fake.FlushCallCount())
	}

	_, prefix := fake.FlushArgsForCall(0)
	if prefix != "cache:" {
		t.Errorf("Flush prefix = %q, want %q", prefix, "cache:")
	}
}
