# kan-474-flow-brainstorm-planner-should-batch — design

## Context

Prose-only change to three skill files. No script, no guard, no spec.

## The edits

### `skills/flow/brainstorm-planner.md`, section B, "The checklist"

Add a bullet beside the existing scoped override of the brainstorming skill's design-approval gate:

- Ask every pending question whose wording does not depend on another pending answer in the same
  turn — each as its own `## Question` block, up to four blocks per turn. A question that only makes
  sense once another is answered waits for the next turn. The convergence confirm and the
  third-round offer may be the last block of such a turn. This is a scoped override of
  `superpowers:brainstorming`'s "Only one question per message", `/flow` only.

### `skills/flow/brainstorm-planner.md`, section B, "Convergence"

State what a batched confirm means: when the confirm rides along with a round's questions,
**approve the design and move on** folds that turn's answers into the design and ends the stage; an
answer to an accompanying question that names something new opens another round regardless of the
confirm's choice. The silence default (another round) and the `⚠ another round — no explicit
answer` marker are unchanged.

### `skills/flow/brainstorm.md`, the relay contract and the relay

- The relay contract: the planner ends every turn with **one to four** `## Question` blocks (was
  "exactly one"), or one of the three returns.
- The relay: the parent puts every `## Question` block of the turn into a single `AskUserQuestion`
  call — one question per block, in block order, each block's named options as that question's
  options — and resumes the planner with every answer, labelled by block order. Prose preceding the
  blocks is still shown first, as today.

The dispatch prompt the parent writes follows this file, so the planner's opening instruction
changes with it.

### `skills/flow-research/SKILL.md`, "Say when you don't know"

Replace "Ask one question at a time." with: ask every pending question whose wording does not
depend on another pending answer in one `AskUserQuestion` call, up to four per call; a dependent
question waits for the next turn. The sentence that follows (don't funnel the user through a fixed
line of questioning) stays.

## Budgets

`scripts/check-contract-budget.sh` bounds all three files: `brainstorm-planner.md` 20,225 of
23,248 bytes, `brainstorm.md` 11,908 of 35,015, `flow-research/SKILL.md` 11,759 of 13,047. The
edits add well under 1,000 bytes to each, so no budget line changes.

## Decisions

### Prose rule at each call site, no shared contract file

**ID:** question-batching-prose-only
**Status:** active
**Chosen:** state the batching rule in `brainstorm-planner.md`, `brainstorm.md` and
`flow-research/SKILL.md` directly — two sentences each, matching the ticket's own fix direction.
**Considered:** a shared `skills/flow-contracts/question-batching.md` cited by every site — rejected,
three two-sentence sites do not earn a file and `operator-prompts.md` already fixes prompt shape; a
mechanical guard — rejected, see `no-mechanical-guard`.

### Scope is the planner and `/flow-research`

**ID:** scope-planner-and-research
**Status:** active
**Chosen:** the planner relay and `/flow-research`'s in-session questioning — both pay the
per-question resend the ticket measured.
**Considered:** planner only — rejected by the operator, research carries the identical
instruction; adding the conductor relay in `implement.md` — rejected, its questions are rare and
each blocks the next.

### A batch is several `## Question` blocks in one turn

**ID:** batch-shape-multiple-blocks
**Status:** active
**Chosen:** up to four `## Question` blocks per planner turn, the parent collecting them into one
`AskUserQuestion` call — each block keeps today's shape, so nothing about a single question changes.
**Considered:** one `## Question` block carrying up to four numbered sub-questions — rejected by the
operator, it changes the block's shape for the single-question case too.

### The convergence confirm and third-round offer may be batched

**ID:** confirm-batchable
**Status:** active
**Chosen:** either may be the last block of a batched turn; **approve** then folds that turn's
answers in and ends the stage, and any accompanying answer that names something new is another
round regardless.
**Considered:** standalone only — rejected by the operator, it forces one extra turn per round for
a confirm whose silence default already protects the gate.

### No mechanical guard

**ID:** no-mechanical-guard
**Status:** active
**Chosen:** prose rule only.
**Considered:** a stats-store query over `flowd` flagging a planner dispatch with more than one
single-question `AskUserQuestion` call — rejected by the operator; nothing in-tree can count calls
per dispatch today and the measurement would need its own change.

## Open questions

None.
