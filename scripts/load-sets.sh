#!/usr/bin/env bash
# load-sets.sh [<repo>] [<rendered-global-CLAUDE.md>] — thin shim over `flow-guard load-sets`
# (stats/internal/guard/loadsets.go), whose header states the report, its arguments and the
# load-set lists to edit when a change moves what a session loads. A report, never a verdict:
# it always exits 0.
#
# <repo> defaults to this script's repository, passed on as FLOW_GUARD_REPO_ROOT.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 with the cause when it cannot — a failed
# build is never an empty report that reads as a measurement.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "load-sets: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec load-sets 2 "load-sets:" "$@"
