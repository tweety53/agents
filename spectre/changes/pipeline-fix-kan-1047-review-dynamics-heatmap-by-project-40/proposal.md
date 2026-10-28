# pipeline-fix-kan-1047-review-dynamics-heatmap-by-project-40

## Why

A fixed Critical or Important finding of panel round N is proved only by a `flow record mutation`
row at round N or N+1 (`uwUnprovedFixRounds`, `stats/internal/guard/unfinishedwork.go`, shared by
`check-panel-findings-closed` and `check-unfinished-work`). `flow record mutation -round N+1`
refuses once round N+1's re-review dispatch has begun (`stats/cmd/flow/record.go`). A parent that
missed recording the fix-mutation rows before launching the re-review can therefore never satisfy
the close guard: `flow stage end -stage flow.review-panel` refuses forever, and integrate's landing
gate too. Observed in kan-1047: findings F1–F3 of round 0, fixed in round 1; the re-review
`panel-1-test-sanity+mutation` began 2026-10-28T17:34:51Z, rows recorded afterwards were refused at
round 1, and rows at round 2 are accepted but do not prove round 0.

## What changes

- `flow record mutation` gains `-late '<reason>'`: it admits a row for a round whose re-review
  already began, and records the reason beside it as a pass note of that round,
  `late mutation proof for <path>: <reason>`, so the rendered pass log shows the proof came late
  and why. Without `-late` the ordering refusal is unchanged.
- The close guards are unchanged: they count round N/N+1 rows, and a late row is one.
- `skills/flow/review-panel-fix-round.md` names the late form as the recovery for a missed row.
