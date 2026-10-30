# pipeline-fix-kan-1060-flow-cost-visual-verify-now-costs-more-than-48

## Why

`TestEnforceAgentBaselineHookRefusesOffPolicyDispatch` case "Explore at its effort" failed inside a
`/flow` inline verify and passes in isolation. `runHook` inherits `os.Environ()`; the verify ran with
`FLOW_RECORDS_ADDR=http://127.0.0.1:4173` (the live dev flowd) exported by `prepare-workspace.sh`, so
the hook's `policyTable` overlaid the live store's open allowed-pairs experiment, which overrides
`explorer` to `sonnet/Explore@low`. Hook tests read live operator state.

## What changes

- `stats/internal/repotest/helpers_test.go` `runHook` pins `FLOW_ADDR` and `FLOW_RECORDS_ADDR` to the
  dead `http://127.0.0.1:1` ahead of the caller's env, so no hook test reaches a live store; callers
  that pass a stub's address still override it.

## Impact

Test harness only. No hook or production code changes.
