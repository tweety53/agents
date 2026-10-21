# kan-887-flow-fix-delete-the-retained-bash-twins-and — session narrative

## 2026-10-21 — creating run

- Inline execution, ten tasks in plan order. The harness audits (tasks 2, 6–10) were done by running each bash harness and diffing its `ok:` labels against the Go subtests, then reading the uncovered cases. Gaps found and closed in Go before deleting: aside's tree/stash state (the parity test compared output only), done-when's empty-worktree usage case, hand-notes' edit above the managed block, project-get's HEAD-duplicate and ambient-git-env cases, fast-route's stderr causes.
- `change-plan.sh`'s 109-line header was canonical for the resolution order; deleting it without a home would have dropped normative text, so it moved into `changeplan.go`'s header.
- Two harnesses were red on the base, unrelated to this change: `test-flow-addr-declaration.sh` missed stage.go's `fset.String` -addr registration; `test-git-config-pins.sh` left a Go composite literal's `}` on a flag token. Fixed as full-suite-fix-1.
- A `**Tests:**` value `TestCheckDoneWhenPaths/case_13b` was refused by the fields guard (subtest paths are not parsed); `Case 13b` then failed `check-plan-shape.sh` (a `Case N` label is numeric); settled on `TestCheckDoneWhenPaths`.
- The pipeline.md sentence edit needed a `verbatim-moves.txt` acknowledgement.
- A panel re-run row closed `-outcome clean` did not satisfy `check-panel-findings-closed.sh`, which reads only `completed`; re-closed.

## 2026-10-21 — integrate run

- Preflight clear; unfinished-work gate clear (every plan item checked, no open finding); no UI paths; origin/main had not moved since the merge base, so no rebase. Route: merge and push, the project's configured default.
