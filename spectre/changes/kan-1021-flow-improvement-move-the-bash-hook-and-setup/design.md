## Context

Operator answers, 2026-10-23: all nine harnesses in one change, `test-protect-main-checkout.sh` first; hooks stay Python; each `.sh` deleted once ported. `setup.sh` stays bash — it runs before the Go binaries exist. `go test ./...` in `.flow/project.md`'s `## test` already runs every new package.

## Decisions

### Hooks stay Python; tests exec them

**ID:** tests-only
**Status:** active
**Chosen:** Go tests exec `python3 <repo>/hooks/<hook>.py` with stdin payloads, as the bash harnesses do — smallest diff, no install change.
**Considered:** porting all four hooks to `flow-guard` subcommands — rewrites `setup.sh`'s hook registration for Claude Code and ZCode, far wider blast radius; porting only `protect-main-checkout.py` — same install churn for one hook.

### Delete each harness, no shim

**ID:** delete-not-shim
**Status:** active
**Chosen:** delete every ported `scripts/test-*.sh`, `test-setup.sh` included; `go test ./...` is the one runner.
**Considered:** a thin `go test` shim per harness (KAN-922/923 convention) — keeps `run-guard-tests.sh` finding them, but runs every test twice per `## test` pass.

### One new package for repo-level harnesses

**ID:** repotest-package
**Status:** active
**Chosen:** `stats/internal/repotest/`, one `_test.go` per harness; `test-setup-agents.sh` joins `stats/internal/setuptest/`, which already sandboxes `setup.sh global`.
**Considered:** spreading tests by subject (`guard/` for git-config pins, `cmd/flow/` for flow-addr) — the harnesses audit source trees and exec hooks, not those packages' code; a package per harness — eight packages for no isolation gain.

### Case parity

**ID:** case-parity
**Status:** active
**Chosen:** one Go subtest per bash harness case, `t.Parallel()`, `t.TempDir()` sandboxes, never the real `$HOME`; `python3`, `git`, `make`, `node` exec'd as the bash did (KAN-922's `carry-923-conventions`).
**Considered:** rewriting cases to new assertions — loses the proof that coverage did not drop.

### pipeline-fix selects packages, not scripts

**ID:** pipelinefix-packages
**Status:** active
**Chosen:** `pfGoAudits` becomes one `cd stats && go test ./internal/repotest/ -run '^Test(GitConfigPins|FlowAddrDeclaration)$'` line; a touched `hooks/*` selects `go test ./internal/repotest/`, a touched `setup.sh` `go test ./internal/setuptest/`.
**Considered:** mapping each hook to its own `-run` filter — a per-hook table nobody needs; the package runs in seconds.

### Fixture strings left alone

**ID:** fixture-strings
**Status:** active
**Chosen:** `scripts/test-setup.sh` inside guard/cmd test fixtures (`check_workspace_isolation_test.go`, `prepare_workspace_test.go`, `record_test.go`) stays — arbitrary fixture text, never executed.
**Considered:** renaming them — churn with no behaviour.

## Open questions
