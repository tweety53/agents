#!/usr/bin/env bash
# Assertion harness for check-done-when-paths.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's exit code
# and stdout. Never touches the real repository tree. Real git repositories
# rather than fixture trees: the guard reads `git ls-files`.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the contract
# stated in the shim's header, never against observed output.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-done-when-paths.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

REPOS=()
cleanup() {
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

# new_repo <name> — a fresh repository with one empty base commit; prints
# its path and registers it for cleanup.
new_repo() {
  local dir
  dir="$(cd "$(mktemp -d)" && pwd -P)"
  REPOS+=("$dir")
  git -C "$dir" init -q
  git -C "$dir" commit -q --allow-empty -m base
  printf '%s' "$dir"
}

# track <repo> <path> <body> — create with parents and stage, so the index
# (the set the guard judges against) carries it.
track() {
  mkdir -p "$(dirname -- "$1/$2")"
  printf '%s\n' "$3" >"$1/$2"
  git -C "$1" add "$2"
}

# run_guard <repo> [args...] -> sets RC, OUT
run_guard() {
  local repo="${1:-}"
  shift || true
  set +e
  OUT="$("$GUARD" "$repo" "$@" 2>/tmp/kan817-guard-err.$$)"
  RC=$?
  ERR="$(cat /tmp/kan817-guard-err.$$ 2>/dev/null || true)"
  rm -f /tmp/kan817-guard-err.$$
  set -e
}

r1="$(new_repo)"
track "$r1" "docs/tickets/kan-x.md" "## Done when

\`shots/27.png\` re-baselined.
"
run_guard "$r1"
if [ "$RC" -eq 1 ] && grep -q "^DONE-WHEN-PATH: shots/27.png — docs/tickets/kan-x.md$" <<<"$OUT" &&
  grep -q "^DONE-WHEN-VIOLATION: $r1 — 1$" <<<"$OUT"; then
  pass "an untracked named path is refused, named with its file"
else
  fail "an untracked named path is refused, named with its file (rc=$RC: $OUT)"
fi

track "$r1" "shots/27.png" "x"
run_guard "$r1"
if [ "$RC" -eq 0 ] && [ "$OUT" = "DONE-WHEN-OK: $r1" ]; then
  pass "the same path once tracked answers DONE-WHEN-OK"
else
  fail "the same path once tracked answers DONE-WHEN-OK (rc=$RC: $OUT)"
fi

r2="$(new_repo)"
track "$r2" "docs/tickets/kan-x.md" "## Done when
Snapshots 27/28/29 re-baselined, see shots/*.png and https://x.test/shots/27.png.
"
run_guard "$r2"
if [ "$RC" -eq 0 ] && [ "$OUT" = "DONE-WHEN-OK: $r2" ]; then
  pass "snapshot numbers, globs and URLs never read as paths"
else
  fail "snapshot numbers, globs and URLs never read as paths (rc=$RC: $OUT)"
fi

track "$r2" "src/Makefile" "x"
run_guard "$r2"
if [ "$RC" -eq 0 ]; then
  pass "a tracked dotless name answers DONE-WHEN-OK"
else
  fail "a tracked dotless name answers DONE-WHEN-OK (rc=$RC: $OUT)"
fi

track "$r2" "docs/tickets/kan-y.md" "## Done when
- (\`shots/27.png\` re-baselined)
- \`src/Makefile\` — untracked here
"
git -C "$r2" rm -q --cached src/Makefile
rm -f "$r2/src/Makefile"
run_guard "$r2"
if [ "$RC" -eq 1 ] &&
  grep -q "^DONE-WHEN-PATH: shots/27.png — docs/tickets/kan-y.md$" <<<"$OUT" &&
  grep -q "^DONE-WHEN-PATH: src/Makefile — docs/tickets/kan-y.md$" <<<"$OUT" &&
  grep -q "^DONE-WHEN-VIOLATION: $r2 — 2$" <<<"$OUT"; then
  pass "a wrapped backticked path and a now-untracked dotless name are both refused"
else
  fail "a wrapped backticked path and a now-untracked dotless name are both refused (rc=$RC: $OUT)"
fi

r3="$(new_repo)"
track "$r3" "docs/tickets/kan-x.md" "## Scope

\`\`\`markdown
## Done when

\`shots/27.png\` re-baselined.
\`\`\`

## Done when

\`\`\`bash
# rebuild the fixtures
\`\`\`

\`shots/27.png\` re-baselined.
"
run_guard "$r3"
if [ "$RC" -eq 1 ] &&
  grep -q "^DONE-WHEN-PATH: shots/27.png — docs/tickets/kan-x.md$" <<<"$OUT" &&
  [ "$(grep -cv "^DONE-WHEN-PATH:" <<<"$OUT")" -eq 1 ]; then
  pass "a quoted template opens no section and a fence line closes nothing"
else
  fail "a quoted template opens no section and a fence line closes nothing (rc=$RC: $OUT)"
fi

r4="$(new_repo)"
run_guard "$r4"
if [ "$RC" -eq 0 ] && [ "$OUT" = "DONE-WHEN-OK: $r4" ]; then
  pass "a repository with no Done-when section is DONE-WHEN-OK"
else
  fail "a repository with no Done-when section is DONE-WHEN-OK (rc=$RC: $OUT)"
fi

run_guard /tmp/kan-817-nonexistent-repo-probe
if [ "$RC" -eq 2 ] && [ -z "$OUT" ] && [ -n "$ERR" ]; then
  pass "an unreadable worktree cannot answer: exit 2, nothing on stdout"
else
  fail "an unreadable worktree cannot answer: exit 2, nothing on stdout (rc=$RC: $OUT)"
fi

run_guard
if [ "$RC" -eq 2 ]; then
  pass "a usage error cannot answer"
else
  fail "a usage error cannot answer (rc=$RC)"
fi

if [ "$FAILURES" -eq 0 ]; then
  printf 'test-check-done-when-paths: all cases pass\n'
  exit 0
fi
printf 'test-check-done-when-paths: %d failure(s)\n' "$FAILURES" >&2
exit 1
