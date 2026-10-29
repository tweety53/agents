#!/usr/bin/env bash
# plan-agent-pair.sh — the model and effort of a planning session's Explore or
# Plan dispatch.
#
# Usage: plan-agent-pair.sh <seed> <n>
#
# Prints one line, `<model> <effort>` — the first pair of the plan-designer
# row of **Allowed pairs** (skills/flow-contracts/model-policy.md). <seed> is the change name — in /flow-plan
# before capture, the session token; <n> counts the session's planning
# dispatches from 1. Exit 0 on the printed line, exit 2 on a usage error
# (<n> not a positive integer, or an argument count other than 2) or an
# unreadable table.
#
# The logic is the Go port in stats/internal/guard/planagentpair.go; flow-guard
# is built from this checkout by scripts/lib/flow-guard.sh, never taken from
# PATH.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "plan-agent-pair: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec plan-agent-pair 2 "plan-agent-pair:" "$@"
