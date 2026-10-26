## Context

The fix report (`panel-fix-report-<round>.md`, REPORT FILE paragraph of `skills/flow/review-panel.md`)
already carries each fixed finding's RED/GREEN/reverted proof and the `fix-mutation:` lines; the
close guard `check-panel-findings-closed` (`stats/internal/guard/panelfindingsclosed.go`) already
reads the worktree's `.superpowers/sdd/` reports and the findings' store rows, and already resolves
a token across the state record's worktrees for cross-repo fixes. The new check sits on both.

## Decisions

### Where the finding→test link lives

**ID:** regression-test-report-line
**Status:** active
**Chosen:** a `regression-test: F<n> — <path> — <test>` / `regression-test: F<n> — none — <reason>` line in the round's fix report — the report already carries per-finding proof, and the guard only needs the link until the panel closes.
**Considered:** a `Regression-Test:` commit trailer — an autosquash `fixup!` fold discards the fixup's message, so the trailer dies on the common fold route; a new store column on findings — a migration, flowd and CLI change for data needed only until the close.

### What the guard proves

**ID:** guard-checks-tracked-not-red
**Status:** active
**Chosen:** the guard checks a line exists and its path is git-tracked (in the worktree or a recorded worktree) and outside `.superpowers/` — the fails-without-the-fix proof stays the report's RED/reverted run, which the re-running slot already checks.
**Considered:** the guard re-running the test at the pre-fix commit — language-specific per project and slow.

### Escape for findings no suite test can cover

**ID:** none-form-allowed
**Status:** active
**Chosen:** `none — <reason>`, mirroring the reproducer and `fix-mutation:` none forms — many agents-repo findings are prose-only contract fixes.
**Considered:** no escape — would block every prose-only fix round.

### Inline fixes

**ID:** inline-parent-writes-lines
**Status:** active
**Chosen:** when the parent fixes inline, it writes the same lines to `panel-fix-report-<round>.md` itself.
**Considered:** exempting inline rounds — leaves the same hole on every `inline` class change.

## Open questions

## Recommended approach

One report line plus one block in the existing close guard; no store change. Rejected: trailer and
store column (`regression-test-report-line`), guard-run RED (`guard-checks-tracked-not-red`).

## Critical files

- `stats/internal/guard/panelfindingsclosed.go` — the close guard; new block beside the fixed-reproducer block.
- `stats/internal/guard/check_panel_findings_closed_test.go` — its tests.
- `scripts/check-panel-findings-closed.sh` — the shim whose header is the guard's contract.
- `skills/flow/review-panel.md` — REPORT FILE paragraph and the close-guard paragraph.
- `skills/flow/review-panel-fix-round.md` — the fix-round rule.

## Existing code to reuse

- `pcWorktreeKeys`, `resolvedRecordedTrees`, `pcFlow` (lazy state-record read) in `panelfindingsclosed.go`'s fixed-reproducer block.
- `cfcFindingRow` (`status`, `severity`, `ref`) and `cfcPrefix` for messages.
- `TestCheckPanelFindingsClosedFixedReproducerExists`'s fixture shape for the new test.

## Verification

- `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckPanelFindingsClosed' -count=1`; `go vet ./...`; `gofmt -l .`.
- `scripts/check-references.sh` and the other `scripts/check-*.sh` guards of `.flow/project.md`'s `## lint`.
- Disable the new block's violation → `TestCheckPanelFindingsClosedFixedNeedsARegressionTest` fails.

## Live check

None — the change is a guard over files and store rows the guard's own tests fake; no running service or persistent state changes.
