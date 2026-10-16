# Operator prompts

**This file is canonical for the shape of an operator-facing prompt.** It is the one prose shape every approval or choice takes — offered as
options, not as open prose — stated once so every call site stops restating it.

## The shape

A prompt in this shape states:

- the question, with named options
- exactly one option marked (recommended)
- what happens if the operator is silent — the safe default, always the recommended option
- a 🤖 marker in the handoff when that silent default actually fired, or when **Auto-resolution**
  below took it

## Explain before asking

**A flow-eval candidate's explanation is written in chat as the turn's final message, and the turn
ends there.** That covers each flow-eval candidate in its kind's template exactly as step 2 of
`skills/flow-self-review/SKILL.md` defines it (impl, review or visual, every field). The
`AskUserQuestion` that offers the candidates fires only in a later turn, after the operator has seen
the explanation; the operator may instead answer in chat directly.

**Every other ask explains and asks in the same turn.** A finding, a filing item, a gate or any other
question the operator must read up on first is explained in chat — observed, what breaks, the fix —
and the `AskUserQuestion` follows in that same turn; the turn is never ended just to split the two.

Either way, the explanation is never carried only inside the `AskUserQuestion`'s question, option or
preview text. Option text records the decision only.

Why: a harness focus mode shows only a turn's final message, so a flow-eval candidate's template
written before a call in the same turn is never seen. Carrying the explanation inside the prompt
instead was rejected by the operator on 2026-10-16: the explanation belongs in chat, in the agreed format, not in the prompt. Splitting every ask across two turns was
rejected by the operator the same day: the split is for flow-eval candidates alone.
An unanswered call of this shape follows **Unanswered mid-run asks**
(`skills/flow-contracts/pipeline.md`).

## Every item offered

**A prompt that asks the operator to choose from a list — findings to file, items to act on —
offers every item on it.** Never a subset the run picked, and never one dropped for the prompt
tool's option limit: group the items into questions (`AskUserQuestion` takes four questions of four
options), and past that, further calls. "Not actionable" is no reason to withhold an item — a
measurement becomes a baseline, a conclusion records why an alternative was rejected; offer it and
let the operator decline. Naming an item in prose is not offering it. A call site whose own rule
fixes or declines items without asking — **3. Fix every finding that is not `big`**
(`skills/flow-self-review/SKILL.md`) — offers every item that rule leaves.

## Batched asks

A call site that holds several of this contract's prompts outstanding at one point — one fix
round's handbacks — presents them as one batched ask, never one ask per prompt as each is
discovered: one ask, one entry per outstanding prompt, each entry a prompt in **The shape**
above, stating its own question and options. **Auto-resolution** below resolves per entry exactly
as it resolves a single prompt — an entry whose recommended option may be taken unasked is taken
and recorded without being asked; only the entries **What still stops** leaves asked join the
ask — and the silent default is per entry, each asked entry's recommended option. One ask is one
stop: the run waits once for the batch, not once per entry.

## The doctrine

Every call site cites this contract for the mechanics and states only its own question text
and options — never the shape itself.

## Auto-resolution

**Load `skills/flow-contracts/operator-prompts-auto-resolution.md`** only when a prompt arises during implementation or a fix run, or during planning when `## decisions` resolves to `recommended`.
The rules are **Auto-resolution** (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
