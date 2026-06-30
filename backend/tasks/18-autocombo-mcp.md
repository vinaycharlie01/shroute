# Task 18: Auto-Combo & MCP

**Complexity**: Very high — Auto-Combo is a 9-factor scoring engine over
live provider/model/usage state (reads everything from Tasks 05-10), and
MCP is a full protocol server (87 tools, 3 transports: stdio/SSE/Streamable
HTTP, 30 scopes) — the largest single surface in the migration.

**TS source**: `OmniRoute/vinaydoc/SLICE_09_AUTOCOMBO_MCP.md` —
`/api/auto-combo/*`, `/api/mcp/*`, `/api/mcp/sse`, `/api/mcp/stream`.
Cross-reference `docs/routing/AUTO-COMBO.md` (9-factor scoring) and
`docs/frameworks/MCP-SERVER.md` (tool/scope/transport contract) in the
OmniRoute TS repo — port the scoring formula and protocol framing exactly,
this is not a place to improvise.

## End-to-end flow — Auto-Combo

1. **Domain** — `internal/domain/autocombo/autocombo.go`: `Score{ComboID
   string, Factors [9]float64, Total float64}`, factor names as named
   constants (latency, cost, reliability, etc. — copy the exact 9 from
   `AUTO-COMBO.md`).
2. **Ports** — reuse `ComboRepository`/`UsageRepository`/`ProviderRepository`
   from Tasks 10/06/07 as read-only inputs; no new persistence port needed,
   scoring is computed on demand.
3. **Application** — `internal/application/autocombo/service.go`: pure
   scoring function over the 9 factors (heavily unit-tested against the TS
   reference values), `Suggest(ctx) ([]Score, error)` ranks combos.
4. **Inbound handler** — `internal/adapters/inbound/http/handlers/autocombo.go`:
   `GET /api/auto-combo/suggestions`, `GET /api/auto-combo/scores`.

## End-to-end flow — MCP

5. **Domain** — `internal/domain/mcp/mcp.go`: `Tool{Name, Description
   string, Schema json.RawMessage, Scopes []string}`, `Session{ID, Transport
   TransportType, AuthScopes []string}`.
6. **Ports** — `McpToolRegistry` (`List/Get(ctx, name)`), `McpAuditRepository`
   (`Append` — every tool invocation logged, matching the TS `mcp_audit`
   table requirement), `McpSessionRepository` (CRUD) in `ports.go`.
7. **Application** — `internal/application/mcp/service.go`: dispatches a
   tool invocation by name, checks the caller's scopes against
   `Tool.Scopes` before executing, always writes to `McpAuditRepository`
   regardless of outcome (success or rejection) — mirrors "tool invocation
   logged to `mcp_audit` table" from the TS "Adding a New MCP Tool"
   convention.
8. **Outbound adapter** — `internal/adapters/outbound/mongodb/mcp.go` for
   sessions/audit; the 87 individual tool implementations are themselves
   adapters or application-layer functions registered into
   `McpToolRegistry` — build this registry to support incremental addition
   (one tool = one small Go file implementing a `mcp.ToolHandler` func
   type), do not attempt all 87 in one pass.
9. **Inbound transports** — `internal/adapters/inbound/mcp/` (new
   subdirectory, parallel to `inbound/http`): `stdio.go`, `sse.go`,
   `streamhttp.go` — three separate inbound adapters all calling the same
   `internal/application/mcp/service.go`, matching the "3 transports" TS
   requirement without tripling the application logic.
10. **Router/DI** — `GET /api/mcp/sse`, `POST /api/mcp/stream` wired in
    `router.go`; stdio transport wired separately in `cmd/server/main.go`
    (or a dedicated `cmd/mcp-stdio/main.go` if it needs a separate process
    entry point — decide based on how the TS stdio transport is invoked).
11. **Tests** — unit tests for the 9-factor scoring formula (exact value
    parity with TS reference cases) and scope-check rejection paths;
    integration test invoking a real tool through each of the 3 transports.

## Checklist

- [ ] `internal/domain/autocombo`, `internal/domain/mcp`
- [ ] Auto-Combo: scoring service + unit tests (parity with TS reference values)
- [ ] MCP: `McpToolRegistry`, `McpAuditRepository`, `McpSessionRepository` ports
- [ ] MCP application service (scope checks, always-audit) + unit tests
- [ ] MCP tool registry pattern + first few tools ported incrementally
- [ ] 3 inbound transports (stdio/SSE/Streamable HTTP) + router/cmd wiring
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
