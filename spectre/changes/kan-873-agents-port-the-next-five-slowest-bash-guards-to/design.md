# kan-873-agents-port-the-next-five-slowest-bash-guards-to

## Context

- Fifth slice of KAN-760's guard port (KAN-778, KAN-841, KAN-842, KAN-850 before it). Base
  `9cd35da8`. The pattern and every per-port convention are KAN-842's (`tasks.md` preamble).
- Ranking: three runs of `scripts/run-guard-tests.sh` at base, per-harness `(Ns)`, KAN-842's
  exclusions (`setup.sh`, Python-backed scripts, flow-guard's own harnesses, `workspace.sh`) plus
  every script already a `flow_guard_exec` shim.
  <!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh x3, grep '(Ns)' per harness @ 9cd35da8 -->

  | Script | run 1 | run 2 | run 3 | median |
  |---|--:|--:|--:|--:|
  | `land-self-review-report` | 10 | 11 | 11 | 11 |
  | `check-self-review-report` | 11 | 8 | 11 | 11 |
  | `break-and-prove` | 10 | 10 | 11 | 10 |
  | `prepare-workspace` | 10 | 9 | 8 | 9 |
  | `check-panel-docs-only` | 15 | 7 | 8 | 8 |
  | `check-worktree-location` | 9 | 7 | 7 | 7 |
  | `check-spec-reach` | 10 | 6 | 7 | 7 |
  | `check-normative-inventory` | 10 | 6 | 7 | 7 |

- Every base run exited 1 on one harness, `test-flow-addr-declaration.sh`: case `pass-unmutated` —
  "expected exactly one FLOW_RECORDS_ADDR -addr registration under
  stats/cmd/flow, found 2". The second is `stats/cmd/flow/lesson.go:66`, added by `79c2f025`.
  Suite wall 79s, 65s, 75s.
  <!-- measured: grep FAIL of each run's log @ 9cd35da8; grep -n '"addr", resolveRecordsAddr()' stats/cmd/flow/*.go -->
- Cannot-answer codes, from each header at base: `break-and-prove` 4, the other four 2.
  <!-- verified: the exit-code paragraph of each script's header @ 9cd35da8 -->
- Sibling helpers at base, each with a Go twin already:
  - `lib/coverage.sh` (`check-self-review-report`) → `coverage.go`; also sourced by
    `check-installed-citations.sh`, `check-guard-symlinks.sh` — stays.
  - `lib/post-mutation-check.sh` (`break-and-prove`) → `postmutationcheck.go`; no other runtime
    caller — deleted (`delete-post-mutation-check-lib`).
  - `lib/panel-touched-paths.sh` (`check-panel-docs-only`) → `paneltouchedpaths.go`; also
    `write-panel-diff.sh` — stays.
  - `check-workspace-isolation.sh` (`prepare-workspace`) — already a shim; exec'd as a sibling
    (`prepare-workspace-execs-sibling`).
  <!-- verified: grep -l 'lib/<name>.sh' scripts/*.sh, grep -n 'SCRIPT_DIR/' of each script @ 9cd35da8 -->

## Measurements

### Parity floors at base

| Harness | floor |
|---|--:|
| `test-land-self-review-report.sh` | 43 |
| `test-check-self-review-report.sh` | 63 |
| `test-break-and-prove.sh` | 38 |
| `test-prepare-workspace.sh` | 25 |
| `test-check-panel-docs-only.sh` | 23 |
| `test-lib-post-mutation-check.sh` | 4 |

- Floor is the harness's `^ok` line count on a green run; all six exited 0.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) bash scripts/test-<name>.sh, grep -c '^ok' of each log, all exit 0 @ 9cd35da8 -->

Each Go test carries at least its harness's floor in subtests (`parity-by-case-count`).

### Suite before/after

**Before** is the ranking's three runs; task 9 records **After** in the same columns.

**Before** — `@ 9cd35da8`:

| Run | suite wall | exit | harnesses | slowest five |
|---|--:|--:|--:|---|
| 1 | 79s | 1 | 46 | go-guards 78s, setup 37s, generate-relocation-comparison 37s, aside-planning-artifacts 22s, check-plan-shape 20s |
| 2 | 65s | 1 | 46 | go-guards 64s, generate-relocation-comparison 27s, setup 16s, lib-flow-guard 15s, run-guard-tests 14s |
| 3 | 75s | 1 | 46 | go-guards 73s, generate-relocation-comparison 29s, setup 18s, check-plan-shape 18s, lib-flow-guard 16s |
| **median** | **75s** | | | |

- Exit 1 in every run is `test-flow-addr-declaration.sh` (**Context**); run 2 also failed
  `test-make-build.sh`'s `build target still compiles every package` with a recipe that contains
  the line — a `pipefail` SIGPIPE race (`tasks.md` task 2).

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh x3, the runner's own "N harnesses, … Ns wall" line and grep '(Ns)' sorted descending @ 9cd35da8 -->

**After** — `@ branch spectre/kan-873-agents-port-the-next-five-slowest-bash-guards-to` (`ea3a5808`):

| Run | suite wall | exit | harnesses | slowest five |
|---|--:|--:|--:|---|
| 1 | 68s | 0 | 40 | go-guards 68s, generate-relocation-comparison 26s, setup 19s, lib-flow-guard 16s, check-plan-shape 16s |
| 2 | 71s | 0 | 40 | go-guards 71s, generate-relocation-comparison 22s, lib-flow-guard 12s, check-plan-shape 12s, run-guard-tests 11s |
| 3 | 54s | 0 | 40 | go-guards 53s, generate-relocation-comparison 16s, lib-flow-guard 13s, lib-parallel 11s, check-plan-shape 11s |
| **median** | **68s** | | | |

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh x3, the runner's own "N harnesses, … Ns wall" line and grep '(Ns)' sorted descending @ branch spectre/kan-873-agents-port-the-next-five-slowest-bash-guards-to -->

Parity (`--- PASS: Test<Name>/`, floor in brackets): CheckPanelDocsOnly 26 (23), PrepareWorkspace
32 (25), BreakAndProve 137 (38), PostMutationCheck 14 (4), LandSelfReviewReport 49 (43),
CheckSelfReviewReport 75 (63) — every port at or above its floor.

<!-- measured: cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '--- PASS: Test<Name>/' @ branch spectre/kan-873-agents-port-the-next-five-slowest-bash-guards-to -->

Judgement: every run green (the two base reds fixed); suite median 68s against 75s; six bash
harnesses gone (46 → 40). `test-go-guards.sh` remains the critical path. Every remaining harness at
or above 10s median is Python-backed (`generate-relocation-comparison`, `check-plan-shape`,
`check-task-records`), already a shim (`aside-planning-artifacts`) or a runner harness
(`lib-parallel`, `run-guard-tests`); the next slice re-ranks the sub-10s bash tail by this
change's three-run median method first.

## Decisions

### Carry the prior slices' port decisions unchanged

**ID:** carry-prior-port-decisions
**Status:** active
**Chosen:** KAN-760's `flow-guard-binary-rationale-in-go`, `go-tests-replace-harnesses`,
`parity-by-case-count`, `inject-deadlines-in-process`, `helper-parity-tests`,
`port-base-moves-into-go`, `guard-binary-built-from-checkout`, KAN-841's
`shared-helper-go-twins`, `exec-unported-siblings`, and KAN-842's
`exclude-workspace-sh` govern this slice as written; the base for `port-base-moves-into-go` is
`9cd35da8` — the operator's instruction (recommended options).
**Considered:** re-deciding each per script — no script here raises a case those decisions do not
cover.

### Scope: the five next bash guards

**ID:** scope-five-next-guards
**Status:** active
**Chosen:** port `land-self-review-report`, `check-self-review-report`, `break-and-prove`,
`prepare-workspace` and `check-panel-docs-only` — the five highest medians of three base runs
(**Context**).
**Considered:** a single run's ranking, as KAN-842 used — run 1 alone ties five scripts at 10s and
puts `check-panel-docs-only` first at 15s, a figure the next two runs halve; the median separates
the ties.

### lib/post-mutation-check.sh is deleted

**ID:** delete-post-mutation-check-lib
**Status:** active
**Chosen:** once `break-and-prove` is ported, delete `scripts/lib/post-mutation-check.sh` and
`scripts/test-lib-post-mutation-check.sh`; `TestPostMutationCheckParity` stops comparing against
the bash library and pins `snapshotTreeState`/`checkTreeRestored` against fixed expected output,
carrying the harness's four cases — KAN-842's `sha256-hex.sh` precedent.
**Considered:** keeping the library — a parity test against a library nothing runs pins the Go
twin to dead code.

### prepare-workspace execs its sibling guard

**ID:** prepare-workspace-execs-sibling
**Status:** active
**Chosen:** the Go port execs `check-workspace-isolation.sh` beside the shim, through
`FLOW_GUARD_REPO_ROOT`, as the bash did — its missing- and unrunnable-sibling cases keep their
meaning and output.
**Considered:** calling the registered Go guard in-process — saves one process start but makes the
sibling cases unreachable, so their harness cases could not be ported.

### flow lesson resolve joins the record family's seam

**ID:** lesson-resolve-through-record-seam
**Status:** active
**Chosen:** `runLessonResolve` registers `-addr`/`-timeout`/`-C` through
`registerRecordConnFlags`, exits 2 with `flow: lesson resolve takes no -C` when `-C` is given (the
read spans every project), and `.flow/project.md`'s record-family sentence names
`flow lesson resolve` — the harness's five assertions hold unchanged.
**Considered:** loosening the harness to allow a second registration — weakens a guard (Lint Fix
Priority); accepting `-C` silently — a flag that does nothing, which `79c2f025`'s own comment
rejects.

## Open questions
