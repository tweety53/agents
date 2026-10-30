# kan-1060-flow-cost-visual-verify-now-costs-more-than — session narrative

## 2026-10-30 — creating run

Implementation resumed from STARTED with the plan ready; all 15 tasks ran under SDD with gated
reviews, the panel's round 0 raised F1–F5 (failure-modes, mutation) and one panel-fix round closed
them, the round-1 re-run clean.

**Task 15's full app suite** went red on the base itself: `ui-test-up` lacked its `policy-copy`
prerequisite and the interactive spec's `Main` locator became ambiguous after a base commit. Both
were fixed on the branch (`Flow-Key: full-suite-fix-1..3`) and the baselines re-recorded; task 15's
collateral was widened by a dated Correction paragraph.

**Inline verify** failed twice before going green: task 12's Files and Tests fields missed
`visual_verify_pass_close_test.go` and its exits-template test, and one run hit a flaky
`TestEnforceLoopDispatchHookRecordsADenial`.

**Three in-run pipeline fixes** landed on main mid-run:
- `3deff9fa` — the loop hook's fixed 10 s `record_incident` timeout killed the incident stub under
  `-race` load; hook tests now lift the hook's wall-clock bounds (`FLOW_HOOK_UNBOUNDED`). Not
  ETXTBSY, the first hypothesis, which the planner disproved by forcing the timeout.
- `81ca7e1b` — hook tests inherited `FLOW_RECORDS_ADDR` from the verify workspace and read the live
  flowd's open allowed-pairs experiment overrides, so `TestEnforceAgentBaselineHookRefusesOffPolicyDispatch`
  failed only inside a `/flow` verify; `runHook` now pins both store addresses dead.
- `dbb6b4f1` — main's Playwright suite was red after the state-board perf commits: Project chips
  lost every all-finished project, finished changes with an open stage dropped off the open poll,
  `ui-test-up` missed `policy-copy`, and specs/baselines were stale. Its review moved the new
  open-poll filter to an `EXISTS` so 70594ba2's speed-up holds.

**Base sync mid-verify.** The visual gate's verify row had no HEAD recorded (`verify-3`), so verify
re-ran; picking up the pipeline fixes needed the branch rebased onto main twice. The second rebase
conflicted only on baseline PNGs this change had re-recorded against the old base; main's side was
taken and the full suite then passed 76/76 unchanged, so the branch's own re-records of non-leaderboard
screens were dropped. Restoring main's PNGs before main itself was fixed was tried first and
abandoned: main's own baselines were stale.

**Live check** ran on a clone against a copy of the store on 4175. The scratch clone sat on the
change's branch, which the main-checkout hook treats as the default branch and refused writes on;
switching the clone to a local `livecheck` branch cleared it. The design's unstamped live token was
refused and a stamped one used. The first `close` refused an incomplete report (a capture on no view
line) and chained a rest pass — the guard working as designed. Main share read 33.4% / $311.90
against the 34% / $332 baseline.

**Visual verification** ran through the change's own `visual-verify-pass.sh` (prepare → one opus
verifier → close) rather than the installed pre-change contract; `--app` had to be written
`"${W}:…"` because zsh's `:s` modifier mangled `$W:stats`.
