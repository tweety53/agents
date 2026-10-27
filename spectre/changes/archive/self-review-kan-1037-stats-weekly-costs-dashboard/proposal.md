# self-review-kan-1037-stats-weekly-costs-dashboard

## Why

The self-review pass of `/flow` change `kan-1037-stats-weekly-costs-dashboard` found eight non-`big`
defects:

- F1: `run-verify-list.sh` ends an all-cache-skipped test group's row `skipped`, and
  `check-unfinished-work.sh` signal seven accepts only a `completed` verify row, so a verify that
  skipped every unchanged test never clears the gate; `verify-and-handoff.md` still says the parent
  ends a failed group's row by hand.
- F2: `RecordDispatch` is idempotent on (change, token, key) and keeps the first `startedAt`, so a
  verify re-run under a reused key never moves past a later code commit.
- F3: `TestProveReproducerSignals` opens its release FIFO with a blocking `O_WRONLY` after the run
  ended; with no reader left the open blocks until the test timeout.
- F4: the panel's mutation slot symlinked `stats/web/node_modules` into its throwaway copy, and the
  copy's `npm ci` emptied the shared directory through the symlink.
- F5: `check-cleanup-complete.sh` matches only `<name>-wave-group-<n>` detached copies, so a
  leftover `<name>-mutation-<n>` throwaway copy passes run 2.
- F6: flow-guard leaves SIGPIPE at Go's default, so `run-verify-list.sh | head` kills the list
  mid-run and leaves its verify row unended.
- F7: `flow stage` takes the change only as a trailing positional while every `flow record` takes
  `-change`.
- F8: `stage_runs.stage_key` (migration 0008) is never written or read, and misled kan-1037's design.

## What changes

- Signal seven accepts a `skipped` verify row; `verify-and-handoff.md` states the script's own end
  outcomes and that a re-run after a new code commit uses a fresh `verify-<k>` key; signal seven's
  reason names the fresh key.
- The prove-reproducer test releases its FIFO without blocking.
- The throwaway-copy guidance forbids symlinking a shared `node_modules`; `stats/Makefile` refuses
  `npm ci` over a symlinked `web/node_modules`.
- The cleanup guard also matches `<name>-mutation-<n>` detached copies.
- `runVerifyList` ignores SIGPIPE.
- `flow stage` accepts `-change <name>` in place of the positional.
- A new migration drops `stage_runs.stage_key` and its index.

## Impact

`stats/internal/guard`, `stats/cmd/flow`, `stats/internal/store`, `stats/Makefile`,
`skills/flow/verify-and-handoff.md`, `skills/flow/review-panel-optional-slots.md`.
