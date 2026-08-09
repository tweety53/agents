# myflow-fast — rationale

This file is the reasoning behind `skills/myflow-fast/SKILL.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

**Load `skills/myflow-contracts/pipeline.md` first** — it is canonical for the states, the
command→state transition table, the wrong-state handoff, git boundaries, and the handoff output
shape.

## State gate

## No state file — brainstorm into implementation

Why the post-design "does this look right?" confirm is skipped here, when `/myflow-start` still
stops at it: `/myflow-fast` exists to remove friction from a command whose whole point is speed,
and the operator explicitly asked for this specific stop to go. Brainstorming's clarifying
questions and the design presentation itself are kept fully interactive — that is where
requirements actually get gathered and where the operator can redirect the design before anything
is built — so nothing about requirements-gathering is weakened. What the confirm added on top of
that was a second, purely procedural pause with nothing left to decide by the time it fires. The
human checkpoint it removed is not gone; it moves downstream to the `IN_PROGRESS`
staged-diff-and-run-instructions review, which already exists in the pipeline and already gates
every `/myflow-do` run the same way. Keeping the confirm but auto-answering it was considered and
rejected — that still leaves an interactive stop on every creating run, which is exactly the
friction being removed. This skip is scoped to `/myflow-fast` alone; `/myflow-start` was never the
command being optimized for speed here, and removing its own gate was never asked for.

## At `IN_PROGRESS`

### After merge-and-push specifically

### After open PR or manual specifically

## Recorded defaults favor speed

Why the planning-effort, model, and review-panel-roster question round is recorded silently
instead of asked: the operator flagged the interactive round itself as friction on a command whose
whole point is speed — every one of `/myflow-fast`'s creating runs paid for four
`AskUserQuestion` stops even when the recommended answer was what most runs wanted anyway. The
defaults recorded here are exactly the recommended answers `/myflow-start`'s own question round
already marks: `sonnet` for the three model roles and `light` for the roster. Keeping the prompts
but pre-selecting the recommendation was considered and rejected for the same reason as the
design-approval confirm above — it still stops every creating run, which is the friction being
removed, not just its cosmetic cost. An explicit session instruction still overrides any one field,
so nothing about operator control is lost — only the default path's silence changed.

## State write and handoff

## Guardrails
