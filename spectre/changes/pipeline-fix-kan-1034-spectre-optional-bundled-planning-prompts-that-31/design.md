# pipeline-fix-kan-1034-spectre-optional-bundled-planning-prompts-that-31

## Context

`flow state get` resolves its store from `FLOW_ADDR` (`stats/cmd/flow/state.go`, `resolveDefaultAddr`); the record family resolves `FLOW_RECORDS_ADDR` first, then `FLOW_ADDR` (`stats/cmd/flow/record.go`). In a workspace-isolated run the two differ, and the state record lives only in the dev store `FLOW_RECORDS_ADDR` names. `trPeerHeads` spawns its child without setting `cmd.Env`, so the child inherits the run's `FLOW_ADDR`.

## Decisions

### Where the address is redirected

**ID:** peer-state-read-follows-records-addr
**Status:** active
**Chosen:** the guard sets the child's `FLOW_ADDR` to `FLOW_RECORDS_ADDR` when that is set and non-empty, else leaves the environment as handed over — smallest change, confined to the one caller the defect names.
**Considered:** making `flow state get` itself honour `FLOW_RECORDS_ADDR` — rejected: it changes the resolution every state write and read in the run uses, including the worktree's own isolated state, a far wider blast radius than the one guard; passing `--addr` on the command line — rejected: `state get` exposes no such flag on every path, and an env override matches how the run already hands the address down.

### Planning effort for the first review

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — the fix is one environment line in one guard plus one test case.
**Considered:** flow-medium — rejected: nothing in the diff calls for wider reasoning.

## Open questions

## Recommended approach

In `trPeerHeads`, build the child environment from `env.Environ` (falling back to `os.Environ`, as `icChildEnv` in `stats/internal/guard/integratechange.go` does), and when `env.Getenv("FLOW_RECORDS_ADDR")` is non-empty append `FLOW_ADDR=<that value>` (the last duplicate wins for `exec`), then assign it to `cmd.Env`. The alternatives are rejected in `peer-state-read-follows-records-addr`.

## Critical files

- `stats/internal/guard/taskrecords.go` — `trPeerHeads`, the state read being fixed.
- `stats/internal/guard/check_task_records_test.go` — `TestCheckTaskRecords`, which gains the case proving the fix.

## Existing code to reuse

- `icChildEnv` (`stats/internal/guard/integratechange.go`) — the `env.Environ`-or-`os.Environ` child environment.
- The cases 17–19 stub-`flow` fixture in `TestCheckTaskRecords` (`stub`, `trRepo`, `pathEnv`).

## Verification

- `cd stats && go test ./internal/guard/ -run 'TestCheckTaskRecords' -count=1` passes, and fails with the `cmd.Env` assignment removed.
- In `/Users/tweety53/Projects/agents-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that`, with the fixed binary built: `FLOW_ADDR=http://127.0.0.1:6783 FLOW_RECORDS_ADDR=http://127.0.0.1:4173 scripts/check-task-records.sh` exits 0 (it exits 1 before the fix with `FLOW_ADDR` alone).
