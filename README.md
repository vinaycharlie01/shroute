# shroute — OmniRoute Go Backend

Go (Hexagonal Architecture) rewrite of the [OmniRoute](./OmniRoute/) Next.js/TypeScript service.
One unified AI proxy/router; 231 LLM providers, auto-fallback, full feature parity with the
original TS app — built slice-by-slice.

---

## Quick Start

```bash
# Build
go build ./...

# Unit tests
go test ./backend/...

# Integration tests (requires Docker)
go test -tags=integration ./backend/test/integration/...

# Lint
golangci-lint run --timeout=10m

# Run
go run ./backend/cmd/server
```

Config is YAML-first (`backend/config/config.base.yaml` + per-env overlay) with automatic
env-var overrides via `APP_*` struct tags. See [`backend/internal/infrastructure/config/`](backend/internal/infrastructure/config/).

---

## Architecture

```
backend/internal/
  domain/<feature>/          Pure types + business rules. Zero I/O. stdlib only.
  application/<feature>/     Use-case services. Depend only on ports interfaces.
  application/ports/         Outbound interfaces owned by the application layer.
  adapters/inbound/http/     Handlers + router. Depend only on application services (via local interfaces).
  adapters/outbound/mongodb/ Concrete MongoDB implementations of ports.
  adapters/outbound/redis/   Concrete Redis implementations of ports.
  infrastructure/di/         ONLY package allowed to import every layer — wires them together.
  infrastructure/config/     YAML + env-var config loading.
```

Dependency direction is one-way: `adapters → application → domain`.
Every new feature is a vertical slice following the 8-step template in
[`backend/tasks/00-INDEX.md`](backend/tasks/00-INDEX.md).

---

## Migration Status

21 slices ported from OmniRoute TS → Go. Slices are ordered easiest → most complex;
each one reuses ports/adapters established by earlier slices.

| # | Slice | Status | Notes |
|---|-------|--------|-------|
| 01 | Health, Logs & Audit | ✅ Done | `GET/POST /api/audit`, `GET /healthz`, `GET /readyz` |
| 02 | Cache Management | ✅ Done | `GET /api/cache/stats`, `GET /api/cache`, `DELETE /api/cache` |
| **03** | **Settings & Feature Flags** | **✅ Done** | See [Task 03 detail](#task-03-settings--feature-flags) below |
| 04 | API Keys | 🔲 Pending | CRUD + scopes |
| 05 | Models & Mappings | 🔲 Pending | Catalog CRUD + mapping tables |
| 06 | Usage & Quota | 🔲 Pending | Tracking/aggregation queries |
| 07 | Providers | 🔲 Pending | CRUD + outbound HTTP health checks |
| 08 | Compression & Context | 🔲 Pending | Token-aware context compression |
| 09 | Webhooks, Compliance & Guardrails | 🔲 Pending | Async dispatch, rules engine |
| 10 | Combos & Routing Strategies | 🔲 Pending | 17 routing strategies + circuit breaker |
| 11 | Batches, Files & Storage | 🔲 Pending | Binary/blob storage, async batch jobs |
| 12 | Skills & Plugins | 🔲 Pending | Sandboxed execution model |
| 13 | Memory & Gamification | 🔲 Pending | Vector search + scoring rules |
| 14 | Analytics & Format Translator | 🔲 Pending | OpenAI↔Claude↔Gemini translation |
| 15 | DevOps & Infra | 🔲 Pending | DB backup/restore, restart, version-manager |
| 16 | Proxy & Network | 🔲 Pending | Proxy chains, tunnels |
| 17 | CLI Tools | 🔲 Pending | Child-process lifecycle |
| 18 | Auto-Combo & MCP | 🔲 Pending | MCP protocol server, 3 transports |
| 19 | OAuth & CLI Auth | 🔲 Pending | Multi-provider OAuth flows |
| 20 | A2A / ACP Protocols | 🔲 Pending | JSON-RPC 2.0 multi-agent protocol |
| 21 | Agent Bridge & Traffic Inspector | 🔲 Pending | MITM/traffic interception |

---

## Task 03: Settings & Feature Flags

**100% migrated.** Full coverage of the `/api/settings` surface from
[`OmniRoute/vinaydoc/SLICE_10_SETTINGS.md`](OmniRoute/vinaydoc/SLICE_10_SETTINGS.md)
that falls within Task 03 scope (general config + feature flags).
The remaining OmniRoute settings routes (`/api/db-backups`, `/api/system`,
`/api/restart`, `/api/shutdown`, `/api/version-manager`) are deferred to Task 15.

### API Routes

| Method | Route | OmniRoute TS | Go handler |
|--------|-------|-------------|------------|
| `GET` | `/api/settings` | `getAll()` — all settings | `Settings.List` |
| `GET` | `/api/settings/{key}` | `get(key)` — single setting | `Settings.Get` |
| `PUT` | `/api/settings/{key}` | `set(key, value)` — upsert | `Settings.Set` |
| `DELETE` | `/api/settings/{key}` | `delete(key)` | `Settings.Delete` |
| `GET` | `/api/settings/flags` | `listFlags()` | `Settings.ListFlags` |
| `GET` | `/api/settings/flags/{name}` | `getFlag(name)` | `Settings.GetFlag` |
| `PUT` | `/api/settings/flags/{name}` | `setFlag(name, enabled)` | `Settings.SetFlag` |

### Setting Keys (known-key registry)

Unknown keys are rejected at the service layer (mirrors Zod-schema discipline from the TS routes).

| Key | Category |
|-----|----------|
| `log_level` | `general` |
| `max_tokens_default` | `general` |
| `theme` | `ui` |
| `require_api_key` | `security` |
| `allowed_origins` | `security` |
| `rate_limit_enabled` | `security` |
| `default_model` | `routing` |
| `fallback_enabled` | `routing` |

### Feature Flags

Seeded on every startup using `$setOnInsert` (operator-modified values are never overwritten).

| Flag | Default | Notes |
|------|---------|-------|
| `PII_REDACTION_ENABLED` | `false` | **Must stay `false`** — opt-in only (Hard Rule #20) |
| `PII_RESPONSE_SANITIZATION` | `false` | **Must stay `false`** — opt-in only (Hard Rule #20) |
| `RATE_LIMIT_ENABLED` | `false` | Off by default |
| `CACHE_ENABLED` | `true` | On by default |
| `AUDIT_LOG_ENABLED` | `true` | On by default |

### Layer Breakdown

| Layer | Package | What it does |
|-------|---------|--------------|
| Domain | `internal/domain/settings` | `Document`, `FeatureFlag`, `Category` enum, `knownKeys` map, `DefaultFeatureFlags` seed list, sentinel errors |
| Ports | `internal/application/ports` | `SettingsRepository` (Get/Set/Delete/List), `FeatureFlagRepository` (Get/List/Set/SeedDefaults) |
| Application | `internal/application/settings` | Known-key validation, empty-value guard, orchestrates repo calls |
| Outbound adapter | `internal/adapters/outbound/mongodb/settings.go` | `settings` collection (string `_id`), `feature_flags` collection (`$set` upsert + `$setOnInsert` seed) |
| Inbound handler | `internal/adapters/inbound/http/handlers/settings.go` | 7 routes via `RegisterRoutes`; local `settingsManager` interface for decoupling |
| DI feature | `internal/infrastructure/di/settings_feature.go` | Wires repos → service → handler; seeds `DefaultFeatureFlags` at startup |

### Test Coverage

| Test | File | What it verifies |
|------|------|-----------------|
| Domain unit | `internal/domain/settings/settings_test.go` | `IsKnownKey`, `CategoryFor`, PII flags `Enabled=false` guard |
| Service unit | `internal/application/settings/service_test.go` | All CRUD paths, unknown-key rejection, empty-value rejection, PII default-false guard |
| Handler unit | `internal/adapters/inbound/http/handlers/settings_test.go` | All 7 route handlers, error status codes |
| Integration | `backend/test/integration/settings_test.go` | `Set→Get→Delete` round-trip; `SeedDefaults` `$setOnInsert` idempotency against real MongoDB |

### What Task 03 Does NOT Cover (→ Task 15)

The following OmniRoute routes from `SLICE_10_SETTINGS.md` are out of scope for Task 03
and will be migrated in **Task 15 (DevOps & Infra)**:

| Route | Reason deferred |
|-------|-----------------|
| `POST /api/db-backups` | Requires `.backup` SQLite/Mongo command + disk I/O |
| `GET /api/db-backups` | Lists backup files on disk |
| `POST /api/db-backups/{id}/restore` | High blast-radius restore operation |
| `GET /api/system` | Process stats (CPU, memory, uptime) |
| `POST /api/restart` | Process restart — loopback-only, requires child-process management |
| `POST /api/shutdown` | Process shutdown |
| `GET /api/version-manager` | Version catalog + current version |
| `POST /api/version-manager/update` | In-place binary update |

---

## Repo Layout

```
backend/          Go service (the product)
  cmd/server/     main.go entrypoint
  config/         YAML config files (base, dev, prod, …)
  internal/       Hexagonal layers (see Architecture above)
  tasks/          Migration task specs (00-INDEX.md … 21-agent-bridge.md)
  test/           Integration tests + testcontainer helpers
frontend/         New frontend (scaffold stage)
OmniRoute/        Reference only — original TS app being migrated from
nava/             Vendored Mage build tooling
```

---

## Running Tests

```bash
# All unit tests
go test ./backend/...

# Single package
go test ./backend/internal/application/settings/...

# Integration (requires Docker for testcontainers)
go test -tags=integration ./backend/test/integration/...

# With coverage
go test -cover ./backend/...
```
