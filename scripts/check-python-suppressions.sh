#!/usr/bin/env bash
# check-python-suppressions.sh — reject linter suppression markers in Python.
#
# Usage: scripts/check-python-suppressions.sh [path ...]     # dirs or files
#
# With no arguments it scans every tracked `*.py` file in the repository —
# `git ls-files '*.py'` from the repo root resolved from this script's own
# location, paths made absolute before the scan so the answer never depends
# on the caller's working directory. With arguments it scans the given files
# or directories instead: the seam its harness
# (test-check-python-suppressions.sh) uses to point the guard at fixture
# trees without touching the real one. Untracked or ignored Python is never
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
# regression guard for a fixed list of markers, exactly like
# check-vocabulary.sh's own header states for retired vocabulary. It does not
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

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# One extended regex, case-insensitive at the grep site. Alternation covers
# every marker the header names; `noqa` alone also covers `# flake8: noqa`,
# `# ruff: noqa` and every coded `# noqa: <codes>` spelling.
MARKER_RE='noqa|nolint|type:[[:space:]]*ignore|pyright:[[:space:]]*ignore|pylint:[[:space:]]*disable|mypy:[[:space:]]*ignore-errors'

# Enumeration writes each source's NUL-delimited output to a temp file whose
# writer's exit status is checked synchronously. Neither source may run in a
# process substitution or command substitution: a substitution severs the
# exit status, so a failed `git ls-files` or `find` would silently read as
# an empty scan set — the vacuous pass this guard's exit-2 contract exists
# to prevent. Symlinked entries (skills/*/scripts/ carry symlinks to
# scripts/'s own copies) are skipped — grep follows them and would report
# the same target file once per link; the target itself is always in the
# list from its real path.
TMP_LISTS=()
cleanup_tmp() { [ "${#TMP_LISTS[@]}" -eq 0 ] || rm -f "${TMP_LISTS[@]}"; }
trap cleanup_tmp EXIT

# Sets LIST to a fresh temp file's path. Called in the parent, never inside
# a command substitution: a substitution runs in a subshell, whose
# TMP_LISTS+= dies with it and leaves the EXIT trap nothing to remove —
# one leaked temp file per guard run (panel round-1 finding F6).
new_list() {
  LIST="$(mktemp "${TMPDIR:-/tmp}/check-python-suppressions.XXXXXX")" || {
    echo "ERROR: mktemp failed — cannot enumerate the scan set" >&2
    exit 2
  }
  TMP_LISTS+=("$LIST")
}

FILES=()
if [ "$#" -eq 0 ]; then
  new_list
  if ! git -C "$REPO_ROOT" ls-files -z -- '*.py' > "$LIST"; then
    echo "ERROR: git ls-files failed — cannot enumerate the tracked Python" >&2
    exit 2
  fi
  while IFS= read -r -d '' f; do
    # The symlink test resolves $f against the caller's cwd unless it is
    # prefixed — from a foreign cwd every tracked symlink would pass the
    # test and re-enter the scan set (panel round-1 finding F7).
    [ -L "$REPO_ROOT/$f" ] && continue
    FILES+=("$REPO_ROOT/$f")
  done < "$LIST"
else
  for arg in "$@"; do
    if [ -f "$arg" ]; then
      [ -L "$arg" ] && continue
      case "$arg" in
        *.py) FILES+=("$arg") ;;
      esac
    elif [ -d "$arg" ]; then
      new_list
      if ! find "$arg" -type f -name '*.py' -print0 > "$LIST"; then
        echo "ERROR: find failed on $arg — cannot enumerate the scan set" >&2
        exit 2
      fi
      while IFS= read -r -d '' f; do
        [ -L "$f" ] && continue
        FILES+=("$f")
      done < "$LIST"
    else
      echo "ERROR: no such file or directory: $arg" >&2
      exit 2
    fi
  done
fi

# No Python in scope is a clean answer, not an error.
if [ "${#FILES[@]}" -eq 0 ]; then
  echo "suppressions-ok: no python file in scope"
  exit 0
fi

# -H keeps every hit line prefixed with its filename even when the scan set
# holds a single file — a hit a reader cannot locate is half a report.
set +e
HITS="$(grep -n -H -i -I -E -e "$MARKER_RE" -- "${FILES[@]}")"
GREP_RC=$?
set -e

# grep's own contract: exit 0 = lines selected, 1 = none, 2 = trouble.
case "$GREP_RC" in
  0)
    printf '%s\n' "$HITS"
    echo "suppressions-found: $(printf '%s\n' "$HITS" | wc -l | tr -d ' ') marker hit(s) — fix the code, never widen this guard" >&2
    exit 1
    ;;
  1)
    echo "suppressions-ok: no linter suppression marker in the python in scope"
    exit 0
    ;;
  *)
    echo "ERROR: grep failed with exit $GREP_RC — cannot answer" >&2
    exit 2
    ;;
esac
