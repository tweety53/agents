# Self-review context bundle for kan-904-flow-fix-a-declared-but-unasserted-premise-still

found: 2 of 7 sources; skipped: 5 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-904-flow-fix-a-declared-but-unasserted-premise-still, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-904-flow-fix-a-declared-but-unasserted-premise-still.md (absent)
skipped: spectre/changes/archive/kan-904-flow-fix-a-declared-but-unasserted-premise-still/tasks.md (absent)
skipped: spectre/changes/archive/kan-904-flow-fix-a-declared-but-unasserted-premise-still/design.md (absent)
skipped: spectre/changes/archive/kan-904-flow-fix-a-declared-but-unasserted-premise-still/narrative.md (absent)

## .superpowers/sdd/reviews/kan-904-flow-fix-a-declared-but-unasserted-premise-still-panel.md

# Review panel — kan-904-flow-fix-a-declared-but-unasserted-premise-still

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | stats/internal/guard/check_panel_reproducer_exit_contract_test.go:600 | case 43's output needle was weakened from "REPRODUCER-EXIT-CONTRACT-OK (1 runnable" to "REPRODUCER-EXIT-CONTRACT-OK" though KAN-904 changes nothing about that case (no premise line; OK-line format untouched) — the stronger needle demonstrably still passes, so the diff lost count-assertion strength for nothing; undocumented in tasks.md's Correction. Fix: restore (1 runnable. |   |
| F2 | principles | Minor | stats/internal/guard/check_panel_reproducer_exit_contract_test.go:600 | same line as F1, citing engineering-principles.md Testing principles (an assertion weakened rather than a bug fixed); same reproducer. |   |

findings-total: 2
finding-status: F1 withdrawn misattributed — merge-base 901ed2ed already carries the weak needle at line 600 ("REPRODUCER-EXIT-CONTRACT-OK"); this diff never touched that line, so no strength was lost. The stronger needle would pass, but restoring it is an out-of-scope assertion change on a line this change does not own; named in the change summary for the operator to file if wanted.
finding-status: F2 withdrawn same defect as F1, same reason — the weakened-needle premise is false at merge-base; see F1.

reproducers-total: 2
finding-reproducer: F1 .superpowers/sdd/reproducers/0-1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-1-1.sh

## Pass log

### Round 0

- diff size: 108 lines, cap not exceeded — proceeding
- docs-only: exit 1 — first non-documentation path scripts/check-panel-reproducer-exit-contract.sh; resolved roster runs unchanged
- roster: compact — 37
- no addition this round — the resolved list ran alone
## git log --stat

commit 4e447a9e4c25101630fa7e6a555b64797bc1122f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 22:24:09 2026 +0300

    docs(flow): premise declarations must be referenced by the script

 scripts/check-panel-reproducer-exit-contract.sh | 11 ++++++++---
 skills/flow/review-panel-fix-round.md           |  3 ++-
 skills/flow/review-panel.md                     |  5 ++++-
 3 files changed, 14 insertions(+), 5 deletions(-)

commit 8c43b249ad28db093b91f94c5eafe2a811723bf0
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 22:22:26 2026 +0300

    fix(guard): declared premises must be referenced by the script

 .../check_panel_reproducer_exit_contract_test.go   | 41 +++++++++++++++++-
 stats/internal/guard/panelexitcontract.go          | 48 ++++++++++++++++++++++
 2 files changed, 88 insertions(+), 1 deletion(-)

## Session narrative

Decided the enforcement level at brainstorm: the guard-visible marker (the script must reference each declared premise path, audited at dispatch), not the runner-side audit KAN-839's design considered and rejected — the dispatch audit (`pcPremiseAudit`) runs immediately before the runner in the same loop, so a runner-side assertion would duplicate a check already performed, and the marker adds a check neither enforcement point performs today. TDD caught two REDs: the first was my own fixture bug (the premise cited other.txt:1 while the content sat on line 2 — the citation miss, not the marker), the second the true vacuous green the change kills. The prefixed-premise subtest moved from `TestPanelExitContractPremiseAudit` into `TestCheckPanelExitContract` as case 44 — `withPeers` is a closure local to that function, discovered on the first build failure. The two prose edits (an addition in review-panel.md, a reword in review-panel-fix-round.md) tripped check-verbatim-moves and are acknowledged in the changeRoot's verbatim-moves.txt; the plan named check-vocabulary.sh, which does not exist at this merge-base, so check-references.sh covered that slot. Mid-panel the flow store dropped for one dispatch-begin row (written to the local journal — the RECORDS LOSS note above reflects it; the store was reachable again for every later row). The panel raised one Minor finding in both roles, which I verified against git before acting: the claimed needle weakening at line 600 of the test file predates the change — merge-base 901ed2ed already carries the weak needle, so the diff lost nothing; both rows are withdrawn with that reason. The underlying observation (case 43/42 could assert the runnable count) is real and pre-existing, deliberately not repaired here, and named in the change summary for the operator to file if wanted.
