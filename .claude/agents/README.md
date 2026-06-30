# .claude/agents/

Custom subagents: one Markdown file per agent, YAML frontmatter
(`name`, `description`, `tools`, optionally `model`) plus the agent's
system prompt as the file body. Each subagent runs in its own context
window, invoked by Claude automatically (matched against `description`)
or directly via `@<agent-name>`.

Empty for now. A natural first candidate for this repo: a read-only
`lint-reviewer` restricted to `Read, Grep, Glob` that checks a diff against
the `.golangci.yml` rules documented in the root `CLAUDE.md` before code is
considered done.
