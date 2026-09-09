# kan-260-guards-are-blind-to-cross-repo-changes — design (brainstorm record)

Date: 2026-09-09 · Change: kan-260-guards-are-blind-to-cross-repo-changes · Jira: KAN-260

## Problem

KAN-260 (self-review finding from KAN-253) reported both guards blind on the hand-shipped KAN-343
cross-repo change. KAN-363 delivered link-based resolution; two residuals remain in today's code:

1. A worktree carrying **no change directory at all** stays blind: `change-plan.sh` consults the
   canonical worktree only through a `link.md`, so `check-task-commit-fields.sh` exits 2 with no
   verdict and `check-unfinished-work.sh` reports `OUTSTANDING: no plan at …`.
2. `check-unfinished-work.sh`'s store calls pass `-C "$WORKTREE"`; the project key derives from
   `-C`'s git common dir, so from the second repo the findings query reads the wrong project
   (`[]` — a false CLEAR on the satellite's own verdict line) and the verdict is recorded where
   the canonical project's tools never look.

## Scope (operator decision)

Fix both residuals. The ticket as written (state-record resolution, explicit change-dir argument)
is superseded by what KAN-363 already shipped; the residue is fixed inside the existing
canonical-worktree pattern.

## Design

1. **Absent-dir canonical resolution** (`scripts/lib/change-plan.sh`): in
   `_change_plan_resolve_dir`, after the local `tasks.md` check and before the `link.md` branch —
   local change directory absent entirely + canonical worktree supplied +
   `<canonical>/<spec-root>/changes/<name>/tasks.md` exists → resolve there. Both guards inherit
   it; `check-task-commit-fields.sh` additionally routes its named-change path through the lib
   when the local dir is absent (its 6th `<name>` argument, which `implement.md` already passes,
   names the change). Without a canonical argument the refusals stay as today.
2. **Plan-anchored store calls** (`scripts/check-unfinished-work.sh`): the findings query and the
   `record verdict` / `record verdicts` calls anchor `-C` at the resolved plan's directory
   (canonical repo for every cross-repo shape), falling back to `$WORKTREE` on the no-plan
   fall-through. `-worktree` keeps naming the judged worktree.
3. **Tests**: real two-repo fixtures in `test-lib-change-plan.sh`,
   `test-check-task-commit-fields.sh`, `test-check-unfinished-work.sh` (stub `flow` asserts the
   `-C` it received); existing cases green unchanged.

## Alternatives rejected

- State-record worktree-list resolution — wrong-project key from the satellite side; store coupling.
- New explicit `<change-dir>` argument — redundant with canonical-worktree; caller drift.
- Project-key arithmetic in bash — re-implements `ProjectKey`, drifts from the CLI.
- Canonical-by-name ahead of link resolution — silent mis-resolution risk; today's linked-satellite
  failures are loud by design.

## Baselines (measured)

- `bash scripts/test-lib-change-plan.sh` → 32 ok
- `bash scripts/test-check-task-commit-fields.sh` → 199 ok
- `bash scripts/test-check-unfinished-work.sh` → 84 ok
