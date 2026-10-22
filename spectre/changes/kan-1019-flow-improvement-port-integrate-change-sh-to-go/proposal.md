# kan-1019-flow-improvement-port-integrate-change-sh-to-go

## Why

`scripts/integrate-change.sh` runs every mechanical step of `/flow`'s run 1 and run 2 — prepare,
commit, land, cleanup — and is still bash, with its own bash harness
`scripts/test-integrate-change.sh`. The project rule puts guard logic in Go in `flow-guard`; this
is the largest remaining bash orchestrator and the highest-risk port (git rewrites, pushes,
worktree removal) of the bash → Go guard migration (KAN-1019).

## What changes

- The four subcommands run in `flow-guard` as `integrate-change`; `scripts/integrate-change.sh`
  becomes a thin `flow_guard_exec` shim. Every output line (`NEXT:`, `STOP:`, `JIRA:`, the
  per-step verdict lines, the relayed guard lines) and every exit code is unchanged byte for byte.
- `scripts/test-integrate-change.sh` is deleted; Go table tests carry every one of its cases.
- The bash at the base commit is the parity oracle: the unchanged bash harness passes against the
  shim before it is deleted.
