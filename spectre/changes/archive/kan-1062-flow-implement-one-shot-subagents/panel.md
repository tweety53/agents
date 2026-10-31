# Review panel — kan-1062-flow-implement-one-shot-subagents

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F7 | 0 | primary | Important | stats/internal/store/aggregate.go:261 | full-suite fix rounds silently drop out of FixRounds: under sdd the full-suite fixer records as role implementer, and FixRounds counts only panel-fix |   |
| F8 | 0 | primary | Important | skills/flow/implement.md:570 | The stub rule says the implementer prompt is these lines alone, but skills/flow/verify-fix-loop.md:77-78 still requires the in-run fix task's implementer dispatch to also carry the MUTATION PROOF and PIXEL PROBE paragraphs |   |
| F9 | 0 | primary | Minor | skills/flow/implement.md:468 | Each still cites The next implementer overlaps the guard, which this diff renamed (also scripts/close-task.sh:3, stats/internal/guard/closetask.go:19, SKILL-rationale.md:609; implement.md:468 says step 3 where begins now ride step 2) |   |
| F10 | 0 | primary | Minor | skills/flow/primary-reviewer-prompt.md:24 | primary-reviewer-prompt.md tells a gated pass its context bundle is the task's own context file; it is now the group's bundle — the description of the input is false |   |
| F11 | 0 | primary | Minor | stats/internal/store/experimentstats.go:97 | the implementer experiment cell now also takes every sdd recommit and gated fix subagent row, mixing fixer-pair dispatches into implementer-pair cells |   |
| F12 | 0 | primary | Minor | skills/flow/implement.md:945 | the keep-warm background sleep 3000 is started and re-armed but nothing stops it once the last child returns, so it still fires up to 50 minutes later |   |
| F13 | 0 | principles | Minor | stats/cmd/flow/record.go:2037 | treats a present but non-string ref ("ref": 4, decoded as json.Number) as ref-less, so numberRefLessFindings silently overwrites it |   |
| F14 | 0 | mutation | Important | stats/cmd/flow/record.go:1965 | -ref= / --ref= branch of the -ref-beside--from refusal is untested (mutated to -xref=; suite green) |   |
| F15 | 0 | mutation | Important | stats/cmd/flow/record.go:1991 | the null-line must be a JSON object exit-2 guard is untested |   |
| F16 | 0 | mutation | Important | stats/cmd/flow/record.go:2081 | recordIdentityArgs dropping -C is untested (a -C dir is not forwarded to the next-record read) |   |
| F17 | 0 | mutation | Important | stats/cmd/flow/record.go:2059 | SetEscapeHTML(false) flip (rewritten line escapes <, >, &) is untested |   |
| F18 | 0 | failure-modes | Important | stats/cmd/flow/record.go:2066 | partial write: WriteFile truncates then writes in place, so a kill mid-write leaves a truncated findings file with no copy; the findings are the only source and are lost |   |
| F19 | 0 | failure-modes | Minor | stats/cmd/flow/record.go:2010 | Concurrent re-entry: two simultaneous -from runs for one change both read the same store state and assign the same refs to different files, overwriting each other's rows |   |

findings-total: 13
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed

reproducers-total: 13
finding-reproducer: F7 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F11 none — stats attribution; nothing to run without a store
finding-reproducer: F12 .superpowers/sdd/reproducers/0-primary-5.sh
finding-reproducer: F13 none — needs the fake daemon; the type assertion on that line is the whole mechanism
finding-reproducer: F14 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F15 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F16 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F17 .superpowers/sdd/reproducers/0-mutation-4.sh
finding-reproducer: F18 .superpowers/sdd/reproducers/0-failure-modes-1.sh
finding-reproducer: F19 none — a race; cannot be forced deterministically

## Pass log

### Round 0

- diff-size: 4032 lines, under cap; docs-only: exit 1 — hooks/enforce-agent-baseline.py; roster: compact — 85; roster dispatched: primary+principles (opus/medium), test-sanity+call-efficiency (opus/low), mutation+failure-modes (sonnet/low); no addition this round — the resolved list ran alone; standards: CLAUDE.md, AGENTS.md
- dispatched: panel-0-primary+principles on opus/medium; panel-0-test-sanity+call-efficiency on opus/low
- dispatched: panel-0-mutation+failure-modes on sonnet/low
- reproducers 0-mutation-1..4: renamed off R-0-mutation-N.sh to the canonical 0-mutation-N.sh shape, and narrowed from the whole cmd/flow package to -run '^TestRecordFinding' (the family prefix) after the whole-package run left a detached child that failed run-reproducer's survivor check; 0-failure-modes-1.sh written by the parent for F18, whose slot gave none
- verdict: fix (0 Critical, 7 Important)

### Round 1

- dispatched: panel-1-primary+principles on opus/low; panel-1-mutation+failure-modes on opus/low
- check-panel-fix-single-dispatch exit 1: panel-fix-1-2's commit 7d3de853 carries Flow-Key: panel-fix-1 (the round, as review-panel-fix-round.md's 'A fix commit's Flow-Key: trailer names that same fix round' states) where the guard wants the chunk key panel-fix-1-2 — auto-resolved Continue (recommended); contract/guard mismatch taken as an in-run pipeline fix
- verdict: clean
fix-mutation: stats/internal/store/aggregate.go — full-suite-fix implementer arm regex → '^NOPE$' — TestLiveStateBoardCountsFixRoundsAndReviewerRuns
fix-mutation: stats/internal/store/experimentstats.go — fixer-key exclusion removed — TestExperimentStatsImplementerFamily
fix-mutation: stats/cmd/flow/record.go — non-string ref guard disabled — TestRecordFindingFromAssignsRefs
fix-mutation: stats/cmd/flow/record.go — -ref= prefix → -xref= — TestRecordFindingFromAssignsRefs
fix-mutation: stats/cmd/flow/record.go — record-loop row == nil → false — TestRecordFindingFromAssignsRefs
fix-mutation: stats/cmd/flow/record.go — "C" dropped from recordIdentityArgs — TestRecordFindingFromAssignsRefs
fix-mutation: stats/cmd/flow/record.go — SetEscapeHTML(false)→true — TestRecordFindingFromAssignsRefs
fix-mutation: stats/cmd/flow/record.go — replaceFile call reverted to os.WriteFile — TestRecordFindingFromAssignsRefs
fix-mutation: stats/cmd/flow/record.go — temp file chmod → 0o644 — TestRecordFindingFromAssignsRefs
fix-mutation: stats/cmd/flow/record.go — none — ignoring os.Rename's error: the rename cannot be failed without fault injection; equivalent in practice
fix-mutation: stats/cmd/flow/record.go — none — deferred temp removal: only reachable when write/close/chmod fails after the temp exists, not cheaply forceable
fix-mutation: skills/flow/implement.md — none — prose/comment-only edits across implement.md, verify-fix-loop.md, primary-reviewer-prompt.md, SKILL-rationale.md, close-task.sh, closetask.go
fix-mutations-total: 12
