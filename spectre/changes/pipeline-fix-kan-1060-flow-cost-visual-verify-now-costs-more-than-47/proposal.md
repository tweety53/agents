# pipeline-fix-kan-1060-flow-cost-visual-verify-now-costs-more-than-47

## Why

`TestEnforceLoopDispatchHookRecordsADenial` fails intermittently under the full parallel
`go test -race -count=1 ./...` load and passes in isolation: the deny is written but the incident
stub logs no call. `hooks/enforce-loop-dispatch.py` bounds its `flow record incident` subprocess at
10 s and swallows `TimeoutExpired`; a stub starved of CPU past that bound is killed before it logs.
`hooks/enforce-agent-baseline.py` had the same failure and gained a `FLOW_HOOK_UNBOUNDED=1` test
seam in kan-1033; the loop hook never did, and the baseline hook's own incident cases never pass it.

## What changes

- `hooks/enforce-loop-dispatch.py` wraps its three wall-clock bounds in the same `bound()` the
  baseline hook uses, lifted by `FLOW_HOOK_UNBOUNDED=1`.
- `stats/internal/repotest/helpers_test.go` `runHook` injects `FLOW_HOOK_UNBOUNDED=1` for every hook
  test, so no hook case depends on wall-clock time.

## Impact

Test harness and one hook's test seam. Production never sets `FLOW_HOOK_UNBOUNDED`, so every bound
production relies on is unchanged.
