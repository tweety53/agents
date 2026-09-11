#!/usr/bin/env bash
# plan-dispatch-groups.sh — thin wrapper.
#
# All classification logic lives in plan-dispatch-groups.py (Python 3,
# standard library only), following the same split plan-dispatch-bundles.sh
# uses and for the same reason. Unlike plan-dispatch-bundles.sh, this
# wrapper carries only the one-argument form: the Decide step
# (skills/flow/brainstorm-planner.md, step 4) always names one explicit
# tasks.md path, and there is no "every open change" scan to resolve.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/plan-dispatch-groups.py"

command -v python3 >/dev/null 2>&1 || {
  echo "plan-dispatch-groups.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

# `command -v` proves the file exists, not that it runs (see
# plan-dispatch-bundles.sh's own comment for the macOS-stub-python3 case
# this probe exists to catch).
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "plan-dispatch-groups.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

if [ "$#" -ne 1 ]; then
  echo "usage: plan-dispatch-groups.sh <path-to-tasks.md>" >&2
  exit 2
fi

exec python3 "$PYTHON_GUARD" "$1"
