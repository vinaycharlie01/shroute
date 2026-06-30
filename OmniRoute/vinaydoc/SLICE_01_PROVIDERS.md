# 🎯 Slice 01: Go Backend for Provider Routes

**Goal**: Migrate all provider CRUD, testing, health, model sync, and bulk operations from TypeScript to Go.

**Routes**: `/api/providers/*`, `/api/v1/providers/*`, `/api/provider-models`, `/api/provider-metrics`, `/api/provider-stats`, `/api/provider-nodes/*`, `/api/free-models`, `/api/free-provider-rankings`

---

## ✅ TASK 1: Types

**Files to create**: `pkg/types/provider.go`

```go
type Provider struct {
    ID              string `json:"id"`
    Name            string `json:"name"`
    Type            string `json:"type"`                 // "openai", "anthropic", "gemini", etc.
    Category        string `json:"category"`             // "api-key", "oauth", "free", "self-hosted"
    BaseURL         string `json:"base_url"`
    IsActive        bool   `json:"is_active"`
    HealthStatus    string `json:"health_status"`         // "healthy", "degraded", "down", "unknown"
    CreatedAt       string `json:"created_at"`
}

type ProviderNode struct {
    ID         string `json:"id"`
    ProviderID string `json:"provider_id"`
    Name       string `json:"name"`
    URL        string `json:"url"`
    Weight     int    `json:"weight"`
    IsActive   bool   `json:"is_active"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create `pkg/types/provider.go` | ☐ |
| 1.2 | Run `go build` | ☐ |

---

## ✅ TASK 2: Provider Repository

**Files to create**: `internal/db/provider.go`

```go
type ProviderRepo struct { db *sql.DB }

// CRUD
func (r *ProviderRepo) List() ([]types.Provider, error)
func (r *ProviderRepo) GetByID(id string) (*types.Provider, error)
func (r *ProviderRepo) Create(provider *types.Provider) error
func (r *ProviderRepo) Update(provider *types.Provider) error
func (r *ProviderRepo) Delete(id string) error

// Bulk
func (r *ProviderRepo) BulkCreate(providers []types.Provider) error
func (r *ProviderRepo) BulkUpdate(providers []types.Provider) error

// Health & testing
func (r *ProviderRepo) Test(providerID string) (*types.TestResult, error)
func (r *ProviderRepo) TestBatch(providerIDs []string) ([]types.TestResult, error)
func (r *ProviderRepo) CheckHealth(providerID string) (string, error)
func (r *ProviderRepo) GetHealthMatrix() (map[string]string, error)
func (r *ProviderRepo) ListExpiring() ([]types.Provider, error)

// Models
func (r *ProviderRepo) SyncModels(providerID string) error
func (r *ProviderRepo) ListModels(providerID string) ([]string, error)
func (r *ProviderRepo) ListProviderModels() ([]types.ProviderModel, error)

// Nodes
func (r *ProviderRepo) ListNodes(providerID string) ([]types.ProviderNode, error)
func (r *ProviderRepo) CreateNode(node *types.ProviderNode) error
func (r *ProviderRepo) DeleteNode(id string) error
func (r *ProviderRepo) ValidateNode(id string) error

// Auth
func (r *ProviderRepo) Login(providerID string) error

// Stats/metrics
func (r *ProviderRepo) GetMetrics(providerID string) (*types.ProviderMetrics, error)
func (r *ProviderRepo) GetStats() (*types.ProviderStats, error)

// Quota windows
func (r *ProviderRepo) GetQuotaWindows(providerID string) ([]types.QuotaWindow, error)

// Client
func (r *ProviderRepo) GetClientProviders() ([]types.Provider, error)

// Expiration
func (r *ProviderRepo) GetExpirationStatus() ([]types.ProviderExpiration, error)

// Health autopilot
func (r *ProviderRepo) GetHealthAutopilotActions() ([]string, error)
func (r *ProviderRepo) UpdateHealthAutopilot(action string) error

// Validate
func (r *ProviderRepo) Validate(provider *types.Provider) error
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement provider CRUD | ☐ |
| 2.2 | Implement models sync & listing | ☐ |
| 2.3 | Implement health checks & matrix | ☐ |
| 2.4 | Implement provider nodes | ☐ |
| 2.5 | Implement quota windows | ☐ |
| 2.6 | Write tests | ☐ |
| 2.7 | `go test ./internal/db/ -run Provider` | ☐ |

---

## ✅ TASK 3: Provider Handlers (~40 routes)

**Files to create**: `api/handlers/provider.go`

| Route | Handler | Done |
|-------|---------|------|
| `GET /api/providers` | `ListProviders` | ☐ |
| `POST /api/providers` | `CreateProvider` | ☐ |
| `GET /api/providers/:id` | `GetProvider` | ☐ |
| `PUT /api/providers/:id` | `UpdateProvider` | ☐ |
| `DELETE /api/providers/:id` | `DeleteProvider` | ☐ |
| `POST /api/providers/bulk` | `BulkCreate` / `BulkUpdate` | ☐ |
| `POST /api/providers/:id/test` | `TestProvider` | ☐ |
| `POST /api/providers/test-batch` | `TestBatch` | ☐ |
| `GET /api/providers/:id/models` | `ListProviderModels` | ☐ |
| `POST /api/providers/:id/sync-models` | `SyncProviderModels` | ☐ |
| `POST /api/providers/:id/login` | `ProviderLogin` | ☐ |
| `POST /api/providers/:id/refresh` | `RefreshProvider` | ☐ |
| `GET /api/providers/health-matrix` | `HealthMatrix` | ☐ |
| `GET /api/providers/health-autopilot` | `HealthAutopilot` | ☐ |
| `POST /api/providers/health-autopilot/actions` | `HealthAutopilotActions` | ☐ |
| `GET /api/providers/expiration` | `ListExpiring` | ☐ |
| `GET /api/providers/quota-windows` | `QuotaWindows` | ☐ |
| `GET /api/providers/client` | `ClientProviders` | ☐ |
| `GET /api/providers/validate` | `ValidateProviders` | ☐ |
| `POST /api/providers/validate` | `ValidateSingleProvider` | ☐ |
| `GET /api/providers/zed/discover` | `ZedDiscover` | ☐ |
| `POST /api/providers/zed/import` | `ZedImport` | ☐ |
| `POST /api/providers/zed/manual-import` | `ZedManualImport` | ☐ |
| `GET /api/provider-models` | `ListAllProviderModels` | ☐ |
| `GET /api/provider-metrics` | `GetProviderMetrics` | ☐ |
| `GET /api/provider-stats` | `GetProviderStats` | ☐ |
| `GET/POST /api/provider-nodes` | `ListNodes` / `CreateNode` | ☐ |
| `DELETE /api/provider-nodes/:id` | `DeleteNode` | ☐ |
| `POST /api/provider-nodes/validate` | `ValidateNode` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all provider routes | ☐ |
| 3.2 | `curl localhost:8080/api/providers` | ☐ |
| 3.3 | `curl localhost:8080/api/provider-models` | ☐ |
| 3.4 | `curl localhost:8080/api/provider-metrics` | ☐ |
| 3.5 | `curl localhost:8080/api/providers/health-matrix` | ☐ |

---

## ✅ TASK 4: Sidecar + Tests + Frontend

| # | Step | Done |
|---|------|------|
| 4.1 | Update nginx: route `/api/providers/*`, `/api/provider-*`, `/api/provider-nodes/*` → Go | ☐ |
| 4.2 | Integration: provider CRUD → sync models → test → health | ☐ |
| 4.3 | `go test ./...` → pass | ☐ |
| 4.4 | Open dashboard → verify provider pages work | ☐ |
| 4.5 | `docker-compose up` → test via nginx | ☐ |

---

## 🚀 QUICK START

```bash
cd omniroute-go && go run .
npm run dev

# Provider CRUD
curl localhost:8080/api/providers
curl localhost:8080/api/providers/health-matrix
curl localhost:8080/api/provider-models

# Create a provider
curl -X POST localhost:8080/api/providers -d '{"name":"test-provider","type":"openai","base_url":"https://api.openai.com/v1"}'