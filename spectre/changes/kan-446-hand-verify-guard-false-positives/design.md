## Context

KAN-446: reproduce KAN-423's hand verification of a guard false positive as a default habit —
whenever a gate guard's verdict conflicts with the pipeline's own structural conventions, the run
shows the operator how to verify by hand rather than leaving the verdict to be trusted blindly or
overridden blindly. One change because the habit is one statement consumed at four gate sites that
already share their prompt shapes. Constraints:

- Non-repetition: the habit is stated once, canonical, and cited at the call sites; per-guard
  procedures live in the guard headers, which this repository already treats as canonical for
  guard semantics (the finish contracts cite headers rather than copying them).
- No behavior change: verdicts, exit codes, verdict vocabularies, option sets and the
  false-positive recording machinery stay exactly as they are; the scripts gain header comments
  only, the skills and contracts prose only.
- Both finish runs must see the habit without cross-loading contracts: run 1 loads
  `finish-contract-run1.md` and run 2 loads `finish-contract-run2.md`, never each other's.
- `check-contract-budget.sh` headroom: `archive.md` has 279 bytes of it, so its citation line is
  worded tight; every other touched file has kilobytes.

## Decisions

### Guidance at the prompt, not an automatic pre-verification

**ID:** hand-verify-at-the-prompt
**Status:** active
**Chosen:** on a firing verdict the run relays the guard's hand-verification procedure alongside
the breakdown; the operator verifies before choosing a course — the operator's explicit choice at
the design gate.
**Considered:** hand-verifying automatically whenever the structural-conflict shape is
recognized, before the prompt — heavier, inserts a verification step into every gate, and takes
the verification out of the operator's hands; rejected by the operator.

### The habit's canonical home is pipeline.md

**ID:** habit-canonical-in-pipeline
**Status:** active
**Chosen:** a **Hand-verifying a guard verdict** section in `skills/flow-contracts/pipeline.md`,
beside **Guard resolution** — the one contract file every `/flow` run loads first, so run 1 and
run 2 both carry it with no cross-loading.
**Considered:** `finish-contract-run1.md` — run 2/archive would need run 1's contract loaded,
against the split-by-run design; restating the habit at each call site — violates non-repetition.

### Per-guard procedures live in the guard headers

**ID:** procedures-in-guard-headers
**Status:** active
**Chosen:** each guard's own header carries its hand-verification paragraph, named so a call site
can cite it — the repo's standing convention that a guard's header is canonical for its
semantics.
**Considered:** listing all four procedures in the pipeline.md section — a second copy of
guard-shaped facts that drifts the moment a guard's signals change.

### The four gate guards, not every guard

**ID:** gate-guards-only
**Status:** active
**Chosen:** `check-finish-preflight.sh` (`REFUSE`), `check-unfinished-work.sh` (`OUTSTANDING`),
`check-base-moved.sh` (`MOVED`) and `check-cleanup-complete.sh` (`LEFTOVER`) — the guards whose
verdicts an operator is asked to act on. Operator confirmed this scope at the design gate.
**Considered:** the panel-internal guards (`check-panel-*`) — their findings reach the operator
through the panel record and its own gates, not a guard-verdict prompt, so no call site exists
for the guidance to land in.

## Open questions
