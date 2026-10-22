#!/usr/bin/env bash
# check-narrative-runs.sh <worktree> <change> <narrative.md>
#
# Compares a change's narrative with the runs the flow store marked for it.
#
# Exit 0: NARRATIVE-RUNS-OK — every marked run has its section.
# Exit 1: NARRATIVE-RUNS-MISSING — a fix or integrate run lacks its section.
# Exit 2: usage error, a store read that did not print one integer, or a
#         narrative that exists and cannot be read.
#
# The mechanism, and its full contract, is
# stats/internal/guard/narrativeruns.go; its test is
# stats/internal/guard/check_narrative_runs_test.go.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-narrative-runs: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-narrative-runs 2 "check-narrative-runs:" "$@"
