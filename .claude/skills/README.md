# .claude/skills/

Reusable, invokable workflows. Each skill is a folder containing a
`SKILL.md` (frontmatter: `description`, optionally `disable-model-invocation`
or `user-invocable: false`) plus any supporting files it needs (checklists,
templates, scripts).

Invoked as `/<skill-name>`, by Claude automatically when the task matches
the skill's `description`, or both — controlled per-skill via frontmatter.

This is the current officially-preferred mechanism for new repeatable
workflows (superseding single-file `commands/` for anything that benefits
from bundled reference docs). Empty for now — add one here, for example,
when the lint-first verification sequence in `CLAUDE.md`
(`go build && go vet && gofmt -l . && golangci-lint run && go test ./...`)
is worth turning into `/verify`.
