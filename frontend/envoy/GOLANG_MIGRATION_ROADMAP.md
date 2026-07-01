# 🚀 OmniRoute → Go + Envoy Migration Roadmap

> **Goal**: Migrate OmniRoute from TypeScript/Next.js to **Go** with **Envoy** as the API gateway and control plane, using a **Controller pattern** for each core concern.
>
> **Duration**: 6 months (approx. 26 weeks)
>
> **Target Architecture**: Envoy Proxy fronting Go-based controllers with xDS control plane, SQLite/PostgreSQL persistence, and the Next.js dashboard as a standalone frontend.

---

## 🏛️ Target Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Clients / IDEs / CLIs                        │
│           (Claude Code, Codex, Cursor, Cline, etc.)              │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Envoy Proxy (L7 Gateway)                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Envoy Configuration:                                    │   │
│  │  - TLS termination                                       │   │
│  │  - Rate limiting (global + per-route)                    │   │
│  │  - Authn/Authz (JWT, API key validation)                 │   │
│  │  - Request routing to Go controllers                     │   │
│  │  - gRPC-web / HTTP/2 bridging                            │   │
│  │  - Access logging + metrics                              │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────┬──────────┬──────────┬──────────┬──────────────────────┘
           │          │          │          │
           ▼          ▼          ▼          ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Go Control Plane (xDS Server)                   │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│  │  Go Runtime   │ │  Go Runtime  │ │  Go Runtime  │   ...      │
│  │  Controller   │ │  Controller  │ │  Controller  │            │
│  │  (env)        │ │  (env)       │ │  (env)       │            │
│  └──────────────┘ └──────────────┘ └──────────────┘            │
└─────────────────────────────────────────────────────────────────┘
           │          │          │          │
           ▼          ▼          ▼          ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Persistence Layer                               │
│         ┌────────────────────────────────────┐                  │
│         │  SQLite / PostgreSQL + Redis Cache  │                  │
│         └────────────────────────────────────┘                  │
└─────────────────────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────┐
│               Upstream Providers (237 endpoints)                 │
│       OpenAI / Anthropic / Gemini / DeepSeek / Groq / etc.      │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🧩 Controller Architecture

Each controller is a separate Go module with its own responsibility, communicating via well-defined interfaces:

```
omniroute-go/
├── cmd/
│   ├── omniroute/              # Main binary (wires everything)
│   ├── envoy-bootstrap/        # Envoy config generator
│   └── controllers/            # Individual controller binaries (optional)
├── internal/
│   ├── api/                    # Shared API types, protobuf definitions
│   ├── envoy/                  # xDS server, Envoy control plane
│   ├── controllers/
│   │   ├── routing/            # Routing controller
│   │   ├── provider/           # Provider controller
│   │   ├── executor/           # Executor controller  
│   │   ├── translator/         # Translation controller
│   │   ├── auth/               # Auth controller
│   │   ├── quota/              # Quota controller
│   │   ├── resilience/         # Circuit breaker controller
│   │   ├── compression/        # Compression controller
│   │   ├── mcp/                # MCP server controller
│   │   ├── a2a/                # A2A protocol controller
│   │   └── sync/               # Cloud sync controller
│   ├── persistence/            # DB layer (SQLite/Postgres)
│   ├── cache/                  # Redis cache layer
│   └── pkg/                    # Shared utilities
├── config/                     # Config files
├── envoy/                      # Envoy YAML configs
├── dashboard/                  # Next.js frontend (unchanged)
└── docs/                       # Documentation
```

### Controller Responsibilities

| Controller | Responsibility | Envoy Integration |
|---|---|---|
| **Routing Controller** | Combo resolution, model selection, 17 strategies | Dynamic route config via xDS |
| **Provider Controller** | Provider registry, credential management, health checks | Upstream cluster management |
| **Executor Controller** | Provider API execution, retry, fallback | Per-provider HTTP filters |
| **Translator Controller** | Format conversion (OpenAI↔Claude↔Gemini) | Request/response transform filters |
| **Auth Controller** | API key validation, OAuth, JWT | Envoy ext_authz filter |
| **Quota Controller** | Rate limiting, token counting, budget tracking | Global rate limit filter |
| **Resilience Controller** | Circuit breaker, connection cooldown, lockout | Circuit breaker config |
| **Compression Controller** | RTK/Caveman prompt compression | LUA/HTTP filter |
| **MCP Controller** | MCP server with 94 tools, SSE/stdio/HTTP | gRPC-web bridge |
| **A2A Controller** | A2A JSON-RPC 2.0, skills, task lifecycle | HTTP route |
| **Sync Controller** | Cloud sync, multi-device state | — |

---

## 📅 6-Month Roadmap

### Month 1: Foundation & Core Framework

**Week 1-2: Project Setup & Envoy Primer**
- [ ] Initialize Go monorepo (`omniroute-go/`)
- [ ] Set up build system (Makefile, go modules, protobuf toolchain)
- [ ] Write Envoy bootstrap config with static listeners
- [ ] Deploy Envoy as reverse proxy (localhost:20128)
- [ ] Implement basic health check endpoint in Go
- [ ] **Milestone**: Envoy forwards `/health` to Go server

**Week 3-4: Core API Gateway & xDS Control Plane**
- [ ] Implement Envoy xDS server (go-control-plane)
- [ ] Define protobuf types for API routes
- [ ] Build dynamic route configuration via xDS
- [ ] Implement `/v1/chat/completions` basic passthrough
- [ ] Add API key validation (ext_authz filter)
- [ ] **Milestone**: Request flows through Envoy → Go → upstream

**Week 5-6: Persistence Layer**
- [ ] Implement SQLite driver (using `modernc.org/sqlite` — pure Go)
- [ ] Schema migration system (similar to current 99 migrations)
- [ ] Implement core models: providers, combos, keys, settings
- [ ] Implement usage history and call logs tables
- [ ] Add Redis cache layer for hot data
- [ ] Write integration tests for persistence
- [ ] **Milestone**: Full persistence layer with migrations

**Week 7-8: Auth Controller**
- [ ] Implement API key generation and validation
- [ ] OAuth provider framework (17 providers)
- [ ] JWT token management
- [ ] Dashboard session auth (cookie-based)
- [ ] Envoy ext_authz gRPC integration
- [ ] **Milestone**: Auth controller fully replaces existing auth

---

### Month 2: Core Routing & Provider Layer

**Week 5-6: Provider Controller**
- [ ] Provider registry (237 providers)
- [ ] Provider configuration and credential storage
- [ ] Health check probes (periodic + on-demand)
- [ ] Provider discovery and model catalog sync
- [ ] **Milestone**: Provider CRUD + health monitoring

**Week 7-8: Executor Controller**
- [ ] Base executor interface and default implementation
- [ ] OpenAI-compatible provider executor
- [ ] Anthropic-compatible provider executor
- [ ] Free/Web-based provider executors (Kiro, Qoder, etc.)
- [ ] Retry logic with exponential backoff
- [ ] Token refresh for OAuth providers
- [ ] **Milestone**: Chat completions execute against 10+ providers

**Week 9-10: Translator Controller**
- [ ] Request translator engine (OpenAI → Claude → Gemini)
- [ ] Response translator engine (reverse direction)
- [ ] Image/video/audio content handling
- [ ] Structured output conversion (json_schema)
- [ ] Role normalization (developer↔system↔user)
- [ ] **Milestone**: Cross-provider translation working

**Week 11-12: Routing Controller — Part 1**
- [ ] Combo model and target resolution
- [ ] Priority and weighted routing strategies
- [ ] Round-robin, random, least-used strategies
- [ ] Fill-first and P2C strategies
- [ ] Dynamic Envoy route updates based on combo config
- [ ] **Milestone**: Basic combo routing with 5 strategies

---

### Month 3: Advanced Routing & Resilience

**Week 13-14: Routing Controller — Part 2**
- [ ] Auto-combo engine with 9-factor scoring
- [ ] Cost-optimized and reset-aware strategies
- [ ] Context-relay and context-optimized strategies
- [ ] Fusion strategy (panel fan-out + judge)
- [ ] Task-aware routing (workflow FSM)
- [ ] **Milestone**: All 17 routing strategies implemented

**Week 15-16: Resilience Controller**
- [ ] Circuit breaker pattern (per-provider)
- [ ] Connection cooldown (per-account/key)
- [ ] Model lockout (per-provider+model)
- [ ] Anti-thundering herd protection (mutex locking)
- [ ] Domain state persistence (fallbacks, budgets, lockouts)
- [ ] Envoy circuit breaker integration
- [ ] **Milestone**: 3-layer resilience fully operational

**Week 17-18: Quota & Rate Limiting**
- [ ] Quota tracker (per-connection, per-model)
- [ ] Rate limit manager with provider-specific profiles
- [ ] Envoy global rate limit filter integration
- [ ] Budget tracking and enforcement
- [ ] Token counting (hybrid: provider-side + estimation)
- [ ] **Milestone**: Quota-aware routing and rate limiting

---

### Month 4: Protocol Servers & Compression

**Week 19-20: MCP Controller**
- [ ] MCP protocol implementation (JSON-RPC 2.0)
- [ ] 3 transports: stdio, SSE, Streamable HTTP
- [ ] 20 core tools (get_health, list_combos, route_request, etc.)
- [ ] Cache tools (cache_stats, cache_flush)
- [ ] Compression tools (compression_status, configure, etc.)
- [ ] 1proxy tools (oneproxy_fetch, rotate, stats)
- [ ] 30-scope authorization system
- [ ] MCP audit logging
- [ ] **Milestone**: MCP server with 94 tools

**Week 21-22: A2A Controller**
- [ ] A2A protocol (JSON-RPC 2.0 + SSE)
- [ ] Task Manager with TTL cleanup
- [ ] Task lifecycle (submitted → working → completed|failed)
- [ ] 6 A2A skills (smart routing, quota, discovery, cost, health, capabilities)
- [ ] Agent Card at `/.well-known/agent.json`
- [ ] **Milestone**: A2A server with full skill set

**Week 23-24: Compression Controller**
- [ ] Caveman compression engine
- [ ] RTK compression engine
- [ ] Stacked pipeline support
- [ ] Compression combos and stats
- [ ] Language packs for Caveman
- [ ] Compression analytics and telemetry
- [ ] **Milestone**: Compression pipeline matching current capabilities

---

### Month 5: Advanced Features & Migration

**Week 25-26: Memory & Skills Systems**
- [ ] FTS5 full-text search (SQLite)
- [ ] Vector-based similarity search
- [ ] Memory extraction, injection, retrieval, summarization
- [ ] Skills registry and executor
- [ ] Custom skill sandbox
- [ ] Built-in skills (quota, routing, etc.)
- [ ] **Milestone**: Memory + skills matching current feature set

**Week 27-28: Guardrails & Security**
- [ ] PII masker guardrail
- [ ] Prompt injection detection
- [ ] Vision bridge guardrail
- [ ] Hot-reloadable guardrail framework
- [ ] SSRF guard for outbound requests
- [ ] TLS fingerprint stealth (JA3/JA4)
- [ ] **Milestone**: Full security layer

**Week 29-30: Infrastructure & Integration**
- [ ] MITM proxy with certificate management
- [ ] Cloud sync (optional multi-device)
- [ ] Tunnels (Cloudflare, ngrok, Tailscale)
- [ ] Webhook dispatcher (HMAC, retry, auto-disable)
- [ ] Eval framework for LLM quality
- [ ] Graceful shutdown, signal handling
- [ ] **Milestone**: Infrastructure layer complete

---

### Month 6: Testing, Dashboard & Polish

**Week 31-32: Dashboard Integration**
- [ ] Next.js dashboard (keep existing frontend)
- [ ] Go API endpoints for dashboard CRUD
- [ ] Real-time SSE updates (circuit breaker status, usage)
- [ ] WebSocket bridge for OpenAI-compatible WS
- [ ] Health dashboard with live provider status
- [ ] **Milestone**: Dashboard connects to Go backend

**Week 33-34: Testing — Unit & Integration**
- [ ] Unit tests for every controller (aim for 80%+ coverage)
- [ ] Integration tests for request pipeline
- [ ] Integration tests for combo routing
- [ ] Integration tests for all 17 strategies
- [ ] Integration tests for MCP/A2A protocols
- [ ] Benchmark tests (latency, throughput)
- [ ] **Milestone**: Test suite with 10,000+ tests

**Week 35-36: Testing — E2E & Migration Validation**
- [ ] E2E tests with real provider calls (using test accounts)
- [ ] Performance benchmarking (compare TS vs Go)
- [ ] Load testing with k6 (target: 10x throughput improvement)
- [ ] Migration validation — run TS and Go in parallel
- [ ] A/B testing framework for migration confidence
- [ ] Documentation update
- [ ] **Milestone**: Full migration validated

---

## 🔄 Migration Strategy

### Phase 1: Side-by-Side (Weeks 1-16)
```
Port 20128 → Envoy → Go controllers (new endpoints)
Port 20129 → Next.js (existing, for dashboard)
Go controllers implement new endpoints alongside existing TS code
```

### Phase 2: Gradual Cutover (Weeks 17-24)
```
Envoy progressively routes more traffic to Go controllers
Start with: /health, /v1/models
Then: /v1/chat/completions (single model)
Then: /v1/chat/completions (combo routing)
Then: /v1/embeddings, /v1/images, /v1/audio
Finally: MCP, A2A endpoints
```

### Phase 3: Full Migration (Weeks 25-36)
```
Port 20128 → Envoy → Go controllers (100% traffic)
Next.js serves only the dashboard frontend
TS backend decommissioned
```

---

## 🛤️ Tech Stack Decisions

| Component | Choice | Rationale |
|---|---|---|
| **Language** | Go 1.24+ | Performance, concurrency, compilation |
| **HTTP Framework** | `net/http` + `chi` or `gin` | Lightweight, standard library compatible |
| **gRPC** | `connect-go` or `grpc-go` | For Envoy xDS and internal communication |
| **Envoy Control Plane** | `go-control-plane` | Standard xDS v3 implementation |
| **SQLite** | `modernc.org/sqlite` | Pure Go, no CGO, matches current DB |
| **PostgreSQL** | `pgx` v5 | For production deployments |
| **Redis** | `go-redis` | Caching, rate limiting, pub/sub |
| **Protobuf** | `buf` build tool | Type definitions for API |
| **Config** | `koanf` + YAML | Flexible configuration |
| **Logging** | `zerolog` | Structured, zero-allocation |
| **Testing** | `testify` + `mockery` | Assertions and mocking |
| **CI/CD** | GitHub Actions | Already in use |

---

## 🎯 Key Metrics & Success Criteria

| Metric | Current (TS) | Target (Go) |
|---|---|---|
| **P50 Latency** | ~50ms | <10ms |
| **P99 Latency** | ~200ms | <50ms |
| **Throughput** | ~1K req/s | >10K req/s |
| **Memory Usage** | ~150MB idle | <30MB idle |
| **Startup Time** | ~5s (Next.js) | <500ms |
| **Binary Size** | N/A (interpreter) | <20MB |
| **Test Count** | ~15,000 | ~15,000+ |
| **Concurrent Connections** | ~500 | >5,000 |

---

## 📦 Envoy Configuration Example

```yaml
# envoy/config/envoy.yaml
admin:
  address:
    socket_address: { address: 0.0.0.0, port_value: 9901 }

static_resources:
  listeners:
  - name: main_listener
    address:
      socket_address: { address: 0.0.0.0, port_value: 20128 }
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          codec_type: AUTO
          route_config_name: dynamic_routes
          http_filters:
          - name: envoy.filters.http.ext_authz
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz
              grpc_service:
                envoy_grpc:
                  cluster_name: auth_controller
          - name: envoy.filters.http.ratelimit
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.ratelimit.v3.RateLimit
              domain: omniroute
              stage: 0
              request_type: both
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router

  clusters:
  - name: auth_controller
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    typed_extension_protocol_options:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
        explicit_http_config:
          http2_protocol_options: {}
    load_assignment:
      cluster_name: auth_controller
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address: { address: 127.0.0.1, port_value: 9001 }

  - name: routing_controller
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: routing_controller
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address: { address: 127.0.0.1, port_value: 9002 }

dynamic_resources:
  lds_config:
    ads: {}
  cds_config:
    ads: {}
  ads_config:
    api_type: DELTA_GRPC
    transport_api_version: V3
    grpc_services:
      envoy_grpc:
        cluster_name: xds_cluster

  clusters:
  - name: xds_cluster
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    typed_extension_protocol_options:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
        explicit_http_config:
          http2_protocol_options: {}
    load_assignment:
      cluster_name: xds_cluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address: { address: 127.0.0.1, port_value: 18000 }
```

---

## 🧪 Go Controller Example

```go
// internal/controllers/routing/controller.go
package routing

import (
    "context"
    "net/http"

    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"

    "omniroute-go/internal/api"
    "omniroute-go/internal/controllers/routing/strategies"
)

// Controller handles routing decisions and model selection
type Controller struct {
    router     *chi.Mux
    resolver   *ComboResolver
    strategies map[string]strategies.Strategy
    logger     *zap.Logger
}

// NewController creates a new routing controller
func NewController(logger *zap.Logger, db *sql.DB) *Controller {
    c := &Controller{
        router:   chi.NewRouter(),
        resolver: NewComboResolver(db),
        strategies: map[string]strategies.Strategy{
            "priority":        &strategies.Priority{},
            "weighted":        &strategies.Weighted{},
            "round-robin":     &strategies.RoundRobin{},
            "cost-optimized":  &strategies.CostOptimized{},
            "auto":            strategies.NewAutoCombo(db),
            // ... 12 more strategies
        },
        logger: logger,
    }
    c.routes()
    return c
}

// ResolveTarget picks the best provider+model for a request
func (c *Controller) ResolveTarget(ctx context.Context, req *api.ChatRequest) (*api.ResolvedTarget, error) {
    strategy := c.resolver.DetectStrategy(req.Model)
    impl, ok := c.strategies[strategy]
    if !ok {
        return nil, fmt.Errorf("unknown strategy: %s", strategy)
    }
    return impl.Select(ctx, req)
}

func (c *Controller) routes() {
    c.router.Post("/v1/chat/completions", c.handleChat)
    c.router.Post("/v1/responses", c.handleResponses)
    c.router.Get("/v1/models", c.handleModels)
}

func (c *Controller) handleChat(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request (Zod-equivalent validation)
    // 2. Resolve combo target
    // 3. Execute via executor controller
    // 4. Stream response back
}

// ServeHTTP implements http.Handler
func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    c.router.ServeHTTP(w, r)
}
```

---

## ⚠️ Risks & Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| **Feature parity gap** | Users lose functionality | Side-by-side run, feature-flag comparison |
| **Performance regression** | Slower than TS | Benchmark every controller; optimize early |
| **Provider compatibility** | 237 providers behave differently | Comprehensive integration tests per provider |
| **OAuth complexity** | 17 OAuth flows to re-implement | Start with top 5 (Claude, Codex, Gemini, GitHub, Kiro) |
| **MCP/A2A protocol drift** | Client incompatibility | Protocol conformance test suite |
| **Team Go inexperience** | Slow development | First month focused on Go training via foundation work |
| **Data migration errors** | Data loss | Dry-run migration, checksums, rollback plan |

---

## 💡 Phased Delivery Summary

| Month | Deliverable |
|---|---|
| **Month 1** | ✅ Envoy gateway + Auth + Persistence + Basic routing |
| **Month 2** | ✅ Provider registry + Executors + Translation + 5 strategies |
| **Month 3** | ✅ All 17 strategies + 3-layer resilience + Quota system |
| **Month 4** | ✅ MCP server (94 tools) + A2A server + Compression engine |
| **Month 5** | ✅ Memory/Skills + Guardrails + Infrastructure + Webhooks |
| **Month 6** | ✅ Dashboard integration + 10K+ tests + Full cutover validation |

> **Total**: ~36 weeks of development for a feature-complete Go + Envoy migration.
>
> **Parallel execution possible**: Months 2-4 can be partially parallelized with 2-3 developers.

---

> *Generated for OmniRoute — v3.8.40 architecture baseline*