# Task 12: Skills & Plugins

**Complexity**: High — introduces a sandboxed execution model (running
user/operator-defined skill code with bounded capabilities), which is a
qualitatively different problem from anything in Tasks 01-11.

**TS source**: `OmniRoute/vinaydoc/SLICE_15_SKILLS_PLUGINS.md` —
`/api/skills/*`, `/api/plugins/*`, `/api/agent-skills/*`. Cross-reference
`docs/frameworks/SKILLS.md` in the OmniRoute TS repo for the sandbox
contract this slice must preserve.

## End-to-end flow

1. **Domain** — `internal/domain/skill/skill.go`: `Skill{ID, Name string,
   Manifest Manifest, Enabled bool}`, `Manifest{EntryPoint string,
   Permissions []string, Timeout time.Duration}`. `internal/domain/plugin/plugin.go`:
   `Plugin{ID, Name, Source string, Installed bool}`.
2. **Ports** — `SkillRepository` (CRUD), `SkillSandbox`
   (`Run(ctx, skill, input []byte) (output []byte, err error)` — the actual
   sandboxed execution, isolated behind its own port exactly like
   `ProviderProbe`/`BlobStore` before it), `PluginRepository` (CRUD +
   install/uninstall) in `ports.go`.
3. **Application** — `internal/application/skill/service.go`: enforces
   `Manifest.Timeout` via `context.WithTimeout` around `SkillSandbox.Run`,
   never trusts the sandbox to self-limit; rejects skills whose
   `Permissions` exceed what the calling API key's scopes allow (reuses
   `ApiKeyRepository` scopes from Task 04). **No `eval()`/dynamic code
   execution in the Go process itself** (CLAUDE.md hard rule #3 carries over
   from TS) — the sandbox adapter must isolate execution via OS-level means
   (separate process, restricted FS/network) rather than evaluating
   arbitrary code in-process.
4. **Outbound adapters** — `internal/adapters/outbound/mongodb/{skill,plugin}.go`
   for metadata; new `internal/adapters/outbound/skillsandbox/` package —
   start with a subprocess-based sandbox (`os/exec` with a restricted
   environment via `env`, never string-interpolated into a shell per
   CLAUDE.md hard rule #13) as the v1 implementation.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/{skill,plugin}.go`:
   CRUD + `POST /api/skills/{id}/invoke`.
6. **Router/DI** — usual extension pattern; sandbox timeout/resource limits
   come from `config.Skills`.
7. **Tests** — unit tests for timeout enforcement and permission-scope
   checks (fake `SkillSandbox`); integration test running a trivial
   subprocess skill end-to-end and asserting it cannot escape its declared
   permissions (e.g. no network access when `Permissions` excludes it).

## Checklist

- [ ] `internal/domain/skill`, `internal/domain/plugin`
- [ ] `SkillRepository`, `SkillSandbox`, `PluginRepository` ports
- [ ] `internal/application/skill/service.go` (timeout + scope enforcement) + unit tests
- [ ] Mongo metadata adapters + subprocess `skillsandbox` adapter (env-based, no shell interpolation) + integration test
- [ ] Handlers + router wiring
- [ ] `config.Skills` + DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
