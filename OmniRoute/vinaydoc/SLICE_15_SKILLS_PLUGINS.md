# 🎯 Slice 15: Go Backend for Skills & Plugin Routes

**Goal**: Migrate skills management, plugin marketplace, and skill execution endpoints from TypeScript to Go.

**Routes**: `/api/skills/*`, `/api/plugins/*`, `/api/agent-skills/*`

---

## ✅ TASK 1: Types

**Files**: `pkg/types/skill.go`, `pkg/types/plugin.go`

```go
type Skill struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Version     string `json:"version"`
    IsEnabled   bool   `json:"is_enabled"`
    Category    string `json:"category"`
    Type        string `json:"type"`   // "built-in", "custom", "skillssh"
}

type SkillExecution struct {
    ID         string `json:"id"`
    SkillID    string `json:"skill_id"`
    Input      string `json:"input"`
    Output     string `json:"output"`
    Status     string `json:"status"`
    StartedAt  string `json:"started_at"`
    CompletedAt string `json:"completed_at"`
    DurationMs int64  `json:"duration_ms"`
}

type Plugin struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    IsActive    bool   `json:"is_active"`
    Config      string `json:"config"` // JSON
    Author      string `json:"author"`
    Description string `json:"description"`
}

type MarketplacePlugin struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    Downloads   int    `json:"downloads"`
    Rating      float64 `json:"rating"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create types | ☐ |

---

## ✅ TASK 2: Repositories

**Files**: `internal/db/skill.go`, `internal/db/plugin.go`

```go
type SkillRepo struct { db *sql.DB }
func (r *SkillRepo) List() ([]types.Skill, error)
func (r *SkillRepo) Enable(id string) error
func (r *SkillRepo) Disable(id string) error
func (r *SkillRepo) Execute(req types.SkillExecution) error
func (r *SkillRepo) GetExecutions() ([]types.SkillExecution, error)
func (r *SkillRepo) Install(name string) error

type PluginRepo struct { db *sql.DB }
func (r *PluginRepo) List() ([]types.Plugin, error)
func (r *PluginRepo) Activate(name string) error
func (r *PluginRepo) Deactivate(name string) error
func (r *PluginRepo) GetConfig(name string) (string, error)
func (r *PluginRepo) Marketplace() ([]types.MarketplacePlugin, error)
func (r *PluginRepo) Scan() error
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement repos | ☐ |

---

## ✅ TASK 3: Handlers

**Files**: `api/handlers/skill.go`, `api/handlers/plugin.go`

| Route | Handler | Done |
|-------|---------|------|
| `GET/PUT /api/skills` | `ListSkills` / `UpdateSkill` | ☐ |
| `DELETE /api/skills/:id` | `DeleteSkill` | ☐ |
| `GET /api/skills/executions` | `ListExecutions` | ☐ |
| `POST /api/skills/install` | `InstallSkill` | ☐ |
| `POST /api/skills/skillssh/install` | `InstallSSHSkill` | ☐ |
| `GET /api/skills/marketplace` | `MarketplaceSkills` | ☐ |
| `POST /api/skills/marketplace/install` | `InstallMarketplaceSkill` | ☐ |
| `GET /api/plugins` | `ListPlugins` | ☐ |
| `POST /api/plugins/:name/activate` | `ActivatePlugin` | ☐ |
| `POST /api/plugins/:name/deactivate` | `DeactivatePlugin` | ☐ |
| `GET /api/plugins/:name/config` | `GetPluginConfig` | ☐ |
| `GET /api/plugins/marketplace` | `PluginMarketplace` | ☐ |
| `POST /api/plugins/scan` | `ScanPlugins` | ☐ |
| `GET /api/agent-skills` | `ListAgentSkills` | ☐ |
| `POST /api/agent-skills/generate` | `GenerateAgentSkill` | ☐ |
| `GET /api/agent-skills/coverage` | `AgentSkillCoverage` | ☐ |
| `GET/DELETE /api/agent-skills/:id` | `GetAgentSkill` / `DeleteAgentSkill` | ☐ |
| `GET /api/agent-skills/:id/raw` | `GetAgentSkillRaw` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes | ☐ |
| 3.2 | `curl localhost:8080/api/skills` | ☐ |
| 3.3 | `curl localhost:8080/api/plugins` | ☐ |
| 3.4 | `curl localhost:8080/api/agent-skills` | ☐ |

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
curl localhost:8080/api/skills
curl localhost:8080/api/plugins
curl localhost:8080/api/agent-skills
curl localhost:8080/api/skills/marketplace