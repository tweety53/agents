# kan-257-assert-mutation-edit-landed

## Why

KAN-257 (from kan-202's self-review, myflow-automation angle): a mutation whose edit silently
does not apply produces exactly the same observable result as a surviving mutant — the suite
passes, the caller records a surviving mutant, and a redundant test gets written for behaviour
that was already covered. The two conclusions are opposite and nothing distinguished them. The
ticket's proposed helper now exists (`scripts/mutate-and-verify.sh`, which refuses a non-applying
patch, reports caught/survived, and verifies the restore), but the two places the *rules* talk
about mutation still permit silent hand mutation:

- the review panel's mutation-testing brief (Mutation and Bugbot slots), and
- the fix round's MUTATION PROOF paragraph, which offers the script but also permits "revert it
  in a scratch tree, or flip the single value" with no assert obligation.

The paragraph also cites the script as `<agents repo>/scripts/mutate-and-verify.sh` — a
repository-root-relative path, which Guard resolution (`skills/flow-contracts/pipeline.md`)
forbids for shipped guards.

## What changes

Normative-only, no new script:

1. `skills/flow/review-panel.md` § The mutation-testing brief — every mutation must confirm its
   edit landed (target found where intended, edit visible before the tests run); a mutation that
   never applied is a **refusal** — redo it with a working mechanism — never a surviving mutant,
   and it never buys a test.
2. The same obligation carried onto the MUTATION PROOF paragraph's permitted hand paths
   (scratch-tree revert, single-value flip), and the script citation fixed to basename form per
   Guard resolution.
3. `scripts/check-dispatch-paragraphs.sh` — the MUTATION PROOF paragraph is already a
   required-paragraph table entry (KAN-464); this change extends that entry's shared phrases from
   three to six so the new wording cannot silently disappear; `scripts/test-check-dispatch-paragraphs.sh`
   gains the matching drop-cases.
4. `scripts/check-contract-budget.sh`'s budget row for `skills/flow/review-panel.md` raised if
   the addition trips it — the correct response to a genuine addition, per `.flow/project.md`.

Operator decisions (design gate, 2026-09-09): scope normative-only (wrapper script and
close-as-done rejected); hand mutation stays allowed with the assert obligation (script-mandatory
rejected).

## Added during implementation

- 2026-09-09 — verify stage blocked on two pre-existing bare `spectre/changes/` citations
  (`skills/flow/archive.md`, `skills/flow-contracts/finish-contract-run2.md`); fixed inline on the
  operator's instruction (commit `463ada4`) with the budget-row raise its growth required.
