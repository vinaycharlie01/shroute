# .claude/commands/

Single-file slash commands (`*.md`, invoked as `/<filename>`,
`$ARGUMENTS`/`$0`/`$1`... for parameters). Still officially supported, but
new workflows should generally go in `skills/` instead — skills use the
same `/name` invocation and additionally let you bundle reference docs,
templates, or scripts alongside the prompt.

Empty for now; add a command here only for something genuinely single-file
that doesn't need supporting files (otherwise prefer `skills/`).
