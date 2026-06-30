# Task 17: CLI Tools

**Complexity**: High — manages child-process lifecycle for arbitrary
external CLI tools across platforms (start/stop/stream output), which
combines the process-spawning risk of Task 15 with broader, less-controlled
external surfaces (`/api/cli-tools/*`, `/api/cli/*` cover many distinct
tools, not one well-known DB-backup binary).

**TS source**: `OmniRoute/vinaydoc/SLICE_12_CLI_TOOLS.md` —
`/api/cli-tools/*` + `/api/cli/*`. Cross-reference
`src/lib/services/installers/` and `src/lib/services/bootstrap.ts` in the
OmniRoute TS repo (the "Adding a New Embedded Service" convention) for the
install/start/stop/restart/update/status/logs lifecycle this slice mirrors.

## End-to-end flow

1. **Domain** — `internal/domain/clitool/clitool.go`: `Tool{ID, Name,
   BinaryPath, Version string, Status ToolStatus}`, `ToolStatus` enum
   (`NotInstalled`/`Installed`/`Running`/`Stopped`/`Failed`).
2. **Ports** — `CliToolRepository` (status/metadata CRUD), `ToolRunner`
   (`Install/Start/Stop/Restart/Update(ctx, t clitool.Tool) error`,
   `Logs(ctx, t clitool.Tool) (io.ReadCloser, error)`) in `ports.go` — one
   port covering the full lifecycle, matching the 7-endpoint pattern
   (`install/start/stop/restart/update/status/logs`) the TS embedded-services
   convention already established, just ported to Go signatures.
3. **Application** — `internal/application/clitool/service.go`: enforces
   valid state transitions (e.g. reject `Start` on a `NotInstalled` tool),
   delegates the actual OS interaction to `ToolRunner`.
4. **Outbound adapters** — `internal/adapters/outbound/mongodb/clitool.go`
   for metadata/status; `internal/adapters/outbound/processrunner/` package
   implementing `ToolRunner` via `os/exec` — same hard rule as Task 15:
   binary paths and arguments passed via `exec.Command` argv slices and the
   `env` option, never built as a shell string.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/clitool.go`:
   the standard 7 endpoints per tool (`install/start/stop/restart/update/
   status/logs`) under `/api/cli-tools/{id}/*`. **Classify this entire
   route prefix as loopback-only** in `router.go`, reusing the
   `middleware.LoopbackOnly` built in Task 15 (CLAUDE.md hard rule #15
   names `/api/cli-tools/` explicitly).
6. **Router/DI** — usual extension pattern, with `middleware.LoopbackOnly`
   applied to this handler's prefix exactly as in Task 15.
7. **Tests** — unit tests for state-transition validation; integration test
   running a trivial real binary (e.g. `echo`) through `processrunner` and
   asserting argv-only invocation (no shell metacharacter interpretation).

## Checklist

- [ ] `internal/domain/clitool`
- [ ] `CliToolRepository`, `ToolRunner` ports
- [ ] `internal/application/clitool/service.go` (state-transition validation) + unit tests
- [ ] Mongo adapter + `processrunner` adapter (argv/env-based, no shell interpolation) + integration test
- [ ] Handlers (7-endpoint lifecycle) + router wiring, loopback-only
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
