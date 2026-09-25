#!/usr/bin/env bash
# Assertion harness for check-planning-commit-location.sh. Builds sandboxed
# git repositories under TMPDIR; never touches the real repository tree.
#
# Asserts the guard's exit contract: 0 with the OK line in a linked worktree
# on spectre/<name>; 1 with the main-checkout line in a main checkout (on any
# branch, spectre/<name> included); 1 with the wrong-branch line in a linked
# worktree on another branch or a detached HEAD; both lines together when
# both hold; 2 with nothing on stdout on bad arguments, a missing directory
# or a directory that is not a git work tree.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/check-planning-commit-location.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/check-planning-commit-location-test.XXXXXX")"
trap 'rm -rf "$SANDBOX"' EXIT

MAIN="$SANDBOX/repo"
git init -q -b main "$MAIN"
git -C "$MAIN" config user.email "test@example.com"
git -C "$MAIN" config user.name "Test"
printf 'seed\n' > "$MAIN/README.md"
git -C "$MAIN" add -A
git -C "$MAIN" commit -q -m seed
git -C "$MAIN" worktree add -q -b spectre/demo "$MAIN/.worktrees/demo"
git -C "$MAIN" worktree add -q -b spectre/other "$MAIN/.worktrees/other"
git -C "$MAIN" worktree add -q --detach "$MAIN/.worktrees/detached"
mkdir -p "$SANDBOX/plain" "$MAIN/.worktrees/demo/sub"

# run <label> <expected-rc> <stdout-pattern|EMPTY> <args...>
run() {
  local label="$1" want_rc="$2" want_out="$3"
  shift 3
  local out rc
  set +e
  out="$("$SCRIPT" "$@" 2>/dev/null)"
  rc=$?
  set -e
  if [ "$rc" -ne "$want_rc" ]; then
    fail "$label: rc=$rc want $want_rc out=$out"
    return
  fi
  if [ "$want_out" = "EMPTY" ]; then
    [ -z "$out" ] && pass "$label" || fail "$label: expected empty stdout, got: $out"
    return
  fi
  case "$out" in
    *"$want_out"*) pass "$label" ;;
    *) fail "$label: stdout lacks '$want_out': $out" ;;
  esac
}

run "linked worktree on spectre/<name> passes" 0 "PLANNING-COMMIT-LOCATION-OK: $MAIN/.worktrees/demo on spectre/demo" \
  "$MAIN/.worktrees/demo" demo
run "main checkout on its default branch is refused" 1 "PLANNING-COMMIT-MAIN-CHECKOUT: $MAIN" "$MAIN" demo
run "main checkout on the default branch is also the wrong branch" 1 "PLANNING-COMMIT-WRONG-BRANCH: $MAIN on main — expected spectre/demo" \
  "$MAIN" demo
run "linked worktree of another change is refused" 1 "PLANNING-COMMIT-WRONG-BRANCH: $MAIN/.worktrees/other on spectre/other — expected spectre/demo" \
  "$MAIN/.worktrees/other" demo
run "detached worktree is refused" 1 "on detached — expected spectre/demo" "$MAIN/.worktrees/detached" demo
run "a subdirectory of the change worktree passes" 0 "PLANNING-COMMIT-LOCATION-OK" "$MAIN/.worktrees/demo/sub" demo

# The main checkout moved onto spectre/<name> is still the main checkout.
git -C "$MAIN/.worktrees/demo" checkout -q --detach
git -C "$MAIN" checkout -q spectre/demo
run "main checkout on spectre/<name> is still refused" 1 "PLANNING-COMMIT-MAIN-CHECKOUT: $MAIN" "$MAIN" demo
set +e
OUT="$("$SCRIPT" "$MAIN" demo 2>/dev/null)"
set -e
case "$OUT" in
  *WRONG-BRANCH*) fail "main checkout on spectre/<name>: wrong-branch line printed: $OUT" ;;
  *) pass "main checkout on spectre/<name>: only the main-checkout line" ;;
esac

run "no arguments cannot answer" 2 EMPTY
run "one argument cannot answer" 2 EMPTY "$MAIN"
run "empty name cannot answer" 2 EMPTY "$MAIN" ""
run "missing directory cannot answer" 2 EMPTY "$SANDBOX/nope" demo
run "non-git directory cannot answer" 2 EMPTY "$SANDBOX/plain" demo

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all check-planning-commit-location.sh cases passed\n'
