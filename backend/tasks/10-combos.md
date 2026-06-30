# Task 10: Combos & Routing Strategies

**Complexity**: High — the first slice with significant orchestration
logic: 17 routing strategies (priority, weighted, fill-first, round-robin,
P2C, random, least-used, cost-optimized, reset-aware, reset-window,
headroom, strict-random, auto, lkgp, context-optimized, context-relay,
fusion) plus circuit-breaker-aware candidate filtering. Build it after
Providers/Models/Usage (Tasks 05-07) since strategy selection reads all
three.

**TS source**: `OmniRoute/vinaydoc/SLICE_02_COMBOS.md` — `/api/combos/*`.
Cross-reference `docs/routing/AUTO-COMBO.md` and
`docs/architecture/RESILIENCE_GUIDE.md` in the OmniRoute TS repo for the
9-factor scoring + 3-layer resilience semantics this slice must reproduce.

## End-to-end flow

1. **Domain** — `internal/domain/combo/combo.go`: `Combo{ID, Name string,
   Strategy StrategyType, Steps []Step}`, `Step{ModelID string, Priority,
   Weight int}`. `StrategyType` as a typed enum with all 17 values. Keep the
   **strategy selection algorithms themselves** as pure functions in
   `internal/domain/combo/strategy/*.go` (one file per strategy or grouped
   by family) — they take `[]Step` + filtered candidate state and return an
   ordered choice, no I/O, fully unit-testable in isolation.
2. **Ports** — `ComboRepository` (CRUD), and reuse `ProviderRepository`/
   `ModelRepository` from Tasks 05/07 as read dependencies (do not duplicate
   provider/model fetching logic here) plus a new `CircuitBreakerStore`
   port:
   ```go
   type CircuitBreakerStore interface {
       Status(ctx context.Context, providerID string) (CircuitState, error)
   }
   ```
3. **Application** — `internal/application/combo/service.go`: `Resolve(ctx,
   comboID) (Target, error)` — loads the combo, filters steps whose provider
   circuit is open (mirrors the TS `canExecute()`/`getStatus()` lazy-recovery
   semantics from `RESILIENCE_GUIDE.md`), dispatches to the right
   `strategy` package function. This is the Go equivalent of
   `resolveComboTargets()`/`handleSingleModel()` in the TS `open-sse/services/combo.ts`.
   The `fusion` strategy (fan-out to N models + judge synthesis) is the one
   exception needing a second port, `Fanout(ctx, models, prompt) ([]Result,
   error)` — stub it behind an interface so it can be implemented once the
   chat-completion proxy itself is ported (later, out of this repo's
   current scope).
4. **Outbound adapter** — `internal/adapters/outbound/mongodb/combo.go`;
   `CircuitBreakerStore` can initially be backed by the existing
   `internal/adapters/outbound/redis` adapter (circuit state as Redis keys
   with TTL, matching the TS `domain_circuit_breakers` table's lazy-expiry
   behavior) — implement as a new file in the `redis` package rather than a
   new outbound package.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/combo.go`:
   CRUD + `POST /api/combos/{id}/resolve` (preview which target a combo
   would pick right now, useful for debugging strategies).
6. **Router/DI** — usual extension pattern.
7. **Tests** — exhaustive unit tests per strategy function (this is the
   highest-value test surface in the whole migration — port the TS
   strategy test cases 1:1 where they exist); integration test for
   `CircuitBreakerStore` lazy-expiry against real Redis.

## Checklist

- [ ] `internal/domain/combo` + `internal/domain/combo/strategy/*` (17 strategies, pure functions)
- [ ] `ComboRepository`, `CircuitBreakerStore` ports
- [ ] `internal/application/combo/service.go` + unit tests (all 17 strategies)
- [ ] Mongo adapter (combos) + Redis-backed circuit breaker store + integration tests
- [ ] Handlers + router wiring (incl. `/resolve` preview)
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
