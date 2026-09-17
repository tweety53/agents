#!/usr/bin/env bash
# test-refresh-main-checkout.sh — harness for refresh-main-checkout.sh.
#
# Each case builds a repo at commit A, then moves refs/heads/main to a
# later commit B with update-ref — exactly what a landing worktree's
# fast-forward does to the main checkout — so the index and worktree lag
# the branch pointer. Cases:
#
#   1. stale index + clean worktree           -> REFRESH-DONE, status clean, HEAD tree == B
#   2. already current                        -> REFRESH-CURRENT, exit 0
#   3. stale + an unstaged edit               -> REFRESH-REFUSED, exit 1, edit kept
#   4. stale + a staged edit (not any tip)    -> REFRESH-REFUSED, exit 1, edit kept
#   5. on another branch                      -> REFRESH-REFUSED, exit 1
#   6. detached HEAD                          -> REFRESH-REFUSED, exit 1
#   7. usage (one argument)                   -> exit 2
#   8. untracked files survive a refresh
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SUT="$SCRIPT_DIR/refresh-main-checkout.sh"
FAILURES=0
ROOTS=()
cleanup() { for r in "${ROOTS[@]:-}"; do [ -n "$r" ] && rm -rf "$r"; done; }
trap cleanup EXIT
fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }
git_q() { git -c user.email=t@t -c user.name=t -c commit.gpgsign=false "$@" >/dev/null 2>&1; }

# stale_repo: main checkout at A with refs/heads/main moved to B. Sets REPO and B.
stale_repo() {
  REPO="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/refresh-test.XXXXXX")" && pwd -P)"
  ROOTS+=("$REPO")
  git_q init -q -b main "$REPO"
  echo a >"$REPO/f.txt"
  git_q -C "$REPO" add f.txt
  git_q -C "$REPO" commit -q -m A
  echo b >"$REPO/f.txt"; echo new >"$REPO/g.txt"
  git_q -C "$REPO" add f.txt g.txt
  git_q -C "$REPO" commit -q -m B
  B="$(git -C "$REPO" rev-parse HEAD)"
  # back to A's tree in index and worktree, then move the pointer to B
  git_q -C "$REPO" reset -q --hard HEAD~1
  git_q -C "$REPO" update-ref refs/heads/main "$B"
}

# 1
stale_repo
[ "$(git -C "$REPO" status --porcelain | wc -l | tr -d ' ')" != 0 ] || fail "1 fixture: expected a stale status"
out="$("$SUT" "$REPO" main)"; rc=$?
if [ "$rc" = 0 ] && [[ "$out" == REFRESH-DONE:* ]] && [ -z "$(git -C "$REPO" status --porcelain)" ] \
  && [ "$(cat "$REPO/f.txt")" = b ] && [ -f "$REPO/g.txt" ]; then pass "1 stale index refreshed"; else fail "1 rc=$rc out=$out"; fi

# 2
out="$("$SUT" "$REPO" main)"; rc=$?
if [ "$rc" = 0 ] && [[ "$out" == REFRESH-CURRENT:* ]]; then pass "2 already current"; else fail "2 rc=$rc out=$out"; fi

# 3
stale_repo
echo edited >>"$REPO/f.txt"
out="$("$SUT" "$REPO" main)" && rc=0 || rc=$?
if [ "$rc" = 1 ] && [[ "$out" == *"unstaged changes"* ]] && grep -q edited "$REPO/f.txt"; then pass "3 unstaged edit refused and kept"; else fail "3 rc=$rc out=$out"; fi

# 4
stale_repo
echo real >"$REPO/h.txt"; git_q -C "$REPO" add h.txt
out="$("$SUT" "$REPO" main)" && rc=0 || rc=$?
if [ "$rc" = 1 ] && [[ "$out" == *"match no recent main tip"* ]] && [ -f "$REPO/h.txt" ]; then pass "4 real staged work refused and kept"; else fail "4 rc=$rc out=$out"; fi

# 5
stale_repo
git_q -C "$REPO" checkout -q -b other
out="$("$SUT" "$REPO" main)" && rc=0 || rc=$?
if [ "$rc" = 1 ] && [[ "$out" == *"is on other, not main"* ]]; then pass "5 other branch refused"; else fail "5 rc=$rc out=$out"; fi

# 6
stale_repo
git_q -C "$REPO" checkout -q --detach
out="$("$SUT" "$REPO" main)" && rc=0 || rc=$?
if [ "$rc" = 1 ] && [[ "$out" == *"is detached"* ]]; then pass "6 detached refused"; else fail "6 rc=$rc out=$out"; fi

# 7
"$SUT" "$REPO" >/dev/null 2>&1 && rc=0 || rc=$?
if [ "$rc" = 2 ]; then pass "7 usage"; else fail "7 rc=$rc"; fi

# 8
stale_repo
echo loose >"$REPO/untracked.txt"
"$SUT" "$REPO" main >/dev/null
if [ -f "$REPO/untracked.txt" ]; then pass "8 untracked survives"; else fail "8 untracked file lost"; fi

if [ "$FAILURES" -gt 0 ]; then echo "$FAILURES failure(s)" >&2; exit 1; fi
echo "all refresh-main-checkout cases pass"
