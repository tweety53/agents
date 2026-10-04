#!/usr/bin/env bash
# plan-dispatch-bundles.sh — thin wrapper.
#
# All classification logic lives in plan-dispatch-bundles.py (Python 3,
# standard library only), following the same split check-task-build-green.sh
# uses and for the same reason: this file exists only so an operator's
# muscle memory invoking this exact filename, and skills/flow/implement.md's own call
# site, keep working, while the block-parsing logic underneath gets a real
# language rather than a hand-rolled Bash ERE allowlist.
#
# Unlike check-plan-provenance (which scans the whole repository tree in
# one call), plan-dispatch-bundles.py's scope is ONE file per invocation.
# This wrapper is what resolves WHICH files that means:
#
#   - no arguments: scan every live change's tasks.md under
#     spectre/changes/*/tasks.md, and that of every archived change under
#     spectre/changes/archive/ the base does not carry yet (`flow-guard
#     unlanded-archives`), calling the Python script once per file and
#     aggregating exit codes — non-zero if ANY file reports a violation or
#     cannot be read, printing each file's own bundles or violations as the
#     Python script emits them;
#   - one argument: treat it as an explicit tasks.md path and scan only
#     that file.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/plan-dispatch-bundles.py"

command -v python3 >/dev/null 2>&1 || {
  echo "plan-dispatch-bundles.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

# `command -v` proves the file exists, not that it runs (see
# check-task-build-green.sh's own comment for the macOS-stub-python3 case
# this probe exists to catch). Probing with a trivial program is what tells
# "python3 is a name on PATH" apart from "python3 is a working interpreter".
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "plan-dispatch-bundles.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

if [ "$#" -eq 1 ]; then
  exec python3 "$PYTHON_GUARD" "$1"
fi

if [ "$#" -gt 1 ]; then
  echo "usage: plan-dispatch-bundles.sh [path-to-tasks.md]" >&2
  exit 2
fi

# No arguments: scan every in-flight change's tasks.md. REPO_ROOT is
# derived from this script's own location so the scan works from any cwd,
# the same convention check-task-build-green.sh uses for its own root —
# unless PLAN_DISPATCH_BUNDLES_ROOT is set, in which case it names the root
# explicitly (the same opt-in override check-task-build-green.sh accepts as
# CHECK_TASK_BUILD_GREEN_ROOT), so a test harness can point this wrapper at
# a sandboxed fixture tree instead of this repository's own
# spectre/changes/.
#
# "One level above $SCRIPT_DIR" is NOT enough to derive that root, because
# this script is now reachable from more than one directory — its real home
# at <repo>/scripts/, and a skills/<name>/scripts/ symlink a command skill
# carries it under. Going up one level from the SECOND of those lands on the
# skill directory, which exists, so a fixed-depth guard would proceed against
# it and return a confident wrong answer (an empty scan) rather than an
# error. Resolving this script's own symlinks first, and deriving the root
# from ITS real physical location, gives the same answer regardless of which
# path invoked it. resolve_file itself is sourced from lib/resolve-file.sh —
# see that file's header for why this guard, unlike the retired
# the retired self-review gather (kan-526), may source a sibling instead of carrying
# its own copy. Sourcing only DEFINES the function; it performs no filesystem
# walk of its own, so doing it here unconditionally does not reintroduce the
# eager resolution the comment below opts out of — only the CALL is deferred.
source "$SCRIPT_DIR/lib/resolve-file.sh"
# Named as $SCRIPT_DIR/<name> and checked before sourcing, so
# check-guard-symlinks.sh rule 2 can see the dependency and a missing
# sibling names itself instead of surfacing as a bash error.
SPEC_ROOT_LIB="$SCRIPT_DIR/lib/spec-root.sh"
if [ ! -f "$SPEC_ROOT_LIB" ]; then
  echo "plan-dispatch-bundles.sh: shared spec-root module not found: $SPEC_ROOT_LIB" >&2
  exit 2
fi
source "$SPEC_ROOT_LIB"
# flow_guard_exec runs `flow-guard unlanded-archives` below; named as
# $SCRIPT_DIR/<name> and checked before sourcing, like every sibling here.
FLOW_GUARD_LIB="$SCRIPT_DIR/lib/flow-guard.sh"
if [ ! -f "$FLOW_GUARD_LIB" ]; then
  echo "plan-dispatch-bundles.sh: shared flow-guard module not found: $FLOW_GUARD_LIB" >&2
  exit 2
fi
source "$FLOW_GUARD_LIB"

# The override is honoured WITHOUT resolving this script's own location, and the
# ordering is deliberate rather than incidental. `${VAR:-expr}` evaluates `expr`
# lazily, so the fixed-depth derivation this replaced never ran at all when
# PLAN_DISPATCH_BUNDLES_ROOT was set. Resolving first and substituting second
# would quietly end that: a caller who sets the override precisely BECAUSE its
# invocation path is exotic — a symlink chain past the loop's 40-hop bound, a
# directory it cannot cd into — would start hitting an exit 2 it never used to
# hit. Resolve only when there is nothing else to fall back to.
if [ -n "${PLAN_DISPATCH_BUNDLES_ROOT:-}" ]; then
  REPO_ROOT="$PLAN_DISPATCH_BUNDLES_ROOT"
else
  SELF_REAL="$(resolve_file "${BASH_SOURCE[0]}")" || {
    echo "plan-dispatch-bundles.sh: cannot resolve this script's own location" >&2
    exit 2
  }
  REPO_ROOT="$(cd "$(dirname "$SELF_REAL")/.." && pwd)"
fi
CHANGES_DIR="$REPO_ROOT/$(spec_root_leaf "$REPO_ROOT")/changes"

STATUS=0

if [ -d "$CHANGES_DIR" ]; then
  # The glob (one directory level under CHANGES_DIR) never descends into an
  # archived change's archive/<name>/tasks.md, a level deeper. An archived
  # change joins the scan only while the base does not carry it — archived
  # on this branch by integrate's run 1, then written into by a fix run — as
  # `flow-guard unlanded-archives` names it
  # (stats/internal/guard/unlandedarchives.go, the one statement of that
  # rule); every archive the base carries stays out.
  ARCHIVED="$(flow_guard_exec unlanded-archives 2 "plan-dispatch-bundles.sh:" "$REPO_ROOT")" || exit 2
  TASKS_FILES=("$CHANGES_DIR"/*/tasks.md)
  while IFS= read -r archived_name; do
    if [ -n "$archived_name" ]; then
      TASKS_FILES+=("$CHANGES_DIR/archive/$archived_name/tasks.md")
    fi
  done <<<"$ARCHIVED"
  for tasks_file in "${TASKS_FILES[@]}"; do
    [ -e "$tasks_file" ] || continue
    rc=0
    python3 "$PYTHON_GUARD" "$tasks_file" || rc=$?
    if [ "$rc" -eq 2 ] || [ "$STATUS" -eq 2 ]; then
      STATUS=2
    elif [ "$rc" -ne 0 ]; then
      STATUS=1
    fi
  done
fi

exit "$STATUS"
