#!/usr/bin/env bash
# check-task-records.sh — thin wrapper.
#
# All parsing and checking logic lives in `flow-guard check-task-records`
# (stats/internal/guard/taskrecords.go), built from this checkout by
# lib/flow-guard.sh. This file exists only so .flow/project.md's declared
# commands and an operator's muscle memory invoking this exact filename keep
# working, while the record checking underneath reads the shared plan grammar
# (stats/internal/guard/planparse.go) that check-task-commit-fields and
# check-plan-shape read. taskrecords.go's header is canonical for the three
# checks (tick state vs commit existence, declared Files vs the commit's real
# files in both directions, Build tag format) and the exit-code contract.
#
#   check-task-records.sh [repo-root] [base-ref]
#
# With no arguments the repo root is the CWD and the base ref is the branch
# refs/remotes/origin/HEAD points at; zero plans in flight under
# spectre/changes/ exits 0, which is what makes the guard a legitimate
# bare-tree lint step.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-task-records: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-task-records 2 "check-task-records:" "$@"
