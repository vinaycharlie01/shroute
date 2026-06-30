# Task 13: Memory & Gamification

**Complexity**: High — needs vector search (the TS side uses Qdrant +
FTS5), which means either standing up a Mongo Atlas Vector Search index (if
available) or adding a genuinely new outbound adapter for a vector DB
client, plus a scoring/rules engine for gamification.

**TS source**: `OmniRoute/vinaydoc/SLICE_16_MEMORY_GAMIFICATION.md` —
`/api/memory/*`, `/api/gamification/*`, `/api/evals/*`. Cross-reference
`docs/frameworks/MEMORY.md` in the OmniRoute TS repo (FTS5 + Qdrant hybrid
search) for the retrieval contract this slice must reproduce.

## End-to-end flow

1. **Domain** — `internal/domain/memory/memory.go`: `Entry{ID, Content
   string, Embedding []float32, Tags []string, CreatedAt time.Time}`.
   `internal/domain/gamification/gamification.go`: `Score{UserID string,
   Points int64, Level int}`, `Achievement{ID, Name string, Criteria Rule}`.
2. **Ports** — `MemoryRepository` (`Append`, `SearchText(ctx, query string)`,
   `SearchVector(ctx, embedding []float32, k int)`, hybrid `Search` combining
   both — mirrors the FTS5+Qdrant hybrid on the TS side), `EmbeddingClient`
   (`Embed(ctx, text string) ([]float32, error)` — calls whatever embedding
   model/provider is configured, kept separate from storage),
   `GamificationRepository` (CRUD scores/achievements) in `ports.go`.
3. **Application** — `internal/application/memory/service.go`: on `Append`,
   calls `EmbeddingClient.Embed` then `MemoryRepository.Append`; `Search`
   embeds the query and merges text + vector results (reciprocal rank
   fusion or similar, ported from the TS hybrid-search ranking logic).
   `internal/application/gamification/service.go`: evaluates `Achievement`
   criteria against a `Score` update, pure rule evaluation.
4. **Outbound adapters** — `internal/adapters/outbound/mongodb/{memory,gamification}.go`.
   For vector search specifically: if MongoDB Atlas Vector Search is
   available in the target deployment, implement `SearchVector` via a
   `$vectorSearch` aggregation stage in the same `mongodb` package (no new
   adapter needed); otherwise add a new
   `internal/adapters/outbound/vectorstore/` package wrapping a self-hosted
   vector DB client — decide this only once the deployment target for this
   slice is confirmed, since it changes which adapter gets built. `EmbeddingClient`
   reuses the existing `providerhttp`-style pattern from Task 07 (an
   outbound HTTP client to whatever embedding provider is configured).
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/{memory,gamification}.go`:
   `POST/GET /api/memory`, `GET /api/memory/search`,
   `GET/POST /api/gamification/score`.
6. **Router/DI** — usual extension pattern.
7. **Tests** — unit tests for hybrid-search ranking and achievement-rule
   evaluation (fakes for `MemoryRepository`/`EmbeddingClient`); integration
   test for whichever vector-search backend is chosen.

## Checklist

- [ ] `internal/domain/memory`, `internal/domain/gamification`
- [ ] `MemoryRepository`, `EmbeddingClient`, `GamificationRepository` ports
- [ ] Application services (hybrid search ranking, achievement rules) + unit tests
- [ ] Vector-search adapter decision recorded + implemented + integration test
- [ ] Handlers + router wiring
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
