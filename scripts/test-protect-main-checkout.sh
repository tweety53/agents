#!/usr/bin/env bash
# test-protect-main-checkout.sh — harness for hooks/protect-main-checkout.py.
#
# Builds one bare origin with a clone whose main checkout sits on `main`
# (origin/HEAD set), a linked worktree of it on a feature branch, and one
# repo with no origin at all, then feeds the hook PreToolUse payloads and
# asserts deny/allow on each. Cases:
#
#    1. Write inside the main checkout on main            -> deny
#    2. Edit inside the linked worktree                   -> allow
#    3. Write in the main checkout once on a feature branch -> allow
#    4. Bash `git commit` with cwd = main checkout        -> deny
#    5. Bash `git -C <main> add .` from elsewhere         -> deny
#    6. Bash `cd <main> && git reset HEAD~1`              -> deny
#    7. Bash `git -C <main> pull --ff-only`               -> allow
#    8. Bash `git -C <main> log`                          -> allow
#    9. Bash `git -C <worktree> commit`                   -> allow
#   10. Bash `sed -i '' ... <main>/file`                  -> deny
#   11. Bash `echo x > <main>/file`                       -> deny
#   12. Bash `cat <main>/file`                            -> allow
#   13. Write outside any repository                      -> allow
#   14. Malformed JSON on stdin                           -> allow, exit 0
#   15. Repo with no origin, checked out on develop       -> deny (fallback)
#   16. Bash `git -C <main> worktree add ...`             -> allow
#   17. Bash `git -C <main> stash`                        -> deny
#   18. Bash `tee <main>/file`                            -> deny
#   19. Bash `rm <main>/file`                             -> deny
#   20. Bash `land-self-review-report.sh <main> main ...` -> allow (a script, not a git verb)
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOK="$SCRIPT_DIR/../hooks/protect-main-checkout.py"
FAILURES=0
ROOT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/protect-main-test.XXXXXX")" && pwd -P)"
trap 'rm -rf "$ROOT"' EXIT

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# run <cwd> <tool> <json tool_input> -> prints "deny" or "allow"
run() {
  local out
  out="$(printf '{"tool_name":"%s","tool_input":%s,"cwd":"%s"}' "$2" "$3" "$1" | python3 "$HOOK")"
  if printf '%s' "$out" | grep -q '"permissionDecision": *"deny"'; then echo deny; else echo allow; fi
}
expect() { # <case> <want> <got>
  if [ "$2" = "$3" ]; then pass "$1"; else fail "$1: wanted $2, got $3"; fi
}
q() { python3 -c 'import json,sys; print(json.dumps(sys.argv[1]))' "$1"; }

git_q() { git -c user.email=t@t -c user.name=t -c commit.gpgsign=false "$@" >/dev/null 2>&1; }

# origin + main checkout on main
git_q init -q --bare "$ROOT/origin.git"
git_q clone -q "$ROOT/origin.git" "$ROOT/main"
MAIN="$ROOT/main"
echo a >"$MAIN/f.txt"
git_q -C "$MAIN" add f.txt
git_q -C "$MAIN" commit -q -m init
git_q -C "$MAIN" branch -M main
git_q -C "$MAIN" push -q -u origin main
git_q -C "$MAIN" remote set-head origin main
git_q -C "$MAIN" worktree add "$MAIN/.worktrees/change" -b change
WT="$MAIN/.worktrees/change"
# repo with no origin, on develop
git_q init -q -b develop "$ROOT/noorigin"
echo a >"$ROOT/noorigin/f.txt"
git_q -C "$ROOT/noorigin" add f.txt
git_q -C "$ROOT/noorigin" commit -q -m init

expect "1 Write inside main checkout on main" deny \
  "$(run "$ROOT" Write "{\"file_path\":$(q "$MAIN/new.txt"),\"content\":\"x\"}")"
expect "2 Edit inside linked worktree" allow \
  "$(run "$ROOT" Edit "{\"file_path\":$(q "$WT/f.txt"),\"old_string\":\"a\",\"new_string\":\"b\"}")"
expect "4 Bash git commit with cwd main" deny "$(run "$MAIN" Bash "{\"command\":\"git commit -m x\"}")"
expect "5 Bash git -C main add from elsewhere" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $MAIN add .")}")"
expect "6 Bash cd main && git reset" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $MAIN && git reset HEAD~1")}")"
expect "7 Bash git -C main pull --ff-only" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $MAIN pull --ff-only")}")"
expect "8 Bash git -C main log" allow "$(run "$ROOT" Bash "{\"command\":$(q "git -C $MAIN log --oneline -3")}")"
expect "9 Bash git -C worktree commit" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $WT commit -m x")}")"
expect "10 Bash sed -i into main" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "sed -i '' s/a/b/ $MAIN/f.txt")}")"
expect "11 Bash redirect into main" deny "$(run "$ROOT" Bash "{\"command\":$(q "echo x > $MAIN/f.txt")}")"
expect "12 Bash cat from main" allow "$(run "$ROOT" Bash "{\"command\":$(q "cat $MAIN/f.txt")}")"
expect "13 Write outside any repo" allow \
  "$(run "$ROOT" Write "{\"file_path\":$(q "$ROOT/loose.txt"),\"content\":\"x\"}")"
out="$(printf 'not json' | python3 "$HOOK")"; rc=$?
if [ "$rc" = 0 ] && [ -z "$out" ]; then pass "14 malformed JSON passes through"; else fail "14 malformed JSON: rc=$rc out=$out"; fi
expect "15 no-origin repo on develop" deny \
  "$(run "$ROOT" Write "{\"file_path\":$(q "$ROOT/noorigin/g.txt"),\"content\":\"x\"}")"
expect "16 Bash git worktree add from main" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $MAIN worktree add $MAIN/.worktrees/other -b other origin/main")}")"
expect "17 Bash git -C main stash" deny "$(run "$ROOT" Bash "{\"command\":$(q "git -C $MAIN stash")}")"
expect "18 Bash tee into main" deny "$(run "$ROOT" Bash "{\"command\":$(q "echo x | tee $MAIN/f.txt")}")"
expect "19 Bash rm in main" deny "$(run "$ROOT" Bash "{\"command\":$(q "rm $MAIN/f.txt")}")"
expect "20 landing script named, not a git verb" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "land-self-review-report.sh $MAIN main 'docs: x' docs/x.md --push main")}")"

# 3 last: moving the main checkout off main lifts the protection
git_q -C "$MAIN" checkout -q -b feature
expect "3 Write in main checkout on a feature branch" allow \
  "$(run "$ROOT" Write "{\"file_path\":$(q "$MAIN/new.txt"),\"content\":\"x\"}")"

if [ "$FAILURES" -gt 0 ]; then
  echo "$FAILURES failure(s)" >&2
  exit 1
fi
echo "all protect-main-checkout cases pass"
