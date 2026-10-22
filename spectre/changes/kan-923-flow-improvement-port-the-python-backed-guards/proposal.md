# kan-923-flow-improvement-port-the-python-backed-guards

## Why

- Every new guard and every extension of one is Go in `flow-guard` (`.flow/project.md`), so a guard
  still in Python must be ported before it can be extended (KAN-923).
- Four `scripts/*.sh` call `flow_guard_exec` only for `unlanded-archives` and take their verdict from
  Python; `check-markdown-integrity.py` has no wrapper at all. None has a Go test — the bash
  harnesses are their only coverage.

## What changes

- Seven Python guards run as Go ports inside `flow-guard`: `check-plan-shape`,
  `check-task-build-green`, `check-task-records`, `plan-dispatch-bundles`,
  `check-markdown-integrity`, `plan-dispatch-groups` and `generate-relocation-comparison` — each CLI
  contract unchanged: arguments, environment overrides, output lines, stdout/stderr split, exit codes.
- Each `scripts/<name>.sh` becomes a `flow_guard_exec` shim; `check-markdown-integrity` gains one, and
  `## lint` runs it in place of the `.py`.
- Each bash harness is replaced by a Go table test carrying every case.
- Every `.py` of the cluster is deleted, `lib/plan_grammar.py` and `check-task-commit-fields.py`
  included; `cache-misses.py` and `loaded-files.py` are untouched.
