#!/usr/bin/env bash
# Assertion harness for check-task-commit-planning-paths.sh. Builds
# throwaway git repositories under a sandboxed TMPDIR, makes real commits
# on them — task commits carrying the pipeline's `Task-Id:` trailer, some
# of which deliberately sweep the planning paths — and asserts the guard's
# TASK-COMMIT-SWEEP lines, its verdict lines and its exit status. Never
# touches the real repository tree.
#
# THE GRAMMAR THIS FILE IS THE EXECUTABLE STATEMENT OF (KAN-553):
#
#   TASK-COMMIT-SWEEP: <short sha> <task id> <path>   per planning path a
#                                                     task commit touched
#   PLANNING-PATHS-CLEAN: <worktree> — <n> task commit(s) checked
#   PLANNING-PATHS-SWEPT: <worktree> — <n> task commit(s)
#
# A "task commit" is a commit in <base>..HEAD carrying a `Task-Id:` trailer
# — the commit an implementer makes for a plan task. The planning paths are
# the spec tree's changes directory (leaf resolved per project, `spectre`
# or `openspec`) and `docs/superpowers/`. A task commit touching either is
# the kan-468 defect: staged planning artifacts swept into the task commit.
#
# The exit-code contract is the guard's own header's, verbatim — stated
# once there, cited here rather than copied.
#
# Shape copied from test-check-foreign-staged.sh: sandboxed TMPDIR,
# pass/fail counters, a run_guard capturing stdout and stderr separately.
# Duplicated rather than shared for the same reason that file's header
# gives — the two suites test unrelated guards, and a shared library would
# mean a change to one guard's contract could only be made by editing a file
# the other one also runs.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

GUARD="$SCRIPT_DIR/check-task-commit-planning-paths.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOXES=()
cleanup() {
  if [ "${#SANDBOXES[@]}" -ne 0 ]; then
    for s in "${SANDBOXES[@]}"; do
      chmod -R u+rwX "$s" 2>/dev/null || true
      rm -rf "$s"
    done
  fi
}
trap cleanup EXIT

WORK="$(mktemp -d "${TMPDIR:-/tmp}/task-commit-planning-paths-test.XXXXXX")"
SANDBOXES+=("$WORK")
# Physical form immediately, for the same reason the guard resolves its own
# argument that way: macOS's /tmp is a symlink to /private/tmp, and an
# expectation built from the symlinked mktemp path would never match a
# verdict that names the real one.
WORK="$(cd "$WORK" && pwd -P)"
ERRFILE="$WORK/stderr"

run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# new_repo <name> [leaf] -> prints the repo path: a git repository with one
# commit on `main` and a configured identity, whose spec tree lives under
# <leaf>/changes/ when a leaf other than the spectre default is wanted.
new_repo() {
  local repo="$WORK/$1"
  git init -q -b main "$repo"
  git -C "$repo" config user.email test@example.com
  git -C "$repo" config user.name test
  if [ "${2:-spectre}" != "spectre" ]; then
    mkdir -p "$repo/$2/changes"
  fi
  echo base >"$repo/base.txt"
  git -C "$repo" add base.txt
  git -C "$repo" commit -qm base
  printf '%s\n' "$repo"
}

# task_commit <repo> <message> <task-id> <path>... — one real commit whose
# trailer identifies the task, so the commits here are shaped exactly the
# way the pipeline's implementers shape them.
task_commit() {
  local repo="$1" msg="$2" tid="$3"; shift 3
  mkdir -p "$(dirname "$repo/$1")"
  echo "$1" >"$repo/$1"
  git -C "$repo" add -A
  git -C "$repo" commit -qm "$msg" -m "Task-Id: $tid"
}

# plain_commit <repo> <message> <path>... — one real commit with no trailer,
# the shape of the planning commit bare /flow makes at integrate.
plain_commit() {
  local repo="$1" msg="$2"; shift 2
  mkdir -p "$(dirname "$repo/$1")"
  echo "$1" >"$repo/$1"
  git -C "$repo" add -A
  git -C "$repo" commit -qm "$msg"
}

# ---- usage: no arguments ------------------------------------------------
run_guard
if [ "$RC" -eq 2 ] && [ -z "$OUT" ] && [ -n "$ERR" ]; then
  pass "no arguments: exit 2, stdout empty, usage on stderr"
else
  fail "no arguments: expected exit 2 with empty stdout and non-empty stderr, got RC=$RC OUT=<$OUT>"
fi

# ---- usage: one argument -------------------------------------------------
ONE="$(new_repo one-arg)"
run_guard "$ONE"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ] && [ -n "$ERR" ]; then
  pass "one argument: exit 2, stdout empty, usage on stderr"
else
  fail "one argument: expected exit 2 with empty stdout and non-empty stderr, got RC=$RC OUT=<$OUT>"
fi

# ---- refusal: not a git repository --------------------------------------
NOTGIT="$WORK/not-git"
mkdir -p "$NOTGIT"
run_guard "$NOTGIT" main
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "non-git directory: exit 2, nothing on stdout"
else
  fail "non-git directory: expected exit 2 with empty stdout, got RC=$RC OUT=<$OUT>"
fi

# ---- refusal: base does not resolve --------------------------------------
NOREF="$(new_repo no-ref)"
run_guard "$NOREF" no-such-ref
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "unresolvable base: exit 2, nothing on stdout"
else
  fail "unresolvable base: expected exit 2 with empty stdout, got RC=$RC OUT=<$OUT>"
fi

# ---- refusal: HEAD does not resolve (unborn repository, KAN-608) ---------
# A repository with no commits has no HEAD: the guard's fourth exit-2 branch.
# The repo is built inline, not through new_repo, because new_repo always
# makes the commit whose absence this branch needs. Base `main` is equally
# unresolvable there, so the stderr grep is what pins THIS branch rather than
# the base one.
UNBORN="$WORK/unborn"
git init -q -b main "$UNBORN"
run_guard "$UNBORN" main
if [ "$RC" -eq 2 ] && [ -z "$OUT" ] &&
   printf '%s' "$ERR" | grep -q "HEAD does not resolve"; then
  pass "unborn repository: exit 2, nothing on stdout, HEAD named on stderr"
else
  fail "unborn repository: expected exit 2 with empty stdout and HEAD named on stderr, got RC=$RC OUT=<$OUT>"
fi

# ---- clean: task commit touching implementation only ---------------------
CLEAN="$(new_repo clean)"
BASE_CLEAN="$(git -C "$CLEAN" rev-parse HEAD)"
task_commit "$CLEAN" "feat(app): do the thing" 1 src/app.go
run_guard "$CLEAN" "$BASE_CLEAN"
if [ "$RC" -eq 0 ] && [ "$OUT" = "PLANNING-PATHS-CLEAN: $CLEAN — 1 task commit(s) checked" ]; then
  pass "implementation-only task commit: exit 0, clean verdict names 1 checked"
else
  fail "implementation-only task commit: expected exit 0 and clean verdict, got RC=$RC OUT=<$OUT>"
fi

# ---- the bite: task commit sweeping spectre/changes/ ---------------------
SWEEP="$(new_repo sweep)"
BASE_SWEEP="$(git -C "$SWEEP" rev-parse HEAD)"
task_commit "$SWEEP" "feat(app): do the thing" 1 src/app.go
task_commit "$SWEEP" "wip" 2 "spectre/changes/kan-1/tasks.md"
run_guard "$SWEEP" "$BASE_SWEEP"
if [ "$RC" -eq 1 ] &&
   [ "$OUT" = "TASK-COMMIT-SWEEP: $(git -C "$SWEEP" rev-parse HEAD | cut -c1-12) 2 spectre/changes/kan-1/tasks.md
PLANNING-PATHS-SWEPT: $SWEEP — 1 task commit(s)" ]; then
  pass "spectre/changes/ sweep: exit 1, commit and path named, swept verdict"
else
  fail "spectre/changes/ sweep: expected exit 1 with sweep + swept lines, got RC=$RC OUT=<$OUT>"
fi

# ---- the bite: task commit sweeping docs/superpowers/ --------------------
DOCS="$(new_repo docs)"
BASE_DOCS="$(git -C "$DOCS" rev-parse HEAD)"
task_commit "$DOCS" "fix(app): patch" 3 "docs/superpowers/research/notes.md"
run_guard "$DOCS" "$BASE_DOCS"
if [ "$RC" -eq 1 ] &&
   printf '%s' "$OUT" | grep -q "TASK-COMMIT-SWEEP: .* 3 docs/superpowers/research/notes.md" &&
   printf '%s' "$OUT" | grep -q "PLANNING-PATHS-SWEPT: $DOCS — 1 task commit(s)"; then
  pass "docs/superpowers/ sweep: exit 1, commit and path named, swept verdict"
else
  fail "docs/superpowers/ sweep: expected exit 1 naming the path, got RC=$RC OUT=<$OUT>"
fi

# ---- mixed commit: implementation and planning path in one ---------------
MIXED="$(new_repo mixed)"
BASE_MIXED="$(git -C "$MIXED" rev-parse HEAD)"
mkdir -p "$MIXED/src" "$MIXED/spectre/changes/kan-1"
echo code >"$MIXED/src/app.go"
echo plan >"$MIXED/spectre/changes/kan-1/design.md"
git -C "$MIXED" add -A
git -C "$MIXED" commit -qm "feat(app): do the thing" -m "Task-Id: 4"
run_guard "$MIXED" "$BASE_MIXED"
if [ "$RC" -eq 1 ] &&
   printf '%s' "$OUT" | grep -q "spectre/changes/kan-1/design.md" &&
   printf '%s' "$OUT" | grep -vq "^src/"; then
  pass "mixed commit: exit 1, only the planning path named"
else
  fail "mixed commit: expected exit 1 naming only the planning path, got RC=$RC OUT=<$OUT>"
fi

# ---- not a task commit: trailer-less commit sweeping spectre/changes/ ----
# The planning commit bare /flow makes at integrate sweeps exactly these
# paths and must stay outside this guard's contract.
PLAIN="$(new_repo plain)"
BASE_PLAIN="$(git -C "$PLAIN" rev-parse HEAD)"
plain_commit "$PLAIN" "chore(spectre): plan" "spectre/changes/kan-1/tasks.md"
run_guard "$PLAIN" "$BASE_PLAIN"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q "0 task commit(s) checked"; then
  pass "trailer-less commit: exit 0, zero task commits checked"
else
  fail "trailer-less commit: expected exit 0 with 0 checked, got RC=$RC OUT=<$OUT>"
fi

# ---- specs are implementation: spectre/specs/ never flagged ---------------
SPECS="$(new_repo specs)"
BASE_SPECS="$(git -C "$SPECS" rev-parse HEAD)"
task_commit "$SPECS" "docs(specs): state the rule" 5 "spectre/specs/guards.md"
run_guard "$SPECS" "$BASE_SPECS"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q "1 task commit(s) checked"; then
  pass "spectre/specs/ commit: exit 0, capability specs are implementation"
else
  fail "spectre/specs/ commit: expected exit 0, got RC=$RC OUT=<$OUT>"
fi

# ---- openspec leaf: the changes directory resolves per project ------------
OPEN="$(new_repo openspec openspec)"
BASE_OPEN="$(git -C "$OPEN" rev-parse HEAD)"
task_commit "$OPEN" "feat(app): do the thing" 6 "openspec/changes/kan-1/tasks.md"
run_guard "$OPEN" "$BASE_OPEN"
if [ "$RC" -eq 1 ] &&
   printf '%s' "$OUT" | grep -q "TASK-COMMIT-SWEEP: .* 6 openspec/changes/kan-1/tasks.md"; then
  pass "openspec leaf: exit 1, the project's own changes directory guarded"
else
  fail "openspec leaf: expected exit 1 naming the openspec path, got RC=$RC OUT=<$OUT>"
fi

# ---- walk: a swept task commit below a clean one is still flagged ---------
# Two task commits in the range, the NEWER clean: only the older one sweeps.
# This is the shape a task-close boundary always sees mid-run, and the case
# whose first implementation mis-parsed (the newline git log emits between
# entries landed at the head of the second sha, diff-tree fatalled, and the
# swept commit went unflagged).
WALK="$(new_repo walk)"
BASE_WALK="$(git -C "$WALK" rev-parse HEAD)"
task_commit "$WALK" "wip" 8 "spectre/changes/kan-1/tasks.md"
task_commit "$WALK" "feat(app): do the thing" 9 src/app.go
OLD_WALK_SHA="$(git -C "$WALK" rev-parse HEAD~1 | cut -c1-12)"
run_guard "$WALK" "$BASE_WALK"
if [ "$RC" -eq 1 ] &&
   printf '%s' "$OUT" | grep -q "TASK-COMMIT-SWEEP: $OLD_WALK_SHA 8 spectre/changes/kan-1/tasks.md" &&
   printf '%s' "$OUT" | grep -q "PLANNING-PATHS-SWEPT: $WALK — 1 task commit(s)"; then
  pass "walk: swept commit below a clean one flagged, newer one parsed"
else
  fail "walk: expected exit 1 naming the older commit, got RC=$RC OUT=<$OUT>"
fi

# ---- the bite: a Task-Id evil merge smuggling a planning path -------------
# diff-tree without a merge flag prints nothing for merges, so this merge
# was counted as a task commit and never diffed — the merge answered CLEAN
# over the sweep (KAN-553 F3, deferred; KAN-607). The merge carries the
# planning path in its result, absent from both parents.
EVIL="$(new_repo evil-merge)"
BASE_EVIL="$(git -C "$EVIL" rev-parse HEAD)"
git -C "$EVIL" checkout -q -b side
task_commit "$EVIL" "feat(app): side work" 10 src/side.go
git -C "$EVIL" checkout -q main
git -C "$EVIL" merge --no-ff --no-commit side >/dev/null 2>&1
task_commit "$EVIL" "merge side" 11 "spectre/changes/kan-1/tasks.md"
[ "$(git -C "$EVIL" cat-file -p HEAD | grep -c '^parent ')" -eq 2 ] ||
  fail "evil merge: fixture HEAD is not a merge commit"
MERGE_EVIL_SHA="$(git -C "$EVIL" rev-parse HEAD | cut -c1-12)"
run_guard "$EVIL" "$BASE_EVIL"
if [ "$RC" -eq 1 ] &&
   printf '%s' "$OUT" | grep -q "TASK-COMMIT-SWEEP: $MERGE_EVIL_SHA 11 spectre/changes/kan-1/tasks.md" &&
   printf '%s' "$OUT" | grep -q "PLANNING-PATHS-SWEPT: $EVIL — 1 task commit(s)"; then
  pass "evil merge: exit 1, the merge's smuggled planning path flagged"
else
  fail "evil merge: expected exit 1 naming the merge's smuggled path, got RC=$RC OUT=<$OUT>"
fi

# ---- combined-diff semantics: carried-over planning content is the --------
# parent's, not the merge's. The planning path exists unchanged on the side
# branch (a trailerless commit, outside the contract); the Task-Id merge
# brings it over verbatim, so its combined diff names nothing — a sweep by
# the merge is what the merge itself introduces, absent from every parent.
CARRY="$(new_repo carry-merge)"
BASE_CARRY="$(git -C "$CARRY" rev-parse HEAD)"
git -C "$CARRY" checkout -q -b side
plain_commit "$CARRY" "chore(spectre): plan" "spectre/changes/kan-1/tasks.md"
git -C "$CARRY" checkout -q main
git -C "$CARRY" merge --no-ff --no-commit side >/dev/null 2>&1
task_commit "$CARRY" "feat(app): do the thing" 12 src/app.go
[ "$(git -C "$CARRY" cat-file -p HEAD | grep -c '^parent ')" -eq 2 ] ||
  fail "carry merge: fixture HEAD is not a merge commit"
run_guard "$CARRY" "$BASE_CARRY"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q "1 task commit(s) checked"; then
  pass "carry merge: exit 0, planning content carried over from a parent is the parent's"
else
  fail "carry merge: expected exit 0 with 1 checked, got RC=$RC OUT=<$OUT>"
fi

# ---- range: a swept commit before <base> stays outside the answer ---------
RANGE="$(new_repo range)"
plain_commit "$RANGE" "chore(spectre): plan" "spectre/changes/kan-1/tasks.md"
BEFORE="$(git -C "$RANGE" rev-parse HEAD)"
task_commit "$RANGE" "feat(app): do the thing" 7 src/app.go
run_guard "$RANGE" "$BEFORE"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q "1 task commit(s) checked"; then
  pass "range: swept planning commit before <base> not flagged"
else
  fail "range: expected exit 0 with 1 checked, got RC=$RC OUT=<$OUT>"
fi

# ---- summary --------------------------------------------------------------
if [ "$FAILURES" -eq 0 ]; then
  echo "all checks passed"
  exit 0
fi
printf '%d check(s) failed\n' "$FAILURES" >&2
exit 1
