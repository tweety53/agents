# self-review-self-review-backlog-flow-6

## Why

The self-review pass of `self-review-backlog-flow-6` found three fixable pipeline defects:

- The stats store gives concurrent inline dispatch rows the wrong share of a run's tokens.
  `DispatchWindowsForSession` (`stats/internal/store/records.go`) caps an open stage-less row at
  the next inline begin in its session, so of the two per-worktree inline verifier rows that
  **Recording.** (`skills/flow/verify-and-handoff.md`) has one Bash call begin together, the first
  covered only the gap before the second began and the second took every record of the verify
  stage. Every multi-repository run hits it.
- `skills/flow/brainstorm-planner.md` section D never tells a planner that a task does not create
  a worktree, and a plan task did — with a bare `git worktree add` on a branch outside the
  `spectre/<name>` convention, duplicating **2. Isolate the workspace**
  (`skills/flow/implement.md`) and leaving a stale decision in the archived design.
- Every verify stage, the parent hand-writes a runner for the `## lint` and `## test` lists that
  **Inline verify** (`skills/flow/verify-and-handoff.md`) prescribes in prose — fence extraction,
  stdin detach, the `## Report` line shape — and the last run patched its stdin detach in on a
  second call.

## What changes

- An inline begin under `minInlineWindow` after an open stage-less row's own begin no longer caps
  that row: both concurrent verifier rows keep their windows open and share the records between
  them. The cap at a real successor is unchanged.
- Section D of `skills/flow/brainstorm-planner.md` carries one rule: a task never creates a
  worktree.
- `scripts/run-verify-list.sh <worktree> lint|test` runs one list as **Inline verify** prescribes
  and prints its `## Report` lines; **Inline verify** names it, and a harness pins its output and
  exit codes.
