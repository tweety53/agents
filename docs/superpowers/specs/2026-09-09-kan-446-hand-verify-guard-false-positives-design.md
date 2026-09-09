# Hand-verifying guard false positives rather than trusting them blindly

Design for `kan-446-hand-verify-guard-false-positives` (KAN-446), approved at the design gate
2026-09-09. `spectre/changes/kan-446-hand-verify-guard-false-positives/design.md` adapts this into
the change's operative decisions.

## Problem

KAN-423's frontend worktree reported `OUTSTANDING` from `check-unfinished-work.sh` while its
canonical plan was 19/19 ticked and the findings store answered `[]` from either worktree — a
structural false positive (a cross-repo change whose plan resolves only in the canonical tree).
The run hand-verified both signals before proceeding, but nothing in the pipeline teaches that
habit: a false verdict could be trusted blindly (blocking verified-complete work, or teaching the
operator to click past the gate) or overridden blindly (silencing a guard that saw real unfinished
work).

## The habit

When a gate guard's verdict fires (`OUTSTANDING`, `MOVED`, `LEFTOVER`, `REFUSE`) while the
situation contradicts the pipeline's own structural conventions, the run relays how to hand-verify
that verdict alongside the breakdown, and the operator verifies before choosing a course. Hand
verification recomputes the guard's own signals from the primary records; it never rewrites the
verdict line or exit code and never bypasses the courses the gate offers. A verdict confirmed
structural is recorded where the call site provides a recording (today: `flow record verdict
false-positive` at the unfinished-work gate).

## Placement

- Canonical habit statement: new **Hand-verifying a guard verdict** section in
  `skills/flow-contracts/pipeline.md`, beside **Guard resolution** — loaded first by every
  `/flow` run, so run 1 and run 2 both carry it without cross-loading contracts.
- Per-guard procedures: each guard's own script header (`scripts/check-unfinished-work.sh`,
  `scripts/check-base-moved.sh`, `scripts/check-cleanup-complete.sh`,
  `scripts/check-finish-preflight.sh`) — headers are canonical for guard semantics.
- Call sites cite both: `integrate.md` (preflight `REFUSE`, unfinished-work `OUTSTANDING`,
  base-moved `MOVED`), `archive.md` step 7 and `finish-contract-run2.md` step 7 (`LEFTOVER`),
  `finish-contract-run1.md` (courses table, base-moved section).

## Scope decisions

- Guidance at the prompt (operator's explicit choice) — not an automatic pre-verification step.
- The four gate guards only; panel-internal guards are out of scope (no operator gate).
- No behavior change: guard code, exit codes, vocabularies, option sets, store schema untouched.
