#!/usr/bin/env bash
# live-store-copy.sh up <name> [<port>] | down <name> — a design's `## Live check`
# against real data without touching the dev stack (kan-1041 self-review: the
# check was hand-built across about 15 calls).
#
#   up    copies the dev `flow` database, read-only on the source, into a
#         scratch database flow_<name, - as _>_live (pg_dump | psql inside the
#         flow-postgres container), builds flowd from this checkout and starts
#         it on <port> (default 4175) against the copy, then prints the address
#         and the FLOW_ADDR / FLOW_RECORDS_ADDR exports that point a session at it.
#         It registers a scratch change <name> in the copy -- a STARTED state
#         record and .superpowers/sdd/<name>/ in this checkout, refused when
#         that directory already exists -- so `flow record -change <name>` works.
#         A `-` in <name> becomes `_` in the database name and the state
#         directory alike, so `a-b` and `a_b` name one copy, never two sharing
#         a database (kan-1067 self-review, 1-E).
#   down  stops that scratch flowd — only the pid its own pidfile names, and
#         only when that pidfile names this script's binary — and drops
#         flow_<name, - as _>_live. A second `down` is a no-op. It removes the
#         scratch change directory `up` created, and no other.
#
# Never the protected dev stack: port 4173 (the dev flowd) and 4174 (the
# UI-test stack) are refused, and the only database dropped is
# flow_<name, - as _>_live. State lives in /tmp/flow-live-<name, - as _>/; the daemon writes
# its own pidfile at /tmp/flowd-<port>.pid (TMPDIR pinned, as ui-test-up does).
#
# Exit codes: 0 done, 1 a step failed (docker, copy, build, not answering),
# 2 usage or a refusal.
set -euo pipefail
usage() { echo "usage: live-store-copy.sh up <name> [<port>] | down <name>" >&2; exit 2; }
[ $# -ge 2 ] || usage
cmd=$1 name=$2 port=${3:-4175}
[[ $name =~ ^[a-z0-9][a-z0-9_-]*$ ]] || { echo "live-store-copy: name '$name' is not [a-z0-9][a-z0-9_-]*" >&2; exit 2; }
db="flow_${name//-/_}_live"
dir="/tmp/flow-live-${name//-/_}"
repo=$(cd "$(dirname "$0")/.." && pwd)
sdd="$repo/.superpowers/sdd/$name"
pg() { docker exec -i flow-postgres "$@"; }
port_ok() {
  [[ $1 =~ ^[0-9]+$ ]] && [ "$1" -ge 1024 ] && [ "$1" -le 65535 ] && [ "$1" != 4173 ] && [ "$1" != 4174 ] \
    || { echo "live-store-copy: port '$1' refused (1024-65535, never 4173 or 4174)" >&2; exit 2; }
}
case $cmd in
up)
  [ $# -le 3 ] || usage
  port_ok "$port"
  if [ -e "$dir" ]; then echo "live-store-copy: $dir exists -- run down $name first" >&2; exit 2; fi
  if [ -e "$sdd" ]; then echo "live-store-copy: $sdd exists -- pick another name" >&2; exit 2; fi
  if curl -sf "http://127.0.0.1:$port/api/v1/changes" >/dev/null 2>&1; then
    echo "live-store-copy: port $port already answers" >&2; exit 2
  fi
  if [ -n "$(pg psql -U flow -d flow -tAc "select 1 from pg_database where datname='$db'")" ]; then
    echo "live-store-copy: database $db exists -- run down $name first" >&2; exit 2
  fi
  mkdir -p "$dir/transcripts" "$dir/state"
  mkdir -p "$sdd"
  echo "$sdd" >"$dir/scratch"
  echo "$port" >"$dir/port"
  pg createdb -U flow "$db"
  if ! pg pg_dump -U flow --no-owner --no-acl flow | pg psql -q -v ON_ERROR_STOP=1 -U flow -d "$db" >/dev/null; then
    echo "live-store-copy: copy into $db failed" >&2; exit 1
  fi
  (cd "$repo/stats" && make -s web-dist && go build -o "$dir/flowd" ./cmd/flowd)
  env -u FLOWD_JIRA_SITE -u FLOWD_JIRA_EMAIL -u FLOWD_JIRA_TOKEN \
    FLOWD_PORT="$port" FLOWD_DSN="postgres://flow:flow@localhost:5433/$db?sslmode=disable" \
    FLOW_TRANSCRIPTS_DIR="$dir/transcripts" FLOW_STATE_DIR="$dir/state" TMPDIR=/tmp \
    nohup "$dir/flowd" >"$dir/log" 2>&1 &
  for _ in $(seq 1 60); do
    if curl -sf "http://127.0.0.1:$port/api/v1/changes" >/dev/null 2>&1; then
      echo "http://127.0.0.1:$port"
      echo "export FLOW_ADDR=http://127.0.0.1:$port FLOW_RECORDS_ADDR=http://127.0.0.1:$port"
      body=$(printf '{"state":"STARTED","branch":"%s","worktrees":{"%s":"%s"},"artifactUrl":null,"jiraIssue":null,"prUrl":null,"updatedBy":"/flow"}' \
        "$(git -C "$repo" branch --show-current)" "$repo" "$(git -C "$repo" rev-parse HEAD)")
      if ! printf '%s' "$body" | FLOW_ADDR="http://127.0.0.1:$port" flow state set -C "$repo" "$name"; then
        echo "live-store-copy: state record for $name failed" >&2; exit 1
      fi
      echo "scratch change: $name"
      exit 0
    fi
    sleep 1
  done
  echo "live-store-copy: flowd did not answer on :$port within 60s -- see $dir/log" >&2; exit 1
  ;;
down)
  [ $# -eq 2 ] || usage
  if [ -f "$dir/port" ]; then
    port=$(cat "$dir/port")
    port_ok "$port"
    pidfile="/tmp/flowd-$port.pid"
    if [ -f "$pidfile" ] && [ "$(sed -n 2p "$pidfile")" -ef "$dir/flowd" ]; then
      pid=$(head -n1 "$pidfile")
      kill "$pid" 2>/dev/null || true
      for _ in $(seq 1 30); do kill -0 "$pid" 2>/dev/null || break; sleep 1; done
      if kill -0 "$pid" 2>/dev/null; then echo "live-store-copy: flowd $pid did not exit within 30s" >&2; exit 1; fi
    fi
  fi
  pg dropdb -U flow --if-exists --force "$db"
  if [ -f "$dir/scratch" ]; then rm -rf "$(cat "$dir/scratch")"; fi
  rm -rf "$dir"
  ;;
*) usage ;;
esac
