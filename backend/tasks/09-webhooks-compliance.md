# Task 09: Webhooks, Compliance & Guardrails

**Complexity**: High-ish — introduces async event dispatch (webhook
delivery with retries) and a rules engine (guardrails/compliance checks),
the first slice where "side effects fan out to other systems" rather than
just CRUD + one outbound probe.

**TS source**: `OmniRoute/vinaydoc/SLICE_17_WEBOOKS_COMPLIANCE.md` —
`/api/webhooks/*`, `/api/compliance/*`, `/api/guardrails/*`, `/api/tags`,
`/api/policies`.

## End-to-end flow

1. **Domain** — `internal/domain/webhook/webhook.go`: `Endpoint{ID, URL
   string, Secret []byte, Events []string, Active bool}`, `Delivery{ID,
   EndpointID string, Payload []byte, Status DeliveryStatus, Attempts int}`.
   `internal/domain/guardrail/guardrail.go`: `Policy{ID, Name string, Rules
   []Rule}` — keep PII-specific rules referencing the opt-in flags from Task
   03, not duplicating their default logic here (Hard Rule #20 stays owned
   by `internal/domain/settings`).
2. **Ports** — `WebhookRepository` (CRUD), `WebhookDispatcher`
   (`Deliver(ctx, endpoint, payload) error` — the actual HTTP POST + HMAC
   signing), `GuardrailRepository` (CRUD policies) in `ports.go`. Splitting
   `WebhookRepository` (persistence) from `WebhookDispatcher` (delivery)
   mirrors the `ProviderRepository`/`ProviderProbe` split in Task 07.
3. **Application** — `internal/application/webhook/service.go`: on a
   triggering event, looks up matching endpoints, calls `WebhookDispatcher`,
   records the `Delivery` outcome, retries failed deliveries with backoff
   (reuse the exponential-backoff shape already established for connection
   cooldown in the TS resilience layer, ported as a small
   `internal/domain/webhook/backoff.go` helper). `internal/application/guardrail/service.go`:
   evaluates a `Policy` against a request payload, returns pass/violation.
4. **Outbound adapters** — `internal/adapters/outbound/mongodb/{webhook,guardrail}.go`
   for persistence; new `internal/adapters/outbound/webhookhttp/dispatcher.go`
   implementing `WebhookDispatcher` (HMAC-SHA256 signs the payload with the
   endpoint's `Secret`, bounded timeout, no retry loop inside the adapter —
   retries are an application-layer concern per step 3).
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/webhook.go`,
   `.../handlers/guardrail.go`: CRUD + `GET /api/webhooks/{id}/deliveries`.
6. **Router/DI** — usual extension pattern; `di.Container` wires the
   dispatcher with a configurable timeout.
7. **Tests** — unit tests for HMAC signing and retry/backoff scheduling
   (table-driven, fake clock); integration test using `httptest.Server` as
   the webhook receiver to assert signature + payload correctness.

## Checklist

- [ ] `internal/domain/webhook`, `internal/domain/guardrail`
- [ ] `WebhookRepository`, `WebhookDispatcher`, `GuardrailRepository` ports
- [ ] Application services + unit tests (incl. retry/backoff, HMAC)
- [ ] Mongo adapters + `webhookhttp` dispatcher adapter + integration tests
- [ ] Handlers + router wiring
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
