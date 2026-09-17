# Self-review context bundle for kan-557-flow-stats-app-host-the-dispatch-finding-store

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-557-flow-stats-app-host-the-dispatch-finding-store/tasks.md (absent)
skipped: spectre/changes/archive/kan-557-flow-stats-app-host-the-dispatch-finding-store/design.md (absent)
skipped: spectre/changes/archive/kan-557-flow-stats-app-host-the-dispatch-finding-store/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-557-flow-stats-app-host-the-dispatch-finding-store.md

# SDD ledger — kan-557-flow-stats-app-host-the-dispatch-finding-store

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles+failure-modes
- Key: panel-0-primary+principles+failure-modes
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T21:15:47Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: fable effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T21:54:24Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: fable effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T21:54:24Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Slot: exp-failure-modes
- Key: panel-1-exp-failure-modes
- Model: fable effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T21:54:24Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-557-flow-stats-app-host-the-dispatch-finding-store-panel.md

# Review panel — kan-557-flow-stats-app-host-the-dispatch-finding-store

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Critical | stats/cmd/flow/record.go:553 | registerRecordConnFlags was re-pointed at resolveRecordsAddr, so flow hazard (3 verbs), flow spec (2), flow suite (2) and flow tasks count also resolve FLOW_RECORDS_ADDR — the plan and four prose statements scope them to FLOW_ADDR; which commands resolve the records address is one fact stated four ways, none matching the code |   |
| F2 | exp-failure-modes | Important | stats/cmd/flow/record.go:553 | concurrent re-entry: two apply worktrees interleave suite record rows into one (project, suite, host)-keyed series in the persistent store with no worktree or change dimension; the median-of-last-10 mixes different trees measurements — wrong but recoverable |   |
| F3 | primary | Minor | .superpowers/sdd/kan-557-flow-stats-app-host-the-dispatch-finding-store/tasks.md | tasks.md checkboxes unticked though every verify step passes, and task 1 Files field omits record_test.go and selfreview.go |   |
| F4 | principles | Minor | stats/cmd/flow/state.go | DRY: the explicit -addr Visit scan is duplicated in noteAddrEnvUsage and noteRecordAddrUsage |   |
| F5 | primary | Minor | stats/cmd/flow/record.go | tasks count and suite record resolve FLOW_RECORDS_ADDR with no stderr note naming it |   |
| F6 | exp-failure-modes | Minor | stats/internal/client | error return: wrong-server failures (missing Flow-Daemon header) on tasks count and suite record name neither the URL nor the deciding variable |   |
| F7 | exp-failure-modes | Minor | .flow/project.md | partial application: FLOW_RECORDS_ADDR exports only at worktree-create, so pre-existing worktrees keep the kan-468 loss silently |   |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 deferred pre-existing — wrong-server error text is the client package, untouched by this diff; the transport-failure half already names the URL
finding-status: F7 deferred pre-existing — the row exports at worktree creation via prepare-workspace.sh; worktrees created before this change predate the variable and are re-derived on their next run

reproducers-total: 7
finding-reproducer: F1 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-exp-failure-modes-1.sh
finding-reproducer: F3 none — plan-file hygiene: checkboxes unticked and task 1 Files omits record_test.go, selfreview.go
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-exp-failure-modes-2.sh
finding-reproducer: F7 none — environmental: the export happens when a workspace is prepared

## Pass log

### Round 0

- roster: compact — primary+principles; experimental slot failure-modes (roll 20) joined the dispatch
- panel diff measured 197, under cap
- docs-only check exit 1 — first non-documentation path stats/cmd/flow/record.go; resolved roster runs unchanged
- no addition this round — the resolved list ran alone
- slot primary+principles+failure-modes exceeded the 15-minute ceiling (~26 min) but returned complete before the dispatcher could stop it; closed completed, no re-dispatch

### Round 1

- fix: F1 narrowed the wiring — hazard/suite/spec/tasks moved to registerConnFlags (FLOW_ADDR alone), one value-derived noteAddrUsage replaced both note functions (F4), riders keep the run address and its note (F5), tasks.md checkboxes and Files completed (F3); F6/F7 stay deferred pre-existing
- agents-ran: parent inline; diff-path: .superpowers/sdd/fix-round-1.diff
- re-runs: primary F1 fixed; principles F1+F4 fixed; exp-failure-modes F2 fixed — reproducer exits 0 on every finding, no defect introduced by the fix diff at the reviewed sites

## Branch log

commit 180026b00f1da55cc9a309597c718ae141cb8921
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 00:09:18 2026 +0300

    docs(flow-config): host the record store at a non-isolated records address

 .flow/project.md | 14 +++++++++++++-
 1 file changed, 13 insertions(+), 1 deletion(-)

commit 33c42cc52e6ca99a9ef863b8b04c67d8247f4a13
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 00:07:29 2026 +0300

    feat(flow): resolve the record family's store address from FLOW_RECORDS_ADDR

 stats/cmd/flow/hazard.go      |   6 +-
 stats/cmd/flow/record.go      |  29 ++++++++-
 stats/cmd/flow/record_test.go |  14 +++++
 stats/cmd/flow/selfreview.go  |   2 +-
 stats/cmd/flow/settings.go    |   4 +-
 stats/cmd/flow/spec.go        |   4 +-
 stats/cmd/flow/stage.go       |   6 +-
 stats/cmd/flow/state.go       |  69 +++++++++++++++------
 stats/cmd/flow/state_test.go  | 135 ++++++++++++++++++++++++++++++++++++++++++
 stats/cmd/flow/suite.go       |   4 +-
 stats/cmd/flow/tasks.go       |   2 +-
 11 files changed, 240 insertions(+), 35 deletions(-)

## Session narrative

This flow-fast run gave the run record a store of its own: the flow CLI's record family and self-review bundle now resolve their store address from FLOW_RECORDS_ADDR (falling back to FLOW_ADDR, then the built-in default), and the project's workspace-isolation table declares that variable as a token-free url row so an apply worktree's dispatch and finding rows land in the persistent store and survive workspace removal — the storage-side remedy beside kan-552's archive-branch copies for the kan-468 incomplete-bundle symptom. The decide roll classified the change small inline with a compact panel plus the rolled experimental failure-modes slot; round 0 found the real defect — registerRecordConnFlags was shared by the hazard, suite, spec and tasks verbs, so four verbs silently changed address behind their own docs (F1 Critical) and suite rows would have interleaved across worktrees in the persistent series (F2 Important) — and the fix narrowed the wiring through a dedicated registerConnFlags for the riders and collapsed the two address-note functions into one value-derived noteAddrUsage (F4, F5), with the plan file completed (F3); two Minors were deferred as pre-existing (F6 client error text, F7 worktrees created before the export existed). All five runnable reproducers demonstrated pre-fix (the exit-contract guard verified against the stashed, pre-fix tree) and exited 0 post-fix; all three raising roles re-ran alone on the fix-round delta and confirmed every finding fixed with no new defect at the reviewed sites. The session struggled most with the usage-text backticks breaking record.go's raw string literal, the round-0 panel bundle exceeding its 15-minute ceiling while still returning complete work, and the fresh-worktree SPA embed needing make web-build before go vet could run; a transient store wedge mid-verify journalled two stage marks that journal flush replayed cleanly.
