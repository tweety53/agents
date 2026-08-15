# SDD ledger — kan-180-separate-db-and-app-for-ui-testing

Every subagent dispatch this change made, with the model it ran on. Recorded per
**Model policy** (`skills/myflow-contracts/pipeline.md`): these lines are the only evidence of the
model a dispatch actually ran on, and the state file's `models` fields record intent rather than
what happened.

**Recorded model roles for this change** (`/myflow-fast` defaults, no operator override):
`implementation: sonnet`, `reviewPanel: sonnet`, `panelFix: sonnet`. Roster `light`.

**This ledger was written at finish rather than incrementally during the implementation run.** That
is a deviation from the contract, which expects a line per dispatch as each completes, and it is
recorded here rather than quietly corrected: the model attribution below is reconstructed from the
run's own dispatch records in the same session, not from memory of a previous one. Every dispatch
named its model explicitly at dispatch time, so no entry is a guess — but a ledger written after the
fact is weaker evidence than one written as it went, and the next run of this change should not
treat this file as proof that the incremental discipline was followed.

## Implementation dispatches

| Task | Outcome | Commit | Review | Model |
|------|---------|--------|--------|-------|
| 1 | complete | `493167a` | combined, clean on pass 1 | sonnet |
| 2 | complete | `ef4e79a` | combined, clean on pass 1 | sonnet |
| 3 | complete | `760eef8` | combined, clean on pass 1 | sonnet |
| 4 | complete | `86dd87a` | combined, 2 findings, fixed, re-reviewed clean | sonnet |
| 5 | complete | `e5ace17` | combined, clean on pass 1 | sonnet |
| 6 | complete | `8723e60` | combined, 1 Critical, fixed, re-reviewed | sonnet |

Each task was dispatched to its own implementer, one in flight against the worktree at a time, per
**4. Execute (SDD + TDD)** (`skills/myflow-do/SKILL.md`). Task 4's and task 6's implementers were
resumed rather than re-dispatched for their fix rounds, so the model is unchanged for those.

## Per-task review dispatches

One combined spec-plus-quality reviewer per task, which is what the `light` roster specifies.

| Reviewing | Model |
|-----------|-------|
| Task 1 | sonnet |
| Task 2 | sonnet |
| Task 3 | sonnet |
| Task 4, and its re-review after the fix | sonnet |
| Task 5 | sonnet |
| Task 6, and its two re-reviews after fixes | sonnet |

## Review panel dispatches

Six passes, three required slots each — Primary, Principles, Code review (low) — all on the panel's
recorded model. No slot was dispatched by `subagent_type`, because the `light` roster's three
required slots are all general-purpose or skill-invoking; no entry therefore reads
`unknown (agent-defined)`.

| Pass | Slots | Model |
|------|-------|-------|
| 1 | Primary · Principles · Code review (low) | sonnet |
| 2 | Primary · Principles · Code review (low) | sonnet |
| 3 | Primary · Principles · Code review (low) | sonnet |
| 4 | Primary · Principles · Code review (low) | sonnet |
| 5 | Primary · Principles · Code review (low) | sonnet |
| 6 | Primary · Principles · Code review (low) | sonnet |

All four optional slots (Security, Adversarial, Lens B, Lens C) had their triggers fire on every
pass and were **declined by the operator**, recorded distinctly from a trigger that never fired.

## Panel fix dispatches

| Round | Findings addressed | Model |
|-------|--------------------|-------|
| 1 | F1–F4 (pass 1) | sonnet |
| 2 | F5 (pass 2) | sonnet |
| 3 | F6 — operator decision to remove the runtime guard | sonnet |
| 4 | F7–F15 (pass 3) | sonnet |
| 5 | F16–F18 (pass 4), fixed structurally rather than per-variable | sonnet |
| 6 | F19–F24 (pass 5) | sonnet |
| 7 | F25–F27 (pass 6), bounded by operator decision | sonnet |

`models.panelFix` is recorded as `sonnet` for this change, which is `/myflow-fast`'s speed-favouring
default rather than the pipeline-wide default of the strongest available model. That is the recorded
value, so it governs, and it is named here so the choice is auditable rather than inferred.
