# .claude

Project-level configuration for Claude Code (claude.ai/code), following the
official directory layout (see each subfolder's own `README.md` for
specifics):

- `rules/` — topic-scoped instructions split out of `CLAUDE.md`, optionally
  gated to specific files via `paths:` frontmatter.
- `skills/` — reusable invokable workflows (`<name>/SKILL.md` + supporting
  files), the current officially-preferred mechanism for new `/<name>`
  workflows.
- `commands/` — single-file slash commands (`*.md`, invoked as
  `/<filename>`); still supported, but prefer `skills/` for anything new
  that needs bundled reference docs.
- `agents/` — custom subagent definitions (`*.md` with frontmatter:
  `name`, `description`, `tools`, optionally `model`), each running in its
  own context window.
- `output-styles/` — team-shared custom output styles (most output styles
  are personal and live in `~/.claude/output-styles/` instead).
- `workflows/` — dynamic multi-subagent orchestration scripts (`.js`),
  saved via the `/workflows` command rather than hand-authored.

All six are currently empty placeholders (each with its own README); add
content to a given folder only when a concrete need shows up, per "create
it now, fill it in if/when needed."

Intentionally **not** created yet, since they require real project-specific
content rather than scaffolding:

- `settings.json` / `settings.local.json` — project-specific permissions,
  hooks, env vars. Until added, Claude Code falls back to user/global
  defaults. `settings.local.json` is auto-gitignored by Claude Code itself
  the first time it's written, so don't hand-create it.
- `agent-memory/` — auto-generated only for subagents that set
  `memory: project` in their frontmatter; nothing under `agents/` does yet.
- `.mcp.json` and `.worktreeinclude` — live at the **project root**, not
  inside `.claude/`; both need real project-specific values (actual MCP
  servers / actual gitignored paths to copy into worktrees) rather than
  placeholder content, so they're left for whoever first needs them.

The actual architecture/testing/lint rules Claude Code auto-loads at session
start live in the root `CLAUDE.md`, not in this folder. Keep `CLAUDE.md` and
`.clinerules/` in sync with each other when either changes.
