# Final review panel — kan-174-anchor-stage-marks-at-command-position

## Diff-size gate

`scripts/check-panel-diff-size.sh . c905998` measured **645** lines against the 2,000 cap: under cap,
exit 0. No operator prompt required.

## Roster

`reviewPanelRoster: light` — required: Primary · Principles · Code review (low).

## Pass 1

| Slot | Result |
|------|--------|
| Primary | clean — read the full diff and checked the test corpus against the literal mark invocations in `skills/myflow-do/SKILL.md` |
| Code review (low) | clean at low effort; **the `code-review` skill was unusable** — its fork's cwd resolved to the main checkout rather than this worktree, so it reported on ambient uncommitted state there, citing a file this diff never touches. Substituted with a direct high-confidence pass, per that slot's own substitution rule |
| Principles | **F1** below, plus a Minor recorded rather than fixed |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Important | `stats/internal/stages/synthetic.go` | the package's doc scopes it to stage vocabulary while it now holds a change-record provenance sentinel, and the placement comment justified that by import convenience rather than purpose |

**F1 fixed in `3e70fcc`** by widening the package doc to state the package's real scope — the stage
vocabulary plus the small set of constants the daemon and the CLI must agree on — and rewriting the
placement comment to give the actual reason: a producer/consumer pair needs one shared definition or
it drifts. Behaviour, constant value and every test assertion unchanged; the count stayed at 338,
which is the evidence it was placement only.

**A Minor recorded and deliberately not fixed.** The ordering rule — read state, then mark — is
carried entirely by `skills/myflow-fast/SKILL.md`'s prose; only its *consequence* (a synthetic-only
record) became machine-checkable. That is how every myflow gate in this repository works, so it is
not a pattern this change introduces. What was corrected is the doc comment's framing: a field
enforces the "is this a state" half, never the sequencing.

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 none — a doc/scope mismatch, read stats/internal/stages/names.go's package comment against synthetic.go's contents
