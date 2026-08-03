## Why

`/myflow-do`'s NO-COMMITS rule forbids committing until a PR exists, so `SKILL.md` currently tracks
per-task and per-fix-round review bases as `git rev-parse HEAD`. Since HEAD never advances while
nothing is committed, every task records the *same* base, so `git diff TASK_BASE` for task 2
silently includes task 1's still-uncommitted work — it is not the isolated diff the review loop
needs. This was worked around by hand during KAN-31's `/myflow-do` run using `git stash create`;
this change makes that reusable rather than re-derived per run.

## What Changes

- Add `skills/myflow-do/scripts/checkpoint`: prints a non-destructive snapshot of the current
  index+worktree state (`git stash create`, falling back to `HEAD` on a clean tree) suitable as an
  isolating BASE, without touching any ref, the index, or the stash list.
- Add `skills/myflow-do/scripts/uncommitted-review-package PLAN_FILE BASE [OUTFILE]`: writes a
  review package (files-changed stat + a `-U10` diff of BASE against the live working tree) into the
  same `.superpowers/sdd/<plan>/` workspace `subagent-driven-development`'s `review-package` already
  uses, but without a `## Commits` section, since none exist under NO-COMMITS.
- Rewrite `skills/myflow-do/SKILL.md` step 4 (per-task review) and the fix-round paragraph in step 5
  to call these two scripts instead of the inline `git rev-parse HEAD` / `git diff TASK_BASE` prose.
- Add `scripts/test-uncommitted-review-package.sh` covering both new scripts.

## Capabilities

### New Capabilities

- `myflow-uncommitted-review-package`: generating an isolated, per-task/per-fix-round review package
  from uncommitted worktree changes during `/myflow-do`'s SDD + TDD execution stage.

### Modified Capabilities

(none — no existing spec's requirements change; this adds a capability the pipeline did not have)

## Impact

- `skills/myflow-do/scripts/` (new directory): `checkpoint`, `uncommitted-review-package`
- `skills/myflow-do/SKILL.md`: step 4 and the fix-round paragraph in step 5
- `scripts/test-uncommitted-review-package.sh` (new)
- No change to `final-review.diff` generation, the review panel roster, or any other myflow command
