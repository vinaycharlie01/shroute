# Task 06: Usage & Quota

**Complexity**: Moderate — still CRUD-shaped, but adds time-series
aggregation (daily rollups, cost breakdowns) which means Mongo aggregation
pipelines instead of single-document reads.

**TS source**: `OmniRoute/vinaydoc/SLICE_04_USAGE.md` — `/api/usage/*`,
`/api/quota/*`, `/api/costs`, `/api/usage/analytics`.

## End-to-end flow

1. **Domain** — `internal/domain/usage/usage.go`: `Record{ID, ProviderID,
   ModelID, KeyID string, PromptTokens, CompletionTokens int64, CostUSD
   float64, CreatedAt time.Time}`, `QuotaSnapshot{KeyID string, Used, Limit
   int64, ResetAt time.Time}`.
2. **Ports** — `UsageRepository` (`Append`, `DailyRollup(ctx, from, to)
   ([]usage.DailyTotal, error)`, `CostBreakdown(ctx, groupBy string) (...)`),
   `QuotaRepository` (`Get/Set(ctx, keyID)`) in `ports.go`.
3. **Application** — `internal/application/usage/service.go`: usage writes
   come from the chat-completion pipeline (out of scope here — this slice
   only covers the read/management API), so the service is mostly read +
   aggregation orchestration plus quota threshold checks.
4. **Outbound adapter** — `internal/adapters/outbound/mongodb/usage.go`:
   `usage_history` collection (TTL or archival policy — high write volume),
   `quota_snapshots` collection. Rollups use Mongo's aggregation pipeline
   (`$group`/`$sum` by day) rather than a separate `usage_daily` table —
   this is a case where Mongo's aggregation framework replaces what was a
   second SQL table in the old plan.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/usage.go`:
   `GET /api/usage`, `GET /api/usage/analytics`, `GET /api/costs`,
   `GET/PUT /api/quota/{keyID}`.
6. **Router/DI** — usual extension pattern.
7. **Tests** — unit tests for quota-threshold logic; integration test
   seeding `usage_history` documents and asserting the aggregation pipeline
   produces correct daily totals.

## Checklist

- [ ] `internal/domain/usage`
- [ ] `UsageRepository`, `QuotaRepository` ports
- [ ] `internal/application/usage/service.go` + unit tests
- [ ] Mongo adapter incl. aggregation pipeline for rollups + integration test
- [ ] Handlers + router wiring
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
