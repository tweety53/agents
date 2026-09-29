# Operator prompts

**This file is canonical for the shape of an operator-facing prompt.** It is the one prose shape every approval or choice takes — offered as
options, not as open prose — stated once so every call site stops restating it.

## The shape

A prompt in this shape states:

- the question, with named options
- exactly one option marked (recommended)
- what happens if the operator is silent — the safe default, always the recommended option
- a ⚠ marker in the handoff when that silent default actually fired, or when **Auto-resolution**
  below took it

## The doctrine

Every call site cites this contract for the mechanics and states only its own question text
and options — never the shape itself.

## Auto-resolution

**Load `skills/flow-contracts/operator-prompts-auto-resolution.md`** only when a prompt arises during implementation or a fix run, or during planning when `## decisions` resolves to `recommended`.
The rules are **Auto-resolution** (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
