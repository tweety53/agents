#!/usr/bin/env bash
# Assertion harness for check-finish-preflight.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's verdict and
# exit status. Never touches the real repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in skills/myflow-contracts/pipeline.md, never against observed
# output. test-check-plan-provenance.sh's header records that suite encoding
# the guard's own defects as its specification more than once, which then made
# each defect look verified.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-finish-preflight.sh"
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
SHIM_DIR="$(mktemp -d "${TMPDIR:-/tmp}/finish-preflight-shim.XXXXXX")"
REPOS+=("$SHIM_DIR")
REAL_GIT="$(command -v git)"
{
  printf '#!/usr/bin/env bash\n'
  printf 'for a in "$@"; do\n'
  printf '  if [ "$a" = "status" ]; then\n'
  printf '    echo "fatal: simulated index.lock contention" >&2\n'
  printf '    exit 128\n'
  printf '  fi\n'
  printf 'done\n'
  printf 'exec %s "$@"\n' "$REAL_GIT"
} > "$SHIM_DIR/git"
chmod +x "$SHIM_DIR/git"
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

# 9. A missing argument is programmer error, reported as exit 2 rather than
#    guessed at.
run_guard "" "" ""
[ "$RC" -eq 2 ] && pass "missing arguments -> exit 2" \
  || fail "missing arguments: expected exit 2, got rc=$RC out=$OUT"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-finish-preflight: all cases pass\n'
