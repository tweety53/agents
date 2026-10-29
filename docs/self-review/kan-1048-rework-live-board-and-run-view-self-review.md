# Self-review report for kan-1048-rework-live-board-and-run-view

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** 499b813a's re-run grouping regex `^verify-(fix-)?[0-9]+` also matched a plain `verify-<repo>` key whose repository name starts with a digit, so a blocked `verify-3d-viewer` was never superseded. The number now ends at a dash or the key's end — fixed: 3d5ec9c3
- **[flow-fix]** 6557ccc1's pin covered only `run-verify-list`'s shim; every guard shim is now checked to name each sibling script its Go source execs — fixed: c8f1093b
- **[flow-fix]** 90a1f2ff waived the break-it proof as "cannot be forced in one isolated run"; it was reverted and 7b10e569 proved the flake does reproduce alone (1 in 8) — noted
- **[flow-fix]** Panel mutation reproducers never checked their edit landed: a failed edit carried on and read as a surviving mutant (F11). The runner now refuses a mutation reproducer that does not run under `set -e` — fixed: c073ad8a
- **[flow-fix]** check-plan-shape passed task 6's `Allowed-collateral` entry under `spectre/changes/<name>/visual-verification/**`, which task-commit-planning-paths then refused. F23 now refuses any Files or Allowed-collateral entry under a planning path other than `visual-verification.md`, and verify-fix-loop step 1 says where evidence is committed — fixed: d3f8a818. The archived kan-1048 plan's offending entry was dropped so the tree scan passes — fixed: 1f9d40cd
- **[flow-fix]** Panel round 0's end times were daemon stamps from before kan-1044's fix was running; the dev store was repaired with `flowd restamp-ends 2026-10-28T00:00:00Z` (`restamped: 14`), and flowd now restamps the last 7 days of stamped ends at startup — fixed: 43ce8d81
- **[flow-fix]** The panel slot named its reproducers `R-0-mutation-<n>.sh` and the parent renamed them; slot variance, caught by the guard — noted
- **[flow-fix]** The plan named a non-existent `FLOW_STORE_ADDR` (the hook reads `FLOW_ADDR`), alongside two other planning-accuracy Corrections; the implementer caught each — noted
- **[flow-fix]** The UI-test sweeper abandoned the seed's past-dated open runs; the change itself sets `FLOW_SWEEP_OFF=1` in `ui-test-up` — noted
- **[flow-fix]** A subagent's PreToolUse/PostToolUse under the parent's `session_id` could overwrite an open `question` activity. The session-activity hook now ignores tool events carrying `agent_id` — fixed: 3bcb5bb1
- **[flow-fix]** Self-review fix review Minors: `rrErrexit` sat inside `declaresMutation`'s doc comment, and planshape.go's finding count was stale — fixed: 9c8679b9

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Inline execution took about 50M of the run's ~62M cache-read tokens (tasks 3–5: 8.6M, 7.2M, 13.6M; panel-fix-1 7.5M; task 6 6.8M at effort low) — noted
- **[flow-cost]** The run-capped flake cost two blocked verifies; fixed on main in-run (7b10e569) — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Task 5's captures caught the half-open `to=now` bug tasks 3 and 4 shipped — noted
- **[flow-improvement]** Visual-verify caught a real prototype departure: Cost in period summed only the filtered rows — noted
- **[flow-improvement]** The sweeper's root cause was read from `stage_runs` rows (`outcome=abandoned`), not guessed — noted
- **[flow-improvement]** The live check used `live-store-copy.sh` with a PUT probe read back; cache-skipped verifies took ~52s; in-run tooling fixes landed with tests — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** Covered by the mutation-reproducer `set -e` refusal and flowd's startup restamp above — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** The startup restamp lives in flowd (43ce8d81) — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** The verify flake cost ~33 min; fixed (7b10e569) — noted
- **[flow-speed]** Panel round 0 took 26 min (suite 11 min before a 14.5 min mutation dispatch), by design per kan-1039's CPU-contention decision — noted
- **[flow-speed]** The one-frame tooltip false positive cost ~33 min; see `flow-prompt-tuning` — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1, the state board polling every 5s while the tab is hidden; not mined (operator paused mining) — noted
- **[flow-eval]** review case — F2, the run view fetching every board row of the project to read one; not mined — noted
- **[flow-eval]** impl case — Task 1, board fields and session activity in the store and API; not mined — noted

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** No case: F1 and F2 fail condition 3, the Cost-tile departure is covered by the mockup rule, and there were no late-round misses — noted
- **[flow-prompt-tuning]** Round 1 reported a one-frame tooltip lag from a single screencast frame that never reproduced under a real pointer; the verifier now confirms a one-frame motion departure with a next-frame DOM read or a second recording before reporting it — fixed: cef3cc2f
