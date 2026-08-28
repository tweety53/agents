#!/usr/bin/env bash
# Assertion harness for check-base-moved.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's verdict and
# exit status. Never touches the real repository tree. Same shape as
# test-check-finish-preflight.sh: an indexed REPOS array (bash 3.2 has no
# associative arrays), removed by an EXIT trap, real git repositories rather
# than fixture trees.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in design.md and tasks.md's task 2, never against observed output.
#
# PROVED BY MUTATION, NOT MERELY ASSERTED (design.md: base-moved-is-a-guard).
# With the intersection step in check-base-moved.sh replaced by "treat every
# moved path as an overlap" (i.e. skip the `comm -12` narrowing and use
# MOVED_SORTED directly as the overlap set), case 2 below — a base moved by
# commits touching only files this change does not — turned from MOVED
# naming "no overlap" into MOVED naming "overlaps: unrelated1.txt,
# unrelated2.txt". Restoring the real intersection made case 2 pass again.
# This is the same manual-mutation-and-revert method
# test-check-unfinished-work.sh's case 8's header records; it is not
# reproduced automatically on every run.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-base-moved.sh"
# shellcheck source=lib/test-git-shim.sh
. "$SCRIPT_DIR/lib/test-git-shim.sh"
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

# run_guard <worktree> <base-ref> <recorded-merge-base> -> sets RC and OUT
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" "$2" "$3" 2>&1)"
  RC=$?
  set -e
}

# new_repo -> sets REPO, BASE_REF, RECORDED_BASE
# A repository on `main` with one commit carrying base.txt and shared.txt,
# plus a branch `demo` checked out at that same commit — shared.txt is
# tracked from the start so an unstaged-only edit on demo (case 5) is
# detected by `git diff` without an intervening `git add`.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/base-moved-test.XXXXXX")"
  REPOS+=("$REPO")
  git -C "$REPO" init -q -b main
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  echo base > "$REPO/base.txt"
  echo shared-base > "$REPO/shared.txt"
  git -C "$REPO" add base.txt shared.txt
  git -C "$REPO" commit -qm "base"
  RECORDED_BASE="$(git -C "$REPO" rev-parse HEAD)"
  BASE_REF=main
  git -C "$REPO" checkout -q -b demo
}

# advance_base <repo> <file>... — one commit per file on main, each touching
# that file (creating it if new, appending if already tracked), then return
# to demo. The commit count this produces is exactly the number of files
# given, so callers can assert an exact count.
advance_base() {
  local repo="$1"
  shift
  git -C "$repo" checkout -q main
  local f
  for f in "$@"; do
    echo moved >> "$repo/$f"
    git -C "$repo" add "$f"
    git -C "$repo" commit -qm "advance: $f"
  done
  git -C "$repo" checkout -q demo
}

# 1. An unmoved base: CLEAR.
new_repo
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  CLEAR*) pass "unmoved base -> CLEAR" ;;
  *) fail "unmoved base: expected CLEAR, got rc=$RC out=$OUT" ;;
esac
[ "$RC" -eq 0 ] && pass "unmoved base: exit 0" \
  || fail "unmoved base: expected exit 0, got rc=$RC"
case "$OUT" in
  *"has not moved since the recorded merge base"*) pass "unmoved base: names the reason" ;;
  *) fail "unmoved base: reason missing: $OUT" ;;
esac

# 2. Base moved touching only files this change does not: MOVED, no overlap,
#    and the commit count is the number actually added.
new_repo
advance_base "$REPO" unrelated1.txt unrelated2.txt
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  MOVED*) pass "moved with no overlap -> MOVED" ;;
  *) fail "moved no overlap: expected MOVED, got rc=$RC out=$OUT" ;;
esac
case "$OUT" in
  *"2 commits"*) pass "moved with no overlap: names the actual commit count" ;;
  *) fail "moved no overlap: commit count wrong: $OUT" ;;
esac
case "$OUT" in
  *"no overlap"*) pass "moved with no overlap: says no overlap" ;;
  *) fail "moved no overlap: does not say no overlap: $OUT" ;;
esac
case "$OUT" in
  *overlaps:*) fail "moved no overlap: names an overlap it should not have: $OUT" ;;
  *) pass "moved with no overlap: names no overlapping path" ;;
esac

# 3. Base moved touching a file this change also committed: MOVED, overlap
#    names that path.
new_repo
advance_base "$REPO" shared.txt
echo demo-committed >> "$REPO/shared.txt"
git -C "$REPO" add shared.txt
git -C "$REPO" commit -qm "demo touches shared, committed"
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  MOVED*) pass "moved with committed overlap -> MOVED" ;;
  *) fail "moved committed overlap: expected MOVED, got rc=$RC out=$OUT" ;;
esac
case "$OUT" in
  *"overlaps: shared.txt"*) pass "moved with committed overlap: names shared.txt" ;;
  *) fail "moved committed overlap: shared.txt not named: $OUT" ;;
esac

# 4. Same, but the change's touch is staged only: the path still appears,
#    proving the index is read.
new_repo
advance_base "$REPO" shared.txt
echo demo-staged >> "$REPO/shared.txt"
git -C "$REPO" add shared.txt
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  *"overlaps: shared.txt"*) pass "moved with staged-only overlap: names shared.txt" ;;
  *) fail "moved staged-only overlap: shared.txt not named: $OUT" ;;
esac

# 5. Same, but the change's touch is unstaged only: the path still appears,
#    proving the working tree is read.
new_repo
advance_base "$REPO" shared.txt
echo demo-unstaged >> "$REPO/shared.txt"
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  *"overlaps: shared.txt"*) pass "moved with unstaged-only overlap: names shared.txt" ;;
  *) fail "moved unstaged-only overlap: shared.txt not named: $OUT" ;;
esac

# 6. Overlap of more than 10 paths: the line carries 10 paths and a
#    "(+N more)" tail.
new_repo
WIDE_FILES=()
for n in 01 02 03 04 05 06 07 08 09 10 11; do
  f="wide$n.txt"
  echo "wide-base" > "$REPO/$f"
  WIDE_FILES+=("$f")
done
git -C "$REPO" add wide*.txt
git -C "$REPO" commit -qm "add wide files"
advance_base "$REPO" "${WIDE_FILES[@]}"
for f in "${WIDE_FILES[@]}"; do
  echo demo-wide >> "$REPO/$f"
  git -C "$REPO" add "$f"
done
git -C "$REPO" commit -qm "demo touches all wide files"
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  *"(+1 more)"*) pass "overlap over 10: carries a (+1 more) tail" ;;
  *) fail "overlap over 10: no (+1 more) tail: $OUT" ;;
esac
case "$OUT" in
  *"wide01.txt"*) pass "overlap over 10: includes the first sorted path" ;;
  *) fail "overlap over 10: wide01.txt missing: $OUT" ;;
esac
case "$OUT" in
  *"wide11.txt"*) fail "overlap over 10: eleventh path should be capped out, not named: $OUT" ;;
  *) pass "overlap over 10: eleventh path is not named individually" ;;
esac

# 6b. Overlap of EXACTLY 10 paths: the cap is `-gt 10`, so 10 must clear it —
#     all ten paths are named and no "(+N more)" tail appears. Case 6 above
#     only exercises 11 overlaps, which leaves the boundary itself unproven:
#     an off-by-one (`-ge 10`) would still pass case 6 but would wrongly cap
#     this exactly-10 case.
new_repo
WIDE_FILES=()
for n in 01 02 03 04 05 06 07 08 09 10; do
  f="wide$n.txt"
  echo "wide-base" > "$REPO/$f"
  WIDE_FILES+=("$f")
done
git -C "$REPO" add wide*.txt
git -C "$REPO" commit -qm "add wide files"
advance_base "$REPO" "${WIDE_FILES[@]}"
for f in "${WIDE_FILES[@]}"; do
  echo demo-wide >> "$REPO/$f"
  git -C "$REPO" add "$f"
done
git -C "$REPO" commit -qm "demo touches all wide files"
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  *"more)"*) fail "overlap of exactly 10: should have no (+N more) tail: $OUT" ;;
  *) pass "overlap of exactly 10: no (+N more) tail" ;;
esac
for f in "${WIDE_FILES[@]}"; do
  case "$OUT" in
    *"$f"*) : ;;
    *) fail "overlap of exactly 10: $f missing from the list: $OUT" ;;
  esac
done
pass "overlap of exactly 10: all ten paths named"

# 7. `-` as the recorded merge base: REFUSE, exit 0.
new_repo
run_guard "$REPO" "$BASE_REF" -
case "$OUT" in
  REFUSE*) pass "'-' recorded merge base -> REFUSE" ;;
  *) fail "'-' recorded merge base: expected REFUSE, got rc=$RC out=$OUT" ;;
esac
[ "$RC" -eq 0 ] && pass "'-' recorded merge base: exit 0" \
  || fail "'-' recorded merge base: expected exit 0, got rc=$RC"
case "$OUT" in
  *"no merge base recorded"*) pass "'-' recorded merge base: names the reason" ;;
  *) fail "'-' recorded merge base: reason missing: $OUT" ;;
esac

# 8. An unresolvable recorded merge base: REFUSE, exit 0.
new_repo
run_guard "$REPO" "$BASE_REF" 0000000000000000000000000000000000000000
case "$OUT" in
  REFUSE*) pass "unresolvable recorded merge base -> REFUSE" ;;
  *) fail "unresolvable recorded merge base: expected REFUSE, got rc=$RC out=$OUT" ;;
esac
[ "$RC" -eq 0 ] && pass "unresolvable recorded merge base: exit 0" \
  || fail "unresolvable recorded merge base: expected exit 0, got rc=$RC"

# 9. An unresolvable base ref: REFUSE, exit 0.
new_repo
run_guard "$REPO" no-such-base "$RECORDED_BASE"
case "$OUT" in
  REFUSE*) pass "unresolvable base ref -> REFUSE" ;;
  *) fail "unresolvable base ref: expected REFUSE, got rc=$RC out=$OUT" ;;
esac
[ "$RC" -eq 0 ] && pass "unresolvable base ref: exit 0" \
  || fail "unresolvable base ref: expected exit 0, got rc=$RC"

# 9b. `git rev-list` failing on the COUNT capture is an environment failure,
#     not a verdict: exit 2 and no verdict line. Same shim technique as
#     test-check-finish-preflight.sh's case 8b — fails only `rev-list`, passes
#     every other subcommand through, so this exercises the COUNT guard alone.
new_repo
advance_base "$REPO" unrelated1.txt
shim_failing_git "base-moved-shim" "rev-list" "simulated pack corruption"
set +e
OUT="$(PATH="$SHIM_DIR:$PATH" "$GUARD" "$REPO" "$BASE_REF" "$RECORDED_BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "unreadable commit count -> exit 2" \
  || fail "unreadable commit count: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  CLEAR*|MOVED*|REFUSE*) fail "unreadable commit count: emitted a verdict line: $OUT" ;;
  *) pass "unreadable commit count: emits no verdict line" ;;
esac
case "$OUT" in
  *check-base-moved:*) pass "unreadable commit count: names the failure" ;;
  *) fail "unreadable commit count: no named message: $OUT" ;;
esac
assert_shim_fired "$SHIM_DIR" "unreadable commit count"

# 9c. `git diff` failing on the MOVED_RAW capture (the base-side path list) is
#     an environment failure, not a verdict: exit 2 and no verdict line.
#     MOVED_RAW's call is given the SAME range string as COUNT's `rev-list`
#     call just before it, so shim_failing_git's single-argument match cannot
#     tell them apart — matching the range alone would fail COUNT's call
#     instead. This case shims directly, matching subcommand ($3) AND range
#     ($6) together: `-C <worktree> diff --name-only --end-of-options
#     <range>` is 6 arguments with "diff" third, while the `rev-list` call
#     has the same range but "rev-list" third. (KAN-88 fix round 2, F8.)
new_repo
advance_base "$REPO" unrelated1.txt
RANGE="${RECORDED_BASE}..${BASE_REF}"
SHIM_DIR="$(mktemp -d "${TMPDIR:-/tmp}/base-moved-shim.XXXXXX")"
REPOS+=("$SHIM_DIR")
{
  printf '#!/usr/bin/env bash\n'
  printf 'if [ "$3" = "diff" ] && [ "$6" = "%s" ]; then\n' "$RANGE"
  printf '  : > "%s/.fired"\n' "$SHIM_DIR"
  printf '  echo "fatal: simulated MOVED_RAW read failure" >&2\n'
  printf '  exit 128\n'
  printf 'fi\n'
  printf 'exec %s "$@"\n' "$TEST_GIT_SHIM_REAL_GIT"
} > "$SHIM_DIR/git"
chmod +x "$SHIM_DIR/git"
set +e
OUT="$(PATH="$SHIM_DIR:$PATH" "$GUARD" "$REPO" "$BASE_REF" "$RECORDED_BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "unreadable moved-paths list -> exit 2" \
  || fail "unreadable moved-paths list: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  CLEAR*|MOVED*|REFUSE*) fail "unreadable moved-paths list: emitted a verdict line: $OUT" ;;
  *) pass "unreadable moved-paths list: emits no verdict line" ;;
esac
case "$OUT" in
  *"changed on"*) pass "unreadable moved-paths list: names the failure" ;;
  *) fail "unreadable moved-paths list: no named message: $OUT" ;;
esac
assert_shim_fired "$SHIM_DIR" "unreadable moved-paths list"

# 9d. `git diff` failing on the COMMITTED_RAW capture (this change's own
#     committed paths) is an environment failure, not a verdict: exit 2 and
#     no verdict line. The match arg (RECORDED_BASE..HEAD) is unique to this
#     call — MOVED_RAW's call names BASE_REF, not HEAD, so it passes through
#     untouched and this exercises the COMMITTED_RAW guard alone.
#     (KAN-88 fix round 2, F8.)
new_repo
advance_base "$REPO" unrelated1.txt
echo work > "$REPO/demo.txt"
git -C "$REPO" add demo.txt
git -C "$REPO" commit -qm "demo work"
shim_failing_git "base-moved-shim" "${RECORDED_BASE}..HEAD" "simulated COMMITTED_RAW read failure"
set +e
OUT="$(PATH="$SHIM_DIR:$PATH" "$GUARD" "$REPO" "$BASE_REF" "$RECORDED_BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "unreadable committed-paths list -> exit 2" \
  || fail "unreadable committed-paths list: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  CLEAR*|MOVED*|REFUSE*) fail "unreadable committed-paths list: emitted a verdict line: $OUT" ;;
  *) pass "unreadable committed-paths list: emits no verdict line" ;;
esac
case "$OUT" in
  *"committed paths"*) pass "unreadable committed-paths list: names the failure" ;;
  *) fail "unreadable committed-paths list: no named message: $OUT" ;;
esac
assert_shim_fired "$SHIM_DIR" "unreadable committed-paths list"

# 9e. `git diff --cached` failing on the STAGED_RAW capture (this change's
#     staged paths) is an environment failure, not a verdict: exit 2 and no
#     verdict line. `--cached` is unique to this one call in the guard, so it
#     is the match arg — every other `diff`/`rev-list` call passes through.
#     (KAN-88 fix round 2, F8.)
new_repo
advance_base "$REPO" unrelated1.txt
echo staged > "$REPO/demo.txt"
git -C "$REPO" add demo.txt
shim_failing_git "base-moved-shim" "--cached" "simulated STAGED_RAW read failure"
set +e
OUT="$(PATH="$SHIM_DIR:$PATH" "$GUARD" "$REPO" "$BASE_REF" "$RECORDED_BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "unreadable staged-paths list -> exit 2" \
  || fail "unreadable staged-paths list: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  CLEAR*|MOVED*|REFUSE*) fail "unreadable staged-paths list: emitted a verdict line: $OUT" ;;
  *) pass "unreadable staged-paths list: emits no verdict line" ;;
esac
case "$OUT" in
  *"staged paths"*) pass "unreadable staged-paths list: names the failure" ;;
  *) fail "unreadable staged-paths list: no named message: $OUT" ;;
esac
assert_shim_fired "$SHIM_DIR" "unreadable staged-paths list"

# 9f. `git diff` failing on the UNSTAGED_RAW capture (this change's unstaged
#     paths) is an environment failure, not a verdict: exit 2 and no verdict
#     line. UNSTAGED_RAW's call is the only `diff --name-only` invocation
#     with no further argument — every other capture adds `--cached` or an
#     `--end-of-options <range>` — so shim_failing_git's single-argument
#     match cannot pick it out (`--name-only` alone would also catch the
#     other three). This case shims directly, matching by argument COUNT
#     instead: `-C <worktree> diff --name-only` is exactly 4 arguments, and
#     no other call in this guard has that shape. (KAN-88 fix round 2, F8.)
new_repo
advance_base "$REPO" unrelated1.txt
echo unstaged >> "$REPO/shared.txt"
SHIM_DIR="$(mktemp -d "${TMPDIR:-/tmp}/base-moved-shim.XXXXXX")"
REPOS+=("$SHIM_DIR")
{
  printf '#!/usr/bin/env bash\n'
  printf 'if [ "$#" -eq 4 ] && [ "$3" = "diff" ] && [ "$4" = "--name-only" ]; then\n'
  printf '  : > "%s/.fired"\n' "$SHIM_DIR"
  printf '  echo "fatal: simulated UNSTAGED_RAW read failure" >&2\n'
  printf '  exit 128\n'
  printf 'fi\n'
  printf 'exec %s "$@"\n' "$TEST_GIT_SHIM_REAL_GIT"
} > "$SHIM_DIR/git"
chmod +x "$SHIM_DIR/git"
set +e
OUT="$(PATH="$SHIM_DIR:$PATH" "$GUARD" "$REPO" "$BASE_REF" "$RECORDED_BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "unreadable unstaged-paths list -> exit 2" \
  || fail "unreadable unstaged-paths list: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  CLEAR*|MOVED*|REFUSE*) fail "unreadable unstaged-paths list: emitted a verdict line: $OUT" ;;
  *) pass "unreadable unstaged-paths list: emits no verdict line" ;;
esac
case "$OUT" in
  *"unstaged paths"*) pass "unreadable unstaged-paths list: names the failure" ;;
  *) fail "unreadable unstaged-paths list: no named message: $OUT" ;;
esac
assert_shim_fired "$SHIM_DIR" "unreadable unstaged-paths list"

# new_repo_with_origin -> sets REPO, ORIGIN, BASE_REF, RECORDED_BASE
# A bare `origin`, a clone of it with one commit on `main` (RECORDED_BASE,
# carrying base.txt and shared.txt), pushed back to origin, and a branch
# `demo` checked out at that same commit.
new_repo_with_origin() {
  ORIGIN="$(mktemp -d "${TMPDIR:-/tmp}/base-moved-test-origin.XXXXXX")"
  REPOS+=("$ORIGIN")
  git init -q --bare -b main "$ORIGIN"

  REPO="$(mktemp -d "${TMPDIR:-/tmp}/base-moved-test.XXXXXX")"
  REPOS+=("$REPO")
  git clone -q "$ORIGIN" "$REPO" 2>/dev/null
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  echo base > "$REPO/base.txt"
  echo shared-base > "$REPO/shared.txt"
  git -C "$REPO" add base.txt shared.txt
  git -C "$REPO" commit -qm "base"
  git -C "$REPO" push -q origin main
  RECORDED_BASE="$(git -C "$REPO" rev-parse HEAD)"
  BASE_REF=main
  git -C "$REPO" checkout -q -b demo
}

# advance_origin_main <repo> <file>... — one commit per file, landed on
# origin's `main` WITHOUT moving the repo's own local `main` branch, so local
# `main` stays at RECORDED_BASE while origin/main advances. Fetches
# afterward so the repo's own origin/main tracking ref reflects the push, and
# returns to `demo` with a clean tree.
advance_origin_main() {
  local repo="$1"
  shift
  git -C "$repo" branch -q tmp-advance main
  git -C "$repo" checkout -q tmp-advance
  local f
  for f in "$@"; do
    echo moved >> "$repo/$f"
    git -C "$repo" add "$f"
    git -C "$repo" commit -qm "advance: $f"
  done
  git -C "$repo" push -q origin tmp-advance:main
  git -C "$repo" checkout -q demo
  git -C "$repo" branch -q -D tmp-advance
  git -C "$repo" fetch -q origin
}

# 10. A bare local name whose origin/ counterpart exists and has moved: the
#     verdict names origin/<name>, proving the shared resolution
#     (scripts/lib/resolve-remote-base.sh) is in force here too.
new_repo_with_origin
advance_origin_main "$REPO" unrelated.txt
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  MOVED*) pass "bare name behind origin, origin moved -> MOVED" ;;
  *) fail "bare name behind origin: expected MOVED, got rc=$RC out=$OUT" ;;
esac
case "$OUT" in
  *"origin/main"*) pass "bare name behind origin: verdict names origin/main" ;;
  *) fail "bare name behind origin: verdict does not name origin/main: $OUT" ;;
esac

# 11. A missing argument: exit 2, no verdict line.
run_guard "" "" ""
[ "$RC" -eq 2 ] && pass "missing arguments -> exit 2" \
  || fail "missing arguments: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  CLEAR*|MOVED*|REFUSE*) fail "missing arguments: emitted a verdict line: $OUT" ;;
  *) pass "missing arguments: emits no verdict line" ;;
esac

# 12. A directory that is not a git worktree: exit 2, no verdict line.
NOT_A_REPO="$(mktemp -d "${TMPDIR:-/tmp}/base-moved-test-notrepo.XXXXXX")"
REPOS+=("$NOT_A_REPO")
run_guard "$NOT_A_REPO" main deadbeef
[ "$RC" -eq 2 ] && pass "non-repository -> exit 2" \
  || fail "non-repository: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  CLEAR*|MOVED*|REFUSE*) fail "non-repository: emitted a verdict line: $OUT" ;;
  *) pass "non-repository: emits no verdict line" ;;
esac


# 13. KAN-88 fix round 3 hardening: shim_failing_git's real-git resolution
#     (TEST_GIT_SHIM_REAL_GIT, captured once when test-git-shim.sh is
#     sourced) is immune to a shim already sitting on PATH. Proven directly:
#     build a first shim, put its directory on PATH the way a caller
#     mistakenly persisting PATH across two shim builds would, then build a
#     second shim while that PATH is in effect. If the second shim's "real
#     git" resolution saw the first shim (the bug this hardening rules out),
#     invoking the second shim with the FIRST shim's own match arg — an
#     argument the second shim's own matcher does not touch at all — would
#     fall through to the first shim and print ITS fatal message with exit
#     128. With the real git correctly resolved once at source time, that
#     same invocation reaches the genuine git binary, which rejects the
#     nonsense argument on its own terms and never emits the first shim's
#     text.
shim_failing_git "chain-shim-1" "chain-marker-one" "first shim fired"
FIRST_SHIM_DIR="$SHIM_DIR"
OLD_PATH="$PATH"
PATH="$FIRST_SHIM_DIR:$PATH"
shim_failing_git "chain-shim-2" "chain-marker-two" "second shim fired"
SECOND_SHIM_DIR="$SHIM_DIR"
PATH="$OLD_PATH"
set +e
CHAIN_OUT="$("$SECOND_SHIM_DIR/git" chain-marker-one 2>&1)"
CHAIN_RC=$?
set -e
case "$CHAIN_OUT" in
  *"first shim fired"*)
    fail "shim chaining: second shim's real-git resolved to the first shim: $CHAIN_OUT" ;;
  *) pass "shim chaining: second shim's real-git did not resolve to the first shim" ;;
esac
[ "$CHAIN_RC" -ne 128 ] \
  && pass "shim chaining: second shim exited a real-git error, not the first shim's 128" \
  || fail "shim chaining: second shim exited 128, matching the first shim's fatal exit"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-base-moved: all cases pass\n'
