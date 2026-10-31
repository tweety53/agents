# pipeline-fix-kan-1051-rework-the-trend-dashboard-51 — design

## Context

Evidence: kan-1051's close of tasks 6 and 7 (commits `6745c964`, `54e12449`) printed
`close-task: task 6: …/stats/web/tests/visual is not the commit's — cannot rebuild stats/web/tests/visual/full-app-suite.zip to compare`
and exited 2. Reproduced in a fixture: task 1 commits a script-built zip over `a.png`, task 2
re-records `a.png` and commits the rebuilt zip; `close-task <wt> demo <base> 1:<s1> 2:<s2>` exits 2
with that line for task 1.
<!-- measured: go test ./internal/guard -run TestCloseTaskJudgesTheLastSuiteZipOfAClose (cd stats) with closetask.go unpatched @ 50e0a1f7 → FAIL, "superseded script-built zip: exit 2, want 0" -->

Cause: `ctCheckSuiteZips` rebuilds every zip-changing commit's zip from the worktree, which holds
only the close's last screenshots root; the root check refuses every earlier commit.

The diff in Task 1 was applied to this worktree at `50e0a1f7` and checked there: every
`TestCloseTask*` passes, the new test fails with `closetask.go` reverted, and
`run-verify-list.sh <worktree> lint` exits clean but for `check-plan-shape.sh` on this then-empty plan.

Blast radius: `ctCheckSuiteZips` is cited in one file.
<!-- measured: git grep -l ctCheckSuiteZips | wc -l @ 50e0a1f7 → 1 -->
The fix touches `stats/internal/guard/closetask.go`, its test and the `scripts/close-task.sh`
header. Nothing the defect wrote needs repair: it only refused, and kan-1051's tasks are re-closed
by that run once this lands.

## Recommended approach

The `judge-last-zip-of-close` decision below: `closeTask` passes each task's peers' commit fields to
`ctCheckSuiteZips`, which skips a zip path a descendant commit of the close, in the same worktree,
changes too.

## Critical files

- `stats/internal/guard/closetask.go` — `closeTask`'s guard loop and `ctCheckSuiteZips`.
- `stats/internal/guard/close_task_test.go` — the new test beside `TestCloseTaskRefusesAHandBuiltSuiteZip`.
- `scripts/close-task.sh` — the header contract's zip paragraph.

## Existing code to reuse

`ctFixture`, `fx.run`, `gitRun`, `tcfGit` and the `rebuild-full-app-suite-zip` registry entry in the
test; `envGit`, `capture` and the existing map-pair split in the guard.

## Verification

Task 1's Step 3: `TestCloseTask*` pass, the new test fails without the guard change, and the Go lint lines exit clean.

## Decisions

### How a close judges two commits that change one zip

**ID:** judge-last-zip-of-close
**Status:** active
**Chosen:** skip a zip path in a commit when another commit of the same close, in the same worktree, descends from it (`git merge-base --is-ancestor`) and changes that path too; the last such commit is judged unchanged, against the worktree. An earlier commit's zip is superseded before it lands, so what lands is what is judged, and a hand-built last zip is still exit 1.
**Considered:** rebuilding each commit's zip in a temporary checkout of its own tree — the resolver reads the canonical worktree's `.flow/project.md` and the zip may sit in the regression checkout, another repository, and `rebuild-full-app-suite-zip` refuses a root its resolved PNGs do not lie under, so a temporary checkout needs a second resolution path; restoring the root to each commit's tree in place, rebuilding, then restoring — mutates the implementer's worktree, and a failed restore loses state; judging only the close's last commit for every guard — the other guards are per-commit and correct as they are.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — the change narrows which commits a close-time guard judges, and a reviewer must confirm it cannot skip the commit that lands.
**Considered:** `flow-low` — the diff is small, but a wrong skip would silently pass a hand-built zip.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

None.
