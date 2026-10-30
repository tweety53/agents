#!/usr/bin/env bash
# test-await-sub-agents.sh — await-sub-agents.sh returns 0 once every named
# agent's meta.json says finished, 1 at its cap while one still runs, and 2
# on a usage error or an agent it cannot find.
set -u
here=$(cd "$(dirname "$0")" && pwd)
aw="$here/await-sub-agents.sh"
fail=0
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
dir="$tmp/proj/sess/subagents"; mkdir -p "$dir"
export CLAUDE_PROJECTS_DIR="$tmp" AWAIT_SUB_AGENTS_SECONDS=2
echo '{"agentType":"Plan","finished":true}' >"$dir/agent-aaa.meta.json"
echo '{"agentType":"Plan"}' >"$dir/agent-bbb.meta.json"

# (a) all finished: 0 at once
bash "$aw" aaa >/dev/null; s=$?
[ "$s" = 0 ] || { echo "FAIL (a): status=$s"; fail=1; }

# (b) one still running: 1 at the cap, naming it
start=$(date +%s)
out=$(bash "$aw" aaa bbb); s=$?
el=$(( $(date +%s) - start ))
[ "$s" = 1 ] && [ "$el" -ge 2 ] && [ "$el" -le 5 ] && grep -q bbb <<<"$out" || { echo "FAIL (b): status=$s elapsed=$el out=$out"; fail=1; }

# (c) finishing mid-wait returns 0 before the cap
AWAIT_SUB_AGENTS_SECONDS=20 bash "$aw" bbb >/dev/null & pid=$!
echo '{"agentType":"Plan","finished":true}' >"$dir/agent-bbb.meta.json"
start=$(date +%s); wait "$pid"; s=$?
el=$(( $(date +%s) - start ))
[ "$s" = 0 ] && [ "$el" -le 8 ] || { echo "FAIL (c): status=$s elapsed=$el"; fail=1; }

# (d) unknown agent and no argument: 2
bash "$aw" zzz 2>/dev/null; s=$?
[ "$s" = 2 ] || { echo "FAIL (d1): status=$s"; fail=1; }
bash "$aw" 2>/dev/null; s=$?
[ "$s" = 2 ] || { echo "FAIL (d2): status=$s"; fail=1; }

# (e) a cap above 240 s is refused: the wait must return inside the 5-minute cache window
AWAIT_SUB_AGENTS_SECONDS=300 bash "$aw" aaa 2>/dev/null; s=$?
[ "$s" = 2 ] || { echo "FAIL (e): status=$s"; fail=1; }

exit "$fail"
