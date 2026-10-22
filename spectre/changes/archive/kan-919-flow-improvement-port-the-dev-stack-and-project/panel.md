# Review panel — kan-919-flow-improvement-port-the-dev-stack-and-project

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/guard/uitestoverrides.go:60 | an unreadable Makefile (top-level or included) is reported clean. |   |
| F2 | primary | Minor | stats/internal/guard/specreach.go:116 | a signal-killed npm run is reported as exited -1, where the bash printed $? (143 for SIGTERM). |   |
| F3 | primary | Minor | stats/internal/guard/devstackfresh.go:63 | the fingerprint subprocess inherits the process environment, not the guard's Env. |   |
| F4 | primary | Minor | stats/internal/guard/check_dev_stack_fresh_test.go:69 | harness sub-labels collapsed. |   |
| F5 | primary | Minor | spectre/changes/kan-919-flow-improvement-port-the-dev-stack-and-project/design.md:47 | design.md decision harness-cases-by-label still says the port runs npm and npx the way the bash did. |   |
| F6 | principles | Minor | stats/internal/guard/specreach.go:116 | the run closure hand-rolls the exit-code mapping that exitCode already states once. | supersedes F2 |
| F7 | principles | Minor | stats/internal/guard/uitestoverrides.go:21 | the C locale's [[:space:]] set is spelled again. |   |
| F8 | test-sanity | Important | stats/internal/guard/check_dev_stack_fresh_test.go:84 | case 5 claims \\| unescapes into a real shell pipe, but it passes with the unescape removed. |   |

findings-total: 8
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed

reproducers-total: 8
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 none — the shim always passes os.Getenv, so the divergence is invisible from any command line
finding-reproducer: F4 none — label naming, no behaviour
finding-reproducer: F5 none — plan artifact wording
finding-reproducer: F6 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F7 none — duplication of a constant, no behaviour difference today
finding-reproducer: F8 .superpowers/sdd/reproducers/0-test-sanity-1.sh

## Pass log

### Round 0

- diff size: 2653 lines, under cap
- docs-only: exit 1 (scripts/check-dev-stack-fresh.sh) — resolved roster primary+principles plus test-sanity+call-efficiency
- roster: compact — 87
- no addition this round — the resolved list ran alone
- standards: CLAUDE.md, AGENTS.md

### Round 1

- no pass row recorded — dispatched: panel-1-primary+principles+test-sanity
