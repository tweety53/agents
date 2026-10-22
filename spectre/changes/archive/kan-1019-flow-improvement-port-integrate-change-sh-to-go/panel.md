# Review panel — kan-1019-flow-improvement-port-integrate-change-sh-to-go

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/guard/integratechange.go:178 | The Go process keeps the caller's cwd, so cleanup run from inside a worktree is HELD by itself |   |
| F2 | primary | Important | stats/internal/guard/integratechange_land.go:95 | recordFixedRows trims each report line before matching, so an indented row is recorded as a fixed store row the bash never wrote |   |
| F3 | primary | Important | stats/internal/guard/integratechange_cleanup.go:124 | The orphan sweep trims each id, so a padded id the bash refused is run through the workspace remove command |   |
| F4 | primary | Minor | stats/internal/guard/integratechange.go:611 | An empty-string worktree value reads as -, the bash's "" |   |
| F5 | primary | Minor | stats/internal/guard/integratechange.go:470 | rebaseLint passes every KEY=value-looking line to run-verify-list, where the bash's export refused an invalid identifier |   |
| F6 | primary | Minor | stats/internal/guard/integratechange.go:263 | A missing unchecked sibling exits 126 in the Go, 127 in the bash |   |
| F7 | primary | Minor | stats/internal/guard/integrate_change_test.go:176 | The Go tests do not isolate $HOME, which the plan's Test isolation clause promises |   |
| F8 | primary | Minor | spectre/changes/kan-1019-flow-improvement-port-integrate-change-sh-to-go/design.md:1 | design.md's go-tests-in-process-two-files no longer describes what shipped |   |
| F9 | principles | Important | stats/internal/guard/integratechange.go:218 | icChildEnv hard-codes the test harness's variable set in production code |   |
| F10 | principles | Minor | stats/internal/guard/integratechange_prepare.go:292 | Filesystem failures are swallowed where the bash surfaced them |   |

findings-total: 10
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed

reproducers-total: 10
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 none — prepare-workspace.sh prints no such line today
finding-reproducer: F6 none — rule 2 requires the three siblings beside the shim
finding-reproducer: F7 none — needs a change to the operator's global git config
finding-reproducer: F8 none — documentation drift
finding-reproducer: F9 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F10 none — needs a filesystem write to fail under the worktree's git dir

## Pass log

### Round 0

- roster: full
- diff size: 5558 lines, over cap — proceeded unasked, the whole diff read
- docs-only: exit 1 — first non-doc path scripts/integrate-change.sh; roster dispatched: primary+principles, test-sanity+call-efficiency
- no addition this round — the resolved list ran alone
- standards passed: CLAUDE.md, AGENTS.md
- relocation comparison: not generated — generate-relocation-comparison.sh could not read the deleted scripts/test-integrate-change.sh

### Round 1

- no pass row recorded — dispatched: panel-1-primary+principles
