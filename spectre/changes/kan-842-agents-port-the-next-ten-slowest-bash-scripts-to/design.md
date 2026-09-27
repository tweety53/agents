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

First judgement: `suite-median-below-before` not met (69.27s against 69.20s);
`guard-package-under-40s` not met (51.14s median), load 18–28 through every run. The Go package,
which `test-go-guards.sh` runs, had become the suite's critical path: its tests held parallel slots
idle behind a per-case `sync.OnceValues`, paid macOS's `/usr/bin/git` xcrun trampoline on every
fixture call, rebuilt `flow-guard` per test, and ran the 33-run workspace-isolation parity loop
serially. `9a6aa00a` fixes each at the source, test files only, every case kept.

**Re-measured after `9a6aa00a`** — Before (`@ c5379c0a`) and After
(`@ branch spectre/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to`) run back to back on
the same machine:

| Run | load before (1/5/15) | suite real/user/sys | slowest five | Go pkg load | Go pkg real/user/sys |
|---|---|---|---|---|---|
| Before 1 | 8.12 7.76 9.70 | 64.68 / 106.93 / 146.30 | go-guards 63s, setup 49s, compose-mockup-frames 28s, check-plan-shape 27s, check-task-reviewer-single-dispatch 20s | 22.14 12.94 11.43 | 27.41 / 33.54 / 53.87 |
| Before 2 | 11.64 9.14 10.09 | 52.38 / 109.48 / 153.41 | setup 50s, go-guards 49s, check-plan-shape 24s, compose-mockup-frames 21s, check-task-reviewer-single-dispatch 18s | 17.77 12.79 11.43 | 24.60 / 34.22 / 54.42 |
| Before 3 | 17.20 10.86 10.66 | 55.16 / 106.14 / 149.05 | go-guards 53s, setup 46s, check-plan-shape 26s, lib-flow-guard 24s, check-visual-verify-dispatched 20s | 17.01 13.03 11.55 | 25.83 / 33.32 / 53.28 |
| **Before median** | | **55.16** (69 harnesses, all exit 0) | | | **25.83** |
| After 1 | 17.77 13.59 11.80 | 45.21 / 100.10 / 141.07 | setup 45s, go-guards 44s, lib-flow-guard 22s, check-plan-shape 20s, compose-mockup-frames 18s | 26.63 18.64 14.10 | 27.16 / 38.51 / 67.12 |
| After 2 | 21.52 15.10 12.44 | 46.99 / 101.16 / 141.36 | setup 46s, go-guards 43s, compose-mockup-frames 38s, generate-relocation-comparison 21s, check-plan-shape 20s | 22.62 18.28 14.09 | 27.64 / 37.11 / 66.24 |
| After 3 | 25.75 17.14 13.33 | 46.11 / 101.59 / 142.89 | setup 45s, go-guards 43s, compose-mockup-frames 25s, check-plan-shape 25s, measure-visual-properties 22s | 16.99 17.35 13.90 | 27.05 / 37.81 / 67.88 |
| **After median** | | **46.11** (59 harnesses, all exit 0) | | | **27.16** |

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p scripts/run-guard-tests.sh x3; cd stats && /usr/bin/time -p go test ./internal/guard/... -count=1 x3; sysctl -n vm.loadavg before each; @ c5379c0a then @ branch spectre/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to (9a6aa00a) -->

Parity after `9a6aa00a`: CheckWorkspaceIsolation 195, CheckTaskReviewerSingleDispatch 27,
RecoverGuardIncident 151, PlanClass 31, CheckInstalledRules 22, ResolveBaseBranch 52,
CheckVisualVerifyDispatched 41, CheckContractBudget 45 (the test was deleted with the guard,
`e595f199`), CheckTaskCommitPlanningPaths 19, CheckPanelCitationTrigger 25 — each at or above its
floor, 0 FAIL. Superseded by the HEAD figures below.

<!-- measured: cd stats && go test ./internal/guard/ -count=1 -v, exit 0, grep -c -- '--- PASS: Test<Name>/' per port @ branch spectre/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to (9a6aa00a) -->

Judgement: `suite-median-below-before` **met** (46.11s against 55.16s);
`guard-package-under-40s` **met** (27.16s). Slowest remaining harness: `test-setup.sh` (45–46s),
with `test-go-guards.sh` beside it (43–44s). Next slice: `test-setup.sh`'s wall time, then the Go
package's, not further bash ports.

**Re-measured at HEAD `2100f149`** — after the origin/main merge (`d09aaaa2`), the KAN-809 port
(`97c6f38b`), the contract-budget removal (`e595f199`) and panel round 1's fixes. The figures above
predate all four.

Parity: CheckWorkspaceIsolation 195 (152), CheckTaskReviewerSingleDispatch 28 (11),
RecoverGuardIncident 151 (67), PlanClass 31 (25), CheckInstalledRules 22 (20),
ResolveBaseBranch 52 (39), CheckVisualVerifyDispatched 80 (18), CheckTaskCommitPlanningPaths 19
(19), CheckPanelCitationTrigger 25 (24) — nine ports, each at or above its floor, 0 FAIL.

<!-- measured: cd stats && go test ./internal/guard/ -count=1 -v, exit 0, grep -c -- '--- PASS: Test<Name>/' per port, grep -c -- '--- FAIL' 0 @ 2100f149 -->

| Run | load before (1/5/15) | Go pkg real/user/sys | control `9a6aa00a` load | control real/user/sys |
|---|---|---|---|---|
| 1 | 15.56 10.63 11.34 | 63.52 / 51.96 / 88.56 | | |
| 2 | 24.91 13.94 12.51 | 90.76 / 49.93 / 84.28 | | |
| 3 | 44.49 23.19 16.21 | 49.90 / 50.15 / 84.22 | | |
| 4 | 29.86 23.50 16.87 | 65.76 / 49.28 / 81.95 | 38.34 26.70 18.52 | 90.76 / 57.74 / 83.53 |
| 5 | 34.37 28.77 20.16 | 44.88 / 50.27 / 83.47 | 30.23 28.47 20.49 | 63.19 / 51.83 / 86.07 |
| 6 | 33.13 29.35 21.39 | 124.23 / 52.43 / 89.60 | 56.74 39.62 26.54 | 66.84 / 51.08 / 85.62 |
| **median** | | **64.64** (runs 1–6); **63.52** (runs 1–3) | | **66.84** |

<!-- measured: cd stats && /usr/bin/time -p go test ./internal/guard/... -count=1, sysctl -n vm.loadavg before each, all exit 0; runs 1–3 back to back @ 2100f149, runs 4–6 interleaved with a detached 9a6aa00a worktree @ 2100f149 and 9a6aa00a -->

Judgement at HEAD: `guard-package-under-40s` **not confirmed** — every run exceeded 40s, under load
15–57 from other work on this machine (webpack, headless Chromium). The `9a6aa00a` control, which
measured 27.16s at load 17–27 above, ran 63–91s interleaved with HEAD's runs, so the load, not the
four later commits, sets these figures; the budget needs a re-take on a quiet machine.
`suite-median-below-before` was not re-taken (suite re-timing not required this round).

At `e616e43d` under load 29–43 the Go package and its `9a6aa00a` control, interleaved, ran alike —
head 85.47 / 85.66 / 44.31s, control 75.57 / 86.54 / 46.03s — so the post-`9a6aa00a` commits added
no wall time; `guard-package-under-40s` rests on the 27.16s median `9a6aa00a` measured at load
17–27, not re-confirmed at this load.

<!-- measured: cd stats && /usr/bin/time -p go test ./internal/guard/... -count=1, head and a 9a6aa00a detached checkout alternated x3, sysctl -n vm.loadavg before each @ branch spectre/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to (e616e43d) -->

Once every guard is in Go, a later slice drops the `scripts/<name>.sh` shims and calls
`flow-guard <name>` directly: the build-on-demand `lib/flow-guard.sh` performs moves into the `flow`
CLI or `setup.sh`, every caller in skills, contracts and `.flow/project.md` is repointed, and the
invoked-path hint (`FLOW_GUARD_SELF`) becomes an explicit argument — the shims exist only so the
ports changed no caller, and cost one bash start-up plus a source hash per call.

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
**Status:** superseded in part by `remove-contract-budget-guard`
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
**Status:** superseded in part by `remove-contract-budget-guard`
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

### The contract-budget guard is removed

**ID:** remove-contract-budget-guard
**Status:** active
**Chosen:** the operator removed the contract-budget guard entirely (`e595f199`, merged into the
branch): `check-contract-budget` leaves this slice's scope after its port landed — the shim,
`contractbudget.go` and its tests are deleted — and `lib/owned-corpus.sh` keeps no Go twin, its
`ownedcorpus.go` and parity test deleted with their only caller; `lib/owned-corpus.sh` stays for
`check-normative-inventory.sh`. Supersedes those parts of `scope-ten-next-scripts` and
`kan842-helper-twins`; nine ports remain.
**Considered:** keeping the guard — the operator judged the per-file byte budget useless.

### check-visual-verify-dispatched parity follows KAN-809

**ID:** visual-verify-parity-follows-main
**Status:** active
**Chosen:** origin/main's KAN-809 changed `check-visual-verify-dispatched.sh` after `c5379c0a`;
the branch merged origin/main (`d09aaaa2`, the operator's choice at the panel's base-movement
check) and `97c6f38b` ports KAN-809's behaviour, so that guard's parity reference is origin/main's
bash at `4a278320`, not `c5379c0a` — the one exception to `port-base-moves-into-go`'s base.
**Considered:** keeping the `c5379c0a` behaviour — the merge would have silently reverted KAN-809.

## Open questions
