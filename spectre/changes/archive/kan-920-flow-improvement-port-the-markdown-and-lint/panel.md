# Review panel — kan-920-flow-improvement-port-the-markdown-and-lint

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | stats/internal/guard/normativeinventory.go:68 | ownedCorpus gets fsRoot, the root after smcAbs, not the root as the caller spelled it, so exit-2 refusals under a relative CHECK_NORMATIVE_INVENTORY_ROOT print an absolute path where the bash printed the caller's spelling. |   |
| F2 | primary | Minor | stats/internal/guard/pythonsuppressions.go:44 | A read error prints Go's open …: <err> text where grep printed grep: f: Permission denied. |   |
| F3 | principles | Minor | stats/internal/guard/ownedcorpus.go:194 | The doc comment still says asRel is… for ocRel, a leftover from the move. |   |
| F4 | principles | Minor | stats/internal/guard/pythonsuppressions.go:52 | The generic line splitter is mrpLines, prefixed for the mutation-pin guard, yet the python guard calls it too, and niScan re-implements the same split inline. |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-P1-1.sh
finding-reproducer: F2 none — needs an unreadable file, environment-dependent
finding-reproducer: F3 none — comment-only
finding-reproducer: F4 none — structural, no behavioural effect

## Pass log

### Round 0

- roster: compact — 20
- diff size: 3891 lines, under cap; docs-only: exit 1 (scripts/check-mutation-reproducer-pin.sh) — roster primary+principles unchanged; standards: CLAUDE.md, AGENTS.md; no addition this round — the resolved list ran alone
