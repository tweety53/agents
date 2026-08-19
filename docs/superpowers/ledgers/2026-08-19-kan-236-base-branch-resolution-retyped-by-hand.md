# SDD ledger — kan-236-base-branch-resolution-retyped-by-hand

Every subagent dispatch this change made, with the model it ran on. Recorded per **Model policy**
(`skills/myflow-contracts/pipeline.md`). No slot was dispatched by `subagent_type`, so no entry
reads `unknown (agent-defined)` — every dispatch named its model explicitly.

**Recorded roles for this change:** `models.implementation: sonnet`, `models.reviewPanel: sonnet`,
`models.panelFix: sonnet`, all three being `/myflow-fast`'s recorded defaults on a creating run.
**No operator override was given for any role**, so every dispatch below ran on the recorded value.

**This ledger was written at integration time, not incrementally during the run** — a gap in this
run's own execution, recorded here rather than papered over. The per-dispatch model was tracked in
the dispatches themselves and in `final-review-panel.md` throughout; this file reconstructs the
ledger from them.

## Implementation

| Task | Dispatch | Model | Outcome |
|---|---|---|---|
| 1 + 2 | implementer (one squash unit — task 1 is `Build: red`, `Squash-with: Task 2`) | sonnet | complete (commit `6f212ee` after all rebases), review clean, review: combined |
| 1 + 2 | same implementer, resumed for the per-task review's `LC_ALL=C` finding | sonnet | complete, folded by `--fixup` + `--autosquash` |
| 3 | implementer | sonnet | complete (commit `6d00330`), review clean, review: combined |
| 4 | implementer | sonnet | complete (commit `0b306dd`), review clean, review: combined |
| 5 | implementer | sonnet | complete (commit `b0886d3`), review clean, review: combined |
| 6 | implementer | sonnet | complete (commit `99c42f8`), review clean, review: combined |

Every implementer dispatch carried the six required blocks: COMMIT-PER-TASK, the TDD sub-skill, the
systematic-debugging sub-skill, REQUIRED READING (`engineering-principles.md`), the CONTEXT BUNDLE
paragraph, and PLAN PROVENANCE.

## Per-task review

One combined reviewer per task, per the `light` roster's per-task shape.

| Task | Model | Result |
|---|---|---|
| 1 + 2 | sonnet | one Major — the locale-collation defect in the name validation; fixed |
| 3 | sonnet | clean |
| 4 | sonnet | clean |
| 5 | sonnet | clean |
| 6 | sonnet | clean |

## Review panel

Roster `light`: Primary, Principles (Merged lens), Code review (low). Four optional slots fired
triggers and all four were declined by the operator. Six passes in total — pass 1 plus four full
re-runs after fix rounds, each escalated automatically rather than chosen.

| Pass | Slots dispatched | Model | Findings |
|---|---|---|---|
| 1 | Primary, Principles, Code review (low) | sonnet | F1, F2, F3, F4 |
| full re-run 1 | all three | sonnet | F5, F6, F7 |
| full re-run 2 | all three | sonnet | F8, F9 |
| full re-run 3 | all three | sonnet | F10 (raised by two slots, deduped), F11, F12 |
| full re-run 4 | all three | sonnet | F13 |
| closing pass | all three | sonnet | F14 |

## Fix rounds

| Round | Findings | Model | Folded into |
|---|---|---|---|
| 1 | F3, F4 | sonnet | tasks 1-2 and task 4 |
| 2 | F5, F6, F7 | sonnet | task 4 |
| 3 | F8 | sonnet | task 4 |
| 4 | F10, F11, F12 | sonnet | tasks 1-2, task 4, task 5 |
| 5 | F13 | sonnet | tasks 1-2 |

F2 and the planning-artifact reconciliations were made by the parent rather than dispatched, since
no implementer may touch `openspec/`.

## Outcome

Fourteen findings: eleven fixed, three withdrawn by the operator (F1, F9, F14). The closing pass was
clean for Primary and Code review, with Principles' only open item withdrawn. **Three of the fourteen
were one defect class** — a consumer of the resolver reading the wrong ref (F3, F7, F11) — all found
by review rather than by the design.
