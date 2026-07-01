# 🌊 OmniRoute Go — Flow Architecture Diagram

> High-level flow architecture showing how data, requests, and configuration move through the **Go + Envoy + xDS + Kubernetes Operator** system.

---

## 🗺️ SYSTEM FLOW ARCHITECTURE — OVERVIEW

```mermaid
flowchart TB
    subgraph Users["👤 Users & Clients"]
        CLI["Claude Code CLI"]
        CURSOR["Cursor IDE"]
        CLINE["Cline"]
        SDK["OpenAI SDK / HTTP Client"]
        DASH["Dashboard UI (Next.js)"]
    end

    subgraph Gateway["🚪 Gateway Layer (Port 20128)"]
        ENV["Envoy Proxy<br/>L7 Reverse Proxy"]
        EXT_AUTH["ext_authz gRPC Server<br/>Port 9001"]
        RL_GRPC["Rate Limit gRPC Server<br/>Port 9003"]
    end

    subgraph Backend["⚙️ Go Backend (Port 8080)"]
        CHI["chi HTTP Router<br/>net/http"]
        MIDDLEWARE["Middleware Pipeline<br/>RequestID → Logger → CORS → Recovery → Validation"]
        
        subgraph Controllers["🎮 Controller Layer"]
            AUTH["Auth Controller<br/>• API Key Validation<br/>• OAuth (14 providers)<br/>• JWT Sessions<br/>• ext_authz Server"]
            ROUTE["Routing Controller<br/>• 17 Strategies<br/>• Combo Resolution<br/>• AutoCombo Scoring<br/>• Task-Aware FSM"]
            TRANS["Translator Controller<br/>• OpenAI↔Claude↔Gemini<br/>• Role Normalizer<br/>• Think Parser<br/>• Structured Output"]
            EXEC["Executor Controller<br/>• 68 Provider Executors<br/>• Retry + Backoff<br/>• OAuth Token Refresh<br/>• SSE Streaming"]
            PROV["Provider Controller<br/>• 237 Providers<br/>• Health Probes<br/>• Registry"]
            RSV["Resilience Controller<br/>• Circuit Breaker<br/>• Cooldown / Lockout<br/>• Herd Protection"]
            QUOTA["Quota Controller<br/>• Rate Limiter (Redis)<br/>• Budget Tracking<br/>• Token Counter"]
            COMP["Compression Controller<br/>• Caveman Engine<br/>• RTK Engine<br/>• Pipeline"]
            MCP["MCP Controller<br/>• 94 Tools<br/>• 3 Transports<br/>• 30 Scopes"]
            A2A["A2A Controller<br/>• JSON-RPC 2.0<br/>• 6 Skills<br/>• Task Manager"]
            MEM["Memory Controller<br/>• FTS5 + Vector<br/>• Extract / Inject"]
            SKILL["Skills Controller<br/>• Registry / Executor<br/>• Sandbox"]
            GRD["Guardrails Controller<br/>• PII Masker<br/>• Injection Detection"]
            WH["Webhook Controller<br/>• HMAC Delivery<br/>• 7 Event Types"]
            EVAL["Eval Controller<br/>• Runner / Runtime<br/>• Targets"]
            TUN["Tunnel Controller<br/>• Cloudflare / ngrok<br/>• Tailscale Funnel"]
            MITM["MITM Controller<br/>• Cert Manager<br/>• TPROXY"]
            SYNC["Sync Controller<br/>• Cloud Sync"]
        end

        subgraph Data["🗄️ Data Layer"]
            SQL["SQLite (WAL Mode)<br/>• 83 Repositories<br/>• 17 Tables<br/>• 99 Migrations"]
            REDIS["Redis Cache<br/>• Rate Counters<br/>• Sessions<br/>• CB State<br/>• Provider Health"]
        end

        subgraph Stream["📡 Streaming Engine"]
            SSE["SSE Writer / Reader<br/>Event Parser + Transform"]
            WS["WebSocket Bridge<br/>OpenAI-Compatible"]
        end
    end

    subgraph K8s["☸️ Kubernetes Operator"]
        OP["Operator<br/>kubebuilder"]
        CRD_PROV["CRD: Provider"]
        CRD_COMBO["CRD: Combo"]
        CRD_KEY["CRD: ApiKey"]
        CRD_CB["CRD: CircuitBreaker"]
        CRD_RL["CRD: RateLimit"]
        CRD_WH["CRD: Webhook"]
        CRD_MCP["CRD: MCPTool"]
        CRD_TUN["CRD: Tunnel"]
    end

    subgraph xDS["🔄 xDS Control Plane (Port 18000)"]
        XDS_SRV["xDS gRPC Server<br/>LDS / RDS / CDS / EDS"]
        SNAP_CACHE["Snapshot Cache<br/>Versioned Configs"]
        RESOURCE_BUILD["Resource Builders<br/>Routes / Clusters / Listeners"]
    end

    subgraph Upstream["🔌 Upstream Providers (237)"]
        UP_OA["OpenAI / Anthropic / Gemini"]
        UP_DS["DeepSeek / Groq / xAI / Mistral"]
        UP_OT["Together / Fireworks / Cerebras / Cohere"]
        UP_MO["+231 More: Free(3) OAuth(14) API Key(120+) Self-Hosted(8+)"]
    end

    %% ─── DATA FLOWS ───
    
    %% Client → Gateway
    CLI -->|"HTTPS :20128"| ENV
    CURSOR -->|"HTTPS :20128"| ENV
    CLINE -->|"HTTPS :20128"| ENV
    SDK -->|"HTTPS :20128"| ENV
    DASH -->|"HTTPS :20128"| ENV

    %% Envoy internal flows
    ENV -->|"gRPC ext_authz"| EXT_AUTH
    ENV -->|"gRPC rate limit"| RL_GRPC
    ENV -->|"xDS config push"| XDS_SRV
    
    %% Envoy → Backend
    ENV -->|"HTTP/2 :8080"| CHI

    %% Router → Middleware → Controllers
    CHI --> MIDDLEWARE
    
    %% Auth flow
    MIDDLEWARE --> AUTH
    AUTH -->|"read/write"| SQL
    AUTH -->|"cache"| REDIS
    EXT_AUTH -->|"key lookup"| AUTH
    RL_GRPC -->|"quota check"| QUOTA

    %% Routing flow
    MIDDLEWARE --> ROUTE
    ROUTE -->|"load combos"| SQL
    ROUTE -->|"cache"| REDIS
    ROUTE -->|"translate"| TRANS
    ROUTE -->|"execute"| EXEC

    %% Executor flow
    EXEC -->|"circuit check"| RSV
    EXEC -->|"provider lookup"| PROV
    EXEC -->|"quota check"| QUOTA
    EXEC -->|"compress"| COMP
    EXEC -->|"SSE stream"| SSE
    EXEC -->|"WS stream"| WS
    
    %% Executor → Upstream
    EXEC -->|"HTTP POST"| UP_OA
    EXEC -->|"HTTP POST"| UP_DS
    EXEC -->|"HTTP POST"| UP_OT
    EXEC -->|"HTTP POST"| UP_MO

    %% Supporting controllers → Data
    RSV -->|"CB state"| REDIS
    RSV -->|"persist"| SQL
    QUOTA -->|"counters"| REDIS
    QUOTA -->|"persist"| SQL
    PROV -->|"registry"| SQL
    COMP -->|"config"| SQL
    MCP -->|"audit + config"| SQL
    A2A -->|"tasks + skills"| SQL
    MEM -->|"vectors + text"| SQL
    SKILL -->|"registry"| SQL
    GRD -->|"rules"| SQL
    WH -->|"events"| SQL
    EVAL -->|"results"| SQL
    TUN -->|"config"| SQL
    MITM -->|"certs"| SQL
    SYNC -->|"state"| SQL

    %% xDS flow
    ROUTE -->|"UpdateRoutes()"| XDS_SRV
    AUTH -->|"UpdateAuthConfig()"| XDS_SRV
    QUOTA -->|"UpdateRateLimits()"| XDS_SRV
    XDS_SRV --> RESOURCE_BUILD
    RESOURCE_BUILD --> SNAP_CACHE
    SNAP_CACHE -->|"DiscoveryResponse"| ENV

    %% Kubernetes Operator → Controllers
    OP -->|"reconcile"| PROV
    OP -->|"reconcile"| ROUTE
    OP -->|"reconcile"| AUTH
    OP -->|"reconcile"| RSV
    OP -->|"reconcile"| QUOTA
    OP -->|"reconcile"| WH
    OP -->|"reconcile"| MCP
    OP -->|"reconcile"| TUN
    
    %% CRDs → Operator
    CRD_PROV --> OP
    CRD_COMBO --> OP
    CRD_KEY --> OP
    CRD_CB --> OP
    CRD_RL --> OP
    CRD_WH --> OP
    CRD_MCP --> OP
    CRD_TUN --> OP
```

---

## 🔄 REQUEST FLOW (Client → Upstream → Response)

```mermaid
flowchart LR
    subgraph Inbound["Inbound"]
        direction LR
        A["Client<br/>HTTP Request"] --> B["Envoy<br/>TLS + Auth + RateLimit"]
        B --> C["Go Router<br/>chi :8080"]
    end

    subgraph Process["Processing Pipeline"]
        direction TB
        C --> D["Middleware<br/>RequestID → Logger → CORS → Recovery → Validate"]
        D --> E{"Auth?<br/>API Key / OAuth / JWT"}
        E -->|"Valid"| F{"Rate Limit?<br/>Redis Token Bucket"}
        E -->|"Invalid"| ERR1["401 Unauthorized"]
        F -->|"Under Limit"| G["Routing<br/>17 Strategies + Combo"]
        F -->|"Over Limit"| ERR2["429 Too Many Requests"]
        G --> H{"Translate?<br/>Cross-Format?"}
        H -->|"Yes"| I["Translator<br/>OpenAI↔Claude↔Gemini"]
        H -->|"No"| J
        I --> J{"Circuit Breaker?<br/>Redis Fast Path"}
        J -->|"CLOSED"| K["Build Request<br/>URL + Headers + Body"]
        J -->|"OPEN"| FALLBACK["Try Fallback Target"]
        FALLBACK --> G
        K --> L["Execute<br/>HTTP POST + Retry x3"]
    end

    subgraph UpstreamCall["Upstream"]
        L --> M["Provider API<br/>237 Providers"]
        M --> N{"Streaming?<br/>SSE Mode?"}
        N -->|"Yes"| O["SSE Reader<br/>Parse Events"]
        N -->|"No"| P["JSON Response<br/>Parse Body"]
    end

    subgraph Response["Response Pipeline"]
        O --> Q{"Translate?<br/>Response Format"}
        P --> Q
        Q -->|"Yes"| R["Response Translator<br/>Claude→OpenAI Format"]
        Q -->|"No"| S
        R --> S["SSE Writer / JSON Writer"]
        S --> T["Add Headers<br/>RateLimit + RequestID"]
        T --> U["Async Logging<br/>Usage → SQLite"]
        U --> V["Response to Client<br/>200 / SSE Stream"]
    end

    subgraph Errors["Error Paths"]
        ERR1 --> V
        ERR2 --> V
        M -->|"401"| TOKEN["OAuth Token Refresh"]
        TOKEN --> L
        M -->|"5xx"| RETRY["Retry with Backoff<br/>100ms→200ms→400ms+jitter"]
        RETRY --> L
    end
```

---

## 🔧 CONTROL PLANE FLOW (xDS Config Updates)

```mermaid
flowchart LR
    subgraph Trigger["Config Change Triggers"]
        ADMIN["Admin API<br/>POST /api/routing/combos"]
        K8S_OP["K8s Operator<br/>CRD Change"]
        DASH_UI["Dashboard UI<br/>Combo Editor"]
        AUTO["AutoCombo<br/>ML Scoring Update"]
    end

    subgraph Controller["Controller Layer"]
        RC["Routing Controller<br/>Reconciler Loop"]
        AC["Auth Controller<br/>Key/Scope Updates"]
        QC["Quota Controller<br/>Limit Updates"]
    end

    subgraph xDS_Path["xDS Update Path"]
        BUILD["Build Resources<br/>Routes / Clusters / Listeners"]
        SNAP["Build Snapshot<br/>Version Increment"]
        CACHE["Snapshot Cache<br/>Stored by Node ID"]
        PUSH["gRPC Stream<br/>DiscoveryResponse"]
    end

    subgraph Envoy["Envoy Proxy"]
        LDS["LDS - Listener Discovery"]
        RDS["RDS - Route Discovery"]
        CDS["CDS - Cluster Discovery"]
        EDS["EDS - Endpoint Discovery"]
        HOT_SWAP["Hot Swap<br/>Zero-Downtime"]
    end

    subgraph Data["Persistence"]
        SQLITE["SQLite<br/>Config Storage"]
    end

    %% Flows
    ADMIN --> RC
    K8S_OP --> RC
    DASH_UI --> RC
    AUTO --> RC
    
    RC --> SQLITE
    AC --> SQLITE
    QC --> SQLITE
    
    SQLITE --> RC
    SQLITE --> AC
    SQLITE --> QC
    
    RC --> BUILD
    AC --> BUILD
    QC --> BUILD
    
    BUILD --> SNAP
    SNAP --> CACHE
    CACHE --> PUSH
    
    PUSH --> LDS
    PUSH --> RDS
    PUSH --> CDS
    PUSH --> EDS
    
    LDS --> HOT_SWAP
    RDS --> HOT_SWAP
    CDS --> HOT_SWAP
    EDS --> HOT_SWAP
    
    HOT_SWAP -->|"ACK (version + nonce)"| CACHE
```

---

## 🏗️ LAYERED ARCHITECTURE FLOW

```mermaid
flowchart TB
    subgraph L7["Layer 7: Client & External"]
        L7A["Claude Code CLI"]
        L7B["Cursor / Cline"]
        L7C["OpenAI SDK"]
        L7D["Browser (Dashboard)"]
    end

    subgraph L6["Layer 6: Edge Gateway"]
        L6A["HTTPS Termination"]
        L6B["ext_authz gRPC<br/>API Key Validation"]
        L6C["Rate Limit gRPC<br/>Redis Token Bucket"]
        L6D["Envoy Routing<br/>xDS Dynamic Config"]
    end

    subgraph L5["Layer 5: HTTP Router & Middleware"]
        L5A["chi Router<br/>Route Dispatch"]
        L5B["Middleware Pipeline<br/>RequestID → Logger → CORS → Recovery → Validation"]
    end

    subgraph L4["Layer 4: Request Controllers (per-request)"]
        L4A["Auth Controller<br/>Key / OAuth / JWT"]
        L4B["Routing Controller<br/>Strategy + Combo"]
        L4C["Translator Controller<br/>Format Conversion"]
        L4D["Executor Controller<br/>Provider Dispatch"]
    end

    subgraph L3["Layer 3: Infrastructure Controllers"]
        L3A["Resilience<br/>Circuit Breaker"]
        L3B["Provider<br/>Registry + Health"]
        L3C["Quota<br/>Budget + Tokens"]
        L3D["Compression<br/>Caveman + RTK"]
    end

    subgraph L2["Layer 2: Persistence & Cache"]
        L2A["SQLite<br/>Primary Store<br/>WAL Mode / 83 Repos"]
        L2B["Redis<br/>Cache Layer<br/>Counters / Sessions / CB"]
    end

    subgraph L1["Layer 1: Upstream Providers"]
        L1A["OpenAI / Anthropic / Gemini"]
        L1B["DeepSeek / Groq / xAI"]
        L1C["Mistral / Together / Fireworks"]
        L1D["+231 More Providers"]
    end

    subgraph L0["Layer 0: Kubernetes Operator & xDS"]
        L0A["K8s Operator<br/>kubebuilder"]
        L0B["xDS Control Plane<br/>gRPC :18000"]
        L0C["Envoy Dynamic Config<br/>LDS / RDS / CDS / EDS"]
    end

    %% L7 → L6
    L7A -->|"HTTPS"| L6A
    L7B -->|"HTTPS"| L6A
    L7C -->|"HTTPS"| L6A
    L7D -->|"HTTPS"| L6A

    %% L6 internal
    L6A --> L6B
    L6A --> L6C
    L6A --> L6D

    %% L6 → L5
    L6A -->|"HTTP/2"| L5A
    L5A --> L5B

    %% L5 → L4
    L5B --> L4A
    L5B --> L4B
    L5B --> L4C

    L4B --> L4D
    L4C --> L4D

    %% L4 → L3
    L4D --> L3A
    L4D --> L3B
    L4D --> L3C
    L4D --> L3D

    %% L3 → L2
    L3A -->|"state"| L2A
    L3A -->|"fast path"| L2B
    L3B -->|"config"| L2A
    L3C -->|"counters"| L2B
    L3C -->|"limits"| L2A
    L3D -->|"config"| L2A

    %% L4 → L2
    L4A -->|"keys"| L2A
    L4B -->|"combos"| L2A

    %% L4 → L1
    L4D -->|"HTTP POST"| L1A
    L4D -->|"HTTP POST"| L1B
    L4D -->|"HTTP POST"| L1C
    L4D -->|"HTTP POST"| L1D

    %% L0 → L6
    L0A -->|"reconcile"| L6D
    L0B -->|"snapshot push"| L6D
    L6D -->|"ACK"| L0C

    %% L0 → L4
    L0A -->|"CRD reconcile"| L4A
    L0A -->|"CRD reconcile"| L4B
    L0A -->|"CRD reconcile"| L3B
    L0A -->|"CRD reconcile"| L3A
    L0A -->|"CRD reconcile"| L3C
```

---

## 🎯 COMPONENT INTERACTION FLOW (Data & Control)

```mermaid
flowchart TB
    subgraph External["External"]
        CLIENT["Client"]
        PROVIDER["Upstream Provider<br/>(OpenAI / Claude / etc.)"]
    end

    subgraph RequestPath["Request Path (Data Plane)"]
        direction LR
        RP1["Envoy<br/>Filter Chain"] --> RP2["chi Router<br/>+ Middleware"]
        RP2 --> RP3["Auth<br/>Controller"]
        RP3 --> RP4["Routing<br/>Controller"]
        RP4 --> RP5["Translator<br/>Controller"]
        RP5 --> RP6["Executor<br/>Controller"]
        RP6 --> RP7["SSE Stream<br/>Engine"]
    end

    subgraph ControlPath["Control Path (Control Plane)"]
        direction LR
        CP1["Admin API<br/>/api/..."] --> CP2["Controller<br/>Reconciler"]
        CP2 --> CP3["xDS Server<br/>Snapshot Builder"]
        CP3 --> CP4["Envoy<br/>Dynamic Update"]
    end

    subgraph ConfigPath["Config Path (Persistence)"]
        direction LR
        CF1["Controller"] --> CF2["SQLite<br/>Repository"]
        CF2 --> CF3["Controller<br/>Reconciler"]
    end

    subgraph CachePath["Cache Path (Fast Data)"]
        direction LR
        CA1["Controller"] --> CA2["Redis<br/>Cache"]
        CA2 --> CA1
    end

    subgraph K8sPath["K8s Operator Path"]
        direction LR
        K1["CRD<br/>Change"] --> K2["Operator<br/>Controller"]
        K2 --> K3["Go<br/>Controller"]
        K3 --> K4["xDS<br/>Cache"]
        K4 --> K5["Envoy<br/>Update"]
    end

    %% Cross-diagram connections
    CLIENT --> RP1
    RP6 -->|"HTTP POST"| PROVIDER
    PROVIDER -->|"SSE / JSON"| RP7
    RP7 -->|"Response"| CLIENT

    CP2 --> CF1
    CP2 --> CA1

    K3 --> CP2
```

---

## 📊 LEGEND

```mermaid
flowchart LR
    A["⚡ Data Flow<br/>(Request/Response)"] --> B["🔄 Control Flow<br/>(Config/Update)"]
    B --> C["💾 Persistence Flow<br/>(Read/Write)"]
    C --> D["⚡ Cache Flow<br/>(Fast Path)"]
```

---

> **See also:**
> - `GOLANG_E2E_FLOW_ARCHITECTURE.md` — End-to-end request sequence (16-step detailed)
> - `ARCHITECTURE_DIAGRAM.md` — Component architecture
> - `GOLANG_ENVOY_K8S_OPERATOR_ROADMAP.md` — 7-month implementation plan