# kan-1021-flow-improvement-move-the-bash-hook-and-setup

## Why

Ten `scripts/test-*.sh` harnesses still test the Python hooks, `setup.sh` and repo wiring in bash, while the guard suite moved to Go in KAN-918..923. Two test runners and two idioms for one suite; bash assertions are the slowest and least readable part of it.

## What changes

- Nine bash harnesses become Go tests under `stats/internal/repotest/` (hooks, git-config pins, flow-addr declaration, make build, flow-task-list landed) and `stats/internal/setuptest/` (setup-agents); `go test ./...` runs them.
- Every ported `scripts/test-*.sh` is deleted, `test-setup.sh` included — no shim.
- `pipeline-fix`'s check selection runs the Go packages instead of the deleted scripts.
- The Python hooks and `setup.sh` are unchanged.
