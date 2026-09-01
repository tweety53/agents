# kan-369-test-setup-sh-s-source-tree-fingerprint-flakes

## Why

KAN-369: `scripts/test-setup.sh`'s "the repo's own skills, rules and commands are unchanged"
assertion (`source_tree_fingerprint`, `scripts/test-setup.sh:377`) flakes under
`scripts/run-guard-tests.sh`'s concurrent runner — 1 failure in 3 worktree runs, never on `main`,
never when `test-setup.sh` runs alone. Root cause: `scripts/__pycache__/` and
`scripts/lib/__pycache__/` hold real files under `$REPO_ROOT/scripts`, and the fingerprint's `find`
is not `.gitignore`-aware, so it picks them up. When a sibling harness invokes one of the repo's
Python guard scripts (`check-task-commit-fields.py`, `check-plan-provenance.py`,
`plan-dispatch-bundles.py`, `check-installed-citations.py`, `plan_grammar.py`) concurrently with
`test-setup.sh`'s run, CPython can rewrite a `.pyc` cache file mid-window, changing that file's
mtime/size between the "before" and "after" snapshots. The guard suite is the one place this fingerprint runs in the `/flow` pipeline (`flow.verify`), so
a flake there either blocks a clean change or trains the operator to re-run until green.

## What changes

`source_tree_fingerprint()` in `scripts/test-setup.sh` excludes `__pycache__` directories from its
`find` — a generated, `.gitignore`d bytecode cache is not source, the same reason `.git` itself is
never scanned. No other script changes; no existing case covered this path (the `.pyc` churn is
invisible to a single-harness run), so a regression case that fabricates a `__pycache__` mtime
change mid-run is added to `scripts/test-setup.sh` proving the fix.
