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
set -u
root=${CLAUDE_PROJECTS_DIR:-$HOME/.claude/projects}
cap=${AWAIT_SUB_AGENTS_SECONDS:-110}
if [ $# -eq 0 ]; then
  echo "usage: await-sub-agents.sh <agent id>..." >&2
  exit 2
fi
if ! [[ $cap =~ ^[0-9]+$ ]] || [ "$cap" -gt 240 ]; then
  echo "await-sub-agents: AWAIT_SUB_AGENTS_SECONDS must be 0-240, got $cap" >&2
  exit 2
fi

metas=()
for id in "$@"; do
  m=$(find "$root" -path "*/subagents/agent-$id.meta.json" -print -quit 2>/dev/null)
  if [ -z "$m" ]; then
    echo "await-sub-agents: no meta.json for agent $id under $root" >&2
    exit 2
  fi
  metas+=("$m")
done

deadline=$(( $(date +%s) + cap ))
while :; do
  running=()
  for i in "${!metas[@]}"; do
    grep -q '"finished"[[:space:]]*:[[:space:]]*true' "${metas[$i]}" || running+=("${@:i+1:1}")
  done
  [ ${#running[@]} -eq 0 ] && exit 0
  if [ "$(date +%s)" -ge "$deadline" ]; then
    echo "still running: ${running[*]}"
    exit 1
  fi
  sleep 2
done
