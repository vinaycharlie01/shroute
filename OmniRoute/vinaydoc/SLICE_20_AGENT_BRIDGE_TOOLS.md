# 🎯 Slice 20: Go Backend for Agent Bridge & Traffic Inspector

**Goal**: Migrate agent bridge MITM proxy, traffic inspector, and related diagnostic tool endpoints from TypeScript to Go.

**Routes**: `/api/tools/agent-bridge/*`, `/api/tools/traffic-inspector/*`, `/api/middleware/*`

---

## ✅ TASK 1: Types

**Files**: `pkg/types/agent_bridge.go`, `pkg/types/traffic.go`

```go
type AgentBridgeConfig struct {
    Enabled   bool   `json:"enabled"`
    Port      int    `json:"port"`
    Mode      string `json:"mode"` // "proxy", "mitm", "dns"
    CertPath  string `json:"cert_path"`
    TProxyPort int   `json:"tproxy_port,omitempty"`
}

type AgentBridgeAgent struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    UserAgent   string `json:"user_agent"`
    Status      string `json:"status"` // "connected", "disconnected"
    LastSeen    string `json:"last_seen"`
    DNS         string `json:"dns,omitempty"`
    Mappings    map[string]string `json:"mappings,omitempty"`
}

type CapturedRequest struct {
    ID          string `json:"id"`
    Method      string `json:"method"`
    URL         string `json:"url"`
    StatusCode  int    `json:"status_code"`
    RequestSize int64  `json:"request_size"`
    DurationMs  int64  `json:"duration_ms"`
    Timestamp   string `json:"timestamp"`
    SessionID   string `json:"session_id"`
}

type CaptureSession struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Mode      string `json:"mode"` // "http-proxy", "system-proxy", "tls-intercept"
    Status    string `json:"status"` // "active", "stopped"
    CreatedAt string `json:"created_at"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create types | ☐ |

---

## ✅ TASK 2: Repositories

**Files**: `internal/db/agent_bridge.go`, `internal/db/traffic.go`

```go
type AgentBridgeRepo struct { db *sql.DB }
func (r *AgentBridgeRepo) GetConfig() (*types.AgentBridgeConfig, error)
func (r *AgentBridgeRepo) UpdateConfig(cfg *types.AgentBridgeConfig) error
func (r *AgentBridgeRepo) ListAgents() ([]types.AgentBridgeAgent, error)
func (r *AgentBridgeRepo) GetAgent(id string) (*types.AgentBridgeAgent, error)
func (r *AgentBridgeRepo) Diagnose() (map[string]string, error)
func (r *AgentBridgeRepo) Repair() error
func (r *AgentBridgeRepo) GetState() (map[string]interface{}, error)
func (r *AgentBridgeRepo) GetCert() (string, error)
func (r *AgentBridgeRepo) RegenerateCert() error
func (r *AgentBridgeRepo) GetUpstreamCA() (string, error)
func (r *AgentBridgeRepo) TestUpstreamCA() error

type TrafficRepo struct { db *sql.DB }
func (r *TrafficRepo) ListSessions() ([]types.CaptureSession, error)
func (r *TrafficRepo) CreateSession(session *types.CaptureSession) error
func (r *TrafficRepo) StopSession(id string) error
func (r *TrafficRepo) ListRequests(opts RequestFilter) ([]types.CapturedRequest, error)
func (r *TrafficRepo) GetRequest(id string) (*types.CapturedRequest, error)
func (r *TrafficRepo) ReplayRequest(id string) (string, error)
func (r *TrafficRepo) GetHosts() ([]string, error)
func (r *TrafficRepo) GetCaptureModes() ([]string, error)
func (r *TrafficRepo) ExportSession(id string) ([]byte, error)
func (r *TrafficRepo) Ingest(data []byte) error
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement agent bridge repos | ☐ |
| 2.2 | Implement traffic inspector repos | ☐ |

---

## ✅ TASK 3: Handlers (~30 routes)

**Files**: `api/handlers/agent_bridge.go`, `api/handlers/traffic.go`

| Agent Bridge Routes | Handler | Done |
|---------------------|---------|------|
| `GET/PUT /api/tools/agent-bridge/config` | `GetBridgeConfig` / `UpdateBridgeConfig` | ☐ |
| `GET /api/tools/agent-bridge/agents` | `ListBridgeAgents` | ☐ |
| `GET/DELETE /api/tools/agent-bridge/agents/:id` | `GetAgent` / `DeleteAgent` | ☐ |
| `POST /api/tools/agent-bridge/agents/:id/detect` | `DetectAgent` | ☐ |
| `POST /api/tools/agent-bridge/agents/:id/dns` | `AgentDNS` | ☐ |
| `GET /api/tools/agent-bridge/agents/:id/mappings` | `AgentMappings` | ☐ |
| `POST /api/tools/agent-bridge/bypass` | `BypassCheck` | ☐ |
| `GET /api/tools/agent-bridge/cert` | `GetCert` | ☐ |
| `POST /api/tools/agent-bridge/cert/regenerate` | `RegenerateCert` | ☐ |
| `POST /api/tools/agent-bridge/cert/download` | `DownloadCert` | ☐ |
| `POST /api/tools/agent-bridge/diagnose` | `Diagnose` | ☐ |
| `POST /api/tools/agent-bridge/repair` | `Repair` | ☐ |
| `GET /api/tools/agent-bridge/state` | `GetState` | ☐ |
| `GET /api/tools/agent-bridge/upstream-ca` | `GetUpstreamCA` | ☐ |
| `POST /api/tools/agent-bridge/upstream-ca/test` | `TestUpstreamCA` | ☐ |
| `GET /api/tools/agent-bridge/server` | `GetServerConfig` | ☐ |
| `PATCH /api/tools/agent-bridge/tproxy` | `UpdateTProxy` | ☐ |

| Traffic Inspector Routes | Handler | Done |
|--------------------------|---------|------|
| `GET /api/tools/traffic-inspector/sessions` | `ListSessions` | ☐ |
| `POST /api/tools/traffic-inspector/sessions` | `CreateSession` | ☐ |
| `GET /api/tools/traffic-inspector/sessions/:id` | `GetSession` | ☐ |
| `DELETE /api/tools/traffic-inspector/sessions/:id` | `StopSession` | ☐ |
| `GET /api/tools/traffic-inspector/sessions/:id/export` | `ExportSession` | ☐ |
| `GET /api/tools/traffic-inspector/sessions/:id/requests` | `SessionRequests` | ☐ |
| `GET /api/tools/traffic-inspector/requests` | `ListRequests` | ☐ |
| `GET /api/tools/traffic-inspector/requests/:id` | `GetRequest` | ☐ |
| `POST /api/tools/traffic-inspector/requests/:id/annotation` | `AnnotateRequest` | ☐ |
| `POST /api/tools/traffic-inspector/requests/:id/replay` | `ReplayRequest` | ☐ |
| `GET /api/tools/traffic-inspector/hosts` | `ListHosts` | ☐ |
| `GET /api/tools/traffic-inspector/hosts/:host` | `GetHostRequests` | ☐ |
| `GET /api/tools/traffic-inspector/capture-modes` | `ListCaptureModes` | ☐ |
| `GET /api/tools/traffic-inspector/export` | `ExportRequest` | ☐ |
| `POST /api/tools/traffic-inspector/internal/ingest` | `IngestData` | ☐ |
| `GET /api/middleware/hooks` | `ListMiddlewareHooks` | ☐ |
| `GET/PUT /api/middleware/hooks/:name` | `GetHook` / `UpdateHook` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes (~35) | ☐ |
| 3.2 | `curl localhost:8080/api/tools/agent-bridge/config` | ☐ |
| 3.3 | `curl localhost:8080/api/tools/traffic-inspector/sessions` | ☐ |
| 3.4 | `curl localhost:8080/api/middleware/hooks` | ☐ |

---

## ✅ TASK 4: Sidecar + Tests

| # | Step | Done |
|---|------|------|
| 4.1 | Update nginx | ☐ |
| 4.2 | `go test ./...` | ☐ |
| 4.3 | Deploy | ☐ |

---

## 🚀 QUICK START

```bash
cd omniroute-go && go run .
curl localhost:8080/api/tools/agent-bridge/config
curl localhost:8080/api/tools/traffic-inspector/sessions
curl localhost:8080/api/tools/agent-bridge/agents