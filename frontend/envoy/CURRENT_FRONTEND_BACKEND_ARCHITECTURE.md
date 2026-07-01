# 🏗️ OmniRoute — Current Frontend & Backend Architecture

> Architecture of the current **TypeScript + Next.js 16** system. The backend is a monolith Next.js app that serves both the dashboard UI and the API/v1 compatibility endpoints, with a shared SSE/routing core in `open-sse/`.

---

## 📊 High-Level Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        🌐 CLIENTS & TOOLS                           │
│                                                                     │
│  Claude Code  Codex CLI  Cursor  Cline  OpenAI SDK  Browser        │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    🚪 ENVOY / REVERSE PROXY (optional)              │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  OmniRoute runs on port 20128 by default                    │   │
│  │  Can be fronted by Envoy, nginx, or Cloudflare Tunnel       │   │
│  └─────────────────────────────────────────────────────────────┘   │
└──────────────┬──────────────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────────────┐
│               ⚡ Next.js 16 App Router (Monolith)                    │
│                                                                      │
│  ┌──────────────────────────┐  ┌───────────────────────────────┐   │
│  │     📱 FRONTEND          │  │       ⚙️ BACKEND              │   │
│  │                          │  │                               │   │
│  │  Dashboard Pages         │  │  /api/v1/* Compatibility API  │   │
│  │  (React Server/Client)   │  │  /api/* Management API        │   │
│  │                          │  │  /api/auth/* Auth Routes      │   │
│  │  i18n (42 locales)       │  │  /api/oauth/* OAuth Routes    │   │
│  │                          │  │  /a2a/* A2A Protocol          │   │
│  │  Tailwind CSS v4         │  │  /api/mcp/* MCP (SSE)         │   │
│  │  next-intl               │  │  /api/settings/* Config       │   │
│  └──────────────────────────┘  └───────────────┬───────────────┘   │
└─────────────────────────────────────────────────┼───────────────────┘
                                                   │
                                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│              🧩 CORE ENGINE — open-sse/ workspace                    │
│                                                                      │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌──────────────┐     │
│  │ Handlers   │ │ Executors  │ │ Services   │ │ Translator   │     │
│  │ (chatCore) │ │ (68 exec)  │ │ (115 mod.) │ │ (format      │     │
│  │            │ │            │ │            │ │  conversion) │     │
│  └────────────┘ └────────────┘ └────────────┘ └──────────────┘     │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌──────────────┐     │
│  │ MCP Server │ │ Transformer│ │ autoCombo  │ │ Compression  │     │
│  │ (94 tools) │ │ (Responses │ │ (9-factor  │ │ (RTK+Caveman)│     │
│  │            │ │  API conv) │ │  scoring)  │ │             │     │
│  └────────────┘ └────────────┘ └────────────┘ └──────────────┘     │
└─────────────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                🏛️ DOMAIN LAYER — src/domain/                        │
│                                                                      │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐                │
│  │ Policy       │ │ Cost Rules   │ │ Fallback     │                │
│  │ Engine       │ │              │ │ Policy       │                │
│  └──────────────┘ └──────────────┘ └──────────────┘                │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐                │
│  │ Lockout      │ │ Combo        │ │ Quota Cache  │                │
│  │ Policy       │ │ Resolver     │ │              │                │
│  └──────────────┘ └──────────────┘ └──────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│              💾 PERSISTENCE LAYER — SQLite                           │
│                                                                      │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐                │
│  │ storage.sqlite│ │ Usage DB     │ │ Domain State │                │
│  │ (providers,  │ │ (usage_hist, │ │ (fallbacks,  │                │
│  │  combos,     │ │  call_logs,  │ │  budgets,    │                │
│  │  keys,       │ │  proxy_logs) │ │  lockouts,   │                │
│  │  settings)   │ │              │ │  breakers)   │                │
│  └──────────────┘ └──────────────┘ └──────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│              🌐 UPSTREAM PROVIDERS — 237 entries                    │
│                                                                      │
│  OAuth: Claude Code  Codex  Gemini  Kiro  Cursor  Antigravity      │
│  API:   OpenAI  Anthropic  DeepSeek  Groq  xAI  Mistral  ...        │
│  Free:  Qoder  Pollinations  LongCat  ...                           │
│  Self:  LM Studio  vLLM  Ollama  Triton  ...                        │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🖥️ FRONTEND — Next.js Dashboard

### Architecture

The frontend is built with **Next.js 16 App Router** using a mix of **Server Components** (for data fetching) and **Client Components** (for interactivity). It's served by the same Next.js process as the backend.

```
src/app/
└── (dashboard)/
    └── dashboard/
        ├── page.tsx                       # Home / Quick Start
        ├── layout.tsx                     # Dashboard layout (sidebar + nav)
        ├── endpoint/
        │   └── page.tsx                   # MCP / A2A / API endpoints tab
        ├── providers/
        │   └── page.tsx                   # Provider connections
        ├── combos/
        │   └── page.tsx                   # Combo strategies, builder
        ├── auto-combo/
        │   └── page.tsx                   # Auto Combo Engine UI
        ├── costs/
        │   └── page.tsx                   # Cost aggregation
        ├── analytics/
        │   └── page.tsx                   # Usage analytics
        ├── health/
        │   └── page.tsx                   # Circuit breaker status
        ├── memory/
        │   └── page.tsx                   # Memory inspection
        ├── skills/
        │   └── page.tsx                   # Skills management
        ├── webhooks/
        │   └── page.tsx                   # Webhook subscriptions
        ├── cache/
        │   └── page.tsx                   # Cache stats
        ├── playground/
        │   └── page.tsx                   # Chat playground
        ├── compression/
        │   └── page.tsx                   # Compression analytics
        ├── context/
        │   ├── caveman/
        │   │   └── page.tsx               # Caveman rules editor
        │   ├── rtk/
        │   │   └── page.tsx               # RTK filter editor
        │   └── combos/
        │       └── page.tsx               # Compression combo assignments
        ├── audit/
        │   └── page.tsx                   # Compliance audit log
        ├── settings/
        │   └── page.tsx                   # System settings
        ├── logs/
        │   └── page.tsx                   # Log viewer
        ├── api-manager/
        │   └── page.tsx                   # API key management
        ├── media/
        │   └── page.tsx                   # Image/video/audio playground
        ├── batch/
        │   └── page.tsx                   # Batch jobs
        ├── cloud-agents/
        │   └── page.tsx                   # Cloud agent tasks
        ├── changelog/
        │   └── page.tsx                   # Changelog viewer
        └── onboarding/
            └── page.tsx                   # First-run wizard
```

### Frontend Component Tree

```
AppLayout
├── Sidebar (navigation)
│   ├── NavItem (providers, combos, health, etc.)
│   └── LanguageSwitcher (42 locales)
├── DashboardPage
│   ├── QuickStartCards
│   ├── ProviderOverview
│   │   ├── ProviderCard (per provider)
│   │   └── ProviderStatusBadge
│   ├── UsageChart (Recharts)
│   └── HealthSummary
│       ├── CircuitBreakerStatus
│       └── RateLimitProgress
├── ProviderPage
│   ├── ProviderList
│   ├── ProviderAddModal
│   └── ProviderConfigForm
├── ComboPage
│   ├── ComboList
│   ├── ComboBuilder (drag-and-drop)
│   │   ├── StepCard (provider + model)
│   │   ├── StrategySelector
│   │   └── TierEditor
│   └── ComboPreview
├── CompressionStudio
│   ├── CavemanRuleEditor
│   ├── RTKFilterEditor
│   ├── LivePreview (request/response)
│   └── CompressionStats
├── MCPPage
│   ├── MCPToolList (94 tools)
│   ├── MCPToolTester
│   └── MCPAuditLog
└── SettingsPage
    ├── GeneralSettings
    ├── RoutingDefaults
    ├── ProxyConfig
    └── IPFilterConfig
```

### Frontend Data Flow

```
User Interaction
    │
    ▼
Client Component (use client)
    │
    ├── fetch() or SWR/React Query ──→ Backend API ──→ JSON Response
    │                                   (src/app/api/*)
    │
    ├── EventSource (SSE) ──→ Backend SSE endpoint
    │                       ──→ Real-time updates (health, logs, usage)
    │
    └── WebSocket ──→ /api/v1/ws ──→ OpenAI-compatible WS
```

### Styling & i18n

| Feature | Library | Details |
|---|---|---|
| **Styling** | Tailwind CSS v4 | Utility-first CSS framework |
| **Charts** | Recharts | React charting library |
| **i18n** | next-intl | 42 locale JSON files in `src/i18n/messages/` |
| **Icons** | lucide-react | SVG icon library |
| **Forms** | React Hook Form | Form validation |

---

## ⚙️ BACKEND — Next.js API Routes + open-sse Engine

### Architecture

The backend is not a separate service — it's the same Next.js process that serves the frontend. The API routes are in `src/app/api/` and the core business logic is in the `open-sse/` workspace package.

```
Next.js Monolith
│
├── src/app/api/              ← REST API endpoints (Route Handlers)
│   ├── v1/                   ← OpenAI-compatible API
│   │   ├── chat/completions/route.ts
│   │   ├── messages/route.ts
│   │   ├── responses/route.ts
│   │   ├── models/route.ts
│   │   ├── embeddings/route.ts
│   │   ├── images/generations/route.ts
│   │   ├── audio/transcriptions/route.ts
│   │   ├── audio/speech/route.ts
│   │   ├── videos/generations/route.ts
│   │   ├── music/generations/route.ts
│   │   ├── moderations/route.ts
│   │   ├── rerank/route.ts
│   │   ├── search/route.ts
│   │   └── ws/route.ts          ← WebSocket bridge
│   ├── auth/*                   ← Login, logout, session
│   ├── oauth/*                  ← OAuth flows
│   ├── providers/*              ← Provider CRUD
│   ├── combos/*                 ← Combo CRUD
│   ├── settings/*               ← System settings
│   ├── keys/*                   ← API key management
│   ├── usage/*                  ← Usage history
│   └── ...                      ← 80+ management endpoints
│
├── open-sse/                 ← Core engine (separate workspace)
│   ├── handlers/              ← Request handlers
│   ├── executors/             ← Provider executors
│   ├── services/              ← Business logic
│   ├── translator/            ← Format translators
│   ├── transformer/           ← Response transformers
│   ├── mcp-server/            ← MCP implementation
│   └── config/                ← Provider configs
│
├── src/domain/               ← Policy engine layer
├── src/lib/                   ← Libraries
│   ├── db/                    ← SQLite persistence (83 modules)
│   ├── guardrails/            ← Security guardrails
│   ├── memory/                ← Conversational memory
│   ├── skills/                ← Skills framework
│   ├── oauth/                 ← OAuth providers (17 modules)
│   ├── a2a/                   ← A2A protocol server
│   ├── acp/                   ← Agent Communication Protocol
│   ├── compression/           ← Compression engine
│   ├── webhooks/              ← Webhook dispatcher
│   └── ...
├── src/middleware/            ← Middleware (prompt injection guard)
├── src/mitm/                  ← MITM proxy
└── src/server/                ← Authz pipeline
```

### Request Pipeline (Backend Flow)

```
HTTP Request (POST /v1/chat/completions)
    │
    ▼
next.config.mjs — Rewrite: /v1/* → /api/v1/*
    │
    ▼
Route: src/app/api/v1/chat/completions/route.ts
    ├── CORS preflight (OPTIONS)
    ├── Body validation (Zod)
    ├── Auth: extractApiKey() / isValidApiKey()
    ├── API Key policy enforcement
    └── Delegate to handler
           │
           ▼
src/sse/handlers/chat.ts — handleChat()
    │
    ├── Model/Combo Resolution
    │   ├── Parse model ID
    │   ├── If combo → open-sse/services/combo.ts::handleComboChat()
    │   │   └── resolveComboTargets() → ordered targets
    │   │       └── For each target: handleSingleModel()
    │   └── If single model → proceed directly
    │
    ├── Credential Selection
    │   └── open-sse/services/accountSelector.ts
    │
    ├── Policy Engine Check (src/domain/policyEngine.ts)
    │   ├── Lockout check
    │   ├── Budget check
    │   └── Fallback check
    │
    ├── Guardrails (src/lib/guardrails/)
    │   ├── PII masker
    │   ├── Prompt injection detection
    │   └── Vision bridge
    │
    ├── Compression (open-sse/services/compression/)
    │   ├── Strategy selection
    │   ├── Caveman / RTK engine
    │   └── Stacked pipeline
    │
    ├── Translation (open-sse/translator/)
    │   ├── Detect source format (OpenAI/Claude/Gemini)
    │   └── translateRequest() → target format
    │
    ├── Execution (open-sse/executors/)
    │   ├── getExecutor() → provider-specific executor
    │   ├── buildUrl() + buildHeaders() + transformRequest()
    │   ├── fetch() to upstream provider
    │   ├── Retry logic (exponential backoff)
    │   └── Token refresh (401 auto-retry)
    │
    ├── Response Translation
    │   ├── Translate back to client format
    │   ├── Think tag parsing
    │   └── Role normalization
    │
    ├── SSE Streaming
    │   └── responsesTransformer.ts → SSE chunks
    │
    └── Usage Extraction
        ├── Token counting
        ├── Cost calculation
        └── SQLite persistence (usage_history)
```

### Backend Routes & Handlers

| API Route | Handler | Description |
|---|---|---|
| **Compatibility API** | | |
| `POST /v1/chat/completions` | `handleChat()` | Chat completions (main endpoint) |
| `POST /v1/responses` | `handleChat()` (unified) | Responses API format |
| `POST /v1/embeddings` | `handleEmbedding()` | Embedding generation |
| `POST /v1/images/generations` | `handleImageGeneration()` | Image generation |
| `POST /v1/audio/transcriptions` | `handleAudioTranscription()` | Audio transcription |
| `POST /v1/audio/speech` | `handleAudioSpeech()` | Text-to-speech |
| `POST /v1/videos/generations` | `handleVideoGeneration()` | Video generation |
| `POST /v1/music/generations` | `handleMusicGeneration()` | Music generation |
| `POST /v1/moderations` | `handleModeration()` | Content moderation |
| `POST /v1/rerank` | `handleRerank()` | Document reranking |
| `POST /v1/search` | `handleSearch()` | Web search |
| `WS /v1/ws` | WebSocket handler | OpenAI-compatible WS |
| **Management API** | | |
| `GET/POST /api/providers` | Provider CRUD | Provider connection management |
| `GET/POST/PUT /api/combos` | Combo CRUD | Combo strategy management |
| `GET/POST /api/keys` | API key management | Key generation & revocation |
| `GET/PUT /api/settings/*` | Settings | System configuration |
| `GET /api/usage/*` | Usage | Usage analytics & budgets |
| `GET /api/resilience` | Resilience | Circuit breaker status |
| **Protocol Endpoints** | | |
| `SSE /api/mcp/sse` | MCP server | MCP over SSE transport |
| `POST /api/mcp/stream` | MCP server | MCP Streamable HTTP |
| `POST /a2a` | A2A server | JSON-RPC 2.0 endpoint |

---

## 🧩 Core Backend Modules (open-sse/)

### Handlers (`open-sse/handlers/`)

```
chatCore.ts           ← Main chat orchestration
  ├── handleChatCore()
  ├── detectSourceFormat()
  ├── translateRequest()
  └── responseHandler()

embeddings.ts         ← Embedding generation
imageGeneration.ts    ← Image generation
audioSpeech.ts        ← Text-to-speech
audioTranscription.ts ← Audio transcription
videoGeneration.ts    ← Video generation
musicGeneration.ts    ← Music generation
moderations.ts        ← Content moderation
rerank.ts             ← Document reranking
search.ts             ← Web search
responseSanitizer.ts  ← Sanitize for OpenAI SDK
responsesHandler.ts   ← Responses API format
responsesTransformer.ts  ← TransformStream for Responses API
sseParser.ts          ← SSE stream parsing
usageExtractor.ts     ← Extract usage from responses
webFetch.ts           ← Web page fetching
```

### Executors (`open-sse/executors/`)

68 provider-specific executors. Each implements the `BaseExecutor` interface:

```
BaseExecutor (abstract)
├── buildUrl()
├── buildHeaders()
├── transformRequest()
├── execute()              ← fetch + retry
└── handleResponse()

DefaultExecutor (OpenAI-compatible)
AnthropicExecutor
GeminiExecutor
ClaudeWebExecutor
CodexExecutor
CursorExecutor
KiroExecutor
QoderExecutor
PollinationsExecutor
... 59 more
```

### Services (`open-sse/services/`)

115+ modules covering:

| Module | Purpose |
|---|---|
| `combo.ts` | Combo routing engine |
| `autoCombo/` | 9-factor scoring engine |
| `accountSelector.ts` | Best account selection (P2C) |
| `accountFallback.ts` | Multi-account fallback |
| `rateLimitManager.ts` | Per-provider rate limiting |
| `tokenRefresh.ts` | OAuth token refresh |
| `circuitBreaker.ts` | Circuit breaker state |
| `contextManager.ts` | Context length management |
| `thinkTagParser.ts` | `<think>` tag extraction |
| `signatureCache.ts` | Request deduplication |
| `wildcardRouter.ts` | Wildcard model routing |
| `workflowFSM.ts` | Workflow state machine |
| `taskAwareRouter.ts` | Task-aware routing |
| `intentClassifier.ts` | Request intent detection |
| `fusion.ts` | Panel + judge synthesis |
| `compression/` | RTK, Caveman, stacked pipelines |
| `reasoningCache.ts` | Reasoning content cache |
| `ipFilter.ts` | IP allowlist/blocklist |
| `systemPrompt.ts` | Global system prompt injection |
| `thinkingBudget.ts` | Thinking budget management |
| `modelFamilyFallback.ts` | Intra-family fallback |
| `emergencyFallback.ts` | Emergency provider fallback |

### Translators (`open-sse/translator/`)

```
index.ts               ← translateRequest() entry
openai-to-claude.ts    ← OpenAI → Anthropic format
openai-to-gemini.ts    ← OpenAI → Gemini format
claude-to-openai.ts    ← Anthropic → OpenAI format
gemini-to-openai.ts    ← Gemini → OpenAI format
helpers.ts             ← Shared translation utilities
```

---

## 🗄️ Persistence Layer

### SQLite Database Modules (`src/lib/db/`)

83 domain-specific modules. Each module owns specific tables:

```
core.ts                ← DB singleton, WAL, schema init
migrationRunner.ts     ← Migration engine (99 migrations)
providers.ts           ← provider_connections table
models.ts              ← Model catalog
combos.ts              ← combos, combo_steps
modelComboMappings.ts  ← Model-combo associations
apiKeys.ts             ← API key storage
settings.ts            ← Key-value settings
secrets.ts             ← Encrypted secrets
usage*.ts              ← Usage history & stats
quotaSnapshots.ts      ← Quota tracking
credits.ts             ← Credit balance
domainState.ts         ← Fallback/budget/lockout state
circuitBreakers.ts     ← Breaker persistence
backup.ts              ← Backup & restore
cleanup.ts             ← Old data cleanup
healthCheck.ts         ← DB health
reasoningCache.ts      ← Reasoning cache
readCache.ts           ← Read cache
batches.ts             ← Batch jobs
files.ts               ← File storage
webhooks.ts            ← Webhook subscriptions
evals.ts               ← Eval persistence
compression.ts         ← Compression config
compressionCombos.ts   ← Compression combo assignments
... (83 total)
```

### Data Flow: API Route → DB

```
Route Handler (src/app/api/*.ts)
    │
    ├── Calls open-sse handler (for v1 endpoints)
    │       └── Handler calls services → services call db modules
    │
    └── Directly calls src/lib/db/ module (for management APIs)
            └── import { getDb } from '@/lib/db/core'
                └── better-sqlite3 queries
```

---

## 🔐 Auth & Security

```
Auth System
├── Dashboard Auth
│   ├── Cookie-based session (src/proxy.ts)
│   └── Login page (src/app/api/auth/login/)
│
├── API Key Auth
│   ├── src/shared/utils/apiKey.ts — Generate/verify
│   ├── src/server/authz/pipeline.ts — Classify → enforce
│   └── src/lib/db/apiKeys.ts — Persistence
│
├── OAuth (17 providers)
│   └── src/lib/oauth/providers/
│       ├── claude.ts, codex.ts, gemini.ts
│       ├── antigravity.ts, qoder.ts, qwen.ts
│       ├── kimi-coding.ts, github.ts, kiro.ts
│       ├── cursor.ts, kilocode.ts, cline.ts
│       ├── windsurf.ts, gitlab-duo.ts, trae.ts
│       └── index.ts (registry)
│
├── Guardrails
│   └── src/lib/guardrails/
│       ├── piiMasker.ts        ← PII detection & masking
│       ├── promptInjection.ts  ← Injection detection
│       └── visionBridge.ts     ← Vision content safety
│
└── SSRF Protection
    └── src/shared/network/outboundUrlGuard.ts
        └── Blocks private/loopback/link-local ranges
```

---

## 🔄 Key Data Flows

### Flow 1: Chat Completion (with Combo)

```
1. POST /v1/chat/completions  { model: "auto/cheap", messages: [...] }
2. next.config.mjs rewrites → /api/v1/chat/completions
3. Route handler:
   a. Zod validation
   b. API key check (optional)
4. chat.ts::handleChat()
5. Resolve model → "auto/cheap" detected as auto prefix
6. open-sse/services/combo.ts::handleComboChat()
   → resolveComboTargets() → [{ glm/glm-5.1 }, { minimax/... }, { kiro/... }]
7. For each target (until success):
   a. Circuit breaker check
   b. Quota check
   c. handleChatCore() → translate request → executor → upstream
   d. If 401: refreshCredentials() → retry
   e. If failure: log → next target
8. Success → SSE stream back to client
9. Usage extracted → SQLite
```

### Flow 2: Dashboard Load

```
1. GET /dashboard/providers
2. Next.js Server Component fetches data
   → GET /api/providers (internal fetch)
3. Route handler:
   a. Session auth check
   b. src/lib/db/providers.ts → SQLite query
   c. Returns JSON
4. Server Component renders ProviderCards
5. Client Component hydrates → adds interactivity
```

### Flow 3: MCP Tool Call

```
1. Client sends JSON-RPC via SSE /api/mcp/sse
2. open-sse/mcp-server/server.ts receives
3. Auth: scope enforcement (30 scopes)
4. Tool registry lookup (94 tools)
5. Zod input validation
6. Handler executes:
   - get_health → check all controllers
   - list_combos → query SQLite
   - route_request → run routing engine
   - web_search → call search provider
7. Response back via SSE
8. Audit log → SQLite (mcp_audit table)
```

---

## 📁 Directory Map (Key Files)

```
src/
├── app/api/v1/
│   ├── chat/completions/route.ts     ← Main chat endpoint
│   ├── responses/route.ts            ← Responses API
│   └── models/route.ts               ← Model listing
├── sse/
│   └── handlers/chat.ts              ← Chat handler
├── lib/
│   ├── localDb.ts                    ← DB re-export facade
│   └── db/
│       ├── core.ts                   ← DB singleton
│       ├── providers.ts              ← Provider CRUD
│       ├── combos.ts                 ← Combo CRUD
│       └── ... (83 modules)
├── server/authz/
│   ├── pipeline.ts                   ← Authz pipeline
│   └── classify.ts                   ← Route classifier
├── middleware/
│   └── promptInjectionGuard.ts       ← Middleware
├── shared/
│   ├── constants/providers.ts        ← Provider registry
│   └── utils/apiKey.ts              ← API key utils
└── domain/
    ├── policyEngine.ts               ← Policy evaluation
    ├── costRules.ts                  ← Cost calculation
    └── fallbackPolicy.ts             ← Fallback logic

open-sse/
├── handlers/chatCore.ts              ← Core orchestration
├── executors/
│   ├── base.ts                       ← Base executor
│   ├── default.ts                    ← OpenAI-compatible
│   └── ... (66 more)
├── services/
│   ├── combo.ts                      ← Combo routing
│   ├── autoCombo/                    ← Auto scoring
│   └── ... (113 more)
├── translator/
│   └── index.ts                      ← Translation entry
├── transformer/
│   └── responsesTransformer.ts       ← Response transform
├── mcp-server/
│   ├── server.ts                     ← MCP server
│   └── tools/                        ← 94 tool handlers
└── config/
    └── providerRegistry.ts           ← Provider configs
```

---

## 📊 Key Numbers (v3.8.40)

| Component | Count |
|---|---|
| **Frontend Dashboard Pages** | 30+ pages |
| **API Routes (v1, management)** | 80+ endpoints |
| **open-sse Handlers** | 12 handlers |
| **Executors** | 68 |
| **Services** | 115+ |
| **MCP Tools** | 94 |
| **DB Modules** | 83 |
| **DB Migrations** | 99 |
| **OAuth Providers** | 17 |
| **i18n Locales** | 42 |
| **Guardrails** | 3 (PII, injection, vision) |
| **A2A Skills** | 6 |

---

> **Source**: OmniRoute v3.8.40 — Current TypeScript/Next.js architecture.