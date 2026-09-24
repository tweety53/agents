#!/usr/bin/env bash
# Assertion harness for check-main-checkout-drift.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR — each with its own bare `origin`
# so `refs/remotes/origin/HEAD` resolves the way a real clone's does — and
# asserts the guard's verdict lines and exit status. Never touches the real
# repository tree.
#
# THE GRAMMAR THIS FILE IS THE EXECUTABLE STATEMENT OF (KAN-647):
#
#   DRIFT-BRANCH: <path> — on <branch>, not <default>   not on the default branch
#   DRIFT-DIRTY: <path> — <n> tracked entries            tracked changes present
#   DRIFT-CLEAN: <path> — <default>                      neither finding
#
# DRIFT-BRANCH and DRIFT-DIRTY are findings: either or both may print, in
# that order, and DRIFT-CLEAN prints exactly when neither does. Exit 0 on
# any verdict; exit 2 with NOTHING on stdout when the argument is missing,
# is not a readable directory, is not a git repository, `origin/HEAD` does
# not resolve there, or the status read fails.
#
# Shape copied from test-check-foreign-staged.sh: sandboxed TMPDIR,
# pass/fail counters, a run_guard capturing stdout and stderr separately.
# Duplicated rather than shared for the same reason that file's header
# gives — the suites test unrelated guards, and a shared library would
# mean a change to one guard's contract could only be made by editing a
# file the other one also runs.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

GUARD="$SCRIPT_DIR/check-main-checkout-drift.sh"
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

WORK="$(mktemp -d "${TMPDIR:-/tmp}/main-checkout-drift-test.XXXXXX")"
SANDBOXES+=("$WORK")
# Physical form immediately, for the same reason the guard resolves its own
# argument that way: macOS's /tmp is a symlink to /private/tmp, and an
# expectation built from the symlinked mktemp path would never match a
# verdict that names the real one.
WORK="$(cd "$WORK" && pwd -P)"
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

# new_repo <name> -> prints the repo path: a git repository with one commit
# on `main`, a configured identity, and a bare `origin` whose HEAD points
# back at main — the shape every real clone's default-branch resolution
# depends on.
new_repo() {
  local repo="$WORK/$1"
  git init -q -b main "$repo"
  git -C "$repo" config user.email test@example.com
  git -C "$repo" config user.name test
  echo base >"$repo/base.txt"
  git -C "$repo" add base.txt
  git -C "$repo" commit -qm base
  git clone -q --bare "$repo" "$WORK/$1-origin.git"
  git -C "$repo" remote add origin "$WORK/$1-origin.git"
  git -C "$repo" fetch -q origin
  # set-head prints its own note on stdout — silent it, or the note rides
  # along inside every new_repo return value and poisons each $repo path.
  git -C "$repo" remote set-head origin -a >/dev/null
  printf '%s\n' "$repo"
}

# ---- usage: no argument ------------------------------------------------
run_guard
if [ "$RC" -eq 2 ] && [ -z "$OUT" ] && [ -n "$ERR" ]; then
  pass "no argument: exit 2, stdout empty, usage on stderr"
else
  fail "no argument: expected exit 2 with empty stdout and non-empty stderr, got RC=$RC OUT=<$OUT>"
fi

# ---- refusal: not a git repository -------------------------------------
NOTGIT="$WORK/not-git"
mkdir -p "$NOTGIT"
run_guard "$NOTGIT"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "non-git directory: exit 2, nothing on stdout"
else
  fail "non-git directory: expected exit 2 with empty stdout, got RC=$RC OUT=<$OUT>"
fi

# ---- refusal: unreadable path ------------------------------------------
run_guard "$WORK/no-such-path"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "unreadable path: exit 2, nothing on stdout"
else
  fail "unreadable path: expected exit 2 with empty stdout, got RC=$RC OUT=<$OUT>"
fi

# ---- refusal: no origin/HEAD -------------------------------------------
# A repository with no remote-tracking HEAD has no default branch the guard
# could name — an inability, never a verdict. `symbolic-ref -d` removes the
# symbolic ref itself; `update-ref -d` on a symref dereferences and would
# delete the branch instead.
NOORIGIN="$(new_repo no-origin)"
git -C "$NOORIGIN" symbolic-ref -d refs/remotes/origin/HEAD
run_guard "$NOORIGIN"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "no origin/HEAD: exit 2, nothing on stdout"
else
  fail "no origin/HEAD: expected exit 2 with empty stdout, got RC=$RC OUT=<$OUT>"
fi

# ---- refusal: a dangling origin/HEAD ------------------------------------
# Deleting the branch a symref points at leaves the symref behind, naming a
# branch that resolves to nothing — no answer either, never a verdict.
DANGLE="$(new_repo dangling)"
git -C "$DANGLE" update-ref -d refs/remotes/origin/main
run_guard "$DANGLE"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "dangling origin/HEAD: exit 2, nothing on stdout"
else
  fail "dangling origin/HEAD: expected exit 2 with empty stdout, got RC=$RC OUT=<$OUT>"
fi

# ---- clean: on the default branch, nothing tracked ----------------------
CLEAN="$(new_repo clean)"
run_guard "$CLEAN"
if [ "$RC" -eq 0 ] && [ "$OUT" = "DRIFT-CLEAN: $CLEAN — main" ]; then
  pass "clean repo: DRIFT-CLEAN verdict names the default branch, exit 0"
else
  fail "clean repo: expected 'DRIFT-CLEAN: $CLEAN — main' at exit 0, got RC=$RC OUT=<$OUT>"
fi

# ---- untracked-only still reads clean ----------------------------------
UNTRACKED="$(new_repo untracked)"
echo stray >"$UNTRACKED/stray.txt"
run_guard "$UNTRACKED"
if [ "$RC" -eq 0 ] && [ "$OUT" = "DRIFT-CLEAN: $UNTRACKED — main" ]; then
  pass "untracked-only repo: DRIFT-CLEAN verdict, exit 0"
else
  fail "untracked-only repo: expected 'DRIFT-CLEAN: $UNTRACKED — main', got RC=$RC OUT=<$OUT>"
fi

# ---- foreign branch ------------------------------------------------------
FOREIGN="$(new_repo foreign)"
git -C "$FOREIGN" checkout -qb feature
run_guard "$FOREIGN"
if [ "$RC" -eq 0 ] \
  && [ "$(printf '%s\n' "$OUT" | grep -c '^DRIFT-')" = "1" ] \
  && [ "$(printf '%s\n' "$OUT" | grep -F 'DRIFT-BRANCH: ')" = "DRIFT-BRANCH: $FOREIGN — on feature, not main" ]; then
  pass "foreign branch: one DRIFT-BRANCH finding names both branches, exit 0"
else
  fail "foreign branch: expected 'DRIFT-BRANCH: $FOREIGN — on feature, not main', got RC=$RC OUT=<$OUT>"
fi

# ---- detached HEAD -------------------------------------------------------
DETACHED="$(new_repo detached)"
git -C "$DETACHED" checkout -q --detach HEAD
run_guard "$DETACHED"
if [ "$RC" -eq 0 ] \
  && printf '%s\n' "$OUT" | grep -qF "DRIFT-BRANCH: $DETACHED — on (detached HEAD), not main"; then
  pass "detached HEAD: the finding says so instead of an empty branch name"
else
  fail "detached HEAD: expected 'on (detached HEAD), not main', got RC=$RC OUT=<$OUT>"
fi

# ---- unstaged drift on the default branch --------------------------------
UNSTAGED="$(new_repo unstaged)"
echo reverted >>"$UNSTAGED/base.txt"
run_guard "$UNSTAGED"
if [ "$RC" -eq 0 ] \
  && [ "$(printf '%s\n' "$OUT" | grep -c '^DRIFT-')" = "1" ] \
  && [ "$(printf '%s\n' "$OUT" | grep -F 'DRIFT-DIRTY: ')" = "DRIFT-DIRTY: $UNSTAGED — 1 tracked entries" ]; then
  pass "unstaged drift: one DRIFT-DIRTY finding counts 1, exit 0"
else
  fail "unstaged drift: expected 'DRIFT-DIRTY: $UNSTAGED — 1 tracked entries', got RC=$RC OUT=<$OUT>"
fi

# ---- staged drift counts too, and staged + unstaged add up ---------------
MIXED="$(new_repo mixed)"
echo staged >"$MIXED/staged.txt"
git -C "$MIXED" add staged.txt
echo reverted >>"$MIXED/base.txt"
run_guard "$MIXED"
if [ "$RC" -eq 0 ] \
  && [ "$(printf '%s\n' "$OUT" | grep -F 'DRIFT-DIRTY: ')" = "DRIFT-DIRTY: $MIXED — 2 tracked entries" ]; then
  pass "staged + unstaged mix: both entries count, verdict says 2"
else
  fail "staged + unstaged mix: expected '— 2 tracked entries', got RC=$RC OUT=<$OUT>"
fi

# ---- foreign branch and dirty content name both findings -----------------
BOTH="$(new_repo both)"
git -C "$BOTH" checkout -qb kan-527
echo reverted >>"$BOTH/base.txt"
run_guard "$BOTH"
if [ "$RC" -eq 0 ] \
  && printf '%s\n' "$OUT" | grep -qF "DRIFT-BRANCH: $BOTH — on kan-527, not main" \
  && printf '%s\n' "$OUT" | grep -qF "DRIFT-DIRTY: $BOTH — 1 tracked entries" \
  && [ "$(printf '%s\n' "$OUT" | grep -c '^DRIFT-')" = "2" ]; then
  pass "foreign branch + reverted content: both findings print, exit 0"
else
  fail "foreign branch + reverted content: expected both findings, got RC=$RC OUT=<$OUT>"
fi

# ---- verdict names the physical path (macOS /tmp symlink) ---------------
LINK="$WORK/via-link"
ln -s "$FOREIGN" "$LINK"
run_guard "$LINK"
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | grep -qF "DRIFT-BRANCH: $FOREIGN — on feature, not main"; then
  pass "physical path: a symlinked argument resolves to the real path in the verdict"
else
  fail "physical path: expected the verdict to name $FOREIGN, got RC=$RC OUT=<$OUT>"
fi

if [ "$FAILURES" -eq 0 ]; then
  echo "all check-main-checkout-drift cases pass"
  exit 0
fi
printf '%d case(s) failed\n' "$FAILURES" >&2
exit 1
