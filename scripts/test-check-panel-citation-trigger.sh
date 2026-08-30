#!/usr/bin/env bash
# Assertion harness for check-panel-citation-trigger.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's exit code.
# Never touches the real repository tree. Same shape as
# test-check-base-moved.sh: an indexed REPOS array (bash 3.2 has no
# associative arrays), removed by an EXIT trap, real git repositories rather
# than fixture trees.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the contract
# stated in tasks.md's task 1, never against observed output. The guard
# prints nothing — the exit code is the whole answer.
#
# PROVED BY MUTATION, NOT MERELY ASSERTED. Case 2 (a committed .mdc change)
# is re-run after replacing the guard's `.mdc?$` pattern with a pattern that
# never matches, confirming the case flips from pass to fail, then the
# guard is restored from git and re-verified clean.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-panel-citation-trigger.sh"
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

# run_guard <worktree> <merge-base> -> sets RC
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" "$2" 2>&1)"
  RC=$?
  set -e
}

# new_repo -> sets REPO, MERGEBASE
# A repository on `main` with one commit carrying base.go and tracked.md —
# the merge base every case starts from. tracked.md is tracked from the
# start (mirroring test-check-base-moved.sh's shared.txt) so an
# unstaged-only edit is a modification `git diff` detects without an
# intervening `git add` — a brand-new untracked file would not appear in
# either the staged or unstaged diff at all.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/citation-trigger-test.XXXXXX")"
  REPOS+=("$REPO")
  git -C "$REPO" init -q -b main
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  echo base > "$REPO/base.go"
  echo "# tracked" > "$REPO/tracked.md"
  git -C "$REPO" add base.go tracked.md
  git -C "$REPO" commit -qm "base"
  MERGEBASE="$(git -C "$REPO" rev-parse HEAD)"
}

# 1. A committed .md change since the merge base -> exit 0.
new_repo
echo "# notes" > "$REPO/notes.md"
git -C "$REPO" add notes.md
git -C "$REPO" commit -qm "add notes.md"
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 0 ] && pass "committed .md change -> exit 0" \
  || fail "committed .md change: expected exit 0, got rc=$RC out=$OUT"

# 2. A committed .mdc change since the merge base -> exit 0.
new_repo
echo "rule" > "$REPO/rule.mdc"
git -C "$REPO" add rule.mdc
git -C "$REPO" commit -qm "add rule.mdc"
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 0 ] && pass "committed .mdc change -> exit 0" \
  || fail "committed .mdc change: expected exit 0, got rc=$RC out=$OUT"

# 3. Only non-Markdown committed changes -> exit 1.
new_repo
echo "package main" > "$REPO/main.go"
git -C "$REPO" add main.go
git -C "$REPO" commit -qm "add main.go"
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 1 ] && pass "only non-Markdown committed changes -> exit 1" \
  || fail "non-Markdown committed changes: expected exit 1, got rc=$RC out=$OUT"

# 4. A staged-only .md change, nothing committed -> exit 0.
new_repo
echo "# staged" > "$REPO/staged.md"
git -C "$REPO" add staged.md
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 0 ] && pass "staged-only .md change -> exit 0" \
  || fail "staged-only .md change: expected exit 0, got rc=$RC out=$OUT"

# 5. An unstaged-only .md change, nothing committed or staged -> exit 0.
# tracked.md is tracked from the merge-base commit (new_repo), so modifying
# it without `git add` is a genuine unstaged change `git diff` detects — a
# brand-new untracked file would not appear in git diff at all.
new_repo
echo "unstaged edit" >> "$REPO/tracked.md"
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 0 ] && pass "unstaged-only .md change -> exit 0" \
  || fail "unstaged-only .md change: expected exit 0, got rc=$RC out=$OUT"

# 6. No changes at all (worktree at the merge base, clean) -> exit 1.
new_repo
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 1 ] && pass "no changes at all -> exit 1" \
  || fail "no changes at all: expected exit 1, got rc=$RC out=$OUT"

# 7. Missing arguments -> exit 2.
run_guard "" ""
[ "$RC" -eq 2 ] && pass "missing arguments -> exit 2" \
  || fail "missing arguments: expected exit 2, got rc=$RC out=$OUT"

# 8. <worktree> not a directory -> exit 2.
NOT_A_DIR="$(mktemp -u "${TMPDIR:-/tmp}/citation-trigger-test-nodir.XXXXXX")"
run_guard "$NOT_A_DIR" deadbeef
[ "$RC" -eq 2 ] && pass "worktree not a directory -> exit 2" \
  || fail "worktree not a directory: expected exit 2, got rc=$RC out=$OUT"

# 9. <worktree> not a git repository -> exit 2.
NOT_A_REPO="$(mktemp -d "${TMPDIR:-/tmp}/citation-trigger-test-notrepo.XXXXXX")"
REPOS+=("$NOT_A_REPO")
run_guard "$NOT_A_REPO" deadbeef
[ "$RC" -eq 2 ] && pass "worktree not a git repository -> exit 2" \
  || fail "worktree not a git repository: expected exit 2, got rc=$RC out=$OUT"

# 10. <merge-base> not resolving in the worktree -> exit 2.
new_repo
run_guard "$REPO" 0000000000000000000000000000000000000000
[ "$RC" -eq 2 ] && pass "merge base not resolving -> exit 2" \
  || fail "merge base not resolving: expected exit 2, got rc=$RC out=$OUT"

# 11. F2/F3 (KAN-304 review round 1): a merge-base value shaped like a git
#     option (starts with `-`) must never be misread as a flag by either
#     `--end-of-options`-protected call site — it is always read as a
#     revision, correctly fails to resolve, and reports "does not resolve"
#     via exit 2, exactly like case 10's unresolving SHA.
new_repo
run_guard "$REPO" "--evil"
[ "$RC" -eq 2 ] && pass "flag-shaped merge base ('--evil') -> exit 2" \
  || fail "flag-shaped merge base: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"does not resolve"*) pass "flag-shaped merge base: reported as not resolving, not a git option error" ;;
  *) fail "flag-shaped merge base: expected a 'does not resolve' message, got: $OUT" ;;
esac

# 11b. F8 (round 4) tried to prove the guard cannot be shadowed by
#      enumerating shadow SYNTAX with a regex — `git() {`, `function git`,
#      `alias git=`. Round 6 found this unbounded the same way rounds 1-3's
#      source-text search for `--end-of-options` was: F11 defeated it with
#      a space inside the parens (`git ( )`), F13 with the definition
#      placed after a leading `;` on the same line — each closing one
#      spelling while the underlying property (does ANY spelling of a
#      shadow intercept the guard's real calls) stayed unproven. F12 found
#      a second, unrelated hole in the same case: xtrace shows the command
#      AS PARSED from source, not what the name actually resolved to, so a
#      scoped `PATH=` swap around one call defeats the trace assertion
#      with no function or alias involved at all.
#
#      Round 6 closes the whole class structurally instead: the guard
#      resolves GIT_BIN="$(type -P git)" as its first executable
#      statement (`type -P` ignores functions and aliases and performs a
#      real PATH search) and invokes "$GIT_BIN" at every call site,
#      never bare `git`. Invoking a command by absolute path never
#      triggers bash's function/alias lookup, so no shadow — in any
#      spacing, any placement, any syntax — can intercept a call that
#      never names the shadowed word. And because GIT_BIN is captured
#      before any later line in the file runs, a PATH swap anywhere
#      after it (F12) resolves against an already-captured absolute
#      path, not a fresh lookup.
#
#      (a) A direct structural check of the fix itself: the guard's real
#      call sites (`-C "$WORKTREE"`) all invoke "$GIT_BIN", never bare
#      `git`. This does not enumerate shadow spellings — after this
#      passes, there is nothing left in the file for a shadow to
#      intercept, which is what makes F11/F13's enumeration problem moot
#      rather than merely narrower.
#
#      Round 6, Bugbot's own re-verification of this fix: a NEGATIVE check
#      (forbid the shapes bare `git` can take) inherits the exact unbounded
#      enumeration problem it was written to end — `"git" -C` (git quoted)
#      slipped past the first version of this regex, since a literal `"`
#      immediately before `git` fell outside the excluded-character class.
#      Replaced with a POSITIVE check instead: every call-site line must
#      contain the literal substring `"$GIT_BIN" -C "$WORKTREE"` verbatim.
#      A positive assertion has no enumeration problem — there is exactly
#      one spelling of "this call uses GIT_BIN", not an open set of
#      spellings of "this call doesn't".
CALL_SITE_LINES="$(grep -n -- '-C[[:space:]]*"\$WORKTREE"' "$GUARD" || true)"
NON_GITBIN_CALLS=0
while IFS= read -r line; do
  [ -z "$line" ] && continue
  case "$line" in
    *'"$GIT_BIN" -C "$WORKTREE"'*) ;;
    *) NON_GITBIN_CALLS=$((NON_GITBIN_CALLS + 1)) ;;
  esac
done <<EOF
$CALL_SITE_LINES
EOF
[ "$NON_GITBIN_CALLS" -eq 0 ] \
  && pass "F8/F11/F13: every real call site is literally \"\$GIT_BIN\" -C \"\$WORKTREE\", never bare git in any spelling" \
  || fail "F8/F11/F13: a call site does not literally invoke \"\$GIT_BIN\" -C \"\$WORKTREE\" — a shadow or quoting trick could intercept it"

# (b) Mutation proof: append a hostile `git` shadow — in the exact
# space-in-parens shape F11 used — AFTER the GIT_BIN resolution line, to a
# throwaway copy of the guard, and confirm its real behaviour (via trace)
# is unaffected: the shadow is defined, but nothing ever calls it.
GIT_SHADOW_COPY="$(mktemp "${TMPDIR:-/tmp}/citation-trigger-guard-shadow.XXXXXX")"
REPOS+=("$GIT_SHADOW_COPY")
awk '/^WORKTREE=/ && !done { print "git ( ) { command git \"$@\" | sed \"s/--end-of-options //\"; }"; done=1 } 1' \
  "$GUARD" > "$GIT_SHADOW_COPY"
chmod +x "$GIT_SHADOW_COPY"
new_repo
echo "notes" > "$REPO/notes.md"
git -C "$REPO" add notes.md
git -C "$REPO" commit -qm "add notes.md"
SHADOW_XTRACE="$(bash -x "$GIT_SHADOW_COPY" "$REPO" "$MERGEBASE" 2>&1 >/dev/null || true)"
printf '%s\n' "$SHADOW_XTRACE" \
  | grep -E "^\\++ [^[:space:]]*/git .*rev-parse --verify --end-of-options .*${MERGEBASE}\\^\\{commit\\}" >/dev/null \
  && pass "F8/F11/F13 mutation proof: a git() shadow (space-in-parens, defined after GIT_BIN) never runs — the real call still carries the flag" \
  || fail "F8/F11/F13 mutation proof: the shadow intercepted a real call despite GIT_BIN"

# 12. F2/F3's mutant, closed definitively (round 4, F6). `--verify`'s own
#     parsing already reads its operand as a revision regardless of a
#     leading `-` (confirmed by direct experiment — case 11 above passes
#     identically with the flag removed from line 37), and line 37 gates
#     line 42's diff call, so no merge-base value can ever be both
#     dash-prefixed (dangerous to `diff`) and resolvable (past line 37) —
#     git forbids a ref or object name that begins with `-`
#     (git-check-ref-format(1); SHAs are hex-only). So this cannot be
#     proved by feeding the guard a hostile value; it has to be proved by
#     observing what the guard's own two calls actually pass to git.
#
#     Rounds 1-3 tried to prove that from the guard's SOURCE TEXT — a
#     grep for the literal string `--end-of-options` — and each round
#     closed one placement Bugbot found the string surviving in while
#     absent from the real call: a whole-line comment (round 1), a decoy
#     comment elsewhere in the file (round 2), a trailing inline comment
#     on the stripped call's own line (round 3). Round 4 found a fourth:
#     the string sitting in a dead, non-comment variable assignment,
#     which no comment-aware filter touches because it is not a comment.
#     No source-text search can ever close every place a substring could
#     be made to sit — that is the source of the recurring bypass, not
#     any one grep's phrasing.
#
#     So this asserts on RUNTIME, not source: `bash -x` traces every
#     command the guard executes, verbatim, to stderr, one `+`/`++`
#     -prefixed line per invocation carrying the real, expanded argv. A
#     trace line is anchored to `+`(s) then `git` — a dead-code
#     assignment traces too (`+ VAR='...'`), but never as a line
#     beginning `+ git`, so anchoring on that shape is what the source
#     text could never be. Verified against all four historical
#     bypasses (F2/F3's plain strip, F4's two comment placements, F5's
#     trailing comment, F6's dead-code assignment): the anchored trace
#     rejects the flag's absence in every one, because it observes what
#     git received, not where a string happens to sit in the file.
#
#     F9 (round 4, Bugbot): a `+ git`-anchored line still does not prove
#     the flag reached THIS test's own COMMITTED/rev-parse call — a
#     decoy `git ... --end-of-options` invocation added anywhere else in
#     the guard, on an unrelated path, also traces as a `+ git` line and
#     satisfies an unanchored match. Closed by requiring the SAME traced
#     line to also carry this run's own resolved `$MERGEBASE` value —
#     `${MERGEBASE}^{commit}` for the rev-parse call, `${MERGEBASE}..HEAD`
#     for the diff call — which only the real call ever emits, since
#     `new_repo` mints a fresh SHA no decoy could have been written
#     against in advance.
#
#     F10 (round 6, Primary): flag+MERGEBASE co-occurrence still does not
#     prove the traced line IS the gating call — a decoy placed inside its
#     own command substitution, at the same subshell depth, referencing
#     the shell variable `${MERGEBASE}` (not a hardcoded value), traces
#     identically to the real call and satisfies a depth-anchored match
#     too, while the real call sits stripped of the flag a few lines
#     away. Depth alone cannot distinguish "this trace line" from "the
#     genuine gating call" when both share it.
#
#     Closed by anchoring to the exact SOURCE LINE NUMBER of the real
#     call, not to its depth or its content. `PS4='+${LINENO}: '` makes
#     `bash -x` prefix every traced line with the line it came from — a
#     decoy on a different line traces a different number no matter what
#     content it carries. The real line numbers are located dynamically
#     in `$GUARD`'s own source, by the ONE structural marker a decoy
#     cannot fake without breaking this guard's actual function: the
#     literal assignment to `COMMITTED` (case 1-6 depend on it holding
#     the real committed diff) and the literal `if` gating on
#     `rev-parse --verify` (case 10/11 depend on it exiting 2 on a bad
#     merge base). `grep -c` on each requires there be exactly one such
#     line in the guard — an attacker cannot add a second real-looking
#     assignment/gate without it either breaking functional cases 1-11 or
#     failing this uniqueness check.
COMMITTED_ASSIGN_LINE="$(grep -n '^COMMITTED="\$(' "$GUARD" || true)"
[ "$(printf '%s\n' "$COMMITTED_ASSIGN_LINE" | grep -c .)" -eq 1 ] \
  || fail "F10: expected exactly one COMMITTED=\"\$(...)\" assignment line in the guard, found: $COMMITTED_ASSIGN_LINE"
COMMITTED_LINE="${COMMITTED_ASSIGN_LINE%%:*}"
REV_PARSE_GATE_LINE="$(grep -n '^if ! "\$GIT_BIN" .*rev-parse --verify --end-of-options' "$GUARD" || true)"
[ "$(printf '%s\n' "$REV_PARSE_GATE_LINE" | grep -c .)" -eq 1 ] \
  || fail "F10: expected exactly one rev-parse --verify gate line in the guard, found: $REV_PARSE_GATE_LINE"
REV_PARSE_LINE="${REV_PARSE_GATE_LINE%%:*}"

new_repo
echo "notes" > "$REPO/notes.md"
git -C "$REPO" add notes.md
git -C "$REPO" commit -qm "add notes.md"
XTRACE="$(PS4='+${LINENO}: ' bash -x "$GUARD" "$REPO" "$MERGEBASE" 2>&1 >/dev/null || true)"
printf '%s\n' "$XTRACE" \
  | grep -E "^\\+${REV_PARSE_LINE}: .*/git .*rev-parse --verify --end-of-options .*${MERGEBASE}\\^\\{commit\\}" >/dev/null \
  && pass "F3/F9/F10: the guard's own rev-parse gate line ($REV_PARSE_LINE) traces carrying the flag for this run's merge base" \
  || fail "F3/F9/F10: line $REV_PARSE_LINE (the guard's rev-parse gate) did not trace carrying the flag"
printf '%s\n' "$XTRACE" \
  | grep -E "^\\+\\+${COMMITTED_LINE}: .*/git .*diff --name-only --end-of-options .*${MERGEBASE}\\.\\.HEAD" >/dev/null \
  && pass "F2/F9/F10: the guard's own COMMITTED assignment line ($COMMITTED_LINE) traces carrying the flag for this run's merge base" \
  || fail "F2/F9/F10: line $COMMITTED_LINE (the guard's COMMITTED assignment) did not trace carrying the flag"

# 13. F7 (round 4, Bugbot's full-mode pass): the exit-2 branch for a
#     failing `git diff` invocation (line 11's own documented contract —
#     "or a git invocation failed") was never exercised by any case
#     above; a mutant that deletes the COMMITTED/STAGED/UNSTAGED error
#     handling entirely (`|| { echo ...; exit 2; }` -> `|| true`) passed
#     every prior case. Shim `git` to fail exactly one of the three
#     `diff --name-only` calls, matching an argv token unique to it:
#     COMMITTED's range string, STAGED's `--cached`. UNSTAGED's own
#     invocation (`git -C <path> diff --name-only`, no other args) has
#     no token that is not also present in the other two, so no
#     single-arg shim can fail it alone without also failing whichever
#     of COMMITTED/STAGED runs first — its handling is the identical
#     three-line pattern already proved for the other two and is left
#     as a stated, structural (not independently exercised) coverage
#     gap rather than a test that would silently fail the wrong call.
# shellcheck source=lib/test-git-shim.sh
. "$SCRIPT_DIR/lib/test-git-shim.sh"

new_repo
echo "notes" > "$REPO/notes.md"
git -C "$REPO" add notes.md
git -C "$REPO" commit -qm "add notes.md"
shim_failing_git "citation-trigger-shim-committed" "${MERGEBASE}..HEAD" "simulated diff failure (COMMITTED)"
COMMITTED_SHIM_DIR="$SHIM_DIR"
PATH="$COMMITTED_SHIM_DIR:$PATH" run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 2 ] && pass "F7: a failing COMMITTED diff call -> exit 2" \
  || fail "F7: a failing COMMITTED diff call: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"committed paths"*) pass "F7: COMMITTED failure reported with its own message" ;;
  *) fail "F7: COMMITTED failure: expected the 'committed paths' message, got: $OUT" ;;
esac
assert_shim_fired "$COMMITTED_SHIM_DIR" "F7 COMMITTED shim"

new_repo
echo "notes" > "$REPO/notes.md"
git -C "$REPO" add notes.md
shim_failing_git "citation-trigger-shim-staged" "--cached" "simulated diff failure (STAGED)"
STAGED_SHIM_DIR="$SHIM_DIR"
PATH="$STAGED_SHIM_DIR:$PATH" run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 2 ] && pass "F7: a failing STAGED diff call -> exit 2" \
  || fail "F7: a failing STAGED diff call: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"staged paths"*) pass "F7: STAGED failure reported with its own message" ;;
  *) fail "F7: STAGED failure: expected the 'staged paths' message, got: $OUT" ;;
esac
assert_shim_fired "$STAGED_SHIM_DIR" "F7 STAGED shim"

# Mutation proof: break the .mdc?$ matching pattern, confirm case 2 flips
# from pass to fail, then restore the guard from a backup copy and
# re-verify clean. Requires the guard to already exist — this proof only
# makes sense once the guard is GREEN, not during the RED phase before it
# is written.
if [ -f "$GUARD" ]; then
  BACKUP="$(mktemp "${TMPDIR:-/tmp}/citation-trigger-guard-backup.XXXXXX")"
  REPOS+=("$BACKUP")
  cp "$GUARD" "$BACKUP"
  sed -i.bak 's/\.mdc?\$/.never-matches$/' "$GUARD"
  rm -f "$GUARD.bak"
  new_repo
  echo "rule" > "$REPO/rule.mdc"
  git -C "$REPO" add rule.mdc
  git -C "$REPO" commit -qm "add rule.mdc"
  run_guard "$REPO" "$MERGEBASE"
  [ "$RC" -ne 0 ] && pass "mutation: breaking .mdc?\$ flips case 2 from pass to fail" \
    || fail "mutation: case 2 still passes with the matching pattern broken"
  cp "$BACKUP" "$GUARD"
  new_repo
  echo "rule" > "$REPO/rule.mdc"
  git -C "$REPO" add rule.mdc
  git -C "$REPO" commit -qm "add rule.mdc"
  run_guard "$REPO" "$MERGEBASE"
  [ "$RC" -eq 0 ] && pass "mutation: guard restored, case 2 passes again" \
    || fail "mutation: guard restore did not bring case 2 back to exit 0"
else
  fail "mutation proof skipped: $GUARD does not exist yet"
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-panel-citation-trigger: all cases pass\n'
