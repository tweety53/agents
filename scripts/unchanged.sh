#!/usr/bin/env bash
# unchanged.sh — tell a `## test` line whether the branch left every path it reads untouched.
#
# Usage: unchanged.sh <pathspec>...
#
# Compares HEAD and the working tree (untracked files included) with its merge-base on
# origin/main. A suite line guards itself as `scripts/unchanged.sh <pathspecs> || { <suite>; }`,
# so the suite runs only when a path matching a <pathspec> (git pathspec syntax, so ':!docs/'
# excludes) changed. The full suite — the suite commands without the guard — runs only when the
# operator asks for it.
#
# Exit codes:
#   0  nothing matching changed: the suite is skipped, and one line says so
#   1  a matching path changed: the suite runs
#   2  usage: no pathspec, or a git call failed
#
# The logic is the Go port in stats/internal/guard/unchanged.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "unchanged: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec unchanged 2 "unchanged:" "$@"
