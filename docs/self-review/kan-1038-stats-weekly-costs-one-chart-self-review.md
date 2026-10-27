# Self-review report for kan-1038-stats-weekly-costs-one-chart

**Deferred:** reasoning pass run on Opus 5.5 from /private/tmp/claude-501/-Users-tweety53-Projects-agents/c96e0ccd-56f2-4b3d-94c9-3299350a571f/scratchpad/bundle.md

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them. The operator waived the rating and the flow-eval mining asks.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** Base red before Task 1: `pipeline-fix.sh verify` never ran the guard package's repository-wide script audits for a touched script, so a pipe into `grep -q` landed main red — fixed: 21c3ea70
- **[flow-fix]** The dispatch docs said to launch first and record `begin` after, contradicting the begin-row hook, which denied six panel-0 launches — fixed: b9468d01
- **[flow-fix]** `flow stage`, `state`, `settings` and `jira` followed the workspace-isolated `FLOW_ADDR` and fell to the local journal — fixed: 767b65e1
- **[flow-fix]** The visual-verify landing gate counted a verifier row the parent backfilled inline — fixed: d28b11e3
- **[flow-fix]** Verify-load flakes (a vitest timeout and the full-app reviewers capture taken before its data loaded) were re-run rather than traced — fixed: 9ff4d08b
- **[flow-fix]** Integrate and self-review asked scope and cost questions that carried a recommended option; the operator ruled such asks are auto-resolved — fixed: 1bf076d2
- **[flow-fix]** The landing HEAD's UI change had no real visual verifier and the archived visual record was stale — fixed: f8782292
- **[flow-fix]** The operator's "no review" was read as skipping the self-review pass; asked once and clarified — noted
- **[flow-fix]** verify-fix-3 used an unset `$SD` in a hand-built runner, which `verify-and-handoff.md` already forbids — noted
- **[flow-fix]** The commit hook denied two fix-run commits for a missing `Flow-Key:` trailer; the hook worked as designed — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Six denied panel-0 launches re-sent ~500-line slot prompts — fixed: b9468d01
- **[flow-cost]** The guard package was tested twice per verify — fixed: 40c5101d
- **[flow-cost]** Seven parent-inline verify rows took about 57 minutes; the remaining cost is the test list's stated design — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The mutation slot found five real Important test gaps, each killed with a recorded `fix-mutation:` line — noted
- **[flow-improvement]** `check-rerun-citations.sh` caught an off-hunk citation, closed by one send-back in 35s — noted
- **[flow-improvement]** The begin-row hook stopped six unrecorded dispatches, keeping the ledger complete — noted
- **[flow-improvement]** Verify caught the crash the panel missed, and in-run fix 1 landed RED→GREEN in 17s — noted
- **[flow-improvement]** `check-dev-stack-fresh.sh` confirmed the stack before a re-run was trusted — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** `run-reproducer` did not notice a reproducer that left the worktree changed — fixed: 0191dc73
- **[flow-automation]** `capture` and `verify` ran against a possibly stale UI-test stack — fixed: 5cf55d85

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** The verify rows `run-verify-list.sh` begins were never settled, so they read `not measured`; the harvester now settles them, and this change's rows were restamped — fixed: 0cb95394

## What can be sped up — `flow-speed`

- **[flow-speed]** `internal/guard` tests ran twice in every verify's test group — fixed: 40c5101d
- **[flow-speed]** verify and visual-verify stay serial in the creating run until the load flakes are traced — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case — Task 5 (in-run fix 1), a series with no measured week crashed the chart; operator waived mining — noted
  - Commit: `2da61e41`; code `stats/web/src/views/WeeklyCosts.tsx:185`
- **[flow-eval]** review case — F8, mutation survivor in the line-break logic, whose fix introduced the crash above; operator waived mining — noted
  - Commit: bad `80e7f26d`, fixed `7ab4ca34`
- **[flow-eval]** review case — F1, the prototype bullet contradicted the unattended planner modes; operator waived mining — noted
  - Commit: bad `782188282d47`, fixed `7ab4ca34`

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** The fix-round re-review was never asked for defects the fix diff itself introduces — fixed: 04b01f38
- **[flow-prompt-tuning]** The approved prototype never became a comparable visual-verify frame — fixed: 21b5d015
- **[flow-prompt-tuning]** An in-run visual fix task could not update the visual record it made stale — fixed: 663bde0e
