#!/usr/bin/env bash
# measure-visual-properties.sh — thin wrapper.
#
# All measurement lives in measure-visual-properties.py
# (Python 3, Pillow the one import beyond the standard library), shaped
# exactly like check-plan-shape.sh: this file checks the interpreter and its
# sibling are usable, then execs the real script with the arguments
# unchanged.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/measure-visual-properties.py"

if [ ! -f "$PYTHON_GUARD" ]; then
  echo "measure-visual-properties.sh: required sibling module not found: $PYTHON_GUARD" >&2
  exit 2
fi

command -v python3 >/dev/null 2>&1 || {
  echo "measure-visual-properties.sh: python3 not found on PATH — cannot run the script" >&2
  exit 2
}

# `command -v` proves the file exists, not that it runs — probing with a
# trivial program is what tells "python3 is a name on PATH" apart from
# "python3 is a working interpreter" (see check-plan-shape.sh's own copy of
# this comment for the macOS-stub-python3 case this probe exists to catch).
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "measure-visual-properties.sh: python3 is present but failed to run a trivial program (see above) — cannot run the script" >&2
  exit 2
fi

exec python3 "$PYTHON_GUARD" "$@"
