# kan-173-make-the-daemon-s-dependencies-required-so-a

Jira: [KAN-173](https://tweety53.atlassian.net/browse/KAN-173)

## Why

`internal/harvest.NewWatcher` takes its four production dependencies as functional options —
optional in the type signature, mandatory in practice. An omitted option compiles, tests green,
and runs inert. It has happened twice: KAN-16 (`Store.Price` never ran in the daemon, so every
currency figure was permanently unavailable) and KAN-172 (`cmd/flowd` built the harvester with no
`WithSessionTokenBinder`, so no stage run could ever bind — 329 tests green, review panel clean).
<!-- measured: cannot be re-run — quoted from KAN-172's own recorded count at the moment its defect was found -->
Both were found by running the daemon. The defect is a line that is *not there*, which is exactly
what a diff review cannot see.

## What changes

- The four dependencies collapse into one required `harvest.Deps` parameter. Dropping it is a
  compile error; `*store.Store` already satisfies it whole.
- `WatcherOption`, `WithPricer`, `WithSessionTokenBinder`, `WithDispatchAttribution`, `HasPricer`,
  `HasSessionTokenBinder` and `HasDispatchAttribution` are removed, along with `RunOnce`'s
  `!= nil` guards on those fields.
- `harvest.NoDeps` is exported for tests that need no dependencies, or one.
- `cmd/flowd/wiring_test.go` asserts the daemon wires the real store rather than `NoDeps` — the
  one hole the compiler cannot close.

No observable behaviour changes for a correctly-wired daemon. The design, its rejected
alternatives and the honest limit of what this closes are in `design.md`.
