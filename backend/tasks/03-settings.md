# Task 03: Settings & Feature Flags

**Complexity**: Easy/moderate — broad surface area (many small config
documents) but each one is plain Mongo CRUD with no external calls; risk is
breadth, not depth. Defer the genuinely risky settings (DB backup/restore,
restart/shutdown, version-manager) to Task 15 (`15-devops-infra.md`).

**TS source**: `OmniRoute/vinaydoc/SLICE_10_SETTINGS.md` — take only the
`/api/settings/*` general config + feature-flag + Qdrant-connection
sub-routes here; its `/api/db-backups`, `/api/system`, `/api/restart`,
`/api/shutdown`, `/api/version-manager` sections belong to Task 15.

## End-to-end flow

1. **Domain** — `internal/domain/settings/settings.go`: `Document{Key string,
   Value bson.Raw, UpdatedAt time.Time}` (settings are heterogeneous
   key→JSON-document pairs, not a fixed struct) plus
   `FeatureFlag{Name string, Enabled bool, DefaultValue string}` — note CLAUDE.md
   Hard Rule #20: any PII-related flag ported from OmniRoute (`PII_REDACTION_ENABLED`,
   `PII_RESPONSE_SANITIZATION`) must keep `Enabled: false` as its seeded default.
2. **Ports** — `SettingsRepository` (`Get/Set/Delete(ctx, key)`) and
   `FeatureFlagRepository` (`Get/List/Set(ctx, name)`) in `ports.go`.
3. **Application** — `internal/application/settings/service.go`: owns
   validation of known setting keys (reject unknown keys rather than
   silently accepting typos — mirrors the Zod-schema discipline from the
   original TS routes).
4. **Outbound adapter** — `internal/adapters/outbound/mongodb/settings.go`:
   single `settings` collection keyed by `_id: key`, plus a `feature_flags`
   collection keyed by `_id: name` seeded with defaults on first run.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/settings.go`:
   `GET/PUT /api/settings/{key}`, `GET /api/settings`,
   `GET/PUT /api/settings/flags/{name}`.
6. **Router/DI** — same pattern as Task 01/02.
7. **Tests** — unit tests per validation rule (especially the PII-flag
   default-false guard, mirroring `tests/unit/pii-opt-in-default.test.ts`
   from the TS side); integration test round-tripping `Set`→`Get`→`Delete`
   against `containers.MongoDB`.

## Checklist

- [ ] `internal/domain/settings`
- [ ] `SettingsRepository` + `FeatureFlagRepository` ports
- [ ] `internal/application/settings/service.go` + unit tests (incl. PII-flag default-false test)
- [ ] Mongo adapter (`settings`, `feature_flags` collections) + integration test
- [ ] Handlers + router wiring
- [ ] DI wiring, with default feature-flag seeding on startup
- [ ] Full gate: build/vet/fmt/lint/test
