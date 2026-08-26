# kan-337-fix-three-flow-finish-phase-friction-points

## Why

Running `/flow` end to end for KAN-271 surfaced three real friction points in the finish/integrate
procedure — none logic bugs in a guard, each an operational rough edge that cost real wall-clock or
extra round-trips: unnecessary `docker compose` cycling during archive, a live-process guard
false-positive caused by the calling procedure's own unstated precondition, and a stray untracked
planning-artifact copy left in the main checkout that blocks `prepare-archive-branch.sh`.

## What changes

- `skills/myflow-contracts/finish-contract.md` gains a clarifying note: once worktree cleanup's
  check 5 stops the project's declared stack, every later run-2 step (workspace removal,
  `check-cleanup-complete.sh`, the state write) already degrades correctly when the store is
  unreachable — that is the intended terminal outcome, not a signal to restart the stack.
- `skills/myflow-contracts/finish-contract.md`'s Worktree cleanup section states, at check 6, that
  the orchestrating shell's own cwd must be outside every worktree in the resolved set before that
  check runs — the guard already treats a shell cwd'd inside as a held process; the calling
  procedure now says so.
- `skills/flow/implement.md`'s isolate-workspace step documents copying the change's planning
  artifacts into the new worktree and removing the main checkout's own now-redundant copy, so
  nothing untracked is left behind for `prepare-archive-branch.sh` to trip over at archive time.
