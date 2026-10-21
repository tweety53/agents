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
#   71. `D=<sibling>; cat > $D/r.md <<'EOF'`, body line `a > the`, then
#       a second command, cwd = main                                  -> allow (kan-929: a report body is data)
#   72. `D=<main>; cat > $D/r.md <<'EOF'`, then a second command      -> deny (the target is still judged)
#   73. `cat > >(sh) <<'EOF'`, body `rm <main>/f.txt`, in a longer command -> deny (the target runs the body)
#   74. `tee /dev/null >(bash) <<'EOF'`, body `rm <main>/f.txt`         -> deny (a process substitution runs it)
#   75. `bash <(` then `cat > /dev/stdout <<'EOF'`, body `rm <main>/f.txt` -> deny (the enclosing command runs it)
#   76. `cat > <sibling>/x.sh <<'EOF'`, body `rm <main>/f.txt`, then
#       `sh "<sibling>"/x.sh`                                           -> deny (written, then run under another spelling)
#   77. Bash `git -C <loop wt> push origin pipeline-fix-x-6:main`, the
#       loop checker refusing                                         -> deny, its lines in the reason
#   78. the same push, the loop checker admitting                     -> allow, checker handed <wt> <branch>
#   79. Bash `git -C <self-review wt> push origin HEAD:main`, refusing -> deny
#   80. the loop push with no runnable loop checker                   -> deny (fails closed)
#   81. Bash `git -C <wt> push origin change:main`, archived           -> allow, loop checker never run
#   82-87. the loop landed under another spelling, the checker refusing -> deny each:
#       `@:main`, `<sha>:main`, `HEAD~0:main`, `git -c x=y push … :main`,
#       `git branch tmp && git push origin tmp:main`, and from the main
#       checkout once `pull --ff-only . <loop>` brought the loop in, `push origin main`
#   88. Bash `git -C <main> push origin spectre/chg:main`, a change branch
#       the checker refuses (self-review rows left open)              -> deny
#   89. two such pushes in one command, each check inside a per-push budget
#       but together over the command's one budget                    -> deny
#   90-93. a landing push the hook cannot read, every existing check
#       admitting it: `git -C $W push origin $B:main`, `"${B}:main"`,
#       `$(git branch --show-current):main`, a backtick refspec       -> deny each
#   94. `cd $W && git push origin HEAD` (repository and destination unknown) -> deny
#   95. `git -C $W push -u origin $B` (destination unknown)          -> deny
#   96. `git -C $W push origin spectre/x` (a literal non-default destination) -> allow
#   97. `B=change; git -C <wt> push origin $B` (assigned in the command) -> allow
#   98. `B=x git -C <wt> push origin $B:main` (prefix assignment)   -> deny
#   99. `git -C <wt> push --all $R` (--all pushes the default branch) -> deny
#  100. `git ${OPTS} push origin HEAD` (expanded git option)         -> deny
#  101. `$GIT push origin HEAD` (expanded git command word)          -> deny
#  102. Bash `cat > <link into main/sub>/../f.txt` from elsewhere    -> deny (kan-963: judged after symlinks resolve)
#  103. Bash `rm <link into main>` / `mv <link into main> ..`       -> allow (they act on the link itself)
#  110. `cat > <sibling>/r.md <<EOF`, body ``see `a` > b``, cwd = main -> allow (only a substitution runs)
#  111. `cat > <sibling>/r.md <<EOF`, body ``see `rm <main>/x` ``      -> deny (the backtick span still runs)
#  112. `echo `rm <main>/x`` from elsewhere                          -> deny (a backtick span is a command)
#  113. `echo "$(rm <main>/x)"` from elsewhere                        -> deny (a quoted substitution still runs)
#  114. `cat > <sibling>/r.md <<EOF` from elsewhere, body ``see `cp a <main>/y` `` -> deny (cp's walk stops at `)`)
#  115. `echo $(cp a <main>/y)` from elsewhere                        -> deny
#  116. `(cp a <main>/y)` from elsewhere                              -> deny
#  117. `cat > <sibling>/r.md <<EOF`, body ``\` `rm <main>/x` \` ``   -> deny (an escaped backtick is a literal)
#  118. `C=<sibling>; cat > $C/t.md <<'EOF'`, body `a > b; rm x`, then
#       `check-plan-shape.sh $C/t.md`, cwd = main                    -> allow (the target as an argument is data)
#  119. the same body, then `~/.claude/.../check-plan-shape.sh /x/o.md` -> allow (a path-named command naming another file)
#  120. the same body, then `chmod +x $C/t.md && $C/t.md`            -> deny (the target run as a command)
#  121. `cat > <sibling>/x.sh <<'EOF'`, body `rm <main>/f.txt`, then
#       `/elsewhere/x.sh`                                             -> deny (the same file name under another path)
#  122-128. the same heredoc, then `/bin/bash <sib>/x.sh`, `/bin/sh <sib>/x.sh`,
#       `cd <sib> && x.sh`, `awk -f`, `make -f`, `watch`, `crontab` on it -> deny (each runs the body)
#  129-134. the same heredoc, then `php`, `setsid`, `find -exec`, `ssh`, `psql -f`, `git -c ... show`
#       naming it                                                     -> deny (not a known reader)
#  135. the same heredoc, then `rg --pre <sib>/x.sh x f`              -> deny (rg's --pre runs the target)
#  136. `cat > <sib>/check-a.sh <<EOF`, body `rm <main>/f.txt`, then `<sib>/check-a.sh` -> deny (a written check script run)
#  137. the x.sh heredoc, then `GIT_EXTERNAL_DIFF=<sib>/x.sh git diff --ext-diff a b` -> deny (an assignment can make a reader run the target)
#  143. `cd <wt> && git commit -qm '<no trailer>' && git push -q`  -> deny (kan-1003: a commit pushed in the same command)
#  144. the same with `-m 'Flow-Key: <key>'`                         -> allow
#  145. `git -C <wt> commit -m ... -- spectre/changes/chg && push`   -> allow (a planning commit)
#  146. the same shape, a `chore(...)` subject                       -> allow
#  147. `git commit -F msg.txt && git push`                          -> allow (a message the command does not state)
#  148. a trailer-less commit pushed to `HEAD:feature/x`             -> allow (no change branch)
#  150. the 145 planning commit with its pathspec given without `--` -> allow
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
expect "71 a report heredoc in a longer command keeps its body as data" allow \
  "$(run "$MAIN" Bash "{\"command\":$(q "D=$SIB; cat > \$D/r.md <<'EOF'
the row reads > the old one
EOF
echo written")}")"
expect "72 a report heredoc in a longer command still judges its target" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "D=$MAIN; cat > \$D/r.md <<'EOF'
x
EOF
echo written")}")"
expect "73 a heredoc into a process substitution keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "echo start; cat > >(sh) <<'EOF'
rm $MAIN/f.txt
EOF
echo done")}")"
expect "74 a tee heredoc feeding a process substitution keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "echo start; tee /dev/null >(bash) <<'EOF'
rm $MAIN/f.txt
EOF
echo done")}")"
expect "75 a heredoc inside bash <( keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "bash <(
cat > /dev/stdout <<'EOF'
rm $MAIN/f.txt
EOF
)")}")"
expect "76 a heredoc written then run under another spelling keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
sh \"$SIB\"/x.sh")}")"

# 77-81: a fix loop lands only once check-loop-landing.sh admits it; the
# checker is a stub here (the guard is covered by check_loop_landing_test.go).
LOOP_LOG="$ROOT/loop-args"
LOOP_RC="$ROOT/loop-rc"
cat >"$ROOT/loopcheck" <<STUB
#!/usr/bin/env bash
printf '%s\n' "\$@" >"$LOOP_LOG"
echo "LOOP-LANDING-REFUSED: REVIEW — stub"
exit "\$(cat "$LOOP_RC")"
STUB
chmod +x "$ROOT/loopcheck"
export LOOP_LANDING_CHECK="$ROOT/loopcheck"
git_q -C "$MAIN" worktree add "$ROOT/main-worktrees/pipeline-fix-x-6" -b pipeline-fix-x-6 origin/main
git_q -C "$MAIN" worktree add "$ROOT/main-worktrees/self-review-x" -b self-review-x origin/main
LWT="$ROOT/main-worktrees/pipeline-fix-x-6"
for w in "$LWT" "$ROOT/main-worktrees/self-review-x"; do
  echo "$w" >"$w/loop.txt"
  git_q -C "$w" add loop.txt
  git_q -C "$w" -c user.name=t -c user.email=t@t commit -q -m "fix: loop"
done
echo 1 >"$LOOP_RC"
out="$(printf '{"tool_name":"Bash","tool_input":{"command":%s},"cwd":"%s"}' "$(q "git -C $LWT push origin pipeline-fix-x-6:main")" "$ROOT" | python3 "$HOOK")"
if grep -q '"permissionDecision": *"deny"' <<<"$out" && grep -q "LOOP-LANDING-REFUSED: REVIEW" <<<"$out"; then
  pass "77 loop push the checker refuses"
else
  fail "77 loop push the checker refuses: $out"
fi
echo 0 >"$LOOP_RC"
expect "78 loop push the checker admits" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $LWT push origin pipeline-fix-x-6:main")}")"
[ "$(tr '\n' ' ' <"$LOOP_LOG")" = "$LWT pipeline-fix-x-6 " ] \
  && pass "78 checker args" || fail "78 checker args: $(tr '\n' ' ' <"$LOOP_LOG")"
echo 1 >"$LOOP_RC"
expect "79 self-review HEAD:main push the checker refuses" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $ROOT/main-worktrees/self-review-x push origin HEAD:main")}")"
expect "80 loop push with no runnable checker" deny \
  "$(LOOP_LANDING_CHECK="$ROOT/missing" run "$ROOT" Bash "{\"command\":$(q "git -C $LWT push origin pipeline-fix-x-6:main")}")"
rm -f "$LOOP_LOG"
expect "81 a non-loop landing" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $WT push origin change:main")}")"
[ ! -e "$LOOP_LOG" ] && pass "81 loop checker not run" || fail "81 loop checker ran"
echo 1 >"$LOOP_RC"
LSHA="$(git -C "$LWT" rev-parse HEAD)"
n=82
for c in "git -C $LWT push origin @:main" "git -C $LWT push origin $LSHA:main" \
  "git -C $LWT push origin HEAD~0:main" "git -C $LWT -c x=y push origin pipeline-fix-x-6:main" \
  "git -C $LWT branch tmp && git -C $LWT push origin tmp:main"; do
  expect "$n loop landed as: $c" deny "$(run "$ROOT" Bash "{\"command\":$(q "$c")}")"
  n=$((n + 1))
done
git_q -C "$MAIN" merge -q --ff-only pipeline-fix-x-6
expect "87 a colon-less push of the default branch the loop was pulled into" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $MAIN push origin main")}")"
git_q -C "$MAIN" reset -q --hard origin/main
git_q -C "$MAIN" branch -q spectre/chg "$LSHA"
git_q -C "$MAIN" update-ref refs/heads/pipeline-fix-x-6 origin/main
rm -f "$LOOP_LOG"
expect "88 a change branch the checker refuses" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $LWT push origin spectre/chg:main")}")"
grep -qx "spectre/chg" "$LOOP_LOG" && pass "88 checker handed the change branch" || fail "88 checker args: $(cat "$LOOP_LOG" 2>/dev/null)"
printf '#!/usr/bin/env bash\nsleep 1.5\n' >"$ROOT/slowcheck"
chmod +x "$ROOT/slowcheck"
expect "89 two landing pushes share one checker budget" deny \
  "$(LOOP_LANDING_CHECK="$ROOT/slowcheck" LOOP_LANDING_TIMEOUT=2 run "$ROOT" Bash "{\"command\":$(q "git -C $LWT push origin spectre/chg:main; git -C $LWT push origin spectre/chg:main")}")"
# 90-97: a push named through an unexpanded shell expansion cannot be checked;
# the stub admits, so only the unreadable-landing check can deny.
echo 0 >"$LOOP_RC"
n=90
for c in "git -C \$W push origin \$B:main" "git -C $WT push origin \"\${B}:main\"" \
  "git -C $WT push origin \$(git branch --show-current):main" \
  "git -C $WT push origin \`git branch --show-current\`:main"; do
  expect "$n unreadable landing: $c" deny "$(run "$ROOT" Bash "{\"command\":$(q "$c")}")"
  n=$((n + 1))
done
expect "94 cd to an unexpanded path, then push HEAD" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q 'cd $W && git push origin HEAD')}")"
expect "95 push of an unexpanded branch" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q 'git -C $W push -u origin $B')}")"
expect "96 unexpanded repository, literal non-default destination" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q 'git -C $W push origin spectre/x')}")"
expect "97 a branch assigned in the command itself" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "B=change; git -C $WT push origin \$B")}")"
expect "98 a prefix assignment is not seen by its own command's words" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "B=x git -C $WT push origin \$B:main")}")"
expect "99 --all with an unexpanded remote pushes the default branch too" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $WT push --all \$R")}")"
expect "100 an expanded git option hides the repository" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git \${OPTS} push origin HEAD")}")"
expect "101 an expanded git command word" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "\$GIT push origin HEAD")}")"
unset LOOP_LANDING_CHECK
mkdir -p "$MAIN/sub"
ln -s "$MAIN/sub" "$ROOT/elsewhere/lsub"
expect "102 a symlink into the checkout, then .., resolves before judging" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $ROOT/elsewhere/lsub/../f.txt")}")"
ln -s "$MAIN/sub" "$ROOT/elsewhere/lsub2"
expect "103 rm and mv of a symlink into the checkout act on the link, outside it" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "rm $ROOT/elsewhere/lsub2 && mv $ROOT/elsewhere/lsub $ROOT/elsewhere/l3")}")"
expect "110 an unquoted data heredoc's literal redirect beside a backtick span is data" allow \
  "$(run "$MAIN" Bash "{\"command\":$(q "cat > $SIB/r.md <<EOF
see \`a\` > b
EOF")}")"
expect "111 a backtick span in an unquoted data heredoc body is still scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q "cat > $SIB/r.md <<EOF
see \`rm $MAIN/x\` here
EOF")}")"
expect "112 a backtick span is scanned as a command" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "echo \`rm $MAIN/x\`")}")"
expect "113 a command substitution inside a double-quoted word is scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "echo \"\$(rm $MAIN/x)\"")}")"
expect "114 a backtick cp in an unquoted data heredoc body is scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/r.md <<EOF
see \`cp a $MAIN/y\` here
EOF")}")"
expect "115 a cp inside an unquoted \$(...) is scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "echo \$(cp a $MAIN/y)")}")"
expect "116 a cp inside a subshell is scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "(cp a $MAIN/y)")}")"
expect "117 an escaped backtick does not pair with a real span" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q "cat > $SIB/r.md <<EOF
\\\` \`rm $MAIN/x\` \\\`
EOF")}")"
# Case 118: the heredoc target named again as an argument is data.
expect "118 a data heredoc's target named again as an argument keeps its body as data" allow \
  "$(run "$MAIN" Bash "{\"command\":$(q "C=$SIB; cat > \$C/t.md <<'EOF'
> **Execution:** a > b; rm x
EOF
check-plan-shape.sh \$C/t.md")}")"
# Case 119: a path-named command naming another file runs that file, not the body.
expect "119 a path-named command naming another file keeps the body as data" allow \
  "$(run "$MAIN" Bash "{\"command\":$(q "C=$SIB; cat > \$C/t.md <<'EOF'
> **Execution:** a > b; rm x
EOF
~/.claude/skills/flow/scripts/check-plan-shape.sh /x/o.md")}")"
# Case 120: the heredoc target run as a command keeps its body scanned.
expect "120 a data heredoc's target run as a command keeps its body scanned" deny \
  "$(run "$MAIN" Bash "{\"command\":$(q "C=$SIB; cat > \$C/t.md <<'EOF'
> **Execution:** a > b; rm x
EOF
chmod +x \$C/t.md && \$C/t.md")}")"
# Case 121: the same file name run under another path keeps its body scanned.
expect "121 a heredoc's file name run under another path keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
/elsewhere/x.sh")}")"
# Cases 122-128: a runner spelled as a path, a bare target name, or a file-executing tool keeps the body scanned.
expect "122 a heredoc body then '/bin/bash $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
/bin/bash $SIB/x.sh")}")"
expect "123 a heredoc body then '/bin/sh $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
/bin/sh $SIB/x.sh")}")"
expect "124 a heredoc body then 'cd $SIB && x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
cd $SIB && x.sh")}")"
expect "125 a heredoc body then 'awk -f $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
awk -f $SIB/x.sh")}")"
expect "126 a heredoc body then 'make -f $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
make -f $SIB/x.sh")}")"
expect "127 a heredoc body then 'watch $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
watch $SIB/x.sh")}")"
expect "128 a heredoc body then 'crontab $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
crontab $SIB/x.sh")}")"
# Cases 129-134: the target named by a command not known to only read it keeps the body scanned.
expect "129 a heredoc body then 'php $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
php $SIB/x.sh")}")"
expect "130 a heredoc body then 'setsid $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
setsid $SIB/x.sh")}")"
expect "131 a heredoc body then 'find $SIB -name x.sh -exec {} ;' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
find $SIB -name x.sh -exec {} ;")}")"
expect "132 a heredoc body then 'ssh localhost $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
ssh localhost $SIB/x.sh")}")"
expect "133 a heredoc body then 'psql -f $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
psql -f $SIB/x.sh")}")"
expect "134 a heredoc body then 'git -c core.pager=x show $SIB/x.sh' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
git -c core.pager=x show $SIB/x.sh")}")"
expect "135 a heredoc body then 'rg --pre $SIB/x.sh x f' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
rg --pre $SIB/x.sh x f")}")"
expect "136 a written check script run afterwards keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/check-a.sh <<EOF
rm $MAIN/f.txt
EOF
$SIB/check-a.sh")}")"
expect "137 a heredoc body then 'GIT_EXTERNAL_DIFF=$SIB/x.sh git diff' keeps its body scanned" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $SIB/x.sh <<'EOF'
rm $MAIN/f.txt
EOF
GIT_EXTERNAL_DIFF=$SIB/x.sh git diff --ext-diff a b")}")"

# 104-107: a push of a change branch runs check-untasked-commits over its
# unpushed commits (kan-941); a stub stands in for the guard.
UNTASKED_RC="$ROOT/untasked-rc"
UNTASKED_LOG="$ROOT/untasked-log"
cat >"$ROOT/untaskedcheck" <<STUB
#!/usr/bin/env bash
printf '%s\n' "\$@" >"$UNTASKED_LOG"
echo "UNTASKED: 1 commit(s) carry no Task-Id: or Flow-Key: trailer"
exit "\$(cat "$UNTASKED_RC")"
STUB
chmod +x "$ROOT/untaskedcheck"
export UNTASKED_COMMITS_CHECK="$ROOT/untaskedcheck"
CWT="$ROOT/main-worktrees/chg2"
git_q -C "$MAIN" worktree add "$CWT" -b spectre/chg2 origin/main
git_q -C "$CWT" push -q -u origin spectre/chg2
echo m >"$CWT/m.txt"
git_q -C "$CWT" add m.txt
git_q -C "$CWT" commit -q -m "fix(app): review Minors"
echo 1 >"$UNTASKED_RC"
expect "104 a change-branch push carrying an untasked commit" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $CWT push origin spectre/chg2")}")"
[ "$(tr '\n' ' ' <"$UNTASKED_LOG")" = "$CWT refs/remotes/origin/spectre/chg2 --not-remotes " ] \
  && pass "104 checker args" || fail "104 checker args: $(tr '\n' ' ' <"$UNTASKED_LOG")"
echo 0 >"$UNTASKED_RC"
expect "105 a change-branch push the checker clears" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git push")}")"
echo 1 >"$UNTASKED_RC"
rm -f "$UNTASKED_LOG"
git_q -C "$CWT" checkout -q -b spectre/chg3
expect "106 a change branch origin does not hold yet" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $CWT push -u origin spectre/chg3")}")"
[ ! -e "$UNTASKED_LOG" ] && pass "106 checker not run" || fail "106 checker ran"
expect "107 a checker that cannot run fails open" allow \
  "$(UNTASKED_COMMITS_CHECK="$ROOT/missing" run "$ROOT" Bash "{\"command\":$(q "git -C $CWT push origin HEAD:spectre/chg2")}")"
unset UNTASKED_COMMITS_CHECK

# 143-148: a commit and its push in one command (kan-1003's self-review: the design handoff
# 0497d943 was committed and pushed in one call, before the push check could see it).
echo h >"$CWT/h.txt"
expect "143 commit and push in one command, no trailer" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git add h.txt && git commit -qm 'docs(design): add the handoff' && git push -q")}")"
expect "144 the same with a Flow-Key trailer" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git commit -q -m 'docs(design): add the handoff' -m 'Flow-Key: flow.brainstorm' && git push -q")}")"
expect "145 a planning commit pathspec-limited to spectre/changes/" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $CWT commit -q -m 'docs(spectre): record' -- spectre/changes/chg && git -C $CWT push -q origin spectre/chg3")}")"
expect "150 a planning commit's pathspec given without '--'" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $CWT commit -q -m 'docs(spectre): record' spectre/changes/chg && git -C $CWT push -q origin spectre/chg3")}")"
expect "146 a chore subject" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git commit -q -m 'chore(spectre): plan chg — plan gate' && git push -q")}")"
expect "147 a message the command does not state" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git commit -q -F msg.txt && git push -q")}")"
expect "148 a trailer-less commit pushed to a branch that is no change branch" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git commit -qm 'docs(x): y' && git push -q origin HEAD:feature/x")}")"

# 108-109: the real checker after a rebase onto a moved base — the base's own
# trailer-less commits are on origin already, never the change's to judge.
RWT="$ROOT/main-worktrees/chg4"
git_q -C "$MAIN" worktree add "$RWT" -b spectre/chg4 origin/main
git_q -C "$RWT" push -q -u origin spectre/chg4
echo b >"$RWT/b.txt"
git_q -C "$RWT" add b.txt
git_q -C "$RWT" commit -q -m "fix(build): someone else's work"
git_q -C "$RWT" push -q origin HEAD:refs/heads/main
echo t >"$RWT/t.txt"
git_q -C "$RWT" add t.txt
git_q -C "$RWT" commit -q -m "feat(app): task one" -m "Task-Id: 1"
expect "108 a push after a rebase onto a moved base" allow \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $RWT push --force-with-lease origin spectre/chg4")}")"
echo u >"$RWT/u.txt"
git_q -C "$RWT" add u.txt
git_q -C "$RWT" commit -q -m "fix(app): review Minors"
expect "109 the real checker still lists the change's own untasked commit" deny \
  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $RWT push --force-with-lease origin spectre/chg4")}")"

# 3 last: moving the main checkout off main lifts the protection
git_q -C "$MAIN" checkout -q -b feature
expect "3 Write in main checkout on a feature branch" allow \
  "$(run "$ROOT" Write "{\"file_path\":$(q "$MAIN/new.txt"),\"content\":\"x\"}")"

if [ "$FAILURES" -gt 0 ]; then
  echo "$FAILURES failure(s)" >&2
  exit 1
fi
echo "all protect-main-checkout cases pass"
