# Operator prompts

**This file is canonical for the shape of an operator-facing prompt.** It consolidates a pattern
already required by `myflow-planning-gate`'s existing "Every approval or choice is offered as
options, not as open prose" requirement (`openspec/specs/myflow-planning-gate/spec.md`) — this is
not a new rule, it is the one prose shape that requirement's options-not-prose mechanics already
imply, stated once so every call site stops restating it.

## The shape

A prompt in this shape states:

- the question, with named options
- exactly one option marked (recommended)
- what happens if the operator is silent — the safe default, always the recommended option
- a ⚠ marker in the handoff when that silent default actually fired

## The doctrine

Every call site below cites this contract for the mechanics and states only its own question text
and options — never the shape itself.
