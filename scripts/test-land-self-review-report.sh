#!/usr/bin/env bash
# test-land-self-review-report.sh — assertion harness for
# land-self-review-report.sh, the one landing chain for a self-review
# report or context bundle (kan-523). Every case builds a throwaway git
# repository (and, where the case pushes, a throwaway remote) under a
# sandboxed TMPDIR and asserts the script's verdict lines and exit codes.
# Nothing here touches a real repository.
#
# The cases pin the behaviours that took two panel rounds to get right in
# prose (F3: the branch re-assert; F4: pull/push inside the same guard):
# a branch mismatch runs nothing at all; a full site-1 landing adds the
# report, removes the context bundle, commits, pull --rebases and pushes
# in that order; the archive shape commits with no push; an empty staged
# diff is a clean no-op, never a failed commit; a rejected push exits
# non-zero with the commit left local; both usage defects exit 2; foreign
# staged work refuses the commit (kan-657); and a branch switched mid-run
# is re-caught before the commit.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_UNDER_TEST="$SCRIPT_DIR/land-self-review-report.sh"
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

# new_repo -> sets REPO to a fresh git repository on branch `main` carrying
# one empty base commit, resolved through pwd -P (macOS mktemp's /var
# prefix resolves to /private/var, and the assertions compare paths).
new_repo() {
  REPO="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/land-self-review-test.XXXXXX")" && pwd -P)"
  ROOTS+=("$REPO")
  git -C "$REPO" init -q -b main
  git -C "$REPO" config user.email land-test@example.com
  git -C "$REPO" config user.name land-test
  git -C "$REPO" commit -q --allow-empty -m base
}

# write_report <name> -> drops an untracked self-review report and keeps a
# tracked context bundle, the mid-flight state both call sites start from.
write_report() {
  local name="$1"
  mkdir -p "$REPO/docs/self-review"
  printf 'report\n' > "$REPO/docs/self-review/$name-self-review.md"
  if [ ! -f "$REPO/docs/self-review/$name-context.md" ]; then
    printf 'bundle\n' > "$REPO/docs/self-review/$name-context.md"
    git -C "$REPO" add "docs/self-review/$name-context.md"
    git -C "$REPO" commit -q -m "docs(self-review): $name context bundle"
  fi
}

run_script() {
  set +e
  OUT="$("$SCRIPT_UNDER_TEST" "$@" 2>&1)"
  RC=$?
  set -e
}

head_subject() { git -C "$REPO" log -1 --format=%s; }

# ---------------------------------------------------------------------------
# 1-2. Usage defects exit 2 with a usage line: too few arguments, and a
#      trailing --push with no base.
# ---------------------------------------------------------------------------
new_repo
run_script "$REPO" main "subject" "docs/self-review/n-self-review.md" "docs/self-review/n-context.md" "extra"
[ "$RC" -eq 2 ] && pass "test_land_self_review_report_usage: too many positionals exits 2" \
  || fail "test_land_self_review_report_usage: rc=$RC out=$OUT"
case "$OUT" in
  *"usage: land-self-review-report.sh"*) pass "test_land_self_review_report_usage: the usage line is printed" ;;
  *) fail "test_land_self_review_report_usage: no usage line: out=$OUT" ;;
esac

new_repo
run_script "$REPO" main "subject" "docs/self-review/n-self-review.md" --push
[ "$RC" -eq 2 ] && pass "test_land_self_review_report_usage: --push without a base exits 2" \
  || fail "test_land_self_review_report_usage: rc=$RC out=$OUT"

new_repo
run_script "$REPO" main "subject"
[ "$RC" -eq 2 ] && pass "test_land_self_review_report_usage: too few arguments exits 2" \
  || fail "test_land_self_review_report_usage: rc=$RC out=$OUT"

# ---------------------------------------------------------------------------
# 3. A branch mismatch runs nothing at all: exit 1, one
#    LAND-BRANCH-MISMATCH line naming the found branch, nothing staged,
#    no commit, the report still untracked.
# ---------------------------------------------------------------------------
new_repo
write_report kan-x
BASE_HEAD="$(git -C "$REPO" rev-parse HEAD)"
run_script "$REPO" "chore/archive-kan-x" "docs(self-review): kan-x self-review report" \
  "docs/self-review/kan-x-self-review.md"
[ "$RC" -eq 1 ] && pass "test_land_branch_mismatch_stops_everything: exits 1" \
  || fail "test_land_branch_mismatch_stops_everything: rc=$RC out=$OUT"
case "$OUT" in
  *"LAND-BRANCH-MISMATCH: expected chore/archive-kan-x, found main"*)
    pass "test_land_branch_mismatch_stops_everything: the mismatch names both branches" ;;
  *) fail "test_land_branch_mismatch_stops_everything: no mismatch line: out=$OUT" ;;
esac
[ -z "$(git -C "$REPO" diff --cached --name-only)" ] \
  && pass "test_land_branch_mismatch_stops_everything: nothing was staged" \
  || fail "test_land_branch_mismatch_stops_everything: the index is dirty: $(git -C "$REPO" diff --cached --name-only)"
[ "$(git -C "$REPO" rev-parse HEAD)" = "$BASE_HEAD" ] \
  && pass "test_land_branch_mismatch_stops_everything: no commit was made" \
  || fail "test_land_branch_mismatch_stops_everything: HEAD moved"
[ -n "$(git -C "$REPO" status --porcelain docs/self-review/kan-x-self-review.md)" ] \
  && pass "test_land_branch_mismatch_stops_everything: the report is still untracked" \
  || fail "test_land_branch_mismatch_stops_everything: the report was staged"

# ---------------------------------------------------------------------------
# 4. The site-1 landing: report added, context bundle removed, one commit
#    with the exact subject, pull --rebase and push both run inside the
#    guard — origin/main carries the commit and the worktree lost the
#    bundle.
# ---------------------------------------------------------------------------
new_repo
write_report kan-y
ORIGIN_BARE="$(mktemp -d "${TMPDIR:-/tmp}/land-self-review-origin.XXXXXX")"
ROOTS+=("$ORIGIN_BARE")
git -C "$ORIGIN_BARE" init -q --bare -b main
git -C "$REPO" remote add origin "$ORIGIN_BARE"
git -C "$REPO" push -q -u origin main
run_script "$REPO" main "docs(self-review): kan-y self-review report" \
  "docs/self-review/kan-y-self-review.md" "docs/self-review/kan-y-context.md" --push main
[ "$RC" -eq 0 ] && pass "test_land_report_lands_and_pushes: exits 0" \
  || fail "test_land_report_lands_and_pushes: rc=$RC out=$OUT"
[ "$(head_subject)" = "docs(self-review): kan-y self-review report" ] \
  && pass "test_land_report_lands_and_pushes: the commit subject is exact" \
  || fail "test_land_report_lands_and_pushes: subject is $(head_subject)"
[ "$(git -C "$REPO" rev-parse origin/main)" = "$(git -C "$REPO" rev-parse HEAD)" ] \
  && pass "test_land_report_lands_and_pushes: origin/main carries the commit" \
  || fail "test_land_report_lands_and_pushes: origin/main is behind HEAD"
[ ! -f "$REPO/docs/self-review/kan-y-context.md" ] \
  && pass "test_land_report_lands_and_pushes: the context bundle was removed" \
  || fail "test_land_report_lands_and_pushes: the context bundle is still on disk"
[ -f "$REPO/docs/self-review/kan-y-self-review.md" ] \
  && pass "test_land_report_lands_and_pushes: the report is on disk" \
  || fail "test_land_report_lands_and_pushes: the report is missing"

# ---------------------------------------------------------------------------
# 5. The archive shape: no rm path, no --push, no remote at all — the
#    commit lands on the asserted branch and nothing tries to pull or
#    push (a remote-less repository would fail if they ran).
# ---------------------------------------------------------------------------
new_repo
write_report kan-z
run_script "$REPO" main "docs(self-review): kan-z self-review report" \
  "docs/self-review/kan-z-self-review.md"
[ "$RC" -eq 0 ] && pass "test_land_report_lands_and_pushes: the archive shape exits 0 with no remote" \
  || fail "test_land_report_lands_and_pushes: archive shape rc=$RC out=$OUT"
[ "$(head_subject)" = "docs(self-review): kan-z self-review report" ] \
  && pass "test_land_report_lands_and_pushes: the archive shape commits" \
  || fail "test_land_report_lands_and_pushes: archive subject is $(head_subject)"

# ---------------------------------------------------------------------------
# 6. Nothing staged is a clean no-op: exit 0, one LAND-NOTHING-TO-COMMIT
#    line, HEAD unchanged.
# ---------------------------------------------------------------------------
new_repo
write_report kan-w
git -C "$REPO" add docs/self-review/kan-w-self-review.md
git -C "$REPO" commit -q -m "docs(self-review): kan-w self-review report"
BASE_HEAD="$(git -C "$REPO" rev-parse HEAD)"
run_script "$REPO" main "docs(self-review): kan-w self-review report" \
  "docs/self-review/kan-w-self-review.md"
[ "$RC" -eq 0 ] && pass "test_land_bundle_skips_empty_commit: an empty staged diff exits 0" \
  || fail "test_land_bundle_skips_empty_commit: rc=$RC out=$OUT"
case "$OUT" in
  *"LAND-NOTHING-TO-COMMIT"*) pass "test_land_bundle_skips_empty_commit: the no-op is named" ;;
  *) fail "test_land_bundle_skips_empty_commit: no no-op line: out=$OUT" ;;
esac
[ "$(git -C "$REPO" rev-parse HEAD)" = "$BASE_HEAD" ] \
  && pass "test_land_bundle_skips_empty_commit: no empty commit was made" \
  || fail "test_land_bundle_skips_empty_commit: HEAD moved"

# ---------------------------------------------------------------------------
# 7. A rejected push leaves the commit local and exits non-zero: the
#    remote refuses the push (a non-bare remote with main checked out),
#    the script reports the failure, and HEAD is still the report commit.
# ---------------------------------------------------------------------------
new_repo
write_report kan-v
# The remote is a clone, not a push target: its main is checked out, so the
# script's own push is the first push into it and is refused.
ORIGIN_LIVE="$(mktemp -d "${TMPDIR:-/tmp}/land-self-review-live.XXXXXX")"
ROOTS+=("$ORIGIN_LIVE")
git clone -q "$REPO" "$ORIGIN_LIVE"
git -C "$REPO" remote add origin "$ORIGIN_LIVE"
git -C "$REPO" fetch -q origin
run_script "$REPO" main "docs(self-review): kan-v self-review report" \
  "docs/self-review/kan-v-self-review.md" "docs/self-review/kan-v-context.md" --push main
[ "$RC" -ne 0 ] && pass "test_land_report_lands_and_pushes: a rejected push exits non-zero" \
  || fail "test_land_report_lands_and_pushes: rejected push exited 0: out=$OUT"
[ "$(head_subject)" = "docs(self-review): kan-v self-review report" ] \
  && pass "test_land_report_lands_and_pushes: the commit stays local on a rejected push" \
  || fail "test_land_report_lands_and_pushes: the commit is gone: $(head_subject)"

# ---------------------------------------------------------------------------
# 8. Foreign staged work refuses the commit: the chain stages only its own
#    paths but a bare git commit takes the whole index, so a shared
#    checkout's foreign staged content would be swept in (kan-657, where
#    121 foreign paths landed as 66ae176). The chain must refuse instead:
#    exit 3, one LAND-FOREIGN-STAGED line naming the foreign path, HEAD
#    unchanged, nothing unstaged and nothing lost.
# ---------------------------------------------------------------------------
new_repo
write_report kan-f
printf 'stray\n' > "$REPO/stray.txt"
git -C "$REPO" add stray.txt
BASE_HEAD="$(git -C "$REPO" rev-parse HEAD)"
run_script "$REPO" main "docs(self-review): kan-f self-review report" \
  "docs/self-review/kan-f-self-review.md"
[ "$RC" -eq 3 ] && pass "test_land_foreign_staged_refuses: exits 3" \
  || fail "test_land_foreign_staged_refuses: rc=$RC out=$OUT"
case "$OUT" in
  *"LAND-FOREIGN-STAGED"*) pass "test_land_foreign_staged_refuses: the refusal is named" ;;
  *) fail "test_land_foreign_staged_refuses: no refusal line: out=$OUT" ;;
esac
case "$OUT" in
  *stray.txt*) pass "test_land_foreign_staged_refuses: the refusal names the foreign path" ;;
  *) fail "test_land_foreign_staged_refuses: foreign path not named: out=$OUT" ;;
esac
[ "$(git -C "$REPO" rev-parse HEAD)" = "$BASE_HEAD" ] \
  && pass "test_land_foreign_staged_refuses: no commit was made" \
  || fail "test_land_foreign_staged_refuses: HEAD moved"
STAGED_NOW="$(git -C "$REPO" diff --cached --name-only)"
case "$STAGED_NOW" in
  *stray.txt*) pass "test_land_foreign_staged_refuses: the foreign path is still staged" ;;
  *) fail "test_land_foreign_staged_refuses: the foreign path left the index: $STAGED_NOW" ;;
esac
case "$STAGED_NOW" in
  *kan-f-self-review.md*) pass "test_land_foreign_staged_refuses: the report is still staged" ;;
  *) fail "test_land_foreign_staged_refuses: the report left the index: $STAGED_NOW" ;;
esac

# ---------------------------------------------------------------------------
# 9. The same refusal under an rm path: a foreign staged file beside the
#    report-and-bundle shape refuses too. The chain's own rm has already
#    run by then, so the bundle is off-disk and its deletion staged — but
#    nothing is committed: HEAD still carries the bundle.
# ---------------------------------------------------------------------------
new_repo
write_report kan-r
printf 'stray\n' > "$REPO/stray.txt"
git -C "$REPO" add stray.txt
BASE_HEAD="$(git -C "$REPO" rev-parse HEAD)"
run_script "$REPO" main "docs(self-review): kan-r self-review report" \
  "docs/self-review/kan-r-self-review.md" "docs/self-review/kan-r-context.md"
[ "$RC" -eq 3 ] && pass "test_land_foreign_staged_with_rm_refuses: exits 3" \
  || fail "test_land_foreign_staged_with_rm_refuses: rc=$RC out=$OUT"
case "$OUT" in
  *"LAND-FOREIGN-STAGED"*) pass "test_land_foreign_staged_with_rm_refuses: the refusal is named" ;;
  *) fail "test_land_foreign_staged_with_rm_refuses: no refusal line: out=$OUT" ;;
esac
[ "$(git -C "$REPO" rev-parse HEAD)" = "$BASE_HEAD" ] \
  && pass "test_land_foreign_staged_with_rm_refuses: no commit was made" \
  || fail "test_land_foreign_staged_with_rm_refuses: HEAD moved"
git -C "$REPO" cat-file -e "$BASE_HEAD:docs/self-review/kan-r-context.md" \
  && pass "test_land_foreign_staged_with_rm_refuses: HEAD still carries the context bundle" \
  || fail "test_land_foreign_staged_with_rm_refuses: the context bundle is gone from HEAD"

# ---------------------------------------------------------------------------
# 10. A branch switched mid-run is re-caught: the start-of-run assert can
#     be stale on a shared checkout — a concurrent session switches it
#     mid-run (kan-657) — so the chain re-asserts before the commit. A
#     PATH git shim delegates every call to the real git but switches the
#     checkout's branch on the chain's first `diff`, between the add and
#     the commit: exit 1, one LAND-BRANCH-MISMATCH line naming both
#     branches, no commit.
# ---------------------------------------------------------------------------
new_repo
write_report kan-s
REAL_GIT="$(command -v git)"
SHIM_DIR="$(mktemp -d "${TMPDIR:-/tmp}/land-self-review-shim.XXXXXX")"
ROOTS+=("$SHIM_DIR")
MARKER="$SHIM_DIR/switched"
cat > "$SHIM_DIR/git" <<EOF
#!/bin/bash
# The chain calls git -C <repo> <verb ...>: skip the -C pair before
# reading the verb.
FIRST="\$1"
[ "\$FIRST" = -C ] && FIRST="\$3"
if [ "\$FIRST" = diff ] && [ ! -f "$MARKER" ]; then
  touch "$MARKER"
  "$REAL_GIT" -C "$REPO" checkout -b other >/dev/null 2>&1
fi
exec "$REAL_GIT" "\$@"
EOF
chmod +x "$SHIM_DIR/git"
BASE_HEAD="$(git -C "$REPO" rev-parse HEAD)"
PATH="$SHIM_DIR:$PATH" run_script "$REPO" main "docs(self-review): kan-s self-review report" \
  "docs/self-review/kan-s-self-review.md"
[ "$RC" -eq 1 ] && pass "test_land_branch_switched_midrun_refuses: exits 1" \
  || fail "test_land_branch_switched_midrun_refuses: rc=$RC out=$OUT"
case "$OUT" in
  *"LAND-BRANCH-MISMATCH: expected main, found other"*)
    pass "test_land_branch_switched_midrun_refuses: the re-assert names both branches" ;;
  *) fail "test_land_branch_switched_midrun_refuses: no mid-run mismatch line: out=$OUT" ;;
esac
[ "$(git -C "$REPO" rev-parse HEAD)" = "$BASE_HEAD" ] \
  && pass "test_land_branch_switched_midrun_refuses: no commit was made" \
  || fail "test_land_branch_switched_midrun_refuses: HEAD moved"

# ---------------------------------------------------------------------------
if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'land-self-review-report: all cases pass\n'
