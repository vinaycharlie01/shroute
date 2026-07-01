# 🎯 Slice 13: Go Backend for Proxy, Network & Tunnel Routes

**Goal**: Migrate proxy configuration, network management, and tunnel endpoints from TypeScript to Go. Covers upstream proxies, egress, Cloudflare/ngrok/Tailscale tunnels, free proxies, 1proxy rotation, and network info.

**Routes**: `/api/settings/proxies/*`, `/api/settings/free-proxies/*`, `/api/settings/proxy/*`, `/api/tunnels/*`, `/api/network/info`, `/api/settings/oneproxy/*`, `/api/upstream-proxy/*`

---

## ✅ TASK 1: Types

**Files to create**: `pkg/types/proxy.go`, `pkg/types/tunnel.go`, `pkg/types/network.go`

```go
type ProxyConfig struct {
    ID       string `json:"id"`
    Type     string `json:"type"`      // "http", "socks5", "ssh"
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username,omitempty"`
    IsActive bool   `json:"is_active"`
    Region   string `json:"region,omitempty"`
    Health   string `json:"health,omitempty"`  // "healthy", "degraded", "down"
}

type ProxyAssignment struct {
    ProxyID    string `json:"proxy_id"`
    ProviderID string `json:"provider_id"`
    ComboID    string `json:"combo_id,omitempty"`
    Priority   int    `json:"priority"`
}

type TunnelConfig struct {
    ID        string `json:"id"`
    Type      string `json:"type"`    // "cloudflared", "ngrok", "tailscale"
    URL       string `json:"url"`
    Status    string `json:"status"`  // "running", "stopped", "error"
    Port      int    `json:"port"`
    CreatedAt string `json:"created_at"`
}

type NetworkInfo struct {
    PublicIP     string `json:"public_ip"`
    Region       string `json:"region"`
    ISP          string `json:"isp"`
    ProxyCount   int    `json:"proxy_count"`
    ActiveTunnels int    `json:"active_tunnels"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create types | ☐ |
| 1.2 | Run `go build` | ☐ |

---

## ✅ TASK 2: Repositories

**Files to create**: `internal/db/proxy.go`, `internal/db/tunnel.go`

```go
type ProxyRepo struct { db *sql.DB }
func (r *ProxyRepo) List() ([]types.ProxyConfig, error)
func (r *ProxyRepo) Assign(assignment *types.ProxyAssignment) error
func (r *ProxyRepo) BulkAssign(assignments []types.ProxyAssignment) error
func (r *ProxyRepo) BulkImport(proxies []types.ProxyConfig) error
func (r *ProxyRepo) MigrateProvider(from, to string) error
func (r *ProxyRepo) CheckHealth() ([]types.ProxyConfig, error)
func (r *ProxyRepo) GetEgress() (*types.ProxyConfig, error)
// Free proxies sub-routes
func (r *FreeProxyRepo) List() ([]types.ProxyConfig, error)
func (r *FreeProxyRepo) AddToPool(id string) error
func (r *FreeProxyRepo) BulkAddToPool(ids []string) error
func (r *FreeProxyRepo) GetStats() (*FreeProxyStats, error)
func (r *FreeProxyRepo) Sync() error

type TunnelRepo struct { db *sql.DB }
func (r *TunnelRepo) GetConfig(tunnelType string) (*types.TunnelConfig, error)
func (r *TunnelRepo) Start(tunnelType string) error
func (r *TunnelRepo) Stop(tunnelType string) error
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement proxy CRUD | ☐ |
| 2.2 | Implement tunnel config | ☐ |
| 2.3 | Write tests | ☐ |

---

## ✅ TASK 3: Handlers

**Files**: `api/handlers/proxy.go`, `api/handlers/tunnel.go`, `api/handlers/network.go`

| Route | Handler | Done |
|-------|---------|------|
| `GET/PUT /api/settings/proxies` | `ListProxies` / `UpdateProxies` | ☐ |
| `POST /api/settings/proxies/assignments` | `AssignProxy` | ☐ |
| `POST /api/settings/proxies/bulk-assign` | `BulkAssign` | ☐ |
| `POST /api/settings/proxies/bulk-import` | `BulkImport` | ☐ |
| `POST /api/settings/proxies/egress` | `SetEgressProxy` | ☐ |
| `GET /api/settings/proxies/health` | `CheckProxyHealth` | ☐ |
| `POST /api/settings/proxies/migrate` | `MigrateProviderProxy` | ☐ |
| `PATCH /api/settings/oneproxy` | `UpdateOneProxy` | ☐ |
| `POST /api/settings/oneproxy/rotate` | `RotateOneProxy` | ☐ |
| `GET /api/settings/proxy` | `GetProxyConfig` | ☐ |
| `POST /api/settings/proxy/test` | `TestProxy` | ☐ |
| `POST /api/settings/proxy/cloudflare-deploy` | `DeployCloudflare` | ☐ |
| `POST /api/settings/proxy/deno-deploy` | `DeployDeno` | ☐ |
| `POST /api/settings/proxy/vercel-deploy` | `DeployVercel` | ☐ |
| `GET /api/settings/free-proxies` | `ListFreeProxies` | ☐ |
| `POST /api/settings/free-proxies/:id/add-to-pool` | `AddToPool` | ☐ |
| `POST /api/settings/free-proxies/bulk-add-to-pool` | `BulkAddToPool` | ☐ |
| `GET /api/settings/free-proxies/stats` | `FreeProxyStats` | ☐ |
| `POST /api/settings/free-proxies/sync` | `SyncFreeProxies` | ☐ |
| `GET/POST /api/tunnels/cloudflared` | `GetCloudflareTunnel` / `Start` | ☐ |
| `GET/POST /api/tunnels/ngrok` | `GetNgrokTunnel` / `Start` | ☐ |
| `POST /api/tunnels/tailscale/enable` | `EnableTailscale` | ☐ |
| `POST /api/tunnels/tailscale/disable` | `DisableTailscale` | ☐ |
| `POST /api/tunnels/tailscale/install` | `InstallTailscale` | ☐ |
| `POST /api/tunnels/tailscale/login` | `TailscaleLogin` | ☐ |
| `POST /api/tunnels/tailscale/start-daemon` | `StartTailscaleDaemon` | ☐ |
| `GET /api/tunnels/tailscale/check` | `CheckTailscale` | ☐ |
| `GET /api/network/info` | `GetNetworkInfo` | ☐ |
| `GET/PUT /api/upstream-proxy/:providerId` | `GetUpstreamProxy` / `SetUpstream` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes | ☐ |
| 3.2 | `curl localhost:8080/api/network/info` | ☐ |
| 3.3 | `curl localhost:8080/api/settings/proxies` | ☐ |
| 3.4 | `curl localhost:8080/api/tunnels/cloudflared` | ☐ |

---

## ✅ TASK 4: Sidecar + Tests + Deploy

| # | Step | Done |
|---|------|------|
| 4.1 | Update nginx | ☐ |
| 4.2 | Integration tests | ☐ |
| 4.3 | `go test ./...` | ☐ |
| 4.4 | Deploy | ☐ |

---

## 🚀 QUICK START

```bash
cd omniroute-go && go run .
curl localhost:8080/api/network/info
curl localhost:8080/api/settings/proxies
curl localhost:8080/api/tunnels/cloudflared
curl localhost:8080/api/settings/free-proxies/stats