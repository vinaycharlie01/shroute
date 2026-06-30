# Task 02: Cache Management

**Complexity**: Easy — pure key/value operations against the existing Redis
adapter (`internal/adapters/outbound/redis`), no Mongo involvement, no
external HTTP calls.

**TS source**: `OmniRoute/vinaydoc/SLICE_06_CACHE.md` — `/api/cache/*`
(stats, flush, entry list). Original tables `reasoning_cache`/`read_cache`/
`api_response_cache` map to Redis key namespaces (`cache:reasoning:*`,
`cache:read:*`), not Mongo collections — there is no SQL here to replace.

## End-to-end flow

1. **Domain** — `internal/domain/cache/cache.go`: `Stats{Hits, Misses,
   SizeBytes, EntryCount int64}`, `Entry{Key string, SizeBytes int64,
   TTL time.Duration}`.
2. **Ports** — add `CacheStore` interface to `ports.go`:
   ```go
   type CacheStore interface {
       Stats(ctx context.Context) (cache.Stats, error)
       List(ctx context.Context, prefix string, limit int) ([]cache.Entry, error)
       Flush(ctx context.Context, prefix string) error
   }
   ```
3. **Application** — `internal/application/cache/service.go`: validates the
   `prefix` argument (reject empty-string "flush everything" unless an
   explicit `all=true` flag is set) before delegating to the port.
4. **Outbound adapter** — extend `internal/adapters/outbound/redis/redis.go`
   (or a sibling file in the same package) with `Stats`/`List`/`Flush`
   methods on the existing `*redis.Adapter`, using `SCAN` (never `KEYS`, to
   avoid blocking) for listing/flushing by prefix, and `INFO`/`DBSIZE` for
   stats. Reuses the single Redis client already wired in `di.Container`.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/cache.go`
   exposing `GET /api/cache/stats`, `GET /api/cache/entries`,
   `POST /api/cache/flush`.
6. **Router/DI** — same extension pattern as Task 01.
7. **Tests** — unit tests against a fake `CacheStore`; integration test
   using `containers.Redis(ctx, t)` (already exists in
   `backend/test/containers/redis.go`) to exercise `Stats`/`List`/`Flush`
   against a real Redis instance.

## Checklist

- [ ] `internal/domain/cache`
- [ ] `CacheStore` port
- [ ] `internal/application/cache/service.go` + unit tests
- [ ] `Stats`/`List`/`Flush` on the Redis adapter (SCAN-based) + integration test
- [ ] Handler + router wiring
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test (unit + integration)
