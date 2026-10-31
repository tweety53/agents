# Review panel — kan-1072-audit-whether-archived-spectre-change-artifacts

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F18 | 0 | primary | Important | skills/spectre-planning/SKILL.md:34 | skills/spectre-planning/SKILL.md:34 and stats/internal/lessons/lessons.go:74 both still read spectre/changes/archive/, which this change deletes at the tip of every repository |   |
| F19 | 0 | primary | Minor | skills/flow-contracts/project-configuration-rationale.md:205 | cites spectre/changes/archive/kan-171-.../design.md, a tip path removed by the archive deletion |   |
| F20 | 0 | principles | Minor | stats/internal/landed/landed.go:31 | Git seam type exported alongside File; justified by the selfreview bounded runner caller. No action. |   |
| F21 | 0 | call-efficiency | Critical | stats/internal/api/stats.go:852 | readPlanFile now falls back to landed.File for every row whose live tasks.md is missing — whole-history git log --all calls on every 5 s board poll |   |
| F22 | 0 | mutation | Important | stats/internal/landed/landed.go:62 | Survived: dropping --all — no test puts the drop/archive commit on a ref other than HEAD |   |
| F23 | 0 | failure-modes | Minor | stats/internal/landed/landed.go:46 | cat-file -e failure of any kind (not just missing object) is reported as ok=false, nil error, masking git faults as absent |   |
| F24 | 1 | call-efficiency | Important | stats/internal/lessons/lessons.go:215 | Resolve now runs landed.File per landed change, each spawning git log --all --grep (1-2x) + ls-tree + show |   |
| F25 | 1 | failure-modes | Minor | stats/internal/lessons/lessons.go:215 | lessons.go landedChanges swallows a git log failure as no landed changes |   |
| F26 | 1 | failure-modes | Minor | stats/internal/landed/landed.go:71 | Root-commit drop (no parent) would surface as an ls-tree error rather than absent |   |

findings-total: 9
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 withdrawn the reviewer marked it no action — the exported seam has a second caller
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed
finding-status: F25 fixed
finding-status: F26 withdrawn a drop commit always has a parent — the change it drops was committed before it

reproducers-total: 9
finding-reproducer: F18 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F19 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F20 none — observation the reviewer marked no action
finding-reproducer: F21 .superpowers/sdd/reproducers/0-call-efficiency-1.sh
finding-reproducer: F22 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F23 none — failure-path minor
finding-reproducer: F24 .superpowers/sdd/reproducers/1-call-efficiency-1.sh
finding-reproducer: F25 none — failure-path minor
finding-reproducer: F26 none — failure-path minor

## Pass log

### Round 0

- diff size 4089 lines, under cap; docs-only exit 1 (first non-doc path hooks/protect-main-checkout.py) — resolved roster primary+principles plus both always-on extra dispatches; roster: full; no addition this round — the resolved list ran alone; standards passed: CLAUDE.md, AGENTS.md
- dispatched: panel-0-primary+principles on opus/low; panel-0-test-sanity+call-efficiency on opus/low
- dispatched: panel-0-mutation+failure-modes on sonnet/low
- verdict: fix (1 Critical, 2 Important)

### Round 1

- dispatched: panel-1-primary+call-efficiency+failure-modes on sonnet/low; panel-1-mutation on sonnet/low
- verdict: fix (0 Critical, 1 Important)
fix-mutation: stats/internal/landed/landed.go — dropped "--all" from Read — TestFileReadsADropOnAnotherRef
fix-mutation: stats/internal/landed/landed.go — ls-tree error returned as absent — TestReadReportsAGitFault
fix-mutation: stats/internal/landed/landed.go — Recent span "-n 50 HEAD" -> "--all" — TestRecentReadsHEADsTip
fix-mutation: stats/internal/api/stats.go — in-flight read landed.Recent -> landed.File — TestPlanTasksInFlightReadsHEADOnly
fix-mutation: stats/internal/api/stats.go — finished read landed.File -> landed.Recent — TestPlanTasksInFlightReadsHEADOnly
fix-mutation: stats/internal/lessons/lessons.go — narrative read stubbed absent — TestResolveBriefOutranksNarrative
fix-mutation: stats/internal/guard/integratechange_prepare.go — removed --no-renames — TestGitConfigPins
fix-mutations-total: 7

### Round 2

- dispatched: panel-2-call-efficiency+failure-modes on sonnet/low
- verdict: clean
fix-mutation: stats/internal/lessons/lessons.go — extra git log -1 per change in the drop branch — TestResolveReadsHistoryInOneWalk
fix-mutation: stats/internal/lessons/lessons.go — git log error path returns nil, nil — TestResolveGitLogFailureIsReported
fix-mutations-total: 2
