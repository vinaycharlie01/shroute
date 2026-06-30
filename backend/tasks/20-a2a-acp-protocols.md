# Task 20: A2A / ACP Protocols

**Complexity**: Very high — a full JSON-RPC 2.0 multi-agent protocol
implementation (A2A) plus the related ACP/agent-card surface, with task
orchestration semantics. Build after MCP (Task 18) since A2A skills often
delegate to MCP tools internally on the TS side.

**TS source**: `OmniRoute/vinaydoc/SLICE_19_A2A_ACP_PROTOCOLS.md` —
`/api/a2a/*`, `/api/acp/*`, `/api/agents/*`. Cross-reference
`docs/frameworks/A2A-SERVER.md` and `docs/frameworks/AGENT_PROTOCOLS_GUIDE.md`
in the OmniRoute TS repo for the exact JSON-RPC method/error-code contract
(this is a protocol compliance surface, not a place for free interpretation)
and the 5 existing skills (smart-routing, quota-management,
provider-discovery, cost-analysis, health-report) this slice ports.

## End-to-end flow

1. **Domain** — `internal/domain/a2a/a2a.go`: `Task{ID, Status TaskStatus,
   Messages []Message, Result *Result}`, `AgentCard{Name, Description
   string, Skills []SkillDescriptor}` (the `.well-known/agent.json` shape).
2. **Ports** — `A2ASkill` (`Execute(ctx, task Task) (Result, error)` — one
   implementation per skill, registered by name, same registry pattern as
   MCP tools in Task 18), `A2ATaskRepository` (CRUD task state) in
   `ports.go`.
3. **Application** — `internal/application/a2a/service.go`: implements the
   JSON-RPC 2.0 method dispatch table (`tasks/send`, `tasks/get`,
   `tasks/cancel`, etc. — copy the exact method names from
   `A2A-SERVER.md`), looks up the requested skill in the `A2ASkill`
   registry, persists task state transitions via `A2ATaskRepository`.
   `internal/application/a2a/skills/{smart-routing,quota-management,
   provider-discovery,cost-analysis,health-report}.go`: one file per
   skill, each a thin `A2ASkill` implementation calling the corresponding
   already-ported application service (`combo` for smart-routing, `usage`/
   `quota` for quota-management, `provider` for provider-discovery, etc.) —
   this slice mostly **wires together work already done in earlier tasks**
   rather than introducing new domain logic.
4. **Outbound adapter** — `internal/adapters/outbound/mongodb/a2a.go` for
   task persistence.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/a2a.go`:
   `POST /api/a2a` (single JSON-RPC endpoint, method routed internally per
   the spec — not one REST route per method), `GET /.well-known/agent.json`
   serving the `AgentCard`. `internal/adapters/inbound/http/handlers/acp.go`
   for the related ACP surface.
6. **Router/DI** — usual extension pattern; the skill registry is
   constructed in `di.Container` from the already-wired `combo`/`usage`/
   `provider` application services — this is the clearest example in the
   whole migration of the hexagonal architecture paying off (no
   reimplementation, just new adapters over existing application services).
7. **Tests** — JSON-RPC envelope/error-code conformance tests (malformed
   request, unknown method, etc.); one unit test per skill verifying it
   delegates to the correct underlying application service; integration
   test exercising a full `tasks/send` → `tasks/get` round trip.

## Checklist

- [ ] `internal/domain/a2a`
- [ ] `A2ASkill`, `A2ATaskRepository` ports
- [ ] `internal/application/a2a/service.go` (JSON-RPC dispatch table) + conformance unit tests
- [ ] 5 skill implementations wired to existing combo/usage/provider services + unit tests
- [ ] Mongo task-persistence adapter + integration test (`tasks/send`→`tasks/get`)
- [ ] Handlers (`/api/a2a`, `/.well-known/agent.json`, `/api/acp/*`) + router wiring
- [ ] DI wiring (skill registry)
- [ ] Full gate: build/vet/fmt/lint/test
