# 🎯 Slice 17: Go Backend for Webhooks, Compliance, Guardrails & Tags

**Goal**: Migrate webhook delivery, compliance audit logs, guardrail management, and tagging endpoints from TypeScript to Go.

**Routes**: `/api/webhooks/*`, `/api/compliance/*`, `/api/guardrails/*`, `/api/tags`, `/api/policies`

---

## ✅ TASK 1: Types

**Files**: `pkg/types/webhook.go`, `pkg/types/compliance.go`, `pkg/types/guardrail.go`

```go
type Webhook struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    URL        string `json:"url"`
    Secret     string `json:"secret,omitempty"`
    Events     []string `json:"events"`
    IsActive   bool   `json:"is_active"`
    CreatedAt  string `json:"created_at"`
}

type WebhookDelivery struct {
    ID         string `json:"id"`
    WebhookID  string `json:"webhook_id"`
    Status     string `json:"status"` // "success", "failed", "retrying"
    ResponseCode int   `json:"response_code"`
    Attempts   int    `json:"attempts"`
    DeliveredAt string `json:"delivered_at"`
}

type ComplianceAuditEntry struct {
    ID        string `json:"id"`
    Action    string `json:"action"`
    Actor     string `json:"actor"`
    Resource  string `json:"resource"`
    Detail    string `json:"detail"`
    CreatedAt string `json:"created_at"`
}

type Guardrail struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Type        string `json:"type"` // "pii-masker", "prompt-injection", "vision-bridge"
    IsEnabled   bool   `json:"is_enabled"`
    Config      string `json:"config"`
    FailOpen    bool   `json:"fail_open"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create types | ☐ |

---

## ✅ TASK 2: Repositories

**Files**: `internal/db/webhook.go`, `internal/db/compliance.go`, `internal/db/guardrail.go`

```go
type WebhookRepo struct { db *sql.DB }
func (r *WebhookRepo) List() ([]types.Webhook, error)
func (r *WebhookRepo) Get(id string) (*types.Webhook, error)
func (r *WebhookRepo) Create(wh *types.Webhook) error
func (r *WebhookRepo) Update(wh *types.Webhook) error
func (r *WebhookRepo) Delete(id string) error
func (r *WebhookRepo) Test(id string) error
func (r *WebhookRepo) GetDeliveries(id string) ([]types.WebhookDelivery, error)
func (r *WebhookRepo) ValidateURL(url string) error

type ComplianceRepo struct { db *sql.DB }
func (r *ComplianceRepo) ListAuditLog() ([]types.ComplianceAuditEntry, error)

type GuardrailRepo struct { db *sql.DB }
func (r *GuardrailRepo) List() ([]types.Guardrail, error)
func (r *GuardrailRepo) Get(id string) (*types.Guardrail, error)
func (r *GuardrailRepo) Update(gr *types.Guardrail) error
func (r *GuardrailRepo) Test(id string, input string) (bool, error)
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement webhook CRUD | ☐ |
| 2.2 | Implement delivery tracking | ☐ |
| 2.3 | Implement compliance audit | ☐ |
| 2.4 | Implement guardrail config | ☐ |

---

## ✅ TASK 3: Handlers

**Files**: `api/handlers/webhook.go`, `api/handlers/compliance.go`, `api/handlers/guardrail.go`

| Route | Handler | Done |
|-------|---------|------|
| `GET/POST /api/webhooks` | `ListWebhooks` / `CreateWebhook` | ☐ |
| `GET/PUT/DELETE /api/webhooks/:id` | `GetWebhook` / `UpdateWebhook` / `DeleteWebhook` | ☐ |
| `GET /api/webhooks/:id/deliveries` | `GetDeliveries` | ☐ |
| `POST /api/webhooks/:id/test` | `TestWebhook` | ☐ |
| `POST /api/webhooks/validate-url` | `ValidateWebhookURL` | ☐ |
| `GET /api/compliance/audit-log` | `GetAuditLog` | ☐ |
| `GET /api/guardrails` | `ListGuardrails` | ☐ |
| `PUT /api/guardrails` | `UpdateGuardrail` | ☐ |
| `POST /api/guardrails/test` | `TestGuardrail` | ☐ |
| `GET /api/policies` | `GetPolicies` | ☐ |
| `GET/POST /api/tags` | `ListTags` / `CreateTag` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes | ☐ |
| 3.2 | `curl localhost:8080/api/webhooks` | ☐ |
| 3.3 | `curl localhost:8080/api/compliance/audit-log` | ☐ |
| 3.4 | `curl localhost:8080/api/guardrails` | ☐ |

---

## ✅ TASK 4: Sidecar + Tests

| # | Step | Done |
|---|------|------|
| 4.1 | Update nginx | ☐ |
| 4.2 | Integration test: webhook CRUD → test → deliveries | ☐ |
| 4.3 | `go test ./...` | ☐ |
| 4.4 | Deploy | ☐ |

---

## 🚀 QUICK START

```bash
cd omniroute-go && go run .
curl localhost:8080/api/webhooks
curl localhost:8080/api/guardrails
curl localhost:8080/api/tags