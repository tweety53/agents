# Self-review report for kan-1039-end-to-end-autonomy-rule

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them. The operator waived the rating and the flow-eval mining asks.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** In-run pipeline fix from the planning session: `flow record dispatch` refused every planning dispatch a creating `/flow` made before `spectre new` — fixed: 35be9d63
- **[flow-fix]** The panel and implement docs still said to record `begin` rows after the launch message, but the hook denies a launch that has no row yet. The panel's first launch was refused — fixed: 20949fce
- **[flow-fix]** The main-checkout hook refused any command with a backtick in it, even one inside single quotes. That made it read a worktree heredoc body as shell and deny the command — fixed: e2960c4d
- **[flow-fix]** `pipeline-fix.sh verify` did not run the guard tests that scan `scripts/` when a script was touched, so 12ceca09 landed main red on `TestNoPipeIntoEarlyExitGrep` — fixed: bcd707b4
- **[flow-fix]** The pass-1 full suite ran at the same time as the mutation slot's `go test` runs, and the guard package hit the 10m timeout. Alone it passed in 332s — fixed: 2bd1928a
- **[flow-fix]** Integrate `commit` stopped because a fix made during planning had no `In-run pipeline fix:` line. The line can be derived from its store row — fixed: 2b236f4b
- **[flow-fix]** A missing fix-mutation row was caught only at the landing gate, 13 minutes after the panel closed — fixed: 1df8394c
- **[flow-fix]** Main fixed the same base-red grep line in 453d9145, and the rebase dropped this run's identical copy — noted
- **[flow-fix]** The F4 rewording failed `check-verbatim-moves.sh` at verify. That gate worked as designed — noted
- **[flow-fix]** The `probe-26101` dispatch row is aborted, 13s and no tokens. It is harmless — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The `full-suite-fix-1` back-fill had no floor at the round's reviewer ends, so it claimed 40m and 4.16M cache-read tokens. Row 8196 was retimed and restamped — fixed: 43aadd7c
- **[flow-cost]** Task 7's zero tokens are real: no transcript records fall in its 18s window — noted
- **[flow-cost]** The mutation slot's 28m is its own cost; the avoidable part is the cached sweep under flow-speed — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The plan's literal blocks were verified at plan time and applied by a script that refuses unless each target appears exactly once, so all 8 tasks applied without drift — noted
- **[flow-improvement]** The mutation slot caught the only Important (the zcode carve-out had no test), and it was killed with a recorded fix-mutation — noted
- **[flow-improvement]** Integrate ran straight on from `IN_PROGRESS` with no operator turn — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** No suite runtimes had been recorded since 2026-10-09; `run-verify-list.sh` now records each declared suite — fixed: 0a4f0234

## What can be sped up — `flow-speed`

- **[flow-speed]** The mutation brief did not warn that Go's test cache misses files a test's subprocess reads, so the slot ran and threw away a whole cached sweep — fixed: 8a2f632c
- **[flow-speed]** Verify's 10-minute re-run of the full test list was needed: the fixes touched Go and the skills its tests read — noted
- **[flow-speed]** The guard package's runtime is now visible through 0a4f0234, and its contention through 2bd1928a — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1 Important, the zcode carve-out in `hooks/enforce-agent-baseline.py:333` had no test; operator waived mining — noted
  - Commit: bad `4eeb47aa`, fixed `3384201b`
- **[flow-eval]** impl case — Task 8, the xhigh refusal in `hooks/enforce-agent-baseline.py:331-339`; operator waived mining — noted

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

_none — this angle produced no findings._
