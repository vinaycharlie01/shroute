# .clinerules

Condensed rules for Cline working in this repository. The full, detailed
version of this ruleset lives in `CLAUDE.md` at the repo root — keep both in
sync if you change one. These rules OVERRIDE default behavior.

## Repo Map

- `backend/` — **the product**. Go service, Hexagonal Architecture. Everything below applies here.
- `frontend/` — new Go-backend's frontend (scaffold stage).
- `backend/tasks/` — per-feature migration task files, each an 8-step end-to-end-flow template (`00-INDEX.md` is the index).
- `OmniRoute/` — **reference only**, the original Next.js/TS app being migrated from. It has its own `CLAUDE.md` with TS-specific rules that apply *only inside that subtree*. Never apply Go rules there; never port its rules here.
- `nava/` — vendored Mage build tooling, reused/extended not rewritten. Excluded from `.golangci.yml` linting.

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

One-way dependency direction: `adapters` → `application` → `domain`. `domain`
never imports `application`; `application` never imports `adapters`. Only
`internal/infrastructure/di/container.go` is allowed to import both sides.

Adding a feature = extending this structure, never bypassing it. Follow the
8-step flow (`backend/tasks/00-INDEX.md`) for every new slice:

1. **Domain** — structs, enums, validation, stdlib only.
2. **Port** (`internal/application/ports/ports.go`) — a narrow, capability-named interface, not a god-interface. Split persistence from external-call concerns into separate ports when both exist.
3. **Application service** — depends only on `ports` interfaces via constructor injection (`NewService(repo ports.XRepository) *Service`). Validates via the domain type, wraps errors with `fmt.Errorf("<pkg>: <op>: %w", err)`.
4. **Outbound adapter** (`internal/adapters/outbound/<dep>/`) — implements the port, owns its own wire-format structs, never leaks them past the boundary.
5. **Inbound handler** — declares its OWN minimal local interface for the service it needs, never imports the concrete `application/<feature>` package (Interface Segregation in practice — see `health.go`'s `healthChecker`, `audit.go`'s `auditRecorder`).
6. **Router wiring** (`router.go`) — add the field to `RouterConfig`, nil-guard if optional.
7. **DI wiring** (`container.go`) — construct adapter → service → handler inside the relevant `if cfg.X.Enabled` block.
8. **Tests** — unit per layer + one integration test per outbound adapter.

Reference implementation to copy conventions from: the **audit log slice**
(`internal/domain/audit/`, `internal/application/audit/`,
`internal/adapters/outbound/mongodb/audit.go`,
`internal/adapters/inbound/http/handlers/audit.go`, plus `router.go`/`container.go` wiring).

## SOLID, Applied Here

- **SRP** — one package per concern (`domain`=rules, `application`=orchestration, `adapters`=I/O).
- **OCP** — add a feature via new `domain`/`application`/`adapters` packages + a new `ports` interface; don't special-case existing services.
- **LSP** — every adapter implementing a port must be fully substitutable, no adapter-specific behavior callers need to know about.
- **ISP** — handlers declare their own narrow local interface; ports stay narrow (2-4 methods), not aggregated god-interfaces.
- **DIP** — `application` depends on `ports` it owns, never on concrete `adapters/outbound/*`. Only `infrastructure/di` wires both sides.

## Go Design Patterns

- Constructor injection: `NewXxx(deps ...)` returning `*Xxx`; dependencies passed in, never constructed inside. Variadic constructors for "zero or more of this dependency type" (see `healthapp.NewService`).
- Adapter pattern: every `internal/adapters/outbound/<dep>/` package adapts a third-party client to a `ports` interface.
- Strategy via interfaces: ports are the strategy contract; swapping stores means a new adapter, zero changes to `application`/`domain`.
- Sentinel errors for domain validation (`internal/domain/<feature>/errors.go`), checked with `errors.Is(...)` at the handler boundary to decide 400 vs 500 — never return raw adapter/driver errors to an HTTP client.
- Functional options only once a constructor needs more than ~3-4 optional params; otherwise a plain typed `Config` struct (YAML-mappable for free) is simpler.

## Config: YAML-First, Dynamic Over Constant

- All environment-shaped config lives in `backend/config/config.{base,local,development,staging,production}.yaml`, layered by `internal/infrastructure/config/loader.go`.
- Env-var overrides are automatic via `env:"APP_..."` struct tags on `Config` fields (`internal/infrastructure/config/config.go`), loaded by a single `env.Parse(cfg)` call (`github.com/caarlos0/env/v11`) in `loader.go`. **Adding override support for a new field is a one-line tag addition — never hand-write another `os.Getenv` branch.** The only exception is cross-field relational logic that can't live on one field's tag (e.g. `APP_MONGO_URI` implies `Mongo.Enabled = true`), which stays as small, explicitly commented code in `applyEnvOverrides`.
- Prefer config/env-driven values over `const`. A `const` is only for values fixed by the code's own logic that could never sensibly vary per deployment. Anything that could differ between local/dev/staging/prod belongs in `Config`.
- Validation lives on `Config` via `validator/v10` tags, not hand-rolled in `Load()`.
- Optional adapters (Mongo, Redis, future) follow `Enabled bool` + `if cfg.X.Enabled { ... }` in `container.go` — the foundation must run with zero external dependencies until a feature needs one.

## Logging: `log/slog` Only

- No custom logger wrapper, no logger struct fields threaded through constructors. Use `slog.Default()` / `slog.InfoContext(ctx, ...)`.
- Always pass `ctx` to `*Context` slog variants when available; structured fields (`slog.String(...)`), never `fmt.Sprintf` a log line.
- Never log a raw secret/credential/full request body. Client-facing errors are generic, internal logs carry the detail.

## Testing: TDD, Table-Driven, Minimal-but-90%

- Write the failing test first when fixing a bug or adding a use-case branch, then make it pass.
- Table-driven is the default shape for any function with more than one meaningful case:
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
  `t.Parallel()` at the top of both the outer test func and every `t.Run` subtest.
- Keep cases minimal, not exhaustive: cover every branch/decision point once and stop. Target **90%+ statement coverage through meaningful branch coverage**, not padding.
- Unit tests per layer: domain (pure-function tests), application (service tests against a hand-written fake of the `ports` interface), inbound handler (stub of the handler's local interface + `httptest.NewRecorder()`).
- **Integration tests are mandatory for every outbound adapter.** Live in `backend/test/integration/`, gated by `//go:build integration`, using `backend/test/containers` testcontainer helpers against a real instance — never mock the driver itself in an integration test.
- Run via Mage targets when available: `mage test`, `mage integration`, `mage coverage`, `mage race`.

## Lint-First: Check `.golangci.yml` BEFORE Writing Code

Read `.golangci.yml` (repo root) before generating code so the first version
already passes. Enabled linters and the gotchas this codebase has hit:

- **`nlreturn`** — every `return`/`continue`/`break` not first in its block needs a blank line immediately before it.
- **`mnd`** — no bare numeric literals for thresholds/limits/sizes; use a named `const` (fixed values) or a config field (env-dependent values).
- **`funlen`** (120 lines / 60 statements), **`cyclop`** (complexity 15), **`gocognit`** — extract a helper before hitting these.
- **`lll`** — 140 char line length.
- **`gosec`** — no weak crypto, no unvalidated file paths from untrusted input, no command injection; annotate deliberate exceptions with `//nolint:gosec // <reason>`.
- **`bodyclose`/`sqlclosecheck`/`rowserrcheck`/`noctx`/`contextcheck`** — close every `Response.Body`/`sql.Rows`/cursor; propagate a real `context.Context`, never invent `context.Background()` mid-stack.
- **`revive`** — `exported` disabled, but write the doc comment anyway.
- **`dupl`/`unparam`/`unconvert`/`wastedassign`/`ineffassign`/`unused`** — no copy-paste blocks, unused params/results, redundant conversions, dead assignments.
- **`godot`** — comments end in a period.
- **`whitespace`/`gofmt`/`goimports`** — formatting enforced; `goimports` groups stdlib → third-party → `github.com/vinaycharlie01/shroute/...`.
- `nava/.*` is excluded from linting (vendored tooling).

Before declaring any change complete, run in order:
```bash
go build ./...
go vet ./...
gofmt -l .                       # must be empty
golangci-lint run --timeout=10m  # must report 0 issues
go test ./...
go test -tags=integration ./backend/test/integration/...   # when Docker is available
```

If `golangci-lint run` complains the Go version it was built with is lower
than `go.mod`'s `go` directive, rebuild it: `GOTOOLCHAIN=go<version> go
install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.

## Git Workflow

- Designated branch per the session's instructions — never commit straight to `main`.
- Conventional, descriptive commit messages; no AI/bot attribution in commit metadata.
- A change to `backend/` code without an accompanying test (unit and, for new outbound adapters, integration) is incomplete.
