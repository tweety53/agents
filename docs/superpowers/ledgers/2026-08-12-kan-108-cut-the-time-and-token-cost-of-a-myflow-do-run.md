# SDD ledger — kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run

**This ledger was reconstructed at `/myflow-finish` run 1, not maintained per dispatch.** That is a
deviation from Model policy, which requires each dispatch to record its model here as it happens. The
values below are accurate — every dispatch named its model explicitly in its own prompt, and none
inherited the parent's — but the record was assembled after the fact from those dispatches rather
than written as they ran, and a reader should weigh it accordingly.

Recorded model choices for this change: `models.implementation` sonnet, `models.reviewPanel` sonnet,
`models.panelFix` sonnet, `reviewPanelRoster` light. `/myflow-fast` records these defaults rather
than asking, which is why the implementer runs on sonnet rather than the Opus the standing policy
defaults it to.

## Task dispatches

| Task | Outcome | Model |
|------|---------|-------|
| 1 — `check-panel-reproducers.sh` + harness | complete (commit rewritten to `c7fb19b` by later fixups; review clean after one fix round) | sonnet |
| 2 — narrow the escalation trigger | complete (`784aa2f`, review clean) | sonnet |
| 3 — reproducer-verified fix dispatch | complete (`451373c`, review clean) | sonnet |
| 4 — reviewer prompts ask for a reproducer | complete (`7cfcff4`, review clean) | sonnet |
| 5 — declare the harness under `## test` | complete (`3580849`, review clean) | sonnet |
| 6 — verification sweep | complete, no commit — the sweep found nothing to fix | sonnet |
| 7 — `run-reproducer.sh` + harness | complete (`54874ac`) — added mid-run at the operator's direction | sonnet |

Tasks 2–6 were dispatched as **one bundle** by `scripts/plan-dispatch-bundles.sh`, which grouped them
on their declared `**Files:**` overlap; task 1 was its own bundle. Implementers were serialized: one
in flight against this worktree at a time.

## Per-task reviews

Under the `light` roster a single combined reviewer covers spec compliance and code quality per task.
Task 1 reviewed on sonnet (3 findings, all fixed); task 2 on sonnet (clean); task 3 on sonnet
(clean); tasks 4 and 5 reviewed together on sonnet (clean).

## Review panel

Seven passes. Every slot dispatched on sonnet, named explicitly. Slot 4 (Security) is normally
`subagent_type: security-review`; that type is unavailable in this session, so it ran as a
`general-purpose` reviewer on this repository's own `security-reviewer-prompt.md` — a recorded
substitution, and its ledger entry names sonnet rather than `unknown (agent-defined)` because this
dispatcher named the model. Bugbot was never dispatched: the `light` roster does not include it.

| Pass | Slots dispatched | Findings |
|------|------------------|----------|
| 1 | Primary, Principles, code-review-low, Security, Adversarial, Lens B, Lens C | 12 |
| 2 | same seven (Full escalation) | 15 |
| 3 | same seven (Full escalation) | 14 |
| 4 | Primary, Principles, code-review-low, Security — Adversarial, Lens B and Lens C stopped by the operator mid-pass | 3 |
| 5 | Primary, Principles, code-review-low | 4 |
| 6 | Primary, Principles, code-review-low | 7, plus 2 from a peer session's reuse review |
| 7 | Adversarial alone, on `run-reproducer.sh` | 3 |

## Fix rounds

Five, each dispatched on sonnet per `models.panelFix`: 11 findings, then 12, then 14, then 7, then 3.
One trim pass, operator-directed, also on sonnet.

## What this ledger cannot tell you

The per-dispatch token cost. Each subagent's usage was reported to the parent as it completed, but
those figures were not accumulated here as they arrived, so the total for this change is recoverable
only from the session transcript. For a change whose own ticket is about token cost, that is the most
useful number this file does not carry.
