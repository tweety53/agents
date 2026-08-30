#!/usr/bin/env bash
# generate-relocation-comparison.sh — thin wrapper.
#
# All classification logic lives in generate-relocation-comparison.py
# (Python 3, standard library only), following the same split
# plan-dispatch-bundles.sh uses and for the same reason: this file exists
# only to resolve which arguments the Python script needs and hand them
# over, while the parsing/classification logic underneath gets a real
# language.
#
# Usage: generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>
#
# `<changeRoot>/tasks.md` is appended here as the fourth argument, so the
# Python script never has to know the plan-directory convention itself —
# it only ever sees a `tasks.md` path.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/generate-relocation-comparison.py"

command -v python3 >/dev/null 2>&1 || {
  echo "generate-relocation-comparison.sh: python3 not found on PATH — cannot run the generator" >&2
  exit 2
}

# `command -v` proves the file exists, not that it runs (see
# plan-dispatch-bundles.sh's own comment for the macOS-stub-python3 case
# this probe exists to catch). Probing with a trivial program is what tells
# "python3 is a name on PATH" apart from "python3 is a working interpreter".
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "generate-relocation-comparison.sh: python3 is present but failed to run a trivial program (see above) — cannot run the generator" >&2
  exit 2
fi

if [ "$#" -ne 3 ]; then
  echo "usage: generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>" >&2
  exit 2
fi

WORKTREE="$1"
CHANGE_ROOT="$2"
BASE_REF="$3"

exec python3 "$PYTHON_GUARD" "$WORKTREE" "$CHANGE_ROOT" "$BASE_REF" "$CHANGE_ROOT/tasks.md"
