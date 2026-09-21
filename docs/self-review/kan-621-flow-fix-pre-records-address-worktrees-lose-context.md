# Self-review context bundle for kan-621-flow-fix-pre-records-address-worktrees-lose

found: 4 of 7 sources; skipped: 3 of 7 sources
skipped: spectre/changes/archive/kan-621-flow-fix-pre-records-address-worktrees-lose/tasks.md (absent)
skipped: spectre/changes/archive/kan-621-flow-fix-pre-records-address-worktrees-lose/design.md (absent)
skipped: spectre/changes/archive/kan-621-flow-fix-pre-records-address-worktrees-lose/narrative.md (absent)

## change summary

## Summary

kan-621-flow-fix-pre-records-address-worktrees-lose fixes KAN-621: self-review bundles for changes whose dispatch and finding rows were written to a pre-records-addr apply worktree's own database (removed at cleanup) read `found: 0 of 6` with no word as to why. The durable fix is in the bundle assembler: when the store shows a completed `flow.review-panel` stage run for the change but holds no dispatch rows, the bundle now leads its notes with a loud RECORDS LOSS line naming the loss and its most plausible cause, instead of plain `skipped (absent)` lines that read as no panel ever having run.

### What changed, grouped by area

- stats/internal/store — new `Store.StageCompleted(ctx, projectKey, change, stage)`: an EXISTS read over `stage_runs` joined to `changes`, matching the named stage and the completed outcome; an unknown change is `(false, nil)`, never an error.
- stats/internal/selfreview — `Bundle` takes a `panelRan bool`; when it holds and the run carries no dispatch rows, a loud `note: RECORDS LOSS` line leads the bundle's notes. The bundle's shape is decided here, where it is produced.
- stats/internal/api — the bundle handler's store interface grows the `StageCompleted` read (also on `RecordStore`); a failing read is a 5xx, never a degraded bundle; the handler passes the boolean into the assembler.
- stats/internal/stages + stats/cmd/flow — the completed outcome value has one home, `stages.OutcomeCompleted`, written by the CLI's end mark and read by the store read alike.
- The sweep the issue asked for ran in-session before implementation: no per-workspace `flow_*` database survives anywhere, so no records were recoverable; one live change (kan-543, STARTED) carries a stale `.flow/project.md` without the `FLOW_RECORDS_ADDR` row and is named here for the operator — its branch needs a rebase onto main (or the leftover worktree removal) before any further apply run there. No code was written for the sweep: its population is empty and a permanent script for a one-time sweep would be speculative.

### What was verified and how

- Test-first per task: store read, bundle note, handler wiring each went red before green; final state: `go test ./internal/{store,selfreview,api,client,reconcile,web,stages}/ ./cmd/flow/ -race -count=1` all green against live flow-postgres.
- Full `## lint` list green in the worktree, including `go vet ./...`, `gofmt -l`, `tsc -b`, and every guard script.
- Review panel (compact: primary+principles) ran one round plus targeted fix-round re-runs: 6 findings (2 critical, 1 important, 3 minor), all fixed in one fix commit and confirmed fixed by both slots; `check-panel-findings-closed.sh` passes.

### Deliberately left out

- No backfill or migration of lost rows: none survive to recover.
- No permanent sweep script: the sweep found an empty population; the operator-facing outcome (kan-543's stale table) is recorded in this summary and the Jira issue.
- No documentation edits in `skills/` or contracts: the bundle's shape is decided where it is produced, and adding a note needs no contract change.
## .superpowers/sdd/ledgers/kan-621-flow-fix-pre-records-address-worktrees-lose.md

# SDD ledger — kan-621-flow-fix-pre-records-address-worktrees-lose

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T19:44:10Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T20:08:50Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T20:08:50Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-621-flow-fix-pre-records-address-worktrees-lose-panel.md

# Review panel — kan-621-flow-fix-pre-records-address-worktrees-lose

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | critical | stats/internal/api/records.go:83 | the new RecordStore.StageCompleted breaks the test build of three untouched packages: stubStageStore (internal/client/client_test.go:1016), nopRecordStore and *fakeRecordStore (internal/reconcile/record_test.go:151, 259) and fakeStore (internal/web/embed_test.go:141) never gained the method; go vet ./... and go test ./... fail to compile internal/client, internal/reconcile and internal/web, so every task's Build-green is false as merged |   |
| F2 | primary | important | .superpowers/sdd/kan-621-flow-fix-pre-records-address-worktrees-lose/tasks.md | tasks.md task 3's Files field no longer reflects the diff: records.go, records_test.go and changes_test.go were touched, none is named, and the unnamed RecordStore widening is exactly what carried the F1 breakage |   |
| F3 | primary | minor | stats/internal/selfreview/git.go:411 | git.go's comment says the records-loss note leads the notes; the code appends it after the unreadable-repo notes |   |
| F4 | principles | critical | stats/internal/api/records.go:83 | hard invariant: the lint/test gate cannot exit clean — the same compile breakage F1 cites under the project's lint list and the global fix-first lint rule |   |
| F5 | principles | minor | stats/internal/store/stageruns.go:777 | Single Source of Truth: the outcome value completed has two unconnected homes, the new SQL reader (stageruns.go) and the CLI writer (cmd/flow/stage.go:786); a drift silences the records-loss predicate with no guard |   |
| F6 | principles | minor | stats/internal/selfreview/git.go:411 | Least astonishment: the note-order comment contradicts the behavior, raised again under this pass's angle |   |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 cd stats && go vet ./...
finding-reproducer: F2 /Users/tweety53/Projects/agents/.worktrees/kan-621-flow-fix-pre-records-address-worktrees-lose/.superpowers/sdd/reproducers/0-primary-2-plan-files.sh
finding-reproducer: F3 /Users/tweety53/Projects/agents/.worktrees/kan-621-flow-fix-pre-records-address-worktrees-lose/.superpowers/sdd/reproducers/0-primary-3-note-order.sh
finding-reproducer: F4 cd stats && go vet ./...
finding-reproducer: F5 /Users/tweety53/Projects/agents/.worktrees/kan-621-flow-fix-pre-records-address-worktrees-lose/.superpowers/sdd/reproducers/0-principles-1-outcome-literal.sh
finding-reproducer: F6 /Users/tweety53/Projects/agents/.worktrees/kan-621-flow-fix-pre-records-address-worktrees-lose/.superpowers/sdd/reproducers/0-principles-2-note-order.sh

## Pass log

### Round 0

- diff-size: 288 lines, cap 750 — under cap, proceeding
- docs-only: no — first non-documentation path stats/internal/api/changes_test.go; resolved roster primary+principles runs unchanged
- roster: compact — 62
- no addition this round — the resolved list ran alone

### Round 1

- fix round 1: F1-F6 fixed in one commit; both re-running slots confirmed all their findings fixed, no new defects at the recorded sites
## git log --stat

commit ef45fbeaa4e9945683ebb8d63efe7ce7d3816323
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 23:06:21 2026 +0300

    fix(stats): close the panel's round-0 findings
    
    F1/F4: stub StageCompleted on the four RecordStore fakes the widening
    left behind, so go vet ./... and the full test build compile again.
    F3/F6: the records-loss note now actually leads the notes, as its own
    comment states. F5: the completed outcome spelling has one home,
    stages.OutcomeCompleted, used by the CLI's end mark and the store read
    alike.

 stats/cmd/flow/stage.go                 |  2 +-
 stats/internal/client/client_test.go    |  4 ++++
 stats/internal/reconcile/record_test.go |  8 ++++++++
 stats/internal/selfreview/git.go        | 12 ++++++------
 stats/internal/stages/outcomes.go       | 11 +++++++++++
 stats/internal/store/stageruns.go       | 12 ++++++++----
 stats/internal/web/embed_test.go        |  4 ++++
 7 files changed, 42 insertions(+), 11 deletions(-)

commit 8d16e31bd9c8822088ee91bf096679d45f9dad69
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 22:41:34 2026 +0300

    feat(api): pass the panel-stage read into the bundle assembler

 stats/internal/api/changes_test.go    |  9 +++++++++
 stats/internal/api/records.go         |  7 +++++++
 stats/internal/api/records_test.go    |  8 ++++++++
 stats/internal/api/selfreview.go      | 19 ++++++++++++++++++-
 stats/internal/api/selfreview_test.go | 32 ++++++++++++++++++++++++++++++++
 5 files changed, 74 insertions(+), 1 deletion(-)

commit 16099ccb081f544eefb99848dc1b6c312158771b
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 22:37:55 2026 +0300

    feat(selfreview): name a records-source loss loudly in the bundle

 stats/internal/api/selfreview.go         |  2 +-
 stats/internal/selfreview/bundle_test.go | 82 ++++++++++++++++++++++++++------
 stats/internal/selfreview/git.go         | 17 ++++++-
 3 files changed, 84 insertions(+), 17 deletions(-)

commit cd5046f9f9004592d7b066a9b3f95d8c22c24c81
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 22:34:06 2026 +0300

    feat(store): read whether a change's named stage run completed

 stats/internal/store/stageruns.go      | 28 +++++++++++
 stats/internal/store/stageruns_test.go | 86 ++++++++++++++++++++++++++++++++++
 2 files changed, 114 insertions(+)

## Session narrative

This /flow-fast run implemented KAN-621 inline end to end: kickoff (worktree, Jira In Progress), brainstorm reading the bundle assembler (stats/internal/selfreview), the flow CLI's records-addr resolution, and the live store's evidence — kan-563 carried a full completed pipeline's stage marks with zero dispatch, finding and summary rows, and no per-workspace database survived anywhere, so the sweep the issue asked for had an empty salvage population and the durable fix became the loud RECORDS LOSS note. The dynamic decide rolled class small, execution inline, and a compact primary+principles panel. Implementation went test-first per task (store StageCompleted read, Bundle panelRan note, handler wiring); the rebase onto a moving main happened three times, each conflict-free. The panel's round 0 earned its keep: the primary slot caught that widening api.RecordStore had broken the test build of three packages the parent never compiled (go vet ./... — the parent had run only go build and targeted tests), plus a plan-drift finding and an outcome-literal SSOT finding; one fix commit closed all six findings, both slots re-ran targeted at their own findings and confirmed each fixed. The run struggled most with the panel contract's own mechanics — the finding writes were refused once for a missing -status, the plan-shape guard correctly rejected indented fields and then an unordered shared-file pair, and the fix-wave paths needed a fourth plan task and a regenerated fix-round diff after the last rebase — each resolved by following the guard rather than around it. Verification closed with the full lint list green and the scoped test set green against live Postgres.
