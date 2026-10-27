# pipeline-fix-kan-1024-templates-warm-up-cool-down-any-workout-37

## Why

The `/flow` parent ends its turn mid-run with no stop the contracts allow. In gymie kan-1024
(2026-10-27) it ticked all 7 tasks, wrote a final report and ended the turn with the full suite,
the review panel, verify and `flow.write-in-progress` still ahead — state `STARTED`,
`reviewBeforeLanding` true, the operator having asked to stop only after implementation was fully
done. `skills/flow-contracts/pipeline.md` already forbids it ("A turn ends only at a gate, a block,
or the run's final handoff"); kan-1015, kan-1017, kan-1018 and kan-1033 are the same defect. Prose
has not held, so the stop must be refused mechanically.

## What changes

- New Claude Code `Stop` hook `hooks/flow-stop-guard.py`: when this session's last `/flow` change
  is mid-phase, it answers `{"decision":"block","reason":"..."}` naming the next stage, so the turn
  cannot end.
- It allows the stop at an AskUserQuestion, an explicit operator stop, a gate (`STARTED` with no
  tasks, the plan-gate handoff, `IN_PROGRESS`, `FINISHED`), and when `stop_hook_active` is true with
  nothing advanced since its last block.
- `setup.sh` warns when the hook is installed but not registered under `"Stop"`, like its siblings.
- `skills/flow-contracts/pipeline.md`'s turn-end paragraph names the hook as its enforcement.
- A Go repotest pins every allow and block case.
