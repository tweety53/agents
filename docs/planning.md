# Planning: design, then a cold plan review

Planning is the one fully interactive part of `/flow`: you answer every ambiguity and approve the
design. One model call after it keeps the plan honest. It is "headless": a one-shot, read-only
call that asks nobody anything.

## Plan review: read cold

Once the plan is written, one call reads the proposal, the design and the task list without ever
having seen the conversation that produced them, so the documents have to stand on their own. It
applies a fixed set of lenses and writes a report of problems.

The planner then triages that report: a problem with one obvious fix is fixed in the plan, one
whose fix would change a design decision comes back to you as a question, and one that still
stands is written into the design's risks as accepted, with the reason. A plan changed this way is
re-decided once, never reviewed a second time.

The last lens is an estimate: how confident the reader is, 1–100%, that the plan is executed as
written, and what would move it to 100%. Below 100% you are asked once — **Raise it**, and the
planner closes each gap by its suggestion, asking you where only you can answer, or **Later**, and
the gaps are accepted as risks. Either way the estimate is filed. Each implementer files the same
estimate for its own tasks before its first edit and continues without asking. The **Confidence**
view of the dashboard lists every filed estimate by tag.

Canonical: [`skills/flow/plan-check-prompt.md`](../skills/flow/plan-check-prompt.md) and the lenses
in [`skills/flow/plan-review-rubric.md`](../skills/flow/plan-review-rubric.md); the triage of the
report is **Plan review** in
[`skills/flow/brainstorm-planner.md`](../skills/flow/brainstorm-planner.md).

## The decision: how it is built and reviewed

The same call also decides three things, so nothing in a project configures them by hand:

- **Implementer groups** — which tasks one agent builds together, and which can run side by side.
- **The review panel** — which reviewers the change needs ([Review panel](review-panel.md)).
- **Visual verification** — required, skipped, or not configured ([Visual verify](visual-verify.md)).

The decision is printed before anything is built.

Canonical: [`skills/flow/plan-check-decision.md`](../skills/flow/plan-check-decision.md); the gate
around it is **Plan check** in
[`skills/flow/brainstorm.md`](../skills/flow/brainstorm.md).
