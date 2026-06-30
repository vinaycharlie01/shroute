# Task 14: Analytics & Format Translator

**Complexity**: High — the format translator (OpenAI↔Claude↔Gemini request/
response shapes) is intricate, format-specific logic with many edge cases
(streaming deltas, tool-call shapes, multimodal content blocks); analytics
adds non-trivial aggregation pipelines on top.

**TS source**: `OmniRoute/vinaydoc/SLICE_21_ANALYTICS_TRANSLATOR.md` —
`/api/analytics/*`, `/api/translator/*`, `/api/playground/*`, `/api/docs/*`,
`/api/openapi/*`, `/api/pricing/*`, `/api/auth/*`, `/api/assess`,
`/api/free-models/*`, `/api/free-tier/*`, `/api/intelligence/sync`,
`/api/telemetry/*`, `/api/token-health`. Cross-reference
`open-sse/translator/` in the OmniRoute TS repo for the exact format
conversion rules to port.

## End-to-end flow

1. **Domain** — `internal/domain/translate/translate.go`: typed request/
   response structs per format (`OpenAIChatRequest`, `ClaudeMessageRequest`,
   `GeminiGenerateRequest`, etc.) plus a `CanonicalRequest`/`CanonicalResponse`
   intermediate form, mirroring the TS translator's hub-and-spoke design
   (every format converts to/from the canonical form, not pairwise).
   `internal/domain/analytics/analytics.go`: `Metric{Name string, Value
   float64, Dimensions map[string]string, Timestamp time.Time}`.
2. **Ports** — `Translator` (`ToCanonical(ctx, format Format, raw []byte)
   (CanonicalRequest, error)`, `FromCanonical(ctx, format Format, resp
   CanonicalResponse) ([]byte, error)`) — note this can plausibly live
   entirely in `internal/application/translate/` as pure functions with no
   port at all, since translation has no I/O dependency; only add a port if
   a concrete implementation needs external state. `AnalyticsRepository`
   (`Append`, `Query(ctx, filter) ([]analytics.Metric, error)`) in `ports.go`.
3. **Application** — `internal/application/translate/service.go`: pure
   format-conversion logic, heavily unit-tested per format pair (this is
   the highest test-count slice — port the TS translator's test fixtures
   1:1 where possible). `internal/application/analytics/service.go`: query
   builder over `AnalyticsRepository.Query`.
4. **Outbound adapter** — `internal/adapters/outbound/mongodb/analytics.go`:
   `analytics_metrics` collection (time-series-pattern: bucketed by
   day/metric name, aggregation pipeline for rollups, same approach as
   Task 06's usage rollups).
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/{translate,analytics}.go`:
   `POST /api/translator/{from}/{to}`, `GET /api/analytics`,
   `GET /api/pricing`, `GET /api/token-health`.
6. **Router/DI** — usual extension pattern.
7. **Tests** — exhaustive table-driven unit tests per format pair
   (OpenAI→Claude, Claude→OpenAI, OpenAI→Gemini, etc.), including streaming
   delta reassembly edge cases; integration test for analytics aggregation.

## Checklist

- [ ] `internal/domain/translate` (canonical + per-format structs), `internal/domain/analytics`
- [ ] `AnalyticsRepository` port (translator likely needs no port — pure functions)
- [ ] `internal/application/translate/service.go` + exhaustive unit tests per format pair
- [ ] `internal/application/analytics/service.go` + unit tests
- [ ] Mongo analytics adapter (aggregation rollups) + integration test
- [ ] Handlers + router wiring
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
