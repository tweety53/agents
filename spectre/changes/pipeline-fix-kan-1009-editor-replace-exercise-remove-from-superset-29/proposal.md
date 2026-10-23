# pipeline-fix-kan-1009-editor-replace-exercise-remove-from-superset-29

## Why

`GET /api/v1/self-review/{project}/{change}/bundle` runs some thirty git calls in series, each
bounded by `DefaultGitBound` (10s) but none bounded together. On a loaded machine each call took
a second or two, the handler ran 40–55s, and the server's 30s `writeTimeout` cut the connection
with an empty reply: `flow self-review bundle` failed with "context deadline exceeded" for
`kan-1009-editor-replace-exercise-remove-from-superset` and `kan-1026-stored-data-first-on-every-open`.

## What changes

- `selfreview.ExecRunner` gains a `Deadline`: a call starting past it fails without running git,
  and a running call is killed at it.
- The bundle handler gives the real runner a deadline `selfreview.BundleBudget` (15s) out, so a
  slow repository degrades the bundle (its sources reported skipped) instead of outrunning
  `writeTimeout`.
- `TestBundleGitBudgetCapsSlowRepository` drives a sleeping git shim through `Bundle`;
  `TestGitBoundInsideWriteBudget` now pins `BundleBudget + DefaultGitBound` inside 30s.
