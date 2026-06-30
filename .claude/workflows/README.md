# .claude/workflows/

Dynamic workflow scripts (`.js`) that orchestrate multiple subagents,
each becoming a `/<name>` command. These are written by Claude and saved
here via the `/workflows` command — not hand-authored from scratch.

Empty for now; nothing in this repo's flow has needed multi-agent
orchestration yet (the audit-slice 8-step flow in `backend/tasks/` is
followed by a single session, not a saved workflow).
