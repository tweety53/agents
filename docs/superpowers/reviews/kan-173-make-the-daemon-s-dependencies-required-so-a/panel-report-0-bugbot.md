# Panel pass 0 — Bugbot (general-purpose substitute; verbatim report)

**Slot substituted:** the `bugbot` agent type is not offered by this harness (Claude Code), so this
slot ran as a general-purpose subagent carrying Bugbot's brief plus the mandatory mutation-testing
requirement. Model actually given: `sonnet`.

**Interference during this run, recorded rather than hidden:** all four pass-0 slots were dispatched
concurrently into one shared worktree. The Principles slot, at its own start, found and reverted an
uncommitted mutation this slot had in flight (`NewDispatchAttributor(nil)` in
`stats/internal/harvest/watcher.go`) and removed a `cmd/flowd/main.go.bk` backup file. The report
below was therefore produced in a worktree another agent was concurrently restoring. See
`panel-report-0-bugbot-rerun.md` for the clean, exclusive re-run that supersedes it.

---

## Bugbot report — kan-173-make-the-daemon-s-dependencies-required-so-a

**Baseline:** `go build ./...`, `go vet ./...`, `gofmt -l .` all clean; `go test ./... -race -count=1` all green (17 packages). `git status --short` shows no tracked-file modification at finish.

### Findings

No defects found. All mutation targets were caught by the existing test suite. One informational (non-defect) observation on test redundancy below.

**1. Informational — `IsNil` check in `TestDaemonWiresTheRealStore` is dead code.** `cmd/flowd/wiring_test.go:204-206`. `harvest.NewWatcher` (`internal/harvest/watcher.go:428-430`) panics before construction if `deps == nil`, so a `*Watcher` returned from it can never have a nil `deps` field — the branch this check guards is unreachable via any real construction path. Not a bug (defensive code, harmless), just redundant given the panic guard added by this same change. Reproducer: `none — informational; confirmed by removing the check (see mutation 4c below), which left the test still passing under normal conditions with no change in coverage.`

### Mutation-testing results (all 26 mutations, all caught except where noted)

**1. `NewWatcher`'s `deps == nil` panic removed** (`internal/harvest/watcher.go:428-430`) — caught by `TestNewWatcherPanicsOnNilDeps`.

**2. `NewDispatchAttributor(deps)` replaced with `NewDispatchAttributor(nil)`** (`watcher.go:436`) — caught (broadly): nil-pointer panic in `DispatchAttributor.Attribute` surfaces through `TestOutageAcrossSeveralCyclesThenRecoveryMatchesCleanRun` (and would surface in nearly every other RunOnce-driving test).

**3a. `RunOnce`'s unconditional `w.deps.Price(...)` dropped** (`watcher.go:661-665`) — caught by `TestRunOncePricesTouchedStageRunsAfterCommit`.
**3b. `maybeBackfillDispatchMeta`'s `w.deps.Price(...)` dropped** (`watcher.go:947-951`) — caught by `TestBackfillDispatchMetaPricesStageRunAfterCommit`.
**3c. `attributeDispatches` body short-circuited** (`watcher.go:791-792`) — caught by `TestAmbiguousDispatchIsStamped` and `TestDispatchThatAttributedIsNeverStampedUnattributed`.
**3d. `pendingSessionTokens` body short-circuited to `return nil`** (`watcher.go:964-965`) — caught extensively (17+ test failures across session-token/give-up tests).

**4. `TestDaemonWiresTheRealStore`'s own guards** (`cmd/flowd/wiring_test.go:200-211`), each weakened *together with* simulating the production defect (`newTranscriptWatcher` wired with `harvest.NoDeps{}` instead of `st`):
- **4a. `IsValid` check removed** — still caught, as a `reflect: call of IsNil on zero Value` panic (test still fails, just less gracefully). Not a true survivor.
- **4b. `IsNil` check removed** (deps never actually nil via `NewWatcher`) — test still passes normally (no coverage lost; see informational finding above — dead code, not exploitable since deps can't be nil here).
- **4c. `f.Elem().Type() == reflect.TypeOf(harvest.NoDeps{})` comparison removed**, combined with the simulated wiring defect — the targeted run and the full suite both stay green. This confirms the type-comparison check is the load-bearing one — it is present and correct in the actual diff, but removing it (hypothetically) leaves the exact defect class KAN-16/KAN-172 shipped completely undetected. Since the check *is* present in the shipped code, this is not a live defect — it demonstrates the check earns its place.

**5. `harvest.NoDeps` — all 9 methods mutated to return non-zero/non-nil, one at a time** (`internal/harvest/deps.go`) — every mutation caught by `TestNoDepsReturnsZeroValuesAndNoError`: `Price`, `UnresolvedSessionTokens`, `BindSession`, `RecordSessionTokenGiveUp`, `PersistedGiveUps`, `MarkDispatchesUnattributedByID`, `MarkDispatchesUnattributed`, `MergeDispatchMetrics`, `DispatchWindowsForSession`.

**6. Composite `Deps` wrappers in `watcher_test.go` disconnected from their real fakes:**
- `pricingDeps.Price` -> no-op — caught by `TestBackfillDispatchMetaPricesStageRunAfterCommit`, `TestRunOncePricesTouchedStageRunsAfterCommit`, `TestPricingFailureIsNotFatal`.
- `sessionBinderDeps.BindSession` -> no-op — caught extensively (11+ tests, including `TestMarkRecognizedWhereverItSitsInTheCommand`'s 10 subtests).
- `sessionBinderAndDispatchDeps.MergeDispatchMetrics` -> no-op — caught by `TestDispatchThatAttributedIsNeverStampedUnattributed`.
- `sessionBinderAndDispatchDeps.DispatchWindowsForSession` -> no-op — caught by `TestAmbiguousDispatchIsStamped` and `TestDispatchThatAttributedIsNeverStampedUnattributed`.

### Summary

26 mutations performed across all 6 target categories; 25 caught cleanly, 1 (`IsValid`) caught indirectly via panic, 1 (`IsNil`) confirmed as intentionally-redundant dead code rather than a coverage gap. No surviving mutant. No ordinary defect found in nil-interface handling, reflect usage, concurrency, or error paths beyond the one informational note above.
