# KAN-173 — Make the daemon's dependencies required, so a dropped wiring is a compile error

## Context

`internal/harvest.NewWatcher` takes its four production dependencies as functional options.
All four are optional in the type signature and mandatory in practice: production supplies
exactly one real implementation of each, and every one of them comes from the same
`*store.Store`.

An omitted option is therefore a defect that compiles, tests green, and runs inert. It has
happened twice:

- **KAN-16** — `Store.Price` was called only from tests. The pricing table was never seeded
  and `Price` never ran in the daemon, so every currency figure in every view was permanently
  unavailable while the whole suite passed.
- **KAN-172** — `cmd/flowd` built the harvester with `harvest.WithPricer(st)` and no
  `harvest.WithSessionTokenBinder(st)`. The binder stayed nil, no stage run could ever bind,
  and the change's entire purpose was inert with 329 tests green and a review panel clean.
  <!-- measured: cannot be re-run — the figure is KAN-172's own recorded count at the moment its defect was found, quoted from that issue rather than taken here -->

Both were found by running the daemon. A review reads a diff, and in both cases the defect is
a line that is *not there*.

## What the code holds today

Three options, four dependency fields:

| Option | Fields set | Production value |
|--------|-----------|------------------|
| `WithPricer` | `pricer Pricer` | `st` |
| `WithSessionTokenBinder` | `sessionTokens SessionTokenBinder` | `st` |
| `WithDispatchAttribution` | `dispatchAttributor *DispatchAttributor`, `dispatchMetrics DispatchMetricsSink` | `NewDispatchAttributor(st)`, `st` |

Three accessors — `HasPricer`, `HasSessionTokenBinder`, `HasDispatchAttribution` — exist only
so `cmd/flowd/wiring_test.go` can assert on the constructed value rather than on `main.go`'s
source text. That pattern grows one predicate per option forever.

44 `NewWatcher(` call sites live in tests. `watcher_test.go` holds 42 of them — 18 passing no
option at all, 4 passing `WithPricer` alone, 18 passing `WithSessionTokenBinder` alone, and 2
passing binder together with `WithDispatchAttribution`; `endtoend_test.go` holds the other 2, both
passing binder and pricer against the real store.
<!-- measured: git show 520e9f6:stats/internal/harvest/watcher_test.go and :endtoend_test.go, then grep 'NewWatcher(' with per-option filters, counting the two multi-line WithDispatchAttribution calls by hand @ merge-base 520e9f6 (the split BEFORE this change) -->
An earlier draft of this section recorded that split as 40/2/2. That was wrong — it came from a
single-line `grep` that could not see an option written on a continuation line — and it is
corrected here rather than quietly replaced, because task 2 was planned against it: the plan's
`unverified:` sketch anticipated one composite `Deps` wrapper for the pricer sites and no other,
where the real shape needs three (`pricingDeps`, `sessionBinderDeps`,
`sessionBinderAndDispatchDeps`).

**No test injects a `*DispatchAttributor` through `NewWatcher`** — the 2 sites naming
`WithDispatchAttribution` pass one built inline from the same fake window source, and
`attribute_test.go` drives `DispatchAttributor` directly. Decision `attributor-built-inside` below
survives that correction: those 2 sites lose nothing by having the attributor built from their
`Deps`, since the source it wraps is what they actually vary.

## Decisions

### The four dependencies collapse into one required `Deps` parameter

**ID:** deps-interface
**Status:** active
**Chosen:** a single `Deps` interface, composed from the four existing interfaces, taken as a
required positional parameter — `NewWatcher(root, sink, attributor, deps, logger)`. `*store.Store`
already satisfies all four whole, so the daemon passes `st`. Dropping the argument is a compile
error; adding a fifth dependency later means adding a method to `Deps`, which is a compile error
for anything that does not satisfy it. The parameter list does not grow.
**Considered:**
- *Required positional parameters, one per dependency* (the issue's own proposal) — gives the
  same compile error, but an eight-parameter signature that grows with every future dependency,
  and edits all 44 call sites to pass four explicit no-ops rather than one.
- *A reflection-based test over the constructed `*Watcher`, options kept* (the primary reviewer's
  cheaper complementary form) — catches a fourth option added and never wired, but stays a test
  failure rather than a compile error, and leaves the option pattern in place.
- *A `Deps` struct rather than an interface* — a dropped **field** is a zero value, not a compile
  error, so it does not close the failure this change exists to close.

### `HarvestSink` stays a separate parameter

**ID:** sink-stays-separate
**Status:** active
**Chosen:** `Deps` carries the four dependencies only. The sink remains its own parameter.
**Considered:** *folding `HarvestSink` into `Deps`* — four parameters instead of five, and the
daemon passes `st` exactly once. Rejected because the 44 tests each inject a *distinct* fake
<!-- measured: grep -c "NewWatcher(" stats/internal/harvest/watcher_test.go stats/internal/harvest/endtoend_test.go @ merge-base 520e9f6, 42 + 2 -->
sink (failing commits, offset outages, atomicity probes); that variation is the whole point of
the sink parameter, where `deps` is a uniform no-op in almost all of them. Folding would force
every test to wrap its fake sink in an embedding struct for no gain.

### `NewWatcher` panics on a nil `deps`

**ID:** nil-deps-panics
**Status:** active
**Chosen:** panic immediately, naming the parameter and pointing at `NoDeps{}`. A nil interface
here is a programming error at wiring time, not a runtime condition — it fails at daemon start,
loudly, before any transcript is read. This is what lets every `if w.pricer != nil` guard inside
`RunOnce` be deleted.
**Considered:**
- *Substitute `NoDeps{}` silently* — re-creates exactly the failure this change closes: a
  dropped wiring runs inert and green.
- *Keep the nil guards and accept nil deps* — the compile-time guarantee would be the only thing
  gained, with the nil-silence path surviving underneath it.

### Tests supply one exported no-op, embedded

**ID:** exported-nodeps
**Status:** active
**Chosen:** `harvest` exports a zero-value `NoDeps` type satisfying every `Deps` method as a
no-op — zero values and nil errors throughout. The 40 option-free call sites become a one-token
edit; a test wanting one real dependency embeds `NoDeps` and overrides that method.
**Considered:**
- *An unexported per-package test helper* — keeps the API surface smaller, but `cmd/flowd` and
  any future package need their own copy.
- *Passing `nil` for unused dependencies* — smallest diff, but re-opens the nil-dereference risk
  `RunOnce` guards against today, and contradicts `nil-deps-panics`.

### The `*DispatchAttributor` is constructed inside `NewWatcher`

**ID:** attributor-built-inside
**Status:** active
**Chosen:** `NewWatcher` builds it from `deps` (which satisfies `DispatchWindowSource`). The
daemon cannot drop it, because there is nothing to pass.
**Considered:** *keeping it a caller-supplied value* — rejected because no test injects one
through `NewWatcher` today, so nothing loses reach, and a caller-supplied value is one more
thing a wiring site can omit.

### The `Has*` accessors are removed, and `cmd/flowd`'s wiring test is rewritten

**ID:** wiring-test-asserts-real-store
**Status:** active
**Chosen:** delete `HasPricer`, `HasSessionTokenBinder` and `HasDispatchAttribution`, and replace
`TestNewTranscriptWatcherWiresBinderAndPricer` with a test asserting the daemon's `deps` is not
`NoDeps`. Route `deps-interface` leaves exactly one hole the compiler cannot close: a daemon
passing `harvest.NoDeps{}` compiles and runs inert. This test closes it.
**Considered:**
- *Deleting the wiring test outright* — honest about what the compiler guarantees, but leaves
  that hole uncovered.
- *A reflection test over every pointer/interface field of `*Watcher`* — catches a future field
  added outside `Deps` and never populated, but does not distinguish a real store from `NoDeps`,
  which is the residual failure that actually remains.
- *One exported `func (w *Watcher) Deps() Deps` accessor, with the test doing a plain type
  assertion and no `reflect` at all* — simpler, and not coupled to a private field name the way
  reading `deps` by reflection is. Ruled out because it reopens the very public surface this
  change shrinks: `wiring-test-asserts-real-store` deletes `HasPricer`,
  `HasSessionTokenBinder` and `HasDispatchAttribution` precisely to stop `Watcher` growing one
  accessor per dependency, and trading three accessors for one is a smaller surface, not none.
  The reflection route's own failure mode — a renamed field making every later assertion pass
  vacuously — is closed by the `IsValid` guard, which `t.Fatal`s rather than passing; that guard
  was mutation-tested during this change's review panel and does fail on a rename.

  Recorded here after the review panel's principles slot raised its absence (F1): the shipped
  choice was judged defensible, but this was the alternative actually foregone, and the record
  named only the two above.

## The honest limit

Stated so this change is not over-sold. None of the routes considered catches:

- a dependency wired to a non-nil but **wrong** value — the `NoDeps` check narrows this to
  "wrong, and not `NoDeps`", it does not eliminate it;
- an omission **outside** `*Watcher` entirely — a second constructor the daemon should call and
  does not.

Only a behavioural end-to-end test catches those. That is KAN-168, and this change does not
claim its ground.

## Open questions

None.
