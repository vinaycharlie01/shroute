package redis

import (
	"context"
	"fmt"

	domaincache "github.com/vinaycharlie01/shroute/backend/internal/domain/cache"
)

const (
	// defaultScanCount is the cursor batch size for SCAN.
	defaultScanCount = 100
)

// CacheStore implements ports.CacheStore against a Redis instance.
type CacheStore struct {
	adapter *Adapter
}

// NewCacheStore builds a CacheStore that reuses the shared Redis adapter.
func NewCacheStore(adapter *Adapter) *CacheStore {
	return &CacheStore{adapter: adapter}
}

// Stats returns aggregate cache statistics using Redis INFO and DBSIZE.
func (s *CacheStore) Stats(ctx context.Context) (domaincache.Stats, error) {
	info, err := s.adapter.client.Info(ctx, "stats").Result()
	if err != nil {
		return domaincache.Stats{}, fmt.Errorf("redis cache: info: %w", err)
	}

	var hits, misses int64
	_, _ = fmt.Sscanf(info, "keyspace_hits:%d\r\nkeyspace_misses:%d", &hits, &misses)

	dbsize, err := s.adapter.client.DBSize(ctx).Result()
	if err != nil {
		return domaincache.Stats{}, fmt.Errorf("redis cache: dbsize: %w", err)
	}

	// Approximate memory usage via MEMORY STATS.
	memory, err := s.adapter.client.MemoryUsage(ctx, "total").Result()
	if err != nil {
		// MEMORY USAGE on a non-existent key returns -2; fall back to INFO.
		var memInfo string
		memInfo, err = s.adapter.client.Info(ctx, "memory").Result()
		if err != nil {
			return domaincache.Stats{}, fmt.Errorf("redis cache: memory info: %w", err)
		}
		var used int64
		_, _ = fmt.Sscanf(memInfo, "used_memory:%d", &used)
		memory = used
	}

	return domaincache.Stats{
		Hits:       hits,
		Misses:     misses,
		SizeBytes:  memory,
		EntryCount: dbsize,
	}, nil
}

// List returns up to limit entries matching the given key prefix using SCAN.
func (s *CacheStore) List(ctx context.Context, prefix string, limit int) ([]domaincache.Entry, error) {
	if limit <= 0 {
		limit = defaultScanCount
	}

	pattern := prefix + "*"
	var cursor uint64
	var entries []domaincache.Entry

	for {
		keys, nextCursor, err := s.adapter.client.Scan(ctx, cursor, pattern, int64(limit-len(entries))).Result()
		if err != nil {
			return nil, fmt.Errorf("redis cache: scan: %w", err)
		}
		cursor = nextCursor

		for _, key := range keys {
			ttl, err := s.adapter.client.TTL(ctx, key).Result()
			if err != nil {
				continue
			}

			mem, err := s.adapter.client.MemoryUsage(ctx, key).Result()
			if err != nil {
				mem = 0
			}

			entries = append(entries, domaincache.Entry{
				Key:       key,
				SizeBytes: mem,
				TTL:       ttl,
			})

			if len(entries) >= limit {
				return entries, nil
			}
		}

		if cursor == 0 {
			break
		}
	}

	return entries, nil
}

// Flush removes every key matching the prefix using SCAN and DEL. An empty
// prefix flushes the entire database via FLUSHDB.
func (s *CacheStore) Flush(ctx context.Context, prefix string) error {
	if prefix == "" {
		return s.adapter.client.FlushDB(ctx).Err()
	}

	pattern := prefix + "*"
	var cursor uint64

	for {
		keys, nextCursor, err := s.adapter.client.Scan(ctx, cursor, pattern, defaultScanCount).Result()
		if err != nil {
			return fmt.Errorf("redis cache: scan flush: %w", err)
		}
		cursor = nextCursor

		if len(keys) > 0 {
			if err := s.adapter.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("redis cache: del: %w", err)
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}
