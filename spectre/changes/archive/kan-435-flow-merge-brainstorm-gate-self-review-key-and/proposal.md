# kan-435-flow-merge-brainstorm-gate-self-review-key-and

**Jira:** KAN-435 — "flow: merge brainstorm gate, self-review key, and self-review filing prompt"
**Source:** `docs/superpowers/research/flow-speedup.md` — attention merges 1, 2a and 2b (§7 round
two, decided in §8).

## Why

Three operator prompts that cost attention and buy nothing, measured in this repository's dev
stats store on 2026-09-04:

- **The brainstorm double gate.** `flow.design-approval`'s median wall clock is 0 s over 49 runs —
<!-- measured: docs/superpowers/research/flow-speedup.md §7 round two, from /api/v1/stats/stage-leaderboard on the dev store @ 2026-09-04 -->
  the approval is a reflex seconds after the convergence confirm, a second answer to the same
  question.
- **The self-review skip prompt.** It fires only after `FINISHED`, when the operator has already
  walked away — 20 minutes on kan-404, a 14.6-minute mean over 44 runs — and this repository's
<!-- measured: docs/superpowers/research/flow-speedup.md §7 and §8, from flow.self-review stage runs on the dev store @ 2026-09-04 -->
  report series is already dead: the newest report is kan-380's, and the six changes since all
  answered "No". Thirty reports yielded nine Jira tickets.
- **The filing prompts.** When self-review does run, five per-angle multi-select prompts plus the
  rating are six prompts for one decision.

## What changes

1. **The convergence confirm is the design approval.** It gains a third option — *Nothing unclear
   — approve the design and move on* (recommended) / *Another round — I have something* (default
   on silence) / *Revise — I have a change to the design* — and the first option both closes the
   checklist and grants the HARD GATE approval. The parent still marks `flow.brainstorm` end and
   `flow.design-approval` begin/end around that one relayed answer; stage keys are unchanged.
2. **A new optional `## self review` key in `<project>/.flow/project.md`**, body `run` or `skip`
   in `## default landing route`'s single-line-literal shape, read by `project-get.sh` and
   resolved by run 2's step 9 before its skip prompt. This repository sets `skip`.
3. **One filing-and-rating prompt.** The five per-angle prompts and the rating collapse into one
   `AskUserQuestion` call: up to three multi-select questions of three angle-labelled findings plus
   *None — file nothing* each, and the rating as the last question. `operator-prompts.md`'s stale
   `skills/myflow-do/SKILL.md` reference is fixed in passing.

Prompts per change in this repository: from n + 5 (self-review declined) to n + 1 with the key
set to `skip`.

Touches: `skills/flow/brainstorm-planner.md`, `skills/flow/brainstorm.md`,
`skills/flow-contracts/pipeline.md`, `skills/flow-contracts/pipeline-rationale.md`,
`skills/flow-contracts/project-configuration.md`, `skills/flow/archive.md`,
`skills/flow-contracts/finish-contract-run2.md`, `skills/flow-contracts/handoff-blocks.md`,
`skills/flow-contracts/operator-prompts.md`, `.flow/project.md`.
