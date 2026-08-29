# kan-173-make-the-daemon-s-dependencies-required-so-a

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Every path below is relative to the repository root. All three tasks are Go-only; nothing in
`stats/web/` is touched, so `## visual verification` does not apply.

- [x] 1. Introduce `harvest.Deps` and `harvest.NoDeps`, wired to nothing yet

**Build:** green
**Files:** `stats/internal/harvest/deps.go`, `stats/internal/harvest/deps_test.go`, `stats/cmd/flowd/main.go`
**Tests:** `TestNoDepsReturnsZeroValuesAndNoError`, `TestNoDepsIsUsableAsEveryConstituentInterface`
**Regression:** Revert this task's commit and `TestNoDepsReturnsZeroValuesAndNoError` disappears with it, so a later edit making a `NoDeps` method return a non-nil error — which would make every option-free test in task 2 fail for a reason unrelated to what it tests — has nothing holding it. `TestNoDepsIsUsableAsEveryConstituentInterface` likewise: without it, a `Deps` that has drifted from the four interfaces it composes is only caught when task 2's call sites stop compiling.
**Baseline:** before=140 after=142
<!-- measured: cd stats && go test ./internal/harvest/ -count=1 -v | grep -c '^=== RUN' @ merge-base 520e9f6 (the count BEFORE this change); the two figures count `=== RUN` lines, subtests included, not top-level functions -->
**Commit:** `feat(harvest): add the Deps interface and its NoDeps no-op`

This task is purely additive: the three `With*` options and the three `Has*` accessors stay
exactly as they are, `NewWatcher`'s signature does not move, and every existing call site
compiles unchanged. Task 2 is what switches over.

  - [x] **Step 1: write `deps_test.go` first, and watch it fail to compile.** Both tests name
    `harvest.Deps` and `harvest.NoDeps`, neither of which exists yet, so the package does not
    build. That is the intended red state for this task.

    `TestNoDepsReturnsZeroValuesAndNoError` calls every method of `NoDeps{}` with a
    `context.Background()` and asserts, per method, that the error is `nil` and the value return
    (where there is one) is the zero value — `nil` map from `UnresolvedSessionTokens`, `0` from
    `BindSession`, `nil` slice from `PersistedGiveUps`, `nil` slice from
    `DispatchWindowsForSession`. It is a real assertion, not a compile check: the whole premise of
    task 2's 40 rewritten call sites is that `NoDeps` is inert, and a method quietly returning an
    error would fail them all with a misleading message.

    `TestNoDepsIsUsableAsEveryConstituentInterface` assigns `NoDeps{}` to a variable of each of
    the four constituent interface types in turn, so a `Deps` that has drifted from what it
    composes is caught here rather than at task 2's call sites.

  - [x] **Step 2: add `stats/internal/harvest/deps.go`.** A new file rather than more of
    `watcher.go`, which is already ~1400 lines and whose top half is the four interfaces this one
    composes.
    <!-- measured: wc -l stats/internal/harvest/watcher.go @ merge-base 520e9f6 -->

```go unverified:confirm every method signature against internal/harvest/watcher.go and attribute.go before relying on this block
// Deps is everything a Watcher needs beyond its root, its sink and its
// Attributor. It is one required parameter rather than a set of
// functional options (KAN-173): each of the four interfaces it composes
// is optional in a type signature but mandatory in practice --
// production supplies exactly one real implementation of each, and all
// four come from the same *store.Store -- so an omitted option
// compiled, tested green, and ran inert. Twice (KAN-16, KAN-172).
//
// Composing the four rather than restating their methods is what makes
// a fifth dependency a compile error too: adding a method to any
// constituent breaks every implementation that has not grown it.
type Deps interface {
    Pricer
    SessionTokenBinder
    DispatchMetricsSink
    DispatchWindowSource
}

// NoDeps satisfies Deps with a no-op for every method: zero values,
// nil errors, nothing recorded. It is exported for tests -- a test
// that needs no dependency passes NoDeps{}, and a test that needs one
// embeds NoDeps and overrides that single method.
//
// A daemon must never wire this. cmd/flowd/wiring_test.go's
// TestDaemonWiresTheRealStore is what says so.
type NoDeps struct{}
```

    Then one method per interface member, each returning the zero value and `nil`. The full
    member list, to be read off the source rather than off this plan: `Price`;
    `UnresolvedSessionTokens`, `BindSession`, `RecordSessionTokenGiveUp`, `PersistedGiveUps`,
    `MarkDispatchesUnattributedByID`, `MarkDispatchesUnattributed`; `MergeDispatchMetrics`;
    `DispatchWindowsForSession`.

  - [x] **Step 3: add the compile-time assertion in `stats/cmd/flowd/main.go`.** Beside the
    existing compile-time checks the file already carries for `HarvestSink` and the rest:

```go unverified:confirm the existing compile-time assertions' exact form and placement in cmd/flowd/main.go
var _ harvest.Deps = (*store.Store)(nil)
```

    This is what proves the premise of decision `deps-interface` in `design.md` — that
    `*store.Store` already satisfies all four whole, so the daemon has one value to pass and no
    adapter to write. It belongs in this task, not task 2, because it must hold *before* the
    signature changes: if the store does not in fact satisfy `Deps`, that is discovered here with
    nothing yet rewritten.

  - [x] **Step 4: verify.** `cd stats && go test ./internal/harvest/ ./cmd/flowd/ -race -count=1`
    passes, and `cd stats && go vet ./... && gofmt -l .` is clean (run `gofmt -w .` first).

- [x] 2. Make `deps` a required `NewWatcher` parameter and delete the option pattern

**Build:** green
**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`, `stats/internal/harvest/endtoend_test.go`, `stats/cmd/flowd/main.go`, `stats/cmd/flowd/wiring_test.go`
**Tests:** `TestNewWatcherPanicsOnNilDeps`
**Regression:** Revert this task's commit and `TestNewWatcherPanicsOnNilDeps` goes with it, leaving a nil `deps` free to reach the `Watcher` and reinstate the exact silent-nil path — a pricer that never prices, a binder that never binds — that KAN-16 and KAN-172 both were. The 42 rewritten `watcher_test.go` call sites and the 2 in `endtoend_test.go` are this task's other net: they are the existing behavioural suite, and they only compile against the new signature, so a revert of the signature alone does not leave them passing.
**Baseline:** before=142 after=143
<!-- measured: cd stats && go test ./internal/harvest/ -count=1 -v | grep -c '^=== RUN' @ branch spectre/kan-173-make-the-daemon-s-dependencies-required-so-a, after task 1 -->
**Commit:** `refactor(harvest): require the watcher's dependencies as one Deps parameter`

The whole switch is one commit because it does not decompose into green steps: removing an
option breaks every caller, and rewriting a caller before the signature moves does not compile.

  - [x] **Step 1: write `TestNewWatcherPanicsOnNilDeps` first, and watch it fail.** It calls
    `harvest.NewWatcher(t.TempDir(), sink, harvest.NewAttributor(nil), nil, nil)` inside a
    `defer func(){ recover() }()` and fails if no panic occurred, asserting the recovered value's
    string names `deps`. Against the current signature it does not compile; against a signature
    that takes `deps` but does not check it, it fails with no panic. Both are the intended red.

  - [x] **Step 2: change the signature and add the panic** in
    `stats/internal/harvest/watcher.go`:

```go unverified:confirm the field names on Watcher and the exact make() initialisers already in NewWatcher
func NewWatcher(root string, sink HarvestSink, attributor *Attributor, deps Deps, logger *slog.Logger) *Watcher {
    if deps == nil {
        panic("harvest.NewWatcher: deps is nil; pass harvest.NoDeps{} to opt out explicitly")
    }
    ...
}
```

    `deps` sits before `logger` because `logger` is documented as nil-able and reads as the
    trailing optional it is; `deps` is the opposite and must not be adjacent to it. Store `deps`
    on the `Watcher` as a single `deps Deps` field, replacing `pricer`, `sessionTokens` and
    `dispatchMetrics`. Build the `*DispatchAttributor` inside the constructor —
    `dispatchAttributor: NewDispatchAttributor(deps)` — per decision `attributor-built-inside`:
    there is then nothing for a caller to omit.

  - [x] **Step 3: delete the option pattern.** Remove `WatcherOption`, `WithPricer`,
    `WithSessionTokenBinder`, `WithDispatchAttribution`, `HasPricer`, `HasSessionTokenBinder` and
    `HasDispatchAttribution`, and delete `TestNewTranscriptWatcherWiresBinderAndPricer` from
    `stats/cmd/flowd/wiring_test.go` — it asserts on accessors that no longer exist. Task 3
    supplies its replacement; this task leaves the file's other four tests untouched.

  - [x] **Step 4: delete the nil guards in `RunOnce` and its callees.** Every
    `if w.pricer != nil`, `if w.sessionTokens != nil` and `if w.dispatchAttributor != nil`
    guard becomes unconditional — `deps` cannot be nil, by step 2. Read each guard before
    deleting it: a guard whose condition is about something other than "was this dependency
    configured" stays. The `dispatchMetrics`/`dispatchAttributor` pairing that
    `WithDispatchAttribution` grouped is now structurally impossible to half-configure, so its
    two-field check has nothing left to check.

  - [x] **Step 5: rewrite the 44 test call sites.** In `stats/internal/harvest/watcher_test.go`,
    the 40 option-free calls take `harvest.NoDeps{}`:

```go unverified:confirm the exact argument list at each site; the sink and attributor expressions differ between them
w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)
```

    The 2 that pass `harvest.WithPricer(pricer)` embed instead:

```go unverified:confirm fakePricer's method set matches Pricer exactly
type pricingDeps struct {
    harvest.NoDeps
    p harvest.Pricer
}

func (d pricingDeps) Price(ctx context.Context, stageRunID int64) error {
    return d.p.Price(ctx, stageRunID)
}
```

    In `stats/internal/harvest/endtoend_test.go`, both sites pass binder and pricer and both are
    the real `st`, so both collapse to passing `st` directly as `deps` — the same value the
    daemon passes, which is the point.

  - [x] **Step 6: rewrite the daemon's wiring** in `stats/cmd/flowd/main.go`:

```go unverified:confirm newTranscriptWatcher's current parameter list before editing
func newTranscriptWatcher(root string, st *store.Store, attributor *harvest.Attributor, logger *slog.Logger) *harvest.Watcher {
    return harvest.NewWatcher(root, st, attributor, st, logger)
}
```

    Rewrite the function's doc comment: its current text explains why the `Has*` accessors exist
    and what KAN-172's dropped option was, and both of those are gone. Keep the KAN-172 and
    KAN-16 history — it is the reason the signature is shaped this way — and say what now
    enforces it.

  - [x] **Step 7: verify.** `cd stats && go test ./... -race -count=1` passes with no test
    skipped or deleted beyond `TestNewTranscriptWatcherWiresBinderAndPricer`, and
    `cd stats && go vet ./... && gofmt -l .` is clean (run `gofmt -w .` first).

- [x] 3. Assert the daemon wires the real store, not `NoDeps`

**Build:** green
**Files:** `stats/cmd/flowd/wiring_test.go`
**Tests:** `TestDaemonWiresTheRealStore`
**Regression:** Revert this task's commit and a daemon built with `harvest.NoDeps{}` in place of `st` compiles, runs, harvests every transcript, and prices nothing, binds nothing and charges no dispatch — silently and with every test green. That is precisely the KAN-172 failure shape surviving task 2's compile-time guarantee, and this test is the only thing in the repository that fails on it.
**Baseline:** before=5 after=5
<!-- measured: cd stats && go test ./cmd/flowd/ -count=1 -v | grep -c '^=== RUN' @ merge-base 520e9f6; the count is unchanged because this task's test replaces the one task 2 deleted -->
**Commit:** `test(flowd): assert the daemon wires the real store, not NoDeps`

Task 2's compile error covers a *dropped* argument. It cannot cover a *wrong* one, and `NoDeps`
is the wrong argument that most resembles a right one. This is that gap and only that gap —
`design.md`'s **The honest limit** states what remains uncovered after it.

  - [x] **Step 1: write the test, and watch it fail.** Prove it fails by temporarily editing
    `newTranscriptWatcher` to pass `harvest.NoDeps{}` instead of `st`; the test must report the
    daemon wired `NoDeps`. Revert that edit and the test must pass. Record both observations in
    the test's own doc comment, the way `TestNewTranscriptWatcherWiresBinderAndPricer` recorded
    the failure it was written against.

```go unverified:confirm reflect can read the unexported deps field's dynamic type from package main; FieldByName returns the zero Value if the field name is wrong, which IsValid must catch
func TestDaemonWiresTheRealStore(t *testing.T) {
    var st *store.Store // never dereferenced: newTranscriptWatcher only stores it behind Deps and HarvestSink.
    w := newTranscriptWatcher(t.TempDir(), st, harvest.NewAttributor(nil), nil)

    f := reflect.ValueOf(w).Elem().FieldByName("deps")
    if !f.IsValid() {
        t.Fatal("harvest.Watcher has no deps field: this test no longer checks anything")
    }
    if f.IsNil() {
        t.Fatal("the daemon's watcher has a nil deps")
    }
    if f.Elem().Type() == reflect.TypeOf(harvest.NoDeps{}) {
        t.Error("the daemon wired harvest.NoDeps: nothing is priced, no session token is bound, no dispatch is charged")
    }
}
```

    The `IsValid` check is not decoration: `FieldByName` on a renamed field returns the zero
    `Value` and every later check would pass vacuously, which is the failure mode a reflection
    test has and a compile-time one does not. `f.Elem()` is what unwraps the interface to its
    dynamic type; comparing `f.Type()` instead would compare `harvest.Deps` and never match.

  - [x] **Step 2: verify.** `cd stats && go test ./cmd/flowd/ -race -count=1` passes and reports
    5 `=== RUN` lines, and `cd stats && go vet ./... && gofmt -l .` is clean.
