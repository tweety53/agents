# pipeline-fix-kan-1051-rework-the-trend-dashboard-53 — design

## Context

Evidence: on kan-1051's branch `git show dffb70df:./spectre/changes/kan-1051-rework-the-trend-dashboard/tasks.md`
exits 128; `dffb70df^` holds the plan with Tasks 1–9 ticked; `integrate-change.sh prepare` stops on
`STOP: unfinished-work … (Task 8, 9)`.
<!-- measured: go test ./internal/guard -run TestCheckUnfinishedWork (cd stats) with unfinishedwork.go unpatched @ df1564a9 → FAIL on the new drop-removes-the-plan case -->

Cause: the drop commit is the range base and no longer holds the plan, so the base read fails silently.

Blast radius: 2 files — `stats/internal/guard/unfinishedwork.go` and its test. Nothing the defect
<!-- measured: git diff --stat (unfinishedwork.go fix applied) @ df1564a9 → 2 files -->
wrote needs repair: it only refused, and kan-1051's prepare is re-run once this lands.

## Recommended approach

The `read-plan-at-drop-parent` decision below.

## Critical files

- `stats/internal/guard/unfinishedwork.go` — signal four's commit half.
- `stats/internal/guard/check_unfinished_work_test.go` — the new case beside the existing dropped-base case.

## Existing code to reuse

`capture`, `envGit`, `uwAppendedIn`; in the test `newUW`, `gitRun`, `tcfGit`, `covered`, `uwVerdict`.

## Verification

Task 1's Step 3.

## Decisions

### Where a dropped re-run's earlier plan is read

**ID:** read-plan-at-drop-parent
**Status:** active
**Chosen:** read the plan at the range base; when that `show` fails, read it at `<base>^`. A base that still holds the plan reads as before; a drop commit that removed it falls back to the plan as it stood just before the drop.
**Considered:** changing `archivedBase` to return the drop's parent — that widens the range to the drop commit itself for every caller of the range, not only this read.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — mechanical: one fallback read and one test case.
**Considered:** `flow-medium` — nothing in the diff needs a judgement the plan leaves open.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

None.
