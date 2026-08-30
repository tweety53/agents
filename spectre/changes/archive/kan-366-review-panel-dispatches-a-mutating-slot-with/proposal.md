# kan-366-review-panel-dispatches-a-mutating-slot-with

## Why

`review-panel.md` dispatches every resolved review slot concurrently into the single apply
worktree. Bugbot's mutation-testing brief requires it to edit code in place (flip a condition,
drop a guard, mutate a branch, then revert). On KAN-173's own `/flow` run this collided with the
concurrently-dispatched reading slots twice: Principles reverted Bugbot's live mutation believing
it an unrelated leftover, and Code review (low) observed intermittent nil-pointer panics from the
same mutation and misattributed them to sandbox contention. The dispatcher had to discard the
compromised pass and re-dispatch Bugbot exclusively — wasted agent time, and a corrupted mutation
gate on any run where the collision goes unnoticed.

## What changes

- Bugbot's dispatch — pass 1 and every fix-round re-run, including a substituted (general-purpose)
  Bugbot — runs against a throwaway git worktree copy of the apply worktree's current state,
  never against the shared `<worktree>` reading slots also read.
- The copy is created immediately before Bugbot's dispatch and removed immediately after that
  dispatch closes (success, timeout, or a stopped run) — it never survives to `/flow`'s
  worktree-cleanup phase.
- Findings, their `file:line` locations, and reproducer verification are unaffected: reproducers
  still run against the real `<worktree>`, exactly as today.
- `artifacts-registry.md` gains one row for the new throwaway copy.
