# Review panel — kan-1021-flow-improvement-move-the-bash-hook-and-setup

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/guard/pipelinefix.go:357 | pfChecks now selects the repotest commands unconditionally for any touched stats/*.go file and any touched hooks/* path, where the code it replaced added them only when they existed in the worktree (exists(h)). |   |
| F2 | primary | Minor | stats/internal/repotest/protect_main_checkout_test.go:3 | Tasks 1-4's Step 5 absence sweep is ticked as printing nothing, but each new test file's doc comment names its deleted script, so the sweep prints one line per task. |   |
| F3 | test-sanity | Minor | stats/internal/repotest/protect_main_checkout_test.go:64 | run discards the hook's exit code and stderr; an empty stdout reads as allow, so a crashing hook passes every allow case. |   |
| F4 | test-sanity | Minor | stats/internal/repotest/flow_addr_declaration_test.go:38 | flow_addr_declaration_test.go mutation cases assert only a non-empty failure reason, not which check fired, so a mutation may fail for a different reason than its name states. |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/r0-P1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/r0-P1-2.sh
finding-reproducer: F3 none — allow cases only, deny cases cover a crash
finding-reproducer: F4 none — unproved

## Pass log

### Round 0

- roster: compact — 77; diff 4986 lines under cap; docs-only exit 1 (mods/flow-task-list/hooks/landed.ts); no operator addition this round — the resolved list ran alone; standards: CLAUDE.md, AGENTS.md; full suite ran green before the pass (run-verify-list test exit 0)

### Round 1

- F1 reproducer demonstrates citation was off by one line (379 -> 380, the add line under the hooks/ case); the parent corrected the line number in place instead of a bounce — content unchanged, the reproducer still demonstrates the defect
- fix round 1: panel-fix-1 (opus/low) fixed F1-F4 in fd52c163; reproducers F1,F2 flipped demonstrated->not; F3,F4 Minor none-reproducers closed on path; re-run: primary (F1,F2) and test-sanity (F3,F4) on fix-round-1.diff; principles and call-efficiency raised nothing — not re-run
fix-mutation: stats/internal/guard/pipelinefix.go — dropped && repotest from the Go-audits gate — TestPipelineFixChecksGoAudits
fix-mutation: stats/internal/guard/pipelinefix.go — dropped && repotest from the hooks/ case — TestPipelineFixChecksGoAudits
fix-mutation: stats/internal/repotest/flow_addr_declaration_test.go — weakened the literal case's want substring to another check's — TestFlowAddrDeclaration
fix-mutation: stats/internal/repotest/protect_main_checkout_test.go — rc assert disabled while the hook crashes — TestProtectMainCheckout
fix-mutation: stats/internal/repotest/protect_main_checkout_test.go — none — comment-only doc comments (F2), four repotest files
fix-mutations-total: 5
