# kan-842-agents-port-the-next-ten-slowest-bash-scripts-to

## Context

- Fourth slice of KAN-760's guard port; KAN-841's After judgement names its candidates. Base
  `c5379c0a`.
- Pattern established by KAN-760/KAN-778/KAN-841: a guard is one Go file registered in
  `guard.Registry` plus a shim calling `flow_guard_exec <name> <code> "<name>:" "$@"`; a shim whose
  Go guard defaults its root to the checkout also exports it (`scripts/check-references.sh` is the
  reference shim).
- Ranking (one `scripts/run-guard-tests.sh` run at base, per-harness `(Ns)`) excluded `setup.sh`
  (installer that must run before Go/flow-guard exist), Python-backed scripts
  (`compose-mockup-frames`, `generate-relocation-comparison`, `measure-visual-properties`,
  `check-plan-shape`, `check-task-records`, `check-task-build-green`, the
  `protect-main-checkout` hook), flow-guard's own harnesses (`test-go-guards.sh`,
  `test-lib-flow-guard.sh`) and `workspace.sh` (**Decision:** exclude-workspace-sh).
  <!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh, grep '(Ns)' sorted descending @ c5379c0a -->
- That base run exited 1 in the main checkout only: `TestCheckGuardSymlinks/7` validates the
  repository's own tree, and the main checkout carried an untracked
  `skills/flow/scripts/guard-autosquash.sh` that is not on the branch.
- Cannot-answer codes, from each header at base: 1 for `check-installed-rules` (it has no exit 2;
  every refusal is 1), 2 for the other nine.
  <!-- verified: grep -n 'exit [0-9]' and the header's exit-code paragraph of each script @ c5379c0a -->
- Sibling helpers and their other callers at base:
  - `lib/resolve-file.sh` (`check-workspace-isolation`) — Go twin `resolvefile.go` exists.
  - `lib/spec-root.sh` (`check-task-commit-planning-paths`) — Go twin `specroot.go` exists.
  - `lib/panel-touched-paths.sh` (`plan-class`, `check-panel-citation-trigger`) — also
    `check-panel-docs-only.sh`; no Go twin.
  - `lib/owned-corpus.sh` (`check-contract-budget`) — also `check-normative-inventory.sh`; no Go
    twin.
  - `lib/sha256-hex.sh` (`plan-class`) — no other bash caller; `sha256.go` mirrors it and
    `TestSHA256Hex` pins it against known vectors and `shasum`, not the bash library.
  - `plan-dispatch-bundles.sh`, `plan-dispatch-groups.sh` (`check-task-reviewer-single-dispatch`)
    and `check-visual-trigger.sh` (`check-visual-verify-dispatched`) — stay bash.
  - `flow record dispatches` (`check-task-reviewer-single-dispatch`,
    `check-visual-verify-dispatched`) — the store read.
  <!-- verified: grep -l 'lib/<name>.sh' scripts/*.sh excluding test-*, grep -n '$SCRIPT_DIR/' and 'flow record' of each script @ c5379c0a -->

## Measurements

### Parity floors at base

| Harness | floor | real (10 run concurrently) |
|---|--:|--:|
| `test-check-workspace-isolation.sh` | 152 | 230.32s |
| `test-check-task-reviewer-single-dispatch.sh` | 11 | 90.12s |
| `test-recover-guard-incident.sh` | 67 | 108.13s |
| `test-plan-class.sh` | 25 | 175.26s |
| `test-check-installed-rules.sh` | 20 | 246.20s |
| `test-resolve-base-branch.sh` | 39 | 101.20s |
| `test-check-visual-verify-dispatched.sh` | 18 | 98.05s |
| `test-check-contract-budget.sh` | 39 | 171.57s |
| `test-check-task-commit-planning-paths.sh` | 19 | 109.52s |
| `test-check-panel-citation-trigger.sh` | 24 | 86.72s |

- Floor is the harness's `^ok` line count on a green run; all ten exited 0.
- Load 32.23 36.07 26.92 after the concurrent run — other work on this machine inflated every
  `real` figure; the floors are unaffected.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p bash scripts/test-<name>.sh, the ten concurrently, grep -c '^ok' of each log, all exit 0 @ c5379c0a -->

Each Go test file carries at least its harness's floor in cases (`parity-by-case-count`).

### Suite before/after

Task 1 records **Before**; the last task records **After**. Shape as KAN-841's: x3 each, load
average read before each run, slowest five harnesses per run.

**Before** — `@ c5379c0a`, detached checkout without `guard-autosquash.sh`:

| Run | load (1/5/15) | suite real/user/sys | exit | harnesses | slowest five | Go pkg real/user/sys |
|---|---|---|--:|--:|---|---|
| 1 | 7.77 9.13 9.57 / Go 23.84 17.30 13.03 | 85.43 / 121.07 / 160.21 | 0 | 69 | go-guards 76s, setup 67s, lib-flow-guard 51s, check-plan-shape 47s, check-task-reviewer-single-dispatch 44s | 35.63 / 42.16 / 66.87 |
| 2 | 20.59 12.80 10.93 / Go 24.15 18.08 13.48 | 64.92 / 119.89 / 171.40 | 0 | 69 | setup 64s, go-guards 51s, generate-relocation-comparison 32s, check-plan-shape 31s, lib-flow-guard 26s | 40.00 / 42.51 / 67.14 |
| 3 | 27.07 15.99 12.24 / Go 25.43 19.22 14.11 | 69.20 / 122.65 / 174.43 | 0 | 69 | setup 67s, go-guards 56s, check-plan-shape 29s, lib-flow-guard 26s, measure-visual-properties 23s | 32.51 / 38.58 / 61.37 |
| **median** | | **69.20** | | | | **35.63** |

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p scripts/run-guard-tests.sh x3; cd stats && /usr/bin/time -p go test ./internal/guard/... -count=1 x3; sysctl -n vm.loadavg before each @ c5379c0a -->

**After** — `@ branch spectre/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to`, run
straight after Before on the same machine:

| Run | load (1/5/15) | suite real/user/sys | exit | harnesses | slowest five | Go pkg real/user/sys |
|---|---|---|--:|--:|---|---|
| 1 | 18.16 18.07 13.90 / Go 27.16 22.91 16.86 | 69.47 / 123.55 / 168.71 | 0 | 59 | go-guards 68s, setup 62s, compose-mockup-frames 24s, check-plan-shape 22s, lib-flow-guard 20s | 51.14 / 55.10 / 88.86 |
| 2 | 22.81 19.63 14.82 / Go 25.42 22.84 17.19 | 69.27 / 123.12 / 175.22 | 0 | 59 | setup 68s, go-guards 68s, compose-mockup-frames 49s, measure-visual-properties 31s, check-plan-shape 27s | 45.37 / 53.35 / 85.32 |
| 3 | 28.08 22.24 16.21 / Go 20.18 22.01 17.19 | 63.78 / 123.54 / 170.58 | 0 | 59 | go-guards 63s, setup 62s, compose-mockup-frames 45s, check-plan-shape 27s, lib-flow-guard 23s | 53.47 / 53.64 / 87.36 |
| **median** | | **69.27** | | | | **51.14** |

<!-- measured: the same commands x3 @ branch spectre/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to -->

Parity (`go test ./internal/guard/ -count=1 -v | grep -c -- '--- PASS: Test<Name>/'`, floor in
brackets): CheckWorkspaceIsolation 162 (152), CheckTaskReviewerSingleDispatch 27 (11),
RecoverGuardIncident 122 (67), PlanClass 31 (25), CheckInstalledRules 22 (20),
ResolveBaseBranch 52 (39), CheckVisualVerifyDispatched 41 (18), CheckContractBudget 45 (39),
CheckTaskCommitPlanningPaths 19 (19), CheckPanelCitationTrigger 25 (24) — every port at or above
its floor.

<!-- measured: cd stats && go test ./internal/guard/ -count=1 -v, exit 0, grep -c per port @ branch spectre/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to -->

Judgement: `suite-median-below-before` **not met** (69.27s against 69.20s);
`guard-package-under-40s` **not met** (51.14s median). Load was 18–28 through every run.
The suite's wall is bound by `test-setup.sh` and `test-go-guards.sh`, both 62–68s, which the ten
retired harnesses never were, so removing them left the median unmoved; `test-go-guards.sh` runs
the Go package, which grew with the ports, so the Go package is now the suite's critical path.
Slowest remaining harness: `test-go-guards.sh`. Next slice: the Go package's wall time and
`test-setup.sh`, not further bash ports.

## Decisions

### Carry the prior slices' port decisions unchanged

**ID:** carry-prior-port-decisions
**Status:** active
**Chosen:** KAN-760's `flow-guard-binary-rationale-in-go`, `go-tests-replace-harnesses`,
`parity-by-case-count`, `inject-deadlines-in-process`, `helper-parity-tests`,
`port-base-moves-into-go`, `guard-binary-built-from-checkout`, KAN-778's
`deferred-findings-to-known-bugs`, `drop-full-rerun-policy` and KAN-841's
`shared-helper-go-twins`, `sole-user-helpers-move-into-go`, `exec-unported-siblings` govern this
slice as written; the base for `port-base-moves-into-go` is `c5379c0a` — the operator's
instruction (recommended options).
**Considered:** re-deciding each per script — no script here raises a case those decisions do not
cover.

### Scope: the ten next bash scripts

**ID:** scope-ten-next-scripts
**Status:** active
**Chosen:** port `check-workspace-isolation`, `check-task-reviewer-single-dispatch`,
`recover-guard-incident`, `plan-class`, `check-installed-rules`, `resolve-base-branch`,
`check-visual-verify-dispatched`, `check-contract-budget`, `check-task-commit-planning-paths` and
`check-panel-citation-trigger` in one change — the ten slowest bash-logic harnesses at base,
KAN-842's scope.
**Considered:** the Python-backed scripts — their logic is not bash, a port is a rewrite;
splitting into two slices — two review cycles for one mechanical pattern.

### workspace.sh stays bash

**ID:** exclude-workspace-sh
**Status:** active
**Chosen:** leave `workspace.sh` out of this slice and take `check-panel-citation-trigger` in its
rank.
**Considered:** porting it — it drives `docker exec … createdb/dropdb` against the
`flow-postgres` container, the dev workspace's storage (`CLAUDE.md`); a port's parity harness is
the one place a slip would reach that container, for a tool that is not a guard.

### New shared helpers get Go twins; sha256-hex.sh is deleted

**ID:** kan842-helper-twins
**Status:** active
**Chosen:** per `shared-helper-go-twins`, `lib/panel-touched-paths.sh` and `lib/owned-corpus.sh`
get Go twins (`paneltouchedpaths.go`, `ownedcorpus.go`) with parity tests running the bash library
and the Go function over the same inputs; the bash libraries stay. Per
`sole-user-helpers-move-into-go`, `lib/sha256-hex.sh` is `git rm`'d in the task that ports
`plan-class`, its last bash caller; `sha256.go` and `TestSHA256Hex` stay as they are, their
comments repointed.
**Considered:** porting `check-panel-docs-only` and `check-normative-inventory` too — out of this
slice's scope.

### Store reads become the injected Env hook

**ID:** dispatches-via-env-hook
**Status:** active
**Chosen:** the `flow record dispatches -change <name> -C <worktree>` call in
`check-task-reviewer-single-dispatch` and `check-visual-verify-dispatched` goes through the
existing `Env.Dispatches` hook (`guard.go`) the way `panelfixsingledispatch.go`'s `pfdRead` does —
nil execs `flow` on PATH with `-C <worktree>` — stubbed in-process by the tests, with each
script's own failure-to-exit-2 mapping and stderr line.
**Considered:** a stub `flow` binary on PATH per case — a process per case, the cost the port
exists to remove.

### The guard test package runs in at most 40s

**ID:** guard-package-under-40s
**Status:** active
**Chosen:** `go test ./internal/guard/... -count=1` runs in ≤40s real on this machine, every case
kept, fixed at the source of the wall time — supersedes KAN-841's `guard-package-under-30s` for
the grown package — never by dropping a case, `-short`, or a loosened assertion.
**Considered:** keeping ≤30s — KAN-841 ended at 25.87s with ten more ports to add; no package
budget — the Go package becomes the next ceiling unseen.

### Suite success: after median below before median

**ID:** suite-median-below-before
**Status:** active
**Chosen:** `scripts/run-guard-tests.sh` timed x3 with `sysctl -n vm.loadavg` before each run, at
base and at the last implementation commit; met when the after median is below the before median;
the slowest remaining harness is named and the next slice identified.
**Considered:** record only — no criterion to fail.

## Open questions
