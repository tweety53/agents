#!/usr/bin/env bash
# check-dispatch-pair.sh <agents-repo> <subagent_type> <model> [<effort>] — admit
# one flow dispatch, its prompt on stdin, against the **Allowed pairs** table of
# skills/flow-contracts/model-policy.md, or refuse it
# (hooks/enforce-agent-baseline.py runs it on every dispatch in a flow run).
# The role is the prompt's first whole line `Role: <role id>`; <effort> is
# passed for Explore and Plan, whose subagent_type carries none.
#
# Exit 0, DISPATCH-PAIR-OK <role>: an unmarked row of the role lists the pair,
# or a marked row lists it and the prompt carries its marker as a whole line.
# Exit 1, one DISPATCH-PAIR-REFUSED line: no `Role:` line, an unknown role, a
# pair off the role's rows, or a marked row's marker missing.
# Exit 2: cannot answer — usage, or the table unreadable.
#
# Go, in flow-guard (stats/internal/guard/modelpolicy.go).
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-dispatch-pair: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-dispatch-pair 2 "check-dispatch-pair:" "$@"
