# myflow-start — rationale

This file is the reasoning behind `skills/myflow-start/SKILL.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## State gate

## Superpowers Basic Workflow (this stage)

## A. Resolve the change

## Ask the planning effort and the models — creating runs only

## B. Basic Workflow #1 — Brainstorming

### Convergence

superpowers:brainstorming's own flow is linear — explore, ask clarifying questions, propose
approaches, present the design, get approval — and its only loop routes a **correction** back to the
design. Nothing in it routes a **question** back to the questions, so a question raised *by* an
answer has nowhere to land. This section is where it lands.

What counts as a planning-stage exchange, the convergence test itself, and why it is one test rather
than a rule per gate are structure, stated once under **Stage exit — never the command's own
judgment** (`skills/myflow-contracts/pipeline.md`) and not restated
here. What is genuinely tuned for this command follows: the two prompts, the threshold, the
no-hard-cap rule, and why their opposite recommendations are not to be harmonised.

**Every planning effort level runs this loop.** A level changes how many questions one round groups
— one at a time at `detailed`, batched at `low` — and never whether another round opens. A level
able to end the loop early would be a way to skip the gate rather than a way to size the thinking
inside it. The levels are **Planning effort** (`skills/myflow-contracts/state-file.md`).

## C. Create the change and its artifacts

### Decisions

### Open questions

**This cross-reference is instruction-only, and no guard checks it.**
`scripts/check-references.sh` verifies a bold token beside a backticked path against a heading in
another *file*; it has no notion of an `answered by <decision-id>` link between two sections of the
same `design.md`, nor of the uniqueness rule just stated. A decision ID that is never created, a
status left at `open` after its question was actually answered, or an ID reused across the two
sections all pass every guard this repository runs — the same limit
**Plan provenance** (`skills/myflow-contracts/plan-provenance.md`)'s "What the guard does not do"
names for a provenance tag: the tool confirms a claim is stated, never that it is true.

## D. Basic Workflow #3 — Writing plans

## E. Publish the proposal artifact

## F. Write state and hand off

The block below is **not** a second definition of the handoff. It is this command's rendering of the
`STARTED` template, which is defined once under
**The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`) and is canonical for the
labels, the field set and their order. What this block adds is the enumeration of the literal
alternatives `/myflow-start` writes for each placeholder that file describes. **Change the template
first and bring this block with it** — a field added here and not there is drift the moment
`/myflow-status <name>` regenerates the same state.

**`missing` is a real alternative on the artifact line, not a defensive one.** This command's
guardrails forbid finishing without publishing, so its own runs print a URL; the alternative is
carried because `/myflow-status <name>` renders this same block from a state file whose
`artifactUrl` may legitimately be `null` — the field is nullable, and a hand-edited or otherwise
incompletely-written file can carry it that way. Omitting the alternative here would narrow the
template and teach the next reader to drop the case the missing-rather-than-dropped rule requires.

## Guardrails
