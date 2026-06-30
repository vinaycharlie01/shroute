# Task 05: Models & Mappings

**Complexity**: Moderate — read-heavy catalog CRUD, but depends on Provider
data existing first conceptually (a `model.ProviderID` foreign reference).
Build it right after API Keys so the reference shape is decided before
Task 07 (Providers) adds the write side that populates it.

**TS source**: `OmniRoute/vinaydoc/SLICE_05_MODELS.md` — `/api/models/*`,
`/api/model-combo-mappings/*`, `/api/synced-available-models`.

## End-to-end flow

1. **Domain** — `internal/domain/model/model.go`: `Model{ID, ProviderID,
   Name, ContextWindow int, Capabilities []string, Deprecated bool}`,
   `ComboMapping{ModelID, ComboID string, Weight int}`.
2. **Ports** — `ModelRepository` (`List/Get/Upsert/MarkDeprecated`),
   `ModelComboMappingRepository` (`List/Set`) in `ports.go`.
3. **Application** — `internal/application/model/service.go`: list/get pass
   through; `Upsert` validates `ProviderID` references an existing provider
   via a narrow `ProviderExists(ctx, id) (bool, error)` method on the
   `ModelRepository` port (or a tiny separate port) rather than importing
   the provider application package directly — keeps Task 05 buildable
   ahead of Task 07.
4. **Outbound adapter** — `internal/adapters/outbound/mongodb/model.go`:
   `models` collection (index on `provider_id`), `model_combo_mappings`
   collection (compound index on `model_id, combo_id`).
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/model.go`:
   `GET /api/models`, `GET /api/models/{id}`, `GET/PUT /api/model-combo-mappings`,
   `GET /api/synced-available-models`.
6. **Router/DI** — usual extension pattern.
7. **Tests** — unit tests for the deprecation/mapping validation rules;
   integration test for Mongo index uniqueness on the mapping collection.

## Checklist

- [ ] `internal/domain/model`
- [ ] `ModelRepository`, `ModelComboMappingRepository` ports
- [ ] `internal/application/model/service.go` + unit tests
- [ ] Mongo adapter (`models`, `model_combo_mappings`) + integration test
- [ ] Handlers + router wiring
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
