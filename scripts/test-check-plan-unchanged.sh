#!/usr/bin/env bash
# Assertion harness for check-plan-unchanged.sh.
#
# Builds throwaway git repositories under TMPDIR and invokes the REAL
# scripts/check-plan-unchanged.sh against them — never a copy of its logic,
# and never a hand-written expected-output string asserted without having
# run the guard to produce it.
#
# The scoping decision the early cases pin: the guard asserts the CHANGE'S
# PLAN TREE (`spectre/changes/<name>/`) and nothing else, because a gated
# per-task reviewer flies BESIDE a working implementer (implement.md's
# bundled dispatch) whose commits and source edits legitimately move the
# rest of the worktree — a whole-worktree assert would fire on every normal
# run. Cases 6 and 7 pin exactly that: an edit outside the plan tree and a
# HEAD move are both clean. Cases 3-5 are the failure shapes the guard
# exists for, case 3 verbatim the KAN-635 incident: uncommitted plan edits
# destroyed by a `git checkout` to HEAD after the snapshot pinned them.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-plan-unchanged.sh"
FAILED=0

WORKTREES=()
SNAPS=()
cleanup() {
  local p
  for p in "${WORKTREES[@]:-}"; do
    [ -n "$p" ] && rm -rf "$p"
  done
  for p in "${SNAPS[@]:-}"; do
    [ -n "$p" ] && rm -rf "$p"
  done
  # The EXIT trap's last status replaces the script's own on this bash —
  # end on success so a green suite exits 0, never the loop's final 1.
  true
}
trap cleanup EXIT

# make_repo -> a git repository with one committed plan file and one
# committed source file. Prints the repo path.
make_repo() {
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-plan-unchanged-test.XXXXXX")" || {
    printf 'make_repo: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  git -C "$wt" init -q
  git -C "$wt" config user.email test@example.com
  git -C "$wt" config user.name test
  mkdir -p "$wt/spectre/changes/demo"
  printf 'plan: original\n' > "$wt/spectre/changes/demo/tasks.md"
  printf 'source: original\n' > "$wt/src.txt"
  git -C "$wt" add -A
  git -C "$wt" commit -q -m original
  printf '%s' "$wt"
}

# new_snap -> a scratch path for a snapshot file. Prints the path.
new_snap() {
  local d
  d="$(mktemp -d "${TMPDIR:-/tmp}/check-plan-unchanged-snap.XXXXXX")" || {
    printf 'new_snap: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  SNAPS+=("$d")
  printf '%s' "$d/snapshot"
}

run_guard() {
  local mode="$1" wt="$2" name="$3" snap="$4"
  "$GUARD" "$mode" "$wt" "$name" "$snap"
}

expect_exit() {
  local label="$1" want="$2"; shift 2
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
  else
    printf 'ok: %s\n' "$label"
  fi
  LAST_OUT="$out"
}

expect_names() {
  local label="$1" needle="$2"
  case "${LAST_OUT:-}" in
    *"$needle"*) printf 'ok: %s\n' "$label" ;;
    *) printf 'FAIL %s: expected %s named in output, got:\n%s\n' "$label" "$needle" "${LAST_OUT:-<empty>}"
       FAILED=1 ;;
  esac
}

# ===========================================================================
# Case 1: snapshot then verify on a clean repo — both exit 0.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
expect_exit 'case 1a: snapshot on a clean repo exits 0' 0 run_guard snapshot "$wt" demo "$snap"
expect_exit 'case 1b: verify on an untouched clean repo exits 0' 0 run_guard verify "$wt" demo "$snap"

# ===========================================================================
# Case 2: a dirty plan tree at snapshot, untouched after — verify exits 0.
# The guard pins whatever state it saw; the parent's pre-dispatch commit is
# what SHOULD have cleaned the tree, but a dirty state honestly carried
# across the dispatch is not a violation.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
printf 'plan: edited\n' >> "$wt/spectre/changes/demo/tasks.md"
expect_exit 'case 2a: snapshot with a dirty plan tree exits 0' 0 run_guard snapshot "$wt" demo "$snap"
expect_exit 'case 2b: verify with the same dirty plan tree exits 0' 0 run_guard verify "$wt" demo "$snap"

# ===========================================================================
# Case 3: THE KAN-635 SHAPE — a dirty plan tree at snapshot, then the
# uncommitted edits destroyed by a checkout to HEAD. Verify exits 1 and
# names the destroyed file.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
printf 'plan: edited\n' >> "$wt/spectre/changes/demo/tasks.md"
expect_exit 'case 3a: snapshot with a dirty plan tree exits 0' 0 run_guard snapshot "$wt" demo "$snap"
git -C "$wt" checkout -- spectre/changes/demo
expect_exit 'case 3b: verify after the edits are checked away exits 1' 1 run_guard verify "$wt" demo "$snap"
expect_names 'case 3c: names the destroyed plan file' 'tasks.md'

# ===========================================================================
# Case 4: a plan file deleted after the snapshot — verify exits 1.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
expect_exit 'case 4a: snapshot exits 0' 0 run_guard snapshot "$wt" demo "$snap"
rm "$wt/spectre/changes/demo/tasks.md"
expect_exit 'case 4b: verify after a plan deletion exits 1' 1 run_guard verify "$wt" demo "$snap"
expect_names 'case 4c: names the deleted plan file' 'tasks.md'

# ===========================================================================
# Case 5: an untracked file appears under the plan tree after the snapshot
# — verify exits 1 (untracked plan files count: nothing legitimate creates
# one mid-flight, and a planted file is exactly the shape a reviewer
# dispatch must not leave behind).
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
expect_exit 'case 5a: snapshot exits 0' 0 run_guard snapshot "$wt" demo "$snap"
printf 'planted\n' > "$wt/spectre/changes/demo/planted.md"
expect_exit 'case 5b: verify after a planted plan file exits 1' 1 run_guard verify "$wt" demo "$snap"
expect_names 'case 5c: names the planted file' 'planted.md'

# ===========================================================================
# Case 6: an edit OUTSIDE the plan tree after the snapshot — verify exits
# 0. The concurrent implementer's legitimate work is not this guard's
# business; a whole-worktree assert would fire on every normal run.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
expect_exit 'case 6a: snapshot exits 0' 0 run_guard snapshot "$wt" demo "$snap"
printf 'source: edited\n' >> "$wt/src.txt"
git -C "$wt" add src.txt
git -C "$wt" commit -q -m implementer-work
expect_exit 'case 6b: verify ignores work outside the plan tree' 0 run_guard verify "$wt" demo "$snap"

# ===========================================================================
# Case 7: HEAD moves after the snapshot with the plan tree untouched —
# verify exits 0. The guard pins plan-tree state, never HEAD: the bundled
# implementer commits beside the reviewer legitimately.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
expect_exit 'case 7a: snapshot exits 0' 0 run_guard snapshot "$wt" demo "$snap"
printf 'more\n' > "$wt/other.txt"
git -C "$wt" add other.txt
git -C "$wt" commit -q -m head-moves
expect_exit 'case 7b: verify ignores a HEAD move' 0 run_guard verify "$wt" demo "$snap"

# ===========================================================================
# Case 8: verify with no snapshot file — exit 2, cannot answer. A snapshot
# runs before the dispatch, never after.
# ===========================================================================
wt="$(make_repo)"
snap="$(new_snap)/absent"
expect_exit 'case 8: verify without a snapshot file exits 2' 2 run_guard verify "$wt" demo "$snap"

# ===========================================================================
# Case 9: usage errors — no arguments, a bad mode, a non-directory
# worktree, a missing snapshot path, and a change name carrying a path
# separator (the containment case) — all exit 2.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
expect_exit 'case 9a: no arguments exits 2' 2 "$GUARD"
expect_exit 'case 9b: a bad mode exits 2' 2 run_guard check "$wt" demo "$snap"
expect_exit 'case 9c: a non-directory worktree exits 2' 2 run_guard snapshot "$wt/no-such-dir" demo "$snap"
expect_exit 'case 9d: a change name with a separator exits 2' 2 run_guard snapshot "$wt" ../escape "$snap"

# ===========================================================================
# Case 10: snapshot into a directory that does not exist — exit 2, cannot
# answer, never a silent skip.
# ===========================================================================
wt="$(make_repo)"
snap="$(new_snap)/no-dir/snapshot"
expect_exit 'case 10: snapshot to an unwritable path exits 2' 2 run_guard snapshot "$wt" demo "$snap"

# ===========================================================================
# Case 11: the worktree is a directory but not a git repository — exit 2.
# ===========================================================================
notrepo="$(mktemp -d "${TMPDIR:-/tmp}/check-plan-unchanged-notrepo.XXXXXX")"
WORKTREES+=("$notrepo")
snap="$(new_snap)"
expect_exit 'case 11: a non-repository worktree exits 2' 2 run_guard snapshot "$notrepo" demo "$snap"

# ===========================================================================
# Case 12: the plan directory absent at snapshot and still absent at
# verify — both exit 0. A change with no plan tree (flow-fast marks the
# reviewers' stages empty) has nothing to lose.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
git -C "$wt" rm -rq spectre
git -C "$wt" commit -q -m plan-removed
expect_exit 'case 12a: snapshot with no plan tree exits 0' 0 run_guard snapshot "$wt" demo "$snap"
expect_exit 'case 12b: verify with the plan tree still absent exits 0' 0 run_guard verify "$wt" demo "$snap"

# ===========================================================================
# Case 13: a plan tree that appears after its absent-at-snapshot state —
# verify exits 1 and names it.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
git -C "$wt" rm -rq spectre
git -C "$wt" commit -q -m plan-removed
expect_exit 'case 13a: snapshot with no plan tree exits 0' 0 run_guard snapshot "$wt" demo "$snap"
mkdir -p "$wt/spectre/changes/demo"
printf 'new\n' > "$wt/spectre/changes/demo/tasks.md"
expect_exit 'case 13b: verify after the plan tree appears exits 1' 1 run_guard verify "$wt" demo "$snap"

if [ "$FAILED" -ne 0 ]; then
  exit 1
fi
exit 0
