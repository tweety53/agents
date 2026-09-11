#!/usr/bin/env bash
# Assertion harness for classify-untracked.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the pre-flight's exit
# status, stdout AND the state it leaves behind for every case. Never
# touches the real repository tree.
#
# As with test-prepare-archive-branch.sh: every case checks the exit code,
# stdout, and — for the cases that move, ignore or report entries — the
# actual resulting state of the worktree, the scratchpad and the
# worktree-local exclude file. An exit code alone proves nothing here.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/classify-untracked.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

REPOS=()
cleanup() {
  # ${REPOS[@]} is unset-expansion-unsafe under `set -u` on bash 3.2 when empty.
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

run_guard() {
  ERRFILE="$(mktemp "${TMPDIR:-/tmp}/classify-untracked-err.XXXXXX")"
  set +e
  OUT="$("$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
  rm -f "$ERRFILE"
}

expect_rc() {
  local case_name="$1" expected="$2"
  [ "$RC" -eq "$expected" ] && pass "$case_name: exit $expected" \
    || fail "$case_name: expected exit $expected, got rc=$RC out=[$OUT] err=[$ERR]"
}

# new_fixture -> sets MAIN, WT, EXCLUDE, SCRATCH. A main checkout with a
# linked worktree at MAIN/.worktrees/_landing-test, one tracked file
# committed on the worktree's branch, and the paths the script is contract-
# bound to resolve: the per-worktree exclude file and the scratchpad.
new_fixture() {
  local fixture
  fixture="$(mktemp -d "${TMPDIR:-/tmp}/classify-untracked-test.XXXXXX")"
  REPOS+=("$fixture")
  MAIN="$fixture/main"
  git init -q -b main "$MAIN"
  # pwd -P: the script resolves paths through `rev-parse --path-format=absolute`,
  # which resolves macOS's /tmp symlink — the assertions below compare against
  # the same physical spelling.
  MAIN="$(cd "$MAIN" && pwd -P)"
  git -C "$MAIN" config user.email test@example.invalid
  git -C "$MAIN" config user.name "Test"
  echo base > "$MAIN/tracked.txt"
  git -C "$MAIN" add tracked.txt
  git -C "$MAIN" commit -qm base
  mkdir -p "$MAIN/.worktrees"
  WT="$MAIN/.worktrees/_landing-test"
  git -C "$MAIN" worktree add -q -b topic "$WT"
  EXCLUDE="$(git -C "$WT" rev-parse --path-format=absolute --git-common-dir)/info/exclude"
  SCRATCH="$MAIN/.worktrees/_scratchpad"
}

# -- -- -- cases -- -- --

# 1. Missing argument: usage on stderr, exit 2, nothing on stdout.
run_guard
expect_rc "missing-argument" 2
case "$ERR" in *usage:*) pass "missing-argument: usage on stderr" ;; *) fail "missing-argument: expected usage on stderr, got [$ERR]" ;; esac
[ -z "$OUT" ] && pass "missing-argument: stdout empty" || fail "missing-argument: expected empty stdout, got '$OUT'"

# 2. Absent worktree path: exit 0, silent — there is nothing to classify,
#    and the guard downstream creates the worktree fresh and clean.
ABSENT="$(mktemp -d "${TMPDIR:-/tmp}/classify-untracked-test.XXXXXX")/absent"
REPOS+=("$(dirname "$ABSENT")")
run_guard "$ABSENT"
expect_rc "absent-worktree" 0
[ -z "$OUT" ] && pass "absent-worktree: stdout empty" || fail "absent-worktree: expected empty stdout, got '$OUT'"

# 3. A plain directory, not a git worktree: exit 2.
NOTGIT="$(mktemp -d "${TMPDIR:-/tmp}/classify-untracked-test.XXXXXX")"
REPOS+=("$NOTGIT")
run_guard "$NOTGIT"
expect_rc "not-a-worktree" 2
case "$ERR" in *"not a git worktree"*) pass "not-a-worktree: names the problem" ;; *) fail "not-a-worktree: expected 'not a git worktree' on stderr, got [$ERR]" ;; esac

# 4. Clean worktree: exactly `CLEAN`, and nothing created anywhere.
new_fixture
run_guard "$WT"
expect_rc "clean-worktree" 0
[ "$OUT" = "CLEAN" ] && pass "clean-worktree: stdout is exactly CLEAN" || fail "clean-worktree: expected 'CLEAN', got '$OUT'"
[ ! -e "$SCRATCH" ] && pass "clean-worktree: no scratchpad created" || fail "clean-worktree: scratchpad exists: $SCRATCH"

# 5. Stray capture: moved to the scratchpad, tree clean again.
new_fixture
echo shot > "$WT/shot.png"
run_guard "$WT"
expect_rc "capture-moved" 0
[ -f "$SCRATCH/shot.png" ] && pass "capture-moved: file in the scratchpad" || fail "capture-moved: no $SCRATCH/shot.png"
[ ! -e "$WT/shot.png" ] && pass "capture-moved: gone from the worktree" || fail "capture-moved: still in the worktree"
case "$OUT" in "CAPTURED: shot.png -> $SCRATCH/shot.png") pass "capture-moved: stdout names the move" ;; *) fail "capture-moved: expected the CAPTURED line, got '$OUT'" ;; esac
[ -z "$(git -C "$WT" status --porcelain)" ] && pass "capture-moved: worktree clean" || fail "capture-moved: worktree still dirty"

# 6. Local config: `.claude/` ignored in the worktree's own exclude, files
#    left on disk, tree clean.
new_fixture
mkdir -p "$WT/.claude"
echo cfg > "$WT/.claude/settings.json"
run_guard "$WT"
expect_rc "config-ignored" 0
[ -f "$WT/.claude/settings.json" ] && pass "config-ignored: files left on disk" || fail "config-ignored: .claude/ contents disturbed"
[ -f "$EXCLUDE" ] && grep -qxF ".claude/" "$EXCLUDE" \
  && pass "config-ignored: local exclude carries .claude/" \
  || fail "config-ignored: exclude missing .claude/ (file: $EXCLUDE)"
[ -z "$(git -C "$WT" status --porcelain)" ] && pass "config-ignored: worktree clean" || fail "config-ignored: worktree still dirty"

# 7. Asset: reported, never touched, tree still dirty.
new_fixture
echo bundle > "$WT/handoff.zip"
run_guard "$WT"
expect_rc "asset-reported" 0
[ -f "$WT/handoff.zip" ] && pass "asset-reported: file untouched" || fail "asset-reported: file disturbed"
[ "$OUT" = "ASSET: handoff.zip (operator decides — commit deliberately or delete)" ] \
  && pass "asset-reported: stdout is the ASSET line" \
  || fail "asset-reported: expected the ASSET line, got '$OUT'"
[ -n "$(git -C "$WT" status --porcelain)" ] && pass "asset-reported: tree still dirty for the guard" || fail "asset-reported: tree unexpectedly clean"

# 8. Mixed: each class settled per class, tracked modifications untouched.
new_fixture
echo shot > "$WT/shot.png"
mkdir -p "$WT/.claude"
echo cfg > "$WT/.claude/settings.json"
echo bundle > "$WT/handoff.zip"
echo edited >> "$WT/tracked.txt"
run_guard "$WT"
expect_rc "mixed-classes" 0
[ -f "$SCRATCH/shot.png" ] && pass "mixed-classes: capture moved" || fail "mixed-classes: capture not moved"
grep -qxF ".claude/" "$EXCLUDE" && pass "mixed-classes: config ignored" || fail "mixed-classes: config not ignored"
[ -f "$WT/handoff.zip" ] && pass "mixed-classes: asset untouched" || fail "mixed-classes: asset disturbed"
grep -q edited "$WT/tracked.txt" && pass "mixed-classes: tracked modification untouched" || fail "mixed-classes: tracked file disturbed"
printf '%s\n' "$OUT" | grep -q "^CAPTURED: shot.png" && printf '%s\n' "$OUT" | grep -q "^IGNORED: .claude/" \
  && printf '%s\n' "$OUT" | grep -q "^ASSET: handoff.zip" \
  && pass "mixed-classes: one report line per class" \
  || fail "mixed-classes: expected three report lines, got '$OUT'"

# 9. Scratchpad collision: an existing same-named file is never clobbered;
#    the incoming copy gains a timestamp prefix.
new_fixture
mkdir -p "$SCRATCH"
echo old > "$SCRATCH/shot.png"
echo new > "$WT/shot.png"
run_guard "$WT"
expect_rc "collision-not-clobbered" 0
[ "$(cat "$SCRATCH/shot.png")" = "old" ] && pass "collision-not-clobbered: original intact" || fail "collision-not-clobbered: original clobbered"
[ "$(ls "$SCRATCH" | wc -l | tr -d ' ')" -eq 2 ] && pass "collision-not-clobbered: both copies present" || fail "collision-not-clobbered: expected two files, got: $(ls "$SCRATCH")"

# -- -- -- verdict -- -- --

if [ "$FAILURES" -eq 0 ]; then
  echo "all classify-untracked cases passed"
  exit 0
fi
printf '%s\n' "$FAILURES case(s) failed" >&2
exit 1
