# Task 15: DevOps & Infra

**Complexity**: High and high-risk — process spawning, DB backup/restore,
system restart/shutdown. CLAUDE.md hard rules #15/#17 (loopback-only
classification for any route that can spawn a child process) apply
directly; this is the slice most likely to introduce a security regression
if rushed.

**TS source**: `OmniRoute/vinaydoc/SLICE_18_DEVOPS_INFRA.md` —
`/api/version-manager/*`, `/api/db-backups/*`, `/api/db/health`,
`/api/sync/*`, `/api/system/*`, `/api/headroom/*`, `/api/monitoring/*`,
`/api/init`, `/api/shutdown`, `/api/restart`. Also the deferred
`/api/settings/db-backups`, `/api/system`, `/api/restart`, `/api/shutdown`,
`/api/version-manager` portion of `SLICE_10_SETTINGS.md` from Task 03.

## End-to-end flow

1. **Domain** — `internal/domain/devops/devops.go`: `BackupJob{ID, Status
   JobStatus, StartedAt, CompletedAt *time.Time, SizeBytes int64}`,
   `VersionInfo{Current, Latest string, AutoStart bool}`.
2. **Ports** — `BackupRepository` (job metadata CRUD), `Backuper`
   (`Run(ctx) (BackupJob, error)`, `Restore(ctx, jobID string) error` — the
   actual mongodump/mongorestore-equivalent invocation), `SystemController`
   (`Restart(ctx) error`, `Shutdown(ctx) error`) in `ports.go`. Every one of
   these ports' adapters spawns a process or affects the whole running
   service — keep them maximally narrow and never let the application
   layer construct a shell command string itself.
3. **Application** — `internal/application/devops/service.go`: orchestrates
   backup scheduling and validates restore requests reference a completed
   backup before calling `Backuper.Restore`; restart/shutdown calls are
   pass-through to `SystemController` with no extra logic (the safety
   boundary is the HTTP-layer loopback check in step 5, not this layer).
4. **Outbound adapter** — `internal/adapters/outbound/mongodb/devops.go` for
   job metadata; new `internal/adapters/outbound/sysops/` package
   implementing `Backuper`/`SystemController` via `os/exec`, passing every
   runtime value (paths, connection URIs) through the `env` option, never
   string-interpolated into a shell script (CLAUDE.md hard rule #13,
   reference `src/mitm/cert/install.ts::updateNssDatabases` in the TS repo
   for the established pattern to mirror).
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/devops.go`:
   `POST /api/db-backups`, `POST /api/db-backups/{id}/restore`,
   `POST /api/restart`, `POST /api/shutdown`, `GET /api/system/*`. **Every
   route in this handler must be registered behind a loopback-only check**
   in `router.go` (a new `middleware.LoopbackOnly` mirroring
   `isLocalOnlyPath()`/`routeGuard.ts` from the TS repo) evaluated
   unconditionally before any auth check, exactly matching CLAUDE.md hard
   rules #15/#17.
6. **Router/DI** — add `middleware.LoopbackOnly` to the chain in
   `router.go`, scoped to this handler's route prefix only (not globally —
   other slices' routes must remain reachable non-locally where intended).
7. **Tests** — unit tests for the loopback-only middleware (table-driven:
   loopback request IDs allowed, tunneled/forwarded IDs rejected — port the
   TS `isLocalOnlyPath()` test cases); integration test for backup/restore
   round-trip against `containers.MongoDB`, asserting `os/exec` invocations
   never receive untrusted interpolated strings (review, not just test).

## Checklist

- [ ] `internal/domain/devops`
- [ ] `BackupRepository`, `Backuper`, `SystemController` ports
- [ ] `internal/application/devops/service.go` + unit tests
- [ ] `sysops` adapter (env-based `os/exec`, no shell interpolation) + integration test
- [ ] `middleware.LoopbackOnly` + unit tests (mirrors `isLocalOnlyPath()`)
- [ ] Handlers + router wiring, scoped loopback-only
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
