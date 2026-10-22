# self-review-kan-922-flow-improvement-port-the-bash-python-tooling

## Why

kan-922's self-review pass found four defects, none `big`:

- **F-A (flow-fix, operator-raised).** `integrate-change.sh prepare` stops at
  `STOP: foreign-staged` on any `DRIFT-BRANCH`, even when the main checkout is otherwise clean
  (`STAGED-CLEAN`, no `DRIFT-DIRTY`). kan-922's integrate run stopped on
  `DRIFT-BRANCH: /Users/tweety53/Projects/agents — on fix/base-missing-on-satellite, not main`
  with nothing staged or dirty; the operator answered "resolve it yourself, update flow to
  resolve such things automatically".
- **F-B (flow-fix).** `land-self-review-report` never runs `check-self-review-report`: KAN-923's
  report reached `main` (aa8c7bc1, e111a883) with two violations, and kan-922's pass-1 full suite
  failed `TestCheckSelfReviewReport` on the base.
- **F-C (flow-fix).** An inline `flow record dispatch begin` checks the session's cwd for tracked
  modifications (`inlineDirtyPaths(f.dir, …)`, `stats/cmd/flow/record.go`), so another session's
  edits in the main checkout refused task 2's begin while the change's worktree was clean.
- **F-D (flow-fix).** The panel fixer brief (`skills/flow/review-panel.md`) describes
  `break-and-prove.sh` without saying `<test-command>` runs from the repository root; the fixer's
  `go test` failed with `go.mod not found` on both legs.

## What changes

- `prepare` switches a clean main checkout on another branch back to its default branch, prints
  one `DRIFT-SWITCHED` line and re-reads its drift; a detached HEAD, a branch the change records
  as its base, a refused switch and a re-read that is not `DRIFT-CLEAN` still stop and ask.
- `land-self-review-report` runs the repository's `check-self-review-report.sh` over a report
  before staging it and refuses a violation (`LAND-REPORT-INVALID`) or a check that cannot answer
  (`LAND-REPORT-UNCHECKED`), exit 5.
- The inline begin's dirt check reads the change's own worktree.
- The fixer brief states where `break-and-prove.sh` runs its test command.
