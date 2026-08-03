## Context

`/myflow-do`'s NO-COMMITS rule forbids committing until a PR exists, so every task and every fix
round happens as uncommitted worktree changes. `skills/myflow-do/SKILL.md` currently tells the
parent to record `TASK_BASE=$(git rev-parse HEAD)` before dispatching each task, and
`FIX_BASE=$(git rev-parse HEAD)` before each fix round, then hand the reviewer `git diff TASK_BASE`
/ `git diff FIX_BASE`.

Because nothing is ever committed, `HEAD` never advances across a run — every task's `TASK_BASE` is
the *same* sha. `git diff TASK_BASE` for task 2 therefore includes task 1's still-uncommitted work
too; it is not the isolated per-task diff the review loop needs. `superpowers:subagent-driven-development`'s
own `review-package` script cannot stand in either: it assumes a real commit range
(`git log BASE..HEAD`, `git diff BASE..HEAD`), which is empty or wrong when nothing is committed.

This was worked around ad hoc during KAN-31's `/myflow-do` run, hand-building each review package
with `git diff <checkpoint>` and inventing `git stash create` on the spot to get a non-destructive
snapshot object usable as an isolating BASE. This change makes that reusable rather than re-derived
per run.

## Goals / Non-Goals

**Goals:**
- Give `/myflow-do` a way to isolate exactly one task's (or one fix round's) own uncommitted changes
  from prior tasks' still-uncommitted changes, without committing anything.
- Give it a review-package file format matching `review-package`'s existing shape (stat + diff,
  written into the same plan-scoped `.superpowers/sdd/<plan>/` workspace) so reviewers read one file
  exactly as they already do for the committed case.

**Non-Goals:**
- Editing `subagent-driven-development`'s `review-package` script itself. It ships from the
  `superpowers` plugin cache, not this repository; a local edit would be silently overwritten by the
  next plugin update and is not this repository's to maintain.
- Changing `final-review.diff` generation, the review panel's slot roster, or any other myflow
  command. `final-review.diff` (`git diff <merge-base>` over the whole accumulated worktree) is
  intentionally cumulative across every task and fix round, not per-task, so it has no isolation bug
  to fix.

## Decisions

### Isolate per-task/per-round diffs via `git stash create`, not a real commit

**ID:** stash-create-isolation
**Status:** active
**Chosen:** record `BASE = git stash create` (falling back to `HEAD` on a clean tree) before each
task/fix-round dispatch, replacing `git rev-parse HEAD` — a non-destructive snapshot that actually
isolates each task's own changes, since HEAD itself never advances under NO-COMMITS.
**Considered:** keeping `git rev-parse HEAD` and asking reviewers to manually exclude prior tasks'
changes — rejected, since it re-creates exactly the hand-built, error-prone process this change
exists to eliminate; making an actual commit per task and squashing later — rejected, since it
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

## Risks / Trade-offs

- **A `git stash create` object is unreachable from any ref and can be garbage-collected.** →
  Mitigation: it only needs to survive from one `checkpoint` call to the next `uncommitted-review-package`
  call within the same session, and nothing in this change holds one across a `git gc`; this is the
  same lifetime `review-package`'s own commit-range arguments already assume for their BASE/HEAD.
- **`git stash create` on a clean tree prints nothing, which a naive caller could mistake for an
  error.** → Mitigation: `checkpoint` explicitly detects this case and prints the current `HEAD` sha
  instead, so its output is always a valid commit-ish; this is exercised directly by the new test.
- **`uncommitted-review-package`'s single-ref diff (`git diff -U10 BASE`) also picks up any changes
  made by a process other than the dispatched subagent (e.g. a stray manual edit in the worktree).**
  → Mitigation: this is the same trust boundary `git diff TASK_BASE` already had; nothing in this
  change widens it.

## Migration Plan

No data or state migration. This is additive: two new scripts and a rewrite of two paragraphs in
`skills/myflow-do/SKILL.md` describing which commands the parent runs during its own dispatch loop.
No existing state file, ledger, or review-package format changes shape. Rollback is reverting the
commit; no other component depends on the new scripts existing.

## Open Questions

None.
