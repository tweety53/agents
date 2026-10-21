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

### Suite before/after

**Before** — `@ c9b8fde8`:

BEFORE_TABLE

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

## Live check

`flow.verify` runs each of the five shims once on this repository through
`~/.claude/skills/flow/scripts/<name>.sh` (the symlinked install path) with the arguments the
skills pass, and compares exit code and output with the same call against the base's bash; any
difference is a failure. A shim that prints nothing and exits 0 where the bash printed a verdict
line is the failure shape to watch for.

## Open questions
