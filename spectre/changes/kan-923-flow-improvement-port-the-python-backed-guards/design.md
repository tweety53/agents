## Context

- Seventh slice of KAN-760's guard port (KAN-778, KAN-841, KAN-842, KAN-850, KAN-873, KAN-918 before
  it). Base `bbfe685e`. The pattern and every per-port convention are KAN-918's (`tasks.md` preamble).
- The Python import graph: `check-plan-shape.py`, `check-task-records.py` and
  `plan-dispatch-bundles.py` load `check-task-commit-fields.py` and `lib/plan_grammar.py`;
  `check-task-build-green.py` loads `lib/plan_grammar.py`; `plan-dispatch-groups.py` and
  `generate-relocation-comparison.py` load `plan-dispatch-bundles.py`.
  <!-- verified: grep -n 'spec_from_file_location\|plan_grammar\|check-task-commit-fields.py\|plan-dispatch-bundles.py' scripts/*.py @ bbfe685e -->
- `check-task-commit-fields.py`'s verdict half is already Go (`taskcommitfields.go`); only its parse
  half and `check_files` remain, kept for the importers above. `planparse.go` holds the Go twins of
  that parse half and of `plan_grammar.py`'s fence rule.
  <!-- verified: stats/internal/guard/taskcommitfields.go header comment, stats/internal/guard/planparse.go @ bbfe685e -->
- `check-plan-shape.sh`'s per-file check runs the Python checks plus three Go guards
  (`plan-suite-steps`, `plan-script-paths`, `plan-task-id-rule`); its no-argument scan, like
  `check-task-build-green.sh`'s, `check-task-records.sh`'s and `plan-dispatch-bundles.sh`'s, adds
  `flow-guard unlanded-archives`. Those wrappers derive the repository root from their own location,
  with a `*_ROOT` environment override for harnesses.
  <!-- verified: scripts/check-plan-shape.sh check_one and no-argument branch; grep -n '_ROOT' scripts/check-task-build-green.sh scripts/plan-dispatch-bundles.sh @ bbfe685e -->

## Measurements

### Parity floors at base

| Harness | floor |
|---|--:|
| `test-check-plan-shape.sh` | 131 |
| `test-check-task-build-green.sh` | 88 |
| `test-check-task-records.sh` | 23 |
| `test-plan-dispatch-bundles.sh` | 50 |
| `test-check-markdown-integrity.sh` | 25 |
| `test-plan-dispatch-groups.sh` | 14 |
| `test-generate-relocation-comparison.sh` | 35 |

- Floor is the harness's `^ok` line count on a green run (`^PASS:` for
  `test-generate-relocation-comparison.sh`, whose cases print that); all seven exited 0.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) bash scripts/test-<name>.sh, grep -c '^ok' (or '^PASS:') of each log, all exit 0 @ bbfe685e -->

Each Go test carries at least its harness's floor in subtests (`parity-by-case-count`).

## Decisions

### Carry the prior slices' port decisions unchanged

**ID:** carry-prior-port-decisions
**Status:** active
**Chosen:** KAN-760's `flow-guard-binary-rationale-in-go`, `go-tests-replace-harnesses`,
`parity-by-case-count`, `inject-deadlines-in-process`, `helper-parity-tests`,
`guard-binary-built-from-checkout`, KAN-841's `shared-helper-go-twins`, KAN-873's shim template and
KAN-918's five-step port and `header-wins-on-exit-code` govern this slice as written; the base is
`bbfe685e`.
**Considered:** re-deciding each per script — no script here raises a case those decisions do not
cover.

### Scope: the five named guards plus the two remaining importers of plan-dispatch-bundles.py

**ID:** scope-seven-guards
**Status:** active
**Chosen:** port the five KAN-923 names plus `plan-dispatch-groups` and
`generate-relocation-comparison`, one task each, so every `.py` the issue says to delete can go —
`plan-dispatch-bundles.py`, `lib/plan_grammar.py` and `check-task-commit-fields.py` have no importer
left only once both are Go.
**Considered:** the five alone — leaves `plan-dispatch-bundles.py`, and through it the grammar and
the commit-fields module, alive as a library for two Python scripts, so the issue's deletions could
not land; a separate change for the two — splits one dependency cluster across two landings for no
gain.

### One Go grammar, in planparse.go

**ID:** grammar-in-planparse
**Status:** active
**Chosen:** `lib/plan_grammar.py`'s functions and `plan-dispatch-bundles.py`'s `parse_tasks` get one
Go twin each, in `stats/internal/guard/planparse.go` beside the existing `tcf*` parse helpers, reused
by every port that loaded them; a twin that already exists is reused, never duplicated.
**Considered:** a twin per port — the Python had one grammar on purpose (parser identity across
guards), and copies would drift; a new package — `planparse.go` already is that home in `guard`.

### check-plan-shape runs its sub-guards in-process

**ID:** plan-shape-subguards-in-process
**Status:** active
**Chosen:** the Go `check-plan-shape` calls `plan-suite-steps`, `plan-script-paths`,
`plan-task-id-rule` and `unlanded-archives` through their `Registry` functions, in the bash order,
combining exit codes as `check_one` did (2 beats 1 beats 0). The other three scanning wrappers call
`unlanded-archives` the same way.
**Considered:** exec'ing `flow-guard` again per sub-guard — a process per file per sub-guard for
output that is already in-process.

### Self-locating shims export FLOW_GUARD_REPO_ROOT

**ID:** shims-export-repo-root
**Status:** active
**Chosen:** a shim whose wrapper derived the repository root from its own location exports
`FLOW_GUARD_REPO_ROOT` as `check-ask-silence.sh` does; each `*_ROOT` override keeps working, read
from `Env`.
**Considered:** resolving the root from the working directory — wrong when run through the installed
`~/.claude/skills/flow/scripts/` symlink from another project.

### check-markdown-integrity gets a shim and ## lint runs it

**ID:** markdown-integrity-shim
**Status:** active
**Chosen:** add `scripts/check-markdown-integrity.sh` (shim template) and change `## lint`'s
`scripts/check-markdown-integrity.py` line to it.
**Considered:** registering the guard without a shim — every other `## lint` guard is a
`scripts/*.sh`, and `run-guard-tests.sh`'s companion rule keys on that name.

### The shared Python modules go last

**ID:** shared-modules-deleted-last
**Status:** active
**Chosen:** each port deletes its own entry `.py` and harness in its own task;
`plan-dispatch-bundles.py`, `lib/plan_grammar.py` and `check-task-commit-fields.py` are deleted in
one final task once every importer is ported, with the citation sweep.
**Considered:** deleting each with the port that ported it — the importers still running would
break between tasks.

## Live check

`flow.verify` runs each of the seven shims on this repository through
`~/.claude/skills/flow/scripts/<name>.sh` (the symlinked install path) with the arguments the skills
pass — no argument for the four scans, this change's `tasks.md` for the bundle and group planners,
a relocation comparison for this change — and compares exit code, stdout and stderr with the same
call against the base's Python (a temporary worktree of `bbfe685e`); any difference is a failure.
A shim that prints nothing and exits 0 where the Python printed a verdict is the failure shape to
watch for. `scripts/check-markdown-integrity.sh` must report what `python3
scripts/check-markdown-integrity.py` reports at base on the same tree.

## Open questions
