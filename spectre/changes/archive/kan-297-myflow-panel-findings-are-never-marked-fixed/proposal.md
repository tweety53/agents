# kan-297-myflow-panel-findings-are-never-marked-fixed

**Jira:** [KAN-297](https://tweety53.atlassian.net/browse/KAN-297) — the fix round never closes
the findings it repaired.

## Why

A review panel finding is written to the store with `-status open`. The fix round repairs it, the
parent verifies the repair, and the panel re-run confirms clean — and nothing ever moves the
finding off `open`. `check-unfinished-work.sh`'s signal two reads the store at integrate time and
reports every one of them as outstanding.

Observed on KAN-295: 30 findings, all fixed and confirmed by three clean panel slots, blocked
`/flow`'s run 1 with `OUTSTANDING: … 30 open finding(s) in the review panel record`. They had to be
closed by hand, one `flow record status` call each, before integration could proceed.

The gap is a prose one. `skills/flow/review-panel.md` shows `flow record status … -status fixed` once,
as a sample in the *recording* section, and states the obligation once more at line 277 — "Fix
subagents record `fixed` when they fixed it" — inside a paragraph about marker-block format, assigned
to the subagent, and absent from the fix subagent's own dispatch prompt. The fix round's procedure
never obliges anyone to close anything, and nothing checks that it happened until integrate.

**The ticket's second defect is already fixed and is out of scope.** `flow record status` no longer
accepts a status outside its vocabulary: `validateFindingStatus`
(`stats/cmd/flow/record.go:45`) rejects anything but `open`, `fixed` and `withdrawn <reason>` before
the store is contacted, naming the accepted set. `check-unfinished-work.sh`'s own header cites that
validation as done.

## What changes

- `skills/flow/review-panel.md`'s fix round records `fixed` on each finding at the point it already
  verifies the repair, per finding, by the parent.
- A new shipped guard, `scripts/check-panel-findings-closed.sh`, refuses to let the review-panel
  stage end while the store still holds an open finding — with its own mutation-tested harness.
- The stale passage describing `check-unfinished-work.sh` as a marker-block parser is corrected.

No bulk close operation is added; see `design.md`'s `no-bulk-close` decision.
