#!/usr/bin/env bash
# Assertion harness for reshape-branch.sh, run together with commit-split.sh
# exactly as integrate runs them. Builds sandboxed git repositories under
# TMPDIR; never touches the real repository tree.
#
# Asserts: every planning commit — plan, link, a later plan — survives the
# reshape as its own commit, in order, with its message and author, carrying
# only the change folder; task and fixup commits collapse into the one
# implementation commit, which carries no planning path; the uncommitted
# planning delta lands as the last planning commit; the final tree equals
# what the branch plus the working tree held before the reshape; a branch
# with no planning commit reshapes exactly as `reset --soft` did; and the
# location guard refuses a main checkout without moving HEAD.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESHAPE="$SCRIPT_DIR/reshape-branch.sh"
SPLIT="$SCRIPT_DIR/commit-split.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/reshape-branch-test.XXXXXX")"
trap 'rm -rf "$SANDBOX"' EXIT

# new_repo <dir> -> a repo on main with one seed commit, and a change
# worktree at <dir>/.worktrees/demo on spectre/demo; sets MAIN, WT, BASE.
new_repo() {
  MAIN="$1"
  git init -q -b main "$MAIN"
  git -C "$MAIN" config user.email "test@example.com"
  git -C "$MAIN" config user.name "Test"
  mkdir -p "$MAIN/spectre/changes/demo" "$MAIN/src"
  printf 'seed\n' > "$MAIN/src/a.txt"
  printf 'other change\n' > "$MAIN/spectre/changes/seed.md"
  git -C "$MAIN" add -A
  git -C "$MAIN" commit -q -m seed
  BASE="$(git -C "$MAIN" rev-parse HEAD)"
  WT="$MAIN/.worktrees/demo"
  git -C "$MAIN" worktree add -q -b spectre/demo "$WT"
  mkdir -p "$WT/spectre/changes/demo"
}

# commit_as <author> <message> <paths...> — a pathspec-scoped commit.
commit_as() {
  local author="$1" msg="$2"
  shift 2
  git -C "$WT" add -A -- "$@"
  git -C "$WT" -c user.name="$author" -c user.email="$author@example.com" \
    commit -q -m "$msg" -- "$@"
}

# expected_tree -> the tree the branch plus working tree holds right now.
expected_tree() {
  local idx="$SANDBOX/expected-index"
  cp "$(git -C "$WT" rev-parse --path-format=absolute --git-dir)/index" "$idx"
  GIT_INDEX_FILE="$idx" git -C "$WT" add -A
  GIT_INDEX_FILE="$idx" git -C "$WT" write-tree
  rm -f "$idx"
}

# ===========================================================================
# 1. Planning commits survive; task and fixup commits collapse.
# ===========================================================================
new_repo "$SANDBOX/one"
printf 'proposal\n' > "$WT/spectre/changes/demo/proposal.md"
commit_as planner "chore(spectre): plan" spectre/changes/demo
printf 'task 1\n' > "$WT/src/a.txt"
commit_as impl $'feat(src): task 1\n\nTask-Id: 1' src
printf 'link\n' > "$WT/spectre/changes/demo/link.md"
commit_as linker "chore(spectre): link peer" spectre/changes/demo/link.md
printf 'task 2\n' > "$WT/src/b.txt"
commit_as impl $'feat(src): task 2\n\nTask-Id: 2' src
printf 'fixup\n' >> "$WT/src/b.txt"
commit_as impl "fixup! feat(src): task 2" src
printf 'tasks ticked\n' > "$WT/spectre/changes/demo/tasks.md"
commit_as reviewer "chore(spectre): plan" spectre/changes/demo
printf 'operator edit\n' >> "$WT/src/a.txt"
printf 'narrative\n' > "$WT/spectre/changes/demo/narrative.md"
WANT_TREE="$(expected_tree)"

set +e
OUT="$("$RESHAPE" "$WT" demo "$BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] || fail "case 1: reshape rc=$RC out=$OUT"
case "$OUT" in
  *"RESHAPED: $WT — 3 planning commit(s) kept on ${BASE:0:12}"*) pass "case 1: reshape verdict names three kept commits" ;;
  *) fail "case 1: verdict: $OUT" ;;
esac
"$SPLIT" "$WT" demo "feat(src): the change" $'chore(spectre): plan\n\noutstanding: none' >/dev/null

SUBJECTS="$(git -C "$WT" log --reverse --format=%s "$BASE..HEAD" | tr '\n' '|')"
WANT='chore(spectre): plan|chore(spectre): link peer|chore(spectre): plan|feat(src): the change|chore(spectre): plan|'
[ "$SUBJECTS" = "$WANT" ] && pass "case 1: planning commits kept, in order, around one implementation commit" \
  || fail "case 1: subjects: $SUBJECTS"

AUTHORS="$(git -C "$WT" log --reverse --format=%an "$BASE..HEAD~2" | tr '\n' ' ')"
[ "$AUTHORS" = "planner linker reviewer " ] && pass "case 1: planning commits keep their authors" \
  || fail "case 1: authors: $AUTHORS"

[ "$(git -C "$WT" rev-parse 'HEAD^{tree}')" = "$WANT_TREE" ] \
  && pass "case 1: final tree equals the pre-reshape branch plus working tree" \
  || fail "case 1: final tree differs"

LEAK=0
for c in $(git -C "$WT" rev-list "$BASE..HEAD"); do
  paths="$(git -C "$WT" diff-tree --no-commit-id --name-only -r "$c")"
  subject="$(git -C "$WT" log -1 --format=%s "$c")"
  case "$subject" in
    chore*) printf '%s\n' "$paths" | grep -qv '^spectre/changes/demo/' && LEAK=1 ;;
    feat*) printf '%s\n' "$paths" | grep -q '^spectre/changes/' && LEAK=1 ;;
  esac
done
[ "$LEAK" -eq 0 ] && pass "case 1: each commit carries only its own side of the split" \
  || fail "case 1: a commit carries the other side's paths"

git -C "$WT" show --name-only --format= "HEAD" | grep -qx 'spectre/changes/demo/narrative.md' \
  && pass "case 1: uncommitted planning delta is the last planning commit" \
  || fail "case 1: narrative not in the last commit"
git -C "$WT" log -1 --format=%B HEAD | grep -q 'outstanding: none' \
  && pass "case 1: last planning commit carries integrate's message" \
  || fail "case 1: last planning message lost"

# ===========================================================================
# 2. No planning commits: the reshape is a plain reset --soft.
# ===========================================================================
new_repo "$SANDBOX/two"
printf 'task\n' > "$WT/src/a.txt"
commit_as impl $'feat(src): task\n\nTask-Id: 1' src
OUT="$("$RESHAPE" "$WT" demo "$BASE")"
case "$OUT" in
  *"0 planning commit(s)"*) pass "case 2: nothing kept" ;;
  *) fail "case 2: verdict: $OUT" ;;
esac
[ "$(git -C "$WT" rev-parse HEAD)" = "$BASE" ] && pass "case 2: HEAD is the merge base" \
  || fail "case 2: HEAD moved elsewhere"
git -C "$WT" diff --cached --quiet -- src/a.txt && fail "case 2: task work lost from the index" \
  || pass "case 2: task work kept staged"

# ===========================================================================
# 3. The guard refuses a main checkout and HEAD does not move.
# ===========================================================================
new_repo "$SANDBOX/three"
BEFORE="$(git -C "$MAIN" rev-parse HEAD)"
set +e
OUT="$("$RESHAPE" "$MAIN" demo "$BASE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 1 ] && pass "case 3: main checkout refused with exit 1" || fail "case 3: rc=$RC"
case "$OUT" in
  *PLANNING-COMMIT-MAIN-CHECKOUT*) pass "case 3: guard's own line printed" ;;
  *) fail "case 3: out=$OUT" ;;
esac
[ "$(git -C "$MAIN" rev-parse HEAD)" = "$BEFORE" ] && pass "case 3: HEAD untouched" || fail "case 3: HEAD moved"

# ===========================================================================
# 4. A base that does not resolve cannot answer.
# ===========================================================================
set +e
"$RESHAPE" "$WT" demo no-such-ref >/dev/null 2>&1
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 4: unresolvable base exits 2" || fail "case 4: rc=$RC"

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all reshape-branch.sh cases passed\n'
