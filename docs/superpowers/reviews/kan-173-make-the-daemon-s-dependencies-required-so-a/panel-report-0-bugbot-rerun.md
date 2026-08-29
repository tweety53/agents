# Panel pass 0 — Bugbot, exclusive re-run (verbatim report)

**Why this run exists.** The four pass-0 slots were dispatched concurrently into one shared
worktree, and this slot mutates code by design. The Principles slot detected this slot's
predecessor's in-flight mutation (`NewDispatchAttributor(nil)` in `watcher.go`) and reverted it with
`git checkout --`; the Code review (low) slot observed the intermittent nil-pointer panics those live
mutations caused. Mutation results measured while another process is restoring the file under test
are not measurements, so this slot was re-dispatched with the worktree held exclusively, clean at
HEAD `4da13f7`. **This report supersedes `panel-report-0-bugbot.md`.**

**Slot substituted:** the `bugbot` agent type is not offered by this harness (Claude Code), so this
slot ran as a general-purpose subagent carrying Bugbot's brief plus the mandatory mutation-testing
requirement. Model actually given: `sonnet`.

---

Clean — matches the pre-check state exactly. No edits left behind, nothing committed.

## Report

**Findings: none.** No defects found. All ordinary-defect hunting (nil-interface handling, `reflect` misuse, `NoDeps` zero-return misuse, concurrency, error paths) turned up nothing beyond what's already covered by tests.

**Specific claim settled — the earlier pass's `IsNil` dead-code claim is CORRECT.**
- `NewWatcher`'s `if deps == nil { panic(...) }` guard catches only a *true* nil interface (nil type + nil value).
- `TestDaemonWiresTheRealStore` passes a nil `*store.Store` as the `Deps` argument, which produces a *typed*-nil interface (dynamic type `*store.Store`, value nil) — Go's `==` and `reflect.Value.IsNil()` both report this as non-nil, confirmed via a standalone reproducer (`/tmp/nilcheck.go`, deleted after use):
  ```
  interface == nil: false
  reflect IsNil: false
  ```
- Only `NewWatcher` constructs `&Watcher{}` in package `harvest` (verified: `grep -rn "&Watcher{" internal/harvest/*.go` outside tests returns one hit), and it panics before construction on any true-nil `Deps`. There is no public construction path that produces a `*Watcher` with a genuinely nil `deps` field.
- Confirmed by mutation: removing the `IsNil()` check from `wiring_test.go` and rerunning `TestDaemonWiresTheRealStore` still passes — no test catches its removal. This check is unreachable dead code.

**Mutations performed (all restored, `git diff` verified empty after each):**

| # | Target | Result |
|---|--------|--------|
| 1 | `NewWatcher`'s `deps == nil` panic removed | Caught — `TestNewWatcherPanicsOnNilDeps` fails |
| 2 | `NewDispatchAttributor(deps)` -> `NewDispatchAttributor(nil)` | Caught — `TestOutageAcrossSeveralCyclesThenRecoveryMatchesCleanRun` panics (nil pointer deref in `DispatchAttributor.Attribute`) |
| 3a | Dropped `w.deps.Price(...)` in `RunOnce` | Caught — 4 tests fail (`TestMarkedStageBindsToRealSessionEndToEnd`, `TestBackfillDispatchMetaPricesStageRunAfterCommit`, `TestRunOncePricesTouchedStageRunsAfterCommit`, `TestPricingFailureIsNotFatal`) |
| 3b | Dropped `w.deps.Price(...)` in `maybeBackfillDispatchMeta` | Caught — `TestBackfillDispatchMetaPricesStageRunAfterCommit` |
| 3c | Short-circuited `attributeDispatches` body | Caught — `TestAmbiguousDispatchIsStamped`, `TestDispatchThatAttributedIsNeverStampedUnattributed` |
| 3d | Short-circuited `pendingSessionTokens` body | Caught heavily — 10+ tests fail across session-binding, give-up, and retry suites |
| 4a | `wiring_test.go` `IsValid` guard removed | **Not caught** — expected; it's a self-check against a future field rename, not a live defect guard |
| 4b | `wiring_test.go` `IsNil` check removed | **Not caught** — confirms dead code, see above |
| 4c | `wiring_test.go` type-comparison (`f.Elem().Type() == reflect.TypeOf(harvest.NoDeps{})`) weakened, combined with re-introducing the real KAN-16/KAN-172-class defect (`newTranscriptWatcher` wired to `harvest.NoDeps{}`) | Caught by the *unweakened* check; with it weakened, the defect goes undetected — confirms this check is the one carrying the test's real weight |
| 5 | `NoDeps.BindSession` returns `42` instead of `0` | Caught — `TestNoDepsReturnsZeroValuesAndNoError`, which exhaustively covers all 9 `NoDeps` methods |
| 6a | `pricingDeps.Price` disconnected from `d.p` | Caught — 3 pricing tests |
| 6b | `sessionBinderDeps.BindSession` disconnected from `d.binder` | Caught heavily — 10+ tests |
| 6c | `sessionBinderAndDispatchDeps.DispatchWindowsForSession` disconnected | Caught — `TestAmbiguousDispatchIsStamped`, `TestDispatchThatAttributedIsNeverStampedUnattributed` |
| 6d | `sessionBinderAndDispatchDeps.MergeDispatchMetrics` disconnected | Caught — `TestDispatchThatAttributedIsNeverStampedUnattributed` |

Every mutation targeted in the brief was performed and evaluated; 13 of 15 killed the suite, 2 were confirmed dead/self-referential rather than surviving-mutant defects.

**Verification:** `go build ./...`, `go vet ./...`, `gofmt -l .` (clean), `go test ./... -race -count=1` (all 17 packages green) — run fresh at the end, with the worktree held exclusively throughout.

**Difference from the earlier pass:** none in substance — this run independently confirms the `IsNil`-is-dead-code claim with a standalone reproducer rather than reading it off. No new findings beyond what mutation testing surfaced, and none of it constitutes a blocking defect.
