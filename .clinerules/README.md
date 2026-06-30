# .clinerules

Cline (the VS Code extension) loads every file in this folder as project
rules — Cline's official multi-file convention, which lets individual rule
files be toggled on/off from the Cline UI without editing content.

- `01-architecture-and-testing.md` — hexagonal architecture, SOLID, Go
  design patterns, TDD/table-driven testing, YAML/env config, slog logging,
  and the lint-first workflow. This is the condensed counterpart to the
  root `CLAUDE.md`, which Claude Code reads instead of this folder.

Keep this folder and `CLAUDE.md` in sync if you change either — they
encode the same ruleset for two different tools. Add new files here
(e.g. `02-<topic>.md`) rather than growing the existing file unboundedly
once a section becomes large enough to stand on its own.
