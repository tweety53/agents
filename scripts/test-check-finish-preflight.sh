#!/usr/bin/env bash
# Assertion harness for check-finish-preflight.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's verdict and
# exit status. Never touches the real repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in skills/flow-contracts/pipeline.md, never against observed
# output. test-check-plan-provenance.sh's header records that suite encoding
# the guard's own defects as its specification more than once, which then made
# each defect look verified.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-finish-preflight.sh"
# shellcheck source=lib/test-git-shim.sh
. "$SCRIPT_DIR/lib/test-git-shim.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# Every case builds a whole git repository rather than a fixture tree of files,
# so the sandbox is heavier than the one the other harnesses leave behind. It
# is removed on exit, including on a failed assertion.
# An indexed array, not a space-separated string: sandbox paths come from
# mktemp under TMPDIR, which may contain spaces, and word-splitting a string
# would then rm -rf the fragments. Bash 3.2 has indexed arrays; only
# associative arrays are unavailable.
REPOS=()
cleanup() {
  # ${REPOS[@]} is unset-expansion-unsafe under `set -u` on bash 3.2 when empty.
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

# run_guard <worktree> <base-ref> <recorded-merge-base> -> sets RC and OUT
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" "$2" "$3" 2>&1)"
  RC=$?
  set -e
}

# new_repo -> sets REPO, BASE_REF, RECORDED_BASE
# A repository on `main` with one commit, plus a branch `openspec/demo`
# checked out at that same commit. RECORDED_BASE is that commit.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/finish-preflight-test.XXXXXX")"
  REPOS+=("$REPO")
  git -C "$REPO" init -q -b main
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  echo base > "$REPO/file.txt"
  git -C "$REPO" add file.txt
  git -C "$REPO" commit -qm "base"
  RECORDED_BASE="$(git -C "$REPO" rev-parse HEAD)"
  BASE_REF=main
  git -C "$REPO" checkout -q -b openspec/demo
}

# merge_demo_into_main <repo> — land openspec/demo on main with a merge commit
# and return to the branch, which is the shape run 2 is allowed to act on.
merge_demo_into_main() {
  git -C "$1" checkout -q main
  git -C "$1" merge -q --no-ff -m "merge" openspec/demo
  git -C "$1" checkout -q openspec/demo
}

# 1. Zero-commit branch with staged work: RUN1, never RUN2.
#    HEAD is still the merge base, so the ancestor test alone says "merged".
new_repo
echo staged > "$REPO/new.txt"
git -C "$REPO" add new.txt
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  RUN1*) pass "zero-commit branch with staged work -> RUN1" ;;
  *) fail "zero-commit branch: expected RUN1, got rc=$RC out=$OUT" ;;
esac
[ "$RC" -eq 0 ] && pass "zero-commit branch: exit 0" \
  || fail "zero-commit branch: expected exit 0, got rc=$RC"

# 1b. The same zero-commit shape, but the base ref does not resolve: still RUN1.
#     Signal (b) needs only HEAD and the recorded merge base, so it must be
#     answerable without the base ref. Deciding REFUSE here would hide the exact
#     KAN-14 shape behind an unrelated environmental failure.
new_repo
echo staged > "$REPO/new.txt"
git -C "$REPO" add new.txt
run_guard "$REPO" no-such-base "$RECORDED_BASE"
case "$OUT" in
  RUN1*) pass "zero-commit branch, unresolvable base ref -> RUN1" ;;
  *) fail "zero-commit branch, unresolvable base ref: expected RUN1, got rc=$RC out=$OUT" ;;
esac

# 1c. Zero-commit branch with a COMPLETELY CLEAN tree: still RUN1.
#     This is the shape signal (b) exists for, with every other signal pointing
#     the wrong way: the ancestor test says "merged" and the dirty check finds
#     nothing to object to, so (b) is the only thing between this worktree and
#     run 2's `git worktree remove --force`. Case 1 above cannot cover it —
#     there, (d)'s dirty check refuses on the staged entry even when (b) is
#     deleted, so case 1 passes a guard with no signal (b) at all.
new_repo
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  RUN1*) pass "zero-commit branch with a clean tree -> RUN1" ;;
  *) fail "zero-commit clean branch: expected RUN1, got rc=$RC out=$OUT" ;;
esac
[ "$RC" -eq 0 ] && pass "zero-commit clean branch: exit 0" \
  || fail "zero-commit clean branch: expected exit 0, got rc=$RC"

# 2. Genuinely merged, clean tree: RUN2.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
merge_demo_into_main "$REPO"
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  RUN2*) pass "merged with clean tree -> RUN2" ;;
  *) fail "merged clean: expected RUN2, got rc=$RC out=$OUT" ;;
esac
[ "$RC" -eq 0 ] && pass "merged clean: exit 0" \
  || fail "merged clean: expected exit 0, got rc=$RC"

# 3. Merged by ancestry but the worktree is dirty: REFUSE.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
merge_demo_into_main "$REPO"
echo dirty > "$REPO/file.txt"
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  REFUSE*) pass "merged but dirty -> REFUSE" ;;
  *) fail "merged dirty: expected REFUSE, got rc=$RC out=$OUT" ;;
esac
[ "$RC" -eq 0 ] && pass "merged dirty: exit 0 — the verdict carries the answer" \
  || fail "merged dirty: expected exit 0, got rc=$RC"

# 4. Unmerged and dirty: RUN1, NOT a refusal. This is the ordinary
#    IN_PROGRESS state and must not prompt the operator on every finish.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
echo dirty > "$REPO/file.txt"
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  RUN1*) pass "unmerged and dirty -> RUN1" ;;
  *) fail "unmerged dirty: expected RUN1, got rc=$RC out=$OUT" ;;
esac

# 5. No recorded merge base: REFUSE, never an inferred verdict.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
merge_demo_into_main "$REPO"
run_guard "$REPO" "$BASE_REF" -
case "$OUT" in
  REFUSE*) pass "absent merge base -> REFUSE" ;;
  *) fail "absent merge base: expected REFUSE, got rc=$RC out=$OUT" ;;
esac
# The REASON, not just the prefix. The `-` sentinel branch can be deleted
# outright and every prefix assertion still passes: `-^{commit}` fails to
# resolve, so the next branch emits a REFUSE of its own. The prefix therefore
# cannot tell "the sentinel is handled" from "it happened to fail to resolve",
# and the operator is shown a ref-resolution problem for a state file that
# simply never recorded a merge base.
case "$OUT" in
  *"no merge base recorded"*) pass "absent merge base: the refusal says the record is absent" ;;
  *) fail "absent merge base: the reason does not say the record is absent: $OUT" ;;
esac
case "$OUT" in
  *"does not resolve"*) fail "absent merge base: reported as a ref-resolution failure: $OUT" ;;
  *) pass "absent merge base: not reported as a ref-resolution failure" ;;
esac

# 6. Abbreviated recorded sha, genuinely merged: RUN2. An abbreviation must
#    not read as a difference from a full HEAD.
new_repo
SHORT="$(git -C "$REPO" rev-parse --short "$RECORDED_BASE")"
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
merge_demo_into_main "$REPO"
run_guard "$REPO" "$BASE_REF" "$SHORT"
case "$OUT" in
  RUN2*) pass "abbreviated merge base -> RUN2" ;;
  *) fail "abbreviated merge base: expected RUN2, got rc=$RC out=$OUT" ;;
esac

# 7. A recorded merge base that no longer resolves is a REFUSE verdict, not an
#    environment failure: the tree was readable, the answer is "cannot tell".
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
run_guard "$REPO" "$BASE_REF" 0000000000000000000000000000000000000000
case "$OUT" in
  REFUSE*) pass "unresolvable merge base -> REFUSE" ;;
  *) fail "unresolvable merge base: expected REFUSE, got rc=$RC out=$OUT" ;;
esac
[ "$RC" -eq 0 ] && pass "unresolvable merge base: exit 0" \
  || fail "unresolvable merge base: expected exit 0, got rc=$RC"

# 7b. A base ref that does not resolve is a REFUSE verdict, not RUN1. The
#     ancestor test fails for both "not merged" and "no such ref", and reading
#     the second as the first would integrate against a base nobody named.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
run_guard "$REPO" no-such-base "$RECORDED_BASE"
case "$OUT" in
  REFUSE*) pass "unresolvable base ref -> REFUSE" ;;
  *) fail "unresolvable base ref: expected REFUSE, got rc=$RC out=$OUT" ;;
esac

# 8. A path that is not a git worktree cannot be judged at all: exit 2, and no
#    verdict line, so a caller that greps for RUN2 cannot read one by accident.
NOT_A_REPO="$(mktemp -d "${TMPDIR:-/tmp}/finish-preflight-test.XXXXXX")"
REPOS+=("$NOT_A_REPO")
run_guard "$NOT_A_REPO" main deadbeef
[ "$RC" -eq 2 ] && pass "non-repository -> exit 2" \
  || fail "non-repository: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  RUN1*|RUN2*|REFUSE*) fail "non-repository: emitted a verdict line: $OUT" ;;
  *) pass "non-repository: emits no verdict line" ;;
esac

# 8b. `git status` failing on the merged-and-otherwise-clean shape is an
#     environment failure, not a verdict: exit 2 and no verdict line. Signal
#     (d) is the LAST thing between a merged branch and run 2, so a status read
#     that cannot be trusted must never fall through to `RUN2` — and it must not
#     abort silently either, which is what an unguarded pipeline under
#     `set -o pipefail` does when `wc` and `tr` succeed after git failed.
#     The shim fails only on `status` and passes every other subcommand through,
#     so the case exercises signal (d) alone.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
merge_demo_into_main "$REPO"
shim_failing_git "finish-preflight-shim" "status" "simulated index.lock contention"
set +e
OUT="$(PATH="$SHIM_DIR:$PATH" "$GUARD" "$REPO" "$BASE_REF" "$RECORDED_BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "unreadable worktree status -> exit 2" \
  || fail "unreadable status: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  RUN1*|RUN2*|REFUSE*) fail "unreadable status: emitted a verdict line: $OUT" ;;
  *) pass "unreadable status: emits no verdict line" ;;
esac
case "$OUT" in
  *check-finish-preflight:*) pass "unreadable status: names the failure" ;;
  *) fail "unreadable status: no named message: $OUT" ;;
esac
assert_shim_fired "$SHIM_DIR" "unreadable status"

# 8c. `git merge-base --is-ancestor` failing for a reason OTHER than "not an
#     ancestor" (exit 1) is an environment failure, not a verdict: exit 2 and
#     no verdict line. Signal (c) treats exit 1 alone as RUN1; any other
#     nonzero exit must fall through to the named exit-2 branch instead of
#     being folded into "not merged". Same shim technique as case 8b — fails
#     only `merge-base`, passes every other subcommand through.
new_repo
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
shim_failing_git "finish-preflight-shim" "merge-base" "simulated repository corruption"
set +e
OUT="$(PATH="$SHIM_DIR:$PATH" "$GUARD" "$REPO" "$BASE_REF" "$RECORDED_BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "merge-base failing (not 'not an ancestor') -> exit 2" \
  || fail "merge-base failure: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  RUN1*|RUN2*|REFUSE*) fail "merge-base failure: emitted a verdict line: $OUT" ;;
  *) pass "merge-base failure: emits no verdict line" ;;
esac
case "$OUT" in
  *check-finish-preflight:*) pass "merge-base failure: names the failure" ;;
  *) fail "merge-base failure: no named message: $OUT" ;;
esac
assert_shim_fired "$SHIM_DIR" "merge-base failure"

# 8d. `git rev-parse --verify HEAD^{commit}` failing (an unresolvable or
#     unborn HEAD) is an environment failure, not a verdict: exit 2 and no
#     verdict line. This is signal HEAD_SHA's own capture guard, which every
#     other case in this suite starts past — every new_repo builds a
#     worktree whose HEAD already resolves. Same shim technique as case 8b —
#     fails only `rev-parse`, passes every other subcommand through. (KAN-88
#     fix round 2, F7.)
new_repo
shim_failing_git "finish-preflight-shim" "HEAD^{commit}" "simulated unresolvable HEAD"
set +e
OUT="$(PATH="$SHIM_DIR:$PATH" "$GUARD" "$REPO" "$BASE_REF" "$RECORDED_BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "unresolvable HEAD -> exit 2" \
  || fail "unresolvable HEAD: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  RUN1*|RUN2*|REFUSE*) fail "unresolvable HEAD: emitted a verdict line: $OUT" ;;
  *) pass "unresolvable HEAD: emits no verdict line" ;;
esac
case "$OUT" in
  *"cannot resolve HEAD"*) pass "unresolvable HEAD: names the failure" ;;
  *) fail "unresolvable HEAD: no named message: $OUT" ;;
esac
assert_shim_fired "$SHIM_DIR" "unresolvable HEAD"

# 9. A missing argument is programmer error, reported as exit 2 rather than
#    guessed at.
run_guard "" "" ""
[ "$RC" -eq 2 ] && pass "missing arguments -> exit 2" \
  || fail "missing arguments: expected exit 2, got rc=$RC out=$OUT"

# --- KAN-88: the preflight resolves the base ref it was handed to its ---
# --- remote-tracking counterpart when one exists, so a bare local base ---
# --- name that has fallen behind origin cannot manufacture a false RUN1. ---
# Every case above builds a repository with no `origin` remote at all, so
# none of it exercises this path; its expectations are left byte-identical
# to prove that.

# new_repo_with_origin -> sets REPO, ORIGIN, BASE_REF, RECORDED_BASE
# A bare `origin`, a clone of it with one commit on `main` (RECORDED_BASE),
# pushed back to origin, and a branch `openspec/demo` checked out at that
# same commit — the same shape new_repo builds, plus a real remote.
new_repo_with_origin() {
  ORIGIN="$(mktemp -d "${TMPDIR:-/tmp}/finish-preflight-test-origin.XXXXXX")"
  REPOS+=("$ORIGIN")
  git init -q --bare -b main "$ORIGIN"

  REPO="$(mktemp -d "${TMPDIR:-/tmp}/finish-preflight-test.XXXXXX")"
  REPOS+=("$REPO")
  git clone -q "$ORIGIN" "$REPO" 2>/dev/null
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  echo base > "$REPO/file.txt"
  git -C "$REPO" add file.txt
  git -C "$REPO" commit -qm "base"
  git -C "$REPO" push -q origin main
  RECORDED_BASE="$(git -C "$REPO" rev-parse HEAD)"
  BASE_REF=main
  git -C "$REPO" checkout -q -b openspec/demo
}

# merge_demo_into_origin_main <repo> — land openspec/demo on origin/main
# WITHOUT moving the repo's own local `main` branch, so local `main` stays at
# RECORDED_BASE while origin/main advances — the exact staleness this task
# fixes. Merges on a throwaway local branch, pushes that to origin's `main`,
# deletes the throwaway, fetches so the repo's own origin/main tracking ref
# reflects the push, and returns to `openspec/demo` with a clean tree.
merge_demo_into_origin_main() {
  local repo="$1"
  git -C "$repo" branch -q tmp-merge main
  git -C "$repo" checkout -q tmp-merge
  git -C "$repo" merge -q --no-ff -m "merge" openspec/demo
  git -C "$repo" push -q origin tmp-merge:main
  git -C "$repo" checkout -q openspec/demo
  git -C "$repo" branch -q -D tmp-merge
  git -C "$repo" fetch -q origin
}

# 10. Merged into origin/main, but local main left behind at the recorded
#     merge base: handed the BARE name "main", must resolve to origin/main
#     and return RUN2 — the ticket's own regression case.
new_repo_with_origin
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
merge_demo_into_origin_main "$REPO"
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  RUN2*) pass "KAN-88: bare 'main' behind origin, merged into origin/main -> RUN2" ;;
  *) fail "KAN-88 bare main merged: expected RUN2, got rc=$RC out=$OUT" ;;
esac
case "$OUT" in
  *"origin/main"*) pass "KAN-88: RUN2 verdict names origin/main" ;;
  *) fail "KAN-88: RUN2 verdict does not name origin/main: $OUT" ;;
esac

# 11. The same repository, handed "origin/main" explicitly: unchanged, still
#     RUN2 naming origin/main — the correct caller (the contract composes
#     origin/$BASE itself) must be unaffected by this task.
run_guard "$REPO" "origin/main" "$RECORDED_BASE"
case "$OUT" in
  RUN2*) pass "KAN-88: explicit origin/main -> RUN2, unchanged" ;;
  *) fail "KAN-88 explicit origin/main: expected RUN2, got rc=$RC out=$OUT" ;;
esac
case "$OUT" in
  *"origin/main"*) pass "KAN-88: explicit origin/main verdict names origin/main" ;;
  *) fail "KAN-88: explicit origin/main verdict does not name origin/main: $OUT" ;;
esac

# 12. The same repository shape, branch NOT merged: handed the bare name
#     "main", must still resolve to origin/main and return RUN1 — proving the
#     substitution does not manufacture a merged verdict.
new_repo_with_origin
echo work > "$REPO/new.txt"
git -C "$REPO" add new.txt
git -C "$REPO" commit -qm "work"
run_guard "$REPO" "$BASE_REF" "$RECORDED_BASE"
case "$OUT" in
  RUN1*) pass "KAN-88: bare 'main' resolved to origin/main, unmerged -> RUN1" ;;
  *) fail "KAN-88 bare main unmerged: expected RUN1, got rc=$RC out=$OUT" ;;
esac
case "$OUT" in
  *"origin/main"*) pass "KAN-88: RUN1 verdict names origin/main" ;;
  *) fail "KAN-88: RUN1 verdict does not name origin/main: $OUT" ;;
esac

# 13. F9 (KAN-88 fix round 2): resolve_remote_base's documented contract for
#     a base ref beginning with a dash, both branches — an origin/ counterpart
#     that exists, and one that does not. Sourced and called directly.
#     resolve_remote_base always composes the lookup as
#     "refs/remotes/origin/<base-ref>^{commit}", which carries a fixed
#     "refs/remotes/origin/" prefix regardless of <base-ref>'s own content —
#     so the string that line 35's --end-of-options flag protects can never
#     itself begin with '-', and no base ref value (dash-prefixed or not)
#     can make that flag's presence externally observable. Confirmed by
#     direct mutation: dropping the flag and rerunning this case still
#     passes. Case 14 below covers the flag itself, on the guard's own
#     source, which is the only place its deletion is detectable.
new_repo_with_origin
DASH_SHA="$(git -C "$REPO" rev-parse HEAD)"
git -C "$REPO" update-ref "refs/remotes/origin/-weird" "$DASH_SHA"
# shellcheck source=lib/resolve-remote-base.sh
. "$SCRIPT_DIR/lib/resolve-remote-base.sh"
RESOLVED="$(resolve_remote_base "$REPO" "-weird")"
[ "$RESOLVED" = "origin/-weird" ] \
  && pass "F9: dash-prefixed base ref with an origin/ counterpart resolves to origin/-weird" \
  || fail "F9: dash-prefixed base ref did not resolve to origin/-weird, got: $RESOLVED"
RESOLVED2="$(resolve_remote_base "$REPO" "-no-such-ref")"
[ "$RESOLVED2" = "-no-such-ref" ] \
  && pass "F9: dash-prefixed base ref with no origin/ counterpart passes through unchanged" \
  || fail "F9: dash-prefixed base ref with no counterpart: expected pass-through, got: $RESOLVED2"

# 14. F9's mutant itself — dropping --end-of-options from line 35 of
#     resolve-remote-base.sh — cannot be caught behaviorally (case 13's
#     header explains why), so this asserts on the guard's own source, the
#     same precedent test-check-panel-reproducers.sh's case 19 sets for a
#     defensive pattern no external input can exercise.
grep -qF -- 'rev-parse --verify --end-of-options' "$SCRIPT_DIR/lib/resolve-remote-base.sh" \
  && pass "F9: resolve-remote-base.sh's rev-parse call still carries --end-of-options" \
  || fail "F9: resolve-remote-base.sh's rev-parse call is missing --end-of-options"

# 15. F14 (KAN-88 fix round 3): shim_failing_git's match is exact equality
#     against a single argv element, not a substring test. An argument that
#     merely CONTAINS the match-arg must pass straight through to the real
#     git rather than firing the shim — otherwise a match-arg like "status"
#     would also catch an unrelated argument such as "some/status/path", and
#     a case built on one subcommand could silently start failing calls it
#     never meant to touch. Exercised directly against the shim's own `git`,
#     not through a guard, since no guard invocation here is guaranteed to
#     hand git an argument that is a superstring of another call's match arg.
shim_failing_git "exact-match-shim" "abc" "should never fire"
set +e
SUPERSTRING_OUT="$("$SHIM_DIR/git" xabcx 2>&1)"
SUPERSTRING_RC=$?
set -e
case "$SUPERSTRING_OUT" in
  *"should never fire"*)
    fail "F14: shim intercepted an argument that merely contains the match arg: $SUPERSTRING_OUT" ;;
  *) pass "F14: an argument that merely contains the match arg is not intercepted" ;;
esac
[ "$SUPERSTRING_RC" -ne 128 ] \
  && pass "F14: superstring argument did not exit 128 (the shim's own fatal exit)" \
  || fail "F14: superstring argument exited 128 — the exact-match matcher is broken"
[ ! -e "$SHIM_DIR/.fired" ] \
  && pass "F14: superstring argument left no .fired sentinel" \
  || fail "F14: superstring argument left a .fired sentinel — the shim fired when it should not have"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-finish-preflight: all cases pass\n'
