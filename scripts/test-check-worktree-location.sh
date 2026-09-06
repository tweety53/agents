#!/usr/bin/env bash
# Assertion harness for check-worktree-location.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR, registers real worktrees in and out
# of <repo>/.worktrees/ with real `git worktree add`, and asserts the guard's
# STRAY lines, its verdict line and its exit status. Never touches the real
# repository tree.
#
# THE GRAMMAR THIS FILE IS THE EXECUTABLE STATEMENT OF (KAN-462 §11):
#
#   STRAY: <path> (<branch>|detached)      per worktree outside .worktrees/
#   LOCATION-OK: <project>                 every worktree is at or under it
#   LOCATION-STRAY: <project> — <n>        <n> worktree(s) are not
#
# Exit 0 on LOCATION-OK, 1 on LOCATION-STRAY, 2 with NOTHING on stdout when
# the argument is not a readable directory or `git worktree list` fails.
#
# Shape copied from test-check-worktree-processes.sh: sandboxed TMPDIR,
# pass/fail counters, a run_guard capturing stdout and stderr separately.
# Duplicated rather than shared for the same reason that file's header
# gives — the two suites test unrelated guards, and a shared library would
# mean a change to one guard's contract could only be made by editing a file
# the other one also runs.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# CHECK_WORKTREE_LOCATION_GUARD points this suite at a guard other than its
# sibling. It exists for ONE caller — the mutation case at the end of this
# file, which re-runs this whole suite against a deliberately broken copy of
# the guard and requires it to go red. Never set it for a normal invocation.
GUARD="${CHECK_WORKTREE_LOCATION_GUARD:-$SCRIPT_DIR/check-worktree-location.sh}"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOXES=()
cleanup() {
  if [ "${#SANDBOXES[@]}" -ne 0 ]; then
    for s in "${SANDBOXES[@]}"; do
      chmod -R u+rwX "$s" 2>/dev/null || true
      rm -rf "$s"
    done
  fi
}
trap cleanup EXIT

WORK="$(mktemp -d "${TMPDIR:-/tmp}/worktree-location-test.XXXXXX")"
SANDBOXES+=("$WORK")
ERRFILE="$WORK/stderr"

# run_guard <arg ...> -> sets OUT (stdout only), ERR, RC. The two streams are
# captured separately, never merged with 2>&1: a refusal puts its message on
# stderr and must leave stdout empty, and a merged capture cannot tell an
# empty stdout from a stdout carrying the message.
run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# new_repo -> sets REPO (physical form) to a freshly-initialised git repo
# with one commit, so `git worktree add` has a HEAD to branch from.
new_repo() {
  local r
  r="$(mktemp -d "${TMPDIR:-/tmp}/worktree-location-repo.XXXXXX")"
  SANDBOXES+=("$r")
  git init -q "$r"
  git -C "$r" commit -q --allow-empty -m init
  REPO="$(cd "$r" && pwd -P)"
}

# assert_ok <label> — exit 0, a LOCATION-OK verdict naming the project, and
# no STRAY line anywhere in stdout.
assert_ok() {
  if [ "$RC" -ne 0 ]; then
    fail "$1: expected exit 0, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  case "$OUT" in
    "LOCATION-OK: $REPO") pass "$1: prints LOCATION-OK for the project" ;;
    *) fail "$1: expected 'LOCATION-OK: $REPO', got: $OUT" ;;
  esac
  case "$OUT" in
    *STRAY*) fail "$1: an OK verdict carried a STRAY line: $OUT" ;;
    *) pass "$1: carries no STRAY line" ;;
  esac
}

# assert_stray <label> <path> <branch-or-detached> <n> — exit 1, a STRAY line
# naming that path and branch, and a LOCATION-STRAY verdict naming the count.
assert_stray() {
  if [ "$RC" -ne 1 ]; then
    fail "$1: expected exit 1, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  case "$OUT" in
    *"STRAY: $2 ($3)"*) pass "$1: reports STRAY for $2 ($3)" ;;
    *) fail "$1: expected 'STRAY: $2 ($3)', got: $OUT" ;;
  esac
  case "$OUT" in
    *"LOCATION-STRAY: $REPO — $4"*) pass "$1: verdict names $4 stray worktree(s)" ;;
    *) fail "$1: expected 'LOCATION-STRAY: $REPO — $4', got: $OUT" ;;
  esac
}

# assert_cannot_answer <label> — exit 2, an empty stdout, no verdict token on
# it, and a named reason on stderr.
assert_cannot_answer() {
  [ "$RC" -eq 2 ] && pass "$1: exits 2" \
    || fail "$1: expected exit 2, got rc=$RC out=$OUT err=$ERR"
  [ -z "$OUT" ] && pass "$1: writes nothing to stdout" \
    || fail "$1: emitted something on stdout: $OUT"
  case "$OUT" in
    *LOCATION-OK*|*LOCATION-STRAY*|*STRAY*) fail "$1: an inability printed a verdict token: $OUT" ;;
    *) pass "$1: prints no verdict token" ;;
  esac
  case "$ERR" in
    *"check-worktree-location: "*) pass "$1: names the failure on stderr" ;;
    *) fail "$1: no named message on stderr: $ERR" ;;
  esac
}

# ---------------------------------------------------------------------------
# Case 1: no worktree beyond the main checkout is trivially OK.
# ---------------------------------------------------------------------------
new_repo
run_guard "$REPO"
assert_ok "case 1: ok with no worktree"

# ---------------------------------------------------------------------------
# Case 2: one worktree, in-tree under .worktrees/.
# ---------------------------------------------------------------------------
new_repo
git -C "$REPO" worktree add -q "$REPO/.worktrees/in-tree" -b in-tree >/dev/null
run_guard "$REPO"
assert_ok "case 2: ok with one under .worktrees"

# ---------------------------------------------------------------------------
# Case 3: a stray sibling directory — the retired <project>-worktrees/ layout.
# ---------------------------------------------------------------------------
new_repo
SIBLING_DIR="$(mktemp -d "${TMPDIR:-/tmp}/worktree-location-sibling.XXXXXX")"
SANDBOXES+=("$SIBLING_DIR")
git -C "$REPO" worktree add -q "$SIBLING_DIR/x" -b sibling-branch >/dev/null
SIBLING_PHYS="$(cd "$SIBLING_DIR/x" && pwd -P)"
run_guard "$REPO"
assert_stray "case 3: stray sibling directory" "$SIBLING_PHYS" "refs/heads/sibling-branch" 1
git -C "$REPO" worktree remove --force "$SIBLING_DIR/x" >/dev/null 2>&1 || true

# ---------------------------------------------------------------------------
# Case 4: a stray detached worktree.
# ---------------------------------------------------------------------------
new_repo
DETACHED_DIR="$(mktemp -d "${TMPDIR:-/tmp}/worktree-location-detached.XXXXXX")"
SANDBOXES+=("$DETACHED_DIR")
git -C "$REPO" worktree add -q --detach "$DETACHED_DIR/x" >/dev/null
DETACHED_PHYS="$(cd "$DETACHED_DIR/x" && pwd -P)"
run_guard "$REPO"
assert_stray "case 4: stray detached" "$DETACHED_PHYS" "detached" 1

# ---------------------------------------------------------------------------
# Case 5: a stray whose path contains a space — read with substr, not $2.
# ---------------------------------------------------------------------------
new_repo
SPACED_DIR="$(mktemp -d "${TMPDIR:-/tmp}/worktree-location-spaced.XXXXXX")"
SANDBOXES+=("$SPACED_DIR")
mkdir -p "$SPACED_DIR/with space"
git -C "$REPO" worktree add -q --detach "$SPACED_DIR/with space/x" >/dev/null
SPACED_PHYS="$(cd "$SPACED_DIR/with space/x" && pwd -P)"
run_guard "$REPO"
assert_stray "case 5: stray path with a space" "$SPACED_PHYS" "detached" 1

# ---------------------------------------------------------------------------
# Case 5b: a worktree registered under `<project>/.worktrees-old/`, a sibling
# of `.worktrees` that shares it as a literal string PREFIX without being
# nested under it — the exact collision "AT OR UNDER IS NOT STARTS WITH"
# (header, line 33) exists to prevent: `root` here is `<project>/.worktrees`,
# and `<project>/.worktrees-old/x` starts with that literal string even
# though it is not at or under `<project>/.worktrees/`. A bare
# `index(w, root) != 1` (rather than `index(w, root "/") != 1`) would treat
# it as in-tree and misreport LOCATION-OK.
# ---------------------------------------------------------------------------
new_repo
PREFIX_SIBLING_DIR="$REPO/.worktrees-old"
mkdir -p "$PREFIX_SIBLING_DIR"
git -C "$REPO" worktree add -q --detach "$PREFIX_SIBLING_DIR/x" >/dev/null
PREFIX_SIBLING_PHYS="$(cd "$PREFIX_SIBLING_DIR/x" && pwd -P)"
run_guard "$REPO"
assert_stray "case 5b: sibling directory sharing .worktrees as a string prefix" "$PREFIX_SIBLING_PHYS" "detached" 1

# ---------------------------------------------------------------------------
# Case 6: the _landing-<name> worktree is in-tree, no rule of its own needed.
# ---------------------------------------------------------------------------
new_repo
git -C "$REPO" worktree add -q "$REPO/.worktrees/_landing-example" -b landing-branch >/dev/null
run_guard "$REPO"
assert_ok "case 6: _landing worktree is in-tree"

# ---------------------------------------------------------------------------
# Case 7: a non-worktree argument cannot be answered.
# ---------------------------------------------------------------------------
NOT_A_REPO="$(mktemp -d "${TMPDIR:-/tmp}/worktree-location-notrepo.XXXXXX")"
SANDBOXES+=("$NOT_A_REPO")
run_guard "$NOT_A_REPO"
assert_cannot_answer "case 7: exit 2 on a non-worktree argument"

# ---------------------------------------------------------------------------
# Case 8: KAN-197-style mutation — this suite must detect a guard that always
# prints LOCATION-OK. A suite that cannot is not a suite.
# ---------------------------------------------------------------------------
if [ -n "${CHECK_WORKTREE_LOCATION_GUARD:-}" ]; then
  # Skipped when this run IS the mutant run, which is how the recursion ends.
  printf 'skip: case 8 (mutation) (this run is itself the mutant run)\n'
else
  MUTANT_DIR="$WORK/mutant"
  mkdir -p "$MUTANT_DIR"
  MUTANT="$MUTANT_DIR/check-worktree-location.sh"
  # The mutation: force the guard's one decision — whether the accumulated
  # STRAY count is greater than zero — to always be false, so every run
  # reports LOCATION-OK regardless of what it actually found.
  sed 's/if \[ "\$COUNT" -gt 0 \]; then/if false; then/' "$GUARD" > "$MUTANT"
  chmod +x "$MUTANT"

  if cmp -s "$GUARD" "$MUTANT"; then
    fail "case 8 (mutation): the mutation did not apply — the guard no longer carries the line this case edits, so nothing was mutated"
  else
    pass "case 8 (mutation): the mutation applied"
    set +e
    CHECK_WORKTREE_LOCATION_GUARD="$MUTANT" "$SCRIPT_DIR/test-check-worktree-location.sh" >/dev/null 2>&1
    MUTANT_RC=$?
    set -e
    [ "$MUTANT_RC" -ne 0 ] && pass "case 8 (mutation): a guard that always prints LOCATION-OK fails this suite" \
      || fail "case 8 (mutation): a guard that always prints LOCATION-OK passed this suite, which therefore detects nothing"
  fi
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-worktree-location: all cases pass\n'
