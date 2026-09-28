# kan-844-agents-speed-up-scripts-test-setup-sh

## Why

`scripts/test-setup.sh`, the `setup.sh` regression harness, is the slowest harness in
`scripts/run-guard-tests.sh` and bounds the guard suite's wall time (KAN-841/842 self-reviews name
it the next slice). Its 26 assertion groups already isolate on their own sandbox HOME, yet run one
after another, and almost all of the time goes to spawning `setup.sh` and small tools (KAN-844).

Implementer effort tops out at `high`. The harness now exposes `xhigh`, and the operator wants the
plan able to put implementers there (added during this run).

## What changes

- `scripts/test-setup.sh` becomes a thin shim over a Go test package,
  `stats/internal/setuptest/`, whose groups run concurrently as parallel top-level tests. Every
  assertion keeps its description; the real-HOME and source-tree containment checks span every
  group; fingerprints and assertions run in-process. A new leak-detection group proves both
  containment checks still fail when a write leaks. Target: ≥ 2× faster than the post-KAN-843
  baseline, measured and recorded in `design.md`.
- A fourth agent definition, `agents/flow-xhigh.md`. The Decide step may give the implementer
  pair and each implementer group `xhigh`; `flow record` accepts `-effort xhigh`.

Out of scope: speeding up `setup.sh` itself; sharing installs between groups; `xhigh` for the
fixer, panel dispatches or rerun pair; `max` effort.
