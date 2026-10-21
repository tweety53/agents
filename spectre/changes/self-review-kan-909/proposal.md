# self-review-kan-909

## Why

KAN-909 (flow-fix): the `stats/internal/setuptest` harness's SIGINT/SIGTERM handler stops and
waits for `setup.sh` runs only. Every test-side sandbox write — group directories, seeded files,
copies, logs, the shared fixture repo — keeps going and can race `cleanup`'s `RemoveAll`, and a
`runSetup` starting while the handler sits in `running.Wait()` calls `running.Add(1)` from a zero
counter concurrently with `Wait`, which breaks the `sync.WaitGroup` contract.

## What changes

- A `gate` in the harness admits sandbox writers until it is closed, and `close` waits for every
  admitted writer to leave. It replaces the `running` WaitGroup.
- Every assertion group holds the gate for its whole life (`newGroup` enters, `t.Cleanup`
  leaves; a group refused entry skips as interrupted), and `runMain` holds it while building the
  shared fixture repo — so nothing writes the sandbox while the handler removes it, and nothing
  starts writing after the handler closes the gate.
- The handler closes the gate between `stopRuns()` and `cleanup()`.
- `TestWritersGateHoldsCloseUntilEveryWriterLeaves` pins the gate deterministically under
  `testing/synctest`, with no timing window.
