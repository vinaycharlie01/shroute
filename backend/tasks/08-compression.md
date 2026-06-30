# Task 08: Compression & Context Management

**Complexity**: Moderate/high — low I/O surface (mostly config CRUD like
Task 03) but the underlying engines (caveman/RTK/lite compression
algorithms) are algorithmically nontrivial. This task only ports the
**configuration/management API**; the compression pipeline itself stays in
TypeScript (`open-sse/`) and reads its config from this Go endpoint over
HTTP, per the existing TS source doc.

**TS source**: `OmniRoute/vinaydoc/SLICE_08_COMPRESSION.md` —
`/api/compression/*`, `/api/context/*`.

## End-to-end flow

1. **Domain** — `internal/domain/compression/compression.go`:
   `Config{ComboID string, Engine EngineType, Threshold int, LanguagePack
   string}`, `EngineType` enum (`Caveman`/`RTK`/`Lite`), `Stats{ComboID
   string, BytesIn, BytesOut int64, Ratio float64}`.
2. **Ports** — `CompressionConfigRepository` (`Get/Set(ctx, comboID)`),
   `CompressionStatsRepository` (`Append`, `Summary(ctx, comboID)`) in
   `ports.go`.
3. **Application** — `internal/application/compression/service.go`:
   validates engine/threshold combinations are coherent (e.g. reject a
   `Threshold` of 0 for `Caveman`) before persisting; pure config
   validation, no compression logic itself lives in Go.
4. **Outbound adapter** — `internal/adapters/outbound/mongodb/compression.go`:
   `compression_config` collection (keyed by `combo_id`),
   `compression_stats` collection (append-only, summarized via aggregation
   like Task 06's usage rollups).
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/compression.go`:
   `GET/PUT /api/compression/config/{comboID}`, `GET /api/compression/stats`,
   `GET/PUT /api/context/*` (context-window assignment, same config shape).
6. **Router/DI** — usual extension pattern.
7. **Tests** — unit tests for the engine/threshold validation matrix;
   integration test for the stats aggregation summary.

## Checklist

- [ ] `internal/domain/compression`
- [ ] `CompressionConfigRepository`, `CompressionStatsRepository` ports
- [ ] `internal/application/compression/service.go` + unit tests
- [ ] Mongo adapter (`compression_config`, `compression_stats`) + integration test
- [ ] Handlers + router wiring
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
