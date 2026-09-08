## Context

`/flow`'s implementation stage dispatches one implementer subagent per dispatch bundle
(`plan-dispatch-bundles.py` groups tasks by declared `**Files:**` overlap and `**Squash-with:**`
partners), but `skills/flow/implement.md` §4 allows at most one implementer in flight per worktree,
so bundles run serially even when file-disjoint. The research note
(`docs/superpowers/research/flow-speedup.md`, sections 4b, 5 and 8) measured the serialization and
the operator reopened its wall-clock lever 1 as this change. Constraints:

- The plan grammar is a Single Source of Truth: every structural pattern lives in
  `scripts/lib/plan_grammar.py` and every guard imports it (the module docstring's F9/F19/F20
  history is why a second regex copy is a defect, not a shortcut).
- Guard scripts are Python 3, standard library only, one file per invocation.
- Backward compatibility is structural, not conventional: a plan carrying no `**After:**` field
  must dispatch bit-for-bit as today.
- `check-contract-budget.sh` ratchets every owned `.md`; `check-normative-inventory.sh` diffs are
  resolved by restoring the sentence; `skills/flow/implement.md` carries required dispatch
  paragraphs (`check-dispatch-paragraphs.sh`) that prose edits must keep intact.
- The KAN-366 collision rule bans two agents mutating one worktree; waves sidestep it by giving
  each concurrent implementer its own worktree, which the existing sentence already permits
  ("dispatches into different worktrees remain free to run concurrently").
- A fresh worktree in this repository cannot run `go test` until the SPA is built
  (`## worktree setup`: `cd stats && make web-build`), so a throwaway wave worktree must run that
  setup once before its implementer dispatches.

## Decisions

### Opt-in parallelism via an `**After:**` task field, absent = serial

**ID:** after-field-opt-in-serial-default
**Status:** active
**Chosen:** a `**After:** Task <ids>` / `**After:** none` field, the line-scoped shape of
`**Squash-with:**`; a bundle dispatches when every id in its resolved after-set has landed —
because `**Files:**`-disjointness cannot see semantic order (task 3 calling task 1's helper adds
a file), so parallelism needs a declared independence, and defaulting to serial keeps every
existing plan and every unannotated task exactly as today.
**Considered:** deriving parallelism from file-disjointness alone — rejected: two file-disjoint
tasks may still depend on each other's symbols; forcing the planner to renumber or merge would
lose real waves; making parallelism the default — rejected: silently changes every existing plan's
dispatch semantics, which the issue promises not to do.

### A bundle alone in its wave runs in the canonical worktree and commits directly

**ID:** singleton-wave-runs-in-canonical-worktree
**Status:** active
**Chosen:** no throwaway worktree and no cherry-pick for a wave of one — because a no-field plan
then produces only singleton waves and its dispatch mechanics are byte-identical to today, making
the backward-compatibility invariant structural rather than a promise.
**Considered:** always throwaway + cherry-pick, even for a wave of one — rejected: one uniform
code path, but every existing plan gains the pick machinery and its failure surface, which is the
behavior change the issue forswears.

### The conductor runs the full `## test` suite after a shared wave's final pick

**ID:** conductor-runs-suite-after-final-wave-picks
**Status:** active
**Chosen:** when the plan-last bundle belongs to a shared wave, the conductor runs the resolved
`## test` list once on the canonical worktree after that wave's final pick passes
`check-task-commit-fields.sh`; a failure is the same verbatim-output `## Question` handback as
today — because it keeps the panel-on-green guarantee at the same serial wall clock as today's
last-bundle run (a singleton plan-last bundle keeps implementer-carried `FULL SUITE` unchanged).
**Considered:** the plan-last implementer keeps carrying the suite in its own throwaway worktree —
rejected: the run then misses same-wave siblings' commits, so the panel can see a red branch;
dropping the in-implementation suite entirely — rejected: abandons the guarantee on every change
to save 3–4 minutes the wave gain dwarfs.
<!-- measured: the ## test list floor, docs/superpowers/research/flow-speedup.md section 8 @ branch main -->

### The presented design stands (merged convergence confirm, round 2)

**ID:** design-approval
**Status:** active
**Chosen:** the operator answered "approve" explicitly after the round-1 answers
(conductor-run suite; canonical-worktree singletons), closing the merged convergence confirm and
the design-approval HARD GATE in one answer.
**Considered:** another round or a revision — neither was raised; the two prior runs' unanswered
asks are superseded by this explicit answer.

## Open questions

### Design approval — does the presented design stand as summarized (approve and move on)?

**ID:** design-approval-confirmation
**Status:** answered by design-approval
**Why it is open:** the merged convergence-and-approval confirm went unanswered — the ask returned empty three times across two runs (2026-09-07/08); the design gate closes only on an explicit operator answer
**What it affects:** whether the run proceeds to flow.create-artifacts (change worktree, enriched design.md/tasks.md) and implementation; the design awaiting approval: `**After:**` grammar in `scripts/lib/plan_grammar.py`, `after <k>:` bundle output in `plan-dispatch-bundles.py`, shape findings F7–F10 in `check-plan-shape.sh`, wave dispatch in `skills/flow/implement.md` §4 (canonical-worktree singletons committing directly, throwaway worktrees + plan-order cherry-picks for shared waves, conductor-run full `## test` suite after a shared wave's final pick), plan-template field in `skills/flow/brainstorm-planner.md`, harness cases; round-1 answers hold: conductor runs the suite after the final wave, singleton waves run in the canonical worktree
