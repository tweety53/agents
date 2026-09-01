## Context

`source_tree_fingerprint()` (`scripts/test-setup.sh:377-388`) hashes `stat -c '%n %s %Y %a'` (or
the BSD equivalent) for every file under `skills/`, `rules/`, `commands/`, `commands-claude/`,
`scripts/`, `setup.sh`, `CLAUDE.md`, `AGENTS.md` and `README.md`, taken once before the harness's
cases run and once after; a mismatch fails "the repo's own skills, rules and commands are
unchanged" (`scripts/test-setup.sh:1138`). `find` walks the real tree with no `.gitignore`
awareness, so it sees `scripts/__pycache__/*.pyc` and `scripts/lib/__pycache__/*.pyc` — CPython's
bytecode cache for `scripts/*.py` and `scripts/lib/plan_grammar.py` — exactly as it sees any other
file under `scripts/`.

`scripts/run-guard-tests.sh` runs every `scripts/test-*.sh` concurrently
(`scripts/lib/parallel.sh`). Several sibling harnesses invoke the repo's own Python guards during
that window — `test-check-task-build-green.sh` → `check-task-commit-fields.py`,
`test-check-plan-provenance.sh` → `check-plan-provenance.py`, `test-check-plan-shape.sh` →
`check-plan-shape.py` (imports `check-task-commit-fields.py`), plus `plan-dispatch-bundles.py` and
`check-installed-citations.py` invoked elsewhere in the suite. CPython rewrites a module's `.pyc`
whenever the cache's embedded source mtime/size no longer matches the `.py` file — a rewrite that
lands between `test-setup.sh`'s before/after snapshots changes that `.pyc`'s own `%s %Y`, and the
fingerprint reports the fingerprinted tree as tampered with, even though nothing under `skills/`,
`rules/`, `commands/` or `commands-claude/` — the paths a write-through incident would actually
threaten — changed.

This reproduces every observed symptom: fails only under the concurrent runner (a solo `.pyc`
rewrite happening outside `test-setup.sh`'s own window is invisible to it), and not reliably even
there (only when a sibling's Python invocation both runs inside the window and actually rewrites a
stale cache entry — most runs the caches are already current and nothing is written).

## Decisions

### Prune `__pycache__` from the fingerprint's `find`

**ID:** prune-pycache-from-fingerprint
**Status:** active
**Chosen:** add `-path '*/__pycache__' -prune -o -type f -print0` to
`source_tree_fingerprint()`'s `find` in `scripts/test-setup.sh` — a generated, `.gitignore`d
bytecode cache is not source, the same reason `.git` itself is never scanned. Targets the exact
mismatch cause with no behavior change to any other script, and preserves tamper detection: a
`__pycache__` rewrite is never itself the shape a write-through incident produces (it is derived,
rebuildable, and already excluded from git), so excluding it opens no blind spot in the containment
guarantee.
**Considered:**
- `PYTHONDONTWRITEBYTECODE=1` in `scripts/run-guard-tests.sh` — stops the write at the source, but
  only for invocations `run-guard-tests.sh` itself launches; a developer running `python3
  scripts/check-plan-shape.py` by hand outside the harness would still churn the same files and
  still trip the fingerprint the next time `test-setup.sh` runs concurrently with something. The
  prune fixes the assertion itself regardless of why a `.pyc` changed.
- Fingerprinting only `git ls-files`-tracked paths — correct in spirit (untracked, gitignored
  content shouldn't count) but a much larger surface change to a security-sensitive assertion for
  a fix that only needs to exclude one known-derived directory.
