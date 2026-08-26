# Tasks

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Two independent doc-only edits: task 1 touches `skills/myflow-contracts/finish-contract.md` (two
related notes — docker-restart guidance and the check-6 `cd`-out precondition, both inside its Run
2 / Worktree cleanup procedure); task 2 touches `skills/flow/implement.md` (the isolate-workspace
copy-and-cleanup step). Neither touches code, so there is no test suite to add to; each task's own
verification is the project's Markdown/prose guards.

- [x] 1. `finish-contract.md` — docker-restart guidance and the check-6 `cd`-out precondition

  Add a note near line 352-367 (step 5's "A failed removal does not stop run 2" and step 7's "A
  `SKIPPED:` clause on a `COMPLETE:` line is relayed" text), stating plainly: once worktree
  cleanup's check 5 has stopped the project's declared stack, every later run-2 step already
  degrades correctly when the store is unreachable — a reported-and-continued removal failure here,
  a `SKIPPED:` clause at step 7, a journaled state write at step 8 — and none of those outcomes is a
  cue to bring the stack back up. Add a second note at check 6 (line ~561, "no process is still
  running FROM this worktree"), stating that the orchestrating shell's own cwd must be outside every
  worktree in the resolved set before this check runs, since the guard's own header already treats a
  shell cwd'd inside a worktree as a process holding it. Run
  `scripts/check-plan-provenance.sh`, `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
  and `scripts/check-contract-budget.sh` and confirm all four pass (the new prose may need a budget
  bump in `check-contract-budget.sh`'s `budgets()` table for this file if it pushes past the
  existing row's cap — check and raise it if so, per that guard's own instructions).

**Build:** green
**Files:** `skills/myflow-contracts/finish-contract.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh` — only its `budgets()` table row for
`skills/myflow-contracts/finish-contract.md`, raised if this task's addition exceeds the existing
budget; no other row.
**Tests:** none named — doc-only change, verified by the four guards named above
**Regression:** reverting this task leaves the two clarifications unstated, which is how the
docker-thrashing and cwd-false-positive frictions this change exists to prevent were reached in the
first place.
**Baseline:** not applicable — no test suite exists for this file; verified by the four guard
commands passing before and after.
**Commit:** `docs(finish-contract): clarify graceful degradation and the check-6 cwd precondition`

- [x] 2. `implement.md` — document copying and cleaning up the change's planning artifacts

  In `skills/flow/implement.md`'s "2. Isolate the workspace (first run only)" section, after the
  worktree is created (following the `superpowers:using-git-worktrees` invocation and the merge-base
  recording), add: copy `<project>/spectre/changes/<name>/` from the main checkout into the same
  path inside the new worktree, then remove the main checkout's own copy of that directory. State
  plainly why: once the worktree's own copy becomes the one that gets edited and eventually
  committed, the main checkout's copy is stale and untracked, and nothing else in the pipeline ever
  removes it — leaving it in place is what made `prepare-archive-branch.sh` refuse on a dirty main
  checkout during KAN-271's own run. Name this as a first-run-only step, matching the section's own
  scope (a fix run resumes the existing worktree and makes no such copy). Run
  `scripts/check-plan-provenance.sh`, `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
  and `scripts/check-contract-budget.sh` and confirm all four pass (same budget-row caveat as task
  1, scoped to `skills/flow/implement.md` here).

**Build:** green
**Files:** `skills/flow/implement.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh` — only its `budgets()` table row for
`skills/flow/implement.md`, raised if needed; no other row.
**Tests:** none named — doc-only change, verified by the four guards named above
**Regression:** reverting this task leaves isolate-workspace silent about the copy-and-cleanup step,
reproducing the exact stray-copy defect this change fixes.
**Baseline:** not applicable — no test suite exists for this file; verified by the four guard
commands passing before and after.
**Commit:** `docs(flow): copy and clean up planning artifacts when isolating the workspace`
