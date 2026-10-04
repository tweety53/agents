#!/usr/bin/env bash
# test-check-task-records.sh — assertion harness for scripts/check-task-records.sh.
#
# Builds throwaway git fixtures under a sandboxed TMPDIR — each with a
# spectre/changes/<name>/tasks.md, an initial base commit snapshotted to
# refs/remotes/origin/main (with origin/HEAD pointing at it, so the guard's
# default base resolution runs), and task commits on top — and asserts the
# guard's exit status and output per case.
#
# Shape: FAILURES counter, fail()/pass() helpers, an EXIT trap removing every
# fixture dir.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-task-records.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# new_repo -> FIX (repo root) and TASKS (the fixture plan). The base commit
# is recorded at refs/remotes/origin/main with origin/HEAD pointing at it,
# so `base..HEAD` is exactly the commits made after new_repo.
new_repo() {
  FIX="$(mktemp -d "${TMPDIR:-/tmp}/task-records-test.XXXXXX")"
  DIRS+=("$FIX")
  git -C "$FIX" init -q -b main
  git -C "$FIX" config user.email fixture@example.com
  git -C "$FIX" config user.name Fixture
  mkdir -p "$FIX/spectre/changes/fixture-change"
  printf 'base\n' >"$FIX/base.txt"
  git -C "$FIX" add -A
  git -C "$FIX" commit -qm "chore: base commit"
  git -C "$FIX" update-ref refs/remotes/origin/main "$(git -C "$FIX" rev-parse HEAD)"
  git -C "$FIX" symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/main
  TASKS="$FIX/spectre/changes/fixture-change/tasks.md"
  : >"$TASKS"
}

# commit_plan — commit the fixture plan itself, so no task commit below
# ever sweeps spectre/changes/*/tasks.md into its diff.
commit_plan() {
  git -C "$FIX" add -A
  git -C "$FIX" commit -qm "docs(plan): fixture plan"
}

# commit_file <path> <subject> — write, add and commit exactly one file on
# the fixture branch; the plan is never included.
commit_file() {
  local path="$1" subject="$2"
  mkdir -p "$FIX/$(dirname "$path")"
  printf '%s\n' "$path" >"$FIX/$path"
  git -C "$FIX" add "$FIX/$path"
  git -C "$FIX" commit -qm "$subject"
}

# commit_all <subject> — commit every pending file (a fold's several files
# land in the one folded commit).
commit_all() {
  git -C "$FIX" add -A
  git -C "$FIX" commit -qm "$1"
}

# task_block <n> <state> <commit-subject-or-dash> <build> [extra lines...]
# — append one task with Files/Tests/Commit/Build/After fields. The extra
# arguments are appended verbatim as additional lines (Squash-with,
# Allowed-collateral, …).
task_block() {
  local n="$1" state="$2" subject="$3" build="$4"
  shift 4
  {
    printf -- '- [%s] %s. Task %s\n' "$state" "$n" "$n"
    printf '**Files:** `src/f%s.go`\n' "$n"
    printf '**Tests:** none\n'
    if [ "$subject" != "-" ]; then
      printf '**Commit:** `%s`\n' "$subject"
    fi
    printf '**Build:** %s\n' "$build"
    local line
    for line in "$@"; do
      printf '%s\n' "$line"
    done
    printf '**After:** none\n'
    printf '\n'
  } >>"$TASKS"
}

run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# --- case 1: clean — ticked task whose commit exists with matching files,
# --- plus an unticked task with no commit -> exit 0 ---
new_repo
task_block 1 x "feat(a): one" green
task_block 2 ' ' "-" green
commit_plan
commit_file src/f1.go "feat(a): one"
run_guard "$FIX"
if [ "$RC" -eq 0 ] && [ -z "$OUT" ]; then
  pass "case 1 clean plan exits 0 silent"
else
  fail "case 1 clean plan: rc=$RC out=$OUT"
fi

# --- case 2: ticked task, no commit with its subject -> exit 1 ---
new_repo
task_block 1 x "feat(a): missing" green
run_guard "$FIX"
if [ "$RC" -eq 1 ] && grep -q "task 1 ticked but no commit" <<<"$OUT"; then
  pass "case 2 ticked without commit reported"
else
  fail "case 2 ticked without commit: rc=$RC out=$OUT"
fi

# --- case 3: unticked task, a commit with its subject exists -> exit 1 ---
new_repo
task_block 1 ' ' "feat(a): landed" green
commit_plan
commit_file src/f1.go "feat(a): landed"
run_guard "$FIX"
if [ "$RC" -eq 1 ] && grep -q "task 1 not ticked but commit" <<<"$OUT"; then
  pass "case 3 commit without tick reported"
else
  fail "case 3 commit without tick: rc=$RC out=$OUT"
fi

# --- case 4: commit touches a file declared nowhere -> exit 1 ---
new_repo
task_block 1 x "feat(a): sneaky" green
commit_plan
commit_file src/f1.go "feat(a): declared"
commit_file src/undeclared.go "feat(a): sneaky"
run_guard "$FIX"
if [ "$RC" -eq 1 ] && grep -q "src/undeclared.go is not declared" <<<"$OUT"; then
  pass "case 4 undeclared commit file reported"
else
  fail "case 4 undeclared commit file: rc=$RC out=$OUT"
fi

# --- case 5: declared file untouched by the task's commit -> exit 1 ---
new_repo
{
  printf -- '- [x] 1. Task 1\n'
  printf '**Files:** `src/f1.go` `src/f9.go`\n'
  printf '**Tests:** none\n'
  printf '**Commit:** `feat(a): partial`\n'
  printf '**Build:** green\n'
  printf '**After:** none\n'
} >"$TASKS"
commit_plan
commit_file src/f1.go "feat(a): partial"
run_guard "$FIX"
if [ "$RC" -eq 1 ] && grep -q "src/f9.go is not touched" <<<"$OUT"; then
  pass "case 5 declared-but-untouched file reported"
else
  fail "case 5 declared-but-untouched file: rc=$RC out=$OUT"
fi

# --- case 6: extra commit file covered by Allowed-collateral -> exit 0 ---
new_repo
task_block 1 x "feat(a): collateral" green '**Allowed-collateral:** `gen/**`'
commit_plan
{
  mkdir -p "$FIX/src" "$FIX/gen"
  printf 'f1\n' >"$FIX/src/f1.go"
  printf 'gen\n' >"$FIX/gen/generated.go"
  commit_all "feat(a): collateral"
}
run_guard "$FIX"
if [ "$RC" -eq 0 ]; then
  pass "case 6 allowed-collateral glob passes"
else
  fail "case 6 allowed-collateral glob: rc=$RC out=$OUT"
fi

# --- case 7: malformed **Build:** value -> exit 1 ---
new_repo
task_block 1 ' ' "-" yellow
run_guard "$FIX"
if [ "$RC" -eq 1 ] && grep -q "malformed \*\*Build:\*\* tag" <<<"$OUT"; then
  pass "case 7 malformed Build tag reported"
else
  fail "case 7 malformed Build tag: rc=$RC out=$OUT"
fi

# --- case 8: no **Build:** line at all -> exit 1 ---
new_repo
{
  printf -- '- [ ] 1. Task 1\n'
  printf '**Files:** `src/f1.go`\n'
  printf '**Tests:** none\n'
  printf '**After:** none\n'
} >"$TASKS"
run_guard "$FIX"
if [ "$RC" -eq 1 ] && grep -q "has no \*\*Build:\*\* tag" <<<"$OUT"; then
  pass "case 8 missing Build tag reported"
else
  fail "case 8 missing Build tag: rc=$RC out=$OUT"
fi

# --- case 9: no in-flight plan under spectre/changes -> exit 0 silent ---
new_repo
rm -rf "$FIX/spectre"
run_guard "$FIX"
if [ "$RC" -eq 0 ] && [ -z "$OUT" ]; then
  pass "case 9 no plan exits 0 silent"
else
  fail "case 9 no plan: rc=$RC out=$OUT"
fi

# --- case 10: base-ref argument resolving nowhere -> exit 2 COULD NOT JUDGE ---
new_repo
run_guard "$FIX" no-such-ref
if [ "$RC" -eq 2 ] && grep -q "COULD NOT JUDGE" <<<"$OUT"; then
  pass "case 10 unusable base-ref refused exit 2"
else
  fail "case 10 unusable base-ref: rc=$RC out=$OUT"
fi

# --- case 11: red fold — partner and red task share one subject, commit
# --- files cover the union -> exit 0 ---
new_repo
task_block 1 x "feat(fold): folded" green
task_block 2 x "feat(fold): folded" red '**Squash-with:** Task 1' '**After:** Task 1'
commit_plan
{
  mkdir -p "$FIX/src"
  printf 'f1\n' >"$FIX/src/f1.go"
  printf 'f2\n' >"$FIX/src/f2.go"
  commit_all "feat(fold): folded"
}
run_guard "$FIX"
if [ "$RC" -eq 0 ]; then
  pass "case 11 red fold union passes"
else
  fail "case 11 red fold union: rc=$RC out=$OUT"
fi

# --- case 13: one subject carried by two commits -> exit 1 ---
new_repo
task_block 1 x "feat(a): twice" green
commit_plan
commit_file src/f1.go "feat(a): twice"
commit_file src/f9.go "feat(a): twice"
run_guard "$FIX"
if [ "$RC" -eq 1 ] && grep -q "carried by 2 commits" <<<"$OUT"; then
  pass "case 13 duplicate-subject commits reported"
else
  fail "case 13 duplicate-subject commits: rc=$RC out=$OUT"
fi

# --- case 14: landed plan — the base moved past the task's commit, so the
# --- commit is reachable from HEAD but absent from base..HEAD -> exit 0 ---
new_repo
task_block 1 x "feat(a): landed" green
commit_plan
commit_file src/f1.go "feat(a): landed"
git -C "$FIX" update-ref refs/remotes/origin/main "$(git -C "$FIX" rev-parse HEAD)"
run_guard "$FIX"
if [ "$RC" -eq 0 ] && [ -z "$OUT" ]; then
  pass "case 14 landed plan past the base reads clean"
else
  fail "case 14 landed plan: rc=$RC out=$OUT"
fi

# archived_plan <name> <state> <subject> — write a one-task plan under
# spectre/changes/archive/<name>/, uncommitted.
archived_plan() {
  mkdir -p "$FIX/spectre/changes/archive/$1"
  {
    printf -- '- [%s] 1. Task 1\n' "$2"
    printf '**Files:** `src/f1.go`\n'
    printf '**Tests:** none\n'
    printf '**Commit:** `%s`\n' "$3"
    printf '**Build:** green\n'
    printf '**After:** none\n'
  } >"$FIX/spectre/changes/archive/$1/tasks.md"
}

# --- case 15: a plan archived at the base has landed and is not in flight:
# --- its unticked task's commit in base..HEAD is never read -> exit 0 ---
new_repo
archived_plan old-change ' ' "feat(a): ghost"
commit_all "chore(spectre): archive old-change"
git -C "$FIX" update-ref refs/remotes/origin/main "$(git -C "$FIX" rev-parse HEAD)"
commit_file src/f1.go "feat(a): ghost"
run_guard "$FIX"
if [ "$RC" -eq 0 ] && [ -z "$OUT" ]; then
  pass "case 15 plan archived at the base ignored"
else
  fail "case 15 plan archived at the base: rc=$RC out=$OUT"
fi

# --- case 15b: a plan archived on this branch, absent from the base, is in
# --- flight: the same unticked task with its commit in base..HEAD -> exit 1 ---
new_repo
archived_plan new-change ' ' "feat(a): ghost"
commit_all "chore(spectre): archive new-change"
commit_file src/f1.go "feat(a): ghost"
run_guard "$FIX"
if [ "$RC" -eq 1 ] && grep -qF "spectre/changes/archive/new-change/tasks.md: task 1 not ticked" <<<"$OUT"; then
  pass "case 15b plan archived on this branch checked"
else
  fail "case 15b plan archived on this branch: rc=$RC out=$OUT"
fi

# --- case 15c: a ticked task of a plan archived on this branch whose commit
# --- integrate's split collapsed away is not a violation -> exit 0 ---
new_repo
archived_plan new-change x "feat(a): collapsed"
commit_all "chore(spectre): archive new-change"
run_guard "$FIX"
if [ "$RC" -eq 0 ] && [ -z "$OUT" ]; then
  pass "case 15c archived ticked task with a collapsed commit reads clean"
else
  fail "case 15c archived ticked task with a collapsed commit: rc=$RC out=$OUT"
fi

# --- case 12: bare base-ref argument resolving via refs/remotes -> exit 0 ---
new_repo
task_block 1 x "feat(a): twelve" green
commit_plan
commit_file src/f1.go "feat(a): twelve"
run_guard "$FIX" main
if [ "$RC" -eq 0 ]; then
  pass "case 12 bare base-ref argument accepted"
else
  fail "case 12 bare base-ref argument: rc=$RC out=$OUT"
fi

# --- case 16: no base-ref argument, a recorded base resolving nowhere ->
# --- exit 2: the recorded base (branch.<cur>.flowBase) is read before
# --- origin/HEAD ---
new_repo
git -C "$FIX" config "branch.$(git -C "$FIX" branch --show-current).flowBase" no-such-ref
run_guard "$FIX"
if [ "$RC" -eq 2 ] && grep -q "COULD NOT JUDGE" <<<"$OUT"; then
  pass "case 16 recorded base read before origin/HEAD"
else
  fail "case 16 recorded base: rc=$RC out=$OUT"
fi

printf '\n'
# ===========================================================================
# Unlanded archives: the scan cannot list the archived changes the base does
# not carry -> exit 2, never a scan that silently skips them. A copy of the
# guard in a scratch tree with no stats/ beside it cannot build flow-guard;
# the same copy with no lib/flow-guard.sh cannot load it.
# ===========================================================================
U="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/unlanded-archives-test.XXXXXX")" && pwd)"
mkdir -p "$U/scripts" "$U/root/spectre/changes/some-change"
cp -R "$SCRIPT_DIR/lib" "$U/scripts/lib"
cp "$SCRIPT_DIR"/*.py "$SCRIPT_DIR/check-task-records.sh" "$U/scripts/"
set +e
OUT="$("$U/scripts/check-task-records.sh" "$U/root" 2>&1)"
RC=$?
set -e
case "$RC:$OUT" in
  2:*"cannot build flow-guard"*) pass "unlanded archives: flow-guard that cannot be built exits 2" ;;
  *) fail "unlanded archives: flow-guard that cannot be built: rc=$RC out=$OUT" ;;
esac
rm -f "$U/scripts/lib/flow-guard.sh"
set +e
OUT="$("$U/scripts/check-task-records.sh" "$U/root" 2>&1)"
RC=$?
set -e
case "$RC:$OUT" in
  2:*"not found: $U/scripts/lib/flow-guard.sh"*) pass "unlanded archives: a missing lib/flow-guard.sh exits 2" ;;
  *) fail "unlanded archives: a missing lib/flow-guard.sh: rc=$RC out=$OUT" ;;
esac
rm -rf "$U"

if [ "$FAILURES" -eq 0 ]; then
  printf 'all cases passed\n'
else
  printf '%d case(s) failed\n' "$FAILURES"
  exit 1
fi
