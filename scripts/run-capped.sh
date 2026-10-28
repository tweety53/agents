#!/usr/bin/env bash
# run-capped.sh — run a command and kill its whole process tree once the
# tree's summed RSS exceeds a memory cap, so a suite, build or load
# reproduction a flow run starts cannot overload the machine.
#
# Usage: run-capped.sh [-m <MB>] -- <cmd...>
#
# Without -m the cap is the body of `<cwd>/.flow/project.md`'s
# `## resource cap` (project-get.sh): a plain integer of MB, nothing else.
# Absent means 4096 MB.
#
# The tree is every process whose ppid chain reaches the command, plus every
# member of the command's process group — so a child that calls setsid (a
# Playwright browser) is measured and killed while its parent lives. A
# process that reparents to PID 1 (a Gradle daemon) escapes it: the wrapper
# therefore sets GRADLE_OPTS `-Dorg.gradle.daemon=false`, so Gradle builds in
# its client's tree.
#
# The command's stdin is /dev/null: the command runs as a background group
# and would stop on a tty read. The cap is per command; a run-capped.sh
# started under another one (RUN_CAPPED_PGID set) execs its command in place,
# so the outer cap covers it.
#
# Exit codes: the command's own status; 137 when the cap killed the tree
# (stderr carries `CAPPED: <cap>MB (tree RSS <n>MB)`); 130 on SIGINT, 143 on
# SIGTERM, 129 on SIGHUP (tree killed); 2 usage, or a `## resource cap` that
# project-get.sh refuses or that is not a plain MB integer.
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
[ -n "${RUN_CAPPED_PGID:-}" ] && exec "$@"
if [ -z "$cap" ]; then
  body=$("$here/project-get.sh" "$PWD" 'resource cap' 2>"${TMPDIR:-/tmp}/run-capped.$$")
  s=$?
  err=$(cat "${TMPDIR:-/tmp}/run-capped.$$"); rm -f "${TMPDIR:-/tmp}/run-capped.$$"
  case $s in
    0) cap=$(printf "%s\n" "$body" | xargs) ;; # trimmed; the check below refuses units
    1) cap=4096 ;;
    *) echo "run-capped.sh: project-get.sh failed: $err" >&2; exit 2 ;;
  esac
fi
case $cap in '' | *[!0-9]*) echo "run-capped.sh: cap must be a plain integer of MB: $cap" >&2; exit 2 ;; esac

export GRADLE_OPTS="${GRADLE_OPTS:-} -Dorg.gradle.daemon=false"
set -m # job control: the background job leads its own process group
RUN_CAPPED_PGID=$$ "$@" </dev/null &
pg=$!
set +m
# Live (non-zombie) pids of the tree with their RSS (KB): descendants of $pg
# by ppid, plus members of group $pg.
tree() {
  ps -A -o pid=,ppid=,pgid=,rss=,stat= | awk -v g="$pg" '
    { pid[NR] = $1; pp[NR] = $2; pgid[NR] = $3; rss[NR] = $4; st[NR] = $5 }
    END {
      in_[g] = 1
      for (i = 1; i <= NR; i++) if (pgid[i] == g) in_[pid[i]] = 1
      do { more = 0; for (i = 1; i <= NR; i++) if (!(pid[i] in in_) && (pp[i] in in_)) { in_[pid[i]] = 1; more = 1 } } while (more)
      for (i = 1; i <= NR; i++) if ((pid[i] in in_) && st[i] !~ /^Z/) print pid[i], rss[i]
    }'
}
kill_tree() {
  local pids
  pids=$(tree | awk '{ print $1 }')
  [ -n "$pids" ] || return 0
  kill -TERM -- "-$pg" $pids 2>/dev/null
  for _ in 1 2 3 4; do [ -n "$(tree)" ] || return 0; sleep 0.5; done
  pids=$(tree | awk '{ print $1 }')
  kill -KILL -- "-$pg" $pids 2>/dev/null
  # Re-signal each pass: a child forked after the snapshot above is killed too.
  while pids=$(tree | awk '{ print $1 }') && [ -n "$pids" ]; do
    kill -KILL $pids 2>/dev/null
    sleep 0.1
  done
}
trap 'kill_tree; exit 130' INT
trap 'kill_tree; exit 143' TERM
trap 'kill_tree; exit 129' HUP

while kill -0 "$pg" 2>/dev/null; do
  kb=$(tree | awk '{ s += $2 } END { print s + 0 }')
  if [ "$kb" -gt $((cap * 1024)) ]; then
    kill_tree
    wait "$pg" 2>/dev/null
    echo "CAPPED: ${cap}MB (tree RSS $((kb / 1024))MB)" >&2
    exit 137
  fi
  sleep 1
done
wait "$pg"
