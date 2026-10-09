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
#   21. Bash `mv <elsewhere>/x ~/.Trash/` with cwd = main   -> allow (tilde expands, not cwd-relative)
#   22. Bash `mv <main>/file /tmp/`                       -> deny (a move out of the checkout removes)
#   23. Bash `cd <worktree>; git commit` (`;` touching the path) -> allow
#   24. Bash `cd <worktree>; npx ... >$S/cap.log 2>&1` with cwd = main -> allow (unexpanded var)
#   25. Bash `cd <main>;git reset HEAD~1` (separators touching)  -> deny
#   38. Bash `git -C <wt> push origin HEAD:main`, change open      -> deny (unarchived landing)
#   39. Bash `git push origin change:main`, change open            -> deny
#   40. Bash `git -C <wt> push origin HEAD:main`, change archived  -> allow
#   41. Bash `git -C <wt> push origin change`, change open         -> allow (not a landing)
#   42. Bash `git -C <wt2> push origin spectre/flowchg:main`, the
#       /flow branch's change open                                   -> deny
#   43. Bash `sed -i '' … <outside file>` from the main checkout     -> allow
#       (BSD sed's empty suffix argument is not a path)
#   44. sibling layout: Edit in an existing <repo>-worktrees/<change> -> allow
#   45. sibling layout: Write into a not-yet-existing <repo>-worktrees/<change>
#       whose parent directory is itself a protected main checkout     -> allow
#   46. sibling layout: the deny reason suggests <repo>-worktrees/<change>,
#       never .worktrees/
#   47. sibling layout: Write into <X>-worktrees/<new> beside a directory X
#       that is no main checkout, inside a protected checkout          -> deny
#   48. heredoc append into a sibling worktree, cwd = main, body naming
#       `rm` and `>`                                                   -> allow (a body is data)
#   49. Bash `rm <sibling worktree>/file` with cwd = main             -> allow
#   50. Bash `rm f.txt` with cwd = main                               -> deny (relative)
#   51. heredoc write to <main>/file from elsewhere                   -> deny
#   52. a line after the heredoc's closing delimiter, `rm f.txt`      -> deny
#   53. `bash <<EOF` whose body runs `rm <main>/file`                 -> deny (a body fed to a shell is code)
#   54. `echo "<<EOF"`, then `rm f.txt`, cwd = main                   -> deny (a quoted `<<` opens nothing)
#   55. `true # <<EOF`, then `rm f.txt`, cwd = main                  -> deny (a commented `<<` opens nothing)
#   56. `echo $((1<<N))`, then `rm f.txt`, cwd = main                -> deny (an arithmetic shift opens nothing)
#   57. `cat <<E"OF"` closed by `EOF`, then `rm f.txt`, cwd = main   -> deny (the delimiter is unquoted whole)
#   58. `/bin/bash <<EOF` whose body runs `rm f.txt`, cwd = main     -> deny (a shell named by path)
#   59. `. /dev/stdin <<EOF` whose body runs `rm f.txt`, cwd = main  -> deny (`.` runs its input)
#   60. `dash <<EOF` whose body runs `rm f.txt`, cwd = main           -> deny (only a plain data write drops a body)
#   61. `cat <<EOF | dash`, body `rm f.txt`, cwd = main               -> deny
#   62. `$SHELL <<EOF`, body `rm f.txt`, cwd = main                    -> deny
#   63. `x=$(cat <<EOF` ... `EOF`, `)`, `eval "$x"`, cwd = main        -> deny
#   64. `read -r -d '' c <<EOF` ... `eval "$c"`, cwd = main           -> deny
#   65. `cat <<EOF > <sibling>/s.sh` ... `sh <sibling>/s.sh`, cwd = main -> deny
#   66. data write closed by `EOF\r`, then `rm f.txt`, cwd = main      -> deny (only blank lines may follow)
#   67. `echo ${x:-<<EOF}`, then `rm f.txt`, cwd = main               -> deny
#   68. `tee -a <sibling>/f <<'EOF'`, body naming `rm`, cwd = main    -> allow (a plain data write)
#   69. `cat >> <main>/f.txt <<EOF` from elsewhere                    -> deny (the target is judged)
#   70. `cat >> <sibling>/f <<EOF`, body `$(rm f.txt)`, cwd = main   -> deny (an unquoted body substitutes)
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
  if grep -q '"permissionDecision": *"deny"' <<<"$out"; then echo deny; else echo allow; fi
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

mkdir -p "$ROOT/elsewhere"; touch "$ROOT/elsewhere/x"
expect "21 tilde destination is not cwd-relative" allow \
  "$(HOME="$ROOT" run "$MAIN" Bash "{\"command\":$(q "mv $ROOT/elsewhere/x ~/.Trash/")}")"
expect "22 mv out of main checkout" deny "$(run "$ROOT" Bash "{\"command\":$(q "mv $MAIN/f.txt $ROOT/")}")"

expect "23 cd worktree; git commit" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $WT; git commit -m x")}")"
expect "24 cd worktree; redirect to an unexpanded var" allow \
  "$(run "$MAIN" Bash "{\"command\":$(q 'S=/tmp/x; cd '"$WT"'; npx playwright test a.spec.ts >$S/cap.log 2>&1; echo cap=$?')}")"
expect "25 cd main;git reset with separators touching" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $MAIN;git reset HEAD~1")}")"
expect "26 multi-line cp cannot swallow a later line's target" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "cp notes.md draft.md
echo done
git add $MAIN/README.md")}")"
expect "27 multi-line second line judged on its own" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "echo start
cd $MAIN && git reset HEAD~1")}")"
expect "28 assignment-set redirect into main resolves and denies" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "C=$MAIN; echo x > \$C/f.txt")}")"
expect "29 nested variable stays let-through" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "B=$MAIN; A=\$B; echo x > \$A/f.txt")}")"
expect "30 cp into a not-yet-existing worktree path" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "cp /tmp/a.md $MAIN/.worktrees/new-landing/spectre/foo.md")}")"
expect "31 loose file directly in .worktrees" allow \
  "$(run "$ROOT" Write "{\"file_path\":$(q "$MAIN/.worktrees/loose.txt"),\"content\":\"x\"}")"
expect "32 cd state threads across lines without a separator" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $MAIN
git reset HEAD~1")}")"
expect "33 longest-name expansion wins the collision" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "CA=$MAIN; C=/tmp; echo x > \$CA/f.txt")}")"
expect "34 braced expansion resolves the variable" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "C=$MAIN; echo x > \${C}/f.txt")}")"
expect "35 rm of the .worktrees root itself stays denied" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "rm -rf $MAIN/.worktrees")}")"
expect "36 unset near-name variable is not expanded greedily" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "D=$MAIN/docs; echo x > \$Dy/f.txt")}")"
expect "37 use before set is not expanded" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "echo x > \$C/f.txt
C=$MAIN")}")"

# 38-41: a landing push must not carry its own change folder open
mkdir -p "$WT/spectre/changes/change"
echo x >"$WT/spectre/changes/change/proposal.md"
git_q -C "$WT" add spectre
git_q -C "$WT" commit -q -m plan
expect "38 HEAD:main push with the change open" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $WT push origin HEAD:main")}")"
expect "39 change:main push with the change open" deny \
  "$(run "$WT" Bash "{\"command\":\"git push origin change:main\"}")"
expect "41 branch push with the change open" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $WT push origin change")}")"
git_q -C "$MAIN" worktree add "$MAIN/.worktrees/flowchg" -b spectre/flowchg
WT2="$MAIN/.worktrees/flowchg"
mkdir -p "$WT2/spectre/changes/flowchg"
echo x >"$WT2/spectre/changes/flowchg/proposal.md"
git_q -C "$WT2" add spectre
git_q -C "$WT2" commit -q -m plan
expect "42 spectre/<name>:main push with the change open" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $WT2 push origin spectre/flowchg:main")}")"
mkdir -p "$WT/spectre/changes/archive"
git_q -C "$WT" mv spectre/changes/change spectre/changes/archive/change
git_q -C "$WT" commit -q -m archive
expect "40 HEAD:main push with the change archived" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $WT push origin HEAD:main")}")"

echo x >"$ROOT/outside.txt"
expect "43 BSD sed -i with an empty suffix, file outside" allow \
  "$(run "$MAIN" Bash "{\"command\":$(q "sed -i '' 's/x/y/' $ROOT/outside.txt")}")"

# 44-47: the sibling layout, <dirname main>/<basename main>-worktrees/<change>
git_q -C "$MAIN" worktree add "$ROOT/main-worktrees/sib" -b sib
expect "44 sibling layout: Edit in an existing sibling worktree" allow \
  "$(run "$ROOT" Edit "{\"file_path\":$(q "$ROOT/main-worktrees/sib/f.txt"),\"old_string\":\"a\",\"new_string\":\"b\"}")"
git_q init -q -b develop "$ROOT/noorigin/inner"
mkdir -p "$ROOT/noorigin/inner-worktrees"
expect "45 sibling layout: Write into a not-yet-existing sibling worktree" allow \
  "$(run "$ROOT" Write "{\"file_path\":$(q "$ROOT/noorigin/inner-worktrees/new/x"),\"content\":\"x\"}")"
mkdir -p "$ROOT/noorigin/lib" "$ROOT/noorigin/lib-worktrees"
expect "47 sibling layout: <X>-worktrees beside a plain directory is main-checkout content" deny \
  "$(run "$ROOT" Write "{\"file_path\":$(q "$ROOT/noorigin/lib-worktrees/new/x"),\"content\":\"x\"}")"
reason="$(printf '{"tool_name":"Write","tool_input":{"file_path":%s,"content":"x"},"cwd":"%s"}' \
  "$(q "$MAIN/new.txt")" "$ROOT" | python3 "$HOOK")"
if grep -qF "worktree add $ROOT/main-worktrees/<change>" <<<"$reason" && ! grep -qF '.worktrees' <<<"$reason"; then
  pass "46 sibling layout: deny reason suggests the sibling path"
else
  fail "46 sibling layout: deny reason does not suggest $ROOT/main-worktrees/<change>: $reason"
fi

SIB="$ROOT/main-worktrees/sib"
expect "48 heredoc append into a sibling worktree, body naming rm and >" allow \
  "$(run "$MAIN" Bash "{\"command\":$(q "cat >> $SIB/f.txt <<'EOF'
then rm the stale file and cp it > there
EOF")}")"
expect "49 rm of an absolute sibling-worktree path from main" allow \
  "$(run "$MAIN" Bash "{\"command\":$(q "rm $SIB/tmp")}")"
expect "50 relative rm with cwd main" deny "$(run "$MAIN" Bash "{\"command\":\"rm f.txt\"}")"
expect "51 heredoc write into main" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $MAIN/f.txt <<EOF
x
EOF")}")"
expect "52 a line after the heredoc is scanned again" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q "cat >> $SIB/f.txt <<-EOF
	x
	EOF
rm f.txt")}")"
expect "53 heredoc fed to a shell stays scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "bash <<'EOF'
rm $MAIN/f.txt
EOF")}")"

expect "54 a quoted << opens no heredoc" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q 'echo "<<EOF"
rm f.txt
EOF')}")"
expect "55 a commented << opens no heredoc" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q 'true # <<EOF
rm f.txt
EOF')}")"
expect "56 an arithmetic shift opens no heredoc" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q 'echo $((1<<N))
rm f.txt')}")"
expect "57 a partly quoted delimiter is unquoted whole" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q 'cat <<E"OF" >/dev/null
x
EOF
rm f.txt')}")"
expect "58 a shell named by path keeps its body scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q '/bin/bash <<EOF
rm f.txt
EOF')}")"
expect "59 a body sourced by . stays scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q '. /dev/stdin <<EOF
rm f.txt
EOF')}")"
expect "60 dash fed a heredoc keeps its body scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q 'dash <<EOF
rm f.txt
EOF')}")"
expect "61 a heredoc piped into a shell keeps its body scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q 'cat <<EOF | dash
rm f.txt
EOF')}")"
expect "62 \$SHELL fed a heredoc keeps its body scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q '$SHELL <<EOF
rm f.txt
EOF')}")"
expect "63 a heredoc captured then eval'd keeps its body scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q 'x=$(cat <<EOF
rm f.txt
EOF
)
eval "$x"')}")"
expect "64 a heredoc read then eval'd keeps its body scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q "read -r -d '' c <<EOF
rm f.txt
EOF
eval \"\$c\"")}")"
expect "65 a heredoc written to a script then run keeps its body scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q "cat <<EOF > $SIB/s.sh
rm f.txt
EOF
sh $SIB/s.sh")}")"
expect "66 a CRLF closing line does not end the scan" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q "cat >> $SIB/f.txt <<EOF
x
EOF"$'\r'"
rm f.txt")}")"
expect "67 a << inside a parameter expansion opens no heredoc" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q 'echo ${x:-<<EOF}
rm f.txt
EOF')}")"
expect "68 tee -a heredoc into a sibling worktree, body naming rm" allow \
  "$(run "$MAIN" Bash "{\"command\":$(q "tee -a $SIB/f.txt <<'EOF'
then rm the stale file
EOF")}")"
expect "69 heredoc append into main" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat >> $MAIN/f.txt <<EOF
x
EOF")}")"
expect "70 an unquoted heredoc body's command substitution stays scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q "cat >> $SIB/f.txt <<EOF
\$(rm f.txt)
EOF")}")"

# 3 last: moving the main checkout off main lifts the protection
git_q -C "$MAIN" checkout -q -b feature
expect "3 Write in main checkout on a feature branch" allow \
  "$(run "$ROOT" Write "{\"file_path\":$(q "$MAIN/new.txt"),\"content\":\"x\"}")"

if [ "$FAILURES" -gt 0 ]; then
  echo "$FAILURES failure(s)" >&2
  exit 1
fi
echo "all protect-main-checkout cases pass"
