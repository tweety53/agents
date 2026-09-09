# kan-257-assert-mutation-edit-landed — design (approved at the gate, 2026-09-09)

Classification: bounded — the mutation-proving flow exists in this repo to read; the change edits
its contract text and the guard that holds it. Canonical decisions with IDs:
`spectre/changes/kan-257-assert-mutation-edit-landed/design.md`. This file records the design
itself; it does not restate the decisions.

## The gap (what the gate approved fixing)

`scripts/mutate-and-verify.sh` already mechanizes KAN-257's assert half (refuses a non-applying
patch, exit 2) and restore half (byte-identical verify, exit 3). Two passages in
`skills/flow/review-panel.md` still permit silent hand mutation, and nothing in the file says
"assert the edit landed":

1. § The mutation-testing brief — Mutation and Bugbot slots mutate with no asserted mechanism; a
   slot whose edit no-ops files phantom surviving-mutant findings.
2. The fix round's MUTATION PROOF paragraph — offers the script but also permits "revert it in a
   scratch tree, or flip the single value" with no assert obligation, and cites the script as
   `<agents repo>/scripts/mutate-and-verify.sh`, a repository-root-relative path Guard resolution
   forbids for shipped guards.

## The design — four edits, no new script

1. **The mutation-testing brief** (`skills/flow/review-panel.md`): add the assert obligation —
   every mutation confirms its edit landed before the tests run (target found where intended, edit
   visible); a mutation that never applied is a **refusal** — redo it with a working mechanism —
   never a surviving mutant, and it never buys a test.
2. **The MUTATION PROOF paragraph** (same file): carry the same obligation onto its permitted
   hand paths (scratch-tree revert, single-value flip); fix the script citation to basename form
   (`mutate-and-verify.sh`), keeping the paragraph's judgment-leaves-the-mechanism posture intact.
3. **`scripts/check-dispatch-paragraphs.sh`**: extend the existing MUTATION PROOF entry (added
   by KAN-464) from three shared phrases to six — the three new assert phrases held as short
   literals, per the guard's existing shape — so the new wording cannot silently disappear;
   `scripts/test-check-dispatch-paragraphs.sh` gains one drop-case per new phrase.
4. **`scripts/check-contract-budget.sh`**: raise the `skills/flow/review-panel.md` budget row if
   the additions trip it — the guard's own documented correct response to a genuine addition.

## Testing

- `scripts/test-check-dispatch-paragraphs.sh` (new cases for the MUTATION PROOF row).
- The `## lint` guard list that reads the touched files (`check-vocabulary.sh`,
  `check-references.sh`, `check-contract-budget.sh`, `check-dispatch-paragraphs.sh`,
  `check-installed-citations.sh`), plus a `check-normative-inventory.sh` before/after inventory
  diff for the bulk normative edit — every differing sentence resolved by intent, and the new
  assert sentences expected in the after-inventory.

## Gate record

Asked at the confirm (batched, with the design summary from the seeded research note): scope —
**normative-only** (wrapper script and close-as-done declined); hand mutation — **allowed with
assert** (script-mandatory declined); convergence — **nothing unclear, approve and move on**.
Refusal semantics for never-applied mutations were planner-judged and stood uncorrected.
