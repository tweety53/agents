# flow-lean-handoff

## Why

- The `IN_PROGRESS` handoff of an implementation or fix run printed a `<details>` block around the
  pre-edit Jira description. Claude Code's terminal renders Markdown without HTML, so the tag
  printed literally and the whole description spilled into the output.
- The same handoff carried fields the operator does not act on: `Panel`, `Visual`,
  `Tooling analysis`, `Staged`, `Records`, `Costs`, `Guards`, the Jira pre-edit echo, `Worktree`,
  `Running`, the review and IntelliJ commands, and a pre-handoff summary of models, findings,
  commits and the live stack.

## What changes

- That handoff prints three parts only: `**Summary:**` (what the round changed for a user, plus
  any line another contract requires in the handoff), `**Decisions:**` (open questions and every
  auto-picked recommended option, or `none`), then `/clear` and `/flow <name>`.
- `pipeline.md` forbids HTML in anything printed to the terminal.
- The Jira pre-edit echo stays on the `STARTED` handoff only; an `IN_PROGRESS` handoff reports an
  append as one `Summary` bullet.
- `/flow-status` keeps its regenerated `IN_PROGRESS` block (`Staged`, `Records`, `Running`, review
  commands); the stack is still started before handoff.
