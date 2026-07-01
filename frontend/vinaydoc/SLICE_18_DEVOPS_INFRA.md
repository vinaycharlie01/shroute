# 🎯 Slice 18: Go Backend for DevOps & Infrastructure Routes

**Goal**: Migrate version manager, DB backups/health, sync, system control, headroom, and monitoring endpoints from TypeScript to Go.

**Routes**: `/api/version-manager/*`, `/api/db-backups/*`, `/api/db/health`, `/api/sync/*`, `/api/system/*`, `/api/headroom/*`, `/api/monitoring/*`, `/api/init`, `/api/shutdown`, `/api/restart`

---

## ✅ TASK 1: Types

**Files**: `pkg/types/devops.go`, `pkg/types/sync.go`

```go
type VersionInfo struct {
    Current    string `json:"current"`
    Latest     string `json:"latest"`
    Changelog  string `json:"changelog"`
    UpdateAvailable bool `json:"update_available"`
}

type DBBackup struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    SizeBytes int64  `json:"size_bytes"`
    CreatedAt string `json:"created_at"`
    Checksum  string `json:"checksum"`
    Format    string `json:"format"` // "json", "sqlite"
}

type SyncToken struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Token     string `json:"token,omitempty"`
    ExpiresAt string `json:"expires_at"`
}

type SystemInfo struct {
    Version    string `json:"version"`
    Uptime     int64  `json:"uptime"`
    MemoryUsed int64  `json:"memory_used"`
    CPUCount   int    `json:"cpu_count"`
    DataDir    string `json:"data_dir"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create types | ☐ |

---

## ✅ TASK 2: Repositories

**Files**: `internal/db/devops.go`, `internal/db/sync.go`

```go
type DevOpsRepo struct { db *sql.DB }
func (r *DevOpsRepo) GetVersion() (*types.VersionInfo, error)
func (r *DevOpsRepo) CheckUpdate() (*types.VersionInfo, error)
func (r *DevOpsRepo) InstallUpdate() error
func (r *DevOpsRepo) ListBackups() ([]types.DBBackup, error)
func (r *DevOpsRepo) CreateBackup() error
func (r *DevOpsRepo) ExportBackup(id string) ([]byte, error)
func (r *DevOpsRepo) ExportAllBackup() ([]byte, error)
func (r *DevOpsRepo) ImportBackup(data []byte) error
func (r *DevOpsRepo) GetDBHealth() (*types.DBHealth, error)
func (r *DevOpsRepo) GetSystemInfo() (*types.SystemInfo, error)

type SyncRepo struct { db *sql.DB }
func (r *SyncRepo) ListTokens() ([]types.SyncToken, error)
func (r *SyncRepo) CreateToken(token *types.SyncToken) error
func (r *SyncRepo) RevokeToken(id string) error
func (r *SyncRepo) Initialize() error
func (r *SyncRepo) SyncCloud() error
func (r *SyncRepo) GetBundle() ([]byte, error)
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement all repos | ☐ |

---

## ✅ TASK 3: Handlers

**Files**: `api/handlers/devops.go`, `api/handlers/sync.go`

| Route | Handler | Done |
|-------|---------|------|
| `GET /api/version-manager/check-update` | `CheckUpdate` | ☐ |
| `POST /api/version-manager/install` | `InstallUpdate` | ☐ |
| `POST /api/version-manager/restart` | `RestartVersion` | ☐ |
| `POST /api/version-manager/start` | `StartVersion` | ☐ |
| `GET /api/version-manager/status` | `VersionStatus` | ☐ |
| `POST /api/version-manager/stop` | `StopVersion` | ☐ |
| `GET /api/db-backups` | `ListDBBackups` | ☐ |
| `POST /api/db-backups` | `CreateDBBackup` | ☐ |
| `GET /api/db-backups/export` | `ExportBackup` | ☐ |
| `GET /api/db-backups/exportAll` | `ExportAllBackup` | ☐ |
| `POST /api/db-backups/import` | `ImportBackup` | ☐ |
| `GET /api/db/health` | `DBHealth` | ☐ |
| `GET/POST /api/sync/tokens` | `ListSyncTokens` / `CreateSyncToken` | ☐ |
| `DELETE /api/sync/tokens/:id` | `RevokeSyncToken` | ☐ |
| `POST /api/sync/cloud` | `SyncCloud` | ☐ |
| `POST /api/sync/initialize` | `InitializeSync` | ☐ |
| `GET /api/sync/bundle` | `GetSyncBundle` | ☐ |
| `GET /api/system/version` | `SystemVersion` | ☐ |
| `POST /api/system/env/repair` | `RepairEnv` | ☐ |
| `POST /api/init` | `InitializeSystem` | ☐ |
| `POST /api/shutdown` | `ShutdownSystem` | ☐ |
| `GET /api/headroom/status` | `HeadroomStatus` | ☐ |
| `POST /api/headroom/start` | `StartHeadroom` | ☐ |
| `POST /api/headroom/stop` | `StopHeadroom` | ☐ |
| `GET /api/monitoring/health` | `MonitoringHealth` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes | ☐ |
| 3.2 | `curl localhost:8080/api/version-manager/status` | ☐ |
| 3.3 | `curl localhost:8080/api/db/health` | ☐ |
| 3.4 | `curl localhost:8080/api/system/version` | ☐ |

---

## ✅ TASK 4: Sidecar + Tests

| # | Step | Done |
|---|------|------|
| 4.1 | Update nginx | ☐ |
| 4.2 | Integration: backup → export → import cycle | ☐ |
| 4.3 | `go test ./...` | ☐ |
| 4.4 | Deploy | ☐ |

---

## 🚀 QUICK START

```bash
cd omniroute-go && go run .
curl localhost:8080/api/db/health
curl localhost:8080/api/version-manager/status
curl localhost:8080/api/system/version
curl localhost:8080/api/sync/tokens