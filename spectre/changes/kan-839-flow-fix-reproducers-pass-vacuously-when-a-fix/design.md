# kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix — design

## Context

A reproducer's checks can enumerate nothing after a fix renames their target and still exit 0 —
the post-fix verification re-run has no premise audit (the exit-contract guard skips non-open
findings by design; its KAN-606 audit covers only the defect-present side). The change gives
reproducers premise declarations, asserted by the script and audited at dispatch, so a renamed
premise lands in the KAN-524 ambiguity refusal instead of a green count. Why now, in
`proposal.md`.

## Mechanism

- **Premise declaration.** Every runnable, non-mutation reproducer carries, in the same
  first-10-lines window as `# demonstrates:`, one `# premise: <path>:<line>:<content>` line per
  file, test/class name or `tasks.md` task id its checks read. Same citation shape and
  resolution machinery as demonstrates; semantics differ: demonstrates names where the defect
  **was** (resolves only against the defect-present tree, pre-fix), a premise names what the
  checks **read** (must resolve in every tree the script runs in).
- **Enforcement point 1 — the script asserts.** Authoring rule: the body asserts every declared
  premise before its real checks run; a missing premise is a loud failure naming it on stderr,
  exiting non-zero — never exit 0. The rule rides every slot's dispatch prompt the way the cwd
  contract does.
- **Enforcement point 2 — the guard audits at dispatch.** `pcAudit`
  (`stats/internal/guard/panelexitcontract.go`) resolves premise citations with the same checks
  demonstrates gets — shape, lexical + resolved containment, file, line, content. Tolerant:
  absence of premise lines never violates; declared-but-unresolvable joins exit 1's violation
  classes.
- **Post-fix effect — no runner change.** A fix renaming a premise makes the verification
  re-run's script fail loudly (non-zero) → the runner reads "demonstrated" → identical to the
  dispatch-time verdict → the KAN-524 ambiguity refusal (exit 2) fires → the re-author path.
- **Coverage and stated limit.** ≥1 premise line mandatory via the authoring rule for every
  newly authored or repaired reproducer; the guard stays tolerant. Residual: a post-change
  reproducer whose author skips premises entirely passes the guard silently — caught by panel
  re-runs and self-review, not mechanically. Accepted.
- **Mutation exemption.** Mutation-declared reproducers declare no premises; their instrument
  audit is the KAN-568 sha pin.

## Files touched

- `skills/flow/review-panel.md` — authoring rule, dispatch-prompt carry, re-run note, the
  guard-invocation section's audit description.
- `stats/internal/guard/panelexitcontract.go` + `check_panel_reproducer_exit_contract_test.go` — premise label in
  the audit, tolerant mode, tests; the shim `check-panel-reproducer-exit-contract.sh` header's
  instrument-audit paragraph.
- Untouched: `runreproducer.go` (the runner), `prove-reproducer.sh`,
  `check-panel-reproducers.sh` (lexical guard).

## Decisions

### Premise enforcement split: script asserts, guard audits at dispatch

**ID:** premise-enforcement-split
**Status:** active
**Chosen:** the reproducer's own body asserts its premises before its checks, and `pcAudit`
resolves premise citations at dispatch time — two points, no runner change; the post-fix rename
is covered by the existing KAN-524 ambiguity refusal.
**Considered:** runner-side audit before every exec — single choke point and clearest verdicts,
but duplicates the audit the script performs and touches `runreproducer.go` and its tests;
post-fix re-run audit only — smallest diff, but the resolution decision lands in the parent's
attention, the pattern this repo replaces with guards.

### One citation form for premises

**ID:** premise-citation-form
**Status:** active
**Chosen:** `# premise: <path>:<line>:<content>` — the exact demonstrates shape and resolution
machinery; a file, a test/class name and a task id all cite their declaration site uniformly.
**Considered:** split `# premise-file:` + citation forms — lighter to author for the common
case, but two shapes to validate and two violation classes to maintain.

### Coverage: mandatory by authoring rule, tolerant in the guard

**ID:** premise-coverage-tolerant-guard
**Status:** active
**Chosen:** every newly authored or repaired runnable non-mutation reproducer carries at least
one premise line; the guard validates only what is declared, so records predating this change
are never re-bounced.
**Considered:** strict from this change on — every open record's reproducer must carry premises
immediately, re-bouncing unrelated changes' panels mid-flight. The accepted residual is stated
under Mechanism.

### Mutation reproducers exempt

**ID:** premise-mutation-exempt
**Status:** active
**Chosen:** mutation-declared reproducers declare no premises; the KAN-568 sha pin is their
instrument audit.
**Considered:** demanding host-tree premises of a reproducer that builds its own mutated tree —
noise for exactly the convention KAN-568 added.

## Open questions

None.
