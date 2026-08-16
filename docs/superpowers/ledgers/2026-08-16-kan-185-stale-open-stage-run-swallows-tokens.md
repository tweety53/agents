# SDD ledger — kan-185-stale-open-stage-run-swallows-tokens

Command: `/myflow-fast` (creating run). Harness: claude-code. Session token: `mf-k185-9q2`.
Recorded models (myflow-fast defaults): implementation `sonnet`, review panel `sonnet`, panel fixes
`sonnet`. Review panel roster: `light`.

Worktree: `/Users/tweety53/Projects/agents/.worktrees/kan-185-stale-open-stage-run-swallows-tokens` ·
branch `openspec/kan-185-stale-open-stage-run-swallows-tokens` · merge base `b85d4f5` · workspace id
`kan-185-190e`.

Dispatch bundles, from `scripts/plan-dispatch-bundles.sh`: bundle 1 = task 1, bundle 2 = task 2 —
one implementer in flight per worktree, task 2 dispatched only after task 1's commit sha came back.

## Dispatches

Task 1: complete (commit `6a9330d`, review clean first time, model: sonnet, review: combined).
`bestWindow` now orders candidates on `(StartedAt, Attempt)` rather than on `Attempt` alone;
`attribute_test.go` went from 16 to 18 top-level tests, matching the plan's predicted baseline.
The reviewer independently re-ran both new tests against the pre-change `attribute.go` and confirmed
`TestOverlappingOpenWindowsPreferTheLatestStarted` fails there, while
`TestSameInstantWindowsFallBackToHighestAttempt` passes both before and after — it pins the delta
spec's same-instant scenario rather than discriminating this diff, which the reviewer reported
rather than leaving implicit.

`scripts/check-task-commit-fields.sh` first exited 2 — it resolves the plan from
`<worktree>/openspec/changes/*/tasks.md`, and the planning artifacts had been written in the main
checkout, where `/myflow-start`'s sections create them. The parent moved the change directory and
the design doc into the worktree (a move, not a copy, so no second divergent copy exists) and
re-ran the guard, which then passed with two advisory skips: `Regression:` and `Baseline:` cannot be
verified because this project's `## test` command is not a single targetable command.

Task 2: complete (commit `29affa4`, model: sonnet, review: combined). Original commit `6d7ac85`;
the reviewer found **F1 — the delta spec's fourth scenario, "A superseded stage is not counted as
rework", was pinned by no test**: `ReworkRate`'s `COUNT(*) FILTER (WHERE sr.outcome = 'abandoned')`
satisfies it structurally, but nothing would have objected had that filter widened. The **parent
repaired the plan** rather than sending the task back against an undeclared file — task 2 gained a
step 6, `stats/internal/store/aggregate_test.go` in its `**Files:**`, the new test in `**Tests:**`,
and `before=20 after=21` for that file in `**Baseline:**` — and the same implementer added
`TestReworkRateDoesNotCountASupersededRunAsAbandoned` as a `--fixup` folded into `6d7ac85` by
`git rebase --autosquash`, which is what rewrote the commit to `29affa4`. The reviewer re-verified
independently, running its own mutation test rather than accepting the implementer's, and confirmed
`aggregate.go` itself is unmodified and `stageruns.go`/`stageruns_test.go` are byte-identical across
the rewrite. F1 fixed; task 2 clean.

`scripts/check-task-commit-fields.sh` passed for both the original and the rewritten commit, with
the same two advisory skips as task 1.

## Panel — pass 1

Roster `light`. Required slots: Primary (slot 0), Principles (slot 2, Merged lens), Code review at
effort low (slot 3) — all three on sonnet, each dispatched separately, each carrying the
per-finding reproducer requirement.

`scripts/check-panel-diff-size.sh` measured **400** changed lines against a cap of **2000** — exit
0, under cap, no operator question asked.

Optional slots: four triggers fired against `final-review.diff` — Security (query construction),
Adversarial (behaviour change to code with existing tests, and >~300 changed lines), Lens B (>~200
changed lines) and Lens C (error handling on the new transaction's failure paths). Under the `light`
preset they went to the operator as one multi-select, and the operator selected **none**: all four
are recorded **declined**, distinctly from a slot whose trigger never fired.

Task 3: complete (commit `00cd04b`, review clean first time, model: sonnet, review: combined).
Migration `0009_stage_run_open_session_token.sql` plus `TestSupersedeIndexExists`; embedded
migrations 8 → 9, and `TestMigrationsAreIdempotent` absorbed the new file unedited because it reads
`EmbeddedMigrationCount()` rather than a literal. The task exists only because the operator answered
panel finding F3 by choosing to index rather than withdraw.

The reviewer measured the index rather than reasoning about it: with `0009` applied the supersede
UPDATE runs `Index Scan using stage_runs_open_session_token`, 1 shared buffer hit, 0.039ms; with the
migration removed from the same seeded database it falls back to `stage_runs_ended_at`, 203 buffer
hits and 200 rows removed by filter, 0.083ms. It also proved `TestSupersedeIndexExists` non-vacuous
twice — once by removing the migration, once by renaming the index inside it — and restored the tree
byte-identically afterwards.

The same dispatch carried task 2's step 8 as a fixup: four references to the pre-rename
`insertStageRun` in `stats/internal/harvest/watcher.go` and `watcher_test.go`, which fix round 2 had
correctly refused to touch as undeclared files. Task 2 became `8c0e95c`.

## Panel — pass 2 and fix round 4

Pass 2 ran all three required slots against the rewritten diff, escalated to Full automatically:
the fixes had altered a delta spec and added a migration, and three fix rounds had already run. The
four conditional slots stayed declined — the operator's pass-1 answer holds for the change.

Every pass-1 finding was re-verified as genuinely fixed by the slot that raised it, each doing its
own measurement rather than reading the record: a fresh `EXPLAIN (ANALYZE, BUFFERS)` for the index,
a fresh mutation test for the boundary test, a repo-wide grep for the old function name.

Pass 2 raised seven new findings, reconciled into one namespace as F6–F12 (all three slots had been
told to number from F6, and all three did, so they collided with each other — a defect in how the
dispatch was written, recorded in the panel record rather than quietly renumbered).

Two were real defects in the mechanism task 2 introduced, and both were demonstrated by hand
against a live database rather than argued: begins for one session token do not serialise under
READ COMMITTED, so two of them can each miss the other's uncommitted insert and both stay open;
and a late end mark could overwrite a run a supersede had already closed, because
`ApplyEndStageMark` looks up and closes in two separate calls while `EndStage` matched on `id`
alone. Those became **task 4** (commit `2e5d83d`), with an advisory lock and a guarded `EndStage`,
and two new decisions in `design.md`.

Fix round 4 dispatched one fix subagent on sonnet for all seven. The three doc-and-test findings
folded into their own task commits as fixups, rewriting the branch to `4bc5597`, `537f05e`,
`0869b30`, and task 4 landed on top as `2e5d83d`. The fixer reported one collateral change it had
to make outside the declared files — `internal/sweep/sweep_test.go`, whose race test pinned the
silent-overwrite behaviour `EndStage`'s new guard replaces — and the parent declared that file in
task 4's `**Files:**` rather than letting an undeclared file ride along. Read against the diff, the
rewritten assertion is **stronger** than the one it replaced: it used to accept either outcome
unconditionally, and now requires the stored outcome to match whichever call reported winning.

The replay paragraph in `bestWindow`'s doc comment has now been corrected three times — the second
correction was one the parent directed, and it was itself wrong. What survives is the accurate
statement: the supersede closes a replayed attempt's predecessor at its own start instant, which
makes that window empty, and `Window.contains` never matches an empty interval — so the
same-instant tie-break is defensive rather than the ordinary replay path.
