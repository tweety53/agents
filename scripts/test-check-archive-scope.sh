#!/usr/bin/env bash
# test-check-archive-scope.sh — assertion harness for
# check-archive-scope.sh, the archive-commit scope guard (KAN-472's
# follow-up). Every case builds a throwaway git repository under a
# sandboxed TMPDIR, stages paths inside and outside the declared allowed
# prefixes, and asserts the guard's verdict lines and exit codes. Nothing
# here stages so much as one path in a real repository.
#
# The cases pin the four behaviours the guard's own header declares: an
# in-scope staged path is SCOPE-OK; an out-of-scope one is named on its own
# OUT-OF-SCOPE line and counted on one SCOPE-VIOLATION line, exit 1; an
# empty staged diff is never a violation; and every inability to answer is
# exit 2 with no verdict line. Case 5 is the mutation-bearing one: a bare
# string-prefix match would let `prefix-sibling/x` through when only
# `prefix` is allowed, so the path-component rule is asserted against
# exactly that pair.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-archive-scope.sh"
FAILURES=0

ROOTS=()
cleanup() {
  [ "${#ROOTS[@]}" -eq 0 ] && return 0
  for r in "${ROOTS[@]}"; do
    rm -rf "$r"
  done
}
trap cleanup EXIT

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# new_repo -> sets REPO to a fresh git repository carrying one empty base
# commit, so a later `stage` has an index to stage into. REPO is resolved
# through pwd -P — the guard answers with the symlink-resolved root (it
# runs `cd "$WORKTREE_ARG" && pwd -P` itself), and on this macOS host
# mktemp's /var/... prefix resolves to /private/var/..., so the raw mktemp
# path would never match the verdict line.
new_repo() {
  REPO="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/check-archive-scope-test.XXXXXX")" && pwd -P)"
  ROOTS+=("$REPO")
  git -C "$REPO" init -q
  git -C "$REPO" config user.email guard-test@example.com
  git -C "$REPO" config user.name guard-test
  git -C "$REPO" commit -q --allow-empty -m base
}

# stage <repo-path>... -> creates each path (with parents) and stages it.
stage() {
  for p in "$@"; do
    mkdir -p "$REPO/$(dirname "$p")"
    printf 'x\n' > "$REPO/$p"
    git -C "$REPO" add "$p"
  done
}

run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# ---------------------------------------------------------------------------
# 1. A staged path inside the allowed prefix is SCOPE-OK, exit 0.
# ---------------------------------------------------------------------------
new_repo
stage "spectre/changes/kan-x/narrative.md"
run_guard "$REPO" "spectre/changes/"
[ "$RC" -eq 0 ] && pass "case 1: staged path inside the prefix exits 0" \
  || fail "case 1: rc=$RC out=$OUT"
case "$OUT" in
  *"SCOPE-OK: $REPO"*) pass "case 1: the verdict line is SCOPE-OK naming the worktree" ;;
  *) fail "case 1: no SCOPE-OK line naming the worktree: out=$OUT" ;;
esac

# ---------------------------------------------------------------------------
# 2. A staged path outside the prefix is named on its own OUT-OF-SCOPE
#    line, counted on one SCOPE-VIOLATION line, exit 1.
# ---------------------------------------------------------------------------
new_repo
stage "stats/web/src/App.tsx"
run_guard "$REPO" "spectre/changes/"
[ "$RC" -eq 1 ] && pass "case 2: staged path outside the prefix exits 1" \
  || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"OUT-OF-SCOPE: stats/web/src/App.tsx"*) pass "case 2: the offending path is named verbatim" ;;
  *) fail "case 2: offending path not named: out=$OUT" ;;
esac
case "$OUT" in
  *"SCOPE-VIOLATION: $REPO — 1"*) pass "case 2: the verdict line counts 1 offender" ;;
  *) fail "case 2: no SCOPE-VIOLATION count of 1: out=$OUT" ;;
esac

# ---------------------------------------------------------------------------
# 3. Two offenders and one in-scope neighbour: both offenders named, the
#    in-scope one never, count 2.
# ---------------------------------------------------------------------------
new_repo
stage "spectre/changes/kan-x/narrative.md" "stats/web/src/App.tsx" "stats/go.mod"
run_guard "$REPO" "spectre/changes/"
[ "$RC" -eq 1 ] && pass "case 3: mixed staged paths exit 1" \
  || fail "case 3: rc=$RC out=$OUT"
# One OUT-OF-SCOPE line per offender, in the guard's own diff order —
# asserted per path on its own line, never as one ordering-dependent glob.
if [ "$(printf '%s\n' "$OUT" | grep -c '^OUT-OF-SCOPE: ')" -eq 2 ]; then
  pass "case 3: exactly two OUT-OF-SCOPE lines, one per offender"
else
  fail "case 3: offenders not one per line: out=$OUT"
fi
case "$OUT" in
  *"OUT-OF-SCOPE: stats/web/src/App.tsx"*) pass "case 3: App.tsx named" ;;
  *) fail "case 3: App.tsx not named: out=$OUT" ;;
esac
case "$OUT" in
  *"OUT-OF-SCOPE: stats/go.mod"*) pass "case 3: go.mod named" ;;
  *) fail "case 3: go.mod not named: out=$OUT" ;;
esac
case "$OUT" in
  *"SCOPE-VIOLATION: $REPO — 2"*) pass "case 3: the verdict line counts 2 offenders" ;;
  *) fail "case 3: no SCOPE-VIOLATION count of 2: out=$OUT" ;;
esac
case "$OUT" in
  *OUT-OF-SCOPE:*narrative.md*) fail "case 3: the in-scope path was reported as an offender: out=$OUT" ;;
  *) pass "case 3: the in-scope path is never named as an offender" ;;
esac

# ---------------------------------------------------------------------------
# 4. Nothing staged at all is SCOPE-OK — an empty diff is never a
#    violation, per the guard's own header.
# ---------------------------------------------------------------------------
new_repo
run_guard "$REPO" "spectre/changes/"
[ "$RC" -eq 0 ] && pass "case 4: an empty staged diff exits 0" \
  || fail "case 4: rc=$RC out=$OUT"
case "$OUT" in
  *SCOPE-OK*) pass "case 4: the empty diff is SCOPE-OK" ;;
  *) fail "case 4: empty diff not reported SCOPE-OK: out=$OUT" ;;
esac

# ---------------------------------------------------------------------------
# 5. THE PATH-COMPONENT RULE: a prefix given without its trailing slash is
#    normalized to one, so `prefix` allows `prefix/ok` but never lets
#    `prefix-sibling/bad` through — the exact pair the guard's header
#    names (`spectre/changes/` vs `spectre/changes-backup/x`). A bare
#    string-prefix mutant fails here.
# ---------------------------------------------------------------------------
new_repo
stage "prefix/ok.txt" "prefix-sibling/bad.txt"
run_guard "$REPO" "prefix"
[ "$RC" -eq 1 ] && pass "case 5: a string-prefix sibling still exits 1" \
  || fail "case 5: rc=$RC out=$OUT"
case "$OUT" in
  *OUT-OF-SCOPE:*prefix-sibling/bad.txt*) pass "case 5: only the sibling path is named" ;;
  *) fail "case 5: sibling path not the named offender: out=$OUT" ;;
esac
case "$OUT" in
  *OUT-OF-SCOPE:*prefix/ok.txt*) fail "case 5: the truly in-scope path was named: out=$OUT" ;;
  *) pass "case 5: the in-scope path under the normalized prefix is not named" ;;
esac

# ---------------------------------------------------------------------------
# 6. Several allowed prefixes: a staged path matching the SECOND one is in
#    scope.
# ---------------------------------------------------------------------------
new_repo
stage "stats/web/src/App.tsx"
run_guard "$REPO" "spectre/changes/" "stats/web/"
[ "$RC" -eq 0 ] && pass "case 6: a path matching a later prefix exits 0" \
  || fail "case 6: rc=$RC out=$OUT"

# ---------------------------------------------------------------------------
# 7. Cannot answer — exit 2, and no verdict line on stdout: a
#    non-directory worktree, a directory that is not a git worktree, and a
#    call with no allowed prefix at all. 7a passes a real second argument
#    so the arity check cannot answer for it — the non-directory branch is
#    what exits 2, not `[ "$#" -ge 2 ]`.
# ---------------------------------------------------------------------------
run_guard "${TMPDIR:-/tmp}/check-archive-scope-missing-$$" "spectre/changes/"
[ "$RC" -eq 2 ] && pass "case 7a: a non-directory worktree exits 2" \
  || fail "case 7a: rc=$RC out=$OUT"
case "$OUT" in
  *SCOPE-OK*|*SCOPE-VIOLATION*) fail "case 7a: a verdict line was printed despite exit 2: out=$OUT" ;;
  *) pass "case 7a: no verdict line on an inability to answer" ;;
esac

PLAIN="$(mktemp -d "${TMPDIR:-/tmp}/check-archive-scope-plain.XXXXXX")"
ROOTS+=("$PLAIN")
run_guard "$PLAIN" "spectre/changes/"
[ "$RC" -eq 2 ] && pass "case 7b: a non-git directory exits 2" \
  || fail "case 7b: rc=$RC out=$OUT"

new_repo
run_guard "$REPO"
[ "$RC" -eq 2 ] && pass "case 7c: no allowed prefix exits 2" \
  || fail "case 7c: rc=$RC out=$OUT"

# ---------------------------------------------------------------------------
if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-archive-scope: all cases pass\n'
