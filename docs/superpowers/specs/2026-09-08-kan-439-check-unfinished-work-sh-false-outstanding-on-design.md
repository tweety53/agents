# kan-439 — check-unfinished-work.sh false OUTSTANDING on cross-repo changes with no spectre link

Date: 2026-09-08 · Status: approved (operator, convergence confirm) · Jira: KAN-439

## Problem

A cross-repo change run under a project convention that keeps the spectre change directory in the
canonical repository only (measured on gymie: backend repo holds the tree, the frontend repo has no
`spectre/` or `openspec/` tree at all) leaves every non-canonical worktree with no
`<spec-root>/changes/<name>/` directory — no `tasks.md` and no `link.md`. At integrate, run 1 calls
`check-unfinished-work.sh <worktree> <name> <canonical-worktree>` once per worktree
(`skills/flow-contracts/finish-contract-run1.md`), passing the canonical worktree. The guard's plan
resolution (`scripts/lib/change-plan.sh`) consults that argument **only** through a satellite
`link.md` carrying `## Part of`; with no `link.md` the resolution falls through to "unresolvable"
and the guard reports `OUTSTANDING: no plan at <path>` over a plan that exists, in the canonical
worktree, under the same change name. The operator must hand-verify every cross-repo change.

The other half of KAN-439 — the `spectre link` step being logged non-gating and never retried — was
already fixed by bf684f7 (2026-09-08): implement.md §2 now runs the link with the working directory
at the repository's primary checkout and treats a refusal as a hard failure of the stage. That fix
cannot serve this shape: a repository with no spec tree has nothing for `spectre link --root` to
write a `link.md` into.

## Design

### 1. Same-name resolution through the supplied canonical worktree (`scripts/lib/change-plan.sh`)

`_change_plan_resolve_dir` gains one branch, after the local-`tasks.md` check and the link
resolution fail with **no usable `## Part of` link**:

- Triggered when the change directory has no `tasks.md`, and either there is no `link.md` at all or
  the `link.md` carries no `## Part of` section — the treeless-satellite shape.
- With a canonical worktree supplied and
  `<canonical-worktree>/<spec_root_leaf(canonical)>/changes/<name>/tasks.md` present, that
  directory resolves — the same change name, same containment allowlist as every other name this
  function concatenates. The canonical-worktree argument is caller-supplied today (finish resolves
  it per `finish-contract-run1.md`) and stays caller-supplied; nothing new is read from project
  configuration.
- Without a canonical worktree the branch does not fire; the guard's ordinary missing-plan verdict
  is unchanged.

**No rescue when a `## Part of` link exists but fails to resolve.** A link that names
`<peer>:<change-id>` and comes up empty — canonical worktree supplied but the plan is not there, or
no canonical worktree and the peer cannot be resolved — keeps its current loud failure, and the
guard keeps its exit-2 "cannot determine anything" triage for it. The same-name branch exists for
the shape that has no link to follow; masking a broken link with a name match would turn a
structural fault into a silent verdict.

Callers inherit the branch unchanged: `check-unfinished-work.sh` and `check-task-commit-fields.sh`
(the only `change_plan_dir`/`change_plan_path` consumers) both resolve plans that are local in the
ordinary case; `gather-dispatch-context.sh` reads `change_plan_ref` only and is untouched.

### 2. Guard triage and signal two (`skills/flow/scripts/check-unfinished-work.sh`)

- The verdict triage is unchanged in shape: a satellite `link.md` with `## Part of` that does not
  resolve still refuses with exit 2; a change with no plan anywhere still reports
  `OUTSTANDING: no plan at …`. The header comment's satellite definition widens to cover the
  canonical-repo-only convention, resolved through the supplied canonical worktree by name.
- **Signal two reads the project the plan lives in.** When the resolved plan directory lies under
  the canonical worktree, `flow record findings -change <name> -C <dir>` runs with
  `-C <canonical-worktree>` instead of the satellite worktree. Today the satellite call queries the
  satellite repository's own project key, finds a change the store has never heard of under that
  key, and answers a permissive `[]` — the "no finding is open" half of a CLEAR verdict would rest
  on a query that cannot see the change. With the canonical `-C`, the satellite call reads the same
  findings store the canonical call reads. The `flow record verdict` write at the end stays
  `-C <worktree>` (advisory, worktree-attributed, per the guard's existing header).

### 3. Tests

- `scripts/test-lib-change-plan.sh`: same-name resolution fires for no-`link.md` and
  link-without-`## Part of` fixtures with a canonical worktree; does not fire without the
  argument; does not fire when a `## Part of` link exists but the canonical plan named by the link
  is absent (returns 1 — the loud case, asserted as unresolvable).
- `scripts/test-check-unfinished-work.sh`: guard-level cases — treeless satellite + canonical
  argument → `CLEAR`/`OUTSTANDING` tracking the canonical plan's checkboxes; argument omitted →
  `OUTSTANDING: no plan at`; `## Part of` link with a missing canonical plan → exit 2 (unchanged);
  signal two surfaces findings recorded under the canonical project through the satellite call, by
  asserting which `-C` the stubbed `flow` received.

### 4. Docs

The two script headers only. No flow-contract text changes: `finish-contract-run1.md` already
requires the canonical worktree on every run-1 call, and no project-configuration key is added.

## Rejected alternatives

- **`plan-repo:` line in `.flow/project.md`** (the issue's option 2): a new optional config key
  needs project-configuration contract prose, a body-shape rule and containment for an
  attacker-influenced path — and a static declaration still cannot name the canonical *apply
  worktree*, which is where the plan lives during a run, so it would need worktree resolution
  underneath anyway. The canonical-worktree argument already carries exactly that fact, resolved by
  the caller.
- **Stop at bf684f7**: the link hard-gating is correct but unreachable for a repository with no
  spec tree; every treeless cross-repo change would keep forcing hand verification — the exact
  recurring cost the issue records.

## Testing strategy

Harness-first, per this repository's guard conventions: extend both harnesses with the cases above
(red), implement the branch and the `-C` switch (green), then the full `## test` list at
`flow.verify`. Lint for the touched files: the guard harnesses themselves plus
`scripts/check-vocabulary.sh`, `scripts/check-references.sh` and `scripts/check-contract-budget.sh`
(header growth stays inside the budget ratchet).
