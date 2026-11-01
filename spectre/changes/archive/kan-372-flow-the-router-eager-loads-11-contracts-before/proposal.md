# kan-372-flow-the-router-eager-loads-11-contracts-before

## Why

`skills/flow/SKILL.md` instructs every `/flow` invocation to eager-load `pipeline.md` plus 10
further contracts before state is even resolved — 62-68% of a run's skill text, paid before a
single line of project code, diff or artifact is read. Measured at `18a89b5`: ~45k tokens of
eager router load, ~73k for a full creating run, ~67k for a full integrate run. A second,
independent cost: every dispatched subagent re-reads `rules/agent-baseline.md` + `CLAUDE.md` +
`<project>/.flow/project.md` (~10k tokens) at full price in its own separate prompt cache — a
five-slot review panel spends ~50k tokens per round on duplicated baseline alone.

KAN-295 (Done) applied this treatment to `pipeline.md` alone. This generalises that result to the
router's load list and to the remaining large contracts.

## What changes

- `skills/flow/SKILL.md`'s eager-load block is removed; each of the 10 non-`pipeline.md` contracts
  is loaded only by the phase file(s) that actually need it, at the point they need it.
- `skills/flow-contracts/finish-contract.md` splits into `finish-contract-run1.md` (preflight
  signals, Run 1, Resolving a change's worktrees) and `finish-contract-run2.md` (Run 2, Worktree
  cleanup) — loaded only by `integrate.md` and `archive.md` respectively.
- `skills/flow-contracts/project-configuration.md` splits into itself (resolution rules a run
  applies) and `project-configuration-authoring.md` (guidance for editing a project's own
  `.flow/project.md`).
- `gather-dispatch-context.sh` gains a `## lint`/`## test`/`## run` extract in the dispatch
  bundle, and the CONTEXT BUNDLE dispatch paragraph tells subagents those commands are already
  provided, so `CLAUDE.md`'s lint-fix-priority rule no longer forces a full `.flow/project.md`
  re-read per dispatch.
- `jira-followups.md`, `review-panel.md`, `pipeline.md`, and `state-file.md` are audited for
  passages restating content canonical elsewhere and cut per the repository's cut-never-paraphrase
  rule.
- A creating run and an integrate run each load a measurably smaller set; the measurement in
  KAN-372's description is re-run and before/after totals recorded; every `scripts/check-*.sh`
  guard stays green.

## Measurement

Re-run at the tip of this change (after Tasks 1-8), same methodology as the ticket's own
measurement at `18a89b5` — sum `wc -c` bytes over each load set, divide by 4 for an approximate
token count.

| Load set | bytes (before) | ~tokens (before) | bytes (after) | ~tokens (after) | reduction |
|---|---|---|---|---|---|
| Router eager-load (SKILL.md + contracts) | 183,309 | ~45k | 43,363 | ~11k | 76.3% |
| Total, creating run | 295,820 | ~73k | 128,130 | ~32k | 56.7% |
| Total, integrate run | 269,536 | ~67k | 101,721 | ~25k | 62.3% |

"Router eager-load" after is `skills/flow/SKILL.md` + `skills/flow-contracts/pipeline.md` — the
only contract SKILL.md's preamble still eagerly names post-Task-1. "Total, creating run" after is
SKILL.md + brainstorm.md + implement.md + review-panel.md + verify-and-handoff.md +
engineering-principles.md + principles-reviewer-prompt.md. "Total, integrate run" after is
SKILL.md + integrate.md + archive.md + finish-contract-run1.md + finish-contract-run2.md +
operator-prompts.md — the split `finish-contract.md` halves both counted, since a full
integrate-to-archive invocation touches both.
