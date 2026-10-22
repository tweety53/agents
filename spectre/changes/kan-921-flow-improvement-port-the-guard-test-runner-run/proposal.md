# kan-921-flow-improvement-port-the-guard-test-runner-run

## Why

The guard test suite's runner, `scripts/run-guard-tests.sh`, is still bash logic
resting on two bash libraries, `scripts/lib/parallel.sh` and `scripts/lib/coverage.sh`,
each with its own bash harness. The project rule is that guard logic lives in Go in
`flow-guard`; this is the last runner-side bash of the bash → Go guard migration (inventory of
2026-10-07, KAN-921).

## What changes

- The runner's discovery, companion check, tree gate, longest-first ordering, bounded-parallel
  capture, list-order failure replay, aggregate exit code and duration file run in `flow-guard` as
  `run-guard-tests`; `scripts/run-guard-tests.sh` becomes a thin `flow_guard_exec` shim. Its
  output and exit codes are unchanged byte for byte.
- `scripts/lib/parallel.sh`, `scripts/lib/coverage.sh`, `scripts/test-lib-parallel.sh`,
  `scripts/test-lib-coverage.sh` and `scripts/test-run-guard-tests.sh` are deleted; Go table tests
  carry their cases.
- Every live mention of the deleted files names the Go file instead.
- `scripts/test-go-guards.sh` and `scripts/test-lib-flow-guard.sh` stay bash — they are the bash
  side of the Go bridge.
