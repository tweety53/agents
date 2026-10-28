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
#   2  usage: no pathspec
set -euo pipefail
if [ $# = 0 ]; then
  echo "usage: unchanged.sh <pathspec>..." >&2
  exit 2
fi
base="$(git merge-base HEAD origin/main)"
if git diff --quiet "$base" -- "$@" && [ -z "$(git ls-files --others --exclude-standard -- "$@")" ]; then
  echo "unchanged: nothing under $* changed since $base — suite skipped"
  exit 0
fi
exit 1
