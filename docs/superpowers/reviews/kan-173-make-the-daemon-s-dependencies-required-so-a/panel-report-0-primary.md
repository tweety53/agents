# Panel pass 0 — Primary (verbatim report)

Correctly historical (referencing a prior commit's compile-time record) — not a stale claim about current code. Everything checks out. Full test suite, vet, gofmt, and reproductions all confirm the branch matches its plan and design exactly. No defects found.

## Review: kan-173-make-the-daemon-s-dependencies-required-so-a — Primary (plan alignment + code quality)

**Verification performed (all reproduced, not just read):**
- `go build ./...`, `go vet ./...`, `gofmt -l .` — all clean.
- `go test ./... -race -count=1` — all packages pass.
- Task-1 baseline (140→142) and task-2 baseline (142→143) `=== RUN` counts on `internal/harvest` reproduced exactly: 143.
- Task-3 baseline (5) on `cmd/flowd` reproduced exactly: 5.
- `TestNewWatcherPanicsOnNilDeps` run directly — passes.
- Reproduced the doc comment's specific claim on `TestDaemonWiresTheRealStore`: temporarily edited `newTranscriptWatcher` to pass `harvest.NoDeps{}` → test failed with the exact logged message `"the daemon wired harvest.NoDeps: nothing is priced, no session token is bound, no dispatch is charged"`; reverted → test passed. Matches the comment verbatim.
- Confirmed every `With*`/`Has*`/`WatcherOption` symbol is gone from live code — only historical mentions remain in comments (main.go's KAN-172 history, endtoend_test.go's pre-task-5 compile record, watcher_test.go's collapse-explanation comments).
- Confirmed `Deps`'s four composed interfaces (`Pricer`, `SessionTokenBinder`, `DispatchMetricsSink`, `DispatchWindowSource`) match `NoDeps`'s 9 methods exhaustively, and the compile-time assertion `_ harvest.Deps = (*store.Store)(nil)` in main.go proves the store satisfies it whole.

**Design decisions vs. code — all six hold:**
- `deps-interface`: `NewWatcher(root, sink, attributor, deps, logger)`, `deps` required, positioned before `logger` exactly as decided.
- `sink-stays-separate`: `HarvestSink` still its own parameter.
- `nil-deps-panics`: panic message names `deps` and points at `NoDeps{}`.
- `exported-nodeps`: `NoDeps` exported, zero-value/nil-error no-ops, `deps_test.go` asserts it directly.
- `attributor-built-inside`: `dispatchAttributor: NewDispatchAttributor(deps)` inside `NewWatcher`.
- `wiring-test-asserts-real-store`: `Has*` deleted, `TestDaemonWiresTheRealStore` added and reproduced above.

(Note: the dispatch prompt says design.md has "seven decisions" — the file's `## Decisions` section only names six. Not a code defect, just flagging the mismatch between the dispatch text and the actual design.md content.)

**"The honest limit" section**: checked every doc comment and test name touched by this diff (`newTranscriptWatcher`, `Deps`, `NoDeps`, `TestDaemonWiresTheRealStore`) — none claims to catch a wrong-but-non-nil dependency or an omission outside `*Watcher`. The limit as stated in design.md is honestly reflected in the code; nothing oversells.

**Composite test wrappers (`pricingDeps`, `sessionBinderDeps`, `sessionBinderAndDispatchDeps`)**: proportionate. Each exists because a fake needs to isolate one dependency axis while leaving the rest inert (`NoDeps` embedded), which is exactly what the pre-existing single-option tests needed. design.md itself discloses that three wrappers were needed instead of the one originally sketched, and explains why (a corrected 44-call-site split) — that's an honestly recorded course-correction, not scope creep.

**Doc-comment quality**: consistent with the package's existing why-not-what standard throughout the diff — every changed comment (Pricer, Watcher struct fields, `pendingSessionTokens`, `attributeDispatches`, `newTranscriptWatcher`) explains the reasoning behind the KAN-173 change and updates historical references correctly rather than leaving stale claims.

No findings — Critical, Major, or Minor. This is a clean result.
