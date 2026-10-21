# self-review-kan-887-flow-fix-delete-the-retained-bash-twins-and

## Why

kan-887's self-review pass found four pipeline defects, none `big`:

- **F-A (flow-fix).** `check-panel-findings-closed` counts a panel slot's re-run only when its
  dispatch closed `completed`, yet `flow record dispatch end` also accepts `clean`, and the dev
  store holds panel reviewer rows closed `clean`. Such a re-run is invisible, and the guard calls
  its finding fixed without verification.
- **F-B (flow-fix).** A `Case <N>` label in a `**Tests:**` field cannot carry a letter suffix:
  `Case 13b` parses as no label, though Go tests in this repository name cases `case 13b:`.
- **F-C (flow-fix).** `pipeline-fix verify` never runs the harnesses that audit the Go tree as a
  whole (`scripts/test-flow-addr-declaration.sh`, `scripts/test-git-config-pins.sh`) when a Go
  file is touched; main landed red twice that way.
- **F-D (flow-speed).** `flow.verify` re-runs the whole `## test` list minutes after a green
  full-suite run, even when the delta since that run touches nothing a command reads.

## What changes

- `check-panel-findings-closed` accepts a re-run closed `clean` as well as `completed`.
- `check-task-commit-fields` (Go and its Python parse half, which `check-plan-shape` imports)
  reads `Case <N>` with an optional letter suffix; the parse-rule message and
  `skills/flow/brainstorm-planner.md` say so.
- `pipeline-fix verify` adds the two Go-tree audit harnesses whenever a `stats/**.go` file is
  touched.
- `run-verify-list.sh <worktree> test` records the sha of a green run on a clean tree in the
  worktree's git dir, and a later run skips every command at that sha, or each command whose
  `# inputs:` pathspecs the delta leaves untouched. The full-suite run after the last group runs
  through it, the verify contract and the `## test` configuration row describe it, and the agents
  project declares inputs on its SPA test line.
