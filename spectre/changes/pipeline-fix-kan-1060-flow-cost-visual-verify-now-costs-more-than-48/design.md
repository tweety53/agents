# Design — pipeline-fix-kan-1060-flow-cost-visual-verify-now-costs-more-than-48

## Context

`runHook` (`stats/internal/repotest/helpers_test.go`) builds the hook's env from `os.Environ()` and
appends the caller's entries last. The off-policy test passes only `FLOW_ADDR=http://127.0.0.1:1`, so
an inherited `FLOW_RECORDS_ADDR` still reaches whatever store it names. `policyTable`
(`stats/internal/guard/modelpolicy.go`) overlays `flow experiment overrides` from that store.

**Cause, established by forcing it** at `3deff9fa`, from `stats/`:
`FLOW_RECORDS_ADDR=http://127.0.0.1:4173 go test -count=1 -run TestEnforceAgentBaselineHookRefusesOffPolicyDispatch ./internal/repotest/`
fails; the same command with the variable unset passes.

Every other hook exec in `stats/internal/repotest/` goes through `runHook`; `sh` and the guard-cache
build run fixtures and `go build`, not hooks, and need no pin.

## Decisions

**ID:** pin-store-addrs-in-runhook
**Status:** active
**Chosen:** `runHook` appends `FLOW_ADDR=http://127.0.0.1:1` and `FLOW_RECORDS_ADDR=http://127.0.0.1:1`
after `os.Environ()` and before the caller's env. The pin replaces the inherited store address with a
dead port, so every hook case reads no live state; store-stub callers keep overriding `FLOW_ADDR`.
**Considered:** pinning per test — rejected: every hook test shares the leak and `runHook` is the one
place all of them route through. Unsetting the variables — rejected: an unset address falls back to
the default dev flowd.

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — a two-line test-harness diff with the cause forced and the proof recorded.
**Considered:** `flow-medium` — no production code moves, so nothing warrants it.

**ID:** fixer-grouping
**Status:** active
**Chosen:** one fixer for task 1, one commit, `opus` on `flow-low`: one file, one mechanism.

## Open questions

None.
