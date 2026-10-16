#!/usr/bin/env bash
# check-loop-dispatch.sh <agents-repo> <key> <subagent_type> <model> — admit
# one dispatch of a pipeline-fix or self-review fix loop, or refuse it, and
# record it (hooks/enforce-loop-dispatch.py runs it on every Agent/Task call
# whose prompt carries a `Key: pipeline-fix-<k>-…` or `Key: self-review-…`
# line).
#
# Refused (exit 1, one LOOP-DISPATCH-REFUSED line):
#   KEY          no row of the pairs table in
#                skills/flow-self-review/SKILL.md — the one place the pairs
#                are read from — covers the key's suffix;
#   NOT-STARTED  a pipeline-fix key with no pipeline-fix-<name>-<k> branch
#                (scripts/pipeline-fix.sh start makes it);
#   PAIR         subagent_type or model off the suffix's row;
#   ESCALATION   the row's escalation type with no recorded escalation
#                (scripts/pipeline-fix.sh escalate);
#   KEY-REUSED   the key already has a row;
#   OPEN         an earlier row of the loop is not closed
#                (scripts/pipeline-fix.sh end);
#   PLAN         a fix or review with no completed plan row.
# Admitted (exit 0, LOOP-DISPATCH-OK): the row is opened with `flow record
# dispatch begin` under the loop's change, its session token the loop's
# branch name — except a `pass`, which is pair-checked and not recorded. A
# self-review on the change's own branch (no self-review-<name> branch in
# <agents-repo>) reads and records in the flowd project of the working
# directory it runs in — the hook runs it in the payload's cwd.
# LOOP-DISPATCH-NA, exit 0: the key is no loop key.
# Exit 2: cannot answer — the table unreadable, flowd unreadable, the begin
# refused (a loop change no checkout carries: pipeline-fix.sh start), usage.
#
# Go, in flow-guard (stats/internal/guard/loop.go). flow-guard is built from
# this checkout, never taken from PATH: scripts/lib/flow-guard.sh derives it,
# and exits 2 with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-loop-dispatch: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-loop-dispatch 2 "check-loop-dispatch:" "$@"
