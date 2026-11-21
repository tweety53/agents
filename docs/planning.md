# Planning: triage, then a cold plan review

Planning is the one fully interactive part of `/flow`: you answer every ambiguity and approve the
design. Two model calls around it keep the effort proportional and the plan honest. Both are
"headless": one-shot, read-only calls that ask nobody anything.

## Triage: easy or hard?

Before anyone designs, one call reads the ticket and the repository's file list and answers
`short` or `full`.

- **Short** — the change sits in one known area, has one obvious approach and makes no visible UI
  change that needs a prototype. One planner does everything, reading the code itself.
- **Full** — it spans repositories or apps, changes a data model or a public contract, is a
  visible change with no mockup, or has real design alternatives or open product questions.

Both still run design rounds with you and end on your approval. When in doubt the answer is `full`.

Canonical: [`skills/flow/plan-triage-prompt.md`](../skills/flow/plan-triage-prompt.md).

## Plan review: read cold

Once the plan is written, one call reads the proposal, the design and the task list without ever
having seen the conversation that produced them, so the documents have to stand on their own. It
applies a fixed set of lenses and writes a report of problems.

The planner then triages that report: a problem with one obvious fix is fixed in the plan, one
whose fix would change a design decision comes back to you as a question, and one that still
stands is written into the design's risks as accepted, with the reason. A plan changed this way is
re-decided once, never reviewed a second time.

Canonical: [`skills/flow/plan-check-prompt.md`](../skills/flow/plan-check-prompt.md) and the lenses
in [`skills/flow/plan-review-rubric.md`](../skills/flow/plan-review-rubric.md); the triage of the
report is **Plan review** in
[`skills/flow/brainstorm-planner.md`](../skills/flow/brainstorm-planner.md).

## The decision: how it is built and reviewed

The same call also decides three things, so nothing in a project configures them by hand:

- **Implementer groups** — which tasks one agent builds together, and which can run side by side.
- **The review panel** — which reviewers the change needs ([Review panel](review-panel.md)).
- **Visual verification** — required, skipped, or not configured ([Visual verify](visual-verify.md)).

The decision is printed with the plan summary before anything is built.

Canonical: [`skills/flow/plan-check-decision.md`](../skills/flow/plan-check-decision.md); the gate
around it is **Plan check gate** in
[`skills/flow/brainstorm-planner.md`](../skills/flow/brainstorm-planner.md).
