## Why

Routine command output (`updated: dispatch 69` lines) fills the operator's terminal; Claude Code has no setting to hide a single tool result (2026-10-02).

## What changes

- `rules/be-brief.mdc`: routine output a command prints and the session does not need goes to `/dev/null`; errors stay visible.
