# Migration Task Index — OmniRoute (Next.js/TS) → `backend/` (Go, Hexagonal)

This folder breaks the OmniRoute → Go migration into one file per feature
area ("slice"). Each file describes the **end-to-end API flow** to build for
that slice on top of the architecture that already exists in `backend/`,
proven out by the health-check vertical slice
(`internal/domain/health` → `internal/application/health` →
`internal/adapters/inbound/http/handlers/health.go` →
`internal/adapters/outbound/{mongodb,redis}`).

Files are numbered **easiest → most complex**, not by business priority.
Work top-down: each slice reuses ports/adapters/middleware the previous ones
established, so later slices get cheaper as the foundation grows.

## Two corrections vs. the older `OmniRoute/vinaydoc/` slice plan

The pre-existing planning docs under `OmniRoute/vinaydoc/SLICE_*.md` are
still useful for the **list of routes and fields per feature** (request/
response shapes, edge cases), but their *target architecture* is outdated.
Apply these two substitutions whenever a `vinaydoc` doc says otherwise:

1. **MongoDB, not SQL/Postgres.** `vinaydoc` assumes `internal/db/*.go` with
   `database/sql` + hand-written SQL queries. The actual backend uses
   `go.mongodb.org/mongo-driver/v2` (see `internal/adapters/outbound/mongodb`).
   Every "repository" becomes a Mongo collection wrapper: `bson` structs,
   `*mongo.Collection` methods, indexes created in adapter `New()` or a
   migration helper — no SQL schema, no query files.
2. **True hexagonal layering, not flat `pkg/types`/`internal/db`/`api/handlers`.**
   `vinaydoc` assumes a flat layout. The real target is:

   ```
   internal/domain/<feature>/        pure types + business rules, no I/O
   internal/application/<feature>/   use-case services, depend only on ports
   internal/application/ports/       outbound interfaces (extend ports.go)
   internal/adapters/inbound/http/   handlers + routes (extend router.go)
   internal/adapters/outbound/mongodb/ or a new outbound/<dep>/ package
   internal/infrastructure/di/       wiring (extend container.go)
   ```

   The Hexagonal Architecture is meant to be **extended**, not replaced: add
   new domain/application packages per feature, add new outbound adapters
   only when a feature needs a dependency `mongodb`/`redis` don't already
   cover (e.g. an object-store client for Slice 14, an OAuth HTTP client for
   Slice 11), and keep `di.Container` the single place that wires everything
   together, exactly as it does today for health.

## End-to-end flow template (apply to every slice)

1. **Domain** (`internal/domain/<feature>/`) — plain Go structs/enums for the
   feature's core concepts, zero imports outside stdlib. No persistence
   concerns here (mirrors `internal/domain/health/health.go`).
2. **Ports** (`internal/application/ports/ports.go`) — add the narrow
   interface(s) the use case needs (a `<Feature>Repository` interface, etc.),
   following the existing `Pinger`/`Closer` pattern: small, capability-named,
   owned by the application layer.
3. **Application service** (`internal/application/<feature>/service.go`) —
   the use case: validation, orchestration, calls only `ports` interfaces.
   Mirrors `internal/application/health/service.go`.
4. **Outbound adapter** (`internal/adapters/outbound/mongodb/<feature>.go` or
   a new file in the existing `mongodb` package) — implements the port
   against `*mongo.Collection`. Reuses the single `mongodb.Adapter.Client()`
   already wired in `di.Container`; do not open a second Mongo connection.
5. **Inbound handler** (`internal/adapters/inbound/http/handlers/<feature>.go`)
   — translates `http.Request`/`ResponseWriter` to/from the application
   service, declares its own minimal local interface for the service it
   depends on (see `healthChecker` in `health.go`), never imports the
   concrete `application/<feature>` package directly.
6. **Router wiring** (`internal/adapters/inbound/http/router.go`) — add the
   new routes to `RouterConfig` + `mux.HandleFunc(...)`, same pattern as
   `GET /healthz`/`GET /readyz`.
7. **DI wiring** (`internal/infrastructure/di/container.go`) — construct the
   adapter, service, handler and pass them into `RouterConfig`, same pattern
   as `healthService`/`healthHandler`.
8. **Tests** — unit tests for domain + application service (table-driven,
   fake port implementations); integration test under `backend/test/integration/`
   using `backend/test/containers.MongoDB(ctx, t)` for the real adapter,
   following `mongodb_test.go`.

## Slice order (easy → complex)

| # | File | Slice | Why this position |
|---|------|-------|---|
| 01 | `01-health-logs-audit.md` | Health, Logs & Audit | Extends the foundation slice directly; storage is append-only logs in Mongo, no business logic |
| 02 | `02-cache.md` | Cache | Pure Redis KV CRUD, adapter already exists |
| 03 | `03-settings.md` | Settings & Feature Flags | Mongo CRUD over config documents, no external calls |
| 04 | `04-api-keys.md` | API Keys | CRUD + scopes; security-sensitive but self-contained |
| 05 | `05-models.md` | Models & Mappings | Catalog CRUD + mapping tables |
| 06 | `06-usage-quota.md` | Usage & Quota | Tracking/aggregation queries, still CRUD-shaped |
| 07 | `07-providers.md` | Providers | CRUD + outbound HTTP health checks + model sync (first real external-call slice) |
| 08 | `08-compression.md` | Compression & Context | Algorithmically complex (token-aware compression), low I/O surface |
| 09 | `09-webhooks-compliance.md` | Webhooks, Compliance & Guardrails | Async dispatch, rules engine, signature verification |
| 10 | `10-combos.md` | Combos & Routing Strategies | 17 routing strategies + circuit breaker integration — first slice with significant orchestration logic |
| 11 | `11-batches-files-storage.md` | Batches, Files & Storage | Binary/blob storage, async batch jobs, new outbound adapter needed |
| 12 | `12-skills-plugins.md` | Skills & Plugins | Sandboxed execution model |
| 13 | `13-memory-gamification.md` | Memory & Gamification | Vector search (Qdrant) + FTS5-equivalent in Mongo, scoring rules |
| 14 | `14-analytics-translator.md` | Analytics & Format Translator | OpenAI↔Claude↔Gemini translation logic + aggregation pipelines |
| 15 | `15-devops-infra.md` | DevOps & Infra | Process spawning, backups, system control — high blast radius, must be loopback-only |
| 16 | `16-proxy-network.md` | Proxy & Network | Proxy chains, tunnels, connection management |
| 17 | `17-cli-tools.md` | CLI Tools | Child-process lifecycle management, cross-platform |
| 18 | `18-autocombo-mcp.md` | Auto-Combo & MCP | MCP protocol server (tools, 3 transports), streaming |
| 19 | `19-oauth-cli-auth.md` | OAuth & CLI Auth | Multi-provider OAuth flows, token refresh, security-critical |
| 20 | `20-a2a-acp-protocols.md` | A2A / ACP Protocols | JSON-RPC 2.0 multi-agent protocol, task orchestration |
| 21 | `21-agent-bridge-traffic.md` | Agent Bridge & Traffic Inspector | MITM/traffic interception, cert install — highest complexity & risk |

## Source material

- Route lists / request-response shapes: `OmniRoute/vinaydoc/SLICE_01..21_*.md`
  (use for field-level detail; ignore their SQL/flat-layout sections).
- Live architecture reference: `internal/domain/health`,
  `internal/application/health`, `internal/application/ports/ports.go`,
  `internal/adapters/inbound/http/{router.go,handlers/health.go}`,
  `internal/adapters/outbound/{mongodb,redis}`, `internal/infrastructure/di/container.go`.
- Coverage/lint/test gates: run `go build ./...`, `go vet ./...`, `gofmt -l .`,
  `golangci-lint run --timeout=10m`, `go test ./...` and
  `go test -tags=integration ./test/integration/...` before considering a
  slice done, same as the foundation and MongoDB-swap commits.
