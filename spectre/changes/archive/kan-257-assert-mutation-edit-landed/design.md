## Context

KAN-257 asks that a mutation which never applied stop being reported as a surviving mutant.
`scripts/mutate-and-verify.sh` (+ `scripts/test-mutate-and-verify.sh`) already mechanizes the
assert half (refuses a non-applying patch, exit 2) and the restore half (byte-identical verify,
exit 3); what it cannot fix is rule text that still tells callers silent hand mutation is fine.
The normative text lives in `skills/flow/review-panel.md` — the review panel contract — which is
also the file `scripts/check-dispatch-paragraphs.sh` guards and `scripts/check-contract-budget.sh`
budgets, so both join the edit. One change: one contract's wording, its guard, its harness, its
budget. The design itself:
`docs/superpowers/specs/2026-09-09-kan-257-assert-mutation-edit-landed-design.md`.

## Decisions

### Scope: normative-only, no new script

**ID:** normative-only-scope
**Status:** active
**Chosen:** edit `skills/flow/review-panel.md`'s two mutation passages plus the guard/harness/budget
that hold the wording in place — the ticket's assert mechanics already exist in
`mutate-and-verify.sh`; the gap is that the rules never demand them.
**Considered:** the ticket's `mutate.sh <file> <old> <new> <harness…>` string front-end — a second
tool for a need the patch-based script already covers, and the assert obligation binds hand paths
anyway; closing the change as already-done — rejected, the phantom-survivor hazard at both rule
sites stays open.

### Hand mutation stays allowed, with an assert obligation

**ID:** hand-mutation-allowed-with-assert
**Status:** active
**Chosen:** scratch-tree reverts and single-value flips remain permitted paths; any mechanism —
script or hand — must confirm the edit landed before the tests run, and a never-applied edit is a
refusal to redo, never a surviving mutant.
**Considered:** making the script mandatory — rejected: the MUTATION PROOF paragraph deliberately
leaves *which* mechanism to mutate to the caller's judgment, and a scratch tree is sometimes the
right tool; what the ticket demands is the assert, not the tool.

### A never-applied mutation is a refusal, not a finding

**ID:** never-applied-is-refusal
**Status:** active
**Chosen:** refusal semantics — the mutation is redone in-round with a working mechanism; it is
never recorded as a surviving mutant and never buys a test.
**Considered:** recording it `unverifiable` like the reproducer-refusal path — rejected: that
record is for mechanisms that cannot run at all, while a non-applied edit is immediately redoable,
so downgrading it would trade a fixable silence for a permanent caveat.

## Open questions

None — the design gate answered scope and hand-mutation policy; nothing was deferred.
