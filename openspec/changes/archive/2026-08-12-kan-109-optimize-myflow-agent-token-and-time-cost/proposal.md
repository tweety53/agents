## Why

KAN-109 records a measurement taken on the KAN-71 fix-1 session: ~5.28M subagent tokens for one
change, of which 14 implementer dispatches accounted for 3.05M (58%) and two full seven-slot review
panel passes accounted for a further 1.70M. The waste it isolated is structural, not workload:
parallel implementers colliding on one Gradle build directory (six assertions left red at file
seams, 60s idle waits, corrupted test-result XML forcing suite re-runs), task slicing finer than the
file graph, a 9877-line diff read twice by a full panel, and conditional panel slots re-run against
subjects that did not change (76k tokens to confirm a comment-only backend diff).

The ticket's fifth proposal — a stronger model for `models.panelFix` — is already in force.
`myflow-model-policy` defaults that role to Opus, or the harness's strongest available model,
deliberately and not to Sonnet. Nothing changes for it.

## What Changes

- **One implementer per worktree, serialized.** `/myflow-do` SHALL keep at most one implementer
  subagent in flight against a given worktree at a time. Parallelism *across* worktrees is
  unchanged. This explicitly overrides `superpowers:subagent-driven-development`'s parallel dispatch
  guidance and `superpowers:dispatching-parallel-agents` wherever the tasks share a worktree.
- **Bundle dispatches by shared files.** New `scripts/plan-dispatch-bundles.sh` (thin Bash wrapper)
  over new `scripts/plan-dispatch-bundles.py`, following the split `check-task-build-green.sh` and
  `check-task-commit-fields.sh` already use for `tasks.md` parsing. It groups unchecked tasks whose
  `**Files:**` sets intersect, transitively, and prints one bundle per line. `/myflow-do` §4
  dispatches one implementer per bundle instead of one per checkbox.
- **Panel diff-size cap.** New `scripts/check-panel-diff-size.sh <worktree> <merge-base> [cap]`,
  pure Bash over `git diff --shortstat`, default cap 2000 changed lines. `/myflow-do` §5 runs it
  before writing `final-review.diff`; over the cap the operator gets a named-options prompt, with
  *proceed with the panel anyway* as the default and recommended answer. The measured size is
  recorded in the panel record on every run, over the cap or not.
- **Conditional slots re-run only when their subject changed.** In the panel's **Full** escalation
  mode, required slots always re-run; a conditional slot (Security, Adversarial, Lens B, Lens C)
  re-runs only when its own trigger-table row still fires against `fix-round-N.diff`, and otherwise
  its pass-1 result stands, recorded as `not re-run — subject unchanged`. This is a definition of
  what non-stale means for a slot with a bounded subject, not a waiver of the bar.
- **Collateral:** `scripts/check-contract-budget.sh`'s `skills/myflow-do/SKILL.md` row is re-anchored
  after the content tasks land; `.myflow/project.md`'s `## test` list gains the two new harnesses.
  Neither new guard joins `## lint` — both need a change in flight, exactly the reason
  `check-finish-preflight.sh` and `check-unfinished-work.sh` are already excluded.

## Capabilities

### New Capabilities

- `myflow-dispatch-economy`: how `/myflow-do` groups and paces implementer dispatches — one
  implementer per worktree at a time, and one dispatch per file-overlap bundle rather than per plan
  task. Both are dispatch properties rather than review-panel properties, which is why they do not
  join `myflow-review-panel-economics`.

### Modified Capabilities

- `myflow-review-panel-economics`: gains **Requirement: The panel measures the diff it is about to
  read** (the cap, the prompt, and the recording of the measured size) and **Requirement: A
  conditional slot re-runs only when its own subject changed** (the scoping of Full-mode escalation,
  and the accompanying definition of non-stale for a slot with a bounded subject). The existing
  zero-open-findings bar, the marker-line rules, the findings-table rules and the roster are
  untouched.

## Impact

- Files: `skills/myflow-do/SKILL.md` (§4 and §5), new `scripts/plan-dispatch-bundles.sh` +
  `scripts/plan-dispatch-bundles.py` + `scripts/test-plan-dispatch-bundles.sh`, new
  `scripts/check-panel-diff-size.sh` + `scripts/test-check-panel-diff-size.sh`,
  `scripts/check-contract-budget.sh`'s table, `.myflow/project.md`'s `## test` list.
- No change to the state file shape, the three-state machine, the review panel roster presets, the
  optional-slot trigger table itself, the escalation ladder's trigger conditions, the panel record
  format, the marker-line rules, or any `/myflow-start` / `/myflow-finish` / `/myflow-status`
  behavior. `/myflow-fast` inherits everything here through the `/myflow-do` sections it cites, with
  no edit of its own.
- Design: `docs/superpowers/specs/2026-08-12-kan-109-optimize-myflow-agent-token-and-time-cost-design.md`.
