# Review panel — kan-922-flow-improvement-port-the-bash-python-tooling

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/guard/workspace.go:155 | Nothing tests the Review Focus contract that a failing docker exec makes survivors exit non-zero; return rc -> return 0 leaves TestWorkspace green. |   |
| F2 | principles | Important | stats/internal/guard/workspace.go:119 | workspace is handed an Env and never reads it; it runs exec.Command("docker") from the process PATH, against the plan preamble and lookPath(env, ...). |   |
| F3 | primary | Minor | stats/internal/guard/loadsets.go:35 | An empty "" argument is not replaced by the default as the script's ${1:-...}/${2:-...} did. |   |
| F4 | principles | Minor | stats/internal/guard/workspace.go:171 | wsExitCode re-implements exec.go's exit-code mapping and drops the 128+signal branch. |   |
| F5 | test-sanity | Minor | stats/internal/guard/workspace_test.go:86 | The round-trip empty/missing-id loop repeats the first four refusal-table rows. |   |
| F6 | call-efficiency | Minor | stats/internal/guard/installersandboxdiff.go:224 | One cksum subprocess per sandbox file, plus an Lstat after WalkDir already has the entry. |   |
| F7 | call-efficiency | Minor | stats/internal/guard/loadsets.go:58 | About 150 re-stats of shared files across the TOTAL rows. |   |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 withdrawn each row must print its own MISSING lines, so a per-row walk is the report contract; ~150 stats on a hand-run report

reproducers-total: 7
finding-reproducer: F1 .superpowers/sdd/reproducers/0-P1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-PR1-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-P2-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-PR2-1.sh
finding-reproducer: F5 none — duplicated assertions, no behaviour defect
finding-reproducer: F6 none — cold path cost, no wrong output
finding-reproducer: F7 none — cold path cost, parity with the script

## Pass log

### Round 0

- roster: full
- diff size: 1978 lines, under cap — proceed
- docs-only: exit 1, first non-doc path scripts/installer-sandbox-diff.sh — roster primary+principles unchanged
- standards: CLAUDE.md, AGENTS.md
- no addition this round — the resolved list ran alone.

### Round 1

- round 1 re-run of all four roles over fix-round-1.diff: clean
