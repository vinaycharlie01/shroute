# .claude

Project-level configuration for Claude Code (claude.ai/code), following the
official directory layout:

- `commands/` — custom slash commands (`*.md` files, invoked as
  `/<filename>`). Empty for now; add a file here when a repeatable workflow
  (e.g. "run the full lint/test/coverage gate") is worth turning into a
  command.
- `agents/` — custom subagent definitions (`*.md` files with frontmatter:
  `name`, `description`, `tools`). Empty for now; add one when a task in
  this repo needs a dedicated, narrowly-scoped agent (e.g. a lint-fixer or
  integration-test runner).
- `settings.json` — not created yet. Add it if/when the repo needs
  project-specific permissions or hooks; until then Claude Code falls back
  to user/global defaults.

The actual architecture/testing/lint rules Claude Code auto-loads live in
the root `CLAUDE.md`, not in this folder — `.claude/` is only for
commands, agents, and settings. Keep `CLAUDE.md` and `.clinerules/` in
sync with each other when either changes.
