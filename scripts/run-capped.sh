#!/usr/bin/env bash
# run-capped.sh — run a command in its own process group and kill the whole
# group once its summed RSS exceeds a memory cap, so no process a flow run
# starts (suite, dev stack, load reproduction) can overload the machine.
#
# Usage: run-capped.sh [-m <MB>] -- <cmd...>
#
# Without -m the cap is the first integer of `<cwd>/.flow/project.md`'s
# `## resource cap` (project-get.sh), else 4096 MB.
#
# Exit codes: the command's own status; 137 when the cap killed the group
# (stderr carries `CAPPED: <cap>MB (group RSS <n>MB)`); 2 usage.
#
# ponytail: polls RSS once a second, so a spike inside one second can
# overshoot the cap; cgroups/ulimit are not available on darwin.
set -u
here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
cap=
if [ "${1:-}" = -m ]; then cap=${2:-}; shift 2 || true; fi
[ "${1:-}" = -- ] || { echo "usage: run-capped.sh [-m <MB>] -- <cmd...>" >&2; exit 2; }
shift
[ $# -gt 0 ] || { echo "usage: run-capped.sh [-m <MB>] -- <cmd...>" >&2; exit 2; }
if [ -z "$cap" ]; then
  cap=$("$here/project-get.sh" "$PWD" 'resource cap' 2>/dev/null | grep -oE '[0-9]+' | head -1)
  cap=${cap:-4096}
fi
case $cap in '' | *[!0-9]*) echo "run-capped.sh: cap must be an integer MB: $cap" >&2; exit 2 ;; esac

set -m # job control: the background job leads its own process group
"$@" &
pg=$!
set +m
# Live (non-zombie) members of the group; `kill -0 -<pg>` cannot be the test,
# darwin answers ESRCH once only the zombie leader is left while a member is
# still exiting.
alive() { ps -A -o pgid=,stat= | awk -v g="$pg" '$1 == g && $2 !~ /^Z/ { n++ } END { exit !n }'; }
kill_group() {
  kill -TERM -- "-$pg" 2>/dev/null
  for _ in 1 2 3 4; do alive || return 0; sleep 0.5; done
  kill -KILL -- "-$pg" 2>/dev/null
  while alive; do sleep 0.1; done
}
trap 'kill_group; exit 130' INT TERM

while kill -0 "$pg" 2>/dev/null; do
  kb=$(ps -A -o pgid=,rss= | awk -v g="$pg" '$1 == g { s += $2 } END { print s + 0 }')
  if [ "$kb" -gt $((cap * 1024)) ]; then
    kill_group
    wait "$pg" 2>/dev/null
    echo "CAPPED: ${cap}MB (group RSS $((kb / 1024))MB)" >&2
    exit 137
  fi
  sleep 1
done
wait "$pg"
