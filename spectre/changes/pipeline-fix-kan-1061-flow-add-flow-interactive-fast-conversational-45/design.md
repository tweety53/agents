## Context

Found by the in-run planner of `kan-1061-flow-add-flow-interactive-fast-conversational` under open
experiment 1 (`plan-designer-xhigh-explorer-low`): `plan-agent-pair.sh <key> <n>` printed
`opus xhigh` for an Explore dispatch, and `hooks/enforce-agent-baseline.py` refused it with
`DISPATCH-PAIR-REFUSED: explorer runs on sonnet/Explore@low`.

Cause: `planAgentPair` reads `rows["plan-designer"]` unconditionally. The hook judges a dispatch by
its prompt's `Role:` line against that role's row of `policyTable` — the same table, overrides
overlaid — so the two agree only while the `explorer` and `plan-designer` rows share their first
pair's model and effort, which the base table happens to (`opus/Explore@low`, `opus/Plan@low`) and
experiment 1 does not.

The diff in Task 1 of `tasks.md` was applied to this worktree at `e6c5c19e` and checked there:
`scripts/plan-agent-pair.sh x 1 explorer` prints `sonnet low` and `... plan-designer` prints
`opus xhigh` against the live store with experiment 1 open; `x 1` alone exits 2 with the new usage
line; `go test ./internal/guard` passes; `TestPlanAgentPairReadsOverride` fails with
`planagentpair.go` reverted; `check-references.sh`, `check-dispatch-paragraphs.sh` and
`check-guard-symlinks.sh` exit 0.
<!-- measured: scripts/plan-agent-pair.sh x 1 explorer|plan-designer in this worktree @ e6c5c19e with the Task 1 diff applied → sonnet low / opus xhigh -->

Blast radius: 35 tracked files cite `plan-agent-pair` (`git grep -l plan-agent-pair`), 23 of them
archived changes or self-review reports that record history and stay as written; of the 12 live
files the fix touches 7 — the shim, the Go port, its test and the four contracts citing the usage.
`model-policy.md`, `SKILL-rationale.md`, `dispatchparagraphs.go`, `check_dispatch_paragraphs_test.go`,
`modelpolicy.go` and `planclass.go` name the script without its arguments and stay unchanged.
<!-- measured: git grep -l plan-agent-pair | wc -l @ e6c5c19e → 35 -->

Nothing to repair: the refused dispatch was denied before any row was recorded. No saved memory
encodes this defect.

## Decisions

### How the script learns the dispatch's row

**ID:** role-argument
**Status:** active
**Chosen:** a required third argument, `<role>`, `explorer` or `plan-designer` — the dispatch's own `Role:` line — reading that row of `policyTable`, overrides overlaid, the row the hook admits on.
**Considered:** printing both roles' pairs on two lines — every caller must then pick a line, the same knowledge passed as parsing; the role as an optional argument defaulting to `plan-designer` — keeps the silent wrong answer for any caller not updated; the role first — every argument position moves for no gain over appending it.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one row lookup made role-keyed, its test and four one-line contract edits.
**Considered:** `flow-medium` — the script feeds every planning dispatch, but the change is mechanical and its test pins both roles under a live-shaped override.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions
