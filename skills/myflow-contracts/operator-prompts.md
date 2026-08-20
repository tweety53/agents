# Operator prompts

**This file is canonical for the shape of an operator-facing prompt.** It consolidates a pattern
already required by `myflow-planning-gate`'s existing "Every approval or choice is offered as
options, not as open prose" requirement (`<agents repo>/openspec/specs/myflow-planning-gate/spec.md`) — this is
not a new rule, it is the one prose shape that requirement's options-not-prose mechanics already
imply, stated once so every call site stops restating it.

## The shape

A prompt in this shape states:

- the question, with named options
- exactly one option marked (recommended)
- what happens if the operator is silent — the safe default, always the recommended option
- a ⚠ marker in the handoff when that silent default actually fired

## The multi-select variant

Some prompts ask the operator to choose any subset of several options, not exactly one. This
variant states:

- the question, with each option listed separately
- that the operator may select any subset of the listed options — none, one, or several
- one explicitly stated default — named by the call site — for what happens if the operator is
  silent

This contract fixes the shape, not the default's polarity: the call site chooses it, and states it
plainly. A safe default may resolve silence to the empty set (an explicit "None" option, marked
recommended) or to the full set (every listed option, silence needing no option of its own to name
it) — whichever matches what the options actually control. The two existing multi-select call
sites choose oppositely: the self-review filing ask's safe default is "file nothing" (silence
selects None), while `skills/myflow-do/SKILL.md`'s optional review-slot ask defaults to "include
all of them" (silence selects every fired trigger).

## The doctrine

Every call site below cites this contract for the mechanics and states only its own question text
and options — never the shape itself.
