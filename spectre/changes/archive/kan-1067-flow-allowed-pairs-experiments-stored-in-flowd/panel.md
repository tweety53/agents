# Review panel — kan-1067-flow-allowed-pairs-experiments-stored-in-flowd

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | primary | Important | stats/internal/store/migrations/0055_task_count_corrections.sql:5 | Every task-count row recorded before 0055, and every change with no row at all, reads as `corrections = 0`, and ExperimentStats COALESCEs a missing row to 0. |   |
| F2 | 0 | call-efficiency | Minor | stats/internal/store/experimentstats.go:157 | ExperimentStats loads every experiment via ListExperiments to find one id. |   |
| F3 | 0 | mutation | Important | stats/internal/store/taskcounts.go:31 | `in.Corrections < 0` -> `< -5`; -1 corrections never tested, so negative-corrections rejection (ErrInvalidTaskCount) is unguarded. |   |
| F4 | 1 | mutation | Important | stats/internal/store/taskcounts.go:58 | no test lists a legacy NULL-corrections row, so removing COALESCE (scan of NULL into int fails every list/board read for pre-0055 changes) passes the suite. |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 none — cold path over a bounded list
finding-reproducer: F3 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/1-mutation-1.sh

## Pass log

### Round 0

- base: MOVED with overlap (skills/flow/implement.md, stats/web/src/styles.css) — auto-resolved Continue; integrate sync settles it
- diff size 4603, under cap; docs-only exit 1 (.gitignore) — resolved roster runs; roster: compact — 21; standards: CLAUDE.md, AGENTS.md; no addition this round — the resolved list ran alone
- dispatched: panel-0-primary+principles on opus/low; panel-0-test-sanity+call-efficiency on sonnet/low
- dispatched: panel-0-mutation+failure-modes on sonnet/low
- verdict: fix (0 Critical, 2 Important)

### Round 1

- panel-fix-1: F1, F3 (Important) and F2 (Minor) to one fix chunk on the fixer pair opus/low
- round 1 re-runs every pass-1 role over fix-round-1.diff (panel-fix-1 plus full-suite-fix-1 delta) on rerun_dispatch sonnet/low
- dispatched: panel-1-primary+principles on sonnet/low over ef89575cf4b4f8a022325123151ff0e1f24b13cf; panel-1-test-sanity+call-efficiency on sonnet/low over ef89575cf4b4f8a022325123151ff0e1f24b13cf
- dispatched: panel-1-mutation+failure-modes on sonnet/low over ef89575cf4b4f8a022325123151ff0e1f24b13cf
- mutation slot could not mutate TestRolesRoute: the throwaway copy lacks the make-copied internal/modelpolicy/embedded/model-policy.md, so internal/api does not build there
- verdict: fix (0 Critical, 1 Important)
fix-mutation: stats/internal/store/experimentstats.go — nil corrections coerced back to 0 — TestExperimentStatsPlannerFamily
fix-mutation: stats/internal/store/taskcounts.go — in.Corrections < 0 flipped to < -1 — TestRecordTaskCountRefusesNegativeCorrections
fix-mutation: stats/internal/store/experiments.go — id filter id = $2 flipped to id <> 0 — TestExperimentStats*
fix-mutations-total: 3

### Round 2

- round 2 re-runs mutation alone over fix-round-2.diff (F4); every other role clean at round 1 and the delta is test-only
- dispatched: panel-2-mutation on sonnet/low over 2cae9c78e161495f9f675fe3b70033997e5b2468
- verdict: clean
fix-mutation: stats/internal/store/taskcounts.go — COALESCE(t.corrections, 0) replaced with t.corrections — TestListTaskCountsLegacyNullCorrections
fix-mutations-total: 1

### Round 3

- late-fix reduction: 6 changed lines since 0a476dc9
- dispatched: panel-3-primary on sonnet/low over 0a476dc94cbc7eba151c80533c8475fc8bfff840; panel-3-test-sanity+call-efficiency on sonnet/low over 0a476dc94cbc7eba151c80533c8475fc8bfff840; panel-3-mutation+failure-modes on sonnet/low over 0a476dc94cbc7eba151c80533c8475fc8bfff840
- failure-modes: slot ran mutation only; parent-written note — prose-only delta, no failure mode
- verdict: clean

### Round 4

- late-fix reduction: Task 15 since bd15eb5597920e39dce69f70518220e14d2e8bda
- dispatched: panel-4-primary on sonnet/low over bd15eb5597920e39dce69f70518220e14d2e8bda; panel-4-test-sanity+call-efficiency on sonnet/low over bd15eb5597920e39dce69f70518220e14d2e8bda; panel-4-mutation+failure-modes on sonnet/low over bd15eb5597920e39dce69f70518220e14d2e8bda
- verdict: clean

### Round 5

- late-fix reduction: Task 16 since 47887f6c
- dispatched: panel-5-primary on sonnet/low over 47887f6c2f9882ea232b5f09ebb1b15df971790b; panel-5-test-sanity+call-efficiency on sonnet/low over 47887f6c2f9882ea232b5f09ebb1b15df971790b; panel-5-mutation+failure-modes on sonnet/low over 47887f6c2f9882ea232b5f09ebb1b15df971790b
- mutation#1 (Experiments.tsx:150 null guard) traced equivalent: mean-after is drawn only when after is non-null, and with before null the unguarded chain still yields --delta-up for any after >= 0
- verdict: clean

### Round 6

- late-fix reduction: Task 17 since 17ed1e1b
- dispatched: panel-6-primary on sonnet/low over 17ed1e1befaed2e4d9e8e7a256e2465f7bc1778f; panel-6-test-sanity+call-efficiency on sonnet/low over 17ed1e1befaed2e4d9e8e7a256e2465f7bc1778f; panel-6-mutation+failure-modes on sonnet/low over 17ed1e1befaed2e4d9e8e7a256e2465f7bc1778f
- verdict: clean
