# CLAUDE.md

Guidance for Claude Code (and any AI agent) working in this repository.
These rules OVERRIDE default behavior — follow them exactly. The same
ruleset, condensed, lives in `.clinerules` for Cline; keep both in sync if
you change one.

## Repo Map — what's actually being developed here

| Path         | What it is                                                                                          |
| ------------ | ---------------------------------------------------------------------------------------------------- |
| `backend/`   | **The product.** Go service, Hexagonal Architecture. Everything below applies here.                   |
| `frontend/`  | New Go-backend's frontend (scaffold stage).                                                           |
| `backend/tasks/` | Per-feature migration task files (easy→complex), each an 8-step end-to-end-flow template.          |
| `OmniRoute/` | **Reference only** — the original Next.js/TS app being migrated from. Has its own `CLAUDE.md` (TS-specific rules) that applies *inside that subtree only*. Do not apply Go rules there; do not port its rules here. |
| `nava/`      | **Vendored** Mage build tooling, reused/extended, not rewritten. Excluded from `.golangci.yml` linting. |

When asked to build a new feature, the work happens in `backend/` (and
`frontend/` once it's live), using `OmniRoute/vinaydoc/SLICE_*.md` only as a
reference for routes/fields — never for target architecture (see
`backend/tasks/00-INDEX.md` for the two corrections: MongoDB not SQL, true
hexagonal not flat layout).

## Architecture: Hexagonal — Extend, Never Replace

```
internal/domain/<feature>/         pure types + business rules, zero I/O, stdlib only
internal/application/<feature>/    use-case services; depend only on internal/application/ports
internal/application/ports/        outbound interfaces, owned by the application layer
internal/adapters/inbound/http/    handlers + router; depend only on application services (via local interfaces)
internal/adapters/outbound/<dep>/  concrete implementations of ports (mongodb, redis, future deps)
internal/infrastructure/di/        the ONLY package allowed to import every layer and wire them together
internal/infrastructure/config/    YAML+env config loading
```

Dependency direction is one-way: `adapters` → `application` → `domain`.
`domain` never imports `application`; `application` never imports
`adapters`. The only place concrete adapters meet application services is
`internal/infrastructure/di/container.go`.

**Adding a feature = extending this structure, never bypassing it.** Follow
the 8-step flow in `backend/tasks/00-INDEX.md` for every new slice:

1. **Domain** (`internal/domain/<feature>/`) — structs, enums, validation. No imports beyond stdlib.
2. **Port** (`internal/application/ports/ports.go`) — add a narrow, capability-named interface (e.g. `AuditRepository`), not a god-interface. Split persistence from external-call concerns into separate ports when both exist (e.g. `ProviderRepository` vs `ProviderProbe`).
3. **Application service** (`internal/application/<feature>/service.go`) — the use case. Depends only on `ports` interfaces via constructor injection (`NewService(repo ports.XRepository) *Service`). Validates via the domain type, wraps errors with `fmt.Errorf("<pkg>: <op>: %w", err)`.
4. **Outbound adapter** (`internal/adapters/outbound/mongodb/<feature>.go`, or a new `outbound/<dep>/` package) — implements the port. Owns its own bson/wire-format structs; never leaks them past the adapter boundary.
5. **Inbound handler** (`internal/adapters/inbound/http/handlers/<feature>.go`) — declares its OWN minimal local interface for the service it needs (e.g. `auditRecorder`), never imports the concrete `application/<feature>` package. This is Interface Segregation in practice — see `health.go`'s `healthChecker` and `audit.go`'s `auditRecorder` as the canonical examples.
6. **Router wiring** (`internal/adapters/inbound/http/router.go`) — add the field to `RouterConfig`, nil-guard it if the feature is optional (depends on an adapter that's only enabled in some environments).
7. **DI wiring** (`internal/infrastructure/di/container.go`) — construct the adapter → service → handler chain inside the relevant `if cfg.X.Enabled` block, pass into `RouterConfig`.
8. **Tests** — unit tests per layer + one integration test per outbound adapter (see Testing below).

Reference implementation to copy conventions from: the **audit log slice**
(`internal/domain/audit/`, `internal/application/audit/`,
`internal/adapters/outbound/mongodb/audit.go`,
`internal/adapters/inbound/http/handlers/audit.go`, plus the wiring diffs in
`router.go`/`container.go`). It was built as the worked example validating
this exact template.

## SOLID, applied concretely in this codebase

- **SRP** — one package per concern (`domain` = rules, `application` = orchestration, `adapters` = I/O). A service method does validation + one repo call, not both validation logic and persistence logic inline.
- **OCP** — add a new feature by adding new `domain`/`application`/`adapters` packages and a new `ports` interface; don't modify existing services to special-case a new feature.
- **LSP** — every adapter that implements a port (e.g. `mongodb.AuditRepository` implementing `ports.AuditRepository`) must be fully substitutable — no adapter-specific behavior the port's callers need to know about.
- **ISP** — handlers declare their own narrow local interface (see step 5 above) instead of depending on the full application service type. Ports themselves stay narrow (2-4 methods), not aggregated god-interfaces.
- **DIP** — `application` depends on `ports` (interfaces it owns), never on concrete `adapters/outbound/*` packages. Only `infrastructure/di` is allowed to import both sides and wire them.

## Go Design Patterns Used Here

- **Constructor injection**: `NewXxx(deps ...)` returning `*Xxx`, dependencies passed in, never constructed inside. Variadic constructors (`NewService(deps ...ports.Pinger)`) for "zero or more of this dependency type" (see `healthapp.NewService`).
- **Adapter pattern**: every `internal/adapters/outbound/<dep>/` package adapts a third-party client to a `ports` interface.
- **Strategy via interfaces**: ports are the strategy contract; swapping Mongo for another store means writing a new adapter, zero changes to `application` or `domain`.
- **Sentinel errors** for domain validation (`internal/domain/<feature>/errors.go`, `errors.New(...)` package vars), checked with `errors.Is(...)` at the handler boundary to decide 400 (safe to show) vs 500 (generic message, no leaked internals) — never return raw adapter/driver errors to an HTTP client.
- **Functional options** only once a constructor would otherwise need more than ~3-4 optional parameters; don't reach for them prematurely — a plain typed `Config` struct is usually simpler and is YAML-mappable for free.

## Config: YAML-first, Dynamic Over Constant

- All environment-shaped configuration lives in `backend/config/config.{base,local,development,staging,production}.yaml`, loaded and layered by `internal/infrastructure/config/loader.go` (`config.base.yaml` + per-env overlay).
- **Automatic env-var overrides via struct tags** — every overridable `Config` field carries an `env:"APP_..."` tag (see `internal/infrastructure/config/config.go`), loaded by a single `env.Parse(cfg)` call (`github.com/caarlos0/env/v11`) in `loader.go`. **Adding override support for a new field is a one-line tag addition — never hand-write another `os.Getenv` branch.** The only exception is relational logic that can't be expressed as one field's tag (e.g. "setting `APP_MONGO_URI` implies `Mongo.Enabled = true`"), which stays as explicit code in `applyEnvOverrides` with a comment explaining why.
- **Prefer config/env-driven values over `const`.** A Go `const` is appropriate only for values that are fixed by the code's own logic and could never sensibly vary per deployment (e.g. a MongoDB collection name that mirrors a domain concept, like `auditCollectionName`). Anything that could plausibly differ between local/dev/staging/prod/an operator's deployment — timeouts, hosts, ports, feature toggles, limits, thresholds — belongs in `Config`, not a `const`.
- Validation lives on the `Config` struct via `validator/v10` tags (`validate:"required,oneof=..."`, `required_if=...`). Don't hand-roll validation in `Load()`.
- Optional adapters (Mongo, Redis, future ones) follow the `Enabled bool` + `if cfg.X.Enabled { ... }` gate in `container.go` — the foundation must run with zero external dependencies until a feature actually needs one.

## Logging: `log/slog` Only

- No custom logger wrapper, no logger struct fields threaded through constructors. Use `slog.Default()` / package-level `slog.InfoContext(ctx, ...)` etc., exactly as already established in `internal/infrastructure/di/container.go` (`slog.SetDefault(log)` once at startup via `internal/infrastructure/logger`).
- Always pass `ctx` to the `*Context` slog variants when one is available; attach structured fields (`slog.String(...)`, `slog.Any(...)`), never `fmt.Sprintf` a log line.
- Never log a raw secret/credential/full request body. Match the OmniRoute TS sibling's error-sanitization spirit even though this is a separate Go service: client-facing errors are generic, internal logs carry the detail.

## Testing: TDD, Table-Driven, Minimal-but-90%

- **Write the failing test first** when fixing a bug or adding a use-case branch; then make it pass. The test is the permanent regression guard.
- **Table-driven tests are the default shape** for any function with more than one meaningful case:
  ```go
  tests := []struct {
      name string
      // inputs
      // want
  }{
      {name: "...", /* ... */},
  }
  for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) {
          t.Parallel()
          // ...
      })
  }
  ```
  Put `t.Parallel()` at the top of both the outer test func and every `t.Run` subtest, matching every existing `*_test.go` in this repo (`health_test.go`, `audit_test.go`, `service_test.go`).
- **Keep cases minimal, not exhaustive.** Cover every branch/decision point once — happy path, each validation failure, each error-wrapping path — and stop. Don't add redundant cases that don't exercise a new branch just to inflate a count. The target is **90%+ statement coverage achieved through meaningful branch coverage**, not padding.
- **Unit tests per layer**: domain (`Validate()`/pure-function tests), application (service tests against a hand-written fake of the `ports` interface — see `fakePinger`/`fakeRepo` patterns), inbound handler (stub of the handler's local interface + `httptest.NewRecorder()`).
- **Integration tests are mandatory for every outbound adapter.** Live in `backend/test/integration/`, gated by `//go:build integration`, using the existing `backend/test/containers` testcontainer helpers (`containers.MongoDB(ctx, t)`, etc.) against a real instance of the dependency — never mock the driver itself in an integration test.
- Run via Mage targets, not raw `go test`, when available: `mage test`, `mage integration`, `mage coverage`, `mage race`. (`go.yaml` defines the underlying package sets/args/timeouts per target.)

## Lint-First: Check `.golangci.yml` BEFORE Writing Code, Not After

`.golangci.yml` (repo root) is the source of truth. Read it before generating
code so the first version you write already passes, instead of writing code
and then chasing lint errors. Currently enabled, with the gotchas this
codebase has actually hit:

- **`nlreturn`** — every `return` (or `continue`/`break`) that is *not* the
  first statement in its block needs a blank line immediately before it.
  This is the single most common miss — write it correctly the first time:
  ```go
  if err != nil {
      writeJSON(w, http.StatusBadRequest, errorResponse{Error: "..."})

      return
  }
  ```
- **`mnd`** (magic number detection) — no bare numeric literals for
  thresholds/limits/sizes; use a named `const` (for fixed values) or a
  config field (for anything env-dependent — see the Config section above).
  `ignored-functions` already covers `time.Duration`, `make`, `fmt.Sprintf`,
  `float64`/`float32`, `v.SetDefault`, `time.NewTicker`.
- **`funlen`** — max 120 lines / 60 statements per function. **`cyclop`** —
  max cyclomatic complexity 15. **`gocognit`** — cognitive complexity cap.
  Extract a helper before you hit these, don't wait for the lint failure.
- **`lll`** — 140 char line length limit.
- **`gosec`** — no weak crypto, no unvalidated file paths from untrusted
  input, no command injection; annotate a deliberate, reviewed exception
  with `//nolint:gosec // <reason>` (see `loader.go`'s `mergeYAML` for the
  pattern) rather than restructuring code to dodge the rule.
- **`bodyclose` / `sqlclosecheck` / `rowserrcheck` / `noctx` / `contextcheck`**
  — every `http.Response.Body`, `sql.Rows`, DB cursor must be closed
  (`defer func() { _ = cur.Close(ctx) }()` pattern, see
  `mongodb/audit.go::List`); every outbound call takes a `context.Context`
  that's actually propagated, not `context.Background()` invented mid-stack.
- **`revive`** — `exported` rule is disabled (no mandatory doc-comment lint
  failure), but write the doc comment anyway; it's still the convention
  throughout this codebase.
- **`dupl` / `unparam` / `unconvert` / `wastedassign` / `ineffassign` /
  `unused`** — no copy-pasted blocks above the duplication threshold, no
  unused params/results, no redundant type conversions, no dead
  assignments.
- **`godot`** — comments must end in a period.
- **`whitespace` / `gofmt` (simplify) / `goimports`** — formatting is
  enforced; run `gofmt -l .` before considering anything done. `goimports`
  groups: stdlib → third-party → this module's `github.com/vinaycharlie01/shroute/...`.
- Exclusion: `nava/.*` is excluded from linting (vendored tooling).

**Before declaring any change complete**, run, in order:
```bash
go build ./...
go vet ./...
gofmt -l .                       # must be empty
golangci-lint run --timeout=10m  # must report 0 issues
go test ./...
go test -tags=integration ./backend/test/integration/...   # when Docker is available
```

**Local environment note**: if `golangci-lint run` fails with *"the Go
language version (go1.2X) used to build golangci-lint is lower than the
targeted Go version"*, the preinstalled binary predates `go.mod`'s `go`
directive. Rebuild it against the right toolchain before trusting its
output:
```bash
GOTOOLCHAIN=go<version-from-go.mod> go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
$(go env GOPATH)/bin/golangci-lint run --timeout=10m
```

## Git Workflow

- Designated branch per the session's instructions — never commit straight to `main`.
- Conventional, descriptive commit messages; no AI/bot attribution in commit metadata.
- A PR/commit that changes `backend/` code without an accompanying test (unit and, for new outbound adapters, integration) is incomplete.
