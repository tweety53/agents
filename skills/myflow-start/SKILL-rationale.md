# myflow-start — rationale

This file is the reasoning behind `skills/myflow-start/SKILL.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## State gate

## Superpowers Basic Workflow (this stage)

## A. Resolve the change

## Ask the planning effort, the models, and the review panel roster — creating runs only

## B. Basic Workflow #1 — Brainstorming

### Convergence

**Why the convergence confirm's default deliberately breaks from the recommended option.**
Recommending *move on* is honest only while an operator has actually said so; a silent or stalled
operator has not said so, which is why the safe default for this one prompt is *another round*
rather than the marked recommendation. The `⚠ another round — no explicit answer` marker exists so a
reader of the handoff can tell an operator-requested round from one nothing could confirm — the two
look identical in the transcript otherwise.

superpowers:brainstorming's own flow is linear — explore, ask clarifying questions, propose
approaches, present the design, get approval — and its only loop routes a **correction** back to the
design. Nothing in it routes a **question** back to the questions, so a question raised *by* an
answer has nowhere to land. This section is where it lands.

**Every planning effort level runs this loop.** A level changes how many questions one round groups
— one at a time at `detailed`, batched at `low` — and never whether another round opens. A level
able to end the loop early would be a way to skip the gate rather than a way to size the thinking
inside it. The levels are **Planning effort** (`skills/myflow-contracts/state-file.md`).

**A session that cannot ask at all is a narrower, bounded exception, not a second version of "an
operator who is present but silent."** The confirm fires only when the convergence test came back
empty, so opening "another round" here has nothing to explore and, with no hard cap, only re-empties
the test and re-fires the confirm — `empty test → confirm → no answer → another round → empty test →
confirm`, without end. **Unrecognised statuses** (`skills/myflow-contracts/jira-integration.md`)
already names this exact failure mode for its own interactive ask — "a session that cannot ask at
all" — and gives it a terminating outcome rather than a retry; cited rather than restated here.

**`3` is a tuned value, and this file is the only place it
is written** — the contract and the pipeline carry the shape of the bound, never the number, so it
can move without amending either. **The threshold counts rounds, not questions**, so it lands at a
different point in the conversation depending on planning effort: at `detailed`, where a round is
one question, the offer can appear after as few as three questions; at `low`, where a round batches
many, the same threshold takes much longer to reach, or is never reached at all. That coupling is
accepted rather than compensated for — the threshold is stated once, in rounds, and each level's own
grouping decides how much a round holds.

The confirm recommends *moving on* precisely because it is unreachable while
this command holds an unanswered question. The offer recommends *another round* for the mirror
reason: it is reachable only while this command genuinely holds one. That is the same shape as the
**Stop** recommendation at the unfinished-work gate of `/myflow-finish` run 1, whose reasoning is
stated under **Finish contract** (`skills/myflow-contracts/finish-contract.md`) and is not
re-argued here.

## C. Create the change and its artifacts

### Decisions

### Open questions

**This cross-reference is instruction-only, and no guard checks it.**
`<agents repo>/scripts/check-references.sh` verifies a bold token beside a backticked path against a heading in
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
