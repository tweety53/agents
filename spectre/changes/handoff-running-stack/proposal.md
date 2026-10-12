# handoff-running-stack

## Why

Operator, 2026-10-12: "I see no running stack url/data. It must be present in implementation summary. always." The `IN_PROGRESS` handoff of a creating, resuming or fix run hands off a stack the run started, but printed only `**Summary:**`, `**Decisions:**` and `Next:` — no URL, no login.

## What changes

- `skills/flow/verify-and-handoff.md`: the handoff block gains a `**Running:**` section between `**Summary:**` and `**Decisions:**` — every started application's worktree-resolved URL, one login per `## credentials` row, or the line stating why nothing was started. It is never omitted.
- `skills/flow-contracts/pipeline.md`: the live-stack line names worktree-resolved URLs and the `## credentials` logins; the `IN_PROGRESS` handoff's `**Running:**` section is its live-stack line.
- `skills/flow-contracts/handoff-blocks.md`: the `IN_PROGRESS` rendering's `Running:` carries a login line; `/flow`'s implement phase prints `Running` too.
