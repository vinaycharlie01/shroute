# 🔄 OmniRoute — Current Workflow Architecture

> How the **current TypeScript/Next.js 16 system** processes requests through its workflow pipeline. No Go, no Envoy, no Controller patterns — just the actual flow as it exists today.

---

## 📊 HIGH-LEVEL WORKFLOW

```mermaid
flowchart TB
    subgraph Client["👤 Clients"]
        direction TB
        C1["Claude Code CLI"]
        C2["Cursor IDE"]
        C3["Cline / Any AI Tool"]
        C4["OpenAI SDK / curl"]
        C5["Browser Dashboard"]
    end

    subgraph Entry["🚪 Entry Points"]
        direction TB
        E1["HTTPS / HTTP<br/>Port 20128"]
        E2["next.config.mjs<br/>Rewrite: /v1/* → /api/v1/*"]
    end

    subgraph Routes["📡 API Routes (src/app/api/)"]
        direction TB
        R1["/api/v1/chat/completions<br/>Main Chat Endpoint"]
        R2["/api/v1/responses<br/>Responses API"]
        R3["/api/v1/embeddings<br/>Embedding Gen"]
        R4["/api/v1/images/generations<br/>Image Gen"]
        R5["/api/v1/audio/*<br/>Speech / Transcribe"]
        R6["/api/v1/videos/generations<br/>Video Gen"]
        R7["/api/v1/music/generations<br/>Music Gen"]
        R8["/api/v1/moderations<br/>Content Safety"]
        R9["/api/v1/rerank<br/>Document Re-rank"]
        R10["/api/v1/search<br/>Web Search"]
        R11["/api/mcp/sse<br/>MCP Server (SSE)"]
        R12["/a2a<br/>A2A Protocol"]
        R13["/api/*<br/>Management APIs"]
    end

    subgraph Pipeline["⚙️ Request Pipeline (per-request)"]
        direction TB
        P1["1. CORS Preflight<br/>OPTIONS check"]
        P2["2. Body Validation<br/>Zod schemas"]
        P3["3. Auth Check<br/>extractApiKey() / isValidApiKey()"]
        P4["4. API Key Policy<br/>enforceApiKeyPolicy()"]
        P5["5. Route-specific<br/>Middleware"]
        P6["6. Handler<br/>Delegation"]
    end

    subgraph Engine["🧩 Core Engine (open-sse/)"]
        direction TB
        S1["Service Selection<br/>• Model Resolution<br/>• Combo Detection"]
        S2["Credential Selection<br/>accountSelector.ts<br/>Best account (P2C)"]
        S3["Policy Engine<br/>policyEngine.ts<br/>• Lockout?<br/>• Budget?<br/>• Fallback?"]
        S4["Guardrails<br/>• PII Masker<br/>• Injection Detection<br/>• Vision Bridge"]
        S5["Compression<br/>• Strategy Select<br/>• Caveman / RTK<br/>• Stacked Pipeline"]
        S6["Translation<br/>translateRequest()<br/>OpenAI↔Claude↔Gemini"]
        S7["Execution<br/>getExecutor()<br/>buildUrl() → buildHeaders() → fetch()"]
        S8["Response Translation<br/>Translate back to client format<br/>Think tag parsing"]
        S9["SSE Streaming<br/>responsesTransformer.ts<br/>Chunk-by-chunk"]
        S10["Usage Extraction<br/>Token counting → Cost calc → SQLite"]
    end

    subgraph Providers["🔌 Upstream (237 Providers)"]
        direction TB
        U1["OpenAI / Anthropic / Gemini"]
        U2["DeepSeek / Groq / xAI / Mistral"]
        U3["Together / Fireworks / Cohere"]
        U4["OAuth: Claude Code, Gemini, Codex..."]
        U5["Self-Hosted: LM Studio, vLLM..."]
    end

    subgraph Storage["💾 SQLite Persistence (src/lib/db/)"]
        direction TB
        D1["83 Domain Modules<br/>17 Base Tables<br/>99 Migrations"]
        D2["Usage History<br/>Tokens / Cost / Latency"]
        D3["Combo Configs<br/>Strategies + Targets"]
        D4["Provider Credentials<br/>Encrypted Secrets"]
        D5["MCP Audit Log<br/>Tool Invocations"]
        D6["Circuit Breaker State<br/>Failures / Recovery"]
    end

    %% Connections
    Client --> Entry
    Entry --> Routes
    
    Routes --> Pipeline
    
    Pipeline --> P1 --> P2 --> P3 --> P4 --> P5 --> P6

    P6 -->|"Delegate to handler"| S1
    
    S1 -->|"Model resolved"| S2
    S2 -->|"Credentials selected"| S3
    S3 -->|"Policies passed"| S4
    S4 -->|"Content safe"| S5
    S5 -->|"Compressed"| S6
    S6 -->|"Translated"| S7
    S7 -->|"Response received"| S8
    S8 -->|"Translated back"| S9
    S9 -->|"Streaming to client"| S10
    
    S7 -->|"HTTP POST"| Providers
    Providers -->|"SSE / JSON"| S8
    
    S3 -->|"read/write"| Storage
    S10 -->|"write"| Storage
    S2 -->|"read"| Storage
    S1 -->|"read"| Storage
```

---

## 🔄 CHAT COMPLETION WORKFLOW (Most Common Path)

```mermaid
sequenceDiagram
    participant Client as Client<br/>(Claude Code / Cursor / SDK)
    participant Next as Next.js 16<br/>(next.config.mjs)
    participant Route as API Route<br/>(src/app/api/v1/chat/completions/route.ts)
    participant Handler as Chat Handler<br/>(open-sse/handlers/chatCore.ts)
    participant Combo as Combo Engine<br/>(open-sse/services/combo.ts)
    participant Policy as Policy Engine<br/>(src/domain/policyEngine.ts)
    participant Guard as Guardrails<br/>(src/lib/guardrails/)
    participant Compress as Compression<br/>(open-sse/services/compression/)
    participant Trans as Translator<br/>(open-sse/translator/)
    participant Exec as Executor<br/>(open-sse/executors/)
    participant Upstream as Upstream Provider<br/>(OpenAI / Claude / etc.)
    participant DB as SQLite<br/>(src/lib/db/)

    Client->>Next: POST /v1/chat/completions<br/>Authorization: Bearer sk-...<br/>{model: "gpt-4o", messages: [...], stream: true}
    
    Note over Next: Rewrite: /v1/* → /api/v1/*
    Next->>Route: Forward to /api/v1/chat/completions
    
    Note over Route: 1. CORS
    Note over Route: 2. Zod validation
    Note over Route: 3. API Key extract + validate
    
    Route->>DB: SHA256(key) → lookup api_keys
    DB-->>Route: Key: id, user_id, scopes
    
    Route->>Handler: Delegate handleChat()
    
    alt Model starts with "auto/" → Combo Route
        Handler->>Combo: handleComboChat(model="auto/cheap")
        Combo->>DB: Load combo config
        DB-->>Combo: Combo: {strategy, targets, weights}
        Combo->>Combo: resolveComboTargets()
        Note over Combo: Returns ordered targets:<br/>[glm/glm-5.1, minimax/..., kiro/...]
        Combo->>Combo: Try each target until success
    else Single Model
        Handler->>Handler: Proceed with requested model
    end
    
    Handler->>Policy: Check lockout, budget, fallback
    Policy->>DB: Read domain state
    DB-->>Policy: State: OK to proceed
    Policy-->>Handler: Allow
    
    Handler->>Guard: Check content safety
    Guard->>Guard: PII detection
    Guard->>Guard: Injection detection
    Guard-->>Handler: Safe to process
    
    Handler->>Compress: Compress prompt (if configured)
    Compress->>Compress: Strategy selection
    Compress->>Compress: Caveman / RTK pipeline
    Compress-->>Handler: Compressed messages
    
    Handler->>Trans: translateRequest(body, sourceFormat, targetFormat)
    Note over Trans: Detect source format (OpenAI)<br/>Convert to target format (Claude format)
    Trans-->>Handler: Translated body
    
    Handler->>Exec: getExecutor(provider) → execute()
    Note over Exec: buildUrl() → https://api.anthropic.com/v1/messages<br/>buildHeaders() → x-api-key, anthropic-version<br/>transformRequest() → Claude format
    
    Exec->>Exec: Circuit breaker check
    Exec->>Exec: Rate limit check
    
    loop Retry up to 3 times
        Exec->>Upstream: HTTP POST (streaming)
        
        alt 401 Unauthorized
            Upstream-->>Exec: 401
            Exec->>Exec: OAuth token refresh
            Note over Exec: If provider supports OAuth,<br/>refresh token and rebuild request
        else 429 / 5xx
            Upstream-->>Exec: Error
            Exec->>Exec: Exponential backoff<br/>100ms → 200ms → 400ms + jitter
        else 200 OK
            Upstream-->>Exec: SSE Stream
            Note over Upstream,Exec: event: message_start<br/>event: content_block_delta<br/>event: content_block_stop<br/>event: message_stop
        end
    end
    
    Exec->>Handler: Raw SSE chunks
    Handler->>Handler: Translate chunks back to client format
    Note over Handler: Claude SSE → OpenAI SSE format<br/>content_block_delta → choices[0].delta.content
    
    Handler->>Client: SSE Stream (OpenAI format)
    Note over Handler,Client: data: {choices:[{delta:{content:"Hello"}}]}<br/>data: [DONE]
    
    par Async Persistence
        Handler->>Handler: Extract usage (tokens, cost, latency)
        Handler->>DB: INSERT usage_history
        Note over Handler,DB: prompt_tokens, completion_tokens,<br/>total_tokens, cost, latency_ms,<br/>provider, model, strategy
    end
```

---

## 🏗️ LAYER-BY-LAYER WORKFLOW

```mermaid
flowchart LR
    subgraph L1["Layer 1: HTTP & Routing"]
        direction LR
        L1A["next.config.mjs<br/>Rewrite /v1/*"]
        L1B["Route Handler<br/>src/app/api/v1/*/route.ts"]
        L1C["Zod Validation<br/>Request Body"]
    end
    
    subgraph L2["Layer 2: Auth & Policy"]
        direction LR
        L2A["API Key Check<br/>extractApiKey()"]
        L2B["Policy Engine<br/>Lockout / Budget"]
        L2C["Guardrails<br/>PII / Injection"]
    end
    
    subgraph L3["Layer 3: Core Processing"]
        direction LR
        L3A["Model/Combo Resolution<br/>chatCore.ts"]
        L3B["Translation<br/>open-sse/translator/"]
        L3C["Compression<br/>Caveman / RTK"]
    end
    
    subgraph L4["Layer 4: Execution"]
        direction LR
        L4A["Executor Selection<br/>getExecutor()"]
        L4B["Provider Call<br/>HTTP POST + Retry"]
        L4C["Response Stream<br/>SSE Transform"]
    end
    
    subgraph L5["Layer 5: Persistence"]
        direction LR
        L5A["Usage Tracking<br/>Token Count"]
        L5B["SQLite Write<br/>83 Modules"]
        L5C["Async Logging<br/>Background"]
    end

    L1 --> L2 --> L3 --> L4 --> L5
```

---

## 🔧 PROVIDER EXECUTOR WORKFLOW

```mermaid
flowchart TB
    subgraph Select["Executor Selection"]
        A["Route Handler calls handleChat()"]
        B{"Model = Combo?"}
        B -->|"Yes"| C["combo.ts<br/>resolveComboTargets()"]
        B -->|"No"| D["Single provider flow"]
        C --> E["For each target in order:"]
    end

    subgraph Execute["Execution Loop"]
        F["Get Provider Config<br/>providerRegistry.ts"]
        G["Select Account<br/>accountSelector.ts (P2C)"]
        H["Circuit Breaker Check<br/>OPEN?"]
        H -->|"OPEN"| I["Skip → Next Target"]
        H -->|"CLOSED"| J
        I --> E
        J["Rate Limit Check<br/>rateLimitManager.ts"]
        J -->|"Under Limit"| K["Get Executor<br/>getExecutor(provider)"]
        J -->|"Over Limit"| L["Skip → Next Target"]
        L --> E
    end

    subgraph Build["Request Building"]
        K --> M["buildUrl()<br/>e.g. POST https://api.anthropic.com/v1/messages"]
        M --> N["buildHeaders()<br/>Auth headers + version"]
        N --> O["transformRequest()<br/>Adapt body for provider"]
    end

    subgraph Call["Upstream Call"]
        O --> P["HTTP POST with retry"]
        P --> Q{"Response?"}
        Q -->|"401"| R["Token Refresh<br/>refreshCredentials()"]
        R --> P
        Q -->|"429 / 5xx"| S["Backoff + Retry<br/>max 3 attempts"]
        S --> P
        Q -->|"200 OK"| T["Process Response"]
    end

    subgraph Process["Response Processing"]
        T --> U{"Streaming?"}
        U -->|"Yes"| V["SSE Parser<br/>Parse events"]
        U -->|"No"| W["JSON Parser<br/>Parse body"]
        V --> X["Translate Response<br/>Back to client format"]
        W --> X
        X --> Y["Stream to Client<br/>responsesTransformer.ts"]
        Y --> Z["Usage Extraction<br/>→ SQLite"]
    end
```

---

## 🗄️ DATA FLOW WORKFLOW (SQLite Reads/Writes)

```mermaid
flowchart TB
    subgraph Reads["Read Path (Combo Resolution)"]
        direction LR
        R1["chatCore.ts<br/>handleChatCore()"] --> R2["combo.ts<br/>handleComboChat()"]
        R2 --> R3["src/lib/db/combos.ts<br/>getComboByKey()"]
        R3 --> R4["SQLite<br/>combos table"]
        R4 --> R5["Combo Config<br/>strategy + targets"]
        R5 --> R2
        R2 --> R6["src/lib/db/providers.ts<br/>getProvider()"]
        R6 --> R4
        R4 --> R7["Provider Config<br/>base_url + auth_type"]
    end

    subgraph Writes["Write Path (Usage Logging)"]
        direction LR
        W1["Executor<br/>Response received"] --> W2["usageExtractor.ts<br/>Extract tokens"]
        W2 --> W3["Calculate cost<br/>provider pricing"]
        W3 --> W4["src/lib/db/usage*.ts<br/>INSERT usage_history"]
        W4 --> W5["SQLite<br/>usage_history table"]
    end

    subgraph State["State Path (Policy/Resilience)"]
        direction LR
        S1["policyEngine.ts<br/>Check lockout/budget"] --> S2["src/lib/db/domainState.ts<br/>read state"]
        S2 --> S3["SQLite<br/>domain_state table"]
        S3 --> S4["Return: ALLOW / DENY"]
        S4 --> S1
        S1 --> S5["circuitBreaker.ts<br/>Check provider"]
        S5 --> S6["src/lib/db/circuitBreakers.ts<br/>read/write"]
    end
```

---

## 🖥️ FRONTEND WORKFLOW (Dashboard)

```mermaid
sequenceDiagram
    participant User as User
    participant Browser as Browser
    participant React as Next.js React<br/>Server Component
    participant Client as Client Component<br/>(use client)
    participant API as API Route<br/>(Management)
    participant DB as SQLite

    User->>Browser: Navigate to /dashboard/providers
    
    Browser->>React: Request page
    React->>API: Server-side fetch /api/providers
    API->>DB: SELECT * FROM provider_connections
    DB-->>API: Provider list
    API-->>React: JSON response
    Note over React: SSR: Render ProviderCards
    
    React-->>Browser: HTML + hydration script
    
    Note over Browser: Page Hydrates
    React-->>Client: Client Component mounts
    
    User->>Client: Click "Add Provider"
    Client->>Client: Show AddProviderModal
    
    User->>Client: Fill form + submit
    Client->>API: POST /api/providers {name, base_url, api_key}
    API->>DB: INSERT provider_connections
    DB-->>API: Success
    API-->>Client: 200 {id, name, ...}
    
    Client->>Client: Close modal + refresh list
    Client->>API: GET /api/providers
    API-->>Client: Updated list
    Client->>Client: Re-render ProviderList
```

---

## 📡 MCP TOOL WORKFLOW

```mermaid
sequenceDiagram
    participant MCPClient as MCP Client<br/>(Claude Code / Cursor)
    participant MCPServer as MCP Server<br/>(open-sse/mcp-server/)
    participant Auth as Auth/Scope Check
    participant Tool as Tool Handler<br/>(94 tools)
    participant Service as Service Layer<br/>(open-sse/services/)
    participant DB as SQLite

    MCPClient->>MCPServer: SSE Connect<br/>GET /api/mcp/sse
    
    Note over MCPServer: Session established<br/>Send endpoint: POST /api/mcp/message
    
    MCPClient->>MCPServer: JSON-RPC Request<br/>{method: "tools/call",<br/> params: {name: "list_combos",<br/>          arguments: {...}}}
    
    MCPServer->>Auth: Check scopes
    Auth->>Auth: Validate key has "combos:read" scope
    Auth-->>MCPServer: Allowed
    
    MCPServer->>Tool: Lookup tool by name
    Note over Tool: Tool registry has 94 tools<br/>Each has: name, description,<br/>inputSchema (Zod), handler
    
    Tool->>Tool: Zod validate arguments
    
    Tool->>Service: Call business logic
    Service->>DB: Query combos table
    DB-->>Service: Combo data
    Service-->>Tool: Results
    
    Tool-->>MCPServer: {content: [{type: "text", text: "..."}]}
    
    MCPServer->>DB: INSERT mcp_audit<br/>(tool_name, args, success, timestamp)
    
    MCPServer-->>MCPClient: JSON-RPC Response<br/>{result: {content: [...]}}
```

---

## ⚡ STREAMING SSE WORKFLOW

```mermaid
flowchart LR
    subgraph Incoming["Incoming SSE from Provider"]
        A1["event: message_start<br/>data: {type: message_start, ...}"]
        A2["event: content_block_start<br/>data: {type: text, text: ''}"]
        A3["event: content_block_delta<br/>data: {type: text_delta, text: 'Hello'}"]
        A4["event: content_block_delta<br/>data: {type: text_delta, text: '! How can'}"]
        A5["event: content_block_stop<br/>data: {type: content_block_stop}"]
        A6["event: message_delta<br/>data: {stop_reason: end_turn}"]
        A7["event: message_stop<br/>data: {type: message_stop}"]
    end

    subgraph Parser["SSE Parser (sseParser.ts)"]
        B["Read event stream<br/>Parse event: + data:"]
    end

    subgraph Transform["Transformer (responsesTransformer.ts)"]
        C{"Event Type?"}
        C -->|"content_block_delta<br/>text_delta"| D["→ OpenAI SSE format<br/>data: {choices:[{delta:{content:'Hello'}}]}"]
        C -->|"content_block_start<br/>tool_use"| E["→ OpenAI tool_calls<br/>data: {choices:[{delta:{tool_calls:...}}]}"]
        C -->|"message_stop"| F["→ [DONE]<br/>data: [DONE]"]
        C -->|"error"| G["→ Error event<br/>data: {error: {message: ...}}"]
    end

    subgraph Output["Output to Client"]
        H["Write transformed<br/>SSE event to response"]
        I["Flush to client"]
    end

    A1 --> B
    A2 --> B
    A3 --> B
    A4 --> B
    A5 --> B
    A6 --> B
    A7 --> B
    
    B --> C
    D --> H --> I
    E --> H --> I
    F --> H --> I
    G --> H --> I
```

---

## 🗺️ FULL SYSTEM WORKFLOW MAP

```mermaid
flowchart TB
    subgraph Input["📥 Input"]
        I1["HTTP Request<br/>POST /v1/chat/completions"]
        I2["MCP Request<br/>JSON-RPC via SSE"]
        I3["A2A Request<br/>POST /a2a"]
        I4["Dashboard UI<br/>Browser"]
        I5["WebSocket<br/>WS /v1/ws"]
    end

    subgraph Gateway["🚪 Gateway"]
        G1["next.config.mjs<br/>Rewrite rules"]
        G2["CORS<br/>Headers"]
        G3["Rate Limit<br/>(optional)"]
    end

    subgraph AuthLayer["🔐 Auth & Validation"]
        A1["Zod Schema<br/>Validation"]
        A2["API Key Check<br/>SHA256 → SQLite"]
        A3["Session Check<br/>Dashboard cookies"]
        A4["Authz Pipeline<br/>Classify → Enforce"]
    end

    subgraph RouteLayer["📡 Route Dispatch"]
        R1["/api/v1/*<br/>Compatibility API"]
        R2["/api/*<br/>Management API"]
        R3["/api/mcp/*<br/>MCP Server"]
        R4["/a2a<br/>A2A Server"]
    end

    subgraph Core["🧩 Core Engine (open-sse/)"]
        C1["chatCore.ts<br/>Orchestration"]
        C2["combo.ts<br/>Routing Engine"]
        C3["autoCombo/<br/>9-Factor Scoring"]
        C4["translator/<br/>Format Conversion"]
        C5["executors/<br/>68 Providers"]
        C6["mcp-server/<br/>94 Tools"]
        C7["services/<br/>115+ Modules"]
    end

    subgraph Domain["🏛️ Domain Layer"]
        D1["policyEngine.ts<br/>Policy Rules"]
        D2["costRules.ts<br/>Cost Calculation"]
        D3["fallbackPolicy.ts<br/>Fallback Logic"]
        D4["lockoutPolicy.ts<br/>Lockout Rules"]
    end

    subgraph Lib["📚 Library Layer"]
        L1["guardrails/<br/>PII / Injection / Vision"]
        L2["memory/<br/>FTS5 + Vector"]
        L3["skills/<br/>Registry + Executor"]
        L4["compression/<br/>Caveman / RTK"]
        L5["webhooks/<br/>HMAC + Retry"]
        L6["oauth/<br/>17 OAuth Providers"]
        L7["a2a/<br/>JSON-RPC + Skills"]
    end

    subgraph DB["💾 Persistence"]
        DB1["core.ts<br/>SQLite Singleton"]
        DB2["83 Domain Modules<br/>17 Tables"]
        DB3["99 Migrations<br/>Version Tracking"]
    end

    subgraph Output["📤 Output"]
        O1["SSE Stream<br/>OpenAI Format"]
        O2["JSON Response"]
        O3["MCP JSON-RPC<br/>Response"]
        O4["A2A JSON-RPC<br/>Response"]
    end

    subgraph Upstream["🔌 Upstream Providers"]
        U1["OpenAI / Anthropic / Gemini"]
        U2["DeepSeek / Groq / xAI"]
        U3["Mistral / Cohere / Together"]
        U4["OAuth Providers (14)"]
        U5["Self-Hosted (8+)"]
        U6["Free Providers (3)"]
    end

    %% Connections
    Input --> Gateway
    Gateway --> AuthLayer
    
    I1 --> G1
    I2 --> G2
    I3 --> G2
    I4 --> G2
    I5 --> G1
    
    G1 --> A1
    G2 --> A2
    G3 --> A2
    
    A1 --> A2
    A2 --> A4
    A3 --> A4
    A4 --> RouteLayer
    
    RouteLayer --> Core
    
    R1 --> C1
    R1 --> C2
    R3 --> C6
    R4 --> L7
    
    C1 --> C2
    C2 --> C3
    C1 --> C4
    C1 --> C5
    C5 --> C7
    
    Core --> Domain
    Core --> Lib
    Core --> DB
    
    C1 --> D1
    C1 --> D2
    C1 --> D3
    C5 --> L1
    C1 --> L2
    C1 --> L3
    C1 --> L4
    
    C5 --> Upstream
    Upstream --> Output
    
    Domain --> DB
    Lib --> DB
    
    Output --> O1
    Output --> O2
    Output --> O3
    Output --> O4
    
    O1 --> I1
    O2 --> I1
    O3 --> I2
    O4 --> I3
```

---

## 📁 KEY FILES IN THE WORKFLOW

| Step | File | What It Does |
|------|------|-------------|
| **Entry** | `next.config.mjs` | Rewrites `/v1/*` → `/api/v1/*` |
| **Route** | `src/app/api/v1/chat/completions/route.ts` | CORS, Zod validation, auth check, delegate |
| **Handler** | `open-sse/handlers/chatCore.ts` | Main orchestration — resolves model, calls services |
| **Combo** | `open-sse/services/combo.ts` | Resolves combo targets, iterates until success |
| **Scoring** | `open-sse/services/autoCombo/` | 9-factor ML scoring engine |
| **Policy** | `src/domain/policyEngine.ts` | Lockout, budget, fallback checks |
| **Guard** | `src/lib/guardrails/` | PII masking, injection detection |
| **Compress** | `open-sse/services/compression/` | Caveman / RTK engines |
| **Translate** | `open-sse/translator/index.ts` | Format conversion between APIs |
| **Execute** | `open-sse/executors/base.ts` | HTTP call with retry, circuit breaker |
| **Stream** | `open-sse/transformer/responsesTransformer.ts` | SSE chunk transformation |
| **Persist** | `src/lib/db/usage*.ts` | Usage logging to SQLite |

---

> **See also:**
> - `CURRENT_FRONTEND_BACKEND_ARCHITECTURE.md` — Full component-level architecture
> - `GOLANG_MIGRATION_ROADMAP.md` — Future Go migration plan