# flow-progress-subagent-board

## Why

Operator, 2026-10-03: `/flow` registers and updates a harness task-list checklist of its steps in
parallel with the `subagent-board` mod (`mods/subagent-board/`), which already renders a run's
progress from its dispatch descriptions and `flow stage` marks. One progress view is enough.

## What changes

- `skills/flow-contracts/pipeline.md`, **Progress visibility**: `/flow` registers nothing with the
  harness's task-list mechanism; the live progress view on Claude Code is the `subagent-board` mod.
  The dispatch-description labelling rule, "the progress view is a view, never a record", the
  `tasks.md` single-source-of-truth and no-third-checkbox-marker statements stay. A harness without
  the mod (ZCode) keeps printing the equivalent block. **Quiet progress** names the mod's view.
- `skills/flow/SKILL.md`: the "register this run's steps" instruction is removed.
- `skills/flow/implement.md`: the per-task task-list granularity paragraph is removed.
- `skills/flow-fast/SKILL.md`: the task-list registration step is removed; the plan's task
  granularity (one task per file or logical unit the change touches) moves to its `writing-plans`
  step.
- `skills/flow-status/SKILL.md`: the "registers nothing" paragraph, moot once no `/flow*` skill
  registers, is removed.
- `skills/flow-contracts/pipeline-rationale.md`: **Progress visibility** records why the mod
  replaced the task list.
