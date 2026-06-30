# Task 07: Providers

**Complexity**: Moderate/high — first slice with real outbound HTTP calls to
third-party APIs (health checks, model sync), not just persistence. Build it
after Models (Task 05) so `ProviderID` references already have somewhere to
point.

**TS source**: `OmniRoute/vinaydoc/SLICE_01_PROVIDERS.md` —
`/api/providers/*`, `/api/v1/providers/*`, `/api/provider-models`,
`/api/provider-metrics`, `/api/provider-stats`, `/api/provider-nodes/*`,
`/api/free-models`, `/api/free-provider-rankings`.

## End-to-end flow

1. **Domain** — `internal/domain/provider/provider.go`: `Provider{ID, Name,
   Type, Category, BaseURL string, IsActive bool, HealthStatus
   HealthState}`, `Node{ID, ProviderID, URL string, Weight int, IsActive
   bool}`. `HealthState` as a typed enum (`Healthy`/`Degraded`/`Down`/
   `Unknown`), same convention as `domainhealth.State` in
   `internal/domain/health/health.go`.
2. **Ports** — `ProviderRepository` (CRUD + bulk + node sub-resource) in
   `ports.go`; a separate `ProviderProbe` port:
   ```go
   type ProviderProbe interface {
       Test(ctx context.Context, p provider.Provider) (provider.HealthState, error)
       SyncModels(ctx context.Context, p provider.Provider) ([]string, error)
   }
   ```
   keeping "talk to the real upstream provider over HTTP" cleanly separate
   from "persist provider config" — this is the application boundary that
   matters most in this slice.
3. **Application** — `internal/application/provider/service.go`: `Test`/
   `TestBatch` call `ProviderProbe`, then persist the resulting
   `HealthState` via `ProviderRepository` — mirrors how
   `application/health/service.go` already aggregates `ports.Pinger`
   results, just persisted instead of returned inline.
4. **Outbound adapters** — `internal/adapters/outbound/mongodb/provider.go`
   (CRUD + bulk + nodes); new `internal/adapters/outbound/providerhttp/probe.go`
   implementing `ProviderProbe` with a `net/http.Client` (bounded timeout,
   no `eval`/dynamic code per CLAUDE.md hard rule #3) — second non-Mongo/Redis
   outbound adapter, demonstrating the hexagonal extension pattern for
   "talks to an arbitrary external HTTP endpoint."
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/provider.go`:
   full CRUD + `/test`, `/sync-models`, `/nodes` sub-routes,
   `/api/free-models`, `/api/free-provider-rankings`.
6. **Router/DI** — usual extension; `di.Container` constructs the
   `providerhttp.Probe` with a sane default timeout from `config.Config`.
7. **Tests** — unit tests with a fake `ProviderProbe` (no real network
   calls in unit tests); integration test against `httptest.Server` for the
   probe, and `containers.MongoDB` for persistence.

## Checklist

- [ ] `internal/domain/provider`
- [ ] `ProviderRepository`, `ProviderProbe` ports
- [ ] `internal/application/provider/service.go` + unit tests
- [ ] Mongo adapter + `providerhttp` outbound adapter + integration tests
- [ ] Handlers + router wiring (incl. `/test`, `/sync-models`, `/nodes`)
- [ ] DI wiring with configurable HTTP timeout
- [ ] Full gate: build/vet/fmt/lint/test
