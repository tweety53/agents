# Self-review report for kan-991-panel-effort-roll-high-gated

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The plan cited the retired `scripts/check-vocabulary.sh` and passed every section-D guard; check-plan-shape now runs a Go guard refusing a cited scripts path the tree lacks — fixed: 15bcf3af
- **[flow-fix]** check-panel-docs-only refused the docs-only reduction on the change's `verbatim-moves.txt`; that file under `<spec-root>/changes/` now counts as documentation — fixed: 772269c9
- **[flow-fix]** The gated per-task reviewer bundle was allowed `high` through the plan gate, but the decision records no pair for it; the bundle now never takes `high` and the gate asks only of panel.dispatches entries — fixed: e0ff7372
- **[flow-fix]** `flow record pass` accepted the one-character note `x` (a silenced placeholder loop wrote two such rows); it now refuses a note of fewer than two words — fixed: ebfb5271
- **[flow-fix]** The plan's Files: omitted SKILL-rationale.md, whose reasoning the change removed; the planner now lists the rationale appendix a changed rule's reasoning lives in — fixed: 852125f7
- **[flow-fix]** The ledger's opus effort=high implementer and verifier rows are the parent's own inline pair, not a high implementer decision — noted
- **[flow-fix]** The landed implementation subject omits the display-formatting bullet; the commit-split subject is written by hand and no guard can judge it — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The parent ran inline at effort high on a plan of exact string substitutions; the parent's effort is set at session launch — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Two panel roles independently caught the stale rationale — noted
- **[flow-improvement]** check-verbatim-moves made every reworded sentence an explicit acknowledgement — noted
- **[flow-improvement]** The missing guard was recorded as a substitution rather than skipped silently — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

_none — this angle produced no findings._

## What would make a good flow-eval case — `flow-eval`

_none — this angle produced no findings._
