# kan-1025-flow-improvement-require-a-suite-regression-test — session narrative

## 2026-10-27 — creating run

- Resumed at STARTED with a ready 2-task plan; inline execution (decision class small).
- Task 1: adding the regression-test check broke four existing fixtures with fixed Importants and no fix report; they now carry a `none` line. Task 2's prose first failed `check-verbatim-moves` (acknowledged in `verbatim-moves.txt`) and `check-installed-citations` (bare `.superpowers/` → `<project>/.superpowers/`).
- Panel round 0: the `failure-modes+mutation` slot ran ~45 min without a report and was stopped, closed `timed-out` and re-dispatched once with a wall-clock budget; root cause was its mutation reproducers running the whole guard package (~455 s) under a 300 s bound. The parent narrowed them to `-run TestCheckPanelFindingsClosed` and swapped their exits to the `# mutation-reproducer` convention (they had generic exits) instead of bouncing — the defect logic was unchanged; recorded in the pass log.
- Round 0 raised F1–F4 Important (pathspec match accepted a directory/`.`, untested recorded-worktree and state-failure branches, fixed Critical never exercised, `.superpowers/` refusal untested in effect) and F5 Minor (bare `none` accepted). Fixed inline in one commit on top; six break-and-prove flips held; re-run clean. The new guard closed the panel off this change's own `regression-test:` lines.
- `flow stage end … flow.review-panel` without `-C` falsely refused from the main checkout (its in-process guard read the main checkout's `.superpowers/sdd/`); closed from the worktree and fixed in-run as pipeline-fix 33.
