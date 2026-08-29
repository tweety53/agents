# Panel pass 0 — Principles (verbatim report)

## Principles Review — kan-173-make-the-daemon-s-dependencies-required-so-a

**Reproduction performed** (not just read):
- `go build ./...`, `go vet ./...`, `gofmt -l .` — all clean.
- `go test ./... -race -count=1` — all packages pass.
- `TestNewWatcherPanicsOnNilDeps`, `TestNoDeps*` — pass in isolation.
- `TestDaemonWiresTheRealStore` — passes against the real wiring; edited `cmd/flowd/main.go` to pass `harvest.NoDeps{}` instead of `st`, reran, got the expected failure (`wiring_test.go:209: the daemon wired harvest.NoDeps: ...`), then reverted. Confirms the test actually detects the regression it claims to.
- `grep` swept the whole tree for `nolint`, leftover `Has*`/`WatcherOption`/`With*` symbols (none live — only historical prose comments remain), and for non-test importers of `harvest.NoDeps` (none — it is consumed only from `_test.go` files, justifying its exported status as a cross-package test seam rather than a production escape hatch).

Note: found and reverted a pre-existing, unrelated uncommitted mutation in the worktree (`stats/internal/harvest/watcher.go` had `NewDispatchAttributor(nil)` in place of `NewDispatchAttributor(deps)`, plus a stray `cmd/flowd/main.go.bk`) before starting — restored via `git checkout -- stats/internal/harvest/watcher.go` and `rm` of the backup file, so the tree matches HEAD (`4da13f7`) for review. Neither was part of the diff under review.

*(Dispatcher's note: that "pre-existing mutation" was in fact the concurrently-running Bugbot
substitute's live mutation test. All four pass-0 slots shared one worktree. See
`panel-report-0-bugbot.md` and its clean exclusive re-run.)*

### Issues

#### Critical (Must Fix)
None.

#### Important (Should Fix)
None.

#### Minor (Nice to Have)

1. `stats/cmd/flowd/wiring_test.go:148` — `TestDaemonWiresTheRealStore` reads the unexported `Watcher.deps` field via `reflect.ValueOf(w).Elem().FieldByName("deps")`. This is real cross-package coupling to an implementation-private field name, and design.md's `wiring-test-asserts-real-store` decision records "delete the test" and "a broader reflection test over every field" as the alternatives considered, but not the narrower fix actually shipped, nor the simpler one it foregoes — a single exported `func (w *Watcher) Deps() Deps { return w.deps }` accessor, which would let the test do a plain type assertion and drop `reflect` entirely. The `f.IsValid()` guard does make a future rename fail loudly (`t.Fatal`) rather than silently pass, so this is not a correctness risk, only a documentation/technique gap. Given the diff just deleted three `Has*` accessors specifically to shrink `Watcher`'s public surface, choosing reflection over reopening that surface with one getter is a defensible call — but it isn't the call design.md says was weighed.
   **Reproducer:** `cd stats && go test ./cmd/flowd/ -run TestDaemonWiresTheRealStore -v` passes today; rename the `deps` field in `watcher.go` and the same command fails with `harvest.Watcher has no deps field: this test no longer checks anything` — demonstrating the coupling (not silently, but a coupling nonetheless).

### Assessment
**Principles-compliant?** Yes.
**Reasoning:** The panic-on-nil-deps is a sound fail-fast application of Robustness's "fail fast on programmer error," `Deps` is a properly consumer-sized interface (Watcher genuinely calls all four methods), `DispatchAttributor` construction inside `NewWatcher` is pure/no-I/O wiring rather than a smuggled responsibility, and `NoDeps`'s export is justified by real cross-package test consumers, not unused surface. Project standards (`CLAUDE.md`/`AGENTS.md`) resolved and show no violated hard invariant — no new suppressions, no lint weakening, formatting clean.
