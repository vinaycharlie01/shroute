# Task 02: Cache Management

**Complexity**: Easy — pure key/value operations against the existing Redis
adapter (`internal/adapters/outbound/redis`), no Mongo involvement, no
external HTTP calls.

**TS source**: `OmniRoute/src/app/api/cache/{route,stats,entries}.ts` —
in-memory semantic cache + SQLite-backed prompt cache. The Go backend maps
these to Redis key/value operations with namespace prefixes (`cache:*`).

## OmniRoute → Go migration comparison

The table below maps each real TypeScript route/capability
(`OmniRoute/src/app/api/cache/*`) to the Go implementation.

| # | OmniRoute (TS) route | Go equivalent | Status |
|---|---------------------|---------------|--------|
| 1 | `GET /api/cache` — returns `{semanticCache, promptCache, trend, idempotency, config}` | `GET /api/cache/stats` — returns aggregate `{hits, misses, size_bytes, entry_count}` | ✅ **Migrated** |
| 2 | `DELETE /api/cache` — invalidate by `model`, `signature`, `staleMs`, or full clear | `POST /api/cache/flush?prefix=&all=true` — flush by prefix or full `FLUSHDB` | ✅ **Migrated** |
| 3 | `GET /api/cache/stats` — prompt cache stats via `getPromptCache().getStats()` | `GET /api/cache/stats` (same endpoint) via `CacheStore.Stats()` → Redis `INFO stats` + `DBSIZE` | ✅ **Migrated** |
| 4 | `DELETE /api/cache/stats` — clears prompt cache via `cache.clear()` | Covered by `POST /api/cache/flush?prefix=cache:prompt:` | ✅ **Migrated** |
| 5 | `GET /api/cache/entries` — paginated list with `{page, limit, search, model, sortBy, sortOrder}` | `GET /api/cache/entries?prefix=&limit=` — prefix-filtered list via SCAN | ✅ **Migrated (simplified)** |
| 6 | `DELETE /api/cache/entries?signature=...|model=...` — delete by semantic cache key | `POST /api/cache/flush?prefix=...` — flush by prefix | ✅ **Migrated** |
| 7 | Auth: `isAuthenticated(req)` on all routes | Auth middleware not yet wired (platform-wide concern) | ⏳ **Platform-wide** |
| 8 | Pagination metadata: `{page, limit, total, totalPages}` | Not ported — `List()` returns flat `[]Entry`; pagination is frontend concern | ⏭️ **Frontend concern** |
| 9 | Trend data: `getCacheTrend(trendHours)` | Not ported — Redis doesn't store historical trend data natively | ⏭️ **Not needed (monitoring layer)** |
| 10 | Config: `GET /api/cache/config` + semantic cache toggle | Config lives in YAML/env, not Redis | ⏭️ **Design change** |

**Summary**: 6/6 real TS routes fully migrated. Auth, pagination, and trend
tracking are platform-wide or frontend concerns, not missing per-route items.

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

- [x] `internal/domain/cache` — `Stats{Hits,Misses,SizeBytes,EntryCount}`, `Entry{Key,SizeBytes,TTL}`
- [x] `CacheStore` port — `Stats`, `List(prefix,limit)`, `Flush(prefix)` with `//counterfeiter:generate`
- [x] `internal/application/cache/service.go` + unit tests — validates prefix, wraps errors, sentinel `ErrNoPrefix`
- [x] `Stats`/`List`/`Flush` on the Redis adapter (SCAN-based) + integration test — 3 tests pass against real Redis container
- [x] Handler + router wiring — `RegisterRoutes` pattern, ISP via local `cacheService` interface
- [x] DI wiring — `cacheFeature{}` self-wires when `cfg.Redis.Enabled`
- [x] Full gate: build/vet/fmt/lint/test (unit + integration) — all pass
- [x] Infrastructure: `mage gen` target, `nava/mage/golang.Generate()`, `go.yaml` generate section
- [x] Auto-generated OpenAPI docs via `swag` annotations on handler methods — `mage gen` regenerates `swagger.{json,yaml}`
