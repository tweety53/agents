#!/usr/bin/env bash
# await-sub-agents.sh — wait, at most 110 seconds (inside the Bash tool's
# default 2-minute timeout), for background agents to
# finish, so a subagent waiting on its own agents keeps its 5-minute prompt
# cache warm: each call returns inside the cache window, and the next call is
# a cache read where ending the turn and waking on the hand-back past 5
# minutes rewrites the whole context (kan-1061's planner: 206k tokens
# rewritten after a 13-minute wait on its Plan agents).
#
# Usage: await-sub-agents.sh <agent id>...
#
# An agent is finished when its subagents/agent-<id>.meta.json, under any
# session of $CLAUDE_PROJECTS_DIR (default ~/.claude/projects), carries
# "finished": true. Exit 0 once every named agent is finished; exit 1 at the
# cap with one line naming those still running — call again; exit 2 on no
# argument, an agent id with no meta.json, or a cap
# ($AWAIT_SUB_AGENTS_SECONDS) above 240.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "await-sub-agents: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec await-sub-agents 2 "await-sub-agents:" "$@"
