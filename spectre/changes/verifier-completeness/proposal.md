# verifier-completeness

## Why

Operator, 2026-10-02, after gymie KAN-870's visual verification ended partly done twice: "Fix all, it must not happen in the future." Three causes:

- The verifier was dispatched on Sonnet (a global browser-driving rule overrode the stage's fixed model in practice) at effort `low`, and both verifiers stopped after the tests and capture, listing steps 9–11 under "Not done".
- The change's substance was motion (exits, slides, drag gestures, entrance removal), and the procedure only compared still captures against mockups; recording frames was a soft ask in the parent's brief.
- The parent accepted a report whose "Not done" listed required steps, and "the second report is final; never a third verifier on the same HEAD" then forbade finishing.

Scope addition, same day: "In general the top priority is the least actions from my side. Less stops. Number of re-review and re visual verify shouldn't be limited."

## What changes

- `skills/flow/visual-verify.md`: the verifier stays on opus at effort `low` (`flow-low`) and is never dispatched on another model whatever any other rule says about browser-driving agents. Every verifier report is checked by the new `check-verify-report.sh` guard before it is read; an incomplete report is never accepted — its remainder is dispatched under the same key suffixed `-rest` (then `-rest-<n>`), counting toward no re-dispatch and no fix round, and repeated until complete. The parent names the change's motions in the dispatch (from `design.md`, `proposal.md` and the diff); a named motion with no strip or entry blocks. Non-zero-exit re-dispatches are no longer capped at one.
- `skills/flow/visual-verify-verifier.md`: a required motion step records frames (CDP screencast or rapid sampling, reduced motion off) for each named motion under `<changeRoot>/visual-verification/motion/` and judges each against the no-blink / same-frame rule; `visual-verification.md` carries an entry per motion; the report carries `motion:` and `steps:` lines.
- `check-verify-report` (Go, `stats/internal/guard/verifyreport.go`, shim `scripts/check-verify-report.sh`): reads a report's `steps:` line and the motion count, exit 0 complete, 1 incomplete, 2 cannot answer.
- `skills/flow-contracts/pipeline.md` **Fewest operator actions**: the operator's actions and the run's stops are minimised; no round count ends a re-review or re-verification; a round without progress changes approach (a fresh fixer on opus/high with every earlier report, a different reproduction) instead of stopping or asking.
- Caps removed: `verify-fix-loop.md`'s five-round cap, `review-panel-fix-round.md`'s "two automatic rounds" escalation, and the auto-resolution contract's exception that pointed at the cap.
