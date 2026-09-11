#!/usr/bin/env bash
# Assertion harness for prepare-archive-branch.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's exit status,
# stdout AND stderr together for every case. Never touches the real
# repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in this repository's own kan-239-run-2-asserts-base-branch-and-archives-via-pr
# change (now archived), its
# "Run 2 positions the main checkout before it archives" requirement, never
# against observed output.
#
# Every case checks stdout, stderr AND exit status together, and — for the
# success cases — the state the guard actually left the checkout in: which
# branch HEAD is on, whether the base was fast-forwarded, and whether an
# existing archive branch's commits survived. A guard that exits 0 while
# leaving HEAD on the wrong branch is the exact failure this whole change
# exists to remove, so a case that only checks the exit code proves nothing.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/prepare-archive-branch.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

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

# run_guard <main-checkout> <base> <archive-branch> -> sets OUT, ERR, RC.
# stdout and stderr are captured separately (never merged) so a case can
# assert both independently.
run_guard() {
  ERRFILE="$(mktemp "${TMPDIR:-/tmp}/prepare-archive-branch-err.XXXXXX")"
  set +e
  OUT="$("$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
  rm -f "$ERRFILE"
}

# expect_positioned <case-name> <expected-from> <expected-to> — exit 0,
# stdout is exactly one line naming both the branch the checkout started
# from and the branch it is now on, stderr is empty.
expect_positioned() {
  local case_name="$1" expected_from="$2" expected_to="$3"
  [ "$RC" -eq 0 ] && pass "$case_name: exit 0" \
    || fail "$case_name: expected exit 0, got rc=$RC out=[$OUT] err=[$ERR]"
  case "$OUT" in
    *"$expected_from"*"$expected_to"*)
      pass "$case_name: stdout names '$expected_from' and '$expected_to'" ;;
    *)
      fail "$case_name: expected stdout naming '$expected_from' and '$expected_to', got '$OUT'" ;;
  esac
  LINES="$(printf '%s\n' "$OUT" | wc -l | tr -d ' ')"
  [ "$LINES" -eq 1 ] && pass "$case_name: stdout is exactly one line" \
    || fail "$case_name: expected exactly one line of stdout, got '$OUT'"
  [ -z "$ERR" ] && pass "$case_name: stderr empty" \
    || fail "$case_name: expected empty stderr, got '$ERR'"
}

# expect_refusal <case-name> <expected-exit> <stderr-substring> — the given
# exit code, empty stdout, and stderr containing the substring.
expect_refusal() {
  local case_name="$1" expected_rc="$2" substring="$3"
  [ "$RC" -eq "$expected_rc" ] && pass "$case_name: exit $expected_rc" \
    || fail "$case_name: expected exit $expected_rc, got rc=$RC out=[$OUT] err=[$ERR]"
  [ -z "$OUT" ] && pass "$case_name: stdout empty" \
    || fail "$case_name: expected empty stdout, got '$OUT'"
  case "$ERR" in
    *"$substring"*) pass "$case_name: stderr names the failure" ;;
    *) fail "$case_name: expected stderr to mention '$substring', got '$ERR'" ;;
  esac
}

current_branch() { git -C "$WT" branch --show-current 2>/dev/null; }

# new_checkout -> sets REMOTE, SEED, WT. A bare "origin" repository with a
# `main` branch, seeded from SEED with one commit and pushed, then cloned
# into WT with refs/remotes/origin/HEAD explicitly set. WT is left checked
# out on `main`, exactly as `git clone` leaves it — this is the guard's
# "main checkout" argument in every case below.
new_checkout() {
  REMOTE="$(mktemp -d "${TMPDIR:-/tmp}/prepare-archive-branch-test.XXXXXX")"
  REPOS+=("$REMOTE")
  rm -rf "$REMOTE"
  git init -q -b main --bare "$REMOTE"

  SEED="$(mktemp -d "${TMPDIR:-/tmp}/prepare-archive-branch-test.XXXXXX")"
  REPOS+=("$SEED")
  git init -q -b main "$SEED"
  git -C "$SEED" config user.email test@example.invalid
  git -C "$SEED" config user.name "Test"
  echo base > "$SEED/file.txt"
  git -C "$SEED" add file.txt
  git -C "$SEED" commit -qm base
  git -C "$SEED" remote add origin "$REMOTE"
  git -C "$SEED" push -q origin main

  WT="$(mktemp -d "${TMPDIR:-/tmp}/prepare-archive-branch-test.XXXXXX")"
  REPOS+=("$WT")
  rm -rf "$WT"
  git clone -q "$REMOTE" "$WT"
  git -C "$WT" remote set-head origin -a >/dev/null
}

# advance_origin -> commits a second, uniquely-named file directly on SEED
# and pushes it, so origin/main moves ahead of WT's local main without WT
# knowing yet — exactly "one commit behind origin/<base>".
advance_origin() {
  echo more > "$SEED/file2.txt"
  git -C "$SEED" add file2.txt
  git -C "$SEED" commit -qm more
  git -C "$SEED" push -q origin main
}

ARCHIVE="chore/archive-fixture"

# 1. on-base-clean: checkout on <base>, clean, one commit behind
#    origin/<base>. Expect exit 0; <base> fast-forwarded; HEAD on
#    <archive-branch>, carrying the fast-forwarded commit.
new_checkout
advance_origin
ORIGIN_TIP="$(git -C "$SEED" rev-parse main)"
run_guard "$WT" main "$ARCHIVE"
expect_positioned "on-base-clean" "main" "$ARCHIVE"
[ "$(current_branch)" = "$ARCHIVE" ] \
  && pass "on-base-clean: HEAD is on $ARCHIVE" \
  || fail "on-base-clean: expected HEAD on $ARCHIVE, got '$(current_branch)'"
git -C "$WT" merge-base --is-ancestor "$ORIGIN_TIP" HEAD \
  && pass "on-base-clean: archive branch carries the fast-forwarded base" \
  || fail "on-base-clean: archive branch does not carry the fast-forwarded base"

# 2. off-base-clean: checkout on an unrelated branch, clean. Expect exit 0;
#    HEAD on <archive-branch>, cut from a fast-forwarded <base>.
new_checkout
advance_origin
ORIGIN_TIP="$(git -C "$SEED" rev-parse main)"
git -C "$WT" checkout -q -b other
run_guard "$WT" main "$ARCHIVE"
expect_positioned "off-base-clean" "other" "$ARCHIVE"
[ "$(current_branch)" = "$ARCHIVE" ] \
  && pass "off-base-clean: HEAD is on $ARCHIVE" \
  || fail "off-base-clean: expected HEAD on $ARCHIVE, got '$(current_branch)'"
git -C "$WT" merge-base --is-ancestor "$ORIGIN_TIP" HEAD \
  && pass "off-base-clean: archive branch carries the fast-forwarded base" \
  || fail "off-base-clean: archive branch does not carry the fast-forwarded base"

# 3. off-base-dirty: checkout on an unrelated branch with an uncommitted
#    modification. Expect exit 1; stderr names both the branch found and
#    <base>; HEAD unchanged; nothing staged.
new_checkout
git -C "$WT" checkout -q -b other
echo dirty >> "$WT/file.txt"
run_guard "$WT" main "$ARCHIVE"
expect_refusal "off-base-dirty" 1 "other"
case "$ERR" in
  *main*) pass "off-base-dirty: stderr names 'main'" ;;
  *) fail "off-base-dirty: expected stderr to mention 'main', got '$ERR'" ;;
esac
[ "$(current_branch)" = "other" ] \
  && pass "off-base-dirty: HEAD unchanged" \
  || fail "off-base-dirty: expected HEAD to stay on 'other', got '$(current_branch)'"
[ -z "$(git -C "$WT" diff --cached --name-only)" ] \
  && pass "off-base-dirty: nothing staged" \
  || fail "off-base-dirty: expected nothing staged"

# 4. on-base-dirty: checkout on <base> with an uncommitted modification.
#    Expect exit 1; HEAD unchanged.
new_checkout
echo dirty >> "$WT/file.txt"
run_guard "$WT" main "$ARCHIVE"
expect_refusal "on-base-dirty" 1 "dirty"
[ "$(current_branch)" = "main" ] \
  && pass "on-base-dirty: HEAD unchanged" \
  || fail "on-base-dirty: expected HEAD to stay on 'main', got '$(current_branch)'"

# 5. detached-head: git checkout --detach. Expect exit 1; stderr says HEAD
#    is detached; HEAD unchanged.
new_checkout
git -C "$WT" checkout -q --detach main
DETACHED_SHA="$(git -C "$WT" rev-parse HEAD)"
run_guard "$WT" main "$ARCHIVE"
expect_refusal "detached-head" 1 "detached"
[ -z "$(current_branch)" ] \
  && pass "detached-head: HEAD stays detached" \
  || fail "detached-head: expected HEAD to stay detached, got '$(current_branch)'"
[ "$(git -C "$WT" rev-parse HEAD)" = "$DETACHED_SHA" ] \
  && pass "detached-head: HEAD sha unchanged" \
  || fail "detached-head: expected HEAD sha unchanged"

# 6. archive-branch-exists-descended: <archive-branch> already exists and is
#    descended from origin/<base>. Expect exit 0; the existing branch is
#    reused, not recreated; its commits survive.
new_checkout
git -C "$WT" checkout -q -b "$ARCHIVE" main
echo extra > "$WT/extra.txt"
git -C "$WT" add extra.txt
git -C "$WT" commit -qm extra
EXISTING_SHA="$(git -C "$WT" rev-parse "$ARCHIVE")"
git -C "$WT" checkout -q main
run_guard "$WT" main "$ARCHIVE"
expect_positioned "archive-branch-exists-descended" "main" "$ARCHIVE"
[ "$(current_branch)" = "$ARCHIVE" ] \
  && pass "archive-branch-exists-descended: HEAD is on $ARCHIVE" \
  || fail "archive-branch-exists-descended: expected HEAD on $ARCHIVE, got '$(current_branch)'"
[ "$(git -C "$WT" rev-parse HEAD)" = "$EXISTING_SHA" ] \
  && pass "archive-branch-exists-descended: existing commit survives, not recreated" \
  || fail "archive-branch-exists-descended: expected the existing branch's commit to survive unchanged"

# 7. archive-branch-exists-unrelated: <archive-branch> exists on an
#    unrelated root commit. Expect exit 1; stderr says it is not descended
#    from origin/<base>; HEAD unchanged.
new_checkout
git -C "$WT" checkout -q --orphan "$ARCHIVE"
git -C "$WT" rm -rf -q --cached . >/dev/null
git -C "$WT" clean -fdq
echo unrelated > "$WT/unrelated.txt"
git -C "$WT" add unrelated.txt
git -C "$WT" commit -qm unrelated
git -C "$WT" checkout -q main
run_guard "$WT" main "$ARCHIVE"
expect_refusal "archive-branch-exists-unrelated" 1 "descended"
[ "$(current_branch)" = "main" ] \
  && pass "archive-branch-exists-unrelated: HEAD unchanged" \
  || fail "archive-branch-exists-unrelated: expected HEAD to stay on 'main', got '$(current_branch)'"

# 8. not-a-worktree: a plain directory under TMPDIR. Expect exit 2; nothing
#    on stdout.
PLAIN_DIR="$(mktemp -d "${TMPDIR:-/tmp}/prepare-archive-branch-test.XXXXXX")"
REPOS+=("$PLAIN_DIR")
WT="$PLAIN_DIR"
run_guard "$PLAIN_DIR" main "$ARCHIVE"
expect_refusal "not-a-worktree" 2 "not a git worktree"

# 9. missing-checkout: a path that does not exist, whose parent is not named
#    `.worktrees`. KAN-462 task 3: an absent landing path is no longer an
#    unconditional inability — the guard tries to create it, and refuses
#    (exit 1, naming the path) before ever touching git, because this one's
#    parent does not match the strict `<project>/.worktrees/_landing-<name>`
#    shape. Expect exit 1; nothing on stdout.
MISSING_DIR="${TMPDIR:-/tmp}/prepare-archive-branch-test-missing.$$"
run_guard "$MISSING_DIR" main "$ARCHIVE"
expect_refusal "missing-checkout" 1 "not a .worktrees directory"

# 10. base-diverged: local <base> carries a commit origin/<base> does not,
#     and origin/<base> also advanced independently, so the fast-forward is
#     genuinely impossible in either direction. Expect exit 3; stderr says
#     the fast-forward is impossible; HEAD unchanged.
new_checkout
echo local > "$WT/local.txt"
git -C "$WT" add local.txt
git -C "$WT" commit -qm local
LOCAL_SHA="$(git -C "$WT" rev-parse main)"
echo remote > "$SEED/remote.txt"
git -C "$SEED" add remote.txt
git -C "$SEED" commit -qm remote
git -C "$SEED" push -q origin main
run_guard "$WT" main "$ARCHIVE"
expect_refusal "base-diverged" 3 "fast-forward"
[ "$(current_branch)" = "main" ] \
  && pass "base-diverged: HEAD unchanged" \
  || fail "base-diverged: expected HEAD to stay on 'main', got '$(current_branch)'"
[ "$(git -C "$WT" rev-parse main)" = "$LOCAL_SHA" ] \
  && pass "base-diverged: local base commit unchanged" \
  || fail "base-diverged: expected local base to stay at its diverged commit"

# 11. no-origin: a repository with no origin remote. The guard's own header
#     documents this as exit 3 — the same verdict resolve-base-branch.sh
#     gives a missing origin, and consistent with "the base cannot be
#     fast-forwarded to origin/<base>" when no origin exists to resolve it
#     against at all.
NOORIGIN_WT="$(mktemp -d "${TMPDIR:-/tmp}/prepare-archive-branch-test.XXXXXX")"
REPOS+=("$NOORIGIN_WT")
rm -rf "$NOORIGIN_WT"
git init -q -b main "$NOORIGIN_WT"
git -C "$NOORIGIN_WT" config user.email test@example.invalid
git -C "$NOORIGIN_WT" config user.name "Test"
echo base > "$NOORIGIN_WT/file.txt"
git -C "$NOORIGIN_WT" add file.txt
git -C "$NOORIGIN_WT" commit -qm base
run_guard "$NOORIGIN_WT" main "$ARCHIVE"
expect_refusal "no-origin" 3 "origin"

# --- KAN-462 task 3: the first argument is a throwaway landing worktree, ---
# --- never the main checkout itself. ---

# Declared test: `case: creates the landing worktree when the path is absent`
# 12. creates-landing-worktree: the landing path is absent, and the main
#     checkout (found by walking up two directories from it) sits on <base>,
#     exactly as check-finish-preflight.sh's new main-checkout assertion
#     leaves it. Expect exit 0; the landing worktree is created under
#     <project>/.worktrees/ and positioned on <archive-branch>, cut from a
#     fast-forwarded <base> — the main checkout is untouched throughout.
new_checkout
advance_origin
ORIGIN_TIP="$(git -C "$SEED" rev-parse main)"
MAIN_BEFORE_BRANCH="$(git -C "$WT" branch --show-current)"
LANDING="$WT/.worktrees/_landing-fixture"
run_guard "$LANDING" main "$ARCHIVE"
expect_positioned "creates-landing-worktree" "main" "$ARCHIVE"
[ -d "$LANDING" ] \
  && pass "creates-landing-worktree: the landing worktree now exists" \
  || fail "creates-landing-worktree: expected $LANDING to exist"
[ "$(git -C "$LANDING" branch --show-current)" = "$ARCHIVE" ] \
  && pass "creates-landing-worktree: landing worktree HEAD is on $ARCHIVE" \
  || fail "creates-landing-worktree: expected landing worktree on $ARCHIVE, got '$(git -C "$LANDING" branch --show-current)'"
git -C "$LANDING" merge-base --is-ancestor "$ORIGIN_TIP" HEAD \
  && pass "creates-landing-worktree: archive branch carries the fast-forwarded base" \
  || fail "creates-landing-worktree: archive branch does not carry the fast-forwarded base"
[ "$(git -C "$WT" branch --show-current)" = "$MAIN_BEFORE_BRANCH" ] \
  && pass "creates-landing-worktree: main checkout branch unchanged" \
  || fail "creates-landing-worktree: expected main checkout to stay on '$MAIN_BEFORE_BRANCH', got '$(git -C "$WT" branch --show-current)'"

# Declared test: `case: refuses a dirty landing worktree`
# 13. dirty-landing-worktree: the landing worktree already exists (from a
#     prior invocation) and carries an uncommitted change. Expect exit 1;
#     the main checkout is untouched; the existing worktree is left dirty,
#     never reset out from under whatever left it that way.
new_checkout
LANDING="$WT/.worktrees/_landing-fixture"
git -C "$WT" worktree add --force --quiet "$LANDING" main
echo dirty >> "$LANDING/file.txt"
MAIN_BEFORE_BRANCH="$(git -C "$WT" branch --show-current)"
run_guard "$LANDING" main "$ARCHIVE"
expect_refusal "dirty-landing-worktree" 1 "dirty"
[ "$(git -C "$WT" branch --show-current)" = "$MAIN_BEFORE_BRANCH" ] \
  && pass "dirty-landing-worktree: main checkout branch unchanged" \
  || fail "dirty-landing-worktree: expected main checkout to stay on '$MAIN_BEFORE_BRANCH'"
git -C "$WT" worktree remove --force "$LANDING" 2>/dev/null || true

# Declared test: `case: main checkout is never checked out`
# 14. main-checkout-never-checked-out: the regression case named in the
#     plan's own baseline — reverting this task puts run 2 back in the main
#     checkout, which this proves by asserting `branch --show-current` in
#     the fixture's main checkout is read once, before the call, and found
#     identical after it, for a call whose landing path had to be created.
new_checkout
BEFORE="$(git -C "$WT" branch --show-current)"
LANDING="$WT/.worktrees/_landing-fixture"
run_guard "$LANDING" main "$ARCHIVE"
[ "$RC" -eq 0 ] && pass "main-checkout-never-checked-out: exit 0" \
  || fail "main-checkout-never-checked-out: expected exit 0, got rc=$RC out=[$OUT] err=[$ERR]"
AFTER="$(git -C "$WT" branch --show-current)"
[ "$BEFORE" = "$AFTER" ] \
  && pass "main-checkout-never-checked-out: branch --show-current in the main checkout is unchanged ($BEFORE)" \
  || fail "main-checkout-never-checked-out: main checkout branch changed from '$BEFORE' to '$AFTER'"

# Declared test: `stash appearing mid-run exits 2 naming the stash`
# 15. stash-mid-run: a stash entry injected between the guard's own tree
#     snapshot (the self-check library's first `stash list`) and its
#     post-run recompute (the second) is residue the guard must name and
#     fail on. A PATH-shim git forwards every invocation to the real git
#     and, before forwarding its second `stash list` call, leaves one real
#     stash entry in the landing worktree — what a killed concurrent
#     mutation would leave behind. Expect exit 2; stderr names the entry;
#     stdout stays empty.
new_checkout
LANDING="$WT/.worktrees/_landing-fixture"
git -C "$WT" worktree add --force --quiet "$LANDING" main
SHIM="$(mktemp -d "${TMPDIR:-/tmp}/prepare-archive-branch-shim.XXXXXX")"
REPOS+=("$SHIM")
REAL_GIT="$(command -v git)"
cat > "$SHIM/git" <<'SHIMEOF'
#!/usr/bin/env bash
# Forwards every invocation to the real git. Counts `stash list` calls;
# before forwarding the second (the post-run check's recompute; the first
# is the snapshot's), plants one real stash entry in the landing worktree.
is_stash_list=0
prev=""
for a in "$@"; do
  [ "$prev" = "stash" ] && [ "$a" = "list" ] && is_stash_list=1
  prev="$a"
done
if [ "$is_stash_list" -eq 1 ]; then
  n="$(cat "$PAB_COUNT" 2>/dev/null || echo 0)"
  n=$((n + 1))
  printf '%s\n' "$n" > "$PAB_COUNT"
  if [ "$n" = "2" ]; then
    printf stray > "$PAB_LANDING/stray.txt"
    "$PAB_REAL" -C "$PAB_LANDING" stash push -q --include-untracked -m kan-448-injected
  fi
fi
exec "$PAB_REAL" "$@"
SHIMEOF
chmod +x "$SHIM/git"
PAB_REAL="$REAL_GIT" PAB_COUNT="$SHIM/count" PAB_LANDING="$LANDING" \
  PATH="$SHIM:$PATH" run_guard "$LANDING" main "$ARCHIVE"
if [ "$RC" -eq 2 ] && printf '%s' "$ERR" | grep -q 'new stash entry: .*kan-448-injected'; then
  pass "stash appearing mid-run exits 2 naming the stash"
else
  fail "stash appearing mid-run exits 2 naming the stash (rc=$RC)"
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'prepare-archive-branch: all cases pass\n'
