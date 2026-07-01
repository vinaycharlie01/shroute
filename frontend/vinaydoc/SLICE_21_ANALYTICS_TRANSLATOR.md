# 🎯 Slice 21: Go Backend for Analytics, Translator & Misc Routes

**Goal**: Migrate analytics dashboard, translation proxy, playground, docs, OpenAPI, pricing, auth/login, and remaining misc endpoints from TypeScript to Go.

**Routes**: `/api/analytics/*`, `/api/translator/*`, `/api/playground/*`, `/api/docs/*`, `/api/openapi/*`, `/api/pricing/*`, `/api/auth/*`, `/api/assess`, `/api/free-models/*`, `/api/free-tier/*`, `/api/intelligence/sync`, `/api/telemetry/*`, `/api/token-health`

---

## ✅ TASK 1: Types

**Files**: `pkg/types/analytics.go`, `pkg/types/translator.go`, `pkg/types/misc.go`

```go
type AnalyticsReport struct {
    PeriodStart  string  `json:"period_start"`
    PeriodEnd    string  `json:"period_end"`
    TotalRequests int64  `json:"total_requests"`
    TotalTokens  int64   `json:"total_tokens"`
    AvgLatencyMs float64 `json:"avg_latency_ms"`
    ErrorRate    float64 `json:"error_rate"`
    ProviderBreakdown map[string]ProviderStats `json:"provider_breakdown"`
}

type TranslateRequest struct {
    SourceLang string `json:"source_lang"`
    TargetLang string `json:"target_lang"`
    Text       string `json:"text"`
    Provider   string `json:"provider,omitempty"`
}

type TranslateHistory struct {
    ID         string `json:"id"`
    SourceText string `json:"source_text"`
    Translated string `json:"translated"`
    SourceLang string `json:"source_lang"`
    TargetLang string `json:"target_lang"`
    CreatedAt  string `json:"created_at"`
}

type PricingModel struct {
    ID        string  `json:"id"`
    Provider  string  `json:"provider"`
    ModelName string  `json:"model_name"`
    InputCost float64 `json:"input_cost_per_1k"`
    OutputCost float64 `json:"output_cost_per_1k"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create types | ☐ |

---

## ✅ TASK 2: Repositories

**Files**: `internal/db/analytics.go`, `internal/db/translator.go`, `internal/db/pricing.go`

```go
type AnalyticsRepo struct { db *sql.DB }
func (r *AnalyticsRepo) GetAutoRouting() (*AnalyticsReport, error)
func (r *AnalyticsRepo) GetCompression() (*AnalyticsReport, error)
func (r *AnalyticsRepo) GetDiversity() (*AnalyticsReport, error)
func (r *AnalyticsRepo) GetSummary() (*AnalyticsReport, error)

type TranslatorRepo struct { db *sql.DB }
func (r *TranslatorRepo) Translate(req *types.TranslateRequest) (string, error)
func (r *TranslatorRepo) Detect(text string) (string, error)
func (r *TranslatorRepo) GetHistory() ([]types.TranslateHistory, error)
func (r *TranslatorRepo) TransformStream(input string) (string, error)

type PricingRepo struct { db *sql.DB }
func (r *PricingRepo) ListModels() ([]types.PricingModel, error)
func (r *PricingRepo) GetDefaults() (map[string]float64, error)
func (r *PricingRepo) Sync() error
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement analytics queries | ☐ |
| 2.2 | Implement translation proxy | ☐ |
| 2.3 | Implement pricing sync | ☐ |

---

## ✅ TASK 3: Handlers

**Files**: `api/handlers/analytics.go`, `api/handlers/translator.go`, `api/handlers/misc.go`

| Route | Handler | Done |
|-------|---------|------|
| `GET /api/analytics/auto-routing` | `AutoRoutingAnalytics` | ☐ |
| `GET /api/analytics/compression` | `CompressionAnalytics` | ☐ |
| `GET /api/analytics/diversity` | `DiversityAnalytics` | ☐ |
| `GET /api/telemetry/summary` | `TelemetrySummary` | ☐ |
| `POST /api/translator/translate` | `Translate` | ☐ |
| `POST /api/translator/detect` | `DetectLanguage` | ☐ |
| `GET /api/translator/history` | `TranslateHistory` | ☐ |
| `POST /api/translator/send` | `TranslateSend` | ☐ |
| `POST /api/translator/transform-stream` | `TransformStream` | ☐ |
| `GET /api/playground/presets` | `ListPlaygroundPresets` | ☐ |
| `GET/PUT /api/playground/presets/:id` | `GetPreset` / `UpdatePreset` | ☐ |
| `POST /api/playground/simulate-route` | `SimulateRoute` | ☐ |
| `POST /api/playground/improve-prompt` | `ImprovePrompt` | ☐ |
| `GET /api/docs` | `ListDocs` | ☐ |
| `GET /api/docs/codex-cli` | `CodexCLIDocs` | ☐ |
| `GET /api/openapi/spec` | `OpenAPISpec` | ☐ |
| `POST /api/openapi/try` | `TryOpenAPI` | ☐ |
| `GET /api/pricing` | `ListPricing` | ☐ |
| `GET /api/pricing/models` | `PricingModels` | ☐ |
| `POST /api/pricing/sync` | `SyncPricing` | ☐ |
| `GET /api/pricing/defaults` | `PricingDefaults` | ☐ |
| `POST /api/auth/login` | `Login` | ☐ |
| `POST /api/auth/logout` | `Logout` | ☐ |
| `GET /api/auth/status` | `AuthStatus` | ☐ |
| `POST /api/assess` | `Assess` | ☐ |
| `GET /api/free-models` | `FreeModels` | ☐ |
| `GET /api/free-provider-rankings` | `FreeProviderRankings` | ☐ |
| `GET /api/free-tier/summary` | `FreeTierSummary` | ☐ |
| `POST /api/intelligence/sync` | `SyncIntelligence` | ☐ |
| `GET /api/token-health` | `TokenHealth` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes | ☐ |
| 3.2 | `curl localhost:8080/api/analytics/auto-routing` | ☐ |
| 3.3 | `curl localhost:8080/api/translator/translate` | ☐ |
| 3.4 | `curl localhost:8080/api/pricing` | ☐ |
| 3.5 | `curl localhost:8080/api/auth/status` | ☐ |

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
curl localhost:8080/api/analytics/auto-routing
curl localhost:8080/api/translator/translate
curl localhost:8080/api/pricing
curl localhost:8080/api/auth/status