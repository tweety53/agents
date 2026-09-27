# withdraw-changes-abandoned-before-planning

## Why

A change abandoned at `STARTED` leaves permanent litter: its record appears in every future
candidate resolution and `/flow-status` report forever, and its worktree, local branch and pushed
remote branch outlive it with no documented path to retire any of them. The route that does end a
creating run early — the `flow-fix`/`flow-cost` reachability check — reports "the issue is the
operator's to close" but leaves the change itself no way to close. The live instance is
`kan-828-flow-fix-refreshing-main-checkouts-before-the`: a `STARTED` record no path can retire and
a zero-commit branch pushed to the remote, both from a finding the base had already delivered.

## What changes

- A withdrawal route in `/flow`'s creating run, offered at two ends: the reachability-check end,
  and a resume of a planless `STARTED` change. The explicit answer names exactly what will be
  deleted — worktree, local branch `spectre/<name>`, remote branch — and is the only consent any
  of them get. No new command.
- The route's steps, in order: per-worktree `git status --porcelain` must be empty → worktree
  remove + prune → `git branch -D spectre/<name>` → `git push origin --delete spectre/<name>`
  (a failed remote delete is one reported line, never a blocker) → one state write. Git first,
  record last, so a crash leaves a re-runnable route.
- The record terminates `FINISHED` carrying a new boolean field `withdrawn` (absent/false = not
  withdrawn). No fourth pipeline state: every FINISHED exclusion — the candidate set,
  `/flow-status`'s open set — already applies. The store refuses `withdrawn: true` on a record
  whose state is not `FINISHED`.
- Store column + migration `0031_withdrawn.sql`; the closed-schema API DTO learns the field so
  writes decode and reads serve it; contract and skill docs. The CLI needs no change
  (`state set` is byte-transparent).
