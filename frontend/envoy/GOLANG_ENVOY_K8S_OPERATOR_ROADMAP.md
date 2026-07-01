# 🚀 OmniRoute → Go + Envoy + Kubernetes Operator Migration (7-Month Plan)

> Complete migration of OmniRoute from TypeScript/Next.js to **Go** with **Envoy Control Plane**, **Kubernetes Operator pattern controllers**, **slog** structured logging, and **xDS** dynamic configuration.
>
> Every day mapped out for 7 months (≈210 days).

---

## 🎯 Target Architecture

```
┌────────────────────────────────────────────────────────────────────────────┐
│                         CLIENTS / IDEs / TOOLS                            │
│    Claude Code  Codex CLI  Cursor  Cline  OpenAI SDK  Browser            │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                         ENVOY PROXY (L7 Gateway)                          │
│                                                                             │
│  ┌────────────────────────────────────────────────────────────────────┐   │
│  │  Envoy Filters (applied in order):                                 │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐ │   │
│  │  │ TLS      │→│ ext_authz│→│ Rate     │→│ LUA      │→│ Router  │ │   │
│  │  │ Term.    │ │ (slog)   │ │ Limit    │ │ (Trans.) │ │         │ │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └─────────┘ │   │
│  └────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  Dynamic config via xDS (Aggregated Discovery Service)                     │
│  ┌────────────────────────────────────────────────────────────────────┐   │
│  │  xDS Resources: LDS (Listeners), RDS (Routes), CDS (Clusters),   │   │
│  │                  EDS (Endpoints), SDS (Secrets), SRDS (Scoped Rds)│   │
│  └────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────┬─────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                    GO CONTROL PLANE — xDS Server                           │
│                                                                             │
│  ┌────────────────────────────────────────────────────────────────────┐   │
│  │  go-control-plane (envoy xDS v3)                                    │   │
│  │  - Snapshot cache for atomic config updates                        │   │
│  │  - Delta gRPC + SotW (State-of-the-World) support                  │   │
│  │  - All controllers push updates through the xDS cache              │   │
│  └────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────┬─────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────────────────┐
│          GO BACKEND — Controller Pattern (like Kubernetes Operators)       │
│                                                                             │
│  Each controller watches its own resource, reconciles state,               │
│  pushes updates to xDS cache / SQLite / Redis                             │
│                                                                             │
│  ┌────────────────────────────────────────────────────────────────────┐   │
│  │  Controller Runtime Pattern:                                       │   │
│  │  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐       │   │
│  │  │  Watcher     │────▶│  Reconciler  │────▶│  Status      │       │   │
│  │  │  (event src) │     │  (business)  │     │  (update)    │       │   │
│  │  └──────────────┘     └──────┬───────┘     └──────────────┘       │   │
│  │                              │                                     │   │
│  │                              ▼                                     │   │
│  │                     ┌────────────────┐                            │   │
│  │                     │  Side Effects  │                            │   │
│  │                     │  - xDS update  │                            │   │
│  │                     │  - DB persist  │                            │   │
│  │                     │  - Cache flush │                            │   │
│  │                     └────────────────┘                            │   │
│  └────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌────────────────────────────────────────────────────────────────────┐   │
│  │  Controllers:                                                      │   │
│  │                                                                     │   │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐          │   │
│  │  │  Auth Controller        │  │  Routing Controller     │          │   │
│  │  │  - API keys, OAuth, JWT │  │  - 17 strategies, combo│          │   │
│  │  │  - Envoy ext_authz gRPC │  │  - xDS route updates   │          │   │
│  │  └─────────────────────────┘  └─────────────────────────┘          │   │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐          │   │
│  │  │  Provider Controller    │  │  Executor Controller    │          │   │
│  │  │  - 237 providers        │  │  - 68 executors         │          │   │
│  │  │  - Health probes        │  │  - Retry, token refresh│          │   │
│  │  └─────────────────────────┘  └─────────────────────────┘          │   │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐          │   │
│  │  │  Translator Controller  │  │  Quota Controller       │          │   │
│  │  │  - Format conversion    │  │  - Rate limiting        │          │   │
│  │  │  - OpenAI↔Claude↔Gemini │  │  - Budget tracking      │          │   │
│  │  └─────────────────────────┘  └─────────────────────────┘          │   │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐          │   │
│  │  │  Resilience Controller  │  │  Compression Controller  │          │   │
│  │  │  - Circuit breakers     │  │  - RTK + Caveman        │          │   │
│  │  │  - Cooldown, lockout    │  │  - Stacked pipeline     │          │   │
│  │  └─────────────────────────┘  └─────────────────────────┘          │   │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐          │   │
│  │  │  MCP Controller         │  │  A2A Controller          │          │   │
│  │  │  - 94 tools, 30 scopes  │  │  - JSON-RPC 2.0, 6 skls │          │   │
│  │  └─────────────────────────┘  └─────────────────────────┘          │   │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐          │   │
│  │  │  Memory Controller      │  │  Skills Controller      │          │   │
│  │  │  - FTS5 + vector        │  │  - Registry, sandbox    │          │   │
│  │  └─────────────────────────┘  └─────────────────────────┘          │   │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐          │   │
│  │  │  Guardrails Controller  │  │  Sync Controller        │          │   │
│  │  │  - PII, injection, vis.│  │  - Cloud sync           │          │   │
│  │  └─────────────────────────┘  └─────────────────────────┘          │   │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐          │   │
│  │  │  Webhook Controller     │  │  Eval Controller        │          │   │
│  │  │  - HMAC, retry, disp.  │  │  - LLM quality scoring  │          │   │
│  │  └─────────────────────────┘  └─────────────────────────┘          │   │
│  │  ┌─────────────────────────┐  ┌─────────────────────────┐          │   │
│  │  │  Tunnel Controller      │  │  MITM Controller        │          │   │
│  │  │  - Cloudflare, ngrok    │  │  - Cert mgmt, TPROXY   │          │   │
│  │  └─────────────────────────┘  └─────────────────────────┘          │   │
│  └────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────┬─────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                   KUBERNETES OPERATOR (optional)                          │
│                                                                             │
│  Custom Resource Definitions (CRDs) for each controller:                   │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐     │
│  │ Provider     │ │ Combo        │ │ ApiKey       │ │ RateLimit    │     │
│  │ (providers.  │ │ (combos.     │ │ (apikeys.    │ │ (ratelimits. │     │
│  │  omniroute)  │ │  omniroute)  │ │  omniroute)  │ │  omniroute)  │     │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘     │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐     │
│  │ CircuitBrekr │ │ Webhook      │ │ MCPTool      │ │ Tunnel       │     │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘     │
│                                                                             │
│  Operator reconciles CRD state → updates controllers → xDS → Envoy       │
└────────────────────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                   PERSISTENCE LAYER                                        │
│                                                                             │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐         │
│  │  SQLite/Postgres  │  │  Redis Cache     │  │  Slog Logger     │         │
│  │  - providers      │  │  - Rate counters  │  │  - Structured    │         │
│  │  - combos         │  │  - Sessions      │  │  - JSON output   │         │
│  │  - api_keys       │  │  - CB state      │  │  - Levels:       │         │
│  │  - usage_history  │  │  - Provider hlth │  │    debug/info/   │         │
│  │  - mcp_audit      │  │  - Combo cache   │  │    warn/error    │         │
│  │  - webhooks       │  │                  │  │  - Attrs:        │         │
│  │  - memory (FTS5)  │  │                  │  │    req_id,       │         │
│  └──────────────────┘  └──────────────────┘  │    provider,      │         │
│                                                │    model,         │         │
│                                                │    latency,       │         │
│                                                │    tokens, cost   │         │
│                                                └──────────────────┘         │
└────────────────────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                     UPSTREAM PROVIDERS (237)                               │
│  OpenAI  Anthropic  Gemini  DeepSeek  Groq  xAI  Mistral  ...              │
│  Claude Code  Codex  Kiro  Qoder  LM Studio  vLLM  ...                    │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 📦 Go Project Structure

```
omniroute/
├── cmd/
│   ├── omniroute/                    # Main binary
│   │   └── main.go                   # Wires everything
│   ├── envoy-bootstrap/              # Envoy config generator
│   │   └── main.go
│   ├── operator/                     # Kubernetes operator binary
│   │   └── main.go
│   └── cli/                          # CLI subcommands
│       ├── main.go
│       ├── serve.go
│       ├── mcp.go
│       └── version.go
├── internal/
│   ├── api/                          # Shared API types (protobuf)
│   │   ├── v1/
│   │   │   ├── chat.proto
│   │   │   ├── models.proto
│   │   │   ├── providers.proto
│   │   │   └── common.proto
│   │   └── buf.gen.yaml
│   ├── envoy/                        # Envoy control plane
│   │   ├── xds_server.go             # xDS gRPC server
│   │   ├── snapshot_cache.go         # Snapshot cache
│   │   ├── resources.go              # Resource generators
│   │   ├── routes.go                 # Route config builder
│   │   ├── clusters.go               # Cluster config builder
│   │   ├── listeners.go             # Listener config builder
│   │   └── types.go                  # Envoy types
│   ├── controllers/                  # All controllers
│   │   ├── auth/
│   │   │   ├── controller.go         # Watcher + Reconciler
│   │   │   ├── reconciler.go         # Business logic
│   │   │   ├── store.go              # SQLite repository
│   │   │   ├── envoy_extauth.go      # gRPC ext_authz server
│   │   │   ├── oauth/
│   │   │   │   ├── handler.go        # OAuth handler interface
│   │   │   │   ├── claude.go
│   │   │   │   ├── codex.go
│   │   │   │   ├── gemini.go
│   │   │   │   ├── github.go
│   │   │   │   ├── kiro.go
│   │   │   │   ├── cursor.go
│   │   │   │   ├── antigravity.go
│   │   │   │   ├── qoder.go
│   │   │   │   ├── qwen.go
│   │   │   │   ├── kimi.go
│   │   │   │   ├── kilocode.go
│   │   │   │   ├── cline.go
│   │   │   │   ├── windsurf.go
│   │   │   │   ├── gitlab-duo.go
│   │   │   │   ├── trae.go
│   │   │   │   └── registry.go
│   │   │   └── test/
│   │   │       ├── controller_test.go
│   │   │       └── oauth_test.go
│   │   ├── routing/
│   │   │   ├── controller.go
│   │   │   ├── reconciler.go
│   │   │   ├── store.go
│   │   │   ├── combo_resolver.go
│   │   │   ├── auto_combo.go         # 9-factor scoring
│   │   │   ├── task_aware.go         # Workflow FSM
│   │   │   └── strategies/
│   │   │       ├── strategy.go       # Interface
│   │   │       ├── priority.go
│   │   │       ├── weighted.go
│   │   │       ├── round_robin.go
│   │   │       ├── fill_first.go
│   │   │       ├── p2c.go
│   │   │       ├── random.go
│   │   │       ├── least_used.go
│   │   │       ├── reset_aware.go
│   │   │       ├── reset_window.go
│   │   │       ├── cost_optimized.go
│   │   │       ├── strict_random.go
│   │   │       ├── auto.go
│   │   │       ├── lkgp.go
│   │   │       ├── context_optimized.go
│   │   │       ├── context_relay.go
│   │   │       ├── fusion.go
│   │   │       └── test/
│   │   └── provider/
│   │       ├── controller.go
│   │       ├── reconciler.go
│   │       ├── store.go
│   │       ├── registry.go           # 237 providers
│   │       ├── health_checker.go
│   │       └── test/
│   ├── executor/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── base_executor.go
│   │   ├── default.go               # OpenAI-compatible
│   │   ├── anthropic.go
│   │   ├── gemini.go
│   │   ├── claude_web.go
│   │   ├── codex.go
│   │   ├── cursor.go
│   │   ├── kiro.go
│   │   ├── qoder.go
│   │   ├── pollinations.go
│   │   ├── puter.go
│   │   ├── deepseek.go
│   │   ├── groq.go
│   │   ├── xai.go
│   │   ├── mistral.go
│   │   ├── together.go
│   │   ├── fireworks.go
│   │   ├── cerebras.go
│   │   ├── cohere.go
│   │   ├── nvidia.go
│   │   ├── nebius.go
│   │   ├── siliconflow.go
│   │   ├── hyperbolic.go
│   │   ├── huggingface.go
│   │   ├── openrouter.go
│   │   ├── vertex.go
│   │   ├── cloudflare.go
│   │   ├── scaleway.go
│   │   ├── pollinations.go
│   │   ├── longcat.go
│   │   ├── alibaba.go
│   │   ├── kimi.go
│   │   ├── minimax.go
│   │   ├── blackbox.go
│   │   ├── synthetic.go
│   │   ├── kilogateway.go
│   │   ├── glm.go
│   │   ├── deepgram.go
│   │   ├── assemblyai.go
│   │   ├── elevenlabs.go
│   │   ├── cartesia.go
│   │   ├── playht.go
│   │   ├── inworld.go
│   │   ├── nanobanana.go
│   │   ├── sdwebui.go
│   │   ├── comfyui.go
│   │   ├── ollamacloud.go
│   │   ├── deepinfra.go
│   │   ├── vercel_aigw.go
│   │   ├── lambda.go
│   │   ├── sambanova.go
│   │   ├── nscale.go
│   │   ├── ovhcloud.go
│   │   ├── baseten.go
│   │   ├── publicai.go
│   │   ├── moonshot.go
│   │   ├── meta_llama.go
│   │   ├── v0.go
│   │   ├── morph.go
│   │   ├── featherless.go
│   │   ├── friendli.go
│   │   ├── llamagate.go
│   │   ├── galadriel.go
│   │   ├── weights_biases.go
│   │   ├── volcengine.go
│   │   ├── ai21.go
│   │   ├── venice.go
│   │   ├── codestral.go
│   │   ├── upstage.go
│   │   ├── maritalk.go
│   │   ├── xiaomi.go
│   │   ├── inference_net.go
│   │   ├── nanogpt.go
│   │   ├── predibase.go
│   │   ├── bytez.go
│   │   ├── heroku.go
│   │   ├── databricks.go
│   │   ├── snowflake.go
│   │   ├── gigachat.go
│   │   ├── crofai.go
│   │   ├── agentrouter.go
│   │   ├── chatgptweb.go
│   │   ├── baidu.go
│   │   ├── awspolly.go
│   │   ├── runwayml.go
│   │   ├── gitlabduo.go
│   │   ├── amazonq.go
│   │   ├── empower.go
│   │   ├── poe.go
│   │   └── ... (remaining executors)
│   ├── translator/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── request.go               # Request translation
│   │   ├── response.go              # Response translation
│   │   ├── openai_to_claude.go
│   │   ├── openai_to_gemini.go
│   │   ├── claude_to_openai.go
│   │   ├── gemini_to_openai.go
│   │   ├── image_handler.go
│   │   ├── audio_handler.go
│   │   ├── video_handler.go
│   │   ├── think_tag_parser.go
│   │   ├── role_normalizer.go
│   │   ├── structured_output.go
│   │   └── test/
│   ├── quota/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── rate_limiter.go           # Token bucket
│   │   ├── budget_tracker.go
│   │   ├── token_counter.go
│   │   ├── envoy_rate_limit.go       # gRPC rate limit service
│   │   └── test/
│   ├── resilience/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── circuit_breaker.go
│   │   ├── cooldown.go
│   │   ├── lockout.go
│   │   ├── herd_protection.go
│   │   ├── domain_state.go
│   │   └── test/
│   ├── compression/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── caveman.go
│   │   ├── caveman_rules.go
│   │   ├── rtk.go
│   │   ├── rtk_filters.go
│   │   ├── pipeline.go               # Stacked engines
│   │   ├── language_packs.go
│   │   ├── stats.go
│   │   └── test/
│   ├── mcp/
│   │   ├── controller.go
│   │   ├── server.go                 # Transport handler
│   │   ├── stdio.go                  # stdio transport
│   │   ├── sse.go                    # SSE transport
│   │   ├── streamable_http.go        # Streamable HTTP transport
│   │   ├── tools.go                  # Tool registry (94)
│   │   ├── scopes.go                 # 30 auth scopes
│   │   ├── audit.go                  # MCP audit logging
│   │   └── tools/
│   │       ├── get_health.go
│   │       ├── list_combos.go
│   │       ├── get_combo_metrics.go
│   │       ├── switch_combo.go
│   │       ├── check_quota.go
│   │       ├── route_request.go
│   │       ├── cost_report.go
│   │       ├── list_models.go
│   │       ├── web_search.go
│   │       ├── simulate_route.go
│   │       ├── set_budget_guard.go
│   │       ├── set_routing_strategy.go
│   │       ├── set_resilience_profile.go
│   │       ├── test_combo.go
│   │       ├── get_provider_metrics.go
│   │       ├── best_combo_for_task.go
│   │       ├── explain_route.go
│   │       ├── get_session_snapshot.go
│   │       ├── db_health_check.go
│   │       ├── sync_pricing.go
│   │       ├── cache_stats.go
│   │       ├── cache_flush.go
│   │       ├── compression_status.go
│   │       ├── compression_configure.go
│   │       ├── set_compression_engine.go
│   │       ├── list_compression_combos.go
│   │       ├── compression_combo_stats.go
│   │       ├── oneproxy_fetch.go
│   │       ├── oneproxy_rotate.go
│   │       ├── oneproxy_stats.go
│   │       ├── memory_search.go
│   │       ├── memory_add.go
│   │       ├── memory_clear.go
│   │       ├── skills_list.go
│   │       ├── skills_enable.go
│   │       ├── skills_execute.go
│   │       ├── skills_executions.go
│   │       ├── notion_list.go
│   │       ├── notion_read.go
│   │       ├── notion_write.go
│   │       ├── notion_search.go
│   │       ├── notion_delete.go
│   │       ├── notion_query.go
│   │       ├── obsidian_search.go
│   │       ├── obsidian_read.go
│   │       ├── obsidian_write.go
│   │       ├── obsidian_delete.go
│   │       ├── obsidian_list.go
│   │       ├── obsidian_tags.go
│   │       ├── obsidian_graph.go
│   │       ├── obsidian_backlinks.go
│   │       ├── obsidian_templates.go
│   │       ├── obsidian_daily_note.go
│   │       ├── obsidian_weekly_note.go
│   │       ├── obsidian_monthly_note.go
│   │       ├── obsidian_canvas.go
│   │       ├── obsidian_properties.go
│   │       ├── obsidian_webdav_list.go
│   │       ├── obsidian_webdav_read.go
│   │       ├── obsidian_webdav_write.go
│   │       ├── obsidian_webdav_delete.go
│   │       ├── obsidian_webdav_mkdir.go
│   │       ├── obsidian_webdav_copy.go
│   │       ├── obsidian_webdav_move.go
│   │       ├── gamification_levels.go
│   │       ├── gamification_badges.go
│   │       ├── gamification_leaderboard.go
│   │       ├── gamification_community.go
│   │       ├── plugin_list.go
│   │       ├── plugin_install.go
│   │       ├── plugin_enable.go
│   │       ├── plugin_disable.go
│   │       ├── plugin_inspect.go
│   │       ├── plugin_marketplace.go
│   │       ├── plugin_config.go
│   │       ├── plugin_uninstall.go
│   │       └── ... (up to 94)
│   ├── a2a/
│   │   ├── controller.go
│   │   ├── server.go                 # JSON-RPC 2.0 handler
│   │   ├── task_manager.go           # Task lifecycle
│   │   ├── agent_card.go             # /.well-known/agent.json
│   │   └── skills/
│   │       ├── smart_routing.go
│   │       ├── quota_management.go
│   │       ├── provider_discovery.go
│   │       ├── cost_analysis.go
│   │       ├── health_report.go
│   │       └── list_capabilities.go
│   ├── memory/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── extractor.go
│   │   ├── injector.go
│   │   ├── retriever.go
│   │   ├── summarizer.go
│   │   ├── store.go                  # FTS5 + vector
│   │   └── test/
│   ├── skills/
│   │   ├── controller.go
│   │   ├── registry.go
│   │   ├── executor.go
│   │   ├── sandbox.go
│   │   ├── builtin/
│   │   │   ├── quota.go
│   │   │   ├── routing.go
│   │   │   └── health.go
│   │   ├── custom/
│   │   │   └── handler.go
│   │   └── test/
│   ├── guardrails/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── pii_masker.go
│   │   ├── prompt_injection.go
│   │   ├── vision_bridge.go
│   │   ├── hot_reload.go
│   │   └── test/
│   ├── webhook/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── dispatcher.go
│   │   ├── hmac.go
│   │   ├── retry.go
│   │   ├── store.go
│   │   └── test/
│   ├── tunnel/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── cloudflare.go
│   │   ├── ngrok.go
│   │   ├── tailscale.go
│   │   └── test/
│   ├── mitm/
│   │   ├── controller.go
│   │   ├── cert_manager.go
│   │   ├── tproxy.go
│   │   └── test/
│   ├── sync/
│   │   ├── controller.go
│   │   ├── reconciler.go
│   │   ├── cloud.go
│   │   └── store.go
│   ├── eval/
│   │   ├── controller.go
│   │   ├── runner.go
│   │   ├── runtime.go
│   │   └── store.go
│   ├── persistence/
│   │   ├── sqlite/
│   │   │   ├── db.go                 # Singleton + WAL
│   │   │   ├── migrations.go         # 99 migrations
│   │   │   ├── repositories/
│   │   │   │   ├── provider_repo.go
│   │   │   │   ├── combo_repo.go
│   │   │   │   ├── api_key_repo.go
│   │   │   │   ├── usage_repo.go
│   │   │   │   ├── settings_repo.go
│   │   │   │   ├── webhook_repo.go
│   │   │   │   ├── mcp_audit_repo.go
│   │   │   │   ├── memory_repo.go
│   │   │   │   ├── eval_repo.go
│   │   │   │   └── ... (83 repos)
│   │   │   └── models/
│   │   │       ├── provider.go
│   │   │       ├── combo.go
│   │   │       ├── api_key.go
│   │   │       └── ... (model structs)
│   │   └── postgres/
│   │       ├── db.go                 # pgx v5
│   │       └── ... (same repos)
│   ├── cache/
│   │   ├── redis.go                  # go-redis
│   │   ├── rate_limiter.go           # Redis-based
│   │   └── session.go                # Session store
│   ├── streaming/
│   │   ├── sse.go                    # SSE writer
│   │   ├── sse_parser.go            # SSE reader
│   │   ├── websocket.go              # WS handler
│   │   └── transformer.go            # Responses API transformer
│   └── pkg/
│       ├── logger/
│       │   └── slog.go               # Slog setup + helpers
│       ├── config/
│       │   └── config.go             # koanf config
│       ├── errors/
│       │   └── errors.go             # Error types
│       ├── middleware/
│       │   ├── cors.go
│       │   ├── logging.go
│       │   ├── recovery.go
│       │   └── request_id.go
│       ├── validator/
│       │   └── validator.go          # Zod-equivalent validation
│       ├── httputil/
│       │   ├── response.go
│       │   ├── stream.go
│       │   └── headers.go
│       └── testutil/
│           ├── fixtures.go
│           └── mocks.go
├── operator/                         # Kubernetes Operator
│   ├── api/
│   │   └── v1alpha1/
│   │       ├── provider_types.go
│   │       ├── combo_types.go
│   │       ├── apikey_types.go
│   │       ├── circuitbreaker_types.go
│   │       ├── ratelimit_types.go
│   │       ├── webhook_types.go
│   │       ├── mcptool_types.go
│   │       ├── tunnel_types.go
│   │       ├── groupversion_info.go
│   │       └── zz_generated.deepcopy.go
│   ├── controllers/                  # Operator controllers
│   │   ├── provider_controller.go
│   │   ├── combo_controller.go
│   │   ├── apikey_controller.go
│   │   ├── circuitbreaker_controller.go
│   │   ├── ratelimit_controller.go
│   │   ├── webhook_controller.go
│   │   ├── mcptool_controller.go
│   │   └── tunnel_controller.go
│   ├── webhooks/                     # Admission webhooks
│   │   └── provider_webhook.go
│   ├── config/
│   │   ├── crd/                      # CRD YAML manifests
│   │   ├── rbac/                     # RBAC manifests
│   │   └── manager/                  # Controller manager config
│   └── main.go                       # Operator entry
├── dashboard/                        # Next.js frontend
│   └── ... (unchanged)
├── envoy/                            # Envoy configs
│   ├── envoy.yaml                    # Bootstrap config
│   └── templates/                    # Config templates
├── proto/                            # Protobuf definitions
│   ├── envoy/
│   │   ├── ext_authz.proto
│   │   ├── rate_limit.proto
│   │   └── xds.proto
│   └── omniroute/
│       ├── v1/
│       └── common/
├── config/
│   ├── config.yaml                   # Default config
│   ├── providers.yaml                # Provider definitions
│   └── strategies.yaml               # Strategy definitions
├── scripts/
│   ├── build.sh
│   ├── test.sh
│   ├── generate.sh                   # protobuf generation
│   └── dev.sh                        # Local dev setup
├── Makefile
├── go.mod
├── go.sum
├── Dockerfile
├── Dockerfile.operator
└── README.md
```

---

## 🪵 Slog Logging Pattern

```go
// internal/pkg/logger/slog.go
package logger

import (
    "context"
    "log/slog"
    "os"
    "runtime"
    "time"
)

// Setup configures slog with the specified level and format
func Setup(level string, json bool) {
    var l slog.Level
    switch level {
    case "debug":
        l = slog.LevelDebug
    case "info":
        l = slog.LevelInfo
    case "warn":
        l = slog.LevelWarn
    case "error":
        l = slog.LevelError
    default:
        l = slog.LevelInfo
    }

    var handler slog.Handler
    if json {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level:       l,
            AddSource:   true,
            ReplaceAttr: replaceAttrs,
        })
    } else {
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level:       l,
            AddSource:   true,
            ReplaceAttr: replaceAttrs,
        })
    }

    slog.SetDefault(slog.New(handler))
}

func replaceAttrs(groups []string, a slog.Attr) slog.Attr {
    // Remove timestamp (handled by JSON serializer)
    if a.Key == "time" {
        return slog.Attr{}
    }
    return a
}

// RequestLogger adds request-scoped context to slog
type RequestLogger struct {
    RequestID string
    Provider  string
    Model     string
    UserID    string
}

func (l *RequestLogger) Log(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
    attrs = append(attrs,
        slog.String("req_id", l.RequestID),
        slog.String("provider", l.Provider),
        slog.String("model", l.Model),
        slog.String("user_id", l.UserID),
    )
    slog.LogAttrs(ctx, level, msg, attrs...)
}

func (l *RequestLogger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
    l.Log(ctx, slog.LevelInfo, msg, attrs...)
}

func (l *RequestLogger) Error(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
    attrs = append(attrs, slog.Any("error", err))
    l.Log(ctx, slog.LevelError, msg, attrs...)
}

func (l *RequestLogger) Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
    l.Log(ctx, slog.LevelDebug, msg, attrs...)
}

func (l *RequestLogger) Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
    l.Log(ctx, slog.LevelWarn, msg, attrs...)
}

// Usage in controllers:
// logger.Info(ctx, "routing decision",
//     slog.String("strategy", "auto"),
//     slog.String("selected_provider", "glm"),
//     slog.String("selected_model", "glm-5.1"),
//     slog.Float64("score", 0.87),
//     slog.Duration("latency", 45*time.Millisecond),
// )
```

### Slog Integration in Controllers

```go
// Example: Auth Controller with slog
type AuthController struct {
    reconciler *AuthReconciler
    store      *AuthStore
    logger     *slog.Logger
}

func NewAuthController(store *AuthStore) *AuthController {
    return &AuthController{
        store:  store,
        logger: slog.With("controller", "auth"),
    }
}

func (c *AuthController) ValidateAPIKey(ctx context.Context, key string) (*APIKey, error) {
    keyHash := sha256.Sum256([]byte(key))
    apiKey, err := c.store.GetByHash(ctx, keyHash[:])
    if err != nil {
        c.logger.ErrorContext(ctx, "api key validation failed",
            slog.String("key_prefix", key[:8]),
            slog.Any("error", err),
        )
        return nil, err
    }
    c.logger.DebugContext(ctx, "api key validated",
        slog.String("key_id", apiKey.ID),
        slog.String("label", apiKey.Label),
    )
    return apiKey, nil
}
```

---

## 🎮 Controller Pattern (Kubernetes Operator Style)

```go
// internal/controllers/base.go — Generic controller pattern
package controllers

import (
    "context"
    "log/slog"
    "sync"
    "time"
)

// Controller is the base interface that all controllers implement
type Controller interface {
    Name() string
    Start(ctx context.Context) error
    Stop() error
}

// Reconciler processes events and returns the desired state
type Reconciler[T any] interface {
    Reconcile(ctx context.Context, req T) (Result, error)
}

type Result struct {
    Requeue      bool
    RequeueAfter time.Duration
}

// Watcher monitors for state changes
type Watcher[T any] struct {
    events chan T
    logger *slog.Logger
}

func NewWatcher[T any](buffer int) *Watcher[T] {
    return &Watcher[T]{
        events: make(chan T, buffer),
        logger: slog.With("component", "watcher"),
    }
}

func (w *Watcher[T]) Events() <-chan T {
    return w.events
}

func (w *Watcher[T]) Notify(event T) {
    select {
    case w.events <- event:
    default:
        w.logger.Warn("watcher buffer full, dropping event")
    }
}

// BaseController provides common controller infrastructure
type BaseController[T any] struct {
    name       string
    watcher    *Watcher[T]
    reconciler Reconciler[T]
    logger     *slog.Logger
    wg         sync.WaitGroup
    workers    int
}

func NewBaseController[T any](name string, reconciler Reconciler[T], workers int) *BaseController[T] {
    return &BaseController[T]{
        name:       name,
        watcher:    NewWatcher[T](1000),
        reconciler: reconciler,
        workers:    workers,
        logger:     slog.With("controller", name),
    }
}

func (bc *BaseController[T]) Name() string {
    return bc.name
}

func (bc *BaseController[T]) Start(ctx context.Context) error {
    bc.logger.InfoContext(ctx, "starting controller",
        slog.Int("workers", bc.workers),
    )
    for i := 0; i < bc.workers; i++ {
        bc.wg.Add(1)
        go bc.worker(ctx, i)
    }
    return nil
}

func (bc *BaseController[T]) Stop() error {
    bc.logger.Info("stopping controller")
    bc.wg.Wait()
    return nil
}

func (bc *BaseController[T]) worker(ctx context.Context, id int) {
    defer bc.wg.Done()
    l := bc.logger.With("worker_id", id)
    l.InfoContext(ctx, "worker started")

    for {
        select {
        case <-ctx.Done():
            l.InfoContext(ctx, "worker stopping")
            return
        case event := <-bc.watcher.Events():
            result, err := bc.reconciler.Reconcile(ctx, event)
            if err != nil {
                l.ErrorContext(ctx, "reconcile failed",
                    slog.Any("error", err),
                    slog.Any("event", event),
                )
            }
            if result.Requeue {
                time.AfterFunc(result.RequeueAfter, func() {
                    bc.watcher.Notify(event)
                })
            }
        }
    }
}

func (bc *BaseController[T]) Watch(event T) {
    bc.watcher.Notify(event)
}
```

### Concrete Controller Example: Auth Controller

```go
// internal/controllers/auth/controller.go
package auth

import (
    "context"
    "log/slog"
    "time"

    "omniroute/internal/controllers"
    "omniroute/internal/pkg/logger"
)

// AuthEvent represents a change in auth state
type AuthEvent struct {
    Type   string // "api_key_created", "api_key_revoked", "oauth_updated"
    KeyID  string
    UserID string
}

// AuthReconciler processes auth events
type AuthReconciler struct {
    store       *AuthStore
    envoyAuth   *EnvoyExtAuthzServer
    oauth       *OAuthRegistry
    requestLog  *logger.RequestLogger
}

func (r *AuthReconciler) Reconcile(ctx context.Context, event AuthEvent) (controllers.Result, error) {
    log := r.requestLog
    log.Info(ctx, "reconciling auth event",
        slog.String("event_type", event.Type),
        slog.String("key_id", event.KeyID),
    )

    switch event.Type {
    case "api_key_created":
        // Reload envoy ext_authz config
        return r.reloadEnvoyAuth(ctx)
    case "api_key_revoked":
        // Remove from envoy cache
        return r.invalidateKey(ctx, event.KeyID)
    case "oauth_updated":
        // Refresh OAuth tokens if needed
        return r.refreshOAuth(ctx, event.UserID)
    default:
        return controllers.Result{}, nil
    }
}

func (r *AuthReconciler) reloadEnvoyAuth(ctx context.Context) (controllers.Result, error) {
    keys, err := r.store.ListAll(ctx)
    if err != nil {
        return controllers.Result{}, err
    }
    r.envoyAuth.UpdateKeyCache(keys)
    return controllers.Result{}, nil
}

// AuthController wraps the base controller pattern
type AuthController struct {
    *controllers.BaseController[AuthEvent]
    store      *AuthStore
    oauth      *OAuthRegistry
}

func NewAuthController(store *AuthStore, oauth *OAuthRegistry) *AuthController {
    reconciler := &AuthReconciler{
        store:     store,
        oauth:     oauth,
        requestLog: &logger.RequestLogger{},
    }
    base := controllers.NewBaseController[AuthEvent]("auth", reconciler, 4)
    return &AuthController{
        BaseController: base,
        store:          store,
        oauth:          oauth,
    }
}
```

---

## 🌐 Envoy Control Plane (xDS Server)

```go
// internal/envoy/xds_server.go
package envoy

import (
    "context"
    "log/slog"
    "net"

    "google.golang.org/grpc"
    discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
    cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
    server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
    "github.com/envoyproxy/go-control-plane/pkg/log"
)

// XDSLogger adapts slog to go-control-plane's log interface
type XDSLogger struct {
    logger *slog.Logger
}

func (l *XDSLogger) Infof(format string, args ...interface{}) {
    l.logger.Info(format, slog.Any("args", args))
}

func (l *XDSLogger) Debugf(format string, args ...interface{}) {
    l.logger.Debug(format, slog.Any("args", args))
}

func (l *XDSLogger) Warnf(format string, args ...interface{}) {
    l.logger.Warn(format, slog.Any("args", args))
}

func (l *XDSLogger) Errorf(format string, args ...interface{}) {
    l.logger.Error(format, slog.Any("args", args))
}

// XDSServer manages the envoy control plane
type XDSServer struct {
    cache     cache.SnapshotCache
    server    server.Server
    grpc      *grpc.Server
    snapshot  *SnapshotManager
    logger    *slog.Logger
}

func NewXDSServer(port int) *XDSServer {
    l := slog.With("component", "xds_server")

    // Create snapshot cache
    snapshotCache := cache.NewSnapshotCache(false, cache.IDHash{}, l)

    // Create callback for connection tracking
    cb := &callback{l}

    // Create xDS server
    srv := server.NewServer(context.Background(), snapshotCache, cb)

    // Create gRPC server
    grpcServer := grpc.NewServer()

    // Register discovery services
    discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
    discovery.RegisterListenerDiscoveryServiceServer(grpcServer, srv)
    discovery.RegisterRouteDiscoveryServiceServer(grpcServer, srv)
    discovery.RegisterClusterDiscoveryServiceServer(grpcServer, srv)
    discovery.RegisterEndpointDiscoveryServiceServer(grpcServer, srv)
    discovery.RegisterSecretDiscoveryServiceServer(grpcServer, srv)

    return &XDSServer{
        cache:    snapshotCache,
        server:   srv,
        grpc:     grpcServer,
        snapshot: NewSnapshotManager(snapshotCache),
        logger:   l,
    }
}

func (x *XDSServer) Start(ctx context.Context) error {
    lis, err := net.Listen("tcp", ":18000")
    if err != nil {
        return err
    }
    x.logger.InfoContext(ctx, "xDS server listening", slog.String("addr", ":18000"))
    return x.grpc.Serve(lis)
}

func (x *XDSServer) Stop() {
    x.grpc.GracefulStop()
}

type callback struct {
    logger *slog.Logger
}

func (c *callback) OnStreamOpen(ctx context.Context, id int64, typ string) error {
    c.logger.DebugContext(ctx, "xDS stream open",
        slog.Int64("stream_id", id),
        slog.String("type", typ),
    )
    return nil
}

func (c *callback) OnStreamClosed(id int64) {
    c.logger.Debug("xDS stream closed", slog.Int64("stream_id", id))
}

func (c *callback) OnStreamRequest(id int64, req *discovery.DiscoveryRequest) error {
    c.logger.Debug("xDS stream request",
        slog.Int64("stream_id", id),
        slog.String("type", req.TypeUrl),
        slog.String("version", req.VersionInfo),
    )
    return nil
}

func (c *callback) OnStreamResponse(ctx context.Context, id int64, req *discovery.DiscoveryRequest, resp *discovery.DiscoveryResponse) {
    c.logger.DebugContext(ctx, "xDS stream response",
        slog.Int64("stream_id", id),
        slog.String("type", resp.TypeUrl),
        slog.Int("resources", len(resp.Resources)),
    )
}

func (c *callback) OnFetchRequest(ctx context.Context, req *discovery.DiscoveryRequest) error {
    return nil
}

func (c *callback) OnFetchResponse(req *discovery.DiscoveryRequest, resp *discovery.DiscoveryResponse) {
}
```

### Snapshot Manager — Pushes Updates to Envoy

```go
// internal/envoy/snapshot_manager.go
package envoy

import (
    "context"
    "log/slog"
    "time"

    corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
    listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
    routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
    clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
    cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
    "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
    "google.golang.org/protobuf/types/known/anypb"
    "google.golang.org/protobuf/types/known/durationpb"
)

type SnapshotManager struct {
    cache  cache.SnapshotCache
    logger *slog.Logger
}

func NewSnapshotManager(c cache.SnapshotCache) *SnapshotManager {
    return &SnapshotManager{
        cache:  c,
        logger: slog.With("component", "snapshot_manager"),
    }
}

// UpdateRoutes pushes new route configuration to Envoy
func (sm *SnapshotManager) UpdateRoutes(ctx context.Context, nodeID string, routes []*routev3.RouteConfiguration) error {
    // Get current snapshot or create new one
    snap, _ := sm.cache.GetSnapshot(nodeID)
    if snap == nil {
        snap = cache.NewSnapshot("1", nil, nil, nil, nil, nil)
    }

    // Marshal routes to anypb
    routeResources := make([]*anypb.Any, len(routes))
    for i, r := range routes {
        any, err := anypb.New(r)
        if err != nil {
            return err
        }
        routeResources[i] = any
    }

    // Create new snapshot with updated routes
    snap, err := cache.NewSnapshot("2",
        nil, // endpoints
        nil, // clusters
        routeResources,
        nil, // listeners
        nil, // secrets
    )
    if err != nil {
        return err
    }

    // Set snapshot for this node
    if err := sm.cache.SetSnapshot(ctx, nodeID, snap); err != nil {
        return err
    }

    sm.logger.InfoContext(ctx, "routes updated",
        slog.String("node_id", nodeID),
        slog.Int("route_count", len(routes)),
    )
    return nil
}

// UpdateListeners pushes new listener configuration to Envoy
func (sm *SnapshotManager) UpdateListeners(ctx context.Context, nodeID string, listeners []*listenerv3.Listener) error {
    snap, _ := sm.cache.GetSnapshot(nodeID)
    if snap == nil {
        snap = cache.NewSnapshot("1", nil, nil, nil, nil, nil)
    }

    listenerResources := make([]*anypb.Any, len(listeners))
    for i, l := range listeners {
        any, err := anypb.New(l)
        if err != nil {
            return err
        }
        listenerResources[i] = any
    }

    snap, err := cache.NewSnapshot("3",
        nil, // endpoints
        nil, // clusters
        snap.Resources[resource.RouteType],
        listenerResources,
        nil, // secrets
    )
    if err != nil {
        return err
    }

    return sm.cache.SetSnapshot(ctx, nodeID, snap)
}

// UpdateClusters pushes new cluster configuration to Envoy
func (sm *SnapshotManager) UpdateClusters(ctx context.Context, nodeID string, clusters []*clusterv3.Cluster) error {
    // Similar pattern...
    return nil
}

---

## 🔧 Envoy Bootstrap Config

```yaml
# envoy/envoy.yaml
admin:
  address:
    socket_address: { address: 0.0.0.0, port_value: 9901 }

node:
  id: omniroute-node-1
  cluster: omniroute
  metadata:
    role: "ai-gateway"

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

static_resources:
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
  - name: rate_limit_controller
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    typed_extension_protocol_options:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
        explicit_http_config:
          http2_protocol_options: {}
    load_assignment:
      cluster_name: rate_limit_controller
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address: { address: 127.0.0.1, port_value: 9003 }
```

---

# 📅 7-MONTH DAY-BY-DAY TASK PLAN

> **Total: 7 months = ~210 working days**
>
> Legend: 🛠️ Build · 🧪 Test · 📝 Document · 🔄 Review · 🚀 Deploy

---

## MONTH 1: Foundation — Go Project Setup + Envoy Control Plane + Auth

### Week 1: Project Scaffolding & Go Foundation (Days 1-7)

| Day | Tasks |
|-----|-------|
| **Day 1** 🛠️ | `git init` omniroute-go monorepo, `go mod init github.com/omniroute/omniroute` |
| | Write `Makefile` with targets: `build`, `test`, `lint`, `generate`, `dev` |
| | Set up directory structure: `cmd/`, `internal/`, `config/`, `proto/`, `envoy/` |
| | Configure `.golangci.yml` with linters (golangci-lint) |
| | Initialize `slog` logger in `internal/pkg/logger/slog.go` |
| **Day 2** 🛠️ | Install tools: `protoc`, `buf`, `golangci-lint`, `mockgen` |
| | Write `config/config.go` using `koanf` (supports YAML, ENV, flags) |
| | Define config struct: Server, Auth, DB, Redis, Envoy, Providers |
| | Write `config.yaml` with default values |
| **Day 3** 🛠️ | Set up `internal/pkg/errors/errors.go` — error types, codes, wrapping |
| | Set up `internal/pkg/middleware/` — CORS, logging, recovery, request_id |
| | Write `internal/pkg/httputil/response.go` — JSON response helpers |
| | Write `internal/pkg/validator/validator.go` — input validation |
| **Day 4** 🛠️ | Write `cmd/omniroute/main.go` — wire everything, signal handling |
| | Implement graceful shutdown (SIGTERM, SIGINT) |
| | Write `cmd/envoy-bootstrap/main.go` — generate envoy.yaml |
| | Test: `go run cmd/omniroute/main.go --help` works |
| **Day 5** 🧪 | Write unit tests for: logger, config, middleware, errors |
| | Write integration test: server starts and responds to health check |
| | Set up CI in GitHub Actions: `make test`, `make lint`, `make build` |
| **Day 6** 🛠️ | Write `internal/pkg/testutil/` — fixtures, mocks, helpers |
| | Write mock SQLite DB for tests |
| | Write mock HTTP server for upstream provider tests |
| **Day 7** 📝 | Document project structure in README.md |
| | Document development setup: `make dev`, `make test`, etc. |
| | ✅ **Milestone: Go project runs, serves /health, passes tests** |

### Week 2: Envoy Control Plane — xDS Server (Days 8-14)

| Day | Tasks |
|-----|-------|
| **Day 8** 🛠️ | Study `go-control-plane` library architecture |
| | Write `internal/envoy/types.go` — constants, types |
| | Write `internal/envoy/xds_server.go` — gRPC server setup |
| **Day 9** 🛠️ | Implement `SnapshotCache` creation and management |
| | Implement `SnapshotManager` — snapshot builder |
| | Write `internal/envoy/snapshot_cache.go` |
| **Day 10** 🛠️ | Implement `internal/envoy/listeners.go` — build Envoy listeners |
| | Implement `internal/envoy/clusters.go` — build upstream clusters |
| | Implement `internal/envoy/routes.go` — build route configs |
| **Day 11** 🛠️ | Write `envoy/envoy.yaml` bootstrap config |
| | Start Envoy container: `docker run envoyproxy/envoy:v3-latest` |
| | Connect xDS server → Envoy, verify streaming updates |
| **Day 12** 🛠️ | Implement delta gRPC (vs SotW) support |
| | Write `internal/envoy/resources.go` — resource generator |
| | Add xDS callback logging (slog) |
| **Day 13** 🧪 | Integration test: xDS server pushes route → Envoy updates dynamically |
| | Test: Envoy proxies request → Go backend → upstream |
| | Test: multiple Envoy nodes with same xDS server |
| **Day 14** 📝 | Document: envoy control plane setup, xDS flow |
| | ✅ **Milestone: Envoy dynamically configured via xDS** |

### Week 3: SQLite Persistence Layer (Days 15-21)

| Day | Tasks |
|-----|-------|
| **Day 15** 🛠️ | Write `internal/persistence/sqlite/db.go` — DB singleton with WAL |
| | Implement connection pooling, busy timeout, WAL mode |
| | Write migration system: `internal/persistence/sqlite/migrations.go` |
| **Day 16** 🛠️ | Create migration 001: `CREATE TABLE providers` |
| | Create migration 002: `CREATE TABLE combos` |
| | Create migration 003: `CREATE TABLE api_keys` |
| **Day 17** 🛠️ | Create migration 004: `CREATE TABLE settings` |
| | Create migration 005: `CREATE TABLE usage_history` |
| | Create migration 006: `CREATE TABLE mcp_audit` |
| **Day 18** 🛠️ | Write `internal/persistence/sqlite/repositories/provider_repo.go` |
| | Write `internal/persistence/sqlite/repositories/combo_repo.go` |
| | Write `internal/persistence/sqlite/repositories/api_key_repo.go` |
| **Day 19** 🛠️ | Write `internal/persistence/sqlite/repositories/usage_repo.go` |
| | Write `internal/persistence/sqlite/repositories/settings_repo.go` |
| | Write `internal/persistence/sqlite/repositories/mcp_audit_repo.go` |
| **Day 20** 🧪 | Unit tests for all repositories (CRUD operations) |
| | Integration test: run all migrations, verify schema |
| | Benchmark: 10K inserts/sec target |
| **Day 21** 📝 | Document: DB schema, migrations, repository pattern |
| | ✅ **Milestone: Full persistence layer with 6 tables + migrations** |

### Week 4: Auth Controller + ext_authz (Days 22-28)

| Day | Tasks |
|-----|-------|
| **Day 22** 🛠️ | Write `internal/controllers/auth/controller.go` — base controller |
| | Implement `internal/controllers/base.go` — generic watcher/reconciler |
| | Write `internal/controllers/auth/reconciler.go` |
| **Day 23** 🛠️ | Implement API key generation (SHA256 hashing) |
| | Implement API key CRUD in `store.go` |
| | Write `internal/controllers/auth/store.go` |
| **Day 24** 🛠️ | Implement Envoy ext_authz gRPC server |
| | Write `internal/controllers/auth/envoy_extauth.go` |
| | Wire ext_authz into Envoy filter chain |
| **Day 25** 🛠️ | Implement dashboard session auth (cookie-based JWT) |
| | Write dashboard login/logout handlers |
| | Write session validation middleware |
| **Day 26** 🛠️ | Start OAuth provider framework |
| | Write `internal/controllers/auth/oauth/handler.go` — interface |
| | Write `internal/controllers/auth/oauth/registry.go` — provider registry |
| **Day 27** 🧪 | Write unit tests: auth controller, API key validation, ext_authz |
| | Integration test: Envoy ext_authz → Go auth controller → SQLite |
| | Test: API key auth end-to-end through Envoy |
| **Day 28** 📝 | Document: auth controller, OAuth setup, API key management |
| | ✅ **Milestone: Auth controller with Envoy ext_authz, API keys, sessions** |

## MONTH 2: Core Routing — Provider + Routing + Executor Controllers

### Week 5: Provider Controller — Registry + Health (Days 29-35)

| Day | Tasks |
|-----|-------|
| **Day 29** 🛠️ | Implement `internal/controllers/provider/controller.go` |
| | Write provider registry: `internal/controllers/provider/registry.go` |
| | Define provider struct: ID, Name, Type, BaseURL, AuthType, Models |
| **Day 30** 🛠️ | Implement provider CRUD in store |
| | Write `internal/controllers/provider/store.go` |
| | Write `internal/persistence/sqlite/repositories/provider_repo.go` |
| **Day 31** 🛠️ | Implement health checker: periodic probes + on-demand |
| | Write `internal/controllers/provider/health_checker.go` |
| | Implement circuit breaker-aware health status |
| **Day 32** 🛠️ | Register first 50 providers from TypeScript constants |
| | Categories: OpenAI, Anthropic, Gemini, DeepSeek, Groq, xAI, Mistral |
| | Write provider definitions in `config/providers.yaml` |
| **Day 33** 🛠️ | Register next 50 providers (Together, Fireworks, Cerebras, Cohere, NVIDIA...) |
| | Register OAuth providers (Claude Code, Codex, Gemini, GitHub, Kiro...) |
| **Day 34** 🛠️ | Register remaining 137 providers (all categories) |
| | Register self-hosted providers (LM Studio, vLLM, Ollama, Triton...) |
| | Validate all 237 providers load without errors |
| **Day 35** 🧪 | Unit tests: provider registry, CRUD, health checker |
| | Integration test: health checker probes real endpoints |
| | ✅ **Milestone: Provider controller with 237 providers + health monitoring** |

### Week 6: Executor Controller — Base + Default (Days 36-42)

| Day | Tasks |
|-----|-------|
| **Day 36** 🛠️ | Write `internal/executor/base_executor.go` — BaseExecutor interface |
| | Define: `Execute(ctx, req) → (*Response, error)` |
| | Implement `buildUrl()`, `buildHeaders()`, `transformRequest()` |
| **Day 37** 🛠️ | Implement `DefaultExecutor` (OpenAI-compatible) |
| | Write `internal/executor/default.go` |
| | Implement streaming (SSE chunk parsing) |
| **Day 38** 🛠️ | Implement retry logic with exponential backoff |
| | Write retry handler: max 3 attempts, jitter, backoff |
| | Implement token refresh on 401 responses |
| **Day 39** 🛠️ | Implement `AnthropicExecutor` |
| | Write `internal/executor/anthropic.go` |
| | Handle Anthropic-specific headers and response format |
| **Day 40** 🛠️ | Implement `GeminiExecutor` |
| | Write `internal/executor/gemini.go` |
| | Handle Gemini API format and auth |
| **Day 41** 🛠️ | Implement `ClaudeWebExecutor` (web/OAuth-based) |
| | Implement `CodexExecutor`, `CursorExecutor` |
| | Write `internal/executor/claude_web.go`, `codex.go`, `cursor.go` |
| **Day 42** 🧪 | Unit tests: all executors with mock HTTP server |
| | Integration test: executor → real upstream (test account) |
| | ✅ **Milestone: 6 executors working + retry + streaming** |

### Week 7: Executor Controller — OAuth + Free + Special (Days 43-49)

| Day | Tasks |
|-----|-------|
| **Day 43** 🛠️ | Implement `KiroExecutor`, `QoderExecutor` |
| | Write `internal/executor/kiro.go`, `qoder.go` |
| | Handle free-tier specific auth and rate limits |
| **Day 44** 🛠️ | Implement `PollinationsExecutor`, `PuterExecutor` |
| | Write `internal/executor/pollinations.go`, `puter.go` |
| | Handle web-based providers |
| **Day 45** 🛠️ | Implement `DeepSeekExecutor`, `GroqExecutor`, `xAIExecutor` |
| | Write `internal/executor/deepseek.go`, `groq.go`, `xai.go` |
| **Day 46** 🛠️ | Implement `MistralExecutor`, `TogetherExecutor`, `FireworksExecutor` |
| | Write `internal/executor/mistral.go`, `together.go`, `fireworks.go` |
| **Day 47** 🛠️ | Implement `CerebrasExecutor`, `CohereExecutor`, `NVIDIAExecutor` |
| | Write `internal/executor/cerebras.go`, `cohere.go`, `nvidia.go` |
| **Day 48** 🛠️ | Implement `HuggingFaceExecutor`, `OpenRouterExecutor` |
| | Implement `VertexExecutor`, `CloudflareExecutor` |
| **Day 49** 🧪 | Integration tests: 20 executors against real providers |
| | Test: streaming works for each executor |
| | ✅ **Milestone: 20 executors implemented and tested** |

### Week 8: Remaining Executors (Days 50-56)

| Day | Tasks |
|-----|-------|
| **Day 50** 🛠️ | Implement executors 21-30: Scaleway, Longcat, Alibaba, Kimi, Minimax, Blackbox, Synthetic, KiloGateway, GLM, Deepgram |
| **Day 51** 🛠️ | Implement executors 31-40: AssemblyAI, ElevenLabs, Cartesia, PlayHT, Inworld, NanoBanana, SDWebUI, ComfyUI, OllamaCloud, DeepInfra |
| **Day 52** 🛠️ | Implement executors 41-50: VercelAIGW, Lambda, SambaNova, NScale, OVHCloud, Baseten, PublicAI, Moonshot, MetaLlama, v0 |
| **Day 53** 🛠️ | Implement executors 51-60: Morph, Featherless, FriendliAI, LlamaGate, Galadriel, WeightsBiases, Volcengine, AI21, Venice, Codestral |
| **Day 54** 🛠️ | Implement executors 61-68: Upstage, Maritalk, Xiaomi, InferenceNet, NanoGPT, Predibase, Bytez, HerokuAI, Databricks, Snowflake, GigaChat, CrofAI, AgentRouter, ChatGPTWeb, Baidu, AWSPolly, RunwayML, GitLabDuo, AmazonQ, Empower, Poe |
| | Implement remaining executor files |
| **Day 55** 🧪 | Integration tests: all 68 executors with HTTP mocks |
| | Test: each executor builds correct URL, headers, body |
| | Test: each executor parses response correctly |
| **Day 56** 📝 | Document: executor architecture, adding new providers |
| | ✅ **Milestone: All 68 executors implemented and tested** |

## MONTH 3: Routing + Translation + Quota Controllers

### Week 9: Routing Controller — Strategies Part 1 (Days 57-63)

| Day | Tasks |
|-----|-------|
| **Day 57** 🛠️ | Write `internal/controllers/routing/controller.go` |
| | Write `internal/controllers/routing/strategies/strategy.go` — interface |
| | Implement `Priority` strategy |
| **Day 58** 🛠️ | Implement `Weighted` strategy |
| | Implement `RoundRobin` strategy |
| | Implement `FillFirst` strategy |
| **Day 59** 🛠️ | Implement `P2C` (Power of Two Choices) strategy |
| | Implement `Random` strategy |
| | Implement `LeastUsed` strategy |
| **Day 60** 🛠️ | Write combo resolver: `internal/controllers/routing/combo_resolver.go` |
| | Implement combo target resolution from DB |
| | Implement Envoy xDS route updates on combo change |
| **Day 61** 🛠️ | Write routing controller reconciler |
| | Wire strategy selector into request pipeline |
| | Implement `handleChat()` in routing controller |
| **Day 62** 🧪 | Unit tests: all 7 strategies (correct selection) |
| | Integration test: combo routing → executor → upstream |
| | Test: Envoy route reflects combo changes |
| **Day 63** 📝 | Document: routing strategies, combo system |
| | ✅ **Milestone: 7 strategies + combo resolver + Envoy integration** |

### Week 10: Routing Controller — Strategies Part 2 (Days 64-70)

| Day | Tasks |
|-----|-------|
| **Day 64** 🛠️ | Implement `ResetAware` strategy |
| | Implement `ResetWindow` strategy |
| | Implement `CostOptimized` strategy |
| **Day 65** 🛠️ | Implement `StrictRandom` strategy |
| | Implement `Auto` strategy (basic version) |
| | Implement `LKGP` (Last Known Good Provider) strategy |
| **Day 66** 🛠️ | Implement `ContextOptimized` strategy |
| | Implement `ContextRelay` strategy |
| | Implement `Fusion` strategy (panel + judge) |
| **Day 67** 🛠️ | Implement `TaskAwareRouter` — FSM-based routing |
| | Write `internal/controllers/routing/task_aware.go` |
| | Implement workflow states for routing decisions |
| **Day 68** 🛠️ | Implement `AutoCombo` engine (9-factor scoring) |
| | Write `internal/controllers/routing/auto_combo.go` |
| | Factors: cost, latency, reliability, quota, context, task, etc. |
| **Day 69** 🧪 | Unit tests: remaining 10 strategies (comprehensive) |
| | Integration test: AutoCombo scoring engine |
| | Benchmark: combo resolution < 5ms |
| **Day 70** 📝 | Document: all 17 strategies, AutoCombo engine |
| | ✅ **Milestone: All 17 routing strategies implemented** |

### Week 11: Translator Controller (Days 71-77)

| Day | Tasks |
|-----|-------|
| **Day 71** 🛠️ | Write `internal/controllers/translator/controller.go` |
| | Define translation interfaces: `RequestTranslator`, `ResponseTranslator` |
| | Write `internal/controllers/translator/request.go` — entry point |
| **Day 72** 🛠️ | Implement `openai_to_claude.go` — request translation |
| | Map: messages, roles, tools, thinking, streaming |
| | Handle: system → developer role mapping |
| **Day 73** 🛠️ | Implement `openai_to_gemini.go` — request translation |
| | Map: contents, safety settings, generation config |
| | Handle: vision/image content parts |
| **Day 74** 🛠️ | Implement `claude_to_openai.go` — response translation |
| | Map: content blocks, tool use, thinking, stop reason |
| | Implement `gemini_to_openai.go` — response translation |
| **Day 75** 🛠️ | Implement `role_normalizer.go` — role mapping across formats |
| | Implement `think_tag_parser.go` — parse <think> tags |
| | Implement `structured_output.go` — json_schema conversion |
| **Day 76** 🛠️ | Implement `image_handler.go` — base64, URL, multipart |
| | Implement `audio_handler.go`, `video_handler.go` |
| | Implement content type detection and conversion |
| **Day 77** 🧪 | Unit tests: all translation paths (8 directions) |
| | Integration test: OpenAI → Claude → OpenAI round trip |
| | ✅ **Milestone: Full translator controller with round-trip validation** |

### Week 12: Quota Controller (Days 78-84)

| Day | Tasks |
|-----|-------|
| **Day 78** 🛠️ | Write `internal/controllers/quota/controller.go` |
| | Implement `internal/controllers/quota/rate_limiter.go` — token bucket |
| | Implement sliding window rate limiter per provider |
| **Day 79** 🛠️ | Implement Envoy global rate limit gRPC service |
| | Write `internal/controllers/quota/envoy_rate_limit.go` |
| | Wire Envoy rate limit filter → Go quota controller |
| **Day 80** 🛠️ | Implement `internal/controllers/quota/budget_tracker.go` |
| | Monthly budget tracking per connection |
| | Auto-disable when budget exhausted |
| **Day 81** 🛠️ | Implement `internal/controllers/quota/token_counter.go` |
| | Hybrid: provider-reported tokens + estimation fallback |
| | Implement token counting for all model families |
| **Day 82** 🛠️ | Implement usage persistence |
| | Write usage_history repository |
| | Implement cost calculation per model |
| **Day 83** 🧪 | Unit tests: rate limiter, budget tracker, token counter |
| | Integration test: Envoy rate limit → Go → Redis → SQLite |
| | Benchmark: rate limiter 100K req/s |
| **Day 84** 📝 | Document: quota controller, rate limiting, budget tracking |
| | ✅ **Milestone: Quota controller with Envoy rate limit integration** |

## MONTH 4: Resilience + Compression + Redis Cache

### Week 13: Resilience Controller (Days 85-91)

| Day | Tasks |
|-----|-------|
| **Day 85** 🛠️ | Write `internal/controllers/resilience/controller.go` |
| | Implement `internal/controllers/resilience/circuit_breaker.go` |
| | Three states: CLOSED → OPEN → HALF_OPEN |
| **Day 86** 🛠️ | Implement failure counting + threshold |
| | Implement recovery timer (OPEN → HALF_OPEN) |
| | Implement half-open probe logic (success/failure decision) |
| **Day 87** 🛠️ | Implement `internal/controllers/resilience/cooldown.go` |
| | Per-account/key cooldown after rate limit hit |
| | Configurable cooldown duration |
| **Day 88** 🛠️ | Implement `internal/controllers/resilience/lockout.go` |
| | Per-provider+model lockout after quota exhaustion |
| | Automatic release after reset window |
| **Day 89** 🛠️ | Implement `internal/controllers/resilience/herd_protection.go` |
| | Mutex-based thundering herd prevention |
| | Request coalescing for identical requests |
| **Day 90** 🛠️ | Implement `internal/controllers/resilience/domain_state.go` |
| | Persist breaker/cooldown/lockout state to SQLite |
| | Redis-cached fast path for state lookups |
| **Day 91** 🧪 | Unit tests: circuit breaker state machine, cooldown, lockout |
| | Integration test: breaker opens → requests blocked → recovers |
| | ✅ **Milestone: 3-layer resilience (breaker + cooldown + lockout)** |

### Week 14: Compression Controller — Caveman Engine (Days 92-98)

| Day | Tasks |
|-----|-------|
| **Day 92** 🛠️ | Write `internal/controllers/compression/controller.go` |
| | Implement compression strategy selector |
| | Write `internal/controllers/compression/caveman.go` — main engine |
| **Day 93** 🛠️ | Implement `collapseWhitespace` technique |
| | Implement `dedupSystemPrompt` technique |
| | Implement `compressToolResults` technique |
| **Day 94** 🛠️ | Implement `removeRedundantContent` técnica |
| | Implement `replaceImageUrls` technique |
| | Implement lite mode: 5 techniques, <1ms latency, 10-15% savings |
| **Day 95** 🛠️ | Implement caveman rules engine |
| | Write `internal/controllers/compression/caveman_rules.go` |
| | Implement semantic condensation rules |
| **Day 96** 🛠️ | Implement language packs |
| | Write `internal/controllers/compression/language_packs.go` |
| | Load per-language compression rules from files |
| **Day 97** 🛠️ | Implement compression stats tracking |
| | Write `internal/controllers/compression/stats.go` |
| | Track: original tokens, compressed tokens, savings %, engine used |
| **Day 98** 🧪 | Unit tests: caveman techniques, rules engine, language packs |
| | Integration test: compression pipeline on real prompts |
| | ✅ **Milestone: Caveman compression engine with rules + language packs** |

### Week 15: Compression Controller — RTK Engine (Days 99-105)

| Day | Tasks |
|-----|-------|
| **Day 99** 🛠️ | Write `internal/controllers/compression/rtk.go` — RTK engine |
| | Implement command output class detection |
| | Implement terminal output pattern recognition |
| **Day 100** 🛠️ | Implement RTK filter packs (JSON DSL) |
| | Write `internal/controllers/compression/rtk_filters.go` |
| | Filters: replace, match-output, strip/keep, truncation |
| **Day 101** 🛠️ | Implement per-line truncation (head/tail) |
| | Implement max-line truncation |
| | Implement ANSI/code noise stripping |
| **Day 102** 🛠️ | Implement deduplication of repeated lines |
| | Implement error-preserving compression |
| | Implement actionable context retention |
| **Day 103** 🛠️ | Implement stacked pipeline |
| | Write `internal/controllers/compression/pipeline.go` |
| | Multiple engines in sequence: Caveman → RTK |
| **Day 104** 🛠️ | Implement compression combos (DB-driven assignments) |
| | Implement compression combo stats tracking |
| | Write compression combo repository |
| **Day 105** 🧪 | Unit tests: RTK filters, stacked pipeline, compression combos |
| | Integration test: end-to-end compression on tool-heavy session |
| | Benchmark: compression latency target <50ms |
| | ✅ **Milestone: Full compression pipeline (Caveman + RTK + stacked)** |

### Week 16: Redis Cache Layer (Days 106-112)

| Day | Tasks |
|-----|-------|
| **Day 106** 🛠️ | Write `internal/cache/redis.go` — Redis client setup |
| | Implement connection pool, retry, health check |
| | Write config for Redis (host, port, password, DB) |
| **Day 107** 🛠️ | Implement rate limiter in Redis |
| | Write `internal/cache/rate_limiter.go` |
| | Use Redis INCR + EXPIRE for sliding window |
| **Day 108** 🛠️ | Implement session store in Redis |
| | Write `internal/cache/session.go` |
| | Cookie-to-session mapping with TTL |
| **Day 109** 🛠️ | Implement circuit breaker state in Redis |
| | Fast path: Redis read → SQLite write on change |
| | Implement pub/sub for cross-instance state sync |
| **Day 110** 🛠️ | Implement provider health snapshot cache |
| | Implement combo resolution cache |
| | Cache invalidation on provider/combo change events |
| **Day 111** 🛠️ | Implement SSE streaming engine |
| | Write `internal/streaming/sse.go` — SSE writer |
| | Write `internal/streaming/sse_parser.go` — SSE reader |
| **Day 112** 🧪 | Unit tests: Redis cache, rate limiter, session store |
| | Integration test: Redis + SQLite dual-write consistency |
| | ✅ **Milestone: Redis cache layer + SSE streaming engine** |

## MONTH 5: Protocol Servers — MCP + A2A

### Week 17: MCP Controller — Core + Transports (Days 113-119)

| Day | Tasks |
|-----|-------|
| **Day 113** 🛠️ | Write `internal/controllers/mcp/controller.go` |
| | Implement JSON-RPC 2.0 message parser |
| | Write `internal/controllers/mcp/server.go` — transport dispatcher |
| **Day 114** 🛠️ | Implement stdio transport |
| | Write `internal/controllers/mcp/stdio.go` |
| | Read/write JSON-RPC over stdin/stdout |
| **Day 115** 🛠️ | Implement SSE transport |
| | Write `internal/controllers/mcp/sse.go` |
| | SSE endpoint: `GET /api/mcp/sse` |
| **Day 116** 🛠️ | Implement Streamable HTTP transport |
| | Write `internal/controllers/mcp/streamable_http.go` |
| | POST endpoint: `/api/mcp/stream` |
| **Day 117** 🛠️ | Implement tool registry |
| | Write `internal/controllers/mcp/tools.go` |
| | Tool definition: name, description, input schema, handler |
| **Day 118** 🛠️ | Implement scope authorization (30 scopes) |
| | Write `internal/controllers/mcp/scopes.go` |
| | Scope enforcement before handler dispatch |
| **Day 119** 🛠️ | Implement MCP audit logging |
| | Write `internal/controllers/mcp/audit.go` |
| | Log: tool name, args, success/failure, key, timestamp → SQLite |
| | ✅ **Milestone: MCP server with 3 transports + auth + audit** |

### Week 18: MCP Tools — Core (Days 120-126)

| Day | Tasks |
|-----|-------|
| **Day 120** 🛠️ | Implement 5 core tools: get_health, list_combos, get_combo_metrics, switch_combo, check_quota |
| **Day 121** 🛠️ | Implement 5 core tools: route_request, cost_report, list_models_catalog, web_search, simulate_route |
| **Day 122** 🛠️ | Implement 5 core tools: set_budget_guard, set_routing_strategy, set_resilience_profile, test_combo, get_provider_metrics |
| **Day 123** 🛠️ | Implement 5 core tools: best_combo_for_task, explain_route, get_session_snapshot, db_health_check, sync_pricing |
| **Day 124** 🛠️ | Implement cache tools: cache_stats, cache_flush |
| | Implement compression tools (5): compression_status, compression_configure, set_compression_engine, list_compression_combos, compression_combo_stats |
| **Day 125** 🛠️ | Implement 1proxy tools (3): oneproxy_fetch, oneproxy_rotate, oneproxy_stats |
| | Implement memory tools (3): memory_search, memory_add, memory_clear |
| **Day 126** 🧪 | Unit tests: all 20 core + 2 cache + 5 compression + 3 1proxy + 3 memory = 33 tools |
| | ✅ **Milestone: 33 MCP tools implemented and tested** |

### Week 19: MCP Tools — Skills + Notion + Obsidian (Days 127-133)

| Day | Tasks |
|-----|-------|
| **Day 127** 🛠️ | Implement skills tools (4): skills_list, skills_enable, skills_execute, skills_executions |
| | Implement A2A skill bridge tools (3) |
| **Day 128** 🛠️ | Implement Notion tools (6): notion_list, notion_read, notion_write, notion_search, notion_delete, notion_query |
| | Wire Notion API client |
| **Day 129** 🛠️ | Implement Obsidian tools Part 1 (10): obsidian_search, obsidian_read, obsidian_write, obsidian_delete, obsidian_list, obsidian_tags, obsidian_graph, obsidian_backlinks, obsidian_templates, obsidian_properties |
| **Day 130** 🛠️ | Implement Obsidian tools Part 2 (12): obsidian_daily_note, obsidian_weekly_note, obsidian_monthly_note, obsidian_canvas, obsidian_webdav_list, obsidian_webdav_read, obsidian_webdav_write, obsidian_webdav_delete, obsidian_webdav_mkdir, obsidian_webdav_copy, obsidian_webdav_move |
| **Day 131** 🛠️ | Implement gamification tools (8): levels, badges, leaderboard, community federation queries |
| | Implement plugin tools (8): marketplace, list, install, enable, disable, inspect, config, uninstall |
| **Day 132** 🧪 | Unit tests: skills (4) + notion (6) + obsidian (22) + gamification (4) + plugin (8) = 44 tools |
| | Integration test: end-to-end MCP tool call through all 3 transports |
| **Day 133** 📝 | Document: MCP server (94 tools), MCP scope system, tool development guide |
| | ✅ **Milestone: All 94 MCP tools implemented and tested** |

### Week 20: A2A Controller (Days 134-140)

| Day | Tasks |
|-----|-------|
| **Day 134** 🛠️ | Write `internal/controllers/a2a/controller.go` |
| | Implement JSON-RPC 2.0 handler |
| | Write `internal/controllers/a2a/server.go` |
| **Day 135** 🛠️ | Implement Task Manager |
| | Write `internal/controllers/a2a/task_manager.go` |
| | Task lifecycle: submitted → working → completed/failed/canceled |
| **Day 136** 🛠️ | Implement message/send (sync) method |
| | Implement message/stream (SSE) method |
| | Implement tasks/get and tasks/cancel methods |
| **Day 137** 🛠️ | Implement Agent Card |
| | Write `internal/controllers/a2a/agent_card.go` |
| | Serve at `/.well-known/agent.json` |
| **Day 138** 🛠️ | Implement A2A skills (3): smart_routing, quota_management, provider_discovery |
| | Write `internal/controllers/a2a/skills/smart_routing.go`, `quota_management.go`, `provider_discovery.go` |
| **Day 139** 🛠️ | Implement A2A skills (3): cost_analysis, health_report, list_capabilities |
| | Write remaining skill files |
| **Day 140** 🧪 | Unit tests: all JSON-RPC methods, task lifecycle, skills |
| | Integration test: A2A client → server → skill → response |
| | ✅ **Milestone: A2A server with 6 skills and Agent Card** |

## MONTH 6: Memory + Skills + Guardrails + Infrastructure

### Week 21: Memory Controller (Days 141-147)

| Day | Tasks |
|-----|-------|
| **Day 141** 🛠️ | Write `internal/controllers/memory/controller.go` |
| | Write `internal/controllers/memory/extractor.go` — memory extraction from conversations |
| **Day 142** 🛠️ | Write `internal/controllers/memory/store.go` — FTS5 full-text search |
| | Implement SQLite FTS5 virtual tables |
| | Write memory CRUD operations |
| **Day 143** 🛠️ | Implement vector-based similarity search |
| | Use simple vector embeddings (or integrate with pgvector / SQLite vector extension) |
| | Implement cosine similarity scoring |
| **Day 144** 🛠️ | Write `internal/controllers/memory/injector.go` — memory injection into prompts |
| | Implement context window-aware injection |
| | Implement memory prioritization (recency, relevance) |
| **Day 145** 🛠️ | Write `internal/controllers/memory/retriever.go` — memory retrieval |
| | Implement hybrid search: FTS5 + vector |
| | Implement memory filtering (by session, user, time range) |
| **Day 146** 🛠️ | Write `internal/controllers/memory/summarizer.go` — memory summarization |
| | Implement conversation summarization for long histories |
| | Implement hierarchical memory (session → day → week → month) |
| **Day 147** 🧪 | Unit tests: memory extraction, injection, retrieval, summarization |
| | Integration test: multi-turn conversation with memory persistence |
| | ✅ **Milestone: Memory system with FTS5 + vector + summarization** |

### Week 22: Skills Controller (Days 148-154)

| Day | Tasks |
|-----|-------|
| **Day 148** 🛠️ | Write `internal/controllers/skills/controller.go` |
| | Write `internal/controllers/skills/registry.go` — DB-backed skill registry |
| **Day 149** 🛠️ | Write `internal/controllers/skills/executor.go` — skill execution engine |
| | Implement configurable timeout and retry |
| | Implement input/output validation |
| **Day 150** 🛠️ | Write `internal/controllers/skills/sandbox.go` — isolation layer |
| | Implement resource limits for custom skills |
| | Implement execution time limits |
| **Day 151** 🛠️ | Implement built-in skills (3): quota, routing, health |
| | Write `internal/controllers/skills/builtin/quota.go`, `routing.go`, `health.go` |
| **Day 152** 🛠️ | Implement custom skill support |
| | Write `internal/controllers/skills/custom/handler.go` |
| | Implement skill installation from plugin registry |
| **Day 153** 🛠️ | Implement skill interception and injection |
| | Skills can intercept requests (pre/post processing) |
| | Skills can inject context into prompts |
| **Day 154** 🧪 | Unit tests: skill registry, executor, sandbox, built-in skills |
| | Integration test: custom skill execution in sandbox |
| | ✅ **Milestone: Skills system with registry + executor + sandbox** |

### Week 23: Guardrails Controller (Days 155-161)

| Day | Tasks |
|-----|-------|
| **Day 155** 🛠️ | Write `internal/controllers/guardrails/controller.go` |
| | Implement hot-reloadable guardrail framework |
| | Write `internal/controllers/guardrails/hot_reload.go` |
| **Day 156** 🛠️ | Implement PII masker guardrail |
| | Write `internal/controllers/guardrails/pii_masker.go` |
| | Regex + ML-based PII detection (emails, SSN, credit cards, etc.) |
| **Day 157** 🛠️ | Implement PII masking options: redact, hash, replace |
| | Implement F/B: opt-in via PII_REDACTION_ENABLED |
| | Implement response sanitization for streaming |
| **Day 158** 🛠️ | Implement prompt injection detection guardrail |
| | Write `internal/controllers/guardrails/prompt_injection.go` |
| | Pattern-based + heuristic detection |
| **Day 159** 🛠️ | Implement vision bridge guardrail |
| | Write `internal/controllers/guardrails/vision_bridge.go` |
| | Content safety check for vision inputs |
| **Day 160** 🛠️ | Implement guardrail opt-out via header (`x-omniroute-disabled-guardrails`) |
| | Implement guardrail chain: run all enabled guardrails in order |
| | Implement guardrail metrics (pass/fail counts, latency) |
| **Day 161** 🧪 | Unit tests: PII masker (100+ patterns), injection detection, vision bridge |
| | Integration test: guardrail chain on real prompts |
| | ✅ **Milestone: 3 guardrails (PII + injection + vision) + hot-reload** |

### Week 24: Webhook + Eval + Tunnel + MITM Controllers (Days 162-168)

| Day | Tasks |
|-----|-------|
| **Day 162** 🛠️ | Write `internal/controllers/webhook/controller.go` |
| | Implement `internal/controllers/webhook/dispatcher.go` — HMAC-signed delivery |
| | Implement event types (7 events) |
| **Day 163** 🛠️ | Implement webhook retry with exponential backoff |
| | Implement auto-disable after 10 failures |
| | Write `internal/controllers/webhook/store.go` |
| **Day 164** 🛠️ | Write `internal/controllers/eval/controller.go` |
| | Implement `internal/controllers/eval/runner.go` — generic eval runner |
| | Implement targets: combo / model / suite-default |
| **Day 165** 🛠️ | Implement `internal/controllers/eval/runtime.go` — eval execution |
| | Implement `internal/controllers/eval/store.go` — eval persistence |
| | Write eval repository |
| **Day 166** 🛠️ | Write `internal/controllers/tunnel/controller.go` |
| | Implement Cloudflare tunnel (Quick + Named) |
| | Write `internal/controllers/tunnel/cloudflare.go` |
| **Day 167** 🛠️ | Implement ngrok tunnel |
| | Implement Tailscale Funnel |
| | Write `internal/controllers/tunnel/ngrok.go`, `tailscale.go` |
| **Day 168** 🛠️ | Write `internal/controllers/mitm/controller.go` |
| | Implement `internal/controllers/mitm/cert_manager.go` — cert generation + installation |
| | Implement `internal/controllers/mitm/tproxy.go` — TPROXY routing |
| | ✅ **Milestone: Webhooks (7 events) + Evals + Tunnels (3) + MITM** |

## MONTH 7: Kubernetes Operator + Testing + Final Migration

### Week 25: Kubernetes Operator — CRDs + API Types (Days 169-175)

| Day | Tasks |
|-----|-------|
| **Day 169** 🛠️ | Initialize operator with `kubebuilder init` |
| | Set up `operator/` directory structure |
| | Create `operator/api/v1alpha1/groupversion_info.go` |
| **Day 170** 🛠️ | Define Provider CRD |
| | Write `operator/api/v1alpha1/provider_types.go` |
| | Spec: name, providerType, baseURL, auth, models, health |
| **Day 171** 🛠️ | Define Combo CRD |
| | Write `operator/api/v1alpha1/combo_types.go` |
| | Spec: name, strategy, targets, fallback, weights |
| **Day 172** 🛠️ | Define ApiKey CRD |
| | Write `operator/api/v1alpha1/apikey_types.go` |
| | Define RateLimit CRD |
| | Write `operator/api/v1alpha1/ratelimit_types.go` |
| **Day 173** 🛠️ | Define CircuitBreaker CRD |
| | Write `operator/api/v1alpha1/circuitbreaker_types.go` |
| | Define Webhook CRD |
| | Write `operator/api/v1alpha1/webhook_types.go` |
| **Day 174** 🛠️ | Define MCPTool CRD |
| | Write `operator/api/v1alpha1/mcptool_types.go` |
| | Define Tunnel CRD |
| | Write `operator/api/v1alpha1/tunnel_types.go` |
| **Day 175** 🛠️ | Generate deepcopy functions: `make generate` |
| | Generate CRD YAML manifests: `make manifests` |
| | ✅ **Milestone: 8 CRDs defined + generated manifests** |

### Week 26: Kubernetes Operator — Controllers (Days 176-182)

| Day | Tasks |
|-----|-------|
| **Day 176** 🛠️ | Implement Provider controller |
| | Write `operator/controllers/provider_controller.go` |
| | Reconcile: CRD changes → Provider controller → xDS update |
| **Day 177** 🛠️ | Implement Combo controller |
| | Write `operator/controllers/combo_controller.go` |
| | Reconcile: CRD changes → Routing controller → Envoy route update |
| **Day 178** 🛠️ | Implement ApiKey controller |
| | Write `operator/controllers/apikey_controller.go` |
| | Reconcile: CRD changes → Auth controller → ext_authz update |
| **Day 179** 🛠️ | Implement CircuitBreaker controller |
| | Write `operator/controllers/circuitbreaker_controller.go` |
| | Reconcile: CRD changes → Resilience controller |
| **Day 180** 🛠️ | Implement RateLimit controller |
| | Write `operator/controllers/ratelimit_controller.go` |
| | Reconcile: CRD changes → Quota controller → Envoy rate limit |
| **Day 181** 🛠️ | Implement Webhook + MCPTool + Tunnel controllers |
| | Write remaining operator controller files |
| **Day 182** 🛠️ | Write admission webhooks (validation + mutation) |
| | Write `operator/webhooks/provider_webhook.go` |
| | Generate RBAC manifests |
| | ✅ **Milestone: All 8 operator controllers + webhooks** |

### Week 27: Testing — Unit + Integration (Days 183-189)

| Day | Tasks |
|-----|-------|
| **Day 183** 🧪 | Write unit tests for Auth controller (API keys, OAuth, ext_authz) |
| | Target: 90% coverage |
| **Day 184** 🧪 | Write unit tests for Routing controller (all 17 strategies) |
| | Write unit tests for Executor controller (retry, streaming, error handling) |
| **Day 185** 🧪 | Write unit tests for Translator controller (all 8 translation paths) |
| | Write unit tests for Quota controller (rate limiter, budget, token counter) |
| **Day 186** 🧪 | Write unit tests for Resilience controller (circuit breaker FSM) |
| | Write unit tests for Compression controller (caveman, RTK, pipeline) |
| **Day 187** 🧪 | Write unit tests for MCP controller (all 94 tools) |
| | Write unit tests for A2A controller (JSON-RPC, task lifecycle) |
| **Day 188** 🧪 | Write unit tests for Memory, Skills, Guardrails controllers |
| | Write unit tests for Webhook, Eval, Tunnel, MITM controllers |
| **Day 189** 🧪 | Write integration tests for full request pipeline |
| | Integration test: Envoy → Auth → Routing → Translator → Executor → Provider |
| | ✅ **Milestone: 15,000+ unit + integration tests** |

### Week 28: Testing — E2E + Performance + Migration (Days 190-196)

| Day | Tasks |
|-----|-------|
| **Day 190** 🧪 | Write E2E tests with real provider calls (test accounts) |
| | E2E: chat completion, streaming, combo routing |
| **Day 191** 🧪 | Write E2E tests for MCP tools (all 94, real execution) |
| | Write E2E tests for A2A (all 6 skills) |
| **Day 192** 🧪 | Performance benchmarks: P50/P99 latency, throughput, memory |
| | Benchmark: single model, combo routing, compression |
| **Day 193** 🧪 | Load testing with k6: 10K concurrent requests |
| | Identify bottlenecks, optimize hot paths |
| **Day 194** 🔄 | Run TS and Go in parallel (side-by-side validation) |
| | A/B test: same request to both — compare responses |
| **Day 195** 🔄 | Fix any discrepancies between TS and Go responses |
| | Validate: SSE format, error codes, streaming behaviour |
| **Day 196** 🚀 | Cutover: route 100% traffic to Go backend |
| | Keep TS as fallback for 48h monitoring |
| | ✅ **Milestone: Full cutover — Go handles 100% traffic** |

### Week 29: Dashboard Integration (Days 197-203)

| Day | Tasks |
|-----|-------|
| **Day 197** 🛠️ | Wire Next.js dashboard to Go API |
| | Update `api-client.ts` to point to Go backend |
| | Test: all dashboard pages load with Go data |
| **Day 198** 🛠️ | Implement real-time SSE updates in dashboard |
| | Circuit breaker status, usage, health — live via EventSource |
| **Day 199** 🛠️ | Implement WebSocket bridge (OpenAI-compatible WS) |
| | Write `internal/streaming/websocket.go` |
| **Day 200** 🛠️ | Test electron desktop app with Go backend |
| | Fix any compatibility issues |
| **Day 201** 🛠️ | Implementation review: feature parity checklist |
| | Verify: all v1 endpoints, management APIs, MCP, A2A |
| **Day 202** 🧪 | Full regression test: all features end-to-end |
| | Document: migration summary, architecture decisions |
| **Day 203** 🚀 | Final deploy: Go + Envoy production rollout |
| | Enable monitoring, alerting, dashboards |
| | ✅ **Milestone: Dashboard fully connected to Go backend** |

### Week 30: Documentation + Cleanup + Release (Days 204-210)

| Day | Tasks |
|-----|-------|
| **Day 204** 📝 | Write Go architecture documentation |
| | Document controller pattern, xDS flow, slog logging |
| **Day 205** 📝 | Write API reference (OpenAPI/Swagger) |
| | Document all v1 and management endpoints |
| **Day 206** 📝 | Write deployment guide |
| | Docker, Docker Compose, Kubernetes, Helm chart |
| **Day 207** 📝 | Write operator documentation |
| | CRDs, controller reconciliation, RBAC setup |
| **Day 208** 🧪 | Final performance benchmark vs TS baseline |
| | Publish results: latency improvement, throughput gain |
| **Day 209** 🛠️ | Clean up: remove TS backend code |
| | Archive old code, update README |
| **Day 210** 📝 | Write CHANGELOG for v4.0.0 (Go/Envoy) |
| | 🎉 **RELEASE: OmniRoute v4.0 — Fully migrated to Go + Envoy** |

---

## 🎯 Migration Strategy

### Phase 1: Side-by-Side (Months 1-4)
```
Port 20128 → Envoy → Go controllers (new endpoints)
Port 20129 → Next.js (existing, for fallback)
Dashboard connects to Go backend
Both systems run in parallel for validation
```

### Phase 2: Gradual Cutover (Months 5-6)
```
Envoy progressively routes more traffic to Go controllers:
  /health → Go
  /v1/models → Go
  /v1/chat/completions (single model) → Go
  /v1/chat/completions (combo routing) → Go
  /v1/embeddings → Go
  MCP, A2A → Go
```

### Phase 3: Full Migration (Month 7)
```
Port 20128 → Envoy → Go controllers (100% traffic)
Next.js serves only dashboard frontend
TS backend decommissioned
Kubernetes operator enables cloud-native deployment
```

---

## 📊 Key Metrics vs Current (TS)

| Metric | Current (TS) | Target (Go) |
|---|---|---|
| **P50 Latency** | ~50ms | <10ms |
| **P99 Latency** | ~200ms | <50ms |
| **Throughput** | ~1K req/s | >10K req/s |
| **Memory Usage** | ~150MB idle | <30MB idle |
| **Startup Time** | ~5s | <500ms |
| **Binary Size** | N/A (interpreter) | <20MB |
| **Concurrent Connections** | ~500 | >5,000 |
| **Test Count** | ~15,000 | ~15,000+ |

---

## ⚠️ Risk Register

| Risk | Probability | Impact | Mitigation |
|---|---|---|---|
| Feature parity gap | Medium | High | Side-by-side validation, A/B testing |
| Performance regression | Low | Medium | Continuous benchmarks, profile early |
| Provider compatibility | Medium | High | Integration tests for all 237 providers |
| OAuth complexity | High | Medium | Implement top 5 first, then rest |
| Team Go learning curve | Medium | Medium | Month 1 is foundation/learning |
| Data migration errors | Low | High | Dry-run, checksums, rollback plan |
| Kubernetes operator complexity | Medium | Medium | Start simple, iterate |
| MCP client compatibility | Low | High | Conformance test suite |

---

## 🛠 Tech Stack

| Component | Choice |
|---|---|
| **Language** | Go 1.24+ |
| **HTTP Framework** | `net/http` + `chi` |
| **Envoy Control Plane** | `go-control-plane` |
| **SQLite** | `modernc.org/sqlite` (pure Go) |
| **PostgreSQL** | `pgx` v5 (optional) |
| **Redis** | `go-redis` |
| **Logging** | `log/slog` (stdlib) |
| **Config** | `koanf` |
| **Protobuf** | `buf` build tool |
| **Kubernetes** | `kubebuilder` + `controller-runtime` |
| **Testing** | `testing` (stdlib) + `testify` + `gomock` |
| **Validation** | `go-playground/validator` |
| **CI/CD** | GitHub Actions |
| **Container** | Docker + Docker Compose |

---

> **Total: 210 days = 7 months**
>
> **Parallel tracks possible**: Months 4-5 (MCP + A2A + Memory + Skills) can run in parallel with DevOps/K8s work.
>
> **v4.0.0 Release**: Fully migrated OmniRoute in Go with Envoy control plane and Kubernetes operator.
</content>
<task_progress>
- [x] Created full architecture diagram with Envoy control plane, Go controllers, Kubernetes operator
- [x] Documented complete Go project structure with all files
- [x] Included slog logging pattern with code examples
- [x] Included controller pattern (watcher/reconciler) with Go code
- [x] Included xDS server and snapshot manager code
- [x] Included Envoy bootstrap config
- [x] Created day-by-day task plan for all 7 months (210 days)
- [ ] Verify the content was appended successfully and the file is complete
</task_progress>

[Response interrupted by user]</fragment>
