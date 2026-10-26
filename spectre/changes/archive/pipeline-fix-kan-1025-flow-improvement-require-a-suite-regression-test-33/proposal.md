# pipeline-fix-kan-1025-flow-improvement-require-a-suite-regression-test-33

## Why

- kan-1025 (2026-10-27): `flow stage end -command '/flow' -stage flow.review-panel -outcome completed <name>`, run from the main checkout as `skills/flow/review-panel.md` prescribes, falsely refused with "completed reviewer dispatch(es) left no report file" — `gatePanelFindingsClosed` (`stats/cmd/flow/stage.go`) defaults its dir to `"."`, and `check-panel-findings-closed` read the main checkout's `.superpowers/sdd/` instead of the change's worktree. The same command run from the worktree printed `FINDINGS-CLOSED`.

## What changes

- `stats/cmd/flow/stage.go`: with no `-C`, `gatePanelFindingsClosed` resolves the checkout carrying the change through `changeCheckouts` (preferring a worktree other than the main checkout) before running the guard.
- `stats/cmd/flow/stage_test.go`: a regression test running the end without `-C` from the main checkout of a repo whose worktree alone carries the report files.
