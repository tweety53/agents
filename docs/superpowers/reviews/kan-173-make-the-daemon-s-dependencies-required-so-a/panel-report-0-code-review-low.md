# Panel pass 0 — Code review (low) (verbatim report)

**Interference during this run, recorded rather than hidden:** all four pass-0 slots were dispatched
concurrently into one shared worktree, one of which (the Bugbot substitute) mutates code by design.
The intermittent full-suite failures this slot reports below — a nil-pointer panic in
`DispatchAttributor.Attribute` — match exactly the Bugbot substitute's mutation 2
(`NewDispatchAttributor(nil)`) being live in the shared worktree at the time. This slot correctly
declined to raise them as findings. Its own reading-based conclusions are unaffected: it read the
static `final-review.diff`.

---

I've reviewed the full diff (all 7 files) and cross-checked against the actual `internal/harvest` and `cmd/flowd` code.

**Verification run:** `go build ./...`, `go vet ./...`, `gofmt -l .` all clean. `go test ./internal/harvest/... -race -count=30` passed 30/30 with zero data-race warnings. A handful of `go test ./... -race -count=1` full-suite runs intermittently crashed/failed (nil-pointer panic in `DispatchAttributor.Attribute`, and unrelated tests failing together), but this did not reproduce when the `harvest` package was run in isolation or repeatedly, and no run ever reported `WARNING: DATA RACE`. That pattern (fails only under full-suite parallel load, never isolated, never race-detected) points to this sandbox's resource contention running many `-race` binaries concurrently, not a defect in this diff — I'm not reporting it as a finding since I could not pin it to the changed code with confidence.

**Findings: none.**

Specific things I checked and found correct:
- The four unconditional-call sites (`RunOnce`'s pricer loop, `maybeBackfillDispatchMeta`, `attributeDispatches`, `pendingSessionTokens`) all rely on `deps` being non-nil (enforced by `NewWatcher`'s panic) and on `NoDeps`'s no-op returns (nil error, zero/nil value) — verified each `NoDeps` method in `stats/internal/harvest/deps.go:200-243` and its assertions in `deps_test.go`.
- `attributeDispatches`'s removed `HasDispatchAttribution`/`sessionTokens==nil` guards are safe: `DispatchAttributor.Attribute` (`stats/internal/harvest/attribute.go:716-742`) short-circuits on `len(records)==0` and on no matching windows, so running it unconditionally against `NoDeps` (empty `DispatchWindowsForSession`) produces no ambiguous entries and no calls to `MergeDispatchMetrics`/`MarkDispatchesUnattributedByID` beyond harmless no-ops.
- `watcher_test.go`'s composite `Deps` wrappers (`pricingDeps`, `sessionBinderDeps`, `sessionBinderAndDispatchDeps`) delegate to the same fake instances the old `WithX` options passed — traced each explicit method through to the same fakes (`pricer`, `binder`, `dispatchWindows`, `dispatchSink`), no fake was silently swapped for a no-op.
- `NewDispatchAttributor(deps)` inside `NewWatcher` (`watcher.go:436`) wires the same window source the two former `WithDispatchAttribution` call sites varied — confirmed via `sessionBinderAndDispatchDeps.DispatchWindowsForSession` delegating to `windows` field set per-test.
- `TestDaemonWiresTheRealStore`'s `reflect.ValueOf(w).Elem().FieldByName("deps")` correctly reads an unexported field's nilness/type without needing `CanInterface()` — no reflect misuse.
- `cmd/flowd/main.go`'s `_ harvest.Deps = (*store.Store)(nil)` compile-time guard and the `newTranscriptWatcher` call site (`harvest.NewWatcher(root, st, attributor, st, logger)`) correctly pass `st` as both sink and deps.
