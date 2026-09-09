# kan-474-flow-brainstorm-planner-should-batch

Linked issue: KAN-474 — "flow: brainstorm planner should batch AskUserQuestion calls".

## Why

In the KAN-445 run's brainstorm phase (session `a7273417-c46a-4128-9f73-619e4a99c84d`, planner
dispatch `ae7017f9938067139`) the planner asked three clarifying questions as three separate
`AskUserQuestion` calls, one question each, at 11:32:59, 11:36:18 and 11:40:00. Every call is its
own turn, and every turn resends the full accumulated context: cache-read grew 51,968 → 54,080 →
58,432 tokens across those three calls. `AskUserQuestion` carries up to four questions per call, so
two of those three resends were avoidable.

The cadence is driven by text, in two places at once:

- `superpowers:brainstorming` (v6.3.0, "Understanding the idea") says "ask questions one at a time"
  and "Only one question per message".
- The planner relay contract in `skills/flow/brainstorm.md` has the planner end every turn with
  "exactly one `## Question` block", and the parent ask "each `## Question` block" through one
  `AskUserQuestion` — the block shape itself has no room for a second question.

`skills/flow-research/SKILL.md` carries the same instruction in-session ("Ask one question at a
time") and pays the same per-question resend without a relay.

## What changes

- The planner asks every pending question whose wording does not depend on another pending answer
  in the same turn, each as its own `## Question` block, up to four blocks per turn — a scoped
  override of `superpowers:brainstorming`'s one-question-per-message rule, `/flow` only. The
  convergence confirm and the third-round offer may ride along as the last block of such a turn.
- The parent collects every `## Question` block of a planner turn into a single `AskUserQuestion`
  call, one question per block in block order, and resumes the planner with all answers labelled by
  block.
- `/flow-research` asks up to four independent pending questions in one `AskUserQuestion` call per
  turn instead of one at a time.
- Nothing else moves: the conductor relay in `skills/flow/implement.md` keeps its single-block
  contract, no mechanical guard is added, and no spec is edited (`spectre/specs/` is empty).
