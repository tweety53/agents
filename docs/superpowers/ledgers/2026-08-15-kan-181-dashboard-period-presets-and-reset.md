# SDD ledger — kan-181-dashboard-period-presets-and-reset

Every subagent dispatch this change made, with the model it ran on, per **Model policy**
(`skills/myflow-contracts/pipeline.md`): these lines are the only evidence of the model a dispatch
actually ran on, and the state file's `models` fields record intent rather than what happened.

**Recorded model roles** (`/myflow-fast` defaults, no operator override): `implementation: sonnet`,
`reviewPanel: sonnet`, `panelFix: sonnet`. Roster `light`.

**Operator instruction for this change:** light roster with capped fix rounds — converge rather than
chase every new finding, and bring anything remaining to the operator as a decision. Recorded here
because it governed how the panel ran.

## Implementation dispatches

Tasks 1, 2 and 5 were one bundle (`scripts/plan-dispatch-bundles.sh` grouped them — all three touch
`App.tsx`), dispatched to a single implementer that made one commit per task. Tasks 3 and 4 were
their own bundles.

| Task | Outcome | Final commit | Model |
|------|---------|--------------|-------|
| 1 | complete | `b32d4e2` | sonnet |
| 2 | complete | `65eb912` | sonnet |
| 3 | complete | `7e0bc97` | sonnet |
| 4 | complete | `a0906b5` | sonnet |
| 5 | complete | `10685a3` | sonnet |

Shas are the final ones after three rebases; earlier shas appear in the panel record's history.

## Per-task review dispatches

One combined spec-plus-quality reviewer per bundle, which is what the `light` roster specifies.

| Reviewing | Model | Result |
|-----------|-------|--------|
| Tasks 1, 2, 5 | sonnet | 1 Important (URL clobber on hashchange) |
| Tasks 3, 4 | sonnet | 2 Critical (moving-clock comparison), 1 Minor (duplicated comparison) |
| Task 2, after its fix | sonnet | clean — mutation-tested the fix and the fallback tests |

## Review panel dispatches

One pass, three required slots, all on the panel's recorded model. No slot was dispatched by
`subagent_type`, so no entry reads `unknown (agent-defined)`.

| Pass | Slots | Model | Result |
|------|-------|-------|--------|
| 1 | Primary · Principles · Code review (low) | sonnet | all three READY; all three raised the same single Minor |

All optional slots were **declined by the operator**, recorded distinctly from a trigger that never
fired.

## Fix dispatches

| Round | Findings | Model |
|-------|----------|-------|
| 1 | the URL-clobber Important, fixed by the bundle's own implementer (resumed, not re-dispatched) | sonnet |
| 2 | both Criticals and the duplication Minor, fixed together by one shared `sameRange` | sonnet |
| 3 | the panel's comment-accuracy Minor — applied directly by the dispatcher, not delegated, because it was a comment with no behavioural effect and three reviewers had already agreed on the exact wording problem | — |

`models.panelFix` is recorded as `sonnet`, which is `/myflow-fast`'s speed-favouring default rather
than the pipeline-wide default of the strongest available model. Named here so the choice is
auditable rather than inferred.

## What this change cost to get right

Five defects, three of them found before the panel ran. Two were Critical and shared one root cause:
a derived comparison run against a clock that moves. The tests could not catch them because every one
used fake timers and asserted immediately after render — so the fix had to include tests that advance
the clock and re-render, or the same class of defect returns unseen. That is the durable lesson from
this change, and it is recorded in the panel record too.
