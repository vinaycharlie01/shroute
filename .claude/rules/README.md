# .claude/rules/

Topic-scoped instructions, split out of `CLAUDE.md` once a section gets
large or only applies to certain files.

- A rule file with no `paths:` frontmatter loads at session start, same as
  `CLAUDE.md`.
- A rule file with `paths:` frontmatter (glob list) loads only when Claude
  reads a matching file — e.g. a `testing.md` rule scoped to `**/*_test.go`.
- Subdirectories are discovered automatically (`rules/backend/mongo.md`).

Empty for now: `CLAUDE.md` is still under the ~200-line point where the
official guidance recommends splitting into `rules/`. Move a section here
(e.g. testing conventions, lint rules) when it grows enough to stand alone,
rather than letting `CLAUDE.md` keep growing unbounded.
