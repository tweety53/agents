#!/usr/bin/env bash
# Assertion harness for commit-split.sh. Builds a sandboxed git repository
# under TMPDIR for every case; never touches the real repository tree.
#
# Asserts against commit-split.sh's own contract (mirrored from
# skills/myflow-contracts/pipeline.md's "Git boundaries" section, the
# canonical spec this script implements): two guarded commits, each skipped
# rather than failed when its staging area is empty, and a tracked symlink at
# either planning path stops the run with git's own exit 128 rather than
# being worked around.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/commit-split.sh"
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
# commit, so `git diff --cached --quiet` and `git commit` both have a HEAD to
# work against from the first case onward.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/commit-split-test.XXXXXX")"
  REPOS+=("$REPO")
  git -C "$REPO" init -q
  git -C "$REPO" config user.email "test@example.com"
  git -C "$REPO" config user.name "Test"
  mkdir -p "$REPO/spectre" "$REPO/docs/superpowers"
  printf 'seed\n' > "$REPO/README.md"
  printf 'seed\n' > "$REPO/spectre/seed.md"
  printf 'seed\n' > "$REPO/docs/superpowers/seed.md"
  git -C "$REPO" add -A
  git -C "$REPO" commit -q -m "seed"
}

log_subjects() {
  git -C "$REPO" log --format=%s
}

# ===========================================================================
# 1. Both commits happen when both staging areas have changes.
# ===========================================================================
new_repo
printf 'impl change\n' > "$REPO/src.md"
printf 'plan change\n' > "$REPO/spectre/plan.md"
set +e
OUT="$("$SCRIPT" "$REPO" demo "impl: case1" "plan: case1" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] || fail "case 1: rc=$RC out=$OUT"
SUBJECTS="$(log_subjects)"
case "$SUBJECTS" in
  *"impl: case1"*) pass "case 1: implementation commit made" ;;
  *) fail "case 1: implementation commit missing: $SUBJECTS" ;;
esac
case "$SUBJECTS" in
  *"plan: case1"*) pass "case 1: planning commit made" ;;
  *) fail "case 1: planning commit missing: $SUBJECTS" ;;
esac
COUNT="$(git -C "$REPO" log --oneline | wc -l | tr -d ' ')"
[ "$COUNT" -eq 3 ] && pass "case 1: exactly two new commits on top of seed" \
  || fail "case 1: expected 3 total commits, got $COUNT"

# ===========================================================================
# 2. Implementation commit skipped when only planning paths changed.
# ===========================================================================
new_repo
printf 'plan only\n' > "$REPO/docs/superpowers/only.md"
set +e
OUT="$("$SCRIPT" "$REPO" demo "impl: case2" "plan: case2" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] || fail "case 2: rc=$RC out=$OUT"
SUBJECTS="$(log_subjects)"
case "$SUBJECTS" in
  *"impl: case2"*) fail "case 2: implementation commit made when nothing implementation-side changed: $SUBJECTS" ;;
  *) pass "case 2: implementation commit skipped" ;;
esac
case "$SUBJECTS" in
  *"plan: case2"*) pass "case 2: planning commit made" ;;
  *) fail "case 2: planning commit missing: $SUBJECTS" ;;
esac

# ===========================================================================
# 3. Planning commit skipped when only implementation paths changed.
# ===========================================================================
new_repo
printf 'impl only\n' > "$REPO/only.md"
set +e
OUT="$("$SCRIPT" "$REPO" demo "impl: case3" "plan: case3" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] || fail "case 3: rc=$RC out=$OUT"
SUBJECTS="$(log_subjects)"
case "$SUBJECTS" in
  *"impl: case3"*) pass "case 3: implementation commit made" ;;
  *) fail "case 3: implementation commit missing: $SUBJECTS" ;;
esac
case "$SUBJECTS" in
  *"plan: case3"*) fail "case 3: planning commit made when nothing planning-side changed: $SUBJECTS" ;;
  *) pass "case 3: planning commit skipped" ;;
esac

# ===========================================================================
# 4. A tracked symlink at spectre/ stops the script with exit 128, no commit
#    made.
# ===========================================================================
new_repo
rm -rf "$REPO/spectre"
ln -s docs "$REPO/spectre"
git -C "$REPO" add -A
git -C "$REPO" commit -q -m "spectre becomes a tracked symlink"
BEFORE_COUNT="$(git -C "$REPO" log --oneline | wc -l | tr -d ' ')"
printf 'impl change\n' > "$REPO/only.md"
set +e
OUT="$("$SCRIPT" "$REPO" demo "impl: case4" "plan: case4" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 128 ] && pass "case 4: exits 128" || fail "case 4: rc=$RC out=$OUT"
AFTER_COUNT="$(git -C "$REPO" log --oneline | wc -l | tr -d ' ')"
[ "$AFTER_COUNT" -eq "$BEFORE_COUNT" ] && pass "case 4: no commit made" \
  || fail "case 4: commit count changed ($BEFORE_COUNT -> $AFTER_COUNT): $OUT"
case "$OUT" in
  *"beyond a symbolic link"* | *symbolic*) pass "case 4: git's own message surfaces" ;;
  *) fail "case 4: expected git's symlink message, got: $OUT" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'commit-split: all cases pass\n'
