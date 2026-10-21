## Context

- Sixth slice of KAN-760's guard port (KAN-778, KAN-841, KAN-842, KAN-850, KAN-873 before it).
  Base `c9b8fde8`. The pattern and every per-port convention are KAN-873's (`tasks.md` preamble).
- Cannot-answer code is `2` for all five; none derives a path from its own location
  (`$SCRIPT_DIR`/`BASH_SOURCE` absent), so no shim exports `FLOW_GUARD_REPO_ROOT`.
  <!-- verified: grep -n 'exit [0-9]\|SCRIPT_DIR\|BASH_SOURCE' scripts/<name>.sh for the five @ c9b8fde8 -->
- `guard-autosquash.sh` has one runtime caller outside the skills' prose: `foldfixup.go` execs it
  through `FLOW_GUARD_AUTOSQUASH`, exported by the `fold-fixup.sh` shim.
  <!-- verified: grep -rn 'FLOW_GUARD_AUTOSQUASH' stats/internal/guard/*.go scripts/*.sh @ c9b8fde8 -->
- `check-worktree-processes`'s only source of an answer is `lsof`; its harness drives real
  processes and resolves every path with `pwd -P`.
  <!-- verified: scripts/check-worktree-processes.sh lines 87-94, scripts/test-check-worktree-processes.sh lines 28-132 @ c9b8fde8 -->

## Measurements

### Parity floors at base

| Harness | floor |
|---|--:|
| `test-check-plan-unchanged.sh` | 30 |
| `test-check-tree-markers.sh` | 30 |
| `test-guard-autosquash.sh` | 12 |
| `test-check-panel-diff-size.sh` | 8 |
| `test-check-worktree-processes.sh` | 27 |

- Floor is the harness's `^ok` line count on a green run; all five exited 0.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) bash scripts/test-<name>.sh, grep -c '^ok' of each log, all exit 0 @ c9b8fde8 -->

Each Go test carries at least its harness's floor in subtests (`parity-by-case-count`).

**After** — subtests per port:

| Go test | `--- PASS` subtests | floor |
|---|--:|--:|
| `TestCheckPlanUnchanged` | 30 | 30 |
| `TestCheckTreeMarkers` | 38 | 30 |
| `TestGuardAutosquash` | 13 | 12 |
| `TestCheckPanelDiffSize` | 21 | 8 |
| `TestCheckWorktreeProcesses` | 44 | 27 |

<!-- measured: cd stats && go test ./internal/guard/ -run '^Test<Name>$' -count=1 -v | grep -c -- '--- PASS: Test<Name>/' @ branch spectre/kan-918-flow-improvement-port-the-git-snapshot-and -->

Full guard package: `ok  	github.com/tweety53/agents/stats/internal/guard	109.802s` (`go test ./internal/guard/ -count=1 -race`).

### Suite before/after

**Before** — `@ c9b8fde8`:

| Run | suite wall | exit | harnesses | slowest five |
|---|--:|--:|--:|---|
| 1 | 313s | 0 | 38 | integrate-change 313s, go-guards 108s, check-plan-shape 34s, protect-main-checkout 29s, lib-flow-guard 25s |
| 2 | 314s | 0 | 38 | integrate-change 313s, go-guards 117s, protect-main-checkout 42s, check-plan-shape 31s, check-task-records 25s |
| 3 | 413s | 0 | 38 | integrate-change 411s, go-guards 89s, protect-main-checkout 30s, check-plan-shape 24s, generate-relocation-comparison 20s |
| **median** | **314s** | | | |

- `test-integrate-change.sh` is the critical path; the five harnesses this change deletes are not
  in any run's slowest five, so the suite wall is not expected to move — the gain is five fewer
  bash harnesses (38 → 33).

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh x3 on a quiet tree, the runner's own "N harnesses, … Ns wall" line and grep '(Ns)' sorted descending @ c9b8fde8 -->

**After** — `@ branch spectre/kan-918-flow-improvement-port-the-git-snapshot-and`:

| Run | suite wall | exit | harnesses | slowest five |
|---|--:|--:|--:|---|
| 1 | 365s | 0 | 33 | integrate-change 364s, go-guards 91s, protect-main-checkout 26s, check-plan-shape 22s, check-task-records 17s |
| 2 | 405s | 0 | 33 | integrate-change 405s, go-guards 118s, protect-main-checkout 47s, check-plan-shape 32s, generate-relocation-comparison 27s |
| 3 | 391s | 0 | 33 | integrate-change 391s, go-guards 120s, protect-main-checkout 34s, check-plan-shape 31s, check-task-records 24s |
| **median** | **391s** | | | |

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh x3, the runner's own "N harnesses, … Ns wall" line and grep '(Ns)' sorted descending @ branch spectre/kan-918-flow-improvement-port-the-git-snapshot-and -->

- Every harness green and every port at or above its floor; the median is 77s above Before's.
  `test-integrate-change.sh`, which this change does not touch, is the critical path in all six
  runs (313–411s Before, 364–405s After), so the difference reads as machine load, not a
  regression — an inference, not a measurement.

## Decisions

### Carry the prior slices' port decisions unchanged

**ID:** carry-prior-port-decisions
**Status:** active
**Chosen:** KAN-760's `flow-guard-binary-rationale-in-go`, `go-tests-replace-harnesses`,
`parity-by-case-count`, `inject-deadlines-in-process`, `helper-parity-tests`,
`guard-binary-built-from-checkout`, KAN-841's `shared-helper-go-twins`, and KAN-873's shim template
and five-step port govern this slice as written; the base is `c9b8fde8`.
**Considered:** re-deciding each per script — no script here raises a case those decisions do not
cover.

### Scope: the five guards the issue names

**ID:** scope-five-issue-guards
**Status:** active
**Chosen:** port exactly the five KAN-918 lists, one task each.
**Considered:** re-ranking the bash tail by timing as KAN-873 did — the issue fixes the set.

### fold-fixup keeps exec'ing guard-autosquash.sh

**ID:** fold-fixup-execs-autosquash-shim
**Status:** active
**Chosen:** `foldfixup.go` keeps exec'ing the path `FLOW_GUARD_AUTOSQUASH` names — now a shim —
and only its "stays bash" comment is corrected.
**Considered:** calling the registered Go guard in-process — saves one process start but changes
fold-fixup's sibling-missing behaviour and its tests; KAN-873's `prepare-workspace-execs-sibling`
took exec for the same reason.

### check-worktree-processes execs lsof from PATH

**ID:** worktree-processes-execs-lsof
**Status:** active
**Chosen:** the Go port runs `lsof` resolved through `Env`'s `PATH`, as the bash did, and its
missing-`lsof` case stays exit 2 with the bash's message; tests drive real child processes.
**Considered:** reading process cwds through `/proc` or libproc — no `/proc` on macOS, and a cgo
or syscall route is a new mechanism the parity contract forbids.

### Where a script's header and body disagree on an exit code, the port follows the header

**ID:** header-wins-on-exit-code
**Status:** active
**Chosen:** two ports depart from their bash body's exit code and follow the script's own header
instead: `guard-autosquash after` exits 2 (cannot answer) on a regular `tasks.md` it cannot read,
where the bash swept zero tokens and could pass (commit 5e2f347f); `check-panel-diff-size` exits 2
when `git diff --shortstat` fails, where the bash exited with git's own code (commit ebd0cbf2).
Every other exit code stays the bash's, so proposal.md's unchanged-exit-codes claim holds for all
but these two cases.
**Considered:** keeping the bash bodies' codes byte-for-byte — that preserves a fail-open pass on an
unreadable plan and an exit code outside the header's 0/1/2 contract, both defects the header
already ruled out.

## Live check

`flow.verify` runs each of the five shims once on this repository through
`~/.claude/skills/flow/scripts/<name>.sh` (the symlinked install path) with the arguments the
skills pass, and compares exit code and output with the same call against the base's bash; any
difference is a failure. A shim that prints nothing and exits 0 where the bash printed a verdict
line is the failure shape to watch for.

## Open questions
