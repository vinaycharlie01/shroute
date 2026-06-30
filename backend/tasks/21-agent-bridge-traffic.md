# Task 21: Agent Bridge & Traffic Inspector

**Complexity**: Highest in the migration — MITM-style traffic interception
and certificate installation (`src/mitm/cert/install.ts` on the TS side),
combined with cross-process "agent bridge" tooling. Do this last: it
depends on the loopback-only middleware (Task 15), the env-based
process-spawning discipline (Tasks 15/17), and benefits from every other
port/adapter convention already being settled.

**TS source**: `OmniRoute/vinaydoc/SLICE_20_AGENT_BRIDGE_TOOLS.md` —
`/api/tools/agent-bridge/*`, `/api/tools/traffic-inspector/*`,
`/api/middleware/*`. Cross-reference `docs/security/STEALTH_GUIDE.md` and
`src/mitm/` in the OmniRoute TS repo — the TLS/fingerprint and cert-install
behavior here is the most security-sensitive code in the entire OmniRoute
codebase; do not improvise around the documented contract.

## End-to-end flow

1. **Domain** — `internal/domain/traffic/traffic.go`: `Capture{ID, Method,
   URL string, RequestHeaders, ResponseHeaders map[string][]string,
   StatusCode int, CapturedAt time.Time}` (bodies stored via the `BlobStore`
   port from Task 11, not inline — these can be large). `internal/domain/agentbridge/agentbridge.go`:
   `BridgeSession{ID, AgentID string, Status SessionStatus}`.
2. **Ports** — `TrafficCaptureRepository` (metadata CRUD, body via
   `BlobStore`), `CertInstaller` (`Install(ctx) error`, `IsInstalled(ctx)
   (bool, error)` — the actual NSS-database/system-trust-store
   manipulation), `AgentBridgeRunner` (`Start/Stop(ctx, s
   BridgeSession) error`) in `ports.go`.
3. **Application** — `internal/application/traffic/service.go`: persists
   captures as they arrive from the inspector adapter, supports filtered
   queries. `internal/application/agentbridge/service.go`: session
   lifecycle, delegating to `AgentBridgeRunner`.
4. **Outbound adapters** — `internal/adapters/outbound/mongodb/traffic.go`
   for capture metadata (bodies via the Task 11 `BlobStore`); new
   `internal/adapters/outbound/mitm/` package implementing `CertInstaller`
   and the actual traffic-interception proxy — this is the one adapter in
   the whole migration where **the implementation detail matters as much
   as the interface**: follow `src/mitm/cert/install.ts::updateNssDatabases`
   exactly for the "runtime values via `env`, never shell-interpolated"
   rule (CLAUDE.md hard rule #13 names this file specifically as the
   reference pattern), and do not weaken any TLS verification step from the
   TS implementation when porting.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/{traffic,agentbridge}.go`:
   `GET /api/tools/traffic-inspector/captures`,
   `POST /api/tools/agent-bridge/sessions`. **Classify under
   `middleware.LoopbackOnly`** (Task 15) — this slice can install system
   certificates and intercept traffic, the highest-privilege capability in
   the migration, so it must never be reachable over a tunnel even with a
   valid JWT.
6. **Router/DI** — usual extension pattern, loopback-only scoped.
7. **Tests** — unit tests for `CertInstaller` argument construction
   (asserting no shell interpolation, same style as the Task 15/17 tests);
   integration test installing a cert into a disposable NSS database in a
   container and verifying it's trusted; a dedicated review/test
   confirming the loopback-only middleware actually rejects a
   forwarded/tunneled request to every route in this handler — this is the
   single most important test in the whole migration to get right.

## Checklist

- [ ] `internal/domain/traffic`, `internal/domain/agentbridge`
- [ ] `TrafficCaptureRepository`, `CertInstaller`, `AgentBridgeRunner` ports
- [ ] Application services + unit tests
- [ ] `mitm` adapter (env-based, no shell interpolation, TLS verification preserved) + integration test (disposable NSS db)
- [ ] Handlers + router wiring, strictly loopback-only
- [ ] Dedicated loopback-rejection test for every route in this handler
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
