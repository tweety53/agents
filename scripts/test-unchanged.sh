#!/usr/bin/env bash
# test-unchanged.sh — unchanged.sh exits 0 only when no path matching its pathspecs changed.
set -u
uc="$(cd "$(dirname "$0")" && pwd)/unchanged.sh"
fail=0
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
cd "$tmp" && git init -q -b main r && cd r && git commit -q --allow-empty -m base &&
  git update-ref refs/remotes/origin/main HEAD
mkdir docs src && touch docs/a.md && git add . && git commit -q -m docs

# (a) only an excluded path changed: 0, and the skip is said
out=$(bash "$uc" . ':!docs/'); s=$?
[ "$s" = 0 ] && grep -q skipped <<<"$out" || { echo "FAIL (a): status=$s $out"; fail=1; }

# (b) an untracked matching file: 1
touch src/x
bash "$uc" . ':!docs/' >/dev/null; s=$?
[ "$s" = 1 ] || { echo "FAIL (b): status=$s"; fail=1; }

# (c) a committed matching file: 1
git add src/x && git commit -q -m src
bash "$uc" . ':!docs/' >/dev/null; s=$?
[ "$s" = 1 ] || { echo "FAIL (c): status=$s"; fail=1; }

# (d) no pathspec: usage
bash "$uc" 2>/dev/null; s=$?
[ "$s" = 2 ] || { echo "FAIL (d): status=$s"; fail=1; }

[ "$fail" = 0 ] && echo "test-unchanged: ok"
exit "$fail"
