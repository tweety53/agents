# pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-16

## Why

b8e0e063 added to `skills/flow-contracts/operator-prompts.md` a section making every operator
prompt's explanation the turn's final message, with the ask in a later turn. The operator corrected
(2026-10-16) that the split applies only to `/flow-self-review`'s flow-eval candidate mining ask
(angle 7); every other flow question — plan gate, landing, convergence, filing — explains in prose
and fires `AskUserQuestion` in the same turn, as before. Pipeline-fix 15 narrowed the section
(**Explain before asking**) but left a self-review-only rule in the shared prompt contract, with
the filing ask in `skills/flow/unfinished-work-gate.md` still citing it.

## What changes

- **Explain before asking** leaves `skills/flow-contracts/operator-prompts.md`; the contract
  returns to its pre-bb273948 text.
- `skills/flow-self-review/SKILL.md` step 4 carries the mining-ask turn split, its rejected
  alternatives and its silence citation, and its finding-explanation sentences return to their
  pre-bb273948 text.
- `skills/flow/unfinished-work-gate.md`'s **Every filing ask explains before it asks** returns to
  its pre-bb273948 text: the message body explains before the prompt fires, same turn.
- `verbatim-moves.txt` acknowledges every sentence removed or restored.
