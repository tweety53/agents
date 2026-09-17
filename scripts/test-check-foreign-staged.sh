#!/usr/bin/env bash
# Assertion harness for check-foreign-staged.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR, stages real entries in them, and
# asserts the guard's FOREIGN-STAGED lines, its verdict line and its exit
# status. Never touches the real repository tree.
#
# THE GRAMMAR THIS FILE IS THE EXECUTABLE STATEMENT OF (KAN-546):
#
#   FOREIGN-STAGED: <porcelain line>       per staged or unmerged entry
#   STAGED-CLEAN: <main-checkout>          nothing staged
#   STAGED-FOREIGN: <main-checkout> — <n>  <n> staged/unmerged entries
#
# Exit 0 on either verdict; exit 2 with NOTHING on stdout when the argument
# is missing, is not a readable directory, is not a git repository, or
# `git status` fails there.
#
# Shape copied from test-check-worktree-location.sh: sandboxed TMPDIR,
# pass/fail counters, a run_guard capturing stdout and stderr separately.
# Duplicated rather than shared for the same reason that file's header
# gives — the two suites test unrelated guards, and a shared library would
# mean a change to one guard's contract could only be made by editing a file
# the other one also runs.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

GUARD="$SCRIPT_DIR/check-foreign-staged.sh"
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

WORK="$(mktemp -d "${TMPDIR:-/tmp}/foreign-staged-test.XXXXXX")"
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

# new_repo <name> -> prints the repo path, a git repository with one commit
# on `main` and a configured identity, so staged entries are the only
# variable between cases.
new_repo() {
  local repo="$WORK/$1"
  git init -q -b main "$repo"
  git -C "$repo" config user.email test@example.com
  git -C "$repo" config user.name test
  echo base >"$repo/base.txt"
  git -C "$repo" add base.txt
  git -C "$repo" commit -qm base
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

# ---- clean: nothing staged ---------------------------------------------
CLEAN="$(new_repo clean)"
run_guard "$CLEAN"
if [ "$RC" -eq 0 ] && [ "$OUT" = "STAGED-CLEAN: $CLEAN" ]; then
  pass "clean repo: STAGED-CLEAN verdict, exit 0"
else
  fail "clean repo: expected 'STAGED-CLEAN: $CLEAN' at exit 0, got RC=$RC OUT=<$OUT>"
fi

# ---- untracked-only still reads clean ----------------------------------
UNTRACKED="$(new_repo untracked)"
echo stray >"$UNTRACKED/stray.txt"
run_guard "$UNTRACKED"
if [ "$RC" -eq 0 ] && [ "$OUT" = "STAGED-CLEAN: $UNTRACKED" ]; then
  pass "untracked-only repo: STAGED-CLEAN verdict, exit 0"
else
  fail "untracked-only repo: expected 'STAGED-CLEAN: $UNTRACKED' at exit 0, got RC=$RC OUT=<$OUT>"
fi

# ---- one staged file ----------------------------------------------------
STAGED="$(new_repo staged)"
echo new >"$STAGED/new.txt"
git -C "$STAGED" add new.txt
run_guard "$STAGED"
if [ "$RC" -eq 0 ] \
  && [ "$(printf '%s\n' "$OUT" | grep -c '^FOREIGN-STAGED: ')" = "1" ] \
  && printf '%s\n' "$OUT" | grep -qF 'FOREIGN-STAGED: A  new.txt' \
  && [ "$(printf '%s\n' "$OUT" | grep -F 'STAGED-FOREIGN: ')" = "STAGED-FOREIGN: $STAGED — 1" ]; then
  pass "staged file: one FOREIGN-STAGED line, verdict counts 1, exit 0"
else
  fail "staged file: expected the A  new.txt line and 'STAGED-FOREIGN: $STAGED — 1', got RC=$RC OUT=<$OUT>"
fi

# ---- staged and unstaged mix: only the staged entry is listed -----------
MIXED="$(new_repo mixed)"
echo a2 >"$MIXED/a.txt"
git -C "$MIXED" add a.txt
echo modified >>"$MIXED/base.txt"
run_guard "$MIXED"
if [ "$RC" -eq 0 ] \
  && [ "$(printf '%s\n' "$OUT" | grep -c '^FOREIGN-STAGED: ')" = "1" ] \
  && printf '%s\n' "$OUT" | grep -qF 'FOREIGN-STAGED: A  a.txt' \
  && ! printf '%s\n' "$OUT" | grep -qF 'base.txt' \
  && [ "$(printf '%s\n' "$OUT" | grep -F 'STAGED-FOREIGN: ')" = "STAGED-FOREIGN: $MIXED — 1" ]; then
  pass "staged+unstaged mix: only the staged entry listed, unstaged named nowhere"
else
  fail "staged+unstaged mix: expected only 'A  a.txt' listed, got RC=$RC OUT=<$OUT>"
fi

# ---- staged deletion and staged addition count separately ---------------
TWO="$(new_repo two)"
git -C "$TWO" rm -q base.txt
echo added >"$TWO/added.txt"
git -C "$TWO" add added.txt
run_guard "$TWO"
if [ "$RC" -eq 0 ] \
  && [ "$(printf '%s\n' "$OUT" | grep -c '^FOREIGN-STAGED: ')" = "2" ] \
  && printf '%s\n' "$OUT" | grep -qF 'FOREIGN-STAGED: D  base.txt' \
  && printf '%s\n' "$OUT" | grep -qF 'FOREIGN-STAGED: A  added.txt' \
  && [ "$(printf '%s\n' "$OUT" | grep -F 'STAGED-FOREIGN: ')" = "STAGED-FOREIGN: $TWO — 2" ]; then
  pass "staged deletion + addition: both listed, verdict counts 2"
else
  fail "staged deletion + addition: expected both entries and '— 2', got RC=$RC OUT=<$OUT>"
fi

# ---- an unmerged entry is listed ----------------------------------------
# Both branches ADD f.txt with different content, so the merge stops on an
# add/add conflict, which porcelain prints as `AA f.txt` — both sides are
# staged, which is exactly the shape this guard lists.
MERGED="$(new_repo conflicted)"
git -C "$MERGED" checkout -qb side
echo side >"$MERGED/f.txt"
git -C "$MERGED" add f.txt
git -C "$MERGED" commit -qm side
git -C "$MERGED" checkout -q main
echo main >"$MERGED/f.txt"
git -C "$MERGED" add f.txt
git -C "$MERGED" commit -qm main
git -C "$MERGED" merge side >/dev/null 2>&1 || true
run_guard "$MERGED"
if [ "$RC" -eq 0 ] \
  && printf '%s\n' "$OUT" | grep -qF 'FOREIGN-STAGED: AA f.txt' \
  && [ "$(printf '%s\n' "$OUT" | grep -F 'STAGED-FOREIGN: ')" = "STAGED-FOREIGN: $MERGED — 1" ]; then
  pass "unmerged entry: AA listed, verdict counts 1"
else
  fail "unmerged entry: expected 'AA f.txt' listed and '— 1', got RC=$RC OUT=<$OUT>"
fi

# ---- an intent-to-add entry is listed -----------------------------------
# `git add -N` stages an index entry, so it is index work like any other
# stage — but porcelain prints its index code blank, ` A ita.txt`, which a
# first-column filter alone would hide. The preflight's identical status read
# refuses on it, so the guard lists it: the second-column A is what gives it
# away (panel finding F2, round 1).
ITA="$(new_repo intent-to-add)"
echo pending >"$ITA/pending.txt"
git -C "$ITA" add -N pending.txt
run_guard "$ITA"
if [ "$RC" -eq 0 ] \
  && printf '%s\n' "$OUT" | grep -qF 'FOREIGN-STAGED:  A pending.txt' \
  && [ "$(printf '%s\n' "$OUT" | grep -F 'STAGED-FOREIGN: ')" = "STAGED-FOREIGN: $ITA — 1" ]; then
  pass "intent-to-add entry: the A-index line listed, verdict counts 1"
else
  fail "intent-to-add entry: expected ' A pending.txt' listed and '— 1', got RC=$RC OUT=<$OUT>"
fi

# ---- verdict names the physical path (macOS /tmp symlink) ---------------
LINK="$WORK/via-link"
ln -s "$STAGED" "$LINK"
run_guard "$LINK"
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | grep -qF "STAGED-FOREIGN: $STAGED — 1"; then
  pass "physical path: a symlinked argument resolves to the real path in the verdict"
else
  fail "physical path: expected the verdict to name $STAGED, got RC=$RC OUT=<$OUT>"
fi

if [ "$FAILURES" -eq 0 ]; then
  echo "all check-foreign-staged cases pass"
  exit 0
fi
printf '%d case(s) failed\n' "$FAILURES" >&2
exit 1
