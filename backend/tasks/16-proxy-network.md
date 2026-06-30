# Task 16: Proxy & Network

**Complexity**: High — proxy chains, tunnel lifecycle, and network
introspection involve long-lived connections and OS-level networking, a
different failure-mode class from request/response CRUD.

**TS source**: `OmniRoute/vinaydoc/SLICE_13_PROXY_NETWORK.md` —
`/api/settings/proxies/*`, `/api/settings/free-proxies/*`,
`/api/settings/proxy/*`, `/api/tunnels/*`, `/api/network/info`,
`/api/settings/oneproxy/*`, `/api/upstream-proxy/*`.

## End-to-end flow

1. **Domain** — `internal/domain/proxy/proxy.go`: `Proxy{ID, URL string,
   Type ProxyType, Active bool, LastChecked time.Time, Healthy bool}`,
   `internal/domain/tunnel/tunnel.go`: `Tunnel{ID, LocalAddr, RemoteAddr
   string, Status TunnelStatus}`.
2. **Ports** — `ProxyRepository` (CRUD), `ProxyChecker`
   (`Check(ctx, p proxy.Proxy) (bool, error)` — reuses the
   probe/repository split pattern from Task 07), `TunnelRepository` (CRUD),
   `TunnelManager` (`Open(ctx, t tunnel.Tunnel) error`, `Close(ctx, id
   string) error` — manages actual long-lived connections) in `ports.go`.
3. **Application** — `internal/application/proxy/service.go`: periodic or
   on-demand health checks via `ProxyChecker`, persists results.
   `internal/application/tunnel/service.go`: `Open` validates no duplicate
   `LocalAddr` binding before delegating to `TunnelManager`, tracks status
   transitions.
4. **Outbound adapters** — `internal/adapters/outbound/mongodb/{proxy,tunnel}.go`
   for metadata; `internal/adapters/outbound/proxyhttp/checker.go`
   implementing `ProxyChecker` (HTTP request through the candidate proxy
   with a bounded timeout); new `internal/adapters/outbound/tunnelmgr/`
   package implementing `TunnelManager` over `net` (the actual
   listener/dialer lifecycle), with explicit context cancellation wired to
   `Close` so tunnels never leak goroutines.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/{proxy,tunnel}.go`:
   CRUD + `POST /api/tunnels/{id}/open`, `POST /api/tunnels/{id}/close`,
   `GET /api/network/info`.
6. **Router/DI** — usual extension pattern; `TunnelManager`'s open tunnels
   must be closed in `di.Container.Close()` alongside the other closers
   (register it as a `ports.Closer`).
7. **Tests** — unit tests for duplicate-binding validation and status
   transitions; integration test opening a real loopback tunnel and
   asserting it closes cleanly on `Container.Close()` (no leaked
   goroutines/listeners — check via `go test -race`).

## Checklist

- [ ] `internal/domain/proxy`, `internal/domain/tunnel`
- [ ] `ProxyRepository`, `ProxyChecker`, `TunnelRepository`, `TunnelManager` ports
- [ ] Application services + unit tests
- [ ] Mongo metadata adapters + `proxyhttp` + `tunnelmgr` adapters + integration tests (incl. `-race`)
- [ ] Handlers + router wiring
- [ ] DI wiring; `TunnelManager` registered as `ports.Closer`
- [ ] Full gate: build/vet/fmt/lint/test
