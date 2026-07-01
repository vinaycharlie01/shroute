# 🎯 Slice 19: Go Backend for A2A, ACP & Protocol Routes

**Goal**: Migrate agent-to-agent protocol, agent communication protocol, and related agent task endpoints from TypeScript to Go.

**Routes**: `/api/a2a/*`, `/api/acp/*`, `/api/agents/*`

---

## ✅ TASK 1: Types

**Files**: `pkg/types/a2a.go`, `pkg/types/acp.go`, `pkg/types/agent.go`

```go
type A2ATask struct {
    ID          string `json:"id"`
    Status      string `json:"status"` // "submitted", "working", "completed", "failed", "canceled"
    Messages    []A2AMessage `json:"messages"`
    Skill       string `json:"skill"`
    TTL         int    `json:"ttl"`
    CreatedAt   string `json:"created_at"`
}

type A2AMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
    Parts   []A2APart `json:"parts,omitempty"`
}

type A2APart struct {
    Type string `json:"type"` // "text", "file", "data"
    Data string `json:"data,omitempty"`
}

type ACPAgent struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Status      string `json:"status"`
    Capabilities []string `json:"capabilities"`
}

type AgentTask struct {
    ID         string `json:"id"`
    AgentID    string `json:"agent_id"`
    Input      string `json:"input"`
    Output     string `json:"output,omitempty"`
    Status     string `json:"status"`
    CreatedAt  string `json:"created_at"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create types | ☐ |

---

## ✅ TASK 2: Repositories

**Files**: `internal/db/a2a.go`, `internal/db/acp.go`, `internal/db/agent.go`

```go
type A2ARepo struct { db *sql.DB }
func (r *A2ARepo) ListTasks() ([]types.A2ATask, error)
func (r *A2ARepo) GetTask(id string) (*types.A2ATask, error)
func (r *A2ARepo) CreateTask(task *types.A2ATask) error
func (r *A2ARepo) CancelTask(id string) error
func (r *A2ARepo) GetStatus() (string, error)

type ACPRepo struct { db *sql.DB }
func (r *ACPRepo) ListAgents() ([]types.ACPAgent, error)
func (r *ACPRepo) GetAgent(id string) (*types.ACPAgent, error)
func (r *ACPRepo) Register(agent *types.ACPAgent) error

type AgentRepo struct { db *sql.DB }
func (r *AgentRepo) ListTasks() ([]types.AgentTask, error)
func (r *AgentRepo) GetTask(id string) (*types.AgentTask, error)
func (r *AgentRepo) CreateTask(task *types.AgentTask) error
func (r *AgentRepo) GetCredentials() (map[string]string, error)
func (r *AgentRepo) Health() (string, error)
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement all repos | ☐ |

---

## ✅ TASK 3: Handlers

**Files**: `api/handlers/a2a.go`, `api/handlers/acp.go`, `api/handlers/agent.go`

| Route | Handler | Done |
|-------|---------|------|
| `POST /api/a2a/message/sync` | `A2AMessageSync` | ☐ |
| `GET /api/a2a/status` | `A2AStatus` | ☐ |
| `GET /api/a2a/tasks` | `A2AListTasks` | ☐ |
| `POST /api/a2a/tasks` | `A2ACreateTask` | ☐ |
| `GET /api/a2a/tasks/:id` | `A2AGetTask` | ☐ |
| `POST /api/a2a/tasks/:id/cancel` | `A2ACancelTask` | ☐ |
| `GET /api/acp/agents` | `ACPListAgents` | ☐ |
| `GET /api/acp/agents/:id` | `ACPGetAgent` | ☐ |
| `POST /api/acp/agents` | `ACPRregisterAgent` | ☐ |
| `GET /api/v1/agents/tasks` | `ListAgentTasks` | ☐ |
| `POST /api/v1/agents/tasks` | `CreateAgentTask` | ☐ |
| `GET /api/v1/agents/tasks/:id` | `GetAgentTask` | ☐ |
| `GET /api/v1/agents/health` | `AgentHealth` | ☐ |
| `GET /api/v1/agents/credentials` | `AgentCredentials` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes | ☐ |
| 3.2 | `curl localhost:8080/api/a2a/status` | ☐ |
| 3.3 | `curl localhost:8080/api/acp/agents` | ☐ |
| 3.4 | `curl localhost:8080/api/v1/agents/tasks` | ☐ |

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
curl localhost:8080/api/a2a/status
curl localhost:8080/api/acp/agents
curl localhost:8080/api/v1/agents/tasks