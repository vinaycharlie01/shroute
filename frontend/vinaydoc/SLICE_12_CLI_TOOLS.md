# 🎯 Slice 12: Go Backend for CLI Tools & CLI Routes

**Goal**: Migrate CLI tool configuration, status, and runtime endpoints (~40 routes) from TypeScript to Go. Covers 15+ CLI tool settings (Codex, Cline, Claude, Kilo, Qwen, etc.) plus CLI token management.

**Routes**: `/api/cli-tools/*` + `/api/cli/*`

---

## ✅ TASK 1: Types

**Files to create**: `pkg/types/cli_tools.go`, `pkg/types/cli.go`

```go
type CLITool struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Version     string `json:"version"`
    Settings    string `json:"settings"`
    IsInstalled bool   `json:"is_installed"`
    LastUsed    string `json:"last_used"`
    ConfigPath  string `json:"config_path"`
}
type CLIToken struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Token     string `json:"token,omitempty"`
    ExpiresAt string `json:"expires_at"`
    Scope     string `json:"scope"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create `pkg/types/cli_tools.go` | ☐ |
| 1.2 | Create `pkg/types/cli.go` | ☐ |
| 1.3 | Run `go build` | ☐ |

---

## ✅ TASK 2: Repositories

**Files to create**: `internal/db/cli_tools.go`, `internal/db/cli_tokens.go`

```go
type CLIToolsRepo struct { db *sql.DB }
func (r *CLIToolsRepo) ListAll() ([]types.CLITool, error)
func (r *CLIToolsRepo) GetConfig(toolID string) (string, error)
func (r *CLIToolsRepo) UpdateConfig(toolID, config string) error
func (r *CLIToolsRepo) GetStatus(toolID string) (*types.CLITool, error)
func (r *CLIToolsRepo) ApplyConfig(toolID string) error
func (r *CLIToolsRepo) DetectTools() ([]types.CLITool, error)
func (r *CLIToolsRepo) ListBackups() ([]types.BackupEntry, error)
func (r *CLIToolsRepo) CreateBackup() error

type CLITokenRepo struct { db *sql.DB }
func (r *CLITokenRepo) List() ([]types.CLIToken, error)
func (r *CLITokenRepo) Create(token *types.CLIToken) error
func (r *CLITokenRepo) Revoke(id string) error
func (r *CLITokenRepo) Get(id string) (*types.CLIToken, error)
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement CLI tool CRUD | ☐ |
| 2.2 | Implement token CRUD (mask on read) | ☐ |
| 2.3 | Write tests | ☐ |
| 2.4 | `go test ./internal/db/ -run CLITool` | ☐ |

---

## ✅ TASK 3: Handlers (~40 routes)

**Files to create**: `api/handlers/cli_tools.go`, `api/handlers/cli.go`

| Route | Handler | Done |
|-------|---------|------|
| `GET /api/cli-tools/all-statuses` | `GetAllStatuses` | ☐ |
| `GET/PUT /api/cli-tools/config` | `GetConfig` / `UpdateConfig` | ☐ |
| `POST /api/cli-tools/detect` | `DetectTools` | ☐ |
| `POST /api/cli-tools/apply` | `ApplyConfig` | ☐ |
| `GET /api/cli-tools/backups` | `ListBackups` | ☐ |
| `GET /api/cli-tools/status` | `GetStatus` | ☐ |
| `GET /api/cli-tools/keys` | `ListKeys` | ☐ |
| `GET /api/cli-tools/logs` | `ListToolLogs` | ☐ |
| `POST /api/cli-tools/antigravity-mitm` | `AntigravityMITM` | ☐ |
| `POST /api/cli-tools/antigravity-mitm/alias` | `SetAlias` | ☐ |
| `GET/PUT /api/cli-tools/claude-settings` | `ClaudeSettings` | ☐ |
| `GET/PUT /api/cli-tools/cline-settings` | `ClineSettings` | ☐ |
| `GET/PUT /api/cli-tools/codex-settings` | `CodexSettings` | ☐ |
| `GET/PUT /api/cli-tools/codex-profiles` | `CodexProfiles` | ☐ |
| `GET/PUT /api/cli-tools/kilo-settings` | `KiloSettings` | ☐ |
| `GET/PUT /api/cli-tools/qwen-settings` | `QwenSettings` | ☐ |
| `GET/PUT /api/cli-tools/deepseek-tui-settings` | `DeepseekTUISettings` | ☐ |
| `GET/PUT /api/cli-tools/forge-settings` | `ForgeSettings` | ☐ |
| `GET/PUT /api/cli-tools/droid-settings` | `DroidSettings` | ☐ |
| `GET/PUT /api/cli-tools/hermes-agent-settings` | `HermesSettings` | ☐ |
| `POST /api/cli-tools/openclaw/auto-order` | `OpenClawAutoOrder` | ☐ |
| `GET /api/cli-tools/guide-settings/:toolId` | `GuideSettings` | ☐ |
| `GET /api/cli-tools/runtime/:toolId` | `RuntimeSettings` | ☐ |
| `GET/POST /api/cli/tokens` | `ListTokens` / `CreateToken` | ☐ |
| `GET/DELETE /api/cli/tokens/:id` | `GetToken` / `RevokeToken` | ☐ |
| `GET /api/cli/whoami` | `WhoAmI` | ☐ |
| `POST /api/cli/connect` | `Connect` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes | ☐ |
| 3.2 | `curl localhost:8080/api/cli-tools/all-statuses` | ☐ |
| 3.3 | `curl localhost:8080/api/cli/whoami` | ☐ |

---

## ✅ TASK 4: Sidecar + Tests + Deploy

| # | Step | Done |
|---|------|------|
| 4.1 | Update nginx: route `/api/cli-tools/*`, `/api/cli/*` → Go | ☐ |
| 4.2 | Integration: tool list → get settings → update | ☐ |
| 4.3 | `go test ./...` → pass | ☐ |
| 4.4 | Deploy, verify dashboard pages work | ☐ |

---

## 🚀 QUICK START

```bash
cd omniroute-go && go run .
curl localhost:8080/api/cli-tools/all-statuses
curl localhost:8080/api/cli-tools/claude-settings
curl localhost:8080/api/cli/whoami
curl localhost:8080/api/cli/tokens