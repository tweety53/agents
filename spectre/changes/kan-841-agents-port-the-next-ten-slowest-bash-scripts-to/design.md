# kan-841-agents-port-the-next-ten-slowest-bash-scripts-to

## Context

- Third slice of KAN-760's guard port, the one KAN-778's After judgement names. Base `d71a2327`.
- Pattern established by KAN-760/KAN-778: a guard is one Go file registered in `guard.Registry`
  plus a shim calling `flow_guard_exec <name> <code> "<name>:" "$@"`; a shim whose Go guard
  defaults its root to the checkout also exports it (KAN-778 task 5 correction —
  `scripts/check-references.sh` is the reference shim).
- Ranking excluded `setup.sh` (installer that must run before Go/flow-guard exist), Python-backed
  wrappers (`compose-mockup-frames`, `generate-relocation-comparison`, `check-plan-shape`) and
  flow-guard's own harnesses (`test-go-guards.sh`, `test-lib-flow-guard.sh`).
- Cannot-answer codes, from each header at base: 4 for `mutate-and-verify`, 2 for the other nine.
  <!-- verified: grep -n 'exit [0-9]' of each script's header @ d71a2327 -->
- Sibling helpers and their other callers at base:
  - `lib/coverage.sh` — Go twin `coverage.go` exists; still sourced by
    `check-self-review-report.sh`, `check-vocabulary.sh`.
  - `lib/resolve-file.sh` (`check-guard-symlinks`) — also `check-dev-stack-fresh.sh`,
    `check-plan-shape.sh`, `check-workspace-isolation.sh`, `plan-dispatch-bundles.sh`.
  - `lib/project-section.sh` (`check-model-keys`) — also `project-get.sh`.
  - `lib/post-mutation-check.sh` (`mutate-and-verify`, `prepare-archive-branch`) — also
    `break-and-prove.sh`.
  - `lib/resolve-remote-base.sh` (`check-base-moved`, `check-finish-preflight`) and
    `lib/reproducer-path.sh` (`prove-reproducer`) — no other caller.
  - `check-worktree-location.sh` (`check-finish-preflight`) — stays bash.
  - `run-reproducer.sh` (`prove-reproducer`) — already a Go guard (`runreproducer.go`).
  <!-- verified: grep -l 'lib/<name>.sh' scripts/*.sh excluding test-*, grep -n '$SCRIPT_DIR/' of each script @ d71a2327 -->

## Measurements

### Parity floors at base

| Harness | floor | real (10 run concurrently) |
|---|--:|--:|
| `test-check-stage-mark-calls.sh` | 74 | 18.88s |
| `test-check-guard-symlinks.sh` | 118 | 13.63s |
| `test-check-dispatch-paragraphs.sh` | 183 | 12.08s |
| `test-mutate-and-verify.sh` | 44 | 22.46s |
| `test-prepare-archive-branch.sh` | 97 | 14.44s |
| `test-check-base-moved.sh` | 63 | 17.96s |
| `test-check-panel-fix-single-dispatch.sh` | 20 | 20.69s |
| `test-check-model-keys.sh` | 30 | 19.14s |
| `test-prove-reproducer.sh` | 8 | 17.24s |
| `test-check-finish-preflight.sh` | 58 | 17.00s |

- Floor is the harness's `^ok` line count on a green run; `test-prove-reproducer.sh` prints no
  `ok` lines, so its floor is its eight `case_N` functions.
- Load 5.59 after the concurrent run.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p bash scripts/test-<name>.sh, the ten concurrently, grep -c '^ok' of each log, all exit 0 @ d71a2327 -->
<!-- measured: grep -cE '^case_[0-9]+$' scripts/test-prove-reproducer.sh = 8 @ d71a2327 -->

Each Go test file carries at least its harness's floor in cases (`parity-by-case-count`).

### Suite before/after

Task 1 records **Before**; the last task records **After**. Shape as KAN-778's: x3 each, load
average read before each run, slowest five harnesses per run.

#### Before

| Run | real (s) | user (s) | sys (s) | load (1/5/15) | exit | harnesses |
|-----|---------:|---------:|--------:|---------------|-----:|-----------|
| suite 1 | 63.25 | 112.57 | 154.22 | 2.26 2.81 3.75 | 0 | 79 (79 passed) |
| suite 2 | 52.75 | 119.55 | 169.11 | 8.95 4.97 4.51 | 0 | 79 (79 passed) |
| suite 3 | 53.58 | 122.36 | 172.87 | 12.03 6.44 5.08 | 0 | 79 (79 passed) |
| go 1 | 11.27 | 16.71 | 30.09 | 17.36 8.93 6.09 | 0 | — |
| go 2 | 13.17 | 16.47 | 28.74 | 18.59 9.45 6.31 | 0 | — |
| go 3 | 13.07 | 17.00 | 29.84 | 16.76 9.48 6.37 | 0 | — |

- Median real: suite **53.58s**; Go package **13.07s**.
- Go exit: each run printed `ok … stats/internal/guard`; the 0 is read from that line.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p scripts/run-guard-tests.sh x3 @ d71a2327 -->
<!-- measured: cd stats && /usr/bin/time -p go test ./internal/guard/... -count=1 x3 @ d71a2327 -->

Slowest five harnesses (runner's per-harness `(Ns)`):

| Rank | Suite 1 | Suite 2 | Suite 3 |
|-----:|---------|---------|---------|
| 1 | test-setup.sh 35s | test-setup.sh 37s | test-setup.sh 42s |
| 2 | test-check-panel-fix-single-dispatch.sh 26s | test-mutate-and-verify.sh 24s | test-compose-mockup-frames.sh 27s |
| 3 | test-compose-mockup-frames.sh 25s | test-compose-mockup-frames.sh 22s | test-mutate-and-verify.sh 23s |
| 4 | test-lib-flow-guard.sh 23s | test-check-guard-symlinks.sh 20s | test-check-guard-symlinks.sh 22s |
| 5 | test-go-guards.sh 23s | test-check-dispatch-paragraphs.sh 19s | test-check-dispatch-paragraphs.sh 22s |

<!-- measured: grep '(Ns)' of each task1-suiteN.log, sorted descending @ d71a2327 -->

#### After

| Run | real (s) | user (s) | sys (s) | load (1/5/15) | exit | harnesses |
|-----|---------:|---------:|--------:|---------------|-----:|-----------|
| suite 1 | 52.43 | 112.70 | 151.44 | 1.93 2.44 2.86 | 0 | 69 (69 passed) |
| suite 2 | 49.48 | 107.76 | 149.37 | 9.77 4.63 3.65 | 0 | 69 (69 passed) |
| suite 3 | 47.14 | 111.93 | 153.93 | 15.87 6.89 4.53 | 0 | 69 (69 passed) |
| go 1 | 26.22 | 32.66 | 54.38 | 13.97 8.05 5.12 | 0 | — |
| go 2 | 25.87 | 32.98 | 54.17 | 13.10 8.32 5.31 | 0 | — |
| go 3 | 23.80 | 33.51 | 54.50 | 10.66 8.11 5.31 | 0 | — |

- Median real: suite **49.48s** (Before 53.58s); Go package **25.87s** (Before 13.07s; ceiling
  30s). Ten harnesses fewer (79 → 69).

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p scripts/run-guard-tests.sh x3 @ branch spectre/kan-841-agents-port-the-next-ten-slowest-bash-scripts-to (455a0f2f) -->
<!-- measured: cd stats && /usr/bin/time -p go test ./internal/guard/... -count=1 x3 @ branch spectre/kan-841-agents-port-the-next-ten-slowest-bash-scripts-to (455a0f2f) -->

Slowest five harnesses (runner's per-harness `(Ns)`):

| Rank | Suite 1 | Suite 2 | Suite 3 |
|-----:|---------|---------|---------|
| 1 | test-go-guards.sh 52s | test-go-guards.sh 44s | test-setup.sh 44s |
| 2 | test-setup.sh 47s | test-setup.sh 42s | test-go-guards.sh 42s |
| 3 | test-check-plan-shape.sh 25s | test-check-plan-shape.sh 28s | test-compose-mockup-frames.sh 26s |
| 4 | test-lib-flow-guard.sh 21s | test-lib-flow-guard.sh 23s | test-lib-flow-guard.sh 23s |
| 5 | test-check-visual-verify-dispatched.sh 17s | test-check-task-reviewer-single-dispatch.sh 20s | test-check-plan-shape.sh 23s |

<!-- measured: grep '(Ns)' of each suite run's log, sorted descending @ branch spectre/kan-841-agents-port-the-next-ten-slowest-bash-scripts-to (455a0f2f) -->

Parity (`cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '--- PASS: Test<Name>/'`,
count / floor): CheckStageMarkCalls 77/74, CheckGuardSymlinks 223/118, CheckDispatchParagraphs
194/183, MutateAndVerify 129/44, PrepareArchiveBranch 213/97, CheckBaseMoved 64/63,
CheckPanelFixSingleDispatch 28/20, CheckModelKeys 45/30, ProveReproducer 11/8,
CheckFinishPreflight 66/58 — every port at or above its floor; no `--- FAIL`.

<!-- measured: cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '--- PASS: Test<Name>/' per port @ branch spectre/kan-841-agents-port-the-next-ten-slowest-bash-scripts-to (455a0f2f) -->

The slowest remaining harness is `test-go-guards.sh`, which runs the Go package itself and now
bounds the suite's wall alongside `test-setup.sh` (excluded from porting). The next slice's bash
candidates are `test-check-plan-shape.sh`, `test-compose-mockup-frames.sh`,
`test-lib-flow-guard.sh`, `test-check-task-reviewer-single-dispatch.sh` and
`test-check-visual-verify-dispatched.sh`; cutting the Go package's own time (its median doubled
here, from process-spawning signal and bash-parity subtests) may pay more than another port slice.

## Decisions

### Carry KAN-760's and KAN-778's port decisions unchanged

**ID:** carry-prior-port-decisions
**Status:** active
**Chosen:** KAN-760's `flow-guard-binary-rationale-in-go`, `go-tests-replace-harnesses`,
`parity-by-case-count`, `inject-deadlines-in-process`, `helper-parity-tests`,
`port-base-moves-into-go`, `guard-binary-built-from-checkout` and KAN-778's
`deferred-findings-to-known-bugs`, `drop-full-rerun-policy` govern this slice as written; the base
for `port-base-moves-into-go` is `d71a2327` — the operator's instruction (recommended options).
**Considered:** re-deciding each per script — no script here raises a case those decisions do not
cover.

### Scope: all ten named scripts

**ID:** scope-ten-next-scripts
**Status:** active
**Chosen:** port the ten KAN-841 scripts in one change — the ticket's scope.
**Considered:** splitting guards from the three tools (`mutate-and-verify`,
`prepare-archive-branch`, `prove-reproducer`) — two review cycles for one mechanical pattern;
including `setup.sh` — the installer runs before Go and flow-guard exist.

### Helpers with remaining bash callers get Go twins

**ID:** shared-helper-go-twins
**Status:** active
**Chosen:** `lib/resolve-file.sh`, `lib/project-section.sh` and `lib/post-mutation-check.sh` get
Go twins in `stats/internal/guard/`, each with a parity test running the bash library and the Go
function over the same inputs and failing on any difference; the bash libraries stay.
`coverage.go` is reused as is.
**Considered:** porting their other bash callers too — out of this slice's scope; no parity test —
silent drift, the failure `helper-parity-tests` exists to prevent.

### Helpers used only by ported scripts move into Go

**ID:** sole-user-helpers-move-into-go
**Status:** active
**Chosen:** `lib/resolve-remote-base.sh` and `lib/reproducer-path.sh` are ported into Go and
`git rm`'d in the task that removes their last bash caller; citations repointed.
**Considered:** keeping them — bash with no caller.

### Unported siblings stay an exec

**ID:** exec-unported-siblings
**Status:** active
**Chosen:** `check-finish-preflight`'s call to `check-worktree-location.sh` stays an exec of the
sibling script, resolved beside the shim, with the same arguments and exit handling.
**Considered:** porting `check-worktree-location` too — scope creep.

### prove-reproducer calls run-reproducer in-process

**ID:** in-process-run-reproducer
**Status:** active
**Chosen:** the Go `prove-reproducer` calls the Go run-reproducer in-process for both legs, with
the verdict and exit mapping the bash read from the shim's exit code kept byte for byte.
**Considered:** exec'ing `run-reproducer.sh` — an extra process and cache-key hash per leg for no
behavioural gain.

### The guard test package runs in at most 30s

**ID:** guard-package-under-30s
**Status:** active
**Chosen:** `go test ./internal/guard/... -count=1` runs in ≤30s real on this machine, every case
kept, fixed at the source of the wall time as KAN-760's `guard-package-under-10s` states — which
this supersedes for the grown package, as KAN-778's `guard-package-under-20s` did — never by
dropping a case, `-short`, or a loosened assertion.
**Considered:** keeping ≤20s — twice the scripts in the same budget; no package budget — the Go
package becomes the next ceiling unseen.

### Suite success: after median below before median

**ID:** suite-median-below-before
**Status:** active
**Chosen:** `scripts/run-guard-tests.sh` timed x3 with `sysctl -n vm.loadavg` before each run, at
base and at the last implementation commit; met when the after median is below the before median;
the slowest remaining harness is named and the next slice identified.
**Considered:** record only — no criterion to fail.

## Open questions
