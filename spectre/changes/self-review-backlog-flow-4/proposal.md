# self-review-backlog-flow-4

## Why

Five tickets in KAN's "Backlog Flow" status, filed by earlier self-review passes, are planned here
as one self-review fix change. Two are already closed (`design.md`, **Context**); three name gaps
the tree still has:

- **KAN-652** — gymie-frontend's `HomeCalendarMarksUiTest` checks that its own tag literal carries
  no `_inverted` infix, a tautology, and asserts no mark's colour, so a mark inverted to the
  disc's ink or drawn in another kind's colour passes.
- **KAN-691** — kan-527's error-path fix passed a unit test against a synthetic `403` the real
  backend never sends on that path, while a global interceptor made the fixed handler
  unreachable; nothing the implementers or the principles reviewer read asks for the real trigger.
- **KAN-698** — kan-573's plan confirmed a prerequisite against `origin/develop` while the
  worktree's own HEAD lacked it; nothing asks a prerequisite to be an ancestor of the worktree's
  HEAD.

## What changes

- **Testing principles** (`skills/flow/engineering-principles.md`) requires an error-path fix's
  test to trace the real response through every interceptor and drive the real trigger.
- **Plan provenance** (`skills/flow-contracts/plan-provenance.md`) measures a prerequisite with
  `git merge-base --is-ancestor <sha> HEAD` in the change's worktree, re-run as the dependent
  task's first step.
- gymie-frontend's `HomeCalendarMarksUiTest` reads each mark's ink from its pixels and asserts its
  kind's colour — own accent, borrowed accent-2, session text.
