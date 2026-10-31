# pipeline-fix-kan-1051-rework-the-trend-dashboard-51

## Why

`close-task.sh` (`stats/internal/guard/closetask.go`, `ctCheckSuiteZips`) judges every task commit
that changes `full-app-suite.zip` by rebuilding the zip in the worktree, and first requires the
worktree's screenshots root to equal that commit's tree (`git diff --quiet <sha> -- <root>`). When
one close carries two such commits — an implementer group whose later task re-records screenshots
and rewrites the zip — the earlier commit's root is no longer the worktree's, so the close exits 2
and nothing is ticked. kan-1051 tasks 6 (`6745c964`, writes the zip) and 7 (`54e12449`, re-records
10 baselines and rewrites the zip) closed together and stopped on
`close-task: task 6: …/stats/web/tests/visual is not the commit's — cannot rebuild stats/web/tests/visual/full-app-suite.zip to compare`.
Passing task 6 as a `<detached-worktree>=<sha>` map pair is no way round it: the map is refused for
leaving the recorded worktree out, and `rebuild-full-app-suite-zip` refuses the detached root.

## What changes

- `ctCheckSuiteZips` skips a zip path that a later commit of the same close, in the same
  worktree, changes again; that later commit's zip is judged as before. A close's last zip is
  therefore still rebuilt and compared, so a hand-built zip is still refused (kan-1016).
- The `scripts/close-task.sh` header states the exception.
- One test, `TestCloseTaskJudgesTheLastSuiteZipOfAClose`, exits 2 without the fix.
