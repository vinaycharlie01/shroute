# 🎯 Slice 16: Go Backend for Memory, Gamification & Evals Routes

**Goal**: Migrate memory extraction/injection/retrieval, gamification (badges, levels, leaderboard), and eval framework endpoints from TypeScript to Go.

**Routes**: `/api/memory/*`, `/api/gamification/*`, `/api/evals/*`

---

## ✅ TASK 1: Types

**Files**: `pkg/types/memory.go`, `pkg/types/gamification.go`, `pkg/types/eval.go`

```go
type MemoryEntry struct {
    ID          string `json:"id"`
    Content     string `json:"content"`
    SessionID   string `json:"session_id"`
    Category    string `json:"category"`
    Importance  int    `json:"importance"`
    CreatedAt   string `json:"created_at"`
    LastUsedAt  string `json:"last_used_at"`
    Embeddings  string `json:"embeddings,omitempty"`
}

type MemoryHealth struct {
    Status       string `json:"status"`
    EntryCount   int    `json:"entry_count"`
    EngineStatus string `json:"engine_status"`
}

type Badge struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    IconURL     string `json:"icon_url"`
    Criteria    string `json:"criteria"`
}

type LeaderboardEntry struct {
    UserID   string `json:"user_id"`
    Score    int64  `json:"score"`
    Level    int    `json:"level"`
    Badges   int    `json:"badges"`
    Rank     int    `json:"rank"`
}

type Eval struct {
    ID      string `json:"id"`
    SuiteID string `json:"suite_id"`
    Name    string `json:"name"`
    Status  string `json:"status"` // "pending", "running", "passed", "failed"
    Score   float64 `json:"score"`
}

type EvalSuite struct {
    ID           string `json:"id"`
    Name         string `json:"name"`
    Target       string `json:"target"` // "combo", "model", "suite-default"
    IsRunning    bool   `json:"is_running"`
    LastRunAt    string `json:"last_run_at"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create all types | ☐ |

---

## ✅ TASK 2: Repositories

**Files**: `internal/db/memory.go`, `internal/db/gamification.go`, `internal/db/eval.go`

```go
type MemoryRepo struct { db *sql.DB }
func (r *MemoryRepo) Search(query string) ([]types.MemoryEntry, error)
func (r *MemoryRepo) Add(entry *types.MemoryEntry) error
func (r *MemoryRepo) Clear(sessionID string) error
func (r *MemoryRepo) Get(id string) (*types.MemoryEntry, error)
func (r *MemoryRepo) RetrievePreview(sessionID string) (string, error)
func (r *MemoryRepo) Summarize(sessionID string) (string, error)
func (r *MemoryRepo) Reindex() error
func (r *MemoryRepo) Health() (*types.MemoryHealth, error)
func (r *MemoryRepo) GetEngineStatus() (string, error)
func (r *MemoryRepo) GetEmbeddingProviders() ([]string, error)

type GamificationRepo struct { db *sql.DB }
func (r *GamificationRepo) GetLevel() (*types.Level, error)
func (r *GamificationRepo) ListBadges() ([]types.Badge, error)
func (r *GamificationRepo) GetEarnedBadges() ([]types.Badge, error)
func (r *GamificationRepo) GetLeaderboard() ([]types.LeaderboardEntry, error)
func (r *GamificationRepo) GetFederationLeaderboard() ([]types.LeaderboardEntry, error)
func (r *GamificationRepo) GetNotifications() ([]string, error)
func (r *GamificationRepo) Rotate() error
func (r *GamificationRepo) Transfer() error
func (r *GamificationRepo) CreateInvite() (string, error)
func (r *GamificationRepo) RedeemInvite(code string) error
func (r *GamificationRepo) ListServers() ([]string, error)
func (r *GamificationRepo) GetAnomalies() ([]string, error)
func (r *GamificationRepo) Stream() error
func (r *GamificationRepo) FederationScore() error

type EvalRepo struct { db *sql.DB }
func (r *EvalRepo) ListSuites() ([]types.EvalSuite, error)
func (r *EvalRepo) GetSuite(id string) (*types.EvalSuite, error)
func (r *EvalRepo) CreateSuite(suite *types.EvalSuite) error
func (r *EvalRepo) List(suiteID string) ([]types.Eval, error)
func (r *EvalRepo) Run(evalID string) error
func (r *EvalRepo) GetResults(suiteID string) ([]types.Eval, error)
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement all repos | ☐ |

---

## ✅ TASK 3: Handlers (~30 routes)

**Files**: `api/handlers/memory.go`, `api/handlers/gamification.go`, `api/handlers/eval.go`

| Memory Routes | Handler | Done |
|---------------|---------|------|
| `GET /api/memory` | `SearchMemory` | ☐ |
| `POST /api/memory` | `AddMemory` | ☐ |
| `DELETE /api/memory` | `ClearMemory` | ☐ |
| `GET /api/memory/:id` | `GetMemory` | ☐ |
| `GET /api/memory/retrieve-preview` | `RetrievePreview` | ☐ |
| `POST /api/memory/summarize` | `SummarizeMemory` | ☐ |
| `POST /api/memory/reindex` | `ReindexMemory` | ☐ |
| `GET /api/memory/health` | `MemoryHealth` | ☐ |
| `GET /api/memory/engine-status` | `EngineStatus` | ☐ |
| `GET /api/memory/embedding-providers` | `EmbeddingProviders` | ☐ |

| Gamification Routes | Handler | Done |
|---------------------|---------|------|
| `GET /api/gamification/level` | `GetLevel` | ☐ |
| `GET /api/gamification/badges` | `ListBadges` | ☐ |
| `GET /api/gamification/badges/earned` | `GetEarnedBadges` | ☐ |
| `GET /api/gamification/leaderboard` | `GetLeaderboard` | ☐ |
| `GET /api/gamification/federation/leaderboard` | `FederationLeaderboard` | ☐ |
| `POST /api/gamification/federation/score` | `FederationScore` | ☐ |
| `GET /api/gamification/notifications` | `GetNotifications` | ☐ |
| `POST /api/gamification/rotate` | `RotateGamification` | ☐ |
| `POST /api/gamification/transfer` | `TransferGamification` | ☐ |
| `GET /api/gamification/invite` | `CreateInvite` | ☐ |
| `POST /api/gamification/invite/redeem` | `RedeemInvite` | ☐ |
| `GET /api/gamification/servers` | `ListServers` | ☐ |
| `GET /api/gamification/anomalies` | `GetAnomalies` | ☐ |
| `POST /api/gamification/stream` | `StreamGamification` | ☐ |

| Eval Routes | Handler | Done |
|-------------|---------|------|
| `GET /api/evals` | `ListEvals` | ☐ |
| `GET /api/evals/suites` | `ListEvalSuites` | ☐ |
| `POST /api/evals/suites` | `CreateEvalSuite` | ☐ |
| `GET/DELETE /api/evals/suites/:suiteId` | `GetSuite` / `DeleteSuite` | ☐ |
| `POST /api/evals/:suiteId` | `RunEval` | ☐ |
| `GET /api/evals/:suiteId` | `GetEvalResults` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes | ☐ |
| 3.2 | `curl localhost:8080/api/memory` | ☐ |
| 3.3 | `curl localhost:8080/api/gamification/level` | ☐ |
| 3.4 | `curl localhost:8080/api/evals/suites` | ☐ |

---

## ✅ TASK 4: Sidecar + Tests + Deploy

| # | Step | Done |
|---|------|------|
| 4.1 | Update nginx | ☐ |
| 4.2 | `go test ./...` | ☐ |
| 4.3 | Deploy | ☐ |

---

## 🚀 QUICK START

```bash
cd omniroute-go && go run .
curl localhost:8080/api/memory/health
curl localhost:8080/api/gamification/level
curl localhost:8080/api/evals/suites