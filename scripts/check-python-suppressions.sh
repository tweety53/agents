#!/usr/bin/env bash
# check-python-suppressions.sh — reject linter suppression markers in Python.
#
# Usage: scripts/check-python-suppressions.sh [path ...]     # dirs or files
#
# With no arguments it scans every tracked `*.py` file in the repository —
# `git ls-files '*.py'` from the repo root resolved from this script's own
# location, paths made absolute before the scan so the answer never depends
# on the caller's working directory. With arguments it scans the given files
# or directories instead: the seam its Go test
# (stats/internal/guard/check_python_suppressions_test.go) uses to point the
# guard at fixture trees without touching the real one. The mechanism is
# stats/internal/guard/pythonsuppressions.go. Untracked or ignored Python is never
# scanned — the guard answers for what the repository owns.
#
# WHAT IT CATCHES: the inline markers whose job is to silence a linter or
# type checker on the line (or file) they sit on — `# noqa` (flake8/ruff,
# bare, coded, or file-level `# flake8: noqa`), `# type: ignore` (mypy and
# pyright), `# pyright: ignore` (pyright's own spelling), `# pylint:
# disable`, `# mypy: ignore-errors`, and `# nolint` (the golangci/eslint
# spelling). Matched case-insensitively, anywhere on the line, docstrings
# included — a marker in prose is still an instruction a future reader can
# copy. The Lint Fix Priority rule forbids every one of these outright; the
# fix for the finding a marker hides is the code, never the marker.
#
# WHAT A GREEN RUN DOES NOT PROVE — read before trusting it. This is a
# regression guard for a fixed list of markers. It does not
# prove the Python is lint-clean: no linter runs over this repository's
# Python, and a suppression form outside the list (`# ruff:
# per-file-ignores` in a config file, a decorator no entry names) passes
# clean, as does every non-Python file. It proves one thing: the markers
# listed above have not come back.
#
# Exit codes: 0 no marker found (or no Python file to scan); 1 at least one
# marker found; 2 the guard cannot answer at all — a path argument that is
# neither file nor directory, a git or find failure while enumerating the
# scan set, or a grep that failed for any reason other than "no match". An
# enumeration failure is never read as an empty scan set: a guard that
# cannot answer must say so loudly, because a vacuous pass is the failure
# mode this exit code exists for.

set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-python-suppressions: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-python-suppressions 2 "check-python-suppressions:" "$@"
