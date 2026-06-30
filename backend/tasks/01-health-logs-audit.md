# Task 01: Health, Logs & Audit (extend the foundation slice)

**Complexity**: Easiest — the domain/application/handler/router/DI pattern
already exists for `/healthz`/`/readyz`; this task only adds persisted
history on top of it (audit trail + request/error logs), no new business
logic.

**TS source**: `OmniRoute/vinaydoc/SLICE_07_HEALTH.md` — `/api/health/*`,
`/api/audit`, `/api/logs/*`, `/api/monitoring`. Use it for field-level
request/response shapes only (its `audit_log`/`request_logs`/`system_events`
SQL tables become Mongo collections below).

## End-to-end flow

1. **Domain** — new `internal/domain/audit/audit.go`: `Entry{ID, Actor,
   Action, Target, Metadata map[string]any, CreatedAt time.Time}` and
   `internal/domain/syslog/syslog.go`: `Record{ID, Level, Message, Source,
   CreatedAt}`. Keep `internal/domain/health/health.go` untouched.
2. **Ports** — add to `internal/application/ports/ports.go`:
   ```go
   type AuditRepository interface {
       Append(ctx context.Context, e audit.Entry) error
       List(ctx context.Context, limit int) ([]audit.Entry, error)
   }
   type LogRepository interface {
       Append(ctx context.Context, r syslog.Record) error
       List(ctx context.Context, limit int) ([]syslog.Record, error)
   }
   ```
3. **Application** — `internal/application/audit/service.go`,
   `internal/application/syslog/service.go`: thin orchestration over the
   ports, mirroring `internal/application/health/service.go`'s shape
   (constructor takes the port, methods delegate with light validation).
4. **Outbound adapter** — new file `internal/adapters/outbound/mongodb/audit.go`
   and `.../mongodb/syslog.go`, both methods on a small wrapper struct that
   takes `*mongo.Collection` from the existing `mongodb.Adapter.Client()` —
   do not add a second `mongodb.New()` call anywhere. Collections:
   `audit_log` (capped or TTL-indexed on `created_at`), `system_logs`
   (TTL-indexed, e.g. 30 days) — create indexes once in a small
   `EnsureIndexes(ctx)` helper called from `di.New`.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/audit.go`
   and `.../handlers/syslog.go`, same local-interface pattern as
   `healthChecker` in `health.go`.
6. **Router** — extend `RouterConfig` in `router.go` with `Audit *handlers.Audit`
   and `Logs *handlers.Syslog`; add `GET /api/audit`, `GET /api/logs`.
7. **DI** — wire both in `container.go` next to `healthService`/`healthHandler`,
   reusing the same `mongo.Adapter` instance already in `deps`.
8. **Tests** — table-driven unit tests with a fake `AuditRepository`/
   `LogRepository`; integration test in `backend/test/integration/audit_test.go`
   using `containers.MongoDB(ctx, t)`, modeled on `mongodb_test.go`.

## Checklist

- [ ] `internal/domain/audit`, `internal/domain/syslog`
- [ ] Ports added to `ports.go`
- [ ] `internal/application/{audit,syslog}/service.go` + unit tests
- [ ] `internal/adapters/outbound/mongodb/{audit,syslog}.go` + integration test
- [ ] Handlers + router wiring
- [ ] DI wiring in `container.go`
- [ ] `go build ./...`, `go vet ./...`, `gofmt -l .`, `golangci-lint run`, `go test ./...`
