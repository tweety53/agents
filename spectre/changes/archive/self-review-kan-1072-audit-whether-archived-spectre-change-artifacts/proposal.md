# self-review-kan-1072-audit-whether-archived-spectre-change-artifacts

## Why

The self-review pass of `kan-1072-audit-whether-archived-spectre-change-artifacts` found four
fixable defects (evidence: `findings.md` under the pass's `<sr-dir>`,
`/Users/tweety53/Agents/flow/state/agents-a740d89c/self-review/kan-1072-audit-whether-archived-spectre-change-artifacts/`):

- F1: `commit-split` scopes the planning commit to the whole `spectre/changes/` tree, not the
  change folder, so integrate's reshape folded the tip's archive deletion into each repository's
  `chore(spectre): plan …` commit (agents: thousands of deleted files under one plan subject), and
  re-added every `archive/*/link.md` deletion to the implementation commit. The six branches are
  already pushed in that shape.
- F2: `skills/flow/implement.md` names a group implementer's key `task-<n>-implementer`, while the
  hook admits only the `+`-joined `task-<ids>-implementer`; the run's first launch was refused.
- F3: `flow self-review bundle` waits only the store-write fallback deadline for a bundle the
  daemon assembles from git history, and has no fallback; the first call timed out.
- F4: `check-plan-shape` does not compare the paths a task's steps edit with its
  `**Files:**`/`**Allowed-collateral:**`; every implementation group of the run needed `**Files:**`
  corrections at close.

Planning found a fifth: on this branch `check-verbatim-moves` no longer sees kan-1072's own
acknowledgements once the archive commit moved them, so the loop's verify is red before any fix.

## What changes

- The kan-1072 branches are re-split locally: the archive deletion is one
  `chore(spectre): remove the change archive` commit and the plan commit carries only
  `spectre/changes/kan-1072-…/`; spectre's implementation subject is the skill re-copy alone.
- `check-verbatim-moves` reads the acknowledgements of a change directory the branch itself
  removed, from the commit before the removal.
- `commit-split` leaves only the change folder and its `-fix-<N>` sub-changes to the planning
  commit; every other `spectre/changes/` path lands in the implementation commit.
- `implement.md` writes the group implementer's key and `-task` as `<ids>`, `+`-joined.
- `flow self-review bundle` has its own, longer default deadline.
- `check-plan-shape` runs a new sub-guard, `plan-step-files`, refusing a path an edit step names in
  an edit position that the task's `**Files:**`/`**Allowed-collateral:**` does not cover.
