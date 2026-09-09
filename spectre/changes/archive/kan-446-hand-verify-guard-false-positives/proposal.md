# kan-446-hand-verify-guard-false-positives

## Why

A gate guard's verdict is computed from files and git state, but the situation it judges can be
structurally innocent: KAN-423's frontend worktree reported `OUTSTANDING` while its canonical plan
was 19/19 ticked and the findings store answered `[]` from either worktree — the plan lived behind
the cross-repo resolution, not in the worktree the guard was pointed at. That run did the right
thing by hand (recomputed both signals, confirmed the verdict was a structural false positive,
proceeded) as an ad hoc habit nothing in the pipeline teaches. Trusted blindly, a false positive
blocks verified-complete work or teaches the operator to click past the gate; overridden blindly,
it silences a guard that may have seen real unfinished work.

## What changes

A default habit, encoded where verdicts are consumed: when a gate guard's verdict fires
(`OUTSTANDING`, `MOVED`, `LEFTOVER`, `REFUSE`) while the situation contradicts the pipeline's own
structural conventions, the run relays **how to hand-verify that verdict** alongside the
breakdown, and the operator verifies before choosing a course. The habit's canonical statement
lands in `skills/flow-contracts/pipeline.md` (beside **Guard resolution** — the one contract file
every `/flow` run loads first); each gate guard's verification procedure lands in that guard's own
script header; the call sites — `integrate.md`'s preflight, unfinished-work and base-moved
prompts, `archive.md` run 2's cleanup stop, and the finish-contract mirrors — cite both. Prose and
header comments only: no guard behavior, exit code, verdict vocabulary, option set or store
schema changes.
