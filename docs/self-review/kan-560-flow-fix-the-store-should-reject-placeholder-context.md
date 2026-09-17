# Self-review context bundle for kan-560-flow-fix-the-store-should-reject-placeholder

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-560-flow-fix-the-store-should-reject-placeholder/tasks.md (absent)
skipped: spectre/changes/archive/kan-560-flow-fix-the-store-should-reject-placeholder/design.md (absent)
skipped: spectre/changes/archive/kan-560-flow-fix-the-store-should-reject-placeholder/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-560-flow-fix-the-store-should-reject-placeholder.md

# SDD ledger — kan-560-flow-fix-the-store-should-reject-placeholder

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: not recorded
- Started: 2026-09-17T22:52:50Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T23:16:51Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T23:16:51Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-560-flow-fix-the-store-should-reject-placeholder-panel.md

# Review panel — kan-560-flow-fix-the-store-should-reject-placeholder

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/.superpowers/sdd/kan-560-flow-fix-the-store-should-reject-placeholder/tasks.md:34,50 | tasks 2 and 3 **Tests:** fields name tests the diff never wrote — neither planned name exists anywhere under stats/, while the planned coverage exists under the names the diff actually wrote; a task field that no longer reflects the diff must be reconciled |   |
| F2 | primary | Minor | stats/cmd/flow/record.go:863 | the documented literal -agent-id none is translated to absent on dispatch begin but sent verbatim on dispatch end, where the new validation refuses it with exit 1 — the same documented word means absence on one half of one command pair and refusal on the other |   |
| F3 | primary | Minor | stats/cmd/flow/record.go:41 | the CLI mirrors none of the new agent-id shape rule before contacting the store, against its own closed-set precedent, so an offline -agent-id pending exits 0, journals, and is silently retired at the next reconcile |   |
| F4 | principles | Minor | stats/cmd/flow/record.go:863 | the knowledge that the literal none means absence lives only in the begin handler while the store rule refuses the same documented word on end — one vocabulary, two meanings, with no recorded reason for the difference |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 deferred the store is the rule's single validator and reconcile retires an offline placeholder before it can land
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-1.sh

## Pass log

### Round 0

- diff-size: 275 lines, under cap
- docs-only: exit 1 — roster unchanged (first non-doc path stats/internal/api/records_test.go); dispatching compact floor bundle primary+principles

### Round 1

- fix round 1 — FIX_BASE db5d9eb1d4dec48a459329e9fa76548d45037cfc; agents-ran: none (execution inline — the parent fixes); why: one Important (F1 stale plan test names) plus two Minors sharing one root (F2/F4 the none literal on dispatch end); F3 deferred
- fix-diff: .superpowers/sdd/fix-round-1.diff (508e3a4, stats/cmd/flow/record.go + record_test.go); F1's fix lives in gitignored tasks.md + repaired reproducer 0-primary-1.sh, disclosed here; all three dispatched reproducers flipped to defect not demonstrated
- round 1 entry: base MOVED no overlap (continues); cap measured from held sha db5d9eb; docs-only exit 1 — roster unchanged; re-running primary and principles alone on the rerun pair (mapped glm-5.3-flash/high)
fix-mutation: stats/cmd/flow/record.go — removed the end handler's none-to-absence mapping — TestDispatchEndAcceptsAgentID/none
fix-mutations-total: 1

## Branch log

commit 508e3a4f7de8fc2a359fb87228a56d1722164cc8
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 02:12:37 2026 +0300

    fix(cli): map the documented -agent-id none to absence on dispatch end

 stats/cmd/flow/record.go      |  9 +++++++++
 stats/cmd/flow/record_test.go | 12 ++++++++++++
 2 files changed, 21 insertions(+)

commit db5d9eb1d4dec48a459329e9fa76548d45037cfc
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 01:49:43 2026 +0300

    feat(reconcile): retire journal entries refused for an invalid agent id

 stats/internal/reconcile/reconcile.go   |  9 ++++++++
 stats/internal/reconcile/record_test.go | 41 +++++++++++++++++++++++++++++++++
 2 files changed, 50 insertions(+)

commit 6c19eba39f0cef2df9649f525d07cd7789092235
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 01:48:01 2026 +0300

    feat(api): answer 400 for an invalid agent id on a record write

 stats/internal/api/records_test.go | 18 ++++++++++++++++++
 stats/internal/api/server.go       |  2 ++
 2 files changed, 20 insertions(+)

commit 81519348c7dddd1e4724115f2552a721dd3a8378
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 01:46:22 2026 +0300

    feat(store): reject placeholder and malformed agent ids at write time

 stats/internal/store/records.go      |  75 ++++++++++++++++++++
 stats/internal/store/records_test.go | 130 +++++++++++++++++++++++++++++++++++
 2 files changed, 205 insertions(+)

## Session narrative

This run implemented KAN-560 inline in one session: the record store now refuses placeholder and
malformed agent ids at write time (closed case-insensitive placeholder word set plus an ASCII
charset rule in validateAgentID, called from RecordDispatch before the seq-retry loop and from
EndDispatch before its UPDATE, behind a new typed ErrAgentIDInvalid), the API maps the sentinel to
400 in mapStoreError, and reconcile retires journalled writes refused this way. TDD held throughout
— every surface got its failing test first, and the store tests run against the real flow-postgres
stack. Where the run struggled: the panel round 0 raised one Important (the plan's **Tests:** fields
named tests the diff renamed; fixed by reconciling tasks.md) and three Minors sharing one root —
the CLI's documented `-agent-id none` literal meant absence on dispatch begin but rode into the new
store refusal on dispatch end; fixed by mirroring the begin handler's mapping into the end handler
(commit 508e3a4), mutation-proved by deleting the mapping and watching the new none subtest fail,
with F3 (CLI-side mirroring of the whole shape rule) deliberately deferred out-of-scope because the
store is the rule's single validator and reconcile retires an offline placeholder before it can
land. Two operational stumbles, both recovered and disclosed: F1's reproducer hardcoded the stale
plan names and could never flip, so the parent repaired it to implement its own documented contract
(parsing tasks.md), and the operator of the marks — this session itself — opened flow.verify while
flow.review-panel was still open, which the store correctly records as superseding the open run; the
superseded attempt-1 rows for review-panel and verify in the stage ledger are that marking error,
not gaps in the work, and both stages carry completed attempt-2 pairs.
