# Uncommitted-diff checkpoint + review-package helper for `/myflow-do`

## Why

`/myflow-do`'s NO-COMMITS rule forbids committing until a PR exists, so every task and every fix
round happens as uncommitted worktree changes. `skills/myflow-do/SKILL.md` currently tells the
parent to record `TASK_BASE=$(git rev-parse HEAD)` before dispatching each task, and again
`FIX_BASE=$(git rev-parse HEAD)` before each fix round, then hand the reviewer `git diff TASK_BASE`
/ `git diff FIX_BASE`.

Because nothing is ever committed, `HEAD` never advances across a run — every task's `TASK_BASE` is
the *same* sha. `git diff TASK_BASE` for task 2 therefore includes task 1's still-uncommitted work
too; it is not the isolated per-task diff the review loop needs. `superpowers:subagent-driven-development`'s
own `review-package` script cannot stand in either: it assumes a real commit range
(`git log BASE..HEAD`, `git diff BASE..HEAD`), which is empty or wrong when nothing is committed.

This was worked around ad hoc during KAN-31's `/myflow-do` run, hand-building each review package
with `git diff <checkpoint>` and inventing `git stash create` on the spot to get a non-destructive
snapshot object usable as an isolating BASE. KAN-53 exists to make that reusable rather than
re-derived per run.

## What Changes

Two new scripts under `skills/myflow-do/scripts/` — this skill's first scripts directory, mirroring
how `subagent-driven-development` already splits its own helpers into single-purpose files
(`sdd-workspace`, `review-package`, `task-brief`):

- **`checkpoint`** — no arguments. Prints a commit-ish representing the current index+worktree
  state, non-destructively: `git stash create` when there are uncommitted changes, or the current
  `HEAD` sha when the tree is clean (`git stash create` prints nothing when there is nothing to
  snapshot, which is the ordinary shape of the very first task in a run). Touches no ref, no index,
  no stash list — nothing about `git status`, `git stash list`, or `HEAD` changes as a result of
  calling it.
- **`uncommitted-review-package PLAN_FILE BASE [OUTFILE]`** — mirrors `review-package`'s package
  shape but for a BASE that is not necessarily an ancestor relationship worth logging:
  - writes into the **same** `.superpowers/sdd/<plan>/` workspace, resolved via
    `subagent-driven-development`'s own `sdd-workspace` script, so files from both scripts sit side
    by side under one plan-scoped directory
  - diffs `BASE` against the **live working tree** with a single-ref `git diff -U10 BASE` — not
    `BASE..HEAD`, since under NO-COMMITS the work never reaches `HEAD`
  - keeps a `## Files changed` section (`git diff --stat BASE`)
  - drops the `## Commits` section entirely (there are none to list) in favor of one line noting
    that the package covers uncommitted work relative to `BASE`

`skills/myflow-do/SKILL.md` step 4 (per-task review) and the fix-round paragraph in step 5 are
rewritten to call these two scripts in place of the inline `git rev-parse HEAD` /
`git diff TASK_BASE` prose. `final-review.diff` (step 5's whole-branch package, `git diff <merge-base>`
over the full accumulated worktree) is unchanged — it is intentionally cumulative across every task
and fix round, not per-task, so it has no isolation bug to fix.

### Out of scope

- Editing `subagent-driven-development`'s `review-package` script itself. It ships from the
  `superpowers` plugin cache, not this repository; a local edit would be silently overwritten by
  the next plugin update and is not this repository's to maintain.
- Any change to the final whole-branch review package or to the review panel's slot roster.

## Data flow

Per task:

```
TASK_BASE=$(skills/myflow-do/scripts/checkpoint)
# ...implementer subagent works, leaves changes uncommitted...
skills/myflow-do/scripts/uncommitted-review-package tasks.md "$TASK_BASE"
# ...task reviewer reads the single package file printed on stdout...
```

Fix rounds follow the identical shape with `FIX_BASE` / `fix-round-N.diff` as the explicit
`OUTFILE` argument, so a re-review after fixes gets a distinct, freshly named file exactly as
`review-package` already guarantees for the committed case.

## Error handling

Both scripts run under `set -euo pipefail`.

- `checkpoint` takes no arguments; it never fails except through `git` itself (e.g. not inside a
  work tree), in which case `git`'s own error propagates.
- `uncommitted-review-package` validates, in order: the plan file exists (else exit 2, "no such plan
  file: `<path>`"); `BASE` resolves via `git rev-parse --verify --quiet` (else exit 2, "bad BASE:
  `<base>`") — the same two checks `review-package` already makes on its own arguments.

## Testing

A new `scripts/test-uncommitted-review-package.sh`, matching this repository's existing
`scripts/test-*.sh` convention, covering:

- clean-tree `checkpoint` falls back to the current `HEAD` sha (stash create prints nothing)
- dirty-tree `checkpoint` returns a stash-create object distinct from `HEAD`
- a second round of changes made *after* a checkpoint is captured by
  `uncommitted-review-package` against that checkpoint, while changes made *before* it are not —
  the actual isolation property this change exists to provide
- the output file's shape: a `## Files changed` section and a `## Diff` section, and no `## Commits`
  section
- the existing-plan-file and resolvable-BASE error paths, matching `review-package`'s own two
  failure modes

## Decisions

### Isolate per-task/per-round diffs via `git stash create`, not a real commit

**ID:** stash-create-isolation
**Status:** active
**Chosen:** record `BASE = git stash create` (falling back to `HEAD` on a clean tree) before each
task/fix-round dispatch, replacing `git rev-parse HEAD` — a non-destructive snapshot that actually
isolates each task's own changes, since HEAD itself never advances under NO-COMMITS.
**Considered:** keeping `git rev-parse HEAD` and asking reviewers to manually exclude prior tasks'
changes — rejected, since it re-creates exactly the hand-built, error-prone process the ticket was
filed to eliminate; making an actual commit per task and squashing later — rejected, since it
directly violates the NO-COMMITS rule this change operates inside of.

### Two small single-purpose scripts, not one combined script

**ID:** two-scripts-not-one
**Status:** active
**Chosen:** a `checkpoint` script and a separate `uncommitted-review-package` script, matching the
existing convention in `subagent-driven-development` (`sdd-workspace` / `review-package` /
`task-brief` are already split this way) and keeping `BASE` an explicit shell variable the parent
tracks, consistent with how `SKILL.md` already documents the per-task loop.
**Considered:** one combined script that internally remembers its own last checkpoint in the sdd
workspace — rejected, since it would make `BASE` implicit state instead of a variable visible in the
dispatch loop, diverging from how step 4 already documents `TASK_BASE`.

## Open questions

None.
