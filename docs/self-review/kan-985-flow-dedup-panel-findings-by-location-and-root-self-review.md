# kan-985-flow-dedup-panel-findings-by-location-and-root — self-review

**Deferred:** reasoning pass run on opus from the in-run bundle
**Rating:** 4 — good

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The cross-role fold rule left linked rows `fixed` with only the first role re-run, which check-panel-findings-closed failed; the guard now accepts a same-round supersedes target's re-run — fixed: 93896288
- **[flow-fix]** The re-run clause credited any `-supersedes` row, so a cross-round re-raise skipped its own re-run; fold and clause limited to the same round — fixed: 86ada3bb
- **[flow-fix]** "Recorded in the same call" was impossible with a one-ref `flow record status`; reworded to one call per linked ref — fixed: c2fcd198
- **[flow-fix]** The flow-fast writing-plans step showed no literal task, and an indented draft was rejected by check-plan-shape — fixed: 39e55b9d
- **[flow-fix]** check-references and check-verbatim-moves rejections were the guards working as designed, each resolved in one retry — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The declared worktree setup runs make web-build on every run, including docs-only changes that need no SPA build — filed: KAN-987

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Reusing the existing `-supersedes` lineage instead of adding a flag kept the change to prose plus one guard condition — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** Passing harness and session token through a shell variable failed; not a defect, the CLI requires the literal token in the transcript by design — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

_none — this angle produced no findings._

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — the panel missed a docs rule contradicting a Go guard; operator chose to mine nothing — noted
  - severity: Important; location: `skills/flow/review-panel.md:387` at cdc4729f
  - sentence: linked rows take fixed while only the first role re-runs; the guard requires each raising slot's own re-run
  - fix: 93896288 fix(guard): cover a same-round superseding finding by its target's re-run
  - why: a cross-file contract contradiction invisible to a diff-only review
