#!/usr/bin/env bash
# Assertion harness for check-panel-diff-size.sh. Builds a sandboxed git
# repository under TMPDIR for every case; never touches the real repository
# tree.
#
# Asserts against check-panel-diff-size.sh's own contract (its header
# comment is canonical): it sums insertions and deletions across committed
# work and the unstaged working tree, exits 0 at or under the cap, 1 over
# the cap, and 2 when it cannot answer at all — a missing or non-repository
# worktree, a merge base that does not resolve, or a cap that is not a
# non-negative decimal integer.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/check-panel-diff-size.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# Every case leaves one sandbox repo behind, removed on exit including on a
# failed assertion. An indexed array, not a space-separated string: mktemp
# paths under TMPDIR may contain spaces, and word-splitting a string would
# leak a sandbox whose path split and `rm -rf` the fragments.
REPOS=()
cleanup() {
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

# new_repo -> sets REPO to a freshly initialized git repo with an initial
# commit, so the merge base used by every case always resolves.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-diff-size-test.XXXXXX")"
  REPOS+=("$REPO")
  git -C "$REPO" init -q
  git -C "$REPO" config user.email "test@example.com"
  git -C "$REPO" config user.name "Test"
  printf 'seed\n' > "$REPO/README.md"
  git -C "$REPO" add -A
  git -C "$REPO" commit -q -m "seed"
}

# lines N -> N newline-terminated numbered lines, for producing a diff of a
# known, predictable size.
lines() {
  local n="$1"
  local i
  for ((i = 1; i <= n; i++)); do
    printf 'line %d\n' "$i"
  done
}

# ===========================================================================
# 1. A diff under the cap exits 0 and prints the count.
# ===========================================================================
new_repo
MERGE_POINT="$(git -C "$REPO" rev-parse HEAD)"
lines 10 > "$REPO/small.txt"
git -C "$REPO" add -A
git -C "$REPO" commit -q -m "small change"
set +e
OUT="$("$SCRIPT" "$REPO" "$MERGE_POINT" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "case 1: exits 0 for a diff under the cap" \
  || fail "case 1: rc=$RC out=$OUT"
case "$OUT" in
  *"10"*) pass "case 1: prints the count" ;;
  *) fail "case 1: expected the count 10 in output, got: $OUT" ;;
esac

# ===========================================================================
# 2. A diff over the cap exits 1 and prints the count.
# ===========================================================================
new_repo
MERGE_POINT="$(git -C "$REPO" rev-parse HEAD)"
# 5200 rather than 2500: this case exists to prove the DEFAULT cap refuses a
# diff over it, so its magnitude has to track that default. It was 2500 while
# the default was 2000; the default is 5000 as of 2026-08-28 and this figure
# moved with it. A case whose size no longer exceeds the default would still
# pass its own name while testing nothing.
lines 5200 > "$REPO/big.txt"
git -C "$REPO" add -A
git -C "$REPO" commit -q -m "big change"
set +e
OUT="$("$SCRIPT" "$REPO" "$MERGE_POINT" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 1 ] && pass "case 2: exits 1 for a diff over the cap" \
  || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"5200"*) pass "case 2: prints the count" ;;
  *) fail "case 2: expected the count 5200 in output, got: $OUT" ;;
esac

# ===========================================================================
# 3. An explicit third argument overrides the default cap.
# ===========================================================================
new_repo
MERGE_POINT="$(git -C "$REPO" rev-parse HEAD)"
lines 10 > "$REPO/small.txt"
git -C "$REPO" add -A
git -C "$REPO" commit -q -m "small change"
set +e
OUT="$("$SCRIPT" "$REPO" "$MERGE_POINT" 5 2>&1)"
RC=$?
set -e
[ "$RC" -eq 1 ] && pass "case 3: explicit cap overrides the default" \
  || fail "case 3: rc=$RC out=$OUT"

# ===========================================================================
# 4. Unstaged working-tree changes count toward the total alongside
#    committed ones. The unstaged edit modifies a file already committed
#    since the merge base, since `git diff <merge-base>` (no --no-index)
#    never reports untracked files — matching the exact command
#    `/myflow-do` §5 uses to write `final-review.diff`.
# ===========================================================================
new_repo
MERGE_POINT="$(git -C "$REPO" rev-parse HEAD)"
lines 10 > "$REPO/committed.txt"
git -C "$REPO" add -A
git -C "$REPO" commit -q -m "committed change"
lines 20 > "$REPO/committed.txt"
set +e
OUT="$("$SCRIPT" "$REPO" "$MERGE_POINT" 15 2>&1)"
RC=$?
set -e
[ "$RC" -eq 1 ] && pass "case 4: unstaged changes count toward the total" \
  || fail "case 4: rc=$RC out=$OUT (expected over a cap of 15 with 10 committed + 10 unstaged insertions)"

# ===========================================================================
# 5. A non-numeric cap exits 2.
# ===========================================================================
new_repo
MERGE_POINT="$(git -C "$REPO" rev-parse HEAD)"
set +e
OUT="$("$SCRIPT" "$REPO" "$MERGE_POINT" "abc" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 5: non-numeric cap exits 2" \
  || fail "case 5: rc=$RC out=$OUT"

# ===========================================================================
# 6. A path that is not a git worktree exits 2.
# ===========================================================================
NOT_A_REPO="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-diff-size-test.XXXXXX")"
REPOS+=("$NOT_A_REPO")
set +e
OUT="$("$SCRIPT" "$NOT_A_REPO" "HEAD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 6: non-repository path exits 2" \
  || fail "case 6: rc=$RC out=$OUT"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-panel-diff-size: all cases pass\n'
