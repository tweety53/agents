# kan-260-guards-are-blind-to-cross-repo-changes

KAN-260 · myflow: guards are blind to cross-repo changes (check-task-commit-fields, check-unfinished-work)

## Why

KAN-260 is the self-review finding from KAN-253; KAN-363 later shipped link-based cross-repo
resolution (`link.md` both ways, `spectre link`, `scripts/lib/change-plan.sh`, canonical-worktree
arguments). Measured against today's code, two residuals remain:

- **A worktree with no change directory at all stays blind.** `change-plan.sh` consults the
  supplied canonical worktree only through a `link.md`; with no local dir there is no link to
  read. `check-task-commit-fields.sh` exits 2 with no verdict; `check-unfinished-work.sh`
  reports `OUTSTANDING: no plan at …`. Reachable whenever guards run outside a current `/flow`
  run — exactly the hand-shipped KAN-343 shape the ticket reported.
- **The store calls anchor at the wrong project.** `check-unfinished-work.sh` queries
  `flow record findings` and writes its verdicts with `-C "$WORKTREE"`; the project key resolves
  from `-C`'s git common dir (`stats/internal/fallback/statefile.go`), so from the second repo's
  worktree the guard reads the second project's findings — `[]`, a false CLEAR on its own verdict
  line — and records the verdict where the canonical project's tools never look.

## What changes

- Both guards reach a verdict for a same-named plan living only in the canonical worktree: when
  the local change directory is absent entirely and a canonical worktree was supplied, the plan
  resolves there by name. No `link.md`, no new argument, no store access.
- `check-unfinished-work.sh`'s findings query and both verdict calls anchor `-C` at the directory
  the plan resolved from — the plan-owning repo's project — instead of always the judged worktree.
- Real two-repo fixtures cover the new shapes in all three harnesses
  (`test-lib-change-plan.sh`, `test-check-task-commit-fields.sh`, `test-check-unfinished-work.sh`);
  existing cases stay green unchanged.
