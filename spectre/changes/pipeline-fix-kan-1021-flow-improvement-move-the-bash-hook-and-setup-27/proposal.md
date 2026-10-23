# pipeline-fix-kan-1021-flow-improvement-move-the-bash-hook-and-setup-27

## Why

kan-1021's plan ended removal tasks on an absence sweep that "prints nothing", but Tasks 5 and 6
swept `stats`, which covered `stats/internal/guard/pipelinefix.go` and
`stats/internal/guard/pipeline_fix_test.go` — Task 10's files, still naming the removed harnesses —
so each sweep printed hits at its own commit and needed a `Correction:` line; Tasks 1–4's sweeps
likewise hit doc comments the plan told them to keep (panel F2). The rule in
`skills/flow/brainstorm-planner.md` (**End a removal task on an absence sweep**) already scopes a
sweep to the task's own paths, but no guard enforces it, so the misfire reaches implementation.

## What changes

- `check-plan-shape.sh` gains finding F18: an unticked task's one-line absence sweep whose `git
  grep` paths cover a later task's `**Files:**` path (that the task does not own) while that file
  still holds one of the sweep's `-e` strings at the change's base is refused, at the sweep's line.
- F15's base resolution is extracted into a shared helper both findings use.
- `TestCheckPlanShape` gains case 43, failing without F18.
- `skills/flow/brainstorm-planner.md` gains one sentence naming the limit and F18; its reasoning
  goes to `skills/flow/SKILL-rationale.md`.
