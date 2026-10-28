# Self-review report for kan-1046-review-dynamics-dashboard

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** Re-ending a closed dispatch row overwrote its `ended_at`. That inflated task 11 to 27m53s and panel-fix-4 to 16m55s, against a real ~3m and ~1m. A closed row now keeps its end, and rows 8415 and 8419 were retimed — fixed: 6560ecb7
- **[flow-fix]** The floor for an inline back-fill ignored the visual-verify verdict the fix answers, so task 11's row started 7m50s before that verdict. The floor now includes the change's `visual-verify*` rows, and 8415's start was repaired — fixed: ad4b7a05
- **[flow-fix]** A panel-fix back-fill got no floor while its round's reviewer rows were still open, so panel-fix-4 started before round 3 began. That back-fill is now refused until the round's reviewers close, and 8419's start was repaired — fixed: c8a0555b
- **[flow-fix]** Fix-mutation rows were accepted after the re-review that contradicted them, and three round-1 rows claimed flips nobody ran. `flow record mutation` now refuses once the round's re-review has begun. The new `flowd drop-mutation` removed rows 1076, 1077 and 1079 — fixed: 41a51c79
- **[flow-fix]** The planning session left `explore-1` open and the run closed it with a guessed end (22m33s against a real 1m36s). The `flow.brainstorm` end is now refused while a dispatch row is open, naming each row's session token, and 8371 was retimed — fixed: f607e76b
- **[flow-fix]** The appended in-run fix task skipped the plan guards, and `close-task` did not run `check-plan-shape`. That cost 2 verify re-runs and a fixup, autosquash and force-push. `close-task` now runs the plan guard on appended and in-run fix tasks — fixed: fbb84fe1
- **[flow-fix]** The plan-shape refusal of a `Task 6, Task 7` **After:** line already names the accepted form — disproved — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Inline verify rows carry zero tokens, which is correct: the parent wrote nothing during those windows — noted
- **[flow-cost]** Task 11's and panel-fix-4's inflated cost was a record artefact, repaired under the `flow-fix` items above — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The mid-run model-version requirement was absorbed as three appended tasks, with a re-published prototype check, and no re-plan — noted
- **[flow-improvement]** The live check on a store copy caught nothing wrong in the product. Its one disagreement traced to the hand query, which proved the cohort definition — noted
- **[flow-improvement]** The mutation slot found real gaps in every round, including the nested reviewer-prompt path defect — noted
- **[flow-improvement]** The visual verifier caught every prototype departure in one pass, so the fix round closed in one scoped and one full run — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** The design's live check spelled out `createdb -T flow`, which fails while flowd is connected, instead of naming `scripts/live-store-copy.sh`. The planner's live-check paragraph now names a project-declared stand-in tool — fixed: 81e31465

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** The mutation slot is the panel's critical path (17m38s and 15m18s), and by design its launch waits for the full suite — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case — Task 2, the ReviewDynamics aggregate; not mined (operator paused mining) — noted
- **[flow-eval]** review case — F18, busiest-first ordering unpinned by the fixture; not mined — noted
- **[flow-eval]** review case — F3, the double fetch on first open; not mined — noted
- **[flow-eval]** visual case — the first-pass review-dynamics frame against the prototype; not mined — noted

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** A test of a choice is fed inputs on which the alternatives agree — implementer — docs/prompt-tuning/kan-1046-review-dynamics-dashboard-1/ — noted
- **[flow-prompt-tuning]** No late-round misses: every round-1+ finding was on a hunk new in that round — noted
- **[flow-prompt-tuning]** The planner's measure-before-commit rule for capture tasks did not cover a prototype-specified view, which cost visual-verify fix round 1. The rule now covers it — fixed: 52495851
- **[flow-prompt-tuning]** The closing full verifier's `--update-snapshots=all` rewrote unrelated baselines with anti-aliasing noise. Chromium rasterisation is now pinned in the Playwright config and every baseline was recaptured. Determinism was proven: two repeat runs left identical bytes, and unpinned runs differ — fixed: 1fdc8c1f
